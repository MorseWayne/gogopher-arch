package fetcher

import (
	"net/http"
	"testing"
	"time"
)

func TestRetryAfterContract(t *testing.T) {
	now := time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name, header string
		maximum      time.Duration
		want         time.Duration
	}{
		{"seconds", "42", time.Minute, 42 * time.Second},
		{"seconds capped", "120", time.Minute, time.Minute},
		{"date", now.Add(45 * time.Second).Format(time.RFC1123), time.Minute, 45 * time.Second},
		{"past date", now.Add(-time.Second).Format(time.RFC1123), time.Minute, 0},
		{"zero", "0", time.Minute, 0},
		{"negative", "-1", time.Minute, 0},
		{"invalid", "later", time.Minute, 0},
		{"non-positive cap", "42", 0, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			response := &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": []string{tc.header}}}
			if got := NewResponseHandler(response).ParseRetryDelay(now, tc.maximum); got != tc.want {
				t.Fatalf("delay=%s want=%s", got, tc.want)
			}
		})
	}
}
