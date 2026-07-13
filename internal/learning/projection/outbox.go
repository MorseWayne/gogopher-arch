package projection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	projectionRequestTopic = "capability_projection.requested"
	reviewSchedulerTopic   = "review_scheduler.requested"
)

type PostgresRequestRepository struct {
	db     *sql.DB
	schema string
	topic  string
}

func NewPostgresRequestRepository(db *sql.DB, options RepositoryOptions) (*PostgresRequestRepository, error) {
	return NewPostgresTopicRequestRepository(db, projectionRequestTopic, options)
}

func NewPostgresReviewSchedulerRequestRepository(db *sql.DB, options RepositoryOptions) (*PostgresRequestRepository, error) {
	return NewPostgresTopicRequestRepository(db, reviewSchedulerTopic, options)
}

func NewPostgresTopicRequestRepository(db *sql.DB, topic string, options RepositoryOptions) (*PostgresRequestRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	if topic == "" {
		return nil, fmt.Errorf("outbox topic is required")
	}
	schema := options.Schema
	if schema == "" {
		schema = "public"
	}
	if !projectionSchemaPattern.MatchString(schema) {
		return nil, fmt.Errorf("invalid PostgreSQL schema %q", schema)
	}
	return &PostgresRequestRepository{db: db, schema: schema, topic: topic}, nil
}

func (r *PostgresRequestRepository) ClaimRequest(ctx context.Context, owner string, now time.Time, lease time.Duration) (Request, bool, error) {
	if owner == "" || lease <= 0 {
		return Request{}, false, fmt.Errorf("outbox lease owner and duration are required")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Request{}, false, fmt.Errorf("begin outbox request claim: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return Request{}, false, err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE learning_outbox
		SET status = 'pending', lease_owner = NULL, lease_expires_at = NULL, available_at = $1
		WHERE topic = $2 AND status = 'processing' AND lease_expires_at <= $1`, now, r.topic); err != nil {
		return Request{}, false, fmt.Errorf("recover expired outbox request: %w", err)
	}
	var request Request
	err = tx.QueryRowContext(ctx, `
		WITH candidate AS (
			SELECT id FROM learning_outbox
			WHERE topic = $1 AND status = 'pending' AND available_at <= $2
			ORDER BY available_at, created_at, id
			FOR UPDATE SKIP LOCKED LIMIT 1
		)
		UPDATE learning_outbox o
		SET status = 'processing', attempt_count = attempt_count + 1,
			lease_owner = $3, lease_expires_at = $4
		FROM candidate
		WHERE o.id = candidate.id
		RETURNING o.id, o.payload, o.attempt_count`,
		r.topic, now, owner, now.Add(lease)).Scan(
		&request.ID, &request.Payload, &request.AttemptCount)
	if errors.Is(err, sql.ErrNoRows) {
		if err := tx.Commit(); err != nil {
			return Request{}, false, fmt.Errorf("commit empty projection claim: %w", err)
		}
		return Request{}, false, nil
	}
	if err != nil {
		return Request{}, false, fmt.Errorf("claim outbox request: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Request{}, false, fmt.Errorf("commit outbox request claim: %w", err)
	}
	return request, true, nil
}

func (r *PostgresRequestRepository) CompleteRequest(ctx context.Context, requestID, owner, consumer string, consumerVersion int, now time.Time) error {
	if requestID == "" || owner == "" || consumer == "" || consumerVersion < 1 {
		return fmt.Errorf("outbox completion identity and consumer version are required")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin outbox request completion: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE learning_outbox
		SET status = 'completed', lease_owner = NULL, lease_expires_at = NULL,
			consumer = $3, consumer_version = $4, completed_at = $5, last_error = NULL
		WHERE id = $1 AND topic = $6 AND status = 'processing'
			AND lease_owner = $2 AND lease_expires_at > $5`,
		requestID, owner, consumer, consumerVersion, now, r.topic)
	if err != nil {
		return fmt.Errorf("complete outbox request: %w", err)
	}
	if rows, err := result.RowsAffected(); err != nil || rows != 1 {
		return fmt.Errorf("outbox request lease was lost")
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit outbox request completion: %w", err)
	}
	return nil
}

func (r *PostgresRequestRepository) RetryRequest(ctx context.Context, requestID, owner string, now time.Time, delay time.Duration, maxAttempts int, summary string) (RetryResult, error) {
	if requestID == "" || owner == "" || delay <= 0 || maxAttempts < 1 {
		return RetryResult{}, fmt.Errorf("outbox retry identity, delay, and attempt limit are required")
	}
	summary = summarizeError(summary)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return RetryResult{}, fmt.Errorf("begin outbox request retry: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return RetryResult{}, err
	}
	var result RetryResult
	var status string
	var availableAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		UPDATE learning_outbox
		SET status = CASE WHEN attempt_count >= $4 THEN 'failed' ELSE 'pending' END,
			available_at = CASE WHEN attempt_count >= $4 THEN available_at ELSE $5 END,
			lease_owner = NULL, lease_expires_at = NULL, last_error = $6,
			failed_at = CASE WHEN attempt_count >= $4 THEN $3 ELSE NULL END
		WHERE id = $1 AND topic = $7 AND status = 'processing'
			AND lease_owner = $2 AND lease_expires_at > $3
		RETURNING attempt_count, status, available_at`,
		requestID, owner, now, maxAttempts, now.Add(delay), summary, r.topic).Scan(
		&result.AttemptCount, &status, &availableAt)
	if errors.Is(err, sql.ErrNoRows) {
		return RetryResult{}, fmt.Errorf("outbox request lease was lost")
	}
	if err != nil {
		return RetryResult{}, fmt.Errorf("retry outbox request: %w", err)
	}
	result.Exhausted = status == "failed"
	if !result.Exhausted && availableAt.Valid {
		value := availableAt.Time.UTC()
		result.AvailableAt = &value
	}
	if err := tx.Commit(); err != nil {
		return RetryResult{}, fmt.Errorf("commit outbox request retry: %w", err)
	}
	return result, nil
}

func (r *PostgresRequestRepository) setSearchPath(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `SET LOCAL search_path TO "`+r.schema+`"`); err != nil {
		return fmt.Errorf("set outbox search path: %w", err)
	}
	return nil
}

func summarizeError(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if value == "" {
		return "projection failed"
	}
	const maximumBytes = 512
	for len(value) > maximumBytes {
		_, size := utf8.DecodeLastRuneInString(value)
		value = value[:len(value)-size]
	}
	return value
}
