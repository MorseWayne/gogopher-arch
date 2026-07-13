package projection

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

func TestStateTransitionsAndIndependenceOrdering(t *testing.T) {
	if !AcquisitionExploring.CanTransition(AcquisitionStable) || AcquisitionStable.CanTransition(AcquisitionVerified) {
		t.Fatal("acquisition transition table is not monotonic")
	}
	ordered := []IndependenceState{
		IndependenceGuided, IndependenceAIAssisted, IndependenceHinted,
		IndependenceReferenced, IndependenceIndependent,
	}
	for index := 0; index < len(ordered)-1; index++ {
		if !ordered[index].CanTransition(ordered[index+1]) || ordered[index+1].CanTransition(ordered[index]) {
			t.Fatalf("independence transition %s -> %s", ordered[index], ordered[index+1])
		}
	}
	if !RetentionFresh.CanTransition(RetentionRusty) || !RetentionRusty.CanTransition(RetentionFresh) {
		t.Fatal("retention base must support review failure and recovery")
	}
}

func TestProjectEvidenceCombinationsAndAsOf(t *testing.T) {
	now := time.Date(2026, time.July, 13, 12, 0, 0, 0, time.UTC)
	policy := definition.CapabilityPolicyView{
		ID: "M1-03", Version: 1,
		RequiredEvidence: []definition.RequiredEvidenceView{{
			Type: "implement", Independence: "independent", Context: "same_context", RuleIDs: []string{"error-chain", "resource-closed"},
		}},
	}
	fact := func(rule, mode string, independence IndependenceState, context TransferState, at time.Time) EvidenceFact {
		return EvidenceFact{
			EvidenceType: "implement", RuleID: rule, Result: "passed", Independence: independence,
			Context: context, ActivityMode: mode, QualifyingReview: mode == "review", OccurredAt: at,
		}
	}
	fullAssessment := []EvidenceFact{
		fact("error-chain", "assessment", IndependenceIndependent, TransferSameContext, now.Add(-2*time.Hour)),
		fact("resource-closed", "assessment", IndependenceIndependent, TransferSameContext, now.Add(-time.Hour)),
	}
	dueAt := now.Add(-time.Minute)
	tests := []struct {
		name         string
		input        Input
		acquisition  AcquisitionState
		independence IndependenceState
		transfer     TransferState
		retention    RetentionState
	}{
		{name: "empty", input: Input{RetentionBase: RetentionFresh, AsOf: now}, acquisition: AcquisitionNotStarted, independence: IndependenceUnverified, transfer: TransferUnverified, retention: RetentionStateFresh},
		{name: "guided", input: Input{Evidence: []EvidenceFact{fact("intro", "guided", IndependenceGuided, TransferSameContext, now)}, RetentionBase: RetentionFresh, AsOf: now}, acquisition: AcquisitionExploring, independence: IndependenceGuided, transfer: TransferSameContext, retention: RetentionStateFresh},
		{name: "practice", input: Input{Evidence: []EvidenceFact{fact("error-chain", "practice", IndependenceHinted, TransferSameContext, now)}, RetentionBase: RetentionFresh, AsOf: now}, acquisition: AcquisitionPracticed, independence: IndependenceHinted, transfer: TransferSameContext, retention: RetentionStateFresh},
		{name: "assessment below independence threshold", input: Input{Evidence: []EvidenceFact{fact("error-chain", "assessment", IndependenceAIAssisted, TransferSameContext, now), fact("resource-closed", "assessment", IndependenceAIAssisted, TransferSameContext, now)}, RetentionBase: RetentionFresh, AsOf: now}, acquisition: AcquisitionNotStarted, independence: IndependenceAIAssisted, transfer: TransferSameContext, retention: RetentionStateFresh},
		{name: "verified", input: Input{Evidence: fullAssessment, RetentionBase: RetentionFresh, AsOf: now}, acquisition: AcquisitionVerified, independence: IndependenceIndependent, transfer: TransferSameContext, retention: RetentionStateFresh},
		{name: "stable after variant review", input: Input{Evidence: append(append([]EvidenceFact(nil), fullAssessment...), fact("review-rule", "review", IndependenceIndependent, TransferVariant, now)), RetentionBase: RetentionFresh, AsOf: now}, acquisition: AcquisitionStable, independence: IndependenceIndependent, transfer: TransferVariant, retention: RetentionStateFresh},
		{name: "due derived at read", input: Input{Evidence: fullAssessment, RetentionBase: RetentionFresh, NextReviewAt: &dueAt, AsOf: now}, acquisition: AcquisitionVerified, independence: IndependenceIndependent, transfer: TransferSameContext, retention: RetentionStateDue},
		{name: "rusty overrides due", input: Input{Evidence: fullAssessment, RetentionBase: RetentionRusty, NextReviewAt: &dueAt, AsOf: now}, acquisition: AcquisitionVerified, independence: IndependenceIndependent, transfer: TransferSameContext, retention: RetentionStateRusty},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := Project(policy, test.input)
			if err != nil {
				t.Fatal(err)
			}
			if result.AcquisitionState != test.acquisition || result.IndependenceState != test.independence || result.TransferState != test.transfer || result.RetentionState != test.retention {
				t.Fatalf("result = %#v", result)
			}
		})
	}
	if _, err := Project(policy, Input{RetentionBase: RetentionFresh}); err == nil {
		t.Fatal("Project without as_of error = nil")
	}
}

func TestProjectIsIndependentOfEvidenceOrder(t *testing.T) {
	now := time.Now().UTC()
	policy := definition.CapabilityPolicyView{ID: "M1-09", RequiredEvidence: []definition.RequiredEvidenceView{{
		Type: "test", Independence: "guided", Context: "same_context", RuleIDs: []string{"a", "b"},
	}}}
	facts := []EvidenceFact{
		{EvidenceType: "test", RuleID: "a", Result: "passed", Independence: IndependenceGuided, Context: TransferSameContext, ActivityMode: "guided", OccurredAt: now.Add(-time.Hour)},
		{EvidenceType: "test", RuleID: "b", Result: "passed", Independence: IndependenceIndependent, Context: TransferVariant, ActivityMode: "review", QualifyingReview: true, OccurredAt: now},
	}
	first, _ := Project(policy, Input{Evidence: facts, RetentionBase: RetentionFresh, AsOf: now})
	second, _ := Project(policy, Input{Evidence: []EvidenceFact{facts[1], facts[0]}, RetentionBase: RetentionFresh, AsOf: now})
	firstJSON, _ := json.Marshal(first)
	secondJSON, _ := json.Marshal(second)
	if string(firstJSON) != string(secondJSON) {
		t.Fatalf("projection depends on evidence order: %s != %s", firstJSON, secondJSON)
	}
}
