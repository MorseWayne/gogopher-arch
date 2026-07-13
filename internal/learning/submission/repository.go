package submission

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"regexp"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
)

var repositorySchemaPattern = regexp.MustCompile("^[a-z_][a-z0-9_]*$")

type RepositoryOptions struct{ Schema string }

type PostgresRepository struct {
	db     *sql.DB
	schema string
}

type FreezeRecord struct {
	SubmissionID       string
	ExecutionID        string
	LearnerID          string
	AttemptID          string
	SubmissionKey      string
	RequestFingerprint string
	Workspace          map[string]string
	WorkspaceRevision  int64
	WorkspaceHash      string
	ReleaseID          string
	ActivityID         string
	ActivityVersion    int
	ActivityHash       string
	TaskID             string
	TaskVersion        int
	TaskHash           string
	RuleSetHash        string
	Spec               execution.ExecutionSpec
	CreatedAt          time.Time
}

type RetryRecord struct {
	ExecutionID        string
	LearnerID          string
	SubmissionID       string
	RequestKey         string
	RequestFingerprint string
	Spec               execution.ExecutionSpec
	CreatedAt          time.Time
}

func NewPostgresRepository(db *sql.DB, options RepositoryOptions) (*PostgresRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	schema := options.Schema
	if schema == "" {
		schema = "public"
	}
	if !repositorySchemaPattern.MatchString(schema) {
		return nil, fmt.Errorf("invalid PostgreSQL schema %q", schema)
	}
	return &PostgresRepository{db: db, schema: schema}, nil
}

func (r *PostgresRepository) Freeze(ctx context.Context, record FreezeRecord) (Result, error) {
	if err := validateFreezeRecord(record); err != nil {
		return Result{}, err
	}
	specJSON, err := json.Marshal(record.Spec)
	if err != nil {
		return Result{}, fmt.Errorf("encode submit execution spec: %w", err)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, fmt.Errorf("begin submission freeze: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return Result{}, err
	}

	var releaseID, activityID, activityHash, taskID, taskHash, status, workspaceHash string
	var activityVersion, taskVersion int
	var workspaceRevision int64
	var workspaceJSON []byte
	err = tx.QueryRowContext(ctx, `
		SELECT release_id, activity_id, activity_version, activity_hash,
			task_id, task_version, task_hash, status, workspace, workspace_revision, workspace_hash
		FROM learning_attempts
		WHERE id = $1 AND learner_id = $2
		FOR UPDATE`, record.AttemptID, record.LearnerID).Scan(
		&releaseID, &activityID, &activityVersion, &activityHash,
		&taskID, &taskVersion, &taskHash, &status, &workspaceJSON, &workspaceRevision, &workspaceHash)
	if errors.Is(err, sql.ErrNoRows) {
		return Result{}, ErrNotFound
	}
	if err != nil {
		return Result{}, fmt.Errorf("lock submission attempt: %w", err)
	}

	existing, err := scanSubmission(tx.QueryRowContext(ctx, submissionSelect+` WHERE s.attempt_id = $1`, record.AttemptID))
	if err == nil {
		switch {
		case existing.SubmissionKey != record.SubmissionKey:
			return Result{}, &AttemptAlreadySubmitted{SubmissionID: existing.ID}
		case existing.RequestFingerprint != record.RequestFingerprint:
			return Result{}, &IdempotencyConflict{SubmissionID: existing.ID}
		default:
			if err := tx.Commit(); err != nil {
				return Result{}, fmt.Errorf("commit idempotent submission read: %w", err)
			}
			return Result{
				Submission: existing, ExecutionID: existing.LatestExecutionID,
				ExecutionSequence: existing.LatestExecutionSeq,
			}, nil
		}
	}
	if !errors.Is(err, ErrNotFound) {
		return Result{}, err
	}
	if status != "active" {
		return Result{}, ErrAttemptInactive
	}
	if releaseID != record.ReleaseID || activityID != record.ActivityID || activityVersion != record.ActivityVersion ||
		activityHash != record.ActivityHash || taskID != record.TaskID || taskVersion != record.TaskVersion || taskHash != record.TaskHash {
		return Result{}, ErrAttemptInactive
	}
	if workspaceRevision != record.WorkspaceRevision || workspaceHash != record.WorkspaceHash {
		return Result{}, &WorkspaceConflict{Revision: workspaceRevision, Hash: workspaceHash}
	}
	var lockedWorkspace map[string]string
	if err := json.Unmarshal(workspaceJSON, &lockedWorkspace); err != nil {
		return Result{}, fmt.Errorf("decode locked submission workspace: %w", err)
	}
	if !maps.Equal(lockedWorkspace, record.Workspace) || attempt.WorkspaceHash(lockedWorkspace) != workspaceHash {
		return Result{}, fmt.Errorf("submission workspace snapshot does not match the locked attempt")
	}
	var cutoff int64
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(event_seq), 0) FROM assistance_events WHERE attempt_id = $1`, record.AttemptID).Scan(&cutoff); err != nil {
		return Result{}, fmt.Errorf("freeze assistance cutoff: %w", err)
	}
	frozenAt := record.CreatedAt
	_, err = tx.ExecContext(ctx, `
		INSERT INTO attempt_submissions (
			id, attempt_id, learner_id, submission_key, request_fingerprint,
			workspace, workspace_revision, workspace_hash, rule_set_hash,
			assistance_cutoff_seq, status, created_at
		) VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7,$8,$9,$10,'executing',$11)`,
		record.SubmissionID, record.AttemptID, record.LearnerID, record.SubmissionKey, record.RequestFingerprint,
		string(workspaceJSON), workspaceRevision, workspaceHash, record.RuleSetHash, cutoff, frozenAt)
	if err != nil {
		return Result{}, fmt.Errorf("insert frozen submission: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO attempt_executions (
			id, attempt_id, submission_id, action, sequence, request_key, request_fingerprint,
			release_id, task_id, task_version, task_hash, workspace_revision, workspace_hash,
			spec, status, created_at, updated_at
		) VALUES ($1,$2,$3,'submit',0,$4,$5,$6,$7,$8,$9,$10,$11,$12::jsonb,'queued',$13,$13)`,
		record.ExecutionID, record.AttemptID, record.SubmissionID, "initial:"+record.SubmissionID, record.RequestFingerprint,
		releaseID, taskID, taskVersion, taskHash, workspaceRevision, workspaceHash, string(specJSON), frozenAt)
	if err != nil {
		return Result{}, fmt.Errorf("insert initial submit execution: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE learning_attempts
		SET status = 'submitted',
			submitted_at = GREATEST(started_at, $3),
			updated_at = GREATEST(updated_at, $3)
		WHERE id = $1 AND learner_id = $2`, record.AttemptID, record.LearnerID, frozenAt)
	if err != nil {
		return Result{}, fmt.Errorf("freeze learning attempt: %w", err)
	}
	created, err := scanSubmission(tx.QueryRowContext(ctx, submissionSelect+` WHERE s.id = $1`, record.SubmissionID))
	if err != nil {
		return Result{}, err
	}
	if err := tx.Commit(); err != nil {
		return Result{}, fmt.Errorf("commit submission freeze: %w", err)
	}
	return Result{
		Submission: created, ExecutionID: created.LatestExecutionID,
		ExecutionSequence: created.LatestExecutionSeq, Created: true,
	}, nil
}

func (r *PostgresRepository) Get(ctx context.Context, learnerID, submissionID string) (Submission, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return Submission{}, fmt.Errorf("begin submission read: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return Submission{}, err
	}
	value, err := scanSubmission(tx.QueryRowContext(ctx, submissionSelect+` WHERE s.id = $1 AND a.learner_id = $2`, submissionID, learnerID))
	if err != nil {
		return Submission{}, err
	}
	if err := tx.Commit(); err != nil {
		return Submission{}, fmt.Errorf("commit submission read: %w", err)
	}
	return value, nil
}

func (r *PostgresRepository) Retry(ctx context.Context, record RetryRecord) (Result, error) {
	if record.Spec.ExecutionID != record.ExecutionID || record.Spec.Action != execution.ActionSubmit {
		return Result{}, fmt.Errorf("retry execution spec identity does not match queue record")
	}
	if err := record.Spec.Validate(); err != nil {
		return Result{}, err
	}
	specJSON, err := json.Marshal(record.Spec)
	if err != nil {
		return Result{}, fmt.Errorf("encode retry execution spec: %w", err)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Result{}, fmt.Errorf("begin submission retry: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return Result{}, err
	}
	var attemptID, attemptStatus string
	err = tx.QueryRowContext(ctx, `
		SELECT a.id, a.status
		FROM learning_attempts a
		JOIN attempt_submissions s ON s.attempt_id = a.id
		WHERE s.id = $1 AND a.learner_id = $2
		FOR UPDATE OF a`, record.SubmissionID, record.LearnerID).Scan(&attemptID, &attemptStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return Result{}, ErrNotFound
	}
	if err != nil {
		return Result{}, fmt.Errorf("lock retry attempt: %w", err)
	}
	frozen, err := scanSubmission(tx.QueryRowContext(ctx, submissionSelect+` WHERE s.id = $1 FOR UPDATE OF s`, record.SubmissionID))
	if err != nil {
		return Result{}, err
	}
	var existingID, existingFingerprint string
	var existingSequence int
	err = tx.QueryRowContext(ctx, `
		SELECT id, sequence, request_fingerprint
		FROM attempt_executions
		WHERE submission_id = $1 AND request_key = $2`,
		record.SubmissionID, record.RequestKey).Scan(&existingID, &existingSequence, &existingFingerprint)
	if err == nil {
		if existingFingerprint != record.RequestFingerprint {
			return Result{}, &IdempotencyConflict{SubmissionID: frozen.ID}
		}
		if err := tx.Commit(); err != nil {
			return Result{}, fmt.Errorf("commit idempotent submission retry read: %w", err)
		}
		return Result{
			Submission: frozen, ExecutionID: existingID, ExecutionSequence: existingSequence,
		}, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Result{}, fmt.Errorf("query idempotent submission retry: %w", err)
	}
	if frozen.Status != StatusInfraFailed || attemptStatus != "submit_infra_failed" ||
		frozen.LatestExecutionStatus != execution.ExecutionInfraFailed {
		return Result{}, ErrRetryUnavailable
	}
	nextSequence := frozen.LatestExecutionSeq + 1
	_, err = tx.ExecContext(ctx, `
		INSERT INTO attempt_executions (
			id, attempt_id, submission_id, action, sequence, request_key, request_fingerprint,
			release_id, task_id, task_version, task_hash, workspace_revision, workspace_hash,
			spec, status, created_at, updated_at
		) VALUES ($1,$2,$3,'submit',$4,$5,$6,$7,$8,$9,$10,$11,$12,$13::jsonb,'queued',$14,$14)`,
		record.ExecutionID, attemptID, record.SubmissionID, nextSequence, record.RequestKey, record.RequestFingerprint,
		frozen.ReleaseID, frozen.TaskID, frozen.TaskVersion, frozen.TaskHash,
		frozen.WorkspaceRevision, frozen.WorkspaceHash, string(specJSON), record.CreatedAt)
	if err != nil {
		return Result{}, fmt.Errorf("insert submit retry execution: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE attempt_submissions SET status = 'executing' WHERE id = $1`, record.SubmissionID); err != nil {
		return Result{}, fmt.Errorf("resume submission execution: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE learning_attempts
		SET status = 'submitted', updated_at = GREATEST(updated_at, $2)
		WHERE id = $1`, attemptID, record.CreatedAt); err != nil {
		return Result{}, fmt.Errorf("resume submitted attempt: %w", err)
	}
	retried, err := scanSubmission(tx.QueryRowContext(ctx, submissionSelect+` WHERE s.id = $1`, record.SubmissionID))
	if err != nil {
		return Result{}, err
	}
	if err := tx.Commit(); err != nil {
		return Result{}, fmt.Errorf("commit submission retry: %w", err)
	}
	return Result{
		Submission: retried, ExecutionID: record.ExecutionID,
		ExecutionSequence: nextSequence, Created: true,
	}, nil
}

func validateFreezeRecord(record FreezeRecord) error {
	if record.Spec.ExecutionID != record.ExecutionID || record.Spec.Action != execution.ActionSubmit {
		return fmt.Errorf("submit execution spec identity does not match freeze record")
	}
	if err := record.Spec.Validate(); err != nil {
		return err
	}
	if record.WorkspaceRevision < 0 || record.WorkspaceHash == "" || record.RuleSetHash == "" {
		return fmt.Errorf("submission workspace revision, hash, and rule set hash are required")
	}
	return nil
}

const submissionSelect = `SELECT
	s.id, s.attempt_id, s.submission_key, s.request_fingerprint, s.workspace,
	s.workspace_revision, s.workspace_hash, s.rule_set_hash, s.assistance_cutoff_seq,
	s.status, s.created_at, s.evaluated_at,
	a.learner_id, a.release_id, a.activity_id, a.activity_version, a.activity_hash,
	a.task_id, a.task_version, a.task_hash, a.mode,
	e.id, e.sequence, e.status
	FROM attempt_submissions s
	JOIN learning_attempts a ON a.id = s.attempt_id
	JOIN LATERAL (
		SELECT id, sequence, status
		FROM attempt_executions
		WHERE submission_id = s.id
		ORDER BY sequence DESC
		LIMIT 1
	) e ON true`

type rowScanner interface{ Scan(...any) error }

func scanSubmission(row rowScanner) (Submission, error) {
	var value Submission
	var workspaceJSON []byte
	var evaluatedAt sql.NullTime
	err := row.Scan(
		&value.ID, &value.AttemptID, &value.SubmissionKey, &value.RequestFingerprint, &workspaceJSON,
		&value.WorkspaceRevision, &value.WorkspaceHash, &value.RuleSetHash, &value.AssistanceCutoff,
		&value.Status, &value.CreatedAt, &evaluatedAt,
		&value.LearnerID, &value.ReleaseID, &value.ActivityID, &value.ActivityVersion, &value.ActivityHash,
		&value.TaskID, &value.TaskVersion, &value.TaskHash, &value.Mode,
		&value.LatestExecutionID, &value.LatestExecutionSeq, &value.LatestExecutionStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return Submission{}, ErrNotFound
	}
	if err != nil {
		return Submission{}, fmt.Errorf("scan learning submission: %w", err)
	}
	if err := json.Unmarshal(workspaceJSON, &value.Workspace); err != nil {
		return Submission{}, fmt.Errorf("decode frozen submission workspace: %w", err)
	}
	if evaluatedAt.Valid {
		value.EvaluatedAt = &evaluatedAt.Time
	}
	return value, nil
}

func (r *PostgresRepository) setSearchPath(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `SET LOCAL search_path TO "`+r.schema+`"`)
	if err != nil {
		return fmt.Errorf("set submission repository search path: %w", err)
	}
	return nil
}
