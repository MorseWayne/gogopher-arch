package assistance

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

const maxPayloadBytes = 16 << 10

type Repository interface {
	Record(context.Context, Record) (RecordResult, error)
	ListThrough(context.Context, string, string, int64) ([]Event, error)
}

type AttemptReader interface {
	Get(context.Context, string, string) (attempt.Attempt, error)
}

type ServiceOptions struct {
	Random io.Reader
	Now    func() time.Time
}

type Service struct {
	repository Repository
	attempts   AttemptReader
	registry   *definition.Registry
	random     io.Reader
	now        func() time.Time
}

type RecordInput struct {
	LearnerID string
	AttemptID string
	EventKey  string
	Type      EventType
	Payload   map[string]any
}

type RevealHintInput struct {
	LearnerID string
	AttemptID string
	EventKey  string
	Hint      Hint
}

func NewService(repository Repository, attempts AttemptReader, registry *definition.Registry, options ServiceOptions) (*Service, error) {
	if repository == nil || attempts == nil || registry == nil {
		return nil, fmt.Errorf("assistance repository, attempt reader, and definition registry are required")
	}
	if options.Random == nil {
		options.Random = rand.Reader
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Service{repository: repository, attempts: attempts, registry: registry, random: options.Random, now: options.Now}, nil
}

func (s *Service) Record(ctx context.Context, input RecordInput) (RecordResult, error) {
	if !eventKeyPattern.MatchString(input.EventKey) {
		return RecordResult{}, fmt.Errorf("%w: event key must be a safe identifier of at most 200 characters", ErrInvalidRequest)
	}
	if !input.Type.valid() {
		return RecordResult{}, fmt.Errorf("%w: unsupported assistance event type %q", ErrInvalidRequest, input.Type)
	}
	current, err := s.attempts.Get(ctx, input.LearnerID, input.AttemptID)
	if err != nil {
		return RecordResult{}, err
	}
	activity, err := s.registry.ActivityView(current.ReleaseID, current.ActivityID, current.ActivityVersion)
	if err != nil {
		return RecordResult{}, fmt.Errorf("resolve frozen assistance policy: %w", err)
	}
	if !eventAllowed(activity.AssistancePolicy, input.Type) {
		return RecordResult{}, ErrEventNotAllowed
	}
	payloadObject := input.Payload
	if payloadObject == nil {
		payloadObject = map[string]any{}
	}
	payload, err := json.Marshal(payloadObject)
	if err != nil || len(payload) > maxPayloadBytes {
		return RecordResult{}, fmt.Errorf("%w: payload must be a JSON object of at most %d bytes", ErrInvalidRequest, maxPayloadBytes)
	}
	payload, err = definition.CanonicalJSON(payload)
	if err != nil {
		return RecordResult{}, fmt.Errorf("canonicalize assistance payload: %w", err)
	}
	id, err := randomEventUUID(s.random)
	if err != nil {
		return RecordResult{}, err
	}
	return s.repository.Record(ctx, Record{
		ID: id, AttemptID: input.AttemptID, LearnerID: input.LearnerID,
		EventKey: input.EventKey, Type: input.Type, Payload: payload, CreatedAt: s.now().UTC(),
	})
}

func (s *Service) RevealHint(ctx context.Context, input RevealHintInput) (Hint, RecordResult, error) {
	if input.Hint.ID == "" || input.Hint.Body == "" {
		return Hint{}, RecordResult{}, fmt.Errorf("%w: hint ID and body are required", ErrInvalidRequest)
	}
	result, err := s.Record(ctx, RecordInput{
		LearnerID: input.LearnerID, AttemptID: input.AttemptID, EventKey: input.EventKey,
		Type: HintRevealed, Payload: map[string]any{"hint_id": input.Hint.ID},
	})
	if err != nil {
		return Hint{}, RecordResult{}, err
	}
	return input.Hint, result, nil
}

func (s *Service) EventsThrough(ctx context.Context, learnerID, attemptID string, cutoff int64) ([]Event, error) {
	if cutoff < 0 {
		return nil, fmt.Errorf("assistance cutoff may not be negative")
	}
	return s.repository.ListThrough(ctx, learnerID, attemptID, cutoff)
}

func eventAllowed(policy definition.AssistancePolicyView, eventType EventType) bool {
	switch eventType {
	case HintRevealed:
		return policy.Hints
	case ReferenceOpened:
		return policy.References
	case SolutionViewed:
		return policy.Solution
	case AIDeclared:
		return policy.AIDeclaration
	default:
		return false
	}
}

func (e EventType) valid() bool {
	return e == HintRevealed || e == ReferenceOpened || e == SolutionViewed || e == AIDeclared
}

func randomEventUUID(source io.Reader) (string, error) {
	var value [16]byte
	if _, err := io.ReadFull(source, value[:]); err != nil {
		return "", fmt.Errorf("generate assistance event UUID: %w", err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}

var eventKeyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,199}$`)
