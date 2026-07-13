package submission

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/assistance"
	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	platformdb "github.com/MorseWayne/gogopher-arch/internal/platform/database"
)

func TestPostgresWorkflowFreezesSubmissionIdempotently(t *testing.T) {
	fixture := setupSubmissionIntegration(t)
	current := fixture.createAttempt(t)
	assistanceRepository, _ := assistance.NewPostgresRepository(fixture.db, assistance.RepositoryOptions{Schema: fixture.schema})
	if _, err := assistanceRepository.Record(fixture.ctx, assistance.Record{
		ID: "00000000-0000-4000-8100-000000000001", AttemptID: current.ID, LearnerID: fixture.learnerID,
		EventKey: "hint:before-submit", Type: assistance.HintRevealed,
		Payload: []byte("{\"hint_id\":\"first\"}"), CreatedAt: time.Now().UTC(),
	}); err != nil {
		t.Fatal(err)
	}
	input := SubmitInput{
		LearnerID: fixture.learnerID, AttemptID: current.ID, SubmissionKey: "submit-1",
		WorkspaceRevision: current.WorkspaceRevision, WorkspaceHash: current.WorkspaceHash,
	}
	created, err := fixture.service.Submit(fixture.ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if !created.Created || created.ExecutionSequence != 0 ||
		created.Submission.Status != StatusExecuting ||
		created.Submission.LatestExecutionStatus != execution.ExecutionQueued ||
		created.Submission.AssistanceCutoff != 1 {
		t.Fatalf("Submit() = %#v", created)
	}
	if created.Submission.RuleSetHash == "" || !maps.Equal(created.Submission.Workspace, current.Workspace) {
		t.Fatalf("frozen submission lost rule set or workspace: %#v", created.Submission)
	}
	frozenAttempt, err := fixture.attempts.Get(fixture.ctx, fixture.learnerID, current.ID)
	if err != nil || frozenAttempt.Status != "submitted" {
		t.Fatalf("frozen attempt = %#v, %v", frozenAttempt, err)
	}

	again, err := fixture.service.Submit(fixture.ctx, input)
	if err != nil || again.Created || again.Submission.ID != created.Submission.ID ||
		again.ExecutionID != created.ExecutionID {
		t.Fatalf("Submit(idempotent retry) = %#v, %v", again, err)
	}
	changed := input
	changed.WorkspaceHash = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	if _, err := fixture.service.Submit(fixture.ctx, changed); err == nil {
		t.Fatal("Submit(changed same key) error = nil")
	} else {
		var conflict *IdempotencyConflict
		if !errors.As(err, &conflict) || conflict.SubmissionID != created.Submission.ID {
			t.Fatalf("Submit(changed same key) error = %v", err)
		}
	}
	different := input
	different.SubmissionKey = "submit-2"
	if _, err := fixture.service.Submit(fixture.ctx, different); err == nil {
		t.Fatal("Submit(different key) error = nil")
	} else {
		var already *AttemptAlreadySubmitted
		if !errors.As(err, &already) || already.SubmissionID != created.Submission.ID {
			t.Fatalf("Submit(different key) error = %v", err)
		}
	}
	if _, err := fixture.service.Get(fixture.ctx, "00000000-0000-4000-8000-000000000999", created.Submission.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get(wrong owner) error = %v", err)
	}
	wrongOwner := input
	wrongOwner.LearnerID = "00000000-0000-4000-8000-000000000999"
	if _, err := fixture.service.Submit(fixture.ctx, wrongOwner); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Submit(wrong owner) error = %v", err)
	}

	var submissionCount, executionCount int
	if err := fixture.db.QueryRowContext(fixture.ctx,
		"SELECT count(*) FROM \""+fixture.schema+"\".attempt_submissions WHERE attempt_id = $1", current.ID,
	).Scan(&submissionCount); err != nil {
		t.Fatal(err)
	}
	if err := fixture.db.QueryRowContext(fixture.ctx,
		"SELECT count(*) FROM \""+fixture.schema+"\".attempt_executions WHERE submission_id = $1", created.Submission.ID,
	).Scan(&executionCount); err != nil {
		t.Fatal(err)
	}
	if submissionCount != 1 || executionCount != 1 {
		t.Fatalf("rows after retries = submissions %d, executions %d", submissionCount, executionCount)
	}
}

func TestPostgresWorkflowSerializesDifferentSubmissionKeys(t *testing.T) {
	fixture := setupSubmissionIntegration(t)
	current := fixture.createAttempt(t)
	base := SubmitInput{
		LearnerID: fixture.learnerID, AttemptID: current.ID,
		WorkspaceRevision: current.WorkspaceRevision, WorkspaceHash: current.WorkspaceHash,
	}
	type outcome struct {
		result Result
		err    error
	}
	start := make(chan struct{})
	outcomes := make(chan outcome, 2)
	var wait sync.WaitGroup
	for _, key := range []string{"concurrent-a", "concurrent-b"} {
		wait.Add(1)
		go func(key string) {
			defer wait.Done()
			<-start
			input := base
			input.SubmissionKey = key
			result, err := fixture.service.Submit(fixture.ctx, input)
			outcomes <- outcome{result: result, err: err}
		}(key)
	}
	close(start)
	wait.Wait()
	close(outcomes)
	var winner Result
	var already *AttemptAlreadySubmitted
	successes, conflicts := 0, 0
	for outcome := range outcomes {
		if outcome.err == nil {
			successes++
			winner = outcome.result
			continue
		}
		if errors.As(outcome.err, &already) {
			conflicts++
			continue
		}
		t.Fatalf("concurrent Submit() error = %v", outcome.err)
	}
	if successes != 1 || conflicts != 1 || !winner.Created || already.SubmissionID != winner.Submission.ID {
		t.Fatalf("concurrent outcomes: winner=%#v successes=%d conflicts=%d already=%#v", winner, successes, conflicts, already)
	}

	staleAttempt := fixture.createAttempt(t)
	if _, err := fixture.service.Submit(fixture.ctx, SubmitInput{
		LearnerID: fixture.learnerID, AttemptID: staleAttempt.ID, SubmissionKey: "stale",
		WorkspaceRevision: staleAttempt.WorkspaceRevision + 1, WorkspaceHash: staleAttempt.WorkspaceHash,
	}); err == nil {
		t.Fatal("Submit(stale revision) error = nil")
	} else {
		var conflict *WorkspaceConflict
		if !errors.As(err, &conflict) || conflict.Revision != staleAttempt.WorkspaceRevision {
			t.Fatalf("Submit(stale revision) error = %v", err)
		}
	}
}

func TestPostgresWorkflowRetriesOnlyInfrastructureFailures(t *testing.T) {
	fixture := setupSubmissionIntegration(t)
	current := fixture.createAttempt(t)
	created, err := fixture.service.Submit(fixture.ctx, SubmitInput{
		LearnerID: fixture.learnerID, AttemptID: current.ID, SubmissionKey: "retryable-submit",
		WorkspaceRevision: current.WorkspaceRevision, WorkspaceHash: current.WorkspaceHash,
	})
	if err != nil {
		t.Fatal(err)
	}
	retryInput := RetryInput{
		LearnerID: fixture.learnerID, SubmissionID: created.Submission.ID, RequestKey: "retry-1",
	}
	if _, err := fixture.service.Retry(fixture.ctx, retryInput); !errors.Is(err, ErrRetryUnavailable) {
		t.Fatalf("Retry(before infra failure) error = %v", err)
	}

	now := time.Now().UTC().Add(time.Second)
	claimed, ok, err := fixture.executions.Claim(fixture.ctx, "submission-worker", now, 10*time.Second, 2)
	if err != nil || !ok || claimed.ID != created.ExecutionID {
		t.Fatalf("Claim() = %#v, %v, %v", claimed, ok, err)
	}
	response := execution.ExecutionResponse{
		ProtocolVersion: execution.ProtocolVersion, ExecutionID: claimed.ID,
		Status: execution.ExecutionInfraFailed, Stages: []execution.StageResult{}, DurationMS: 1,
		Policy: execution.PolicyReport{Network: execution.NetworkPolicyReport{
			Requested: claimed.Spec.Policy.Network, Enforcement: execution.EnforcementPolicyOnly,
		}},
		Failure: &execution.Failure{Code: "sandbox_unreachable", Message: "Sandbox unavailable"},
	}
	if err := fixture.executions.Complete(fixture.ctx, claimed.ID, "submission-worker", response, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	failed, err := fixture.service.Get(fixture.ctx, fixture.learnerID, created.Submission.ID)
	if err != nil || failed.Status != StatusInfraFailed || failed.LatestExecutionStatus != execution.ExecutionInfraFailed {
		t.Fatalf("failed submission = %#v, %v", failed, err)
	}
	failedAttempt, err := fixture.attempts.Get(fixture.ctx, fixture.learnerID, current.ID)
	if err != nil || failedAttempt.Status != "submit_infra_failed" {
		t.Fatalf("failed attempt = %#v, %v", failedAttempt, err)
	}

	retried, err := fixture.service.Retry(fixture.ctx, retryInput)
	if err != nil || !retried.Created || retried.ExecutionSequence != 1 ||
		retried.ExecutionID == created.ExecutionID ||
		retried.Submission.Status != StatusExecuting ||
		retried.Submission.LatestExecutionStatus != execution.ExecutionQueued {
		t.Fatalf("Retry() = %#v, %v", retried, err)
	}
	if !maps.Equal(retried.Submission.Workspace, created.Submission.Workspace) ||
		retried.Submission.WorkspaceHash != created.Submission.WorkspaceHash ||
		retried.Submission.AssistanceCutoff != created.Submission.AssistanceCutoff {
		t.Fatalf("Retry() changed frozen facts: %#v", retried.Submission)
	}
	again, err := fixture.service.Retry(fixture.ctx, retryInput)
	if err != nil || again.Created || again.ExecutionID != retried.ExecutionID || again.ExecutionSequence != 1 {
		t.Fatalf("Retry(idempotent) = %#v, %v", again, err)
	}
	secondKey := retryInput
	secondKey.RequestKey = "retry-2"
	if _, err := fixture.service.Retry(fixture.ctx, secondKey); !errors.Is(err, ErrRetryUnavailable) {
		t.Fatalf("Retry(while executing) error = %v", err)
	}
	resumedAttempt, err := fixture.attempts.Get(fixture.ctx, fixture.learnerID, current.ID)
	if err != nil || resumedAttempt.Status != "submitted" {
		t.Fatalf("resumed attempt = %#v, %v", resumedAttempt, err)
	}
}

func TestPostgresWorkflowMarksExhaustedSubmitLeaseAsInfrastructureFailure(t *testing.T) {
	fixture := setupSubmissionIntegration(t)
	current := fixture.createAttempt(t)
	created, err := fixture.service.Submit(fixture.ctx, SubmitInput{
		LearnerID: fixture.learnerID, AttemptID: current.ID, SubmissionKey: "exhausted-submit",
		WorkspaceRevision: current.WorkspaceRevision, WorkspaceHash: current.WorkspaceHash,
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Add(time.Second)
	claimed, ok, err := fixture.executions.Claim(fixture.ctx, "lost-worker", now, time.Second, 1)
	if err != nil || !ok || claimed.ID != created.ExecutionID {
		t.Fatalf("Claim() = %#v, %v, %v", claimed, ok, err)
	}
	if _, ok, err := fixture.executions.Claim(fixture.ctx, "replacement-worker", now.Add(2*time.Second), time.Second, 1); err != nil || ok {
		t.Fatalf("Claim(after exhausted lease) ok=%v error=%v", ok, err)
	}
	failed, err := fixture.service.Get(fixture.ctx, fixture.learnerID, created.Submission.ID)
	if err != nil || failed.Status != StatusInfraFailed ||
		failed.LatestExecutionStatus != execution.ExecutionInfraFailed {
		t.Fatalf("exhausted submission = %#v, %v", failed, err)
	}
	failedAttempt, err := fixture.attempts.Get(fixture.ctx, fixture.learnerID, current.ID)
	if err != nil || failedAttempt.Status != "submit_infra_failed" {
		t.Fatalf("exhausted attempt = %#v, %v", failedAttempt, err)
	}
}

type submissionIntegrationFixture struct {
	ctx        context.Context
	db         *sql.DB
	schema     string
	learnerID  string
	attempts   *attempt.Service
	service    *Service
	executions *execution.PostgresRepository
}

func setupSubmissionIntegration(t *testing.T) submissionIntegrationFixture {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	t.Cleanup(cancel)
	db, err := platformdb.Open(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("submission_test_%d", time.Now().UTC().UnixNano())
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA \""+schema+"\""); err != nil {
		db.Close()
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DROP SCHEMA \""+schema+"\" CASCADE")
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
	learnerID := fmt.Sprintf("00000000-0000-4000-8200-%012d", time.Now().UnixNano()%1_000_000_000_000)
	if _, err := db.ExecContext(ctx, "INSERT INTO \""+schema+"\".learners (id) VALUES ($1)", learnerID); err != nil {
		t.Fatal(err)
	}
	attemptRepository, _ := attempt.NewPostgresRepository(db, attempt.RepositoryOptions{Schema: schema})
	attemptService, _ := attempt.NewService(attemptRepository, registry, attempt.ServiceOptions{})
	specBuilder, _ := execution.NewSpecBuilder(registry)
	repository, _ := NewPostgresRepository(db, RepositoryOptions{Schema: schema})
	service, _ := NewService(repository, attemptService, registry, specBuilder, ServiceOptions{})
	executionRepository, _ := execution.NewPostgresRepository(db, execution.RepositoryOptions{Schema: schema})
	return submissionIntegrationFixture{
		ctx: ctx, db: db, schema: schema, learnerID: learnerID,
		attempts: attemptService, service: service, executions: executionRepository,
	}
}

func (f submissionIntegrationFixture) createAttempt(t *testing.T) attempt.Attempt {
	t.Helper()
	current, err := f.attempts.Create(f.ctx, attempt.CreateInput{
		LearnerID: f.learnerID, ActivityID: "assessment-check-config", ActivityVersion: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	return current
}
