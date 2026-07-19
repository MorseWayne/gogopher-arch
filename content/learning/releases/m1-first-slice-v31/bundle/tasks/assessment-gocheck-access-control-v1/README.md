# gocheck-hub project 访问控制

完成 `internal/projectapi/handler.go` 和自己的表格测试。接口为 `GET /v1/projects/{id}`：严格认证 Bearer API key，仅允许 owner 读取 project，并让跨租户访问与资源不存在共用 404 响应。只允许 1–64 位小写字母、数字、`-` 和 `_` 组成的 canonical ID，首字符必须是小写字母或数字。

审计器只能收到 `authentication_failed` 或 `resource_not_found` 这类稳定 reason，不得传入 token、Authorization header、project 内容或底层错误。提交时说明 authentication、authorization、constant-time、404 与 secret 边界。
