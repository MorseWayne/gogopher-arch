package attemptview

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	platformdb "github.com/MorseWayne/gogopher-arch/internal/platform/database"
)

func TestPostgresRepositoryLoadsOwnedAttemptRelatedState(t *testing.T) {
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
	schema := fmt.Sprintf("attempt_view_test_%d", time.Now().UnixNano())
	if _, err := db.ExecContext(ctx, `CREATE SCHEMA "`+schema+`"`); err != nil {
		t.Fatal(err)
	}
	defer func() {
		_, _ = db.ExecContext(context.Background(), `DROP SCHEMA "`+schema+`" CASCADE`)
		_ = db.Close()
	}()
	migrator, err := platformdb.NewMigrator(db, os.DirFS("../../../db/migrations"), platformdb.MigratorOptions{Schema: schema})
	if err != nil {
		t.Fatal(err)
	}
	if err := migrator.Up(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	hash := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	learnerID := "00000000-0000-4000-8000-000000000001"
	attemptID := "00000000-0000-4000-8000-000000000002"
	submissionID := "00000000-0000-4000-8000-000000000003"
	executionID := "00000000-0000-4000-8000-000000000004"
	batchID := "00000000-0000-4000-8000-000000000005"
	evidenceID := "00000000-0000-4000-8000-000000000006"
	assistanceID := "00000000-0000-4000-8000-000000000007"
	statements := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO definition_releases (release_id,schema_version,manifest,bundle_hash,status,created_at,published_at) VALUES ('release-v1',1,'{}',$1,'current',$2,$2)`, []any{hash, now}},
		{`INSERT INTO learners (id,created_at) VALUES ($1,$2)`, []any{learnerID, now}},
		{`INSERT INTO learning_attempts (id,learner_id,release_id,activity_id,activity_version,activity_hash,task_id,task_version,task_hash,capability_refs,mode,status,workspace,workspace_revision,workspace_hash,started_at,updated_at,submitted_at,completed_at) VALUES ($1,$2,'release-v1','activity',1,$3,'task-v1',1,$3,'[]','assessment','completed','{}',2,$3,$4,$4,$4,$4)`, []any{attemptID, learnerID, hash, now}},
		{`INSERT INTO assistance_events (id,attempt_id,learner_id,event_key,event_seq,event_type,payload,created_at) VALUES ($1,$2,$3,'hint:trace-contract',1,'hint_revealed','{"hint_id":"trace-contract"}',$4)`, []any{assistanceID, attemptID, learnerID, now}},
		{`INSERT INTO attempt_submissions (id,attempt_id,learner_id,submission_key,request_fingerprint,workspace,workspace_revision,workspace_hash,rule_set_hash,assistance_cutoff_seq,status,created_at,evaluated_at) VALUES ($1,$2,$3,'submit-1',$4,'{}',2,$4,$4,1,'evaluated',$5,$5)`, []any{submissionID, attemptID, learnerID, hash, now}},
		{`INSERT INTO attempt_executions (id,attempt_id,submission_id,action,sequence,request_key,request_fingerprint,release_id,task_id,task_version,task_hash,workspace_revision,workspace_hash,spec,status,result,claim_count,started_at,finished_at,created_at,updated_at) VALUES ($1,$2,$3,'submit',1,'submit:1',$4,'release-v1','task-v1',1,$4,2,$4,'{}','succeeded',$5,1,$6,$6,$6,$6)`, []any{executionID, attemptID, submissionID, hash, `{"protocol_version":1,"execution_id":"` + executionID + `","status":"succeeded","stages":[],"duration_ms":1,"policy":{"network":{"requested":"none","enforcement":"policy_only"}}}`, now}},
		{`INSERT INTO evaluation_batches (id,submission_id,execution_id,rule_set_hash,rule_results,created_at) VALUES ($1,$2,$3,$4,$5,$6)`, []any{batchID, submissionID, executionID, hash, `[{"rule_id":"held-out-tests-pass","status":"passed","stage":"held_out_test","test":"HiddenCase","summary":"passed","execution_id":"` + executionID + `"}]`, now}},
		{`INSERT INTO evidence_records (id,evaluation_batch_id,learner_id,capability_id,capability_version,capability_hash,attempt_id,activity_id,evidence_rule_id,evidence_type,result,independence,context_level,evaluator,rule_version,reason,occurred_at,created_at) VALUES ($1,$2,$3,'M1-09',1,$4,$5,'activity','held-out-tests-pass','test','passed','hinted','same_context','deterministic',1,'passed',$6,$6)`, []any{evidenceID, batchID, learnerID, hash, attemptID, now}},
	}
	fixtureTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer fixtureTx.Rollback()
	if _, err := fixtureTx.ExecContext(ctx, `SET LOCAL search_path TO "`+schema+`"`); err != nil {
		t.Fatal(err)
	}
	for _, statement := range statements {
		if _, err := fixtureTx.ExecContext(ctx, statement.query, statement.args...); err != nil {
			t.Fatal(err)
		}
	}
	if err := fixtureTx.Commit(); err != nil {
		t.Fatal(err)
	}
	repository, err := NewPostgresRepository(db, RepositoryOptions{Schema: schema})
	if err != nil {
		t.Fatal(err)
	}
	related, err := repository.Load(ctx, learnerID, attemptID)
	if err != nil {
		t.Fatal(err)
	}
	if related.Submission == nil || related.Submission.ID != submissionID || len(related.Executions) != 1 ||
		len(related.Assistance) != 1 || related.Assistance[0].EventKey != "hint:trace-contract" ||
		len(related.RuleResults) != 1 || len(related.Evidence) != 1 || related.Evidence[0].Independence != "hinted" {
		t.Fatalf("related = %#v", related)
	}
	other, err := repository.Load(ctx, "00000000-0000-4000-8000-000000000999", attemptID)
	if err != nil || other.Submission != nil || len(other.Assistance) != 0 || len(other.Executions) != 0 || len(other.Evidence) != 0 {
		t.Fatalf("cross-owner related = %#v, %v", other, err)
	}
}
