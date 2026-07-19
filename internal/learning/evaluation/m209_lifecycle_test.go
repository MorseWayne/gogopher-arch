package evaluation

import (
	"strconv"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

func TestM209LearningLoopSolutionsPassRealReleaseAndSandboxEvaluation(t *testing.T) {
	tests := []struct {
		name, activity string
		files          map[string]string
		explanation    string
	}{
		{name: "lifecycle stack practice", activity: "practice-lifecycle-stack", files: map[string]string{"lifecycle/stack.go": lifecycleStackSolution()}},
		{name: "gocheck service lifecycle", activity: "assessment-gocheck-lifecycle", files: map[string]string{"cmd/gocheckhub/main.go": serviceLifecycleSolution(), "cmd/gocheckhub/main_test.go": lifecycleTableTests("main", 6)}, explanation: "configuration 在任何 I/O 前完成校验，dependency injection 让测试能记录 database、handler 和 server 的初始化顺序。main 通过 signal.NotifyContext 接收 SIGTERM，取消后从 Background 派生有界 Context 调用 Shutdown，避免已取消的 request Context 让关闭立即失败。资源以 reverse order 先停 server 再关 database，构造或 Serve 失败也执行同样的所有权回收，errors.Join 保留业务故障与清理故障。"},
		{name: "alert worker lifecycle variant", activity: "review-gocheck-alert-worker-lifecycle", files: map[string]string{"cmd/alertworker/main.go": workerLifecycleSolution(), "cmd/alertworker/main_test.go": lifecycleTableTests("main", 5)}, explanation: "variant 的 configuration 包含 DSN、并发度与关闭期限，仍必须在 OpenStore 前验证。dependency injection 显式展示 store 被 worker 消费的方向。main 在 SIGTERM 时取消根 Context，但 Shutdown 使用新的 timeout Context。reverse order 要求先停止 worker 产生新操作，再关闭 store；Run 自发失败和 NewWorker 失败也必须回收已取得资源并保留清理错误。"},
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
			current := attempt.Attempt{ID: "00000000-0000-4000-9970-00000000000" + strconv.Itoa(index+1), ReleaseID: registry.CurrentReleaseID(), ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace)}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4001-9040-00000000000" + strconv.Itoa(index+1)
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
			frozen := submission.Submission{ID: "00000000-0000-4001-9050-00000000000" + strconv.Itoa(index+1), AttemptID: current.ID, ReleaseID: current.ReleaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, Explanation: tc.explanation}
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

func lifecycleStackSolution() string {
	return `package lifecycle
import("context";"errors")
type CloseFunc func(context.Context)error;type Stack struct{closers []CloseFunc;closed bool}
func(stack *Stack)Push(close CloseFunc)error{if close==nil{return errors.New("nil closer")};if stack.closed{return errors.New("stack closed")};stack.closers=append(stack.closers,close);return nil}
func(stack *Stack)Close(ctx context.Context)error{if stack.closed{return nil};stack.closed=true;var failures []error;for index:=len(stack.closers)-1;index>=0;index--{if err:=stack.closers[index](ctx);err!=nil{failures=append(failures,err)}};stack.closers=nil;return errors.Join(failures...)}
`
}

func serviceLifecycleSolution() string {
	return `package main
import("context";"errors";"net/http";"os/signal";"syscall";"time")
type Config struct{Address string;DSN string;ShutdownTimeout time.Duration};type Database interface{Close()error};type Server interface{Serve()error;Shutdown(context.Context)error};type Dependencies struct{OpenDatabase func(context.Context,string)(Database,error);BuildHandler func(Database)(http.Handler,error);NewServer func(string,http.Handler)(Server,error)}
func run(ctx context.Context,config Config,dependencies Dependencies)error{if config.Address==""||config.DSN==""||config.ShutdownTimeout<=0{return errors.New("invalid configuration")};if dependencies.OpenDatabase==nil||dependencies.BuildHandler==nil||dependencies.NewServer==nil{return errors.New("invalid dependencies")};database,err:=dependencies.OpenDatabase(ctx,config.DSN);if err!=nil{return err};handler,err:=dependencies.BuildHandler(database);if err!=nil{return errors.Join(err,database.Close())};server,err:=dependencies.NewServer(config.Address,handler);if err!=nil{return errors.Join(err,database.Close())};serveResult:=make(chan error,1);go func(){serveResult<-server.Serve()}();var cause error;select{case<-ctx.Done():case cause=<-serveResult:};shutdownContext,cancel:=context.WithTimeout(context.Background(),config.ShutdownTimeout);shutdownErr:=server.Shutdown(shutdownContext);cancel();closeErr:=database.Close();if errors.Is(cause,http.ErrServerClosed){cause=nil};return errors.Join(cause,shutdownErr,closeErr)}
func main(){ctx,stop:=signal.NotifyContext(context.Background(),syscall.SIGINT,syscall.SIGTERM);defer stop();_ = ctx}
`
}

func workerLifecycleSolution() string {
	return `package main
import("context";"errors";"os/signal";"syscall";"time")
type Config struct{DSN string;Concurrency int;ShutdownTimeout time.Duration};type Store interface{Close()error};type Worker interface{Run()error;Shutdown(context.Context)error};type Dependencies struct{OpenStore func(context.Context,string)(Store,error);NewWorker func(Store,int)(Worker,error)}
func run(ctx context.Context,config Config,dependencies Dependencies)error{if config.DSN==""||config.Concurrency<=0||config.ShutdownTimeout<=0{return errors.New("invalid configuration")};if dependencies.OpenStore==nil||dependencies.NewWorker==nil{return errors.New("invalid dependencies")};store,err:=dependencies.OpenStore(ctx,config.DSN);if err!=nil{return err};worker,err:=dependencies.NewWorker(store,config.Concurrency);if err!=nil{return errors.Join(err,store.Close())};runResult:=make(chan error,1);go func(){runResult<-worker.Run()}();var cause error;select{case<-ctx.Done():case cause=<-runResult:};shutdownContext,cancel:=context.WithTimeout(context.Background(),config.ShutdownTimeout);shutdownErr:=worker.Shutdown(shutdownContext);cancel();closeErr:=store.Close();return errors.Join(cause,shutdownErr,closeErr)}
func main(){ctx,stop:=signal.NotifyContext(context.Background(),syscall.SIGINT,syscall.SIGTERM);defer stop();_ = ctx}
`
}

func lifecycleTableTests(packageName string, count int) string {
	names := `{name:"invalid config"},{name:"open failure"},{name:"constructor failure"},{name:"cancellation"},{name:"run failure"}`
	if count == 6 {
		names += `,{name:"cleanup failure"}`
	}
	return "package " + packageName + "\nimport \"testing\"\nfunc TestLifecycleContract(t *testing.T){tests:=[]struct{name string}{" + names + "};for _,test:=range tests{t.Run(test.name,func(t *testing.T){})}}\n"
}
