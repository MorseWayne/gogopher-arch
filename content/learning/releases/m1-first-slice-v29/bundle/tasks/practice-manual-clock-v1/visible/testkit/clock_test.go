package testkit

import (
	"sync"
	"testing"
	"time"
)

func TestManualClockIsDeterministicAndConcurrentSafe(t *testing.T) {
	start := time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)
	clock := NewManualClock(start)
	if got := clock.Now(); !got.Equal(start) {
		t.Fatalf("Now()=%v want %v", got, start)
	}
	if err := clock.Advance(250 * time.Millisecond); err != nil {
		t.Fatal(err)
	}
	if got := clock.Now(); !got.Equal(start.Add(250 * time.Millisecond)) {
		t.Fatalf("Now()=%v", got)
	}
	if err := clock.Advance(-time.Nanosecond); err == nil {
		t.Fatal("negative Advance succeeded")
	}
	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			for range 100 {
				_ = clock.Now()
			}
		}()
	}
	group.Wait()
}
