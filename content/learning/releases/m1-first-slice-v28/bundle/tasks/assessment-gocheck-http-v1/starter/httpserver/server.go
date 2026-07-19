package httpserver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

type TargetLookup func(context.Context, string) (string, bool)

type Dependencies struct {
	LookupTarget  TargetLookup
	NextRequestID func() string
}

type Timeouts struct {
	ReadHeader time.Duration
	Read       time.Duration
	Write      time.Duration
	Idle       time.Duration
}

func NewHandler(dependencies Dependencies) (http.Handler, error) {
	return nil, errors.New("TODO: build routes and request ID middleware")
}

func RequestID(ctx context.Context) string {
	return ""
}

func NewServer(handler http.Handler, timeouts Timeouts) (*http.Server, error) {
	return nil, errors.New("TODO: validate and configure http.Server")
}

func Serve(ctx context.Context, server *http.Server, listener net.Listener, shutdownTimeout time.Duration) error {
	return errors.New("TODO: serve until cancellation and shut down gracefully")
}
