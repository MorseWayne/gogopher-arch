package schema

import (
	"os"
	"strings"
	"testing"
)

func TestAlertMigrationAssetsCompleted(t *testing.T) {
	paths := []string{"../../migrations/0001_create_alert_rules.up.sql", "../../migrations/0002_add_alert_rules_lookup_index.up.sql", "../../queries/list_active_alert_rules.sql", "../../queries/explain_list_active_alert_rules.sql"}
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
