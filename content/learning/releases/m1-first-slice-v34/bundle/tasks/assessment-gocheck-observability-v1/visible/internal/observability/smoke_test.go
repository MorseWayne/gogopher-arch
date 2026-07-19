package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type smokeLogger struct{ event Event }

func (logger *smokeLogger) Log(_ context.Context, event Event) { logger.event = event }

type smokeMetrics struct{ route string }

func (metrics *smokeMetrics) Observe(_, route, _ string, _ time.Duration) { metrics.route = route }

type smokeReadiness struct{}

func (smokeReadiness) Check(context.Context) error { return nil }

func TestServiceEmitsTemplateTelemetry(t *testing.T) {
	logger, metrics := &smokeLogger{}, &smokeMetrics{}
	service, err := New(logger, metrics, smokeReadiness{}, Options{Route: "/api/v1/checks/{id}", Now: time.Now, NewRequestID: func() string { return "request-1" }})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	service.Middleware(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if RequestID(request.Context()) == "" {
			t.Fatal("missing request ID")
		}
		response.WriteHeader(http.StatusAccepted)
	})).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/checks/raw-id?secret=value", nil))
	if logger.event.Route != "/api/v1/checks/{id}" || logger.event.Status != http.StatusAccepted || metrics.route != "/api/v1/checks/{id}" {
		t.Fatalf("event=%#v metric route=%q", logger.event, metrics.route)
	}
}
