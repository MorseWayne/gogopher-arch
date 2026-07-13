package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/assistance"
	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/attemptview"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/evaluation"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
)

type AttemptService interface {
	Create(context.Context, attempt.CreateInput) (attempt.Attempt, error)
	Get(context.Context, string, string) (attempt.Attempt, error)
	Save(context.Context, attempt.SaveInput) (attempt.Attempt, error)
}

type AttemptDetails interface {
	Load(context.Context, string, string) (attemptview.Related, error)
}

type AttemptHandler struct {
	service AttemptService
	details AttemptDetails
}

func NewAttemptHandler(service AttemptService, details ...AttemptDetails) (*AttemptHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("attempt service is required")
	}
	handler := &AttemptHandler{service: service}
	if len(details) > 0 {
		handler.details = details[0]
	}
	return handler, nil
}

func (h *AttemptHandler) Create(w http.ResponseWriter, request *http.Request) {
	owner, ok := SessionFromContext(request.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "learning session is required")
		return
	}
	var body struct {
		ActivityID      string `json:"activity_id"`
		ActivityVersion int    `json:"activity_version"`
	}
	if err := decodeJSON(request, &body); err != nil || body.ActivityID == "" || body.ActivityVersion < 1 {
		writeError(w, http.StatusUnprocessableEntity, "validation_failed", "activity_id and activity_version are required")
		return
	}
	created, err := h.service.Create(request.Context(), attempt.CreateInput{LearnerID: owner.LearnerID, ActivityID: body.ActivityID, ActivityVersion: body.ActivityVersion})
	if err != nil {
		if errors.Is(err, definition.ErrDefinitionNotFound) {
			writeError(w, http.StatusUnprocessableEntity, "unknown_activity", "activity does not exist")
			return
		}
		writeError(w, http.StatusInternalServerError, "attempt_unavailable", "learning attempt is unavailable")
		return
	}
	writeJSON(w, http.StatusCreated, attemptResponse(created))
}

func (h *AttemptHandler) Get(w http.ResponseWriter, request *http.Request, attemptID string) {
	owner, ok := SessionFromContext(request.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "learning session is required")
		return
	}
	value, err := h.service.Get(request.Context(), owner.LearnerID, attemptID)
	if err != nil {
		h.writeAttemptError(w, err)
		return
	}
	response := attemptResponse(value)
	if h.details != nil {
		related, err := h.details.Load(request.Context(), owner.LearnerID, attemptID)
		if err != nil {
			writeLearningError(w, err)
			return
		}
		response = attemptDetailResponse(response, related)
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *AttemptHandler) SaveWorkspace(w http.ResponseWriter, request *http.Request, attemptID string) {
	owner, ok := SessionFromContext(request.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthenticated", "learning session is required")
		return
	}
	var body struct {
		BaseRevision int64             `json:"base_revision"`
		Files        map[string]string `json:"files"`
	}
	if err := decodeJSON(request, &body); err != nil || body.BaseRevision < 0 || body.Files == nil {
		writeError(w, http.StatusUnprocessableEntity, "validation_failed", "base_revision and complete files map are required")
		return
	}
	value, err := h.service.Save(request.Context(), attempt.SaveInput{
		LearnerID: owner.LearnerID, AttemptID: attemptID, BaseRevision: body.BaseRevision, Files: body.Files,
	})
	if err != nil {
		h.writeAttemptError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, attemptResponse(value))
}

func (h *AttemptHandler) writeAttemptError(w http.ResponseWriter, err error) {
	writeLearningError(w, err)
}

type attemptDTO struct {
	APIVersion        string                              `json:"api_version"`
	ID                string                              `json:"id"`
	ReleaseID         string                              `json:"release_id"`
	ActivityID        string                              `json:"activity_id"`
	ActivityVersion   int                                 `json:"activity_version"`
	ActivityHash      string                              `json:"activity_hash"`
	TaskID            string                              `json:"task_id"`
	TaskVersion       int                                 `json:"task_version"`
	TaskHash          string                              `json:"task_hash"`
	CapabilityRefs    []definition.VersionedDefinitionRef `json:"capability_refs"`
	Mode              string                              `json:"mode"`
	Status            string                              `json:"status"`
	Workspace         map[string]string                   `json:"workspace"`
	WorkspaceRevision int64                               `json:"workspace_revision"`
	WorkspaceHash     string                              `json:"workspace_hash"`
	Submission        *submissionDTO                      `json:"submission,omitempty"`
	Executions        []executionDTO                      `json:"executions"`
	RuleResults       []ruleResultDTO                     `json:"rule_results"`
	Evidence          []evidenceDTO                       `json:"evidence"`
}

func attemptResponse(value attempt.Attempt) attemptDTO {
	return attemptDTO{
		APIVersion: APIVersion, ID: value.ID, ReleaseID: value.ReleaseID,
		ActivityID: value.ActivityID, ActivityVersion: value.ActivityVersion, ActivityHash: value.ActivityHash,
		TaskID: value.TaskID, TaskVersion: value.TaskVersion, TaskHash: value.TaskHash,
		CapabilityRefs: value.CapabilityRefs, Mode: value.Mode, Status: value.Status,
		Workspace: value.Workspace, WorkspaceRevision: value.WorkspaceRevision, WorkspaceHash: value.WorkspaceHash,
		Executions: []executionDTO{}, RuleResults: []ruleResultDTO{}, Evidence: []evidenceDTO{},
	}
}

type submissionDTO struct {
	ID                    string                    `json:"id"`
	WorkspaceRevision     int64                     `json:"workspace_revision"`
	WorkspaceHash         string                    `json:"workspace_hash"`
	RuleSetHash           string                    `json:"rule_set_hash"`
	AssistanceCutoff      int64                     `json:"assistance_cutoff_seq"`
	Status                string                    `json:"status"`
	LatestExecutionID     string                    `json:"latest_execution_id"`
	LatestExecutionSeq    int                       `json:"latest_execution_sequence"`
	LatestExecutionStatus execution.ExecutionStatus `json:"latest_execution_status"`
	CreatedAt             time.Time                 `json:"created_at"`
	EvaluatedAt           *time.Time                `json:"evaluated_at,omitempty"`
}

type ruleResultDTO struct {
	RuleID      string               `json:"rule_id"`
	Status      execution.RuleStatus `json:"status"`
	Stage       execution.Stage      `json:"stage"`
	Package     string               `json:"package,omitempty"`
	Test        string               `json:"test,omitempty"`
	Analyzer    string               `json:"analyzer,omitempty"`
	Summary     string               `json:"summary"`
	ExecutionID string               `json:"execution_id"`
}

type evidenceDTO struct {
	ID                string                  `json:"id"`
	EvaluationBatchID string                  `json:"evaluation_batch_id"`
	CapabilityID      string                  `json:"capability_id"`
	CapabilityVersion int                     `json:"capability_version"`
	EvidenceRuleID    string                  `json:"evidence_rule_id"`
	EvidenceType      string                  `json:"evidence_type"`
	Result            execution.RuleStatus    `json:"result"`
	Independence      assistance.Independence `json:"independence"`
	ContextLevel      string                  `json:"context_level"`
	Evaluator         string                  `json:"evaluator"`
	Reason            string                  `json:"reason"`
	OccurredAt        time.Time               `json:"occurred_at"`
}

func attemptDetailResponse(response attemptDTO, related attemptview.Related) attemptDTO {
	if related.Submission != nil {
		value := related.Submission
		response.Submission = &submissionDTO{
			ID: value.ID, WorkspaceRevision: value.WorkspaceRevision, WorkspaceHash: value.WorkspaceHash,
			RuleSetHash: value.RuleSetHash, AssistanceCutoff: value.AssistanceCutoff, Status: string(value.Status),
			LatestExecutionID: value.LatestExecutionID, LatestExecutionSeq: value.LatestExecutionSeq,
			LatestExecutionStatus: value.LatestExecutionStatus, CreatedAt: value.CreatedAt, EvaluatedAt: value.EvaluatedAt,
		}
	}
	for _, value := range related.Executions {
		response.Executions = append(response.Executions, executionResponse(value))
	}
	for _, value := range related.RuleResults {
		public := ruleResultDTO{
			RuleID: value.RuleID, Status: value.Status, Stage: value.Stage, Analyzer: value.Analyzer,
			Summary: value.Summary, ExecutionID: value.ExecutionID,
		}
		if value.Stage != execution.StageHeldOutTest {
			public.Package, public.Test = value.Package, value.Test
		}
		response.RuleResults = append(response.RuleResults, public)
	}
	for _, value := range related.Evidence {
		response.Evidence = append(response.Evidence, evidenceResponse(value))
	}
	return response
}

func evidenceResponse(value evaluation.Evidence) evidenceDTO {
	return evidenceDTO{
		ID: value.ID, EvaluationBatchID: value.EvaluationBatchID,
		CapabilityID: value.CapabilityID, CapabilityVersion: value.CapabilityVersion,
		EvidenceRuleID: value.EvidenceRuleID, EvidenceType: value.EvidenceType,
		Result: value.Result, Independence: value.Independence, ContextLevel: value.ContextLevel,
		Evaluator: value.Evaluator, Reason: value.Reason, OccurredAt: value.OccurredAt,
	}
}

func decodeJSON(request *http.Request, target any) error {
	defer request.Body.Close()
	contents, err := io.ReadAll(io.LimitReader(request.Body, 2<<20))
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return fmt.Errorf("trailing JSON")
	}
	return nil
}
