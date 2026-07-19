package textstats

type Category string

const (
	CategoryAccepted Category = "accepted"
	CategoryEmpty    Category = "empty"
	CategoryTooLong  Category = "too_long"
)

type Entry struct {
	Text     string
	Runes    int
	Category Category
}

type Report struct {
	Entries    []Entry
	Accepted   int
	Rejected   int
	TotalRunes int
}

func Analyze(lines []string, maxRunes int) Report {
	// TODO: 按 README 契约实现。
	return Report{}
}
