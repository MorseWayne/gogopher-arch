package memory

import (
	"context"
	"errors"

	"gocheckhub/internal/alerts"
)

type Store struct{}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Save(ctx context.Context, rule alerts.Rule) error {
	return errors.New("TODO: save alert rules safely")
}
