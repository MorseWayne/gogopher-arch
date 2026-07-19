package checks_test

import (
	"context"
	"testing"

	"gocheckhub/internal/checks"
)

type smokeRepository struct {
	created checks.Check
}

func (r *smokeRepository) Create(_ context.Context, check checks.Check) error {
	r.created = check
	return nil
}

func TestPublicUseCaseSmoke(t *testing.T) {
	repository := &smokeRepository{}
	service, err := checks.NewService(repository, func() string { return "public-1" })
	if err != nil {
		t.Fatal(err)
	}
	created, err := service.Create(context.Background(), checks.NewCheck{Target: " api.example.com "})
	if err != nil {
		t.Fatal(err)
	}
	if created.ID != "public-1" || created.Target != "api.example.com" || repository.created != created {
		t.Fatalf("created = %#v stored = %#v", created, repository.created)
	}
}
