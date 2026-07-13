package httpapi

import (
	"errors"
	"net/http"

	"github.com/MorseWayne/gogopher-arch/internal/learning/assistance"
	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

func writeLearningError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, attempt.ErrNotFound), errors.Is(err, assistance.ErrAttemptNotFound):
		writeError(w, http.StatusNotFound, "attempt_not_found", "learning attempt not found")
	case errors.Is(err, execution.ErrExecutionNotFound):
		writeError(w, http.StatusNotFound, "execution_not_found", "learning execution not found")
	case errors.Is(err, submission.ErrNotFound):
		writeError(w, http.StatusNotFound, "submission_not_found", "learning submission not found")
	case errors.Is(err, definition.ErrHintNotFound):
		writeError(w, http.StatusNotFound, "hint_not_found", "task hint not found")
	case errors.Is(err, execution.ErrInvalidRequest), errors.Is(err, submission.ErrInvalidRequest),
		errors.Is(err, assistance.ErrInvalidRequest), errors.Is(err, attempt.ErrInvalidWorkspace),
		errors.Is(err, assistance.ErrEventNotAllowed):
		writeError(w, http.StatusUnprocessableEntity, "validation_failed", err.Error())
	case errors.Is(err, attempt.ErrInactive), errors.Is(err, assistance.ErrAttemptInactive),
		errors.Is(err, execution.ErrAttemptUnavailable), errors.Is(err, submission.ErrAttemptInactive):
		writeError(w, http.StatusConflict, "attempt_inactive", "learning attempt is not active")
	case errors.Is(err, submission.ErrRetryUnavailable):
		writeError(w, http.StatusConflict, "retry_unavailable", "submission is not awaiting an infrastructure retry")
	default:
		var attemptConflict *attempt.RevisionConflict
		if errors.As(err, &attemptConflict) {
			writeJSON(w, http.StatusConflict, map[string]any{
				"error":            errorDTO{Code: "revision_conflict", Message: "workspace revision is stale"},
				"current_revision": attemptConflict.Revision, "current_hash": attemptConflict.Hash,
			})
			return
		}
		var executionWorkspace *execution.WorkspaceConflict
		if errors.As(err, &executionWorkspace) {
			writeJSON(w, http.StatusConflict, map[string]any{
				"error":            errorDTO{Code: "workspace_conflict", Message: "workspace revision or hash is stale"},
				"current_revision": executionWorkspace.Revision, "current_hash": executionWorkspace.Hash,
			})
			return
		}
		var submissionWorkspace *submission.WorkspaceConflict
		if errors.As(err, &submissionWorkspace) {
			writeJSON(w, http.StatusConflict, map[string]any{
				"error":            errorDTO{Code: "workspace_conflict", Message: "workspace revision or hash is stale"},
				"current_revision": submissionWorkspace.Revision, "current_hash": submissionWorkspace.Hash,
			})
			return
		}
		var executionConflict *execution.IdempotencyConflict
		if errors.As(err, &executionConflict) {
			writeJSON(w, http.StatusConflict, map[string]any{
				"error":        errorDTO{Code: "idempotency_conflict", Message: "request key conflicts with its original request"},
				"execution_id": executionConflict.ExecutionID,
			})
			return
		}
		var submissionConflict *submission.IdempotencyConflict
		if errors.As(err, &submissionConflict) {
			writeJSON(w, http.StatusConflict, map[string]any{
				"error":         errorDTO{Code: "idempotency_conflict", Message: "submission key conflicts with its original request"},
				"submission_id": submissionConflict.SubmissionID,
			})
			return
		}
		var alreadySubmitted *submission.AttemptAlreadySubmitted
		if errors.As(err, &alreadySubmitted) {
			writeJSON(w, http.StatusConflict, map[string]any{
				"error":         errorDTO{Code: "attempt_already_submitted", Message: "learning attempt already has a frozen submission"},
				"submission_id": alreadySubmitted.SubmissionID,
			})
			return
		}
		var assistanceConflict *assistance.IdempotencyConflict
		if errors.As(err, &assistanceConflict) {
			writeJSON(w, http.StatusConflict, map[string]any{
				"error":    errorDTO{Code: "idempotency_conflict", Message: "event key conflicts with its original event"},
				"event_id": assistanceConflict.EventID,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, "learning_unavailable", "learning service is unavailable")
	}
}

type errorDTO struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
