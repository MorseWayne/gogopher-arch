package sandbox

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
)

type goTestRecord struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
	Output  string  `json:"Output"`
}

func (r *Runner) runSubmit(ctx context.Context, spec execution.ExecutionSpec, response execution.ExecutionResponse, executionRoot string, started, deadline time.Time) execution.ExecutionResponse {
	visibleRoot := filepath.Join(executionRoot, "visible-source")
	if err := materializeAssets(visibleRoot, spec.Files, func(asset execution.FileAsset) bool {
		return asset.Role != execution.RoleHeldOutTest
	}); err != nil {
		return infraResponse(response, started, "workspace_materialize_failed", "Sandbox could not materialize the visible test workspace")
	}
	remainingOutput := spec.Policy.MaxOutputBytes
	for _, stage := range []struct {
		name      execution.Stage
		label     string
		arguments []string
	}{
		{name: execution.StageBuild, label: "build", arguments: []string{"build", "./..."}},
		{name: execution.StageVet, label: "vet", arguments: []string{"vet", "./..."}},
	} {
		process := r.runProcess(ctx, time.Until(deadline), executionRoot, visibleRoot, stage.arguments, remainingOutput)
		if process.infrastructureFailure != nil {
			return infraResponse(response, started, process.infrastructureFailure.Code, process.infrastructureFailure.Message)
		}
		stageResult := process.stageResult(stage.name, stage.label)
		stageResult.Stdout = sanitizeOutput(stageResult.Stdout, executionRoot)
		stageResult.Stderr = sanitizeOutput(stageResult.Stderr, executionRoot)
		if process.outputTruncated {
			stageResult.Status = execution.StageFailed
			stageResult.PublicSummary = stage.label + " output exceeded the configured limit"
		}
		response.Stages = append(response.Stages, stageResult)
		remainingOutput = remainingAfter(remainingOutput, process)
		if stageResult.Status == execution.StageFailed {
			response.Status = execution.ExecutionUserFailed
			response.DurationMS = time.Since(started).Milliseconds()
			return response
		}
	}

	visibleProcess := r.runProcess(ctx, time.Until(deadline), executionRoot, visibleRoot, []string{"test", "-json", "./..."}, remainingOutput)
	if visibleProcess.infrastructureFailure != nil {
		return infraResponse(response, started, visibleProcess.infrastructureFailure.Code, visibleProcess.infrastructureFailure.Message)
	}
	visibleEvents, visibleOutput, parseError := parseGoTestJSON([]byte(visibleProcess.stdout), visibleProcess.outputTruncated)
	visibleStage := visibleProcess.stageResult(execution.StageVisibleTest, "visible tests")
	visibleStage.Stdout = sanitizeOutput(visibleOutput, executionRoot)
	visibleStage.Stderr = sanitizeOutput(visibleStage.Stderr, executionRoot)
	visibleStage.TestEvents = visibleEvents
	if parseError != nil {
		return infraResponse(response, started, "test_result_parse_failed", "Sandbox could not parse structured visible test results")
	}
	if visibleProcess.outputTruncated {
		visibleStage.Status = execution.StageFailed
		visibleStage.PublicSummary = "visible test output exceeded the configured limit"
	}
	response.Stages = append(response.Stages, visibleStage)
	remainingOutput = remainingAfter(remainingOutput, visibleProcess)
	if visibleStage.Status == execution.StageFailed {
		response.Status = execution.ExecutionUserFailed
		response.DurationMS = time.Since(started).Milliseconds()
		return response
	}

	packages := heldOutPackages(spec.Files)
	if len(packages) == 0 {
		response.Status = execution.ExecutionSucceeded
		response.DurationMS = time.Since(started).Milliseconds()
		return response
	}
	heldOutStage, failure := r.runHeldOut(ctx, spec, executionRoot, packages, deadline, remainingOutput)
	if failure != nil {
		return infraResponse(response, started, failure.Code, failure.Message)
	}
	response.Stages = append(response.Stages, heldOutStage)
	if heldOutStage.Status == execution.StagePassed {
		response.Status = execution.ExecutionSucceeded
	} else {
		response.Status = execution.ExecutionUserFailed
	}
	response.DurationMS = time.Since(started).Milliseconds()
	return response
}

func (r *Runner) runHeldOut(ctx context.Context, spec execution.ExecutionSpec, executionRoot string, packages []string, deadline time.Time, maxOutputBytes int) (execution.StageResult, *execution.Failure) {
	started := time.Now()
	buildRoot := filepath.Join(executionRoot, "held-out-source")
	if err := materializeAssets(buildRoot, spec.Files, func(execution.FileAsset) bool { return true }); err != nil {
		return execution.StageResult{}, &execution.Failure{Code: "workspace_materialize_failed", Message: "Sandbox could not materialize held-out test sources"}
	}
	binaryRoot := filepath.Join(executionRoot, "test-binaries")
	if err := os.Mkdir(binaryRoot, 0o700); err != nil {
		return execution.StageResult{}, &execution.Failure{Code: "test_binary_directory_failed", Message: "Sandbox could not prepare held-out test binaries"}
	}
	binaries := make([]string, 0, len(packages))
	remainingOutput := maxOutputBytes
	truncated := false
	for index, packagePath := range packages {
		binary := filepath.Join(binaryRoot, fmt.Sprintf("package-%d.test", index))
		result := r.runProcess(ctx, time.Until(deadline), executionRoot, buildRoot, []string{"test", "-c", "-o", binary, packageArgument(packagePath)}, remainingOutput)
		if result.infrastructureFailure != nil {
			return execution.StageResult{}, result.infrastructureFailure
		}
		remainingOutput = remainingAfter(remainingOutput, result)
		truncated = truncated || result.outputTruncated
		if result.exitCode != 0 || result.timedOut || result.outputTruncated {
			return heldOutFailureStage(started, result.exitCode, result.timedOut, truncated, "held-out test build failed"), nil
		}
		binaries = append(binaries, binary)
	}
	if err := os.RemoveAll(buildRoot); err != nil {
		return execution.StageResult{}, &execution.Failure{Code: "held_out_source_cleanup_failed", Message: "Sandbox could not remove held-out test sources before execution"}
	}

	allEvents := make([]execution.TestEvent, 0)
	for index, binary := range binaries {
		runtimeRoot := filepath.Join(executionRoot, fmt.Sprintf("runtime-%d", index))
		if err := materializeAssets(runtimeRoot, spec.Files, func(asset execution.FileAsset) bool {
			return asset.Role == execution.RoleFixture
		}); err != nil {
			return execution.StageResult{}, &execution.Failure{Code: "runtime_fixture_failed", Message: "Sandbox could not prepare held-out runtime fixtures"}
		}
		result := r.runProcess(ctx, time.Until(deadline), executionRoot, runtimeRoot, []string{"tool", "test2json", "-t", "-p", packageArgument(packages[index]), binary, "-test.v=test2json"}, remainingOutput)
		if result.infrastructureFailure != nil {
			return execution.StageResult{}, result.infrastructureFailure
		}
		remainingOutput = remainingAfter(remainingOutput, result)
		truncated = truncated || result.outputTruncated
		events, _, parseError := parseGoTestJSON([]byte(result.stdout), result.outputTruncated)
		if parseError != nil {
			return execution.StageResult{}, &execution.Failure{Code: "test_result_parse_failed", Message: "Sandbox could not parse structured held-out test results"}
		}
		allEvents = append(allEvents, events...)
		if result.exitCode != 0 || result.timedOut || result.outputTruncated {
			stage := heldOutFailureStage(started, result.exitCode, result.timedOut, truncated, "held-out checks failed")
			stage.TestEvents = allEvents
			return stage, nil
		}
	}
	return execution.StageResult{
		Stage: execution.StageHeldOutTest, Status: execution.StagePassed, ExitCode: 0,
		DurationMS: time.Since(started).Milliseconds(), OutputTruncated: truncated,
		PublicSummary: fmt.Sprintf("held-out checks passed for %d package(s)", len(packages)), TestEvents: allEvents,
	}, nil
}

func heldOutFailureStage(started time.Time, exitCode int, timedOut, truncated bool, summary string) execution.StageResult {
	if timedOut {
		summary = "held-out checks timed out"
	}
	return execution.StageResult{
		Stage: execution.StageHeldOutTest, Status: execution.StageFailed, ExitCode: exitCode,
		DurationMS: time.Since(started).Milliseconds(), TimedOut: timedOut,
		OutputTruncated: truncated, PublicSummary: summary,
	}
}

func heldOutPackages(assets []execution.FileAsset) []string {
	unique := make(map[string]struct{})
	for _, asset := range assets {
		if asset.Role == execution.RoleHeldOutTest {
			unique[path.Dir(asset.Path)] = struct{}{}
		}
	}
	packages := make([]string, 0, len(unique))
	for packagePath := range unique {
		packages = append(packages, packagePath)
	}
	sort.Strings(packages)
	return packages
}

func packageArgument(packagePath string) string {
	if packagePath == "." {
		return "."
	}
	return "./" + packagePath
}

func remainingAfter(remaining int, result processResult) int {
	remaining -= len(result.stdout) + len(result.stderr)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func parseGoTestJSON(raw []byte, truncated bool) ([]execution.TestEvent, string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 64*1024), execution.MaxProtocolOutputBytes)
	events := make([]execution.TestEvent, 0)
	var output strings.Builder
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var record goTestRecord
		if err := json.Unmarshal(line, &record); err != nil {
			if truncated {
				break
			}
			return nil, "", fmt.Errorf("decode go test event: %w", err)
		}
		if record.Action == "output" {
			output.WriteString(record.Output)
			continue
		}
		if record.Action == "run" || record.Action == "pass" || record.Action == "fail" || record.Action == "skip" {
			events = append(events, execution.TestEvent{Action: record.Action, Package: record.Package, Test: record.Test, Elapsed: record.Elapsed})
		}
	}
	if err := scanner.Err(); err != nil && !truncated {
		return nil, "", fmt.Errorf("scan go test events: %w", err)
	}
	return events, output.String(), nil
}

func sanitizeOutput(value, executionRoot string) string {
	return strings.ReplaceAll(value, executionRoot, "<sandbox>")
}
