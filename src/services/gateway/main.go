package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MorseWayne/gogopher-arch/src/pkg/common"
)

const sandboxURL = "http://localhost:8081/execute"

func main() {
	// 简单的路由设置
	http.HandleFunc("/api/v1/execute", executeHandler)
	
	// 静态文件服务 (生产环境可以使用这个，MVP 开发时可以由 Vite 处理)
	// http.Handle("/", http.FileServer(http.Dir("./web/dist")))

	port := ":8080"
	fmt.Printf("Gogopher Arch Gateway listening on %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Failed to start gateway: %v\n", err)
	}
}

func executeHandler(w http.ResponseWriter, r *http.Request) {
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

	// 1. 读取前端请求
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}

	// 2. 转发给 Sandbox Engine
	fmt.Printf("[%s] Forwarding execution request to sandbox...\n", time.Now().Format(time.RFC3339))
	resp, err := http.Post(sandboxURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "Sandbox engine unreachable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	// 3. 将结果返回给前端
	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}
