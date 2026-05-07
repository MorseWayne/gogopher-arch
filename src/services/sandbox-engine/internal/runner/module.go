package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (r *Runner) ensureGoModule(tmpDir, code string) error {
	if !strings.Contains(code, "github.com/lib/pq") && !strings.Contains(code, "github.com/redis/go-redis") {
		return nil
	}
	for _, f := range []string{"go.mod", "go.sum"} {
		src := filepath.Join("/app/sandbox-module", f)
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("read %s: %w", src, err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, f), data, 0644); err != nil {
			return fmt.Errorf("write %s: %w", f, err)
		}
	}
	return nil
}