package alertobserve

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type privateLogger struct {
	events    []Event
	contextID string
}

func (logger *privateLogger) Log(ctx context.Context, event Event) {
	logger.events = append(logger.events, event)
	logger.contextID = RequestID(ctx)
}

type privatePoint struct {
	method, route, class string
	duration             time.Duration
}
type privateMetrics struct{ points []privatePoint }

func (metrics *privateMetrics) Observe(method, route, class string, duration time.Duration) {
	metrics.points = append(metrics.points, privatePoint{method, route, class, duration})
}

type privateReady struct {
	err        error
	calls      int
	contextErr error
}

func (ready *privateReady) Check(ctx context.Context) error {
	ready.calls++
	ready.contextErr = ctx.Err()
	return ready.err
}

func TestRequestIDCorrelatesResponseContextAndStructuredLog(t *testing.T) {
	service, logger, _ := newPrivateService(t, &privateReady{}, []time.Time{time.Unix(1, 0), time.Unix(1, int64(time.Millisecond))})
	var contextID string
	request := httptest.NewRequest(http.MethodDelete, "/api/v1/alerts/a-9", nil)
	request.Header.Set("X-Request-ID", "alert-trace-8")
	response := httptest.NewRecorder()
	service.Middleware(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		contextID = RequestID(request.Context())
		response.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(response, request)
	if contextID != "alert-trace-8" || response.Header().Get("X-Request-ID") != "alert-trace-8" || logger.contextID != "alert-trace-8" || logger.events[0].Message != "http_request_completed" {
		t.Fatalf("context=%q response=%q logger=%#v", contextID, response.Header().Get("X-Request-ID"), logger)
	}
}

func TestMissingOrUnsafeRequestIDIsGenerated(t *testing.T) {
	for _, incoming := range []string{"", "alert id with spaces", strings.Repeat("x", 65)} {
		service, logger, _ := newPrivateService(t, &privateReady{}, []time.Time{time.Unix(11, 0), time.Unix(11, 1)})
		request := httptest.NewRequest(http.MethodGet, "/api/v1/alerts/a-1", nil)
		request.Header.Set("X-Request-ID", incoming)
		response := httptest.NewRecorder()
		service.Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(response, request)
		if response.Header().Get("X-Request-ID") != "generated-alert-id" || logger.events[0].RequestID != "generated-alert-id" {
			t.Fatalf("incoming=%q response=%q event=%#v", incoming, response.Header().Get("X-Request-ID"), logger.events)
		}
	}
}

func TestMetricsUseOnlyBoundedRouteMethodAndStatusLabels(t *testing.T) {
	service, logger, metrics := newPrivateService(t, &privateReady{}, []time.Time{time.Unix(2, 0), time.Unix(2, int64(2*time.Millisecond))})
	request := httptest.NewRequest("ALERT-CUSTOM-ID", "/api/v1/alerts/customer-991?api_key=secret", nil)
	request.Header.Set("X-Request-ID", "bad request id with spaces")
	response := httptest.NewRecorder()
	service.Middleware(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) { response.WriteHeader(http.StatusConflict) })).ServeHTTP(response, request)
	point := metrics.points[0]
	if point.method != "OTHER" || point.route != "/api/v1/alerts/{id}" || point.class != "4xx" || logger.events[0].RequestID != "generated-alert-id" {
		t.Fatalf("point=%#v event=%#v", point, logger.events[0])
	}
	if strings.Contains(point.route, "customer-991") || strings.Contains(point.route, "secret") {
		t.Fatalf("high-cardinality labels=%#v", point)
	}
}

func TestImplicitStatusBytesAndDurationAreRecorded(t *testing.T) {
	service, logger, metrics := newPrivateService(t, &privateReady{}, []time.Time{time.Unix(3, 0), time.Unix(3, int64(15*time.Millisecond))})
	response := httptest.NewRecorder()
	service.Middleware(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(response, "queued") })).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/alerts", nil))
	event := logger.events[0]
	if event.Status != http.StatusOK || event.Bytes != 6 || event.Duration != 15*time.Millisecond || metrics.points[0].duration != 15*time.Millisecond {
		t.Fatalf("event=%#v metrics=%#v", event, metrics.points)
	}
}

func TestLivenessDoesNotProbeDependencies(t *testing.T) {
	ready := &privateReady{err: errors.New("queue unavailable")}
	service, _, _ := newPrivateService(t, ready, nil)
	response := httptest.NewRecorder()
	service.Liveness(response, httptest.NewRequest(http.MethodGet, "/livez", nil))
	if response.Code != http.StatusOK || response.Body.String() != "live\n" || ready.calls != 0 {
		t.Fatalf("status=%d body=%q calls=%d", response.Code, response.Body.String(), ready.calls)
	}
}

func TestReadinessReflectsDependencyWithoutLeakingError(t *testing.T) {
	ready := &privateReady{err: errors.New("webhook token secret unavailable")}
	service, _, _ := newPrivateService(t, ready, nil)
	response := httptest.NewRecorder()
	service.Readiness(response, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if response.Code != http.StatusServiceUnavailable || response.Body.String() != "not ready\n" || ready.calls != 1 || strings.Contains(response.Body.String(), "secret") {
		t.Fatalf("status=%d body=%q ready=%#v", response.Code, response.Body.String(), ready)
	}
	ready.err = nil
	response = httptest.NewRecorder()
	service.Readiness(response, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if response.Code != http.StatusOK || response.Body.String() != "ready\n" {
		t.Fatalf("status=%d body=%q", response.Code, response.Body.String())
	}
}

func TestConstructorRejectsInvalidDependenciesAndRoute(t *testing.T) {
	logger, metrics, ready := &privateLogger{}, &privateMetrics{}, &privateReady{}
	valid := Options{Route: "/api/v1/alerts/{id}", Now: time.Now, NewRequestID: func() string { return "id" }}
	for _, test := range []struct {
		name    string
		logger  Logger
		metrics Metrics
		ready   Readiness
		options Options
	}{{"logger", nil, metrics, ready, valid}, {"metrics", logger, nil, ready, valid}, {"readiness", logger, metrics, nil, valid}, {"route", logger, metrics, ready, Options{Now: time.Now, NewRequestID: valid.NewRequestID}}, {"query route", logger, metrics, ready, Options{Route: "/alerts?id=raw", Now: time.Now, NewRequestID: valid.NewRequestID}}, {"clock", logger, metrics, ready, Options{Route: valid.Route, NewRequestID: valid.NewRequestID}}, {"generator", logger, metrics, ready, Options{Route: valid.Route, Now: time.Now}}} {
		t.Run(test.name, func(t *testing.T) {
			if service, err := New(test.logger, test.metrics, test.ready, test.options); err == nil || service != nil {
				t.Fatalf("service=%#v error=%v", service, err)
			}
		})
	}
}

func newPrivateService(t *testing.T, ready Readiness, times []time.Time) (*Service, *privateLogger, *privateMetrics) {
	t.Helper()
	logger, metrics := &privateLogger{}, &privateMetrics{}
	now := time.Now
	if len(times) > 0 {
		now = func() time.Time { value := times[0]; times = times[1:]; return value }
	}
	service, err := New(logger, metrics, ready, Options{Route: "/api/v1/alerts/{id}", Now: now, NewRequestID: func() string { return "generated-alert-id" }})
	if err != nil {
		t.Fatal(err)
	}
	return service, logger, metrics
}
