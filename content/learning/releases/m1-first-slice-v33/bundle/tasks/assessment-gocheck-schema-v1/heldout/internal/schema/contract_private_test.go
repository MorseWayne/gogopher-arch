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
	sql := normalizedSQL(t, "migrations/0001_create_checks.up.sql")
	requireSQL(t, sql, `create\s+table\s+checks`, `id\s+uuid\s+primary\s+key`, `owner_id\s+uuid\s+not\s+null`, `references\s+owners\s*\(\s*id\s*\)`, `target\s+text\s+not\s+null`, `schedule\s+text\s+not\s+null`, `enabled\s+boolean\s+not\s+null\s+default\s+true`, `created_at\s+timestamptz\s+not\s+null\s+default\s+now\s*\(\s*\)`, `updated_at\s+timestamptz\s+not\s+null\s+default\s+now\s*\(\s*\)`, `check\s*\(\s*length\s*\(\s*btrim\s*\(\s*target\s*\)\s*\)\s*>\s*0\s*\)`, `unique\s*\(\s*owner_id\s*,\s*target\s*\)`)
}

func TestMigrationForwardOnly(t *testing.T) {
	first := normalizedSQL(t, "migrations/0001_create_checks.up.sql")
	second := normalizedSQL(t, "migrations/0002_add_checks_lookup_index.up.sql")
	for path, sql := range map[string]string{"0001": first, "0002": second} {
		if regexp.MustCompile(`\b(drop|delete|truncate)\b`).MatchString(sql) {
			t.Fatalf("%s contains destructive SQL", path)
		}
	}
	if strings.Contains(second, "create table") || strings.Contains(second, "alter table") {
		t.Fatal("0002 must add only the lookup index")
	}
	requireSQL(t, second, `create\s+index\s+(if\s+not\s+exists\s+)?checks_owner_enabled_created_idx\s+on\s+checks\s*\(\s*owner_id\s*,\s*enabled\s*,\s*created_at\s+desc\s*\)`)
}

func TestLookupIndexPlanAligned(t *testing.T) {
	query := normalizedSQL(t, "queries/list_enabled_checks.sql")
	explain := normalizedSQL(t, "queries/explain_list_enabled_checks.sql")
	shape := []string{`select\s+id\s*,\s*owner_id\s*,\s*target\s*,\s*schedule\s*,\s*enabled\s*,\s*created_at\s*,\s*updated_at\s+from\s+checks`, `where\s+owner_id\s*=\s*\$1\s+and\s+enabled\s*=\s*true`, `order\s+by\s+created_at\s+desc`, `limit\s+\$2`}
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
