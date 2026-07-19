package workerpool

import (
	"runtime"
	"sync/atomic"
	"testing"
	"time"
)

func TestProcessBoundsConcurrency(t *testing.T) {
	var active atomic.Int32
	var peak atomic.Int32
	values := make([]int, 24)
	for index := range values {
		values[index] = index
	}
	count := 0
	for range Process(values, 4, func(value int) int {
		current := active.Add(1)
		updatePeak(&peak, current)
		time.Sleep(2 * time.Millisecond)
		active.Add(-1)
		return value
	}) {
		count++
	}
	if count != len(values) {
		t.Fatalf("received %d results, want %d", count, len(values))
	}
	if observed := peak.Load(); observed < 2 || observed > 4 {
		t.Fatalf("peak concurrency = %d, want 2..4", observed)
	}
}

func TestProcessClosesAndReleasesWorkers(t *testing.T) {
	baseline := runtime.NumGoroutine()
	for iteration := 0; iteration < 20; iteration++ {
		count := 0
		for range Process([]int{1, 2, 3, 4, 5}, 3, func(value int) int { return value }) {
			count++
		}
		if count != 5 {
			t.Fatalf("iteration %d received %d results", iteration, count)
		}
	}
	deadline := time.Now().Add(300 * time.Millisecond)
	for runtime.NumGoroutine() > baseline+3 && time.Now().Before(deadline) {
		runtime.Gosched()
		time.Sleep(time.Millisecond)
	}
	if current := runtime.NumGoroutine(); current > baseline+3 {
		t.Fatalf("goroutines after completion = %d, baseline %d", current, baseline)
	}
	for range Process([]int{1}, 0, func(value int) int { return value }) {
		t.Fatal("invalid worker count produced a result")
	}
}

func updatePeak(peak *atomic.Int32, current int32) {
	for {
		previous := peak.Load()
		if current <= previous || peak.CompareAndSwap(previous, current) {
			return
		}
	}
}
