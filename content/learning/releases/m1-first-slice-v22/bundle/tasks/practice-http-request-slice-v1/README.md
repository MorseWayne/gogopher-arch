# HTTP 请求纵向切片

补全 `httpslice.NewHandler` 与 `httpslice.RequestID`，只使用标准库完成一条可测试的 HTTP 请求链：

- `GET /healthz` 返回 `204 No Content`；
- `GET /targets/{id}` 调用 `TargetLookup`，找到时返回 `text/plain` 的 `name\n`，否则返回 404；
- 其他 method 由 `http.ServeMux` 返回 405；
- 请求带有非空 `X-Request-ID` 时复用它，否则调用 `RequestIDGenerator`；
- 中间件把 request ID 同时写入响应头和新的 request context，handler 通过 `RequestID(r.Context())` 读取。

先画出 `request → middleware → mux → handler → response`，再实现。用 `httptest.NewRequest` 和 `httptest.NewRecorder` 运行公开测试，不要启动真实端口。
