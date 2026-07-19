package checkapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type creatorFunc func(context.Context, NewCheck) (Check, error)
func (f creatorFunc) Create(ctx context.Context, in NewCheck) (Check, error) { return f(ctx, in) }

func TestJSONContractAndValidation(t *testing.T) {
	creator := creatorFunc(func(_ context.Context, in NewCheck) (Check, error) { return Check{ID: "c-1", Target: in.Target, TimeoutMS: in.TimeoutMS}, nil })
	tests := []struct{ name, body string; want int }{
		{"success", `{"target":"https://example.com","timeout_ms":1200}`, 201},
		{"unknown", `{"target":"https://example.com","timeout_ms":1200,"secret":true}`, 400},
		{"trailing", `{"target":"https://example.com","timeout_ms":1200}{}`, 400},
		{"empty target", `{"target":" ","timeout_ms":1200}`, 400},
		{"timeout low", `{"target":"https://example.com","timeout_ms":0}`, 400},
		{"timeout high", `{"target":"https://example.com","timeout_ms":60001}`, 400},
	}
	for _, tt := range tests { t.Run(tt.name, func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/checks", strings.NewReader(tt.body)); r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder(); NewHandler(creator).ServeHTTP(w, r)
		if w.Code != tt.want { t.Fatalf("status=%d body=%s", w.Code, w.Body.String()) }
		if !strings.HasPrefix(w.Header().Get("Content-Type"), "application/json") { t.Fatalf("content-type=%q", w.Header().Get("Content-Type")) }
	}) }
}

func TestDomainErrorsMapToStableProtocol(t *testing.T) {
	tests := []struct{ name string; err error; status int; code string }{
		{"conflict", ErrCheckExists, 409, "check_exists"},
		{"internal", errors.New("password=secret"), 500, "internal_error"},
	}
	for _, tt := range tests { t.Run(tt.name, func(t *testing.T) {
		creator := creatorFunc(func(context.Context, NewCheck) (Check, error) { return Check{}, tt.err })
		r := httptest.NewRequest(http.MethodPost, "/checks", strings.NewReader(`{"target":"https://example.com","timeout_ms":1000}`))
		w := httptest.NewRecorder(); NewHandler(creator).ServeHTTP(w, r)
		if w.Code != tt.status || !strings.Contains(w.Body.String(), `"code":"`+tt.code+`"`) { t.Fatalf("status=%d body=%s", w.Code, w.Body.String()) }
		if strings.Contains(w.Body.String(), "password=secret") { t.Fatal("internal error leaked") }
	}) }
}
