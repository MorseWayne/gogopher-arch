package health

type Result struct {
	Name string
	OK   bool
}

type Summary struct {
	Names  []string
	Failed int
}

func Summarize(results []Result) Summary {
	// TODO: 按公开契约实现。
	return Summary{}
}

func (s Summary) ExitCode() int {
	return 0
}
