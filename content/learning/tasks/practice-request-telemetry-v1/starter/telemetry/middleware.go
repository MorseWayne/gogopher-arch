package telemetry

import (
	"context"
	"errors"
	"net/http"
	"time"
)

type Event struct {
	RequestID   string
	Method      string
	Route       string
	Status      int
	StatusClass string
	Duration    time.Duration
}

type Logger interface {
	Log(context.Context, Event)
}

type Metrics interface {
	Observe(method, route, statusClass string, duration time.Duration)
}

type contextKey struct{}

func RequestID(ctx context.Context) string {
	value, _ := ctx.Value(contextKey{}).(string)
	return value
}

func New(route string, logger Logger, metrics Metrics, now func() time.Time, newRequestID func() string) (func(http.Handler) http.Handler, error) {
	if route == "" || logger == nil || metrics == nil || now == nil || newRequestID == nil {
		return nil, errors.New("invalid telemetry dependencies")
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			// TODO: correlate the request, capture status/duration, then emit one event and one metric.
			next.ServeHTTP(response, request)
		})
	}, nil
}

func statusClass(status int) string {
	// TODO: return a bounded label such as 2xx or 5xx.
	return ""
}
