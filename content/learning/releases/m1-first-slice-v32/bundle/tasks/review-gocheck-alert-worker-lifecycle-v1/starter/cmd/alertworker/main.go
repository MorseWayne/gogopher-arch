package main

import (
	"context"
	"errors"
	"os/signal"
	"time"
)

type Config struct {
	DSN             string
	Concurrency     int
	ShutdownTimeout time.Duration
}

type Store interface{ Close() error }
type Worker interface {
	Run() error
	Shutdown(context.Context) error
}
type Dependencies struct {
	OpenStore func(context.Context, string) (Store, error)
	NewWorker func(Store, int) (Worker, error)
}

func run(ctx context.Context, config Config, dependencies Dependencies) error {
	return errors.New("TODO: initialize and own dependencies")
}

func main() {
	_, stop := signal.NotifyContext(context.Background())
	defer stop()
}
