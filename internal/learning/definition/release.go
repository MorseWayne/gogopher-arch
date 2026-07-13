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
	for i, asset := range ordered {
		if i > 0 && asset.Source == ordered[i-1].Source {
			return "", fmt.Errorf("duplicate asset source %q", asset.Source)
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
