package definition

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"
)

var releaseIDPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*-v[0-9]+$`)

type ReleaseOptions struct {
	ContentDir  string
	ActivitySet string
	ReleaseID   string
	CreatedAt   time.Time
	OutputDir   string
}

func ValidateDrafts(contentDir, activitySet string) error {
	_, err := loadDraftSet(contentDir, activitySet)
	return err
}

func BuildRelease(options ReleaseOptions) (string, error) {
	if options.ContentDir == "" {
		return "", fmt.Errorf("content directory is required")
	}
	if !releaseIDPattern.MatchString(options.ReleaseID) {
		return "", fmt.Errorf("release ID %q must end in a numeric version", options.ReleaseID)
	}
	if options.CreatedAt.IsZero() {
		return "", fmt.Errorf("release creation time is required")
	}
	createdAt := options.CreatedAt.UTC().Truncate(time.Second).Format(time.RFC3339)
	if options.OutputDir == "" {
		options.OutputDir = filepath.Join(options.ContentDir, "releases")
	}

	targetDir := filepath.Join(options.OutputDir, options.ReleaseID)
	if _, err := os.Lstat(targetDir); err == nil {
		return "", fmt.Errorf("release %q already exists", options.ReleaseID)
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("inspect release target: %w", err)
	}
	if err := os.MkdirAll(options.OutputDir, 0o755); err != nil {
		return "", fmt.Errorf("create release output directory: %w", err)
	}

	drafts, err := loadDraftSet(options.ContentDir, options.ActivitySet)
	if err != nil {
		return "", err
	}
	stagingDir, err := os.MkdirTemp(options.OutputDir, "."+options.ReleaseID+"-")
	if err != nil {
		return "", fmt.Errorf("create release staging directory: %w", err)
	}
	defer os.RemoveAll(stagingDir)

	manifest, err := writeReleaseBundle(stagingDir, options.ReleaseID, options.ActivitySet, createdAt, drafts)
	if err != nil {
		return "", err
	}
	manifestJSON, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode release manifest: %w", err)
	}
	manifestJSON = append(manifestJSON, '\n')
	validator, err := NewValidator(os.DirFS(filepath.Join(options.ContentDir, "schemas")))
	if err != nil {
		return "", err
	}
	if err := validator.Validate(KindReleaseManifest, manifestJSON); err != nil {
		return "", err
	}
	if err := os.WriteFile(filepath.Join(stagingDir, "manifest.json"), manifestJSON, 0o644); err != nil {
		return "", fmt.Errorf("write release manifest: %w", err)
	}
	if err := VerifyRelease(stagingDir, filepath.Join(options.ContentDir, "schemas")); err != nil {
		return "", fmt.Errorf("verify generated release: %w", err)
	}
	if err := os.Rename(stagingDir, targetDir); err != nil {
		return "", fmt.Errorf("publish release: %w", err)
	}
	return targetDir, nil
}

func writeReleaseBundle(stagingDir, releaseID, activitySet, createdAt string, drafts draftSet) (ReleaseManifest, error) {
	manifest := ReleaseManifest{
		ReleaseID:     releaseID,
		CreatedAt:     createdAt,
		SchemaVersion: ReleaseManifestSchemaVersion,
		ActivitySet:   activitySet,
		Definitions:   make([]ManifestDefinition, 0, len(drafts.Capabilities)+len(drafts.Activities)+len(drafts.Tasks)),
		Assets:        []ManifestAsset{},
	}
	bundleDir := filepath.Join(stagingDir, "bundle")
	fileDigests := make([]FileDigest, 0)
	capabilityDigests := make(map[string]DefinitionDigest, len(drafts.Capabilities))
	taskBundleHashes := make(map[string]string, len(drafts.Tasks))

	for _, capability := range drafts.Capabilities {
		contentHash, fileDigest, err := writeCanonicalDefinition(bundleDir, capability.BundlePath, capability.Raw)
		if err != nil {
			return ReleaseManifest{}, err
		}
		manifest.Definitions = append(manifest.Definitions, ManifestDefinition{
			Kind: KindCapability, ID: capability.Document.ID, Version: capability.Document.Version,
			Path: capability.BundlePath, ContentHash: contentHash,
		})
		fileDigests = append(fileDigests, fileDigest)
		capabilityDigests[referenceKey(capability.Document.ID, capability.Document.Version)] = DefinitionDigest{
			ID: capability.Document.ID, Version: capability.Document.Version, ContentHash: contentHash,
		}
	}

	for _, task := range drafts.Tasks {
		contentHash, fileDigest, err := writeCanonicalDefinition(bundleDir, task.BundlePath, task.Raw)
		if err != nil {
			return ReleaseManifest{}, err
		}
		fileDigests = append(fileDigests, fileDigest)
		assetDigests := make([]AssetDigest, 0, len(task.Document.Files))
		for _, file := range task.Document.Files {
			sourcePath := filepath.Join(filepath.Dir(task.SourcePath), filepath.FromSlash(file.Source))
			contents, err := os.ReadFile(sourcePath)
			if err != nil {
				return ReleaseManifest{}, fmt.Errorf("read task asset %s: %w", sourcePath, err)
			}
			if actual := SHA256Hex(contents); actual != file.SHA256 {
				return ReleaseManifest{}, fmt.Errorf("task %s asset %q hash mismatch: definition=%s actual=%s", task.Document.ID, file.Source, file.SHA256, actual)
			}
			bundlePath := filepath.ToSlash(filepath.Join("tasks", task.Document.ID, filepath.FromSlash(file.Source)))
			if err := writeBundleFile(bundleDir, bundlePath, contents); err != nil {
				return ReleaseManifest{}, err
			}
			manifest.Assets = append(manifest.Assets, ManifestAsset{
				TaskID: task.Document.ID, TaskVersion: task.Document.Version, Source: file.Source,
				WorkspacePath: file.Path, BundlePath: bundlePath, Role: file.Role, SHA256: file.SHA256,
			})
			assetDigests = append(assetDigests, AssetDigest{Source: file.Source, Path: file.Path, SHA256: file.SHA256})
			fileDigests = append(fileDigests, FileDigest{Path: bundlePath, SHA256: file.SHA256})
		}
		bundleHash, err := TaskBundleHash(task.Raw, assetDigests)
		if err != nil {
			return ReleaseManifest{}, fmt.Errorf("hash task %s version %d: %w", task.Document.ID, task.Document.Version, err)
		}
		manifest.Definitions = append(manifest.Definitions, ManifestDefinition{
			Kind: KindTask, ID: task.Document.ID, Version: task.Document.Version,
			Path: task.BundlePath, ContentHash: contentHash, BundleHash: bundleHash,
		})
		taskBundleHashes[referenceKey(task.Document.ID, task.Document.Version)] = bundleHash
	}

	for _, activity := range drafts.Activities {
		contentHash, fileDigest, err := writeCanonicalDefinition(bundleDir, activity.BundlePath, activity.Raw)
		if err != nil {
			return ReleaseManifest{}, err
		}
		fileDigests = append(fileDigests, fileDigest)
		capabilities := make([]DefinitionDigest, 0, len(activity.Document.CapabilityRefs))
		for _, capabilityRef := range activity.Document.CapabilityRefs {
			capability, exists := capabilityDigests[referenceKey(capabilityRef.ID, capabilityRef.Version)]
			if !exists {
				return ReleaseManifest{}, fmt.Errorf("activity %s references capability %s version %d outside release", activity.Document.ID, capabilityRef.ID, capabilityRef.Version)
			}
			capabilities = append(capabilities, capability)
		}
		taskBundleHash, exists := taskBundleHashes[referenceKey(activity.Document.TaskRef.ID, activity.Document.TaskRef.Version)]
		if !exists {
			return ReleaseManifest{}, fmt.Errorf("activity %s references task %s version %d outside release", activity.Document.ID, activity.Document.TaskRef.ID, activity.Document.TaskRef.Version)
		}
		ruleSetHash, err := RuleSetHash(activity.Raw, capabilities, taskBundleHash)
		if err != nil {
			return ReleaseManifest{}, fmt.Errorf("hash activity %s version %d rules: %w", activity.Document.ID, activity.Document.Version, err)
		}
		manifest.Definitions = append(manifest.Definitions, ManifestDefinition{
			Kind: KindActivity, ID: activity.Document.ID, Version: activity.Document.Version,
			Path: activity.BundlePath, ContentHash: contentHash, RuleSetHash: ruleSetHash,
		})
	}

	sort.Slice(manifest.Definitions, func(i, j int) bool {
		if manifest.Definitions[i].Kind != manifest.Definitions[j].Kind {
			return manifest.Definitions[i].Kind < manifest.Definitions[j].Kind
		}
		if manifest.Definitions[i].ID != manifest.Definitions[j].ID {
			return manifest.Definitions[i].ID < manifest.Definitions[j].ID
		}
		return manifest.Definitions[i].Version < manifest.Definitions[j].Version
	})
	sort.Slice(manifest.Assets, func(i, j int) bool {
		if manifest.Assets[i].TaskID != manifest.Assets[j].TaskID {
			return manifest.Assets[i].TaskID < manifest.Assets[j].TaskID
		}
		return manifest.Assets[i].Source < manifest.Assets[j].Source
	})
	bundleHash, err := FullBundleHash(fileDigests)
	if err != nil {
		return ReleaseManifest{}, err
	}
	manifest.BundleHash = bundleHash
	return manifest, nil
}

func writeCanonicalDefinition(bundleDir, bundlePath string, raw []byte) (string, FileDigest, error) {
	canonical, err := CanonicalJSON(raw)
	if err != nil {
		return "", FileDigest{}, err
	}
	if err := writeBundleFile(bundleDir, bundlePath, canonical); err != nil {
		return "", FileDigest{}, err
	}
	hash := SHA256Hex(canonical)
	return hash, FileDigest{Path: bundlePath, SHA256: hash}, nil
}

func writeBundleFile(bundleDir, bundlePath string, contents []byte) error {
	if !isCleanRelativePath(bundlePath) {
		return fmt.Errorf("invalid bundle path %q", bundlePath)
	}
	targetPath := filepath.Join(bundleDir, filepath.FromSlash(bundlePath))
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("create bundle directory: %w", err)
	}
	if _, err := os.Lstat(targetPath); err == nil {
		return fmt.Errorf("duplicate bundle path %q", bundlePath)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("inspect bundle path %q: %w", bundlePath, err)
	}
	if err := os.WriteFile(targetPath, contents, 0o644); err != nil {
		return fmt.Errorf("write bundle path %q: %w", bundlePath, err)
	}
	return nil
}
