package registry

import (
	"sync"
	"testing"
)

func TestRegistryConcurrentAccessIsRaceFree(t *testing.T) {
	registry := New()
	var group sync.WaitGroup
	group.Add(24)
	for worker := 0; worker < 24; worker++ {
		worker := worker
		go func() {
			defer group.Done()
			for iteration := 0; iteration < 300; iteration++ {
				registry.Record("shared", 1)
				registry.Record(string(rune('a'+worker%4)), 1)
				_ = registry.Snapshot()
			}
		}()
	}
	group.Wait()
	if got := registry.Snapshot()["shared"]; got != 24*300 {
		t.Fatalf("shared count = %d, want %d", got, 24*300)
	}
}
