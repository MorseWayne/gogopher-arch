package execution

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
)

func TestServiceCreatesIdempotentNormalExecution(t *testing.T) {
	registry := releaseRegistry(t)
	current := frozenAttempt(t, registry, "guided-run-model-v2", 4)
	current.ID = "00000000-0000-4000-8000-000000000101"
	current.LearnerID = "00000000-0000-4000-8000-000000000102"
	current.Status = "active"
	builder, err := NewSpecBuilder(registry)
	if err != nil {
		t.Fatal(err)
	}
	repository := &serviceRepositoryStub{}
	service, err := NewService(repository, attemptReaderStub{value: current}, builder, ServiceOptions{
		Random: strings.NewReader(strings.Repeat("a", 32)),
		Now:    func() time.Time { return time.Date(2026, 7, 13, 7, 0, 0, 0, time.UTC) },
	})
	if err != nil {
		t.Fatal(err)
	}
	input := CreateInput{
		LearnerID: current.LearnerID, AttemptID: current.ID, Action: ActionBuild,
		RequestKey: "build-request-1", WorkspaceRevision: current.WorkspaceRevision, WorkspaceHash: current.WorkspaceHash,
	}
	created, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if created.Status != ExecutionQueued || repository.created.Spec.Action != ActionBuild || repository.created.Spec.ExecutionID != created.ID {
		t.Fatalf("created = %#v, record = %#v", created, repository.created)
	}
	if created.RequestFingerprint != RequestFingerprint(ActionBuild, 0, current.WorkspaceHash, current.TaskHash) {
		t.Fatalf("fingerprint = %q", created.RequestFingerprint)
	}
	repository.existing = &created
	again, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if again.ID != created.ID || repository.createCalls != 1 {
		t.Fatalf("idempotent result = %#v, create calls = %d", again, repository.createCalls)
	}
	input.Action = ActionTest
	if _, err := service.Create(context.Background(), input); err == nil {
		t.Fatal("Create(conflicting key) error = nil")
	} else {
		var conflict *IdempotencyConflict
		if !errors.As(err, &conflict) || conflict.ExecutionID != created.ID {
			t.Fatalf("Create(conflicting key) error = %v", err)
		}
	}
}

func TestServiceRejectsStaleWorkspaceBeforeQueueing(t *testing.T) {
	registry := releaseRegistry(t)
	current := frozenAttempt(t, registry, "guided-run-model-v2", 4)
	current.ID, current.LearnerID, current.Status = "attempt-1", "learner-1", "active"
	current.WorkspaceRevision = 3
	builder, _ := NewSpecBuilder(registry)
	repository := &serviceRepositoryStub{}
	service, err := NewService(repository, attemptReaderStub{value: current}, builder, ServiceOptions{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.Create(context.Background(), CreateInput{
		LearnerID: current.LearnerID, AttemptID: current.ID, Action: ActionBuild,
		RequestKey: "stale", WorkspaceRevision: 2, WorkspaceHash: current.WorkspaceHash,
	})
	var conflict *WorkspaceConflict
	if !errors.As(err, &conflict) || conflict.Revision != 3 || repository.createCalls != 0 {
		t.Fatalf("Create() error = %v, create calls = %d", err, repository.createCalls)
	}
}

type attemptReaderStub struct {
	value attempt.Attempt
	err   error
}

func (s attemptReaderStub) Get(context.Context, string, string) (attempt.Attempt, error) {
	return s.value, s.err
}

type serviceRepositoryStub struct {
	existing    *Execution
	created     CreateNormalRecord
	createCalls int
}

func (s *serviceRepositoryStub) FindNormal(context.Context, string, string, string) (Execution, error) {
	if s.existing != nil {
		return *s.existing, nil
	}
	return Execution{}, ErrExecutionNotFound
}

func (s *serviceRepositoryStub) CreateNormal(_ context.Context, record CreateNormalRecord) (Execution, bool, error) {
	s.created = record
	s.createCalls++
	return Execution{
		ID: record.ID, AttemptID: record.AttemptID, Action: record.Action,
		RequestKey: record.RequestKey, RequestFingerprint: record.RequestFingerprint,
		TaskHash: record.TaskHash, Spec: record.Spec, Status: ExecutionQueued,
	}, true, nil
}

func (s *serviceRepositoryStub) Get(context.Context, string, string) (Execution, error) {
	if s.existing == nil {
		return Execution{}, ErrExecutionNotFound
	}
	return *s.existing, nil
}
