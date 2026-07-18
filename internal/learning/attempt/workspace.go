package attempt

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"path"
	"regexp"
	"sort"
	"strings"

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
	if len(files) > view.WorkspaceLimits.MaxFiles {
		return fmt.Errorf("workspace may contain at most %d files", view.WorkspaceLimits.MaxFiles)
	}
	for _, filePath := range view.ReadonlyPaths {
		contents, exists := files[filePath]
		if !exists || contents != baseline[filePath] {
			return fmt.Errorf("workspace path %q is readonly", filePath)
		}
	}
	if !view.WorkspacePolicy.AllowDeleteFiles {
		for _, filePath := range view.EditablePaths {
			if _, exists := files[filePath]; !exists {
				return fmt.Errorf("workspace path %q is required", filePath)
			}
		}
	}
	total := 0
	for filePath, contents := range files {
		editable, exists := allowed[filePath]
		if !exists {
			if !view.WorkspacePolicy.AllowNewFiles || !validLearnerPath(filePath) {
				return fmt.Errorf("workspace path %q is not allowed", filePath)
			}
			editable = true
		}
		if len(contents) > view.WorkspaceLimits.MaxFileBytes {
			return fmt.Errorf("workspace path %q exceeds %d bytes", filePath, view.WorkspaceLimits.MaxFileBytes)
		}
		total += len(contents)
		if !editable && contents != baseline[filePath] {
			return fmt.Errorf("workspace path %q is readonly", filePath)
		}
	}
	if total > view.WorkspaceLimits.MaxTotalBytes {
		return fmt.Errorf("workspace exceeds %d total bytes", view.WorkspaceLimits.MaxTotalBytes)
	}
	return nil
}

var learnerPathPattern = regexp.MustCompile(`^(?:[A-Za-z0-9._-]+/)*[A-Za-z0-9._-]+$`)

func validLearnerPath(value string) bool {
	if value == "" || !learnerPathPattern.MatchString(value) || strings.ContainsAny(value, "\\\x00") || strings.HasPrefix(value, "/") || path.Clean(value) != value || value == "." {
		return false
	}
	for _, segment := range strings.Split(value, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
	}
	return true
}
