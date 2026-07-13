package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/assistance"
	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	learningsession "github.com/MorseWayne/gogopher-arch/internal/learning/session"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

func TestWorkflowExecuteUsesOwnerAndRedactsHeldOutDetails(t *testing.T) {
	executions := &workflowExecutionStub{value: execution.Execution{
		ID: "execution-id", AttemptID: "attempt-id", Action: execution.ActionTest,
		Status: execution.ExecutionUserFailed, Response: &execution.ExecutionResponse{
			Stages: []execution.StageResult{
				{Stage: execution.StageVisibleTest, Status: execution.StagePassed, Stdout: "visible output"},
				{Stage: execution.StageHeldOutTest, Status: execution.StageFailed, Stdout: "secret input", TestEvents: []execution.TestEvent{{Action: "fail", Package: "private", Test: "SecretTest"}}, PublicSummary: "one private check failed"},
			},
		},
	}}
	handler := newWorkflowTestHandler(t, executions, &workflowSubmissionStub{}, &workflowAssistanceStub{})
	request := authenticatedRequest(http.MethodPost, "/execute", `{"request_key":"test-1","action":"test","workspace_revision":2,"workspace_hash":"hash"}`)
	response := httptest.NewRecorder()
	handler.Execute(response, request, "attempt-id")

	if response.Code != http.StatusAccepted || executions.input.LearnerID != "learner-id" || executions.input.AttemptID != "attempt-id" {
		t.Fatalf("response=%d input=%#v", response.Code, executions.input)
	}
	body := response.Body.String()
	if !strings.Contains(body, "visible output") || !strings.Contains(body, "one private check failed") || strings.Contains(body, "secret input") || strings.Contains(body, "SecretTest") {
		t.Fatalf("public execution = %s", body)
	}
}

func TestWorkflowSubmitAndRetryReturnStableSubmission(t *testing.T) {
	submissions := &workflowSubmissionStub{result: submission.Result{
		Submission:  submission.Submission{ID: "submission-id", LatestExecutionStatus: execution.ExecutionQueued},
		ExecutionID: "execution-id", ExecutionSequence: 1, Created: true,
	}}
	handler := newWorkflowTestHandler(t, &workflowExecutionStub{}, submissions, &workflowAssistanceStub{})

	response := httptest.NewRecorder()
	handler.Submit(response, authenticatedRequest(http.MethodPost, "/submit", `{"submission_key":"submit-1","workspace_revision":3,"workspace_hash":"hash"}`), "attempt-id")
	if response.Code != http.StatusAccepted || submissions.submitInput.LearnerID != "learner-id" || submissions.submitInput.AttemptID != "attempt-id" {
		t.Fatalf("submit=%d input=%#v", response.Code, submissions.submitInput)
	}

	response = httptest.NewRecorder()
	handler.RetrySubmission(response, authenticatedRequest(http.MethodPost, "/retry", `{"request_key":"retry-1"}`), "submission-id")
	if response.Code != http.StatusAccepted || submissions.retryInput.LearnerID != "learner-id" || submissions.retryInput.SubmissionID != "submission-id" {
		t.Fatalf("retry=%d input=%#v", response.Code, submissions.retryInput)
	}
}

func TestWorkflowAssistanceRecordsAllowedEventsAndRevealsAfterCommit(t *testing.T) {
	assistanceCommands := &workflowAssistanceStub{
		recordResult: assistance.RecordResult{Created: true, Event: assistance.Event{ID: "event-id", Type: assistance.ReferenceOpened}},
		revealResult: assistance.RecordResult{Created: true, Event: assistance.Event{ID: "hint-event"}},
	}
	handler := newWorkflowTestHandler(t, &workflowExecutionStub{}, &workflowSubmissionStub{}, assistanceCommands)

	response := httptest.NewRecorder()
	handler.RecordAssistance(response, authenticatedRequest(http.MethodPost, "/events", `{"event_key":"reference-1","event_type":"reference_opened","payload":{"reference_id":"docs"}}`), "attempt-id")
	if response.Code != http.StatusCreated || assistanceCommands.recordInput.Type != assistance.ReferenceOpened {
		t.Fatalf("record=%d input=%#v", response.Code, assistanceCommands.recordInput)
	}

	response = httptest.NewRecorder()
	handler.RevealHint(response, authenticatedRequest(http.MethodPost, "/hint", `{"event_key":"hint:first"}`), "attempt-id", "first")
	if response.Code != http.StatusCreated || assistanceCommands.revealInput.Hint.Body != "Inspect the failing contract." || !strings.Contains(response.Body.String(), "Inspect the failing contract.") {
		t.Fatalf("reveal=%d input=%#v body=%s", response.Code, assistanceCommands.revealInput, response.Body.String())
	}

	assistanceCommands.revealErr = assistance.ErrAttemptInactive
	response = httptest.NewRecorder()
	handler.RevealHint(response, authenticatedRequest(http.MethodPost, "/hint", `{"event_key":"hint:late"}`), "attempt-id", "first")
	if response.Code != http.StatusConflict || strings.Contains(response.Body.String(), "Inspect the failing contract.") {
		t.Fatalf("failed reveal=%d body=%s", response.Code, response.Body.String())
	}
}

func TestLearningErrorMapsDomainConflicts(t *testing.T) {
	tests := []struct {
		err  error
		code int
		body string
	}{
		{err: execution.ErrInvalidRequest, code: http.StatusUnprocessableEntity, body: "validation_failed"},
		{err: &execution.IdempotencyConflict{ExecutionID: "existing"}, code: http.StatusConflict, body: `"execution_id":"existing"`},
		{err: &submission.AttemptAlreadySubmitted{SubmissionID: "submitted"}, code: http.StatusConflict, body: `"submission_id":"submitted"`},
		{err: submission.ErrNotFound, code: http.StatusNotFound, body: "submission_not_found"},
	}
	for _, test := range tests {
		response := httptest.NewRecorder()
		writeLearningError(response, test.err)
		if response.Code != test.code || !strings.Contains(response.Body.String(), test.body) {
			t.Fatalf("error=%v response=%d %s", test.err, response.Code, response.Body.String())
		}
	}
}

func authenticatedRequest(method, target, body string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	return request.WithContext(context.WithValue(request.Context(), sessionContextKey{}, learningsession.Session{LearnerID: "learner-id"}))
}

func newWorkflowTestHandler(t *testing.T, executions ExecutionCommands, submissions SubmissionCommands, assistanceCommands AssistanceCommands) *WorkflowHandler {
	t.Helper()
	handler, err := NewWorkflowHandler(
		executions, submissions, assistanceCommands,
		&workflowAttemptStub{value: attempt.Attempt{ReleaseID: "release", TaskID: "task", TaskVersion: 1}},
		&workflowHintStub{value: definition.HintView{ID: "first", Title: "First", Body: "Inspect the failing contract."}},
	)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

type workflowExecutionStub struct {
	input execution.CreateInput
	value execution.Execution
	err   error
}

func (s *workflowExecutionStub) Create(_ context.Context, input execution.CreateInput) (execution.Execution, error) {
	s.input = input
	return s.value, s.err
}

type workflowSubmissionStub struct {
	submitInput submission.SubmitInput
	retryInput  submission.RetryInput
	result      submission.Result
	err         error
}

func (s *workflowSubmissionStub) Submit(_ context.Context, input submission.SubmitInput) (submission.Result, error) {
	s.submitInput = input
	return s.result, s.err
}

func (s *workflowSubmissionStub) Retry(_ context.Context, input submission.RetryInput) (submission.Result, error) {
	s.retryInput = input
	return s.result, s.err
}

type workflowAssistanceStub struct {
	recordInput  assistance.RecordInput
	revealInput  assistance.RevealHintInput
	recordResult assistance.RecordResult
	revealResult assistance.RecordResult
	recordErr    error
	revealErr    error
}

func (s *workflowAssistanceStub) Record(_ context.Context, input assistance.RecordInput) (assistance.RecordResult, error) {
	s.recordInput = input
	return s.recordResult, s.recordErr
}

func (s *workflowAssistanceStub) RevealHint(_ context.Context, input assistance.RevealHintInput) (assistance.Hint, assistance.RecordResult, error) {
	s.revealInput = input
	if s.revealErr != nil {
		return assistance.Hint{}, assistance.RecordResult{}, s.revealErr
	}
	return input.Hint, s.revealResult, nil
}

type workflowAttemptStub struct {
	value attempt.Attempt
	err   error
}

func (s *workflowAttemptStub) Get(context.Context, string, string) (attempt.Attempt, error) {
	return s.value, s.err
}

type workflowHintStub struct {
	value definition.HintView
	err   error
}

func (s *workflowHintStub) Hint(string, string, int, string) (definition.HintView, error) {
	return s.value, s.err
}
