package targetmodel

import "testing"

func TestTargetOpensAtFailureLimit(t *testing.T) {
	target, err := NewTarget(" api ", 2)
	if err != nil {
		t.Fatal(err)
	}
	target.RecordFailure()
	if target.State() != StateReady {
		t.Fatalf("state after one failure = %q", target.State())
	}
	target.RecordFailure()
	if target.State() != StateOpen || target.Label() != "api:open" {
		t.Fatalf("target at limit = %#v label=%q", target, target.Label())
	}
}
