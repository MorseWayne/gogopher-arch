package sandbox

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
)

const maxExecutionRequestBytes = 16 << 20

type Executor interface {
	Run(context.Context, execution.ExecutionSpec) (execution.ExecutionResponse, error)
}

type HTTPHandler struct {
	executor Executor
}

func NewHTTPHandler(executor Executor) http.Handler {
	handler := &HTTPHandler{executor: executor}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /v1/executions", handler.execute)
	return mux
}

func (h *HTTPHandler) execute(w http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(w, request.Body, maxExecutionRequestBytes)
	spec, err := execution.DecodeSpec(request.Body)
	if err != nil {
		var validationError *execution.ValidationError
		if errors.As(err, &validationError) {
			writeAPIError(w, http.StatusUnprocessableEntity, "invalid_execution_spec", validationError.Error())
			return
		}
		writeAPIError(w, http.StatusBadRequest, "invalid_execution_json", "Execution request must be one complete JSON object")
		return
	}
	response, err := h.executor.Run(request.Context(), spec)
	if err != nil {
		writeAPIError(w, http.StatusUnprocessableEntity, "invalid_execution_spec", err.Error())
		return
	}
	if err := response.Validate(); err != nil {
		writeAPIError(w, http.StatusInternalServerError, "invalid_execution_response", "Sandbox produced an invalid execution response")
		return
	}
	writeJSON(w, http.StatusOK, response)
}

func writeAPIError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{Error: struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
