package alerts_test

import (
	"context"
	"errors"
	"testing"

	"gocheckhub/internal/alerts"
)

type controlledStore struct {
	saved alerts.Rule
	err   error
}

func (s *controlledStore) Save(_ context.Context, rule alerts.Rule) error {
	s.saved = rule
	return s.err
}

func TestUseCaseContract(t *testing.T) {
	if _, err := alerts.NewManager(nil, func() string { return "id" }); err == nil {
		t.Fatal("NewManager accepted nil store")
	}
	if _, err := alerts.NewManager(&controlledStore{}, nil); err == nil {
		t.Fatal("NewManager accepted nil nextID")
	}
	store := &controlledStore{}
	manager, err := alerts.NewManager(store, func() string { return "alert-42" })
	if err != nil {
		t.Fatal(err)
	}
	created, err := manager.Publish(context.Background(), alerts.NewRule{Destination: "  OPS@EXAMPLE.COM  "})
	if err != nil {
		t.Fatal(err)
	}
	want := alerts.Rule{ID: "alert-42", Destination: "OPS@EXAMPLE.COM"}
	if created != want || store.saved != want {
		t.Fatalf("created = %#v saved = %#v, want %#v", created, store.saved, want)
	}
	if _, err := manager.Publish(context.Background(), alerts.NewRule{Destination: "\t "}); !errors.Is(err, alerts.ErrInvalidDestination) {
		t.Fatalf("empty destination error = %v", err)
	}
	sentinel := errors.New("store unavailable")
	store.err = sentinel
	if _, err := manager.Publish(context.Background(), alerts.NewRule{Destination: "other@example.com"}); !errors.Is(err, sentinel) {
		t.Fatalf("storage error = %v", err)
	}
}
