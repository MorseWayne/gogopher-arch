package alertworker

import (
	"context"
	"testing"
	"time"
)

type emptyQueue struct{}

func (emptyQueue) Claim(context.Context, string, time.Time, time.Duration) (Delivery, bool, error) {
	return Delivery{}, false, nil
}
func (emptyQueue) Ack(context.Context, string, string) error              { return nil }
func (emptyQueue) Retry(context.Context, string, string, time.Time) error { return nil }
func (emptyQueue) Fail(context.Context, string, string) error             { return nil }

type noopSender struct{}

func (noopSender) Send(context.Context, Delivery) error { return nil }
func TestEmptyQueue(t *testing.T) {
	worker, err := New(emptyQueue{}, noopSender{}, Options{Owner: "alerts", Concurrency: 1, Lease: time.Second, SendTimeout: time.Second / 2, PollInterval: time.Millisecond, MaxAttempts: 3, RetryDelay: func(int) time.Duration { return time.Second }, Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	if processed, err := worker.RunOnce(t.Context()); err != nil || processed {
		t.Fatalf("processed=%v err=%v", processed, err)
	}
}
