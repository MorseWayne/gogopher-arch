# 独立评估：外部 probe client

实现 `Client.Probe`：构造 `GET /v1/probe?target=...`，使用调用方 Context 与 `Accept: application/json`，通过注入的 `http.Client` 只发送一次请求。

200 时最多读取 `maxBody+1` 字节，超限返回 `ErrBodyTooLarge`，否则解析 `{"status":"..."}`。429 返回 `ErrRateLimited`，其他 4xx 返回 `ErrRejected`，5xx 返回 `ErrUpstream`。所有响应都必须关闭 Body，不得自动重试。补充至少五个命名 case。
