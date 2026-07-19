package byteframe

import "testing"

func TestDetachPayloadOwnsStorage(t *testing.T) {
	frame := []byte{9, 8, 1, 2, 3}
	payload := DetachPayload(frame, 2)
	frame[2] = 7
	payload[1] = 6
	if payload[0] != 1 || frame[3] != 2 {
		t.Fatalf("frame and payload alias: frame=%v payload=%v", frame, payload)
	}
	if DetachPayload(frame, -1) != nil || DetachPayload(frame, len(frame)+1) != nil {
		t.Fatal("invalid header size must return nil")
	}
}

func TestFrequencyReturnsWritableMap(t *testing.T) {
	empty := Frequency(nil)
	if empty == nil || len(empty) != 0 {
		t.Fatalf("Frequency(nil) = %#v", empty)
	}
	empty[0] = 1
	got := Frequency([]byte{1, 2, 1, 1})
	if got[1] != 3 || got[2] != 1 || len(got) != 2 {
		t.Fatalf("Frequency() = %#v", got)
	}
}
