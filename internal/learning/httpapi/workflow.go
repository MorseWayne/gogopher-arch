package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/assistance"
	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

type ExecutionCommands interface {
	Create(context.Context, execution.CreateInput) (execution.Execution, error)
}

type SubmissionCommands interface {
	Submit(context.Context, submission.SubmitInput) (submission.Result, error)
	Retry(context.Context, submission.RetryInput) (submission.Result, error)
}

type AssistanceCommands interface {
	Record(context.Context, assistance.RecordInput) (assistance.RecordResult, error)
	RevealHint(context.Context, assistance.RevealHintInput) (assistance.Hint, assistance.RecordResult, error)
}

type AttemptLookup interface {
	Get(context.Context, string, string) (attempt.Attempt, error)
}

type HintLookup interface {
	Hint(string, string, int, string) (definition.HintView, error)
}

type WorkflowObserver interface {
	SubmissionQueued(submission.Result)
}

type WorkflowHandler struct {
	executions  ExecutionCommands
	submissions SubmissionCommands
	assistance  AssistanceCommands
	attempts    AttemptLookup
	hints       HintLookup
	observer    WorkflowObserver
}

func NewWorkflowHandler(executions ExecutionCommands, submissions SubmissionCommands, assistanceCommands AssistanceCommands, attempts AttemptLookup, hints HintLookup, observers ...WorkflowObserver) (*WorkflowHandler, error) {
	if executions == nil || submissions == nil || assistanceCommands == nil || attempts == nil || hints == nil {
		return nil, fmt.Errorf("learning workflow services and hint lookup are required")
	}
	handler := &WorkflowHandler{
		executions: executions, submissions: submissions, assistance: assistanceCommands,
		attempts: attempts, hints: hints,
	}
	if len(observers) > 0 {
		handler.observer = observers[0]
	}
	return handler, nil
}

func (h *WorkflowHandler) Execute(w http.ResponseWriter, request *http.Request, attemptID string) {
	learnerID, ok := requestLearner(w, request)
	if !ok {
		return
	}
	var body struct {
		RequestKey        string           `json:"request_key"`
		Action            execution.Action `json:"action"`
		WorkspaceRevision int64            `json:"workspace_revision"`
		WorkspaceHash     string           `json:"workspace_hash"`
	}
	if decodeJSON(request, &body) != nil || body.RequestKey == "" || body.WorkspaceRevision < 0 || body.WorkspaceHash == "" {
		writeError(w, http.StatusUnprocessableEntity, "validation_failed", "request_key, action, workspace_revision, and workspace_hash are required")
		return
	}
	value, err := h.executions.Create(request.Context(), execution.CreateInput{
		LearnerID: learnerID, AttemptID: attemptID, Action: body.Action, RequestKey: body.RequestKey,
		WorkspaceRevision: body.WorkspaceRevision, WorkspaceHash: body.WorkspaceHash,
	})
	if err != nil {
		writeLearningError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, executionResponse(value))
}

func (h *WorkflowHandler) Submit(w http.ResponseWriter, request *http.Request, attemptID string) {
	learnerID, ok := requestLearner(w, request)
	if !ok {
		return
	}
	var body struct {
		SubmissionKey     string `json:"submission_key"`
		WorkspaceRevision int64  `json:"workspace_revision"`
		WorkspaceHash     string `json:"workspace_hash"`
		Explanation       string `json:"explanation"`
	}
	if decodeJSON(request, &body) != nil || body.SubmissionKey == "" || body.WorkspaceRevision < 0 || body.WorkspaceHash == "" {
		writeError(w, http.StatusUnprocessableEntity, "validation_failed", "submission_key, workspace_revision, and workspace_hash are required")
		return
	}
	result, err := h.submissions.Submit(request.Context(), submission.SubmitInput{
		LearnerID: learnerID, AttemptID: attemptID, SubmissionKey: body.SubmissionKey,
		WorkspaceRevision: body.WorkspaceRevision, WorkspaceHash: body.WorkspaceHash,
		Explanation: body.Explanation,
	})
	if err != nil {
		writeLearningError(w, err)
		return
	}
	if h.observer != nil {
		h.observer.SubmissionQueued(result)
	}
	status := http.StatusOK
	if result.Created {
		status = http.StatusAccepted
	}
	writeJSON(w, status, submissionResponse(result))
}

func (h *WorkflowHandler) RetrySubmission(w http.ResponseWriter, request *http.Request, submissionID string) {
	learnerID, ok := requestLearner(w, request)
	if !ok {
		return
	}
	var body struct {
		RequestKey string `json:"request_key"`
	}
	if decodeJSON(request, &body) != nil || body.RequestKey == "" {
		writeError(w, http.StatusUnprocessableEntity, "validation_failed", "request_key is required")
		return
	}
	result, err := h.submissions.Retry(request.Context(), submission.RetryInput{
		LearnerID: learnerID, SubmissionID: submissionID, RequestKey: body.RequestKey,
	})
	if err != nil {
		writeLearningError(w, err)
		return
	}
	status := http.StatusOK
	if result.Created {
		status = http.StatusAccepted
	}
	writeJSON(w, status, submissionResponse(result))
}

func (h *WorkflowHandler) RecordAssistance(w http.ResponseWriter, request *http.Request, attemptID string) {
	learnerID, ok := requestLearner(w, request)
	if !ok {
		return
	}
	var body struct {
		EventKey string               `json:"event_key"`
		Type     assistance.EventType `json:"event_type"`
		Payload  map[string]any       `json:"payload"`
	}
	if decodeJSON(request, &body) != nil || body.EventKey == "" ||
		(body.Type != assistance.ReferenceOpened && body.Type != assistance.SolutionViewed && body.Type != assistance.AIDeclared) {
		writeError(w, http.StatusUnprocessableEntity, "validation_failed", "event_key and a recordable event_type are required")
		return
	}
	result, err := h.assistance.Record(request.Context(), assistance.RecordInput{
		LearnerID: learnerID, AttemptID: attemptID, EventKey: body.EventKey, Type: body.Type, Payload: body.Payload,
	})
	if err != nil {
		writeLearningError(w, err)
		return
	}
	status := http.StatusOK
	if result.Created {
		status = http.StatusCreated
	}
	writeJSON(w, status, map[string]any{"api_version": APIVersion, "event": result.Event})
}

func (h *WorkflowHandler) RevealHint(w http.ResponseWriter, request *http.Request, attemptID, hintID string) {
	learnerID, ok := requestLearner(w, request)
	if !ok {
		return
	}
	var body struct {
		EventKey string `json:"event_key"`
	}
	if decodeJSON(request, &body) != nil || body.EventKey == "" || hintID == "" {
		writeError(w, http.StatusUnprocessableEntity, "validation_failed", "event_key and hint ID are required")
		return
	}
	current, err := h.attempts.Get(request.Context(), learnerID, attemptID)
	if err != nil {
		writeLearningError(w, err)
		return
	}
	hint, err := h.hints.Hint(current.ReleaseID, current.TaskID, current.TaskVersion, hintID)
	if err != nil {
		writeLearningError(w, err)
		return
	}
	revealed, result, err := h.assistance.RevealHint(request.Context(), assistance.RevealHintInput{
		LearnerID: learnerID, AttemptID: attemptID, EventKey: body.EventKey,
		Hint: assistance.Hint{ID: hint.ID, Title: hint.Title, Body: hint.Body},
	})
	if err != nil {
		writeLearningError(w, err)
		return
	}
	status := http.StatusOK
	if result.Created {
		status = http.StatusCreated
	}
	writeJSON(w, status, map[string]any{"api_version": APIVersion, "hint": revealed, "event": result.Event})
}

func requestLearner(w http.ResponseWriter, request *http.Request) (string, bool) {
	owner, ok := SessionFromContext(request.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "learning session is required")
		return "", false
	}
	return owner.LearnerID, true
}

type executionDTO struct {
	APIVersion        string                    `json:"api_version"`
	ID                string                    `json:"id"`
	AttemptID         string                    `json:"attempt_id"`
	SubmissionID      string                    `json:"submission_id,omitempty"`
	Action            execution.Action          `json:"action"`
	Sequence          int                       `json:"sequence"`
	Status            execution.ExecutionStatus `json:"status"`
	WorkspaceRevision int64                     `json:"workspace_revision"`
	WorkspaceHash     string                    `json:"workspace_hash"`
	Stages            []stageDTO                `json:"stages"`
	Failure           *executionFailureDTO      `json:"failure,omitempty"`
	StartedAt         *time.Time                `json:"started_at,omitempty"`
	FinishedAt        *time.Time                `json:"finished_at,omitempty"`
	CreatedAt         time.Time                 `json:"created_at"`
	UpdatedAt         time.Time                 `json:"updated_at"`
}

type stageDTO struct {
	Stage           execution.Stage       `json:"stage"`
	Status          execution.StageStatus `json:"status"`
	ExitCode        int                   `json:"exit_code"`
	Stdout          string                `json:"stdout,omitempty"`
	Stderr          string                `json:"stderr,omitempty"`
	DurationMS      int64                 `json:"duration_ms"`
	TimedOut        bool                  `json:"timed_out"`
	OutputTruncated bool                  `json:"output_truncated"`
	PublicSummary   string                `json:"public_summary,omitempty"`
	TestEvents      []execution.TestEvent `json:"test_events,omitempty"`
}

type executionFailureDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func executionResponse(value execution.Execution) executionDTO {
	result := executionDTO{
		APIVersion: APIVersion, ID: value.ID, AttemptID: value.AttemptID, SubmissionID: value.SubmissionID,
		Action: value.Action, Sequence: value.Sequence, Status: value.Status,
		WorkspaceRevision: value.WorkspaceRevision, WorkspaceHash: value.WorkspaceHash,
		Stages: []stageDTO{}, StartedAt: value.StartedAt, FinishedAt: value.FinishedAt,
		CreatedAt: value.CreatedAt, UpdatedAt: value.UpdatedAt,
	}
	if value.Response == nil {
		return result
	}
	for _, stage := range value.Response.Stages {
		public := stageDTO{
			Stage: stage.Stage, Status: stage.Status, ExitCode: stage.ExitCode, DurationMS: stage.DurationMS,
			TimedOut: stage.TimedOut, OutputTruncated: stage.OutputTruncated, PublicSummary: stage.PublicSummary,
		}
		if stage.Stage != execution.StageHeldOutTest {
			public.Stdout, public.Stderr = stage.Stdout, stage.Stderr
			public.TestEvents = append([]execution.TestEvent(nil), stage.TestEvents...)
		}
		result.Stages = append(result.Stages, public)
	}
	if value.Response.Failure != nil {
		result.Failure = &executionFailureDTO{Code: value.Response.Failure.Code, Message: "execution infrastructure failed"}
	}
	return result
}

func submissionResponse(result submission.Result) map[string]any {
	return map[string]any{
		"api_version": APIVersion,
		"submission":  result.Submission,
		"execution": map[string]any{
			"id": result.ExecutionID, "sequence": result.ExecutionSequence,
			"status": result.Submission.LatestExecutionStatus,
		},
	}
}
