# 为 target API 建立稳定 JSON 契约

补全 `targetapi.Handler`：`POST /targets` 接收 `name`、`url` 和 `interval_seconds`。只接受一个 JSON 对象，拒绝未知字段；名称不能为空，URL 必须使用 HTTP(S)，间隔必须在 1–3600 秒。成功返回 `201`，失败统一返回 `{"error":{"code","message"}}`。
