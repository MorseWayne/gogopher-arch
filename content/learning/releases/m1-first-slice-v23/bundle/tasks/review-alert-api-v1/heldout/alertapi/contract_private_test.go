package alertapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type creatorFunc func(context.Context, NewRule) (Rule, error)
func (f creatorFunc) Create(ctx context.Context, in NewRule) (Rule, error) { return f(ctx, in) }

func TestVariantAPIContract(t *testing.T) {
	tests := []struct{ name, body string; err error; status int; code string }{
		{"success", `{"destination":"https://hooks.example.com/a","threshold":80}`, nil, 201, ""},
		{"invalid scheme", `{"destination":"file:///tmp/a","threshold":80}`, nil, 400, "invalid_request"},
		{"unknown field", `{"destination":"https://hooks.example.com/a","threshold":80,"token":"x"}`, nil, 400, "invalid_request"},
		{"conflict", `{"destination":"https://hooks.example.com/a","threshold":80}`, ErrRuleExists, 409, "rule_exists"},
		{"internal", `{"destination":"https://hooks.example.com/a","threshold":80}`, errors.New("secret backend"), 500, "internal_error"},
	}
	for _, tt := range tests { t.Run(tt.name, func(t *testing.T) {
		creator := creatorFunc(func(_ context.Context, in NewRule) (Rule, error) { return Rule{ID:"r-1",Destination:in.Destination,Threshold:in.Threshold}, tt.err })
		r := httptest.NewRequest(http.MethodPost, "/rules", strings.NewReader(tt.body)); w := httptest.NewRecorder(); NewHandler(creator).ServeHTTP(w, r)
		if w.Code != tt.status { t.Fatalf("status=%d body=%s", w.Code, w.Body.String()) }
		if !strings.HasPrefix(w.Header().Get("Content-Type"), "application/json") { t.Fatalf("content-type=%q", w.Header().Get("Content-Type")) }
		if tt.code != "" && !strings.Contains(w.Body.String(), `"code":"`+tt.code+`"`) { t.Fatalf("body=%s", w.Body.String()) }
		if strings.Contains(w.Body.String(), "secret backend") { t.Fatal("internal error leaked") }
	}) }
}
