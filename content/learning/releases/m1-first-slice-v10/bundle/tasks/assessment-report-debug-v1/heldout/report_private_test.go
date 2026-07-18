package report

import (
	"bytes"
	"testing"
)

func TestRenderIncludesAllEntries(t *testing.T) {
	entries := []Entry{{Name: "api", Value: 1}, {Name: "db", Value: 2}, {Name: "cache", Value: 3}}
	if got, want := Render(entries), "api=1\ndb=2\ncache=3\n"; got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
	if got := Render(nil); got != "" {
		t.Fatalf("Render(nil) = %q", got)
	}
	var output bytes.Buffer
	LogSummary(&output, "api=1\n")
	if output.String() != "rendered=api=1\n\n" {
		t.Fatalf("LogSummary() = %q", output.String())
	}
}
