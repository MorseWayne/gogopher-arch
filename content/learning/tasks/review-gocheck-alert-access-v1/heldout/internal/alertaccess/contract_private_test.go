package alertaccess

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type ruleStoreStub struct {
	rules      map[string]Rule
	findErr    error
	deleteErr  error
	finds      int
	deletedIDs []string
}

func (store *ruleStoreStub) FindRule(_ context.Context, id string) (Rule, error) {
	store.finds++
	if store.findErr != nil {
		return Rule{}, store.findErr
	}
	rule, exists := store.rules[id]
	if !exists {
		return Rule{}, ErrNotFound
	}
	return rule, nil
}

func (store *ruleStoreStub) DeleteRule(_ context.Context, id string) error {
	if store.deleteErr != nil {
		return store.deleteErr
	}
	store.deletedIDs = append(store.deletedIDs, id)
	return nil
}

type auditLoggerSpy struct{ reasons []string }

func (logger *auditLoggerSpy) Denied(_ context.Context, reason string) {
	logger.reasons = append(logger.reasons, reason)
}

func TestCredentialConfigurationSafe(t *testing.T) {
	store, logger := &ruleStoreStub{}, &auditLoggerSpy{}
	unsafe := [][]Credential{nil, {{Subject: "", Token: "secret"}}, {{Subject: "alice", Token: ""}}, {{Subject: "alice", Token: "same"}, {Subject: "bob", Token: "same"}}}
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
	store := &ruleStoreStub{rules: map[string]Rule{"rule-1": {ID: "rule-1", OwnerID: "alice"}}}
	logger := &auditLoggerSpy{}
	handler := newAlertHandler(t, store, logger)
	for _, authorization := range []string{"", "Basic alice-secret", "Bearer alice-secret extra", "Bearer attacker-secret"} {
		recorder := requestAlert(handler, authorization, http.MethodDelete, "/v1/alert-rules/rule-1")
		if recorder.Code != http.StatusUnauthorized || recorder.Header().Get("WWW-Authenticate") != "Bearer" {
			t.Fatalf("status=%d headers=%v body=%q", recorder.Code, recorder.Header(), recorder.Body.String())
		}
		assertJSONCode(t, recorder, "unauthorized")
		if strings.Contains(recorder.Body.String(), "secret") {
			t.Fatalf("response leaked secret: %q", recorder.Body.String())
		}
	}
	if store.finds != 0 || len(store.deletedIDs) != 0 {
		t.Fatalf("unauthenticated request touched store: finds=%d deletes=%v", store.finds, store.deletedIDs)
	}
	if len(logger.reasons) != 4 {
		t.Fatalf("audit reasons = %#v", logger.reasons)
	}
	for _, reason := range logger.reasons {
		if reason != "authentication_failed" || strings.Contains(reason, "secret") {
			t.Fatalf("unsafe reason %q", reason)
		}
	}
}

func TestResourceAuthorizationPreventsIDOR(t *testing.T) {
	store := &ruleStoreStub{rules: map[string]Rule{"rule-1": {ID: "rule-1", OwnerID: "alice"}}}
	logger := &auditLoggerSpy{}
	handler := newAlertHandler(t, store, logger)
	owner := requestAlert(handler, "Bearer alice-secret", http.MethodDelete, "/v1/alert-rules/rule-1")
	if owner.Code != http.StatusNoContent || len(store.deletedIDs) != 1 || store.deletedIDs[0] != "rule-1" {
		t.Fatalf("owner response=%d deletes=%v body=%q", owner.Code, store.deletedIDs, owner.Body.String())
	}
	store.deletedIDs = nil
	other := requestAlert(handler, "Bearer bob-secret", http.MethodDelete, "/v1/alert-rules/rule-1")
	missing := requestAlert(handler, "Bearer alice-secret", http.MethodDelete, "/v1/alert-rules/missing")
	if other.Code != http.StatusNotFound || missing.Code != http.StatusNotFound || other.Body.String() != missing.Body.String() {
		t.Fatalf("other=%d %q missing=%d %q", other.Code, other.Body.String(), missing.Code, missing.Body.String())
	}
	if len(store.deletedIDs) != 0 {
		t.Fatalf("unauthorized rule deleted: %v", store.deletedIDs)
	}
	if len(logger.reasons) != 1 || logger.reasons[0] != "resource_not_found" {
		t.Fatalf("audit reasons = %#v", logger.reasons)
	}
}

func TestCanonicalInputAndMethodBoundaries(t *testing.T) {
	store := &ruleStoreStub{rules: map[string]Rule{}}
	handler := newAlertHandler(t, store, &auditLoggerSpy{})
	paths := []string{"/v1/alert-rules/", "/v1/alert-rules/UPPER", "/v1/alert-rules/a.b", "/v1/alert-rules/bad%2Fid", "/v1/alert-rules/a/b", "/v1/alert-rules/" + strings.Repeat("a", 65), "/other/rule-1"}
	for _, path := range paths {
		before := store.finds
		recorder := requestAlert(handler, "Bearer alice-secret", http.MethodDelete, path)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("path=%q status=%d body=%q", path, recorder.Code, recorder.Body.String())
		}
		assertJSONCode(t, recorder, "invalid_resource_id")
		if store.finds != before {
			t.Fatal("invalid ID reached store")
		}
	}
	before := store.finds
	method := requestAlert(handler, "Bearer alice-secret", http.MethodGet, "/v1/alert-rules/rule-1")
	if method.Code != http.StatusMethodNotAllowed || method.Header().Get("Allow") != http.MethodDelete || store.finds != before {
		t.Fatalf("method response=%d headers=%v finds=%d", method.Code, method.Header(), store.finds)
	}
}

func TestSecurityFailureResponsesStable(t *testing.T) {
	tests := []struct {
		name   string
		store  *ruleStoreStub
		status int
		code   string
	}{
		{name: "find failure", store: &ruleStoreStub{findErr: errors.New("database alice-secret detail")}, status: http.StatusInternalServerError, code: "internal_error"},
		{name: "delete disappeared", store: &ruleStoreStub{rules: map[string]Rule{"rule-1": {ID: "rule-1", OwnerID: "alice"}}, deleteErr: ErrNotFound}, status: http.StatusNotFound, code: "not_found"},
		{name: "delete failure", store: &ruleStoreStub{rules: map[string]Rule{"rule-1": {ID: "rule-1", OwnerID: "alice"}}, deleteErr: errors.New("delete alice-secret detail")}, status: http.StatusInternalServerError, code: "internal_error"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := newAlertHandler(t, test.store, &auditLoggerSpy{})
			recorder := requestAlert(handler, "Bearer alice-secret", http.MethodDelete, "/v1/alert-rules/rule-1")
			if recorder.Code != test.status {
				t.Fatalf("status=%d body=%q", recorder.Code, recorder.Body.String())
			}
			assertJSONCode(t, recorder, test.code)
			if strings.Contains(recorder.Body.String(), "alice-secret") || strings.Contains(recorder.Body.String(), "database") || strings.Contains(recorder.Body.String(), "delete") {
				t.Fatalf("response leaked internal detail: %q", recorder.Body.String())
			}
		})
	}
}

func newAlertHandler(t *testing.T, store RuleStore, logger AuditLogger) *Handler {
	t.Helper()
	handler, err := New([]Credential{{Subject: "alice", Token: "alice-secret"}, {Subject: "bob", Token: "bob-secret"}}, store, logger)
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func requestAlert(handler http.Handler, authorization, method, path string) *httptest.ResponseRecorder {
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
