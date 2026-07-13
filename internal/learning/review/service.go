package review

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

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

var (
	ErrNotFound    = errors.New("review item not found")
	ErrUnavailable = errors.New("review item is unavailable")
	schemaPattern  = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)
)

type ClaimResult struct {
	Attempt attempt.Attempt
	Created bool
}

type TransitionObserver interface {
	ReviewItemsTransitioned(string, int)
}

type ServiceOptions struct {
	Schema   string
	Random   io.Reader
	Now      func() time.Time
	Observer TransitionObserver
}

type Service struct {
	db       *sql.DB
	schema   string
	registry *definition.Registry
	random   io.Reader
	now      func() time.Time
	observer TransitionObserver
}

type reviewItem struct {
	ID                string
	ReleaseID         string
	ActivityID        string
	ActivityVersion   int
	ActivityHash      string
	GroupKey          string
	CapabilityID      string
	CapabilityVersion int
	Reason            string
	Status            string
	DueAt             time.Time
	AttemptID         sql.NullString
}

func NewService(db *sql.DB, registry *definition.Registry, options ServiceOptions) (*Service, error) {
	if db == nil || registry == nil {
		return nil, fmt.Errorf("database and definition registry are required")
	}
	schema := options.Schema
	if schema == "" {
		schema = "public"
	}
	if !schemaPattern.MatchString(schema) {
		return nil, fmt.Errorf("invalid PostgreSQL schema %q", schema)
	}
	if options.Random == nil {
		options.Random = rand.Reader
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Service{
		db: db, schema: schema, registry: registry, random: options.Random,
		now: options.Now, observer: options.Observer,
	}, nil
}

func (s *Service) Claim(ctx context.Context, learnerID, reviewItemID string) (ClaimResult, error) {
	return s.ClaimAt(ctx, learnerID, reviewItemID, s.now().UTC())
}

func (s *Service) ClaimAt(ctx context.Context, learnerID, reviewItemID string, asOf time.Time) (ClaimResult, error) {
	if learnerID == "" || reviewItemID == "" {
		return ClaimResult{}, ErrNotFound
	}
	if asOf.IsZero() {
		return ClaimResult{}, fmt.Errorf("review claim time is required")
	}
	asOf = asOf.UTC()
	lifecycleAt := s.now().UTC()
	if lifecycleAt.IsZero() {
		return ClaimResult{}, fmt.Errorf("review lifecycle time is required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ClaimResult{}, fmt.Errorf("begin review claim: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `SET LOCAL search_path TO "`+s.schema+`"`); err != nil {
		return ClaimResult{}, fmt.Errorf("set review claim search path: %w", err)
	}
	requested, err := loadRequestedItem(ctx, tx, learnerID, reviewItemID)
	if err != nil {
		return ClaimResult{}, err
	}
	if requested.Status == "completed" || requested.Status == "replaced" {
		return ClaimResult{}, ErrUnavailable
	}
	items, err := lockActiveGroup(ctx, tx, learnerID, requested, asOf)
	if err != nil {
		return ClaimResult{}, err
	}
	if len(items) == 0 {
		return ClaimResult{}, ErrUnavailable
	}
	foundRequested := false
	for _, item := range items {
		if item.ID == reviewItemID {
			requested = item
			foundRequested = true
			break
		}
	}
	if !foundRequested {
		return ClaimResult{}, ErrUnavailable
	}

	activity, err := s.registry.ActivityView(requested.ReleaseID, requested.ActivityID, requested.ActivityVersion)
	if err != nil {
		return ClaimResult{}, fmt.Errorf("resolve frozen review activity: %w", err)
	}
	if (activity.Mode != "review" && activity.Mode != "practice" && activity.Mode != "guided") || activity.ContentHash != requested.ActivityHash {
		return ClaimResult{}, fmt.Errorf("review item does not match a frozen review or remediation activity")
	}
	if err := validateGroupCapabilities(items, activity.CapabilityRefs); err != nil {
		return ClaimResult{}, err
	}

	attemptID, err := claimedAttemptID(items)
	if err != nil {
		return ClaimResult{}, err
	}
	created := attemptID == ""
	var value attempt.Attempt
	if created {
		value, err = s.buildAttempt(requested.ReleaseID, learnerID, activity, lifecycleAt)
		if err != nil {
			return ClaimResult{}, err
		}
		attemptID = value.ID
		if err := insertAttempt(ctx, tx, value); err != nil {
			return ClaimResult{}, err
		}
	} else {
		value, err = loadAttempt(ctx, tx, learnerID, attemptID)
		if err != nil {
			return ClaimResult{}, err
		}
	}
	claimedCount := 0
	for _, item := range items {
		if item.Status == "open" {
			result, err := tx.ExecContext(ctx, `
				UPDATE review_items
				SET status='claimed',claimed_attempt_id=$2,updated_at=$3
				WHERE id=$1 AND status='open'`, item.ID, attemptID, lifecycleAt)
			if err != nil {
				return ClaimResult{}, fmt.Errorf("claim review item %s: %w", item.ID, err)
			}
			if affected, err := result.RowsAffected(); err != nil || affected != 1 {
				return ClaimResult{}, fmt.Errorf("claim review item %s changed %d rows", item.ID, affected)
			}
			if _, err := tx.ExecContext(ctx, `
			INSERT INTO attempt_review_items (attempt_id,review_item_id,created_at)
			VALUES ($1,$2,$3)`, attemptID, item.ID, lifecycleAt); err != nil {
				return ClaimResult{}, fmt.Errorf("link review item %s: %w", item.ID, err)
			}
			claimedCount++
			continue
		}
		var linkedAttemptID string
		if err := tx.QueryRowContext(ctx, `
			SELECT attempt_id FROM attempt_review_items WHERE review_item_id=$1`, item.ID).Scan(&linkedAttemptID); err != nil {
			return ClaimResult{}, fmt.Errorf("load claimed review item link %s: %w", item.ID, err)
		}
		if linkedAttemptID != attemptID {
			return ClaimResult{}, fmt.Errorf("review item %s link conflicts with claimed attempt", item.ID)
		}
	}
	if err := tx.Commit(); err != nil {
		return ClaimResult{}, fmt.Errorf("commit review claim: %w", err)
	}
	if s.observer != nil && claimedCount > 0 {
		s.observer.ReviewItemsTransitioned("claimed", claimedCount)
	}
	return ClaimResult{Attempt: value, Created: created}, nil
}

func loadRequestedItem(ctx context.Context, tx *sql.Tx, learnerID, reviewItemID string) (reviewItem, error) {
	var item reviewItem
	err := tx.QueryRowContext(ctx, `
		SELECT id,release_id,activity_id,activity_version,activity_hash,review_group_key,
			capability_id,capability_version,reason,status,due_at,claimed_attempt_id
		FROM review_items WHERE id=$1 AND learner_id=$2`, reviewItemID, learnerID).Scan(
		&item.ID, &item.ReleaseID, &item.ActivityID, &item.ActivityVersion, &item.ActivityHash,
		&item.GroupKey, &item.CapabilityID, &item.CapabilityVersion, &item.Reason, &item.Status, &item.DueAt, &item.AttemptID)
	if errors.Is(err, sql.ErrNoRows) {
		return reviewItem{}, ErrNotFound
	}
	if err != nil {
		return reviewItem{}, fmt.Errorf("load requested review item: %w", err)
	}
	return item, nil
}

func lockActiveGroup(ctx context.Context, tx *sql.Tx, learnerID string, requested reviewItem, asOf time.Time) ([]reviewItem, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id,release_id,activity_id,activity_version,activity_hash,review_group_key,
			capability_id,capability_version,reason,status,due_at,claimed_attempt_id
		FROM review_items
		WHERE learner_id=$1 AND review_group_key=$2 AND release_id=$3
			AND activity_id=$4 AND activity_version=$5
			AND (status='claimed' OR (status='open' AND due_at <= $6))
		ORDER BY id FOR UPDATE`, learnerID, requested.GroupKey, requested.ReleaseID,
		requested.ActivityID, requested.ActivityVersion, asOf)
	if err != nil {
		return nil, fmt.Errorf("lock active review group: %w", err)
	}
	defer rows.Close()
	var items []reviewItem
	for rows.Next() {
		var item reviewItem
		if err := rows.Scan(
			&item.ID, &item.ReleaseID, &item.ActivityID, &item.ActivityVersion, &item.ActivityHash,
			&item.GroupKey, &item.CapabilityID, &item.CapabilityVersion, &item.Reason, &item.Status, &item.DueAt, &item.AttemptID,
		); err != nil {
			return nil, fmt.Errorf("scan active review group: %w", err)
		}
		if item.ActivityHash != requested.ActivityHash {
			return nil, fmt.Errorf("review group contains conflicting activity hashes")
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate active review group: %w", err)
	}
	return items, nil
}

func validateGroupCapabilities(items []reviewItem, refs []definition.VersionedDefinitionRef) error {
	allowed := make(map[string]struct{}, len(refs))
	for _, ref := range refs {
		allowed[fmt.Sprintf("%s@%d", ref.ID, ref.Version)] = struct{}{}
	}
	for _, item := range items {
		if _, ok := allowed[fmt.Sprintf("%s@%d", item.CapabilityID, item.CapabilityVersion)]; !ok {
			return fmt.Errorf("review group capability %s@%d is absent from frozen activity", item.CapabilityID, item.CapabilityVersion)
		}
	}
	return nil
}

func claimedAttemptID(items []reviewItem) (string, error) {
	var result string
	for _, item := range items {
		if item.Status != "claimed" || !item.AttemptID.Valid {
			continue
		}
		if result != "" && result != item.AttemptID.String {
			return "", fmt.Errorf("review group is split across attempts")
		}
		result = item.AttemptID.String
	}
	return result, nil
}

func (s *Service) buildAttempt(releaseID, learnerID string, activity definition.ActivityView, claimedAt time.Time) (attempt.Attempt, error) {
	task, err := s.registry.TaskView(releaseID, activity.TaskRef.ID, activity.TaskRef.Version)
	if err != nil {
		return attempt.Attempt{}, fmt.Errorf("resolve frozen review task: %w", err)
	}
	workspace, err := s.registry.PublicWorkspace(releaseID, task.ID, task.Version)
	if err != nil {
		return attempt.Attempt{}, fmt.Errorf("restore review workspace: %w", err)
	}
	id, err := randomUUID(s.random)
	if err != nil {
		return attempt.Attempt{}, err
	}
	now := claimedAt.UTC()
	return attempt.Attempt{
		ID: id, LearnerID: learnerID, ReleaseID: releaseID,
		ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash,
		TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
		CapabilityRefs: append([]definition.VersionedDefinitionRef(nil), activity.CapabilityRefs...),
		Mode:           activity.Mode, Status: "active", Workspace: workspace, WorkspaceHash: attempt.WorkspaceHash(workspace),
		StartedAt: now, UpdatedAt: now,
	}, nil
}

func insertAttempt(ctx context.Context, tx *sql.Tx, value attempt.Attempt) error {
	capabilities, err := json.Marshal(value.CapabilityRefs)
	if err != nil {
		return err
	}
	workspace, err := json.Marshal(value.Workspace)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO learning_attempts (
			id,learner_id,release_id,activity_id,activity_version,activity_hash,
			task_id,task_version,task_hash,capability_refs,mode,status,
			workspace,workspace_revision,workspace_hash,started_at,updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11,'active',$12::jsonb,0,$13,$14,$14)`,
		value.ID, value.LearnerID, value.ReleaseID, value.ActivityID, value.ActivityVersion, value.ActivityHash,
		value.TaskID, value.TaskVersion, value.TaskHash, string(capabilities), value.Mode,
		string(workspace), value.WorkspaceHash, value.StartedAt)
	if err != nil {
		return fmt.Errorf("insert review attempt: %w", err)
	}
	return nil
}

func loadAttempt(ctx context.Context, tx *sql.Tx, learnerID, attemptID string) (attempt.Attempt, error) {
	var value attempt.Attempt
	var capabilities, workspace []byte
	err := tx.QueryRowContext(ctx, `
		SELECT id,learner_id,release_id,activity_id,activity_version,activity_hash,
			task_id,task_version,task_hash,capability_refs,mode,status,workspace,
			workspace_revision,workspace_hash,started_at,updated_at
		FROM learning_attempts WHERE id=$1 AND learner_id=$2`, attemptID, learnerID).Scan(
		&value.ID, &value.LearnerID, &value.ReleaseID, &value.ActivityID, &value.ActivityVersion,
		&value.ActivityHash, &value.TaskID, &value.TaskVersion, &value.TaskHash, &capabilities,
		&value.Mode, &value.Status, &workspace, &value.WorkspaceRevision, &value.WorkspaceHash,
		&value.StartedAt, &value.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return attempt.Attempt{}, fmt.Errorf("claimed review attempt is unavailable")
	}
	if err != nil {
		return attempt.Attempt{}, fmt.Errorf("load claimed review attempt: %w", err)
	}
	if err := json.Unmarshal(capabilities, &value.CapabilityRefs); err != nil {
		return attempt.Attempt{}, fmt.Errorf("decode review attempt capabilities: %w", err)
	}
	if err := json.Unmarshal(workspace, &value.Workspace); err != nil {
		return attempt.Attempt{}, fmt.Errorf("decode review attempt workspace: %w", err)
	}
	return value, nil
}

func randomUUID(source io.Reader) (string, error) {
	var value [16]byte
	if _, err := io.ReadFull(source, value[:]); err != nil {
		return "", fmt.Errorf("generate review attempt UUID: %w", err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}
