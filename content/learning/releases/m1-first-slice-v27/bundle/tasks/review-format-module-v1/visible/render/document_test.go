package render

import (
	"reflect"
	"testing"
)

func TestLinesPreservesOrder(t *testing.T) {
	got := Lines([]Record{{Key: "api", Value: "up"}, {Key: "db", Value: "down"}})
	if !reflect.DeepEqual(got.Lines, []string{"api=up", "db=down"}) || got.Count() != 2 {
		t.Fatalf("Lines() = %#v count=%d", got, got.Count())
	}
}
