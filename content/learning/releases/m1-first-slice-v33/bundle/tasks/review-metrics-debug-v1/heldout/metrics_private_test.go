package metrics

import (
	"bytes"
	"testing"
)

func TestRenderIncludesAllEntries(t *testing.T) {
	values := []Metric{{Name: "requests", Value: 5}, {Name: "errors", Value: 1}}
	if got, want := Render(values), "requests:5\nerrors:1\n"; got != want {
		t.Fatalf("Render() = %q, want %q", got, want)
	}
	var output bytes.Buffer
	Debug(&output, "requests:5\n")
	if output.String() != "metrics=requests:5\n\n" {
		t.Fatalf("Debug() = %q", output.String())
	}
}
