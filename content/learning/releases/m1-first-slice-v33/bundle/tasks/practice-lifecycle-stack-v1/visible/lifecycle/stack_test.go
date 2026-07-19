package lifecycle

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestStackClosesOnceInReverseOrderAndJoinsErrors(t *testing.T) {
	var stack Stack
	order := []string{}
	firstErr, secondErr := errors.New("first"), errors.New("second")
	for _, entry := range []struct {
		name string
		err  error
	}{{name: "database", err: firstErr}, {name: "worker"}, {name: "server", err: secondErr}} {
		entry := entry
		if err := stack.Push(func(context.Context) error { order = append(order, entry.name); return entry.err }); err != nil {
			t.Fatal(err)
		}
	}
	err := stack.Close(context.Background())
	if !reflect.DeepEqual(order, []string{"server", "worker", "database"}) || !errors.Is(err, firstErr) || !errors.Is(err, secondErr) {
		t.Fatalf("order=%v error=%v", order, err)
	}
	if err := stack.Close(context.Background()); err != nil || len(order) != 3 {
		t.Fatalf("second Close: order=%v error=%v", order, err)
	}
	if err := stack.Push(nil); err == nil {
		t.Fatal("Push accepted nil closer")
	}
}
