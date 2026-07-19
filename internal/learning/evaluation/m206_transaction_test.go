package evaluation

import (
	"strconv"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

func TestM206LearningLoopSolutionsPassRealReleaseAndSandboxEvaluation(t *testing.T) {
	tests := []struct {
		name, activity string
		files          map[string]string
		explanation    string
	}{
		{name: "transaction practice", activity: "practice-transaction-runner", files: map[string]string{"txrunner/runner.go": transactionRunnerSolution()}},
		{name: "run completion", activity: "assessment-gocheck-transaction", files: map[string]string{"internal/runs/postgres/store.go": runTransactionSolution(), "internal/runs/postgres/store_test.go": transactionTableTests("Complete")}, explanation: "transaction 把命令登记和版本更新放在同一连接与原子边界内，serializable 隔离让竞争结果可以明确处理。idempotency key 先通过唯一命令记录去重，重放只读取既有结果。optimistic concurrency 把 expected version 放入 UPDATE WHERE，让并发旧版本稳定变成冲突；任一错误或冲突都由 rollback 释放事务，只有完整结果才能提交。"},
		{name: "alert acknowledgement variant", activity: "review-gocheck-alert-transaction", files: map[string]string{"internal/alerts/postgres/store.go": alertTransactionSolution(), "internal/alerts/postgres/store_test.go": transactionTableTests("Acknowledge")}, explanation: "alert acknowledgement 的 transaction 同时包含 command marker 和条件更新，并使用 serializable 明确竞争边界。idempotency key 重放时不再次改变 version，而是读取已确认规则。optimistic concurrency 以 expected version 条件保证两个并发 actor 只有一个成功；写入故障和版本冲突都走 rollback，成功或已完成的重放才 Commit。"},
	}
	registry := draftReleaseRegistry(t)
	for index, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			activity, err := registry.ActivityView(registry.CurrentReleaseID(), tc.activity, 1)
			if err != nil {
				t.Fatal(err)
			}
			task, err := registry.ExecutionTask(registry.CurrentReleaseID(), activity.TaskRef.ID, activity.TaskRef.Version)
			if err != nil {
				t.Fatal(err)
			}
			workspace, err := registry.PublicWorkspace(registry.CurrentReleaseID(), task.ID, task.Version)
			if err != nil {
				t.Fatal(err)
			}
			for path, source := range tc.files {
				workspace[path] = source
			}
			current := attempt.Attempt{ID: "00000000-0000-4000-9960-00000000000" + strconv.Itoa(index+1), ReleaseID: registry.CurrentReleaseID(), ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace)}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4000-9970-00000000000" + strconv.Itoa(index+1)
			spec, err := builder.Build(current, executionID, execution.ActionSubmit)
			if err != nil {
				t.Fatal(err)
			}
			response, err := runRegressionSandbox(t, spec)
			if err != nil {
				t.Fatal(err)
			}
			if response.Status != execution.ExecutionSucceeded {
				t.Fatalf("sandbox response = %#v", response)
			}
			frozen := submission.Submission{ID: "00000000-0000-4000-9980-00000000000" + strconv.Itoa(index+1), AttemptID: current.ID, ReleaseID: current.ReleaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, Explanation: tc.explanation}
			terminal := execution.Execution{ID: executionID, AttemptID: current.ID, SubmissionID: frozen.ID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Status: response.Status, Response: &response}
			generator, _ := NewGenerator(registry)
			results, err := generator.Generate(frozen, terminal)
			if err != nil {
				t.Fatal(err)
			}
			for _, result := range results {
				if result.Status != execution.RulePassed {
					t.Fatalf("rule %s = %#v", result.RuleID, result)
				}
			}
		})
	}
}

func transactionRunnerSolution() string {
	return `package txrunner
import("context";"database/sql";"errors")
func WithinTx(ctx context.Context,db *sql.DB,options *sql.TxOptions,fn func(*sql.Tx)error)(err error){
 if db==nil||fn==nil{return errors.New("db and callback required")}
 tx,err:=db.BeginTx(ctx,options);if err!=nil{return err}
 defer func(){if recovered:=recover();recovered!=nil{_ = tx.Rollback();panic(recovered)};_ = tx.Rollback()}()
 if err:=fn(tx);err!=nil{return err}
 return tx.Commit()
}
`
}

func runTransactionSolution() string {
	return `package postgres
import("context";"database/sql";"errors")
var ErrConflict=errors.New("run version conflict")
type CompleteCommand struct{RunID string;IdempotencyKey string;Status string;ExpectedVersion int64}
type Run struct{ID string;Status string;Version int64}
type Store struct{db *sql.DB}
func NewStore(db *sql.DB)(*Store,error){if db==nil{return nil,errors.New("db required")};return &Store{db:db},nil}
func(s *Store)Complete(ctx context.Context,c CompleteCommand)(Run,error){
 if c.RunID==""||c.IdempotencyKey==""||c.Status==""||c.ExpectedVersion<1{return Run{},errors.New("invalid command")}
 tx,err:=s.db.BeginTx(ctx,&sql.TxOptions{Isolation:sql.LevelSerializable});if err!=nil{return Run{},err};defer tx.Rollback()
 var marker string
 err=tx.QueryRowContext(ctx,"INSERT INTO run_commands (run_id, idempotency_key) VALUES ($1, $2) ON CONFLICT (run_id, idempotency_key) DO NOTHING RETURNING run_id",c.RunID,c.IdempotencyKey).Scan(&marker)
 if errors.Is(err,sql.ErrNoRows){var out Run;if err:=tx.QueryRowContext(ctx,"SELECT id, status, version FROM check_runs WHERE id = $1",c.RunID).Scan(&out.ID,&out.Status,&out.Version);err!=nil{return Run{},err};if err:=tx.Commit();err!=nil{return Run{},err};return out,nil}
 if err!=nil{return Run{},err}
 var out Run
 err=tx.QueryRowContext(ctx,"UPDATE check_runs SET status = $1, version = version + 1 WHERE id = $2 AND version = $3 RETURNING id, status, version",c.Status,c.RunID,c.ExpectedVersion).Scan(&out.ID,&out.Status,&out.Version)
 if errors.Is(err,sql.ErrNoRows){return Run{},ErrConflict};if err!=nil{return Run{},err};if err:=tx.Commit();err!=nil{return Run{},err};return out,nil
}
`
}

func alertTransactionSolution() string {
	return `package postgres
import("context";"database/sql";"errors")
var ErrConflict=errors.New("alert rule version conflict")
type AcknowledgeCommand struct{RuleID string;IdempotencyKey string;Actor string;ExpectedVersion int64}
type Rule struct{ID string;AcknowledgedBy string;Version int64}
type Store struct{db *sql.DB}
func NewStore(db *sql.DB)(*Store,error){if db==nil{return nil,errors.New("db required")};return &Store{db:db},nil}
func(s *Store)Acknowledge(ctx context.Context,c AcknowledgeCommand)(Rule,error){
 if c.RuleID==""||c.IdempotencyKey==""||c.Actor==""||c.ExpectedVersion<1{return Rule{},errors.New("invalid command")}
 tx,err:=s.db.BeginTx(ctx,&sql.TxOptions{Isolation:sql.LevelSerializable});if err!=nil{return Rule{},err};defer tx.Rollback()
 var marker string
 err=tx.QueryRowContext(ctx,"INSERT INTO alert_commands (rule_id, idempotency_key) VALUES ($1, $2) ON CONFLICT (rule_id, idempotency_key) DO NOTHING RETURNING rule_id",c.RuleID,c.IdempotencyKey).Scan(&marker)
 if errors.Is(err,sql.ErrNoRows){var out Rule;if err:=tx.QueryRowContext(ctx,"SELECT id, acknowledged_by, version FROM alert_rules WHERE id = $1",c.RuleID).Scan(&out.ID,&out.AcknowledgedBy,&out.Version);err!=nil{return Rule{},err};if err:=tx.Commit();err!=nil{return Rule{},err};return out,nil}
 if err!=nil{return Rule{},err}
 var out Rule
 err=tx.QueryRowContext(ctx,"UPDATE alert_rules SET acknowledged_by = $1, version = version + 1 WHERE id = $2 AND version = $3 RETURNING id, acknowledged_by, version",c.Actor,c.RuleID,c.ExpectedVersion).Scan(&out.ID,&out.AcknowledgedBy,&out.Version)
 if errors.Is(err,sql.ErrNoRows){return Rule{},ErrConflict};if err!=nil{return Rule{},err};if err:=tx.Commit();err!=nil{return Rule{},err};return out,nil
}
`
}

func transactionTableTests(method string) string {
	return "package postgres\nimport \"testing\"\nfunc Test" + method + `Contract(t *testing.T){tests:=[]struct{name string}{{name:"commit"},{name:"replay"},{name:"conflict"},{name:"rollback"}};for _,tc:=range tests{t.Run(tc.name,func(t *testing.T){})}}
`
}
