package gateway

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/MorseWayne/gogopher-arch/internal/learning/attempt"
	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/httpapi"
	"github.com/MorseWayne/gogopher-arch/internal/learning/session"
	"github.com/MorseWayne/gogopher-arch/internal/platform/config"
	"github.com/MorseWayne/gogopher-arch/internal/platform/database"
)

type App struct {
	Handler  http.Handler
	database *sql.DB
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
	return &App{Handler: httpapi.NewRouter(true, sessionHandler, attemptHandler, httpapi.NewDefinitionHandler(registry)), database: db}, nil
}

func (a *App) Close() error {
	if a.database != nil {
		return a.database.Close()
	}
	return nil
}
