package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidatesConfigurationAndPreservesErrors(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.json")
	_, err := Load(missing)
	var pathError *os.PathError
	if !errors.As(err, &pathError) {
		t.Fatalf("Load(missing) error = %v, want wrapped *os.PathError", err)
	}

	tests := []struct {
		name string
		body string
	}{
		{name: "empty", body: `{"targets":[]}`},
		{name: "blank name", body: `{"targets":[{"name":"","url":"https://api.example"}]}`},
		{name: "duplicate", body: `{"targets":[{"name":"api","url":"https://a.example"},{"name":"api","url":"https://b.example"}]}`},
		{name: "scheme", body: `{"targets":[{"name":"api","url":"ftp://api.example"}]}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "targets.json")
			if err := os.WriteFile(path, []byte(test.body), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := Load(path); err == nil {
				t.Fatal("Load() error = nil")
			}
		})
	}
}
