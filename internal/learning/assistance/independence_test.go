package assistance

import "testing"

func TestCalculateIndependenceUsesStrongestEventThroughCutoff(t *testing.T) {
	events := []Event{
		{Sequence: 1, Type: ReferenceOpened},
		{Sequence: 2, Type: HintRevealed},
		{Sequence: 3, Type: AIDeclared},
		{Sequence: 4, Type: SolutionViewed},
	}
	tests := []struct {
		name   string
		mode   string
		cutoff int64
		want   Independence
	}{
		{name: "guided mode is always guided", mode: "guided", cutoff: 0, want: IndependenceGuided},
		{name: "no assistance", mode: "assessment", cutoff: 0, want: IndependenceIndependent},
		{name: "reference", mode: "assessment", cutoff: 1, want: IndependenceReferenced},
		{name: "hint", mode: "assessment", cutoff: 2, want: IndependenceHinted},
		{name: "AI", mode: "assessment", cutoff: 3, want: IndependenceAIAssisted},
		{name: "solution", mode: "assessment", cutoff: 4, want: IndependenceGuided},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := CalculateIndependence(test.mode, events, test.cutoff); got != test.want {
				t.Fatalf("CalculateIndependence(%q, cutoff %d) = %q, want %q", test.mode, test.cutoff, got, test.want)
			}
		})
	}
}
