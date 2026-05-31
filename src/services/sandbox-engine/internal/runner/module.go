package runner

import (
	"os"
	"path/filepath"
	"strings"
)

const sandboxModule = `module sandbox-task

go 1.24.0

require (
	github.com/lib/pq v1.10.9
	github.com/redis/go-redis/v9 v9.18.0
)
`

func (r *Runner) ensureGoModule(tmpDir, code string) error {
	if !strings.Contains(code, "github.com/lib/pq") && !strings.Contains(code, "github.com/redis/go-redis") {
		return nil
	}

	return os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(sandboxModule), 0644)
}
