package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/MorseWayne/gogopher-arch/src/pkg/common"
)

func getSandboxURL() string {
	url := os.Getenv("SANDBOX_URL")
	if url == "" {
		return "http://localhost:8081/execute"
	}
	return url
}

func main() {
	sandboxURL := getSandboxURL()
	fmt.Printf("Gogopher Arch Gateway using Sandbox URL: %s\n", sandboxURL)

	http.HandleFunc("/api/v1/execute", func(w http.ResponseWriter, r *http.Request) {
		// 允许跨域 (MVP 阶段)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			return
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusInternalServerError)
			return
		}

		fmt.Printf("[%s] Forwarding execution request to sandbox...\n", time.Now().Format(time.RFC3339))
		resp, err := http.Post(sandboxURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			http.Error(w, "Sandbox engine unreachable: "+err.Error(), http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		io.Copy(w, resp.Body)
	})

	port := ":8080"
	fmt.Printf("Gogopher Arch Gateway listening on %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Failed to start gateway: %v\n", err)
	}
}
