package schema

import (
	"os"
	"strings"
	"testing"
)

func TestMigrationAssetsCompleted(t *testing.T) {
	paths := []string{
		"../../migrations/0001_create_checks.up.sql",
		"../../migrations/0002_add_checks_lookup_index.up.sql",
		"../../queries/list_enabled_checks.sql",
		"../../queries/explain_list_enabled_checks.sql",
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(strings.ToUpper(string(data)), "TODO") {
			t.Fatalf("%s is incomplete", path)
		}
	}
}
