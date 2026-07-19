package evaluation

import (
	"strconv"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

func TestM207LearningLoopSolutionsPassRealReleaseAndSandboxEvaluation(t *testing.T) {
	tests := []struct {
		name, activity string
		files          map[string]string
		explanation    string
	}{
		{name: "client policy practice", activity: "practice-http-client-policy", files: map[string]string{"clientpolicy/client.go": httpClientPolicySolution()}},
		{name: "probe client", activity: "assessment-gocheck-http-client", files: map[string]string{"internal/probe/httpclient/client.go": probeClientSolution(), "internal/probe/httpclient/client_test.go": httpClientTableTests("Probe")}, explanation: "受控 http.Client 提供整体 timeout，而每个 request 继续携带调用方 context，使取消能更早终止 RoundTrip。Do 成功后立即安排 Body Close，并用读取上限防止无界响应。429 与 5xx 只映射为稳定错误；retry 必须由拥有预算、退避和幂等信息的上层决定，client 单次调用不自动重试。"},
		{name: "alert delivery variant", activity: "review-gocheck-alert-delivery", files: map[string]string{"internal/alerts/webhook/client.go": deliveryClientSolution(), "internal/alerts/webhook/client_test.go": httpClientTableTests("Deliver")}, explanation: "delivery 使用独立 http.Client 与调用方 context，POST 发出后 transport 错误不能证明对方未接收。只要收到 response 就 defer Body Close，并限制响应大小。429 交给调度层读取策略后处理；5xx 也不在 client 内 retry，未来重试必须结合 delivery idempotency key、预算与退避。"},
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
			current := attempt.Attempt{ID: "00000000-0000-4000-9990-00000000000" + strconv.Itoa(index+1), ReleaseID: registry.CurrentReleaseID(), ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace)}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4001-9000-00000000000" + strconv.Itoa(index+1)
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
			frozen := submission.Submission{ID: "00000000-0000-4001-9010-00000000000" + strconv.Itoa(index+1), AttemptID: current.ID, ReleaseID: current.ReleaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, Explanation: tc.explanation}
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

func httpClientPolicySolution() string {
	return `package clientpolicy
import("errors";"net/http";"time")
type Config struct{Timeout time.Duration;MaxIdleConns int;MaxIdleConnsPerHost int;IdleConnTimeout time.Duration;TLSHandshakeTimeout time.Duration;ResponseHeaderTimeout time.Duration}
func NewClient(c Config)(*http.Client,error){if c.Timeout<=0||c.MaxIdleConns<=0||c.MaxIdleConnsPerHost<=0||c.IdleConnTimeout<=0||c.TLSHandshakeTimeout<=0||c.ResponseHeaderTimeout<=0{return nil,errors.New("invalid policy")};transport:=http.DefaultTransport.(*http.Transport).Clone();transport.MaxIdleConns=c.MaxIdleConns;transport.MaxIdleConnsPerHost=c.MaxIdleConnsPerHost;transport.IdleConnTimeout=c.IdleConnTimeout;transport.TLSHandshakeTimeout=c.TLSHandshakeTimeout;transport.ResponseHeaderTimeout=c.ResponseHeaderTimeout;return &http.Client{Transport:transport,Timeout:c.Timeout},nil}
`
}

func probeClientSolution() string {
	return `package httpclient
import("context";"encoding/json";"errors";"io";"net/http";"net/url";"strings")
var(ErrRateLimited=errors.New("probe rate limited");ErrRejected=errors.New("probe request rejected");ErrUpstream=errors.New("probe upstream failed");ErrBodyTooLarge=errors.New("probe response too large"))
type Result struct{Status string ` + "`json:\"status\"`" + `};type Client struct{client *http.Client;baseURL string;maxBody int64}
func New(client *http.Client,baseURL string,maxBody int64)(*Client,error){parsed,err:=url.Parse(baseURL);if client==nil||err!=nil||parsed.Host==""||(parsed.Scheme!="http"&&parsed.Scheme!="https")||maxBody<=0{return nil,errors.New("invalid client")};return &Client{client:client,baseURL:strings.TrimRight(baseURL,"/"),maxBody:maxBody},nil}
func(c *Client)Probe(ctx context.Context,target string)(Result,error){if strings.TrimSpace(target)==""{return Result{},errors.New("target required")};endpoint,err:=url.Parse(c.baseURL+"/v1/probe");if err!=nil{return Result{},err};query:=endpoint.Query();query.Set("target",target);endpoint.RawQuery=query.Encode();request,err:=http.NewRequestWithContext(ctx,http.MethodGet,endpoint.String(),nil);if err!=nil{return Result{},err};request.Header.Set("Accept","application/json");response,err:=c.client.Do(request);if err!=nil{return Result{},err};body:=response.Body;defer body.Close();if response.StatusCode==http.StatusTooManyRequests{return Result{},ErrRateLimited};if response.StatusCode>=500{return Result{},ErrUpstream};if response.StatusCode>=400{return Result{},ErrRejected};if response.StatusCode!=http.StatusOK{return Result{},ErrRejected};data,err:=io.ReadAll(io.LimitReader(body,c.maxBody+1));if err!=nil{return Result{},err};if int64(len(data))>c.maxBody{return Result{},ErrBodyTooLarge};var result Result;if err:=json.Unmarshal(data,&result);err!=nil{return Result{},err};if result.Status==""{return Result{},errors.New("missing status")};return result,nil}
`
}

func deliveryClientSolution() string {
	return `package webhook
import("bytes";"context";"encoding/json";"errors";"io";"net/http";"net/url";"strings")
var(ErrRateLimited=errors.New("delivery rate limited");ErrRejected=errors.New("delivery rejected");ErrUpstream=errors.New("delivery upstream failed");ErrBodyTooLarge=errors.New("delivery response too large"))
type Command struct{Destination string ` + "`json:\"destination\"`" + `;Message string ` + "`json:\"message\"`" + `};type Result struct{DeliveryID string ` + "`json:\"delivery_id\"`" + `};type Client struct{client *http.Client;baseURL string;maxBody int64}
func New(client *http.Client,baseURL string,maxBody int64)(*Client,error){parsed,err:=url.Parse(baseURL);if client==nil||err!=nil||parsed.Host==""||(parsed.Scheme!="http"&&parsed.Scheme!="https")||maxBody<=0{return nil,errors.New("invalid client")};return &Client{client:client,baseURL:strings.TrimRight(baseURL,"/"),maxBody:maxBody},nil}
func(c *Client)Deliver(ctx context.Context,command Command)(Result,error){if strings.TrimSpace(command.Destination)==""||strings.TrimSpace(command.Message)==""{return Result{},errors.New("invalid command")};payload,err:=json.Marshal(command);if err!=nil{return Result{},err};request,err:=http.NewRequestWithContext(ctx,http.MethodPost,c.baseURL+"/v1/deliveries",bytes.NewReader(payload));if err!=nil{return Result{},err};request.Header.Set("Content-Type","application/json");request.Header.Set("Accept","application/json");response,err:=c.client.Do(request);if err!=nil{return Result{},err};body:=response.Body;defer body.Close();if response.StatusCode==http.StatusTooManyRequests{return Result{},ErrRateLimited};if response.StatusCode>=500{return Result{},ErrUpstream};if response.StatusCode>=400{return Result{},ErrRejected};if response.StatusCode!=http.StatusAccepted{return Result{},ErrRejected};data,err:=io.ReadAll(io.LimitReader(body,c.maxBody+1));if err!=nil{return Result{},err};if int64(len(data))>c.maxBody{return Result{},ErrBodyTooLarge};var result Result;if err:=json.Unmarshal(data,&result);err!=nil{return Result{},err};if result.DeliveryID==""{return Result{},errors.New("missing delivery id")};return result,nil}
`
}

func httpClientTableTests(method string) string {
	return "package " + map[string]string{"Probe": "httpclient", "Deliver": "webhook"}[method] + "\nimport \"testing\"\nfunc Test" + method + `Contract(t *testing.T){tests:=[]struct{name string}{{name:"request"},{name:"cancel"},{name:"limit"},{name:"rate-limit"},{name:"upstream"}};for _,tc:=range tests{t.Run(tc.name,func(t *testing.T){})}}
`
}
