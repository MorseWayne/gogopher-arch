package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

type AttemptService interface {
	Create(context.Context, attempt.CreateInput) (attempt.Attempt, error)
	Get(context.Context, string, string) (attempt.Attempt, error)
	Save(context.Context, attempt.SaveInput) (attempt.Attempt, error)
}

type AttemptHandler struct{ service AttemptService }

func NewAttemptHandler(service AttemptService) (*AttemptHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("attempt service is required")
	}
	return &AttemptHandler{service: service}, nil
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
	writeJSON(w, http.StatusOK, attemptResponse(value))
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
	Executions        []any                               `json:"executions"`
	Evidence          []any                               `json:"evidence"`
}

func attemptResponse(value attempt.Attempt) attemptDTO {
	return attemptDTO{
		APIVersion: APIVersion, ID: value.ID, ReleaseID: value.ReleaseID,
		ActivityID: value.ActivityID, ActivityVersion: value.ActivityVersion, ActivityHash: value.ActivityHash,
		TaskID: value.TaskID, TaskVersion: value.TaskVersion, TaskHash: value.TaskHash,
		CapabilityRefs: value.CapabilityRefs, Mode: value.Mode, Status: value.Status,
		Workspace: value.Workspace, WorkspaceRevision: value.WorkspaceRevision, WorkspaceHash: value.WorkspaceHash,
		Executions: []any{}, Evidence: []any{},
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
