package adminserver_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"jobwatch/adminserver"
)

func TestRouteContract(t *testing.T) {
	ready := true
	handler, err := adminserver.NewHandler(adminserver.Dependencies{
		Ready: func(context.Context) bool { return ready },
		LookupJob: func(_ context.Context, id string) (string, bool) {
			if id == "job-7" {
				return "running", true
			}
			return "", false
		},
		NextTraceID: func() string { return "generated-route" },
	})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name        string
		method      string
		path        string
		ready       bool
		wantStatus  int
		wantBody    string
		contentType string
	}{
		{name: "ready", method: http.MethodGet, path: "/readyz", ready: true, wantStatus: http.StatusNoContent},
		{name: "not-ready", method: http.MethodGet, path: "/readyz", ready: false, wantStatus: http.StatusServiceUnavailable},
		{name: "job", method: http.MethodGet, path: "/jobs/job-7", ready: true, wantStatus: http.StatusOK, wantBody: "running\n", contentType: "text/plain; charset=utf-8"},
		{name: "missing", method: http.MethodGet, path: "/jobs/missing", ready: true, wantStatus: http.StatusNotFound, wantBody: "404 page not found\n", contentType: "text/plain; charset=utf-8"},
		{name: "method", method: http.MethodPost, path: "/jobs/job-7", ready: true, wantStatus: http.StatusMethodNotAllowed, wantBody: "Method Not Allowed\n", contentType: "text/plain; charset=utf-8"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			ready = testCase.ready
			request := httptest.NewRequest(testCase.method, testCase.path, nil)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			response := recorder.Result()
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != testCase.wantStatus || string(body) != testCase.wantBody {
				t.Fatalf("response = status %d body %q", response.StatusCode, body)
			}
			if got := response.Header.Get("Content-Type"); got != testCase.contentType {
				t.Fatalf("Content-Type = %q, want %q", got, testCase.contentType)
			}
		})
	}

	if _, err := adminserver.NewHandler(adminserver.Dependencies{}); err == nil {
		t.Fatal("NewHandler accepted missing dependencies")
	}
}

func TestRequestIDMiddlewarePropagatesContext(t *testing.T) {
	seen := ""
	generated := 0
	handler, err := adminserver.NewHandler(adminserver.Dependencies{
		Ready: func(context.Context) bool { return true },
		LookupJob: func(ctx context.Context, _ string) (string, bool) {
			seen = adminserver.TraceID(ctx)
			return "queued", true
		},
		NextTraceID: func() string {
			generated++
			return "generated-trace"
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, testCase := range []struct {
		name       string
		incoming   string
		want       string
		wantCalled int
	}{
		{name: "generated", want: "generated-trace", wantCalled: 1},
		{name: "incoming", incoming: "  trace-9  ", want: "trace-9", wantCalled: 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			seen = ""
			request := httptest.NewRequest(http.MethodGet, "/jobs/job-9", nil)
			if testCase.incoming != "" {
				request.Header.Set("X-Trace-ID", testCase.incoming)
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if got := recorder.Header().Get("X-Trace-ID"); got != testCase.want {
				t.Fatalf("response trace ID = %q, want %q", got, testCase.want)
			}
			if seen != testCase.want || generated != testCase.wantCalled {
				t.Fatalf("context trace ID = %q generated calls = %d", seen, generated)
			}
		})
	}
}

func TestNewServerConfiguresAllTimeouts(t *testing.T) {
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	timeouts := adminserver.Timeouts{
		ReadHeader: 1200 * time.Millisecond,
		Read:       2300 * time.Millisecond,
		Write:      3400 * time.Millisecond,
		Idle:       4500 * time.Millisecond,
	}
	server, err := adminserver.NewServer(handler, timeouts)
	if err != nil {
		t.Fatal(err)
	}
	if server.Handler == nil || server.ReadHeaderTimeout != timeouts.ReadHeader || server.ReadTimeout != timeouts.Read || server.WriteTimeout != timeouts.Write || server.IdleTimeout != timeouts.Idle {
		t.Fatalf("configured server = %#v", server)
	}
	invalid := []adminserver.Timeouts{
		{ReadHeader: 0, Read: time.Second, Write: time.Second, Idle: time.Second},
		{ReadHeader: time.Second, Read: 0, Write: time.Second, Idle: time.Second},
		{ReadHeader: time.Second, Read: time.Second, Write: 0, Idle: time.Second},
		{ReadHeader: time.Second, Read: time.Second, Write: time.Second, Idle: 0},
	}
	for index, value := range invalid {
		if _, err := adminserver.NewServer(handler, value); err == nil {
			t.Fatalf("NewServer accepted invalid timeout set %d", index)
		}
	}
	if _, err := adminserver.NewServer(nil, timeouts); err == nil {
		t.Fatal("NewServer accepted nil handler")
	}
}

func TestServeWaitsForActiveRequestAndReturns(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	started := make(chan struct{})
	release := make(chan struct{})
	released := false
	defer func() {
		if !released {
			close(release)
		}
	}()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusNoContent)
	})
	server, err := adminserver.NewServer(handler, adminserver.Timeouts{
		ReadHeader: time.Second,
		Read:       time.Second,
		Write:      time.Second,
		Idle:       time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	serveDone := make(chan error, 1)
	go func() {
		serveDone <- adminserver.Serve(ctx, server, listener, 2*time.Second)
	}()
	requestDone := make(chan error, 1)
	go func() {
		response, requestErr := (&http.Client{Timeout: 3 * time.Second}).Get("http://" + listener.Addr().String() + "/slow")
		if requestErr == nil {
			_, requestErr = io.Copy(io.Discard, response.Body)
			response.Body.Close()
		}
		requestDone <- requestErr
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not accept the active request")
	}
	cancel()
	select {
	case err := <-serveDone:
		t.Fatalf("Serve returned before the active request completed: %v", err)
	case <-time.After(75 * time.Millisecond):
	}
	close(release)
	released = true
	select {
	case err := <-requestDone:
		if err != nil {
			t.Fatalf("active request failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("active request did not finish")
	}
	select {
	case err := <-serveDone:
		if err != nil {
			t.Fatalf("Serve() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Serve did not return after graceful shutdown")
	}
}
