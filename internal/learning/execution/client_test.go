package execution

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSandboxClientUsesVersionedEndpointAndValidatesResponse(t *testing.T) {
	var received ExecutionSpec
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/v1/executions" {
			t.Fatalf("request = %s %s", request.Method, request.URL.Path)
		}
		if err := json.NewDecoder(request.Body).Decode(&received); err != nil {
			t.Fatal(err)
		}
		_ = json.NewEncoder(w).Encode(successResponse(received))
	}))
	defer server.Close()
	client, err := NewSandboxClient(SandboxClientOptions{Endpoint: server.URL + "/v1/executions"})
	if err != nil {
		t.Fatal(err)
	}
	spec := validSpec(ActionBuild)
	response, err := client.Execute(t.Context(), spec)
	if err != nil {
		t.Fatal(err)
	}
	if received.ExecutionID != spec.ExecutionID || response.Status != ExecutionSucceeded {
		t.Fatalf("received = %#v, response = %#v", received, response)
	}
}

func TestSandboxClientRejectsRedirectAndInvalidResponse(t *testing.T) {
	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Redirect(w, &http.Request{}, "https://example.com", http.StatusTemporaryRedirect)
	}))
	defer redirect.Close()
	client, _ := NewSandboxClient(SandboxClientOptions{Endpoint: redirect.URL})
	if _, err := client.Execute(t.Context(), validSpec(ActionBuild)); err == nil || !strings.Contains(err.Error(), "307") {
		t.Fatalf("redirect Execute() error = %v", err)
	}

	invalid := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"status":"succeeded","secret":"unexpected"}`))
	}))
	defer invalid.Close()
	client, _ = NewSandboxClient(SandboxClientOptions{Endpoint: invalid.URL})
	if _, err := client.Execute(t.Context(), validSpec(ActionBuild)); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("invalid Execute() error = %v", err)
	}
}

func TestSandboxClientRejectsOversizedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("x", 65)))
	}))
	defer server.Close()
	client, err := NewSandboxClient(SandboxClientOptions{Endpoint: server.URL, MaxResponseBytes: 64})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := client.Execute(t.Context(), validSpec(ActionBuild)); err == nil || !strings.Contains(err.Error(), "exceeds 64 bytes") {
		t.Fatalf("Execute() error = %v", err)
	}
}

func successResponse(spec ExecutionSpec) ExecutionResponse {
	stage := StageBuild
	if spec.Action == ActionTest || spec.Action == ActionSubmit {
		stage = StageVisibleTest
	} else if spec.Action == ActionVet {
		stage = StageVet
	}
	return ExecutionResponse{
		ProtocolVersion: ProtocolVersion, ExecutionID: spec.ExecutionID,
		Status: ExecutionSucceeded, Stages: []StageResult{{Stage: stage, Status: StagePassed}},
		Policy: PolicyReport{Network: NetworkPolicyReport{Requested: NetworkNone, Enforcement: EnforcementPolicyOnly}},
	}
}
