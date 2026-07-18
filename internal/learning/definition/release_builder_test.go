package definition

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testActivitySet = "m1-first-slice"
	testReleaseID   = "m1-first-slice-v13"
)

var testCreatedAt = time.Date(2026, time.July, 18, 0, 0, 0, 0, time.UTC)

func TestBuildReleaseIsDeterministicAndVerifiable(t *testing.T) {
	contentDir := repositoryContentDir(t)
	firstOutput := filepath.Join(t.TempDir(), "first")
	secondOutput := filepath.Join(t.TempDir(), "second")

	first, err := BuildRelease(ReleaseOptions{
		ContentDir: contentDir, ActivitySet: testActivitySet, ReleaseID: testReleaseID,
		CreatedAt: testCreatedAt, OutputDir: firstOutput,
	})
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildRelease(ReleaseOptions{
		ContentDir: contentDir, ActivitySet: testActivitySet, ReleaseID: testReleaseID,
		CreatedAt: testCreatedAt, OutputDir: secondOutput,
	})
	if err != nil {
		t.Fatal(err)
	}

	firstFiles := readDirectoryFiles(t, first)
	secondFiles := readDirectoryFiles(t, second)
	if len(firstFiles) != len(secondFiles) {
		t.Fatalf("release file count differs: %d != %d", len(firstFiles), len(secondFiles))
	}
	for path, firstContents := range firstFiles {
		secondContents, exists := secondFiles[path]
		if !exists {
			t.Fatalf("second release is missing %s", path)
		}
		if !bytes.Equal(firstContents, secondContents) {
			t.Fatalf("release file %s differs between builds", path)
		}
	}
	if err := VerifyRelease(first, filepath.Join(contentDir, "schemas")); err != nil {
		t.Fatalf("VerifyRelease() error = %v", err)
	}
}

func TestVerifyReleaseRejectsChangedAsset(t *testing.T) {
	contentDir := repositoryContentDir(t)
	releaseDir, err := BuildRelease(ReleaseOptions{
		ContentDir: contentDir, ActivitySet: testActivitySet, ReleaseID: testReleaseID,
		CreatedAt: testCreatedAt, OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	assetPath := filepath.Join(releaseDir, "bundle", "tasks", "assessment-check-config-v2", "heldout", "internal", "config", "heldout_test.go")
	if err := os.WriteFile(assetPath, []byte("package config\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err = VerifyRelease(releaseDir, filepath.Join(contentDir, "schemas"))
	if err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("VerifyRelease() error = %v, want hash mismatch", err)
	}
}

func TestBuildReleaseRefusesOverwrite(t *testing.T) {
	contentDir := repositoryContentDir(t)
	outputDir := t.TempDir()
	options := ReleaseOptions{
		ContentDir: contentDir, ActivitySet: testActivitySet, ReleaseID: testReleaseID,
		CreatedAt: testCreatedAt, OutputDir: outputDir,
	}
	if _, err := BuildRelease(options); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildRelease(options); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("BuildRelease() error = %v, want already exists", err)
	}
}

func TestVerifyFrontendBundleRejectsHeldOutContent(t *testing.T) {
	contentDir := repositoryContentDir(t)
	releaseDir, err := BuildRelease(ReleaseOptions{
		ContentDir: contentDir, ActivitySet: testActivitySet, ReleaseID: testReleaseID,
		CreatedAt: testCreatedAt, OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	webDist := t.TempDir()
	if err := os.WriteFile(filepath.Join(webDist, "app.js"), []byte("console.log('public bundle')\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := VerifyFrontendBundle(releaseDir, webDist); err != nil {
		t.Fatalf("clean frontend bundle rejected: %v", err)
	}
	heldOut, err := os.ReadFile(filepath.Join(releaseDir, "bundle", "tasks", "assessment-check-config-v2", "heldout", "internal", "config", "heldout_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(webDist, "app.js"), heldOut, 0o644); err != nil {
		t.Fatal(err)
	}
	err = VerifyFrontendBundle(releaseDir, webDist)
	if err == nil || !strings.Contains(err.Error(), "private test asset fingerprint") {
		t.Fatalf("VerifyFrontendBundle() error = %v, want private test fingerprint", err)
	}
}

func TestVerifyFrontendBundleRejectsRaceTestContent(t *testing.T) {
	contentDir := repositoryContentDir(t)
	releaseDir, err := BuildRelease(ReleaseOptions{
		ContentDir: contentDir, ActivitySet: testActivitySet, ReleaseID: testReleaseID,
		CreatedAt: testCreatedAt, OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	webDist := t.TempDir()
	raceTest, err := os.ReadFile(filepath.Join(releaseDir, "bundle", "tasks", "assessment-concurrent-registry-v1", "race", "registry_race_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(webDist, "app.js"), raceTest, 0o644); err != nil {
		t.Fatal(err)
	}
	err = VerifyFrontendBundle(releaseDir, webDist)
	if err == nil || !strings.Contains(err.Error(), "private test asset fingerprint") {
		t.Fatalf("VerifyFrontendBundle() error = %v, want private test fingerprint", err)
	}
}

func TestValidateHardPrerequisitesRejectsCycle(t *testing.T) {
	first := capabilityDocument{ID: "M1-01", Version: 1}
	first.Prerequisites.Hard = []versionedRef{{ID: "M1-03", Version: 1}}
	second := capabilityDocument{ID: "M1-03", Version: 1}
	second.Prerequisites.Hard = []versionedRef{{ID: "M1-01", Version: 1}}

	err := validateHardPrerequisites([]draftDefinition[capabilityDocument]{
		{Document: first},
		{Document: second},
	})
	if err == nil || !strings.Contains(err.Error(), "hard prerequisite cycle") {
		t.Fatalf("validateHardPrerequisites() error = %v, want cycle", err)
	}
}

func TestCommittedReleaseMatchesDeterministicBuild(t *testing.T) {
	contentDir := repositoryContentDir(t)
	committedDir := filepath.Join(contentDir, "releases", testReleaseID)
	if _, err := os.Stat(committedDir); err != nil {
		t.Fatalf("committed release is unavailable: %v", err)
	}
	generatedDir, err := BuildRelease(ReleaseOptions{
		ContentDir: contentDir, ActivitySet: testActivitySet, ReleaseID: testReleaseID,
		CreatedAt: testCreatedAt, OutputDir: t.TempDir(),
	})
	if err != nil {
		t.Fatal(err)
	}
	committedFiles := readDirectoryFiles(t, committedDir)
	generatedFiles := readDirectoryFiles(t, generatedDir)
	if len(committedFiles) != len(generatedFiles) {
		t.Fatalf("committed file count = %d, generated = %d", len(committedFiles), len(generatedFiles))
	}
	for path, committed := range committedFiles {
		generated, exists := generatedFiles[path]
		if !exists || !bytes.Equal(committed, generated) {
			t.Fatalf("committed release differs at %s", path)
		}
	}
}

func repositoryContentDir(t *testing.T) string {
	t.Helper()
	dir, err := filepath.Abs("../../../content/learning")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func readDirectoryFiles(t *testing.T, root string) map[string][]byte {
	t.Helper()
	files := make(map[string][]byte)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(relative)], err = os.ReadFile(path)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}
