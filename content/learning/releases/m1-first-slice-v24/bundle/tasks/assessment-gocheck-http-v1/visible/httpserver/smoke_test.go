package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPublicSmokeContract(t *testing.T) {
	handler, err := NewHandler(Dependencies{
		LookupTarget:  func(context.Context, string) (string, bool) { return "", false },
		NextRequestID: func() string { return "public-1" },
	})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNoContent || recorder.Header().Get("X-Request-ID") != "public-1" {
		t.Fatalf("health response = status %d request-id %q", recorder.Code, recorder.Header().Get("X-Request-ID"))
	}

	timeouts := Timeouts{ReadHeader: time.Second, Read: 2 * time.Second, Write: 3 * time.Second, Idle: 4 * time.Second}
	server, err := NewServer(handler, timeouts)
	if err != nil {
		t.Fatal(err)
	}
	if server.Handler == nil || server.ReadHeaderTimeout != timeouts.ReadHeader || server.IdleTimeout != timeouts.Idle {
		t.Fatalf("server = %#v", server)
	}
}
