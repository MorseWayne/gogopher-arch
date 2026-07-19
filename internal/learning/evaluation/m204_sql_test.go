package evaluation

import (
	"strconv"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

func TestM204LearningLoopSolutionsPassRealReleaseAndSandboxEvaluation(t *testing.T) {
	tests := []struct {
		name, activity string
		files          map[string]string
		explanation    string
	}{
		{name: "pool practice", activity: "practice-sql-pool", files: map[string]string{"sqlpool/pool.go": sqlPoolSolution()}},
		{name: "checks SQL", activity: "assessment-gocheck-sql", files: map[string]string{"internal/checks/postgres/repository.go": checksSQLSolution(), "internal/checks/postgres/repository_test.go": sqlTableTests("postgres")}, explanation: "connection pool 由长生命周期的 sql.DB 管理容量和连接寿命。每次 I/O 都沿用调用方 Context：多行读取用 QueryContext，成功后立即 defer Close，逐行 Scan，并在循环后检查 Rows.Err；单行查询直接 Scan，只映射 ErrNoRows。这样等待连接、driver 执行、取消和资源归还形成完整边界。"},
		{name: "alerts SQL variant", activity: "review-gocheck-alert-sql", files: map[string]string{"internal/alerts/sqlstore/repository.go": alertsSQLSolution(), "internal/alerts/sqlstore/repository_test.go": sqlTableTests("sqlstore")}, explanation: "connection pool 是共享的 sql.DB，而不是每次请求新建的连接。alert 列表通过 QueryContext 获得 Rows 后立刻安排 Close，每一行独立 Scan，结束后检查迭代错误；Find 在 Scan 时区分无记录。所有操作保留上游 Context，因此连接池等待和 driver 阻塞都能随取消结束。"},
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
			current := attempt.Attempt{ID: "00000000-0000-4000-9900-00000000000" + strconv.Itoa(index+1), ReleaseID: registry.CurrentReleaseID(), ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace)}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4000-9910-00000000000" + strconv.Itoa(index+1)
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
			frozen := submission.Submission{ID: "00000000-0000-4000-9920-00000000000" + strconv.Itoa(index+1), AttemptID: current.ID, ReleaseID: current.ReleaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, Explanation: tc.explanation}
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

func sqlPoolSolution() string {
	return `package sqlpool
import ("database/sql"; "errors"; "time")
type Config struct { MaxOpen int; MaxIdle int; MaxLifetime time.Duration; MaxIdleTime time.Duration }
func Configure(db *sql.DB,c Config) error { if db==nil||c.MaxOpen<1||c.MaxIdle<1||c.MaxIdle>c.MaxOpen||c.MaxLifetime<=0||c.MaxIdleTime<=0{return errors.New("invalid pool")}; db.SetMaxOpenConns(c.MaxOpen); db.SetMaxIdleConns(c.MaxIdle); db.SetConnMaxLifetime(c.MaxLifetime); db.SetConnMaxIdleTime(c.MaxIdleTime); return nil }
`
}
func checksSQLSolution() string {
	return `package postgres
import("context";"database/sql";"errors";"time";"gocheckhub/internal/checks")
type PoolConfig struct{MaxOpen,MaxIdle int; MaxLifetime,MaxIdleTime time.Duration}
func ConfigurePool(db *sql.DB,c PoolConfig)error{if db==nil||c.MaxOpen<1||c.MaxIdle<1||c.MaxIdle>c.MaxOpen||c.MaxLifetime<=0||c.MaxIdleTime<=0{return errors.New("invalid pool")};db.SetMaxOpenConns(c.MaxOpen);db.SetMaxIdleConns(c.MaxIdle);db.SetConnMaxLifetime(c.MaxLifetime);db.SetConnMaxIdleTime(c.MaxIdleTime);return nil}
type Repository struct{db *sql.DB}
func NewRepository(db *sql.DB)(*Repository,error){if db==nil{return nil,errors.New("db required")};return &Repository{db:db},nil}
func(r *Repository)Create(ctx context.Context,c checks.Check)error{_,err:=r.db.ExecContext(ctx,` + "`INSERT INTO checks (id,target) VALUES ($1,$2)`" + `,c.ID,c.Target);return err}
func(r *Repository)List(ctx context.Context)([]checks.Check,error){rows,err:=r.db.QueryContext(ctx,` + "`SELECT id,target FROM checks ORDER BY id`" + `);if err!=nil{return nil,err};defer rows.Close();var out []checks.Check;for rows.Next(){var c checks.Check;if err:=rows.Scan(&c.ID,&c.Target);err!=nil{return nil,err};out=append(out,c)};if err:=rows.Err();err!=nil{return nil,err};return out,nil}
func(r *Repository)Find(ctx context.Context,id string)(checks.Check,error){var c checks.Check;err:=r.db.QueryRowContext(ctx,` + "`SELECT id,target FROM checks WHERE id=$1`" + `,id).Scan(&c.ID,&c.Target);if errors.Is(err,sql.ErrNoRows){return checks.Check{},checks.ErrNotFound};if err!=nil{return checks.Check{},err};return c,nil}
`
}
func alertsSQLSolution() string {
	return `package sqlstore
import("context";"database/sql";"errors";"time";"gocheckhub/internal/alerts")
type PoolConfig struct{MaxOpen,MaxIdle int; MaxLifetime,MaxIdleTime time.Duration}
func ConfigurePool(db *sql.DB,c PoolConfig)error{if db==nil||c.MaxOpen<1||c.MaxIdle<1||c.MaxIdle>c.MaxOpen||c.MaxLifetime<=0||c.MaxIdleTime<=0{return errors.New("invalid pool")};db.SetMaxOpenConns(c.MaxOpen);db.SetMaxIdleConns(c.MaxIdle);db.SetConnMaxLifetime(c.MaxLifetime);db.SetConnMaxIdleTime(c.MaxIdleTime);return nil}
type Repository struct{db *sql.DB}
func NewRepository(db *sql.DB)(*Repository,error){if db==nil{return nil,errors.New("db required")};return &Repository{db:db},nil}
func(r *Repository)Save(ctx context.Context,c alerts.Rule)error{_,err:=r.db.ExecContext(ctx,` + "`INSERT INTO alerts (id,destination) VALUES ($1,$2)`" + `,c.ID,c.Destination);return err}
func(r *Repository)List(ctx context.Context)([]alerts.Rule,error){rows,err:=r.db.QueryContext(ctx,` + "`SELECT id,destination FROM alerts ORDER BY id`" + `);if err!=nil{return nil,err};defer rows.Close();var out []alerts.Rule;for rows.Next(){var c alerts.Rule;if err:=rows.Scan(&c.ID,&c.Destination);err!=nil{return nil,err};out=append(out,c)};if err:=rows.Err();err!=nil{return nil,err};return out,nil}
func(r *Repository)Find(ctx context.Context,id string)(alerts.Rule,error){var c alerts.Rule;err:=r.db.QueryRowContext(ctx,` + "`SELECT id,destination FROM alerts WHERE id=$1`" + `,id).Scan(&c.ID,&c.Destination);if errors.Is(err,sql.ErrNoRows){return alerts.Rule{},alerts.ErrNotFound};if err!=nil{return alerts.Rule{},err};return c,nil}
`
}
func sqlTableTests(pkg string) string {
	return "package " + pkg + ` 
import "testing"
func TestRepositoryContract(t *testing.T){tests:=[]struct{name string}{{name:"pool"},{name:"rows"},{name:"cancel"}};for _,tc:=range tests{t.Run(tc.name,func(t *testing.T){})}}
`
}
