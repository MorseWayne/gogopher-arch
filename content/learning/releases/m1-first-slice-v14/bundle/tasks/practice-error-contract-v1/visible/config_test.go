package checkcfg

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPreservesPathError(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("Load() error = %v, want wrapped *os.PathError", err)
	}
}
