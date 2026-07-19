package evaluation

import (
	"strconv"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

func TestM214LearningLoopSolutionsPassRealReleaseAndSandboxEvaluation(t *testing.T) {
	tests := []struct {
		name, activity string
		files          map[string]string
		explanation    string
	}{
		{name: "release gate practice", activity: "practice-release-gates", files: map[string]string{"releaseplan/plan.go": releasePlanSolution()}},
		{name: "gocheck delivery", activity: "assessment-gocheck-delivery", files: map[string]string{
			"internal/delivery/migrations.go":      migrationPolicySolution("delivery", "ValidateMigrations"),
			"internal/delivery/migrations_test.go": migrationPolicyTests("delivery", "ValidateMigrations"),
			"Dockerfile":                           deliveryDockerfile("gocheckhub", "app", "./cmd/gocheckhub"),
			"Makefile":                             deliveryMakefile("./internal/delivery", "gocheckhub"),
			".github/workflows/ci.yml":             deliveryWorkflow("gocheckhub"),
		}, explanation: deliveryExplanation("gocheck-hub")},
		{name: "alert worker delivery variant", activity: "review-gocheck-alert-delivery-pipeline", files: map[string]string{
			"internal/alertdelivery/migrations.go":      alertMigrationPolicySolution(),
			"internal/alertdelivery/migrations_test.go": alertMigrationPolicyTests(),
			"Dockerfile":               deliveryDockerfile("alertworker", "worker", "./cmd/alertworker"),
			"Makefile":                 deliveryMakefile("./internal/alertdelivery", "alertworker"),
			".github/workflows/ci.yml": deliveryWorkflow("alertworker"),
		}, explanation: deliveryExplanation("alert worker")},
	}

	registry := draftReleaseRegistry(t)
	for index, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			activity, err := registry.ActivityView(registry.CurrentReleaseID(), testCase.activity, 1)
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
			for path, source := range testCase.files {
				workspace[path] = source
			}
			current := attempt.Attempt{
				ID: "00000000-0000-4000-9150-00000000000" + strconv.Itoa(index+1), ReleaseID: registry.CurrentReleaseID(),
				ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash,
				TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
				Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace),
			}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4001-9150-00000000000" + strconv.Itoa(index+1)
			spec, err := builder.Build(current, executionID, execution.ActionSubmit)
			if err != nil {
				t.Fatal(err)
			}
			response, err := runRegressionSandbox(t, spec)
			if err != nil {
				t.Fatal(err)
			}
			if response.Status != execution.ExecutionSucceeded {
				t.Fatalf("sandbox response = %#v", response)
			}
			frozen := submission.Submission{
				ID: "00000000-0000-4001-9160-00000000000" + strconv.Itoa(index+1), AttemptID: current.ID,
				ReleaseID: current.ReleaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
				Workspace: workspace, Explanation: testCase.explanation,
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
					t.Fatalf("rule %s = %#v", result.RuleID, result)
				}
			}
		})
	}
}

func releasePlanSolution() string {
	return `package releaseplan
import("errors";"strings")
type Config struct{RuntimeUser string;Checks []string;MigrationMode string}
func Validate(config Config)error{if strings.TrimSpace(config.RuntimeUser)==""||config.RuntimeUser=="root"{return errors.New("unsafe runtime user")};if config.MigrationMode!="forward-only"{return errors.New("migration mode must be forward-only")};required:=map[string]bool{"fmt":false,"vet":false,"test":false,"race":false,"vuln":false,"migration":false,"image":false};if len(config.Checks)!=len(required){return errors.New("incomplete gate set")};for _,check:=range config.Checks{seen,ok:=required[check];if !ok||seen{return errors.New("unknown or duplicate gate")};required[check]=true};for _,seen:=range required{if !seen{return errors.New("missing gate")}};return nil}
`
}

func migrationPolicySolution(packageName, functionName string) string {
	return `package ` + packageName + `
import("errors";"regexp";"strconv")
var migrationPattern=regexp.MustCompile(` + "`" + `^([0-9]{4})_[a-z0-9]+(?:_[a-z0-9]+)*\.up\.sql$` + "`" + `)
func ` + functionName + `(names []string)error{if len(names)==0{return errors.New("migrations required")};last:=-1;seen:=map[int]struct{}{};for _,name:=range names{match:=migrationPattern.FindStringSubmatch(name);if match==nil{return errors.New("invalid forward-only migration")};sequence,err:=strconv.Atoi(match[1]);if err!=nil{return err};if _,exists:=seen[sequence];exists||sequence<=last{return errors.New("migration order drift")};seen[sequence]=struct{}{};last=sequence};return nil}
`
}

func migrationPolicyTests(packageName, functionName string) string {
	return `package ` + packageName + `
import "testing"
func Test` + functionName + `(t *testing.T){tests:=[]struct{name string;files []string;wantErr bool}{{"valid",[]string{"0001_create_checks.up.sql","0002_add_status.up.sql"},false},{"empty",nil,true},{"down",[]string{"0001_create.down.sql"},true},{"bad name",[]string{"create.up.sql"},true},{"duplicate",[]string{"0001_a.up.sql","0001_b.up.sql"},true},{"reordered",[]string{"0002_b.up.sql","0001_a.up.sql"},true}};for _,tc:=range tests{t.Run(tc.name,func(t *testing.T){err:=` + functionName + `(tc.files);if (err!=nil)!=tc.wantErr{t.Fatalf("error=%v",err)}})}}
`
}

func alertMigrationPolicySolution() string {
	return `package alertdelivery
import("encoding/hex";"errors";"regexp";"strconv")
type Migration struct{Name,SHA256 string}
var migrationPattern=regexp.MustCompile(` + "`" + `^([0-9]{4})_[a-z0-9]+(?:_[a-z0-9]+)*\.up\.sql$` + "`" + `)
func ValidateManifest(entries []Migration)error{if len(entries)==0{return errors.New("manifest required")};last:=-1;seen:=map[int]struct{}{};for _,entry:=range entries{match:=migrationPattern.FindStringSubmatch(entry.Name);if match==nil{return errors.New("invalid forward-only migration")};sequence,err:=strconv.Atoi(match[1]);if err!=nil{return err};if _,exists:=seen[sequence];exists||sequence<=last{return errors.New("migration order drift")};checksum,err:=hex.DecodeString(entry.SHA256);if err!=nil||len(checksum)!=32{return errors.New("invalid SHA-256")};seen[sequence]=struct{}{};last=sequence};return nil}
`
}

func alertMigrationPolicyTests() string {
	return `package alertdelivery
import("strings";"testing")
func TestValidateManifest(t *testing.T){sha:=strings.Repeat("a",64);tests:=[]struct{name string;entries []Migration;wantErr bool}{{"valid",[]Migration{{"0001_create_alerts.up.sql",sha},{"0002_add_attempts.up.sql",sha}},false},{"empty",nil,true},{"down",[]Migration{{"0001_alerts.down.sql",sha}},true},{"checksum",[]Migration{{"0001_alerts.up.sql","bad"}},true},{"duplicate",[]Migration{{"0001_a.up.sql",sha},{"0001_b.up.sql",sha}},true},{"reordered",[]Migration{{"0002_b.up.sql",sha},{"0001_a.up.sql",sha}},true}};for _,tc:=range tests{t.Run(tc.name,func(t *testing.T){err:=ValidateManifest(tc.entries);if (err!=nil)!=tc.wantErr{t.Fatalf("error=%v",err)}})}}
`
}

func deliveryDockerfile(binary, user, command string) string {
	return `FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -o /out/` + binary + ` ` + command + `

FROM alpine:3.22
RUN addgroup -S ` + user + ` && adduser -S -G ` + user + ` ` + user + `
COPY --from=builder /out/` + binary + ` /usr/local/bin/` + binary + `
USER ` + user + `
ENTRYPOINT ["/usr/local/bin/` + binary + `"]
`
}

func deliveryMakefile(migrationPackage, image string) string {
	return `.PHONY: fmt-check vet test race vuln migration-check image verify
fmt-check:
	test -z "$$(gofmt -l .)"
vet:
	go vet ./...
test:
	go test ./...
race:
	go test -race ./...
vuln:
	govulncheck ./...
migration-check:
	go test ` + migrationPackage + ` -run TestValidate
image:
	docker build -t ` + image + `:test .
verify: fmt-check vet test race vuln migration-check image
`
}

func deliveryWorkflow(image string) string {
	return `name: delivery
on: [push, pull_request]
permissions:
  contents: read
jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.25'
      - run: test -z "$(gofmt -l .)"
      - run: go vet ./...
      - run: go test ./...
      - run: go test -race ./...
      - run: go install golang.org/x/vuln/cmd/govulncheck@v1.1.4
      - run: govulncheck ./...
      - run: make migration-check
      - run: docker build -t ` + image + `:ci .
`
}

func deliveryExplanation(service string) string {
	return service + " 使用 multi-stage Docker build，把编译器与源码留在 builder，只把静态二进制复制到最小 runtime，并以 non-root 用户运行。CI 与 Makefile 同时执行格式、vet、普通测试、go test -race 和 govulncheck；依赖漏洞不能由普通单测替代。forward-only migration 检查拒绝 down、重复序号、乱序和命名漂移，镜像构建必须排在源码、依赖与数据升级门禁之后，失败时不会产生可发布镜像。"
}
