package projection

import (
	"encoding/json"
	"time"
)

const (
	ProjectionRequestEventVersion  = 1
	ProjectionTargetEventVersion   = 2
	ReviewSchedulerEventVersion    = 1
	ReviewOutcomeEventVersion      = 2
	ProjectionConsumerVersion      = 1
	ReviewSchedulerConsumerVersion = 2
)

type AcquisitionState string

const (
	AcquisitionNotStarted AcquisitionState = "not_started"
	AcquisitionExploring  AcquisitionState = "exploring"
	AcquisitionPracticed  AcquisitionState = "practiced"
	AcquisitionVerified   AcquisitionState = "verified"
	AcquisitionStable     AcquisitionState = "stable"
)

type IndependenceState string

const (
	IndependenceUnverified  IndependenceState = "unverified"
	IndependenceGuided      IndependenceState = "guided"
	IndependenceAIAssisted  IndependenceState = "ai_assisted"
	IndependenceHinted      IndependenceState = "hinted"
	IndependenceReferenced  IndependenceState = "referenced"
	IndependenceIndependent IndependenceState = "independent"
)

type TransferState string

const (
	TransferUnverified  TransferState = "unverified"
	TransferSameContext TransferState = "same_context"
	TransferVariant     TransferState = "variant"
	TransferNewProject  TransferState = "new_project"
)

type RetentionBaseState string

const (
	RetentionFresh RetentionBaseState = "fresh"
	RetentionRusty RetentionBaseState = "rusty"
)

type RetentionState string

const (
	RetentionStateFresh RetentionState = "fresh"
	RetentionStateDue   RetentionState = "due"
	RetentionStateRusty RetentionState = "rusty"
)

type EvidenceFact struct {
	EvidenceType     string
	RuleID           string
	Result           string
	Independence     IndependenceState
	Context          TransferState
	ActivityMode     string
	QualifyingReview bool
	OccurredAt       time.Time
}

type Input struct {
	Evidence      []EvidenceFact
	RetentionBase RetentionBaseState
	NextReviewAt  *time.Time
	AsOf          time.Time
}

type Result struct {
	AcquisitionState  AcquisitionState   `json:"acquisition_state"`
	IndependenceState IndependenceState  `json:"independence_state"`
	TransferState     TransferState      `json:"transfer_state"`
	RetentionBase     RetentionBaseState `json:"retention_base_state"`
	RetentionState    RetentionState     `json:"retention_state"`
	LastEvidenceAt    *time.Time         `json:"last_evidence_at,omitempty"`
	LastIndependentAt *time.Time         `json:"last_independent_at,omitempty"`
	NextReviewAt      *time.Time         `json:"next_review_at,omitempty"`
}

type Snapshot struct {
	LearnerID         string    `json:"learner_id"`
	CapabilityID      string    `json:"capability_id"`
	CapabilityVersion int       `json:"capability_version"`
	CapabilityHash    string    `json:"capability_hash"`
	ProjectionVersion int       `json:"projection_version"`
	ProjectedAt       time.Time `json:"projected_at"`
	Result
}

type RebuildInput struct {
	LearnerID         string    `json:"learner_id"`
	ReleaseID         string    `json:"release_id"`
	CapabilityID      string    `json:"capability_id"`
	CapabilityVersion int       `json:"capability_version"`
	AsOf              time.Time `json:"as_of"`
}

type Request struct {
	ID           string
	Payload      json.RawMessage
	AttemptCount int
	CreatedAt    time.Time
}

type ProjectionRequestPayload struct {
	EventVersion      int    `json:"event_version"`
	EvaluationBatchID string `json:"evaluation_batch_id"`
	LearnerID         string `json:"learner_id"`
	ReleaseID         string `json:"release_id,omitempty"`
	CapabilityID      string `json:"capability_id,omitempty"`
	CapabilityVersion int    `json:"capability_version,omitempty"`
}

type ReviewSchedulerRequestPayload struct {
	EventVersion      int                `json:"event_version"`
	ProjectionVersion int                `json:"projection_version"`
	LearnerID         string             `json:"learner_id"`
	ReleaseID         string             `json:"release_id"`
	CapabilityID      string             `json:"capability_id"`
	CapabilityVersion int                `json:"capability_version"`
	CapabilityHash    string             `json:"capability_hash"`
	AcquisitionState  AcquisitionState   `json:"acquisition_state"`
	IndependenceState IndependenceState  `json:"independence_state"`
	TransferState     TransferState      `json:"transfer_state"`
	RetentionBase     RetentionBaseState `json:"retention_base_state"`
}

type ReviewOutcomeRequestPayload struct {
	EventVersion      int    `json:"event_version"`
	EvaluationBatchID string `json:"evaluation_batch_id"`
	LearnerID         string `json:"learner_id"`
}

type RetryResult struct {
	AttemptCount int
	AvailableAt  *time.Time
	Exhausted    bool
}
