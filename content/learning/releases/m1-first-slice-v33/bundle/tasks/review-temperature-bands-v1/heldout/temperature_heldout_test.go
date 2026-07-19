package temperature

import "testing"

func TestClassifyReadingsHandlesBoundariesAndControlFlow(t *testing.T) {
	got := ClassifyReadings([]Celsius{-1.5, 0, 10, 10.5}, 0, 10)
	want := []Band{BandLow, BandNormal, BandNormal, BandHigh}
	if got.Alerts != 2 || len(got.Readings) != len(want) {
		t.Fatalf("ClassifyReadings() = %#v", got)
	}
	for index, band := range want {
		if got.Readings[index].Band != band || got.Readings[index].Value != []Celsius{-1.5, 0, 10, 10.5}[index] {
			t.Fatalf("reading %d = %#v, want band %q", index, got.Readings[index], band)
		}
	}
}

func TestClassifyReadingsReturnsZeroForInvalidRange(t *testing.T) {
	var band Band = BandNormal
	if band != "normal" {
		t.Fatalf("BandNormal = %q", band)
	}
	got := ClassifyReadings([]Celsius{1, 2}, 5, 5)
	if len(got.Readings) != 0 || got.Alerts != 0 {
		t.Fatalf("invalid range result = %#v, want zero summary", got)
	}
}
