# 独立评估：gocheck-hub 可观测请求与健康检查

补全 `internal/observability/service.go`，交付 gocheck-hub 的 HTTP 可观测边界：

- `New` 拒绝缺失依赖、空 route template 或不可用 clock/request ID generator；
- `Middleware` 复用有效的 `X-Request-ID`，否则生成新 ID，并同步写入 response、context 与结构化 `Event`；
- `Metrics.Observe` 只接收规范 method、构造时给定的 route template、status class 与 duration，禁止原始 path、query、request ID 或底层错误成为 label；
- `Liveness` 只证明进程能够服务，不访问依赖；
- `Readiness` 使用请求 context 检查依赖，失败返回 `503` 和稳定正文，不泄露底层错误；
- 捕获 handler 的隐式 `200`、显式状态码和响应字节数。

在 `service_test.go` 至少列出 8 个命名场景，并提交说明，串起 `structured log`、`request ID`、`low cardinality`、`liveness` 与 `readiness`。
