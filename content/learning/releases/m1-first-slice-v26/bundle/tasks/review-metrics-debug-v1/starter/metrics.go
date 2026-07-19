package metrics

import (
	"fmt"
	"io"
)

type Metric struct {
	Name  string
	Value int
}

func Render(values []Metric) string {
	output := ""
	for index := 1; index < len(values); index++ {
		output += fmt.Sprintf("%s:%d\n", values[index].Name, values[index].Value)
	}
	return output
}

func Debug(writer io.Writer, rendered string) {
	fmt.Fprintf(writer, "metrics=%d\n", rendered)
}
