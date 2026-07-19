package delivery

import "testing"

type recordingSender struct {
	messages []Message
}

func (s *recordingSender) Send(message Message) error {
	s.messages = append(s.messages, message)
	return nil
}

func TestDeliverUsesSmallSenderContract(t *testing.T) {
	spy := &recordingSender{}
	service := New(spy)
	want := []Message{{Recipient: "a", Body: "one"}, {Recipient: "b", Body: "two"}}
	if err := service.Deliver(want); err != nil {
		t.Fatal(err)
	}
	if len(spy.messages) != len(want) || spy.messages[1] != want[1] {
		t.Fatalf("recorded messages = %#v", spy.messages)
	}
}
