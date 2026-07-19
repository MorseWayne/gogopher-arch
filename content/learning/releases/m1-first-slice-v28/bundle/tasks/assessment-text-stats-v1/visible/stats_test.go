package textstats

import "testing"

func TestAnalyzeClassifiesASCII(t *testing.T) {
	got := Analyze([]string{"go", "", "gopher"}, 2)
	if got.Accepted != 1 || got.Rejected != 2 || got.TotalRunes != 2 {
		t.Fatalf("Analyze() summary = %#v", got)
	}
	if len(got.Entries) != 3 || got.Entries[0].Category != CategoryAccepted || got.Entries[1].Category != CategoryEmpty || got.Entries[2].Category != CategoryTooLong {
		t.Fatalf("Analyze() entries = %#v", got.Entries)
	}
}
