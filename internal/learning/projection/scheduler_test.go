package projection

import (
	"strings"
	"testing"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/execution"
)

func TestCapabilityOutcomeHandlesMixedRuleResults(t *testing.T) {
	task := definition.ExecutionTask{AssessmentRules: []definition.AssessmentRule{
		{RuleID: "first", CapabilityRefs: []definition.VersionedDefinitionRef{{ID: "M1-07", Version: 1}}},
		{RuleID: "second", CapabilityRefs: []definition.VersionedDefinitionRef{{ID: "M1-07", Version: 1}}},
	}}
	tests := []struct {
		name    string
		results map[string]execution.RuleStatus
		want    string
	}{
		{name: "all passed", results: map[string]execution.RuleStatus{"first": execution.RulePassed, "second": execution.RulePassed}, want: "passed"},
		{name: "failed wins", results: map[string]execution.RuleStatus{"first": execution.RuleFailed, "second": execution.RuleNotEvaluated}, want: "failed"},
		{name: "not evaluated is incomplete", results: map[string]execution.RuleStatus{"first": execution.RulePassed, "second": execution.RuleNotEvaluated}, want: "incomplete"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := capabilityOutcome(task, test.results, "M1-07", 1)
			if err != nil || got != test.want {
				t.Fatalf("capabilityOutcome() = %q, %v; want %q", got, err, test.want)
			}
		})
	}
}

func TestValidatedRuleResultsRejectsIncompleteDuplicateAndUnknownSets(t *testing.T) {
	task := definition.ExecutionTask{AssessmentRules: []definition.AssessmentRule{{RuleID: "first"}, {RuleID: "second"}}}
	tests := []struct {
		name    string
		results []execution.RuleResult
		want    string
	}{
		{name: "incomplete", results: []execution.RuleResult{{RuleID: "first", Status: execution.RulePassed}}, want: "incomplete"},
		{name: "duplicate", results: []execution.RuleResult{{RuleID: "first", Status: execution.RulePassed}, {RuleID: "first", Status: execution.RuleFailed}}, want: "duplicate"},
		{name: "unknown", results: []execution.RuleResult{{RuleID: "first", Status: execution.RulePassed}, {RuleID: "unknown", Status: execution.RulePassed}}, want: "unknown"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := validatedRuleResults(task, test.results)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("validatedRuleResults() error = %v", err)
			}
		})
	}
}
