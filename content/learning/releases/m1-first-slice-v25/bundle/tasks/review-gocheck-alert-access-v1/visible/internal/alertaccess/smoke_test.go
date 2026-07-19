package alertaccess

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type smokeStore struct{}

func (smokeStore) FindRule(context.Context, string) (Rule, error) { return Rule{}, ErrNotFound }
func (smokeStore) DeleteRule(context.Context, string) error       { return nil }

type smokeLogger struct{}

func (smokeLogger) Denied(context.Context, string) {}

func TestHandlerRequiresAuthentication(t *testing.T) {
	handler, err := New([]Credential{{Subject: "learner", Token: "learner-secret"}}, smokeStore{}, smokeLogger{})
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodDelete, "/v1/alert-rules/rule-1", nil))
	if recorder.Code != http.StatusUnauthorized || recorder.Header().Get("WWW-Authenticate") != "Bearer" {
		t.Fatalf("status=%d headers=%v body=%q", recorder.Code, recorder.Header(), recorder.Body.String())
	}
}
