package deliveryreview

import "testing"

func TestReviewSafePlan(t *testing.T) {
	plan := Plan{AIDeclared: true, AuthBeforeLookup: true, SourceOfTruth: "postgres", CacheFailureMode: "degrade", WorkerConcurrency: 4, RetryLimit: 3, MigrationMode: "forward-only", Gates: []string{"fmt", "vet", "test", "race", "vuln", "migration", "image"}, RuntimeUser: "app", Rollback: "forward-revert"}
	if findings := Review(plan); len(findings) != 0 {
		t.Fatalf("safe plan findings = %v", findings)
	}
}
