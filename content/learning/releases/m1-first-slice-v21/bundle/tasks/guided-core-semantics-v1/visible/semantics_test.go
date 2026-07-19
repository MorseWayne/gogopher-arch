package semantics

import (
	"reflect"
	"testing"
)

func TestClassifyNamesCountsRunes(t *testing.T) {
	got := ClassifyNames([]string{" Go ", "你好", "", "gopher"}, 2)
	want := Summary{Accepted: []Label{"Go", "你好"}, Rejected: 2, RuneCount: 4}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ClassifyNames() = %#v, want %#v", got, want)
	}
}
