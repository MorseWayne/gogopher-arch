package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	learningsession "github.com/MorseWayne/gogopher-arch/internal/learning/session"
)

func TestAttemptHandlerCreatesForAuthenticatedOwner(t *testing.T) {
	service := &attemptServiceStub{created: attempt.Attempt{ID: "attempt-id", LearnerID: "owner", Status: "active", Workspace: map[string]string{}, CapabilityRefs: nil}}
	handler, _ := NewAttemptHandler(service)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/learning/attempts", strings.NewReader(`{"activity_id":"assessment-check-config","activity_version":1}`))
	request = request.WithContext(context.WithValue(request.Context(), sessionContextKey{}, learningsession.Session{LearnerID: "owner"}))
	response := httptest.NewRecorder()
	handler.Create(response, request)
	if response.Code != http.StatusCreated || service.createInput.LearnerID != "owner" {
		t.Fatalf("response=%d input=%#v", response.Code, service.createInput)
	}
	if !strings.Contains(response.Body.String(), `"executions":[]`) || !strings.Contains(response.Body.String(), `"evidence":[]`) {
		t.Fatalf("body = %s", response.Body.String())
	}
}

func TestAttemptHandlerMapsOwnershipAndRevisionConflict(t *testing.T) {
	service := &attemptServiceStub{getErr: attempt.ErrNotFound, saveErr: &attempt.RevisionConflict{Revision: 3, Hash: "current-hash"}}
	handler, _ := NewAttemptHandler(service)
	ctx := context.WithValue(context.Background(), sessionContextKey{}, learningsession.Session{LearnerID: "owner"})
	request := httptest.NewRequest(http.MethodGet, "/attempts/id", nil).WithContext(ctx)
	response := httptest.NewRecorder()
	handler.Get(response, request, "id")
	if response.Code != http.StatusNotFound {
		t.Fatalf("Get status = %d", response.Code)
	}
	request = httptest.NewRequest(http.MethodPut, "/attempts/id/workspace", strings.NewReader(`{"base_revision":0,"files":{"main.go":"package main"}}`)).WithContext(ctx)
	response = httptest.NewRecorder()
	handler.SaveWorkspace(response, request, "id")
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), `"current_revision":3`) || !strings.Contains(response.Body.String(), `"current_hash":"current-hash"`) {
		t.Fatalf("conflict = %d %s", response.Code, response.Body.String())
	}
}

type attemptServiceStub struct {
	created     attempt.Attempt
	createInput attempt.CreateInput
	createErr   error
	got         attempt.Attempt
	getErr      error
	saved       attempt.Attempt
	saveErr     error
}

func (s *attemptServiceStub) Create(_ context.Context, input attempt.CreateInput) (attempt.Attempt, error) {
	s.createInput = input
	return s.created, s.createErr
}
func (s *attemptServiceStub) Get(context.Context, string, string) (attempt.Attempt, error) {
	return s.got, s.getErr
}
func (s *attemptServiceStub) Save(context.Context, attempt.SaveInput) (attempt.Attempt, error) {
	return s.saved, s.saveErr
}
