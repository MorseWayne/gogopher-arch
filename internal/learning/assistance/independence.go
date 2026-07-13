package assistance

func CalculateIndependence(mode string, events []Event, cutoff int64) Independence {
	if mode == "guided" {
		return IndependenceGuided
	}
	var solution, ai, hint, reference bool
	for _, event := range events {
		if event.Sequence > cutoff {
			continue
		}
		switch event.Type {
		case SolutionViewed:
			solution = true
		case AIDeclared:
			ai = true
		case HintRevealed:
			hint = true
		case ReferenceOpened:
			reference = true
		}
	}
	switch {
	case solution:
		return IndependenceGuided
	case ai:
		return IndependenceAIAssisted
	case hint:
		return IndependenceHinted
	case reference:
		return IndependenceReferenced
	default:
		return IndependenceIndependent
	}
}
