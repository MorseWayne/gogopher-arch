package checkapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type smokeCreator struct{}
func (smokeCreator) Create(_ context.Context, in NewCheck) (Check, error) { return Check{ID: "check-1", Target: in.Target, TimeoutMS: in.TimeoutMS}, nil }

func TestHandlerExposesChecksRoute(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/checks", strings.NewReader(`{"target":"https://example.com","timeout_ms":1000}`))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	NewHandler(smokeCreator{}).ServeHTTP(w, r)
	if w.Code != http.StatusCreated { t.Fatalf("status=%d body=%s", w.Code, w.Body.String()) }
}
