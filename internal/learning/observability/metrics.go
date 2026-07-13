package observability

import (
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/evaluation"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

type executionKey struct {
	Action execution.Action
	Status execution.ExecutionStatus
}

type truncationKey struct {
	Action execution.Action
	Stage  execution.Stage
}

type evidenceKey struct {
	Type         string
	Independence string
	Result       execution.RuleStatus
}

type executionMetrics struct {
	Count      uint64
	DurationMS uint64
}

type Collector struct {
	mu          sync.Mutex
	attempts    map[string]uint64
	executions  map[executionKey]executionMetrics
	failures    map[string]uint64
	truncations map[truncationKey]uint64
	evidence    map[evidenceKey]uint64
	projection  map[string]uint64
}

func NewCollector() *Collector {
	return &Collector{
		attempts: make(map[string]uint64), executions: make(map[executionKey]executionMetrics),
		failures: make(map[string]uint64), truncations: make(map[truncationKey]uint64),
		evidence: make(map[evidenceKey]uint64), projection: make(map[string]uint64),
	}
}

func (c *Collector) ProjectionRetried(exhausted bool) {
	outcome := "scheduled"
	if exhausted {
		outcome = "exhausted"
	}
	c.mu.Lock()
	c.projection[outcome]++
	c.mu.Unlock()
}

func (c *Collector) AttemptCreated(value attempt.Attempt) {
	c.mu.Lock()
	c.attempts["created"]++
	c.mu.Unlock()
	slog.Info("learning attempt created",
		"learner_id", value.LearnerID, "attempt_id", value.ID,
		"activity_id", value.ActivityID, "task_id", value.TaskID)
}

func (c *Collector) SubmissionQueued(result submission.Result) {
	if !result.Created {
		return
	}
	c.mu.Lock()
	c.attempts["submitted"]++
	c.mu.Unlock()
	slog.Info("learning submission queued",
		"learner_id", result.Submission.LearnerID, "attempt_id", result.Submission.AttemptID,
		"submission_id", result.Submission.ID, "execution_id", result.ExecutionID)
}

func (c *Collector) ExecutionCompleted(value execution.Execution, response execution.ExecutionResponse) {
	key := executionKey{Action: value.Action, Status: response.Status}
	duration := response.DurationMS
	if duration < 0 {
		duration = 0
	}
	timedOut, truncated := false, false
	c.mu.Lock()
	metrics := c.executions[key]
	metrics.Count++
	metrics.DurationMS += uint64(duration)
	c.executions[key] = metrics
	for _, stage := range response.Stages {
		if stage.TimedOut {
			timedOut = true
			c.failures["action_timeout"]++
		}
		if stage.OutputTruncated {
			truncated = true
			c.truncations[truncationKey{Action: value.Action, Stage: stage.Stage}]++
		}
	}
	failureCode := ""
	if response.Failure != nil {
		failureCode = response.Failure.Code
		c.failures[failureCode]++
	}
	if value.Action == execution.ActionSubmit && response.Status == execution.ExecutionInfraFailed {
		c.attempts["infra_failed"]++
	}
	c.mu.Unlock()
	slog.Info("learning execution completed",
		"attempt_id", value.AttemptID, "submission_id", value.SubmissionID,
		"execution_id", value.ID, "task_id", value.TaskID,
		"action", value.Action, "status", response.Status,
		"duration_ms", duration, "timed_out", timedOut, "output_truncated", truncated,
		"failure_code", failureCode)
}

func (c *Collector) EvaluationCompleted(batch evaluation.Batch) {
	c.mu.Lock()
	c.attempts["completed"]++
	for _, value := range batch.Evidence {
		c.evidence[evidenceKey{
			Type: value.EvidenceType, Independence: string(value.Independence), Result: value.Result,
		}]++
	}
	c.mu.Unlock()
	learnerID, attemptID := "", ""
	if len(batch.Evidence) > 0 {
		learnerID, attemptID = batch.Evidence[0].LearnerID, batch.Evidence[0].AttemptID
	}
	slog.Info("learning evaluation completed",
		"learner_id", learnerID, "attempt_id", attemptID, "submission_id", batch.SubmissionID,
		"execution_id", batch.ExecutionID, "evaluation_batch_id", batch.ID,
		"rule_result_count", len(batch.RuleResults), "evidence_count", len(batch.Evidence))
}

func (c *Collector) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	c.mu.Lock()
	lines := c.metricLines()
	c.mu.Unlock()
	sort.Strings(lines)
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprintln(w, strings.Join(lines, "\n"))
}

func (c *Collector) metricLines() []string {
	lines := make([]string, 0, len(c.attempts)+len(c.executions)*3+len(c.failures)+len(c.truncations)+len(c.evidence)+len(c.projection))
	for status, count := range c.attempts {
		lines = append(lines, fmt.Sprintf(`gogopher_learning_attempt_total{outcome=%q} %d`, status, count))
	}
	for key, metrics := range c.executions {
		labels := fmt.Sprintf(`action=%q,status=%q`, key.Action, key.Status)
		lines = append(lines,
			fmt.Sprintf(`gogopher_learning_execution_total{%s} %d`, labels, metrics.Count),
			fmt.Sprintf(`gogopher_learning_execution_duration_milliseconds_count{%s} %d`, labels, metrics.Count),
			fmt.Sprintf(`gogopher_learning_execution_duration_milliseconds_sum{%s} %d`, labels, metrics.DurationMS),
		)
	}
	for code, count := range c.failures {
		lines = append(lines, fmt.Sprintf(`gogopher_learning_execution_failure_total{code=%q} %d`, code, count))
	}
	for key, count := range c.truncations {
		lines = append(lines, fmt.Sprintf(`gogopher_learning_execution_output_truncated_total{action=%q,stage=%q} %d`, key.Action, key.Stage, count))
	}
	for key, count := range c.evidence {
		lines = append(lines, fmt.Sprintf(`gogopher_learning_evidence_total{type=%q,independence=%q,result=%q} %d`, key.Type, key.Independence, key.Result, count))
	}
	for outcome, count := range c.projection {
		lines = append(lines, fmt.Sprintf(`gogopher_learning_projection_retry_total{outcome=%q} %d`, outcome, count))
	}
	return lines
}
