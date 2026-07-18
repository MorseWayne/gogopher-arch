package registry

import "testing"

func TestRegistrySnapshotOwnsState(t *testing.T) {
	registry := New()
	registry.Record("success", 4)
	registry.Record("failed", 1)
	first := registry.Snapshot()
	first["success"] = 100
	first["new"] = 7
	second := registry.Snapshot()
	if second["success"] != 4 || second["failed"] != 1 {
		t.Fatalf("second snapshot = %v", second)
	}
	if _, exists := second["new"]; exists {
		t.Fatalf("snapshot mutation reached registry: %v", second)
	}
}
