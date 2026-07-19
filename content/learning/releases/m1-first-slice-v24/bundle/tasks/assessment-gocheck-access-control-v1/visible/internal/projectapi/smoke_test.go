package projectapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

type smokeStore struct{}

func (smokeStore) FindProject(context.Context, string) (Project, error) {
	return Project{}, ErrNotFound
}

type smokeLogger struct{}

func (smokeLogger) Denied(context.Context, string) {}

func TestHandlerRequiresAuthentication(t *testing.T) {
	handler, err := New([]Credential{{Subject: "learner", Token: "learner-secret"}}, smokeStore{}, smokeLogger{})
	if err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/v1/projects/project-1", nil))
	if recorder.Code != http.StatusUnauthorized || recorder.Header().Get("WWW-Authenticate") != "Bearer" {
		t.Fatalf("status=%d headers=%v body=%q", recorder.Code, recorder.Header(), recorder.Body.String())
	}
}
