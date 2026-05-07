package common

import (
	"encoding/json"
	"net/http"
)

// ValidationError 表示请求参数校验失败
type ValidationError struct {
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return e.Message
}

// InternalError 表示内部服务错误
type InternalError struct {
	Message string `json:"message"`
}

func (e InternalError) Error() string {
	return e.Message
}

// WriteError 向响应写入统一的 JSON 错误格式
func WriteError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
}