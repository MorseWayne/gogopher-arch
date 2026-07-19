package checks_test

import (
	"context"
	"errors"
	"testing"

	"gocheckhub/internal/checks"
)

type controlledRepository struct {
	created checks.Check
	err     error
}

func (r *controlledRepository) Create(_ context.Context, check checks.Check) error {
	r.created = check
	return r.err
}

func TestUseCaseContract(t *testing.T) {
	if _, err := checks.NewService(nil, func() string { return "id" }); err == nil {
		t.Fatal("NewService accepted nil repository")
	}
	if _, err := checks.NewService(&controlledRepository{}, nil); err == nil {
		t.Fatal("NewService accepted nil nextID")
	}

	repository := &controlledRepository{}
	service, err := checks.NewService(repository, func() string { return "check-42" })
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.Create(context.Background(), checks.NewCheck{Target: "  API.EXAMPLE.COM  "})
	if err != nil {
		t.Fatal(err)
	}
	want := checks.Check{ID: "check-42", Target: "API.EXAMPLE.COM"}
	if created != want || repository.created != want {
		t.Fatalf("created = %#v stored = %#v, want %#v", created, repository.created, want)
	}
	if _, err := service.Create(context.Background(), checks.NewCheck{Target: "\t "}); !errors.Is(err, checks.ErrInvalidTarget) {
		t.Fatalf("empty target error = %v", err)
	}
	sentinel := errors.New("storage unavailable")
	repository.err = sentinel
	if _, err := service.Create(context.Background(), checks.NewCheck{Target: "other"}); !errors.Is(err, sentinel) {
		t.Fatalf("storage error = %v", err)
	}
}
