package definition

import (
	"path/filepath"
	"testing"
)

func TestRegistryReturnsVersionedCapabilityPolicies(t *testing.T) {
	contentDir, _ := filepath.Abs("../../../content/learning")
	registry, err := LoadRegistry(RegistryOptions{ContentDir: contentDir})
	if err != nil {
		t.Fatal(err)
	}
	tooling, err := registry.CapabilityPolicy(registry.CurrentReleaseID(), "M1-01", 1)
	if err != nil {
		t.Fatal(err)
	}
	errorsPolicy, err := registry.CapabilityPolicy(registry.CurrentReleaseID(), "M1-03", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(tooling.RequiredEvidence) != 2 || len(errorsPolicy.RequiredEvidence) != 1 ||
		tooling.RequiredEvidence[0].RuleIDs[0] == errorsPolicy.RequiredEvidence[0].RuleIDs[0] {
		t.Fatalf("capability policies were flattened: tooling=%#v errors=%#v", tooling, errorsPolicy)
	}
	if tooling.ReviewPolicy.FirstReviewAfterDays != 3 || tooling.ReviewPolicy.SuccessIntervalDays != 14 ||
		tooling.ReviewPolicy.FailureRemediationAfterDays != 1 {
		t.Fatalf("review policy = %#v", tooling.ReviewPolicy)
	}
}
