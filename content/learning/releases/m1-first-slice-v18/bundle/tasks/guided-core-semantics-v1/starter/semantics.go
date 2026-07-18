package semantics

import "strings"

type Label string

type Summary struct {
	Accepted  []Label
	Rejected  int
	RuneCount int
}

func ClassifyNames(inputs []string, maxRunes int) Summary {
	var result Summary
	for _, raw := range inputs {
		name := strings.TrimSpace(raw)
		if name == "" || maxRunes <= 0 {
			result.Rejected++
			continue
		}
		count := len(name) // TODO: 这里统计的是字节，不是 Unicode code point。
		if count > maxRunes {
			result.Rejected++
			continue
		}
		result.Accepted = append(result.Accepted, Label(name))
		result.RuneCount += count
	}
	return result
}
