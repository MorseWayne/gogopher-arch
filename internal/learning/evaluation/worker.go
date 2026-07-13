package evaluation

import (
	"context"
	"fmt"
	"time"
)

type RequestRepository interface {
	ClaimRequest(context.Context, string, time.Time, time.Duration) (Request, bool, error)
	CompleteRequest(context.Context, string, string, time.Time) error
}

type Evaluator interface {
	Evaluate(context.Context, string, string, string) (Batch, bool, error)
}

type WorkerOptions struct {
	Owner        string
	Lease        time.Duration
	PollInterval time.Duration
	Now          func() time.Time
}

type Worker struct {
	repository RequestRepository
	evaluator  Evaluator
	options    WorkerOptions
}

func NewWorker(repository RequestRepository, evaluator Evaluator, options WorkerOptions) (*Worker, error) {
	if repository == nil || evaluator == nil || options.Owner == "" || options.Lease <= 0 || options.PollInterval <= 0 {
		return nil, fmt.Errorf("evaluation worker repository, evaluator, owner, lease, and poll interval are required")
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Worker{repository: repository, evaluator: evaluator, options: options}, nil
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
	request, ok, err := w.repository.ClaimRequest(
		ctx, w.options.Owner, w.options.Now().UTC(), w.options.Lease,
	)
	if err != nil || !ok {
		return false, err
	}
	if _, _, err := w.evaluator.Evaluate(ctx, request.LearnerID, request.SubmissionID, request.ExecutionID); err != nil {
		return true, err
	}
	if err := w.repository.CompleteRequest(ctx, request.ID, w.options.Owner, w.options.Now().UTC()); err != nil {
		return true, err
	}
	return true, nil
}
