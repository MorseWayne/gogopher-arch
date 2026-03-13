package common

import "time"

// SandboxRequest 定义了发送给沙盒引擎的任务
type SandboxRequest struct {
	ID        string `json:"id"`
	Code      string `json:"code"`
	Language  string `json:"language"`
	Timeout   int    `json:"timeout"` // 秒
}

// SandboxResponse 定义了沙盒执行后的返回结果
type SandboxResponse struct {
	ID         string        `json:"id"`
	Stdout     string        `json:"stdout"`
	Stderr     string        `json:"stderr"`
	ExitCode   int           `json:"exit_code"`
	Duration   time.Duration `json:"duration"`
	Status     string        `json:"status"` // "success", "error", "timeout", "panic"
}
