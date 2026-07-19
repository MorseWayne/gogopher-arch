package integration_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"endpointaudit/internal/app"
	"endpointaudit/internal/output"
	"endpointaudit/internal/probe"
	"endpointaudit/internal/spec"
)

func TestInputContractAndCLIExitCodes(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.json")
	_, err := spec.Load(missing)
	var pathError *os.PathError
	if !errors.As(err, &pathError) {
		t.Fatalf("Load(missing) error = %v", err)
	}
	invalid := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(invalid, []byte(`{"checks":[{"name":"api","url":"ftp://api.example","accept_status":99}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := spec.Load(invalid); err == nil {
		t.Fatal("Load(invalid) error = nil")
	}

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	path := filepath.Join(t.TempDir(), "checks.json")
	body := fmt.Sprintf(`{"checks":[{"name":"health","url":%q,"accept_status":204}]}`, server.URL)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	dependencies := app.Dependencies{Client: probe.Client(server.Client())}
	if code := app.Run(context.Background(), []string{"-spec", path, "-workers", "1", "-output", "text"}, &stdout, &stderr, dependencies); code != 0 || stdout.String() != "health\tpass\t204\t204\n" {
		t.Fatalf("Run() code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if code := app.Run(context.Background(), []string{"-spec", missing}, &stdout, &stderr, dependencies); code != 2 || stderr.Len() == 0 {
		t.Fatalf("Run(missing) code=%d stdout=%q stderr=%q", code, stdout.String(), stderr.String())
	}
}

func TestConcurrentWorkflowAndStableOutput(t *testing.T) {
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
		time.Sleep(10 * time.Millisecond)
		writer.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	checks := []probe.Check{{Name: "one", URL: server.URL, AcceptStatus: 201}, {Name: "two", URL: server.URL, AcceptStatus: 204}, {Name: "three", URL: server.URL, AcceptStatus: 201}}
	results, err := probe.All(context.Background(), server.Client(), checks, 2, time.Second)
	if err != nil || maximum.Load() > 2 || len(results) != 3 || results[0].Name != "one" {
		t.Fatalf("All() results=%#v error=%v max=%d", results, err, maximum.Load())
	}
	if got, want := output.Text(results), "one\tpass\t201\t201\ntwo\tfail\t204\t201\nthree\tpass\t201\t201\n"; got != want {
		t.Fatalf("Text() = %q, want %q", got, want)
	}
	encoded, err := output.JSON(results[:1])
	if err != nil || !strings.HasSuffix(encoded, "\n") || !strings.Contains(encoded, `"expected":201`) {
		t.Fatalf("JSON() = %q, %v", encoded, err)
	}

	started := make(chan struct{}, 1)
	blocking := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started <- struct{}{}
		<-request.Context().Done()
	}))
	defer blocking.Close()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := probe.All(ctx, blocking.Client(), []probe.Check{{Name: "one", URL: blocking.URL, AcceptStatus: 200}, {Name: "two", URL: blocking.URL, AcceptStatus: 200}}, 1, time.Second)
		done <- err
	}()
	<-started
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("All(cancel) error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("All() leaked work after cancellation")
	}
}
