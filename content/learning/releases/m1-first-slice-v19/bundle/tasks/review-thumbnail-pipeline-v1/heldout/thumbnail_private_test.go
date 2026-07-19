package thumbnail

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGenerateBoundsConcurrency(t *testing.T) {
	paths := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	var active atomic.Int32
	var peak atomic.Int32
	seen := make(map[string]bool)
	var seenMu sync.Mutex
	for thumbnail := range Generate(paths, 3, func(path string) string {
		current := active.Add(1)
		for {
			previous := peak.Load()
			if current <= previous || peak.CompareAndSwap(previous, current) { break }
		}
		time.Sleep(3 * time.Millisecond)
		active.Add(-1)
		return "image:" + path
	}) {
		seenMu.Lock()
		seen[thumbnail.Path] = thumbnail.Data == "image:"+thumbnail.Path
		seenMu.Unlock()
	}
	if len(seen) != len(paths) || peak.Load() < 2 || peak.Load() > 3 {
		t.Fatalf("seen=%v peak=%d", seen, peak.Load())
	}
}

func TestGenerateClosesAndReleasesWorkers(t *testing.T) {
	baseline := runtime.NumGoroutine()
	for iteration := 0; iteration < 20; iteration++ {
		for range Generate([]string{"a", "b", "c"}, 2, func(path string) string { return path }) {
		}
	}
	deadline := time.Now().Add(300 * time.Millisecond)
	for runtime.NumGoroutine() > baseline+3 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if current := runtime.NumGoroutine(); current > baseline+3 {
		t.Fatalf("goroutines after completion = %d, baseline %d", current, baseline)
	}
	for range Generate([]string{"a"}, 0, func(path string) string { return path }) {
		t.Fatal("invalid worker count produced a thumbnail")
	}
}
