package checks_test

import (
	"context"
	"errors"
	"testing"

	"gocheckhub/internal/checks"
	"gocheckhub/internal/checks/memory"
)

type recordingRepository struct {
	created checks.Check
	err     error
}

func (r *recordingRepository) Create(_ context.Context, check checks.Check) error {
	r.created = check
	return r.err
}

func TestServiceAndMemoryBoundaries(t *testing.T) {
	if _, err := checks.NewService(nil, func() string { return "check-1" }); err == nil {
		t.Fatal("NewService accepted a nil repository")
	}
	if _, err := checks.NewService(&recordingRepository{}, nil); err == nil {
		t.Fatal("NewService accepted a nil ID generator")
	}

	storage := &recordingRepository{}
	service, err := checks.NewService(storage, func() string { return "check-1" })
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.Create(context.Background(), checks.NewCheck{Target: "  api.example.com  "})
	if err != nil {
		t.Fatal(err)
	}
	want := checks.Check{ID: "check-1", Target: "api.example.com"}
	if created != want || storage.created != want {
		t.Fatalf("created = %#v stored = %#v, want %#v", created, storage.created, want)
	}
	if _, err := service.Create(context.Background(), checks.NewCheck{Target: "  "}); !errors.Is(err, checks.ErrInvalidTarget) {
		t.Fatalf("invalid target error = %v", err)
	}

	repository := memory.NewRepository()
	if err := repository.Create(context.Background(), checks.Check{ID: "one", Target: "API.EXAMPLE.COM"}); err != nil {
		t.Fatal(err)
	}
	if err := repository.Create(context.Background(), checks.Check{ID: "two", Target: "api.example.com"}); !errors.Is(err, checks.ErrCheckExists) {
		t.Fatalf("duplicate error = %v", err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := repository.Create(cancelled, checks.Check{ID: "three", Target: "other.example.com"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled create error = %v", err)
	}
}
