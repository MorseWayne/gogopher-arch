package httpslice

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerRoutesAndPropagatesRequestID(t *testing.T) {
	seenRequestID := ""
	handler := NewHandler(func(ctx context.Context, id string) (string, bool) {
		seenRequestID = RequestID(ctx)
		if id == "api" {
			return "primary", true
		}
		return "", false
	}, func() string { return "generated-1" })

	tests := []struct {
		name          string
		method        string
		path          string
		requestID     string
		wantStatus    int
		wantBody      string
		wantRequestID string
	}{
		{name: "health", method: http.MethodGet, path: "/healthz", wantStatus: http.StatusNoContent, wantRequestID: "generated-1"},
		{name: "target", method: http.MethodGet, path: "/targets/api", requestID: "caller-7", wantStatus: http.StatusOK, wantBody: "primary\n", wantRequestID: "caller-7"},
		{name: "missing", method: http.MethodGet, path: "/targets/missing", wantStatus: http.StatusNotFound, wantBody: "404 page not found\n", wantRequestID: "generated-1"},
		{name: "method", method: http.MethodPost, path: "/targets/api", wantStatus: http.StatusMethodNotAllowed, wantBody: "Method Not Allowed\n", wantRequestID: "generated-1"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			seenRequestID = ""
			request := httptest.NewRequest(testCase.method, testCase.path, nil)
			if testCase.requestID != "" {
				request.Header.Set("X-Request-ID", testCase.requestID)
			}
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
			if got := response.Header.Get("X-Request-ID"); got != testCase.wantRequestID {
				t.Fatalf("X-Request-ID = %q, want %q", got, testCase.wantRequestID)
			}
			if testCase.method == http.MethodGet && testCase.path == "/targets/api" && seenRequestID != testCase.wantRequestID {
				t.Fatalf("lookup request ID = %q, want %q", seenRequestID, testCase.wantRequestID)
			}
		})
	}
}
