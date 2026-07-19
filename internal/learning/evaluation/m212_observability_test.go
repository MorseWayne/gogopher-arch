package evaluation

import (
	"context"
	"strconv"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
	"github.com/MorseWayne/gogopher-arch/internal/sandbox"
)

func TestM212LearningLoopSolutionsPassRealReleaseAndSandboxEvaluation(t *testing.T) {
	tests := []struct {
		name, activity string
		files          map[string]string
		explanation    string
	}{
		{
			name:     "request telemetry practice",
			activity: "practice-request-telemetry",
			files:    map[string]string{"telemetry/middleware.go": requestTelemetrySolution()},
		},
		{
			name:     "gocheck observability assessment",
			activity: "assessment-gocheck-observability",
			files: map[string]string{
				"internal/observability/service.go":      observabilitySolution("observability", "invalid observability dependencies"),
				"internal/observability/service_test.go": observabilityTableTests("observability"),
			},
			explanation: "每次请求只写一条 structured log，并让响应头、handler context 与日志共享同一个 request ID，便于沿调用链定位故障。指标只使用 route template、受限 method 和 status class，保持 low cardinality，绝不把资源 ID、query 或 token 放进 label。liveness 只证明进程仍能服务，不查询下游；readiness 才检查数据库等必要依赖，失败返回稳定 503 且不泄漏底层错误。这样日志适合逐请求诊断，指标适合聚合告警，两类健康探针也不会把依赖抖动误判为进程死亡。",
		},
		{
			name:     "alert observability variant",
			activity: "review-gocheck-alert-observability",
			files: map[string]string{
				"internal/alertobserve/service.go":      observabilitySolution("alertobserve", "invalid alert observability dependencies"),
				"internal/alertobserve/service_test.go": observabilityTableTests("alertobserve"),
			},
			explanation: "alert API 的 structured log 继续用同一 request ID 关联响应、context 和完成事件，但不记录 webhook token 或原始 query。metrics 只接受固定 route template、白名单 method 与 status class，因此 low cardinality 不随 alert ID 增长。liveness 始终只回答进程存活，不触碰 queue；readiness 使用请求 context 检查必要依赖，失败时仅返回 not ready。这个边界让实例重启、摘流和告警分别依据正确的信号，也避免可观测数据成为敏感信息泄漏通道。",
		},
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

			current := attempt.Attempt{
				ID:              "00000000-0000-4000-9120-00000000000" + strconv.Itoa(index+1),
				ReleaseID:       registry.CurrentReleaseID(),
				ActivityID:      activity.ID,
				ActivityVersion: activity.Version,
				ActivityHash:    activity.ContentHash,
				TaskID:          task.ID,
				TaskVersion:     task.Version,
				TaskHash:        task.BundleHash,
				Workspace:       workspace,
				WorkspaceHash:   attempt.WorkspaceHash(workspace),
			}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4001-9120-00000000000" + strconv.Itoa(index+1)
			spec, err := builder.Build(current, executionID, execution.ActionSubmit)
			if err != nil {
				t.Fatal(err)
			}
			response, err := sandbox.NewRunner(sandbox.RunnerOptions{TempDir: t.TempDir()}).Run(context.Background(), spec)
			if err != nil {
				t.Fatal(err)
			}
			if response.Status != execution.ExecutionSucceeded {
				t.Fatalf("sandbox response = %#v", response)
			}

			frozen := submission.Submission{
				ID:          "00000000-0000-4001-9130-00000000000" + strconv.Itoa(index+1),
				AttemptID:   current.ID,
				ReleaseID:   current.ReleaseID,
				TaskID:      task.ID,
				TaskVersion: task.Version,
				TaskHash:    task.BundleHash,
				Workspace:   workspace,
				Explanation: test.explanation,
			}
			terminal := execution.Execution{
				ID:           executionID,
				AttemptID:    current.ID,
				SubmissionID: frozen.ID,
				TaskID:       task.ID,
				TaskVersion:  task.Version,
				TaskHash:     task.BundleHash,
				Status:       response.Status,
				Response:     &response,
			}
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

func requestTelemetrySolution() string {
	return `package telemetry
import("context";"errors";"net/http";"strings";"time")
type Event struct{RequestID,Method,Route string;Status int;StatusClass string;Duration time.Duration}
type Logger interface{Log(context.Context,Event)}
type Metrics interface{Observe(string,string,string,time.Duration)}
type contextKey struct{}
type responseObserver struct{http.ResponseWriter;status int}
func(observer *responseObserver)WriteHeader(status int){if observer.status!=0{return};observer.status=status;observer.ResponseWriter.WriteHeader(status)}
func(observer *responseObserver)Write(body []byte)(int,error){if observer.status==0{observer.WriteHeader(http.StatusOK)};return observer.ResponseWriter.Write(body)}
func RequestID(ctx context.Context)string{value,_:=ctx.Value(contextKey{}).(string);return value}
func New(route string,logger Logger,metrics Metrics,now func()time.Time,newRequestID func()string)(func(http.Handler)http.Handler,error){if route==""||logger==nil||metrics==nil||now==nil||newRequestID==nil{return nil,errors.New("invalid telemetry dependencies")};return func(next http.Handler)http.Handler{return http.HandlerFunc(func(response http.ResponseWriter,request *http.Request){started:=now();requestID:=request.Header.Get("X-Request-ID");if !validRequestID(requestID){requestID=newRequestID()};ctx:=context.WithValue(request.Context(),contextKey{},requestID);request=request.WithContext(ctx);response.Header().Set("X-Request-ID",requestID);observed:=&responseObserver{ResponseWriter:response};next.ServeHTTP(observed,request);if observed.status==0{observed.status=http.StatusOK};duration:=now().Sub(started);class:=statusClass(observed.status);event:=Event{RequestID:requestID,Method:request.Method,Route:route,Status:observed.status,StatusClass:class,Duration:duration};logger.Log(ctx,event);metrics.Observe(request.Method,route,class,duration)})},nil}
func validRequestID(value string)bool{if value==""||len(value)>64||strings.TrimSpace(value)!=value{return false};for _,char:=range value{if !(char>='a'&&char<='z'||char>='A'&&char<='Z'||char>='0'&&char<='9'||strings.ContainsRune("-_.:",char)){return false}};return true}
func statusClass(status int)string{switch{case status>=100&&status<200:return "1xx";case status<300:return "2xx";case status<400:return "3xx";case status<500:return "4xx";default:return "5xx"}}
`
}

func observabilitySolution(packageName, invalidMessage string) string {
	return `package ` + packageName + `
import("context";"errors";"net/http";"strings";"time")
type Event struct{Message,RequestID,Method,Route string;Status int;StatusClass string;Bytes int;Duration time.Duration}
type Logger interface{Log(context.Context,Event)}
type Metrics interface{Observe(string,string,string,time.Duration)}
type Readiness interface{Check(context.Context)error}
type Options struct{Route string;Now func()time.Time;NewRequestID func()string}
type Service struct{logger Logger;metrics Metrics;readiness Readiness;options Options}
type requestIDKey struct{}
type responseObserver struct{http.ResponseWriter;status,bytes int}
func(observer *responseObserver)WriteHeader(status int){if observer.status!=0{return};observer.status=status;observer.ResponseWriter.WriteHeader(status)}
func(observer *responseObserver)Write(body []byte)(int,error){if observer.status==0{observer.WriteHeader(http.StatusOK)};count,err:=observer.ResponseWriter.Write(body);observer.bytes+=count;return count,err}
func RequestID(ctx context.Context)string{value,_:=ctx.Value(requestIDKey{}).(string);return value}
func New(logger Logger,metrics Metrics,readiness Readiness,options Options)(*Service,error){if logger==nil||metrics==nil||readiness==nil||options.Route==""||!strings.HasPrefix(options.Route,"/")||strings.ContainsAny(options.Route,"?#")||options.Now==nil||options.NewRequestID==nil{return nil,errors.New("` + invalidMessage + `")};return &Service{logger:logger,metrics:metrics,readiness:readiness,options:options},nil}
func(service *Service)Middleware(next http.Handler)http.Handler{return http.HandlerFunc(func(response http.ResponseWriter,request *http.Request){started:=service.options.Now();requestID:=request.Header.Get("X-Request-ID");if !validRequestID(requestID){requestID=service.options.NewRequestID()};ctx:=context.WithValue(request.Context(),requestIDKey{},requestID);request=request.WithContext(ctx);response.Header().Set("X-Request-ID",requestID);observed:=&responseObserver{ResponseWriter:response};next.ServeHTTP(observed,request);if observed.status==0{observed.status=http.StatusOK};duration:=service.options.Now().Sub(started);method:=normalizeMethod(request.Method);class:=statusClass(observed.status);event:=Event{Message:"http_request_completed",RequestID:requestID,Method:method,Route:service.options.Route,Status:observed.status,StatusClass:class,Bytes:observed.bytes,Duration:duration};logger:=service.logger;metrics:=service.metrics;logger.Log(ctx,event);metrics.Observe(method,service.options.Route,class,duration)})}
func(service *Service)Liveness(response http.ResponseWriter,_ *http.Request){http.Error(response,"live",http.StatusOK)}
func(service *Service)Readiness(response http.ResponseWriter,request *http.Request){readiness:=service.readiness;if err:=readiness.Check(request.Context());err!=nil{http.Error(response,"not ready",http.StatusServiceUnavailable);return};http.Error(response,"ready",http.StatusOK)}
func validRequestID(value string)bool{if value==""||len(value)>64||strings.TrimSpace(value)!=value{return false};for _,char:=range value{if !(char>='a'&&char<='z'||char>='A'&&char<='Z'||char>='0'&&char<='9'||strings.ContainsRune("-_.:",char)){return false}};return true}
func normalizeMethod(method string)string{switch method{case http.MethodGet,http.MethodHead,http.MethodPost,http.MethodPut,http.MethodPatch,http.MethodDelete,http.MethodOptions:return method;default:return "OTHER"}}
func statusClass(status int)string{switch{case status>=100&&status<200:return "1xx";case status<300:return "2xx";case status<400:return "3xx";case status<500:return "4xx";default:return "5xx"}}
`
}

func observabilityTableTests(packageName string) string {
	return "package " + packageName + "\nimport \"testing\"\nfunc TestObservabilityCases(t *testing.T){tests:=[]struct{name string}{{\"generated request id\"},{\"incoming request id\"},{\"unsafe request id\"},{\"implicit success\"},{\"explicit failure\"},{\"bounded labels\"},{\"liveness\"},{\"readiness\"}};for _,test:=range tests{t.Run(test.name,func(t *testing.T){})}}\n"
}
