package evaluation

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

var evaluationSchemaPattern = regexp.MustCompile("^[a-z_][a-z0-9_]*$")

type RepositoryOptions struct{ Schema string }

type PostgresRepository struct {
	db     *sql.DB
	schema string
}

func (r *PostgresRepository) ClaimRequest(ctx context.Context, owner string, now time.Time, lease time.Duration) (Request, bool, error) {
	if owner == "" || lease <= 0 {
		return Request{}, false, fmt.Errorf("evaluation lease owner and duration are required")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Request{}, false, fmt.Errorf("begin evaluation request claim: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return Request{}, false, err
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE learning_outbox
		SET status = 'pending', lease_owner = NULL, lease_expires_at = NULL, available_at = $1
		WHERE topic = 'submission.evaluate' AND status = 'processing' AND lease_expires_at <= $1`, now)
	if err != nil {
		return Request{}, false, fmt.Errorf("recover expired evaluation request: %w", err)
	}
	var request Request
	err = tx.QueryRowContext(ctx, `
		WITH candidate AS (
			SELECT id FROM learning_outbox
			WHERE topic = 'submission.evaluate' AND status = 'pending' AND available_at <= $1
			ORDER BY available_at, created_at, id
			FOR UPDATE SKIP LOCKED LIMIT 1
		)
		UPDATE learning_outbox o
		SET status = 'processing', attempt_count = attempt_count + 1,
			lease_owner = $2, lease_expires_at = $3
		FROM candidate
		WHERE o.id = candidate.id
		RETURNING o.id, o.payload->>'learner_id', o.payload->>'submission_id', o.payload->>'execution_id'`,
		now, owner, now.Add(lease)).Scan(
		&request.ID, &request.LearnerID, &request.SubmissionID, &request.ExecutionID)
	if errors.Is(err, sql.ErrNoRows) {
		if err := tx.Commit(); err != nil {
			return Request{}, false, err
		}
		return Request{}, false, nil
	}
	if err != nil {
		return Request{}, false, fmt.Errorf("claim evaluation request: %w", err)
	}
	if request.LearnerID == "" || request.SubmissionID == "" || request.ExecutionID == "" {
		return Request{}, false, fmt.Errorf("evaluation request payload is incomplete")
	}
	if err := tx.Commit(); err != nil {
		return Request{}, false, fmt.Errorf("commit evaluation request claim: %w", err)
	}
	return request, true, nil
}

func (r *PostgresRepository) CompleteRequest(ctx context.Context, requestID, owner string, now time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin evaluation request completion: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return err
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE learning_outbox
		SET status = 'completed', lease_owner = NULL, lease_expires_at = NULL, completed_at = $3
		WHERE id = $1 AND status = 'processing' AND lease_owner = $2 AND lease_expires_at > $3`,
		requestID, owner, now)
	if err != nil {
		return fmt.Errorf("complete evaluation request: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil || rows != 1 {
		return fmt.Errorf("evaluation request lease was lost")
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit evaluation request completion: %w", err)
	}
	return nil
}

func NewPostgresRepository(db *sql.DB, options RepositoryOptions) (*PostgresRepository, error) {
	if db == nil {
		return nil, fmt.Errorf("database is required")
	}
	schema := options.Schema
	if schema == "" {
		schema = "public"
	}
	if !evaluationSchemaPattern.MatchString(schema) {
		return nil, fmt.Errorf("invalid PostgreSQL schema %q", schema)
	}
	return &PostgresRepository{db: db, schema: schema}, nil
}

func (r *PostgresRepository) Persist(ctx context.Context, record PersistRecord) (Batch, bool, error) {
	if err := validatePersistArtifacts(record); err != nil {
		return Batch{}, false, err
	}
	ruleResultsJSON, err := json.Marshal(record.Batch.RuleResults)
	if err != nil {
		return Batch{}, false, fmt.Errorf("encode rule results: %w", err)
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Batch{}, false, fmt.Errorf("begin evaluation batch: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return Batch{}, false, err
	}
	var submissionStatus, ruleSetHash, attemptID, learnerID, latestExecutionID, latestExecutionStatus string
	err = tx.QueryRowContext(ctx, `
		SELECT s.status, s.rule_set_hash, s.attempt_id, s.learner_id, e.id, e.status
		FROM attempt_submissions s
		JOIN LATERAL (
			SELECT id, status FROM attempt_executions
			WHERE submission_id = s.id ORDER BY sequence DESC LIMIT 1
		) e ON true
		WHERE s.id = $1 AND s.learner_id = $2
		FOR UPDATE OF s`, record.Batch.SubmissionID, record.LearnerID).Scan(
		&submissionStatus, &ruleSetHash, &attemptID, &learnerID, &latestExecutionID, &latestExecutionStatus)
	if errors.Is(err, sql.ErrNoRows) {
		return Batch{}, false, fmt.Errorf("evaluation submission not found")
	}
	if err != nil {
		return Batch{}, false, fmt.Errorf("lock evaluation submission: %w", err)
	}
	existing, err := r.readBatch(ctx, tx, record.Batch.SubmissionID, record.Batch.RuleSetHash)
	if err == nil {
		if err := tx.Commit(); err != nil {
			return Batch{}, false, fmt.Errorf("commit idempotent evaluation read: %w", err)
		}
		return existing, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Batch{}, false, err
	}
	if submissionStatus != "executing" || ruleSetHash != record.Batch.RuleSetHash ||
		attemptID != record.AttemptID || learnerID != record.LearnerID ||
		latestExecutionID != record.Batch.ExecutionID ||
		(latestExecutionStatus != "succeeded" && latestExecutionStatus != "user_failed") {
		return Batch{}, false, fmt.Errorf("submission is not ready for deterministic evaluation")
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO evaluation_batches (id, submission_id, execution_id, rule_set_hash, rule_results, created_at)
		VALUES ($1,$2,$3,$4,$5::jsonb,$6)`,
		record.Batch.ID, record.Batch.SubmissionID, record.Batch.ExecutionID,
		record.Batch.RuleSetHash, string(ruleResultsJSON), record.Batch.CreatedAt)
	if err != nil {
		return Batch{}, false, fmt.Errorf("insert evaluation batch: %w", err)
	}
	for _, artifact := range record.Batch.Artifacts {
		_, err := tx.ExecContext(ctx, `
			INSERT INTO artifacts (
				id, attempt_id, submission_id, kind, content, content_bytes, content_hash, created_at
			) VALUES ($1,$2,$3,$4,$5::jsonb,$6,$7,$8)`,
			artifact.ID, artifact.AttemptID, artifact.SubmissionID, artifact.Kind,
			string(artifact.Content), artifact.ContentBytes, artifact.ContentHash, artifact.CreatedAt)
		if err != nil {
			return Batch{}, false, fmt.Errorf("insert %s artifact: %w", artifact.Kind, err)
		}
	}
	for _, evidence := range record.Batch.Evidence {
		var artifactID any
		if evidence.ArtifactID != "" {
			artifactID = evidence.ArtifactID
		}
		_, err := tx.ExecContext(ctx, `
			INSERT INTO evidence_records (
				id, evaluation_batch_id, learner_id, capability_id, capability_version, capability_hash,
				attempt_id, activity_id, artifact_id, evidence_rule_id, evidence_type, result,
				independence, context_level, evaluator, rule_version, reason, occurred_at, created_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19)`,
			evidence.ID, record.Batch.ID, evidence.LearnerID, evidence.CapabilityID,
			evidence.CapabilityVersion, evidence.CapabilityHash, evidence.AttemptID, evidence.ActivityID,
			artifactID, evidence.EvidenceRuleID, evidence.EvidenceType, string(evidence.Result),
			string(evidence.Independence), evidence.ContextLevel, evidence.Evaluator,
			evidence.RuleVersion, evidence.Reason, evidence.OccurredAt, evidence.CreatedAt)
		if err != nil {
			return Batch{}, false, fmt.Errorf("insert evidence record: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE attempt_submissions
		SET status = 'evaluated', evaluated_at = $2
		WHERE id = $1`, record.Batch.SubmissionID, record.OccurredAt); err != nil {
		return Batch{}, false, fmt.Errorf("complete evaluated submission: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE learning_attempts
		SET status = 'completed', completed_at = $3, updated_at = GREATEST(updated_at, $3)
		WHERE id = $1 AND learner_id = $2`, record.AttemptID, record.LearnerID, record.OccurredAt); err != nil {
		return Batch{}, false, fmt.Errorf("complete evaluated attempt: %w", err)
	}
	projectionPayload, _ := json.Marshal(map[string]string{
		"evaluation_batch_id": record.Batch.ID,
		"learner_id":          record.LearnerID,
	})
	_, err = tx.ExecContext(ctx, `
		INSERT INTO learning_outbox (
			id, topic, aggregate_type, aggregate_id, idempotency_key, payload,
			status, available_at, created_at
		) VALUES ($1,'capability_projection.requested','evaluation_batch',$1,$2,$3::jsonb,'pending',$4,$4)
		ON CONFLICT (idempotency_key) DO NOTHING`,
		record.Batch.ID, "capability-projection:"+record.Batch.ID, string(projectionPayload), record.Batch.CreatedAt)
	if err != nil {
		return Batch{}, false, fmt.Errorf("enqueue capability projection: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Batch{}, false, fmt.Errorf("commit evaluation batch: %w", err)
	}
	return record.Batch, true, nil
}

func validatePersistArtifacts(record PersistRecord) error {
	expected := map[string]bool{
		"workspace": false, "diff": false, "explanation": false, "test_report": false,
	}
	if len(record.Batch.Artifacts) != len(expected) {
		return fmt.Errorf("evaluation batch must contain exactly four artifacts")
	}
	artifactIDs := make(map[string]struct{}, len(record.Batch.Artifacts))
	for _, artifact := range record.Batch.Artifacts {
		if _, exists := expected[artifact.Kind]; !exists || expected[artifact.Kind] {
			return fmt.Errorf("evaluation batch contains duplicate or unsupported artifact kind %q", artifact.Kind)
		}
		if artifact.AttemptID != record.AttemptID || artifact.SubmissionID != record.Batch.SubmissionID {
			return fmt.Errorf("%s artifact does not belong to the evaluated submission", artifact.Kind)
		}
		canonical, err := definition.CanonicalJSON(artifact.Content)
		if err != nil {
			return fmt.Errorf("canonicalize %s artifact: %w", artifact.Kind, err)
		}
		if len(canonical) > maxArtifactContentBytes {
			return fmt.Errorf("%s artifact contains %d bytes, limit is %d", artifact.Kind, len(canonical), maxArtifactContentBytes)
		}
		if !bytes.Equal(canonical, artifact.Content) || artifact.ContentBytes != len(canonical) ||
			artifact.ContentHash != definition.SHA256Hex(canonical) {
			return fmt.Errorf("%s artifact content does not match its metadata", artifact.Kind)
		}
		if _, exists := artifactIDs[artifact.ID]; exists || artifact.ID == "" {
			return fmt.Errorf("evaluation batch contains duplicate or empty artifact ID")
		}
		artifactIDs[artifact.ID] = struct{}{}
		expected[artifact.Kind] = true
	}
	for _, evidence := range record.Batch.Evidence {
		if _, exists := artifactIDs[evidence.ArtifactID]; !exists {
			return fmt.Errorf("evidence %q does not reference a batch artifact", evidence.ID)
		}
	}
	return nil
}

func (r *PostgresRepository) readBatch(ctx context.Context, tx *sql.Tx, submissionID, ruleSetHash string) (Batch, error) {
	var batch Batch
	var ruleResultsJSON []byte
	err := tx.QueryRowContext(ctx, `
		SELECT id, submission_id, execution_id, rule_set_hash, rule_results, created_at
		FROM evaluation_batches
		WHERE submission_id = $1 AND rule_set_hash = $2`, submissionID, ruleSetHash).Scan(
		&batch.ID, &batch.SubmissionID, &batch.ExecutionID, &batch.RuleSetHash, &ruleResultsJSON, &batch.CreatedAt)
	if err != nil {
		return Batch{}, err
	}
	if err := json.Unmarshal(ruleResultsJSON, &batch.RuleResults); err != nil {
		return Batch{}, fmt.Errorf("decode stored rule results: %w", err)
	}
	artifactRows, err := tx.QueryContext(ctx, `
		SELECT id, attempt_id, submission_id, kind, content, content_bytes, content_hash, created_at
		FROM artifacts
		WHERE submission_id = $1
		ORDER BY kind, id`, batch.SubmissionID)
	if err != nil {
		return Batch{}, fmt.Errorf("query stored artifacts: %w", err)
	}
	for artifactRows.Next() {
		var artifact Artifact
		var storedContent []byte
		if err := artifactRows.Scan(
			&artifact.ID, &artifact.AttemptID, &artifact.SubmissionID, &artifact.Kind,
			&storedContent, &artifact.ContentBytes, &artifact.ContentHash, &artifact.CreatedAt,
		); err != nil {
			artifactRows.Close()
			return Batch{}, fmt.Errorf("scan stored artifact: %w", err)
		}
		content, err := definition.CanonicalJSON(storedContent)
		if err != nil {
			artifactRows.Close()
			return Batch{}, fmt.Errorf("canonicalize stored %s artifact: %w", artifact.Kind, err)
		}
		if len(content) != artifact.ContentBytes || definition.SHA256Hex(content) != artifact.ContentHash {
			artifactRows.Close()
			return Batch{}, fmt.Errorf("stored %s artifact content does not match its metadata", artifact.Kind)
		}
		artifact.Content = append(json.RawMessage(nil), content...)
		batch.Artifacts = append(batch.Artifacts, artifact)
	}
	if err := artifactRows.Close(); err != nil {
		return Batch{}, fmt.Errorf("close stored artifacts: %w", err)
	}
	if err := artifactRows.Err(); err != nil {
		return Batch{}, fmt.Errorf("read stored artifacts: %w", err)
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT id, evaluation_batch_id, learner_id, capability_id, capability_version, capability_hash,
			attempt_id, activity_id, artifact_id, evidence_rule_id, evidence_type, result, independence,
			context_level, evaluator, rule_version, reason, occurred_at, created_at
		FROM evidence_records
		WHERE evaluation_batch_id = $1
		ORDER BY capability_id, capability_version, evidence_rule_id, evidence_type`, batch.ID)
	if err != nil {
		return Batch{}, fmt.Errorf("query stored evidence: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var evidence Evidence
		var artifactID sql.NullString
		if err := rows.Scan(
			&evidence.ID, &evidence.EvaluationBatchID, &evidence.LearnerID,
			&evidence.CapabilityID, &evidence.CapabilityVersion, &evidence.CapabilityHash,
			&evidence.AttemptID, &evidence.ActivityID, &artifactID, &evidence.EvidenceRuleID,
			&evidence.EvidenceType, &evidence.Result, &evidence.Independence,
			&evidence.ContextLevel, &evidence.Evaluator, &evidence.RuleVersion,
			&evidence.Reason, &evidence.OccurredAt, &evidence.CreatedAt); err != nil {
			return Batch{}, fmt.Errorf("scan stored evidence: %w", err)
		}
		if artifactID.Valid {
			evidence.ArtifactID = artifactID.String
		}
		batch.Evidence = append(batch.Evidence, evidence)
	}
	return batch, rows.Err()
}

func (r *PostgresRepository) setSearchPath(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `SET LOCAL search_path TO "`+r.schema+`"`)
	if err != nil {
		return fmt.Errorf("set evaluation repository search path: %w", err)
	}
	return nil
}
