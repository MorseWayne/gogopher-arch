package releaseplan

import "testing"

func TestValidateAcceptsCompleteForwardOnlyPlan(t *testing.T) {
	config := Config{RuntimeUser: "app", MigrationMode: "forward-only", Checks: []string{"fmt", "vet", "test", "race", "vuln", "migration", "image"}}
	if err := Validate(config); err != nil {
		t.Fatal(err)
	}
}
