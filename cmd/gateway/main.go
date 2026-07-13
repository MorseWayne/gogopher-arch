package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/app/gateway"
	"github.com/MorseWayne/gogopher-arch/internal/platform/config"
)

func main() {
	if err := run(); err != nil {
		slog.Error("gateway stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	app, err := gateway.Build(ctx, cfg)
	if err != nil {
		return err
	}
	defer app.Close()
	server := &http.Server{Addr: cfg.ListenAddress, Handler: app.Handler, ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second}
	errorsChannel := make(chan error, 1)
	go func() {
		slog.Info("gateway listening", "address", cfg.ListenAddress, "learning_enabled", cfg.LearningSliceEnabled)
		errorsChannel <- server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdown); err != nil {
			return err
		}
		return nil
	case err := <-errorsChannel:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
