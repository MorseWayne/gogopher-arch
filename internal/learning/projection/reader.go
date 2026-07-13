package projection

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

const defaultRecentEvidenceLimit = 10

type ReaderOptions struct {
	Schema              string
	RecentEvidenceLimit int
}

type Reader struct {
	db                  *sql.DB
	schema              string
	registry            *definition.Registry
	recentEvidenceLimit int
}

type EvidenceSummary struct {
	ID                string    `json:"id"`
	EvaluationBatchID string    `json:"evaluation_batch_id"`
	AttemptID         string    `json:"attempt_id"`
	ActivityID        string    `json:"activity_id"`
	ActivityMode      string    `json:"activity_mode"`
	RuleID            string    `json:"rule_id"`
	EvidenceType      string    `json:"evidence_type"`
	Result            string    `json:"result"`
	Independence      string    `json:"independence"`
	Context           string    `json:"context"`
	Reason            string    `json:"reason"`
	OccurredAt        time.Time `json:"occurred_at"`
}

type CapabilityRead struct {
	ReleaseID      string                    `json:"release_id"`
	Capability     definition.CapabilityView `json:"capability"`
	Snapshot       *Snapshot                 `json:"snapshot"`
	RecentEvidence []EvidenceSummary         `json:"recent_evidence"`
}

type ReviewItemView struct {
	ID                string    `json:"id"`
	ReleaseID         string    `json:"release_id"`
	CapabilityID      string    `json:"capability_id"`
	CapabilityVersion int       `json:"capability_version"`
	GroupKey          string    `json:"group_key"`
	DueAt             time.Time `json:"due_at"`
	Priority          int       `json:"priority"`
	Reason            string    `json:"reason"`
	Status            string    `json:"status"`
	ClaimedAttemptID  string    `json:"claimed_attempt_id,omitempty"`
}

type PrerequisiteStatus struct {
	ID               string `json:"id"`
	RequiredVersion  int    `json:"required_version"`
	Satisfied        bool   `json:"satisfied"`
	SatisfiedVersion int    `json:"satisfied_version,omitempty"`
}

type NextRecommendation struct {
	Kind                     string                             `json:"kind"`
	Reason                   string                             `json:"reason"`
	Activity                 definition.ActivityView            `json:"activity"`
	TargetCapability         *definition.VersionedDefinitionRef `json:"target_capability,omitempty"`
	ReviewItem               *ReviewItemView                    `json:"review_item,omitempty"`
	HardPrerequisites        []PrerequisiteStatus               `json:"hard_prerequisites"`
	RecommendedPrerequisites []PrerequisiteStatus               `json:"recommended_prerequisites"`
}

func NewReader(db *sql.DB, registry *definition.Registry, options ReaderOptions) (*Reader, error) {
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
	limit := options.RecentEvidenceLimit
	if limit == 0 {
		limit = defaultRecentEvidenceLimit
	}
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("recent evidence limit must be between 1 and 100")
	}
	return &Reader{db: db, schema: schema, registry: registry, recentEvidenceLimit: limit}, nil
}

func (r *Reader) Capability(ctx context.Context, learnerID, capabilityID string, asOf time.Time) (CapabilityRead, error) {
	if learnerID == "" || capabilityID == "" || asOf.IsZero() {
		return CapabilityRead{}, fmt.Errorf("learner, capability, and as_of are required")
	}
	releaseID := r.registry.CurrentReleaseID()
	stored, err := r.registry.Latest(releaseID, definition.KindCapability, capabilityID)
	if err != nil {
		return CapabilityRead{}, err
	}
	capability, err := r.registry.CapabilityView(releaseID, stored.ID, stored.Version)
	if err != nil {
		return CapabilityRead{}, err
	}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return CapabilityRead{}, fmt.Errorf("begin capability query: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return CapabilityRead{}, err
	}
	snapshot, err := r.readSnapshot(ctx, tx, learnerID, capability, asOf.UTC())
	if err != nil {
		return CapabilityRead{}, err
	}
	evidence, err := r.readRecentEvidence(ctx, tx, learnerID, capability)
	if err != nil {
		return CapabilityRead{}, err
	}
	if err := tx.Commit(); err != nil {
		return CapabilityRead{}, fmt.Errorf("commit capability query: %w", err)
	}
	return CapabilityRead{ReleaseID: releaseID, Capability: capability, Snapshot: snapshot, RecentEvidence: evidence}, nil
}

func (r *Reader) Next(ctx context.Context, learnerID string, asOf time.Time) (*NextRecommendation, error) {
	if learnerID == "" || asOf.IsZero() {
		return nil, fmt.Errorf("learner and as_of are required")
	}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, fmt.Errorf("begin next learning query: %w", err)
	}
	defer tx.Rollback()
	if err := r.setSearchPath(ctx, tx); err != nil {
		return nil, err
	}
	review, err := r.readNextReview(ctx, tx, learnerID, asOf.UTC())
	if err != nil {
		return nil, err
	}
	if review != nil {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit next review query: %w", err)
		}
		return review, nil
	}
	states, err := r.readCurrentStates(ctx, tx, learnerID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit next acquisition query: %w", err)
	}
	capabilities, activities, err := r.currentDefinitions()
	if err != nil {
		return nil, err
	}
	return selectAcquisition(capabilities, activities, states), nil
}

func (r *Reader) DueReviewCount(ctx context.Context, asOf time.Time) (int, error) {
	if asOf.IsZero() {
		return 0, fmt.Errorf("as_of is required")
	}
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT count(*) FROM "`+r.schema+`".review_items
		WHERE status='claimed' OR (status='open' AND due_at <= $1)`, asOf.UTC()).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count due review items: %w", err)
	}
	return count, nil
}

func (r *Reader) setSearchPath(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `SET LOCAL search_path TO "`+r.schema+`"`); err != nil {
		return fmt.Errorf("set learning query search path: %w", err)
	}
	return nil
}

func (r *Reader) readSnapshot(ctx context.Context, tx *sql.Tx, learnerID string, capability definition.CapabilityView, asOf time.Time) (*Snapshot, error) {
	var value Snapshot
	var lastEvidence, lastIndependent, nextReview sql.NullTime
	err := tx.QueryRowContext(ctx, `
		SELECT learner_id,capability_id,capability_version,capability_hash,
			acquisition_state,independence_state,transfer_state,retention_base_state,
			last_evidence_at,last_independent_at,next_review_at,projection_version,projected_at
		FROM capability_snapshots
		WHERE learner_id=$1 AND capability_id=$2 AND capability_version=$3`,
		learnerID, capability.ID, capability.Version).Scan(
		&value.LearnerID, &value.CapabilityID, &value.CapabilityVersion, &value.CapabilityHash,
		&value.AcquisitionState, &value.IndependenceState, &value.TransferState, &value.RetentionBase,
		&lastEvidence, &lastIndependent, &nextReview, &value.ProjectionVersion, &value.ProjectedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read capability snapshot: %w", err)
	}
	if value.CapabilityHash != capability.ContentHash {
		return nil, fmt.Errorf("capability snapshot hash does not match current definition")
	}
	value.LastEvidenceAt = nullTime(lastEvidence)
	value.LastIndependentAt = nullTime(lastIndependent)
	value.NextReviewAt = nullTime(nextReview)
	value.RetentionState = derivedRetention(value.RetentionBase, value.NextReviewAt, asOf)
	return &value, nil
}

func (r *Reader) readRecentEvidence(ctx context.Context, tx *sql.Tx, learnerID string, capability definition.CapabilityView) ([]EvidenceSummary, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT e.id,e.evaluation_batch_id,e.attempt_id,e.activity_id,a.mode,
			e.evidence_rule_id,e.evidence_type,e.result,e.independence,e.context_level,e.reason,e.occurred_at,e.capability_hash
		FROM evidence_records e
		JOIN learning_attempts a ON a.id=e.attempt_id AND a.learner_id=e.learner_id
		WHERE e.learner_id=$1 AND e.capability_id=$2 AND e.capability_version=$3
		ORDER BY e.occurred_at DESC,e.id DESC
		LIMIT $4`, learnerID, capability.ID, capability.Version, r.recentEvidenceLimit)
	if err != nil {
		return nil, fmt.Errorf("query recent capability evidence: %w", err)
	}
	defer rows.Close()
	result := make([]EvidenceSummary, 0)
	for rows.Next() {
		var value EvidenceSummary
		var capabilityHash string
		if err := rows.Scan(
			&value.ID, &value.EvaluationBatchID, &value.AttemptID, &value.ActivityID, &value.ActivityMode,
			&value.RuleID, &value.EvidenceType, &value.Result, &value.Independence, &value.Context,
			&value.Reason, &value.OccurredAt, &capabilityHash,
		); err != nil {
			return nil, fmt.Errorf("scan recent capability evidence: %w", err)
		}
		if capabilityHash != capability.ContentHash {
			return nil, fmt.Errorf("capability evidence hash does not match current definition")
		}
		result = append(result, value)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent capability evidence: %w", err)
	}
	return result, nil
}

func (r *Reader) readNextReview(ctx context.Context, tx *sql.Tx, learnerID string, asOf time.Time) (*NextRecommendation, error) {
	var item ReviewItemView
	var activityID, activityHash string
	var activityVersion int
	var claimedAttemptID sql.NullString
	err := tx.QueryRowContext(ctx, `
		SELECT id,capability_id,capability_version,review_group_key,due_at,priority,reason,status,
			claimed_attempt_id,release_id,activity_id,activity_version,activity_hash
		FROM review_items
		WHERE learner_id=$1 AND (status='claimed' OR (status='open' AND due_at <= $2))
		ORDER BY CASE status WHEN 'claimed' THEN 0 ELSE 1 END,priority DESC,due_at,created_at,id
		LIMIT 1`, learnerID, asOf).Scan(
		&item.ID, &item.CapabilityID, &item.CapabilityVersion, &item.GroupKey, &item.DueAt,
		&item.Priority, &item.Reason, &item.Status, &claimedAttemptID,
		&item.ReleaseID, &activityID, &activityVersion, &activityHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query next review item: %w", err)
	}
	activity, err := r.registry.ActivityView(item.ReleaseID, activityID, activityVersion)
	if err != nil {
		return nil, fmt.Errorf("resolve next review activity: %w", err)
	}
	if activity.ContentHash != activityHash {
		return nil, fmt.Errorf("next review item does not match frozen activity")
	}
	if claimedAttemptID.Valid {
		item.ClaimedAttemptID = claimedAttemptID.String
	}
	reason := "due_review"
	if item.Status == "claimed" {
		reason = "claimed_review"
	}
	return &NextRecommendation{
		Kind: "review", Reason: reason, Activity: activity, ReviewItem: &item,
		HardPrerequisites: []PrerequisiteStatus{}, RecommendedPrerequisites: []PrerequisiteStatus{},
	}, nil
}

type currentState struct {
	Version     int
	Hash        string
	Acquisition AcquisitionState
}

func (r *Reader) readCurrentStates(ctx context.Context, tx *sql.Tx, learnerID string) (map[string]currentState, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT capability_id,capability_version,capability_hash,acquisition_state
		FROM capability_snapshots WHERE learner_id=$1
		ORDER BY capability_id,capability_version DESC`, learnerID)
	if err != nil {
		return nil, fmt.Errorf("query learner capability states: %w", err)
	}
	defer rows.Close()
	states := make(map[string]currentState)
	for rows.Next() {
		var id string
		var state currentState
		if err := rows.Scan(&id, &state.Version, &state.Hash, &state.Acquisition); err != nil {
			return nil, fmt.Errorf("scan learner capability state: %w", err)
		}
		if _, exists := states[id]; !exists {
			states[id] = state
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate learner capability states: %w", err)
	}
	return states, nil
}

func (r *Reader) currentDefinitions() ([]definition.CapabilityView, []definition.ActivityView, error) {
	releaseID := r.registry.CurrentReleaseID()
	definitions, err := r.registry.Definitions(releaseID)
	if err != nil {
		return nil, nil, err
	}
	latestCapabilities := make(map[string]definition.Definition)
	latestActivities := make(map[string]definition.Definition)
	for _, stored := range definitions {
		switch stored.Kind {
		case definition.KindCapability:
			if current, exists := latestCapabilities[stored.ID]; !exists || stored.Version > current.Version {
				latestCapabilities[stored.ID] = stored
			}
		case definition.KindActivity:
			if current, exists := latestActivities[stored.ID]; !exists || stored.Version > current.Version {
				latestActivities[stored.ID] = stored
			}
		}
	}
	capabilities := make([]definition.CapabilityView, 0, len(latestCapabilities))
	for _, stored := range latestCapabilities {
		value, err := r.registry.CapabilityView(releaseID, stored.ID, stored.Version)
		if err != nil {
			return nil, nil, err
		}
		capabilities = append(capabilities, value)
	}
	sort.Slice(capabilities, func(i, j int) bool { return capabilities[i].ID < capabilities[j].ID })
	activities := make([]definition.ActivityView, 0, len(latestActivities))
	for _, stored := range latestActivities {
		value, err := r.registry.ActivityView(releaseID, stored.ID, stored.Version)
		if err != nil {
			return nil, nil, err
		}
		activities = append(activities, value)
	}
	sort.Slice(activities, func(i, j int) bool { return activities[i].ID < activities[j].ID })
	return capabilities, activities, nil
}

func selectAcquisition(capabilities []definition.CapabilityView, activities []definition.ActivityView, states map[string]currentState) *NextRecommendation {
	currentStates := make(map[string]currentState, len(capabilities))
	for _, capability := range capabilities {
		state, exists := states[capability.ID]
		if exists && state.Version == capability.Version && state.Hash == capability.ContentHash {
			currentStates[capability.ID] = state
		}
	}
	for _, capability := range capabilities {
		state := currentStates[capability.ID]
		if acquisitionRanks[state.Acquisition] >= acquisitionRanks[AcquisitionVerified] {
			continue
		}
		hard := prerequisiteStatuses(capability.Prerequisites.Hard, currentStates)
		if hasUnsatisfied(hard) {
			continue
		}
		activity, found := acquisitionActivity(capability, state.Acquisition, activities)
		if !found {
			continue
		}
		target := definition.VersionedDefinitionRef{ID: capability.ID, Version: capability.Version}
		return &NextRecommendation{
			Kind: "acquisition", Reason: "acquisition_path", Activity: activity, TargetCapability: &target,
			HardPrerequisites:        hard,
			RecommendedPrerequisites: prerequisiteStatuses(capability.Prerequisites.Recommended, currentStates),
		}
	}
	return nil
}

func acquisitionActivity(capability definition.CapabilityView, state AcquisitionState, activities []definition.ActivityView) (definition.ActivityView, bool) {
	ranks := map[string]int{}
	switch state {
	case "", AcquisitionNotStarted:
		ranks = map[string]int{"guided": 0, "practice": 1, "assessment": 2}
	case AcquisitionExploring:
		ranks = map[string]int{"practice": 0, "assessment": 1}
	case AcquisitionPracticed:
		ranks = map[string]int{"assessment": 0}
	default:
		return definition.ActivityView{}, false
	}
	bestRank := len(ranks) + 1
	var best definition.ActivityView
	found := false
	for _, activity := range activities {
		rank, allowed := ranks[activity.Mode]
		if !allowed || !containsRef(activity.CapabilityRefs, capability.ID, capability.Version) {
			continue
		}
		if !found || rank < bestRank || (rank == bestRank && activity.ID < best.ID) {
			best, bestRank, found = activity, rank, true
		}
	}
	return best, found
}

func containsRef(refs []definition.VersionedDefinitionRef, id string, version int) bool {
	for _, ref := range refs {
		if ref.ID == id && ref.Version == version {
			return true
		}
	}
	return false
}

func prerequisiteStatuses(refs []definition.VersionedDefinitionRef, states map[string]currentState) []PrerequisiteStatus {
	result := make([]PrerequisiteStatus, 0, len(refs))
	for _, ref := range refs {
		state, exists := states[ref.ID]
		satisfied := exists && acquisitionRanks[state.Acquisition] >= acquisitionRanks[AcquisitionVerified]
		value := PrerequisiteStatus{ID: ref.ID, RequiredVersion: ref.Version, Satisfied: satisfied}
		if satisfied {
			value.SatisfiedVersion = state.Version
		}
		result = append(result, value)
	}
	return result
}

func hasUnsatisfied(values []PrerequisiteStatus) bool {
	for _, value := range values {
		if !value.Satisfied {
			return true
		}
	}
	return false
}

func derivedRetention(base RetentionBaseState, nextReviewAt *time.Time, asOf time.Time) RetentionState {
	if base == RetentionRusty {
		return RetentionStateRusty
	}
	if nextReviewAt != nil && !nextReviewAt.After(asOf) {
		return RetentionStateDue
	}
	return RetentionStateFresh
}
