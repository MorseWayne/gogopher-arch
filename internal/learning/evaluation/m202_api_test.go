package evaluation

import (
	"strconv"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

func TestM202LearningLoopSolutionsPassRealReleaseAndSandboxEvaluation(t *testing.T) {
	tests := []struct {
		name        string
		activityID  string
		solution    map[string]string
		explanation string
	}{
		{
			name:       "independent checks API",
			activityID: "assessment-gocheck-api",
			solution: map[string]string{
				"checkapi/handler.go":      gocheckAPISolution(),
				"checkapi/handler_test.go": gocheckAPITableTestSolution(),
			},
			explanation: "我先用严格 JSON 解码拒绝未知字段和多余值，再校验 target 与 timeout。边界输入统一返回 400，领域冲突映射为 409，未知依赖错误只返回固定的 500 协议，避免把内部错误泄漏给客户端。",
		},
		{
			name:       "delayed alert API variant",
			activityID: "review-alert-api",
			solution: map[string]string{
				"alertapi/handler.go":      alertAPISolution(),
				"alertapi/handler_test.go": alertAPITableTestSolution(),
			},
			explanation: "我把传输层 JSON DTO、输入校验和 Creator 调用分开：非法 URL 或 threshold 返回 400，重复规则返回 409，未识别的服务错误返回不含原始错误文本的 500；所有分支都保持同一个 JSON 错误包络。",
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
				ID: "00000000-0000-4000-9300-00000000000" + strconv.Itoa(index+1), ReleaseID: registry.CurrentReleaseID(),
				ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash,
				TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
				Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace),
			}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4000-9400-00000000000" + strconv.Itoa(index+1)
			specification, err := builder.Build(current, executionID, execution.ActionSubmit)
			if err != nil {
				t.Fatal(err)
			}
			response, err := runRegressionSandbox(t, specification)
			if err != nil {
				t.Fatal(err)
			}
			if response.Status != execution.ExecutionSucceeded {
				t.Fatalf("sandbox response = %#v", response)
			}
			frozen := submission.Submission{
				ID: "00000000-0000-4000-9500-00000000000" + strconv.Itoa(index+1), AttemptID: current.ID,
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

func gocheckAPISolution() string {
	return `package checkapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

var ErrCheckExists = errors.New("check already exists")

type NewCheck struct { Target string; TimeoutMS int }
type Check struct { ID, Target string; TimeoutMS int }
type Creator interface { Create(context.Context, NewCheck) (Check, error) }

type createCheckRequest struct {
	Target string ` + "`json:\"target\"`" + `
	TimeoutMS int ` + "`json:\"timeout_ms\"`" + `
}

type checkResponse struct {
	ID string ` + "`json:\"id\"`" + `
	Target string ` + "`json:\"target\"`" + `
	TimeoutMS int ` + "`json:\"timeout_ms\"`" + `
}

func NewHandler(creator Creator) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /checks", func(w http.ResponseWriter, r *http.Request) {
		var request createCheckRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if decoder.Decode(&request) != nil || decoder.Decode(&struct{}{}) != io.EOF {
			writeError(w, http.StatusBadRequest, "invalid_request", "request body is invalid")
			return
		}
		request.Target = strings.TrimSpace(request.Target)
		if request.Target == "" || request.TimeoutMS < 1 || request.TimeoutMS > 60000 {
			writeError(w, http.StatusBadRequest, "invalid_request", "request fields are invalid")
			return
		}
		created, err := creator.Create(r.Context(), NewCheck{Target: request.Target, TimeoutMS: request.TimeoutMS})
		if errors.Is(err, ErrCheckExists) {
			writeError(w, http.StatusConflict, "check_exists", "check already exists")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
			return
		}
		writeJSON(w, http.StatusCreated, checkResponse{ID: created.ID, Target: created.Target, TimeoutMS: created.TimeoutMS})
	})
	return mux
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
`
}

func gocheckAPITableTestSolution() string {
	return `package checkapi

import "testing"

func TestHandlerContract(t *testing.T) {
	tests := []struct { name string }{
		{name: "success"},
		{name: "invalid request"},
		{name: "domain conflict"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {})
	}
}
`
}

func alertAPISolution() string {
	return `package alertapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

var ErrRuleExists = errors.New("rule already exists")
type NewRule struct { Destination string; Threshold int }
type Rule struct { ID, Destination string; Threshold int }
type Creator interface { Create(context.Context, NewRule) (Rule, error) }

type createRuleRequest struct {
	Destination string ` + "`json:\"destination\"`" + `
	Threshold int ` + "`json:\"threshold\"`" + `
}

type ruleResponse struct {
	ID string ` + "`json:\"id\"`" + `
	Destination string ` + "`json:\"destination\"`" + `
	Threshold int ` + "`json:\"threshold\"`" + `
}

func NewHandler(creator Creator) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /rules", func(w http.ResponseWriter, r *http.Request) {
		var request createRuleRequest
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		if decoder.Decode(&request) != nil || decoder.Decode(&struct{}{}) != io.EOF || !validDestination(request.Destination) || request.Threshold < 1 || request.Threshold > 100 {
			writeError(w, http.StatusBadRequest, "invalid_request", "request is invalid")
			return
		}
		created, err := creator.Create(r.Context(), NewRule{Destination: request.Destination, Threshold: request.Threshold})
		if errors.Is(err, ErrRuleExists) {
			writeError(w, http.StatusConflict, "rule_exists", "rule already exists")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
			return
		}
		writeJSON(w, http.StatusCreated, ruleResponse{ID: created.ID, Destination: created.Destination, Threshold: created.Threshold})
	})
	return mux
}

func validDestination(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
`
}

func alertAPITableTestSolution() string {
	return `package alertapi

import "testing"

func TestRuleAPI(t *testing.T) {
	tests := []struct { name string }{
		{name: "success"},
		{name: "invalid destination"},
		{name: "rule conflict"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {})
	}
}
`
}
