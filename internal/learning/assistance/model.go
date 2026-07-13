package assistance

import (
	"encoding/json"
	"errors"
	"time"
)

type EventType string

const (
	HintRevealed    EventType = "hint_revealed"
	ReferenceOpened EventType = "reference_opened"
	SolutionViewed  EventType = "solution_viewed"
	AIDeclared      EventType = "ai_declared"
)

var (
	ErrAttemptNotFound = errors.New("assistance attempt not found")
	ErrAttemptInactive = errors.New("assistance attempt is no longer active")
	ErrEventNotAllowed = errors.New("assistance event is not allowed by the frozen activity")
)

type IdempotencyConflict struct{ EventID string }

func (e *IdempotencyConflict) Error() string {
	return "assistance event key conflicts with its original payload"
}

type Event struct {
	ID        string          `json:"id"`
	AttemptID string          `json:"attempt_id"`
	LearnerID string          `json:"-"`
	EventKey  string          `json:"event_key"`
	Sequence  int64           `json:"event_seq"`
	Type      EventType       `json:"event_type"`
	Payload   json.RawMessage `json:"payload"`
	CreatedAt time.Time       `json:"created_at"`
}

type Record struct {
	ID        string
	AttemptID string
	LearnerID string
	EventKey  string
	Type      EventType
	Payload   json.RawMessage
	CreatedAt time.Time
}

type RecordResult struct {
	Event   Event
	Created bool
}

type Independence string

const (
	IndependenceGuided      Independence = "guided"
	IndependenceHinted      Independence = "hinted"
	IndependenceReferenced  Independence = "referenced"
	IndependenceAIAssisted  Independence = "ai_assisted"
	IndependenceIndependent Independence = "independent"
)

type Hint struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}
