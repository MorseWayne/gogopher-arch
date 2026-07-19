# 独立交付 checks JSON API

实现 `checkapi.NewHandler(Creator)`，提供 `POST /checks`。请求字段为 `target`、`timeout_ms`；成功返回 `201` 和 `id`、`target`、`timeout_ms`。严格拒绝未知字段、多余 JSON、空 target 和范围外 timeout。

错误协议固定为 `{"error":{"code":"...","message":"..."}}`：输入错误映射 `400 invalid_request`，`ErrCheckExists` 映射 `409 check_exists`，其他错误映射 `500 internal_error`，不得把内部错误文本返回给客户端。请在 `handler_test.go` 编写至少三个命名表格用例。
