package submission

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
	"strings"
	"time"
	"unicode/utf8"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
)

type Repository interface {
	Freeze(context.Context, FreezeRecord) (Result, error)
	Get(context.Context, string, string) (Submission, error)
	Retry(context.Context, RetryRecord) (Result, error)
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
	builder    *execution.SpecBuilder
	random     io.Reader
	now        func() time.Time
}

type SubmitInput struct {
	LearnerID         string
	AttemptID         string
	SubmissionKey     string
	WorkspaceRevision int64
	WorkspaceHash     string
	Explanation       string
}

type RetryInput struct {
	LearnerID    string
	SubmissionID string
	RequestKey   string
}

func NewService(repository Repository, attempts AttemptReader, registry *definition.Registry, builder *execution.SpecBuilder, options ServiceOptions) (*Service, error) {
	if repository == nil || attempts == nil || registry == nil || builder == nil {
		return nil, fmt.Errorf("submission repository, attempt reader, definition registry, and spec builder are required")
	}
	if options.Random == nil {
		options.Random = rand.Reader
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Service{repository: repository, attempts: attempts, registry: registry, builder: builder, random: options.Random, now: options.Now}, nil
}

func (s *Service) Submit(ctx context.Context, input SubmitInput) (Result, error) {
	if !submissionKeyPattern.MatchString(input.SubmissionKey) {
		return Result{}, fmt.Errorf("%w: submission key must be a safe identifier of at most 200 characters", ErrInvalidRequest)
	}
	current, err := s.attempts.Get(ctx, input.LearnerID, input.AttemptID)
	if errors.Is(err, attempt.ErrNotFound) {
		return Result{}, ErrNotFound
	}
	if err != nil {
		return Result{}, err
	}
	activity, err := s.registry.ActivityView(current.ReleaseID, current.ActivityID, current.ActivityVersion)
	if err != nil {
		return Result{}, fmt.Errorf("resolve frozen submission rules: %w", err)
	}
	if activity.ContentHash != current.ActivityHash {
		return Result{}, fmt.Errorf("frozen activity content hash mismatch")
	}
	explanation := strings.TrimSpace(input.Explanation)
	if utf8.RuneCountInString(explanation) > 4000 {
		return Result{}, fmt.Errorf("%w: explanation must contain at most 4000 characters", ErrInvalidRequest)
	}
	if current.Mode == "guided" && utf8.RuneCountInString(explanation) < 20 {
		return Result{}, fmt.Errorf("%w: guided learning explanation must contain at least 20 characters", ErrInvalidRequest)
	}
	submissionID, err := randomUUID(s.random, "submission")
	if err != nil {
		return Result{}, err
	}
	executionID, err := randomUUID(s.random, "submission execution")
	if err != nil {
		return Result{}, err
	}
	spec, err := s.builder.Build(current, executionID, execution.ActionSubmit)
	if err != nil {
		return Result{}, err
	}
	fingerprint := RequestFingerprint(input.WorkspaceRevision, input.WorkspaceHash, current.TaskHash, activity.RuleSetHash, explanation)
	return s.repository.Freeze(ctx, FreezeRecord{
		SubmissionID: submissionID, ExecutionID: executionID, LearnerID: input.LearnerID, AttemptID: input.AttemptID,
		SubmissionKey: input.SubmissionKey, RequestFingerprint: fingerprint,
		Workspace:         cloneWorkspace(current.Workspace),
		WorkspaceRevision: input.WorkspaceRevision, WorkspaceHash: input.WorkspaceHash,
		Explanation: explanation,
		ReleaseID:   current.ReleaseID, ActivityID: current.ActivityID, ActivityVersion: current.ActivityVersion,
		ActivityHash: current.ActivityHash, TaskID: current.TaskID, TaskVersion: current.TaskVersion, TaskHash: current.TaskHash,
		RuleSetHash: activity.RuleSetHash, Spec: spec, CreatedAt: s.now().UTC(),
	})
}

func (s *Service) Get(ctx context.Context, learnerID, submissionID string) (Submission, error) {
	return s.repository.Get(ctx, learnerID, submissionID)
}

func (s *Service) Retry(ctx context.Context, input RetryInput) (Result, error) {
	if !retryKeyPattern.MatchString(input.RequestKey) {
		return Result{}, fmt.Errorf("%w: retry request key must be a safe identifier of at most 190 characters", ErrInvalidRequest)
	}
	frozen, err := s.repository.Get(ctx, input.LearnerID, input.SubmissionID)
	if err != nil {
		return Result{}, err
	}
	executionID, err := randomUUID(s.random, "submission retry execution")
	if err != nil {
		return Result{}, err
	}
	current := attempt.Attempt{
		ID: frozen.AttemptID, LearnerID: frozen.LearnerID, ReleaseID: frozen.ReleaseID,
		ActivityID: frozen.ActivityID, ActivityVersion: frozen.ActivityVersion, ActivityHash: frozen.ActivityHash,
		TaskID: frozen.TaskID, TaskVersion: frozen.TaskVersion, TaskHash: frozen.TaskHash,
		Mode: frozen.Mode, Status: "submit_infra_failed", Workspace: cloneWorkspace(frozen.Workspace),
		WorkspaceRevision: frozen.WorkspaceRevision, WorkspaceHash: frozen.WorkspaceHash,
	}
	spec, err := s.builder.Build(current, executionID, execution.ActionSubmit)
	if err != nil {
		return Result{}, err
	}
	return s.repository.Retry(ctx, RetryRecord{
		ExecutionID: executionID, LearnerID: input.LearnerID, SubmissionID: input.SubmissionID,
		RequestKey: "retry:" + input.RequestKey, RequestFingerprint: frozen.RequestFingerprint,
		Spec: spec, CreatedAt: s.now().UTC(),
	})
}

func RequestFingerprint(workspaceRevision int64, workspaceHash, taskHash, ruleSetHash, explanation string) string {
	hash := sha256.New()
	var revision [8]byte
	binary.BigEndian.PutUint64(revision[:], uint64(workspaceRevision))
	for _, value := range [][]byte{revision[:], []byte(workspaceHash), []byte(taskHash), []byte(ruleSetHash), []byte(explanation)} {
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(len(value)))
		_, _ = hash.Write(size[:])
		_, _ = hash.Write(value)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func randomUUID(source io.Reader, object string) (string, error) {
	var value [16]byte
	if _, err := io.ReadFull(source, value[:]); err != nil {
		return "", fmt.Errorf("generate %s UUID: %w", object, err)
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}

func cloneWorkspace(value map[string]string) map[string]string {
	clone := make(map[string]string, len(value))
	for path, contents := range value {
		clone[path] = contents
	}
	return clone
}

var (
	submissionKeyPattern = regexp.MustCompile("^[A-Za-z0-9][A-Za-z0-9._:-]{0,199}$")
	retryKeyPattern      = regexp.MustCompile("^[A-Za-z0-9][A-Za-z0-9._:-]{0,189}$")
)
