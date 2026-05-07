package main

import (
	"fmt"
	"net/http"

	"github.com/MorseWayne/gogopher-arch/src/internal/config"
	"github.com/MorseWayne/gogopher-arch/src/services/gateway/internal/routes"
)

func main() {
	cfg := config.Load()
	h := routes.New(cfg)

	fmt.Printf("Gogopher Arch Gateway listening on %s...\n", cfg.Port)
	if err := http.ListenAndServe(cfg.Port, h); err != nil {
		fmt.Printf("Failed to start gateway: %v\n", err)
	}
}