package main

import (
	"context"
	"errors"
	"net/http"
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

type fakeDatabase struct {
	log      *eventLog
	closeErr error
}

func (database *fakeDatabase) Close() error {
	database.log.add("database.close")
	return database.closeErr
}

type fakeServer struct {
	log         *eventLog
	started     chan struct{}
	result      chan error
	shutdownErr error
	hasDeadline bool
}

func newFakeServer(log *eventLog) *fakeServer {
	return &fakeServer{log: log, started: make(chan struct{}), result: make(chan error, 1)}
}
func (server *fakeServer) Serve() error {
	server.log.add("server.serve")
	close(server.started)
	return <-server.result
}
func (server *fakeServer) Shutdown(ctx context.Context) error {
	server.log.add("server.shutdown")
	_, server.hasDeadline = ctx.Deadline()
	select {
	case server.result <- http.ErrServerClosed:
	default:
	}
	return server.shutdownErr
}

func TestConfigurationValidatedBeforeInitialization(t *testing.T) {
	for _, config := range []Config{{DSN: "dsn", ShutdownTimeout: time.Second}, {Address: ":8080", ShutdownTimeout: time.Second}, {Address: ":8080", DSN: "dsn"}} {
		called := false
		err := run(context.Background(), config, Dependencies{OpenDatabase: func(context.Context, string) (Database, error) { called = true; return nil, nil }})
		if err == nil || called {
			t.Fatalf("config=%#v error=%v called=%v", config, err, called)
		}
	}
}

func TestDependenciesInitializedExplicitly(t *testing.T) {
	log := &eventLog{}
	server := newFakeServer(log)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- run(ctx, validConfig(), lifecycleDependencies(log, server)) }()
	<-server.started
	cancel()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	want := []string{"database.open", "handler.build", "server.new", "server.serve", "server.shutdown", "database.close"}
	if got := log.snapshot(); !reflect.DeepEqual(got, want) {
		t.Fatalf("events=%v want=%v", got, want)
	}
	if !server.hasDeadline {
		t.Fatal("Shutdown context has no deadline")
	}
}

func TestStartupFailureRollsBackAcquiredResources(t *testing.T) {
	for _, stage := range []string{"open", "handler", "server"} {
		t.Run(stage, func(t *testing.T) {
			log := &eventLog{}
			deps := lifecycleDependencies(log, newFakeServer(log))
			failure := errors.New(stage + " failed")
			switch stage {
			case "open":
				deps.OpenDatabase = func(context.Context, string) (Database, error) { log.add("database.open"); return nil, failure }
			case "handler":
				deps.BuildHandler = func(Database) (http.Handler, error) { log.add("handler.build"); return nil, failure }
			case "server":
				deps.NewServer = func(string, http.Handler) (Server, error) { log.add("server.new"); return nil, failure }
			}
			err := run(context.Background(), validConfig(), deps)
			if !errors.Is(err, failure) {
				t.Fatalf("error=%v", err)
			}
			wantClose := stage != "open"
			closed := false
			for _, event := range log.snapshot() {
				closed = closed || event == "database.close"
			}
			if closed != wantClose {
				t.Fatalf("events=%v", log.snapshot())
			}
		})
	}
}

func TestCancellationShutdownOrderedAndBounded(t *testing.T) {
	log := &eventLog{}
	shutdownErr, closeErr := errors.New("shutdown"), errors.New("close")
	server := newFakeServer(log)
	server.shutdownErr = shutdownErr
	deps := lifecycleDependencies(log, server)
	baseOpen := deps.OpenDatabase
	deps.OpenDatabase = func(ctx context.Context, dsn string) (Database, error) {
		_, _ = baseOpen(ctx, dsn)
		return &fakeDatabase{log: log, closeErr: closeErr}, nil
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- run(ctx, validConfig(), deps) }()
	<-server.started
	cancel()
	err := <-done
	if !errors.Is(err, shutdownErr) || !errors.Is(err, closeErr) || !server.hasDeadline {
		t.Fatalf("error=%v deadline=%v", err, server.hasDeadline)
	}
	events := log.snapshot()
	if events[len(events)-2] != "server.shutdown" || events[len(events)-1] != "database.close" {
		t.Fatalf("events=%v", events)
	}
}

func TestServeFailureCleansUpResources(t *testing.T) {
	log := &eventLog{}
	server := newFakeServer(log)
	serveErr := errors.New("serve failed")
	server.result <- serveErr
	err := run(context.Background(), validConfig(), lifecycleDependencies(log, server))
	if !errors.Is(err, serveErr) {
		t.Fatalf("error=%v", err)
	}
	events := log.snapshot()
	if len(events) < 2 || events[len(events)-2] != "server.shutdown" || events[len(events)-1] != "database.close" {
		t.Fatalf("events=%v", events)
	}
}

func validConfig() Config {
	return Config{Address: "127.0.0.1:8080", DSN: "postgres://test", ShutdownTimeout: time.Second}
}
func lifecycleDependencies(log *eventLog, server *fakeServer) Dependencies {
	return Dependencies{
		OpenDatabase: func(context.Context, string) (Database, error) {
			log.add("database.open")
			return &fakeDatabase{log: log}, nil
		},
		BuildHandler: func(Database) (http.Handler, error) {
			log.add("handler.build")
			return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), nil
		},
		NewServer: func(string, http.Handler) (Server, error) { log.add("server.new"); return server, nil },
	}
}
