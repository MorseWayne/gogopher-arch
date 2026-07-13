package projection

import (
	"fmt"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

var acquisitionRanks = map[AcquisitionState]int{
	AcquisitionNotStarted: 0, AcquisitionExploring: 1, AcquisitionPracticed: 2,
	AcquisitionVerified: 3, AcquisitionStable: 4,
}

var independenceRanks = map[IndependenceState]int{
	IndependenceUnverified: 0, IndependenceGuided: 1, IndependenceAIAssisted: 2,
	IndependenceHinted: 3, IndependenceReferenced: 4, IndependenceIndependent: 5,
}

var transferRanks = map[TransferState]int{
	TransferUnverified: 0, TransferSameContext: 1, TransferVariant: 2, TransferNewProject: 3,
}

func (s AcquisitionState) Valid() bool   { _, ok := acquisitionRanks[s]; return ok }
func (s IndependenceState) Valid() bool  { _, ok := independenceRanks[s]; return ok }
func (s TransferState) Valid() bool      { _, ok := transferRanks[s]; return ok }
func (s RetentionBaseState) Valid() bool { return s == RetentionFresh || s == RetentionRusty }

func (s AcquisitionState) CanTransition(to AcquisitionState) bool {
	return s.Valid() && to.Valid() && acquisitionRanks[to] >= acquisitionRanks[s]
}

func (s IndependenceState) CanTransition(to IndependenceState) bool {
	return s.Valid() && to.Valid() && independenceRanks[to] >= independenceRanks[s]
}

func (s TransferState) CanTransition(to TransferState) bool {
	return s.Valid() && to.Valid() && transferRanks[to] >= transferRanks[s]
}

func (s RetentionBaseState) CanTransition(to RetentionBaseState) bool {
	return s.Valid() && to.Valid()
}

func Project(policy definition.CapabilityPolicyView, input Input) (Result, error) {
	if input.AsOf.IsZero() {
		return Result{}, fmt.Errorf("projection as_of is required")
	}
	if !input.RetentionBase.Valid() {
		return Result{}, fmt.Errorf("invalid retention base state %q", input.RetentionBase)
	}
	for _, requirement := range policy.RequiredEvidence {
		if !IndependenceState(requirement.Independence).Valid() || !TransferState(requirement.Context).Valid() || len(requirement.RuleIDs) == 0 {
			return Result{}, fmt.Errorf("invalid required evidence policy for capability %s", policy.ID)
		}
	}
	result := Result{
		AcquisitionState: AcquisitionNotStarted, IndependenceState: IndependenceUnverified,
		TransferState: TransferUnverified, RetentionBase: input.RetentionBase,
		RetentionState: RetentionStateFresh, NextReviewAt: cloneTime(input.NextReviewAt),
	}
	for _, fact := range input.Evidence {
		if !fact.Independence.Valid() || !fact.Context.Valid() {
			return Result{}, fmt.Errorf("invalid evidence state for rule %q", fact.RuleID)
		}
		result.LastEvidenceAt = later(result.LastEvidenceAt, fact.OccurredAt)
		if fact.Result != "passed" {
			continue
		}
		if independenceRanks[fact.Independence] > independenceRanks[result.IndependenceState] {
			result.IndependenceState = fact.Independence
		}
		if transferRanks[fact.Context] > transferRanks[result.TransferState] {
			result.TransferState = fact.Context
		}
		if fact.Independence == IndependenceIndependent {
			result.LastIndependentAt = later(result.LastIndependentAt, fact.OccurredAt)
		}
		switch fact.ActivityMode {
		case "guided":
			promote(&result.AcquisitionState, AcquisitionExploring)
		case "practice":
			promote(&result.AcquisitionState, AcquisitionPracticed)
		}
	}
	if requirementsSatisfied(policy.RequiredEvidence, input.Evidence, TransferSameContext) {
		promote(&result.AcquisitionState, AcquisitionVerified)
	}
	if result.AcquisitionState == AcquisitionVerified && hasIndependentVariantReview(input.Evidence) {
		promote(&result.AcquisitionState, AcquisitionStable)
	}
	if result.RetentionBase == RetentionRusty {
		result.RetentionState = RetentionStateRusty
	} else if result.NextReviewAt != nil && !result.NextReviewAt.After(input.AsOf) {
		result.RetentionState = RetentionStateDue
	}
	return result, nil
}

func hasIndependentVariantReview(facts []EvidenceFact) bool {
	for _, fact := range facts {
		if fact.ActivityMode == "review" && fact.Result == "passed" &&
			fact.Independence == IndependenceIndependent && transferRanks[fact.Context] >= transferRanks[TransferVariant] {
			return true
		}
	}
	return false
}

func requirementsSatisfied(requirements []definition.RequiredEvidenceView, facts []EvidenceFact, minimumContext TransferState) bool {
	if len(requirements) == 0 {
		return false
	}
	for _, requirement := range requirements {
		requiredIndependence := IndependenceState(requirement.Independence)
		requiredContext := TransferState(requirement.Context)
		if transferRanks[minimumContext] > transferRanks[requiredContext] {
			requiredContext = minimumContext
		}
		for _, ruleID := range requirement.RuleIDs {
			matched := false
			for _, fact := range facts {
				if fact.RuleID == ruleID && fact.EvidenceType == requirement.Type && fact.Result == "passed" &&
					independenceRanks[fact.Independence] >= independenceRanks[requiredIndependence] &&
					transferRanks[fact.Context] >= transferRanks[requiredContext] {
					matched = true
					break
				}
			}
			if !matched {
				return false
			}
		}
	}
	return true
}

func promote(current *AcquisitionState, candidate AcquisitionState) {
	if acquisitionRanks[candidate] > acquisitionRanks[*current] {
		*current = candidate
	}
}

func later(current *time.Time, candidate time.Time) *time.Time {
	value := candidate.UTC()
	if current == nil || value.After(*current) {
		return &value
	}
	return current
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copy := value.UTC()
	return &copy
}
