# gocheck-hub alert rule 访问控制变式

完成 `internal/alertaccess/handler.go` 和自己的表格测试。接口为 `DELETE /v1/alert-rules/{id}`：严格认证 Bearer API key，只允许 owner 删除 alert rule，跨租户访问和资源不存在共用 404。只允许 1–64 位小写字母、数字、`-` 和 `_` 组成的 canonical ID，首字符必须是小写字母或数字。

审计器只能收到稳定 reason，不得记录 token、Authorization header、rule 内容或 store 错误。提交时说明 authentication、authorization、constant-time、404 与 secret 边界。
