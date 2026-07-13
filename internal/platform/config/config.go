package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv                string
	ListenAddress         string
	DatabaseURL           string
	LearningSliceEnabled  bool
	LearningContentDir    string
	SessionTTL            time.Duration
	DBMaxOpenConnections  int
	DBMaxIdleConnections  int
	DBConnectionLifetime  time.Duration
	SandboxEndpoint       string
	SandboxResponseGrace  time.Duration
	SandboxRPCDeadline    time.Duration
	ExecutionPersistGrace time.Duration
	ExecutionLease        time.Duration
	ExecutionHeartbeat    time.Duration
	ExecutionPoll         time.Duration
	ExecutionMaxClaims    int
	ExecutionWorkerID     string
	ProjectionLease       time.Duration
	ProjectionPoll        time.Duration
	ProjectionMaxAttempts int
	ProjectionBaseBackoff time.Duration
	ProjectionMaxBackoff  time.Duration
}

func Load() (Config, error) {
	config := Config{
		AppEnv: env("APP_ENV", "local"), ListenAddress: env("LISTEN_ADDRESS", "127.0.0.1:8080"),
		DatabaseURL:        env("DATABASE_URL", "postgres://user:pass@localhost:5432/gogopher?sslmode=disable"),
		LearningContentDir: env("LEARNING_CONTENT_DIR", "content/learning"),
		SandboxEndpoint:    env("SANDBOX_ENDPOINT", "http://127.0.0.1:8081/v1/executions"),
		ExecutionWorkerID:  env("EXECUTION_WORKER_ID", defaultWorkerID()),
	}
	var err error
	if config.LearningSliceEnabled, err = strconv.ParseBool(env("LEARNING_SLICE_ENABLED", "false")); err != nil {
		return Config{}, fmt.Errorf("parse LEARNING_SLICE_ENABLED: %w", err)
	}
	if config.SessionTTL, err = time.ParseDuration(env("LEARNING_SESSION_TTL", "720h")); err != nil {
		return Config{}, fmt.Errorf("parse LEARNING_SESSION_TTL: %w", err)
	}
	if config.DBConnectionLifetime, err = time.ParseDuration(env("DB_CONNECTION_LIFETIME", "30m")); err != nil {
		return Config{}, fmt.Errorf("parse DB_CONNECTION_LIFETIME: %w", err)
	}
	if config.SandboxResponseGrace, err = time.ParseDuration(env("SANDBOX_RESPONSE_GRACE", "2s")); err != nil {
		return Config{}, fmt.Errorf("parse SANDBOX_RESPONSE_GRACE: %w", err)
	}
	if config.SandboxRPCDeadline, err = time.ParseDuration(env("SANDBOX_RPC_DEADLINE", "35s")); err != nil {
		return Config{}, fmt.Errorf("parse SANDBOX_RPC_DEADLINE: %w", err)
	}
	if config.ExecutionPersistGrace, err = time.ParseDuration(env("EXECUTION_PERSIST_GRACE", "5s")); err != nil {
		return Config{}, fmt.Errorf("parse EXECUTION_PERSIST_GRACE: %w", err)
	}
	if config.ExecutionLease, err = time.ParseDuration(env("EXECUTION_WORKER_LEASE", "45s")); err != nil {
		return Config{}, fmt.Errorf("parse EXECUTION_WORKER_LEASE: %w", err)
	}
	if config.ExecutionHeartbeat, err = time.ParseDuration(env("EXECUTION_WORKER_HEARTBEAT", "10s")); err != nil {
		return Config{}, fmt.Errorf("parse EXECUTION_WORKER_HEARTBEAT: %w", err)
	}
	if config.ExecutionPoll, err = time.ParseDuration(env("EXECUTION_WORKER_POLL", "250ms")); err != nil {
		return Config{}, fmt.Errorf("parse EXECUTION_WORKER_POLL: %w", err)
	}
	if config.ProjectionLease, err = time.ParseDuration(env("PROJECTION_WORKER_LEASE", "30s")); err != nil {
		return Config{}, fmt.Errorf("parse PROJECTION_WORKER_LEASE: %w", err)
	}
	if config.ProjectionPoll, err = time.ParseDuration(env("PROJECTION_WORKER_POLL", "250ms")); err != nil {
		return Config{}, fmt.Errorf("parse PROJECTION_WORKER_POLL: %w", err)
	}
	if config.ProjectionBaseBackoff, err = time.ParseDuration(env("PROJECTION_WORKER_BASE_BACKOFF", "1s")); err != nil {
		return Config{}, fmt.Errorf("parse PROJECTION_WORKER_BASE_BACKOFF: %w", err)
	}
	if config.ProjectionMaxBackoff, err = time.ParseDuration(env("PROJECTION_WORKER_MAX_BACKOFF", "1m")); err != nil {
		return Config{}, fmt.Errorf("parse PROJECTION_WORKER_MAX_BACKOFF: %w", err)
	}
	if config.DBMaxOpenConnections, err = parsePositiveInt("DB_MAX_OPEN_CONNECTIONS", 10); err != nil {
		return Config{}, err
	}
	if config.DBMaxIdleConnections, err = parsePositiveInt("DB_MAX_IDLE_CONNECTIONS", 5); err != nil {
		return Config{}, err
	}
	if config.ExecutionMaxClaims, err = parsePositiveInt("EXECUTION_MAX_CLAIMS", 3); err != nil {
		return Config{}, err
	}
	if config.ProjectionMaxAttempts, err = parsePositiveInt("PROJECTION_WORKER_MAX_ATTEMPTS", 5); err != nil {
		return Config{}, err
	}
	if config.SessionTTL <= 0 || config.DBConnectionLifetime <= 0 || config.SandboxResponseGrace <= 0 ||
		config.SandboxRPCDeadline <= 0 || config.ExecutionPersistGrace <= 0 || config.ExecutionLease <= 0 ||
		config.ExecutionHeartbeat <= 0 || config.ExecutionPoll <= 0 || config.ProjectionLease <= 0 ||
		config.ProjectionPoll <= 0 || config.ProjectionBaseBackoff <= 0 || config.ProjectionMaxBackoff <= 0 {
		return Config{}, fmt.Errorf("session, database, Sandbox, and worker durations must be positive")
	}
	if config.ProjectionMaxBackoff < config.ProjectionBaseBackoff {
		return Config{}, fmt.Errorf("PROJECTION_WORKER_MAX_BACKOFF must be at least PROJECTION_WORKER_BASE_BACKOFF")
	}
	if config.DBMaxIdleConnections > config.DBMaxOpenConnections {
		return Config{}, fmt.Errorf("DB_MAX_IDLE_CONNECTIONS may not exceed DB_MAX_OPEN_CONNECTIONS")
	}
	if _, _, err := net.SplitHostPort(config.ListenAddress); err != nil {
		return Config{}, fmt.Errorf("invalid LISTEN_ADDRESS: %w", err)
	}
	if config.LearningSliceEnabled && config.AppEnv != "local" {
		return Config{}, fmt.Errorf("Learning slice may only be enabled when APP_ENV=local")
	}
	if config.SandboxRPCDeadline+config.ExecutionPersistGrace >= config.ExecutionLease {
		return Config{}, fmt.Errorf("EXECUTION_WORKER_LEASE must exceed SANDBOX_RPC_DEADLINE plus EXECUTION_PERSIST_GRACE")
	}
	if config.ExecutionHeartbeat >= config.ExecutionLease/2 {
		return Config{}, fmt.Errorf("EXECUTION_WORKER_HEARTBEAT must be less than half EXECUTION_WORKER_LEASE")
	}
	return config, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func parsePositiveInt(key string, fallback int) (int, error) {
	value, err := strconv.Atoi(env(key, strconv.Itoa(fallback)))
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return value, nil
}

func defaultWorkerID() string {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "localhost"
	}
	return fmt.Sprintf("%s-%d", hostname, os.Getpid())
}
