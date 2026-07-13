package review

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	platformdb "github.com/MorseWayne/gogopher-arch/internal/platform/database"
)

func TestServiceClaimsReviewGroupOnceAcrossConcurrentItems(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	db, err := platformdb.Open(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	schema := fmt.Sprintf("review_test_%d", time.Now().UnixNano())
	if _, err := db.ExecContext(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
		t.Fatal(err)
	}
	defer db.ExecContext(context.Background(), `DROP SCHEMA "`+schema+`" CASCADE`)
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

	now := time.Date(2026, time.July, 13, 12, 0, 0, 0, time.UTC)
	learnerID := "00000000-0000-4000-8000-000000000101"
	otherLearnerID := "00000000-0000-4000-8000-000000000102"
	assessmentAttemptID := "00000000-0000-4000-8000-000000000103"
	submissionID := "00000000-0000-4000-8000-000000000104"
	executionID := "00000000-0000-4000-8000-000000000105"
	batchID := "00000000-0000-4000-8000-000000000106"
	releaseID := registry.CurrentReleaseID()
	assessment, _ := registry.ActivityView(releaseID, "assessment-check-config", 3)
	reviewActivity, _ := registry.ActivityView(releaseID, "review-check-config-variant", 2)
	hash := strings.Repeat("a", 64)

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SET LOCAL search_path TO "`+schema+`"`); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO learners (id,created_at) VALUES ($1,$3),($2,$3)`, learnerID, otherLearnerID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO learning_attempts (
			id,learner_id,release_id,activity_id,activity_version,activity_hash,
			task_id,task_version,task_hash,capability_refs,mode,status,workspace,
			workspace_revision,workspace_hash,started_at,updated_at,submitted_at,completed_at
		) VALUES ($1,$2,$3,$4,$5,$6,'assessment-check-config-v2',2,$7,'[]','assessment','completed','{}',0,$7,$8,$8,$8,$8)`,
		assessmentAttemptID, learnerID, releaseID, assessment.ID, assessment.Version, assessment.ContentHash, hash, now); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO attempt_submissions (
			id,attempt_id,learner_id,submission_key,request_fingerprint,workspace,
			workspace_revision,workspace_hash,rule_set_hash,status,created_at,evaluated_at
		) VALUES ($1,$2,$3,'submit',$4,'{}',0,$4,$4,'evaluated',$5,$5)`,
		submissionID, assessmentAttemptID, learnerID, hash, now); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO attempt_executions (
			id,attempt_id,submission_id,action,sequence,request_key,request_fingerprint,
			release_id,task_id,task_version,task_hash,workspace_revision,workspace_hash,
			spec,status,result,finished_at,created_at,updated_at
		) VALUES ($1,$2,$3,'submit',0,'submit:0',$4,$5,'assessment-check-config-v2',2,$4,0,$4,'{}','succeeded','{}',$6,$6,$6)`,
		executionID, assessmentAttemptID, submissionID, hash, releaseID, now); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO evaluation_batches (id,submission_id,execution_id,rule_set_hash,rule_results,created_at)
		VALUES ($1,$2,$3,$4,'[]',$5)`, batchID, submissionID, executionID, hash, now); err != nil {
		t.Fatal(err)
	}

	itemIDs := []string{
		"00000000-0000-4000-8000-000000000201",
		"00000000-0000-4000-8000-000000000202",
		"00000000-0000-4000-8000-000000000203",
		"00000000-0000-4000-8000-000000000204",
	}
	evidenceIDs := []string{
		"00000000-0000-4000-8000-000000000301",
		"00000000-0000-4000-8000-000000000302",
		"00000000-0000-4000-8000-000000000303",
		"00000000-0000-4000-8000-000000000304",
	}
	groupKey := "assessment:" + batchID + ":review:review-check-config-variant@2:policy:1"
	for index, ref := range reviewActivity.CapabilityRefs {
		policy, err := registry.CapabilityPolicy(releaseID, ref.ID, ref.Version)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO evidence_records (
				id,evaluation_batch_id,learner_id,capability_id,capability_version,capability_hash,
				attempt_id,activity_id,evidence_rule_id,evidence_type,result,independence,
				context_level,evaluator,rule_version,reason,occurred_at,created_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'implement','passed','independent','same_context','deterministic',1,'passed',$10,$10)`,
			evidenceIDs[index], batchID, learnerID, ref.ID, ref.Version, policy.ContentHash,
			assessmentAttemptID, assessment.ID, fmt.Sprintf("rule-%d", index), now); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO review_items (
				id,learner_id,capability_id,capability_version,source_evidence_id,
				release_id,activity_id,activity_version,activity_hash,review_group_key,
				due_at,priority,reason,status,policy_version,created_at,updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,100,'first_review','open',1,$12,$12)`,
			itemIDs[index], learnerID, ref.ID, ref.Version, evidenceIDs[index], releaseID,
			reviewActivity.ID, reviewActivity.Version, reviewActivity.ContentHash, groupKey, now.Add(72*time.Hour), now); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	earlyService, err := NewService(db, registry, ServiceOptions{Schema: schema, Now: func() time.Time { return now.Add(time.Hour) }})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := earlyService.Claim(ctx, learnerID, itemIDs[0]); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("early Claim() error = %v", err)
	}
	lifecycleAt := now.Add(2 * time.Hour)
	service, err := NewService(db, registry, ServiceOptions{Schema: schema, Now: func() time.Time { return lifecycleAt }})
	if err != nil {
		t.Fatal(err)
	}
	results := make(chan ClaimResult, 2)
	errorsFound := make(chan error, 2)
	var wait sync.WaitGroup
	for _, itemID := range itemIDs[:2] {
		wait.Add(1)
		go func() {
			defer wait.Done()
			result, err := service.ClaimAt(ctx, learnerID, itemID, now.Add(73*time.Hour))
			results <- result
			errorsFound <- err
		}()
	}
	wait.Wait()
	close(results)
	close(errorsFound)
	for err := range errorsFound {
		if err != nil {
			t.Fatalf("concurrent Claim() error = %v", err)
		}
	}
	var attemptID string
	created := 0
	for result := range results {
		if attemptID == "" {
			attemptID = result.Attempt.ID
		}
		if result.Attempt.ID != attemptID || result.Attempt.Mode != "review" {
			t.Fatalf("concurrent Claim() = %#v, attempt = %s", result, attemptID)
		}
		if !result.Attempt.StartedAt.Equal(lifecycleAt) {
			t.Fatalf("attempt started_at = %s, want lifecycle clock %s", result.Attempt.StartedAt, lifecycleAt)
		}
		if result.Created {
			created++
		}
	}
	if created != 1 {
		t.Fatalf("created results = %d, want 1", created)
	}

	replayed, err := service.Claim(ctx, learnerID, itemIDs[3])
	if err != nil || replayed.Created || replayed.Attempt.ID != attemptID {
		t.Fatalf("replayed Claim() = %#v, %v", replayed, err)
	}
	if _, err := service.Claim(ctx, otherLearnerID, itemIDs[0]); !errors.Is(err, ErrNotFound) {
		t.Fatalf("cross-owner Claim() error = %v", err)
	}
	var attempts, claimed, links, linkedAttempts int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".learning_attempts WHERE learner_id=$1 AND mode='review'`, learnerID).Scan(&attempts); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*),count(DISTINCT claimed_attempt_id) FROM "`+schema+`".review_items WHERE learner_id=$1 AND status='claimed'`, learnerID).Scan(&claimed, &linkedAttempts); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM "`+schema+`".attempt_review_items WHERE attempt_id=$1`, attemptID).Scan(&links); err != nil {
		t.Fatal(err)
	}
	if attempts != 1 || claimed != 4 || linkedAttempts != 1 || links != 4 {
		t.Fatalf("attempts=%d claimed=%d linked_attempts=%d links=%d", attempts, claimed, linkedAttempts, links)
	}
}
