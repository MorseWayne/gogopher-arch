package cancellable

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMapPreservesOrderAndStopsOnCancellation(t *testing.T) {
	got, err := Map(context.Background(), []int{1, 2, 3}, 2, func(_ context.Context, value int) (int, error) {
		return value * 2, nil
	})
	if err != nil || len(got) != 3 || got[0] != 2 || got[1] != 4 || got[2] != 6 {
		t.Fatalf("Map() = %v, %v", got, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	started := time.Now()
	_, err = Map(ctx, []int{1, 2, 3}, 2, func(ctx context.Context, value int) (int, error) {
		<-ctx.Done()
		return 0, ctx.Err()
	})
	if !errors.Is(err, context.Canceled) || time.Since(started) > 200*time.Millisecond {
		t.Fatalf("cancelled Map() error = %v, duration = %s", err, time.Since(started))
	}
}
