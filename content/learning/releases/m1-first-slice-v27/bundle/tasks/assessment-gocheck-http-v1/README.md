# 独立评估：gocheck-hub HTTP 入口

完成 `httpserver` 包，把 M1 的 gocheck 演进为第一个可运行的服务边界。只使用标准库，不引入第三方 router 或 middleware。

## Handler 契约

`NewHandler` 必须拒绝缺失的 `LookupTarget` 或 `NextRequestID`，并返回一个独立 `http.ServeMux` 组成的 handler：

- `GET /healthz` 返回 `204 No Content`；
- `GET /targets/{id}` 调用 `LookupTarget(ctx, id)`，成功时返回 `200`、`Content-Type: text/plain; charset=utf-8` 和 `name\n`，不存在时返回 404；
- method 不匹配时保留 `http.ServeMux` 的 405 语义；
- request ID middleware 优先复用去除首尾空白后非空的 `X-Request-ID`，否则调用 `NextRequestID`；它必须在调用下游前把 ID 放入新的 request context，并在响应头写入 `X-Request-ID`；`RequestID` 用来读取该值。

## Server 契约

`NewServer` 拒绝 nil handler 或任何非正数 timeout，并把 `Timeouts` 的四个值逐一映射到 `http.Server` 的 `ReadHeaderTimeout`、`ReadTimeout`、`WriteTimeout` 和 `IdleTimeout`。

`Serve` 使用传入 listener 启动 server。server 自己结束时返回真实错误；`ctx.Done()` 后用独立的 `shutdownTimeout` 调用 `server.Shutdown`，等待活动请求结束，并把正常的 `http.ErrServerClosed` 收敛为 nil。不要用 `Close` 冒充优雅关闭，也不要让函数在 `Shutdown` 完成前返回。

## 你必须补充的测试与解释

修改 `httpserver/server_test.go`，至少写一张包含 3 个命名 case 且使用 `t.Run` 的表格测试，覆盖路由或 middleware 边界。公开 smoke test 只覆盖最短成功路径，Submit 还会在真实 loopback listener 上验证完整契约。

提交说明不少于 120 字，并明确串起 `ServeMux`、`context`、`timeout`、`Shutdown`：一次请求如何经过 middleware 与 handler，以及停止信号到达后 listener、空闲连接、活动请求和 `Serve` 返回值如何变化。
