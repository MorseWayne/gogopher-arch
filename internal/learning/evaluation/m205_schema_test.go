package evaluation

import (
	"strconv"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

func TestM205LearningLoopSolutionsPassRealReleaseAndSandboxEvaluation(t *testing.T) {
	tests := []struct {
		name, activity string
		files          map[string]string
		explanation    string
	}{
		{
			name:     "schema practice",
			activity: "practice-schema-migration",
			files: map[string]string{
				"migrations/0001_create_projects.up.sql": projectSchemaSolution(),
				"queries/list_active_projects.sql":       projectLookupSolution(),
			},
		},
		{
			name:     "checks schema",
			activity: "assessment-gocheck-schema",
			files: map[string]string{
				"migrations/0001_create_checks.up.sql":           checksSchemaSolution(),
				"migrations/0002_add_checks_lookup_index.up.sql": checksIndexSolution(),
				"queries/list_enabled_checks.sql":                checksLookupSolution(false),
				"queries/explain_list_enabled_checks.sql":        checksLookupSolution(true),
				"internal/schema/migrations_test.go":             migrationTableTests(),
			},
			explanation: "foreign key 把 owner 归属交给数据库验证，constraint 同时守住非空、非空白与 owner 内唯一。复合 index 依次放置 owner_id、enabled 和 created_at DESC，对齐等值过滤与排序；EXPLAIN 使用完全相同的 SELECT 检查访问路径，避免用另一条查询得出错误结论。migration 只向前建表和加索引，不删除历史数据。",
		},
		{
			name:     "alerts schema variant",
			activity: "review-gocheck-alert-schema",
			files: map[string]string{
				"migrations/0001_create_alert_rules.up.sql":           alertsSchemaSolution(),
				"migrations/0002_add_alert_rules_lookup_index.up.sql": alertsIndexSolution(),
				"queries/list_active_alert_rules.sql":                 alertsLookupSolution(false),
				"queries/explain_list_active_alert_rules.sql":         alertsLookupSolution(true),
				"internal/schema/migrations_test.go":                  migrationTableTests(),
			},
			explanation: "foreign key 保证每条 alert rule 都属于存在的 tenant，constraint 负责必填、非空白 destination 与 tenant 内唯一。复合 index 按 tenant_id、active、severity、created_at DESC 排列，与查询的三段等值过滤和倒序一致；EXPLAIN 检查同一生产查询的计划。升级 migration 只新增关系和索引，保留已有审计与业务数据。",
		},
	}
	registry := draftReleaseRegistry(t)
	for index, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			activity, err := registry.ActivityView(registry.CurrentReleaseID(), tc.activity, 1)
			if err != nil {
				t.Fatal(err)
			}
			task, err := registry.ExecutionTask(registry.CurrentReleaseID(), activity.TaskRef.ID, activity.TaskRef.Version)
			if err != nil {
				t.Fatal(err)
			}
			workspace, err := registry.PublicWorkspace(registry.CurrentReleaseID(), task.ID, task.Version)
			if err != nil {
				t.Fatal(err)
			}
			for path, source := range tc.files {
				workspace[path] = source
			}
			current := attempt.Attempt{ID: "00000000-0000-4000-9930-00000000000" + strconv.Itoa(index+1), ReleaseID: registry.CurrentReleaseID(), ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace)}
			builder, err := execution.NewSpecBuilder(registry)
			if err != nil {
				t.Fatal(err)
			}
			executionID := "00000000-0000-4000-9940-00000000000" + strconv.Itoa(index+1)
			spec, err := builder.Build(current, executionID, execution.ActionSubmit)
			if err != nil {
				t.Fatal(err)
			}
			response, err := runRegressionSandbox(t, spec)
			if err != nil {
				t.Fatal(err)
			}
			if response.Status != execution.ExecutionSucceeded {
				t.Fatalf("sandbox response = %#v", response)
			}
			frozen := submission.Submission{ID: "00000000-0000-4000-9950-00000000000" + strconv.Itoa(index+1), AttemptID: current.ID, ReleaseID: current.ReleaseID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Workspace: workspace, Explanation: tc.explanation}
			terminal := execution.Execution{ID: executionID, AttemptID: current.ID, SubmissionID: frozen.ID, TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash, Status: response.Status, Response: &response}
			generator, _ := NewGenerator(registry)
			results, err := generator.Generate(frozen, terminal)
			if err != nil {
				t.Fatal(err)
			}
			for _, result := range results {
				if result.Status != execution.RulePassed {
					t.Fatalf("rule %s = %#v", result.RuleID, result)
				}
			}
		})
	}
}

func projectSchemaSolution() string {
	return `CREATE TABLE projects (
  id UUID PRIMARY KEY,
  owner_id UUID NOT NULL,
  name TEXT NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CHECK (length(btrim(name)) > 0),
  UNIQUE (owner_id, name)
);
CREATE INDEX projects_owner_active_created_idx ON projects (owner_id, active, created_at DESC);
`
}

func projectLookupSolution() string {
	return `SELECT id, owner_id, name, active, created_at
FROM projects
WHERE owner_id = $1 AND active = TRUE
ORDER BY created_at DESC;
`
}

func checksSchemaSolution() string {
	return `CREATE TABLE checks (
  id UUID PRIMARY KEY,
  owner_id UUID NOT NULL REFERENCES owners(id),
  target TEXT NOT NULL,
  schedule TEXT NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT checks_target_not_blank CHECK (length(btrim(target)) > 0),
  CONSTRAINT checks_owner_target_key UNIQUE (owner_id, target)
);
`
}

func checksIndexSolution() string {
	return `CREATE INDEX checks_owner_enabled_created_idx
ON checks (owner_id, enabled, created_at DESC);
`
}

func checksLookupSolution(explain bool) string {
	prefix := ""
	if explain {
		prefix = "EXPLAIN (ANALYZE, BUFFERS) "
	}
	return prefix + `SELECT id, owner_id, target, schedule, enabled, created_at, updated_at
FROM checks
WHERE owner_id = $1 AND enabled = TRUE
ORDER BY created_at DESC
LIMIT $2;
`
}

func alertsSchemaSolution() string {
	return `CREATE TABLE alert_rules (
  id UUID PRIMARY KEY,
  tenant_id UUID NOT NULL REFERENCES tenants(id),
  destination TEXT NOT NULL,
  severity TEXT NOT NULL,
  active BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT alert_rules_destination_not_blank CHECK (length(btrim(destination)) > 0),
  CONSTRAINT alert_rules_tenant_destination_key UNIQUE (tenant_id, destination)
);
`
}

func alertsIndexSolution() string {
	return `CREATE INDEX alert_rules_tenant_active_severity_created_idx
ON alert_rules (tenant_id, active, severity, created_at DESC);
`
}

func alertsLookupSolution(explain bool) string {
	prefix := ""
	if explain {
		prefix = "EXPLAIN (ANALYZE, BUFFERS) "
	}
	return prefix + `SELECT id, tenant_id, destination, severity, active, created_at, updated_at
FROM alert_rules
WHERE tenant_id = $1 AND active = TRUE AND severity = $2
ORDER BY created_at DESC
LIMIT $3;
`
}

func migrationTableTests() string {
	return `package schema
import "testing"
func TestMigrationContract(t *testing.T) {
  tests := []struct{name string}{{name:"constraints"},{name:"forward-only"},{name:"query-plan"}}
  for _, test := range tests { t.Run(test.name, func(t *testing.T){}) }
}
`
}
