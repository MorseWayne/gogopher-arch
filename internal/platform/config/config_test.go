package config

import "testing"

func TestLoadUsesSafeLocalDefaults(t *testing.T) {
	for _, key := range []string{"APP_ENV", "LISTEN_ADDRESS", "DATABASE_URL", "LEARNING_SLICE_ENABLED", "LEARNING_CONTENT_DIR", "LEARNING_SESSION_TTL", "DB_MAX_OPEN_CONNECTIONS", "DB_MAX_IDLE_CONNECTIONS", "DB_CONNECTION_LIFETIME"} {
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

func TestLoadRejectsEnabledSliceOutsideLocal(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("LEARNING_SLICE_ENABLED", "true")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil")
	}
}
