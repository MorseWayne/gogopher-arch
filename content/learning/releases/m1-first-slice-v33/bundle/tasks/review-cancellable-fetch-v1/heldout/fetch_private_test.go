package fetch

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestFetchAllCancelsSiblings(t *testing.T) {
	failure := errors.New("upstream unavailable")
	var started atomic.Int32
	var active atomic.Int32
	err := FetchAll(context.Background(), []string{"bad", "slow-a", "slow-b", "queued"}, 3, func(ctx context.Context, url string) error {
		started.Add(1)
		active.Add(1)
		defer active.Add(-1)
		if url == "bad" {
			deadline := time.Now().Add(300 * time.Millisecond)
			for started.Load() < 3 && time.Now().Before(deadline) { time.Sleep(time.Millisecond) }
			return failure
		}
		<-ctx.Done()
		return ctx.Err()
	})
	if !errors.Is(err, failure) || active.Load() != 0 {
		t.Fatalf("FetchAll() error=%v active=%d", err, active.Load())
	}
}

func TestFetchAllHonorsDeadlineAndReleasesWorkers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	var active atomic.Int32
	started := time.Now()
	err := FetchAll(ctx, []string{"a", "b", "c", "d"}, 2, func(ctx context.Context, url string) error {
		active.Add(1)
		defer active.Add(-1)
		<-ctx.Done()
		return ctx.Err()
	})
	if !errors.Is(err, context.DeadlineExceeded) || active.Load() != 0 || time.Since(started) > 300*time.Millisecond {
		t.Fatalf("FetchAll() error=%v active=%d duration=%s", err, active.Load(), time.Since(started))
	}
}
