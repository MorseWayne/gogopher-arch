package attempt

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

func WorkspaceHash(files map[string]string) string {
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	hash := sha256.New()
	for _, path := range paths {
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(len(path)))
		hash.Write(size[:])
		hash.Write([]byte(path))
		binary.BigEndian.PutUint64(size[:], uint64(len(files[path])))
		hash.Write(size[:])
		hash.Write([]byte(files[path]))
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func ValidateWorkspace(view definition.TaskView, baseline, files map[string]string) error {
	allowed := make(map[string]bool, len(view.EditablePaths)+len(view.ReadonlyPaths))
	for _, path := range view.EditablePaths {
		allowed[path] = true
	}
	for _, path := range view.ReadonlyPaths {
		allowed[path] = false
	}
	if len(files) != len(allowed) || len(files) > view.WorkspaceLimits.MaxFiles {
		return fmt.Errorf("workspace must contain exactly %d public files", len(allowed))
	}
	total := 0
	for path, contents := range files {
		editable, exists := allowed[path]
		if !exists {
			return fmt.Errorf("workspace path %q is not allowed", path)
		}
		if len(contents) > view.WorkspaceLimits.MaxFileBytes {
			return fmt.Errorf("workspace path %q exceeds %d bytes", path, view.WorkspaceLimits.MaxFileBytes)
		}
		total += len(contents)
		if !editable && contents != baseline[path] {
			return fmt.Errorf("workspace path %q is readonly", path)
		}
	}
	if total > view.WorkspaceLimits.MaxTotalBytes {
		return fmt.Errorf("workspace exceeds %d total bytes", view.WorkspaceLimits.MaxTotalBytes)
	}
	return nil
}
