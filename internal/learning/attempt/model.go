package attempt

import (
	"errors"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

var (
	ErrNotFound            = errors.New("learning attempt not found")
	ErrInactive            = errors.New("learning attempt is not active")
	ErrInvalidWorkspace    = errors.New("invalid workspace")
	ErrReviewClaimRequired = errors.New("review activity requires a ReviewItem claim")
)

type RevisionConflict struct {
	Revision int64
	Hash     string
}

func (e *RevisionConflict) Error() string { return "workspace revision conflict" }

type Attempt struct {
	ID                string
	LearnerID         string
	ReleaseID         string
	ActivityID        string
	ActivityVersion   int
	ActivityHash      string
	TaskID            string
	TaskVersion       int
	TaskHash          string
	CapabilityRefs    []definition.VersionedDefinitionRef
	Mode              string
	Status            string
	Workspace         map[string]string
	WorkspaceRevision int64
	WorkspaceHash     string
	StartedAt         time.Time
	UpdatedAt         time.Time
}

type CreateRecord struct {
	Attempt
}

// CreateResult distinguishes a newly-created Attempt from an existing open
// Attempt that was resumed for the same learner and frozen Activity.
type CreateResult struct {
	Attempt
	Created bool
}

type SaveRecord struct {
	AttemptID     string
	LearnerID     string
	BaseRevision  int64
	Workspace     map[string]string
	WorkspaceHash string
	UpdatedAt     time.Time
}

type CreateInput struct {
	LearnerID       string
	ActivityID      string
	ActivityVersion int
}

type SaveInput struct {
	LearnerID    string
	AttemptID    string
	BaseRevision int64
	Files        map[string]string
}
