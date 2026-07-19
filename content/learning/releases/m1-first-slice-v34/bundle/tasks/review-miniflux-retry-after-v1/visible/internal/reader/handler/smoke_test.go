package handler

import (
	"testing"
	"time"
)

type smokeRateLimitResponse struct{ calls int }

func (s *smokeRateLimitResponse) IsRateLimited() bool { return false }
func (s *smokeRateLimitResponse) ParseRetryDelay(time.Time, time.Duration) time.Duration {
	s.calls++
	return time.Hour
}

func TestOrdinaryResponseDoesNotAffectSchedule(t *testing.T) {
	response := &smokeRateLimitResponse{}
	if got := RateLimitDelay(response, time.Unix(1_700_000_000, 0), time.Minute); got != 0 || response.calls != 0 {
		t.Fatalf("delay=%s calls=%d", got, response.calls)
	}
}
