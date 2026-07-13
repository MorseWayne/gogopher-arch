package sandbox

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
)

func TestRunnerExecutesBuildTestAndVetInFreshModules(t *testing.T) {
	for _, action := range []execution.Action{execution.ActionBuild, execution.ActionTest, execution.ActionVet} {
		t.Run(string(action), func(t *testing.T) {
			tempRoot := t.TempDir()
			runner := NewRunner(RunnerOptions{TempDir: tempRoot})
			spec := runnerSpec(action, []execution.FileAsset{
				runnerAsset("go.mod", "module task\n\ngo 1.25.0\n", execution.OriginReleaseBundle, execution.AccessReadonly, execution.RoleWorkspace),
				runnerAsset("math.go", "package task\n\nfunc Add(a, b int) int { return a + b }\n", execution.OriginLearnerWorkspace, execution.AccessEditable, execution.RoleWorkspace),
				runnerAsset("math_test.go", "package task\n\nimport \"testing\"\n\nfunc TestAdd(t *testing.T) { if Add(2, 3) != 5 { t.Fatal(\"bad sum\") } }\n", execution.OriginReleaseBundle, execution.AccessReadonly, execution.RoleVisibleTest),
			})
			response, err := runner.Run(context.Background(), spec)
			if err != nil {
				t.Fatal(err)
			}
			if response.Status != execution.ExecutionSucceeded || len(response.Stages) != 1 || response.Stages[0].Status != execution.StagePassed {
				t.Fatalf("response = %#v", response)
			}
			if response.Policy.Network.Enforcement != execution.EnforcementPolicyOnly {
				t.Fatalf("network enforcement = %q", response.Policy.Network.Enforcement)
			}
			entries, err := os.ReadDir(tempRoot)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 0 {
				t.Fatalf("execution workspace was not removed: %v", entries)
			}
		})
	}
}

func TestRunnerClassifiesUserFailureTimeoutAndOutputTruncation(t *testing.T) {
	t.Run("compile failure", func(t *testing.T) {
		runner := NewRunner(RunnerOptions{TempDir: t.TempDir()})
		spec := runnerSpec(execution.ActionBuild, []execution.FileAsset{
			runnerAsset("go.mod", "module task\n\ngo 1.25.0\n", execution.OriginReleaseBundle, execution.AccessReadonly, execution.RoleWorkspace),
			runnerAsset("broken.go", "package task\n\nfunc Broken( {\n", execution.OriginLearnerWorkspace, execution.AccessEditable, execution.RoleWorkspace),
		})
		response, err := runner.Run(context.Background(), spec)
		if err != nil {
			t.Fatal(err)
		}
		if response.Status != execution.ExecutionUserFailed || response.Stages[0].ExitCode == 0 {
			t.Fatalf("response = %#v", response)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		runner := NewRunner(RunnerOptions{TempDir: t.TempDir()})
		spec := runnerSpec(execution.ActionTest, []execution.FileAsset{
			runnerAsset("go.mod", "module task\n\ngo 1.25.0\n", execution.OriginReleaseBundle, execution.AccessReadonly, execution.RoleWorkspace),
			runnerAsset("slow_test.go", "package task\n\nimport (\"testing\"; \"time\")\n\nfunc TestSlow(t *testing.T) { time.Sleep(5 * time.Second) }\n", execution.OriginReleaseBundle, execution.AccessReadonly, execution.RoleVisibleTest),
		})
		spec.Policy.TimeoutMS = execution.MinTimeoutMS
		response, err := runner.Run(context.Background(), spec)
		if err != nil {
			t.Fatal(err)
		}
		if response.Status != execution.ExecutionUserFailed || !response.Stages[0].TimedOut {
			t.Fatalf("response = %#v", response)
		}
	})

	t.Run("output truncation", func(t *testing.T) {
		runner := NewRunner(RunnerOptions{TempDir: t.TempDir()})
		spec := runnerSpec(execution.ActionTest, []execution.FileAsset{
			runnerAsset("go.mod", "module task\n\ngo 1.25.0\n", execution.OriginReleaseBundle, execution.AccessReadonly, execution.RoleWorkspace),
			runnerAsset("spam_test.go", "package task\n\nimport (\"fmt\"; \"testing\")\n\nfunc TestSpam(t *testing.T) { for i := 0; i < 5000; i++ { fmt.Println(\"lots of output\") }; t.Fail() }\n", execution.OriginReleaseBundle, execution.AccessReadonly, execution.RoleVisibleTest),
		})
		spec.Policy.MaxOutputBytes = 1_024
		response, err := runner.Run(context.Background(), spec)
		if err != nil {
			t.Fatal(err)
		}
		stage := response.Stages[0]
		if !stage.OutputTruncated || len(stage.Stdout)+len(stage.Stderr) > spec.Policy.MaxOutputBytes {
			t.Fatalf("stage = %#v", stage)
		}
	})
}

func TestRunnerRejectsInvalidSpecBeforeCreatingWorkspace(t *testing.T) {
	tempRoot := t.TempDir()
	runner := NewRunner(RunnerOptions{TempDir: tempRoot, GoBinary: filepath.Join(tempRoot, "must-not-run")})
	spec := runnerSpec(execution.ActionBuild, []execution.FileAsset{
		runnerAsset("../main.go", "package task\n", execution.OriginLearnerWorkspace, execution.AccessEditable, execution.RoleWorkspace),
	})
	if _, err := runner.Run(context.Background(), spec); err == nil || !strings.Contains(err.Error(), "files[0].path") {
		t.Fatalf("Run() error = %v", err)
	}
	entries, err := os.ReadDir(tempRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("invalid request created files: %v", entries)
	}
}

func TestMaterializerRejectsSymlinkParentsAndHashMismatch(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "linked")); err != nil {
		t.Fatal(err)
	}
	if err := makeParentsWithoutSymlinks(root, filepath.Join(root, "linked", "child")); err == nil || !strings.Contains(err.Error(), "not a real directory") {
		t.Fatalf("makeParentsWithoutSymlinks() error = %v", err)
	}

	asset := runnerAsset("main.go", "package task\n", execution.OriginLearnerWorkspace, execution.AccessEditable, execution.RoleWorkspace)
	asset.SHA256 = strings.Repeat("0", 64)
	if err := materializeAssets(filepath.Join(t.TempDir(), "workspace"), []execution.FileAsset{asset}, func(execution.FileAsset) bool { return true }); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("materializeAssets() error = %v", err)
	}
}

func TestActionMappingUsesFixedGoArguments(t *testing.T) {
	tests := map[execution.Action]string{
		execution.ActionBuild: "build ./...", execution.ActionTest: "test ./...", execution.ActionVet: "vet ./...",
	}
	for action, expected := range tests {
		_, arguments := actionCommand(action)
		if strings.Join(arguments, " ") != expected {
			t.Fatalf("actionCommand(%s) = %q", action, arguments)
		}
	}
}

func TestSandboxEnvironmentDoesNotInheritProcessSecrets(t *testing.T) {
	t.Setenv("DATABASE_URL", "secret-database-url")
	t.Setenv("SESSION_TOKEN", "secret-session-token")
	environment := strings.Join(sandboxEnvironment(t.TempDir()), "\n")
	if strings.Contains(environment, "DATABASE_URL") || strings.Contains(environment, "SESSION_TOKEN") || strings.Contains(environment, "secret-") {
		t.Fatalf("sandbox environment inherited a process secret: %s", environment)
	}
	for _, required := range []string{"PATH=", "GOPROXY=off", "GOWORK=off", "GOENV=off", "TMPDIR="} {
		if !strings.Contains(environment, required) {
			t.Fatalf("sandbox environment is missing %q: %s", required, environment)
		}
	}
}

func TestCallerCancellationIsInfrastructureFailure(t *testing.T) {
	runner := NewRunner(RunnerOptions{TempDir: t.TempDir()})
	spec := runnerSpec(execution.ActionTest, []execution.FileAsset{
		runnerAsset("go.mod", "module task\n\ngo 1.25.0\n", execution.OriginReleaseBundle, execution.AccessReadonly, execution.RoleWorkspace),
		runnerAsset("slow_test.go", "package task\n\nimport (\"testing\"; \"time\")\n\nfunc TestSlow(t *testing.T) { time.Sleep(5 * time.Second) }\n", execution.OriginReleaseBundle, execution.AccessReadonly, execution.RoleVisibleTest),
	})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	response, err := runner.Run(ctx, spec)
	if err != nil {
		t.Fatal(err)
	}
	if response.Status != execution.ExecutionInfraFailed || response.Failure == nil || response.Failure.Code != "execution_context_cancelled" {
		t.Fatalf("response = %#v", response)
	}
}

func runnerSpec(action execution.Action, files []execution.FileAsset) execution.ExecutionSpec {
	return execution.ExecutionSpec{
		ProtocolVersion: execution.ProtocolVersion, ExecutionID: fmt.Sprintf("%s-execution", action),
		Language: execution.GoLanguage, WorkspaceRoot: execution.WorkspaceRoot, Action: action, Files: files,
		Limits: execution.WorkspaceLimits{MaxFiles: 16, MaxFileBytes: 1 << 16, MaxTotalBytes: 1 << 18},
		Policy: execution.ActionPolicy{TimeoutMS: 5_000, MaxOutputBytes: 65_536, Network: execution.NetworkNone},
	}
}

func runnerAsset(path, content string, origin execution.AssetOrigin, access execution.AssetAccess, role execution.AssetRole) execution.FileAsset {
	return execution.FileAsset{Path: path, Content: content, SHA256: execution.SHA256Hex(content), Origin: origin, Access: access, Role: role}
}
