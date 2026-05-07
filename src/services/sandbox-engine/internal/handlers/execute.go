package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MorseWayne/gogopher-arch/src/pkg/common"
	"github.com/MorseWayne/gogopher-arch/src/services/sandbox-engine/internal/runner"
)

type ExecuteHandler struct {
	runner *runner.Runner
}

func NewExecuteHandler(r *runner.Runner) *ExecuteHandler {
	return &ExecuteHandler{runner: r}
}

func (h *ExecuteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SandboxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Printf("[%s] Executing code task: %s\n", time.Now().Format(time.RFC3339), req.ID)
	resp := h.runner.Run(req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}