package httpserver_test

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gocheckhub/httpserver"
)

func TestRouteContract(t *testing.T) {
	handler, err := httpserver.NewHandler(httpserver.Dependencies{
		LookupTarget: func(_ context.Context, id string) (string, bool) {
			if id == "api" {
				return "primary", true
			}
			return "", false
		},
		NextRequestID: func() string { return "generated-route" },
	})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name        string
		method      string
		path        string
		wantStatus  int
		wantBody    string
		contentType string
	}{
		{name: "health", method: http.MethodGet, path: "/healthz", wantStatus: http.StatusNoContent},
		{name: "target", method: http.MethodGet, path: "/targets/api", wantStatus: http.StatusOK, wantBody: "primary\n", contentType: "text/plain; charset=utf-8"},
		{name: "missing", method: http.MethodGet, path: "/targets/unknown", wantStatus: http.StatusNotFound, wantBody: "404 page not found\n", contentType: "text/plain; charset=utf-8"},
		{name: "method", method: http.MethodPost, path: "/targets/api", wantStatus: http.StatusMethodNotAllowed, wantBody: "Method Not Allowed\n", contentType: "text/plain; charset=utf-8"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
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

	if _, err := httpserver.NewHandler(httpserver.Dependencies{}); err == nil {
		t.Fatal("NewHandler accepted missing dependencies")
	}
}

func TestRequestIDMiddlewarePropagatesContext(t *testing.T) {
	seen := ""
	generated := 0
	handler, err := httpserver.NewHandler(httpserver.Dependencies{
		LookupTarget: func(ctx context.Context, _ string) (string, bool) {
			seen = httpserver.RequestID(ctx)
			return "worker", true
		},
		NextRequestID: func() string {
			generated++
			return "generated-context"
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
		{name: "generated", want: "generated-context", wantCalled: 1},
		{name: "incoming", incoming: "  caller-42  ", want: "caller-42", wantCalled: 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			seen = ""
			request := httptest.NewRequest(http.MethodGet, "/targets/worker", nil)
			if testCase.incoming != "" {
				request.Header.Set("X-Request-ID", testCase.incoming)
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, request)
			if got := recorder.Header().Get("X-Request-ID"); got != testCase.want {
				t.Fatalf("response request ID = %q, want %q", got, testCase.want)
			}
			if seen != testCase.want || generated != testCase.wantCalled {
				t.Fatalf("context request ID = %q generated calls = %d", seen, generated)
			}
		})
	}
}

func TestNewServerConfiguresAllTimeouts(t *testing.T) {
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	timeouts := httpserver.Timeouts{
		ReadHeader: 1100 * time.Millisecond,
		Read:       2200 * time.Millisecond,
		Write:      3300 * time.Millisecond,
		Idle:       4400 * time.Millisecond,
	}
	server, err := httpserver.NewServer(handler, timeouts)
	if err != nil {
		t.Fatal(err)
	}
	if server.Handler == nil || server.ReadHeaderTimeout != timeouts.ReadHeader || server.ReadTimeout != timeouts.Read || server.WriteTimeout != timeouts.Write || server.IdleTimeout != timeouts.Idle {
		t.Fatalf("configured server = %#v", server)
	}

	invalid := []httpserver.Timeouts{
		{ReadHeader: 0, Read: time.Second, Write: time.Second, Idle: time.Second},
		{ReadHeader: time.Second, Read: 0, Write: time.Second, Idle: time.Second},
		{ReadHeader: time.Second, Read: time.Second, Write: 0, Idle: time.Second},
		{ReadHeader: time.Second, Read: time.Second, Write: time.Second, Idle: 0},
	}
	for index, value := range invalid {
		if _, err := httpserver.NewServer(handler, value); err == nil {
			t.Fatalf("NewServer accepted invalid timeout set %d", index)
		}
	}
	if _, err := httpserver.NewServer(nil, timeouts); err == nil {
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
	server, err := httpserver.NewServer(handler, httpserver.Timeouts{
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
		serveDone <- httpserver.Serve(ctx, server, listener, 2*time.Second)
	}()
	requestDone := make(chan error, 1)
	go func() {
		response, requestErr := (&http.Client{Timeout: 3 * time.Second}).Get("http://" + listener.Addr().String() + "/slow")
		if requestErr == nil {
			_, requestErr = io.Copy(io.Discard, response.Body)
			response.Body.Close()
			if requestErr == nil && response.StatusCode != http.StatusNoContent {
				requestErr = errors.New("unexpected response status")
			}
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
		if err != nil && !strings.Contains(err.Error(), "server closed") {
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
