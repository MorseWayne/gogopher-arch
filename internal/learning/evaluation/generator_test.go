package evaluation

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
	"github.com/MorseWayne/gogopher-arch/internal/sandbox"
)

func TestGeneratorProducesMixedRuleResultsFromExecutedStages(t *testing.T) {
	generator, frozen := setupGenerator(t)
	terminal := terminalExecution(frozen, execution.ExecutionUserFailed, []execution.StageResult{
		passedStage(execution.StageBuild),
		passedStage(execution.StageVet),
		{
			Stage: execution.StageVisibleTest, Status: execution.StageFailed, ExitCode: 1,
			PublicSummary: "visible tests failed",
			TestEvents: []execution.TestEvent{
				{Action: "run", Package: "assessment/internal/config", Test: "TestNormalizeSortsTargets"},
				{Action: "pass", Package: "assessment/internal/config", Test: "TestNormalizeSortsTargets"},
				{Action: "fail", Package: "assessment/internal/config", Test: "TestOther"},
			},
		},
	})
	results, err := generator.Generate(frozen, terminal)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]execution.RuleStatus{
		"module-builds":          execution.RulePassed,
		"toolchain-checks-pass":  execution.RulePassed,
		"error-chain-preserved":  execution.RuleNotEvaluated,
		"resource-closed":        execution.RulePassed,
		"invalid-input-rejected": execution.RuleNotEvaluated,
		"stable-output":          execution.RulePassed,
		"cli-failure-contract":   execution.RuleNotEvaluated,
		"learner-tests-present":  execution.RuleFailed,
		"visible-tests-pass":     execution.RuleFailed,
		"held-out-tests-pass":    execution.RuleNotEvaluated,
	}
	assertRuleStatuses(t, results, want)
	if got := resultByID(t, results, "stable-output"); got.Test != "TestNormalizeSortsTargets" {
		t.Fatalf("stable-output selector = %#v", got)
	}
	if got := resultByID(t, results, "resource-closed"); got.Analyzer != "go_ast_deferred_call" {
		t.Fatalf("resource-closed analyzer = %q", got.Analyzer)
	}
}

func TestGeneratorPassesAllRulesWithFullStructuredResponse(t *testing.T) {
	generator, frozen := setupGenerator(t)
	terminal := terminalExecution(frozen, execution.ExecutionSucceeded, []execution.StageResult{
		passedStage(execution.StageBuild),
		passedStage(execution.StageVet),
		{
			Stage: execution.StageVisibleTest, Status: execution.StagePassed,
			TestEvents: []execution.TestEvent{
				{Action: "pass", Package: "assessment/internal/config", Test: "TestNormalizeSortsTargets"},
			},
		},
		{
			Stage: execution.StageHeldOutTest, Status: execution.StagePassed,
			TestEvents: []execution.TestEvent{
				{Action: "pass", Package: "assessment/internal/config", Test: "TestLoadRejectsInvalidSchemeAndPreservesPathErrors"},
			},
		},
	})
	results, err := generator.Generate(frozen, terminal)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 10 {
		t.Fatalf("len(results) = %d, want 10", len(results))
	}
	for _, result := range results {
		if result.Status != execution.RulePassed {
			t.Fatalf("result %s = %#v, want passed", result.RuleID, result)
		}
	}
}

func TestGeneratorMarksRulesAfterBuildFailureNotEvaluated(t *testing.T) {
	generator, frozen := setupGenerator(t)
	terminal := terminalExecution(frozen, execution.ExecutionUserFailed, []execution.StageResult{{
		Stage: execution.StageBuild, Status: execution.StageFailed, ExitCode: 1,
		PublicSummary: "build failed",
	}})
	results, err := generator.Generate(frozen, terminal)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		want := execution.RuleNotEvaluated
		if result.RuleID == "module-builds" {
			want = execution.RuleFailed
		}
		if result.Status != want {
			t.Fatalf("result %s = %q, want %q", result.RuleID, result.Status, want)
		}
	}
}

func TestGeneratorDoesNotFailNamedRuleWhenItsTestNeverRan(t *testing.T) {
	rule := definition.AssessmentRule{
		RuleID: "targeted-test", Stage: string(execution.StageVisibleTest),
		Selector: definition.AssessmentSelector{Test: "TestTarget"}, Condition: "passed",
	}
	status := evaluateExecutionRule(rule, map[execution.Stage]execution.StageResult{
		execution.StageVisibleTest: {
			Stage: execution.StageVisibleTest, Status: execution.StageFailed, ExitCode: 1,
			TestEvents: []execution.TestEvent{{Action: "fail", Package: "task", Test: "TestBeforeTarget"}},
		},
	})
	if status != execution.RuleNotEvaluated {
		t.Fatalf("targeted test status = %q, want not_evaluated", status)
	}
}

func TestGeneratorRejectsInfrastructureFailure(t *testing.T) {
	generator, frozen := setupGenerator(t)
	terminal := terminalExecution(frozen, execution.ExecutionInfraFailed, nil)
	terminal.Response.Failure = &execution.Failure{Code: "sandbox_unreachable", Message: "Sandbox unavailable"}
	if _, err := generator.Generate(frozen, terminal); err == nil {
		t.Fatal("Generate(infra failure) error = nil")
	}
}

func TestDocumentedExportsAnalyzerRequiresPackageAndDeclarationComments(t *testing.T) {
	valid := `// Package report creates stable reports.
package report

// Summary describes a report.
type Summary struct{}

// Build creates a Summary.
func Build() Summary { return Summary{} }

// Count returns the number of entries.
func (Summary) Count() int { return 0 }
`
	if !hasDocumentedExports(valid) {
		t.Fatal("hasDocumentedExports(valid) = false")
	}
	for name, source := range map[string]string{
		"package":  strings.Replace(valid, "// Package report creates stable reports.\n", "", 1),
		"type":     strings.Replace(valid, "// Summary describes a report.\n", "", 1),
		"function": strings.Replace(valid, "// Build creates a Summary.\n", "// Create builds a Summary.\n", 1),
		"method":   strings.Replace(valid, "// Count returns the number of entries.\n", "", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if hasDocumentedExports(source) {
				t.Fatalf("hasDocumentedExports(%s without valid comment) = true", name)
			}
		})
	}
}

func TestInterfaceAndGenericASTAnalyzersRequireExplicitContracts(t *testing.T) {
	valid := `package service

type Notifier interface { Notify(string) error }

func IndexBy[T any, K comparable](values []T, key func(T) K) map[K]T { return nil }
`
	if !hasMinimalInterface(valid, "Notifier", 1) {
		t.Fatal("hasMinimalInterface(valid) = false")
	}
	if !hasGenericFunction(valid, "IndexBy") {
		t.Fatal("hasGenericFunction(valid) = false")
	}
	tooWide := `package service; type Notifier interface { Notify(string) error; Close() error }`
	if hasMinimalInterface(tooWide, "Notifier", 1) {
		t.Fatal("hasMinimalInterface(tooWide) = true")
	}
	nongeneric := `package service; func IndexBy(values []string) map[string]string { return nil }`
	if hasGenericFunction(nongeneric, "IndexBy") {
		t.Fatal("hasGenericFunction(nongeneric) = true")
	}
}

func TestExplanationSelectorRequiresLengthAndEvidenceTerms(t *testing.T) {
	selector := definition.AssessmentSelector{MinimumChars: 20, RequiredTerms: []string{"alloc_space", "strings.Builder"}}
	if !explanationSatisfies("根据 alloc_space profile，我选择 strings.Builder 降低分配。", selector) {
		t.Fatal("valid evidence explanation was rejected")
	}
	if explanationSatisfies(strings.Repeat("无关内容", 20), selector) {
		t.Fatal("explanation without required evidence terms passed")
	}
}

func TestGeneratorReplaysHistoricalGuidedExplanationRule(t *testing.T) {
	contentDir, err := filepath.Abs("../../../content/learning")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := definition.LoadRegistry(definition.RegistryOptions{ContentDir: contentDir})
	if err != nil {
		t.Fatal(err)
	}
	generator, err := NewGenerator(registry)
	if err != nil {
		t.Fatal(err)
	}
	releaseID := "m1-first-slice-v6"
	task, err := registry.ExecutionTask(releaseID, "guided-run-model-v2", 5)
	if err != nil {
		t.Fatal(err)
	}
	workspace, err := registry.PublicWorkspace(releaseID, task.ID, task.Version)
	if err != nil {
		t.Fatal(err)
	}
	frozen := submission.Submission{
		ID: "submission-guided", AttemptID: "attempt-guided", ReleaseID: releaseID,
		TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
		Workspace: workspace, Explanation: strings.Repeat("学", 20),
	}
	terminal := terminalExecution(frozen, execution.ExecutionSucceeded, []execution.StageResult{passedStage(execution.StageBuild)})

	results, err := generator.Generate(frozen, terminal)
	if err != nil {
		t.Fatal(err)
	}
	if resultByID(t, results, "toolchain-baseline-builds").Status != execution.RulePassed ||
		resultByID(t, results, "toolchain-reflection-recorded").Status != execution.RulePassed {
		t.Fatalf("guided results = %#v", results)
	}
	frozen.Explanation = strings.Repeat("学", 19)
	results, err = generator.Generate(frozen, terminal)
	if err != nil {
		t.Fatal(err)
	}
	if resultByID(t, results, "toolchain-reflection-recorded").Status != execution.RuleFailed {
		t.Fatalf("short explanation results = %#v", results)
	}
}

func TestGeneratorDoesNotScoreCurrentGuidedExplanation(t *testing.T) {
	contentDir, err := filepath.Abs("../../../content/learning")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := definition.LoadRegistry(definition.RegistryOptions{ContentDir: contentDir})
	if err != nil {
		t.Fatal(err)
	}
	generator, err := NewGenerator(registry)
	if err != nil {
		t.Fatal(err)
	}
	releaseID := registry.CurrentReleaseID()
	task, err := registry.ExecutionTask(releaseID, "guided-run-model-v2", 6)
	if err != nil {
		t.Fatal(err)
	}
	workspace, err := registry.PublicWorkspace(releaseID, task.ID, task.Version)
	if err != nil {
		t.Fatal(err)
	}
	frozen := submission.Submission{
		ID: "submission-guided-current", AttemptID: "attempt-guided-current", ReleaseID: releaseID,
		TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
		Workspace: workspace, Explanation: strings.Repeat("学", 20),
	}
	terminal := terminalExecution(frozen, execution.ExecutionSucceeded, []execution.StageResult{passedStage(execution.StageBuild)})

	results, err := generator.Generate(frozen, terminal)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].RuleID != "toolchain-baseline-builds" {
		t.Fatalf("current guided results = %#v", results)
	}
}

func TestGeneratorConsumesRealAssessmentSandboxResult(t *testing.T) {
	contentDir, err := filepath.Abs("../../../content/learning")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := definition.LoadRegistry(definition.RegistryOptions{ContentDir: contentDir})
	if err != nil {
		t.Fatal(err)
	}
	releaseID := registry.CurrentReleaseID()
	activity, err := registry.ActivityView(releaseID, "assessment-check-config", 4)
	if err != nil {
		t.Fatal(err)
	}
	task, err := registry.ExecutionTask(releaseID, activity.TaskRef.ID, activity.TaskRef.Version)
	if err != nil {
		t.Fatal(err)
	}
	workspace, err := registry.PublicWorkspace(releaseID, task.ID, task.Version)
	if err != nil {
		t.Fatal(err)
	}
	workspace["internal/config/config.go"] = assessmentConfigSolution()
	workspace["internal/config/config_test.go"] = assessmentTableTests()
	current := attempt.Attempt{
		ID: "00000000-0000-4000-8400-000000000001", ReleaseID: releaseID,
		ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash,
		TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
		Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace),
	}
	builder, _ := execution.NewSpecBuilder(registry)
	executionID := "00000000-0000-4000-8400-000000000002"
	spec, err := builder.Build(current, executionID, execution.ActionSubmit)
	if err != nil {
		t.Fatal(err)
	}
	response, err := sandbox.NewRunner(sandbox.RunnerOptions{TempDir: t.TempDir()}).Run(context.Background(), spec)
	if err != nil {
		t.Fatal(err)
	}
	if response.Status != execution.ExecutionSucceeded || len(response.Stages) != 4 {
		t.Fatalf("real Sandbox response = %#v", response)
	}
	frozen := submission.Submission{
		ID: "00000000-0000-4000-8400-000000000003", AttemptID: current.ID,
		ReleaseID: releaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
		Workspace: workspace,
	}
	terminal := execution.Execution{
		ID: executionID, AttemptID: current.ID, SubmissionID: frozen.ID,
		TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
		Status: response.Status, Response: &response,
	}
	generator, _ := NewGenerator(registry)
	results, err := generator.Generate(frozen, terminal)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		if result.Status != execution.RulePassed {
			t.Fatalf("real rule result %s = %#v", result.RuleID, result)
		}
	}
}

func TestConcurrencyAssessmentSolutionsPassRealSandbox(t *testing.T) {
	contentDir, err := filepath.Abs("../../../content/learning")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := definition.LoadRegistry(definition.RegistryOptions{ContentDir: contentDir})
	if err != nil {
		t.Fatal(err)
	}
	for _, testCase := range []struct {
		name       string
		activityID string
		path       string
		source     string
	}{
		{name: "bounded worker pool", activityID: "assessment-worker-pool", path: "pool.go", source: workerPoolSolution()},
		{name: "cancellable checks", activityID: "assessment-cancellable-checks", path: "checker.go", source: cancellableChecksSolution()},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			activity, err := registry.ActivityView(registry.CurrentReleaseID(), testCase.activityID, 1)
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
			workspace[testCase.path] = testCase.source
			current := attempt.Attempt{
				ID: "concurrency-attempt-" + strings.ReplaceAll(testCase.name, " ", "-"), ReleaseID: registry.CurrentReleaseID(),
				ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash,
				TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
				Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace),
			}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			spec, err := builder.Build(current, "concurrency-execution-"+strings.ReplaceAll(testCase.name, " ", "-"), execution.ActionSubmit)
			if err != nil {
				t.Fatal(err)
			}
			response, err := sandbox.NewRunner(sandbox.RunnerOptions{TempDir: t.TempDir()}).Run(context.Background(), spec)
			if err != nil {
				t.Fatal(err)
			}
			if response.Status != execution.ExecutionSucceeded {
				t.Fatalf("Sandbox response = %#v", response)
			}
		})
	}
}

func TestAbstractionAndDebugAssessmentSolutionsPassRealEvaluation(t *testing.T) {
	contentDir, err := filepath.Abs("../../../content/learning")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := definition.LoadRegistry(definition.RegistryOptions{ContentDir: contentDir})
	if err != nil {
		t.Fatal(err)
	}
	generator, err := NewGenerator(registry)
	if err != nil {
		t.Fatal(err)
	}
	testCases := []struct {
		name        string
		activityID  string
		path        string
		source      string
		explanation string
	}{
		{name: "minimal interface and generics", activityID: "assessment-delivery-service", path: "delivery.go", source: deliveryServiceSolution()},
		{
			name: "debug evidence", activityID: "assessment-report-debug", path: "report.go", source: reportDebugSolution(),
			explanation: "失败测试确认最后一项丢失，breakpoint 显示循环提前退出；alloc_space 指向重复拼接，所以改用 strings.Builder，最后用隐藏测试和 Vet 验证行为与格式参数。",
		},
	}
	for index, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			activity, err := registry.ActivityView(registry.CurrentReleaseID(), testCase.activityID, 1)
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
			workspace[testCase.path] = testCase.source
			current := attempt.Attempt{
				ID: "00000000-0000-4000-8500-00000000000" + strconv.Itoa(index+1), ReleaseID: registry.CurrentReleaseID(),
				ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash,
				TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
				Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace),
			}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4000-8600-00000000000" + strconv.Itoa(index+1)
			spec, err := builder.Build(current, executionID, execution.ActionSubmit)
			if err != nil {
				t.Fatal(err)
			}
			response, err := sandbox.NewRunner(sandbox.RunnerOptions{TempDir: t.TempDir()}).Run(context.Background(), spec)
			if err != nil {
				t.Fatal(err)
			}
			if response.Status != execution.ExecutionSucceeded {
				t.Fatalf("Sandbox response = %#v", response)
			}
			frozen := submission.Submission{
				ID: "00000000-0000-4000-8700-00000000000" + strconv.Itoa(index+1), AttemptID: current.ID,
				ReleaseID: current.ReleaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
				Workspace: workspace, Explanation: testCase.explanation,
			}
			terminal := execution.Execution{
				ID: executionID, AttemptID: current.ID, SubmissionID: frozen.ID,
				TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
				Status: response.Status, Response: &response,
			}
			results, err := generator.Generate(frozen, terminal)
			if err != nil {
				t.Fatal(err)
			}
			for _, result := range results {
				if result.Status != execution.RulePassed {
					t.Fatalf("rule %s = %#v", result.RuleID, result)
				}
			}
		})
	}
}

func workerPoolSolution() string {
	return `package workerpool

import "sync"

type Result struct {
	Index int
	Value int
}

type job struct {
	index int
	value int
}

func Process(values []int, workers int, transform func(int) int) <-chan Result {
	results := make(chan Result)
	if workers <= 0 || len(values) == 0 {
		close(results)
		return results
	}
	jobs := make(chan job)
	go func() {
		defer close(jobs)
		for index, value := range values {
			jobs <- job{index: index, value: value}
		}
	}()
	var group sync.WaitGroup
	group.Add(workers)
	for range workers {
		go func() {
			defer group.Done()
			for item := range jobs {
				results <- Result{Index: item.index, Value: transform(item.value)}
			}
		}()
	}
	go func() {
		group.Wait()
		close(results)
	}()
	return results
}
`
}

func cancellableChecksSolution() string {
	return `package checker

import (
	"context"
	"fmt"
	"sync"
)

func CheckAll(ctx context.Context, targets []string, workers int, check func(context.Context, string) error) error {
	if workers <= 0 {
		return fmt.Errorf("workers must be positive")
	}
	if len(targets) == 0 {
		return nil
	}
	batch, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan string)
	results := make(chan error, workers)
	var group sync.WaitGroup
	group.Add(workers + 1)
	go func() {
		defer group.Done()
		defer close(jobs)
		for _, target := range targets {
			select {
			case jobs <- target:
			case <-batch.Done():
				return
			}
		}
	}()
	for range workers {
		go func() {
			defer group.Done()
			for {
				select {
				case <-batch.Done():
					return
				case target, ok := <-jobs:
					if !ok {
						return
					}
					err := check(batch, target)
					select {
					case results <- err:
					case <-batch.Done():
						return
					}
				}
			}
		}()
	}
	go func() {
		group.Wait()
		close(results)
	}()
	var first error
	for err := range results {
		if err != nil && first == nil {
			if ctx.Err() != nil {
				first = ctx.Err()
			} else {
				first = err
			}
			cancel()
		}
	}
	if first != nil {
		return first
	}
	return ctx.Err()
}
`
}

func deliveryServiceSolution() string {
	return `package delivery

type Message struct {
	Recipient string
	Body      string
}

type Sender interface {
	Send(Message) error
}

type Service struct {
	sender Sender
}

func New(sender Sender) Service {
	return Service{sender: sender}
}

func (s Service) Deliver(messages []Message) error {
	for _, message := range messages {
		if err := s.sender.Send(message); err != nil {
			return err
		}
	}
	return nil
}

func IndexBy[T any, K comparable](values []T, key func(T) K) map[K]T {
	indexed := make(map[K]T, len(values))
	for _, value := range values {
		indexed[key(value)] = value
	}
	return indexed
}
`
}

func reportDebugSolution() string {
	return `package report

import (
	"fmt"
	"io"
	"strings"
)

type Entry struct {
	Name  string
	Value int
}

func Render(entries []Entry) string {
	var output strings.Builder
	for _, entry := range entries {
		fmt.Fprintf(&output, "%s=%d\n", entry.Name, entry.Value)
	}
	return output.String()
}

func LogSummary(writer io.Writer, rendered string) {
	fmt.Fprintf(writer, "rendered=%s\n", rendered)
}
`
}

func setupGenerator(t *testing.T) (*Generator, submission.Submission) {
	t.Helper()
	contentDir, err := filepath.Abs("../../../content/learning")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := definition.LoadRegistry(definition.RegistryOptions{ContentDir: contentDir})
	if err != nil {
		t.Fatal(err)
	}
	generator, err := NewGenerator(registry)
	if err != nil {
		t.Fatal(err)
	}
	releaseID := registry.CurrentReleaseID()
	task, err := registry.ExecutionTask(releaseID, "assessment-check-config-v2", 3)
	if err != nil {
		t.Fatal(err)
	}
	workspace, err := registry.PublicWorkspace(releaseID, task.ID, task.Version)
	if err != nil {
		t.Fatal(err)
	}
	workspace["internal/config/config.go"] = "package config\n\nimport \"os\"\n\nfunc inspect(path string) error {\n\tfile, err := os.Open(path)\n\tif err != nil { return err }\n\tdefer file.Close()\n\treturn nil\n}\n"
	workspace["internal/config/config_test.go"] = "package config\n\nimport \"testing\"\n\nfunc TestLearnerCases(t *testing.T) {\n\ttests := []struct{name string}{\n\t\t{name: \"valid\"},\n\t\t{name: \"invalid URL\"},\n\t\t{name: \"missing file\"},\n\t}\n\tfor _, test := range tests {\n\t\tt.Run(test.name, func(t *testing.T) {})\n\t}\n}\n"
	return generator, submission.Submission{
		ID:        "00000000-0000-4000-8300-000000000001",
		AttemptID: "00000000-0000-4000-8300-000000000002",
		ReleaseID: releaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
		Workspace: workspace,
	}
}

func terminalExecution(frozen submission.Submission, status execution.ExecutionStatus, stages []execution.StageResult) execution.Execution {
	response := &execution.ExecutionResponse{
		ProtocolVersion: execution.ProtocolVersion,
		ExecutionID:     "00000000-0000-4000-8300-000000000003",
		Status:          status,
		Stages:          stages,
		Policy: execution.PolicyReport{Network: execution.NetworkPolicyReport{
			Requested: execution.NetworkNone, Enforcement: execution.EnforcementPolicyOnly,
		}},
	}
	return execution.Execution{
		ID: response.ExecutionID, AttemptID: frozen.AttemptID, SubmissionID: frozen.ID,
		TaskID: frozen.TaskID, TaskVersion: frozen.TaskVersion, TaskHash: frozen.TaskHash,
		Status: status, Response: response,
	}
}

func passedStage(stage execution.Stage) execution.StageResult {
	return execution.StageResult{Stage: stage, Status: execution.StagePassed}
}

func assertRuleStatuses(t *testing.T, results []execution.RuleResult, want map[string]execution.RuleStatus) {
	t.Helper()
	if len(results) != len(want) {
		t.Fatalf("len(results) = %d, want %d", len(results), len(want))
	}
	for _, result := range results {
		status, exists := want[result.RuleID]
		if !exists || result.Status != status {
			t.Fatalf("result %s = %q, want %q", result.RuleID, result.Status, status)
		}
	}
}

func resultByID(t *testing.T, results []execution.RuleResult, ruleID string) execution.RuleResult {
	t.Helper()
	for _, result := range results {
		if result.RuleID == ruleID {
			return result
		}
	}
	t.Fatalf("missing result %q", ruleID)
	return execution.RuleResult{}
}

func assessmentConfigSolution() string {
	return "package config\n\nimport (\n\t\"encoding/json\"\n\t\"fmt\"\n\t\"net/url\"\n\t\"os\"\n\t\"sort\"\n\t\"strings\"\n)\n\n" +
		"type Config struct { Targets []Target }\n" +
		"type Target struct { Name string; URL string; TimeoutMS int }\n\n" +
		"func Load(path string) (Config, error) {\n" +
		"\tfile, err := os.Open(path)\n\tif err != nil { return Config{}, fmt.Errorf(\"open config: %w\", err) }\n" +
		"\tdefer file.Close()\n\tvar raw struct { Targets []struct { Name string; URL string; Timeout_ms int } }\n" +
		"\tif err := json.NewDecoder(file).Decode(&raw); err != nil { return Config{}, fmt.Errorf(\"decode config: %w\", err) }\n" +
		"\tcfg := Config{Targets: make([]Target, len(raw.Targets))}\n\tfor index, target := range raw.Targets { cfg.Targets[index] = Target{Name: target.Name, URL: target.URL, TimeoutMS: target.Timeout_ms} }\n" +
		"\treturn Normalize(cfg)\n}\n\n" +
		"func Normalize(cfg Config) (Config, error) {\n" +
		"\tif len(cfg.Targets) == 0 { return Config{}, fmt.Errorf(\"at least one target is required\") }\n" +
		"\tseen := map[string]bool{}\n\tfor index := range cfg.Targets {\n\t\ttarget := &cfg.Targets[index]\n\t\ttarget.Name = strings.TrimSpace(target.Name)\n" +
		"\t\tparsed, err := url.Parse(target.URL)\n\t\tif target.Name == \"\" || err != nil || (parsed.Scheme != \"http\" && parsed.Scheme != \"https\") || target.TimeoutMS <= 0 || seen[target.Name] { return Config{}, fmt.Errorf(\"invalid target %q\", target.Name) }\n" +
		"\t\tseen[target.Name] = true\n\t}\n\tsort.Slice(cfg.Targets, func(i, j int) bool { return cfg.Targets[i].Name < cfg.Targets[j].Name })\n\treturn cfg, nil\n}\n"
}

func assessmentTableTests() string {
	return "package config\n\nimport \"testing\"\n\nfunc TestLearnerValidation(t *testing.T) {\n" +
		"\ttests := []struct { name string; cfg Config; wantError bool }{\n" +
		"\t\t{name: \"valid\", cfg: Config{Targets: []Target{{Name: \"api\", URL: \"https://api.example\", TimeoutMS: 1}}}},\n" +
		"\t\t{name: \"empty\", cfg: Config{}, wantError: true},\n" +
		"\t\t{name: \"bad scheme\", cfg: Config{Targets: []Target{{Name: \"db\", URL: \"file:///tmp/db\", TimeoutMS: 1}}}, wantError: true},\n\t}\n" +
		"\tfor _, test := range tests { t.Run(test.name, func(t *testing.T) { _, err := Normalize(test.cfg); if (err != nil) != test.wantError { t.Fatalf(\"error = %v\", err) } }) }\n}\n"
}
