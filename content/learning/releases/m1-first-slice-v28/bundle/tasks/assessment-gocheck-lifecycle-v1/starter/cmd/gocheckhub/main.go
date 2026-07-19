package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"time"
)

type Config struct {
	Address         string
	DSN             string
	ShutdownTimeout time.Duration
}

type Database interface{ Close() error }

type Server interface {
	Serve() error
	Shutdown(context.Context) error
}

type Dependencies struct {
	OpenDatabase func(context.Context, string) (Database, error)
	BuildHandler func(Database) (http.Handler, error)
	NewServer    func(string, http.Handler) (Server, error)
}

func run(ctx context.Context, config Config, dependencies Dependencies) error {
	return errors.New("TODO: initialize and own dependencies")
}

func main() {
	_, stop := signal.NotifyContext(context.Background())
	defer stop()
}
