package session

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

var schemaPattern = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

type RepositoryOptions struct {
	Schema string
}

type PostgresRepository struct {
	db     *sql.DB
	schema string
}

func NewPostgresRepository(db *sql.DB, options RepositoryOptions) (*PostgresRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	schema := options.Schema
	if schema == "" {
		schema = "public"
	}
	if !schemaPattern.MatchString(schema) {
		return nil, fmt.Errorf("invalid PostgreSQL schema %q", schema)
	}
	return &PostgresRepository{db: db, schema: schema}, nil
}

func (r *PostgresRepository) FindActive(ctx context.Context, tokenHash string, now time.Time) (Session, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Session{}, fmt.Errorf("begin session lookup: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return Session{}, err
	}
	var active Session
	err = tx.QueryRowContext(ctx, `
		UPDATE learner_sessions
		SET last_used_at = GREATEST(last_used_at, $2)
		WHERE token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > $2
		RETURNING id, learner_id, created_at, expires_at, last_used_at`, tokenHash, now).
		Scan(&active.ID, &active.LearnerID, &active.CreatedAt, &active.ExpiresAt, &active.LastUsedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Session{}, ErrNotFound
		}
		return Session{}, fmt.Errorf("find active session: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Session{}, fmt.Errorf("commit session lookup: %w", err)
	}
	return active, nil
}

func (r *PostgresRepository) Create(ctx context.Context, input NewSession) (Session, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Session{}, fmt.Errorf("begin session creation: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return Session{}, err
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO learners (id, created_at)
		VALUES ($1, $2)`, input.LearnerID, input.CreatedAt); err != nil {
		return Session{}, fmt.Errorf("insert learner: %w", err)
	}
	var created Session
	err = tx.QueryRowContext(ctx, `
		INSERT INTO learner_sessions (
			id, learner_id, token_hash, created_at, expires_at, last_used_at
		) VALUES ($1, $2, $3, $4, $5, $4)
		RETURNING id, learner_id, created_at, expires_at, last_used_at`,
		input.ID, input.LearnerID, input.TokenHash, input.CreatedAt, input.ExpiresAt).
		Scan(&created.ID, &created.LearnerID, &created.CreatedAt, &created.ExpiresAt, &created.LastUsedAt)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" && postgresError.ConstraintName == "learner_sessions_token_hash_key" {
			return Session{}, ErrTokenCollision
		}
		return Session{}, fmt.Errorf("insert learner session: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Session{}, fmt.Errorf("commit session creation: %w", err)
	}
	return created, nil
}

func (r *PostgresRepository) setSearchPath(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `SET LOCAL search_path TO `+quoteIdentifier(r.schema)); err != nil {
		return fmt.Errorf("set session repository search path: %w", err)
	}
	return nil
}

func quoteIdentifier(identifier string) string {
	return `"` + identifier + `"`
}
