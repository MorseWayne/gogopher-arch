package assistance

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

var schemaPattern = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

var errEventNotFound = errors.New("assistance event not found")

type RepositoryOptions struct{ Schema string }

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

func (r *PostgresRepository) Record(ctx context.Context, record Record) (RecordResult, error) {
	payload, err := canonicalPayload(record.Payload)
	if err != nil {
		return RecordResult{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return RecordResult{}, fmt.Errorf("begin assistance event: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return RecordResult{}, err
	}
	var status string
	err = tx.QueryRowContext(ctx, `
		SELECT status
		FROM learning_attempts
		WHERE id = $1 AND learner_id = $2
		FOR UPDATE`, record.AttemptID, record.LearnerID).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return RecordResult{}, ErrAttemptNotFound
	}
	if err != nil {
		return RecordResult{}, fmt.Errorf("lock assistance attempt: %w", err)
	}
	existing, err := scanEvent(tx.QueryRowContext(ctx, eventSelect+` WHERE attempt_id = $1 AND event_key = $2`, record.AttemptID, record.EventKey))
	if err == nil {
		existingPayload, canonicalError := canonicalPayload(existing.Payload)
		if canonicalError != nil {
			return RecordResult{}, canonicalError
		}
		if existing.Type != record.Type || !bytes.Equal(existingPayload, payload) {
			return RecordResult{}, &IdempotencyConflict{EventID: existing.ID}
		}
		if err := tx.Commit(); err != nil {
			return RecordResult{}, fmt.Errorf("commit idempotent assistance event: %w", err)
		}
		return RecordResult{Event: existing}, nil
	}
	if !errors.Is(err, errEventNotFound) {
		return RecordResult{}, err
	}
	if status != "active" {
		return RecordResult{}, ErrAttemptInactive
	}
	var sequence int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(event_seq), 0) + 1 FROM assistance_events WHERE attempt_id = $1`, record.AttemptID).Scan(&sequence); err != nil {
		return RecordResult{}, fmt.Errorf("allocate assistance event sequence: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO assistance_events (id, attempt_id, learner_id, event_key, event_seq, event_type, payload, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7::jsonb,$8)`,
		record.ID, record.AttemptID, record.LearnerID, record.EventKey, sequence, string(record.Type), string(payload), record.CreatedAt)
	if err != nil {
		return RecordResult{}, fmt.Errorf("insert assistance event: %w", err)
	}
	created, err := scanEvent(tx.QueryRowContext(ctx, eventSelect+` WHERE id = $1`, record.ID))
	if err != nil {
		return RecordResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return RecordResult{}, fmt.Errorf("commit assistance event: %w", err)
	}
	return RecordResult{Event: created, Created: true}, nil
}

func (r *PostgresRepository) ListThrough(ctx context.Context, learnerID, attemptID string, cutoff int64) ([]Event, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("begin assistance event read: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return nil, err
	}
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM learning_attempts WHERE id = $1 AND learner_id = $2)`, attemptID, learnerID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check assistance attempt ownership: %w", err)
	}
	if !exists {
		return nil, ErrAttemptNotFound
	}
	rows, err := tx.QueryContext(ctx, eventSelect+` WHERE attempt_id = $1 AND event_seq <= $2 ORDER BY event_seq`, attemptID, cutoff)
	if err != nil {
		return nil, fmt.Errorf("query assistance events: %w", err)
	}
	defer rows.Close()
	var events []Event
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate assistance events: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit assistance event read: %w", err)
	}
	return events, nil
}

const eventSelect = `SELECT id, attempt_id, learner_id, event_key, event_seq, event_type, payload, created_at FROM assistance_events`

type rowScanner interface{ Scan(...any) error }

func scanEvent(row rowScanner) (Event, error) {
	var event Event
	if err := row.Scan(&event.ID, &event.AttemptID, &event.LearnerID, &event.EventKey, &event.Sequence, &event.Type, &event.Payload, &event.CreatedAt); errors.Is(err, sql.ErrNoRows) {
		return Event{}, errEventNotFound
	} else if err != nil {
		return Event{}, fmt.Errorf("scan assistance event: %w", err)
	}
	return event, nil
}

func canonicalPayload(payload []byte) ([]byte, error) {
	if len(payload) == 0 {
		payload = []byte(`{}`)
	}
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil || object == nil {
		return nil, fmt.Errorf("assistance payload must be a JSON object")
	}
	canonical, err := definition.CanonicalJSON(payload)
	if err != nil {
		return nil, fmt.Errorf("canonicalize assistance payload: %w", err)
	}
	return canonical, nil
}

func (r *PostgresRepository) setSearchPath(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `SET LOCAL search_path TO "`+r.schema+`"`)
	if err != nil {
		return fmt.Errorf("set assistance repository search path: %w", err)
	}
	return nil
}
