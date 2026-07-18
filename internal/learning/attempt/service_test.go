package attempt

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

func TestServiceCreatesFrozenStarterAndEnforcesOwner(t *testing.T) {
	service, repository := newTestService(t)
	created, err := service.Create(context.Background(), CreateInput{
		LearnerID: "owner", ActivityID: "assessment-check-config", ActivityVersion: 4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.ReleaseID != "m1-first-slice-v13" || created.ActivityHash == "" || created.TaskHash == "" || len(created.CapabilityRefs) != 4 {
		t.Fatalf("frozen attempt = %#v", created)
	}
	if _, exists := created.Workspace["internal/config/heldout_test.go"]; exists {
		t.Fatal("starter exposed held-out test")
	}
	if created.WorkspaceHash != WorkspaceHash(created.Workspace) {
		t.Fatal("workspace hash mismatch")
	}
	if _, err := service.Get(context.Background(), "other-owner", created.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-owner Get() error = %v", err)
	}
	if len(repository.attempts) != 1 {
		t.Fatalf("attempt count = %d", len(repository.attempts))
	}
}

func TestServiceRequiresReviewItemClaimForReviewActivity(t *testing.T) {
	service, _ := newTestService(t)
	_, err := service.Create(context.Background(), CreateInput{
		LearnerID: "owner", ActivityID: "review-check-config-variant", ActivityVersion: 3,
	})
	if !errors.Is(err, ErrReviewClaimRequired) {
		t.Fatalf("Create(review) error = %v", err)
	}
}

func TestServiceSavesCompleteWorkspaceWithRevisionCAS(t *testing.T) {
	service, _ := newTestService(t)
	created, err := service.Create(context.Background(), CreateInput{LearnerID: "owner", ActivityID: "assessment-check-config", ActivityVersion: 4})
	if err != nil {
		t.Fatal(err)
	}
	files := cloneWorkspace(created.Workspace)
	files["internal/config/config.go"] += "\n// learner change\n"
	saved, err := service.Save(context.Background(), SaveInput{LearnerID: "owner", AttemptID: created.ID, BaseRevision: 0, Files: files})
	if err != nil {
		t.Fatal(err)
	}
	if saved.WorkspaceRevision != 1 || saved.WorkspaceHash != WorkspaceHash(files) {
		t.Fatalf("saved = %#v", saved)
	}
	_, err = service.Save(context.Background(), SaveInput{LearnerID: "owner", AttemptID: created.ID, BaseRevision: 0, Files: files})
	var conflict *RevisionConflict
	if !errors.As(err, &conflict) || conflict.Revision != 1 || conflict.Hash != saved.WorkspaceHash {
		t.Fatalf("stale Save() error = %#v", err)
	}
}

func TestServiceRejectsWorkspaceContractViolations(t *testing.T) {
	service, _ := newTestService(t)
	created, err := service.Create(context.Background(), CreateInput{LearnerID: "owner", ActivityID: "assessment-check-config", ActivityVersion: 4})
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string]map[string]string{
		"missing file":     cloneWorkspace(created.Workspace),
		"extra file":       cloneWorkspace(created.Workspace),
		"readonly changed": cloneWorkspace(created.Workspace),
	}
	delete(tests["missing file"], "go.mod")
	tests["extra file"]["secret.go"] = "package secret"
	tests["readonly changed"]["go.mod"] = "module changed"
	for name, files := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := service.Save(context.Background(), SaveInput{LearnerID: "owner", AttemptID: created.ID, BaseRevision: 0, Files: files}); err == nil {
				t.Fatal("Save() error = nil")
			}
		})
	}
}

func TestWorkspaceHashIsIndependentOfMapOrder(t *testing.T) {
	first := map[string]string{"b.go": "b", "a.go": "a"}
	second := map[string]string{"a.go": "a", "b.go": "b"}
	if WorkspaceHash(first) != WorkspaceHash(second) {
		t.Fatal("WorkspaceHash depends on map order")
	}
}

func TestValidateWorkspaceEnforcesFileAndTotalByteLimits(t *testing.T) {
	view := definition.TaskView{
		EditablePaths:   []string{"main.go", "notes.txt"},
		WorkspaceLimits: definition.WorkspaceLimitsView{MaxFiles: 2, MaxFileBytes: 5, MaxTotalBytes: 8},
	}
	baseline := map[string]string{"main.go": "", "notes.txt": ""}
	if err := ValidateWorkspace(view, baseline, map[string]string{"main.go": "123456", "notes.txt": ""}); err == nil {
		t.Fatal("ValidateWorkspace() accepted oversized file")
	}
	if err := ValidateWorkspace(view, baseline, map[string]string{"main.go": "12345", "notes.txt": "1234"}); err == nil {
		t.Fatal("ValidateWorkspace() accepted oversized workspace")
	}
}

func TestValidateWorkspaceSupportsBlankProjectTasksWithoutWeakeningReadonlyAssets(t *testing.T) {
	view := definition.TaskView{
		ReadonlyPaths:   []string{"README.md"},
		WorkspacePolicy: definition.WorkspacePolicyView{AllowNewFiles: true, AllowDeleteFiles: true},
		WorkspaceLimits: definition.WorkspaceLimitsView{MaxFiles: 5, MaxFileBytes: 1024, MaxTotalBytes: 4096},
	}
	baseline := map[string]string{"README.md": "contract"}
	workspace := map[string]string{
		"README.md":        "contract",
		"go.mod":           "module example.com/tool\n",
		"cmd/tool/main.go": "package main\n",
	}
	if err := ValidateWorkspace(view, baseline, workspace); err != nil {
		t.Fatalf("ValidateWorkspace(dynamic files) error = %v", err)
	}
	delete(workspace, "README.md")
	if err := ValidateWorkspace(view, baseline, workspace); err == nil {
		t.Fatal("ValidateWorkspace() accepted a missing readonly asset")
	}
	workspace["README.md"] = "contract"
	workspace["../escape.go"] = "package escape"
	if err := ValidateWorkspace(view, baseline, workspace); err == nil {
		t.Fatal("ValidateWorkspace() accepted path traversal")
	}
}

func newTestService(t *testing.T) (*Service, *memoryRepository) {
	t.Helper()
	contentDir, err := filepath.Abs("../../../content/learning")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := definition.LoadRegistry(definition.RegistryOptions{ContentDir: contentDir})
	if err != nil {
		t.Fatal(err)
	}
	repository := &memoryRepository{attempts: make(map[string]Attempt)}
	service, err := NewService(repository, registry, ServiceOptions{
		Random: bytes.NewReader(make([]byte, 128)),
		Now:    func() time.Time { return time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatal(err)
	}
	return service, repository
}

type memoryRepository struct{ attempts map[string]Attempt }

func (r *memoryRepository) Create(_ context.Context, record CreateRecord) (CreateResult, error) {
	r.attempts[record.ID] = cloneAttempt(record.Attempt)
	return CreateResult{Attempt: cloneAttempt(record.Attempt), Created: true}, nil
}
func (r *memoryRepository) Get(_ context.Context, learnerID, attemptID string) (Attempt, error) {
	value, exists := r.attempts[attemptID]
	if !exists || value.LearnerID != learnerID {
		return Attempt{}, ErrNotFound
	}
	return cloneAttempt(value), nil
}
func (r *memoryRepository) Save(_ context.Context, record SaveRecord) (Attempt, error) {
	value, exists := r.attempts[record.AttemptID]
	if !exists || value.LearnerID != record.LearnerID {
		return Attempt{}, ErrNotFound
	}
	if value.Status != "active" {
		return Attempt{}, ErrInactive
	}
	if value.WorkspaceRevision != record.BaseRevision {
		return Attempt{}, &RevisionConflict{Revision: value.WorkspaceRevision, Hash: value.WorkspaceHash}
	}
	value.Workspace = cloneWorkspace(record.Workspace)
	value.WorkspaceHash = record.WorkspaceHash
	value.WorkspaceRevision++
	value.UpdatedAt = record.UpdatedAt
	r.attempts[value.ID] = value
	return cloneAttempt(value), nil
}
