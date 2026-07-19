package check

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestAllBoundsConcurrencyAndReleasesOnCancel(t *testing.T) {
	var active atomic.Int32
	var maximum atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		current := active.Add(1)
		defer active.Add(-1)
		for {
			previous := maximum.Load()
			if current <= previous || maximum.CompareAndSwap(previous, current) {
				break
			}
		}
		time.Sleep(15 * time.Millisecond)
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	targets := make([]Target, 8)
	for index := range targets {
		targets[index] = Target{Name: string(rune('a' + index)), URL: server.URL}
	}
	results, err := All(context.Background(), server.Client(), targets, 2, time.Second)
	if err != nil || len(results) != len(targets) || maximum.Load() > 2 {
		t.Fatalf("All() results=%#v error=%v maximum=%d", results, err, maximum.Load())
	}
	for index := range results {
		if results[index].Name != targets[index].Name {
			t.Fatalf("result order = %#v", results)
		}
	}

	started := make(chan struct{}, 2)
	blocking := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started <- struct{}{}
		<-request.Context().Done()
	}))
	defer blocking.Close()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := All(ctx, blocking.Client(), []Target{{Name: "one", URL: blocking.URL}, {Name: "two", URL: blocking.URL}, {Name: "three", URL: blocking.URL}}, 2, time.Second)
		done <- err
	}()
	<-started
	<-started
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("All(cancel) error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("All() did not release workers after cancellation")
	}
}
