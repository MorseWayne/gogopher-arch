# 异题复验：gocheck-hub alert 边界

在空白业务切片中实现 alert rule 的 transport、use case、storage 与 composition root。只使用标准库。

- `internal/alerts/manager.go` 定义消费方接口 `Store`，只有 `Save(context.Context, Rule) error`；`NewManager` 拒绝 nil store 或 nil `nextID`。`Publish` trim destination，空值返回 `ErrInvalidDestination`，否则分配 ID 并保存。
- `internal/alerts/memory/store.go` 提供并发安全且响应 context 取消的实现；destination 忽略大小写唯一，重复返回 `alerts.ErrRuleExists`。
- `internal/alertapi/handler.go` 自己定义单方法 `Publisher`。`POST /alerts` 严格接收恰好一个 `{"destination":"..."}`，成功返回 `201`；非法请求映射 `400/invalid_request`，重复映射 `409/alert_exists`，未知错误映射不泄漏细节的 `500/internal_error`。所有 body 都是 JSON，Content-Type 为 `application/json; charset=utf-8`。
- `cmd/gocheckalerts/main.go` 的 `buildHandler` 直接调用 `memory.NewStore`、`alerts.NewManager`、`alertapi.NewHandler`。

在 `internal/alerts/manager_test.go` 写至少三个命名 case 和 `t.Run`。提交说明不少于 120 字，并包含 `transport`、`use case`、`storage`、`constructor`，说明接口为什么属于使用方以及 main 为什么是唯一组装点。
