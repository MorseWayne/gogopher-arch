package main

import (
	"flag"
	"fmt"
	"os"

	"checkcfg/internal/config"
)

func main() {
	path := flag.String("config", "", "path to targets JSON")
	flag.Parse()
	if *path == "" {
		fmt.Fprintln(os.Stderr, "-config is required")
		os.Exit(2)
	}
	if _, err := config.Load(*path); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
