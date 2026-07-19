package evaluation

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

func TestM213LearningLoopSolutionsPassRealReleaseAndSandboxEvaluation(t *testing.T) {
	tests := []struct {
		name, activity string
		version        int
		files          map[string]string
		explanation    string
	}{
		{"manual clock practice", "practice-manual-clock", 1, map[string]string{"testkit/clock.go": manualClockSolution()}, ""},
		{"gocheck layered tests", "assessment-gocheck-test-layers", 2, map[string]string{"internal/quality/service_test.go": qualityServiceTests(), "internal/quality/handler_test.go": qualityHandlerTests(), "internal/quality/postgres_integration_test.go": postgresIntegrationTest("quality")}, "unit test 用 deterministic clock、ID source 与 Store fake 验证业务决策，不需要 sleep。handler 层用 httptest 和 service fake 固定 JSON/status contract。testdata fixture 提供跨层场景，但真实 PostgreSQL integration test 仍必须通过 TEST_DATABASE_URL 连接数据库，执行 Ping、隔离 setup 与 Cleanup，才能覆盖驱动、schema 和 SQL，而不是让 unit test 依赖共享数据库。"},
		{"alert layered test variant", "review-gocheck-alert-test-layers", 2, map[string]string{"internal/alertquality/service_test.go": alertServiceTests(), "internal/alertquality/handler_test.go": alertHandlerTests(), "internal/alertquality/postgres_integration_test.go": postgresIntegrationTest("alertquality")}, "alert 的 unit test 同样注入 deterministic clock、ID 与 Store fake；HTTP contract 用 httptest 和 Creator fake，不启动 listener。alerts.json fixture 统一场景输入，真实 PostgreSQL integration test 只在 TEST_DATABASE_URL 已配置时运行，并负责 Ping、setup 与 Cleanup。这样快速测试保持确定性，数据库证据仍覆盖 driver、schema 和 SQL 的真实组合。"},
	}
	registry := draftReleaseRegistry(t)
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			activity, err := registry.ActivityView(registry.CurrentReleaseID(), test.activity, test.version)
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
			for path, source := range test.files {
				workspace[path] = source
			}
			current := attempt.Attempt{ID: "00000000-0000-4000-9140-00000000000" + strconv.Itoa(index+1), ReleaseID: registry.CurrentReleaseID(), ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace)}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4001-9140-00000000000" + strconv.Itoa(index+1)
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
			frozen := submission.Submission{ID: "00000000-0000-4001-9150-00000000000" + strconv.Itoa(index+1), AttemptID: current.ID, ReleaseID: current.ReleaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, Explanation: test.explanation}
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

func TestM213LayeredSolutionsPassRealPostgreSQLIntegration(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run the M2-13 PostgreSQL integration regression")
	}
	tests := []struct {
		activity string
		version  int
		files    map[string]string
	}{
		{"assessment-gocheck-test-layers", 2, map[string]string{"internal/quality/service_test.go": qualityServiceTests(), "internal/quality/handler_test.go": qualityHandlerTests(), "internal/quality/postgres_integration_test.go": postgresIntegrationTest("quality")}},
		{"review-gocheck-alert-test-layers", 2, map[string]string{"internal/alertquality/service_test.go": alertServiceTests(), "internal/alertquality/handler_test.go": alertHandlerTests(), "internal/alertquality/postgres_integration_test.go": postgresIntegrationTest("alertquality")}},
	}
	registry := draftReleaseRegistry(t)
	for _, test := range tests {
		t.Run(test.activity, func(t *testing.T) {
			activity, err := registry.ActivityView(registry.CurrentReleaseID(), test.activity, test.version)
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
			for path, source := range test.files {
				workspace[path] = source
			}
			workspaceRoot := t.TempDir()
			for path, source := range workspace {
				target := filepath.Join(workspaceRoot, filepath.FromSlash(path))
				if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(target, []byte(source), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			command := exec.Command("go", "test", "-tags=integration", "./...")
			command.Dir = workspaceRoot
			command.Env = append(os.Environ(), "GOWORK=off", "TEST_DATABASE_URL="+databaseURL)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("real PostgreSQL integration failed: %v\n%s", err, output)
			}
		})
	}
}

func manualClockSolution() string {
	return `package testkit
import("errors";"sync";"time")
type ManualClock struct{mu sync.RWMutex;now time.Time}
func NewManualClock(start time.Time)*ManualClock{return &ManualClock{now:start}}
func(clock *ManualClock)Now()time.Time{clock.mu.RLock();defer clock.mu.RUnlock();return clock.now}
func(clock *ManualClock)Advance(duration time.Duration)error{if duration<0{return errors.New("manual clock cannot move backwards")};clock.mu.Lock();defer clock.mu.Unlock();clock.now=clock.now.Add(duration);return nil}
`
}

func qualityServiceTests() string {
	return `package quality
import("context";"errors";"testing";"time")
type testClock struct{value time.Time};func(clock testClock)Now()time.Time{return clock.value};type testIDs struct{value string};func(ids testIDs)NewID()string{return ids.value};type testStore struct{saved []Check;err error};func(store *testStore)Save(_ context.Context,check Check)error{store.saved=append(store.saved,check);return store.err}
func TestServiceCases(t *testing.T){tests:=[]struct{name,inputName,target,id string;storeErr error;wantErr bool}{{"created","api","https://a","id-1",nil,false},{"trims"," api "," target ","id-2",nil,false},{"empty name","","target","id",nil,true},{"empty target","name","","id",nil,true},{"empty id","name","target","",nil,true},{"store failure","name","target","id",errors.New("store"),true}};for _,test:=range tests{t.Run(test.name,func(t *testing.T){clock:=testClock{time.Unix(10,0)};_ = clock.Now();store:=&testStore{err:test.storeErr};service,_:=NewService(store,clock,testIDs{test.id});_,err:=service.Create(context.Background(),test.inputName,test.target);if(err!=nil)!=test.wantErr{t.Fatalf("error=%v",err)}})}}
`
}

func alertServiceTests() string {
	return `package alertquality
import("context";"errors";"testing";"time")
type testClock struct{value time.Time};func(clock testClock)Now()time.Time{return clock.value};type testIDs struct{value string};func(ids testIDs)NewID()string{return ids.value};type testStore struct{saved []Alert;err error};func(store *testStore)Save(_ context.Context,alert Alert)error{store.saved=append(store.saved,alert);return store.err}
func TestServiceCases(t *testing.T){tests:=[]struct{name,inputName,destination,id string;storeErr error;wantErr bool}{{"created","latency","https://h","id-1",nil,false},{"trims"," latency "," hook ","id-2",nil,false},{"empty name","","hook","id",nil,true},{"empty destination","name","","id",nil,true},{"empty id","name","hook","",nil,true},{"store failure","name","hook","id",errors.New("store"),true}};for _,test:=range tests{t.Run(test.name,func(t *testing.T){clock:=testClock{time.Unix(10,0)};_ = clock.Now();store:=&testStore{err:test.storeErr};service,_:=NewService(store,clock,testIDs{test.id});_,err:=service.Create(context.Background(),test.inputName,test.destination);if(err!=nil)!=test.wantErr{t.Fatalf("error=%v",err)}})}}
`
}

func qualityHandlerTests() string {
	return `package quality
import("bytes";"context";"errors";"net/http";"net/http/httptest";"testing")
type handlerCreator struct{err error};func(fake handlerCreator)Create(context.Context,string,string)(Check,error){return Check{ID:"check-1"},fake.err}
func TestHandlerCases(t *testing.T){for _,test:=range[]struct{name,body string;err error;want int}{{"created",` + "`" + `{"name":"api","target":"https://a"}` + "`" + `,nil,201},{"bad json","{",nil,400},{"invalid",` + "`" + `{"name":"","target":"x"}` + "`" + `,ErrInvalid,400},{"internal",` + "`" + `{"name":"a","target":"x"}` + "`" + `,errors.New("db"),500}}{t.Run(test.name,func(t *testing.T){handler,_:=NewHandler(handlerCreator{test.err});request:=httptest.NewRequest(http.MethodPost,"/api/v1/checks",bytes.NewBufferString(test.body));response:=httptest.NewRecorder();handler.ServeHTTP(response,request);if response.Code!=test.want{t.Fatalf("status=%d",response.Code)}})}}
`
}

func alertHandlerTests() string {
	return `package alertquality
import("bytes";"context";"errors";"net/http";"net/http/httptest";"testing")
type handlerCreator struct{err error};func(fake handlerCreator)Create(context.Context,string,string)(Alert,error){return Alert{ID:"alert-1"},fake.err}
func TestHandlerCases(t *testing.T){for _,test:=range[]struct{name,body string;err error;want int}{{"created",` + "`" + `{"name":"latency","destination":"https://h"}` + "`" + `,nil,201},{"bad json","{",nil,400},{"invalid",` + "`" + `{"name":"","destination":"x"}` + "`" + `,ErrInvalid,400},{"internal",` + "`" + `{"name":"a","destination":"x"}` + "`" + `,errors.New("db"),500}}{t.Run(test.name,func(t *testing.T){handler,_:=NewHandler(handlerCreator{test.err});request:=httptest.NewRequest(http.MethodPost,"/api/v1/alerts",bytes.NewBufferString(test.body));response:=httptest.NewRecorder();handler.ServeHTTP(response,request);if response.Code!=test.want{t.Fatalf("status=%d",response.Code)}})}}
`
}

func postgresIntegrationTest(packageName string) string {
	return `package ` + packageName + `
import("context";"database/sql";"os";"testing";"time")
func TestPostgreSQLIntegration(t *testing.T){dsn:=os.Getenv("TEST_DATABASE_URL");if dsn==""{t.Skip("TEST_DATABASE_URL is not configured")};db,err:=sql.Open("pgx",dsn);if err!=nil{t.Fatal(err)};t.Cleanup(func(){_ = db.Close()});ctx,cancel:=context.WithTimeout(context.Background(),5*time.Second);defer cancel();if err:=db.PingContext(ctx);err!=nil{t.Fatal(err)};if _,err:=db.ExecContext(ctx,"CREATE TEMP TABLE gogopher_fixture(id text primary key)");err!=nil{t.Fatal(err)}}
`
}
