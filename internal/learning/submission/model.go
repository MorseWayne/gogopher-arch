package submission

import (
	"errors"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
)

var (
	ErrNotFound         = errors.New("learning submission not found")
	ErrAttemptInactive  = errors.New("learning attempt cannot be submitted")
	ErrRetryUnavailable = errors.New("learning submission is not awaiting an infrastructure retry")
)

type IdempotencyConflict struct{ SubmissionID string }

func (e *IdempotencyConflict) Error() string {
	return "submission key conflicts with its original request"
}

type AttemptAlreadySubmitted struct{ SubmissionID string }

func (e *AttemptAlreadySubmitted) Error() string {
	return "learning attempt already has a frozen submission"
}

type WorkspaceConflict struct {
	Revision int64
	Hash     string
}

func (e *WorkspaceConflict) Error() string {
	return "submission workspace revision or hash is stale"
}

type Status string

const (
	StatusFrozen      Status = "frozen"
	StatusExecuting   Status = "executing"
	StatusEvaluated   Status = "evaluated"
	StatusInfraFailed Status = "infra_failed"
)

type Submission struct {
	ID                    string                    `json:"id"`
	AttemptID             string                    `json:"attempt_id"`
	LearnerID             string                    `json:"-"`
	SubmissionKey         string                    `json:"submission_key"`
	RequestFingerprint    string                    `json:"-"`
	Workspace             map[string]string         `json:"-"`
	WorkspaceRevision     int64                     `json:"workspace_revision"`
	WorkspaceHash         string                    `json:"workspace_hash"`
	RuleSetHash           string                    `json:"rule_set_hash"`
	AssistanceCutoff      int64                     `json:"assistance_cutoff_seq"`
	Status                Status                    `json:"status"`
	LatestExecutionID     string                    `json:"latest_execution_id"`
	LatestExecutionSeq    int                       `json:"latest_execution_sequence"`
	LatestExecutionStatus execution.ExecutionStatus `json:"latest_execution_status"`
	CreatedAt             time.Time                 `json:"created_at"`
	EvaluatedAt           *time.Time                `json:"evaluated_at,omitempty"`

	ReleaseID       string `json:"-"`
	ActivityID      string `json:"-"`
	ActivityVersion int    `json:"-"`
	ActivityHash    string `json:"-"`
	TaskID          string `json:"-"`
	TaskVersion     int    `json:"-"`
	TaskHash        string `json:"-"`
	Mode            string `json:"-"`
}

type Result struct {
	Submission        Submission
	ExecutionID       string
	ExecutionSequence int
	Created           bool
}
