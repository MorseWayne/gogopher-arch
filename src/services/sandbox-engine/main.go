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
	executeHandler := handlers.NewExecuteHandler(r)

	mux := http.NewServeMux()
	mux.Handle("/execute", executeHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	port := ":8081"
	// Sandbox engine uses fixed port, config.Port is for gateway
	fmt.Printf("Gogopher Arch Sandbox Engine listening on %s...\n", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
	_ = cfg // silence unused variable warning for now
}
