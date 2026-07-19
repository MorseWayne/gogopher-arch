package temperature

import "testing"

func TestClassifyReadingsBasicBands(t *testing.T) {
	got := ClassifyReadings([]Celsius{5, 20, 40}, 10, 30)
	if got.Alerts != 2 || len(got.Readings) != 3 || got.Readings[0].Band != BandLow || got.Readings[1].Band != BandNormal || got.Readings[2].Band != BandHigh {
		t.Fatalf("ClassifyReadings() = %#v", got)
	}
}
