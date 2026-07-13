package attempt

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

type Repository interface {
	Create(context.Context, CreateRecord) (Attempt, error)
	Get(context.Context, string, string) (Attempt, error)
	Save(context.Context, SaveRecord) (Attempt, error)
}

type ServiceOptions struct {
	Random io.Reader
	Now    func() time.Time
}

type Service struct {
	repository Repository
	registry   *definition.Registry
	random     io.Reader
	now        func() time.Time
}

func NewService(repository Repository, registry *definition.Registry, options ServiceOptions) (*Service, error) {
	if repository == nil || registry == nil {
		return nil, fmt.Errorf("attempt repository and definition registry are required")
	}
	if options.Random == nil {
		options.Random = rand.Reader
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Service{repository: repository, registry: registry, random: options.Random, now: options.Now}, nil
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Attempt, error) {
	releaseID := s.registry.CurrentReleaseID()
	activity, err := s.registry.ActivityView(releaseID, input.ActivityID, input.ActivityVersion)
	if err != nil {
		return Attempt{}, fmt.Errorf("resolve activity: %w", err)
	}
	task, err := s.registry.TaskView(releaseID, activity.TaskRef.ID, activity.TaskRef.Version)
	if err != nil {
		return Attempt{}, fmt.Errorf("resolve task: %w", err)
	}
	workspace, err := s.registry.PublicWorkspace(releaseID, task.ID, task.Version)
	if err != nil {
		return Attempt{}, fmt.Errorf("restore starter workspace: %w", err)
	}
	id, err := randomUUID(s.random)
	if err != nil {
		return Attempt{}, err
	}
	now := s.now().UTC()
	record := Attempt{
		ID: id, LearnerID: input.LearnerID, ReleaseID: releaseID,
		ActivityID: activity.ID, ActivityVersion: activity.Version, ActivityHash: activity.ContentHash,
		TaskID: task.ID, TaskVersion: task.Version, TaskHash: task.BundleHash,
		CapabilityRefs: append([]definition.VersionedDefinitionRef(nil), activity.CapabilityRefs...),
		Mode:           activity.Mode, Status: "active", Workspace: cloneWorkspace(workspace),
		WorkspaceHash: WorkspaceHash(workspace), StartedAt: now, UpdatedAt: now,
	}
	return s.repository.Create(ctx, CreateRecord{Attempt: record})
}

func (s *Service) Get(ctx context.Context, learnerID, attemptID string) (Attempt, error) {
	return s.repository.Get(ctx, learnerID, attemptID)
}

func (s *Service) Save(ctx context.Context, input SaveInput) (Attempt, error) {
	current, err := s.repository.Get(ctx, input.LearnerID, input.AttemptID)
	if err != nil {
		return Attempt{}, err
	}
	view, err := s.registry.TaskView(current.ReleaseID, current.TaskID, current.TaskVersion)
	if err != nil {
		return Attempt{}, fmt.Errorf("resolve frozen task: %w", err)
	}
	baseline, err := s.registry.PublicWorkspace(current.ReleaseID, current.TaskID, current.TaskVersion)
	if err != nil {
		return Attempt{}, fmt.Errorf("restore frozen workspace: %w", err)
	}
	if err := ValidateWorkspace(view, baseline, input.Files); err != nil {
		return Attempt{}, fmt.Errorf("%w: %v", ErrInvalidWorkspace, err)
	}
	return s.repository.Save(ctx, SaveRecord{
		AttemptID: input.AttemptID, LearnerID: input.LearnerID, BaseRevision: input.BaseRevision,
		Workspace: cloneWorkspace(input.Files), WorkspaceHash: WorkspaceHash(input.Files), UpdatedAt: s.now().UTC(),
	})
}

func cloneWorkspace(files map[string]string) map[string]string {
	clone := make(map[string]string, len(files))
	for path, contents := range files {
		clone[path] = contents
	}
	return clone
}

func randomUUID(source io.Reader) (string, error) {
	var value [16]byte
	if _, err := io.ReadFull(source, value[:]); err != nil {
		return "", fmt.Errorf("generate attempt UUID: %w", err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}
