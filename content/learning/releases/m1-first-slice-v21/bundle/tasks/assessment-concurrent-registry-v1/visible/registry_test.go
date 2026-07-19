package registry

import "testing"

func TestRegistryRecordsAndCopiesSnapshot(t *testing.T) {
	registry := New()
	registry.Record("ok", 2)
	registry.Record("ok", 3)
	snapshot := registry.Snapshot()
	if snapshot["ok"] != 5 {
		t.Fatalf("snapshot = %v", snapshot)
	}
	snapshot["ok"] = 99
	if got := registry.Snapshot()["ok"]; got != 5 {
		t.Fatalf("Snapshot() shares internal state: %d", got)
	}
}
