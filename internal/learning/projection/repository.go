package projection

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

const ProjectionVersion = 1

var projectionSchemaPattern = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

type RepositoryOptions struct {
	Schema string
	Random io.Reader
	Now    func() time.Time
}

type PostgresProjector struct {
	db       *sql.DB
	schema   string
	registry *definition.Registry
	random   io.Reader
	now      func() time.Time
}

func NewPostgresProjector(db *sql.DB, registry *definition.Registry, options RepositoryOptions) (*PostgresProjector, error) {
	if db == nil || registry == nil {
		return nil, fmt.Errorf("database and definition registry are required")
	}
	schema := options.Schema
	if schema == "" {
		schema = "public"
	}
	if !projectionSchemaPattern.MatchString(schema) {
		return nil, fmt.Errorf("invalid PostgreSQL schema %q", schema)
	}
	if options.Random == nil {
		options.Random = rand.Reader
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &PostgresProjector{db: db, schema: schema, registry: registry, random: options.Random, now: options.Now}, nil
}

func (p *PostgresProjector) Rebuild(ctx context.Context, input RebuildInput) (Snapshot, bool, error) {
	if input.LearnerID == "" || input.ReleaseID == "" || input.CapabilityID == "" || input.CapabilityVersion < 1 || input.AsOf.IsZero() {
		return Snapshot{}, false, fmt.Errorf("learner, release, capability version, and as_of are required")
	}
	policy, err := p.registry.CapabilityPolicy(input.ReleaseID, input.CapabilityID, input.CapabilityVersion)
	if err != nil {
		return Snapshot{}, false, err
	}
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return Snapshot{}, false, fmt.Errorf("begin capability projection: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SET LOCAL search_path TO "`+p.schema+`"`); err != nil {
		return Snapshot{}, false, fmt.Errorf("set projection search path: %w", err)
	}
	facts, err := p.readEvidence(ctx, tx, input, policy.ContentHash)
	if err != nil {
		return Snapshot{}, false, err
	}
	retentionBase, err := p.readRetentionBase(ctx, tx, input)
	if err != nil {
		return Snapshot{}, false, err
	}
	nextReviewAt, err := p.readNextReview(ctx, tx, input)
	if err != nil {
		return Snapshot{}, false, err
	}
	projected, err := Project(policy, Input{
		Evidence: facts, RetentionBase: retentionBase, NextReviewAt: nextReviewAt, AsOf: input.AsOf.UTC(),
	})
	if err != nil {
		return Snapshot{}, false, err
	}
	now := p.now().UTC()
	snapshot := Snapshot{
		LearnerID: input.LearnerID, CapabilityID: input.CapabilityID,
		CapabilityVersion: input.CapabilityVersion, CapabilityHash: policy.ContentHash,
		ProjectionVersion: ProjectionVersion, ProjectedAt: now, Result: projected,
	}
	existing, found, err := p.readStored(ctx, tx, input)
	if err != nil {
		return Snapshot{}, false, err
	}
	changed := !found || !sameStoredProjection(existing, snapshot)
	if changed {
		if err := p.upsert(ctx, tx, snapshot); err != nil {
			return Snapshot{}, false, err
		}
		if err := p.enqueueScheduler(ctx, tx, snapshot); err != nil {
			return Snapshot{}, false, err
		}
	} else {
		snapshot.ProjectedAt = existing.ProjectedAt
	}
	if err := tx.Commit(); err != nil {
		return Snapshot{}, false, fmt.Errorf("commit capability projection: %w", err)
	}
	return snapshot, changed, nil
}

func (p *PostgresProjector) readEvidence(ctx context.Context, tx *sql.Tx, input RebuildInput, capabilityHash string) ([]EvidenceFact, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT e.evidence_type, e.evidence_rule_id, e.result, e.independence,
			e.context_level, a.mode, e.occurred_at, e.capability_hash
		FROM evidence_records e
		JOIN learning_attempts a ON a.id = e.attempt_id AND a.learner_id = e.learner_id
		WHERE e.learner_id = $1 AND e.capability_id = $2 AND e.capability_version = $3
		ORDER BY e.occurred_at, e.id`, input.LearnerID, input.CapabilityID, input.CapabilityVersion)
	if err != nil {
		return nil, fmt.Errorf("query capability evidence: %w", err)
	}
	defer rows.Close()
	var facts []EvidenceFact
	for rows.Next() {
		var fact EvidenceFact
		var storedHash string
		if err := rows.Scan(
			&fact.EvidenceType, &fact.RuleID, &fact.Result, &fact.Independence,
			&fact.Context, &fact.ActivityMode, &fact.OccurredAt, &storedHash,
		); err != nil {
			return nil, fmt.Errorf("scan capability evidence: %w", err)
		}
		if storedHash != capabilityHash {
			return nil, fmt.Errorf("capability evidence hash does not match frozen definition")
		}
		facts = append(facts, fact)
	}
	return facts, rows.Err()
}

func (p *PostgresProjector) readRetentionBase(ctx context.Context, tx *sql.Tx, input RebuildInput) (RetentionBaseState, error) {
	var failed bool
	err := tx.QueryRowContext(ctx, `
		SELECT bool_or(e.result = 'failed')
		FROM review_items r
		JOIN evidence_records e ON e.evaluation_batch_id = r.evaluation_batch_id
			AND e.capability_id = r.capability_id AND e.capability_version = r.capability_version
		WHERE r.learner_id = $1 AND r.capability_id = $2 AND r.capability_version = $3
			AND r.status = 'completed'
		GROUP BY r.id, r.completed_at
		ORDER BY r.completed_at DESC, r.id DESC
		LIMIT 1`, input.LearnerID, input.CapabilityID, input.CapabilityVersion).Scan(&failed)
	if errors.Is(err, sql.ErrNoRows) {
		return RetentionFresh, nil
	}
	if err != nil {
		return "", fmt.Errorf("derive capability retention base: %w", err)
	}
	if failed {
		return RetentionRusty, nil
	}
	return RetentionFresh, nil
}

func (p *PostgresProjector) readNextReview(ctx context.Context, tx *sql.Tx, input RebuildInput) (*time.Time, error) {
	var dueAt sql.NullTime
	err := tx.QueryRowContext(ctx, `
		SELECT min(due_at) FROM review_items
		WHERE learner_id = $1 AND capability_id = $2 AND capability_version = $3
			AND status IN ('open','claimed')`, input.LearnerID, input.CapabilityID, input.CapabilityVersion).Scan(&dueAt)
	if err != nil {
		return nil, fmt.Errorf("derive next capability review: %w", err)
	}
	if !dueAt.Valid {
		return nil, nil
	}
	value := dueAt.Time.UTC()
	return &value, nil
}

func (p *PostgresProjector) readStored(ctx context.Context, tx *sql.Tx, input RebuildInput) (Snapshot, bool, error) {
	var value Snapshot
	var lastEvidence, lastIndependent, nextReview sql.NullTime
	err := tx.QueryRowContext(ctx, `
		SELECT learner_id, capability_id, capability_version, capability_hash,
			acquisition_state, independence_state, transfer_state, retention_base_state,
			last_evidence_at, last_independent_at, next_review_at, projection_version, projected_at
		FROM capability_snapshots
		WHERE learner_id = $1 AND capability_id = $2 AND capability_version = $3
		FOR UPDATE`, input.LearnerID, input.CapabilityID, input.CapabilityVersion).Scan(
		&value.LearnerID, &value.CapabilityID, &value.CapabilityVersion, &value.CapabilityHash,
		&value.AcquisitionState, &value.IndependenceState, &value.TransferState, &value.RetentionBase,
		&lastEvidence, &lastIndependent, &nextReview, &value.ProjectionVersion, &value.ProjectedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Snapshot{}, false, nil
	}
	if err != nil {
		return Snapshot{}, false, fmt.Errorf("read stored capability snapshot: %w", err)
	}
	value.LastEvidenceAt = nullTime(lastEvidence)
	value.LastIndependentAt = nullTime(lastIndependent)
	value.NextReviewAt = nullTime(nextReview)
	return value, true, nil
}

func (p *PostgresProjector) upsert(ctx context.Context, tx *sql.Tx, value Snapshot) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO capability_snapshots (
			learner_id, capability_id, capability_version, capability_hash,
			acquisition_state, independence_state, transfer_state, retention_base_state,
			last_evidence_at, last_independent_at, next_review_at, projection_version, projected_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (learner_id, capability_id, capability_version) DO UPDATE SET
			capability_hash = EXCLUDED.capability_hash,
			acquisition_state = EXCLUDED.acquisition_state,
			independence_state = EXCLUDED.independence_state,
			transfer_state = EXCLUDED.transfer_state,
			retention_base_state = EXCLUDED.retention_base_state,
			last_evidence_at = EXCLUDED.last_evidence_at,
			last_independent_at = EXCLUDED.last_independent_at,
			next_review_at = EXCLUDED.next_review_at,
			projection_version = EXCLUDED.projection_version,
			projected_at = EXCLUDED.projected_at`,
		value.LearnerID, value.CapabilityID, value.CapabilityVersion, value.CapabilityHash,
		value.AcquisitionState, value.IndependenceState, value.TransferState, value.RetentionBase,
		value.LastEvidenceAt, value.LastIndependentAt, value.NextReviewAt,
		value.ProjectionVersion, value.ProjectedAt)
	if err != nil {
		return fmt.Errorf("upsert capability snapshot: %w", err)
	}
	return nil
}

func (p *PostgresProjector) enqueueScheduler(ctx context.Context, tx *sql.Tx, value Snapshot) error {
	payload, err := json.Marshal(map[string]any{
		"projection_version": value.ProjectionVersion, "learner_id": value.LearnerID,
		"capability_id": value.CapabilityID, "capability_version": value.CapabilityVersion,
		"acquisition_state": value.AcquisitionState, "retention_base_state": value.RetentionBase,
	})
	if err != nil {
		return err
	}
	canonical, err := definition.CanonicalJSON(payload)
	if err != nil {
		return err
	}
	id, err := projectionUUID(p.random)
	if err != nil {
		return err
	}
	idempotencyKey := fmt.Sprintf("review-scheduler:v%d:%s:%s:%d:%s",
		value.ProjectionVersion, value.LearnerID, value.CapabilityID,
		value.CapabilityVersion, definition.SHA256Hex(canonical))
	_, err = tx.ExecContext(ctx, `
		INSERT INTO learning_outbox (
			id, topic, aggregate_type, aggregate_id, idempotency_key, payload,
			status, available_at, created_at
		) VALUES ($1,'review_scheduler.requested','capability_snapshot',$2,$3,$4::jsonb,'pending',$5,$5)
		ON CONFLICT (idempotency_key) DO NOTHING`,
		id, value.LearnerID, idempotencyKey, string(canonical), value.ProjectedAt)
	if err != nil {
		return fmt.Errorf("enqueue review scheduler: %w", err)
	}
	return nil
}

func sameStoredProjection(left, right Snapshot) bool {
	return left.CapabilityHash == right.CapabilityHash &&
		left.AcquisitionState == right.AcquisitionState && left.IndependenceState == right.IndependenceState &&
		left.TransferState == right.TransferState && left.RetentionBase == right.RetentionBase &&
		sameTime(left.LastEvidenceAt, right.LastEvidenceAt) &&
		sameTime(left.LastIndependentAt, right.LastIndependentAt) &&
		sameTime(left.NextReviewAt, right.NextReviewAt) && left.ProjectionVersion == right.ProjectionVersion
}

func sameTime(left, right *time.Time) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Equal(*right)
}

func nullTime(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time.UTC()
	return &result
}

func projectionUUID(source io.Reader) (string, error) {
	var value [16]byte
	if _, err := io.ReadFull(source, value[:]); err != nil {
		return "", fmt.Errorf("generate projection UUID: %w", err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}
