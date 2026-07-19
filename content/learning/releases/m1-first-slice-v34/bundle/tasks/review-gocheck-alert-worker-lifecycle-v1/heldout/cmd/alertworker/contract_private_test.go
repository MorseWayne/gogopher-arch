package main

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"
)

type eventLog struct {
	mu     sync.Mutex
	values []string
}

func (log *eventLog) add(value string) {
	log.mu.Lock()
	defer log.mu.Unlock()
	log.values = append(log.values, value)
}
func (log *eventLog) snapshot() []string {
	log.mu.Lock()
	defer log.mu.Unlock()
	return append([]string(nil), log.values...)
}

type fakeStore struct {
	log      *eventLog
	closeErr error
}

func (store *fakeStore) Close() error { store.log.add("store.close"); return store.closeErr }

type fakeWorker struct {
	log         *eventLog
	started     chan struct{}
	result      chan error
	shutdownErr error
	hasDeadline bool
}

func newFakeWorker(log *eventLog) *fakeWorker {
	return &fakeWorker{log: log, started: make(chan struct{}), result: make(chan error, 1)}
}
func (worker *fakeWorker) Run() error {
	worker.log.add("worker.run")
	close(worker.started)
	return <-worker.result
}
func (worker *fakeWorker) Shutdown(ctx context.Context) error {
	worker.log.add("worker.shutdown")
	_, worker.hasDeadline = ctx.Deadline()
	select {
	case worker.result <- nil:
	default:
	}
	return worker.shutdownErr
}

func TestConfigurationValidatedBeforeInitialization(t *testing.T) {
	for _, config := range []Config{{Concurrency: 1, ShutdownTimeout: time.Second}, {DSN: "dsn", ShutdownTimeout: time.Second}, {DSN: "dsn", Concurrency: 1}} {
		called := false
		err := run(context.Background(), config, Dependencies{OpenStore: func(context.Context, string) (Store, error) { called = true; return nil, nil }})
		if err == nil || called {
			t.Fatalf("config=%#v error=%v called=%v", config, err, called)
		}
	}
}

func TestDependenciesInitializedExplicitly(t *testing.T) {
	log := &eventLog{}
	worker := newFakeWorker(log)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- run(ctx, validConfig(), lifecycleDependencies(log, worker)) }()
	<-worker.started
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	want := []string{"store.open", "worker.new", "worker.run", "worker.shutdown", "store.close"}
	if got := log.snapshot(); !reflect.DeepEqual(got, want) {
		t.Fatalf("events=%v want=%v", got, want)
	}
	if !worker.hasDeadline {
		t.Fatal("Shutdown context has no deadline")
	}
}

func TestStartupFailureRollsBackAcquiredResources(t *testing.T) {
	for _, stage := range []string{"open", "worker"} {
		t.Run(stage, func(t *testing.T) {
			log := &eventLog{}
			deps := lifecycleDependencies(log, newFakeWorker(log))
			failure := errors.New(stage + " failed")
			if stage == "open" {
				deps.OpenStore = func(context.Context, string) (Store, error) { log.add("store.open"); return nil, failure }
			} else {
				deps.NewWorker = func(Store, int) (Worker, error) { log.add("worker.new"); return nil, failure }
			}
			err := run(context.Background(), validConfig(), deps)
			if !errors.Is(err, failure) {
				t.Fatalf("error=%v", err)
			}
			closed := false
			for _, event := range log.snapshot() {
				closed = closed || event == "store.close"
			}
			if closed != (stage != "open") {
				t.Fatalf("events=%v", log.snapshot())
			}
		})
	}
}

func TestCancellationShutdownOrderedAndBounded(t *testing.T) {
	log := &eventLog{}
	shutdownErr, closeErr := errors.New("shutdown"), errors.New("close")
	worker := newFakeWorker(log)
	worker.shutdownErr = shutdownErr
	deps := lifecycleDependencies(log, worker)
	deps.OpenStore = func(context.Context, string) (Store, error) {
		log.add("store.open")
		return &fakeStore{log: log, closeErr: closeErr}, nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- run(ctx, validConfig(), deps) }()
	<-worker.started
	cancel()
	err := <-done
	if !errors.Is(err, shutdownErr) || !errors.Is(err, closeErr) || !worker.hasDeadline {
		t.Fatalf("error=%v deadline=%v", err, worker.hasDeadline)
	}
	events := log.snapshot()
	if events[len(events)-2] != "worker.shutdown" || events[len(events)-1] != "store.close" {
		t.Fatalf("events=%v", events)
	}
}

func TestServeFailureCleansUpResources(t *testing.T) {
	log := &eventLog{}
	worker := newFakeWorker(log)
	runErr := errors.New("run failed")
	worker.result <- runErr
	err := run(context.Background(), validConfig(), lifecycleDependencies(log, worker))
	if !errors.Is(err, runErr) {
		t.Fatalf("error=%v", err)
	}
	events := log.snapshot()
	if len(events) < 2 || events[len(events)-2] != "worker.shutdown" || events[len(events)-1] != "store.close" {
		t.Fatalf("events=%v", events)
	}
}

func validConfig() Config {
	return Config{DSN: "postgres://test", Concurrency: 4, ShutdownTimeout: time.Second}
}
func lifecycleDependencies(log *eventLog, worker *fakeWorker) Dependencies {
	return Dependencies{OpenStore: func(context.Context, string) (Store, error) { log.add("store.open"); return &fakeStore{log: log}, nil }, NewWorker: func(Store, int) (Worker, error) { log.add("worker.new"); return worker, nil }}
}
