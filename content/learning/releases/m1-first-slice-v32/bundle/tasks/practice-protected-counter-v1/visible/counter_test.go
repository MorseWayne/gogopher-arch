package counter

import (
	"sync"
	"testing"
)

func TestCounterSupportsConcurrentAdd(t *testing.T) {
	var counter Counter
	var group sync.WaitGroup
	group.Add(32)
	for range 32 {
		go func() {
			defer group.Done()
			for range 1000 {
				counter.Add(1)
			}
		}()
	}
	group.Wait()
	if got := counter.Value(); got != 32_000 {
		t.Fatalf("Value() = %d, want 32000", got)
	}
}
