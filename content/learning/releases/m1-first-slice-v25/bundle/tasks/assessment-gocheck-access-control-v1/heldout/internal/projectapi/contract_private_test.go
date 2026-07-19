package projectapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type projectStoreStub struct {
	projects map[string]Project
	err      error
	finds    int
}

func (store *projectStoreStub) FindProject(_ context.Context, id string) (Project, error) {
	store.finds++
	if store.err != nil {
		return Project{}, store.err
	}
	project, exists := store.projects[id]
	if !exists {
		return Project{}, ErrNotFound
	}
	return project, nil
}

type auditLoggerSpy struct{ reasons []string }

func (logger *auditLoggerSpy) Denied(_ context.Context, reason string) {
	logger.reasons = append(logger.reasons, reason)
}

func TestCredentialConfigurationSafe(t *testing.T) {
	store, logger := &projectStoreStub{}, &auditLoggerSpy{}
	unsafe := [][]Credential{
		nil,
		{{Subject: "", Token: "secret"}},
		{{Subject: "alice", Token: ""}},
		{{Subject: "alice", Token: "same"}, {Subject: "bob", Token: "same"}},
	}
	for _, credentials := range unsafe {
		if _, err := New(credentials, store, logger); err == nil {
			t.Fatalf("New(%#v) succeeded", credentials)
		}
	}
	if _, err := New([]Credential{{Subject: "alice", Token: "secret"}}, nil, logger); err == nil {
		t.Fatal("New accepted nil store")
	}
	if _, err := New([]Credential{{Subject: "alice", Token: "secret"}}, store, nil); err == nil {
		t.Fatal("New accepted nil logger")
	}
	if _, err := New([]Credential{{Subject: "alice", Token: "secret"}, {Subject: "alice", Token: "rotated"}}, store, logger); err != nil {
		t.Fatalf("New rejected key rotation: %v", err)
	}
}

func TestAuthenticationContractAndSecretSafety(t *testing.T) {
	store := &projectStoreStub{projects: map[string]Project{"project-1": {ID: "project-1", OwnerID: "alice", Name: "private"}}}
	logger := &auditLoggerSpy{}
	handler := newProjectHandler(t, store, logger)
	tests := []struct{ name, authorization string }{
		{name: "missing"},
		{name: "wrong scheme", authorization: "Basic alice-secret"},
		{name: "extra field", authorization: "Bearer alice-secret extra"},
		{name: "unknown", authorization: "Bearer attacker-secret"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := requestProject(handler, test.authorization, http.MethodGet, "/v1/projects/project-1")
			if recorder.Code != http.StatusUnauthorized || recorder.Header().Get("WWW-Authenticate") != "Bearer" {
				t.Fatalf("status=%d headers=%v body=%q", recorder.Code, recorder.Header(), recorder.Body.String())
			}
			assertJSONCode(t, recorder, "unauthorized")
			for _, secret := range []string{"alice-secret", "attacker-secret", "private"} {
				if strings.Contains(recorder.Body.String(), secret) {
					t.Fatalf("response leaked %q: %q", secret, recorder.Body.String())
				}
			}
		})
	}
	if store.finds != 0 {
		t.Fatalf("unauthenticated requests reached store %d times", store.finds)
	}
	if len(logger.reasons) != len(tests) {
		t.Fatalf("audit reasons = %#v", logger.reasons)
	}
	for _, reason := range logger.reasons {
		if reason != "authentication_failed" || strings.Contains(reason, "secret") {
			t.Fatalf("unsafe audit reason %q", reason)
		}
	}
}

func TestResourceAuthorizationPreventsIDOR(t *testing.T) {
	store := &projectStoreStub{projects: map[string]Project{"project-1": {ID: "project-1", OwnerID: "alice", Name: "Alice project"}}}
	logger := &auditLoggerSpy{}
	handler := newProjectHandler(t, store, logger)
	owner := requestProject(handler, "Bearer alice-secret", http.MethodGet, "/v1/projects/project-1")
	if owner.Code != http.StatusOK || !strings.Contains(owner.Body.String(), "Alice project") {
		t.Fatalf("owner response = %d %q", owner.Code, owner.Body.String())
	}
	other := requestProject(handler, "Bearer bob-secret", http.MethodGet, "/v1/projects/project-1")
	missing := requestProject(handler, "Bearer alice-secret", http.MethodGet, "/v1/projects/missing")
	if other.Code != http.StatusNotFound || missing.Code != http.StatusNotFound || other.Body.String() != missing.Body.String() {
		t.Fatalf("other=%d %q missing=%d %q", other.Code, other.Body.String(), missing.Code, missing.Body.String())
	}
	if strings.Contains(other.Body.String(), "Alice") || strings.Contains(other.Body.String(), "alice") {
		t.Fatalf("authorization response leaked project: %q", other.Body.String())
	}
	if len(logger.reasons) != 1 || logger.reasons[0] != "resource_not_found" {
		t.Fatalf("audit reasons = %#v", logger.reasons)
	}
}

func TestCanonicalInputAndMethodBoundaries(t *testing.T) {
	store := &projectStoreStub{projects: map[string]Project{}}
	handler := newProjectHandler(t, store, &auditLoggerSpy{})
	paths := []string{"/v1/projects/", "/v1/projects/UPPER", "/v1/projects/a.b", "/v1/projects/bad%2Fid", "/v1/projects/a/b", "/v1/projects/" + strings.Repeat("a", 65), "/other/project-1"}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			before := store.finds
			recorder := requestProject(handler, "Bearer alice-secret", http.MethodGet, path)
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("status=%d body=%q", recorder.Code, recorder.Body.String())
			}
			assertJSONCode(t, recorder, "invalid_resource_id")
			if store.finds != before {
				t.Fatal("invalid ID reached store")
			}
		})
	}
	before := store.finds
	method := requestProject(handler, "Bearer alice-secret", http.MethodPost, "/v1/projects/project-1")
	if method.Code != http.StatusMethodNotAllowed || method.Header().Get("Allow") != http.MethodGet || store.finds != before {
		t.Fatalf("method response=%d headers=%v finds=%d", method.Code, method.Header(), store.finds)
	}
}

func TestSecurityFailureResponsesStable(t *testing.T) {
	store := &projectStoreStub{err: errors.New("database alice-secret internal detail")}
	handler := newProjectHandler(t, store, &auditLoggerSpy{})
	recorder := requestProject(handler, "Bearer alice-secret", http.MethodGet, "/v1/projects/project-1")
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%q", recorder.Code, recorder.Body.String())
	}
	assertJSONCode(t, recorder, "internal_error")
	if strings.Contains(recorder.Body.String(), "alice-secret") || strings.Contains(recorder.Body.String(), "database") {
		t.Fatalf("response leaked internal detail: %q", recorder.Body.String())
	}
	if recorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("content type = %q", recorder.Header().Get("Content-Type"))
	}
}

func newProjectHandler(t *testing.T, store ProjectStore, logger AuditLogger) *Handler {
	t.Helper()
	handler, err := New([]Credential{{Subject: "alice", Token: "alice-secret"}, {Subject: "bob", Token: "bob-secret"}}, store, logger)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func requestProject(handler http.Handler, authorization, method, path string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	if authorization != "" {
		request.Header.Set("Authorization", authorization)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func assertJSONCode(t *testing.T, recorder *httptest.ResponseRecorder, expected string) {
	t.Helper()
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil || envelope.Error.Code != expected {
		t.Fatalf("body=%q error=%v code=%q", recorder.Body.String(), err, envelope.Error.Code)
	}
}
