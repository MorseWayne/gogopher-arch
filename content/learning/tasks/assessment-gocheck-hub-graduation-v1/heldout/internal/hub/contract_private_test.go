package hub

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type contractStore struct {
	mu sync.Mutex
	projects map[string]Project
	lookups []string
	creates int
	readyErr error
	jobs chan Job
	completed chan string
}
func (store *contractStore) CreateProject(_ context.Context, project Project) (Project,error) { store.mu.Lock(); defer store.mu.Unlock(); if store.projects==nil { store.projects=map[string]Project{} }; for _, current := range store.projects { if current.TenantID==project.TenantID && current.Name==project.Name { return Project{},ErrConflict } }; store.projects[project.TenantID+"/"+project.ID]=project; store.creates++; return project,nil }
func (store *contractStore) Project(_ context.Context, tenantID,id string)(Project,error){store.mu.Lock();defer store.mu.Unlock();store.lookups=append(store.lookups,tenantID+"/"+id);project,ok:=store.projects[tenantID+"/"+id];if !ok{return Project{},ErrNotFound};return project,nil}
func(store *contractStore)Ready(context.Context)error{return store.readyErr}
func(store *contractStore)Claim(ctx context.Context)(Job,error){select{case job:=<-store.jobs:return job,nil;case<-ctx.Done():return Job{},ctx.Err()}}
func(store *contractStore)Complete(_ context.Context,id string,result error)error{value:=id+":ok";if result!=nil{value=id+":failed"};store.completed<-value;return nil}

type contractCache struct{mu sync.Mutex;values map[string]Project;gets,sets,deletes int;err error}
func(cache *contractCache)Get(_ context.Context,key string)(Project,bool,error){cache.mu.Lock();defer cache.mu.Unlock();cache.gets++;if cache.err!=nil{return Project{},false,cache.err};project,ok:=cache.values[key];return project,ok,nil}
func(cache *contractCache)Set(_ context.Context,key string,project Project,_ time.Duration)error{cache.mu.Lock();defer cache.mu.Unlock();cache.sets++;if cache.values==nil{cache.values=map[string]Project{}};cache.values[key]=project;return cache.err}
func(cache *contractCache)Delete(_ context.Context,key string)error{cache.mu.Lock();defer cache.mu.Unlock();cache.deletes++;delete(cache.values,key);return cache.err}

func newContractService(t *testing.T,store Store,cache Cache)*Service{t.Helper();service,err:=NewService(store,cache,map[string]string{"tenant-a":"alpha-secret","tenant-b":"beta-secret"},slog.New(slog.NewJSONHandler(io.Discard,nil)));if err!=nil{t.Fatal(err)};return service}
func perform(handler http.Handler,method,path,key,body string)*httptest.ResponseRecorder{request:=httptest.NewRequest(method,path,strings.NewReader(body));if key!=""{request.Header.Set("X-API-Key",key)};response:=httptest.NewRecorder();handler.ServeHTTP(response,request);return response}

func TestAuthenticationPrecedesLookupAndScopesTenant(t *testing.T){store:=&contractStore{projects:map[string]Project{"tenant-a/p-7":{ID:"p-7",TenantID:"tenant-a",Name:"private"}}};handler:=newContractService(t,store,nil).Handler();if got:=perform(handler,http.MethodGet,"/v1/projects/p-7","","");got.Code!=http.StatusUnauthorized{t.Fatalf("unauthorized=%d",got.Code)};if len(store.lookups)!=0{t.Fatalf("lookup before auth: %v",store.lookups)};got:=perform(handler,http.MethodGet,"/v1/projects/p-7","beta-secret","");if got.Code!=http.StatusNotFound{t.Fatalf("cross tenant=%d body=%s",got.Code,got.Body.String())};if strings.Contains(got.Body.String(),"private")||strings.Contains(got.Body.String(),"tenant-a"){t.Fatal("resource leaked")};if len(store.lookups)!=1||store.lookups[0]!="tenant-b/p-7"{t.Fatalf("lookups=%v",store.lookups)}}

func TestProjectAPIUsesStrictJSONAndTenantCacheAside(t *testing.T){store:=&contractStore{projects:map[string]Project{"tenant-a/p-9":{ID:"p-9",TenantID:"tenant-a",Name:"cached"}}};cache:=&contractCache{values:map[string]Project{}};handler:=newContractService(t,store,cache).Handler();first:=perform(handler,http.MethodGet,"/v1/projects/p-9","alpha-secret","");second:=perform(handler,http.MethodGet,"/v1/projects/p-9","alpha-secret","");if first.Code!=200||second.Code!=200||len(store.lookups)!=1||cache.sets!=1{t.Fatalf("cache aside status=%d/%d lookup=%d sets=%d",first.Code,second.Code,len(store.lookups),cache.sets)};cache.err=errors.New("cache unavailable");if got:=perform(handler,http.MethodGet,"/v1/projects/p-9","alpha-secret","");got.Code!=200{t.Fatalf("cache outage status=%d",got.Code)};bad:=perform(handler,http.MethodPost,"/v1/projects","alpha-secret",`{"name":"new","extra":1}`);if bad.Code!=http.StatusBadRequest{t.Fatalf("unknown field=%d",bad.Code)};created:=perform(handler,http.MethodPost,"/v1/projects","alpha-secret",`{"name":" new "}`);if created.Code!=http.StatusCreated||!strings.Contains(created.Body.String(),`"name":"new"`){t.Fatalf("create=%d %s",created.Code,created.Body.String())};duplicate:=perform(handler,http.MethodPost,"/v1/projects","alpha-secret",`{"name":"new"}`);if duplicate.Code!=http.StatusConflict{t.Fatalf("duplicate=%d",duplicate.Code)}}

func TestWorkerConcurrencyIsBoundedAndCancellationJoins(t *testing.T){store:=&contractStore{jobs:make(chan Job,8),completed:make(chan string,8)};for index:=0;index<6;index++{store.jobs<-Job{ID:fmt.Sprintf("j-%d",index),Target:"https://example.test"}};service:=newContractService(t,store,nil);ctx,cancel:=context.WithCancel(context.Background());var active,maxActive atomic.Int32;release:=make(chan struct{});done:=make(chan error,1);go func(){done<-service.RunWorker(ctx,2,func(context.Context,string)error{current:=active.Add(1);for{maximum:=maxActive.Load();if current<=maximum||maxActive.CompareAndSwap(maximum,current){break}};<-release;active.Add(-1);return nil})}();deadline:=time.After(time.Second);for maxActive.Load()<2{select{case<-deadline:t.Fatal("worker did not reach concurrency");default:time.Sleep(time.Millisecond)}};if maxActive.Load()>2{t.Fatalf("max=%d",maxActive.Load())};close(release);for index:=0;index<2;index++{<-store.completed};cancel();select{case err:=<-done:if err!=nil{t.Fatalf("worker error=%v",err)};case<-time.After(time.Second):t.Fatal("worker did not join")};if active.Load()!=0{t.Fatalf("active=%d",active.Load())}}

func TestHealthRequestCorrelationAndLowCardinalityMetrics(t *testing.T){store:=&contractStore{projects:map[string]Project{}};service:=newContractService(t,store,nil);handler:=service.Handler();request:=httptest.NewRequest(http.MethodGet,"/livez",nil);request.Header.Set("X-Request-ID","request-123");response:=httptest.NewRecorder();handler.ServeHTTP(response,request);if response.Code!=200||response.Header().Get("X-Request-ID")!="request-123"{t.Fatalf("live=%d id=%q",response.Code,response.Header().Get("X-Request-ID"))};store.readyErr=errors.New("database password=secret");ready:=perform(handler,http.MethodGet,"/readyz","","");if ready.Code!=http.StatusServiceUnavailable||strings.Contains(ready.Body.String(),"password"){t.Fatalf("ready=%d %s",ready.Code,ready.Body.String())};metrics:=perform(handler,http.MethodGet,"/metrics","","");if metrics.Code!=200||!strings.Contains(metrics.Body.String(),"gocheckhub_http_requests_total")||strings.Contains(metrics.Body.String(),"request-123"){t.Fatalf("metrics=%d %s",metrics.Code,metrics.Body.String())}}
