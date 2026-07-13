package evaluation

import (
	"encoding/json"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/assistance"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
)

type Evidence struct {
	ID                string
	EvaluationBatchID string
	LearnerID         string
	CapabilityID      string
	CapabilityVersion int
	CapabilityHash    string
	AttemptID         string
	ActivityID        string
	ArtifactID        string
	EvidenceRuleID    string
	EvidenceType      string
	Result            execution.RuleStatus
	Independence      assistance.Independence
	ContextLevel      string
	Evaluator         string
	RuleVersion       int
	Reason            string
	OccurredAt        time.Time
	CreatedAt         time.Time
}

type Artifact struct {
	ID           string
	AttemptID    string
	SubmissionID string
	Kind         string
	Content      json.RawMessage
	ContentBytes int
	ContentHash  string
	CreatedAt    time.Time
}

type Batch struct {
	ID           string
	SubmissionID string
	ExecutionID  string
	RuleSetHash  string
	RuleResults  []execution.RuleResult
	Artifacts    []Artifact
	Evidence     []Evidence
	CreatedAt    time.Time
}

type PersistRecord struct {
	Batch           Batch
	AttemptID       string
	LearnerID       string
	ReviewRequestID string
	OccurredAt      time.Time
}

type Request struct {
	ID           string
	LearnerID    string
	SubmissionID string
	ExecutionID  string
}
