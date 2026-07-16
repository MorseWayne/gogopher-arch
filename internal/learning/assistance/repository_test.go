package assistance

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	platformdb "github.com/MorseWayne/gogopher-arch/internal/platform/database"
)

func TestPostgresRepositoryEnforcesEventIdempotencySequenceAndOwnership(t *testing.T) {
	fixture := setupAssistanceIntegration(t)
	first := fixture.record("00000000-0000-4000-8000-000000000101", "hint:first", HintRevealed, `{"hint_id":"first"}`)
	created, err := fixture.repository.Record(fixture.ctx, first)
	if err != nil || !created.Created || created.Event.Sequence != 1 {
		t.Fatalf("Record(first) = %#v, %v", created, err)
	}
	again, err := fixture.repository.Record(fixture.ctx, first)
	if err != nil || again.Created || again.Event.ID != created.Event.ID {
		t.Fatalf("Record(retry) = %#v, %v", again, err)
	}
	conflicting := first
	conflicting.Payload = []byte(`{"hint_id":"changed"}`)
	if _, err := fixture.repository.Record(fixture.ctx, conflicting); err == nil {
		t.Fatal("Record(conflicting retry) error = nil")
	} else {
		var conflict *IdempotencyConflict
		if !errors.As(err, &conflict) || conflict.EventID != first.ID {
			t.Fatalf("Record(conflicting retry) error = %v", err)
		}
	}

	records := []Record{
		fixture.record("00000000-0000-4000-8000-000000000102", "reference:first", ReferenceOpened, `{}`),
		fixture.record("00000000-0000-4000-8000-000000000103", "ai:first", AIDeclared, `{}`),
	}
	type result struct {
		value RecordResult
		err   error
	}
	results := make(chan result, len(records))
	var wait sync.WaitGroup
	for _, record := range records {
		wait.Add(1)
		go func(record Record) {
			defer wait.Done()
			value, err := fixture.repository.Record(fixture.ctx, record)
			results <- result{value: value, err: err}
		}(record)
	}
	wait.Wait()
	close(results)
	sequences := map[int64]bool{1: true}
	for result := range results {
		if result.err != nil || !result.value.Created {
			t.Fatalf("concurrent Record() = %#v", result)
		}
		sequences[result.value.Event.Sequence] = true
	}
	if len(sequences) != 3 || !sequences[2] || !sequences[3] {
		t.Fatalf("event sequences = %#v, want 1, 2, 3", sequences)
	}
	events, err := fixture.repository.ListThrough(fixture.ctx, fixture.learnerID, fixture.current.ID, 2)
	if err != nil || len(events) != 2 || events[0].Sequence != 1 || events[1].Sequence != 2 {
		t.Fatalf("ListThrough(cutoff 2) = %#v, %v", events, err)
	}
	if _, err := fixture.repository.ListThrough(fixture.ctx, "00000000-0000-4000-8000-000000000999", fixture.current.ID, 3); !errors.Is(err, ErrAttemptNotFound) {
		t.Fatalf("ListThrough(wrong owner) error = %v", err)
	}

	fixture.freeze(fixture.current.ID)
	if retried, err := fixture.repository.Record(fixture.ctx, first); err != nil || retried.Event.ID != first.ID {
		t.Fatalf("Record(existing after submit) = %#v, %v", retried, err)
	}
	late := fixture.record("00000000-0000-4000-8000-000000000104", "hint:late", HintRevealed, `{}`)
	if _, err := fixture.repository.Record(fixture.ctx, late); !errors.Is(err, ErrAttemptInactive) {
		t.Fatalf("Record(after submit) error = %v", err)
	}
}

func TestPostgresRepositorySharesAttemptLockWithSubmissionCutoff(t *testing.T) {
	fixture := setupAssistanceIntegration(t)
	for index := 0; index < 12; index++ {
		current, err := fixture.attempts.Create(fixture.ctx, attempt.CreateInput{
			LearnerID: fixture.learnerID, ActivityID: "assessment-check-config", ActivityVersion: 3,
		})
		if err != nil {
			t.Fatal(err)
		}
		event := fixture.recordFor(current.ID,
			fmt.Sprintf("00000000-0000-4000-8001-%012d", index+1),
			fmt.Sprintf("hint:race:%d", index), HintRevealed, `{"hint_id":"race"}`)
		start := make(chan struct{})
		type eventResult struct {
			result RecordResult
			err    error
		}
		eventDone := make(chan eventResult, 1)
		cutoffDone := make(chan struct {
			cutoff int64
			err    error
		}, 1)
		go func() {
			<-start
			result, err := fixture.repository.Record(fixture.ctx, event)
			eventDone <- eventResult{result: result, err: err}
		}()
		go func() {
			<-start
			cutoff, err := fixture.freezeWithCutoff(current.ID)
			cutoffDone <- struct {
				cutoff int64
				err    error
			}{cutoff: cutoff, err: err}
		}()
		close(start)
		eventOutcome, cutoffOutcome := <-eventDone, <-cutoffDone
		if cutoffOutcome.err != nil {
			t.Fatal(cutoffOutcome.err)
		}
		switch {
		case eventOutcome.err == nil:
			if !eventOutcome.result.Created || cutoffOutcome.cutoff != 1 {
				t.Fatalf("event won race: event=%#v cutoff=%d", eventOutcome.result, cutoffOutcome.cutoff)
			}
		case errors.Is(eventOutcome.err, ErrAttemptInactive):
			if cutoffOutcome.cutoff != 0 {
				t.Fatalf("submission won race: cutoff=%d, want 0", cutoffOutcome.cutoff)
			}
		default:
			t.Fatalf("Record(race) error = %v", eventOutcome.err)
		}
		events, err := fixture.repository.ListThrough(fixture.ctx, fixture.learnerID, current.ID, cutoffOutcome.cutoff)
		if err != nil || int64(len(events)) != cutoffOutcome.cutoff {
			t.Fatalf("persisted events through cutoff %d = %#v, %v", cutoffOutcome.cutoff, events, err)
		}
		independence := CalculateIndependence(current.Mode, events, cutoffOutcome.cutoff)
		want := IndependenceIndependent
		if cutoffOutcome.cutoff == 1 {
			want = IndependenceHinted
		}
		if independence != want {
			t.Fatalf("independence = %q, want %q", independence, want)
		}
	}
}

type assistanceIntegrationFixture struct {
	ctx        context.Context
	db         *sql.DB
	schema     string
	learnerID  string
	current    attempt.Attempt
	attempts   *attempt.Service
	repository *PostgresRepository
}

func setupAssistanceIntegration(t *testing.T) assistanceIntegrationFixture {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	t.Cleanup(cancel)
	db, err := platformdb.Open(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("assistance_test_%d", time.Now().UTC().UnixNano())
	if _, err := db.ExecContext(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA "`+schema+`" CASCADE`)
		_ = db.Close()
	})
	migrator, err := platformdb.NewMigrator(db, os.DirFS("../../../db/migrations"), platformdb.MigratorOptions{Schema: schema})
	if err != nil {
		t.Fatal(err)
	}
	if err := migrator.Up(ctx); err != nil {
		t.Fatal(err)
	}
	contentDir, _ := filepath.Abs("../../../content/learning")
	registry, err := definition.LoadRegistry(definition.RegistryOptions{ContentDir: contentDir})
	if err != nil {
		t.Fatal(err)
	}
	history, _ := definition.NewReleaseStore(db, definition.ReleaseStoreOptions{Schema: schema})
	if err := history.Register(ctx, registry); err != nil {
		t.Fatal(err)
	}
	learnerID := fmt.Sprintf("00000000-0000-4000-8000-%012d", time.Now().UnixNano()%1_000_000_000_000)
	if _, err := db.ExecContext(ctx, `INSERT INTO "`+schema+`".learners (id) VALUES ($1)`, learnerID); err != nil {
		t.Fatal(err)
	}
	attemptRepository, _ := attempt.NewPostgresRepository(db, attempt.RepositoryOptions{Schema: schema})
	attemptService, _ := attempt.NewService(attemptRepository, registry, attempt.ServiceOptions{})
	current, err := attemptService.Create(ctx, attempt.CreateInput{
		LearnerID: learnerID, ActivityID: "assessment-check-config", ActivityVersion: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	repository, _ := NewPostgresRepository(db, RepositoryOptions{Schema: schema})
	return assistanceIntegrationFixture{
		ctx: ctx, db: db, schema: schema, learnerID: learnerID, current: current.Attempt,
		attempts: attemptService, repository: repository,
	}
}

func (f assistanceIntegrationFixture) record(id, key string, eventType EventType, payload string) Record {
	return f.recordFor(f.current.ID, id, key, eventType, payload)
}

func (f assistanceIntegrationFixture) recordFor(attemptID, id, key string, eventType EventType, payload string) Record {
	return Record{
		ID: id, AttemptID: attemptID, LearnerID: f.learnerID, EventKey: key,
		Type: eventType, Payload: []byte(payload), CreatedAt: time.Now().UTC(),
	}
}

func (f assistanceIntegrationFixture) freeze(attemptID string) {
	if _, err := f.freezeWithCutoff(attemptID); err != nil {
		panic(err)
	}
}

func (f assistanceIntegrationFixture) freezeWithCutoff(attemptID string) (int64, error) {
	tx, err := f.db.BeginTx(f.ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(f.ctx, `SET LOCAL search_path TO "`+f.schema+`"`); err != nil {
		return 0, err
	}
	var status string
	if err := tx.QueryRowContext(f.ctx, `SELECT status FROM learning_attempts WHERE id = $1 AND learner_id = $2 FOR UPDATE`, attemptID, f.learnerID).Scan(&status); err != nil {
		return 0, err
	}
	if status != "active" {
		return 0, fmt.Errorf("attempt status = %s", status)
	}
	var cutoff int64
	if err := tx.QueryRowContext(f.ctx, `SELECT COALESCE(MAX(event_seq), 0) FROM assistance_events WHERE attempt_id = $1`, attemptID).Scan(&cutoff); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(f.ctx, `UPDATE learning_attempts SET status = 'submitted', submitted_at = now(), updated_at = GREATEST(updated_at, now()) WHERE id = $1`, attemptID); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return cutoff, nil
}
