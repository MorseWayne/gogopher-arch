package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"gocheckhub/internal/alerts"
)

type PoolConfig struct {
	MaxOpen, MaxIdle         int
	MaxLifetime, MaxIdleTime time.Duration
}

func ConfigurePool(db *sql.DB, config PoolConfig) error { return errors.New("TODO: configure pool") }

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) (*Repository, error) { return nil, errors.New("TODO: store DB") }
func (r *Repository) Save(ctx context.Context, rule alerts.Rule) error {
	return errors.New("TODO: ExecContext")
}
func (r *Repository) List(ctx context.Context) ([]alerts.Rule, error) {
	return nil, errors.New("TODO: QueryContext")
}
func (r *Repository) Find(ctx context.Context, id string) (alerts.Rule, error) {
	return alerts.Rule{}, errors.New("TODO: QueryRowContext")
}
