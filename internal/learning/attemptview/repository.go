package attemptview

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/MorseWayne/gogopher-arch/internal/learning/evaluation"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
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

func (r *PostgresRepository) Load(ctx context.Context, learnerID, attemptID string) (Related, error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return Related{}, fmt.Errorf("begin attempt detail read: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SET LOCAL search_path TO "`+r.schema+`"`); err != nil {
		return Related{}, fmt.Errorf("set attempt detail search path: %w", err)
	}
	result := Related{Executions: []execution.Execution{}, RuleResults: []execution.RuleResult{}, Evidence: []evaluation.Evidence{}}
	if err := r.readSubmission(ctx, tx, learnerID, attemptID, &result); err != nil {
		return Related{}, err
	}
	if err := r.readExecutions(ctx, tx, learnerID, attemptID, &result); err != nil {
		return Related{}, err
	}
	if err := r.readEvaluation(ctx, tx, learnerID, attemptID, &result); err != nil {
		return Related{}, err
	}
	if err := tx.Commit(); err != nil {
		return Related{}, fmt.Errorf("commit attempt detail read: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) readSubmission(ctx context.Context, tx *sql.Tx, learnerID, attemptID string, result *Related) error {
	var value Submission
	var evaluatedAt sql.NullTime
	err := tx.QueryRowContext(ctx, `
		SELECT s.id, s.workspace_revision, s.workspace_hash, s.rule_set_hash,
			s.assistance_cutoff_seq, s.status, s.created_at, s.evaluated_at,
			COALESCE(e.id::text, ''), COALESCE(e.sequence, 0), COALESCE(e.status, '')
		FROM attempt_submissions s
		JOIN learning_attempts a ON a.id = s.attempt_id AND a.learner_id = $1
		LEFT JOIN LATERAL (
			SELECT id, sequence, status FROM attempt_executions
			WHERE submission_id = s.id ORDER BY sequence DESC LIMIT 1
		) e ON true
		WHERE s.attempt_id = $2`, learnerID, attemptID).Scan(
		&value.ID, &value.WorkspaceRevision, &value.WorkspaceHash, &value.RuleSetHash,
		&value.AssistanceCutoff, &value.Status, &value.CreatedAt, &evaluatedAt,
		&value.LatestExecutionID, &value.LatestExecutionSeq, &value.LatestExecutionStatus,
	)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return fmt.Errorf("read attempt submission detail: %w", err)
	}
	if evaluatedAt.Valid {
		value.EvaluatedAt = &evaluatedAt.Time
	}
	result.Submission = &value
	return nil
}

func (r *PostgresRepository) readExecutions(ctx context.Context, tx *sql.Tx, learnerID, attemptID string, result *Related) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT e.id, e.attempt_id, COALESCE(e.submission_id::text, ''), e.action, e.sequence,
			e.workspace_revision, e.workspace_hash, e.status, e.result,
			e.started_at, e.finished_at, e.created_at, e.updated_at
		FROM attempt_executions e
		JOIN learning_attempts a ON a.id = e.attempt_id AND a.learner_id = $1
		WHERE e.attempt_id = $2
		ORDER BY e.created_at, e.id`, learnerID, attemptID)
	if err != nil {
		return fmt.Errorf("read attempt executions: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var value execution.Execution
		var responseJSON []byte
		var startedAt, finishedAt sql.NullTime
		if err := rows.Scan(
			&value.ID, &value.AttemptID, &value.SubmissionID, &value.Action, &value.Sequence,
			&value.WorkspaceRevision, &value.WorkspaceHash, &value.Status, &responseJSON,
			&startedAt, &finishedAt, &value.CreatedAt, &value.UpdatedAt,
		); err != nil {
			return fmt.Errorf("scan attempt execution: %w", err)
		}
		if len(responseJSON) > 0 {
			var response execution.ExecutionResponse
			if err := json.Unmarshal(responseJSON, &response); err != nil {
				return fmt.Errorf("decode attempt execution response: %w", err)
			}
			value.Response = &response
		}
		if startedAt.Valid {
			value.StartedAt = &startedAt.Time
		}
		if finishedAt.Valid {
			value.FinishedAt = &finishedAt.Time
		}
		result.Executions = append(result.Executions, value)
	}
	return rows.Err()
}

func (r *PostgresRepository) readEvaluation(ctx context.Context, tx *sql.Tx, learnerID, attemptID string, result *Related) error {
	batchRows, err := tx.QueryContext(ctx, `
		SELECT b.rule_results
		FROM evaluation_batches b
		JOIN attempt_submissions s ON s.id = b.submission_id
		JOIN learning_attempts a ON a.id = s.attempt_id AND a.learner_id = $1
		WHERE s.attempt_id = $2
		ORDER BY b.created_at, b.id`, learnerID, attemptID)
	if err != nil {
		return fmt.Errorf("read attempt rule results: %w", err)
	}
	for batchRows.Next() {
		var encoded []byte
		if err := batchRows.Scan(&encoded); err != nil {
			batchRows.Close()
			return fmt.Errorf("scan attempt rule results: %w", err)
		}
		var values []execution.RuleResult
		if err := json.Unmarshal(encoded, &values); err != nil {
			batchRows.Close()
			return fmt.Errorf("decode attempt rule results: %w", err)
		}
		result.RuleResults = append(result.RuleResults, values...)
	}
	if err := batchRows.Close(); err != nil {
		return err
	}
	if err := batchRows.Err(); err != nil {
		return err
	}

	evidenceRows, err := tx.QueryContext(ctx, `
		SELECT e.id, e.evaluation_batch_id, e.learner_id, e.capability_id, e.capability_version,
			e.attempt_id, e.activity_id, e.evidence_rule_id, e.evidence_type, e.result,
			e.independence, e.context_level, e.evaluator, e.rule_version, e.reason,
			e.occurred_at, e.created_at
		FROM evidence_records e
		WHERE e.attempt_id = $2 AND e.learner_id = $1
		ORDER BY e.occurred_at, e.id`, learnerID, attemptID)
	if err != nil {
		return fmt.Errorf("read attempt evidence: %w", err)
	}
	defer evidenceRows.Close()
	for evidenceRows.Next() {
		var value evaluation.Evidence
		if err := evidenceRows.Scan(
			&value.ID, &value.EvaluationBatchID, &value.LearnerID, &value.CapabilityID,
			&value.CapabilityVersion, &value.AttemptID, &value.ActivityID, &value.EvidenceRuleID,
			&value.EvidenceType, &value.Result, &value.Independence, &value.ContextLevel,
			&value.Evaluator, &value.RuleVersion, &value.Reason, &value.OccurredAt, &value.CreatedAt,
		); err != nil {
			return fmt.Errorf("scan attempt evidence: %w", err)
		}
		result.Evidence = append(result.Evidence, value)
	}
	return evidenceRows.Err()
}
