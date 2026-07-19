package retry

import (
	"testing"
	"time"
)

func TestPolicyNext(t *testing.T) {
	policy := Policy{Base: time.Second, Max: 5 * time.Second, MaxAttempts: 5}
	tests := []struct {
		attempt int
		delay   time.Duration
		ok      bool
	}{{1, time.Second, true}, {2, 2 * time.Second, true}, {3, 4 * time.Second, true}, {4, 5 * time.Second, true}, {5, 0, false}}
	for _, test := range tests {
		if delay, ok := policy.Next(test.attempt); delay != test.delay || ok != test.ok {
			t.Fatalf("Next(%d) = (%s, %v), want (%s, %v)", test.attempt, delay, ok, test.delay, test.ok)
		}
	}
}

func TestPolicyRejectsInvalidInput(t *testing.T) {
	for _, policy := range []Policy{{}, {Base: time.Second, Max: time.Second}} {
		if _, ok := policy.Next(1); ok {
			t.Fatal("invalid policy allowed a retry")
		}
	}
}
