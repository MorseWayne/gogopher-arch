package projection

import "time"

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
	EvidenceType string
	RuleID       string
	Result       string
	Independence IndependenceState
	Context      TransferState
	ActivityMode string
	OccurredAt   time.Time
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
