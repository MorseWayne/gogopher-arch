package adminserver

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

type JobLookup func(context.Context, string) (string, bool)

type Dependencies struct {
	Ready       func(context.Context) bool
	LookupJob   JobLookup
	NextTraceID func() string
}

type Timeouts struct {
	ReadHeader time.Duration
	Read       time.Duration
	Write      time.Duration
	Idle       time.Duration
}

func NewHandler(dependencies Dependencies) (http.Handler, error) {
	return nil, errors.New("TODO: build admin routes and trace middleware")
}

func TraceID(ctx context.Context) string {
	return ""
}

func NewServer(handler http.Handler, timeouts Timeouts) (*http.Server, error) {
	return nil, errors.New("TODO: validate and configure http.Server")
}

func Serve(ctx context.Context, server *http.Server, listener net.Listener, shutdownTimeout time.Duration) error {
	return errors.New("TODO: serve until cancellation and shut down gracefully")
}
