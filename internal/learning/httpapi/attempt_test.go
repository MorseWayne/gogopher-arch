package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/assistance"
	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/attemptview"
	"github.com/MorseWayne/gogopher-arch/internal/learning/evaluation"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	learningsession "github.com/MorseWayne/gogopher-arch/internal/learning/session"
)

func TestAttemptHandlerCreatesForAuthenticatedOwner(t *testing.T) {
	service := &attemptServiceStub{created: attempt.Attempt{ID: "attempt-id", LearnerID: "owner", Status: "active", Workspace: map[string]string{}, CapabilityRefs: nil}}
	handler, _ := NewAttemptHandler(service, nil)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/learning/attempts", strings.NewReader(`{"activity_id":"assessment-check-config","activity_version":1}`))
	request = request.WithContext(context.WithValue(request.Context(), sessionContextKey{}, learningsession.Session{LearnerID: "owner"}))
	response := httptest.NewRecorder()
	handler.Create(response, request)
	if response.Code != http.StatusCreated || service.createInput.LearnerID != "owner" {
		t.Fatalf("response=%d input=%#v", response.Code, service.createInput)
	}
	if !strings.Contains(response.Body.String(), `"executions":[]`) || !strings.Contains(response.Body.String(), `"evidence":[]`) {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestAttemptHandlerResumesOpenAttemptWithoutCreatingAnotherFact(t *testing.T) {
	resumed := attempt.CreateResult{Attempt: attempt.Attempt{
		ID: "attempt-existing", LearnerID: "owner", Status: "active", Workspace: map[string]string{},
	}}
	service := &attemptServiceStub{createdResult: &resumed}
	observer := &attemptObserverStub{}
	handler, _ := NewAttemptHandler(service, nil, observer)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/learning/attempts", strings.NewReader(`{"activity_id":"guided-run-model","activity_version":5}`))
	request = request.WithContext(context.WithValue(request.Context(), sessionContextKey{}, learningsession.Session{LearnerID: "owner"}))
	response := httptest.NewRecorder()

	handler.Create(response, request)

	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"id":"attempt-existing"`) {
		t.Fatalf("response=%d %s", response.Code, response.Body.String())
	}
	if observer.created != 0 {
		t.Fatalf("AttemptCreated calls = %d, want 0", observer.created)
	}
}

func TestAttemptHandlerRequiresReviewItemClaim(t *testing.T) {
	service := &attemptServiceStub{createErr: attempt.ErrReviewClaimRequired}
	handler, _ := NewAttemptHandler(service, nil)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/learning/attempts", strings.NewReader(`{"activity_id":"review-check-config-variant","activity_version":3}`))
	request = request.WithContext(context.WithValue(request.Context(), sessionContextKey{}, learningsession.Session{LearnerID: "owner"}))
	response := httptest.NewRecorder()
	handler.Create(response, request)
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "review_claim_required") {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func TestAttemptHandlerReturnsPublicRelatedState(t *testing.T) {
	service := &attemptServiceStub{got: attempt.Attempt{ID: "attempt-id", LearnerID: "owner", Mode: "assessment", Workspace: map[string]string{}}}
	details := &attemptDetailsStub{related: attemptview.Related{
		Submission: &attemptview.Submission{ID: "submission-id", Status: "evaluated"},
		Assistance: []assistance.Event{{
			ID: "assistance-id", AttemptID: "attempt-id", EventKey: "hint:trace-contract",
			Sequence: 1, Type: assistance.HintRevealed, Payload: []byte(`{"hint_id":"trace-contract"}`),
		}},
		Executions: []execution.Execution{{
			ID: "execution-id", AttemptID: "attempt-id", Status: execution.ExecutionSucceeded,
			Response: &execution.ExecutionResponse{Stages: []execution.StageResult{{
				Stage: execution.StageHeldOutTest, Status: execution.StagePassed,
				Stdout: "hidden output", TestEvents: []execution.TestEvent{{Action: "pass", Test: "HiddenCase"}},
				PublicSummary: "private checks passed",
			}}},
		}},
		RuleResults: []execution.RuleResult{{
			RuleID: "held-out-tests-pass", Status: execution.RulePassed, Stage: execution.StageHeldOutTest,
			Package: "private/package", Test: "HiddenCase", Summary: "passed", ExecutionID: "execution-id",
		}},
		Evidence: []evaluation.Evidence{{
			ID: "evidence-id", CapabilityID: "M1-09", CapabilityVersion: 1,
			Result: execution.RulePassed, Independence: assistance.IndependenceHinted,
		}},
	}}
	handler, _ := NewAttemptHandler(service, details)
	ctx := context.WithValue(context.Background(), sessionContextKey{}, learningsession.Session{LearnerID: "owner"})
	response := httptest.NewRecorder()
	handler.Get(response, httptest.NewRequest(http.MethodGet, "/attempts/attempt-id", nil).WithContext(ctx), "attempt-id")

	body := response.Body.String()
	if response.Code != http.StatusOK || !strings.Contains(body, `"submission":{"id":"submission-id"`) ||
		!strings.Contains(body, `"level":"hinted"`) || !strings.Contains(body, `"event_key":"hint:trace-contract"`) ||
		!strings.Contains(body, `"payload":{"hint_id":"trace-contract"}`) ||
		!strings.Contains(body, `"capability_id":"M1-09"`) || !strings.Contains(body, "private checks passed") ||
		strings.Contains(body, "learner_id") || strings.Contains(body, "hidden output") || strings.Contains(body, "HiddenCase") || strings.Contains(body, "private/package") {
		t.Fatalf("detail response = %d %s", response.Code, body)
	}
}

func TestAttemptHandlerReturnsEmptyRelatedCollectionsAsArrays(t *testing.T) {
	service := &attemptServiceStub{got: attempt.Attempt{ID: "attempt-id", LearnerID: "owner", Mode: "guided", Workspace: map[string]string{}}}
	handler, _ := NewAttemptHandler(service, &attemptDetailsStub{})
	ctx := context.WithValue(context.Background(), sessionContextKey{}, learningsession.Session{LearnerID: "owner"})
	response := httptest.NewRecorder()
	handler.Get(response, httptest.NewRequest(http.MethodGet, "/attempts/attempt-id", nil).WithContext(ctx), "attempt-id")

	body := response.Body.String()
	for _, collection := range []string{`"events":[]`, `"executions":[]`, `"rule_results":[]`, `"evidence":[]`} {
		if !strings.Contains(body, collection) {
			t.Fatalf("detail response collection %s = %d %s", collection, response.Code, body)
		}
	}
}

func TestAttemptHandlerMapsOwnershipAndRevisionConflict(t *testing.T) {
	service := &attemptServiceStub{getErr: attempt.ErrNotFound, saveErr: &attempt.RevisionConflict{Revision: 3, Hash: "current-hash"}}
	handler, _ := NewAttemptHandler(service, nil)
	ctx := context.WithValue(context.Background(), sessionContextKey{}, learningsession.Session{LearnerID: "owner"})
	request := httptest.NewRequest(http.MethodGet, "/attempts/id", nil).WithContext(ctx)
	response := httptest.NewRecorder()
	handler.Get(response, request, "id")
	if response.Code != http.StatusNotFound {
		t.Fatalf("Get status = %d", response.Code)
	}
	request = httptest.NewRequest(http.MethodPut, "/attempts/id/workspace", strings.NewReader(`{"base_revision":0,"files":{"main.go":"package main"}}`)).WithContext(ctx)
	response = httptest.NewRecorder()
	handler.SaveWorkspace(response, request, "id")
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"current_revision":3`) || !strings.Contains(response.Body.String(), `"current_hash":"current-hash"`) {
		t.Fatalf("conflict = %d %s", response.Code, response.Body.String())
	}
}

func TestDisabledRouterReturnsExplicitUnavailableAndRemovesExecuteAPI(t *testing.T) {
	router := NewRouter(false, nil, nil, nil, nil, nil, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/learning/session", nil))
	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), "learning_disabled") {
		t.Fatalf("disabled = %d %s", response.Code, response.Body.String())
	}
	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/execute", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("legacy execute status = %d", response.Code)
	}
}

type attemptServiceStub struct {
	created       attempt.Attempt
	createdResult *attempt.CreateResult
	createInput   attempt.CreateInput
	createErr     error
	got           attempt.Attempt
	getErr        error
	saved         attempt.Attempt
	saveErr       error
}

type attemptDetailsStub struct {
	related attemptview.Related
	err     error
}

func (s *attemptDetailsStub) Load(context.Context, string, string) (attemptview.Related, error) {
	return s.related, s.err
}

func (s *attemptServiceStub) Create(_ context.Context, input attempt.CreateInput) (attempt.CreateResult, error) {
	s.createInput = input
	if s.createdResult != nil {
		return *s.createdResult, s.createErr
	}
	return attempt.CreateResult{Attempt: s.created, Created: true}, s.createErr
}
func (s *attemptServiceStub) Get(context.Context, string, string) (attempt.Attempt, error) {
	return s.got, s.getErr
}
func (s *attemptServiceStub) Save(context.Context, attempt.SaveInput) (attempt.Attempt, error) {
	return s.saved, s.saveErr
}

type attemptObserverStub struct{ created int }

func (s *attemptObserverStub) AttemptCreated(attempt.Attempt) { s.created++ }
