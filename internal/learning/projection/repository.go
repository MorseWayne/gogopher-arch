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
	"sort"
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

type ProjectionChange struct {
	Input  RebuildInput `json:"input"`
	Change string       `json:"change"`
	Before *Snapshot    `json:"before"`
	After  Snapshot     `json:"after"`
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

func (p *PostgresProjector) RebuildRequest(ctx context.Context, request Request, asOf time.Time) error {
	var payload ProjectionRequestPayload
	if err := json.Unmarshal(request.Payload, &payload); err != nil {
		return fmt.Errorf("decode projection request payload: %w", err)
	}
	if asOf.IsZero() || payload.LearnerID == "" {
		return fmt.Errorf("projection request learner and as_of are required")
	}
	var targets []RebuildInput
	switch payload.EventVersion {
	case ProjectionRequestEventVersion:
		if payload.EvaluationBatchID == "" {
			return fmt.Errorf("projection request evaluation batch is required")
		}
		var err error
		targets, err = p.requestTargets(ctx, payload, asOf.UTC())
		if err != nil {
			return err
		}
	case ProjectionTargetEventVersion:
		if payload.ReleaseID == "" || payload.CapabilityID == "" || payload.CapabilityVersion < 1 {
			return fmt.Errorf("targeted projection request release and capability are required")
		}
		targets = []RebuildInput{{
			LearnerID: payload.LearnerID, ReleaseID: payload.ReleaseID,
			CapabilityID: payload.CapabilityID, CapabilityVersion: payload.CapabilityVersion,
			AsOf: asOf.UTC(),
		}}
	default:
		return fmt.Errorf("unsupported projection request event version %d", payload.EventVersion)
	}
	for _, target := range targets {
		if _, _, err := p.Rebuild(ctx, target); err != nil {
			return fmt.Errorf("rebuild requested capability %s@%d: %w", target.CapabilityID, target.CapabilityVersion, err)
		}
	}
	return nil
}

func (p *PostgresProjector) ProcessRequest(ctx context.Context, request Request, asOf time.Time) error {
	return p.RebuildRequest(ctx, request, asOf)
}

func (p *PostgresProjector) requestTargets(ctx context.Context, payload ProjectionRequestPayload, asOf time.Time) ([]RebuildInput, error) {
	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("begin projection request target query: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SET LOCAL search_path TO "`+p.schema+`"`); err != nil {
		return nil, fmt.Errorf("set projection request search path: %w", err)
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT a.release_id, e.capability_id, e.capability_version
		FROM evaluation_batches b
		JOIN attempt_submissions s ON s.id = b.submission_id
		JOIN learning_attempts a ON a.id = s.attempt_id AND a.learner_id = s.learner_id
		LEFT JOIN evidence_records e ON e.evaluation_batch_id = b.id AND e.learner_id = a.learner_id
		WHERE b.id = $1 AND a.learner_id = $2
		ORDER BY e.capability_id NULLS LAST, e.capability_version NULLS LAST`,
		payload.EvaluationBatchID, payload.LearnerID)
	if err != nil {
		return nil, fmt.Errorf("query projection request targets: %w", err)
	}
	defer rows.Close()
	var targets []RebuildInput
	foundBatch := false
	seen := make(map[string]struct{})
	for rows.Next() {
		foundBatch = true
		var releaseID string
		var capabilityID sql.NullString
		var capabilityVersion sql.NullInt64
		if err := rows.Scan(&releaseID, &capabilityID, &capabilityVersion); err != nil {
			return nil, fmt.Errorf("scan projection request target: %w", err)
		}
		if !capabilityID.Valid || !capabilityVersion.Valid {
			continue
		}
		key := fmt.Sprintf("%s@%d", capabilityID.String, capabilityVersion.Int64)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		targets = append(targets, RebuildInput{
			LearnerID: payload.LearnerID, ReleaseID: releaseID,
			CapabilityID: capabilityID.String, CapabilityVersion: int(capabilityVersion.Int64), AsOf: asOf,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projection request targets: %w", err)
	}
	if !foundBatch {
		return nil, fmt.Errorf("projection request evaluation batch was not found for learner")
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit projection request target query: %w", err)
	}
	return targets, nil
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
		if err := p.enqueueScheduler(ctx, tx, input.ReleaseID, snapshot); err != nil {
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

func (p *PostgresProjector) Preview(ctx context.Context, input RebuildInput) (ProjectionChange, error) {
	if input.LearnerID == "" || input.ReleaseID == "" || input.CapabilityID == "" || input.CapabilityVersion < 1 || input.AsOf.IsZero() {
		return ProjectionChange{}, fmt.Errorf("learner, release, capability version, and as_of are required")
	}
	policy, err := p.registry.CapabilityPolicy(input.ReleaseID, input.CapabilityID, input.CapabilityVersion)
	if err != nil {
		return ProjectionChange{}, err
	}
	tx, err := p.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return ProjectionChange{}, fmt.Errorf("begin capability projection preview: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SET LOCAL search_path TO "`+p.schema+`"`); err != nil {
		return ProjectionChange{}, fmt.Errorf("set projection preview search path: %w", err)
	}
	facts, err := p.readEvidence(ctx, tx, input, policy.ContentHash)
	if err != nil {
		return ProjectionChange{}, err
	}
	retentionBase, err := p.readRetentionBase(ctx, tx, input)
	if err != nil {
		return ProjectionChange{}, err
	}
	nextReviewAt, err := p.readNextReview(ctx, tx, input)
	if err != nil {
		return ProjectionChange{}, err
	}
	projected, err := Project(policy, Input{
		Evidence: facts, RetentionBase: retentionBase, NextReviewAt: nextReviewAt, AsOf: input.AsOf.UTC(),
	})
	if err != nil {
		return ProjectionChange{}, err
	}
	after := Snapshot{
		LearnerID: input.LearnerID, CapabilityID: input.CapabilityID,
		CapabilityVersion: input.CapabilityVersion, CapabilityHash: policy.ContentHash,
		ProjectionVersion: ProjectionVersion, ProjectedAt: p.now().UTC(), Result: projected,
	}
	existing, found, err := p.readStoredValue(ctx, tx, input, false)
	if err != nil {
		return ProjectionChange{}, err
	}
	change := "create"
	var before *Snapshot
	if found {
		existing.RetentionState = derivedRetention(existing.RetentionBase, existing.NextReviewAt, input.AsOf.UTC())
		before = &existing
		change = "update"
		if sameStoredProjection(existing, after) {
			change = "none"
			after.ProjectedAt = existing.ProjectedAt
		}
	}
	if err := tx.Commit(); err != nil {
		return ProjectionChange{}, fmt.Errorf("commit capability projection preview: %w", err)
	}
	return ProjectionChange{Input: input, Change: change, Before: before, After: after}, nil
}

func (p *PostgresProjector) RebuildTargets(ctx context.Context, learnerID, capabilityID string, asOf time.Time) ([]RebuildInput, error) {
	if asOf.IsZero() {
		return nil, fmt.Errorf("as_of is required")
	}
	type targetDefinition struct {
		releaseID string
		stored    definition.Definition
	}
	available := make(map[string]targetDefinition)
	currentLatest := make(map[string]targetDefinition)
	currentReleaseID := p.registry.CurrentReleaseID()
	for _, releaseID := range p.registry.ReleaseIDs() {
		definitions, err := p.registry.Definitions(releaseID)
		if err != nil {
			return nil, err
		}
		for _, stored := range definitions {
			if stored.Kind != definition.KindCapability || (capabilityID != "" && stored.ID != capabilityID) {
				continue
			}
			key := fmt.Sprintf("%s@%d", stored.ID, stored.Version)
			if _, exists := available[key]; !exists || releaseID == currentReleaseID {
				available[key] = targetDefinition{releaseID: releaseID, stored: stored}
			}
			if releaseID == currentReleaseID {
				if current, exists := currentLatest[stored.ID]; !exists || stored.Version > current.stored.Version {
					currentLatest[stored.ID] = targetDefinition{releaseID: releaseID, stored: stored}
				}
			}
		}
	}
	if capabilityID != "" && len(available) == 0 {
		return nil, fmt.Errorf("current Capability %s: %w", capabilityID, definition.ErrDefinitionNotFound)
	}
	if learnerID != "" {
		var exists bool
		if err := p.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM "`+p.schema+`".learners WHERE id=$1)`, learnerID).Scan(&exists); err != nil {
			return nil, fmt.Errorf("query rebuild learner: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("rebuild learner was not found")
		}
	}
	rows, err := p.db.QueryContext(ctx, `
		SELECT learner_id::text,capability_id,capability_version FROM "`+p.schema+`".evidence_records
		UNION
		SELECT learner_id::text,capability_id,capability_version FROM "`+p.schema+`".capability_snapshots
		UNION
		SELECT learner_id::text,capability_id,capability_version FROM "`+p.schema+`".review_items
		ORDER BY 1,2,3`)
	if err != nil {
		return nil, fmt.Errorf("query full projection rebuild targets: %w", err)
	}
	defer rows.Close()
	targetsByKey := make(map[string]RebuildInput)
	if learnerID != "" {
		for _, current := range currentLatest {
			key := fmt.Sprintf("%s:%s@%d", learnerID, current.stored.ID, current.stored.Version)
			targetsByKey[key] = RebuildInput{
				LearnerID: learnerID, ReleaseID: current.releaseID, CapabilityID: current.stored.ID,
				CapabilityVersion: current.stored.Version, AsOf: asOf.UTC(),
			}
		}
	}
	for rows.Next() {
		var target RebuildInput
		if err := rows.Scan(&target.LearnerID, &target.CapabilityID, &target.CapabilityVersion); err != nil {
			return nil, fmt.Errorf("scan full projection rebuild target: %w", err)
		}
		if learnerID != "" && target.LearnerID != learnerID {
			continue
		}
		if capabilityID != "" && target.CapabilityID != capabilityID {
			continue
		}
		frozen, exists := available[fmt.Sprintf("%s@%d", target.CapabilityID, target.CapabilityVersion)]
		if !exists {
			return nil, fmt.Errorf("definition for projected Capability %s@%d is unavailable", target.CapabilityID, target.CapabilityVersion)
		}
		target.ReleaseID, target.AsOf = frozen.releaseID, asOf.UTC()
		key := fmt.Sprintf("%s:%s@%d", target.LearnerID, target.CapabilityID, target.CapabilityVersion)
		targetsByKey[key] = target
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate full projection rebuild targets: %w", err)
	}
	targets := make([]RebuildInput, 0, len(targetsByKey))
	for _, target := range targetsByKey {
		targets = append(targets, target)
	}
	sortRebuildTargets(targets)
	return targets, nil
}

func sortRebuildTargets(values []RebuildInput) {
	sort.Slice(values, func(i, j int) bool {
		if values[i].LearnerID != values[j].LearnerID {
			return values[i].LearnerID < values[j].LearnerID
		}
		if values[i].CapabilityID != values[j].CapabilityID {
			return values[i].CapabilityID < values[j].CapabilityID
		}
		return values[i].CapabilityVersion < values[j].CapabilityVersion
	})
}

func (p *PostgresProjector) readEvidence(ctx context.Context, tx *sql.Tx, input RebuildInput, capabilityHash string) ([]EvidenceFact, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT e.evidence_type, e.evidence_rule_id, e.result, e.independence,
			e.context_level, a.mode,
			COALESCE((
				SELECT r.outcome = 'passed'
				FROM attempt_review_items link
				JOIN review_items r ON r.id = link.review_item_id
				WHERE link.attempt_id = a.id
					AND r.capability_id = e.capability_id
					AND r.capability_version = e.capability_version
					AND r.evaluation_batch_id = e.evaluation_batch_id
				LIMIT 1
			), false),
			e.occurred_at, e.capability_hash
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
			&fact.Context, &fact.ActivityMode, &fact.QualifyingReview, &fact.OccurredAt, &storedHash,
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
	var outcome string
	err := tx.QueryRowContext(ctx, `
		SELECT r.outcome
		FROM review_items r
		JOIN learning_attempts a ON a.id = r.claimed_attempt_id AND a.learner_id = r.learner_id
		JOIN attempt_review_items link ON link.review_item_id = r.id AND link.attempt_id = a.id
		WHERE r.learner_id = $1 AND r.capability_id = $2 AND r.capability_version = $3
			AND r.status = 'completed' AND a.mode = 'review'
			AND r.outcome IN ('passed','failed')
		ORDER BY r.completed_at DESC, r.id DESC
		LIMIT 1`, input.LearnerID, input.CapabilityID, input.CapabilityVersion).Scan(&outcome)
	if errors.Is(err, sql.ErrNoRows) {
		return RetentionFresh, nil
	}
	if err != nil {
		return "", fmt.Errorf("derive capability retention base: %w", err)
	}
	if outcome == "failed" {
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
	return p.readStoredValue(ctx, tx, input, true)
}

func (p *PostgresProjector) readStoredValue(ctx context.Context, tx *sql.Tx, input RebuildInput, lock bool) (Snapshot, bool, error) {
	var value Snapshot
	var lastEvidence, lastIndependent, nextReview sql.NullTime
	query := `
		SELECT learner_id, capability_id, capability_version, capability_hash,
			acquisition_state, independence_state, transfer_state, retention_base_state,
			last_evidence_at, last_independent_at, next_review_at, projection_version, projected_at
		FROM capability_snapshots
		WHERE learner_id = $1 AND capability_id = $2 AND capability_version = $3`
	if lock {
		query += ` FOR UPDATE`
	}
	err := tx.QueryRowContext(ctx, query, input.LearnerID, input.CapabilityID, input.CapabilityVersion).Scan(
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

func (p *PostgresProjector) enqueueScheduler(ctx context.Context, tx *sql.Tx, releaseID string, value Snapshot) error {
	payload, err := json.Marshal(map[string]any{
		"event_version": ReviewSchedulerEventVersion, "projection_version": value.ProjectionVersion,
		"learner_id": value.LearnerID, "release_id": releaseID,
		"capability_id": value.CapabilityID, "capability_version": value.CapabilityVersion,
		"capability_hash": value.CapabilityHash, "acquisition_state": value.AcquisitionState,
		"independence_state": value.IndependenceState, "transfer_state": value.TransferState,
		"retention_base_state": value.RetentionBase,
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
