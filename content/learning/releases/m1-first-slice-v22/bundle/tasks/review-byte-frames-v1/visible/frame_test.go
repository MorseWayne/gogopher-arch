package byteframe

import "testing"

func TestDetachPayloadAndFrequency(t *testing.T) {
	frame := []byte{0x01, 0x02, 'a', 'b', 'a'}
	payload := DetachPayload(frame, 2)
	if string(payload) != "aba" {
		t.Fatalf("DetachPayload() = %q", payload)
	}
	counts := Frequency(payload)
	if counts['a'] != 2 || counts['b'] != 1 {
		t.Fatalf("Frequency() = %#v", counts)
	}
}
