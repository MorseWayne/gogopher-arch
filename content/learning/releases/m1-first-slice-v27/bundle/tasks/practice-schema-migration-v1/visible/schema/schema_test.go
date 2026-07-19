package schema

import (
	"os"
	"strings"
	"testing"
)

func TestProjectSchemaAndLookup(t *testing.T) {
	migration := read(t, "../migrations/0001_create_projects.up.sql")
	query := read(t, "../queries/list_active_projects.sql")
	for _, required := range []string{"create table projects", "primary key", "not null", "check", "unique", "create index", "owner_id", "active", "created_at desc"} {
		if !strings.Contains(normalize(migration), required) {
			t.Fatalf("migration must contain %q", required)
		}
	}
	for _, required := range []string{"where owner_id = $1", "active = true", "order by created_at desc"} {
		if !strings.Contains(normalize(query), required) {
			t.Fatalf("query must contain %q", required)
		}
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func normalize(value string) string { return strings.Join(strings.Fields(strings.ToLower(value)), " ") }
