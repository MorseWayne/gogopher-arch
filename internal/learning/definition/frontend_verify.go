package definition

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func VerifyFrontendBundle(releaseDir, webDist string) error {
	manifestJSON, err := os.ReadFile(filepath.Join(releaseDir, "manifest.json"))
	if err != nil {
		return fmt.Errorf("read release manifest: %w", err)
	}
	var manifest ReleaseManifest
	if err := json.Unmarshal(manifestJSON, &manifest); err != nil {
		return fmt.Errorf("parse release manifest: %w", err)
	}

	type forbiddenValue struct {
		asset string
		value []byte
	}
	var forbidden []forbiddenValue
	for _, asset := range manifest.Assets {
		if asset.Role != "held_out_test" {
			continue
		}
		contents, err := readBundleFile(filepath.Join(releaseDir, "bundle"), asset.BundlePath)
		if err != nil {
			return err
		}
		encodedJSON, err := json.Marshal(string(contents))
		if err != nil {
			return fmt.Errorf("encode held-out asset fingerprint: %w", err)
		}
		forbidden = append(forbidden,
			forbiddenValue{asset: asset.Source, value: []byte(filepath.Base(asset.Source))},
			forbiddenValue{asset: asset.Source, value: []byte(asset.Source)},
			forbiddenValue{asset: asset.Source, value: []byte(asset.WorkspacePath)},
			forbiddenValue{asset: asset.Source, value: []byte(asset.SHA256)},
			forbiddenValue{asset: asset.Source, value: contents},
			forbiddenValue{asset: asset.Source, value: encodedJSON[1 : len(encodedJSON)-1]},
			forbiddenValue{asset: asset.Source, value: []byte(base64.StdEncoding.EncodeToString(contents))},
		)
	}
	if len(forbidden) == 0 {
		return nil
	}

	info, err := os.Stat(webDist)
	if err != nil {
		return fmt.Errorf("inspect frontend build directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("frontend build path %q is not a directory", webDist)
	}
	return filepath.WalkDir(webDist, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("frontend build may not contain symlink %q", path)
		}
		if entry.IsDir() {
			return nil
		}
		contents, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, candidate := range forbidden {
			if len(candidate.value) > 0 && bytes.Contains(contents, candidate.value) {
				return fmt.Errorf("frontend build file %q contains held-out asset fingerprint from %q", path, candidate.asset)
			}
		}
		return nil
	})
}
