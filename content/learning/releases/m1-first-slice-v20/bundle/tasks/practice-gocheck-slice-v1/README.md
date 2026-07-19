# 把已学能力接成一个可运行切片

补全 `CheckAll` 和 `RenderText`，完成一个最小服务状态检查器：

- `CheckAll` 使用给定 `http.Client` 和 `context.Context` 请求全部目标；
- 并发数不得超过 `workers`，结果保持输入顺序；
- 每个响应都关闭 Body；取消后不再派发新工作，并等待已启动 worker 退出；
- `RenderText` 每行输出 `<name>\t<status>\t<status-code>\n`；2xx/3xx 为 `ok`，其他状态为 `fail`，请求错误为 `error` 和状态码 0。

先画出输入、执行、输出三段数据流，再实现。公开测试全部使用本地 `httptest.Server`，不依赖公网。
