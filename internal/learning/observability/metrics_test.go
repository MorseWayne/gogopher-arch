package observability

import (
	"bytes"
	"context"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/assistance"
	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/evaluation"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/submission"
)

func TestCollectorExportsBoundedMetricsWithoutSensitivePayloads(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logs, nil)))
	defer slog.SetDefault(previous)

	collector := NewCollector(stateProviderStub{due: 3})
	collector.AttemptCreated(attempt.Attempt{
		ID: "attempt-id", LearnerID: "learner-id", ActivityID: "activity-id", TaskID: "task-id",
		Workspace: map[string]string{"main.go": "SECRET_USER_CODE"},
	})
	collector.SubmissionQueued(submission.Result{
		Created: true, ExecutionID: "submit-execution",
		Submission: submission.Submission{ID: "submission-id", AttemptID: "attempt-id", LearnerID: "learner-id"},
	})
	collector.ExecutionCompleted(execution.Execution{
		ID: "execution-id", AttemptID: "attempt-id", TaskID: "task-id", Action: execution.ActionTest,
		RequestKey: "SECRET_REQUEST_KEY",
	}, execution.ExecutionResponse{
		Status: execution.ExecutionUserFailed, DurationMS: 125,
		Stages: []execution.StageResult{{
			Stage: execution.StageVisibleTest, Status: execution.StageFailed,
			TimedOut: true, OutputTruncated: true, Stdout: "SECRET_EXECUTION_OUTPUT",
		}},
	})
	collector.ExecutionCompleted(execution.Execution{
		ID: "submit-execution", AttemptID: "attempt-id", SubmissionID: "submission-id",
		TaskID: "task-id", Action: execution.ActionSubmit,
	}, execution.ExecutionResponse{
		Status:  execution.ExecutionInfraFailed,
		Failure: &execution.Failure{Code: "sandbox_rpc_deadline", Message: "SECRET_INTERNAL_ADDRESS"},
	})
	collector.EvaluationCompleted(evaluation.Batch{
		ID: "batch-id", SubmissionID: "submission-id", ExecutionID: "submit-execution",
		Evidence: []evaluation.Evidence{{
			LearnerID: "learner-id", AttemptID: "attempt-id", EvidenceType: "test",
			Independence: assistance.IndependenceHinted, Result: execution.RulePassed,
			Reason: "SECRET_EVIDENCE_REASON",
		}},
	})
	collector.OutboxRetried("capability_projector", false)
	collector.OutboxRetried("review_scheduler", true)
	collector.OutboxCompleted("capability_projector", 2500*time.Millisecond)
	collector.ReviewItemsTransitioned("created", 2)
	collector.ReviewItemsTransitioned("claimed", 1)
	collector.ReviewItemsTransitioned("completed", 1)
	collector.ReviewItemsTransitioned("replaced", 1)

	response := httptest.NewRecorder()
	collector.ServeHTTP(response, httptest.NewRequest("GET", "/metrics", nil))
	body := response.Body.String()
	for _, metric := range []string{
		`gogopher_learning_attempt_total{outcome="created"} 1`,
		`gogopher_learning_attempt_total{outcome="submitted"} 1`,
		`gogopher_learning_attempt_total{outcome="completed"} 1`,
		`gogopher_learning_attempt_total{outcome="infra_failed"} 1`,
		`gogopher_learning_execution_duration_milliseconds_sum{action="test",status="user_failed"} 125`,
		`gogopher_learning_execution_failure_total{code="action_timeout"} 1`,
		`gogopher_learning_execution_failure_total{code="sandbox_rpc_deadline"} 1`,
		`gogopher_learning_execution_output_truncated_total{action="test",stage="visible_test"} 1`,
		`gogopher_learning_evidence_total{type="test",independence="hinted",result="passed"} 1`,
		`gogopher_learning_outbox_retry_total{consumer="capability_projector",outcome="scheduled"} 1`,
		`gogopher_learning_outbox_retry_total{consumer="review_scheduler",outcome="exhausted"} 1`,
		`gogopher_learning_projection_lag_seconds{consumer="capability_projector"} 2.5`,
		`gogopher_learning_review_transition_total{outcome="created"} 2`,
		`gogopher_learning_review_transition_total{outcome="claimed"} 1`,
		`gogopher_learning_review_transition_total{outcome="completed"} 1`,
		`gogopher_learning_review_transition_total{outcome="replaced"} 1`,
		`gogopher_learning_review_due 3`,
	} {
		if !strings.Contains(body, metric) {
			t.Fatalf("metrics missing %q:\n%s", metric, body)
		}
	}
	for _, secret := range []string{"SECRET_USER_CODE", "SECRET_REQUEST_KEY", "SECRET_EXECUTION_OUTPUT", "SECRET_INTERNAL_ADDRESS", "SECRET_EVIDENCE_REASON"} {
		if strings.Contains(logs.String(), secret) {
			t.Fatalf("logs contain sensitive value %q: %s", secret, logs.String())
		}
	}
}

type stateProviderStub struct{ due int }

func (s stateProviderStub) DueReviewCount(context.Context, time.Time) (int, error) { return s.due, nil }
