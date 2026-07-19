package checkworker

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type storeCall struct {
	kind, id, owner string
	at              time.Time
}
type fakeStore struct {
	mu      sync.Mutex
	tasks   []Task
	calls   []storeCall
	claimed int
}

func (store *fakeStore) Claim(_ context.Context, owner string, now time.Time, _ time.Duration) (Task, bool, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if len(store.tasks) == 0 {
		return Task{}, false, nil
	}
	task := store.tasks[0]
	store.tasks = store.tasks[1:]
	store.claimed++
	store.calls = append(store.calls, storeCall{kind: "claim", id: task.ID, owner: owner, at: now})
	return task, true, nil
}
func (store *fakeStore) Ack(_ context.Context, id, owner string) error {
	store.record("ack", id, owner, time.Time{})
	return nil
}
func (store *fakeStore) Retry(_ context.Context, id, owner string, at time.Time) error {
	store.record("retry", id, owner, at)
	return nil
}
func (store *fakeStore) Fail(_ context.Context, id, owner string) error {
	store.record("fail", id, owner, time.Time{})
	return nil
}
func (store *fakeStore) record(kind, id, owner string, at time.Time) {
	store.mu.Lock()
	defer store.mu.Unlock()
	store.calls = append(store.calls, storeCall{kind: kind, id: id, owner: owner, at: at})
}
func (store *fakeStore) last() storeCall {
	store.mu.Lock()
	defer store.mu.Unlock()
	return store.calls[len(store.calls)-1]
}

type processFunc func(context.Context, Task) error

func (function processFunc) Process(ctx context.Context, task Task) error { return function(ctx, task) }

type temporaryError struct{ retry bool }

func (failure temporaryError) Error() string   { return "process failed" }
func (failure temporaryError) Temporary() bool { return failure.retry }
func options(now time.Time) Options {
	return Options{Owner: "worker-a", Concurrency: 2, Lease: time.Second, ProcessTimeout: 500 * time.Millisecond, PollInterval: time.Millisecond, MaxAttempts: 3, RetryDelay: func(attempt int) time.Duration { return time.Duration(attempt) * time.Minute }, Now: func() time.Time { return now }}
}

func TestSuccessfulTaskAcknowledged(t *testing.T) {
	store := &fakeStore{tasks: []Task{{ID: "t1", Key: "check:1", Attempt: 1}}}
	called := 0
	worker, err := New(store, processFunc(func(context.Context, Task) error { called++; return nil }), options(time.Unix(10, 0)))
	if err != nil {
		t.Fatal(err)
	}
	processed, err := worker.RunOnce(t.Context())
	if err != nil || !processed || called != 1 || store.last().kind != "ack" {
		t.Fatalf("processed=%v called=%d last=%+v err=%v", processed, called, store.last(), err)
	}
}
func TestDuplicateAcknowledgedWithoutSideEffect(t *testing.T) {
	store := &fakeStore{tasks: []Task{{ID: "t2", Key: "check:1", Attempt: 2, Duplicate: true}}}
	worker, _ := New(store, processFunc(func(context.Context, Task) error { t.Fatal("duplicate processed"); return nil }), options(time.Now()))
	if processed, err := worker.RunOnce(t.Context()); err != nil || !processed || store.last().kind != "ack" {
		t.Fatalf("processed=%v last=%+v err=%v", processed, store.last(), err)
	}
}
func TestTemporaryFailureScheduledForRetry(t *testing.T) {
	now := time.Unix(20, 0)
	store := &fakeStore{tasks: []Task{{ID: "t3", Attempt: 2}}}
	worker, _ := New(store, processFunc(func(context.Context, Task) error { return temporaryError{retry: true} }), options(now))
	if _, err := worker.RunOnce(t.Context()); err != nil {
		t.Fatal(err)
	}
	last := store.last()
	if last.kind != "retry" || !last.at.Equal(now.Add(2*time.Minute)) {
		t.Fatalf("last=%+v", last)
	}
}
func TestPermanentAndExhaustedFailuresAreTerminal(t *testing.T) {
	for _, test := range []struct {
		name    string
		task    Task
		failure error
	}{{"permanent", Task{ID: "p", Attempt: 1}, temporaryError{}}, {"exhausted", Task{ID: "e", Attempt: 3}, temporaryError{retry: true}}} {
		t.Run(test.name, func(t *testing.T) {
			store := &fakeStore{tasks: []Task{test.task}}
			worker, _ := New(store, processFunc(func(context.Context, Task) error { return test.failure }), options(time.Now()))
			if _, err := worker.RunOnce(t.Context()); err != nil {
				t.Fatal(err)
			}
			if store.last().kind != "fail" {
				t.Fatalf("last=%+v", store.last())
			}
		})
	}
}
func TestRunAppliesBackpressureAndStopsOnCancellation(t *testing.T) {
	tasks := make([]Task, 12)
	for index := range tasks {
		tasks[index] = Task{ID: string(rune('a' + index)), Attempt: 1}
	}
	store := &fakeStore{tasks: tasks}
	ctx, cancel := context.WithCancel(t.Context())
	var active, maxActive atomic.Int32
	started := make(chan struct{}, 12)
	release := make(chan struct{})
	processor := processFunc(func(ctx context.Context, _ Task) error {
		current := active.Add(1)
		defer active.Add(-1)
		for {
			old := maxActive.Load()
			if current <= old || maxActive.CompareAndSwap(old, current) {
				break
			}
		}
		started <- struct{}{}
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	configured := options(time.Now())
	configured.Concurrency = 3
	worker, _ := New(store, processor, configured)
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()
	for range 3 {
		<-started
	}
	time.Sleep(10 * time.Millisecond)
	store.mu.Lock()
	claimed := store.claimed
	store.mu.Unlock()
	if claimed != 3 || maxActive.Load() > 3 {
		t.Fatalf("claimed=%d max=%d", claimed, maxActive.Load())
	}
	cancel()
	close(release)
	select {
	case err := <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("Run did not join workers")
	}
	if active.Load() != 0 {
		t.Fatalf("active=%d", active.Load())
	}
}
func TestExpiredLeaseCanBeReclaimedAfterRestart(t *testing.T) {
	now := time.Unix(30, 0)
	store := &fakeStore{tasks: []Task{{ID: "leased", Attempt: 2}}}
	configured := options(now)
	configured.Owner = "replacement"
	worker, _ := New(store, processFunc(func(context.Context, Task) error { return nil }), configured)
	if _, err := worker.RunOnce(t.Context()); err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	first := store.calls[0]
	store.mu.Unlock()
	if first.kind != "claim" || first.owner != "replacement" || !first.at.Equal(now) {
		t.Fatalf("claim=%+v", first)
	}
}
