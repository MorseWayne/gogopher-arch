package attemptview

import (
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/assistance"
	"github.com/MorseWayne/gogopher-arch/internal/learning/evaluation"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

type Submission struct {
	ID                    string
	WorkspaceRevision     int64
	WorkspaceHash         string
	RuleSetHash           string
	AssistanceCutoff      int64
	Status                submission.Status
	LatestExecutionID     string
	LatestExecutionSeq    int
	LatestExecutionStatus execution.ExecutionStatus
	CreatedAt             time.Time
	EvaluatedAt           *time.Time
}

type Related struct {
	Submission  *Submission
	Assistance  []assistance.Event
	Executions  []execution.Execution
	RuleResults []execution.RuleResult
	Evidence    []evaluation.Evidence
}
