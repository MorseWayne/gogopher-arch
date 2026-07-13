package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	learningsession "github.com/MorseWayne/gogopher-arch/internal/learning/session"
)

func TestSessionHandlerEstablishesSecurelyScopedCookie(t *testing.T) {
	now := time.Date(2026, time.July, 13, 8, 0, 0, 0, time.UTC)
	service := &sessionServiceStub{establishment: learningsession.Establishment{
		Created: true, Token: "raw-session-token",
		Session: learningsession.Session{
			ID: "session-id", LearnerID: "learner-id", CreatedAt: now, ExpiresAt: now.Add(24 * time.Hour),
		},
	}}
	handler, err := NewSessionHandler(service, SessionHandlerOptions{})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, SessionCookiePath+"/session", nil)
	handler.Establish(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("cookies = %#v, want one", cookies)
	}
	cookie := cookies[0]
	if cookie.Name != SessionCookieName || cookie.Value != "raw-session-token" || cookie.Path != SessionCookiePath {
		t.Fatalf("cookie identity/scope = %#v", cookie)
	}
	if !cookie.HttpOnly || cookie.Secure || cookie.SameSite != http.SameSiteLaxMode || cookie.MaxAge != 86400 {
		t.Fatalf("cookie security attributes = %#v", cookie)
	}
	if strings.Contains(response.Body.String(), "raw-session-token") {
		t.Fatal("response body exposed raw session token")
	}
	if cache := response.Header().Get("Cache-Control"); cache != "no-store" {
		t.Fatalf("Cache-Control = %q", cache)
	}
}

func TestSessionHandlerReusesCookieWithoutRotatingIt(t *testing.T) {
	service := &sessionServiceStub{establishment: learningsession.Establishment{
		Session: learningsession.Session{LearnerID: "learner-id", ExpiresAt: time.Now().Add(time.Hour)},
	}}
	handler, err := NewSessionHandler(service, SessionHandlerOptions{})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, SessionCookiePath+"/session", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "existing-token"})
	handler.Establish(response, request)

	if service.establishToken != "existing-token" {
		t.Fatalf("Establish token = %q", service.establishToken)
	}
	if header := response.Header().Get("Set-Cookie"); header != "" {
		t.Fatalf("Set-Cookie = %q, reused session must not rotate", header)
	}
}

func TestSessionMiddlewareRejectsInvalidCookieAndSetsOwnerContext(t *testing.T) {
	service := &sessionServiceStub{authenticateErr: learningsession.ErrUnauthenticated}
	handler, err := NewSessionHandler(service, SessionHandlerOptions{})
	if err != nil {
		t.Fatal(err)
	}
	nextCalled := false
	next := handler.Authenticate(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { nextCalled = true }))
	request := httptest.NewRequest(http.MethodGet, SessionCookiePath+"/attempts", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "forged-token"})
	response := httptest.NewRecorder()
	next.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized || nextCalled {
		t.Fatalf("forged response = %d called=%v", response.Code, nextCalled)
	}
	if strings.Contains(response.Body.String(), "forged-token") {
		t.Fatal("unauthorized response exposed raw token")
	}

	service.authenticateErr = nil
	service.authenticated = learningsession.Session{ID: "session-id", LearnerID: "owner-id"}
	next = handler.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		active, ok := SessionFromContext(request.Context())
		if !ok || active.LearnerID != "owner-id" {
			t.Fatalf("session context = %#v, %v", active, ok)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	request = httptest.NewRequest(http.MethodGet, SessionCookiePath+"/attempts", nil)
	request.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "valid-token"})
	response = httptest.NewRecorder()
	next.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent || service.authenticateToken != "valid-token" {
		t.Fatalf("valid response = %d token=%q", response.Code, service.authenticateToken)
	}
}

type sessionServiceStub struct {
	establishment     learningsession.Establishment
	establishErr      error
	establishToken    string
	authenticated     learningsession.Session
	authenticateErr   error
	authenticateToken string
}

func (s *sessionServiceStub) Establish(_ context.Context, token string) (learningsession.Establishment, error) {
	s.establishToken = token
	return s.establishment, s.establishErr
}

func (s *sessionServiceStub) Authenticate(_ context.Context, token string) (learningsession.Session, error) {
	s.authenticateToken = token
	return s.authenticated, s.authenticateErr
}
