package temperature

type Celsius float64
type Band string

const (
	BandLow    Band = "low"
	BandNormal Band = "normal"
	BandHigh   Band = "high"
)

type Reading struct {
	Value Celsius
	Band  Band
}

type Summary struct {
	Readings []Reading
	Alerts   int
}

func ClassifyReadings(values []Celsius, low, high Celsius) Summary {
	// TODO: 按 README 契约实现。
	return Summary{}
}
