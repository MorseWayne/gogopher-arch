package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type eventSink struct{ event Event }

func (sink *eventSink) Log(_ context.Context, event Event) { sink.event = event }

type metricSink struct {
	method, route, class string
	duration             time.Duration
}

func (sink *metricSink) Observe(method, route, class string, duration time.Duration) {
	sink.method, sink.route, sink.class, sink.duration = method, route, class, duration
}

func TestMiddlewareCorrelatesStructuredEventAndMetric(t *testing.T) {
	logger, metrics := &eventSink{}, &metricSink{}
	times := []time.Time{time.Unix(100, 0), time.Unix(100, int64(25*time.Millisecond))}
	now := func() time.Time { value := times[0]; times = times[1:]; return value }
	wrap, err := New("/api/v1/checks/{id}", logger, metrics, now, func() string { return "req-generated" })
	if err != nil {
		t.Fatal(err)
	}
	var contextID string
	handler := wrap(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		contextID = RequestID(request.Context())
		response.WriteHeader(http.StatusCreated)
	}))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/checks/check-42?token=secret", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Header().Get("X-Request-ID") != "req-generated" || contextID != "req-generated" {
		t.Fatalf("response request id=%q context id=%q", response.Header().Get("X-Request-ID"), contextID)
	}
	if logger.event.RequestID != "req-generated" || logger.event.Route != "/api/v1/checks/{id}" || logger.event.Status != http.StatusCreated || logger.event.StatusClass != "2xx" || logger.event.Duration != 25*time.Millisecond {
		t.Fatalf("event=%#v", logger.event)
	}
	if metrics.method != http.MethodPost || metrics.route != "/api/v1/checks/{id}" || metrics.class != "2xx" || metrics.duration != 25*time.Millisecond {
		t.Fatalf("metric=%#v", metrics)
	}
}
