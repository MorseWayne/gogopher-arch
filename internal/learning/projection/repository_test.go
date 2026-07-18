package projection

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/assistance"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/evaluation"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	reviewflow "github.com/MorseWayne/gogopher-arch/internal/learning/review"
	platformdb "github.com/MorseWayne/gogopher-arch/internal/platform/database"
)

func TestPostgresProjectorRebuildsFactsIdempotently(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	db, err := platformdb.Open(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("projection_test_%d", time.Now().UnixNano())
	if _, err := db.ExecContext(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA "`+schema+`" CASCADE`)
		_ = db.Close()
	}()
	migrator, _ := platformdb.NewMigrator(db, os.DirFS("../../../db/migrations"), platformdb.MigratorOptions{Schema: schema})
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
	policies := make(map[string]definition.CapabilityPolicyView)
	capabilityVersions := map[string]int{"M1-01": 2, "M1-03": 1, "M1-07": 1, "M1-09": 1}
	for _, capabilityID := range []string{"M1-01", "M1-03", "M1-07", "M1-09"} {
		policies[capabilityID], _ = registry.CapabilityPolicy(registry.CurrentReleaseID(), capabilityID, capabilityVersions[capabilityID])
	}
	policy := policies["M1-03"]
	assessmentActivity, _ := registry.ActivityView(registry.CurrentReleaseID(), "assessment-check-config", 5)
	reviewActivity, _ := registry.ActivityView(registry.CurrentReleaseID(), "review-check-config-variant", 4)
	now := time.Date(2026, time.July, 13, 12, 0, 0, 0, time.UTC)
	learnerID := "00000000-0000-4000-8000-000000000101"
	attemptID := "00000000-0000-4000-8000-000000000102"
	submissionID := "00000000-0000-4000-8000-000000000103"
	executionID := "00000000-0000-4000-8000-000000000104"
	batchID := "00000000-0000-4000-8000-000000000105"
	evidenceOne := "00000000-0000-4000-8000-000000000106"
	evidenceTwo := "00000000-0000-4000-8000-000000000107"
	m101EvidenceOne := "00000000-0000-4000-8000-000000000201"
	m101EvidenceTwo := "00000000-0000-4000-8000-000000000202"
	projectionRequestID := "00000000-0000-4000-8000-000000000115"
	tx, _ := db.BeginTx(ctx, nil)
	defer tx.Rollback()
	_, _ = tx.ExecContext(ctx, `SET LOCAL search_path TO "`+schema+`"`)
	statements := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO learners (id,created_at) VALUES ($1,$2)`, []any{learnerID, now}},
		{`INSERT INTO learning_attempts (id,learner_id,release_id,activity_id,activity_version,activity_hash,task_id,task_version,task_hash,capability_refs,mode,status,workspace,workspace_revision,workspace_hash,started_at,updated_at,submitted_at,completed_at) VALUES ($1,$2,$3,'assessment-check-config',5,$4,'assessment-check-config-v2',4,$5,'[]','assessment','completed','{}',0,$5,$6,$6,$6,$6)`, []any{attemptID, learnerID, registry.CurrentReleaseID(), assessmentActivity.ContentHash, hash64("c"), now}},
		{`INSERT INTO attempt_submissions (id,attempt_id,learner_id,submission_key,request_fingerprint,workspace,workspace_revision,workspace_hash,rule_set_hash,status,created_at,evaluated_at) VALUES ($1,$2,$3,'submit',$4,'{}',0,$4,$4,'evaluated',$5,$5)`, []any{submissionID, attemptID, learnerID, hash64("d"), now}},
		{`INSERT INTO attempt_executions (id,attempt_id,submission_id,action,sequence,request_key,request_fingerprint,release_id,task_id,task_version,task_hash,workspace_revision,workspace_hash,spec,status,result,finished_at,created_at,updated_at) VALUES ($1,$2,$3,'submit',0,'submit:0',$4,$5,'assessment-check-config-v2',4,$4,0,$4,'{}','succeeded','{}',$6,$6,$6)`, []any{executionID, attemptID, submissionID, hash64("e"), registry.CurrentReleaseID(), now}},
		{`INSERT INTO evaluation_batches (id,submission_id,execution_id,rule_set_hash,rule_results,created_at) VALUES ($1,$2,$3,$4,'[]',$5)`, []any{batchID, submissionID, executionID, hash64("f"), now}},
		{`INSERT INTO evidence_records (id,evaluation_batch_id,learner_id,capability_id,capability_version,capability_hash,attempt_id,activity_id,evidence_rule_id,evidence_type,result,independence,context_level,evaluator,rule_version,reason,occurred_at,created_at) VALUES ($1,$2,$3,'M1-03',1,$4,$5,'assessment-check-config',$6,'implement','passed','independent','same_context','deterministic',1,'passed',$7,$7)`, []any{evidenceOne, batchID, learnerID, policy.ContentHash, attemptID, "error-chain-preserved", now.Add(-time.Minute)}},
		{`INSERT INTO evidence_records (id,evaluation_batch_id,learner_id,capability_id,capability_version,capability_hash,attempt_id,activity_id,evidence_rule_id,evidence_type,result,independence,context_level,evaluator,rule_version,reason,occurred_at,created_at) VALUES ($1,$2,$3,'M1-03',1,$4,$5,'assessment-check-config',$6,'implement','passed','independent','same_context','deterministic',1,'passed',$7,$7)`, []any{evidenceTwo, batchID, learnerID, policy.ContentHash, attemptID, "resource-closed", now}},
		{`INSERT INTO learning_outbox (id,topic,aggregate_type,aggregate_id,idempotency_key,payload,status,available_at,created_at) VALUES ($1,'capability_projection.requested','evaluation_batch',$2,$3,$4::jsonb,'pending',$5,$5)`, []any{projectionRequestID, batchID, "capability-projection:" + batchID, fmt.Sprintf(`{"event_version":1,"evaluation_batch_id":%q,"learner_id":%q}`, batchID, learnerID), now}},
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	additionalEvidence := []struct {
		id, capabilityID, ruleID, evidenceType string
		capabilityVersion                      int
	}{
		{m101EvidenceOne, "M1-01", "module-builds", "implement", 2},
		{m101EvidenceTwo, "M1-01", "toolchain-checks-pass", "diagnose", 2},
		{"00000000-0000-4000-8000-000000000203", "M1-07", "invalid-input-rejected", "implement", 1},
		{"00000000-0000-4000-8000-000000000204", "M1-07", "stable-output", "implement", 1},
		{"00000000-0000-4000-8000-000000000205", "M1-07", "cli-failure-contract", "implement", 1},
		{"00000000-0000-4000-8000-000000000206", "M1-09", "learner-tests-present", "test", 1},
		{"00000000-0000-4000-8000-000000000207", "M1-09", "visible-tests-pass", "test", 1},
		{"00000000-0000-4000-8000-000000000208", "M1-09", "held-out-tests-pass", "test", 1},
	}
	for _, evidence := range additionalEvidence {
		if _, err := tx.ExecContext(ctx, `INSERT INTO evidence_records (id,evaluation_batch_id,learner_id,capability_id,capability_version,capability_hash,attempt_id,activity_id,evidence_rule_id,evidence_type,result,independence,context_level,evaluator,rule_version,reason,occurred_at,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,'assessment-check-config',$8,$9,'passed','independent','same_context','deterministic',1,'passed',$10,$10)`,
			evidence.id, batchID, learnerID, evidence.capabilityID, evidence.capabilityVersion,
			policies[evidence.capabilityID].ContentHash, attemptID, evidence.ruleID, evidence.evidenceType, now); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	projector, err := NewPostgresProjector(db, registry, RepositoryOptions{Schema: schema, Now: func() time.Time { return now.Add(time.Hour) }})
	if err != nil {
		t.Fatal(err)
	}
	requestRepository, _ := NewPostgresRequestRepository(db, RepositoryOptions{Schema: schema})
	if _, ok, err := requestRepository.ClaimRequest(ctx, "crashed-projector", now.Add(time.Minute), time.Minute); err != nil || !ok {
		t.Fatalf("crashed ClaimRequest() ok=%v error=%v", ok, err)
	}
	projectionClock := now.Add(3 * time.Minute)
	projectionWorker, _ := NewWorker(requestRepository, projector, WorkerOptions{
		Owner: "replacement-projector", Lease: time.Minute, PollInterval: time.Millisecond,
		MaxAttempts: 3, BaseBackoff: time.Second, MaxBackoff: time.Minute,
		Consumer: ProjectionConsumer, ConsumerVersion: ProjectionConsumerVersion,
		Now: func() time.Time { return projectionClock },
	})
	if processed, err := projectionWorker.RunOnce(ctx); err != nil || !processed {
		t.Fatalf("replacement RunOnce() processed=%v error=%v", processed, err)
	}
	input := RebuildInput{LearnerID: learnerID, ReleaseID: registry.CurrentReleaseID(), CapabilityID: "M1-03", CapabilityVersion: 1, AsOf: projectionClock}
	snapshot, changed, err := projector.Rebuild(ctx, input)
	if err != nil || changed || snapshot.AcquisitionState != AcquisitionVerified || snapshot.IndependenceState != IndependenceIndependent {
		t.Fatalf("worker Rebuild() = %#v, changed=%v, error=%v", snapshot, changed, err)
	}
	again, changed, err := projector.Rebuild(ctx, input)
	if err != nil || changed || !again.ProjectedAt.Equal(snapshot.ProjectedAt) {
		t.Fatalf("replayed Rebuild() = %#v, changed=%v, error=%v", again, changed, err)
	}
	var snapshots, schedulerRequests int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".capability_snapshots`).Scan(&snapshots); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".learning_outbox WHERE topic='review_scheduler.requested'`).Scan(&schedulerRequests); err != nil {
		t.Fatal(err)
	}
	if snapshots != 4 || schedulerRequests != 4 {
		t.Fatalf("counts = snapshots %d scheduler requests %d", snapshots, schedulerRequests)
	}
	var requestStatus, consumer string
	var consumerVersion, requestAttempts int
	if err := db.QueryRowContext(ctx, `SELECT status,consumer,consumer_version,attempt_count FROM "`+schema+`".learning_outbox WHERE id=$1`, projectionRequestID).Scan(
		&requestStatus, &consumer, &consumerVersion, &requestAttempts,
	); err != nil || requestStatus != "completed" || consumer != ProjectionConsumer || consumerVersion != ProjectionConsumerVersion || requestAttempts != 2 {
		t.Fatalf("recovered request = status %q consumer %q version %d attempts %d error %v", requestStatus, consumer, consumerVersion, requestAttempts, err)
	}
	scheduler, _ := NewReviewScheduler(db, registry, SchedulerOptions{Schema: schema, Now: func() time.Time { return now.Add(time.Hour) }})
	schedulerRepository, _ := NewPostgresReviewSchedulerRequestRepository(db, RepositoryOptions{Schema: schema})
	schedulerWorker, _ := NewWorker(schedulerRepository, scheduler, WorkerOptions{
		Owner: "review-scheduler", Lease: time.Minute, PollInterval: time.Millisecond,
		MaxAttempts: 3, BaseBackoff: time.Second, MaxBackoff: time.Minute,
		Consumer: ReviewSchedulerConsumer, ConsumerVersion: ReviewSchedulerConsumerVersion,
		Now: func() time.Time { return now.Add(time.Hour) },
	})
	var concurrentPayload []byte
	if err := db.QueryRowContext(ctx, `SELECT payload FROM "`+schema+`".learning_outbox WHERE topic='review_scheduler.requested' AND payload->>'capability_id'='M1-03' ORDER BY created_at,id LIMIT 1`).Scan(&concurrentPayload); err != nil {
		t.Fatal(err)
	}
	concurrentErrors := make(chan error, 2)
	for range 2 {
		go func() {
			concurrentErrors <- scheduler.ProcessRequest(ctx, Request{ID: "concurrent", Payload: concurrentPayload}, now.Add(time.Hour))
		}()
	}
	for range 2 {
		if err := <-concurrentErrors; err != nil {
			t.Fatalf("concurrent scheduler: %v", err)
		}
	}
	for range 4 {
		if processed, err := schedulerWorker.RunOnce(ctx); err != nil || !processed {
			t.Fatalf("scheduler RunOnce() processed=%v error=%v", processed, err)
		}
	}
	var activeReviews, allReviews, reviewGroups int
	var earliestDue, latestDue time.Time
	if err := db.QueryRowContext(ctx, `SELECT count(*),count(DISTINCT review_group_key),min(due_at),max(due_at) FROM "`+schema+`".review_items WHERE status='open'`).Scan(
		&activeReviews, &reviewGroups, &earliestDue, &latestDue,
	); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".review_items`).Scan(&allReviews); err != nil {
		t.Fatal(err)
	}
	if activeReviews != 4 || allReviews != 4 || reviewGroups != 1 ||
		!earliestDue.Equal(now.AddDate(0, 0, 3)) || !latestDue.Equal(earliestDue) {
		t.Fatalf("first reviews = active %d all %d groups %d due %s..%s", activeReviews, allReviews, reviewGroups, earliestDue, latestDue)
	}
	var openReviewID string
	if err := db.QueryRowContext(ctx, `SELECT id FROM "`+schema+`".review_items WHERE capability_id='M1-03' AND status='open'`).Scan(&openReviewID); err != nil {
		t.Fatal(err)
	}
	var m101SchedulerPayload []byte
	if err := db.QueryRowContext(ctx, `SELECT payload FROM "`+schema+`".learning_outbox WHERE topic='review_scheduler.requested' AND payload->>'capability_id'='M1-01' ORDER BY created_at,id LIMIT 1`).Scan(&m101SchedulerPayload); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE "`+schema+`".review_items SET source_evidence_id=$1 WHERE capability_id='M1-01' AND status='open'`, m101EvidenceOne); err != nil {
		t.Fatal(err)
	}
	if err := scheduler.ProcessRequest(ctx, Request{ID: "replacement", Payload: m101SchedulerPayload}, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := scheduler.ProcessRequest(ctx, Request{ID: "replay", Payload: m101SchedulerPayload}, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	var replacedReviews int
	if err := db.QueryRowContext(ctx, `SELECT count(*),count(*) FILTER (WHERE status='open'),count(*) FILTER (WHERE status='replaced') FROM "`+schema+`".review_items`).Scan(&allReviews, &activeReviews, &replacedReviews); err != nil {
		t.Fatal(err)
	}
	if allReviews != 5 || activeReviews != 4 || replacedReviews != 1 {
		t.Fatalf("review replacement = all %d active %d replaced %d", allReviews, activeReviews, replacedReviews)
	}
	projectionClock = now.Add(time.Hour)
	for range 5 {
		if processed, err := projectionWorker.RunOnce(ctx); err != nil || !processed {
			t.Fatalf("target projection RunOnce() processed=%v error=%v", processed, err)
		}
	}
	var snapshotsWithReview int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".capability_snapshots WHERE next_review_at=$1`, now.AddDate(0, 0, 3)).Scan(&snapshotsWithReview); err != nil || snapshotsWithReview != 4 {
		t.Fatalf("snapshots with next review = %d, %v", snapshotsWithReview, err)
	}
	dueInput := input
	dueInput.AsOf = now.AddDate(0, 0, 4)
	dueSnapshot, changed, err := projector.Rebuild(ctx, dueInput)
	if err != nil || changed || dueSnapshot.RetentionState != RetentionStateDue || dueSnapshot.NextReviewAt == nil {
		t.Fatalf("due Rebuild() = %#v, changed=%v, error=%v", dueSnapshot, changed, err)
	}
	reader, _ := NewReader(db, registry, ReaderOptions{Schema: schema})
	capabilityRead, err := reader.Capability(ctx, learnerID, CapabilitySelection{ID: "M1-03"}, dueInput.AsOf)
	if err != nil || capabilityRead.Snapshot == nil || capabilityRead.Snapshot.RetentionState != RetentionStateDue || len(capabilityRead.RecentEvidence) != 2 {
		t.Fatalf("Capability(M1-03) = %#v, %v", capabilityRead, err)
	}
	guidedActivity, _ := registry.ActivityView(registry.CurrentReleaseID(), "guided-run-model", 7)
	guidedTask, _ := registry.TaskView(registry.CurrentReleaseID(), guidedActivity.TaskRef.ID, guidedActivity.TaskRef.Version)
	openAttemptID := "00000000-0000-4000-8000-000000000116"
	if _, err := db.ExecContext(ctx, `INSERT INTO "`+schema+`".learning_attempts (
		id,learner_id,release_id,activity_id,activity_version,activity_hash,
		task_id,task_version,task_hash,capability_refs,mode,status,workspace,workspace_hash,started_at,updated_at
	) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'[]','guided','active','{}',$9,$10,$10)`,
		openAttemptID, learnerID, registry.CurrentReleaseID(), guidedActivity.ID, guidedActivity.Version,
		guidedActivity.ContentHash, guidedTask.ID, guidedTask.Version, guidedTask.BundleHash, dueInput.AsOf); err != nil {
		t.Fatal(err)
	}
	continuedRecommendation, err := reader.Next(ctx, learnerID, dueInput.AsOf)
	if err != nil || continuedRecommendation == nil || continuedRecommendation.Reason != "continue_attempt" ||
		continuedRecommendation.OpenAttempt == nil || continuedRecommendation.OpenAttempt.ID != openAttemptID ||
		continuedRecommendation.OpenAttempt.ReleaseID != registry.CurrentReleaseID() ||
		continuedRecommendation.Activity.ID != guidedActivity.ID {
		t.Fatalf("Next(open attempt) = %#v, %v", continuedRecommendation, err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM "`+schema+`".learning_attempts WHERE id=$1`, openAttemptID); err != nil {
		t.Fatal(err)
	}
	dueRecommendation, err := reader.Next(ctx, learnerID, dueInput.AsOf)
	if err != nil || dueRecommendation == nil || dueRecommendation.Kind != "review" || dueRecommendation.Reason != "due_review" || dueRecommendation.Activity.ID != reviewActivity.ID {
		t.Fatalf("Next(due review) = %#v, %v", dueRecommendation, err)
	}

	reviewService, _ := reviewflow.NewService(db, registry, reviewflow.ServiceOptions{
		Schema: schema, Now: func() time.Time { return now.AddDate(0, 0, 4).Add(-2 * time.Minute) },
	})
	claimedReview, err := reviewService.Claim(ctx, learnerID, openReviewID)
	if err != nil || !claimedReview.Created || claimedReview.Attempt.Mode != "review" {
		t.Fatalf("Claim(review) = %#v, %v", claimedReview, err)
	}
	reviewAttemptID := claimedReview.Attempt.ID
	claimedRecommendation, err := reader.Next(ctx, learnerID, dueInput.AsOf)
	if err != nil || claimedRecommendation == nil || claimedRecommendation.Reason != "claimed_review" || claimedRecommendation.ReviewItem == nil || claimedRecommendation.ReviewItem.ClaimedAttemptID != reviewAttemptID {
		t.Fatalf("Next(claimed review) = %#v, %v", claimedRecommendation, err)
	}
	reviewSubmissionID := "00000000-0000-4000-8000-000000000110"
	reviewExecutionID := "00000000-0000-4000-8000-000000000111"
	reviewBatchID := "00000000-0000-4000-8000-000000000112"
	reviewRequestID := "00000000-0000-4000-8000-000000000113"
	reviewFinishedAt := now.AddDate(0, 0, 4)
	reviewStatements := []struct {
		query string
		args  []any
	}{
		{`UPDATE "` + schema + `".learning_attempts SET status='submitted',submitted_at=$2,updated_at=$2 WHERE id=$1`, []any{reviewAttemptID, reviewFinishedAt.Add(-time.Minute)}},
		{`INSERT INTO "` + schema + `".attempt_submissions (id,attempt_id,learner_id,submission_key,request_fingerprint,workspace,workspace_revision,workspace_hash,rule_set_hash,status,created_at) VALUES ($1,$2,$3,'review-submit',$4,'{}',0,$4,$5,'executing',$6)`, []any{reviewSubmissionID, reviewAttemptID, learnerID, hash64("c"), reviewActivity.RuleSetHash, reviewFinishedAt.Add(-time.Minute)}},
		{`INSERT INTO "` + schema + `".attempt_executions (id,attempt_id,submission_id,action,sequence,request_key,request_fingerprint,release_id,task_id,task_version,task_hash,workspace_revision,workspace_hash,spec,status,result,finished_at,created_at,updated_at) VALUES ($1,$2,$3,'submit',0,'submit:0',$4,$5,'review-check-config-variant-v2',3,$6,0,$4,'{}','user_failed','{}',$7,$8,$7)`, []any{reviewExecutionID, reviewAttemptID, reviewSubmissionID, hash64("d"), registry.CurrentReleaseID(), claimedReview.Attempt.TaskHash, reviewFinishedAt, reviewFinishedAt.Add(-time.Minute)}},
	}
	for _, statement := range reviewStatements {
		if _, err := db.ExecContext(ctx, statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	ruleResults := []execution.RuleResult{
		{RuleID: "module-builds", Status: execution.RulePassed, Stage: execution.StageBuild, ExecutionID: reviewExecutionID},
		{RuleID: "error-chain-preserved", Status: execution.RuleFailed, Stage: execution.StageHeldOutTest, ExecutionID: reviewExecutionID},
		{RuleID: "invalid-input-rejected", Status: execution.RulePassed, Stage: execution.StageHeldOutTest, ExecutionID: reviewExecutionID},
		{RuleID: "stable-output", Status: execution.RuleNotEvaluated, Stage: execution.StageHeldOutTest, ExecutionID: reviewExecutionID},
		{RuleID: "learner-tests-present", Status: execution.RulePassed, Stage: execution.StageAST, ExecutionID: reviewExecutionID},
		{RuleID: "held-out-tests-pass", Status: execution.RulePassed, Stage: execution.StageHeldOutTest, ExecutionID: reviewExecutionID},
	}
	artifacts := evaluationArtifactFixtures(t, reviewAttemptID, reviewSubmissionID, reviewFinishedAt, 130)
	evidenceSpecs := []struct {
		id, capabilityID, ruleID, evidenceType string
		capabilityVersion                      int
		result                                 execution.RuleStatus
	}{
		{"00000000-0000-4000-8000-000000000121", "M1-01", "module-builds", "implement", 2, execution.RulePassed},
		{"00000000-0000-4000-8000-000000000122", "M1-03", "error-chain-preserved", "implement", 1, execution.RuleFailed},
		{"00000000-0000-4000-8000-000000000123", "M1-07", "invalid-input-rejected", "implement", 1, execution.RulePassed},
		{"00000000-0000-4000-8000-000000000124", "M1-09", "learner-tests-present", "test", 1, execution.RulePassed},
		{"00000000-0000-4000-8000-000000000125", "M1-09", "held-out-tests-pass", "test", 1, execution.RulePassed},
	}
	var reviewEvidence []evaluation.Evidence
	for _, spec := range evidenceSpecs {
		reviewEvidence = append(reviewEvidence, evaluation.Evidence{
			ID: spec.id, EvaluationBatchID: reviewBatchID, LearnerID: learnerID,
			CapabilityID: spec.capabilityID, CapabilityVersion: spec.capabilityVersion,
			CapabilityHash: policies[spec.capabilityID].ContentHash, AttemptID: reviewAttemptID,
			ActivityID: reviewActivity.ID, ArtifactID: artifacts[1].ID, EvidenceRuleID: spec.ruleID,
			EvidenceType: spec.evidenceType, Result: spec.result,
			Independence: assistance.IndependenceIndependent, ContextLevel: "variant",
			Evaluator: "deterministic", RuleVersion: 1, Reason: string(spec.result),
			OccurredAt: reviewFinishedAt, CreatedAt: reviewFinishedAt,
		})
	}
	evaluationRepository, _ := evaluation.NewPostgresRepository(db, evaluation.RepositoryOptions{Schema: schema})
	_, created, err := evaluationRepository.Persist(ctx, evaluation.PersistRecord{
		Batch: evaluation.Batch{
			ID: reviewBatchID, SubmissionID: reviewSubmissionID, ExecutionID: reviewExecutionID,
			RuleSetHash: reviewActivity.RuleSetHash, RuleResults: ruleResults,
			Artifacts: artifacts, Evidence: reviewEvidence, CreatedAt: reviewFinishedAt,
		},
		AttemptID: reviewAttemptID, LearnerID: learnerID,
		ReviewRequestID: reviewRequestID, OccurredAt: reviewFinishedAt,
	})
	if err != nil || !created {
		t.Fatalf("Persist(review outcome) created=%v error=%v", created, err)
	}
	outcomeClock := reviewFinishedAt.Add(time.Minute)
	outcomeScheduler, _ := NewReviewScheduler(db, registry, SchedulerOptions{Schema: schema, Now: func() time.Time { return outcomeClock }})
	outcomeWorker, _ := NewWorker(schedulerRepository, outcomeScheduler, WorkerOptions{
		Owner: "review-outcome-scheduler", Lease: time.Minute, PollInterval: time.Millisecond,
		MaxAttempts: 3, BaseBackoff: time.Second, MaxBackoff: time.Minute,
		Consumer: ReviewSchedulerConsumer, ConsumerVersion: ReviewSchedulerConsumerVersion,
		Now: func() time.Time { return outcomeClock },
	})
	for attempt := 0; attempt < 10; attempt++ {
		if processed, err := outcomeWorker.RunOnce(ctx); err != nil || !processed {
			t.Fatalf("review outcome RunOnce() processed=%v error=%v", processed, err)
		}
		var status string
		if err := db.QueryRowContext(ctx, `SELECT status FROM "`+schema+`".learning_outbox WHERE id=$1`, reviewRequestID).Scan(&status); err != nil {
			t.Fatal(err)
		}
		if status == "completed" {
			break
		}
		if attempt == 9 {
			t.Fatal("review outcome request was not processed")
		}
	}
	var completed, successors int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".review_items WHERE claimed_attempt_id=$1 AND status='completed'`, reviewAttemptID).Scan(&completed); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".review_items WHERE predecessor_review_item_id IS NOT NULL AND status='open'`).Scan(&successors); err != nil {
		t.Fatal(err)
	}
	if completed != 4 || successors != 4 {
		t.Fatalf("review transitions = completed %d successors %d", completed, successors)
	}
	transitionAssertions := []struct {
		capabilityID, outcome, reason, activityID string
		dueAt                                     time.Time
	}{
		{"M1-01", "passed", "maintenance", reviewActivity.ID, reviewFinishedAt.AddDate(0, 0, 14)},
		{"M1-03", "failed", "remediation", "practice-error-contract", reviewFinishedAt.AddDate(0, 0, 1)},
		{"M1-07", "incomplete", "review_incomplete", reviewActivity.ID, now.AddDate(0, 0, 3)},
		{"M1-09", "passed", "maintenance", reviewActivity.ID, reviewFinishedAt.AddDate(0, 0, 14)},
	}
	for _, assertion := range transitionAssertions {
		var outcome, reason, activityID string
		var dueAt time.Time
		if err := db.QueryRowContext(ctx, `
			SELECT previous.outcome,next.reason,next.activity_id,next.due_at
			FROM "`+schema+`".review_items previous
			JOIN "`+schema+`".review_items next ON next.predecessor_review_item_id=previous.id
			WHERE previous.claimed_attempt_id=$1 AND previous.capability_id=$2`,
			reviewAttemptID, assertion.capabilityID).Scan(&outcome, &reason, &activityID, &dueAt); err != nil {
			t.Fatal(err)
		}
		if outcome != assertion.outcome || reason != assertion.reason || activityID != assertion.activityID || !dueAt.Equal(assertion.dueAt) {
			t.Fatalf("%s transition = %s/%s/%s/%s", assertion.capabilityID, outcome, reason, activityID, dueAt)
		}
	}
	var outcomePayload []byte
	if err := db.QueryRowContext(ctx, `SELECT payload FROM "`+schema+`".learning_outbox WHERE id=$1 AND status='completed'`, reviewRequestID).Scan(&outcomePayload); err != nil {
		t.Fatal(err)
	}
	if err := outcomeScheduler.ProcessRequest(ctx, Request{ID: reviewRequestID, Payload: outcomePayload}, outcomeClock); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".review_items WHERE predecessor_review_item_id IS NOT NULL`).Scan(&successors); err != nil || successors != 4 {
		t.Fatalf("replayed successors = %d, %v", successors, err)
	}
	projectionClock = outcomeClock.Add(time.Minute)
	for range 5 {
		if processed, err := projectionWorker.RunOnce(ctx); err != nil || !processed {
			t.Fatalf("review projection RunOnce() processed=%v error=%v", processed, err)
		}
	}
	for _, assertion := range []struct {
		capabilityID string
		version      int
		acquisition  AcquisitionState
		retention    RetentionState
	}{
		{"M1-01", 2, AcquisitionStable, RetentionStateFresh},
		{"M1-03", 1, AcquisitionVerified, RetentionStateRusty},
		{"M1-07", 1, AcquisitionVerified, RetentionStateDue},
		{"M1-09", 1, AcquisitionStable, RetentionStateFresh},
	} {
		projected, _, err := projector.Rebuild(ctx, RebuildInput{
			LearnerID: learnerID, ReleaseID: registry.CurrentReleaseID(),
			CapabilityID: assertion.capabilityID, CapabilityVersion: assertion.version, AsOf: projectionClock,
		})
		if err != nil || projected.AcquisitionState != assertion.acquisition || projected.RetentionState != assertion.retention {
			t.Fatalf("%s review projection = %#v, %v", assertion.capabilityID, projected, err)
		}
	}
	var remediationItemID string
	if err := db.QueryRowContext(ctx, `SELECT id FROM "`+schema+`".review_items WHERE capability_id='M1-03' AND reason='remediation' AND status='open'`).Scan(&remediationItemID); err != nil {
		t.Fatal(err)
	}
	remediationFinishedAt := reviewFinishedAt.AddDate(0, 0, 1).Add(3 * time.Minute)
	remediationClaimService, _ := reviewflow.NewService(db, registry, reviewflow.ServiceOptions{
		Schema: schema, Now: func() time.Time { return remediationFinishedAt.Add(-2 * time.Minute) },
	})
	claimedRemediation, err := remediationClaimService.Claim(ctx, learnerID, remediationItemID)
	if err != nil || !claimedRemediation.Created || claimedRemediation.Attempt.Mode != "practice" {
		t.Fatalf("Claim(remediation) = %#v, %v", claimedRemediation, err)
	}
	remediationActivity, _ := registry.ActivityView(registry.CurrentReleaseID(), "practice-error-contract", 3)
	remediationSubmissionID := "00000000-0000-4000-8000-000000000140"
	remediationExecutionID := "00000000-0000-4000-8000-000000000141"
	remediationBatchID := "00000000-0000-4000-8000-000000000142"
	remediationRequestID := "00000000-0000-4000-8000-000000000143"
	remediationStatements := []struct {
		query string
		args  []any
	}{
		{`UPDATE "` + schema + `".learning_attempts SET status='submitted',submitted_at=$2,updated_at=$2 WHERE id=$1`, []any{claimedRemediation.Attempt.ID, remediationFinishedAt.Add(-time.Minute)}},
		{`INSERT INTO "` + schema + `".attempt_submissions (id,attempt_id,learner_id,submission_key,request_fingerprint,workspace,workspace_revision,workspace_hash,rule_set_hash,status,created_at) VALUES ($1,$2,$3,'remediation-submit',$4,'{}',0,$4,$5,'executing',$6)`, []any{remediationSubmissionID, claimedRemediation.Attempt.ID, learnerID, hash64("e"), remediationActivity.RuleSetHash, remediationFinishedAt.Add(-time.Minute)}},
		{`INSERT INTO "` + schema + `".attempt_executions (id,attempt_id,submission_id,action,sequence,request_key,request_fingerprint,release_id,task_id,task_version,task_hash,workspace_revision,workspace_hash,spec,status,result,finished_at,created_at,updated_at) VALUES ($1,$2,$3,'submit',0,'submit:0',$4,$5,'practice-error-contract-v1',2,$6,0,$4,'{}','succeeded','{}',$7,$8,$7)`, []any{remediationExecutionID, claimedRemediation.Attempt.ID, remediationSubmissionID, hash64("f"), registry.CurrentReleaseID(), claimedRemediation.Attempt.TaskHash, remediationFinishedAt, remediationFinishedAt.Add(-time.Minute)}},
	}
	for _, statement := range remediationStatements {
		if _, err := db.ExecContext(ctx, statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	remediationArtifacts := evaluationArtifactFixtures(t, claimedRemediation.Attempt.ID, remediationSubmissionID, remediationFinishedAt, 150)
	remediationRules := []execution.RuleResult{
		{RuleID: "error-chain-preserved", Status: execution.RulePassed, Stage: execution.StageVisibleTest, ExecutionID: remediationExecutionID},
		{RuleID: "resource-closed", Status: execution.RulePassed, Stage: execution.StageAST, ExecutionID: remediationExecutionID},
	}
	remediationEvidence := []evaluation.Evidence{
		{
			ID: "00000000-0000-4000-8000-000000000144", EvaluationBatchID: remediationBatchID,
			LearnerID: learnerID, CapabilityID: "M1-03", CapabilityVersion: 1,
			CapabilityHash: policies["M1-03"].ContentHash, AttemptID: claimedRemediation.Attempt.ID,
			ActivityID: remediationActivity.ID, ArtifactID: remediationArtifacts[1].ID,
			EvidenceRuleID: "error-chain-preserved", EvidenceType: "implement", Result: execution.RulePassed,
			Independence: assistance.IndependenceIndependent, ContextLevel: "same_context",
			Evaluator: "deterministic", RuleVersion: 1, Reason: "passed",
			OccurredAt: remediationFinishedAt, CreatedAt: remediationFinishedAt,
		},
		{
			ID: "00000000-0000-4000-8000-000000000145", EvaluationBatchID: remediationBatchID,
			LearnerID: learnerID, CapabilityID: "M1-03", CapabilityVersion: 1,
			CapabilityHash: policies["M1-03"].ContentHash, AttemptID: claimedRemediation.Attempt.ID,
			ActivityID: remediationActivity.ID, ArtifactID: remediationArtifacts[1].ID,
			EvidenceRuleID: "resource-closed", EvidenceType: "implement", Result: execution.RulePassed,
			Independence: assistance.IndependenceIndependent, ContextLevel: "same_context",
			Evaluator: "deterministic", RuleVersion: 1, Reason: "passed",
			OccurredAt: remediationFinishedAt, CreatedAt: remediationFinishedAt,
		},
	}
	_, created, err = evaluationRepository.Persist(ctx, evaluation.PersistRecord{
		Batch: evaluation.Batch{
			ID: remediationBatchID, SubmissionID: remediationSubmissionID, ExecutionID: remediationExecutionID,
			RuleSetHash: remediationActivity.RuleSetHash, RuleResults: remediationRules,
			Artifacts: remediationArtifacts, Evidence: remediationEvidence, CreatedAt: remediationFinishedAt,
		},
		AttemptID: claimedRemediation.Attempt.ID, LearnerID: learnerID,
		ReviewRequestID: remediationRequestID, OccurredAt: remediationFinishedAt,
	})
	if err != nil || !created {
		t.Fatalf("Persist(remediation outcome) created=%v error=%v", created, err)
	}
	outcomeClock = remediationFinishedAt.Add(time.Minute)
	for attempt := 0; attempt < 10; attempt++ {
		if processed, err := outcomeWorker.RunOnce(ctx); err != nil || !processed {
			t.Fatalf("remediation outcome RunOnce() processed=%v error=%v", processed, err)
		}
		var status string
		if err := db.QueryRowContext(ctx, `SELECT status FROM "`+schema+`".learning_outbox WHERE id=$1`, remediationRequestID).Scan(&status); err != nil {
			t.Fatal(err)
		}
		if status == "completed" {
			break
		}
		if attempt == 9 {
			t.Fatal("remediation outcome request was not processed")
		}
	}
	var remediationOutcome, followupReason, followupActivity string
	var followupDueAt time.Time
	if err := db.QueryRowContext(ctx, `
		SELECT previous.outcome,next.reason,next.activity_id,next.due_at
		FROM "`+schema+`".review_items previous
		JOIN "`+schema+`".review_items next ON next.predecessor_review_item_id=previous.id
		WHERE previous.id=$1`, remediationItemID).Scan(
		&remediationOutcome, &followupReason, &followupActivity, &followupDueAt,
	); err != nil {
		t.Fatal(err)
	}
	if remediationOutcome != "passed" || followupReason != "remediation_review" ||
		followupActivity != reviewActivity.ID || !followupDueAt.Equal(remediationFinishedAt.AddDate(0, 0, 3)) {
		t.Fatalf("remediation transition = %s/%s/%s/%s", remediationOutcome, followupReason, followupActivity, followupDueAt)
	}
	projectionClock = outcomeClock.Add(time.Minute)
	for range 2 {
		if processed, err := projectionWorker.RunOnce(ctx); err != nil || !processed {
			t.Fatalf("remediation projection RunOnce() processed=%v error=%v", processed, err)
		}
	}
	remediated, _, err := projector.Rebuild(ctx, RebuildInput{
		LearnerID: learnerID, ReleaseID: registry.CurrentReleaseID(),
		CapabilityID: "M1-03", CapabilityVersion: 1, AsOf: projectionClock,
	})
	if err != nil || remediated.AcquisitionState != AcquisitionVerified || remediated.RetentionState != RetentionStateRusty ||
		remediated.NextReviewAt == nil || !remediated.NextReviewAt.Equal(remediationFinishedAt.AddDate(0, 0, 3)) {
		t.Fatalf("remediated projection = %#v, %v", remediated, err)
	}
	input.AsOf = projectionClock
	var evidenceBeforePreview, outboxBeforePreview int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".evidence_records`).Scan(&evidenceBeforePreview); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".learning_outbox`).Scan(&outboxBeforePreview); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM "`+schema+`".capability_snapshots`); err != nil {
		t.Fatal(err)
	}
	preview, err := projector.Preview(ctx, input)
	if err != nil || preview.Change != "create" || preview.Before != nil {
		t.Fatalf("Preview(create) = %#v, %v", preview, err)
	}
	var evidenceAfterPreview, outboxAfterPreview, snapshotsAfterPreview int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".evidence_records`).Scan(&evidenceAfterPreview); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".learning_outbox`).Scan(&outboxAfterPreview); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".capability_snapshots`).Scan(&snapshotsAfterPreview); err != nil {
		t.Fatal(err)
	}
	if evidenceAfterPreview != evidenceBeforePreview || outboxAfterPreview != outboxBeforePreview || snapshotsAfterPreview != 0 {
		t.Fatalf("preview mutated state: evidence %d/%d outbox %d/%d snapshots %d", evidenceBeforePreview, evidenceAfterPreview, outboxBeforePreview, outboxAfterPreview, snapshotsAfterPreview)
	}
	if _, changed, err := projector.Rebuild(ctx, input); err != nil || !changed {
		t.Fatalf("rebuild after delete changed=%v error=%v", changed, err)
	}
	if repeated, err := projector.Preview(ctx, input); err != nil || repeated.Change != "none" {
		t.Fatalf("Preview(repeated) = %#v, %v", repeated, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".learning_outbox WHERE id=$1 AND topic='review_scheduler.requested' AND status='completed'`, reviewRequestID).Scan(&schedulerRequests); err != nil || schedulerRequests != 1 {
		t.Fatalf("completed review outcome requests = %d, %v", schedulerRequests, err)
	}

	poisonRequestID := "00000000-0000-4000-8000-000000000116"
	poisonAvailableAt := now.Add(2 * time.Hour)
	if _, err := db.ExecContext(ctx, `INSERT INTO "`+schema+`".learning_outbox (id,topic,aggregate_type,aggregate_id,idempotency_key,payload,status,available_at,created_at) VALUES ($1,'capability_projection.requested','evaluation_batch',$2,$3,'{"event_version":999}'::jsonb,'pending',$4,$4)`, poisonRequestID, batchID, "poison-projection", poisonAvailableAt); err != nil {
		t.Fatal(err)
	}
	poisonClock := poisonAvailableAt
	poisonWorker, _ := NewWorker(requestRepository, projector, WorkerOptions{
		Owner: "poison-projector", Lease: time.Minute, PollInterval: time.Millisecond,
		MaxAttempts: 2, BaseBackoff: time.Second, MaxBackoff: time.Second,
		Consumer: ProjectionConsumer, ConsumerVersion: ProjectionConsumerVersion,
		Now: func() time.Time { return poisonClock },
	})
	if processed, err := poisonWorker.RunOnce(ctx); err != nil || !processed {
		t.Fatalf("first poison RunOnce() processed=%v error=%v", processed, err)
	}
	poisonClock = poisonClock.Add(time.Second)
	if processed, err := poisonWorker.RunOnce(ctx); err != nil || !processed {
		t.Fatalf("second poison RunOnce() processed=%v error=%v", processed, err)
	}
	var poisonStatus, lastError string
	var poisonAttempts int
	var failedAt time.Time
	if err := db.QueryRowContext(ctx, `SELECT status,attempt_count,last_error,failed_at FROM "`+schema+`".learning_outbox WHERE id=$1`, poisonRequestID).Scan(
		&poisonStatus, &poisonAttempts, &lastError, &failedAt,
	); err != nil || poisonStatus != "failed" || poisonAttempts != 2 || lastError == "" || failedAt.IsZero() {
		t.Fatalf("poison request = status %q attempts %d last_error %q failed_at %s error %v", poisonStatus, poisonAttempts, lastError, failedAt, err)
	}
	assertOldReleaseReview(t, ctx, db, schema, registry, now)
}

func assertOldReleaseReview(t *testing.T, ctx context.Context, db *sql.DB, schema string, registry *definition.Registry, now time.Time) {
	t.Helper()
	const (
		learnerID    = "00000000-0000-4000-8000-000000008001"
		attemptID    = "00000000-0000-4000-8000-000000008002"
		submissionID = "00000000-0000-4000-8000-000000008003"
		executionID  = "00000000-0000-4000-8000-000000008004"
		batchID      = "00000000-0000-4000-8000-000000008005"
	)
	releaseID := "m1-first-slice-v1"
	activity, _ := registry.ActivityView(releaseID, "assessment-check-config", 1)
	task, _ := registry.TaskView(releaseID, activity.TaskRef.ID, activity.TaskRef.Version)
	policy, _ := registry.CapabilityPolicy(releaseID, "M1-03", 1)
	statements := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO "` + schema + `".learners (id,created_at) VALUES ($1,$2)`, []any{learnerID, now}},
		{`INSERT INTO "` + schema + `".learning_attempts (id,learner_id,release_id,activity_id,activity_version,activity_hash,task_id,task_version,task_hash,capability_refs,mode,status,workspace,workspace_revision,workspace_hash,started_at,updated_at,submitted_at,completed_at) VALUES ($1,$2,$3,$4,1,$5,$6,1,$7,$8::jsonb,'assessment','completed','{}',0,$9,$10,$10,$10,$10)`, []any{attemptID, learnerID, releaseID, activity.ID, activity.ContentHash, task.ID, task.BundleHash, `[ {"id":"M1-03","version":1} ]`, hash64("8"), now}},
		{`INSERT INTO "` + schema + `".attempt_submissions (id,attempt_id,learner_id,submission_key,request_fingerprint,workspace,workspace_revision,workspace_hash,rule_set_hash,status,created_at,evaluated_at) VALUES ($1,$2,$3,'old-submit',$4,'{}',0,$4,$5,'evaluated',$6,$6)`, []any{submissionID, attemptID, learnerID, hash64("9"), activity.RuleSetHash, now}},
		{`INSERT INTO "` + schema + `".attempt_executions (id,attempt_id,submission_id,action,sequence,request_key,request_fingerprint,release_id,task_id,task_version,task_hash,workspace_revision,workspace_hash,spec,status,result,finished_at,created_at,updated_at) VALUES ($1,$2,$3,'submit',0,'old-submit:0',$4,$5,$6,1,$7,0,$4,'{}','succeeded','{}',$8,$8,$8)`, []any{executionID, attemptID, submissionID, hash64("a"), releaseID, task.ID, task.BundleHash, now}},
		{`INSERT INTO "` + schema + `".evaluation_batches (id,submission_id,execution_id,rule_set_hash,rule_results,created_at) VALUES ($1,$2,$3,$4,'[]',$5)`, []any{batchID, submissionID, executionID, activity.RuleSetHash, now}},
		{`INSERT INTO "` + schema + `".capability_snapshots (learner_id,capability_id,capability_version,capability_hash,acquisition_state,independence_state,transfer_state,retention_base_state,last_evidence_at,last_independent_at,projection_version,projected_at) VALUES ($1,'M1-03',1,$2,'verified','independent','same_context','fresh',$3,$3,1,$3)`, []any{learnerID, policy.ContentHash, now}},
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	for index, ruleID := range []string{"error-chain-preserved", "resource-closed"} {
		id := fmt.Sprintf("00000000-0000-4000-8000-%012d", 8006+index)
		if _, err := db.ExecContext(ctx, `INSERT INTO "`+schema+`".evidence_records (id,evaluation_batch_id,learner_id,capability_id,capability_version,capability_hash,attempt_id,activity_id,evidence_rule_id,evidence_type,result,independence,context_level,evaluator,rule_version,reason,occurred_at,created_at) VALUES ($1,$2,$3,'M1-03',1,$4,$5,$6,$7,'implement','passed','independent','same_context','deterministic',1,'passed',$8,$8)`, id, batchID, learnerID, policy.ContentHash, attemptID, activity.ID, ruleID, now); err != nil {
			t.Fatal(err)
		}
	}
	scheduler, _ := NewReviewScheduler(db, registry, SchedulerOptions{Schema: schema, Now: func() time.Time { return now }})
	payload, _ := json.Marshal(ReviewSchedulerRequestPayload{
		EventVersion: ReviewSchedulerEventVersion, ProjectionVersion: ProjectionVersion,
		LearnerID: learnerID, ReleaseID: releaseID, CapabilityID: "M1-03", CapabilityVersion: 1,
		CapabilityHash: policy.ContentHash, AcquisitionState: AcquisitionVerified,
		IndependenceState: IndependenceIndependent, TransferState: TransferSameContext, RetentionBase: RetentionFresh,
	})
	if err := scheduler.ProcessRequest(ctx, Request{ID: "old-release", Payload: payload}, now); err != nil {
		t.Fatal(err)
	}
	var reviewItemID string
	if err := db.QueryRowContext(ctx, `SELECT id FROM "`+schema+`".review_items WHERE learner_id=$1 AND release_id=$2 AND status='open'`, learnerID, releaseID).Scan(&reviewItemID); err != nil {
		t.Fatal(err)
	}
	claimer, _ := reviewflow.NewService(db, registry, reviewflow.ServiceOptions{Schema: schema, Now: func() time.Time { return now.AddDate(0, 0, 4) }})
	claimed, err := claimer.Claim(ctx, learnerID, reviewItemID)
	if err != nil || !claimed.Created || claimed.Attempt.ReleaseID != releaseID || claimed.Attempt.ActivityVersion != 1 || claimed.Attempt.TaskVersion != 1 {
		t.Fatalf("old release Claim() = %#v, %v", claimed, err)
	}
}

func hash64(character string) string {
	return strings.Repeat(character, 64)
}

func evaluationArtifactFixtures(t *testing.T, attemptID, submissionID string, createdAt time.Time, idBase int) []evaluation.Artifact {
	t.Helper()
	var artifacts []evaluation.Artifact
	for index, kind := range []string{"workspace", "diff", "explanation", "test_report"} {
		content, err := definition.CanonicalJSON([]byte(fmt.Sprintf(`{"kind":%q}`, kind)))
		if err != nil {
			t.Fatal(err)
		}
		artifacts = append(artifacts, evaluation.Artifact{
			ID:        fmt.Sprintf("00000000-0000-4000-8000-%012d", idBase+index),
			AttemptID: attemptID, SubmissionID: submissionID, Kind: kind,
			Content: content, ContentBytes: len(content), ContentHash: definition.SHA256Hex(content), CreatedAt: createdAt,
		})
	}
	return artifacts
}
