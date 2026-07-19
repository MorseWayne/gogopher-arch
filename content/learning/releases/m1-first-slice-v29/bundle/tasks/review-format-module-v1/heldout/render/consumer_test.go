package render_test

import (
	"reflect"
	"testing"

	"formatmodule/render"
)

func TestExportedAPIUsableByConsumer(t *testing.T) {
	got := render.Lines([]render.Record{{Key: "region", Value: "cn"}, {Key: "mode", Value: "prod"}})
	if !reflect.DeepEqual(got.Lines, []string{"region=cn", "mode=prod"}) || got.Count() != 2 {
		t.Fatalf("consumer document = %#v count=%d", got, got.Count())
	}
}
