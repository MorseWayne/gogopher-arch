package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/MorseWayne/gogopher-arch/src/pkg/common"
)

type GopherRunner struct{}

func NewGopherRunner() *GopherRunner {
	return &GopherRunner{}
}

func (r *GopherRunner) Run(req common.SandboxRequest) common.SandboxResponse {
	start := time.Now()
	tmpDir, err := os.MkdirTemp("", "gopher-task-*")
	if err != nil {
		return r.errorResponse(req.ID, "Failed to create temp directory: "+err.Error())
	}
	defer os.RemoveAll(tmpDir)

	codePath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(codePath, []byte(req.Code), 0644); err != nil {
		return r.errorResponse(req.ID, "Failed to write code file: "+err.Error())
	}

	timeout := time.Duration(req.Timeout) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", "run", codePath)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	duration := time.Since(start)

	resp := common.SandboxResponse{
		ID:       req.ID,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	if ctx.Err() == context.DeadlineExceeded {
		resp.Status = "timeout"
		resp.ExitCode = -1
	} else if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			resp.ExitCode = exitError.ExitCode()
		} else {
			resp.ExitCode = 1
		}
		resp.Status = "error"
	} else {
		resp.Status = "success"
		resp.ExitCode = 0
	}

	return resp
}

func (r *GopherRunner) errorResponse(id, msg string) common.SandboxResponse {
	return common.SandboxResponse{ID: id, Status: "internal_error", Stderr: msg}
}

func main() {
	runner := NewGopherRunner()
	
	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
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
		resp := runner.Run(req)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	port := ":8081"
	fmt.Printf("Gogopher Arch Sandbox Engine listening on %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
