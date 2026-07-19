# 独立评估：gocheck-hub 业务边界

把创建 check 的流程分成 `internal/checks` use case、`internal/checks/memory` storage、`internal/httpapi` transport，并在 `cmd/gocheckhub/main.go` 显式组装。只使用标准库。

## Use case 与 storage

- `checks.Repository` 必须保留在消费 storage 的 `checks/service.go` 中，并且只声明 `Create(context.Context, Check) error`。
- `NewService` 拒绝 nil repository 或 nil `nextID`。
- `Service.Create` 去除 target 首尾空白；空 target 返回 `ErrInvalidTarget`。有效输入先用 `nextID` 生成 ID，再持久化完整 `Check`；storage 错误原样返回。
- memory repository 由 `NewRepository` 创建，支持并发调用和 context 取消。target 按忽略大小写唯一，重复时返回 `checks.ErrCheckExists`。

## Transport

`httpapi.Creator` 定义在 transport 中，只包含 transport 使用的 `Create` 方法。`NewHandler` 拒绝 nil creator，并提供 `POST /checks`：

- body 必须是恰好一个 JSON 值，未知字段、空 target 返回 `400` 与 `{"error":{"code":"invalid_request","message":"request is invalid"}}`；
- 成功返回 `201` 与 `{"id":"...","target":"..."}`；
- `ErrCheckExists` 返回 `409/check_exists`；未知错误返回不泄漏原始文本的 `500/internal_error`；
- 所有响应使用 `application/json; charset=utf-8`。

## Wiring、测试与解释

`cmd/gocheckhub/main.go` 的 `buildHandler` 必须直接调用 `memory.NewRepository`、`checks.NewService`、`httpapi.NewHandler`，将具体依赖留在 composition root；业务和 transport 包不得反向导入 cmd 或具体 storage。

修改 `internal/checks/service_test.go`，至少写三个命名 case 并使用 `t.Run`。提交说明不少于 120 字，必须明确包含 `transport`、`use case`、`storage`、`constructor`，解释接口归属与依赖组装，而不是复述目录名。
