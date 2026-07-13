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
	policy, _ := registry.CapabilityPolicy(registry.CurrentReleaseID(), "M1-03", 1)
	now := time.Date(2026, time.July, 13, 12, 0, 0, 0, time.UTC)
	learnerID := "00000000-0000-4000-8000-000000000101"
	attemptID := "00000000-0000-4000-8000-000000000102"
	submissionID := "00000000-0000-4000-8000-000000000103"
	executionID := "00000000-0000-4000-8000-000000000104"
	batchID := "00000000-0000-4000-8000-000000000105"
	evidenceOne := "00000000-0000-4000-8000-000000000106"
	evidenceTwo := "00000000-0000-4000-8000-000000000107"
	tx, _ := db.BeginTx(ctx, nil)
	defer tx.Rollback()
	_, _ = tx.ExecContext(ctx, `SET LOCAL search_path TO "`+schema+`"`)
	statements := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO learners (id,created_at) VALUES ($1,$2)`, []any{learnerID, now}},
		{`INSERT INTO learning_attempts (id,learner_id,release_id,activity_id,activity_version,activity_hash,task_id,task_version,task_hash,capability_refs,mode,status,workspace,workspace_revision,workspace_hash,started_at,updated_at,submitted_at,completed_at) VALUES ($1,$2,$3,'assessment-check-config',1,$4,'assessment-check-config-v1',1,$5,'[]','assessment','completed','{}',0,$5,$6,$6,$6,$6)`, []any{attemptID, learnerID, registry.CurrentReleaseID(), hash64("b"), hash64("c"), now}},
		{`INSERT INTO attempt_submissions (id,attempt_id,learner_id,submission_key,request_fingerprint,workspace,workspace_revision,workspace_hash,rule_set_hash,status,created_at,evaluated_at) VALUES ($1,$2,$3,'submit',$4,'{}',0,$4,$4,'evaluated',$5,$5)`, []any{submissionID, attemptID, learnerID, hash64("d"), now}},
		{`INSERT INTO attempt_executions (id,attempt_id,submission_id,action,sequence,request_key,request_fingerprint,release_id,task_id,task_version,task_hash,workspace_revision,workspace_hash,spec,status,result,finished_at,created_at,updated_at) VALUES ($1,$2,$3,'submit',0,'submit:0',$4,$5,'assessment-check-config-v1',1,$4,0,$4,'{}','succeeded','{}',$6,$6,$6)`, []any{executionID, attemptID, submissionID, hash64("e"), registry.CurrentReleaseID(), now}},
		{`INSERT INTO evaluation_batches (id,submission_id,execution_id,rule_set_hash,rule_results,created_at) VALUES ($1,$2,$3,$4,'[]',$5)`, []any{batchID, submissionID, executionID, hash64("f"), now}},
		{`INSERT INTO evidence_records (id,evaluation_batch_id,learner_id,capability_id,capability_version,capability_hash,attempt_id,activity_id,evidence_rule_id,evidence_type,result,independence,context_level,evaluator,rule_version,reason,occurred_at,created_at) VALUES ($1,$2,$3,'M1-03',1,$4,$5,'assessment-check-config',$6,'implement','passed','independent','same_context','deterministic',1,'passed',$7,$7)`, []any{evidenceOne, batchID, learnerID, policy.ContentHash, attemptID, "error-chain-preserved", now.Add(-time.Minute)}},
		{`INSERT INTO evidence_records (id,evaluation_batch_id,learner_id,capability_id,capability_version,capability_hash,attempt_id,activity_id,evidence_rule_id,evidence_type,result,independence,context_level,evaluator,rule_version,reason,occurred_at,created_at) VALUES ($1,$2,$3,'M1-03',1,$4,$5,'assessment-check-config',$6,'implement','passed','independent','same_context','deterministic',1,'passed',$7,$7)`, []any{evidenceTwo, batchID, learnerID, policy.ContentHash, attemptID, "resource-closed", now}},
	}
	for _, statement := range statements {
		if _, err := tx.ExecContext(ctx, statement.query, statement.args...); err != nil {
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
	input := RebuildInput{LearnerID: learnerID, ReleaseID: registry.CurrentReleaseID(), CapabilityID: "M1-03", CapabilityVersion: 1, AsOf: now}
	snapshot, changed, err := projector.Rebuild(ctx, input)
	if err != nil || !changed || snapshot.AcquisitionState != AcquisitionVerified || snapshot.IndependenceState != IndependenceIndependent {
		t.Fatalf("first Rebuild() = %#v, changed=%v, error=%v", snapshot, changed, err)
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
	if snapshots != 1 || schedulerRequests != 1 {
		t.Fatalf("counts = snapshots %d scheduler requests %d", snapshots, schedulerRequests)
	}
	openReviewID := "00000000-0000-4000-8000-000000000108"
	if _, err := db.ExecContext(ctx, `
		INSERT INTO "`+schema+`".review_items (
			id,learner_id,capability_id,capability_version,source_evidence_id,release_id,
			activity_id,activity_version,activity_hash,review_group_key,due_at,priority,
			reason,status,policy_version,created_at,updated_at
		) VALUES ($1,$2,'M1-03',1,$3,$4,'review-check-config-variant',1,$5,'group-1',$6,10,'first_review','open',1,$7,$7)`,
		openReviewID, learnerID, evidenceOne, registry.CurrentReleaseID(), hash64("a"), now.Add(-time.Minute), now.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	dueSnapshot, changed, err := projector.Rebuild(ctx, input)
	if err != nil || !changed || dueSnapshot.RetentionState != RetentionStateDue || dueSnapshot.NextReviewAt == nil {
		t.Fatalf("due Rebuild() = %#v, changed=%v, error=%v", dueSnapshot, changed, err)
	}

	reviewAttemptID := "00000000-0000-4000-8000-000000000109"
	reviewSubmissionID := "00000000-0000-4000-8000-000000000110"
	reviewExecutionID := "00000000-0000-4000-8000-000000000111"
	reviewBatchID := "00000000-0000-4000-8000-000000000112"
	reviewEvidenceID := "00000000-0000-4000-8000-000000000113"
	completedReviewID := "00000000-0000-4000-8000-000000000114"
	reviewStatements := []struct {
		query string
		args  []any
	}{
		{`UPDATE "` + schema + `".review_items SET status='replaced', replaced_at=$2, updated_at=$2 WHERE id=$1`, []any{openReviewID, now}},
		{`INSERT INTO "` + schema + `".learning_attempts (id,learner_id,release_id,activity_id,activity_version,activity_hash,task_id,task_version,task_hash,capability_refs,mode,status,workspace,workspace_revision,workspace_hash,started_at,updated_at,submitted_at,completed_at) VALUES ($1,$2,$3,'review-check-config-variant',1,$4,'review-check-config-variant-v1',1,$5,'[]','review','completed','{}',0,$5,$6,$6,$6,$6)`, []any{reviewAttemptID, learnerID, registry.CurrentReleaseID(), hash64("a"), hash64("b"), now}},
		{`INSERT INTO "` + schema + `".attempt_submissions (id,attempt_id,learner_id,submission_key,request_fingerprint,workspace,workspace_revision,workspace_hash,rule_set_hash,status,created_at,evaluated_at) VALUES ($1,$2,$3,'review-submit',$4,'{}',0,$4,$4,'evaluated',$5,$5)`, []any{reviewSubmissionID, reviewAttemptID, learnerID, hash64("c"), now}},
		{`INSERT INTO "` + schema + `".attempt_executions (id,attempt_id,submission_id,action,sequence,request_key,request_fingerprint,release_id,task_id,task_version,task_hash,workspace_revision,workspace_hash,spec,status,result,finished_at,created_at,updated_at) VALUES ($1,$2,$3,'submit',0,'submit:0',$4,$5,'review-check-config-variant-v1',1,$4,0,$4,'{}','user_failed','{}',$6,$6,$6)`, []any{reviewExecutionID, reviewAttemptID, reviewSubmissionID, hash64("d"), registry.CurrentReleaseID(), now}},
		{`INSERT INTO "` + schema + `".evaluation_batches (id,submission_id,execution_id,rule_set_hash,rule_results,created_at) VALUES ($1,$2,$3,$4,'[]',$5)`, []any{reviewBatchID, reviewSubmissionID, reviewExecutionID, hash64("e"), now}},
		{`INSERT INTO "` + schema + `".evidence_records (id,evaluation_batch_id,learner_id,capability_id,capability_version,capability_hash,attempt_id,activity_id,evidence_rule_id,evidence_type,result,independence,context_level,evaluator,rule_version,reason,occurred_at,created_at) VALUES ($1,$2,$3,'M1-03',1,$4,$5,'review-check-config-variant','error-chain-preserved','implement','failed','independent','variant','deterministic',1,'failed',$6,$6)`, []any{reviewEvidenceID, reviewBatchID, learnerID, policy.ContentHash, reviewAttemptID, now}},
		{`INSERT INTO "` + schema + `".review_items (id,learner_id,capability_id,capability_version,source_evidence_id,release_id,activity_id,activity_version,activity_hash,review_group_key,due_at,priority,reason,status,claimed_attempt_id,evaluation_batch_id,policy_version,created_at,updated_at,completed_at) VALUES ($1,$2,'M1-03',1,$3,$4,'review-check-config-variant',1,$5,'group-1',$6,10,'first_review','completed',$7,$8,1,$6,$6,$6)`, []any{completedReviewID, learnerID, evidenceTwo, registry.CurrentReleaseID(), hash64("a"), now, reviewAttemptID, reviewBatchID}},
	}
	for _, statement := range reviewStatements[:len(reviewStatements)-1] {
		if _, err := db.ExecContext(ctx, statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	unlinkedFailure, changed, err := projector.Rebuild(ctx, input)
	if err != nil || !changed || unlinkedFailure.RetentionState != RetentionStateFresh {
		t.Fatalf("unlinked review failure = %#v, changed=%v, error=%v", unlinkedFailure, changed, err)
	}
	completedReview := reviewStatements[len(reviewStatements)-1]
	if _, err := db.ExecContext(ctx, completedReview.query, completedReview.args...); err != nil {
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
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".learning_outbox WHERE topic='review_scheduler.requested'`).Scan(&schedulerRequests); err != nil || schedulerRequests != 2 {
		t.Fatalf("scheduler requests after rebuild = %d, %v", schedulerRequests, err)
	}
}

func hash64(character string) string {
	return strings.Repeat(character, 64)
}
