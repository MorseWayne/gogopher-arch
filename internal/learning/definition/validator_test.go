package definition

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatorAcceptsValidDefinitions(t *testing.T) {
	validator := loadValidator(t)

	tests := []struct {
		kind Kind
		doc  string
	}{
		{KindCapability, `{
			"id":"M1-01","version":1,"name":"运行模型与工具链",
			"description":"理解 Go module 和基础工具反馈。","milestone":"M1","domain":"tooling",
			"prerequisites":{"hard":[],"recommended":[]},
			"required_evidence":[{"type":"implement","independence":"independent","context":"same_context","rule_ids":["module-builds"]}],
			"review_policy":{"first_review_after_days":3,"success_interval_days":14,"failure_remediation_after_days":1},
			"resource_refs":[]
		}`},
		{KindActivity, `{
			"id":"guided-run-model","version":1,"title":"读懂工具反馈","kind":"learning_task","mode":"guided",
			"capability_refs":[{"id":"M1-01","version":1}],
			"task_ref":{"id":"guided-run-model-v1","version":1},
			"assistance_policy":{"hints":true,"references":true,"solution":true,"ai_declaration":true},
			"evidence_rules":[{"rule_id":"module-builds","capability_id":"M1-01","capability_version":1,"evidence_type":"diagnose","result":"passed"}]
		}`},
		{KindTask, `{
			"id":"guided-run-model-v1","version":1,"language":"go1.25","workspace_root":"workspace",
			"files":[{"source":"starter/go.mod","path":"go.mod","role":"starter","editable":false,"sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}],
			"editable_paths":[],"readonly_paths":["go.mod"],
			"limits":{"max_files":8,"max_file_bytes":65536,"max_total_bytes":262144},
			"actions":{"build":{"timeout_ms":5000,"max_output_bytes":65536,"network":"none"}},
			"visible_tests":[],"held_out_tests":[],"assessment_rules":[],
			"artifact_rules":{"workspace":"full","logs":["build"]}
		}`},
		{KindTask, `{
			"id":"blank-project-v1","version":1,"language":"go1.25","workspace_root":"workspace",
			"files":[{"source":"README.md","path":"README.md","role":"readme","editable":false,"sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}],
			"editable_paths":[],"readonly_paths":["README.md"],
			"workspace_policy":{"allow_new_files":true,"allow_delete_files":true},
			"limits":{"max_files":16,"max_file_bytes":65536,"max_total_bytes":262144},
			"actions":{"build":{"timeout_ms":5000,"max_output_bytes":65536,"network":"none"}},
			"visible_tests":[],"held_out_tests":[],"assessment_rules":[],
			"artifact_rules":{"workspace":"full","logs":["build"]}
		}`},
		{KindTask, `{
			"id":"miniflux-slice-v1","version":1,"language":"go1.26","workspace_root":"workspace",
			"files":[{"source":"starter/go.mod","path":"go.mod","role":"starter","editable":false,"sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}],
			"editable_paths":[],"readonly_paths":["go.mod"],
			"limits":{"max_files":8,"max_file_bytes":65536,"max_total_bytes":262144},
			"actions":{"build":{"timeout_ms":5000,"max_output_bytes":65536,"network":"none"}},
			"visible_tests":[],"held_out_tests":[],"assessment_rules":[],
			"artifact_rules":{"workspace":"full","logs":["build"]}
		}`},
	}

	for _, tt := range tests {
		if err := validator.Validate(tt.kind, []byte(tt.doc)); err != nil {
			t.Errorf("Validate(%s) error = %v", tt.kind, err)
		}
	}
}

func TestValidatorRejectsInvalidDefinitions(t *testing.T) {
	validator := loadValidator(t)

	tests := []struct {
		name string
		kind Kind
		doc  string
		want string
	}{
		{
			name: "capability missing version",
			kind: KindCapability,
			doc:  `{"id":"M1-01"}`,
			want: "version",
		},
		{
			name: "activity has unknown mode",
			kind: KindActivity,
			doc: `{
				"id":"activity","version":1,"title":"title","kind":"learning_task","mode":"legacy",
				"capability_refs":[],"task_ref":{"id":"task","version":1},
				"assistance_policy":{"hints":false,"references":false,"solution":false,"ai_declaration":false},
				"evidence_rules":[]
			}`,
			want: "mode",
		},
		{
			name: "task permits path traversal",
			kind: KindTask,
			doc: `{
				"id":"task","version":1,"language":"go1.25","workspace_root":"workspace",
				"files":[{"source":"starter/secret","path":"../secret","role":"starter","editable":true,"sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}],
				"editable_paths":["../secret"],"readonly_paths":[],
				"limits":{"max_files":1,"max_file_bytes":1024,"max_total_bytes":1024},
				"actions":{"build":{"timeout_ms":1000,"max_output_bytes":1024,"network":"none"}},
				"visible_tests":[],"held_out_tests":[],"assessment_rules":[],
				"artifact_rules":{"workspace":"full","logs":[]}
			}`,
			want: "path",
		},
		{
			name: "unknown property",
			kind: KindTask,
			doc: `{
				"id":"task","version":1,"language":"go1.25","workspace_root":"workspace","command":"rm -rf /",
				"files":[],"editable_paths":[],"readonly_paths":[],
				"limits":{"max_files":1,"max_file_bytes":1024,"max_total_bytes":1024},
				"actions":{"build":{"timeout_ms":1000,"max_output_bytes":1024,"network":"none"}},
				"visible_tests":[],"held_out_tests":[],"assessment_rules":[],
				"artifact_rules":{"workspace":"full","logs":[]}
			}`,
			want: "command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.kind, []byte(tt.doc))
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate() error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func loadValidator(t *testing.T) *Validator {
	t.Helper()
	schemas := os.DirFS("../../../content/learning/schemas")
	validator, err := NewValidator(schemas)
	if err != nil {
		t.Fatalf("NewValidator() error = %v", err)
	}
	return validator
}

func TestRepositoryDefinitionsValidate(t *testing.T) {
	validator := loadValidator(t)
	root := filepath.Clean("../../../content/learning")
	tests := []struct {
		kind Kind
		dir  string
	}{
		{KindCapability, filepath.Join(root, "capabilities")},
		{KindActivity, filepath.Join(root, "activities")},
		{KindTask, filepath.Join(root, "tasks")},
	}

	for _, tt := range tests {
		err := filepath.WalkDir(tt.dir, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				if os.IsNotExist(walkErr) && tt.kind == KindTask {
					return nil
				}
				return walkErr
			}
			if entry.IsDir() || filepath.Ext(path) != ".json" || (tt.kind == KindTask && entry.Name() != "task.json") {
				return nil
			}
			document, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if err := validator.Validate(tt.kind, document); err != nil {
				t.Errorf("%s: %v", path, err)
			}
			if tt.kind == KindTask {
				if err := ValidateTaskAssets(filepath.Dir(path), document); err != nil {
					t.Errorf("%s assets: %v", path, err)
				}
			}
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			t.Fatalf("walk %s: %v", tt.dir, err)
		}
	}
}
