package execution

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestExecutionSpecJSONRoundTripForEveryAction(t *testing.T) {
	for _, action := range []Action{ActionBuild, ActionTest, ActionVet, ActionSubmit} {
		t.Run(string(action), func(t *testing.T) {
			spec := validSpec(action)
			encoded, err := json.Marshal(spec)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := DecodeSpec(bytes.NewReader(encoded))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(decoded, spec) {
				t.Fatalf("round trip = %#v, want %#v", decoded, spec)
			}
		})
	}
}

func TestExecutionSpecRejectsUntrustedExecutionFields(t *testing.T) {
	for _, field := range []string{"command", "environment", "mounts", "host_path"} {
		spec := validSpec(ActionTest)
		encoded, err := json.Marshal(spec)
		if err != nil {
			t.Fatal(err)
		}
		payload := strings.TrimSuffix(string(encoded), "}") + `,"` + field + `":"forbidden"}`
		if _, err := DecodeSpec(strings.NewReader(payload)); err == nil || !strings.Contains(err.Error(), "unknown field") {
			t.Fatalf("DecodeSpec(%s) error = %v", field, err)
		}
	}
}

func TestExecutionSpecValidationInvariants(t *testing.T) {
	large := strings.Repeat("x", 65)
	tests := []struct {
		name   string
		mutate func(*ExecutionSpec)
		field  string
	}{
		{name: "protocol version", mutate: func(s *ExecutionSpec) { s.ProtocolVersion = 2 }, field: "protocol_version"},
		{name: "arbitrary action", mutate: func(s *ExecutionSpec) { s.Action = "go test; rm -rf /" }, field: "action"},
		{name: "absolute path", mutate: func(s *ExecutionSpec) { s.Files[0].Path = "/tmp/main.go" }, field: "files[0].path"},
		{name: "parent path", mutate: func(s *ExecutionSpec) { s.Files[0].Path = "../main.go" }, field: "files[0].path"},
		{name: "windows path", mutate: func(s *ExecutionSpec) { s.Files[0].Path = `cmd\main.go` }, field: "files[0].path"},
		{name: "path overlap", mutate: func(s *ExecutionSpec) {
			s.Files = append(s.Files, asset("main.go/child", "x", OriginLearnerWorkspace, AccessEditable, RoleWorkspace))
		}, field: "files[1].path"},
		{name: "digest mismatch", mutate: func(s *ExecutionSpec) { s.Files[0].SHA256 = strings.Repeat("0", 64) }, field: "files[0].sha256"},
		{name: "learner readonly", mutate: func(s *ExecutionSpec) { s.Files[0].Access = AccessReadonly }, field: "files[0].access"},
		{name: "release editable", mutate: func(s *ExecutionSpec) { s.Files[0].Origin = OriginReleaseBundle }, field: "files[0].access"},
		{name: "held out before submit", mutate: func(s *ExecutionSpec) {
			s.Files[0] = asset("hidden_test.go", "package task", OriginReleaseBundle, AccessReadonly, RoleHeldOutTest)
		}, field: "files[0].role"},
		{name: "file count", mutate: func(s *ExecutionSpec) {
			s.Limits.MaxFiles = 1
			s.Files = append(s.Files, asset("other.go", "package task", OriginLearnerWorkspace, AccessEditable, RoleWorkspace))
		}, field: "files"},
		{name: "file bytes", mutate: func(s *ExecutionSpec) {
			s.Limits.MaxFileBytes = 64
			s.Files[0] = asset("main.go", large, OriginLearnerWorkspace, AccessEditable, RoleWorkspace)
		}, field: "files[0].content"},
		{name: "total bytes", mutate: func(s *ExecutionSpec) {
			s.Limits.MaxFileBytes = 65
			s.Limits.MaxTotalBytes = 65
			s.Files = append(s.Files, asset("other.go", "x", OriginLearnerWorkspace, AccessEditable, RoleWorkspace))
			s.Files[0] = asset("main.go", large, OriginLearnerWorkspace, AccessEditable, RoleWorkspace)
		}, field: "files"},
		{name: "timeout", mutate: func(s *ExecutionSpec) { s.Policy.TimeoutMS = 31_000 }, field: "policy.timeout_ms"},
		{name: "output", mutate: func(s *ExecutionSpec) { s.Policy.MaxOutputBytes = MaxProtocolOutputBytes + 1 }, field: "policy.max_output_bytes"},
		{name: "network", mutate: func(s *ExecutionSpec) { s.Policy.Network = "host" }, field: "policy.network"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := validSpec(ActionTest)
			test.mutate(&spec)
			if err := spec.Validate(); err == nil || !strings.HasPrefix(err.Error(), test.field+":") {
				t.Fatalf("Validate() error = %v, want field %s", err, test.field)
			}
		})
	}
}

func TestExecutionResponseRequiresPolicyOnlyNetworkReport(t *testing.T) {
	response := ExecutionResponse{
		ProtocolVersion: ProtocolVersion,
		ExecutionID:     "execution-1",
		Status:          ExecutionSucceeded,
		Stages:          []StageResult{{Stage: StageBuild, Status: StagePassed}},
		Policy: PolicyReport{Network: NetworkPolicyReport{
			Requested: NetworkNone, Enforcement: EnforcementPolicyOnly,
		}},
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	var decoded ExecutionResponse
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if err := decoded.Validate(); err != nil {
		t.Fatal(err)
	}
	decoded.Policy.Network.Enforcement = "enforced"
	if err := decoded.Validate(); err == nil || !strings.HasPrefix(err.Error(), "policy.network:") {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestRuleResultValidation(t *testing.T) {
	result := RuleResult{
		RuleID: "module-builds", Status: RulePassed, Stage: StageBuild,
		Package: "./...", Summary: "module builds", ExecutionID: "execution-1",
	}
	if err := result.Validate(); err != nil {
		t.Fatal(err)
	}
	result.Status = "partial"
	if err := result.Validate(); err == nil || !strings.HasPrefix(err.Error(), "status:") {
		t.Fatalf("Validate() error = %v", err)
	}
}

func validSpec(action Action) ExecutionSpec {
	files := []FileAsset{asset("main.go", "package task\n", OriginLearnerWorkspace, AccessEditable, RoleWorkspace)}
	if action == ActionSubmit {
		files = append(files, asset("hidden_test.go", "package task\n", OriginReleaseBundle, AccessReadonly, RoleHeldOutTest))
	}
	return ExecutionSpec{
		ProtocolVersion: ProtocolVersion, ExecutionID: "execution-1", Language: GoLanguage,
		WorkspaceRoot: WorkspaceRoot, Action: action, Files: files,
		Limits: WorkspaceLimits{MaxFiles: 8, MaxFileBytes: 1 << 16, MaxTotalBytes: 1 << 18},
		Policy: ActionPolicy{TimeoutMS: 5_000, MaxOutputBytes: 65_536, Network: NetworkNone},
	}
}

func asset(path, content string, origin AssetOrigin, access AssetAccess, role AssetRole) FileAsset {
	return FileAsset{Path: path, Content: content, SHA256: SHA256Hex(content), Origin: origin, Access: access, Role: role}
}
