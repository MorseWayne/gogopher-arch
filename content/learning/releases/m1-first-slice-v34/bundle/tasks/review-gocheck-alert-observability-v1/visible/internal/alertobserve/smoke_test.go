package alertobserve

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type visibleLogger struct{ event Event }

func (logger *visibleLogger) Log(_ context.Context, event Event) { logger.event = event }

type visibleMetrics struct{ route string }

func (metrics *visibleMetrics) Observe(_, route, _ string, _ time.Duration) { metrics.route = route }

type visibleReady struct{}

func (visibleReady) Check(context.Context) error { return nil }

func TestAlertRouteUsesTemplateTelemetry(t *testing.T) {
	logger, metrics := &visibleLogger{}, &visibleMetrics{}
	service, err := New(logger, metrics, visibleReady{}, Options{Route: "/api/v1/alerts/{id}", Now: time.Now, NewRequestID: func() string { return "alert-request-1" }})
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	service.Middleware(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if RequestID(request.Context()) == "" {
			t.Fatal("missing request ID")
		}
		response.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(response, httptest.NewRequest(http.MethodDelete, "/api/v1/alerts/raw-alert-id?key=secret", nil))
	if logger.event.Route != "/api/v1/alerts/{id}" || logger.event.Status != http.StatusNoContent || metrics.route != "/api/v1/alerts/{id}" {
		t.Fatalf("event=%#v route=%q", logger.event, metrics.route)
	}
}
