package evaluation

import (
	"strconv"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

func TestM215LearningLoopSolutionsPassGo126SandboxEvaluation(t *testing.T) {
	tests := []struct {
		name, activity string
		files          map[string]string
		explanation    string
	}{
		{name: "fixed Miniflux call chains", activity: "practice-miniflux-call-chain", files: map[string]string{"trace/trace.go": minifluxTraceSolution()}},
		{name: "Miniflux category validation patch", activity: "assessment-miniflux-category-validation", files: map[string]string{
			"internal/validator/category.go":      minifluxCategorySolution(),
			"internal/validator/category_test.go": minifluxCategoryTests(),
		}, explanation: minifluxPatchExplanation("API chain")},
		{name: "Miniflux Retry-After variant", activity: "review-miniflux-retry-after", files: map[string]string{
			"internal/reader/fetcher/response_handler.go":      minifluxRetryAfterSolution(),
			"internal/reader/fetcher/response_handler_test.go": minifluxRetryAfterTests(),
			"internal/reader/handler/retry_policy.go":          minifluxRateLimitSolution(),
			"internal/reader/handler/retry_policy_test.go":     minifluxRateLimitTests(),
		}, explanation: minifluxPatchExplanation("background chain")},
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
			current := attempt.Attempt{
				ID: "00000000-0000-4000-9250-00000000000" + strconv.Itoa(index+1), ReleaseID: registry.CurrentReleaseID(),
				ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash,
				TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
				Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace),
			}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4001-9250-00000000000" + strconv.Itoa(index+1)
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
			frozen := submission.Submission{
				ID: "00000000-0000-4001-9260-00000000000" + strconv.Itoa(index+1), AttemptID: current.ID,
				ReleaseID: current.ReleaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
				Workspace: workspace, Explanation: testCase.explanation,
			}
			terminal := execution.Execution{
				ID: executionID, AttemptID: current.ID, SubmissionID: frozen.ID,
				TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
				Status: response.Status, Response: &response,
			}
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

func minifluxTraceSolution() string {
	return `package trace
type Step struct{Path,Symbol,Responsibility string}
func CategoryCreation()[]Step{return []Step{{"internal/api/api.go","POST /v1/categories","route"},{"internal/api/category_handlers.go","createCategoryHandler","decode-and-respond"},{"internal/validator/category.go","ValidateCategoryCreation","validate"},{"internal/storage/category.go","CreateCategory","persist"}}}
func FeedRefresh()[]Step{return []Step{{"internal/cli/scheduler.go","feedScheduler","schedule"},{"internal/storage/batch.go","FetchJobs","select-due-feeds"},{"internal/worker/pool.go","Push","backpressure"},{"internal/worker/worker.go","Run","consume-job"},{"internal/reader/handler/handler.go","RefreshFeed","refresh-and-store"}}}
`
}

func minifluxCategorySolution() string {
	return `package validator
import("strings";"unicode/utf8";"miniflux.app/v2/internal/locale";"miniflux.app/v2/internal/model")
const maxCategoryTitleRunes=100
type CategoryStore interface{CategoryTitleExists(int64,string)bool;AnotherCategoryExists(int64,int64,string)bool}
func normalizeCategoryTitle(title string)(string,*locale.LocalizedError){title=strings.TrimSpace(title);if title==""{return "",locale.NewLocalizedError("error.title_required")};if utf8.RuneCountInString(title)>maxCategoryTitleRunes{return "",locale.NewLocalizedError("error.title_too_long")};return title,nil}
func ValidateCategoryCreation(store CategoryStore,userID int64,request *model.CategoryCreationRequest)*locale.LocalizedError{title,err:=normalizeCategoryTitle(request.Title);if err!=nil{return err};request.Title=title;if store.CategoryTitleExists(userID,title){return locale.NewLocalizedError("error.category_already_exists")};return nil}
func ValidateCategoryModification(store CategoryStore,userID,categoryID int64,request *model.CategoryModificationRequest)*locale.LocalizedError{if request.Title==nil{return nil};title,err:=normalizeCategoryTitle(*request.Title);if err!=nil{return err};request.Title=&title;if store.AnotherCategoryExists(userID,categoryID,title){return locale.NewLocalizedError("error.category_already_exists")};return nil}
`
}

func minifluxCategoryTests() string {
	return `package validator
import("strings";"testing";"miniflux.app/v2/internal/model")
type learnerStore struct{duplicate bool}
func(s learnerStore)CategoryTitleExists(int64,string)bool{return s.duplicate}
func(s learnerStore)AnotherCategoryExists(int64,int64,string)bool{return s.duplicate}
func TestCategoryValidationCases(t *testing.T){tests:=[]struct{name,title string;duplicate,wantErr bool}{{"valid","News",false,false},{"trim"," News ",false,false},{"blank","  ",false,true},{"long",strings.Repeat("界",101),false,true},{"duplicate","News",true,true},{"unicode limit",strings.Repeat("界",100),false,false}};for _,tc:=range tests{t.Run(tc.name,func(t *testing.T){request:=&model.CategoryCreationRequest{Title:tc.title};err:=ValidateCategoryCreation(learnerStore{tc.duplicate},1,request);if(err!=nil)!=tc.wantErr{t.Fatalf("err=%v",err)}})}}
`
}

func minifluxRetryAfterSolution() string {
	return `package fetcher
import("net/http";"strconv";"time")
type ResponseHandler struct{httpResponse *http.Response}
func NewResponseHandler(response *http.Response)*ResponseHandler{return &ResponseHandler{httpResponse:response}}
func(r *ResponseHandler)ParseRetryDelay(now time.Time,maximum time.Duration)time.Duration{if r.httpResponse==nil||maximum<=0{return 0};value:=r.httpResponse.Header.Get("Retry-After");var delay time.Duration;if seconds,err:=strconv.Atoi(value);err==nil{delay=time.Duration(seconds)*time.Second}else if parsed,err:=time.Parse(time.RFC1123,value);err==nil{delay=parsed.Sub(now)};if delay<=0{return 0};if delay>maximum{return maximum};return delay}
func(r *ResponseHandler)IsRateLimited()bool{return r.httpResponse!=nil&&r.httpResponse.StatusCode==http.StatusTooManyRequests}
`
}

func minifluxRetryAfterTests() string {
	return `package fetcher
import("net/http";"testing";"time")
func TestRetryAfterCases(t *testing.T){now:=time.Date(2026,7,19,12,0,0,0,time.UTC);tests:=[]struct{name,value string;maximum,want time.Duration}{{"seconds","42",time.Minute,42*time.Second},{"cap","120",time.Minute,time.Minute},{"date",now.Add(time.Second).Format(time.RFC1123),time.Minute,time.Second},{"past",now.Add(-time.Second).Format(time.RFC1123),time.Minute,0},{"invalid","later",time.Minute,0},{"zero","0",time.Minute,0}};for _,tc:=range tests{t.Run(tc.name,func(t *testing.T){response:=&http.Response{Header:http.Header{"Retry-After":[]string{tc.value}}};if got:=NewResponseHandler(response).ParseRetryDelay(now,tc.maximum);got!=tc.want{t.Fatalf("got=%s",got)}})}}
`
}

func minifluxRateLimitSolution() string {
	return `package handler
import "time"
type RateLimitResponse interface{IsRateLimited()bool;ParseRetryDelay(time.Time,time.Duration)time.Duration}
func RateLimitDelay(response RateLimitResponse,now time.Time,maximum time.Duration)time.Duration{if !response.IsRateLimited(){return 0};return response.ParseRetryDelay(now,maximum)}
`
}

func minifluxRateLimitTests() string {
	return `package handler
import("testing";"time")
type learnerResponse struct{limited bool;delay time.Duration}
func(r learnerResponse)IsRateLimited()bool{return r.limited}
func(r learnerResponse)ParseRetryDelay(time.Time,time.Duration)time.Duration{return r.delay}
func TestRateLimitDelayCases(t *testing.T){tests:=[]struct{name string;limited bool;delay,want time.Duration}{{"ordinary",false,time.Hour,0},{"limited",true,time.Second,time.Second}};for _,tc:=range tests{t.Run(tc.name,func(t *testing.T){if got:=RateLimitDelay(learnerResponse{tc.limited,tc.delay},time.Time{},time.Minute);got!=tc.want{t.Fatalf("got=%s",got)}})}}
`
}

func minifluxPatchExplanation(chain string) string {
	return "The fixed commit is Miniflux v2.3.2 at 51f2e0d8199ea8fa305081f6e175bba64b0ef94b. I traced the " + chain + " by concrete file and symbol before changing behavior. This is a training patch, not an upstream contribution: the seam removes database or public-feed dependencies while preserving the caller direction. The test boundary covers pure policy with deterministic fakes and clocks, while the real repository would still need package and integration gates. Rollback is a forward source revert of this isolated behavior change; it does not delete migrations, attribution, or audit evidence."
}
