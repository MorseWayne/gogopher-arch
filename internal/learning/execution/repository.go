package execution

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"
)

var executionSchemaPattern = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

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
	if !executionSchemaPattern.MatchString(schema) {
		return nil, fmt.Errorf("invalid PostgreSQL schema %q", schema)
	}
	return &PostgresRepository{db: db, schema: schema}, nil
}

func (r *PostgresRepository) FindNormal(ctx context.Context, learnerID, attemptID, requestKey string) (Execution, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return Execution{}, fmt.Errorf("begin execution idempotency read: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return Execution{}, err
	}
	value, err := scanExecution(tx.QueryRowContext(ctx, executionSelect+`
		JOIN learning_attempts a ON a.id = e.attempt_id
		WHERE e.attempt_id = $1 AND e.request_key = $2 AND e.submission_id IS NULL AND a.learner_id = $3`,
		attemptID, requestKey, learnerID))
	if err != nil {
		return Execution{}, err
	}
	if err := tx.Commit(); err != nil {
		return Execution{}, fmt.Errorf("commit execution idempotency read: %w", err)
	}
	return value, nil
}

func (r *PostgresRepository) Get(ctx context.Context, learnerID, executionID string) (Execution, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return Execution{}, fmt.Errorf("begin execution read: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return Execution{}, err
	}
	value, err := scanExecution(tx.QueryRowContext(ctx, executionSelect+`
		JOIN learning_attempts a ON a.id = e.attempt_id
		WHERE e.id = $1 AND a.learner_id = $2`, executionID, learnerID))
	if err != nil {
		return Execution{}, err
	}
	if err := tx.Commit(); err != nil {
		return Execution{}, fmt.Errorf("commit execution read: %w", err)
	}
	return value, nil
}

func (r *PostgresRepository) CreateNormal(ctx context.Context, record CreateNormalRecord) (Execution, bool, error) {
	if record.Action != ActionBuild && record.Action != ActionTest && record.Action != ActionVet {
		return Execution{}, false, fmt.Errorf("normal execution action must be build, test, or vet")
	}
	if err := record.Spec.Validate(); err != nil {
		return Execution{}, false, err
	}
	if record.Spec.ExecutionID != record.ID || record.Spec.Action != record.Action {
		return Execution{}, false, fmt.Errorf("execution spec identity does not match queue record")
	}
	specJSON, err := json.Marshal(record.Spec)
	if err != nil {
		return Execution{}, false, fmt.Errorf("encode execution spec: %w", err)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Execution{}, false, fmt.Errorf("begin execution creation: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return Execution{}, false, err
	}
	var releaseID, taskID, taskHash, workspaceHash, status string
	var taskVersion int
	var workspaceRevision int64
	err = tx.QueryRowContext(ctx, `
		SELECT release_id, task_id, task_version, task_hash, workspace_revision, workspace_hash, status
		FROM learning_attempts
		WHERE id = $1 AND learner_id = $2
		FOR UPDATE`, record.AttemptID, record.LearnerID).Scan(
		&releaseID, &taskID, &taskVersion, &taskHash, &workspaceRevision, &workspaceHash, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return Execution{}, false, ErrExecutionNotFound
	}
	if err != nil {
		return Execution{}, false, fmt.Errorf("lock execution attempt: %w", err)
	}
	existing, err := scanExecution(tx.QueryRowContext(ctx, executionSelect+`
		WHERE e.attempt_id = $1 AND e.request_key = $2 AND e.submission_id IS NULL`, record.AttemptID, record.RequestKey))
	if err == nil {
		if existing.RequestFingerprint != record.RequestFingerprint {
			return Execution{}, false, &IdempotencyConflict{ExecutionID: existing.ID}
		}
		if err := tx.Commit(); err != nil {
			return Execution{}, false, fmt.Errorf("commit idempotent execution read: %w", err)
		}
		return existing, false, nil
	}
	if !errors.Is(err, ErrExecutionNotFound) {
		return Execution{}, false, err
	}
	if status != "active" || releaseID != record.ReleaseID || taskID != record.TaskID || taskVersion != record.TaskVersion || taskHash != record.TaskHash {
		return Execution{}, false, ErrAttemptUnavailable
	}
	if workspaceRevision != record.WorkspaceRevision || workspaceHash != record.WorkspaceHash {
		return Execution{}, false, &WorkspaceConflict{Revision: workspaceRevision, Hash: workspaceHash}
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO attempt_executions (
			id, attempt_id, action, sequence, request_key, request_fingerprint,
			release_id, task_id, task_version, task_hash, workspace_revision, workspace_hash,
			spec, status, created_at, updated_at
		) VALUES ($1,$2,$3,0,$4,$5,$6,$7,$8,$9,$10,$11,$12::jsonb,'queued',$13,$13)`,
		record.ID, record.AttemptID, string(record.Action), record.RequestKey, record.RequestFingerprint,
		record.ReleaseID, record.TaskID, record.TaskVersion, record.TaskHash,
		record.WorkspaceRevision, record.WorkspaceHash, string(specJSON), record.CreatedAt)
	if err != nil {
		return Execution{}, false, fmt.Errorf("insert queued execution: %w", err)
	}
	created, err := scanExecution(tx.QueryRowContext(ctx, executionSelect+` WHERE e.id = $1`, record.ID))
	if err != nil {
		return Execution{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Execution{}, false, fmt.Errorf("commit execution creation: %w", err)
	}
	return created, true, nil
}

func (r *PostgresRepository) Claim(ctx context.Context, owner string, now time.Time, leaseDuration time.Duration, maxClaims int) (Execution, bool, error) {
	if owner == "" || leaseDuration <= 0 || maxClaims < 1 {
		return Execution{}, false, fmt.Errorf("lease owner, positive duration, and max claims are required")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Execution{}, false, fmt.Errorf("begin execution claim: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return Execution{}, false, err
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE attempt_executions
		SET status = 'infra_failed',
			result = jsonb_build_object(
				'protocol_version', 1,
				'execution_id', id::text,
				'status', 'infra_failed',
				'stages', '[]'::jsonb,
				'duration_ms', GREATEST(0, floor(extract(epoch FROM ($1 - started_at)) * 1000)::bigint),
				'policy', jsonb_build_object('network', jsonb_build_object(
					'requested', spec #>> '{policy,network}', 'enforcement', 'policy_only')),
				'failure', jsonb_build_object('code', 'lease_attempts_exhausted', 'message', 'Execution exceeded its worker claim limit')
			),
			lease_owner = NULL, lease_expires_at = NULL, lease_heartbeat_at = NULL,
			finished_at = $1, updated_at = $1
		WHERE status = 'running' AND lease_expires_at <= $1 AND claim_count >= $2`, now, maxClaims)
	if err != nil {
		return Execution{}, false, fmt.Errorf("expire exhausted execution leases: %w", err)
	}
	leaseUntil := now.Add(leaseDuration)
	claimed, err := scanExecution(tx.QueryRowContext(ctx, `
		WITH candidate AS (
			SELECT id
			FROM attempt_executions
			WHERE (status = 'queued' OR (status = 'running' AND lease_expires_at <= $1))
				AND claim_count < $2
			ORDER BY created_at, id
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		UPDATE attempt_executions e
		SET status = 'running', claim_count = claim_count + 1,
			lease_owner = $3, lease_expires_at = $4, lease_heartbeat_at = $1,
			started_at = COALESCE(started_at, $1), updated_at = $1
		FROM candidate
		WHERE e.id = candidate.id
		RETURNING `+executionColumns, now, maxClaims, owner, leaseUntil))
	if errors.Is(err, ErrExecutionNotFound) {
		if err := tx.Commit(); err != nil {
			return Execution{}, false, fmt.Errorf("commit empty execution claim: %w", err)
		}
		return Execution{}, false, nil
	}
	if err != nil {
		return Execution{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return Execution{}, false, fmt.Errorf("commit execution claim: %w", err)
	}
	return claimed, true, nil
}

func (r *PostgresRepository) Renew(ctx context.Context, executionID, owner string, now time.Time, leaseDuration time.Duration) (bool, error) {
	if leaseDuration <= 0 {
		return false, fmt.Errorf("lease duration must be positive")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin execution lease renewal: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return false, err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE attempt_executions
		SET lease_expires_at = $4, lease_heartbeat_at = $3, updated_at = $3
		WHERE id = $1 AND status = 'running' AND lease_owner = $2 AND lease_expires_at > $3`,
		executionID, owner, now, now.Add(leaseDuration))
	if err != nil {
		return false, fmt.Errorf("renew execution lease: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("read execution lease renewal result: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit execution lease renewal: %w", err)
	}
	return rows == 1, nil
}

func (r *PostgresRepository) Complete(ctx context.Context, executionID, owner string, response ExecutionResponse, now time.Time) error {
	if response.ExecutionID != executionID {
		return fmt.Errorf("execution response identity mismatch")
	}
	if err := response.Validate(); err != nil {
		return err
	}
	encoded, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("encode execution response: %w", err)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin execution completion: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE attempt_executions
		SET status = $3, result = $4::jsonb,
			lease_owner = NULL, lease_expires_at = NULL, lease_heartbeat_at = NULL,
			finished_at = $5, updated_at = $5
		WHERE id = $1 AND status = 'running' AND lease_owner = $2 AND lease_expires_at > $5`,
		executionID, owner, string(response.Status), string(encoded), now)
	if err != nil {
		return fmt.Errorf("complete execution: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read execution completion result: %w", err)
	}
	if rows != 1 {
		return ErrLeaseLost
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit execution completion: %w", err)
	}
	return nil
}

const executionColumns = `e.id, e.attempt_id, e.submission_id, e.action, e.sequence,
	e.request_key, e.request_fingerprint, e.release_id, e.task_id, e.task_version,
	e.task_hash, e.workspace_revision, e.workspace_hash, e.spec, e.status, e.result,
	e.claim_count, e.lease_owner, e.lease_expires_at, e.lease_heartbeat_at,
	e.started_at, e.finished_at, e.created_at, e.updated_at`

const executionSelect = `SELECT ` + executionColumns + ` FROM attempt_executions e `

type rowScanner interface{ Scan(...any) error }

func scanExecution(row rowScanner) (Execution, error) {
	var value Execution
	var submissionID, leaseOwner sql.NullString
	var leaseExpiresAt, leaseHeartbeatAt, startedAt, finishedAt sql.NullTime
	var specJSON, responseJSON []byte
	err := row.Scan(
		&value.ID, &value.AttemptID, &submissionID, &value.Action, &value.Sequence,
		&value.RequestKey, &value.RequestFingerprint, &value.ReleaseID, &value.TaskID, &value.TaskVersion,
		&value.TaskHash, &value.WorkspaceRevision, &value.WorkspaceHash, &specJSON, &value.Status, &responseJSON,
		&value.ClaimCount, &leaseOwner, &leaseExpiresAt, &leaseHeartbeatAt,
		&startedAt, &finishedAt, &value.CreatedAt, &value.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Execution{}, ErrExecutionNotFound
	}
	if err != nil {
		return Execution{}, fmt.Errorf("scan learning execution: %w", err)
	}
	if err := json.Unmarshal(specJSON, &value.Spec); err != nil {
		return Execution{}, fmt.Errorf("decode execution spec: %w", err)
	}
	if err := value.Spec.Validate(); err != nil {
		return Execution{}, fmt.Errorf("validate stored execution spec: %w", err)
	}
	if len(responseJSON) > 0 {
		var response ExecutionResponse
		if err := json.Unmarshal(responseJSON, &response); err != nil {
			return Execution{}, fmt.Errorf("decode execution response: %w", err)
		}
		if err := response.Validate(); err != nil {
			return Execution{}, fmt.Errorf("validate stored execution response: %w", err)
		}
		value.Response = &response
	}
	if submissionID.Valid {
		value.SubmissionID = submissionID.String
	}
	if leaseOwner.Valid {
		value.LeaseOwner = leaseOwner.String
	}
	value.LeaseExpiresAt = nullableTime(leaseExpiresAt)
	value.LeaseHeartbeatAt = nullableTime(leaseHeartbeatAt)
	value.StartedAt = nullableTime(startedAt)
	value.FinishedAt = nullableTime(finishedAt)
	return value, nil
}

func nullableTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	copy := value.Time
	return &copy
}

func (r *PostgresRepository) setSearchPath(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `SET LOCAL search_path TO "`+r.schema+`"`)
	if err != nil {
		return fmt.Errorf("set execution repository search path: %w", err)
	}
	return nil
}
