package main

import (
	"fmt"
	"os"

	"statusmodule/health"
)

func main() {
	summary := health.Summarize([]health.Result{{Name: "api", OK: true}})
	fmt.Printf("checks=%d failed=%d\n", len(summary.Names), summary.Failed)
	os.Exit(summary.ExitCode())
}
