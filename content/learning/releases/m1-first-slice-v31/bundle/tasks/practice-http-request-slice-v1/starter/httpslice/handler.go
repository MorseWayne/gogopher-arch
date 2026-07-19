package httpslice

import (
	"context"
	"net/http"
)

type TargetLookup func(context.Context, string) (string, bool)

type RequestIDGenerator func() string

func NewHandler(lookup TargetLookup, nextRequestID RequestIDGenerator) http.Handler {
	// TODO: 组装 request ID middleware、ServeMux 与两个 route。
	return http.NotFoundHandler()
}

func RequestID(ctx context.Context) string {
	// TODO: 返回中间件写入 context 的 request ID。
	return ""
}
