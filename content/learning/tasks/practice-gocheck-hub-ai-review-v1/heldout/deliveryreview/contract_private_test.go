package deliveryreview

import (
	"reflect"
	"testing"
)

func safePlan() Plan {
	return Plan{AIDeclared: true, AuthBeforeLookup: true, SourceOfTruth: "postgres", CacheFailureMode: "degrade", WorkerConcurrency: 4, RetryLimit: 3, MigrationMode: "forward-only", Gates: []string{"fmt", "vet", "test", "race", "vuln", "migration", "image"}, RuntimeUser: "app", Rollback: "forward-revert"}
}

func TestAIReviewFindsEveryUnsafeDeliveryDecision(t *testing.T) {
	tests := []struct {
		name string
		mutate func(*Plan)
		want string
	}{
		{"AI use undeclared", func(plan *Plan) { plan.AIDeclared = false }, "ai-undeclared"},
		{"lookup before auth", func(plan *Plan) { plan.AuthBeforeLookup = false }, "auth-after-lookup"},
		{"cache is truth", func(plan *Plan) { plan.SourceOfTruth = "cache" }, "invalid-source-of-truth"},
		{"cache outage fails closed", func(plan *Plan) { plan.CacheFailureMode = "fail" }, "cache-does-not-degrade"},
		{"unbounded workers", func(plan *Plan) { plan.WorkerConcurrency = 0 }, "worker-unbounded"},
		{"unbounded retries", func(plan *Plan) { plan.RetryLimit = -1 }, "retry-unbounded"},
		{"down migration", func(plan *Plan) { plan.MigrationMode = "up-and-down" }, "migration-not-forward-only"},
		{"missing race gate", func(plan *Plan) { plan.Gates = plan.Gates[:3] }, "delivery-gates-incomplete"},
		{"root runtime", func(plan *Plan) { plan.RuntimeUser = "root" }, "runtime-is-root"},
		{"destructive rollback", func(plan *Plan) { plan.Rollback = "drop-tables" }, "rollback-not-forward"},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			plan := safePlan()
			testCase.mutate(&plan)
			if got := Review(plan); !contains(got, testCase.want) {
				t.Fatalf("findings = %v, want %q", got, testCase.want)
			}
		})
	}
	bad := Plan{}
	first, second := Review(bad), Review(bad)
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("findings are unstable: %v / %v", first, second)
	}
	seen := map[string]bool{}
	for _, finding := range first {
		if seen[finding] { t.Fatalf("duplicate finding %q", finding) }
		seen[finding] = true
	}
}

func contains(values []string, want string) bool {
	for _, value := range values { if value == want { return true } }
	return false
}
