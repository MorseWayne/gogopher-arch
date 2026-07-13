package execution

import (
	"errors"
	"time"
)

var (
	ErrExecutionNotFound  = errors.New("learning execution not found")
	ErrAttemptUnavailable = errors.New("learning attempt is unavailable for execution")
	ErrInvalidRequest     = errors.New("invalid execution request")
	ErrLeaseLost          = errors.New("execution worker lease is no longer valid")
)

type IdempotencyConflict struct {
	ExecutionID string
}

func (e *IdempotencyConflict) Error() string {
	return "execution idempotency key conflicts with its original request"
}

type WorkspaceConflict struct {
	Revision int64
	Hash     string
}

func (e *WorkspaceConflict) Error() string { return "execution workspace revision or hash is stale" }

type Execution struct {
	ID                 string
	AttemptID          string
	SubmissionID       string
	Action             Action
	Sequence           int
	RequestKey         string
	RequestFingerprint string
	ReleaseID          string
	TaskID             string
	TaskVersion        int
	TaskHash           string
	WorkspaceRevision  int64
	WorkspaceHash      string
	Spec               ExecutionSpec
	Status             ExecutionStatus
	Response           *ExecutionResponse
	ClaimCount         int
	LeaseOwner         string
	LeaseExpiresAt     *time.Time
	LeaseHeartbeatAt   *time.Time
	StartedAt          *time.Time
	FinishedAt         *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CreateNormalRecord struct {
	ID                 string
	LearnerID          string
	AttemptID          string
	Action             Action
	RequestKey         string
	RequestFingerprint string
	WorkspaceRevision  int64
	WorkspaceHash      string
	ReleaseID          string
	TaskID             string
	TaskVersion        int
	TaskHash           string
	Spec               ExecutionSpec
	CreatedAt          time.Time
}
