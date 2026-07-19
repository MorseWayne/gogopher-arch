package handler

import (
	"testing"
	"time"
)

type contractRateLimitResponse struct {
	limited bool
	delay   time.Duration
	calls   int
	now     time.Time
	maximum time.Duration
}

func (s *contractRateLimitResponse) IsRateLimited() bool { return s.limited }
func (s *contractRateLimitResponse) ParseRetryDelay(now time.Time, maximum time.Duration) time.Duration {
	s.calls++
	s.now = now
	s.maximum = maximum
	return s.delay
}

func TestRateLimitChainContract(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	ordinary := &contractRateLimitResponse{delay: time.Hour}
	if got := RateLimitDelay(ordinary, now, time.Minute); got != 0 || ordinary.calls != 0 {
		t.Fatalf("ordinary delay=%s calls=%d", got, ordinary.calls)
	}
	limited := &contractRateLimitResponse{limited: true, delay: 30 * time.Second}
	if got := RateLimitDelay(limited, now, time.Minute); got != 30*time.Second || limited.calls != 1 || !limited.now.Equal(now) || limited.maximum != time.Minute {
		t.Fatalf("limited delay=%s response=%#v", got, limited)
	}
}
