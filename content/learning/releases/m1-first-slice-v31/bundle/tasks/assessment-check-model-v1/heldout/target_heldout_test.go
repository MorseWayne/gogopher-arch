package targetmodel

import "testing"

func TestTargetEnforcesConstructionInvariants(t *testing.T) {
	for _, input := range []struct {
		name  string
		limit int
	}{{"", 1}, {"   ", 1}, {"api", 0}, {"api", -1}} {
		if _, err := NewTarget(input.name, input.limit); err == nil {
			t.Fatalf("NewTarget(%q, %d) error = nil", input.name, input.limit)
		}
	}
	target, err := NewTarget(" worker ", 3)
	if err != nil || target.Label() != "worker:ready" {
		t.Fatalf("valid target = %#v, %v", target, err)
	}
}

func TestTargetBehaviorMethodsDriveTransitions(t *testing.T) {
	target, err := NewTarget("db", 2)
	if err != nil {
		t.Fatal(err)
	}
	target.RecordFailure()
	target.RecordFailure()
	target.RecordFailure()
	if target.State() != StateOpen || target.Label() != "db:open" {
		t.Fatalf("open target = %#v label=%q", target, target.Label())
	}
	target.RecordSuccess()
	if target.State() != StateReady || target.Label() != "db:ready" {
		t.Fatalf("recovered target = %#v label=%q", target, target.Label())
	}
}
