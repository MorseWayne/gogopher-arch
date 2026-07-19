package releaseplan

import "testing"

func TestReleasePlanRejectsUnsafeOrIncompleteDelivery(t *testing.T) {
	valid := Config{RuntimeUser: "app", MigrationMode: "forward-only", Checks: []string{"fmt", "vet", "test", "race", "vuln", "migration", "image"}}
	if err := Validate(valid); err != nil {
		t.Fatalf("valid plan: %v", err)
	}
	cases := []Config{
		{RuntimeUser: "root", MigrationMode: valid.MigrationMode, Checks: valid.Checks},
		{RuntimeUser: valid.RuntimeUser, MigrationMode: "up-and-down", Checks: valid.Checks},
		{RuntimeUser: valid.RuntimeUser, MigrationMode: valid.MigrationMode, Checks: valid.Checks[:6]},
		{RuntimeUser: valid.RuntimeUser, MigrationMode: valid.MigrationMode, Checks: append(append([]string{}, valid.Checks...), "test")},
		{RuntimeUser: valid.RuntimeUser, MigrationMode: valid.MigrationMode, Checks: append(append([]string{}, valid.Checks[:6]...), "deploy")},
	}
	for index, testCase := range cases {
		if err := Validate(testCase); err == nil {
			t.Fatalf("case %d accepted", index)
		}
	}
}
