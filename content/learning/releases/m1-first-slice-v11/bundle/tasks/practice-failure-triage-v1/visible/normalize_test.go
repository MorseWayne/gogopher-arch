package normalize

import "testing"

func TestNormalizeHandlesEveryInputWithoutPanic(t *testing.T) {
	got := Normalize([]string{" Alice ", "BOB", " 陈 "})
	want := []string{"alice", "bob", "陈"}
	if len(got) != len(want) {
		t.Fatalf("Normalize() length = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("Normalize()[%d] = %q, want %q", index, got[index], want[index])
		}
	}
}
