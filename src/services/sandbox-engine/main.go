package main

import (
	"fmt"
	"net/http"

	"github.com/MorseWayne/gogopher-arch/src/internal/config"
	"github.com/MorseWayne/gogopher-arch/src/services/sandbox-engine/internal/handlers"
	"github.com/MorseWayne/gogopher-arch/src/services/sandbox-engine/internal/runner"
)

func main() {
	cfg := config.Load()
	r := runner.New()
	h := handlers.NewExecuteHandler(r)

	port := ":8081"
	// Sandbox engine uses fixed port, config.Port is for gateway
	fmt.Printf("Gogopher Arch Sandbox Engine listening on %s...\n", port)
	if err := http.ListenAndServe(port, h); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
	_ = cfg // silence unused variable warning for now
}