package definition

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	"github.com/gowebpki/jcs"
)

type AssetDigest struct {
	Source string
	Path   string
	SHA256 string
}

type DefinitionDigest struct {
	ID          string
	Version     int
	ContentHash string
}

type FileDigest struct {
	Path   string
	SHA256 string
}

func CanonicalJSON(document []byte) ([]byte, error) {
	canonical, err := jcs.Transform(document)
	if err != nil {
		return nil, fmt.Errorf("canonicalize JSON: %w", err)
	}
	return canonical, nil
}

func SHA256Hex(contents []byte) string {
	digest := sha256.Sum256(contents)
	return hex.EncodeToString(digest[:])
}

func TaskBundleHash(taskDefinition []byte, assets []AssetDigest) (string, error) {
	canonical, err := CanonicalJSON(taskDefinition)
	if err != nil {
		return "", err
	}

	ordered := append([]AssetDigest(nil), assets...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].Source == ordered[j].Source {
			return ordered[i].Path < ordered[j].Path
		}
		return ordered[i].Source < ordered[j].Source
	})

	preimage := append([]byte(nil), canonical...)
	preimage = append(preimage, '\n')
	workspacePaths := make(map[string]struct{}, len(ordered))
	for i, asset := range ordered {
		if i > 0 && asset.Source == ordered[i-1].Source {
			return "", fmt.Errorf("duplicate asset source %q", asset.Source)
		}
		if _, exists := workspacePaths[asset.Path]; exists {
			return "", fmt.Errorf("duplicate asset workspace path %q", asset.Path)
		}
		workspacePaths[asset.Path] = struct{}{}
		if err := validateSHA256("asset hash", asset.SHA256); err != nil {
			return "", fmt.Errorf("asset %q: %w", asset.Source, err)
		}
		preimage = append(preimage, asset.Source...)
		preimage = append(preimage, 0)
		preimage = append(preimage, asset.Path...)
		preimage = append(preimage, 0)
		preimage = append(preimage, asset.SHA256...)
		preimage = append(preimage, '\n')
	}
	return SHA256Hex(preimage), nil
}

func RuleSetHash(activityDefinition []byte, capabilities []DefinitionDigest, taskBundleHash string) (string, error) {
	canonical, err := CanonicalJSON(activityDefinition)
	if err != nil {
		return "", err
	}
	if err := validateSHA256("task bundle hash", taskBundleHash); err != nil {
		return "", err
	}

	ordered := append([]DefinitionDigest(nil), capabilities...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].ID == ordered[j].ID {
			return ordered[i].Version < ordered[j].Version
		}
		return ordered[i].ID < ordered[j].ID
	})

	preimage := append([]byte(nil), canonical...)
	preimage = append(preimage, '\n')
	for i, capability := range ordered {
		if i > 0 && capability.ID == ordered[i-1].ID && capability.Version == ordered[i-1].Version {
			return "", fmt.Errorf("duplicate capability %s version %d", capability.ID, capability.Version)
		}
		if err := validateSHA256("capability content hash", capability.ContentHash); err != nil {
			return "", fmt.Errorf("capability %s version %d: %w", capability.ID, capability.Version, err)
		}
		preimage = append(preimage, capability.ID...)
		preimage = append(preimage, 0)
		preimage = append(preimage, fmt.Sprintf("%d", capability.Version)...)
		preimage = append(preimage, 0)
		preimage = append(preimage, capability.ContentHash...)
		preimage = append(preimage, '\n')
	}
	preimage = append(preimage, taskBundleHash...)
	preimage = append(preimage, '\n')
	return SHA256Hex(preimage), nil
}

func FullBundleHash(files []FileDigest) (string, error) {
	ordered := append([]FileDigest(nil), files...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].Path < ordered[j].Path
	})

	var preimage []byte
	for i, file := range ordered {
		if i > 0 && file.Path == ordered[i-1].Path {
			return "", fmt.Errorf("duplicate bundle path %q", file.Path)
		}
		if file.Path == "" {
			return "", fmt.Errorf("bundle path is required")
		}
		if err := validateSHA256("file hash", file.SHA256); err != nil {
			return "", fmt.Errorf("bundle path %q: %w", file.Path, err)
		}
		preimage = append(preimage, file.Path...)
		preimage = append(preimage, 0)
		preimage = append(preimage, file.SHA256...)
		preimage = append(preimage, '\n')
	}
	return SHA256Hex(preimage), nil
}

func validateSHA256(name, digest string) error {
	if len(digest) != sha256.Size*2 {
		return fmt.Errorf("%s must be a lowercase SHA-256 digest", name)
	}
	for _, character := range digest {
		if (character < '0' || character > '9') && (character < 'a' || character > 'f') {
			return fmt.Errorf("%s must be a lowercase SHA-256 digest", name)
		}
	}
	return nil
}
