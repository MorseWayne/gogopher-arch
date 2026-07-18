package health_test

import (
	"reflect"
	"testing"

	"statusmodule/health"
)

func TestExportedAPIUsableByConsumer(t *testing.T) {
	got := health.Summarize([]health.Result{{Name: "queue", OK: true}, {Name: "cache", OK: true}})
	if !reflect.DeepEqual(got.Names, []string{"queue", "cache"}) || got.Failed != 0 || got.ExitCode() != 0 {
		t.Fatalf("consumer summary = %#v exit=%d", got, got.ExitCode())
	}
}
