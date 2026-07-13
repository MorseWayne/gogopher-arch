package assistance

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

func TestServiceRecordsOnlyEventsAllowedByFrozenActivity(t *testing.T) {
	registry := loadAssistanceRegistry(t)
	repository := &repositoryStub{result: RecordResult{Created: true}}
	service := newAssistanceService(t, repository, attempt.Attempt{
		ID: "attempt-1", LearnerID: "learner-1", ReleaseID: registry.CurrentReleaseID(),
		ActivityID: "review-check-config-variant", ActivityVersion: 1,
	}, registry)

	if _, err := service.Record(context.Background(), RecordInput{
		LearnerID: "learner-1", AttemptID: "attempt-1", EventKey: "hint-1",
		Type: HintRevealed, Payload: map[string]any{"hint_id": "first"},
	}); !errors.Is(err, ErrEventNotAllowed) {
		t.Fatalf("Record(disallowed hint) error = %v", err)
	}
	result, err := service.Record(context.Background(), RecordInput{
		LearnerID: "learner-1", AttemptID: "attempt-1", EventKey: "ai-1", Type: AIDeclared,
	})
	if err != nil || !result.Created {
		t.Fatalf("Record(AI declaration) = %#v, %v", result, err)
	}
	if got := string(repository.record.Payload); got != "{}" {
		t.Fatalf("nil payload persisted as %s, want {}", got)
	}
	if repository.record.CreatedAt.Location() != time.UTC {
		t.Fatalf("CreatedAt location = %v, want UTC", repository.record.CreatedAt.Location())
	}
}

func TestRevealHintReturnsContentOnlyAfterEventCommit(t *testing.T) {
	registry := loadAssistanceRegistry(t)
	current := attempt.Attempt{
		ID: "attempt-1", LearnerID: "learner-1", ReleaseID: registry.CurrentReleaseID(),
		ActivityID: "assessment-check-config", ActivityVersion: 1,
	}
	hint := Hint{ID: "first", Title: "First step", Body: "Inspect the failing contract."}
	repository := &repositoryStub{err: errors.New("commit failed")}
	service := newAssistanceService(t, repository, current, registry)
	got, _, err := service.RevealHint(context.Background(), RevealHintInput{
		LearnerID: current.LearnerID, AttemptID: current.ID, EventKey: "hint:first", Hint: hint,
	})
	if err == nil || got != (Hint{}) {
		t.Fatalf("RevealHint(failed event) = %#v, %v; content must remain hidden", got, err)
	}

	repository.err = nil
	repository.result = RecordResult{Created: true}
	got, result, err := service.RevealHint(context.Background(), RevealHintInput{
		LearnerID: current.LearnerID, AttemptID: current.ID, EventKey: "hint:first", Hint: hint,
	})
	if err != nil || got != hint || !result.Created {
		t.Fatalf("RevealHint() = %#v, %#v, %v", got, result, err)
	}
	if repository.record.Type != HintRevealed || string(repository.record.Payload) != `{"hint_id":"first"}` {
		t.Fatalf("hint event = %#v", repository.record)
	}
}

type repositoryStub struct {
	record Record
	result RecordResult
	err    error
}

func (s *repositoryStub) Record(_ context.Context, record Record) (RecordResult, error) {
	s.record = record
	return s.result, s.err
}

func (s *repositoryStub) ListThrough(context.Context, string, string, int64) ([]Event, error) {
	return nil, s.err
}

type attemptReaderStub struct{ current attempt.Attempt }

func (s attemptReaderStub) Get(context.Context, string, string) (attempt.Attempt, error) {
	return s.current, nil
}

func loadAssistanceRegistry(t *testing.T) *definition.Registry {
	t.Helper()
	contentDir, err := filepath.Abs("../../../content/learning")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := definition.LoadRegistry(definition.RegistryOptions{ContentDir: contentDir})
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func newAssistanceService(t *testing.T, repository Repository, current attempt.Attempt, registry *definition.Registry) *Service {
	t.Helper()
	service, err := NewService(repository, attemptReaderStub{current: current}, registry, ServiceOptions{
		Random: bytes.NewReader(make([]byte, 64)),
		Now:    func() time.Time { return time.Date(2026, 7, 13, 8, 30, 0, 0, time.FixedZone("test", 8*60*60)) },
	})
	if err != nil {
		t.Fatal(err)
	}
	return service
}
