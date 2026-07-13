package gateway

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
	"github.com/MorseWayne/gogopher-arch/internal/learning/httpapi"
	"github.com/MorseWayne/gogopher-arch/internal/learning/session"
	"github.com/MorseWayne/gogopher-arch/internal/platform/config"
	"github.com/MorseWayne/gogopher-arch/internal/platform/database"
)

type App struct {
	Handler          http.Handler
	database         *sql.DB
	executionService *execution.Service
	workerCancel     context.CancelFunc
	workerDone       chan error
	closeOnce        sync.Once
	closeError       error
}

func Build(ctx context.Context, cfg config.Config) (*App, error) {
	if !cfg.LearningSliceEnabled {
		return &App{Handler: httpapi.NewRouter(false, nil, nil, nil)}, nil
	}
	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.DBMaxOpenConnections)
	db.SetMaxIdleConns(cfg.DBMaxIdleConnections)
	db.SetConnMaxLifetime(cfg.DBConnectionLifetime)
	fail := func(err error) (*App, error) { db.Close(); return nil, err }
	history, err := definition.NewReleaseStore(db, definition.ReleaseStoreOptions{})
	if err != nil {
		return fail(err)
	}
	registry, err := definition.BootstrapRegistry(ctx, cfg.LearningContentDir, history)
	if err != nil {
		return fail(err)
	}
	sessionRepository, err := session.NewPostgresRepository(db, session.RepositoryOptions{})
	if err != nil {
		return fail(err)
	}
	sessionService, err := session.NewService(sessionRepository, session.ServiceOptions{TTL: cfg.SessionTTL})
	if err != nil {
		return fail(err)
	}
	sessionHandler, err := httpapi.NewSessionHandler(sessionService, httpapi.SessionHandlerOptions{SecureCookie: false})
	if err != nil {
		return fail(err)
	}
	attemptRepository, err := attempt.NewPostgresRepository(db, attempt.RepositoryOptions{})
	if err != nil {
		return fail(err)
	}
	attemptService, err := attempt.NewService(attemptRepository, registry, attempt.ServiceOptions{})
	if err != nil {
		return fail(err)
	}
	attemptHandler, err := httpapi.NewAttemptHandler(attemptService)
	if err != nil {
		return fail(err)
	}
	if registry.CurrentReleaseID() == "" {
		return fail(fmt.Errorf("current release is required"))
	}
	specBuilder, err := execution.NewSpecBuilder(registry)
	if err != nil {
		return fail(err)
	}
	executionRepository, err := execution.NewPostgresRepository(db, execution.RepositoryOptions{})
	if err != nil {
		return fail(err)
	}
	executionService, err := execution.NewService(executionRepository, attemptService, specBuilder, execution.ServiceOptions{})
	if err != nil {
		return fail(err)
	}
	sandboxClient, err := execution.NewSandboxClient(execution.SandboxClientOptions{Endpoint: cfg.SandboxEndpoint})
	if err != nil {
		return fail(err)
	}
	maximumActionTimeout, err := registry.MaximumActionTimeout()
	if err != nil {
		return fail(err)
	}
	worker, err := execution.NewWorker(executionRepository, sandboxClient, execution.WorkerOptions{
		Owner: cfg.ExecutionWorkerID, MaxActionTimeout: maximumActionTimeout,
		SandboxResponseGrace: cfg.SandboxResponseGrace, RPCDeadline: cfg.SandboxRPCDeadline,
		PersistenceGrace: cfg.ExecutionPersistGrace, LeaseDuration: cfg.ExecutionLease,
		HeartbeatInterval: cfg.ExecutionHeartbeat, PollInterval: cfg.ExecutionPoll,
		MaxClaims: cfg.ExecutionMaxClaims,
	})
	if err != nil {
		return fail(err)
	}
	workerContext, workerCancel := context.WithCancel(ctx)
	workerDone := make(chan error, 1)
	go func() {
		for {
			err := worker.Run(workerContext)
			if errors.Is(err, context.Canceled) || workerContext.Err() != nil {
				workerDone <- workerContext.Err()
				return
			}
			slog.Error("execution worker restarting", "error", err)
			timer := time.NewTimer(time.Second)
			select {
			case <-workerContext.Done():
				timer.Stop()
				workerDone <- workerContext.Err()
				return
			case <-timer.C:
			}
		}
	}()
	return &App{
		Handler:  httpapi.NewRouter(true, sessionHandler, attemptHandler, httpapi.NewDefinitionHandler(registry)),
		database: db, executionService: executionService,
		workerCancel: workerCancel, workerDone: workerDone,
	}, nil
}

func (a *App) Close() error {
	a.closeOnce.Do(func() {
		if a.workerCancel != nil {
			a.workerCancel()
			if err := <-a.workerDone; err != nil && !errors.Is(err, context.Canceled) {
				a.closeError = err
			}
		}
		if a.database != nil {
			a.closeError = errors.Join(a.closeError, a.database.Close())
		}
	})
	return a.closeError
}
