package testkit

import (
	"errors"
	"sync"
	"time"
)

type ManualClock struct {
	mu  sync.RWMutex
	now time.Time
}

func NewManualClock(start time.Time) *ManualClock {
	// TODO: retain the explicit test start time.
	return &ManualClock{}
}

func (clock *ManualClock) Now() time.Time {
	// TODO: return a concurrency-safe snapshot.
	return time.Time{}
}

func (clock *ManualClock) Advance(duration time.Duration) error {
	if duration < 0 {
		return errors.New("manual clock cannot move backwards")
	}
	// TODO: advance atomically.
	return nil
}
