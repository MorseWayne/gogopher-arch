package checkmodel

import "testing"

func TestCheckProtectsInvariantsAndTransitions(t *testing.T) {
	if _, err := NewCheck("", 2); err == nil {
		t.Fatal("NewCheck(empty name) error = nil")
	}
	check, err := NewCheck("api", 2)
	if err != nil {
		t.Fatal(err)
	}
	check.Record(false)
	if check.Health() != HealthHealthy {
		t.Fatalf("health after one failure = %q", check.Health())
	}
	check.Record(false)
	if check.Health() != HealthUnhealthy || check.String() != "api:unhealthy" {
		t.Fatalf("check after two failures = %#v, %q", check, check.String())
	}
	check.Record(true)
	if check.Health() != HealthHealthy {
		t.Fatalf("health after success = %q", check.Health())
	}
}
