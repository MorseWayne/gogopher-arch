package main

import (
	"fmt"

	"packagepractice/internal/report"
)

func main() {
	summary := report.Summarize([]report.Result{{Name: "api", OK: true}})
	fmt.Printf("checks=%d failed=%d\n", len(summary.Names), summary.Failed)
}
