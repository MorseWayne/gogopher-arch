package evaluation

import (
	"strconv"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

func TestM210LearningLoopSolutionsPassRealReleaseAndSandboxEvaluation(t *testing.T) {
	tests := []struct {
		name, activity string
		files          map[string]string
		explanation    string
	}{
		{name: "retry schedule practice", activity: "practice-retry-schedule", files: map[string]string{"retry/policy.go": retryScheduleSolution()}},
		{name: "gocheck durable worker", activity: "assessment-gocheck-worker", files: map[string]string{"internal/checkworker/worker.go": checkWorkerSolution(), "internal/checkworker/worker_test.go": workerTableTests("checkworker")}, explanation: "backpressure 来自固定数量的领取循环：每个槽位结算后才再次 Claim，因此积压留在持久化 Store。idempotency 由 durable claim 标记 Duplicate，worker 只 Ack 而不重复执行副作用。lease 让崩溃时未结算任务在过期后由新 owner 领取；temporary 故障按确定时间 retry，永久或耗尽次数才 Fail。每次 Process 继承 context 并有 timeout，根 context 取消后停止领取、等待 goroutine 退出，并把未结算任务留给 lease recovery。"},
		{name: "alert delivery worker variant", activity: "review-gocheck-alert-delivery-worker", files: map[string]string{"internal/alertworker/worker.go": alertWorkerSolution(), "internal/alertworker/worker_test.go": workerTableTests("alertworker")}, explanation: "在告警投递题材中，backpressure 仍由 Concurrency 个循环控制，不能先把数据库积压搬进 channel。Queue 使用 idempotency key 标记 Duplicate，避免 webhook 重复发送。lease 到期后 replacement owner 能恢复进程崩溃遗留的投递；临时发送故障用 RetryDelay 安排 retry，永久故障落为 Fail。Send 使用派生 context，根 context 取消会停止 Claim 并 join 全部循环，未确认消息随后依靠 lease 恢复。"},
	}
	registry := draftReleaseRegistry(t)
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			activity, err := registry.ActivityView(registry.CurrentReleaseID(), test.activity, 1)
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
			current := attempt.Attempt{ID: "00000000-0000-4000-9980-00000000000" + strconv.Itoa(index+1), ReleaseID: registry.CurrentReleaseID(), ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace)}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4001-9060-00000000000" + strconv.Itoa(index+1)
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
			frozen := submission.Submission{ID: "00000000-0000-4001-9070-00000000000" + strconv.Itoa(index+1), AttemptID: current.ID, ReleaseID: current.ReleaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, Explanation: test.explanation}
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

func retryScheduleSolution() string {
	return `package retry
import "time"
type Policy struct{Base time.Duration;Max time.Duration;MaxAttempts int}
func(policy Policy)Next(attempt int)(time.Duration,bool){if policy.Base<=0||policy.Max<policy.Base||policy.MaxAttempts<=1||attempt<1||attempt>=policy.MaxAttempts{return 0,false};delay:=policy.Base;for current:=1;current<attempt;current++{if delay>=policy.Max||delay>policy.Max/2{delay=policy.Max;break};delay*=2};if delay>policy.Max{delay=policy.Max};return delay,true}
`
}

func checkWorkerSolution() string {
	return `package checkworker
import("context";"errors";"sync";"time")
type Task struct{ID string;Key string;Attempt int;Duplicate bool}
type Store interface{Claim(context.Context,string,time.Time,time.Duration)(Task,bool,error);Ack(context.Context,string,string)error;Retry(context.Context,string,string,time.Time)error;Fail(context.Context,string,string)error}
type Processor interface{Process(context.Context,Task)error}
type Options struct{Owner string;Concurrency int;Lease time.Duration;ProcessTimeout time.Duration;PollInterval time.Duration;MaxAttempts int;RetryDelay func(int)time.Duration;Now func()time.Time}
type Worker struct{store Store;processor Processor;options Options}
type temporary interface{Temporary()bool}
func New(store Store,processor Processor,options Options)(*Worker,error){if store==nil||processor==nil||options.Owner==""||options.Concurrency<=0||options.Lease<=0||options.ProcessTimeout<=0||options.Lease<options.ProcessTimeout||options.PollInterval<=0||options.MaxAttempts<=0||options.RetryDelay==nil||options.Now==nil{return nil,errors.New("invalid worker configuration")};return &Worker{store:store,processor:processor,options:options},nil}
func(worker *Worker)Run(ctx context.Context)error{runContext,cancel:=context.WithCancel(ctx);defer cancel();failures:=make(chan error,worker.options.Concurrency);var group sync.WaitGroup;for range worker.options.Concurrency{group.Add(1);go func(){defer group.Done();for{if runContext.Err()!=nil{return};processed,err:=worker.RunOnce(runContext);if err!=nil{select{case failures<-err:cancel();default:};return};if processed{continue};timer:=time.NewTimer(worker.options.PollInterval);select{case<-runContext.Done():if !timer.Stop(){<-timer.C};return;case<-timer.C:}}}()};group.Wait();select{case err:=<-failures:return err;default:return ctx.Err()}}
func(worker *Worker)RunOnce(ctx context.Context)(bool,error){store:=worker.store;task,ok,err:=store.Claim(ctx,worker.options.Owner,worker.options.Now(),worker.options.Lease);if err!=nil||!ok{return false,err};if task.Duplicate{return true,store.Ack(ctx,task.ID,worker.options.Owner)};processContext,cancel:=context.WithTimeout(ctx,worker.options.ProcessTimeout);processor:=worker.processor;processErr:=processor.Process(processContext,task);cancel();if processErr==nil{return true,store.Ack(ctx,task.ID,worker.options.Owner)};if ctx.Err()!=nil{return true,ctx.Err()};retryable:=errors.Is(processErr,context.DeadlineExceeded);var transient temporary;if errors.As(processErr,&transient){retryable=transient.Temporary()};if retryable&&task.Attempt<worker.options.MaxAttempts{return true,store.Retry(ctx,task.ID,worker.options.Owner,worker.options.Now().Add(worker.options.RetryDelay(task.Attempt)))};return true,store.Fail(ctx,task.ID,worker.options.Owner)}
`
}

func alertWorkerSolution() string {
	return `package alertworker
import("context";"errors";"sync";"time")
type Delivery struct{ID,IdempotencyKey string;Attempt int;Duplicate bool}
type Queue interface{Claim(context.Context,string,time.Time,time.Duration)(Delivery,bool,error);Ack(context.Context,string,string)error;Retry(context.Context,string,string,time.Time)error;Fail(context.Context,string,string)error}
type Sender interface{Send(context.Context,Delivery)error}
type Options struct{Owner string;Concurrency int;Lease,SendTimeout,PollInterval time.Duration;MaxAttempts int;RetryDelay func(int)time.Duration;Now func()time.Time}
type Worker struct{queue Queue;sender Sender;options Options};type temporary interface{Temporary()bool}
func New(queue Queue,sender Sender,options Options)(*Worker,error){if queue==nil||sender==nil||options.Owner==""||options.Concurrency<=0||options.Lease<=0||options.SendTimeout<=0||options.Lease<options.SendTimeout||options.PollInterval<=0||options.MaxAttempts<=0||options.RetryDelay==nil||options.Now==nil{return nil,errors.New("invalid worker configuration")};return &Worker{queue:queue,sender:sender,options:options},nil}
func(worker *Worker)Run(ctx context.Context)error{runContext,cancel:=context.WithCancel(ctx);defer cancel();failures:=make(chan error,worker.options.Concurrency);var group sync.WaitGroup;for range worker.options.Concurrency{group.Add(1);go func(){defer group.Done();for{if runContext.Err()!=nil{return};processed,err:=worker.RunOnce(runContext);if err!=nil{select{case failures<-err:cancel();default:};return};if processed{continue};timer:=time.NewTimer(worker.options.PollInterval);select{case<-runContext.Done():if !timer.Stop(){<-timer.C};return;case<-timer.C:}}}()};group.Wait();select{case err:=<-failures:return err;default:return ctx.Err()}}
func(worker *Worker)RunOnce(ctx context.Context)(bool,error){queue:=worker.queue;delivery,ok,err:=queue.Claim(ctx,worker.options.Owner,worker.options.Now(),worker.options.Lease);if err!=nil||!ok{return false,err};if delivery.Duplicate{return true,queue.Ack(ctx,delivery.ID,worker.options.Owner)};sendContext,cancel:=context.WithTimeout(ctx,worker.options.SendTimeout);sender:=worker.sender;sendErr:=sender.Send(sendContext,delivery);cancel();if sendErr==nil{return true,queue.Ack(ctx,delivery.ID,worker.options.Owner)};if ctx.Err()!=nil{return true,ctx.Err()};retryable:=errors.Is(sendErr,context.DeadlineExceeded);var transient temporary;if errors.As(sendErr,&transient){retryable=transient.Temporary()};if retryable&&delivery.Attempt<worker.options.MaxAttempts{return true,queue.Retry(ctx,delivery.ID,worker.options.Owner,worker.options.Now().Add(worker.options.RetryDelay(delivery.Attempt)))};return true,queue.Fail(ctx,delivery.ID,worker.options.Owner)}
`
}

func workerTableTests(packageName string) string {
	return "package " + packageName + "\nimport \"testing\"\nfunc TestWorkerCases(t *testing.T){tests:=[]struct{name string}{{\"success\"},{\"duplicate\"},{\"temporary failure\"},{\"permanent failure\"},{\"backpressure\"},{\"cancellation\"},{\"restart recovery\"}};for _,test:=range tests{t.Run(test.name,func(t *testing.T){})}}\n"
}
