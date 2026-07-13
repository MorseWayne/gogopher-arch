package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	learningsession "github.com/MorseWayne/gogopher-arch/internal/learning/session"
)

const (
	SessionCookieName = "gogopher_learning_session"
	SessionCookiePath = "/api/v1/learning"
	APIVersion        = "v1"
)

type SessionService interface {
	Establish(context.Context, string) (learningsession.Establishment, error)
	Authenticate(context.Context, string) (learningsession.Session, error)
}

type SessionHandlerOptions struct {
	SecureCookie bool
}

type SessionHandler struct {
	service      SessionService
	secureCookie bool
}

func NewSessionHandler(service SessionService, options SessionHandlerOptions) (*SessionHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("session service is required")
	}
	return &SessionHandler{service: service, secureCookie: options.SecureCookie}, nil
}

func (h *SessionHandler) Establish(w http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
		return
	}
	token := ""
	if cookie, err := request.Cookie(SessionCookieName); err == nil {
		token = cookie.Value
	}
	established, err := h.service.Establish(request.Context(), token)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "session_unavailable", "learning session is unavailable")
		return
	}
	if established.Created {
		maxAge := int(established.Session.ExpiresAt.Sub(established.Session.CreatedAt).Seconds())
		http.SetCookie(w, &http.Cookie{
			Name: SessionCookieName, Value: established.Token, Path: SessionCookiePath,
			Expires: established.Session.ExpiresAt, MaxAge: maxAge,
			HttpOnly: true, Secure: h.secureCookie, SameSite: http.SameSiteLaxMode,
		})
	}
	w.Header().Set("Cache-Control", "no-store")
	writeJSON(w, http.StatusOK, struct {
		APIVersion string `json:"api_version"`
		Learner    struct {
			ID string `json:"id"`
		} `json:"learner"`
		Session struct {
			ExpiresAt time.Time `json:"expires_at"`
		} `json:"session"`
	}{
		APIVersion: APIVersion,
		Learner: struct {
			ID string `json:"id"`
		}{ID: established.Session.LearnerID},
		Session: struct {
			ExpiresAt time.Time `json:"expires_at"`
		}{ExpiresAt: established.Session.ExpiresAt},
	})
}

func (h *SessionHandler) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		cookie, err := request.Cookie(SessionCookieName)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthenticated", "learning session is required")
			return
		}
		active, err := h.service.Authenticate(request.Context(), cookie.Value)
		if err != nil {
			if errors.Is(err, learningsession.ErrUnauthenticated) {
				writeError(w, http.StatusUnauthorized, "unauthenticated", "learning session is required")
				return
			}
			writeError(w, http.StatusInternalServerError, "session_unavailable", "learning session is unavailable")
			return
		}
		next.ServeHTTP(w, request.WithContext(context.WithValue(request.Context(), sessionContextKey{}, active)))
	})
}

func SessionFromContext(ctx context.Context) (learningsession.Session, bool) {
	active, ok := ctx.Value(sessionContextKey{}).(learningsession.Session)
	return active, ok
}

type sessionContextKey struct{}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}{Error: struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}{Code: code, Message: message}})
}
