package observability

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type capturedLogger struct {
	events     []Event
	contextIDs []string
}

func (logger *capturedLogger) Log(ctx context.Context, event Event) {
	logger.events = append(logger.events, event)
	logger.contextIDs = append(logger.contextIDs, RequestID(ctx))
}

type metricPoint struct {
	method, route, class string
	duration             time.Duration
}

type capturedMetrics struct{ points []metricPoint }

func (metrics *capturedMetrics) Observe(method, route, class string, duration time.Duration) {
	metrics.points = append(metrics.points, metricPoint{method: method, route: route, class: class, duration: duration})
}

type readinessProbe struct {
	err    error
	calls  int
	marker any
}

func (probe *readinessProbe) Check(ctx context.Context) error {
	probe.calls++
	probe.marker = ctx.Value(probeContextKey{})
	return probe.err
}

type probeContextKey struct{}

func TestRequestIDCorrelatesResponseContextAndStructuredLog(t *testing.T) {
	service, logger, _ := testService(t, &readinessProbe{}, []time.Time{time.Unix(10, 0), time.Unix(10, int64(12*time.Millisecond))})
	var handlerID string
	handler := service.Middleware(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		handlerID = RequestID(request.Context())
		response.WriteHeader(http.StatusCreated)
	}))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/checks/check-9", nil)
	request.Header.Set("X-Request-ID", "trace-42")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if handlerID != "trace-42" || response.Header().Get("X-Request-ID") != "trace-42" {
		t.Fatalf("handler id=%q response id=%q", handlerID, response.Header().Get("X-Request-ID"))
	}
	if len(logger.events) != 1 || logger.events[0].RequestID != "trace-42" || logger.events[0].Message != "http_request_completed" || logger.contextIDs[0] != "trace-42" {
		t.Fatalf("events=%#v context IDs=%v", logger.events, logger.contextIDs)
	}
}

func TestMissingOrUnsafeRequestIDIsGenerated(t *testing.T) {
	for _, incoming := range []string{"", "contains spaces and ?query", strings.Repeat("x", 65)} {
		t.Run(fmt.Sprintf("length-%d", len(incoming)), func(t *testing.T) {
			service, logger, _ := testService(t, &readinessProbe{}, []time.Time{time.Unix(20, 0), time.Unix(20, 1)})
			request := httptest.NewRequest(http.MethodGet, "/api/v1/checks/check-1", nil)
			request.Header.Set("X-Request-ID", incoming)
			response := httptest.NewRecorder()
			service.Middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})).ServeHTTP(response, request)
			if response.Header().Get("X-Request-ID") != "generated-7" || logger.events[0].RequestID != "generated-7" {
				t.Fatalf("response=%q event=%#v", response.Header().Get("X-Request-ID"), logger.events)
			}
		})
	}
}

func TestMetricsUseOnlyBoundedRouteMethodAndStatusLabels(t *testing.T) {
	service, _, metrics := testService(t, &readinessProbe{}, []time.Time{time.Unix(30, 0), time.Unix(30, int64(time.Millisecond))})
	request := httptest.NewRequest("BREW-CHECK-123", "/api/v1/checks/raw-project-id?token=super-secret", nil)
	request.Header.Set("X-Request-ID", "request-high-cardinality-99")
	response := httptest.NewRecorder()
	service.Middleware(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusBadGateway)
	})).ServeHTTP(response, request)
	if len(metrics.points) != 1 {
		t.Fatalf("points=%#v", metrics.points)
	}
	point := metrics.points[0]
	if point.method != "OTHER" || point.route != "/api/v1/checks/{id}" || point.class != "5xx" {
		t.Fatalf("point=%#v", point)
	}
	joined := point.method + point.route + point.class
	for _, forbidden := range []string{"raw-project-id", "super-secret", "high-cardinality"} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("metric labels leak %q: %#v", forbidden, point)
		}
	}
}

func TestImplicitStatusBytesAndDurationAreRecorded(t *testing.T) {
	service, logger, metrics := testService(t, &readinessProbe{}, []time.Time{time.Unix(40, 0), time.Unix(40, int64(40*time.Millisecond))})
	response := httptest.NewRecorder()
	service.Middleware(http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(response, "hello")
	})).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/checks/check-2", nil))
	event := logger.events[0]
	if event.Status != http.StatusOK || event.StatusClass != "2xx" || event.Bytes != 5 || event.Duration != 40*time.Millisecond {
		t.Fatalf("event=%#v", event)
	}
	if len(metrics.points) != 1 || metrics.points[0].duration != 40*time.Millisecond {
		t.Fatalf("metrics=%#v", metrics.points)
	}
}

func TestLivenessDoesNotProbeDependencies(t *testing.T) {
	probe := &readinessProbe{err: errors.New("database unavailable")}
	service, _, _ := testService(t, probe, nil)
	response := httptest.NewRecorder()
	service.Liveness(response, httptest.NewRequest(http.MethodGet, "/livez", nil))
	if response.Code != http.StatusOK || response.Body.String() != "live\n" || probe.calls != 0 {
		t.Fatalf("status=%d body=%q calls=%d", response.Code, response.Body.String(), probe.calls)
	}
}

func TestReadinessReflectsDependencyWithoutLeakingError(t *testing.T) {
	for _, test := range []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{{"ready", nil, http.StatusOK, "ready\n"}, {"not ready", errors.New("postgres password=secret refused"), http.StatusServiceUnavailable, "not ready\n"}} {
		t.Run(test.name, func(t *testing.T) {
			probe := &readinessProbe{err: test.err}
			service, _, _ := testService(t, probe, nil)
			request := httptest.NewRequest(http.MethodGet, "/readyz", nil).WithContext(context.WithValue(context.Background(), probeContextKey{}, "marker"))
			response := httptest.NewRecorder()
			service.Readiness(response, request)
			if response.Code != test.wantStatus || response.Body.String() != test.wantBody || probe.calls != 1 || probe.marker != "marker" {
				t.Fatalf("status=%d body=%q probe=%#v", response.Code, response.Body.String(), probe)
			}
			if strings.Contains(response.Body.String(), "password") || strings.Contains(response.Body.String(), "secret") {
				t.Fatalf("readiness leaked dependency error: %q", response.Body.String())
			}
		})
	}
}

func TestConstructorRejectsInvalidDependenciesAndRoute(t *testing.T) {
	logger, metrics, readiness := &capturedLogger{}, &capturedMetrics{}, &readinessProbe{}
	valid := Options{Route: "/api/v1/checks/{id}", Now: time.Now, NewRequestID: func() string { return "id" }}
	tests := []struct {
		name      string
		logger    Logger
		metrics   Metrics
		readiness Readiness
		options   Options
	}{{"logger", nil, metrics, readiness, valid}, {"metrics", logger, nil, readiness, valid}, {"readiness", logger, metrics, nil, valid}, {"route", logger, metrics, readiness, Options{Now: time.Now, NewRequestID: valid.NewRequestID}}, {"raw route", logger, metrics, readiness, Options{Route: "/checks?id=raw", Now: time.Now, NewRequestID: valid.NewRequestID}}, {"clock", logger, metrics, readiness, Options{Route: valid.Route, NewRequestID: valid.NewRequestID}}, {"generator", logger, metrics, readiness, Options{Route: valid.Route, Now: time.Now}}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if service, err := New(test.logger, test.metrics, test.readiness, test.options); err == nil || service != nil {
				t.Fatalf("service=%#v error=%v", service, err)
			}
		})
	}
}

func testService(t *testing.T, readiness Readiness, times []time.Time) (*Service, *capturedLogger, *capturedMetrics) {
	t.Helper()
	logger, metrics := &capturedLogger{}, &capturedMetrics{}
	now := time.Now
	if len(times) > 0 {
		now = func() time.Time {
			value := times[0]
			times = times[1:]
			return value
		}
	}
	service, err := New(logger, metrics, readiness, Options{Route: "/api/v1/checks/{id}", Now: now, NewRequestID: func() string { return "generated-7" }})
	if err != nil {
		t.Fatal(err)
	}
	return service, logger, metrics
}
