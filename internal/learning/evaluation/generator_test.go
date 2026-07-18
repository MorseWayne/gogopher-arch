package evaluation

import (
	"context"
	"path/filepath"
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
