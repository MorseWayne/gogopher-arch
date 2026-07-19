package report

import (
	"reflect"
	"testing"
)

func TestSummarizeKeepsDomainLogicInPackage(t *testing.T) {
	got := Summarize([]Result{{Name: "api", OK: true}, {Name: "db", OK: false}})
	if !reflect.DeepEqual(got.Names, []string{"api", "db"}) || got.Failed != 1 {
		t.Fatalf("Summarize() = %#v", got)
	}
}
