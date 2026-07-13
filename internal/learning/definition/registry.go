package definition

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

var ErrDefinitionNotFound = errors.New("definition not found")

type DefinitionRef struct {
	ReleaseID string
	Kind      Kind
	ID        string
	Version   int
}

type Definition struct {
	DefinitionRef
	ContentHash string
	BundleHash  string
	RuleSetHash string
	Document    json.RawMessage
}

type RegistryOptions struct {
	ContentDir         string
	RequiredReleaseIDs []string
}

type Registry struct {
	contentDir  string
	current     string
	releases    map[string]ReleaseManifest
	definitions map[DefinitionRef]Definition
}

type RegistryHistory interface {
	ReferencedReleaseIDs(context.Context) ([]string, error)
	Register(context.Context, *Registry) error
}

func BootstrapRegistry(ctx context.Context, contentDir string, history RegistryHistory) (*Registry, error) {
	if history == nil {
		return nil, fmt.Errorf("definition history store is required")
	}
	required, err := history.ReferencedReleaseIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("load referenced definition releases: %w", err)
	}
	registry, err := LoadRegistry(RegistryOptions{ContentDir: contentDir, RequiredReleaseIDs: required})
	if err != nil {
		return nil, err
	}
	if err := history.Register(ctx, registry); err != nil {
		return nil, fmt.Errorf("register definition history: %w", err)
	}
	return registry, nil
}

func LoadRegistry(options RegistryOptions) (*Registry, error) {
	if options.ContentDir == "" {
		return nil, fmt.Errorf("learning content directory is required")
	}
	current, err := loadCurrentRelease(filepath.Join(options.ContentDir, "current-release.json"))
	if err != nil {
		return nil, err
	}
	releasesDir := filepath.Join(options.ContentDir, "releases")
	entries, err := os.ReadDir(releasesDir)
	if err != nil {
		return nil, fmt.Errorf("read releases directory: %w", err)
	}
	registry := &Registry{
		contentDir: options.ContentDir, current: current,
		releases: make(map[string]ReleaseManifest), definitions: make(map[DefinitionRef]Definition),
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("release directories may not be symlinks: %s", entry.Name())
		}
		releaseDir := filepath.Join(releasesDir, entry.Name())
		if err := VerifyRelease(releaseDir, filepath.Join(options.ContentDir, "schemas")); err != nil {
			return nil, fmt.Errorf("release %s: %w", entry.Name(), err)
		}
		manifest, err := readManifest(releaseDir)
		if err != nil {
			return nil, err
		}
		if manifest.ReleaseID != entry.Name() {
			return nil, fmt.Errorf("release directory %q contains manifest for %q", entry.Name(), manifest.ReleaseID)
		}
		if _, exists := registry.releases[manifest.ReleaseID]; exists {
			return nil, fmt.Errorf("duplicate release %q", manifest.ReleaseID)
		}
		registry.releases[manifest.ReleaseID] = manifest
		for _, metadata := range manifest.Definitions {
			ref := DefinitionRef{ReleaseID: manifest.ReleaseID, Kind: metadata.Kind, ID: metadata.ID, Version: metadata.Version}
			document, err := readBundleFile(filepath.Join(releaseDir, "bundle"), metadata.Path)
			if err != nil {
				return nil, err
			}
			registry.definitions[ref] = Definition{
				DefinitionRef: ref, ContentHash: metadata.ContentHash, BundleHash: metadata.BundleHash,
				RuleSetHash: metadata.RuleSetHash, Document: append(json.RawMessage(nil), document...),
			}
		}
	}
	if len(registry.releases) == 0 {
		return nil, fmt.Errorf("no learning releases found")
	}
	if _, exists := registry.releases[current]; !exists {
		return nil, fmt.Errorf("current release %q is unavailable", current)
	}
	for _, required := range options.RequiredReleaseIDs {
		if _, exists := registry.releases[required]; !exists {
			return nil, fmt.Errorf("required historical release %q is unavailable", required)
		}
	}
	return registry, nil
}

func (r *Registry) CurrentReleaseID() string {
	return r.current
}

func (r *Registry) ReleaseIDs() []string {
	result := make([]string, 0, len(r.releases))
	for releaseID := range r.releases {
		result = append(result, releaseID)
	}
	sort.Strings(result)
	return result
}

func (r *Registry) Manifest(releaseID string) (ReleaseManifest, error) {
	manifest, exists := r.releases[releaseID]
	if !exists {
		return ReleaseManifest{}, fmt.Errorf("release %q: %w", releaseID, ErrDefinitionNotFound)
	}
	return cloneManifest(manifest), nil
}

func (r *Registry) Get(ref DefinitionRef) (Definition, error) {
	definition, exists := r.definitions[ref]
	if !exists {
		return Definition{}, fmt.Errorf("release %q %s %s version %d: %w", ref.ReleaseID, ref.Kind, ref.ID, ref.Version, ErrDefinitionNotFound)
	}
	definition.Document = append(json.RawMessage(nil), definition.Document...)
	return definition, nil
}

func (r *Registry) Latest(releaseID string, kind Kind, id string) (Definition, error) {
	var latest Definition
	found := false
	for ref, value := range r.definitions {
		if ref.ReleaseID != releaseID || ref.Kind != kind || ref.ID != id || (found && ref.Version <= latest.Version) {
			continue
		}
		latest = value
		found = true
	}
	if !found {
		return Definition{}, fmt.Errorf("release %q %s %s: %w", releaseID, kind, id, ErrDefinitionNotFound)
	}
	latest.Document = append(json.RawMessage(nil), latest.Document...)
	return latest, nil
}

func (r *Registry) Definitions(releaseID string) ([]Definition, error) {
	if _, exists := r.releases[releaseID]; !exists {
		return nil, fmt.Errorf("release %q: %w", releaseID, ErrDefinitionNotFound)
	}
	definitions := make([]Definition, 0)
	for ref, definition := range r.definitions {
		if ref.ReleaseID != releaseID {
			continue
		}
		definition.Document = append(json.RawMessage(nil), definition.Document...)
		definitions = append(definitions, definition)
	}
	sort.Slice(definitions, func(i, j int) bool {
		if definitions[i].Kind != definitions[j].Kind {
			return definitions[i].Kind < definitions[j].Kind
		}
		if definitions[i].ID != definitions[j].ID {
			return definitions[i].ID < definitions[j].ID
		}
		return definitions[i].Version < definitions[j].Version
	})
	return definitions, nil
}

func (r *Registry) releaseDir(releaseID string) (string, error) {
	if _, exists := r.releases[releaseID]; !exists {
		return "", fmt.Errorf("release %q: %w", releaseID, ErrDefinitionNotFound)
	}
	return filepath.Join(r.contentDir, "releases", releaseID), nil
}

func loadCurrentRelease(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read current release pointer: %w", err)
	}
	var pointer struct {
		ReleaseID string `json:"release_id"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&pointer); err != nil {
		return "", fmt.Errorf("parse current release pointer: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return "", fmt.Errorf("current release pointer contains trailing JSON")
		}
		return "", fmt.Errorf("parse current release pointer trailing data: %w", err)
	}
	if !releaseIDPattern.MatchString(pointer.ReleaseID) {
		return "", fmt.Errorf("current release pointer contains invalid release ID %q", pointer.ReleaseID)
	}
	return pointer.ReleaseID, nil
}

func readManifest(releaseDir string) (ReleaseManifest, error) {
	raw, err := os.ReadFile(filepath.Join(releaseDir, "manifest.json"))
	if err != nil {
		return ReleaseManifest{}, fmt.Errorf("read release manifest: %w", err)
	}
	var manifest ReleaseManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return ReleaseManifest{}, fmt.Errorf("parse release manifest: %w", err)
	}
	return manifest, nil
}

func cloneManifest(manifest ReleaseManifest) ReleaseManifest {
	manifest.Definitions = append([]ManifestDefinition(nil), manifest.Definitions...)
	manifest.Assets = append([]ManifestAsset(nil), manifest.Assets...)
	return manifest
}
