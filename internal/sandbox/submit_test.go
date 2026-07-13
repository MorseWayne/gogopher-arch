package sandbox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
)

func TestSubmitRunsVisibleThenHeldOutTestsFromBinaries(t *testing.T) {
	runner := NewRunner(RunnerOptions{TempDir: t.TempDir()})
	spec := submitSpec([]execution.FileAsset{
		runnerAsset("go.mod", "module task\n\ngo 1.25.0\n", execution.OriginReleaseBundle, execution.AccessReadonly, execution.RoleWorkspace),
		runnerAsset("math.go", "package task\n\nfunc Add(a, b int) int { return a + b }\n", execution.OriginLearnerWorkspace, execution.AccessEditable, execution.RoleWorkspace),
		runnerAsset("math_test.go", "package task\n\nimport \"testing\"\n\nfunc TestVisible(t *testing.T) { if Add(2, 3) != 5 { t.Fatal(\"bad sum\") } }\n", execution.OriginReleaseBundle, execution.AccessReadonly, execution.RoleVisibleTest),
		runnerAsset("hidden_test.go", hiddenEnumerationTest(), execution.OriginReleaseBundle, execution.AccessReadonly, execution.RoleHeldOutTest),
	})
	response, err := runner.Run(context.Background(), spec)
	if err != nil {
		t.Fatal(err)
	}
	if response.Status != execution.ExecutionSucceeded || len(response.Stages) != 4 {
		t.Fatalf("response = %#v", response)
	}
	if response.Stages[0].Stage != execution.StageBuild || response.Stages[1].Stage != execution.StageVet ||
		response.Stages[2].Stage != execution.StageVisibleTest || response.Stages[3].Stage != execution.StageHeldOutTest {
		t.Fatalf("stages = %#v", response.Stages)
	}
	if !hasTestEvent(response.Stages[2].TestEvents, "TestVisible", "pass") || !hasTestEvent(response.Stages[3].TestEvents, "TestHeldOutSourceIsAbsent", "pass") {
		t.Fatalf("test events = %#v", response.Stages)
	}
	if response.Stages[3].Stdout != "" || response.Stages[3].Stderr != "" || strings.Contains(response.Stages[3].PublicSummary, "hidden_test.go") {
		t.Fatalf("held-out stage leaks details: %#v", response.Stages[3])
	}
}

func TestSubmitDoesNotInjectHeldOutTestsAfterVisibleFailure(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "held-out-ran")
	runner := NewRunner(RunnerOptions{TempDir: t.TempDir()})
	hidden := fmt.Sprintf("package task\n\nimport (\"os\"; \"testing\")\n\nfunc TestHidden(t *testing.T) { if err := os.WriteFile(%q, []byte(\"ran\"), 0600); err != nil { t.Fatal(err) } }\n", marker)
	spec := submitSpec([]execution.FileAsset{
		runnerAsset("go.mod", "module task\n\ngo 1.25.0\n", execution.OriginReleaseBundle, execution.AccessReadonly, execution.RoleWorkspace),
		runnerAsset("math.go", "package task\n", execution.OriginLearnerWorkspace, execution.AccessEditable, execution.RoleWorkspace),
		runnerAsset("visible_test.go", "package task\n\nimport \"testing\"\n\nfunc TestVisible(t *testing.T) { t.Fatal(\"stop\") }\n", execution.OriginReleaseBundle, execution.AccessReadonly, execution.RoleVisibleTest),
		runnerAsset("hidden_test.go", hidden, execution.OriginReleaseBundle, execution.AccessReadonly, execution.RoleHeldOutTest),
	})
	response, err := runner.Run(context.Background(), spec)
	if err != nil {
		t.Fatal(err)
	}
	if response.Status != execution.ExecutionUserFailed || len(response.Stages) != 3 || response.Stages[2].Stage != execution.StageVisibleTest {
		t.Fatalf("response = %#v", response)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("held-out test ran before visible stage passed: %v", err)
	}
}

func TestParseGoTestJSONProducesStableEvents(t *testing.T) {
	raw := strings.Join([]string{
		`{"Time":"2026-07-13T00:00:00Z","Action":"run","Package":"task/pkg","Test":"TestOne"}`,
		`{"Time":"2026-07-13T00:00:01Z","Action":"output","Package":"task/pkg","Test":"TestOne","Output":"hello\n"}`,
		`{"Time":"2026-07-13T00:00:02Z","Action":"pass","Package":"task/pkg","Test":"TestOne","Elapsed":0.01}`,
	}, "\n")
	events, output, err := parseGoTestJSON([]byte(raw), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 || events[0].Action != "run" || events[1].Action != "pass" || output != "hello\n" {
		t.Fatalf("events = %#v, output = %q", events, output)
	}
	for _, event := range events {
		encoded := fmt.Sprintf("%#v", event)
		if strings.Contains(encoded, "2026-07-13") {
			t.Fatalf("event contains unstable timestamp: %s", encoded)
		}
	}
}

func submitSpec(files []execution.FileAsset) execution.ExecutionSpec {
	spec := runnerSpec(execution.ActionSubmit, files)
	spec.Policy.TimeoutMS = 15_000
	return spec
}

func hasTestEvent(events []execution.TestEvent, test, action string) bool {
	for _, event := range events {
		if event.Test == test && event.Action == action {
			return true
		}
	}
	return false
}

func hiddenEnumerationTest() string {
	return `package task

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHeldOutSourceIsAbsent(t *testing.T) {
	err := filepath.Walk("..", func(path string, info os.FileInfo, err error) error {
		if err != nil { return err }
		if info.Name() == "hidden_test.go" { t.Fatalf("held-out source is visible at %s", path) }
		return nil
	})
	if err != nil { t.Fatal(err) }
}
`
}
