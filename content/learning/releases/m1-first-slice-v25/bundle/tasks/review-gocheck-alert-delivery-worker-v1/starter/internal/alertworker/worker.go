package alertworker

import (
	"context"
	"errors"
	"time"
)

type Delivery struct {
	ID, IdempotencyKey string
	Attempt            int
	Duplicate          bool
}
type Queue interface {
	Claim(context.Context, string, time.Time, time.Duration) (Delivery, bool, error)
	Ack(context.Context, string, string) error
	Retry(context.Context, string, string, time.Time) error
	Fail(context.Context, string, string) error
}
type Sender interface {
	Send(context.Context, Delivery) error
}
type Options struct {
	Owner                            string
	Concurrency                      int
	Lease, SendTimeout, PollInterval time.Duration
	MaxAttempts                      int
	RetryDelay                       func(int) time.Duration
	Now                              func() time.Time
}
type Worker struct {
	queue   Queue
	sender  Sender
	options Options
}

func New(queue Queue, sender Sender, options Options) (*Worker, error) {
	return nil, errors.New("TODO")
}
func (worker *Worker) Run(ctx context.Context) error             { return errors.New("TODO") }
func (worker *Worker) RunOnce(ctx context.Context) (bool, error) { return false, errors.New("TODO") }

type temporary interface{ Temporary() bool }
