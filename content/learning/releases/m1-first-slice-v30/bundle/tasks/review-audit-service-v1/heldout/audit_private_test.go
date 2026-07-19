package audit

import (
	"errors"
	"testing"
)

type sinkSubstitute struct {
	events []Event
	err    error
}

func (s *sinkSubstitute) Write(event Event) error {
	s.events = append(s.events, event)
	if len(s.events) == 2 {
		return s.err
	}
	return nil
}

func TestServiceUsesSubstituteAndPropagatesError(t *testing.T) {
	wantErr := errors.New("audit unavailable")
	sink := &sinkSubstitute{err: wantErr}
	err := New(sink).Append([]Event{{Actor: "a"}, {Actor: "b"}, {Actor: "c"}})
	if !errors.Is(err, wantErr) || len(sink.events) != 2 {
		t.Fatalf("Append() error=%v events=%v", err, sink.events)
	}
}

func TestGroupByIsReusableAcrossTypes(t *testing.T) {
	type sample struct {
		Group int
		Name  string
	}
	grouped := GroupBy([]sample{{Group: 1, Name: "a"}, {Group: 2, Name: "b"}, {Group: 1, Name: "c"}}, func(value sample) int { return value.Group })
	if len(grouped) != 2 || len(grouped[1]) != 2 || grouped[1][1].Name != "c" {
		t.Fatalf("GroupBy() = %#v", grouped)
	}
	empty := GroupBy([]int(nil), func(value int) int { return value })
	empty[1] = []int{1}
}
