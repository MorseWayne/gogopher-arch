package httpclient

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return fn(request) }

type trackedBody struct {
	*bytes.Reader
	closed atomic.Bool
}

func newTrackedBody(value string) *trackedBody {
	return &trackedBody{Reader: bytes.NewReader([]byte(value))}
}
func (body *trackedBody) Close() error { body.closed.Store(true); return nil }

func TestRequestContractAndBodyLifecycle(t *testing.T) {
	body := newTrackedBody(`{"status":"ok"}`)
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.Method != http.MethodGet || request.URL.Path != "/v1/probe" || request.URL.Query().Get("target") != "https://service.example/a b" || request.Header.Get("Accept") != "application/json" {
			t.Fatalf("request = %#v", request)
		}
		return &http.Response{StatusCode: http.StatusOK, Body: body, Header: make(http.Header)}, nil
	})}
	client, err := New(httpClient, "https://probe.example", 1024)
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Probe(context.Background(), "https://service.example/a b")
	if err != nil || result.Status != "ok" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	if !body.closed.Load() {
		t.Fatal("response body was not closed")
	}
}

func TestCancellationPropagates(t *testing.T) {
	started := make(chan struct{})
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		close(started)
		<-request.Context().Done()
		return nil, request.Context().Err()
	})}
	client, err := New(httpClient, "https://probe.example", 1024)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { _, err := client.Probe(ctx, "https://service.example"); done <- err }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error = %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("request ignored cancellation")
	}
}

func TestResponseBodyBounded(t *testing.T) {
	body := newTrackedBody("123456789")
	httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: body, Header: make(http.Header)}, nil
	})}
	client, _ := New(httpClient, "https://probe.example", 8)
	_, err := client.Probe(context.Background(), "https://service.example")
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Fatalf("error = %v", err)
	}
	if !body.closed.Load() {
		t.Fatal("oversized body was not closed")
	}
}

func TestFailureBoundariesAndNoRetry(t *testing.T) {
	tests := []struct {
		status   int
		expected error
	}{{http.StatusTooManyRequests, ErrRateLimited}, {http.StatusBadRequest, ErrRejected}, {http.StatusServiceUnavailable, ErrUpstream}}
	for _, test := range tests {
		t.Run(http.StatusText(test.status), func(t *testing.T) {
			var calls atomic.Int32
			body := newTrackedBody("failure")
			httpClient := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
				calls.Add(1)
				return &http.Response{StatusCode: test.status, Body: body, Header: make(http.Header)}, nil
			})}
			client, _ := New(httpClient, "https://probe.example", 1024)
			_, err := client.Probe(context.Background(), "https://service.example")
			if !errors.Is(err, test.expected) {
				t.Fatalf("error = %v", err)
			}
			if calls.Load() != 1 {
				t.Fatalf("requests = %d", calls.Load())
			}
			if !body.closed.Load() {
				t.Fatal("failure body was not closed")
			}
		})
	}
}

var _ io.ReadCloser = (*trackedBody)(nil)
