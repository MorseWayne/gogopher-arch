package definition

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type Kind string

const (
	KindCapability      Kind = "capability"
	KindActivity        Kind = "activity"
	KindTask            Kind = "task"
	KindReleaseManifest Kind = "release_manifest"
)

var schemaFiles = map[Kind]string{
	KindCapability:      "capability.schema.json",
	KindActivity:        "activity.schema.json",
	KindTask:            "task.schema.json",
	KindReleaseManifest: "release-manifest.schema.json",
}

type taskAssets struct {
	Files         []taskFile `json:"files"`
	EditablePaths []string   `json:"editable_paths"`
	ReadonlyPaths []string   `json:"readonly_paths"`
	VisibleTests  []string   `json:"visible_tests"`
	HeldOutTests  []string   `json:"held_out_tests"`
	RaceTests     []string   `json:"race_tests"`
}

type taskFile struct {
	Source   string `json:"source"`
	Path     string `json:"path"`
	Role     string `json:"role"`
	Editable bool   `json:"editable"`
	SHA256   string `json:"sha256"`
}

func ValidateTaskAssets(taskDir string, document []byte) error {
	var task taskAssets
	if err := json.Unmarshal(document, &task); err != nil {
		return fmt.Errorf("parse task assets: %w", err)
	}

	declaredSources := make(map[string]taskFile, len(task.Files))
	declaredPaths := make(map[string]taskFile, len(task.Files))
	for _, file := range task.Files {
		if _, exists := declaredSources[file.Source]; exists {
			return fmt.Errorf("duplicate asset source %q", file.Source)
		}
		if _, exists := declaredPaths[file.Path]; exists {
			return fmt.Errorf("duplicate workspace path %q", file.Path)
		}
		contents, err := os.ReadFile(filepath.Join(taskDir, filepath.FromSlash(file.Source)))
		if err != nil {
			return fmt.Errorf("read declared asset %q: %w", file.Source, err)
		}
		digest := sha256.Sum256(contents)
		if actual := hex.EncodeToString(digest[:]); actual != file.SHA256 {
			return fmt.Errorf("asset %q hash mismatch: definition=%s actual=%s", file.Source, file.SHA256, actual)
		}
		declaredSources[file.Source] = file
		declaredPaths[file.Path] = file
	}

	if err := filepath.WalkDir(taskDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("task assets may not contain symlink %q", path)
		}
		if entry.IsDir() || entry.Name() == "task.json" || (path == filepath.Join(taskDir, "go.mod")) {
			return nil
		}
		relative, err := filepath.Rel(taskDir, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if _, exists := declaredSources[relative]; !exists {
			return fmt.Errorf("undeclared task asset %q", relative)
		}
		return nil
	}); err != nil {
		return err
	}

	for _, path := range task.EditablePaths {
		file, exists := declaredPaths[path]
		if !exists || !file.Editable {
			return fmt.Errorf("editable path %q does not reference an editable file", path)
		}
	}
	for _, path := range task.ReadonlyPaths {
		file, exists := declaredPaths[path]
		if !exists || file.Editable || file.Role == "held_out_test" || file.Role == "race_test" {
			return fmt.Errorf("readonly path %q does not reference a public readonly file", path)
		}
	}
	for _, path := range task.VisibleTests {
		if file, exists := declaredPaths[path]; !exists || file.Role != "visible_test" {
			return fmt.Errorf("visible test %q has no matching asset", path)
		}
	}
	for _, path := range task.HeldOutTests {
		if file, exists := declaredPaths[path]; !exists || file.Role != "held_out_test" {
			return fmt.Errorf("held-out test %q has no matching asset", path)
		}
	}
	for _, path := range task.RaceTests {
		if file, exists := declaredPaths[path]; !exists || file.Role != "race_test" {
			return fmt.Errorf("race test %q has no matching asset", path)
		}
	}
	return nil
}

type Validator struct {
	schemas map[Kind]*jsonschema.Schema
}

func NewValidator(source fs.FS) (*Validator, error) {
	if source == nil {
		return nil, fmt.Errorf("schema source is required")
	}

	compiled := make(map[Kind]*jsonschema.Schema, len(schemaFiles))
	for kind, filename := range schemaFiles {
		contents, err := fs.ReadFile(source, filename)
		if err != nil {
			return nil, fmt.Errorf("read %s schema: %w", kind, err)
		}
		document, err := jsonschema.UnmarshalJSON(bytes.NewReader(contents))
		if err != nil {
			return nil, fmt.Errorf("parse %s schema: %w", kind, err)
		}

		compiler := jsonschema.NewCompiler()
		if err := compiler.AddResource(filename, document); err != nil {
			return nil, fmt.Errorf("register %s schema: %w", kind, err)
		}
		schema, err := compiler.Compile(filename)
		if err != nil {
			return nil, fmt.Errorf("compile %s schema: %w", kind, err)
		}
		compiled[kind] = schema
	}

	return &Validator{schemas: compiled}, nil
}

func (v *Validator) Validate(kind Kind, document []byte) error {
	schema, ok := v.schemas[kind]
	if !ok {
		return fmt.Errorf("unknown definition kind %q", kind)
	}
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(document))
	if err != nil {
		return fmt.Errorf("parse %s definition: %w", kind, err)
	}
	if err := schema.Validate(instance); err != nil {
		return fmt.Errorf("validate %s definition: %w", kind, err)
	}
	return nil
}
