package postgres

import (
	"context"
	"database/sql"
	"errors"
)

var ErrConflict = errors.New("alert rule version conflict")

type AcknowledgeCommand struct {
	RuleID          string
	IdempotencyKey  string
	Actor           string
	ExpectedVersion int64
}

type Rule struct {
	ID             string
	AcknowledgedBy string
	Version        int64
}

type Store struct{ db *sql.DB }

func NewStore(db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("db required")
	}
	return &Store{db: db}, nil
}

func (s *Store) Acknowledge(ctx context.Context, command AcknowledgeCommand) (Rule, error) {
	return Rule{}, errors.New("TODO: acknowledge the alert transactionally")
}
