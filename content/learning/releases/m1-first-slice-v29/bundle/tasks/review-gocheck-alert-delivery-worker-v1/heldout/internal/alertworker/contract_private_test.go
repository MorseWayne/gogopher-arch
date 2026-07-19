package alertworker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type call struct {
	kind, id, owner string
	at              time.Time
}
type queueFake struct {
	mu      sync.Mutex
	items   []Delivery
	calls   []call
	claimed int
}

func (q *queueFake) Claim(_ context.Context, owner string, now time.Time, _ time.Duration) (Delivery, bool, error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return Delivery{}, false, nil
	}
	item := q.items[0]
	q.items = q.items[1:]
	q.claimed++
	q.calls = append(q.calls, call{"claim", item.ID, owner, now})
	return item, true, nil
}
func (q *queueFake) Ack(_ context.Context, id, owner string) error {
	q.add("ack", id, owner, time.Time{})
	return nil
}
func (q *queueFake) Retry(_ context.Context, id, owner string, at time.Time) error {
	q.add("retry", id, owner, at)
	return nil
}
func (q *queueFake) Fail(_ context.Context, id, owner string) error {
	q.add("fail", id, owner, time.Time{})
	return nil
}
func (q *queueFake) add(kind, id, owner string, at time.Time) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.calls = append(q.calls, call{kind, id, owner, at})
}
func (q *queueFake) last() call { q.mu.Lock(); defer q.mu.Unlock(); return q.calls[len(q.calls)-1] }

type sendFunc func(context.Context, Delivery) error

func (f sendFunc) Send(ctx context.Context, d Delivery) error { return f(ctx, d) }

type sendError struct{ temporary bool }

func (e sendError) Error() string   { return "send" }
func (e sendError) Temporary() bool { return e.temporary }
func opts(now time.Time) Options {
	return Options{Owner: "alerts-2", Concurrency: 2, Lease: time.Second, SendTimeout: time.Second / 2, PollInterval: time.Millisecond, MaxAttempts: 3, RetryDelay: func(attempt int) time.Duration { return time.Duration(attempt) * time.Second }, Now: func() time.Time { return now }}
}
func TestDeliveryOutcomes(t *testing.T) {
	now := time.Unix(40, 0)
	tests := []struct {
		name     string
		delivery Delivery
		failure  error
		want     string
	}{{"success", Delivery{ID: "a", Attempt: 1}, nil, "ack"}, {"retry", Delivery{ID: "b", Attempt: 2}, sendError{true}, "retry"}, {"permanent", Delivery{ID: "c", Attempt: 1}, sendError{}, "fail"}, {"exhausted", Delivery{ID: "d", Attempt: 3}, sendError{true}, "fail"}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			q := &queueFake{items: []Delivery{test.delivery}}
			worker, _ := New(q, sendFunc(func(context.Context, Delivery) error { return test.failure }), opts(now))
			if _, err := worker.RunOnce(t.Context()); err != nil {
				t.Fatal(err)
			}
			last := q.last()
			if last.kind != test.want {
				t.Fatalf("last=%+v", last)
			}
			if test.want == "retry" && !last.at.Equal(now.Add(2*time.Second)) {
				t.Fatalf("retry=%s", last.at)
			}
		})
	}
}
func TestDuplicateDeliveryIsNotSent(t *testing.T) {
	q := &queueFake{items: []Delivery{{ID: "dup", Duplicate: true}}}
	worker, _ := New(q, sendFunc(func(context.Context, Delivery) error { t.Fatal("duplicate sent"); return nil }), opts(time.Now()))
	if _, err := worker.RunOnce(t.Context()); err != nil || q.last().kind != "ack" {
		t.Fatalf("last=%+v err=%v", q.last(), err)
	}
}
func TestBacklogIsBoundedAndShutdownJoins(t *testing.T) {
	items := make([]Delivery, 8)
	for i := range items {
		items[i] = Delivery{ID: string(rune('a' + i)), Attempt: 1}
	}
	q := &queueFake{items: items}
	ctx, cancel := context.WithCancel(t.Context())
	var active, max atomic.Int32
	started := make(chan struct{}, 8)
	sender := sendFunc(func(ctx context.Context, _ Delivery) error {
		n := active.Add(1)
		defer active.Add(-1)
		for {
			old := max.Load()
			if n <= old || max.CompareAndSwap(old, n) {
				break
			}
		}
		started <- struct{}{}
		<-ctx.Done()
		return ctx.Err()
	})
	configured := opts(time.Now())
	configured.Concurrency = 2
	worker, _ := New(q, sender, configured)
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()
	<-started
	<-started
	time.Sleep(10 * time.Millisecond)
	q.mu.Lock()
	claimed := q.claimed
	q.mu.Unlock()
	if claimed != 2 || max.Load() > 2 {
		t.Fatalf("claimed=%d max=%d", claimed, max.Load())
	}
	cancel()
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("worker did not stop")
	}
	if active.Load() != 0 {
		t.Fatalf("active=%d", active.Load())
	}
}
func TestReplacementOwnerReclaimsExpiredLease(t *testing.T) {
	now := time.Unix(50, 0)
	q := &queueFake{items: []Delivery{{ID: "recover", Attempt: 2}}}
	worker, _ := New(q, sendFunc(func(context.Context, Delivery) error { return nil }), opts(now))
	if _, err := worker.RunOnce(t.Context()); err != nil {
		t.Fatal(err)
	}
	q.mu.Lock()
	first := q.calls[0]
	q.mu.Unlock()
	if first.owner != "alerts-2" || !first.at.Equal(now) {
		t.Fatalf("claim=%+v", first)
	}
}
