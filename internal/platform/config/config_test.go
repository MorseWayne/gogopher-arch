package config

import "testing"

func TestLoadUsesSafeLocalDefaults(t *testing.T) {
	for _, key := range []string{"APP_ENV", "LISTEN_ADDRESS", "DATABASE_URL", "LEARNING_SLICE_ENABLED", "LEARNING_CONTENT_DIR", "LEARNING_SESSION_TTL", "DB_MAX_OPEN_CONNECTIONS", "DB_MAX_IDLE_CONNECTIONS", "DB_CONNECTION_LIFETIME", "SANDBOX_ENDPOINT", "SANDBOX_RESPONSE_GRACE", "SANDBOX_RPC_DEADLINE", "EXECUTION_PERSIST_GRACE", "EXECUTION_WORKER_LEASE", "EXECUTION_WORKER_HEARTBEAT", "EXECUTION_WORKER_POLL", "EXECUTION_MAX_CLAIMS", "EXECUTION_WORKER_ID"} {
		t.Setenv(key, "")
	}
	config, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if config.ListenAddress != "127.0.0.1:8080" || config.LearningSliceEnabled {
		t.Fatalf("defaults = %#v", config)
	}
}

func TestLoadRejectsAmbiguousExecutionTiming(t *testing.T) {
	t.Setenv("SANDBOX_RPC_DEADLINE", "40s")
	t.Setenv("EXECUTION_PERSIST_GRACE", "5s")
	t.Setenv("EXECUTION_WORKER_LEASE", "45s")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil")
	}
}

func TestLoadRejectsEnabledSliceOutsideLocal(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("LEARNING_SLICE_ENABLED", "true")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil")
	}
}
