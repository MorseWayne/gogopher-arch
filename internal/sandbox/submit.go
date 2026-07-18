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
		return !isPrivateTestRole(asset.Role)
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

	privateStages := []struct {
		role   execution.AssetRole
		config privateTestConfig
	}{
		{role: execution.RoleHeldOutTest, config: privateTestConfig{
			stage: execution.StageHeldOutTest, label: "held-out checks", sourceDirectory: "held-out-source",
			binaryDirectory: "held-out-binaries", runtimePrefix: "held-out-runtime",
			includeSource: func(asset execution.FileAsset) bool { return asset.Role != execution.RoleRaceTest },
		}},
		{role: execution.RoleRaceTest, config: privateTestConfig{
			stage: execution.StageRace, label: "race checks", sourceDirectory: "race-source",
			binaryDirectory: "race-binaries", runtimePrefix: "race-runtime", buildFlags: []string{"-race"},
			buildEnvironment: map[string]string{"CGO_ENABLED": "1"},
			includeSource:    func(asset execution.FileAsset) bool { return asset.Role != execution.RoleHeldOutTest },
		}},
	}
	for _, private := range privateStages {
		packages := privateTestPackages(spec.Files, private.role)
		if len(packages) == 0 {
			continue
		}
		stage, failure := r.runPrivateTests(ctx, spec, executionRoot, packages, deadline, remainingOutput, private.config)
		if failure != nil {
			return infraResponse(response, started, failure.Code, failure.Message)
		}
		response.Stages = append(response.Stages, stage)
		if stage.Status == execution.StageFailed {
			response.Status = execution.ExecutionUserFailed
			response.DurationMS = time.Since(started).Milliseconds()
			return response
		}
	}
	response.Status = execution.ExecutionSucceeded
	response.DurationMS = time.Since(started).Milliseconds()
	return response
}

type privateTestConfig struct {
	stage            execution.Stage
	label            string
	sourceDirectory  string
	binaryDirectory  string
	runtimePrefix    string
	buildFlags       []string
	buildEnvironment map[string]string
	includeSource    func(execution.FileAsset) bool
}

func (r *Runner) runPrivateTests(ctx context.Context, spec execution.ExecutionSpec, executionRoot string, packages []string, deadline time.Time, maxOutputBytes int, config privateTestConfig) (execution.StageResult, *execution.Failure) {
	started := time.Now()
	buildRoot := filepath.Join(executionRoot, config.sourceDirectory)
	if err := materializeAssets(buildRoot, spec.Files, config.includeSource); err != nil {
		return execution.StageResult{}, &execution.Failure{Code: "workspace_materialize_failed", Message: "Sandbox could not materialize private test sources"}
	}
	binaryRoot := filepath.Join(executionRoot, config.binaryDirectory)
	if err := os.Mkdir(binaryRoot, 0o700); err != nil {
		return execution.StageResult{}, &execution.Failure{Code: "test_binary_directory_failed", Message: "Sandbox could not prepare private test binaries"}
	}
	binaries := make([]string, 0, len(packages))
	remainingOutput := maxOutputBytes
	truncated := false
	for index, packagePath := range packages {
		binary := filepath.Join(binaryRoot, fmt.Sprintf("package-%d.test", index))
		arguments := append([]string{"test"}, config.buildFlags...)
		arguments = append(arguments, "-c", "-o", binary, packageArgument(packagePath))
		result := r.runProcessWithEnvironment(ctx, time.Until(deadline), executionRoot, buildRoot, arguments, remainingOutput, config.buildEnvironment)
		if result.infrastructureFailure != nil {
			return execution.StageResult{}, result.infrastructureFailure
		}
		remainingOutput = remainingAfter(remainingOutput, result)
		truncated = truncated || result.outputTruncated
		if result.exitCode != 0 || result.timedOut || result.outputTruncated {
			return privateTestFailureStage(config.stage, config.label, started, result.exitCode, result.timedOut, truncated, config.label+" build failed"), nil
		}
		binaries = append(binaries, binary)
	}
	if err := os.RemoveAll(buildRoot); err != nil {
		return execution.StageResult{}, &execution.Failure{Code: "private_test_source_cleanup_failed", Message: "Sandbox could not remove private test sources before execution"}
	}

	allEvents := make([]execution.TestEvent, 0)
	for index, binary := range binaries {
		runtimeRoot := filepath.Join(executionRoot, fmt.Sprintf("%s-%d", config.runtimePrefix, index))
		if err := materializeAssets(runtimeRoot, spec.Files, func(asset execution.FileAsset) bool {
			return asset.Role == execution.RoleFixture
		}); err != nil {
			return execution.StageResult{}, &execution.Failure{Code: "runtime_fixture_failed", Message: "Sandbox could not prepare private test runtime fixtures"}
		}
		result := r.runProcess(ctx, time.Until(deadline), executionRoot, runtimeRoot, []string{"tool", "test2json", "-t", "-p", packageArgument(packages[index]), binary, "-test.v=test2json"}, remainingOutput)
		if result.infrastructureFailure != nil {
			return execution.StageResult{}, result.infrastructureFailure
		}
		remainingOutput = remainingAfter(remainingOutput, result)
		truncated = truncated || result.outputTruncated
		events, _, parseError := parseGoTestJSON([]byte(result.stdout), result.outputTruncated)
		if parseError != nil {
			return execution.StageResult{}, &execution.Failure{Code: "test_result_parse_failed", Message: "Sandbox could not parse structured private test results"}
		}
		allEvents = append(allEvents, events...)
		if result.exitCode != 0 || result.timedOut || result.outputTruncated {
			stage := privateTestFailureStage(config.stage, config.label, started, result.exitCode, result.timedOut, truncated, config.label+" failed")
			stage.TestEvents = allEvents
			return stage, nil
		}
	}
	return execution.StageResult{
		Stage: config.stage, Status: execution.StagePassed, ExitCode: 0,
		DurationMS: time.Since(started).Milliseconds(), OutputTruncated: truncated,
		PublicSummary: fmt.Sprintf("%s passed for %d package(s)", config.label, len(packages)), TestEvents: allEvents,
	}, nil
}

func privateTestFailureStage(stage execution.Stage, label string, started time.Time, exitCode int, timedOut, truncated bool, summary string) execution.StageResult {
	if timedOut {
		summary = label + " timed out"
	}
	return execution.StageResult{
		Stage: stage, Status: execution.StageFailed, ExitCode: exitCode,
		DurationMS: time.Since(started).Milliseconds(), TimedOut: timedOut,
		OutputTruncated: truncated, PublicSummary: summary,
	}
}

func privateTestPackages(assets []execution.FileAsset, role execution.AssetRole) []string {
	unique := make(map[string]struct{})
	for _, asset := range assets {
		if asset.Role == role {
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
