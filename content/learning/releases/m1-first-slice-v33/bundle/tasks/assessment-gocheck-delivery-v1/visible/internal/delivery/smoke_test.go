package delivery

import "testing"

func TestMigrationSequenceSmoke(t *testing.T) {
	if err := ValidateMigrations([]string{"0001_create_checks.up.sql", "0002_add_status.up.sql"}); err != nil {
		t.Fatal(err)
	}
}
