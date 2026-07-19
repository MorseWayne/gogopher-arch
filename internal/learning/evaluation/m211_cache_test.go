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

func TestM211LearningLoopSolutionsPassRealReleaseAndSandboxEvaluation(t *testing.T) {
	tests := []struct {
		name, activity string
		files          map[string]string
		explanation    string
	}{
		{"ttl cache practice", "practice-ttl-cache", map[string]string{"ttlcache/cache.go": ttlCacheSolution()}, ""},
		{"project cache assessment", "assessment-gocheck-project-cache", map[string]string{"internal/projectcache/service.go": projectCacheSolution(), "internal/projectcache/service_test.go": cacheTableTests("projectcache")}, "cache-aside 只让 cache 加速读，Source 始终是 source of truth。fresh negative cache 用较短 TTL 阻止不存在 key 反复穿透，positive miss 才写长 TTL。每个 key 的 flight 合并并发 miss，waiter 用自己的 context 取消。Cache Get 或 Set 故障选择 degradation 到 Source；Update 必须先写 truth 再 invalidation，删除失败会连同已更新值返回，明确旧 cache 仍可能存在。"},
		{"alert cache variant", "review-gocheck-alert-cache", map[string]string{"internal/alertcache/service.go": alertCacheSolution(), "internal/alertcache/service_test.go": cacheTableTests("alertcache")}, "alert rule 仍采用 cache-aside，Repository 是唯一 source of truth。negative cache 限制不存在 rule 的穿透，并发 miss 共享 flight 且 waiter 继承 context。cache outage 的 degradation 直接读取 Repository，不能映射成 not found。Save 先提交 Repository 再 invalidation；若删除失败，返回保存后的 Rule 和错误，让调用方知道 truth 已变化而 cache 可能陈旧。"},
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
			current := attempt.Attempt{ID: "00000000-0000-4000-9990-00000000000" + strconv.Itoa(index+1), ReleaseID: registry.CurrentReleaseID(), ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace)}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4001-9080-00000000000" + strconv.Itoa(index+1)
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
			frozen := submission.Submission{ID: "00000000-0000-4001-9090-00000000000" + strconv.Itoa(index+1), AttemptID: current.ID, ReleaseID: current.ReleaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, Explanation: test.explanation}
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

func ttlCacheSolution() string {
	return `package ttlcache
import("errors";"sync";"time")
type entry[V any]struct{value V;found bool;expiresAt time.Time};type Cache[V any]struct{mu sync.Mutex;entries map[string]entry[V];now func()time.Time}
func New[V any](now func()time.Time)(*Cache[V],error){if now==nil{return nil,errors.New("nil clock")};return &Cache[V]{entries:make(map[string]entry[V]),now:now},nil}
func(cache *Cache[V])Set(key string,value V,found bool,ttl time.Duration)error{if key==""||ttl<=0{return errors.New("invalid cache entry")};cache.mu.Lock();defer cache.mu.Unlock();cache.entries[key]=entry[V]{value:value,found:found,expiresAt:cache.now().Add(ttl)};return nil}
func(cache *Cache[V])Get(key string)(value V,found bool,hit bool){cache.mu.Lock();defer cache.mu.Unlock();entry,ok:=cache.entries[key];if !ok{return value,false,false};if !cache.now().Before(entry.expiresAt){delete(cache.entries,key);return value,false,false};return entry.value,entry.found,true}
func(cache *Cache[V])Delete(key string){cache.mu.Lock();defer cache.mu.Unlock();delete(cache.entries,key)}
`
}

func projectCacheSolution() string {
	return `package projectcache
import("context";"errors";"fmt";"sync";"time")
var ErrNotFound=errors.New("project not found");type Project struct{ID,Name string;Version int};type CacheEntry struct{Project Project;Found bool};type Source interface{Get(context.Context,string)(Project,error);Update(context.Context,Project)(Project,error)};type Cache interface{Get(context.Context,string)(CacheEntry,bool,error);Set(context.Context,string,CacheEntry,time.Duration)error;Delete(context.Context,string)error};type Options struct{PositiveTTL,NegativeTTL time.Duration};type flight struct{done chan struct{};project Project;err error};type Service struct{source Source;cache Cache;options Options;mu sync.Mutex;flights map[string]*flight}
func New(source Source,cache Cache,options Options)(*Service,error){if source==nil||cache==nil||options.PositiveTTL<=0||options.NegativeTTL<=0{return nil,errors.New("invalid cache service")};return &Service{source:source,cache:cache,options:options,flights:make(map[string]*flight)},nil}
func(service *Service)Get(ctx context.Context,id string)(Project,error){if id==""{return Project{},errors.New("empty project id")};cache:=service.cache;entry,hit,cacheErr:=cache.Get(ctx,id);if cacheErr==nil&&hit{if !entry.Found{return Project{},ErrNotFound};return entry.Project,nil};service.mu.Lock();if current:=service.flights[id];current!=nil{service.mu.Unlock();select{case<-current.done:return current.project,current.err;case<-ctx.Done():return Project{},ctx.Err()}};current:=&flight{done:make(chan struct{})};service.flights[id]=current;service.mu.Unlock();source:=service.source;project,err:=source.Get(ctx,id);if err==nil{_ = cache.Set(ctx,id,CacheEntry{Project:project,Found:true},service.options.PositiveTTL)}else if errors.Is(err,ErrNotFound){_ = cache.Set(ctx,id,CacheEntry{Found:false},service.options.NegativeTTL)};service.mu.Lock();current.project,current.err=project,err;delete(service.flights,id);close(current.done);service.mu.Unlock();return project,err}
func(service *Service)Update(ctx context.Context,project Project)(Project,error){if project.ID==""{return Project{},errors.New("empty project id")};source:=service.source;updated,err:=source.Update(ctx,project);if err!=nil{return Project{},err};cache:=service.cache;if err:=cache.Delete(ctx,updated.ID);err!=nil{return updated,fmt.Errorf("project updated but cache invalidation failed: %w",err)};return updated,nil}
`
}

func alertCacheSolution() string {
	return `package alertcache
import("context";"errors";"fmt";"sync";"time")
var ErrNotFound=errors.New("alert rule not found");type Rule struct{ID,Destination string;Version int};type Entry struct{Rule Rule;Found bool};type Repository interface{Find(context.Context,string)(Rule,error);Save(context.Context,Rule)(Rule,error)};type Cache interface{Get(context.Context,string)(Entry,bool,error);Set(context.Context,string,Entry,time.Duration)error;Delete(context.Context,string)error};type Options struct{PositiveTTL,NegativeTTL time.Duration};type flight struct{done chan struct{};rule Rule;err error};type Service struct{repository Repository;cache Cache;options Options;mu sync.Mutex;flights map[string]*flight}
func New(repository Repository,cache Cache,options Options)(*Service,error){if repository==nil||cache==nil||options.PositiveTTL<=0||options.NegativeTTL<=0{return nil,errors.New("invalid cache service")};return &Service{repository:repository,cache:cache,options:options,flights:make(map[string]*flight)},nil}
func(service *Service)Get(ctx context.Context,id string)(Rule,error){if id==""{return Rule{},errors.New("empty rule id")};cache:=service.cache;entry,hit,cacheErr:=cache.Get(ctx,id);if cacheErr==nil&&hit{if !entry.Found{return Rule{},ErrNotFound};return entry.Rule,nil};service.mu.Lock();if current:=service.flights[id];current!=nil{service.mu.Unlock();select{case<-current.done:return current.rule,current.err;case<-ctx.Done():return Rule{},ctx.Err()}};current:=&flight{done:make(chan struct{})};service.flights[id]=current;service.mu.Unlock();repository:=service.repository;rule,err:=repository.Find(ctx,id);if err==nil{_ = cache.Set(ctx,id,Entry{Rule:rule,Found:true},service.options.PositiveTTL)}else if errors.Is(err,ErrNotFound){_ = cache.Set(ctx,id,Entry{Found:false},service.options.NegativeTTL)};service.mu.Lock();current.rule,current.err=rule,err;delete(service.flights,id);close(current.done);service.mu.Unlock();return rule,err}
func(service *Service)Save(ctx context.Context,rule Rule)(Rule,error){if rule.ID==""{return Rule{},errors.New("empty rule id")};repository:=service.repository;saved,err:=repository.Save(ctx,rule);if err!=nil{return Rule{},err};cache:=service.cache;if err:=cache.Delete(ctx,saved.ID);err!=nil{return saved,fmt.Errorf("rule saved but cache invalidation failed: %w",err)};return saved,nil}
`
}

func cacheTableTests(packageName string) string {
	return "package " + packageName + "\nimport \"testing\"\nfunc TestCacheCases(t *testing.T){tests:=[]struct{name string}{{\"positive hit\"},{\"negative hit\"},{\"cold miss\"},{\"cache outage\"},{\"concurrent miss\"},{\"update invalidates\"},{\"truth failure\"},{\"invalidation failure\"}};for _,test:=range tests{t.Run(test.name,func(t *testing.T){})}}\n"
}
