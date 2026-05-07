package runner

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/MorseWayne/gogopher-arch/src/pkg/common"
)

type Runner struct{}

func New() *Runner {
	return &Runner{}
}

func (r *Runner) Run(req common.SandboxRequest) common.SandboxResponse {
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

func (r *Runner) errorResponse(id, msg string) common.SandboxResponse {
	return common.SandboxResponse{ID: id, Status: "internal_error", Stderr: msg}
}