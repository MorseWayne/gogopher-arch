package pipeline

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestMapUsesBoundedWorkersAndPreservesOrder(t *testing.T) {
	var active atomic.Int32
	var peak atomic.Int32
	values := []int{1, 2, 3, 4, 5, 6, 7, 8}
	got := Map(values, 3, func(value int) int {
		current := active.Add(1)
		for current > peak.Load() && !peak.CompareAndSwap(peak.Load(), current) {
		}
		time.Sleep(5 * time.Millisecond)
		active.Add(-1)
		return value * value
	})
	want := []int{1, 4, 9, 16, 25, 36, 49, 64}
	if len(got) != len(want) {
		t.Fatalf("Map() length = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("Map()[%d] = %d, want %d", index, got[index], want[index])
		}
	}
	if observed := peak.Load(); observed < 2 || observed > 3 {
		t.Fatalf("peak concurrency = %d, want 2..3", observed)
	}
	if active.Load() != 0 {
		t.Fatalf("active workers after return = %d", active.Load())
	}
}

func TestMapHandlesInvalidWorkerCount(t *testing.T) {
	if got := Map([]int{1}, 0, func(value int) int { return value }); len(got) != 0 {
		t.Fatalf("Map() = %v, want empty result", got)
	}
}
