package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv               string
	ListenAddress        string
	DatabaseURL          string
	LearningSliceEnabled bool
	LearningContentDir   string
	SessionTTL           time.Duration
	DBMaxOpenConnections int
	DBMaxIdleConnections int
	DBConnectionLifetime time.Duration
}

func Load() (Config, error) {
	config := Config{
		AppEnv: env("APP_ENV", "local"), ListenAddress: env("LISTEN_ADDRESS", "127.0.0.1:8080"),
		DatabaseURL:        env("DATABASE_URL", "postgres://user:pass@localhost:5432/gogopher?sslmode=disable"),
		LearningContentDir: env("LEARNING_CONTENT_DIR", "content/learning"),
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
	if config.DBMaxOpenConnections, err = parsePositiveInt("DB_MAX_OPEN_CONNECTIONS", 10); err != nil {
		return Config{}, err
	}
	if config.DBMaxIdleConnections, err = parsePositiveInt("DB_MAX_IDLE_CONNECTIONS", 5); err != nil {
		return Config{}, err
	}
	if config.SessionTTL <= 0 || config.DBConnectionLifetime <= 0 {
		return Config{}, fmt.Errorf("session TTL and DB connection lifetime must be positive")
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
