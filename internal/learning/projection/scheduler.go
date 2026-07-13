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
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
)

const firstReviewPolicyVersion = 1

type ReviewTransitionObserver interface {
	ReviewItemsTransitioned(string, int)
}

type SchedulerOptions struct {
	Schema   string
	Random   io.Reader
	Now      func() time.Time
	Observer ReviewTransitionObserver
}

type ReviewScheduler struct {
	db       *sql.DB
	schema   string
	registry *definition.Registry
	random   io.Reader
	now      func() time.Time
	observer ReviewTransitionObserver
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
	return &ReviewScheduler{
		db: db, schema: schema, registry: registry, random: options.Random,
		now: options.Now, observer: options.Observer,
	}, nil
}

func (s *ReviewScheduler) ProcessRequest(ctx context.Context, request Request, _ time.Time) error {
	var header struct {
		EventVersion int `json:"event_version"`
	}
	if err := json.Unmarshal(request.Payload, &header); err != nil {
		return fmt.Errorf("decode review scheduler payload: %w", err)
	}
	switch header.EventVersion {
	case ReviewSchedulerEventVersion:
		var payload ReviewSchedulerRequestPayload
		if err := json.Unmarshal(request.Payload, &payload); err != nil {
			return fmt.Errorf("decode first review scheduler payload: %w", err)
		}
		return s.processFirstReview(ctx, payload)
	case ReviewOutcomeEventVersion:
		var payload ReviewOutcomeRequestPayload
		if err := json.Unmarshal(request.Payload, &payload); err != nil {
			return fmt.Errorf("decode review outcome payload: %w", err)
		}
		return s.processReviewOutcome(ctx, payload)
	default:
		return fmt.Errorf("unsupported review scheduler event %d", header.EventVersion)
	}
}

func (s *ReviewScheduler) processFirstReview(ctx context.Context, payload ReviewSchedulerRequestPayload) error {
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
	created, replaced, reviewItemID, err := s.insertFirstReview(ctx, tx, payload, source, reviewActivity, groupKey, dueAt, createdAt)
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
	if s.observer != nil {
		if replaced {
			s.observer.ReviewItemsTransitioned("replaced", 1)
		}
		if created {
			s.observer.ReviewItemsTransitioned("created", 1)
		}
	}
	return nil
}

type reviewEvaluation struct {
	BatchID         string
	AttemptID       string
	ReleaseID       string
	ActivityID      string
	ActivityVersion int
	ActivityHash    string
	TaskID          string
	TaskVersion     int
	TaskHash        string
	RuleSetHash     string
	Mode            string
	AttemptStatus   string
	FinishedAt      time.Time
	RuleResults     []execution.RuleResult
}

type outcomeReviewItem struct {
	ID                string
	CapabilityID      string
	CapabilityVersion int
	ActivityID        string
	ActivityVersion   int
	ActivityHash      string
	DueAt             time.Time
	Priority          int
	Status            string
	PolicyVersion     int
	ClaimedAttemptID  sql.NullString
	EvaluationBatchID sql.NullString
	Outcome           sql.NullString
}

func (s *ReviewScheduler) processReviewOutcome(ctx context.Context, payload ReviewOutcomeRequestPayload) error {
	if payload.EventVersion != ReviewOutcomeEventVersion || payload.EvaluationBatchID == "" || payload.LearnerID == "" {
		return fmt.Errorf("review outcome batch and learner are required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin review outcome: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SET LOCAL search_path TO "`+s.schema+`"`); err != nil {
		return fmt.Errorf("set review outcome search path: %w", err)
	}
	evaluation, err := s.loadReviewEvaluation(ctx, tx, payload)
	if err != nil {
		return err
	}
	activity, err := s.registry.ActivityView(evaluation.ReleaseID, evaluation.ActivityID, evaluation.ActivityVersion)
	if err != nil {
		return fmt.Errorf("resolve review outcome activity: %w", err)
	}
	if activity.ContentHash != evaluation.ActivityHash || activity.RuleSetHash != evaluation.RuleSetHash || activity.Mode != evaluation.Mode {
		return fmt.Errorf("review outcome activity does not match frozen Attempt")
	}
	task, err := s.registry.ExecutionTask(evaluation.ReleaseID, evaluation.TaskID, evaluation.TaskVersion)
	if err != nil {
		return fmt.Errorf("resolve review outcome task: %w", err)
	}
	if task.BundleHash != evaluation.TaskHash || activity.TaskRef.ID != task.ID || activity.TaskRef.Version != task.Version {
		return fmt.Errorf("review outcome task does not match frozen Attempt")
	}
	results, err := validatedRuleResults(task, evaluation.RuleResults)
	if err != nil {
		return err
	}
	items, err := s.lockOutcomeItems(ctx, tx, payload.LearnerID, evaluation)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return fmt.Errorf("review outcome Attempt has no linked ReviewItem")
	}
	createdAt := s.now().UTC()
	transitioned := 0
	for _, item := range items {
		if item.Status == "completed" {
			if !item.EvaluationBatchID.Valid || item.EvaluationBatchID.String != evaluation.BatchID || !item.Outcome.Valid {
				return fmt.Errorf("ReviewItem %s was completed by a different EvaluationBatch", item.ID)
			}
			var successorID string
			if err := tx.QueryRowContext(ctx, `SELECT id FROM review_items WHERE predecessor_review_item_id=$1`, item.ID).Scan(&successorID); err != nil {
				return fmt.Errorf("load replayed ReviewItem successor %s: %w", item.ID, err)
			}
			continue
		}
		if item.Status != "claimed" || !item.ClaimedAttemptID.Valid || item.ClaimedAttemptID.String != evaluation.AttemptID {
			return fmt.Errorf("ReviewItem %s is not claimed by evaluated Attempt", item.ID)
		}
		if item.ActivityID != evaluation.ActivityID || item.ActivityVersion != evaluation.ActivityVersion || item.ActivityHash != evaluation.ActivityHash {
			return fmt.Errorf("ReviewItem %s does not match evaluated Activity", item.ID)
		}
		if item.PolicyVersion != firstReviewPolicyVersion {
			return fmt.Errorf("ReviewItem %s uses unsupported policy version %d", item.ID, item.PolicyVersion)
		}
		outcome, err := capabilityOutcome(task, results, item.CapabilityID, item.CapabilityVersion)
		if err != nil {
			return err
		}
		policy, err := s.registry.CapabilityPolicy(evaluation.ReleaseID, item.CapabilityID, item.CapabilityVersion)
		if err != nil {
			return err
		}
		sourceEvidenceID, err := outcomeSourceEvidence(ctx, tx, evaluation.BatchID, item, outcome)
		if err != nil {
			return err
		}
		successorActivity, reason, dueAt, err := s.reviewSuccessor(evaluation, activity, item, policy, outcome)
		if err != nil {
			return err
		}
		result, err := tx.ExecContext(ctx, `
			UPDATE review_items
			SET status='completed',evaluation_batch_id=$2,outcome=$3,
				completed_at=$4,updated_at=GREATEST(updated_at,$4)
			WHERE id=$1 AND status='claimed' AND claimed_attempt_id=$5`,
			item.ID, evaluation.BatchID, outcome, evaluation.FinishedAt, evaluation.AttemptID)
		if err != nil {
			return fmt.Errorf("complete ReviewItem %s: %w", item.ID, err)
		}
		if affected, err := result.RowsAffected(); err != nil || affected != 1 {
			return fmt.Errorf("complete ReviewItem %s changed %d rows", item.ID, affected)
		}
		successorID, err := projectionUUID(s.random)
		if err != nil {
			return err
		}
		groupKey := fmt.Sprintf("review-outcome:%s:%s:%s@%d", evaluation.BatchID, reason, successorActivity.ID, successorActivity.Version)
		var sourceEvidence any
		if sourceEvidenceID != "" {
			sourceEvidence = sourceEvidenceID
		}
		_, err = tx.ExecContext(ctx, `
			INSERT INTO review_items (
				id,learner_id,capability_id,capability_version,source_evidence_id,
				predecessor_review_item_id,release_id,activity_id,activity_version,activity_hash,
				review_group_key,due_at,priority,reason,status,policy_version,created_at,updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,'open',$15,$16,$16)`,
			successorID, payload.LearnerID, item.CapabilityID, item.CapabilityVersion,
			sourceEvidence, item.ID, evaluation.ReleaseID, successorActivity.ID,
			successorActivity.Version, successorActivity.ContentHash, groupKey, dueAt.UTC(),
			item.Priority, reason, item.PolicyVersion, createdAt)
		if err != nil {
			return fmt.Errorf("insert %s successor for ReviewItem %s: %w", reason, item.ID, err)
		}
		target := ReviewSchedulerRequestPayload{
			LearnerID: payload.LearnerID, ReleaseID: evaluation.ReleaseID,
			CapabilityID: item.CapabilityID, CapabilityVersion: item.CapabilityVersion,
		}
		if err := s.enqueueTargetProjection(ctx, tx, target, successorID, createdAt); err != nil {
			return err
		}
		transitioned++
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit review outcome: %w", err)
	}
	if s.observer != nil && transitioned > 0 {
		s.observer.ReviewItemsTransitioned("completed", transitioned)
		s.observer.ReviewItemsTransitioned("created", transitioned)
	}
	return nil
}

func (s *ReviewScheduler) loadReviewEvaluation(ctx context.Context, tx *sql.Tx, payload ReviewOutcomeRequestPayload) (reviewEvaluation, error) {
	var value reviewEvaluation
	var ruleResults []byte
	err := tx.QueryRowContext(ctx, `
		SELECT b.id,a.id,a.release_id,a.activity_id,a.activity_version,a.activity_hash,
			a.task_id,a.task_version,a.task_hash,b.rule_set_hash,a.mode,a.status,e.finished_at,b.rule_results
		FROM evaluation_batches b
		JOIN attempt_submissions sub ON sub.id=b.submission_id
		JOIN learning_attempts a ON a.id=sub.attempt_id AND a.learner_id=sub.learner_id
		JOIN attempt_executions e ON e.id=b.execution_id AND e.submission_id=b.submission_id
		WHERE b.id=$1 AND a.learner_id=$2`, payload.EvaluationBatchID, payload.LearnerID).Scan(
		&value.BatchID, &value.AttemptID, &value.ReleaseID, &value.ActivityID,
		&value.ActivityVersion, &value.ActivityHash, &value.TaskID, &value.TaskVersion,
		&value.TaskHash, &value.RuleSetHash, &value.Mode, &value.AttemptStatus, &value.FinishedAt, &ruleResults)
	if errors.Is(err, sql.ErrNoRows) {
		return reviewEvaluation{}, fmt.Errorf("review outcome EvaluationBatch was not found")
	}
	if err != nil {
		return reviewEvaluation{}, fmt.Errorf("load review outcome EvaluationBatch: %w", err)
	}
	if value.AttemptStatus != "completed" || (value.Mode != "review" && value.Mode != "practice" && value.Mode != "guided") {
		return reviewEvaluation{}, fmt.Errorf("review outcome Attempt is not a completed review or remediation")
	}
	if err := json.Unmarshal(ruleResults, &value.RuleResults); err != nil {
		return reviewEvaluation{}, fmt.Errorf("decode review outcome RuleResult: %w", err)
	}
	return value, nil
}

func (s *ReviewScheduler) lockOutcomeItems(ctx context.Context, tx *sql.Tx, learnerID string, evaluation reviewEvaluation) ([]outcomeReviewItem, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT r.id,r.capability_id,r.capability_version,r.activity_id,r.activity_version,
			r.activity_hash,r.due_at,r.priority,r.status,r.policy_version,
			r.claimed_attempt_id,r.evaluation_batch_id,r.outcome
		FROM attempt_review_items link
		JOIN review_items r ON r.id=link.review_item_id
		WHERE link.attempt_id=$1 AND r.learner_id=$2
		ORDER BY r.id FOR UPDATE OF r`, evaluation.AttemptID, learnerID)
	if err != nil {
		return nil, fmt.Errorf("lock evaluated ReviewItems: %w", err)
	}
	defer rows.Close()
	var items []outcomeReviewItem
	for rows.Next() {
		var item outcomeReviewItem
		if err := rows.Scan(
			&item.ID, &item.CapabilityID, &item.CapabilityVersion, &item.ActivityID,
			&item.ActivityVersion, &item.ActivityHash, &item.DueAt, &item.Priority,
			&item.Status, &item.PolicyVersion, &item.ClaimedAttemptID,
			&item.EvaluationBatchID, &item.Outcome,
		); err != nil {
			return nil, fmt.Errorf("scan evaluated ReviewItem: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate evaluated ReviewItems: %w", err)
	}
	return items, nil
}

func validatedRuleResults(task definition.ExecutionTask, values []execution.RuleResult) (map[string]execution.RuleStatus, error) {
	expected := make(map[string]struct{}, len(task.AssessmentRules))
	for _, rule := range task.AssessmentRules {
		expected[rule.RuleID] = struct{}{}
	}
	results := make(map[string]execution.RuleStatus, len(values))
	for _, result := range values {
		if _, ok := expected[result.RuleID]; !ok {
			return nil, fmt.Errorf("review outcome contains unknown RuleResult %q", result.RuleID)
		}
		if _, duplicate := results[result.RuleID]; duplicate {
			return nil, fmt.Errorf("review outcome contains duplicate RuleResult %q", result.RuleID)
		}
		if result.Status != execution.RulePassed && result.Status != execution.RuleFailed && result.Status != execution.RuleNotEvaluated {
			return nil, fmt.Errorf("review outcome contains invalid status %q", result.Status)
		}
		results[result.RuleID] = result.Status
	}
	if len(results) != len(expected) {
		return nil, fmt.Errorf("review outcome RuleResult set is incomplete")
	}
	return results, nil
}

func capabilityOutcome(task definition.ExecutionTask, results map[string]execution.RuleStatus, capabilityID string, capabilityVersion int) (string, error) {
	matched, passed := 0, 0
	hasFailed := false
	for _, rule := range task.AssessmentRules {
		applies := false
		for _, ref := range rule.CapabilityRefs {
			if ref.ID == capabilityID && ref.Version == capabilityVersion {
				applies = true
				break
			}
		}
		if !applies {
			continue
		}
		matched++
		switch results[rule.RuleID] {
		case execution.RulePassed:
			passed++
		case execution.RuleFailed:
			hasFailed = true
		}
	}
	if matched == 0 {
		return "", fmt.Errorf("review Activity has no rule for Capability %s@%d", capabilityID, capabilityVersion)
	}
	if hasFailed {
		return "failed", nil
	}
	if passed == matched {
		return "passed", nil
	}
	return "incomplete", nil
}

func outcomeSourceEvidence(ctx context.Context, tx *sql.Tx, batchID string, item outcomeReviewItem, outcome string) (string, error) {
	if outcome == "incomplete" {
		return "", nil
	}
	var evidenceID string
	err := tx.QueryRowContext(ctx, `
		SELECT id FROM evidence_records
		WHERE evaluation_batch_id=$1 AND capability_id=$2 AND capability_version=$3 AND result=$4
		ORDER BY id LIMIT 1`, batchID, item.CapabilityID, item.CapabilityVersion, outcome).Scan(&evidenceID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%s review outcome for Capability %s@%d has no Evidence", outcome, item.CapabilityID, item.CapabilityVersion)
	}
	if err != nil {
		return "", fmt.Errorf("load review outcome Evidence: %w", err)
	}
	return evidenceID, nil
}

func (s *ReviewScheduler) reviewSuccessor(evaluation reviewEvaluation, currentActivity definition.ActivityView, item outcomeReviewItem, policy definition.CapabilityPolicyView, outcome string) (definition.ActivityView, string, time.Time, error) {
	ref := definition.VersionedDefinitionRef{ID: item.CapabilityID, Version: item.CapabilityVersion}
	if outcome == "incomplete" {
		return currentActivity, "review_incomplete", item.DueAt.UTC(), nil
	}
	if evaluation.Mode == "review" {
		if outcome == "passed" {
			return currentActivity, "maintenance", evaluation.FinishedAt.AddDate(0, 0, policy.ReviewPolicy.SuccessIntervalDays), nil
		}
		activity, err := s.registry.RemediationActivity(evaluation.ReleaseID, ref)
		if err != nil {
			return definition.ActivityView{}, "", time.Time{}, err
		}
		return activity, "remediation", evaluation.FinishedAt.AddDate(0, 0, policy.ReviewPolicy.FailureRemediationAfterDays), nil
	}
	if outcome == "failed" {
		return currentActivity, "remediation", evaluation.FinishedAt.AddDate(0, 0, policy.ReviewPolicy.FailureRemediationAfterDays), nil
	}
	activity, err := s.registry.VariantReviewActivity(evaluation.ReleaseID, ref)
	if err != nil {
		return definition.ActivityView{}, "", time.Time{}, err
	}
	return activity, "remediation_review", evaluation.FinishedAt.AddDate(0, 0, policy.ReviewPolicy.FirstReviewAfterDays), nil
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

func (s *ReviewScheduler) insertFirstReview(ctx context.Context, tx *sql.Tx, payload ReviewSchedulerRequestPayload, source assessmentEvidenceBatch, activity definition.ActivityView, groupKey string, dueAt, createdAt time.Time) (bool, bool, string, error) {
	var existingID string
	err := tx.QueryRowContext(ctx, `
		SELECT id FROM review_items
		WHERE learner_id=$1 AND capability_id=$2 AND capability_version=$3
			AND source_evidence_id=$4 AND policy_version=$5`,
		payload.LearnerID, payload.CapabilityID, payload.CapabilityVersion,
		source.SourceEvidence, firstReviewPolicyVersion).Scan(&existingID)
	if err == nil {
		return false, false, existingID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, false, "", fmt.Errorf("query existing first review: %w", err)
	}
	replaced := false
	var activeID, activeStatus string
	err = tx.QueryRowContext(ctx, `
		SELECT id,status FROM review_items
		WHERE learner_id=$1 AND capability_id=$2 AND capability_version=$3
			AND policy_version=$4 AND status IN ('open','claimed')
		FOR UPDATE`, payload.LearnerID, payload.CapabilityID,
		payload.CapabilityVersion, firstReviewPolicyVersion).Scan(&activeID, &activeStatus)
	if err == nil {
		if activeStatus == "claimed" {
			return false, false, activeID, nil
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE review_items SET status='replaced',replaced_at=$2,updated_at=$2
			WHERE id=$1 AND status='open'`, activeID, createdAt); err != nil {
			return false, false, "", fmt.Errorf("replace stale first review: %w", err)
		}
		replaced = true
	} else if !errors.Is(err, sql.ErrNoRows) {
		return false, false, "", fmt.Errorf("query active first review: %w", err)
	}
	id, err := projectionUUID(s.random)
	if err != nil {
		return false, false, "", err
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
		return false, false, "", fmt.Errorf("insert first review item: %w", err)
	}
	return true, replaced, id, nil
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
