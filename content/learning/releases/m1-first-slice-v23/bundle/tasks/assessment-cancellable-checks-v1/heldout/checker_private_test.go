package checker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestCheckAllCancelsSiblings(t *testing.T) {
	failure := errors.New("dependency rejected")
	var started atomic.Int32
	var active atomic.Int32
	check := func(ctx context.Context, target string) error {
		started.Add(1)
		active.Add(1)
		defer active.Add(-1)
		if target == "fail" {
			deadline := time.Now().Add(300 * time.Millisecond)
			for started.Load() < 3 && time.Now().Before(deadline) {
				time.Sleep(time.Millisecond)
			}
			return failure
		}
		<-ctx.Done()
		return ctx.Err()
	}
	err := CheckAll(context.Background(), []string{"fail", "slow-a", "slow-b", "queued"}, 3, check)
	if !errors.Is(err, failure) {
		t.Fatalf("CheckAll() error = %v, want wrapped failure", err)
	}
	if active.Load() != 0 {
		t.Fatalf("active checks after return = %d", active.Load())
	}
}

func TestCheckAllHonorsDeadlineAndReleasesWorkers(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	var active atomic.Int32
	started := time.Now()
	err := CheckAll(ctx, []string{"a", "b", "c", "d", "e"}, 3, func(ctx context.Context, target string) error {
		active.Add(1)
		defer active.Add(-1)
		<-ctx.Done()
		return ctx.Err()
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("CheckAll() error = %v, want deadline exceeded", err)
	}
	if elapsed := time.Since(started); elapsed > 300*time.Millisecond {
		t.Fatalf("CheckAll() returned after %s", elapsed)
	}
	if active.Load() != 0 {
		t.Fatalf("active checks after return = %d", active.Load())
	}
}
