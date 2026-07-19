# 练习：关联一次 HTTP 请求

补全 `telemetry/middleware.go`，让 middleware 为每次请求建立稳定的可观测记录：

- 没有 `X-Request-ID` 时生成 ID，并把它写入 response header 与 request context；
- 结构化事件记录 request ID、method、route template、status class 和 duration；
- 指标只使用 method、route template 与 status class，不使用原始 URL 或 request ID；
- 正确捕获隐式 `200` 和 handler 显式写出的状态码。

`route` 是注册路由时提供的模板（例如 `/api/v1/checks/{id}`），不能用 `request.URL.Path` 替代。
