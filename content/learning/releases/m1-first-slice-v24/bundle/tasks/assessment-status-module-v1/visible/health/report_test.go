package health

import (
	"reflect"
	"testing"
)

func TestSummarizePreservesOrderAndCountsFailures(t *testing.T) {
	got := Summarize([]Result{{Name: "api", OK: true}, {Name: "db", OK: false}})
	if !reflect.DeepEqual(got.Names, []string{"api", "db"}) || got.Failed != 1 || got.ExitCode() != 1 {
		t.Fatalf("Summarize() = %#v exit=%d", got, got.ExitCode())
	}
}
