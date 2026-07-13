package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRejectsInvalidSchemeAndPreservesPathErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "targets.json")
	if err := os.WriteFile(path, []byte(`{"targets":[{"name":"db","url":"file:///tmp/db","timeout_ms":0}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Fatal("Load(invalid) error = nil")
	}
	_, err := Load(filepath.Join(dir, "missing.json"))
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("Load(missing) error = %v, want wrapped *os.PathError", err)
	}
}
