package memory

import (
	"context"
	"errors"

	"gocheckhub/internal/checks"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) Create(ctx context.Context, check checks.Check) error {
	return errors.New("TODO: store checks safely")
}
