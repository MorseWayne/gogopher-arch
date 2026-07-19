package schema

import (
	"embed"
	"regexp"
	"strings"
	"testing"
)

//go:embed migrations/*.sql queries/*.sql
var migrationAssets embed.FS

func TestSchemaRelationsConstrained(t *testing.T) {
	sql := normalizedSQL(t, "migrations/0001_create_alert_rules.up.sql")
	requireSQL(t, sql, `create\s+table\s+alert_rules`, `id\s+uuid\s+primary\s+key`, `tenant_id\s+uuid\s+not\s+null`, `references\s+tenants\s*\(\s*id\s*\)`, `destination\s+text\s+not\s+null`, `severity\s+text\s+not\s+null`, `active\s+boolean\s+not\s+null\s+default\s+true`, `created_at\s+timestamptz\s+not\s+null\s+default\s+now\s*\(\s*\)`, `updated_at\s+timestamptz\s+not\s+null\s+default\s+now\s*\(\s*\)`, `check\s*\(\s*length\s*\(\s*btrim\s*\(\s*destination\s*\)\s*\)\s*>\s*0\s*\)`, `unique\s*\(\s*tenant_id\s*,\s*destination\s*\)`)
}

func TestMigrationForwardOnly(t *testing.T) {
	first := normalizedSQL(t, "migrations/0001_create_alert_rules.up.sql")
	second := normalizedSQL(t, "migrations/0002_add_alert_rules_lookup_index.up.sql")
	for path, sql := range map[string]string{"0001": first, "0002": second} {
		if regexp.MustCompile(`\b(drop|delete|truncate)\b`).MatchString(sql) {
			t.Fatalf("%s contains destructive SQL", path)
		}
	}
	if strings.Contains(second, "create table") || strings.Contains(second, "alter table") {
		t.Fatal("0002 must add only the lookup index")
	}
	requireSQL(t, second, `create\s+index\s+(if\s+not\s+exists\s+)?alert_rules_tenant_active_severity_created_idx\s+on\s+alert_rules\s*\(\s*tenant_id\s*,\s*active\s*,\s*severity\s*,\s*created_at\s+desc\s*\)`)
}

func TestLookupIndexPlanAligned(t *testing.T) {
	query := normalizedSQL(t, "queries/list_active_alert_rules.sql")
	explain := normalizedSQL(t, "queries/explain_list_active_alert_rules.sql")
	shape := []string{`select\s+id\s*,\s*tenant_id\s*,\s*destination\s*,\s*severity\s*,\s*active\s*,\s*created_at\s*,\s*updated_at\s+from\s+alert_rules`, `where\s+tenant_id\s*=\s*\$1\s+and\s+active\s*=\s*true\s+and\s+severity\s*=\s*\$2`, `order\s+by\s+created_at\s+desc`, `limit\s+\$3`}
	requireSQL(t, query, shape...)
	requireSQL(t, explain, append([]string{`explain\s*(\([^)]*\))?\s+select`}, shape[1:]...)...)
}

func normalizedSQL(t *testing.T, path string) string {
	t.Helper()
	data, err := migrationAssets.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return strings.Join(strings.Fields(strings.ToLower(string(data))), " ")
}
func requireSQL(t *testing.T, sql string, patterns ...string) {
	t.Helper()
	for _, pattern := range patterns {
		if !regexp.MustCompile(pattern).MatchString(sql) {
			t.Fatalf("SQL does not satisfy %q", pattern)
		}
	}
}
