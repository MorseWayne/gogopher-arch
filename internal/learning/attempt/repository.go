package attempt

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

var schemaPattern = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

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

func (r *PostgresRepository) Create(ctx context.Context, record CreateRecord) (Attempt, error) {
	capabilities, err := json.Marshal(record.CapabilityRefs)
	if err != nil {
		return Attempt{}, err
	}
	workspace, err := json.Marshal(record.Workspace)
	if err != nil {
		return Attempt{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Attempt{}, fmt.Errorf("begin attempt creation: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return Attempt{}, err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO learning_attempts (
			id, learner_id, release_id, activity_id, activity_version, activity_hash,
			task_id, task_version, task_hash, capability_refs, mode, status,
			workspace, workspace_revision, workspace_hash, started_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11,$12,$13::jsonb,0,$14,$15,$15)`,
		record.ID, record.LearnerID, record.ReleaseID, record.ActivityID, record.ActivityVersion, record.ActivityHash,
		record.TaskID, record.TaskVersion, record.TaskHash, string(capabilities), record.Mode, record.Status,
		string(workspace), record.WorkspaceHash, record.StartedAt)
	if err != nil {
		return Attempt{}, fmt.Errorf("insert learning attempt: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Attempt{}, fmt.Errorf("commit attempt creation: %w", err)
	}
	return cloneAttempt(record.Attempt), nil
}

func (r *PostgresRepository) Get(ctx context.Context, learnerID, attemptID string) (Attempt, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return Attempt{}, fmt.Errorf("begin attempt read: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return Attempt{}, err
	}
	attempt, err := scanAttempt(tx.QueryRowContext(ctx, attemptSelect+` WHERE id = $1 AND learner_id = $2`, attemptID, learnerID))
	if err != nil {
		return Attempt{}, err
	}
	if err := tx.Commit(); err != nil {
		return Attempt{}, fmt.Errorf("commit attempt read: %w", err)
	}
	return attempt, nil
}

func (r *PostgresRepository) Save(ctx context.Context, record SaveRecord) (Attempt, error) {
	workspace, err := json.Marshal(record.Workspace)
	if err != nil {
		return Attempt{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Attempt{}, fmt.Errorf("begin workspace save: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return Attempt{}, err
	}
	var revision int64
	var hash, status string
	err = tx.QueryRowContext(ctx, `
		SELECT workspace_revision, workspace_hash, status
		FROM learning_attempts
		WHERE id = $1 AND learner_id = $2
		FOR UPDATE`, record.AttemptID, record.LearnerID).Scan(&revision, &hash, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return Attempt{}, ErrNotFound
	}
	if err != nil {
		return Attempt{}, fmt.Errorf("lock learning attempt: %w", err)
	}
	if status != "active" {
		return Attempt{}, ErrInactive
	}
	if revision != record.BaseRevision {
		return Attempt{}, &RevisionConflict{Revision: revision, Hash: hash}
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE learning_attempts
		SET workspace = $3::jsonb, workspace_revision = workspace_revision + 1,
			workspace_hash = $4, updated_at = GREATEST(updated_at, $5)
		WHERE id = $1 AND learner_id = $2`, record.AttemptID, record.LearnerID, string(workspace), record.WorkspaceHash, record.UpdatedAt)
	if err != nil {
		return Attempt{}, fmt.Errorf("update attempt workspace: %w", err)
	}
	attempt, err := scanAttempt(tx.QueryRowContext(ctx, attemptSelect+` WHERE id = $1 AND learner_id = $2`, record.AttemptID, record.LearnerID))
	if err != nil {
		return Attempt{}, err
	}
	if err := tx.Commit(); err != nil {
		return Attempt{}, fmt.Errorf("commit workspace save: %w", err)
	}
	return attempt, nil
}

const attemptSelect = `SELECT id, learner_id, release_id, activity_id, activity_version, activity_hash,
	task_id, task_version, task_hash, capability_refs, mode, status, workspace,
	workspace_revision, workspace_hash, started_at, updated_at FROM learning_attempts`

type rowScanner interface{ Scan(...any) error }

func scanAttempt(row rowScanner) (Attempt, error) {
	var result Attempt
	var capabilities, workspace []byte
	err := row.Scan(&result.ID, &result.LearnerID, &result.ReleaseID, &result.ActivityID, &result.ActivityVersion,
		&result.ActivityHash, &result.TaskID, &result.TaskVersion, &result.TaskHash, &capabilities,
		&result.Mode, &result.Status, &workspace, &result.WorkspaceRevision, &result.WorkspaceHash,
		&result.StartedAt, &result.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Attempt{}, ErrNotFound
	}
	if err != nil {
		return Attempt{}, fmt.Errorf("scan learning attempt: %w", err)
	}
	if err := json.Unmarshal(capabilities, &result.CapabilityRefs); err != nil {
		return Attempt{}, fmt.Errorf("decode capability refs: %w", err)
	}
	if err := json.Unmarshal(workspace, &result.Workspace); err != nil {
		return Attempt{}, fmt.Errorf("decode workspace: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) setSearchPath(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `SET LOCAL search_path TO "`+r.schema+`"`)
	if err != nil {
		return fmt.Errorf("set attempt repository search path: %w", err)
	}
	return nil
}

func cloneAttempt(value Attempt) Attempt {
	value.Workspace = cloneWorkspace(value.Workspace)
	value.CapabilityRefs = append([]definition.VersionedDefinitionRef(nil), value.CapabilityRefs...)
	return value
}
