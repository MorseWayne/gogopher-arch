package execution

import (
	"context"
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

func TestPostgresExecutionQueueEnforcesIdempotencyAndLeaseOwnership(t *testing.T) {
	fixture := setupExecutionIntegration(t)
	ctx, learnerID, current := fixture.ctx, fixture.learnerID, fixture.current
	repository, service := fixture.repository, fixture.service

	input := CreateInput{
		LearnerID: learnerID, AttemptID: current.ID, Action: ActionBuild,
		RequestKey: "build-1", WorkspaceRevision: current.WorkspaceRevision, WorkspaceHash: current.WorkspaceHash,
	}
	created, err := service.Create(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	again, err := service.Create(ctx, input)
	if err != nil || again.ID != created.ID {
		t.Fatalf("idempotent Create() = %#v, %v", again, err)
	}
	conflicting := input
	conflicting.Action = ActionTest
	if _, err := service.Create(ctx, conflicting); err == nil {
		t.Fatal("Create(conflicting key) error = nil")
	} else {
		var conflict *IdempotencyConflict
		if !errors.As(err, &conflict) || conflict.ExecutionID != created.ID {
			t.Fatalf("Create(conflicting key) error = %v", err)
		}
	}

	now := time.Now().UTC().Add(time.Second)
	claimed, ok, err := repository.Claim(ctx, "worker-1", now, 2*time.Second, 2)
	if err != nil || !ok || claimed.ID != created.ID || claimed.Status != ExecutionRunning || claimed.ClaimCount != 1 {
		t.Fatalf("Claim() = %#v, %v, %v", claimed, ok, err)
	}
	if renewed, err := repository.Renew(ctx, created.ID, "wrong-worker", now.Add(time.Second), 2*time.Second); err != nil || renewed {
		t.Fatalf("Renew(wrong owner) = %v, %v", renewed, err)
	}
	response := successResponse(created.Spec)
	if err := repository.Complete(ctx, created.ID, "wrong-worker", response, now.Add(time.Second)); !errors.Is(err, ErrLeaseLost) {
		t.Fatalf("Complete(wrong owner) error = %v", err)
	}
	if err := repository.Complete(ctx, created.ID, "worker-1", response, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	finished, err := service.Get(ctx, learnerID, created.ID)
	if err != nil || finished.Status != ExecutionSucceeded || finished.Response == nil {
		t.Fatalf("finished = %#v, %v", finished, err)
	}

	input.Action, input.RequestKey = ActionTest, "test-expiring"
	expiring, err := service.Create(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	first, ok, err := repository.Claim(ctx, "worker-1", now.Add(3*time.Second), time.Second, 2)
	if err != nil || !ok || first.ID != expiring.ID || first.ClaimCount != 1 {
		t.Fatalf("first expiring Claim() = %#v, %v, %v", first, ok, err)
	}
	second, ok, err := repository.Claim(ctx, "worker-2", now.Add(5*time.Second), time.Second, 2)
	if err != nil || !ok || second.ID != expiring.ID || second.ClaimCount != 2 {
		t.Fatalf("second expiring Claim() = %#v, %v, %v", second, ok, err)
	}
	if _, ok, err := repository.Claim(ctx, "worker-3", now.Add(7*time.Second), time.Second, 2); err != nil || ok {
		t.Fatalf("exhausted Claim() ok=%v error=%v", ok, err)
	}
	exhausted, err := service.Get(ctx, learnerID, expiring.ID)
	if err != nil || exhausted.Status != ExecutionInfraFailed || exhausted.Response == nil || exhausted.Response.Failure.Code != "lease_attempts_exhausted" {
		t.Fatalf("exhausted = %#v, %v", exhausted, err)
	}
	if _, err := service.Get(ctx, "00000000-0000-4000-8000-000000000999", created.ID); !errors.Is(err, ErrExecutionNotFound) {
		t.Fatalf("cross-owner Get() error = %v", err)
	}

	input.Action, input.RequestKey = ActionBuild, "concurrent-build"
	firstQueued, err := service.Create(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	input.Action, input.RequestKey = ActionVet, "concurrent-vet"
	secondQueued, err := service.Create(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	type claimResult struct {
		value Execution
		ok    bool
		err   error
	}
	results := make(chan claimResult, 2)
	var wait sync.WaitGroup
	for _, owner := range []string{"concurrent-worker-1", "concurrent-worker-2"} {
		wait.Add(1)
		go func() {
			defer wait.Done()
			value, ok, err := repository.Claim(ctx, owner, now.Add(8*time.Second), 2*time.Second, 2)
			results <- claimResult{value: value, ok: ok, err: err}
		}()
	}
	wait.Wait()
	close(results)
	claimedIDs := make(map[string]bool)
	for result := range results {
		if result.err != nil || !result.ok {
			t.Fatalf("concurrent Claim() = %#v", result)
		}
		claimedIDs[result.value.ID] = true
	}
	if len(claimedIDs) != 2 || !claimedIDs[firstQueued.ID] || !claimedIDs[secondQueued.ID] {
		t.Fatalf("concurrent claimed IDs = %#v", claimedIDs)
	}
}

func TestPostgresWorkerExecutesRealSandboxProcess(t *testing.T) {
	endpoint := os.Getenv("TEST_SANDBOX_ENDPOINT")
	if endpoint == "" {
		t.Skip("set TEST_SANDBOX_ENDPOINT to run Sandbox process integration test")
	}
	fixture := setupExecutionIntegration(t)
	created, err := fixture.service.Create(fixture.ctx, CreateInput{
		LearnerID: fixture.learnerID, AttemptID: fixture.current.ID, Action: ActionBuild,
		RequestKey: "real-sandbox-build", WorkspaceRevision: fixture.current.WorkspaceRevision,
		WorkspaceHash: fixture.current.WorkspaceHash,
	})
	if err != nil {
		t.Fatal(err)
	}
	client, err := NewSandboxClient(SandboxClientOptions{Endpoint: endpoint})
	if err != nil {
		t.Fatal(err)
	}
	maximum, err := fixture.registry.MaximumActionTimeout()
	if err != nil {
		t.Fatal(err)
	}
	worker, err := NewWorker(fixture.repository, client, WorkerOptions{
		Owner: "integration-worker", MaxActionTimeout: maximum, SandboxResponseGrace: 2 * time.Second,
		RPCDeadline: 20 * time.Second, PersistenceGrace: 2 * time.Second, LeaseDuration: 25 * time.Second,
		HeartbeatInterval: time.Second, PollInterval: 10 * time.Millisecond, MaxClaims: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if processed, err := worker.RunOnce(fixture.ctx); err != nil || !processed {
		t.Fatalf("RunOnce() processed=%v error=%v", processed, err)
	}
	finished, err := fixture.service.Get(fixture.ctx, fixture.learnerID, created.ID)
	if err != nil || finished.Status != ExecutionSucceeded || finished.Response == nil || len(finished.Response.Stages) != 1 {
		t.Fatalf("finished = %#v, %v", finished, err)
	}
}

type executionIntegrationFixture struct {
	ctx        context.Context
	learnerID  string
	current    attempt.Attempt
	registry   *definition.Registry
	repository *PostgresRepository
	service    *Service
}

func setupExecutionIntegration(t *testing.T) executionIntegrationFixture {
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
	schema := fmt.Sprintf("execution_test_%d", time.Now().UTC().UnixNano())
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
	current, err := attemptService.Create(ctx, attempt.CreateInput{LearnerID: learnerID, ActivityID: "guided-run-model", ActivityVersion: 3})
	if err != nil {
		t.Fatal(err)
	}
	repository, _ := NewPostgresRepository(db, RepositoryOptions{Schema: schema})
	builder, _ := NewSpecBuilder(registry)
	service, _ := NewService(repository, attemptService, builder, ServiceOptions{})
	return executionIntegrationFixture{
		ctx: ctx, learnerID: learnerID, current: current, registry: registry,
		repository: repository, service: service,
	}
}
