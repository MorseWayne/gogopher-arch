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

func TestM208LearningLoopSolutionsPassRealReleaseAndSandboxEvaluation(t *testing.T) {
	tests := []struct {
		name, activity string
		files          map[string]string
		explanation    string
	}{
		{name: "API key authentication practice", activity: "practice-api-key-authentication", files: map[string]string{"authn/authenticator.go": apiKeyAuthenticatorSolution()}},
		{name: "project access control", activity: "assessment-gocheck-access-control", files: map[string]string{"internal/projectapi/handler.go": projectAccessSolution(), "internal/projectapi/handler_test.go": securityTableTests("projectapi", "AccessControl")}, explanation: "authentication 先严格解析 Bearer credential，构造时只保留 SHA-256 digest，并用 constant-time 比较降低时序侧信道。authentication 通过只说明调用者身份，authorization 仍必须在 project 资源上比较 owner。跨租户与不存在都返回 404，避免泄露资源是否存在。原始 secret、Authorization header、project 和底层错误都不进入日志或响应，审计只记录稳定 reason。"},
		{name: "alert rule access variant", activity: "review-gocheck-alert-access", files: map[string]string{"internal/alertaccess/handler.go": alertAccessSolution(), "internal/alertaccess/handler_test.go": securityTableTests("alertaccess", "AlertAccess")}, explanation: "variant 仍要把 authentication 与 authorization 分开：API key 通过 SHA-256 digest 和 constant-time 比较只确定 subject，DELETE 前还要检查 alert rule owner。不是 owner 时不调用 DeleteRule，而是与不存在共用 404，阻止 IDOR 探测。日志与错误包络只包含稳定代码，不包含 secret、header、rule 内容或 store 故障细节。"},
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
			current := attempt.Attempt{ID: "00000000-0000-4000-9980-00000000000" + strconv.Itoa(index+1), ReleaseID: registry.CurrentReleaseID(), ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace)}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4001-9020-00000000000" + strconv.Itoa(index+1)
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
			frozen := submission.Submission{ID: "00000000-0000-4001-9030-00000000000" + strconv.Itoa(index+1), AttemptID: current.ID, ReleaseID: current.ReleaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, Explanation: tc.explanation}
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

func apiKeyAuthenticatorSolution() string {
	return `package authn
import("crypto/sha256";"crypto/subtle";"errors";"strings")
type Credential struct{Subject string;Token string};type storedCredential struct{subject string;digest [32]byte};type Authenticator struct{credentials []storedCredential}
func New(credentials []Credential)(*Authenticator,error){if len(credentials)==0{return nil,errors.New("credentials required")};stored:=make([]storedCredential,0,len(credentials));seen:=map[[32]byte]struct{}{};for _,credential:=range credentials{if strings.TrimSpace(credential.Subject)==""||strings.TrimSpace(credential.Subject)!=credential.Subject||credential.Token==""||strings.ContainsAny(credential.Token," \t\r\n"){return nil,errors.New("invalid credential")};digest:=sha256.Sum256([]byte(credential.Token));if _,exists:=seen[digest];exists{return nil,errors.New("duplicate token")};seen[digest]=struct{}{};stored=append(stored,storedCredential{subject:credential.Subject,digest:digest})};return &Authenticator{credentials:stored},nil}
func(authenticator *Authenticator)Authenticate(authorization string)(string,bool){fields:=strings.Fields(authorization);if len(fields)!=2||fields[0]!="Bearer"{return "",false};digest:=sha256.Sum256([]byte(fields[1]));matched:=0;subject:="";for _,credential:=range authenticator.credentials{equal:=subtle.ConstantTimeCompare(digest[:],credential.digest[:]);if equal==1{subject=credential.subject};matched|=equal};return subject,matched==1}
`
}

func projectAccessSolution() string {
	return `package projectapi
import("context";"crypto/sha256";"crypto/subtle";"encoding/json";"errors";"net/http";"strings")
var ErrNotFound=errors.New("project not found")
type Credential struct{Subject string;Token string};type Project struct{ID string ` + "`json:\"id\"`" + `;OwnerID string ` + "`json:\"owner_id\"`" + `;Name string ` + "`json:\"name\"`" + `};type ProjectStore interface{FindProject(context.Context,string)(Project,error)};type AuditLogger interface{Denied(context.Context,string)};type storedCredential struct{subject string;digest [32]byte};type Handler struct{credentials []storedCredential;store ProjectStore;logger AuditLogger}
func New(credentials []Credential,store ProjectStore,logger AuditLogger)(*Handler,error){if len(credentials)==0||store==nil||logger==nil{return nil,errors.New("invalid dependencies")};stored:=make([]storedCredential,0,len(credentials));seen:=map[[32]byte]struct{}{};for _,credential:=range credentials{if strings.TrimSpace(credential.Subject)==""||strings.TrimSpace(credential.Subject)!=credential.Subject||credential.Token==""||strings.ContainsAny(credential.Token," \t\r\n"){return nil,errors.New("invalid credential")};digest:=sha256.Sum256([]byte(credential.Token));if _,exists:=seen[digest];exists{return nil,errors.New("duplicate token")};seen[digest]=struct{}{};stored=append(stored,storedCredential{subject:credential.Subject,digest:digest})};return &Handler{credentials:stored,store:store,logger:logger},nil}
func(handler *Handler)ServeHTTP(response http.ResponseWriter,request *http.Request){if request.Method!=http.MethodGet{response.Header().Set("Allow",http.MethodGet);writeError(response,http.StatusMethodNotAllowed,"method_not_allowed");return};subject,ok:=handler.authenticate(request.Header.Get("Authorization"));if !ok{handler.logger.Denied(request.Context(),"authentication_failed");response.Header().Set("WWW-Authenticate","Bearer");writeError(response,http.StatusUnauthorized,"unauthorized");return};id,ok:=canonicalID(request.URL.EscapedPath(),"/v1/projects/");if !ok{writeError(response,http.StatusBadRequest,"invalid_resource_id");return};project,err:=handler.store.FindProject(request.Context(),id);if errors.Is(err,ErrNotFound){writeError(response,http.StatusNotFound,"not_found");return};if err!=nil{writeError(response,http.StatusInternalServerError,"internal_error");return};if project.OwnerID!=subject{handler.logger.Denied(request.Context(),"resource_not_found");writeError(response,http.StatusNotFound,"not_found");return};response.Header().Set("Content-Type","application/json");json.NewEncoder(response).Encode(project)}
func(handler *Handler)authenticate(authorization string)(string,bool){fields:=strings.Fields(authorization);if len(fields)!=2||fields[0]!="Bearer"{return "",false};digest:=sha256.Sum256([]byte(fields[1]));matched:=0;subject:="";for _,credential:=range handler.credentials{equal:=subtle.ConstantTimeCompare(digest[:],credential.digest[:]);if equal==1{subject=credential.subject};matched|=equal};return subject,matched==1}
func canonicalID(path,prefix string)(string,bool){if !strings.HasPrefix(path,prefix){return "",false};id:=strings.TrimPrefix(path,prefix);if len(id)<1||len(id)>64{return "",false};for index,value:=range []byte(id){letter:=value>='a'&&value<='z';digit:=value>='0'&&value<='9';if !letter&&!digit&&(index==0||(value!='-'&&value!='_')){return "",false}};return id,true}
func writeError(response http.ResponseWriter,status int,code string){response.Header().Set("Content-Type","application/json");response.WriteHeader(status);json.NewEncoder(response).Encode(map[string]any{"error":map[string]string{"code":code}})}
`
}

func alertAccessSolution() string {
	return `package alertaccess
import("context";"crypto/sha256";"crypto/subtle";"encoding/json";"errors";"net/http";"strings")
var ErrNotFound=errors.New("alert rule not found")
type Credential struct{Subject string;Token string};type Rule struct{ID string;OwnerID string};type RuleStore interface{FindRule(context.Context,string)(Rule,error);DeleteRule(context.Context,string)error};type AuditLogger interface{Denied(context.Context,string)};type storedCredential struct{subject string;digest [32]byte};type Handler struct{credentials []storedCredential;store RuleStore;logger AuditLogger}
func New(credentials []Credential,store RuleStore,logger AuditLogger)(*Handler,error){if len(credentials)==0||store==nil||logger==nil{return nil,errors.New("invalid dependencies")};stored:=make([]storedCredential,0,len(credentials));seen:=map[[32]byte]struct{}{};for _,credential:=range credentials{if strings.TrimSpace(credential.Subject)==""||strings.TrimSpace(credential.Subject)!=credential.Subject||credential.Token==""||strings.ContainsAny(credential.Token," \t\r\n"){return nil,errors.New("invalid credential")};digest:=sha256.Sum256([]byte(credential.Token));if _,exists:=seen[digest];exists{return nil,errors.New("duplicate token")};seen[digest]=struct{}{};stored=append(stored,storedCredential{subject:credential.Subject,digest:digest})};return &Handler{credentials:stored,store:store,logger:logger},nil}
func(handler *Handler)ServeHTTP(response http.ResponseWriter,request *http.Request){if request.Method!=http.MethodDelete{response.Header().Set("Allow",http.MethodDelete);writeError(response,http.StatusMethodNotAllowed,"method_not_allowed");return};subject,ok:=handler.authenticate(request.Header.Get("Authorization"));if !ok{handler.logger.Denied(request.Context(),"authentication_failed");response.Header().Set("WWW-Authenticate","Bearer");writeError(response,http.StatusUnauthorized,"unauthorized");return};id,ok:=canonicalID(request.URL.EscapedPath(),"/v1/alert-rules/");if !ok{writeError(response,http.StatusBadRequest,"invalid_resource_id");return};rule,err:=handler.store.FindRule(request.Context(),id);if errors.Is(err,ErrNotFound){writeError(response,http.StatusNotFound,"not_found");return};if err!=nil{writeError(response,http.StatusInternalServerError,"internal_error");return};if rule.OwnerID!=subject{handler.logger.Denied(request.Context(),"resource_not_found");writeError(response,http.StatusNotFound,"not_found");return};if err:=handler.store.DeleteRule(request.Context(),id);errors.Is(err,ErrNotFound){writeError(response,http.StatusNotFound,"not_found");return}else if err!=nil{writeError(response,http.StatusInternalServerError,"internal_error");return};response.WriteHeader(http.StatusNoContent)}
func(handler *Handler)authenticate(authorization string)(string,bool){fields:=strings.Fields(authorization);if len(fields)!=2||fields[0]!="Bearer"{return "",false};digest:=sha256.Sum256([]byte(fields[1]));matched:=0;subject:="";for _,credential:=range handler.credentials{equal:=subtle.ConstantTimeCompare(digest[:],credential.digest[:]);if equal==1{subject=credential.subject};matched|=equal};return subject,matched==1}
func canonicalID(path,prefix string)(string,bool){if !strings.HasPrefix(path,prefix){return "",false};id:=strings.TrimPrefix(path,prefix);if len(id)<1||len(id)>64{return "",false};for index,value:=range []byte(id){letter:=value>='a'&&value<='z';digit:=value>='0'&&value<='9';if !letter&&!digit&&(index==0||(value!='-'&&value!='_')){return "",false}};return id,true}
func writeError(response http.ResponseWriter,status int,code string){response.Header().Set("Content-Type","application/json");response.WriteHeader(status);json.NewEncoder(response).Encode(map[string]any{"error":map[string]string{"code":code}})}
`
}

func securityTableTests(packageName, method string) string {
	return "package " + packageName + "\nimport \"testing\"\nfunc Test" + method + `Contract(t *testing.T){tests:=[]struct{name string}{{name:"owner"},{name:"missing credential"},{name:"invalid credential"},{name:"other owner"},{name:"missing resource"},{name:"invalid id"}};for _,test:=range tests{t.Run(test.name,func(t *testing.T){})}}
`
}
