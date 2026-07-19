# 延迟复验：jobwatch admin server

在不同题材下重新实现标准库 HTTP 服务边界。不能复制 gocheck-hub 的标识符或 route；请从下面的契约重新建立 request lifecycle。

`NewHandler` 拒绝缺失的 `Ready`、`LookupJob` 或 `NextTraceID`：

- `GET /readyz` 调用 `Ready(ctx)`；ready 时返回 204，否则返回 503；
- `GET /jobs/{id}` 调用 `LookupJob(ctx, id)`；找到时返回 `text/plain; charset=utf-8` 的 `status\n`，否则返回 404；
- method 不匹配保留 `http.ServeMux` 的 405；
- middleware 复用去除首尾空白后的非空 `X-Trace-ID`，缺失时调用 `NextTraceID`，并同时写入响应头和 request context；`TraceID` 读取该值。

`NewServer` 拒绝 nil handler 和任何非正数 timeout，把四个值映射到 `http.Server` 的对应字段。`Serve` 在 server 自行结束时返回真实错误；`ctx.Done()` 后使用独立的正数 `shutdownTimeout` 调用 `Shutdown`，等待活动请求结束，并把正常的 `http.ErrServerClosed` 转成 nil。

修改 `adminserver/server_test.go`，写一张至少包含 3 个命名 case 且调用 `t.Run` 的表格测试。提交说明不少于 120 字，并明确包含 `ServeMux`、`context`、`timeout`、`Shutdown`，解释请求路径和关闭路径，而不是罗列 API 名称。
