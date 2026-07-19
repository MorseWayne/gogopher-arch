package retry

import "time"

type Policy struct {
	Base        time.Duration
	Max         time.Duration
	MaxAttempts int
}

// Next returns the delay after a failed attempt and whether another attempt is allowed.
func (policy Policy) Next(attempt int) (time.Duration, bool) {
	// TODO: Please implement
	return 0, false
}
