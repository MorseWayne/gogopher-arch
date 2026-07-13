package execution

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
)

type ServiceRepository interface {
	FindNormal(context.Context, string, string, string) (Execution, error)
	CreateNormal(context.Context, CreateNormalRecord) (Execution, bool, error)
	Get(context.Context, string, string) (Execution, error)
}

type AttemptReader interface {
	Get(context.Context, string, string) (attempt.Attempt, error)
}

type ServiceOptions struct {
	Random io.Reader
	Now    func() time.Time
}

type Service struct {
	repository ServiceRepository
	attempts   AttemptReader
	builder    *SpecBuilder
	random     io.Reader
	now        func() time.Time
}

type CreateInput struct {
	LearnerID         string
	AttemptID         string
	Action            Action
	RequestKey        string
	WorkspaceRevision int64
	WorkspaceHash     string
}

func NewService(repository ServiceRepository, attempts AttemptReader, builder *SpecBuilder, options ServiceOptions) (*Service, error) {
	if repository == nil || attempts == nil || builder == nil {
		return nil, fmt.Errorf("execution repository, attempt reader, and spec builder are required")
	}
	if options.Random == nil {
		options.Random = rand.Reader
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Service{repository: repository, attempts: attempts, builder: builder, random: options.Random, now: options.Now}, nil
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Execution, error) {
	if input.Action != ActionBuild && input.Action != ActionTest && input.Action != ActionVet {
		return Execution{}, fmt.Errorf("%w: %v", ErrInvalidRequest, ErrActionNotAllowed)
	}
	if !requestKeyPattern.MatchString(input.RequestKey) {
		return Execution{}, fmt.Errorf("%w: request key must be a safe identifier of at most 200 characters", ErrInvalidRequest)
	}
	existing, err := s.repository.FindNormal(ctx, input.LearnerID, input.AttemptID, input.RequestKey)
	if err == nil {
		fingerprint := RequestFingerprint(input.Action, input.WorkspaceRevision, input.WorkspaceHash, existing.TaskHash)
		if fingerprint != existing.RequestFingerprint {
			return Execution{}, &IdempotencyConflict{ExecutionID: existing.ID}
		}
		return existing, nil
	}
	if !errors.Is(err, ErrExecutionNotFound) {
		return Execution{}, err
	}
	current, err := s.attempts.Get(ctx, input.LearnerID, input.AttemptID)
	if err != nil {
		return Execution{}, err
	}
	if current.Status != "active" {
		return Execution{}, ErrAttemptUnavailable
	}
	if current.WorkspaceRevision != input.WorkspaceRevision || current.WorkspaceHash != input.WorkspaceHash {
		return Execution{}, &WorkspaceConflict{Revision: current.WorkspaceRevision, Hash: current.WorkspaceHash}
	}
	id, err := randomExecutionUUID(s.random)
	if err != nil {
		return Execution{}, err
	}
	spec, err := s.builder.Build(current, id, input.Action)
	if err != nil {
		return Execution{}, err
	}
	fingerprint := RequestFingerprint(input.Action, input.WorkspaceRevision, input.WorkspaceHash, current.TaskHash)
	created, _, err := s.repository.CreateNormal(ctx, CreateNormalRecord{
		ID: id, LearnerID: input.LearnerID, AttemptID: input.AttemptID,
		Action: input.Action, RequestKey: input.RequestKey, RequestFingerprint: fingerprint,
		WorkspaceRevision: input.WorkspaceRevision, WorkspaceHash: input.WorkspaceHash,
		ReleaseID: current.ReleaseID, TaskID: current.TaskID, TaskVersion: current.TaskVersion, TaskHash: current.TaskHash,
		Spec: spec, CreatedAt: s.now().UTC(),
	})
	return created, err
}

func (s *Service) Get(ctx context.Context, learnerID, executionID string) (Execution, error) {
	return s.repository.Get(ctx, learnerID, executionID)
}

func RequestFingerprint(action Action, workspaceRevision int64, workspaceHash, taskHash string) string {
	hash := sha256.New()
	writeFingerprintField(hash, []byte(action))
	var revision [8]byte
	binary.BigEndian.PutUint64(revision[:], uint64(workspaceRevision))
	writeFingerprintField(hash, revision[:])
	writeFingerprintField(hash, []byte(workspaceHash))
	writeFingerprintField(hash, []byte(taskHash))
	return hex.EncodeToString(hash.Sum(nil))
}

func writeFingerprintField(target io.Writer, value []byte) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	_, _ = target.Write(size[:])
	_, _ = target.Write(value)
}

func randomExecutionUUID(source io.Reader) (string, error) {
	var value [16]byte
	if _, err := io.ReadFull(source, value[:]); err != nil {
		return "", fmt.Errorf("generate execution UUID: %w", err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}

var requestKeyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,199}$`)
