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
	"strings"
	"time"

	"github.com/MorseWayne/gogopher-arch/src/pkg/common"
)

type GopherRunner struct{}

func NewGopherRunner() *GopherRunner {
	return &GopherRunner{}
}

// ensureGoModule copies a pre-cached go.mod/go.sum into tmpDir if the user code
// imports external packages (e.g. github.com/lib/pq, github.com/redis/go-redis).
func (r *GopherRunner) ensureGoModule(tmpDir, code string) error {
	if !strings.Contains(code, "github.com/lib/pq") && !strings.Contains(code, "github.com/redis/go-redis") {
		return nil
	}
	for _, f := range []string{"go.mod", "go.sum"} {
		src := filepath.Join("/app/sandbox-module", f)
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("read %s: %w", src, err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, f), data, 0644); err != nil {
			return fmt.Errorf("write %s: %w", f, err)
		}
	}
	return nil
}

func (r *GopherRunner) Run(req common.SandboxRequest) common.SandboxResponse {
	start := time.Now()
	tmpDir, err := os.MkdirTemp("", "gopher-task-*")
	if err != nil {
		return r.errorResponse(req.ID, "Failed to create temp directory: "+err.Error())
	}
	defer os.RemoveAll(tmpDir)

	if err := r.ensureGoModule(tmpDir, req.Code); err != nil {
		return r.errorResponse(req.ID, "Failed to setup Go module: "+err.Error())
	}

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
