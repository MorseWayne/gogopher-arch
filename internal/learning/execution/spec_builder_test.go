package execution

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

func TestSpecBuilderUsesWorkspaceOnlyForEditableAssets(t *testing.T) {
	registry := releaseRegistry(t)
	current := frozenAttempt(t, registry, "assessment-check-config-v1")
	current.Workspace["internal/config/config.go"] += "\n// learner change\n"
	current.WorkspaceHash = attempt.WorkspaceHash(current.Workspace)
	builder, err := NewSpecBuilder(registry)
	if err != nil {
		t.Fatal(err)
	}
	spec, err := builder.Build(current, "execution-1", ActionSubmit)
	if err != nil {
		t.Fatal(err)
	}
	editable := findAsset(t, spec.Files, "internal/config/config.go")
	if editable.Origin != OriginLearnerWorkspace || editable.Access != AccessEditable || editable.SHA256 != SHA256Hex(editable.Content) {
		t.Fatalf("editable asset = %#v", editable)
	}
	visible := findAsset(t, spec.Files, "internal/config/visible_test.go")
	if visible.Origin != OriginReleaseBundle || visible.Access != AccessReadonly || visible.Role != RoleVisibleTest {
		t.Fatalf("visible asset = %#v", visible)
	}
	heldOut := findAsset(t, spec.Files, "internal/config/heldout_test.go")
	if heldOut.Origin != OriginReleaseBundle || heldOut.Access != AccessReadonly || heldOut.Role != RoleHeldOutTest {
		t.Fatalf("held-out asset = %#v", heldOut)
	}
	if heldOut.Content == "" || heldOut.SHA256 != SHA256Hex(heldOut.Content) {
		t.Fatal("held-out asset was not restored from the verified release")
	}
}

func TestSpecBuilderEnforcesFrozenActionAndWorkspace(t *testing.T) {
	registry := releaseRegistry(t)
	current := frozenAttempt(t, registry, "practice-error-contract-v1")
	builder, err := NewSpecBuilder(registry)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := builder.Build(current, "execution-1", ActionBuild); !errors.Is(err, ErrActionNotAllowed) {
		t.Fatalf("Build(disallowed action) error = %v", err)
	}
	current.Workspace["go.mod"] = "module attacker\n"
	current.WorkspaceHash = attempt.WorkspaceHash(current.Workspace)
	if _, err := builder.Build(current, "execution-2", ActionTest); err == nil {
		t.Fatal("Build(modified readonly asset) error = nil")
	}
}

func releaseRegistry(t *testing.T) *definition.Registry {
	t.Helper()
	registry, err := definition.LoadRegistry(definition.RegistryOptions{ContentDir: filepath.Join("..", "..", "..", "content", "learning")})
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func frozenAttempt(t *testing.T, registry *definition.Registry, taskID string) attempt.Attempt {
	t.Helper()
	task, err := registry.TaskView(registry.CurrentReleaseID(), taskID, 1)
	if err != nil {
		t.Fatal(err)
	}
	workspace, err := registry.PublicWorkspace(registry.CurrentReleaseID(), taskID, 1)
	if err != nil {
		t.Fatal(err)
	}
	return attempt.Attempt{
		ReleaseID: registry.CurrentReleaseID(), TaskID: task.ID, TaskVersion: task.Version,
		TaskHash: task.BundleHash, Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace),
	}
}

func findAsset(t *testing.T, assets []FileAsset, path string) FileAsset {
	t.Helper()
	for _, asset := range assets {
		if asset.Path == path {
			return asset
		}
	}
	t.Fatalf("asset %q not found", path)
	return FileAsset{}
}
