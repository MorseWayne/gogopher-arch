package projection

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

const (
	ProjectionConsumer      = "capability_projector"
	ReviewSchedulerConsumer = "review_scheduler"
)

type RequestRepository interface {
	ClaimRequest(context.Context, string, time.Time, time.Duration) (Request, bool, error)
	CompleteRequest(context.Context, string, string, string, int, time.Time) error
	RetryRequest(context.Context, string, string, time.Time, time.Duration, int, string) (RetryResult, error)
}

type RequestProcessor interface {
	ProcessRequest(context.Context, Request, time.Time) error
}

type WorkerObserver interface {
	OutboxRetried(string, bool)
}

type workerCompletionObserver interface {
	OutboxCompleted(string, time.Duration)
}

type WorkerOptions struct {
	Owner           string
	Lease           time.Duration
	PollInterval    time.Duration
	MaxAttempts     int
	BaseBackoff     time.Duration
	MaxBackoff      time.Duration
	Consumer        string
	ConsumerVersion int
	Now             func() time.Time
	Observer        WorkerObserver
}

type Worker struct {
	repository RequestRepository
	processor  RequestProcessor
	options    WorkerOptions
}

func NewWorker(repository RequestRepository, processor RequestProcessor, options WorkerOptions) (*Worker, error) {
	if repository == nil || processor == nil || options.Owner == "" || options.Lease <= 0 ||
		options.PollInterval <= 0 || options.MaxAttempts < 1 || options.BaseBackoff <= 0 ||
		options.MaxBackoff < options.BaseBackoff || options.Consumer == "" || options.ConsumerVersion < 1 {
		return nil, fmt.Errorf("outbox worker dependencies, identity, lease, poll, attempts, consumer, and backoff are required")
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Worker{repository: repository, processor: processor, options: options}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		processed, err := w.RunOnce(ctx)
		if err != nil {
			return err
		}
		if processed {
			continue
		}
		timer := time.NewTimer(w.options.PollInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (w *Worker) RunOnce(ctx context.Context) (bool, error) {
	now := w.options.Now().UTC()
	request, ok, err := w.repository.ClaimRequest(ctx, w.options.Owner, now, w.options.Lease)
	if err != nil || !ok {
		return false, err
	}
	if err := w.processor.ProcessRequest(ctx, request, now); err != nil {
		delay := retryDelay(request.AttemptCount, w.options.BaseBackoff, w.options.MaxBackoff)
		result, retryErr := w.repository.RetryRequest(
			ctx, request.ID, w.options.Owner, w.options.Now().UTC(), delay,
			w.options.MaxAttempts, err.Error(),
		)
		if retryErr != nil {
			return true, fmt.Errorf("persist outbox retry after %v: %w", err, retryErr)
		}
		if w.options.Observer != nil {
			w.options.Observer.OutboxRetried(w.options.Consumer, result.Exhausted)
		}
		slog.Warn("learning outbox request failed",
			"request_id", request.ID, "attempt_count", result.AttemptCount,
			"consumer", w.options.Consumer, "retry_exhausted", result.Exhausted)
		return true, nil
	}
	completedAt := w.options.Now().UTC()
	if err := w.repository.CompleteRequest(
		ctx, request.ID, w.options.Owner, w.options.Consumer,
		w.options.ConsumerVersion, completedAt,
	); err != nil {
		return true, err
	}
	if observer, ok := w.options.Observer.(workerCompletionObserver); ok && !request.CreatedAt.IsZero() {
		lag := completedAt.Sub(request.CreatedAt)
		if lag < 0 {
			lag = 0
		}
		observer.OutboxCompleted(w.options.Consumer, lag)
	}
	return true, nil
}

func retryDelay(attempt int, base, maximum time.Duration) time.Duration {
	delay := base
	for step := 1; step < attempt && delay < maximum; step++ {
		if delay > maximum/2 {
			return maximum
		}
		delay *= 2
	}
	if delay > maximum {
		return maximum
	}
	return delay
}
