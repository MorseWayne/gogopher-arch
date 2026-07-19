package postgres

import (
	"context"
	"database/sql"
	"errors"
)

var ErrConflict = errors.New("run version conflict")

type CompleteCommand struct {
	RunID           string
	IdempotencyKey  string
	Status          string
	ExpectedVersion int64
}

type Run struct {
	ID      string
	Status  string
	Version int64
}

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("db required")
	}
	return &Store{db: db}, nil
}

func (s *Store) Complete(ctx context.Context, command CompleteCommand) (Run, error) {
	return Run{}, errors.New("TODO: complete the run transactionally")
}
