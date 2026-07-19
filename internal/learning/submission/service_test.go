package submission

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
)

func TestServiceRequiresAndFreezesGuidedExplanation(t *testing.T) {
	contentDir, err := filepath.Abs("../../../content/learning")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := definition.LoadRegistry(definition.RegistryOptions{ContentDir: contentDir})
	if err != nil {
		t.Fatal(err)
	}
	releaseID := registry.CurrentReleaseID()
	activity, err := registry.ActivityView(releaseID, "guided-run-model", 8)
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
	current := attempt.Attempt{
		ID: "attempt-guided", LearnerID: "learner-guided", ReleaseID: releaseID,
		ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash,
		TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
		Mode: activity.Mode, Status: "active", Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace),
	}
	repository := &submissionRepositoryStub{}
	builder, err := execution.NewSpecBuilder(registry)
	if err != nil {
		t.Fatal(err)
	}
	service, err := NewService(repository, submissionAttemptStub{value: current}, registry, builder, ServiceOptions{
		Random: bytes.NewReader(make([]byte, 32)),
	})
	if err != nil {
		t.Fatal(err)
	}
	base := SubmitInput{
		LearnerID: current.LearnerID, AttemptID: current.ID, SubmissionKey: "guided-submit",
		WorkspaceRevision: current.WorkspaceRevision, WorkspaceHash: current.WorkspaceHash,
	}

	for _, explanation := range []string{"", strings.Repeat("学", 19)} {
		input := base
		input.Explanation = explanation
		if _, err := service.Submit(context.Background(), input); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("Submit(explanation length %d) error = %v, want ErrInvalidRequest", len([]rune(explanation)), err)
		}
	}
	input := base
	input.Explanation = "  " + strings.Repeat("学", 20) + "  "
	if _, err := service.Submit(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if repository.freezeCalls != 1 || repository.record.Explanation != strings.Repeat("学", 20) {
		t.Fatalf("Freeze() calls=%d explanation=%q", repository.freezeCalls, repository.record.Explanation)
	}
	if RequestFingerprint(0, current.WorkspaceHash, current.TaskHash, activity.RuleSetHash, "first") ==
		RequestFingerprint(0, current.WorkspaceHash, current.TaskHash, activity.RuleSetHash, "second") {
		t.Fatal("request fingerprint does not bind the frozen explanation")
	}
}

type submissionAttemptStub struct{ value attempt.Attempt }

func (s submissionAttemptStub) Get(context.Context, string, string) (attempt.Attempt, error) {
	return s.value, nil
}

type submissionRepositoryStub struct {
	record      FreezeRecord
	freezeCalls int
}

func (s *submissionRepositoryStub) Freeze(_ context.Context, record FreezeRecord) (Result, error) {
	s.record = record
	s.freezeCalls++
	return Result{Submission: Submission{ID: record.SubmissionID, Explanation: record.Explanation}, Created: true}, nil
}

func (s *submissionRepositoryStub) Get(context.Context, string, string) (Submission, error) {
	return Submission{}, ErrNotFound
}

func (s *submissionRepositoryStub) Retry(context.Context, RetryRecord) (Result, error) {
	return Result{}, ErrRetryUnavailable
}
