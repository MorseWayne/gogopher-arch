package textstats

import "testing"

func TestAnalyzeHandlesUnicodeAndControlFlow(t *testing.T) {
	got := Analyze([]string{"你好", "é", "世界好"}, 2)
	if got.Accepted != 2 || got.Rejected != 1 || got.TotalRunes != 3 {
		t.Fatalf("Analyze() summary = %#v", got)
	}
	if got.Entries[0].Runes != 2 || got.Entries[1].Runes != 1 || got.Entries[2].Category != CategoryTooLong {
		t.Fatalf("Analyze() entries = %#v", got.Entries)
	}
}

func TestAnalyzeUsesNamedCategoriesAndZeroDefaults(t *testing.T) {
	var category Category = CategoryEmpty
	if category != "empty" {
		t.Fatalf("CategoryEmpty = %q", category)
	}
	if got := Analyze(nil, 3); len(got.Entries) != 0 || got.Accepted != 0 || got.Rejected != 0 || got.TotalRunes != 0 {
		t.Fatalf("Analyze(nil) = %#v, want zero report", got)
	}
	got := Analyze([]string{"go", ""}, 0)
	if got.Accepted != 0 || got.Rejected != 2 || got.Entries[0].Category != CategoryTooLong || got.Entries[1].Category != CategoryEmpty {
		t.Fatalf("Analyze(maxRunes=0) = %#v", got)
	}
}
