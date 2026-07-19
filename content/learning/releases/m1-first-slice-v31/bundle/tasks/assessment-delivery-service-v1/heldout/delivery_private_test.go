package delivery

import (
	"errors"
	"testing"
)

type senderSubstitute struct {
	messages []Message
	failAt   int
	err      error
}

func (s *senderSubstitute) Send(message Message) error {
	s.messages = append(s.messages, message)
	if len(s.messages) == s.failAt {
		return s.err
	}
	return nil
}

func TestServiceUsesSubstituteAndPropagatesError(t *testing.T) {
	wantErr := errors.New("delivery rejected")
	stub := &senderSubstitute{failAt: 2, err: wantErr}
	messages := []Message{{Recipient: "a"}, {Recipient: "b"}, {Recipient: "c"}}
	err := New(stub).Deliver(messages)
	if !errors.Is(err, wantErr) {
		t.Fatalf("Deliver() error = %v, want %v", err, wantErr)
	}
	if len(stub.messages) != 2 {
		t.Fatalf("Send calls = %d, want 2", len(stub.messages))
	}
}

func TestIndexByIsReusableAcrossTypes(t *testing.T) {
	type item struct {
		ID    int
		Value string
	}
	indexed := IndexBy([]item{{ID: 1, Value: "old"}, {ID: 2, Value: "two"}, {ID: 1, Value: "new"}}, func(value item) int { return value.ID })
	if len(indexed) != 2 || indexed[1].Value != "new" {
		t.Fatalf("IndexBy() = %#v", indexed)
	}
	empty := IndexBy([]int(nil), func(value int) int { return value })
	empty[1] = 1
}
