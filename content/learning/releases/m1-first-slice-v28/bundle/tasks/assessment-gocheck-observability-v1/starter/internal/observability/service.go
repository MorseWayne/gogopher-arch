package observability

import (
	"context"
	"errors"
	"net/http"
	"time"
)

type Event struct {
	Message     string
	RequestID   string
	Method      string
	Route       string
	Status      int
	StatusClass string
	Bytes       int
	Duration    time.Duration
}

type Logger interface {
	Log(context.Context, Event)
}

type Metrics interface {
	Observe(method, route, statusClass string, duration time.Duration)
}

type Readiness interface {
	Check(context.Context) error
}

type Options struct {
	Route        string
	Now          func() time.Time
	NewRequestID func() string
}

type Service struct {
	logger    Logger
	metrics   Metrics
	readiness Readiness
	options   Options
}

type requestIDKey struct{}

func RequestID(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey{}).(string)
	return value
}

func New(logger Logger, metrics Metrics, readiness Readiness, options Options) (*Service, error) {
	if logger == nil || metrics == nil || readiness == nil || options.Route == "" || options.Now == nil || options.NewRequestID == nil {
		return nil, errors.New("invalid observability dependencies")
	}
	return &Service{logger: logger, metrics: metrics, readiness: readiness, options: options}, nil
}

func (service *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		// TODO: correlate request, capture response metadata and emit bounded telemetry.
		next.ServeHTTP(response, request)
	})
}

func (service *Service) Liveness(response http.ResponseWriter, _ *http.Request) {
	// TODO: process-only probe.
	http.Error(response, "not implemented", http.StatusNotImplemented)
}

func (service *Service) Readiness(response http.ResponseWriter, request *http.Request) {
	// TODO: dependency-aware probe with a stable public failure.
	http.Error(response, "not implemented", http.StatusNotImplemented)
}

func normalizeMethod(method string) string { return method }
func statusClass(status int) string        { return "" }
