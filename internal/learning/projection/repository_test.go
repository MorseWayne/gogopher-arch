package projection

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
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
	assessmentActivity, _ := registry.ActivityView(registry.CurrentReleaseID(), "assessment-check-config", 2)
	reviewActivity, _ := registry.ActivityView(registry.CurrentReleaseID(), "review-check-config-variant", 2)
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
		{`INSERT INTO learning_attempts (id,learner_id,release_id,activity_id,activity_version,activity_hash,task_id,task_version,task_hash,capability_refs,mode,status,workspace,workspace_revision,workspace_hash,started_at,updated_at,submitted_at,completed_at) VALUES ($1,$2,$3,'assessment-check-config',2,$4,'assessment-check-config-v2',2,$5,'[]','assessment','completed','{}',0,$5,$6,$6,$6,$6)`, []any{attemptID, learnerID, registry.CurrentReleaseID(), assessmentActivity.ContentHash, hash64("c"), now}},
		{`INSERT INTO attempt_submissions (id,attempt_id,learner_id,submission_key,request_fingerprint,workspace,workspace_revision,workspace_hash,rule_set_hash,status,created_at,evaluated_at) VALUES ($1,$2,$3,'submit',$4,'{}',0,$4,$4,'evaluated',$5,$5)`, []any{submissionID, attemptID, learnerID, hash64("d"), now}},
		{`INSERT INTO attempt_executions (id,attempt_id,submission_id,action,sequence,request_key,request_fingerprint,release_id,task_id,task_version,task_hash,workspace_revision,workspace_hash,spec,status,result,finished_at,created_at,updated_at) VALUES ($1,$2,$3,'submit',0,'submit:0',$4,$5,'assessment-check-config-v2',2,$4,0,$4,'{}','succeeded','{}',$6,$6,$6)`, []any{executionID, attemptID, submissionID, hash64("e"), registry.CurrentReleaseID(), now}},
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

	reviewAttemptID := "00000000-0000-4000-8000-000000000109"
	reviewSubmissionID := "00000000-0000-4000-8000-000000000110"
	reviewExecutionID := "00000000-0000-4000-8000-000000000111"
	reviewBatchID := "00000000-0000-4000-8000-000000000112"
	reviewEvidenceID := "00000000-0000-4000-8000-000000000113"
	reviewStatements := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO "` + schema + `".learning_attempts (id,learner_id,release_id,activity_id,activity_version,activity_hash,task_id,task_version,task_hash,capability_refs,mode,status,workspace,workspace_revision,workspace_hash,started_at,updated_at,submitted_at,completed_at) VALUES ($1,$2,$3,'review-check-config-variant',2,$4,'review-check-config-variant-v2',2,$5,'[]','review','completed','{}',0,$5,$6,$6,$6,$6)`, []any{reviewAttemptID, learnerID, registry.CurrentReleaseID(), reviewActivity.ContentHash, hash64("b"), now}},
		{`INSERT INTO "` + schema + `".attempt_submissions (id,attempt_id,learner_id,submission_key,request_fingerprint,workspace,workspace_revision,workspace_hash,rule_set_hash,status,created_at,evaluated_at) VALUES ($1,$2,$3,'review-submit',$4,'{}',0,$4,$4,'evaluated',$5,$5)`, []any{reviewSubmissionID, reviewAttemptID, learnerID, hash64("c"), now}},
		{`INSERT INTO "` + schema + `".attempt_executions (id,attempt_id,submission_id,action,sequence,request_key,request_fingerprint,release_id,task_id,task_version,task_hash,workspace_revision,workspace_hash,spec,status,result,finished_at,created_at,updated_at) VALUES ($1,$2,$3,'submit',0,'submit:0',$4,$5,'review-check-config-variant-v2',2,$4,0,$4,'{}','user_failed','{}',$6,$6,$6)`, []any{reviewExecutionID, reviewAttemptID, reviewSubmissionID, hash64("d"), registry.CurrentReleaseID(), now}},
		{`INSERT INTO "` + schema + `".evaluation_batches (id,submission_id,execution_id,rule_set_hash,rule_results,created_at) VALUES ($1,$2,$3,$4,'[]',$5)`, []any{reviewBatchID, reviewSubmissionID, reviewExecutionID, hash64("e"), now}},
		{`INSERT INTO "` + schema + `".evidence_records (id,evaluation_batch_id,learner_id,capability_id,capability_version,capability_hash,attempt_id,activity_id,evidence_rule_id,evidence_type,result,independence,context_level,evaluator,rule_version,reason,occurred_at,created_at) VALUES ($1,$2,$3,'M1-03',1,$4,$5,'review-check-config-variant','error-chain-preserved','implement','failed','independent','variant','deterministic',1,'failed',$6,$6)`, []any{reviewEvidenceID, reviewBatchID, learnerID, policy.ContentHash, reviewAttemptID, now.Add(2 * time.Hour)}},
	}
	for _, statement := range reviewStatements {
		if _, err := db.ExecContext(ctx, statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	input.AsOf = now.Add(2 * time.Hour)
	unlinkedFailure, changed, err := projector.Rebuild(ctx, input)
	if err != nil || !changed || unlinkedFailure.RetentionState != RetentionStateFresh {
		t.Fatalf("unlinked review failure = %#v, changed=%v, error=%v", unlinkedFailure, changed, err)
	}
	if _, err := db.ExecContext(ctx, `UPDATE "`+schema+`".review_items SET status='completed',claimed_attempt_id=$2,evaluation_batch_id=$3,completed_at=$4,updated_at=$4 WHERE id=$1 AND status='open'`, openReviewID, reviewAttemptID, reviewBatchID, now.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	rustySnapshot, changed, err := projector.Rebuild(ctx, input)
	if err != nil || !changed || rustySnapshot.RetentionState != RetentionStateRusty || rustySnapshot.RetentionBase != RetentionRusty {
		t.Fatalf("rusty Rebuild() = %#v, changed=%v, error=%v", rustySnapshot, changed, err)
	}
	if _, err := db.ExecContext(ctx, `DELETE FROM "`+schema+`".capability_snapshots`); err != nil {
		t.Fatal(err)
	}
	if _, changed, err := projector.Rebuild(ctx, input); err != nil || !changed {
		t.Fatalf("rebuild after delete changed=%v error=%v", changed, err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".learning_outbox WHERE topic='review_scheduler.requested'`).Scan(&schedulerRequests); err != nil || schedulerRequests != 5 {
		t.Fatalf("scheduler requests after rebuild = %d, %v", schedulerRequests, err)
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
}

func hash64(character string) string {
	return strings.Repeat(character, 64)
}
