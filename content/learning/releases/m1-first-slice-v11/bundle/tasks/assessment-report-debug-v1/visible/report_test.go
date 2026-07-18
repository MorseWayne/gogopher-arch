package report

import "testing"

func TestRenderIncludesLastEntry(t *testing.T) {
	got := Render([]Entry{{Name: "api", Value: 1}, {Name: "db", Value: 2}})
	want := "api=1\ndb=2\n"
	if got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
}
