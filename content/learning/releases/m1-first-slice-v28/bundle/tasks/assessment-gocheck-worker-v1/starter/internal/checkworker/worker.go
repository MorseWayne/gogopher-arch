package checkworker

import (
	"context"
	"errors"
	"time"
)

type Task struct {
	ID        string
	Key       string
	Attempt   int
	Duplicate bool
}

type Store interface {
	Claim(context.Context, string, time.Time, time.Duration) (Task, bool, error)
	Ack(context.Context, string, string) error
	Retry(context.Context, string, string, time.Time) error
	Fail(context.Context, string, string) error
}

type Processor interface {
	Process(context.Context, Task) error
}

type Options struct {
	Owner          string
	Concurrency    int
	Lease          time.Duration
	ProcessTimeout time.Duration
	PollInterval   time.Duration
	MaxAttempts    int
	RetryDelay     func(int) time.Duration
	Now            func() time.Time
}

type Worker struct {
	store     Store
	processor Processor
	options   Options
}

func New(store Store, processor Processor, options Options) (*Worker, error) {
	// TODO: validate every dependency and timing boundary.
	return nil, errors.New("TODO")
}

func (worker *Worker) Run(ctx context.Context) error {
	// TODO: start exactly options.Concurrency loops and join them before returning.
	return errors.New("TODO")
}

func (worker *Worker) RunOnce(ctx context.Context) (bool, error) {
	// TODO: claim, suppress duplicates, process, then ack/retry/fail.
	return false, errors.New("TODO")
}

type temporary interface{ Temporary() bool }
