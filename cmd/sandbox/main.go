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

	sandboxapp "github.com/MorseWayne/gogopher-arch/internal/app/sandbox"
	"github.com/MorseWayne/gogopher-arch/internal/platform/config"
)

func main() {
	if err := run(); err != nil {
		slog.Error("sandbox stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadSandbox()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	server := &http.Server{
		Addr: cfg.ListenAddress, Handler: sandboxapp.Build(),
		ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second,
	}
	errorsChannel := make(chan error, 1)
	go func() {
		slog.Info("sandbox listening", "address", cfg.ListenAddress, "network_enforcement", "policy_only")
		errorsChannel <- server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdown)
	case err := <-errorsChannel:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}
