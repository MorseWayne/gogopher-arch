package evaluation

import (
	"context"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
	"github.com/MorseWayne/gogopher-arch/internal/sandbox"
)

func TestBlankEndpointAuditVariantPassesRealReleaseAndSandboxEvaluation(t *testing.T) {
	registry := draftReleaseRegistry(t)
	activity, err := registry.ActivityView(registry.CurrentReleaseID(), "review-endpoint-audit", 2)
	if err != nil {
		t.Fatal(err)
	}
	task, err := registry.ExecutionTask(registry.CurrentReleaseID(), activity.TaskRef.ID, activity.TaskRef.Version)
	if err != nil {
		t.Fatal(err)
	}
	workspace, err := registry.PublicWorkspace(registry.CurrentReleaseID(), task.ID, task.Version)
	if err != nil {
		t.Fatal(err)
	}
	for path, source := range endpointAuditProjectSolution() {
		workspace[path] = source
	}
	current := attempt.Attempt{
		ID: "00000000-0000-4000-8900-000000000001", ReleaseID: registry.CurrentReleaseID(),
		ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash,
		TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
		Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace),
	}
	builder, err := execution.NewSpecBuilder(registry)
	if err != nil {
		t.Fatal(err)
	}
	executionID := "00000000-0000-4000-8900-000000000002"
	specification, err := builder.Build(current, executionID, execution.ActionSubmit)
	if err != nil {
		t.Fatal(err)
	}
	response, err := sandbox.NewRunner(sandbox.RunnerOptions{TempDir: t.TempDir()}).Run(context.Background(), specification)
	if err != nil {
		t.Fatal(err)
	}
	if response.Status != execution.ExecutionSucceeded {
		t.Fatalf("endpointaudit sandbox response = %#v", response)
	}
	frozen := submission.Submission{
		ID: "00000000-0000-4000-8900-000000000003", AttemptID: current.ID,
		ReleaseID: current.ReleaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
		Workspace:   workspace,
		Explanation: "我重新按 package 划分 spec、probe、output 与 app，让 context 同时约束派发过程和每个 HTTP 请求；使用 httptest 覆盖状态不符、请求错误、取消和稳定输出，再由 main 将 Run 的结果转换为 exit code。并发实现只启动固定数量的 worker，关闭响应体并等待全部协程退出，最终按照配置顺序汇总结果，避免完成顺序改变命令行输出。",
	}
	terminal := execution.Execution{
		ID: executionID, AttemptID: current.ID, SubmissionID: frozen.ID,
		TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
		Status: response.Status, Response: &response,
	}
	generator, err := NewGenerator(registry)
	if err != nil {
		t.Fatal(err)
	}
	results, err := generator.Generate(frozen, terminal)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		if result.Status != execution.RulePassed {
			t.Fatalf("endpointaudit rule %s = %#v", result.RuleID, result)
		}
	}
}

func endpointAuditProjectSolution() map[string]string {
	return map[string]string{
		"go.mod": `module endpointaudit

go 1.25
`,
		"README.md": `# endpointaudit

Build with go build ./cmd/endpointaudit and test with go test ./....
`,
		"examples/checks.json": `{"checks":[{"name":"health","url":"http://127.0.0.1:8080/healthz","accept_status":200}]}
`,
		"internal/spec/spec.go": `package spec

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Check struct {
	Name         string ` + "`json:\"name\"`" + `
	URL          string ` + "`json:\"url\"`" + `
	AcceptStatus int    ` + "`json:\"accept_status\"`" + `
}

type Document struct {
	Checks []Check ` + "`json:\"checks\"`" + `
}

func Load(path string) (Document, error) {
	file, err := os.Open(path)
	if err != nil {
		return Document{}, fmt.Errorf("open spec %q: %w", path, err)
	}
	defer file.Close()
	var document Document
	if err := json.NewDecoder(file).Decode(&document); err != nil {
		return Document{}, fmt.Errorf("decode spec %q: %w", path, err)
	}
	if len(document.Checks) == 0 {
		return Document{}, fmt.Errorf("at least one check is required")
	}
	seen := make(map[string]struct{}, len(document.Checks))
	for index := range document.Checks {
		check := &document.Checks[index]
		check.Name = strings.TrimSpace(check.Name)
		check.URL = strings.TrimSpace(check.URL)
		parsed, err := url.ParseRequestURI(check.URL)
		if check.Name == "" || err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
			return Document{}, fmt.Errorf("invalid check %q", check.Name)
		}
		if check.AcceptStatus < 100 || check.AcceptStatus > 599 {
			return Document{}, fmt.Errorf("invalid status for %q", check.Name)
		}
		if _, exists := seen[check.Name]; exists {
			return Document{}, fmt.Errorf("duplicate check %q", check.Name)
		}
		seen[check.Name] = struct{}{}
	}
	return document, nil
}
`,
		"internal/probe/probe.go": `package probe

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

type Check struct {
	Name         string
	URL          string
	AcceptStatus int
}

type Result struct {
	Name     string ` + "`json:\"name\"`" + `
	Expected int    ` + "`json:\"expected\"`" + `
	Actual   int    ` + "`json:\"actual\"`" + `
	Error    string ` + "`json:\"error,omitempty\"`" + `
}

type job struct {
	index int
	check Check
}

func All(ctx context.Context, client Client, checks []Check, workers int, timeout time.Duration) ([]Result, error) {
	if workers <= 0 || timeout <= 0 {
		return nil, fmt.Errorf("workers and timeout must be positive")
	}
	results := make([]Result, len(checks))
	jobs := make(chan job)
	var group sync.WaitGroup
	group.Add(workers)
	for range workers {
		go func() {
			defer group.Done()
			for item := range jobs {
				result := Result{Name: item.check.Name, Expected: item.check.AcceptStatus}
				requestContext, cancel := context.WithTimeout(ctx, timeout)
				request, err := http.NewRequestWithContext(requestContext, http.MethodGet, item.check.URL, nil)
				if err == nil {
					var response *http.Response
					response, err = client.Do(request)
					if response != nil {
						result.Actual = response.StatusCode
						response.Body.Close()
					}
				}
				cancel()
				if err != nil {
					result.Error = err.Error()
				}
				results[item.index] = result
			}
		}()
	}
	dispatching := true
	for index, check := range checks {
		select {
		case jobs <- job{index: index, check: check}:
		case <-ctx.Done():
			dispatching = false
		}
		if !dispatching {
			break
		}
	}
	close(jobs)
	group.Wait()
	if err := ctx.Err(); err != nil {
		return results, err
	}
	return results, nil
}
`,
		"internal/output/output.go": `package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"endpointaudit/internal/probe"
)

func Text(results []probe.Result) string {
	var output strings.Builder
	for _, result := range results {
		status := "error"
		if result.Error == "" {
			status = "fail"
			if result.Actual == result.Expected {
				status = "pass"
			}
		}
		fmt.Fprintf(&output, "%s\t%s\t%d\t%d\n", result.Name, status, result.Expected, result.Actual)
	}
	return output.String()
}

func JSON(results []probe.Result) (string, error) {
	encoded, err := json.Marshal(results)
	if err != nil {
		return "", err
	}
	return string(encoded) + "\n", nil
}
`,
		"internal/app/app.go": `package app

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"endpointaudit/internal/output"
	"endpointaudit/internal/probe"
	"endpointaudit/internal/spec"
)

type Dependencies struct {
	Client probe.Client
}

func Run(ctx context.Context, args []string, stdout, stderr io.Writer, dependencies Dependencies) int {
	flags := flag.NewFlagSet("endpointaudit", flag.ContinueOnError)
	flags.SetOutput(stderr)
	specPath := flags.String("spec", "", "path to check specification")
	deadline := flags.Duration("deadline", time.Second, "per-check deadline")
	workers := flags.Int("workers", 4, "maximum concurrent probes")
	format := flags.String("output", "text", "text or json")
	if err := flags.Parse(args); err != nil || *specPath == "" || *deadline <= 0 || *workers <= 0 || (*format != "text" && *format != "json") {
		fmt.Fprintln(stderr, "invalid arguments")
		return 2
	}
	document, err := spec.Load(*specPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	checks := make([]probe.Check, len(document.Checks))
	for index, check := range document.Checks {
		checks[index] = probe.Check{Name: check.Name, URL: check.URL, AcceptStatus: check.AcceptStatus}
	}
	client := dependencies.Client
	if client == nil {
		client = http.DefaultClient
	}
	results, err := probe.All(ctx, client, checks, *workers, *deadline)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	rendered := ""
	if *format == "json" {
		rendered, err = output.JSON(results)
	} else {
		rendered = output.Text(results)
	}
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if _, err := io.WriteString(stdout, rendered); err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	for _, result := range results {
		if result.Error != "" || result.Actual != result.Expected {
			return 1
		}
	}
	return 0
}
`,
		"cmd/endpointaudit/main.go": `package main

import (
	"context"
	"os"
	"os/signal"

	"endpointaudit/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	os.Exit(app.Run(ctx, os.Args[1:], os.Stdout, os.Stderr, app.Dependencies{}))
}
`,
		"internal/output/output_learner_test.go": `package output

import (
	"testing"

	"endpointaudit/internal/probe"
)

func TestTextCases(t *testing.T) {
	tests := []struct {
		name   string
		result probe.Result
		want   string
	}{
		{name: "pass", result: probe.Result{Name: "api", Expected: 204, Actual: 204}, want: "api\tpass\t204\t204\n"},
		{name: "fail", result: probe.Result{Name: "api", Expected: 200, Actual: 503}, want: "api\tfail\t200\t503\n"},
		{name: "error", result: probe.Result{Name: "api", Expected: 200, Error: "down"}, want: "api\terror\t200\t0\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Text([]probe.Result{test.result}); got != test.want {
				t.Fatalf("Text() = %q, want %q", got, test.want)
			}
		})
	}
}
`,
	}
}
