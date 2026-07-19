package evaluation

import (
	"strconv"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

func TestM216LearningLoopSolutionsPassGo126SandboxEvaluation(t *testing.T) {
	tests := []struct {
		name, activity string
		files          map[string]string
		explanation    string
	}{
		{name: "AI delivery review", activity: "practice-gocheck-hub-ai-review", files: map[string]string{
			"deliveryreview/review.go":      aiDeliveryReviewSolution(),
			"deliveryreview/review_test.go": deliveryReviewTests(),
		}, explanation: "AI generated the first delivery proposal, but a human remains accountable for verification. I checked authentication before lookup, kept PostgreSQL as the source of truth with cache degradation, bounded worker concurrency and retries, required forward-only migration and the complete quality gate set, and rejected root runtime. The rollback is a forward source revert that preserves migrations and audit data rather than a destructive database rollback."},
		{name: "independent gocheck-hub delivery", activity: "assessment-gocheck-hub-graduation", files: gocheckHubSolution(), explanation: "authentication runs before any resource lookup and maps a digest to one tenant, so cross-tenant absence is stable. PostgreSQL remains the source of truth while cache-aside is only an outage-tolerant read optimization. The worker has fixed concurrency, propagates context cancellation, and joins before shutdown. Liveness is process-only, readiness checks the database boundary, and metrics use low-cardinality labels. Two forward-only migrations preserve deployed data. A multi-stage non-root image and matching local/CI gates run vet, tests, race detection, migration checks and vulnerability scanning."},
		{name: "delayed alertboard variant", activity: "review-alertboard-delivery", files: map[string]string{
			"internal/alertboard/service.go":      alertboardSolution(),
			"internal/alertboard/service_test.go": alertboardTests(),
		}, explanation: "authentication hashes the supplied key and resolves exactly one tenant before store or cache access. The read path uses tenant-aware cache-aside, but the Store is the source of truth and cache failures degrade. Acknowledge commits the store before invalidation, so a failed write cannot evict valid data. The bounded worker uses a fixed goroutine count, passes context to Next and delivery, records completion, and joins after cancellation."},
	}

	registry := draftReleaseRegistry(t)
	for index, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			activity, err := registry.ActivityView(registry.CurrentReleaseID(), testCase.activity, 1)
			if err != nil {
				t.Fatal(err)
			}
			task, err := registry.ExecutionTask(registry.CurrentReleaseID(), activity.TaskRef.ID, activity.TaskRef.Version)
			if err != nil {
				t.Fatal(err)
			}
			if task.Language != execution.GoLanguage126 {
				t.Fatalf("task language = %q", task.Language)
			}
			workspace, err := registry.PublicWorkspace(registry.CurrentReleaseID(), task.ID, task.Version)
			if err != nil {
				t.Fatal(err)
			}
			for path, source := range testCase.files {
				workspace[path] = source
			}
			current := attempt.Attempt{ID: "00000000-0000-4000-9350-00000000000" + strconv.Itoa(index+1), ReleaseID: registry.CurrentReleaseID(), ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace)}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4001-9350-00000000000" + strconv.Itoa(index+1)
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
			frozen := submission.Submission{ID: "00000000-0000-4001-9360-00000000000" + strconv.Itoa(index+1), AttemptID: current.ID, ReleaseID: current.ReleaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, Explanation: testCase.explanation}
			terminal := execution.Execution{ID: executionID, AttemptID: current.ID, SubmissionID: frozen.ID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Status: response.Status, Response: &response}
			generator, err := NewGenerator(registry)
			if err != nil {
				t.Fatal(err)
			}
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

func aiDeliveryReviewSolution() string {
	return `package deliveryreview
type Plan struct{AIDeclared bool;AuthBeforeLookup bool;SourceOfTruth string;CacheFailureMode string;WorkerConcurrency int;RetryLimit int;MigrationMode string;Gates []string;RuntimeUser string;Rollback string}
func Review(plan Plan)[]string{findings:=[]string{};if !plan.AIDeclared{findings=append(findings,"ai-undeclared")};if !plan.AuthBeforeLookup{findings=append(findings,"auth-after-lookup")};if plan.SourceOfTruth!="postgres"{findings=append(findings,"invalid-source-of-truth")};if plan.CacheFailureMode!="degrade"{findings=append(findings,"cache-does-not-degrade")};if plan.WorkerConcurrency<=0{findings=append(findings,"worker-unbounded")};if plan.RetryLimit<0{findings=append(findings,"retry-unbounded")};if plan.MigrationMode!="forward-only"{findings=append(findings,"migration-not-forward-only")};required:=map[string]bool{"fmt":false,"vet":false,"test":false,"race":false,"vuln":false,"migration":false,"image":false};complete:=len(plan.Gates)==len(required);for _,gate:=range plan.Gates{seen,ok:=required[gate];if !ok||seen{complete=false;continue};required[gate]=true};for _,seen:=range required{if !seen{complete=false}};if !complete{findings=append(findings,"delivery-gates-incomplete")};if plan.RuntimeUser==""||plan.RuntimeUser=="root"{findings=append(findings,"runtime-is-root")};if plan.Rollback!="forward-revert"{findings=append(findings,"rollback-not-forward")};return findings}
`
}

func deliveryReviewTests() string {
	return `package deliveryreview
import "testing"
func TestReviewCases(t *testing.T){base:=Plan{AIDeclared:true,AuthBeforeLookup:true,SourceOfTruth:"postgres",CacheFailureMode:"degrade",WorkerConcurrency:2,RetryLimit:2,MigrationMode:"forward-only",Gates:[]string{"fmt","vet","test","race","vuln","migration","image"},RuntimeUser:"app",Rollback:"forward-revert"};tests:=[]struct{name string;mutate func(*Plan);want int}{{"safe",func(*Plan){},0},{"AI",func(plan *Plan){plan.AIDeclared=false},1},{"auth",func(plan *Plan){plan.AuthBeforeLookup=false},1},{"cache",func(plan *Plan){plan.SourceOfTruth="cache"},1},{"worker",func(plan *Plan){plan.WorkerConcurrency=0},1},{"migration",func(plan *Plan){plan.MigrationMode="down"},1},{"runtime",func(plan *Plan){plan.RuntimeUser="root"},1}};for _,tc:=range tests{t.Run(tc.name,func(t *testing.T){plan:=base;tc.mutate(&plan);if got:=len(Review(plan));got!=tc.want{t.Fatalf("got=%d",got)}})}}
`
}

func gocheckHubSolution() map[string]string {
	return map[string]string{
		"go.mod":                                 "module gocheckhub.local/service\n\ngo 1.26.0\n",
		"README.md":                              "# gocheck-hub\n\nHermetic M2 graduation service with an injectable SQL boundary.\n",
		"internal/hub/service.go":                gocheckHubService(),
		"internal/hub/memory.go":                 gocheckHubMemory(),
		"internal/hub/sqlstore.go":               gocheckHubSQLStore(),
		"internal/hub/service_test.go":           "package hub\nimport \"testing\"\nfunc TestLearnerCases(t *testing.T){tests:=[]struct{name string}{{\"auth\"},{\"strict JSON\"},{\"tenant\"},{\"cache\"},{\"worker\"},{\"health\"}};for _,tc:=range tests{t.Run(tc.name,func(t *testing.T){})}}\n",
		"cmd/gocheckhub/main.go":                 gocheckHubMain(),
		"migrations/0001_create_projects.up.sql": "CREATE TABLE projects (tenant_id text NOT NULL, id text NOT NULL, name text NOT NULL, PRIMARY KEY (tenant_id,id), UNIQUE (tenant_id,name));\n",
		"migrations/0002_add_jobs.up.sql":        "CREATE TABLE jobs (id text PRIMARY KEY, tenant_id text NOT NULL, target text NOT NULL, status text NOT NULL CHECK (status IN ('pending','running','done','failed')));\nCREATE INDEX jobs_claim_idx ON jobs (status,id);\n",
		"Dockerfile":                             "FROM golang:1.26-alpine AS builder\nWORKDIR /src\nCOPY . .\nRUN CGO_ENABLED=0 go test ./... && CGO_ENABLED=0 go build -trimpath -o /out/gocheckhub ./cmd/gocheckhub\nFROM alpine:3.23\nRUN addgroup -S app && adduser -S -G app app\nCOPY --from=builder /out/gocheckhub /usr/local/bin/gocheckhub\nUSER app\nENTRYPOINT [\"/usr/local/bin/gocheckhub\"]\n",
		"Makefile":                               ".PHONY: verify migration-check\nverify:\n\ttest -z \"$$(gofmt -l .)\"\n\tgo vet ./...\n\tgo test ./...\n\tgo test -race ./...\n\tgovulncheck ./...\n\t$(MAKE) migration-check\nmigration-check:\n\ttest ! -e migrations/0001_create_projects.down.sql\n",
		".github/workflows/ci.yml":               "name: verify\non: [push, pull_request]\npermissions:\n  contents: read\njobs:\n  test:\n    runs-on: ubuntu-latest\n    steps:\n      - uses: actions/checkout@v4\n      - uses: actions/setup-go@v5\n        with:\n          go-version: '1.26.0'\n      - run: make verify\n      - run: docker build -t gocheckhub:ci .\n",
	}
}

func gocheckHubService() string {
	return `package hub
import("context";"crypto/sha256";"crypto/subtle";"encoding/json";"errors";"fmt";"io";"log/slog";"net/http";"strings";"sync";"sync/atomic";"time")
var(ErrNotFound=errors.New("not found");ErrConflict=errors.New("conflict");ErrNoJob=errors.New("no job"))
type Project struct{ID string ` + "`json:\"id\"`" + `;TenantID string ` + "`json:\"tenant_id\"`" + `;Name string ` + "`json:\"name\"`" + `}
type Job struct{ID,TenantID,Target string}
type Store interface{CreateProject(context.Context,Project)(Project,error);Project(context.Context,string,string)(Project,error);Ready(context.Context)error;Claim(context.Context)(Job,error);Complete(context.Context,string,error)error}
type Cache interface{Get(context.Context,string)(Project,bool,error);Set(context.Context,string,Project,time.Duration)error;Delete(context.Context,string)error}
type credential struct{tenant string;digest [32]byte}
type Service struct{store Store;cache Cache;credentials []credential;logger *slog.Logger;ids atomic.Uint64;requests atomic.Uint64}
func NewService(store Store,cache Cache,credentials map[string]string,logger *slog.Logger)(*Service,error){if store==nil||logger==nil||len(credentials)==0{return nil,errors.New("invalid dependencies")};service:=&Service{store:store,cache:cache,logger:logger};for tenant,key:=range credentials{if strings.TrimSpace(tenant)==""||key==""{return nil,errors.New("invalid credential")};service.credentials=append(service.credentials,credential{tenant:tenant,digest:sha256.Sum256([]byte(key))})};return service,nil}
func(service *Service)Handler()http.Handler{mux:=http.NewServeMux();mux.HandleFunc("GET /livez",service.livez);mux.HandleFunc("GET /readyz",service.readyz);mux.HandleFunc("GET /metrics",service.metrics);mux.HandleFunc("POST /v1/projects",service.createProject);mux.HandleFunc("GET /v1/projects/{id}",service.getProject);return http.HandlerFunc(func(writer http.ResponseWriter,request *http.Request){service.requests.Add(1);id:=request.Header.Get("X-Request-ID");if id==""{id=fmt.Sprintf("req-%d",service.ids.Add(1))};writer.Header().Set("X-Request-ID",id);mux.ServeHTTP(writer,request)})}
func(service *Service)tenant(request *http.Request)(string,bool){supplied:=sha256.Sum256([]byte(request.Header.Get("X-API-Key")));tenant:="";matched:=0;for _,candidate:=range service.credentials{equal:=subtle.ConstantTimeCompare(supplied[:],candidate.digest[:]);if equal==1{tenant=candidate.tenant};matched|=equal};return tenant,matched==1}
func writeJSON(writer http.ResponseWriter,status int,value any){writer.Header().Set("Content-Type","application/json");writer.WriteHeader(status);_ = json.NewEncoder(writer).Encode(value)}
func writeError(writer http.ResponseWriter,status int,code string){writeJSON(writer,status,map[string]string{"error":code})}
func(service *Service)authenticate(writer http.ResponseWriter,request *http.Request)(string,bool){tenant,ok:=service.tenant(request);if !ok{writeError(writer,http.StatusUnauthorized,"unauthorized")};return tenant,ok}
func(service *Service)createProject(writer http.ResponseWriter,request *http.Request){tenant,ok:=service.authenticate(writer,request);if !ok{return};var input struct{Name string ` + "`json:\"name\"`" + `};decoder:=json.NewDecoder(request.Body);decoder.DisallowUnknownFields();if err:=decoder.Decode(&input);err!=nil{writeError(writer,400,"invalid_json");return};var extra any;if err:=decoder.Decode(&extra);err!=io.EOF{writeError(writer,400,"invalid_json");return};input.Name=strings.TrimSpace(input.Name);if input.Name==""{writeError(writer,400,"invalid_name");return};project:=Project{ID:fmt.Sprintf("p-%d",service.ids.Add(1)),TenantID:tenant,Name:input.Name};created,err:=service.store.CreateProject(request.Context(),project);if errors.Is(err,ErrConflict){writeError(writer,409,"conflict");return};if err!=nil{writeError(writer,500,"internal");return};writeJSON(writer,201,created)}
func(service *Service)getProject(writer http.ResponseWriter,request *http.Request){tenant,ok:=service.authenticate(writer,request);if !ok{return};id:=request.PathValue("id");key:=tenant+"/"+id;if service.cache!=nil{if project,hit,err:=service.cache.Get(request.Context(),key);err==nil&&hit{writeJSON(writer,200,project);return}};project,err:=service.store.Project(request.Context(),tenant,id);if errors.Is(err,ErrNotFound){writeError(writer,404,"not_found");return};if err!=nil{writeError(writer,500,"internal");return};if service.cache!=nil{_ = service.cache.Set(request.Context(),key,project,time.Minute)};writeJSON(writer,200,project)}
func(service *Service)livez(writer http.ResponseWriter,_ *http.Request){writer.WriteHeader(200);_,_=writer.Write([]byte("ok"))}
func(service *Service)readyz(writer http.ResponseWriter,request *http.Request){if service.store.Ready(request.Context())!=nil{writeError(writer,503,"not_ready");return};writer.WriteHeader(200);_,_=writer.Write([]byte("ready"))}
func(service *Service)metrics(writer http.ResponseWriter,_ *http.Request){writer.Header().Set("Content-Type","text/plain");_,_=fmt.Fprintf(writer,"gocheckhub_http_requests_total{route=%q} %d\n","all",service.requests.Load())}
func(service *Service)RunWorker(ctx context.Context,concurrency int,probe func(context.Context,string)error)error{if concurrency<=0||probe==nil{return errors.New("invalid worker")};runCtx,cancel:=context.WithCancel(ctx);defer cancel();failures:=make(chan error,concurrency);var wait sync.WaitGroup;for index:=0;index<concurrency;index++{wait.Add(1);go func(){defer wait.Done();for{job,err:=service.store.Claim(runCtx);if err!=nil{if runCtx.Err()!=nil{return};if errors.Is(err,ErrNoJob){timer:=time.NewTimer(10*time.Millisecond);select{case<-timer.C:case<-runCtx.Done():if !timer.Stop(){<-timer.C};return};continue};select{case failures<-err:default:};cancel();return};result:=probe(runCtx,job.Target);if err:=service.store.Complete(runCtx,job.ID,result);err!=nil{select{case failures<-err:default:};cancel();return}}}()};wait.Wait();select{case err:=<-failures:return err;default:return nil}}
`
}

func gocheckHubMemory() string {
	return `package hub
import("context";"sync")
type MemoryStore struct{mu sync.RWMutex;projects map[string]Project}
func NewMemoryStore()*MemoryStore{return &MemoryStore{projects:map[string]Project{}}}
func(store *MemoryStore)CreateProject(_ context.Context,project Project)(Project,error){store.mu.Lock();defer store.mu.Unlock();for _,current:=range store.projects{if current.TenantID==project.TenantID&&current.Name==project.Name{return Project{},ErrConflict}};store.projects[project.TenantID+"/"+project.ID]=project;return project,nil}
func(store *MemoryStore)Project(_ context.Context,tenant,id string)(Project,error){store.mu.RLock();defer store.mu.RUnlock();project,ok:=store.projects[tenant+"/"+id];if !ok{return Project{},ErrNotFound};return project,nil}
func(store *MemoryStore)Ready(context.Context)error{return nil}
func(store *MemoryStore)Claim(ctx context.Context)(Job,error){<-ctx.Done();return Job{},ctx.Err()}
func(store *MemoryStore)Complete(context.Context,string,error)error{return nil}
`
}

func gocheckHubSQLStore() string {
	return `package hub
import("context";"database/sql";"errors")
type SQLStore struct{db *sql.DB}
func NewSQLStore(db *sql.DB)*SQLStore{return &SQLStore{db:db}}
func(store *SQLStore)Ready(ctx context.Context)error{db:=store.db;return db.PingContext(ctx)}
func(store *SQLStore)CreateProject(ctx context.Context,project Project)(Project,error){db:=store.db;err:=db.QueryRowContext(ctx,"INSERT INTO projects(tenant_id,id,name) VALUES($1,$2,$3) RETURNING tenant_id,id,name",project.TenantID,project.ID,project.Name).Scan(&project.TenantID,&project.ID,&project.Name);if err!=nil{return Project{},err};return project,nil}
func(store *SQLStore)Project(ctx context.Context,tenant,id string)(Project,error){db:=store.db;var project Project;err:=db.QueryRowContext(ctx,"SELECT tenant_id,id,name FROM projects WHERE tenant_id=$1 AND id=$2",tenant,id).Scan(&project.TenantID,&project.ID,&project.Name);if errors.Is(err,sql.ErrNoRows){return Project{},ErrNotFound};return project,err}
func(store *SQLStore)Claim(context.Context)(Job,error){return Job{},ErrNoJob}
func(store *SQLStore)Complete(ctx context.Context,id string,result error)error{db:=store.db;status:="done";if result!=nil{status="failed"};_,err:=db.ExecContext(ctx,"UPDATE jobs SET status=$2 WHERE id=$1",id,status);return err}
`
}

func gocheckHubMain() string {
	return `package main
import("io";"log";"log/slog";"net/http";"os";"time";"gocheckhub.local/service/internal/hub")
func main(){key:=os.Getenv("API_KEY");if key==""{key="local-development-key"};logger:=slog.New(slog.NewJSONHandler(io.Discard,nil));service,err:=hub.NewService(hub.NewMemoryStore(),nil,map[string]string{"local":key},logger);if err!=nil{log.Fatal(err)};server:=&http.Server{Addr:":8080",Handler:service.Handler(),ReadHeaderTimeout:5*time.Second,ReadTimeout:10*time.Second,WriteTimeout:10*time.Second,IdleTimeout:30*time.Second};log.Fatal(server.ListenAndServe())}
`
}

func alertboardSolution() string {
	return `package alertboard
import("context";"crypto/sha256";"crypto/subtle";"encoding/json";"errors";"net/http";"sync";"time")
var ErrNotFound=errors.New("not found")
type Alert struct{ID, TenantID, Message string;Acknowledged bool};type Delivery struct{ID,TenantID,Target string}
type Store interface{Alert(context.Context,string,string)(Alert,error);Acknowledge(context.Context,string,string)error;Next(context.Context)(Delivery,error);Complete(context.Context,string,error)error}
type Cache interface{Get(context.Context,string)(Alert,bool,error);Set(context.Context,string,Alert,time.Duration)error;Delete(context.Context,string)error}
type credential struct{tenant string;digest [32]byte};type Service struct{store Store;cache Cache;credentials []credential}
func NewService(store Store,cache Cache,values map[string]string)(*Service,error){if store==nil||len(values)==0{return nil,errors.New("invalid dependencies")};service:=&Service{store:store,cache:cache};for tenant,key:=range values{if tenant==""||key==""{return nil,errors.New("invalid credential")};service.credentials=append(service.credentials,credential{tenant,sha256.Sum256([]byte(key))})};return service,nil}
func(service *Service)tenant(request *http.Request)(string,bool){digest:=sha256.Sum256([]byte(request.Header.Get("X-API-Key")));tenant:="";matched:=0;for _,candidate:=range service.credentials{equal:=subtle.ConstantTimeCompare(digest[:],candidate.digest[:]);if equal==1{tenant=candidate.tenant};matched|=equal};return tenant,matched==1}
func(service *Service)Handler()http.Handler{mux:=http.NewServeMux();mux.HandleFunc("GET /v1/alerts/{id}",service.get);mux.HandleFunc("POST /v1/alerts/{id}/ack",service.ack);return mux}
func(service *Service)authenticate(writer http.ResponseWriter,request *http.Request)(string,bool){tenant,ok:=service.tenant(request);if !ok{http.Error(writer,"unauthorized",401)};return tenant,ok}
func(service *Service)get(writer http.ResponseWriter,request *http.Request){tenant,ok:=service.authenticate(writer,request);if !ok{return};key:=tenant+"/"+request.PathValue("id");if service.cache!=nil{if alert,hit,err:=service.cache.Get(request.Context(),key);err==nil&&hit{_ = json.NewEncoder(writer).Encode(alert);return}};alert,err:=service.store.Alert(request.Context(),tenant,request.PathValue("id"));if errors.Is(err,ErrNotFound){http.Error(writer,"not found",404);return};if err!=nil{http.Error(writer,"internal",500);return};if service.cache!=nil{_ = service.cache.Set(request.Context(),key,alert,time.Minute)};_ = json.NewEncoder(writer).Encode(alert)}
func(service *Service)ack(writer http.ResponseWriter,request *http.Request){tenant,ok:=service.authenticate(writer,request);if !ok{return};id:=request.PathValue("id");if err:=service.store.Acknowledge(request.Context(),tenant,id);errors.Is(err,ErrNotFound){http.Error(writer,"not found",404);return}else if err!=nil{http.Error(writer,"internal",500);return};if service.cache!=nil{if err:=service.cache.Delete(request.Context(),tenant+"/"+id);err!=nil{http.Error(writer,"committed but cache invalidation failed",500);return}};writer.WriteHeader(204)}
func(service *Service)Run(ctx context.Context,concurrency int,deliver func(context.Context,Delivery)error)error{if concurrency<=0||deliver==nil{return errors.New("invalid worker")};runCtx,cancel:=context.WithCancel(ctx);defer cancel();failures:=make(chan error,concurrency);var wait sync.WaitGroup;for index:=0;index<concurrency;index++{wait.Add(1);go func(){defer wait.Done();for{delivery,err:=service.store.Next(runCtx);if err!=nil{if runCtx.Err()!=nil{return};select{case failures<-err:default:};cancel();return};result:=deliver(runCtx,delivery);if err:=service.store.Complete(runCtx,delivery.ID,result);err!=nil{select{case failures<-err:default:};cancel();return}}}()};wait.Wait();select{case err:=<-failures:return err;default:return nil}}
`
}

func alertboardTests() string {
	return `package alertboard
import "testing"
func TestAlertboardCases(t *testing.T){tests:=[]struct{name string}{{"auth first"},{"tenant isolation"},{"cache hit"},{"cache outage"},{"ack order"},{"bounded cancellation"}};for _,tc:=range tests{t.Run(tc.name,func(t *testing.T){})}}
`
}
