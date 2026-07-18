package report

type Result struct {
	Name string
	OK   bool
}

type Summary struct {
	Names  []string
	Failed int
}

func Summarize(results []Result) Summary {
	// TODO: 只处理报告数据，不依赖 cmd 包。
	return Summary{}
}
