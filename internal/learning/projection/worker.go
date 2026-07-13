package projection

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

const projectionConsumer = "capability_projector"

type RequestRepository interface {
	ClaimRequest(context.Context, string, time.Time, time.Duration) (Request, bool, error)
	CompleteRequest(context.Context, string, string, string, int, time.Time) error
	RetryRequest(context.Context, string, string, time.Time, time.Duration, int, string) (RetryResult, error)
}

type RequestProjector interface {
	RebuildRequest(context.Context, Request, time.Time) error
}

type WorkerObserver interface {
	ProjectionRetried(bool)
}

type WorkerOptions struct {
	Owner        string
	Lease        time.Duration
	PollInterval time.Duration
	MaxAttempts  int
	BaseBackoff  time.Duration
	MaxBackoff   time.Duration
	Now          func() time.Time
	Observer     WorkerObserver
}

type Worker struct {
	repository RequestRepository
	projector  RequestProjector
	options    WorkerOptions
}

func NewWorker(repository RequestRepository, projector RequestProjector, options WorkerOptions) (*Worker, error) {
	if repository == nil || projector == nil || options.Owner == "" || options.Lease <= 0 ||
		options.PollInterval <= 0 || options.MaxAttempts < 1 || options.BaseBackoff <= 0 ||
		options.MaxBackoff < options.BaseBackoff {
		return nil, fmt.Errorf("projection worker dependencies, identity, lease, poll, attempts, and backoff are required")
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Worker{repository: repository, projector: projector, options: options}, nil
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
	if err := w.projector.RebuildRequest(ctx, request, now); err != nil {
		delay := retryDelay(request.AttemptCount, w.options.BaseBackoff, w.options.MaxBackoff)
		result, retryErr := w.repository.RetryRequest(
			ctx, request.ID, w.options.Owner, w.options.Now().UTC(), delay,
			w.options.MaxAttempts, err.Error(),
		)
		if retryErr != nil {
			return true, fmt.Errorf("persist projection retry after %v: %w", err, retryErr)
		}
		if w.options.Observer != nil {
			w.options.Observer.ProjectionRetried(result.Exhausted)
		}
		slog.Warn("capability projection request failed",
			"request_id", request.ID, "attempt_count", result.AttemptCount,
			"retry_exhausted", result.Exhausted)
		return true, nil
	}
	if err := w.repository.CompleteRequest(
		ctx, request.ID, w.options.Owner, projectionConsumer,
		ProjectionConsumerVersion, w.options.Now().UTC(),
	); err != nil {
		return true, err
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
