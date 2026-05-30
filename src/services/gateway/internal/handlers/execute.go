package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MorseWayne/gogopher-arch/src/internal/config"
)

type ExecuteHandler struct {
	cfg config.Config
}

func NewExecuteHandler(cfg config.Config) *ExecuteHandler {
	return &ExecuteHandler{cfg: cfg}
}

func (h *ExecuteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}

	fmt.Printf("[%s] Forwarding execution request to sandbox...\n", time.Now().Format(time.RFC3339))
	resp, err := http.Post(h.cfg.SandboxURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "Sandbox engine unreachable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		fmt.Printf("Failed to copy sandbox response: %v\n", err)
	}
}
