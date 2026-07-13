package sandbox

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
)

func TestHTTPHandlerAcceptsOnlyVersionedExecutionSpecs(t *testing.T) {
	executor := &executorStub{}
	handler := NewHTTPHandler(executor)
	spec := runnerSpec(execution.ActionBuild, []execution.FileAsset{
		runnerAsset("go.mod", "module task\n\ngo 1.25.0\n", execution.OriginReleaseBundle, execution.AccessReadonly, execution.RoleWorkspace),
	})
	payload, err := json.Marshal(spec)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/executions", bytes.NewReader(payload)))
	if response.Code != http.StatusOK || executor.calls != 1 {
		t.Fatalf("response = %d %s, calls = %d", response.Code, response.Body.String(), executor.calls)
	}
	var decoded execution.ExecutionResponse
	if err := json.NewDecoder(response.Body).Decode(&decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Policy.Network.Enforcement != execution.EnforcementPolicyOnly {
		t.Fatalf("response = %#v", decoded)
	}
}

func TestHTTPHandlerRejectsCommandsAndLegacyPayload(t *testing.T) {
	handler := NewHTTPHandler(&executorStub{})
	for name, payload := range map[string]string{
		"command": `{"protocol_version":1,"command":"go test ./..."}`,
		"legacy":  `{"id":"legacy","code":"package main","language":"go","timeout":5}`,
	} {
		t.Run(name, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/executions", strings.NewReader(payload)))
			if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "invalid_execution_json") {
				t.Fatalf("response = %d %s", response.Code, response.Body.String())
			}
		})
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/execute", strings.NewReader(`{}`)))
	if response.Code != http.StatusNotFound {
		t.Fatalf("legacy route status = %d", response.Code)
	}
}

type executorStub struct {
	calls int
}

func (s *executorStub) Run(_ context.Context, spec execution.ExecutionSpec) (execution.ExecutionResponse, error) {
	s.calls++
	return execution.ExecutionResponse{
		ProtocolVersion: execution.ProtocolVersion, ExecutionID: spec.ExecutionID,
		Status: execution.ExecutionSucceeded,
		Stages: []execution.StageResult{{Stage: execution.StageBuild, Status: execution.StagePassed}},
		Policy: execution.PolicyReport{Network: execution.NetworkPolicyReport{
			Requested: execution.NetworkNone, Enforcement: execution.EnforcementPolicyOnly,
		}},
	}, nil
}
