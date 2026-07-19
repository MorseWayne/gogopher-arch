package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsAValidTargetFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "targets.json")
	if err := os.WriteFile(path, []byte(`{"targets":[{"name":"api","url":"https://api.example/health"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Targets) != 1 || loaded.Targets[0].Name != "api" || loaded.Targets[0].URL != "https://api.example/health" {
		t.Fatalf("Load() = %#v", loaded)
	}
}
