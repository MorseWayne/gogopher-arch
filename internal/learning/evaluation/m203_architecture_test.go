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

func TestM203LearningLoopSolutionsPassRealReleaseAndSandboxEvaluation(t *testing.T) {
	tests := []struct {
		name        string
		activityID  string
		solution    map[string]string
		explanation string
	}{
		{
			name:       "guided gocheck layering",
			activityID: "practice-gocheck-architecture",
			solution: map[string]string{
				"internal/checks/service.go":           gocheckServiceSolution(),
				"internal/checks/memory/repository.go": gocheckMemorySolution(),
			},
			explanation: "先让 use case 定义最小 Repository，再让 memory storage 实现它，并通过 constructor 注入替身或真实依赖。",
		},
		{
			name:       "independent gocheck architecture",
			activityID: "assessment-gocheck-architecture",
			solution: map[string]string{
				"internal/checks/service.go":           gocheckServiceSolution(),
				"internal/checks/memory/repository.go": gocheckMemorySolution(),
				"internal/httpapi/handler.go":          gocheckTransportSolution(),
				"cmd/gocheckhub/main.go":               gocheckWiringSolution(),
				"internal/checks/service_test.go":      gocheckBoundaryTestSolution(),
			},
			explanation: "transport 只把严格 JSON 翻译为 use case 输入并映射稳定错误，不选择 concrete storage。use case 在消费方定义单方法 Repository，constructor 显式接收它和 ID 生成器，因此测试可传替身。memory storage 反向依赖业务契约并实现接口，cmd 中的 constructor 调用才决定具体实现和组装顺序，依赖始终朝业务边界流动。",
		},
		{
			name:       "delayed gocheck alert variant",
			activityID: "review-gocheck-alert-architecture",
			solution: map[string]string{
				"internal/alerts/manager.go":      alertManagerSolution(),
				"internal/alerts/memory/store.go": alertMemorySolution(),
				"internal/alertapi/handler.go":    alertTransportSolution(),
				"cmd/gocheckalerts/main.go":       alertWiringSolution(),
				"internal/alerts/manager_test.go": alertBoundaryTestSolution(),
			},
			explanation: "alert transport 只处理请求响应并通过自己消费的 Publisher 调用 use case；业务层定义所需 Store，不反向认识 memory storage。两个 constructor 都拒绝缺失依赖，使替身和真实实现走同一条路径；最后只有 cmd composition root 选择 memory，并按 storage、use case、transport 顺序调用 constructor，从而避免全局状态和跨层循环依赖。",
		},
	}

	registry := draftReleaseRegistry(t)
	for index, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			activity, err := registry.ActivityView(registry.CurrentReleaseID(), testCase.activityID, 1)
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
			for path, source := range testCase.solution {
				workspace[path] = source
			}
			current := attempt.Attempt{
				ID: "00000000-0000-4000-9600-00000000000" + strconv.Itoa(index+1), ReleaseID: registry.CurrentReleaseID(),
				ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash,
				TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
				Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace),
			}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4000-9700-00000000000" + strconv.Itoa(index+1)
			specification, err := builder.Build(current, executionID, execution.ActionSubmit)
			if err != nil {
				t.Fatal(err)
			}
			response, err := sandbox.NewRunner(sandbox.RunnerOptions{TempDir: t.TempDir()}).Run(context.Background(), specification)
			if err != nil {
				t.Fatal(err)
			}
			if response.Status != execution.ExecutionSucceeded {
				t.Fatalf("sandbox response = %#v", response)
			}
			frozen := submission.Submission{
				ID: "00000000-0000-4000-9800-00000000000" + strconv.Itoa(index+1), AttemptID: current.ID,
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

func gocheckServiceSolution() string {
	return `package checks

import (
	"context"
	"errors"
	"strings"
)

var (
	ErrInvalidTarget = errors.New("invalid target")
	ErrCheckExists = errors.New("check already exists")
)
type NewCheck struct { Target string }
type Check struct { ID string; Target string }
type Repository interface { Create(context.Context, Check) error }
type Service struct { repository Repository; nextID func() string }

func NewService(repository Repository, nextID func() string) (*Service, error) {
	if repository == nil || nextID == nil { return nil, errors.New("repository and nextID are required") }
	return &Service{repository: repository, nextID: nextID}, nil
}
func (s *Service) Create(ctx context.Context, input NewCheck) (Check, error) {
	target := strings.TrimSpace(input.Target)
	if target == "" { return Check{}, ErrInvalidTarget }
	created := Check{ID: s.nextID(), Target: target}
	if err := s.repository.Create(ctx, created); err != nil { return Check{}, err }
	return created, nil
}
`
}

func gocheckMemorySolution() string {
	return `package memory

import (
	"context"
	"strings"
	"sync"

	"gocheckhub/internal/checks"
)
type Repository struct { mu sync.Mutex; checks map[string]checks.Check }
func NewRepository() *Repository { return &Repository{checks: make(map[string]checks.Check)} }
func (r *Repository) Create(ctx context.Context, check checks.Check) error {
	select { case <-ctx.Done(): return ctx.Err(); default: }
	key := strings.ToLower(check.Target)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.checks[key]; exists { return checks.ErrCheckExists }
	r.checks[key] = check
	return nil
}
`
}

func gocheckTransportSolution() string {
	return `package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"gocheckhub/internal/checks"
)
type Creator interface { Create(context.Context, checks.NewCheck) (checks.Check, error) }
type createRequest struct { Target string ` + "`json:\"target\"`" + ` }
type checkResponse struct { ID string ` + "`json:\"id\"`" + `; Target string ` + "`json:\"target\"`" + ` }
func NewHandler(creator Creator) (http.Handler, error) {
	if creator == nil { return nil, errors.New("creator is required") }
	mux := http.NewServeMux()
	mux.HandleFunc("POST /checks", func(w http.ResponseWriter, r *http.Request) {
		var request createRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if decoder.Decode(&request) != nil || decoder.Decode(&struct{}{}) != io.EOF || strings.TrimSpace(request.Target) == "" {
			writeError(w, http.StatusBadRequest, "invalid_request", "request is invalid"); return
		}
		created, err := creator.Create(r.Context(), checks.NewCheck{Target: request.Target})
		switch {
		case errors.Is(err, checks.ErrInvalidTarget): writeError(w, 400, "invalid_request", "request is invalid")
		case errors.Is(err, checks.ErrCheckExists): writeError(w, 409, "check_exists", "check already exists")
		case err != nil: writeError(w, 500, "internal_error", "internal server error")
		default: writeJSON(w, 201, checkResponse{ID: created.ID, Target: created.Target})
		}
	})
	return mux, nil
}
func writeError(w http.ResponseWriter, status int, code, message string) { writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}}) }
func writeJSON(w http.ResponseWriter, status int, body any) { w.Header().Set("Content-Type", "application/json; charset=utf-8"); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(body) }
`
}

func gocheckWiringSolution() string {
	return `package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"gocheckhub/internal/checks"
	"gocheckhub/internal/checks/memory"
	"gocheckhub/internal/httpapi"
)
func buildHandler() (http.Handler, error) {
	repository := memory.NewRepository()
	var sequence atomic.Uint64
	service, err := checks.NewService(repository, func() string { return fmt.Sprintf("check-%d", sequence.Add(1)) })
	if err != nil { return nil, err }
	return httpapi.NewHandler(service)
}
func main() { handler, err := buildHandler(); if err != nil { log.Fatal(err) }; log.Fatal(http.ListenAndServe(":8080", handler)) }
`
}

func gocheckBoundaryTestSolution() string {
	return `package checks

import "testing"
func TestServiceCreate(t *testing.T) {
	tests := []struct { name string; target string }{
		{name: "valid", target: "api"},
		{name: "trimmed", target: " api "},
		{name: "invalid", target: " "},
	}
	for _, testCase := range tests { t.Run(testCase.name, func(t *testing.T) { _ = testCase.target }) }
}
`
}

func alertManagerSolution() string {
	return `package alerts

import (
	"context"
	"errors"
	"strings"
)
var (
	ErrInvalidDestination = errors.New("invalid destination")
	ErrRuleExists = errors.New("alert rule already exists")
)
type NewRule struct { Destination string }
type Rule struct { ID string; Destination string }
type Store interface { Save(context.Context, Rule) error }
type Manager struct { store Store; nextID func() string }
func NewManager(store Store, nextID func() string) (*Manager, error) {
	if store == nil || nextID == nil { return nil, errors.New("store and nextID are required") }
	return &Manager{store: store, nextID: nextID}, nil
}
func (m *Manager) Publish(ctx context.Context, input NewRule) (Rule, error) {
	destination := strings.TrimSpace(input.Destination)
	if destination == "" { return Rule{}, ErrInvalidDestination }
	created := Rule{ID: m.nextID(), Destination: destination}
	if err := m.store.Save(ctx, created); err != nil { return Rule{}, err }
	return created, nil
}
`
}

func alertMemorySolution() string {
	return `package memory

import (
	"context"
	"strings"
	"sync"

	"gocheckhub/internal/alerts"
)
type Store struct { mu sync.Mutex; rules map[string]alerts.Rule }
func NewStore() *Store { return &Store{rules: make(map[string]alerts.Rule)} }
func (s *Store) Save(ctx context.Context, rule alerts.Rule) error {
	select { case <-ctx.Done(): return ctx.Err(); default: }
	key := strings.ToLower(rule.Destination)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.rules[key]; exists { return alerts.ErrRuleExists }
	s.rules[key] = rule
	return nil
}
`
}

func alertTransportSolution() string {
	return `package alertapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"gocheckhub/internal/alerts"
)
type Publisher interface { Publish(context.Context, alerts.NewRule) (alerts.Rule, error) }
type publishRequest struct { Destination string ` + "`json:\"destination\"`" + ` }
type alertResponse struct { ID string ` + "`json:\"id\"`" + `; Destination string ` + "`json:\"destination\"`" + ` }
func NewHandler(publisher Publisher) (http.Handler, error) {
	if publisher == nil { return nil, errors.New("publisher is required") }
	mux := http.NewServeMux()
	mux.HandleFunc("POST /alerts", func(w http.ResponseWriter, r *http.Request) {
		var request publishRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if decoder.Decode(&request) != nil || decoder.Decode(&struct{}{}) != io.EOF || strings.TrimSpace(request.Destination) == "" {
			writeError(w, 400, "invalid_request", "request is invalid"); return
		}
		created, err := publisher.Publish(r.Context(), alerts.NewRule{Destination: request.Destination})
		switch {
		case errors.Is(err, alerts.ErrInvalidDestination): writeError(w, 400, "invalid_request", "request is invalid")
		case errors.Is(err, alerts.ErrRuleExists): writeError(w, 409, "alert_exists", "alert rule already exists")
		case err != nil: writeError(w, 500, "internal_error", "internal server error")
		default: writeJSON(w, 201, alertResponse{ID: created.ID, Destination: created.Destination})
		}
	})
	return mux, nil
}
func writeError(w http.ResponseWriter, status int, code, message string) { writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}}) }
func writeJSON(w http.ResponseWriter, status int, body any) { w.Header().Set("Content-Type", "application/json; charset=utf-8"); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(body) }
`
}

func alertWiringSolution() string {
	return `package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"gocheckhub/internal/alertapi"
	"gocheckhub/internal/alerts"
	"gocheckhub/internal/alerts/memory"
)
func buildHandler() (http.Handler, error) {
	store := memory.NewStore()
	var sequence atomic.Uint64
	manager, err := alerts.NewManager(store, func() string { return fmt.Sprintf("alert-%d", sequence.Add(1)) })
	if err != nil { return nil, err }
	return alertapi.NewHandler(manager)
}
func main() { handler, err := buildHandler(); if err != nil { log.Fatal(err) }; log.Fatal(http.ListenAndServe(":8080", handler)) }
`
}

func alertBoundaryTestSolution() string {
	return `package alerts

import "testing"
func TestManagerPublish(t *testing.T) {
	tests := []struct { name string; destination string }{
		{name: "valid", destination: "ops@example.com"},
		{name: "trimmed", destination: " ops@example.com "},
		{name: "invalid", destination: " "},
	}
	for _, testCase := range tests { t.Run(testCase.name, func(t *testing.T) { _ = testCase.destination }) }
}
`
}
