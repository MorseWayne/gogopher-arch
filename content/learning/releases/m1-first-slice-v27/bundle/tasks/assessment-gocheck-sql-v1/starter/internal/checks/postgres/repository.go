package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gocheckhub/internal/checks"
)

type PoolConfig struct {
	MaxOpen, MaxIdle         int
	MaxLifetime, MaxIdleTime time.Duration
}

func ConfigurePool(db *sql.DB, config PoolConfig) error { return errors.New("TODO: configure pool") }

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) (*Repository, error) { return nil, errors.New("TODO: store DB") }
func (r *Repository) Create(ctx context.Context, check checks.Check) error {
	return errors.New("TODO: ExecContext")
}
func (r *Repository) List(ctx context.Context) ([]checks.Check, error) {
	return nil, errors.New("TODO: QueryContext and close Rows")
}
func (r *Repository) Find(ctx context.Context, id string) (checks.Check, error) {
	return checks.Check{}, errors.New("TODO: QueryRowContext and Scan")
}
