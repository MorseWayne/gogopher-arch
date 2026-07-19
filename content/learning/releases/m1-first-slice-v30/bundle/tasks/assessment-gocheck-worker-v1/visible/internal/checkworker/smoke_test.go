package checkworker

import (
	"context"
	"testing"
	"time"
)

type emptyStore struct{}

func (emptyStore) Claim(context.Context, string, time.Time, time.Duration) (Task, bool, error) {
	return Task{}, false, nil
}
func (emptyStore) Ack(context.Context, string, string) error              { return nil }
func (emptyStore) Retry(context.Context, string, string, time.Time) error { return nil }
func (emptyStore) Fail(context.Context, string, string) error             { return nil }

type noopProcessor struct{}

func (noopProcessor) Process(context.Context, Task) error { return nil }

func TestRunOnceWithEmptyStore(t *testing.T) {
	worker, err := New(emptyStore{}, noopProcessor{}, Options{Owner: "worker-1", Concurrency: 1, Lease: time.Second, ProcessTimeout: time.Second / 2, PollInterval: time.Millisecond, MaxAttempts: 3, RetryDelay: func(int) time.Duration { return time.Second }, Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	processed, err := worker.RunOnce(t.Context())
	if err != nil || processed {
		t.Fatalf("processed=%v err=%v", processed, err)
	}
}
