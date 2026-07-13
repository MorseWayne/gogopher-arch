package definition

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func VerifyRelease(releaseDir, schemasDir string) error {
	manifestPath := filepath.Join(releaseDir, "manifest.json")
	manifestInfo, err := os.Lstat(manifestPath)
	if err != nil {
		return fmt.Errorf("inspect release manifest: %w", err)
	}
	if manifestInfo.Mode()&os.ModeSymlink != 0 || !manifestInfo.Mode().IsRegular() {
		return fmt.Errorf("release manifest must be a regular file")
	}
	manifestJSON, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("read release manifest: %w", err)
	}
	validator, err := NewValidator(os.DirFS(schemasDir))
	if err != nil {
		return err
	}
	if err := validator.Validate(KindReleaseManifest, manifestJSON); err != nil {
		return err
	}
	var manifest ReleaseManifest
	if err := json.Unmarshal(manifestJSON, &manifest); err != nil {
		return fmt.Errorf("parse release manifest: %w", err)
	}
	if !releaseIDPattern.MatchString(manifest.ReleaseID) {
		return fmt.Errorf("invalid release ID %q", manifest.ReleaseID)
	}

	bundleDir := filepath.Join(releaseDir, "bundle")
	expectedPaths := make(map[string]string, len(manifest.Definitions)+len(manifest.Assets))
	fileDigests := make([]FileDigest, 0, len(expectedPaths))
	definitionByKey := make(map[string]ManifestDefinition, len(manifest.Definitions))
	capabilities := make([]draftDefinition[capabilityDocument], 0)
	capabilityDigests := make(map[string]DefinitionDigest)
	activityDocuments := make(map[string]activityDocument)
	activityRaw := make(map[string][]byte)
	taskDocuments := make(map[string]taskDocument)
	taskRaw := make(map[string][]byte)

	for _, definition := range manifest.Definitions {
		if !isCleanRelativePath(definition.Path) {
			return fmt.Errorf("definition %s version %d has invalid bundle path %q", definition.ID, definition.Version, definition.Path)
		}
		if previous, exists := expectedPaths[definition.Path]; exists {
			return fmt.Errorf("bundle path %q is shared by %s and definition %s version %d", definition.Path, previous, definition.ID, definition.Version)
		}
		expectedPaths[definition.Path] = fmt.Sprintf("definition %s version %d", definition.ID, definition.Version)
		key := definitionKey(definition.Kind, definition.ID, definition.Version)
		if _, exists := definitionByKey[key]; exists {
			return fmt.Errorf("duplicate %s definition %s version %d", definition.Kind, definition.ID, definition.Version)
		}
		definitionByKey[key] = definition

		raw, err := readBundleFile(bundleDir, definition.Path)
		if err != nil {
			return err
		}
		if err := validator.Validate(definition.Kind, raw); err != nil {
			return fmt.Errorf("definition %s version %d: %w", definition.ID, definition.Version, err)
		}
		canonical, err := CanonicalJSON(raw)
		if err != nil {
			return fmt.Errorf("definition %s version %d: %w", definition.ID, definition.Version, err)
		}
		contentHash := SHA256Hex(canonical)
		if contentHash != definition.ContentHash {
			return fmt.Errorf("definition %s version %d content hash mismatch: manifest=%s actual=%s", definition.ID, definition.Version, definition.ContentHash, contentHash)
		}
		fileDigests = append(fileDigests, FileDigest{Path: definition.Path, SHA256: SHA256Hex(raw)})

		switch definition.Kind {
		case KindCapability:
			if definition.BundleHash != "" || definition.RuleSetHash != "" {
				return fmt.Errorf("capability %s version %d contains task or activity hashes", definition.ID, definition.Version)
			}
			var document capabilityDocument
			if err := json.Unmarshal(raw, &document); err != nil {
				return fmt.Errorf("parse capability %s version %d: %w", definition.ID, definition.Version, err)
			}
			if document.ID != definition.ID || document.Version != definition.Version {
				return fmt.Errorf("capability manifest identity does not match %s", definition.Path)
			}
			capabilities = append(capabilities, draftDefinition[capabilityDocument]{Raw: raw, Document: document})
			capabilityDigests[referenceKey(document.ID, document.Version)] = DefinitionDigest{ID: document.ID, Version: document.Version, ContentHash: definition.ContentHash}
		case KindActivity:
			if definition.BundleHash != "" {
				return fmt.Errorf("activity %s version %d contains task bundle hash", definition.ID, definition.Version)
			}
			var document activityDocument
			if err := json.Unmarshal(raw, &document); err != nil {
				return fmt.Errorf("parse activity %s version %d: %w", definition.ID, definition.Version, err)
			}
			if document.ID != definition.ID || document.Version != definition.Version {
				return fmt.Errorf("activity manifest identity does not match %s", definition.Path)
			}
			activityDocuments[referenceKey(document.ID, document.Version)] = document
			activityRaw[referenceKey(document.ID, document.Version)] = raw
		case KindTask:
			if definition.RuleSetHash != "" {
				return fmt.Errorf("task %s version %d contains activity rule set hash", definition.ID, definition.Version)
			}
			var document taskDocument
			if err := json.Unmarshal(raw, &document); err != nil {
				return fmt.Errorf("parse task %s version %d: %w", definition.ID, definition.Version, err)
			}
			if document.ID != definition.ID || document.Version != definition.Version {
				return fmt.Errorf("task manifest identity does not match %s", definition.Path)
			}
			taskDocuments[referenceKey(document.ID, document.Version)] = document
			taskRaw[referenceKey(document.ID, document.Version)] = raw
		default:
			return fmt.Errorf("unsupported definition kind %q", definition.Kind)
		}
	}

	assetsByTask := make(map[string][]ManifestAsset)
	for _, asset := range manifest.Assets {
		if !isCleanRelativePath(asset.Source) || !isCleanRelativePath(asset.WorkspacePath) || !isCleanRelativePath(asset.BundlePath) {
			return fmt.Errorf("task %s version %d contains invalid asset path", asset.TaskID, asset.TaskVersion)
		}
		if previous, exists := expectedPaths[asset.BundlePath]; exists {
			return fmt.Errorf("bundle path %q is shared by %s and task asset %s", asset.BundlePath, previous, asset.Source)
		}
		expectedPath := path.Join("tasks", asset.TaskID, asset.Source)
		if asset.BundlePath != expectedPath {
			return fmt.Errorf("task %s asset %q bundle path is %q, want %q", asset.TaskID, asset.Source, asset.BundlePath, expectedPath)
		}
		expectedPaths[asset.BundlePath] = fmt.Sprintf("task %s asset %s", asset.TaskID, asset.Source)
		raw, err := readBundleFile(bundleDir, asset.BundlePath)
		if err != nil {
			return err
		}
		actualHash := SHA256Hex(raw)
		if actualHash != asset.SHA256 {
			return fmt.Errorf("task %s asset %q hash mismatch: manifest=%s actual=%s", asset.TaskID, asset.Source, asset.SHA256, actualHash)
		}
		fileDigests = append(fileDigests, FileDigest{Path: asset.BundlePath, SHA256: actualHash})
		key := referenceKey(asset.TaskID, asset.TaskVersion)
		assetsByTask[key] = append(assetsByTask[key], asset)
	}

	if err := verifyBundleCoverage(bundleDir, expectedPaths); err != nil {
		return err
	}

	taskBundleHashes := make(map[string]string, len(taskDocuments))
	for key, task := range taskDocuments {
		definition := definitionByKey[definitionKey(KindTask, task.ID, task.Version)]
		assets := assetsByTask[key]
		if len(assets) != len(task.Files) {
			return fmt.Errorf("task %s version %d manifest has %d assets, definition declares %d", task.ID, task.Version, len(assets), len(task.Files))
		}
		manifestBySource := make(map[string]ManifestAsset, len(assets))
		for _, asset := range assets {
			if _, exists := manifestBySource[asset.Source]; exists {
				return fmt.Errorf("task %s version %d has duplicate manifest asset %q", task.ID, task.Version, asset.Source)
			}
			manifestBySource[asset.Source] = asset
		}
		digests := make([]AssetDigest, 0, len(task.Files))
		for _, file := range task.Files {
			asset, exists := manifestBySource[file.Source]
			if !exists {
				return fmt.Errorf("task %s version %d asset %q is absent from manifest", task.ID, task.Version, file.Source)
			}
			if asset.WorkspacePath != file.Path || asset.Role != file.Role || asset.SHA256 != file.SHA256 {
				return fmt.Errorf("task %s version %d asset %q metadata differs from definition", task.ID, task.Version, file.Source)
			}
			digests = append(digests, AssetDigest{Source: file.Source, Path: file.Path, SHA256: file.SHA256})
		}
		bundleHash, err := TaskBundleHash(taskRaw[key], digests)
		if err != nil {
			return fmt.Errorf("task %s version %d: %w", task.ID, task.Version, err)
		}
		if bundleHash != definition.BundleHash {
			return fmt.Errorf("task %s version %d bundle hash mismatch: manifest=%s actual=%s", task.ID, task.Version, definition.BundleHash, bundleHash)
		}
		taskBundleHashes[key] = bundleHash
	}
	for key, assets := range assetsByTask {
		if _, exists := taskDocuments[key]; !exists {
			return fmt.Errorf("manifest contains %d assets for missing task", len(assets))
		}
	}

	for key, activity := range activityDocuments {
		definition := definitionByKey[definitionKey(KindActivity, activity.ID, activity.Version)]
		capabilityHashes := make([]DefinitionDigest, 0, len(activity.CapabilityRefs))
		for _, capabilityRef := range activity.CapabilityRefs {
			capability, exists := capabilityDigests[referenceKey(capabilityRef.ID, capabilityRef.Version)]
			if !exists {
				return fmt.Errorf("activity %s version %d references missing capability %s version %d", activity.ID, activity.Version, capabilityRef.ID, capabilityRef.Version)
			}
			capabilityHashes = append(capabilityHashes, capability)
		}
		taskKey := referenceKey(activity.TaskRef.ID, activity.TaskRef.Version)
		task, exists := taskDocuments[taskKey]
		if !exists {
			return fmt.Errorf("activity %s version %d references missing task %s version %d", activity.ID, activity.Version, activity.TaskRef.ID, activity.TaskRef.Version)
		}
		if err := validateEvidenceRules(activity, task); err != nil {
			return err
		}
		ruleSetHash, err := RuleSetHash(activityRaw[key], capabilityHashes, taskBundleHashes[taskKey])
		if err != nil {
			return fmt.Errorf("activity %s version %d: %w", activity.ID, activity.Version, err)
		}
		if ruleSetHash != definition.RuleSetHash {
			return fmt.Errorf("activity %s version %d rule set hash mismatch: manifest=%s actual=%s", activity.ID, activity.Version, definition.RuleSetHash, ruleSetHash)
		}
	}
	if err := validateHardPrerequisites(capabilities); err != nil {
		return err
	}

	bundleHash, err := FullBundleHash(fileDigests)
	if err != nil {
		return err
	}
	if bundleHash != manifest.BundleHash {
		return fmt.Errorf("release bundle hash mismatch: manifest=%s actual=%s", manifest.BundleHash, bundleHash)
	}
	return nil
}

func readBundleFile(bundleDir, bundlePath string) ([]byte, error) {
	if !isCleanRelativePath(bundlePath) {
		return nil, fmt.Errorf("invalid bundle path %q", bundlePath)
	}
	filePath := filepath.Join(bundleDir, filepath.FromSlash(bundlePath))
	info, err := os.Lstat(filePath)
	if err != nil {
		return nil, fmt.Errorf("inspect bundle path %q: %w", bundlePath, err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("bundle path %q must be a regular file", bundlePath)
	}
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read bundle path %q: %w", bundlePath, err)
	}
	return raw, nil
}

func verifyBundleCoverage(bundleDir string, expected map[string]string) error {
	seen := make(map[string]struct{}, len(expected))
	err := filepath.WalkDir(bundleDir, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("release bundle may not contain symlink %q", filePath)
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(bundleDir, filePath)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if _, exists := expected[relative]; !exists {
			return fmt.Errorf("release bundle contains undeclared file %q", relative)
		}
		seen[relative] = struct{}{}
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk release bundle: %w", err)
	}
	for expectedPath := range expected {
		if _, exists := seen[expectedPath]; !exists {
			return fmt.Errorf("release bundle is missing declared file %q", expectedPath)
		}
	}
	return nil
}

func isCleanRelativePath(value string) bool {
	if value == "" || value == "." || strings.Contains(value, `\`) || path.IsAbs(value) {
		return false
	}
	cleaned := path.Clean(value)
	return cleaned == value && cleaned != ".." && !strings.HasPrefix(cleaned, "../")
}
