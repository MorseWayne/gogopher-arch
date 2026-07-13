package projection

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

const firstReviewPolicyVersion = 1

type SchedulerOptions struct {
	Schema string
	Random io.Reader
	Now    func() time.Time
}

type ReviewScheduler struct {
	db       *sql.DB
	schema   string
	registry *definition.Registry
	random   io.Reader
	now      func() time.Time
}

type assessmentEvidenceBatch struct {
	ID              string
	ActivityID      string
	ActivityVersion int
	ActivityHash    string
	SourceEvidence  string
	OccurredAt      time.Time
	Facts           []EvidenceFact
}

func NewReviewScheduler(db *sql.DB, registry *definition.Registry, options SchedulerOptions) (*ReviewScheduler, error) {
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
	return &ReviewScheduler{db: db, schema: schema, registry: registry, random: options.Random, now: options.Now}, nil
}

func (s *ReviewScheduler) ProcessRequest(ctx context.Context, request Request, _ time.Time) error {
	var payload ReviewSchedulerRequestPayload
	if err := json.Unmarshal(request.Payload, &payload); err != nil {
		return fmt.Errorf("decode review scheduler payload: %w", err)
	}
	if payload.EventVersion != ReviewSchedulerEventVersion || payload.ProjectionVersion != ProjectionVersion {
		return fmt.Errorf("unsupported review scheduler event %d projection %d", payload.EventVersion, payload.ProjectionVersion)
	}
	if payload.LearnerID == "" || payload.ReleaseID == "" || payload.CapabilityID == "" || payload.CapabilityVersion < 1 {
		return fmt.Errorf("review scheduler learner, release, and capability are required")
	}
	if !payload.AcquisitionState.Valid() || !payload.IndependenceState.Valid() ||
		!payload.TransferState.Valid() || !payload.RetentionBase.Valid() {
		return fmt.Errorf("review scheduler projection state is invalid")
	}
	if acquisitionRanks[payload.AcquisitionState] < acquisitionRanks[AcquisitionVerified] ||
		payload.IndependenceState != IndependenceIndependent ||
		transferRanks[payload.TransferState] < transferRanks[TransferSameContext] {
		return nil
	}
	policy, err := s.registry.CapabilityPolicy(payload.ReleaseID, payload.CapabilityID, payload.CapabilityVersion)
	if err != nil {
		return err
	}
	if policy.ContentHash != payload.CapabilityHash {
		return fmt.Errorf("review scheduler capability hash does not match frozen definition")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin first review scheduling: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SET LOCAL search_path TO "`+s.schema+`"`); err != nil {
		return fmt.Errorf("set review scheduler search path: %w", err)
	}
	var locked bool
	if err := tx.QueryRowContext(ctx, `
		SELECT true FROM capability_snapshots
		WHERE learner_id=$1 AND capability_id=$2 AND capability_version=$3
		FOR UPDATE`, payload.LearnerID, payload.CapabilityID, payload.CapabilityVersion).Scan(&locked); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("review scheduler capability snapshot was not found")
		}
		return fmt.Errorf("lock capability snapshot for review: %w", err)
	}
	source, found, err := s.firstQualifyingAssessment(ctx, tx, payload, policy)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("independent assessment evidence does not satisfy review policy")
	}
	sourceActivity, err := s.registry.ActivityView(payload.ReleaseID, source.ActivityID, source.ActivityVersion)
	if err != nil {
		return err
	}
	if sourceActivity.ContentHash != source.ActivityHash || sourceActivity.Mode != "assessment" {
		return fmt.Errorf("review source activity does not match frozen assessment")
	}
	reviewActivity, err := s.registry.ReviewActivity(payload.ReleaseID, sourceActivity.CapabilityRefs)
	if err != nil {
		return err
	}
	createdAt := s.now().UTC()
	dueAt := source.OccurredAt.AddDate(0, 0, policy.ReviewPolicy.FirstReviewAfterDays)
	groupKey := fmt.Sprintf("assessment:%s:review:%s@%d:policy:%d",
		source.ID, reviewActivity.ID, reviewActivity.Version, firstReviewPolicyVersion)
	created, reviewItemID, err := s.insertFirstReview(ctx, tx, payload, source, reviewActivity, groupKey, dueAt, createdAt)
	if err != nil {
		return err
	}
	if created {
		if err := s.enqueueTargetProjection(ctx, tx, payload, reviewItemID, createdAt); err != nil {
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit first review scheduling: %w", err)
	}
	return nil
}

func (s *ReviewScheduler) firstQualifyingAssessment(ctx context.Context, tx *sql.Tx, payload ReviewSchedulerRequestPayload, policy definition.CapabilityPolicyView) (assessmentEvidenceBatch, bool, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT e.id,e.evaluation_batch_id,e.evidence_type,e.evidence_rule_id,e.result,
			e.independence,e.context_level,e.occurred_at,
			a.activity_id,a.activity_version,a.activity_hash
		FROM evidence_records e
		JOIN learning_attempts a ON a.id=e.attempt_id AND a.learner_id=e.learner_id
		WHERE e.learner_id=$1 AND e.capability_id=$2 AND e.capability_version=$3
			AND a.release_id=$4 AND a.mode='assessment'
		ORDER BY e.occurred_at,e.evaluation_batch_id,e.id`,
		payload.LearnerID, payload.CapabilityID, payload.CapabilityVersion, payload.ReleaseID)
	if err != nil {
		return assessmentEvidenceBatch{}, false, fmt.Errorf("query assessment evidence for first review: %w", err)
	}
	defer rows.Close()
	byID := make(map[string]*assessmentEvidenceBatch)
	for rows.Next() {
		var evidenceID string
		var batchID string
		var fact EvidenceFact
		var activityID, activityHash string
		var activityVersion int
		if err := rows.Scan(
			&evidenceID, &batchID, &fact.EvidenceType, &fact.RuleID, &fact.Result,
			&fact.Independence, &fact.Context, &fact.OccurredAt,
			&activityID, &activityVersion, &activityHash,
		); err != nil {
			return assessmentEvidenceBatch{}, false, fmt.Errorf("scan assessment evidence for first review: %w", err)
		}
		fact.ActivityMode = "assessment"
		batch := byID[batchID]
		if batch == nil {
			batch = &assessmentEvidenceBatch{
				ID: batchID, ActivityID: activityID, ActivityVersion: activityVersion,
				ActivityHash: activityHash, OccurredAt: fact.OccurredAt.UTC(),
			}
			byID[batchID] = batch
		}
		batch.Facts = append(batch.Facts, fact)
		if fact.OccurredAt.After(batch.OccurredAt) {
			batch.OccurredAt = fact.OccurredAt.UTC()
		}
		if fact.Result == "passed" && fact.Independence == IndependenceIndependent &&
			transferRanks[fact.Context] >= transferRanks[TransferSameContext] {
			batch.SourceEvidence = evidenceID
		}
	}
	if err := rows.Err(); err != nil {
		return assessmentEvidenceBatch{}, false, fmt.Errorf("iterate assessment evidence for first review: %w", err)
	}
	batches := make([]assessmentEvidenceBatch, 0, len(byID))
	for _, batch := range byID {
		batches = append(batches, *batch)
	}
	sort.Slice(batches, func(i, j int) bool {
		if batches[i].OccurredAt.Equal(batches[j].OccurredAt) {
			return batches[i].ID < batches[j].ID
		}
		return batches[i].OccurredAt.Before(batches[j].OccurredAt)
	})
	for _, batch := range batches {
		independent := make([]EvidenceFact, 0, len(batch.Facts))
		for _, fact := range batch.Facts {
			if fact.Independence == IndependenceIndependent {
				independent = append(independent, fact)
			}
		}
		if batch.SourceEvidence != "" && requirementsSatisfied(policy.RequiredEvidence, independent, TransferSameContext) {
			return batch, true, nil
		}
	}
	return assessmentEvidenceBatch{}, false, nil
}

func (s *ReviewScheduler) insertFirstReview(ctx context.Context, tx *sql.Tx, payload ReviewSchedulerRequestPayload, source assessmentEvidenceBatch, activity definition.ActivityView, groupKey string, dueAt, createdAt time.Time) (bool, string, error) {
	var existingID string
	err := tx.QueryRowContext(ctx, `
		SELECT id FROM review_items
		WHERE learner_id=$1 AND capability_id=$2 AND capability_version=$3
			AND source_evidence_id=$4 AND policy_version=$5`,
		payload.LearnerID, payload.CapabilityID, payload.CapabilityVersion,
		source.SourceEvidence, firstReviewPolicyVersion).Scan(&existingID)
	if err == nil {
		return false, existingID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, "", fmt.Errorf("query existing first review: %w", err)
	}
	var activeID, activeStatus string
	err = tx.QueryRowContext(ctx, `
		SELECT id,status FROM review_items
		WHERE learner_id=$1 AND capability_id=$2 AND capability_version=$3
			AND policy_version=$4 AND status IN ('open','claimed')
		FOR UPDATE`, payload.LearnerID, payload.CapabilityID,
		payload.CapabilityVersion, firstReviewPolicyVersion).Scan(&activeID, &activeStatus)
	if err == nil {
		if activeStatus == "claimed" {
			return false, activeID, nil
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE review_items SET status='replaced',replaced_at=$2,updated_at=$2
			WHERE id=$1 AND status='open'`, activeID, createdAt); err != nil {
			return false, "", fmt.Errorf("replace stale first review: %w", err)
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return false, "", fmt.Errorf("query active first review: %w", err)
	}
	id, err := projectionUUID(s.random)
	if err != nil {
		return false, "", err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO review_items (
			id,learner_id,capability_id,capability_version,source_evidence_id,
			release_id,activity_id,activity_version,activity_hash,review_group_key,
			due_at,priority,reason,status,policy_version,created_at,updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,100,'first_review','open',$12,$13,$13)`,
		id, payload.LearnerID, payload.CapabilityID, payload.CapabilityVersion,
		source.SourceEvidence, payload.ReleaseID, activity.ID, activity.Version,
		activity.ContentHash, groupKey, dueAt.UTC(), firstReviewPolicyVersion, createdAt)
	if err != nil {
		return false, "", fmt.Errorf("insert first review item: %w", err)
	}
	return true, id, nil
}

func (s *ReviewScheduler) enqueueTargetProjection(ctx context.Context, tx *sql.Tx, payload ReviewSchedulerRequestPayload, reviewItemID string, createdAt time.Time) error {
	target, err := json.Marshal(ProjectionRequestPayload{
		EventVersion: ProjectionTargetEventVersion, LearnerID: payload.LearnerID,
		ReleaseID: payload.ReleaseID, CapabilityID: payload.CapabilityID,
		CapabilityVersion: payload.CapabilityVersion,
	})
	if err != nil {
		return err
	}
	id, err := projectionUUID(s.random)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO learning_outbox (
			id,topic,aggregate_type,aggregate_id,idempotency_key,payload,status,available_at,created_at
		) VALUES ($1,$2,'review_item',$3,$4,$5::jsonb,'pending',$6,$6)
		ON CONFLICT (idempotency_key) DO NOTHING`,
		id, projectionRequestTopic, payload.LearnerID,
		"capability-projection:review-item:"+reviewItemID, string(target), createdAt)
	if err != nil {
		return fmt.Errorf("enqueue review item projection: %w", err)
	}
	return nil
}
