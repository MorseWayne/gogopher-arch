package fetcher

import (
	"net/http"
	"testing"
	"time"
)

func TestRetryAfterSecondsSmoke(t *testing.T) {
	response := &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": []string{"42"}}}
	got := NewResponseHandler(response).ParseRetryDelay(time.Unix(1_700_000_000, 0), time.Minute)
	if got != 42*time.Second {
		t.Fatalf("delay=%s", got)
	}
}
