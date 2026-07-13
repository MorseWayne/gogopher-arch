package sandbox

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
)

type RunnerOptions struct {
	TempDir  string
	GoBinary string
}

type Runner struct {
	tempDir  string
	goBinary string
}

func NewRunner(options RunnerOptions) *Runner {
	if options.GoBinary == "" {
		options.GoBinary = "go"
	}
	return &Runner{tempDir: options.TempDir, goBinary: options.GoBinary}
}

func (r *Runner) Run(ctx context.Context, spec execution.ExecutionSpec) (execution.ExecutionResponse, error) {
	if err := spec.Validate(); err != nil {
		return execution.ExecutionResponse{}, err
	}
	started := time.Now()
	deadline := started.Add(time.Duration(spec.Policy.TimeoutMS) * time.Millisecond)
	response := baseResponse(spec)
	executionRoot, err := os.MkdirTemp(r.tempDir, "gogopher-execution-*")
	if err != nil {
		return infraResponse(response, started, "workspace_create_failed", "Sandbox could not create an execution workspace"), nil
	}
	defer os.RemoveAll(executionRoot)

	if spec.Action == execution.ActionSubmit {
		return r.runSubmit(ctx, spec, response, executionRoot, started, deadline), nil
	}
	workspace := filepath.Join(executionRoot, execution.WorkspaceRoot)
	if err := materializeAssets(workspace, spec.Files, func(asset execution.FileAsset) bool {
		return asset.Role != execution.RoleHeldOutTest
	}); err != nil {
		return infraResponse(response, started, "workspace_materialize_failed", "Sandbox could not materialize the execution workspace"), nil
	}
	stage, arguments := actionCommand(spec.Action)
	result := r.runProcess(ctx, time.Until(deadline), executionRoot, workspace, arguments, spec.Policy.MaxOutputBytes)
	if result.infrastructureFailure != nil {
		return infraResponse(response, started, result.infrastructureFailure.Code, result.infrastructureFailure.Message), nil
	}
	stageResult := result.stageResult(stage, string(spec.Action))
	stageResult.Stdout = sanitizeOutput(stageResult.Stdout, executionRoot)
	stageResult.Stderr = sanitizeOutput(stageResult.Stderr, executionRoot)
	response.Stages = []execution.StageResult{stageResult}
	if stageResult.Status == execution.StagePassed {
		response.Status = execution.ExecutionSucceeded
	} else {
		response.Status = execution.ExecutionUserFailed
	}
	response.DurationMS = time.Since(started).Milliseconds()
	return response, nil
}

func baseResponse(spec execution.ExecutionSpec) execution.ExecutionResponse {
	return execution.ExecutionResponse{
		ProtocolVersion: execution.ProtocolVersion,
		ExecutionID:     spec.ExecutionID,
		Stages:          []execution.StageResult{},
		Policy: execution.PolicyReport{Network: execution.NetworkPolicyReport{
			Requested: spec.Policy.Network, Enforcement: execution.EnforcementPolicyOnly,
		}},
	}
}

func infraResponse(response execution.ExecutionResponse, started time.Time, code, message string) execution.ExecutionResponse {
	response.Status = execution.ExecutionInfraFailed
	response.DurationMS = time.Since(started).Milliseconds()
	response.Failure = &execution.Failure{Code: code, Message: message}
	return response
}

func actionCommand(action execution.Action) (execution.Stage, []string) {
	switch action {
	case execution.ActionBuild:
		return execution.StageBuild, []string{"build", "./..."}
	case execution.ActionTest:
		return execution.StageVisibleTest, []string{"test", "./..."}
	case execution.ActionVet:
		return execution.StageVet, []string{"vet", "./..."}
	default:
		panic("actionCommand called with unsupported action " + action)
	}
}

type processResult struct {
	stdout                string
	stderr                string
	exitCode              int
	durationMS            int64
	timedOut              bool
	outputTruncated       bool
	infrastructureFailure *execution.Failure
}

func (r processResult) stageResult(stage execution.Stage, label string) execution.StageResult {
	status := execution.StagePassed
	summary := label + " completed"
	if r.exitCode != 0 || r.timedOut {
		status = execution.StageFailed
		summary = label + " failed"
	}
	if r.timedOut {
		summary = label + " timed out"
	}
	return execution.StageResult{
		Stage: stage, Status: status, ExitCode: r.exitCode,
		Stdout: r.stdout, Stderr: r.stderr, DurationMS: r.durationMS,
		TimedOut: r.timedOut, OutputTruncated: r.outputTruncated, PublicSummary: summary,
	}
}

func (r *Runner) runProcess(ctx context.Context, timeout time.Duration, executionRoot, workingDirectory string, arguments []string, maxOutputBytes int) processResult {
	started := time.Now()
	runContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	capture := newOutputCapture(maxOutputBytes)
	command := exec.CommandContext(runContext, r.goBinary, arguments...)
	command.Dir = workingDirectory
	command.Stdout = capture.stdoutWriter()
	command.Stderr = capture.stderrWriter()
	if err := prepareGoEnvironment(executionRoot); err != nil {
		return processResult{infrastructureFailure: &execution.Failure{Code: "runner_environment_failed", Message: "Sandbox could not prepare the Go environment"}}
	}
	command.Env = sandboxEnvironment(executionRoot)
	err := command.Run()
	stdout, stderr, truncated := capture.values()
	result := processResult{stdout: stdout, stderr: stderr, exitCode: 0, durationMS: time.Since(started).Milliseconds(), outputTruncated: truncated}
	if err == nil {
		return result
	}
	if ctx.Err() != nil {
		result.infrastructureFailure = &execution.Failure{Code: "execution_context_cancelled", Message: "Sandbox execution was cancelled by its caller"}
		return result
	}
	if errors.Is(runContext.Err(), context.DeadlineExceeded) {
		result.exitCode = -1
		result.timedOut = true
		return result
	}
	var exitError *exec.ExitError
	if errors.As(err, &exitError) {
		result.exitCode = exitError.ExitCode()
		return result
	}
	result.infrastructureFailure = &execution.Failure{Code: "process_start_failed", Message: "Sandbox could not start the Go tool"}
	return result
}

func prepareGoEnvironment(executionRoot string) error {
	for _, directory := range []string{"home", "gocache", "gomodcache", "gotmp"} {
		if err := os.MkdirAll(filepath.Join(executionRoot, directory), 0o700); err != nil {
			return err
		}
	}
	return nil
}

func sandboxEnvironment(executionRoot string) []string {
	overrides := map[string]string{
		"CGO_ENABLED": "0", "GOENV": "off", "GOFLAGS": "", "GONOSUMDB": "*", "GOPROXY": "off",
		"GOSUMDB": "off", "GOTOOLCHAIN": "local", "GOWORK": "off",
		"HOME": filepath.Join(executionRoot, "home"), "GOCACHE": filepath.Join(executionRoot, "gocache"),
		"GOMODCACHE": filepath.Join(executionRoot, "gomodcache"), "GOTMPDIR": filepath.Join(executionRoot, "gotmp"),
		"TMPDIR": filepath.Join(executionRoot, "gotmp"),
	}
	keys := make([]string, 0, len(overrides))
	for key := range overrides {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	environment := make([]string, 0, len(overrides)+1)
	environment = append(environment, "PATH="+os.Getenv("PATH"))
	for _, key := range keys {
		environment = append(environment, key+"="+overrides[key])
	}
	return environment
}

func materializeAssets(root string, assets []execution.FileAsset, include func(execution.FileAsset) bool) error {
	if err := os.Mkdir(root, 0o700); err != nil {
		return fmt.Errorf("create workspace root: %w", err)
	}
	rootInfo, err := os.Lstat(root)
	if err != nil {
		return fmt.Errorf("inspect workspace root: %w", err)
	}
	if rootInfo.Mode()&os.ModeSymlink != 0 || !rootInfo.IsDir() {
		return fmt.Errorf("workspace root must be a real directory")
	}
	for _, asset := range assets {
		if !include(asset) {
			continue
		}
		if execution.SHA256Hex(asset.Content) != asset.SHA256 {
			return fmt.Errorf("asset %q hash mismatch", asset.Path)
		}
		target := filepath.Join(root, filepath.FromSlash(asset.Path))
		if err := makeParentsWithoutSymlinks(root, filepath.Dir(target)); err != nil {
			return fmt.Errorf("asset %q parent: %w", asset.Path, err)
		}
		if _, err := os.Lstat(target); err == nil {
			return fmt.Errorf("asset %q path already exists", asset.Path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("inspect asset %q: %w", asset.Path, err)
		}
		file, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if err != nil {
			return fmt.Errorf("create asset %q: %w", asset.Path, err)
		}
		_, writeError := file.WriteString(asset.Content)
		closeError := file.Close()
		if writeError != nil || closeError != nil {
			os.Remove(target)
			return fmt.Errorf("write asset %q: %v", asset.Path, errors.Join(writeError, closeError))
		}
		if asset.Access == execution.AccessReadonly {
			if err := os.Chmod(target, 0o400); err != nil {
				return fmt.Errorf("mark asset %q readonly: %w", asset.Path, err)
			}
		}
	}
	return nil
}

func makeParentsWithoutSymlinks(root, directory string) error {
	relative, err := filepath.Rel(root, directory)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return fmt.Errorf("directory escapes workspace")
	}
	current := root
	if relative == "." {
		return nil
	}
	for _, segment := range strings.Split(relative, string(filepath.Separator)) {
		current = filepath.Join(current, segment)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			if err := os.Mkdir(current, 0o700); err != nil {
				return err
			}
			info, err = os.Lstat(current)
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return fmt.Errorf("%q is not a real directory", current)
		}
	}
	return nil
}

type outputCapture struct {
	mutex     sync.Mutex
	remaining int
	truncated bool
	stdout    bytes.Buffer
	stderr    bytes.Buffer
}

type captureWriter struct {
	capture *outputCapture
	target  *bytes.Buffer
}

func newOutputCapture(limit int) *outputCapture {
	return &outputCapture{remaining: limit}
}

func (c *outputCapture) stdoutWriter() *captureWriter {
	return &captureWriter{capture: c, target: &c.stdout}
}
func (c *outputCapture) stderrWriter() *captureWriter {
	return &captureWriter{capture: c, target: &c.stderr}
}

func (w *captureWriter) Write(value []byte) (int, error) {
	w.capture.mutex.Lock()
	defer w.capture.mutex.Unlock()
	accepted := len(value)
	if accepted > w.capture.remaining {
		accepted = w.capture.remaining
		w.capture.truncated = true
	}
	if accepted > 0 {
		_, _ = w.target.Write(value[:accepted])
		w.capture.remaining -= accepted
	}
	return len(value), nil
}

func (c *outputCapture) values() (string, string, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.stdout.String(), c.stderr.String(), c.truncated
}
