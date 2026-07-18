package report

import (
	"fmt"
	"io"
)

type Entry struct {
	Name  string
	Value int
}

func Render(entries []Entry) string {
	output := ""
	for index := 0; index < len(entries)-1; index++ {
		output += fmt.Sprintf("%s=%d\n", entries[index].Name, entries[index].Value)
	}
	return output
}

func LogSummary(writer io.Writer, rendered string) {
	fmt.Fprintf(writer, "rendered=%d\n", rendered)
}
