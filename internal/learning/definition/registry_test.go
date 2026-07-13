package definition

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRegistryIndexesCurrentAndHistoricalDefinitions(t *testing.T) {
	registry, err := LoadRegistry(RegistryOptions{ContentDir: repositoryContentDir(t)})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := registry.CurrentReleaseID(), testReleaseID; got != want {
		t.Fatalf("CurrentReleaseID() = %q, want %q", got, want)
	}
	definitions, err := registry.Definitions(testReleaseID)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(definitions), 17; got != want {
		t.Fatalf("len(Definitions()) = %d, want %d", got, want)
	}
	ref := DefinitionRef{ReleaseID: testReleaseID, Kind: KindActivity, ID: "assessment-check-config", Version: 3}
	definition, err := registry.Get(ref)
	if err != nil {
		t.Fatal(err)
	}
	if definition.ContentHash == "" || definition.RuleSetHash == "" {
		t.Fatalf("activity hashes are incomplete: %#v", definition)
	}
	definition.Document[0] = '['
	again, err := registry.Get(ref)
	if err != nil {
		t.Fatal(err)
	}
	if again.Document[0] != '{' {
		t.Fatal("Get() exposed mutable registry document storage")
	}
	if _, err := registry.Get(DefinitionRef{ReleaseID: testReleaseID, Kind: KindTask, ID: "missing", Version: 1}); !errors.Is(err, ErrDefinitionNotFound) {
		t.Fatalf("Get(missing) error = %v, want ErrDefinitionNotFound", err)
	}
}

func TestRegistryViewsExcludePrivateEvaluationContract(t *testing.T) {
	registry, err := LoadRegistry(RegistryOptions{ContentDir: repositoryContentDir(t)})
	if err != nil {
		t.Fatal(err)
	}
	activity, err := registry.ActivityView(testReleaseID, "assessment-check-config", 3)
	if err != nil {
		t.Fatal(err)
	}
	review, err := registry.ReviewActivity(testReleaseID, activity.CapabilityRefs)
	if err != nil || review.ID != "review-check-config-variant" || review.Mode != "review" {
		t.Fatalf("ReviewActivity() = %#v, %v", review, err)
	}
	toolingRemediation, err := registry.RemediationActivity(testReleaseID, VersionedDefinitionRef{ID: "M1-01", Version: 2})
	if err != nil || toolingRemediation.ID != "guided-run-model" || toolingRemediation.Mode != "guided" {
		t.Fatalf("RemediationActivity(M1-01) = %#v, %v", toolingRemediation, err)
	}
	errorRemediation, err := registry.RemediationActivity(testReleaseID, VersionedDefinitionRef{ID: "M1-03", Version: 1})
	if err != nil || errorRemediation.ID != "practice-error-contract" || errorRemediation.Mode != "practice" {
		t.Fatalf("RemediationActivity(M1-03) = %#v, %v", errorRemediation, err)
	}
	variant, err := registry.VariantReviewActivity(testReleaseID, VersionedDefinitionRef{ID: "M1-03", Version: 1})
	if err != nil || variant.ID != review.ID || variant.Version != review.Version {
		t.Fatalf("VariantReviewActivity(M1-03) = %#v, %v", variant, err)
	}
	task, err := registry.TaskView(testReleaseID, activity.TaskRef.ID, activity.TaskRef.Version)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(struct {
		Activity ActivityView `json:"activity"`
		Task     TaskView     `json:"task"`
	}{Activity: activity, Task: task})
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"held_out_tests", "assessment_rules", "evidence_rules", `"actions":`, "bundle_path", "source", "sha256", "先阅读 README"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("public view contains private field %q: %s", forbidden, encoded)
		}
	}
	if task.Readme == "" || len(task.AllowedActions) == 0 {
		t.Fatalf("public Task context is incomplete: %#v", task)
	}
	if len(task.Hints) != 3 || task.Hints[0].ID != "trace-contract" || task.Hints[0].Level != 1 {
		t.Fatalf("public Task hint summaries = %#v", task.Hints)
	}

	workspace, err := registry.PublicWorkspace(testReleaseID, activity.TaskRef.ID, activity.TaskRef.Version)
	if err != nil {
		t.Fatal(err)
	}
	if _, exists := workspace["internal/config/visible_test.go"]; !exists {
		t.Fatal("public workspace is missing visible test")
	}
	if _, exists := workspace["internal/config/heldout_test.go"]; exists {
		t.Fatal("public workspace contains held-out test")
	}
}

func TestLoadRegistryRejectsDamagedAndMissingRequiredRelease(t *testing.T) {
	contentDir := copyRegistryContent(t)
	assetPath := filepath.Join(contentDir, "releases", testReleaseID, "bundle", "tasks", "assessment-check-config-v2", "heldout", "internal", "config", "heldout_test.go")
	if err := os.WriteFile(assetPath, []byte("package config\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadRegistry(RegistryOptions{ContentDir: contentDir}); err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("LoadRegistry(damaged) error = %v, want hash mismatch", err)
	}

	if _, err := LoadRegistry(RegistryOptions{
		ContentDir: repositoryContentDir(t), RequiredReleaseIDs: []string{"m1-first-slice-v999"},
	}); err == nil || !strings.Contains(err.Error(), "historical release") {
		t.Fatalf("LoadRegistry(missing required) error = %v, want historical release", err)
	}
}

func TestBootstrapRegistryChecksHistoryBeforeRegistration(t *testing.T) {
	history := &registryHistoryStub{referenced: []string{testReleaseID}}
	registry, err := BootstrapRegistry(context.Background(), repositoryContentDir(t), history)
	if err != nil {
		t.Fatal(err)
	}
	if history.registered != registry {
		t.Fatal("BootstrapRegistry() did not register the verified registry")
	}

	history = &registryHistoryStub{referenced: []string{"m1-first-slice-v999"}}
	if _, err := BootstrapRegistry(context.Background(), repositoryContentDir(t), history); err == nil || !strings.Contains(err.Error(), "historical release") {
		t.Fatalf("BootstrapRegistry(missing history) error = %v", err)
	}
	if history.registered != nil {
		t.Fatal("BootstrapRegistry() registered content before checking historical releases")
	}
}

type registryHistoryStub struct {
	referenced []string
	registered *Registry
}

func (s *registryHistoryStub) ReferencedReleaseIDs(context.Context) ([]string, error) {
	return append([]string(nil), s.referenced...), nil
}

func (s *registryHistoryStub) Register(_ context.Context, registry *Registry) error {
	s.registered = registry
	return nil
}

func copyRegistryContent(t *testing.T) string {
	t.Helper()
	source := repositoryContentDir(t)
	target := t.TempDir()
	for _, relative := range []string{"current-release.json", "schemas", filepath.Join("releases", testReleaseID)} {
		if err := copyPath(filepath.Join(source, relative), filepath.Join(target, relative)); err != nil {
			t.Fatal(err)
		}
	}
	return target
}

func copyPath(source, target string) error {
	info, err := os.Lstat(source)
	if err != nil {
		return err
	}
	if info.IsDir() {
		if err := os.MkdirAll(target, info.Mode().Perm()); err != nil {
			return err
		}
		entries, err := os.ReadDir(source)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			if err := copyPath(filepath.Join(source, entry.Name()), filepath.Join(target, entry.Name())); err != nil {
				return err
			}
		}
		return nil
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return errors.New("test fixture may not contain symlink")
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	if _, err := io.Copy(output, input); err != nil {
		output.Close()
		return err
	}
	return output.Close()
}
