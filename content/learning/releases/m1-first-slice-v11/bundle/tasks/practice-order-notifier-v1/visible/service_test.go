package order

import (
	"errors"
	"testing"
)

type notifierSpy struct {
	userID  string
	message string
	calls   int
	err     error
}

func (s *notifierSpy) Notify(userID, message string) error {
	s.userID, s.message = userID, message
	s.calls++
	return s.err
}

func TestPlaceUsesNotifierContractAndPropagatesError(t *testing.T) {
	wantErr := errors.New("mail unavailable")
	spy := &notifierSpy{err: wantErr}
	err := NewService(spy).Place("u-7", "book")
	if !errors.Is(err, wantErr) {
		t.Fatalf("Place() error = %v, want %v", err, wantErr)
	}
	if spy.calls != 1 || spy.userID != "u-7" || spy.message != "order placed: book" {
		t.Fatalf("notifier calls=%d user=%q message=%q", spy.calls, spy.userID, spy.message)
	}
}
