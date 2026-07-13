package evaluation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/assistance"
	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
	platformdb "github.com/MorseWayne/gogopher-arch/internal/platform/database"
)

func TestPostgresEvaluationWorkerCommitsAtomicEvidenceBatch(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	db, err := platformdb.Open(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("evaluation_test_%d", time.Now().UTC().UnixNano())
	if _, err := db.ExecContext(ctx, "CREATE SCHEMA \""+schema+"\""); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.ExecContext(context.Background(), "DROP SCHEMA \""+schema+"\" CASCADE")
		_ = db.Close()
	}()
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
	learnerID := fmt.Sprintf("00000000-0000-4000-8500-%012d", time.Now().UnixNano()%1_000_000_000_000)
	if _, err := db.ExecContext(ctx, "INSERT INTO \""+schema+"\".learners (id) VALUES ($1)", learnerID); err != nil {
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
	assistanceRepository, _ := assistance.NewPostgresRepository(db, assistance.RepositoryOptions{Schema: schema})
	assistanceService, _ := assistance.NewService(assistanceRepository, attemptService, registry, assistance.ServiceOptions{})
	if _, err := assistanceService.Record(ctx, assistance.RecordInput{
		LearnerID: learnerID, AttemptID: current.ID, EventKey: "hint:before-evaluation",
		Type: assistance.HintRevealed, Payload: map[string]any{"hint_id": "first"},
	}); err != nil {
		t.Fatal(err)
	}
	specBuilder, _ := execution.NewSpecBuilder(registry)
	executionRepository, _ := execution.NewPostgresRepository(db, execution.RepositoryOptions{Schema: schema})
	executionService, _ := execution.NewService(executionRepository, attemptService, specBuilder, execution.ServiceOptions{})
	submissionRepository, _ := submission.NewPostgresRepository(db, submission.RepositoryOptions{Schema: schema})
	submissionService, _ := submission.NewService(submissionRepository, attemptService, registry, specBuilder, submission.ServiceOptions{})
	submitted, err := submissionService.Submit(ctx, submission.SubmitInput{
		LearnerID: learnerID, AttemptID: current.ID, SubmissionKey: "evaluate-submit",
		WorkspaceRevision: current.WorkspaceRevision, WorkspaceHash: current.WorkspaceHash,
	})
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC().Add(time.Second)
	claimed, ok, err := executionRepository.Claim(ctx, "execution-worker", now, 10*time.Second, 2)
	if err != nil || !ok || claimed.ID != submitted.ExecutionID {
		t.Fatalf("Claim() = %#v, %v, %v", claimed, ok, err)
	}
	response := successfulAssessmentResponse(claimed)
	if err := executionRepository.Complete(ctx, claimed.ID, "execution-worker", response, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}

	generator, _ := NewGenerator(registry)
	evaluationRepository, _ := NewPostgresRepository(db, RepositoryOptions{Schema: schema})
	evaluationService, _ := NewService(
		evaluationRepository, submissionService, executionService, assistanceService,
		registry, generator, ServiceOptions{},
	)
	worker, _ := NewWorker(evaluationRepository, evaluationService, WorkerOptions{
		Owner: "evaluation-worker", Lease: 10 * time.Second, PollInterval: time.Millisecond,
		Now: func() time.Time { return now.Add(2 * time.Second) },
	})
	if processed, err := worker.RunOnce(ctx); err != nil || !processed {
		t.Fatalf("RunOnce() processed=%v error=%v", processed, err)
	}
	batch, created, err := evaluationService.Evaluate(ctx, learnerID, submitted.Submission.ID, claimed.ID)
	if err != nil || created {
		t.Fatalf("Evaluate(replay) = %#v, created=%v, error=%v", batch, created, err)
	}
	if len(batch.RuleResults) != 10 || len(batch.Artifacts) != 4 || len(batch.Evidence) != 10 {
		t.Fatalf("batch sizes = rules %d, artifacts %d, evidence %d",
			len(batch.RuleResults), len(batch.Artifacts), len(batch.Evidence))
	}
	artifactKinds := make(map[string]string, len(batch.Artifacts))
	for _, artifact := range batch.Artifacts {
		if artifact.ContentBytes != len(artifact.Content) || artifact.ContentHash != definition.SHA256Hex(artifact.Content) {
			t.Fatalf("artifact metadata = %#v", artifact)
		}
		artifactKinds[artifact.Kind] = artifact.ID
	}
	for _, kind := range []string{"workspace", "diff", "explanation", "test_report"} {
		if artifactKinds[kind] == "" {
			t.Fatalf("artifact kinds = %#v", artifactKinds)
		}
	}
	for _, evidence := range batch.Evidence {
		if evidence.Independence != assistance.IndependenceHinted ||
			evidence.Evaluator != "deterministic" || evidence.EvaluationBatchID != batch.ID ||
			(evidence.ArtifactID != artifactKinds["diff"] && evidence.ArtifactID != artifactKinds["test_report"]) {
			t.Fatalf("evidence = %#v", evidence)
		}
	}
	finishedSubmission, err := submissionService.Get(ctx, learnerID, submitted.Submission.ID)
	if err != nil || finishedSubmission.Status != submission.StatusEvaluated {
		t.Fatalf("finished submission = %#v, %v", finishedSubmission, err)
	}
	finishedAttempt, err := attemptService.Get(ctx, learnerID, current.ID)
	if err != nil || finishedAttempt.Status != "completed" {
		t.Fatalf("finished attempt = %#v, %v", finishedAttempt, err)
	}
	var batches, artifacts, evidenceRecords, completedRequests, projectionRequests int
	for query, target := range map[string]*int{
		"SELECT count(*) FROM \"" + schema + "\".evaluation_batches":                                                                     &batches,
		"SELECT count(*) FROM \"" + schema + "\".artifacts":                                                                              &artifacts,
		"SELECT count(*) FROM \"" + schema + "\".evidence_records":                                                                       &evidenceRecords,
		"SELECT count(*) FROM \"" + schema + "\".learning_outbox WHERE topic = 'submission.evaluate' AND status = 'completed'":           &completedRequests,
		"SELECT count(*) FROM \"" + schema + "\".learning_outbox WHERE topic = 'capability_projection.requested' AND status = 'pending'": &projectionRequests,
	} {
		if err := db.QueryRowContext(ctx, query).Scan(target); err != nil {
			t.Fatal(err)
		}
	}
	if batches != 1 || artifacts != 4 || evidenceRecords != 10 || completedRequests != 1 || projectionRequests != 1 {
		t.Fatalf("persisted counts = batch %d artifacts %d evidence %d completed %d projection %d",
			batches, artifacts, evidenceRecords, completedRequests, projectionRequests)
	}
	var projectionEventVersion int
	if err := db.QueryRowContext(ctx, `SELECT (payload->>'event_version')::integer FROM "`+schema+`".learning_outbox WHERE topic='capability_projection.requested'`).Scan(&projectionEventVersion); err != nil || projectionEventVersion != 1 {
		t.Fatalf("projection event version = %d, %v", projectionEventVersion, err)
	}

	rollbackAttempt, err := attemptService.Create(ctx, attempt.CreateInput{
		LearnerID: learnerID, ActivityID: "assessment-check-config", ActivityVersion: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	rollbackSubmission, err := submissionService.Submit(ctx, submission.SubmitInput{
		LearnerID: learnerID, AttemptID: rollbackAttempt.ID, SubmissionKey: "rollback-submit",
		WorkspaceRevision: rollbackAttempt.WorkspaceRevision, WorkspaceHash: rollbackAttempt.WorkspaceHash,
	})
	if err != nil {
		t.Fatal(err)
	}
	rollbackExecution, ok, err := executionRepository.Claim(ctx, "rollback-worker", now.Add(3*time.Second), 10*time.Second, 2)
	if err != nil || !ok || rollbackExecution.ID != rollbackSubmission.ExecutionID {
		t.Fatalf("rollback Claim() = %#v, %v, %v", rollbackExecution, ok, err)
	}
	if err := executionRepository.Complete(ctx, rollbackExecution.ID, "rollback-worker", successfulAssessmentResponse(rollbackExecution), now.Add(4*time.Second)); err != nil {
		t.Fatal(err)
	}
	capability, err := registry.Get(definition.DefinitionRef{
		ReleaseID: registry.CurrentReleaseID(), Kind: definition.KindCapability, ID: "M1-01", Version: 2,
	})
	if err != nil {
		t.Fatal(err)
	}
	badBatchID := "00000000-0000-4000-8500-000000000901"
	rollbackArtifacts := persistedArtifactFixtures(t, rollbackAttempt.ID, rollbackSubmission.Submission.ID, now.Add(5*time.Second))
	_, _, err = evaluationRepository.Persist(ctx, PersistRecord{
		Batch: Batch{
			ID: badBatchID, SubmissionID: rollbackSubmission.Submission.ID,
			ExecutionID: rollbackExecution.ID, RuleSetHash: rollbackSubmission.Submission.RuleSetHash,
			RuleResults: []execution.RuleResult{}, Artifacts: rollbackArtifacts, CreatedAt: now.Add(5 * time.Second),
			Evidence: []Evidence{{
				ID: "00000000-0000-4000-8500-000000000902", EvaluationBatchID: badBatchID,
				LearnerID: learnerID, CapabilityID: "M1-01", CapabilityVersion: 2,
				CapabilityHash: capability.ContentHash, AttemptID: rollbackAttempt.ID,
				ActivityID: "assessment-check-config", ArtifactID: rollbackArtifacts[1].ID,
				EvidenceRuleID: "module-builds",
				EvidenceType:   "invalid", Result: execution.RulePassed,
				Independence: assistance.IndependenceIndependent, ContextLevel: "same_context",
				Evaluator: "deterministic", RuleVersion: 1, Reason: "invalid fixture",
				OccurredAt: now.Add(5 * time.Second), CreatedAt: now.Add(5 * time.Second),
			}},
		},
		AttemptID: rollbackAttempt.ID, LearnerID: learnerID, OccurredAt: now.Add(5 * time.Second),
	})
	if err == nil {
		t.Fatal("Persist(invalid evidence) error = nil")
	}
	var rolledBack int
	if err := db.QueryRowContext(ctx,
		"SELECT count(*) FROM \""+schema+"\".evaluation_batches WHERE id = $1", badBatchID,
	).Scan(&rolledBack); err != nil || rolledBack != 0 {
		t.Fatalf("rolled back batches = %d, error=%v", rolledBack, err)
	}
	stillExecuting, err := submissionService.Get(ctx, learnerID, rollbackSubmission.Submission.ID)
	if err != nil || stillExecuting.Status != submission.StatusExecuting {
		t.Fatalf("submission after rollback = %#v, %v", stillExecuting, err)
	}

	infraAttempt, err := attemptService.Create(ctx, attempt.CreateInput{
		LearnerID: learnerID, ActivityID: "assessment-check-config", ActivityVersion: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	infraSubmission, err := submissionService.Submit(ctx, submission.SubmitInput{
		LearnerID: learnerID, AttemptID: infraAttempt.ID, SubmissionKey: "infra-submit",
		WorkspaceRevision: infraAttempt.WorkspaceRevision, WorkspaceHash: infraAttempt.WorkspaceHash,
	})
	if err != nil {
		t.Fatal(err)
	}
	infraExecution, ok, err := executionRepository.Claim(ctx, "infra-worker", now.Add(6*time.Second), 10*time.Second, 2)
	if err != nil || !ok || infraExecution.ID != infraSubmission.ExecutionID {
		t.Fatalf("infra Claim() = %#v, %v, %v", infraExecution, ok, err)
	}
	if err := executionRepository.Complete(ctx, infraExecution.ID, "infra-worker", infrastructureFailureResponse(infraExecution), now.Add(7*time.Second)); err != nil {
		t.Fatal(err)
	}
	failedSubmission, err := submissionService.Get(ctx, learnerID, infraSubmission.Submission.ID)
	if err != nil || failedSubmission.Status != submission.StatusInfraFailed {
		t.Fatalf("infra submission = %#v, %v", failedSubmission, err)
	}
	failedAttempt, err := attemptService.Get(ctx, learnerID, infraAttempt.ID)
	if err != nil || failedAttempt.Status != "submit_infra_failed" {
		t.Fatalf("infra attempt = %#v, %v", failedAttempt, err)
	}
	var evaluationRequests, infraBatches, infraEvidence, infraArtifacts int
	for query, target := range map[string]*int{
		"SELECT count(*) FROM \"" + schema + "\".learning_outbox WHERE id = '" + infraExecution.ID + "' AND topic = 'submission.evaluate'": &evaluationRequests,
		"SELECT count(*) FROM \"" + schema + "\".evaluation_batches WHERE submission_id = '" + infraSubmission.Submission.ID + "'":         &infraBatches,
		"SELECT count(*) FROM \"" + schema + "\".evidence_records WHERE attempt_id = '" + infraAttempt.ID + "'":                            &infraEvidence,
		"SELECT count(*) FROM \"" + schema + "\".artifacts WHERE attempt_id = '" + infraAttempt.ID + "'":                                   &infraArtifacts,
	} {
		if err := db.QueryRowContext(ctx, query).Scan(target); err != nil {
			t.Fatal(err)
		}
	}
	if evaluationRequests != 0 || infraBatches != 0 || infraEvidence != 0 || infraArtifacts != 0 {
		t.Fatalf("infra side effects = requests %d batches %d evidence %d artifacts %d",
			evaluationRequests, infraBatches, infraEvidence, infraArtifacts)
	}
}

func TestCanonicalArtifactRejectsOversizedContent(t *testing.T) {
	_, err := canonicalArtifact(
		"00000000-0000-4000-8500-000000000001",
		"00000000-0000-4000-8500-000000000002",
		"00000000-0000-4000-8500-000000000003",
		"workspace", map[string]string{"content": strings.Repeat("x", maxArtifactContentBytes)}, time.Now(),
	)
	if err == nil || !strings.Contains(err.Error(), "limit is 4194304") {
		t.Fatalf("canonicalArtifact() error = %v", err)
	}
}

func persistedArtifactFixtures(t *testing.T, attemptID, submissionID string, createdAt time.Time) []Artifact {
	t.Helper()
	kinds := []string{"workspace", "diff", "explanation", "test_report"}
	artifacts := make([]Artifact, 0, len(kinds))
	for index, kind := range kinds {
		artifact, err := canonicalArtifact(
			fmt.Sprintf("00000000-0000-4000-8500-%012d", 903+index), attemptID, submissionID,
			kind, map[string]any{"kind": kind}, createdAt,
		)
		if err != nil {
			t.Fatal(err)
		}
		artifacts = append(artifacts, artifact)
	}
	return artifacts
}

func successfulAssessmentResponse(claimed execution.Execution) execution.ExecutionResponse {
	return execution.ExecutionResponse{
		ProtocolVersion: execution.ProtocolVersion, ExecutionID: claimed.ID,
		Status: execution.ExecutionSucceeded,
		Stages: []execution.StageResult{
			{Stage: execution.StageBuild, Status: execution.StagePassed},
			{Stage: execution.StageVet, Status: execution.StagePassed},
			{
				Stage: execution.StageVisibleTest, Status: execution.StagePassed,
				TestEvents: []execution.TestEvent{{
					Action: "pass", Package: "assessment/internal/config", Test: "TestNormalizeSortsTargets",
				}},
			},
			{
				Stage: execution.StageHeldOutTest, Status: execution.StagePassed,
				TestEvents: []execution.TestEvent{{
					Action: "pass", Package: "assessment/internal/config",
					Test: "TestLoadRejectsInvalidSchemeAndPreservesPathErrors",
				}},
			},
		},
		Policy: execution.PolicyReport{Network: execution.NetworkPolicyReport{
			Requested: claimed.Spec.Policy.Network, Enforcement: execution.EnforcementPolicyOnly,
		}},
	}
}

func infrastructureFailureResponse(claimed execution.Execution) execution.ExecutionResponse {
	return execution.ExecutionResponse{
		ProtocolVersion: execution.ProtocolVersion, ExecutionID: claimed.ID,
		Status: execution.ExecutionInfraFailed,
		Policy: execution.PolicyReport{Network: execution.NetworkPolicyReport{
			Requested: claimed.Spec.Policy.Network, Enforcement: execution.EnforcementPolicyOnly,
		}},
		Failure: &execution.Failure{Code: "sandbox_unavailable", Message: "sandbox unavailable"},
	}
}
