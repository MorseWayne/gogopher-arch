package targetapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerCreatesTargetAndRejectsInvalidInput(t *testing.T) {
	tests := []struct{ name, body string; want int }{
		{"valid", `{"name":"homepage","url":"https://example.com","interval_seconds":30}`, http.StatusCreated},
		{"unknown field", `{"name":"homepage","url":"https://example.com","interval_seconds":30,"debug":true}`, http.StatusBadRequest},
		{"invalid URL", `{"name":"homepage","url":"file:///tmp/x","interval_seconds":30}`, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/targets", strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			Handler().ServeHTTP(w, r)
			if w.Code != tt.want { t.Fatalf("status=%d body=%s", w.Code, w.Body.String()) }
			if got := w.Header().Get("Content-Type"); !strings.HasPrefix(got, "application/json") { t.Fatalf("content-type=%q", got) }
			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil { t.Fatalf("invalid JSON: %v", err) }
		})
	}
}
