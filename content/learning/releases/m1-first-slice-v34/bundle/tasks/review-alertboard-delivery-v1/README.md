# alertboard 延迟复验

只修改 `internal/alertboard/service.go` 与 `service_test.go`。保留既有 module、migration 和发布资产。

`GET /v1/alerts/{id}` 与 `POST /v1/alerts/{id}/ack` 使用 `X-API-Key`；`NewService(store, cache, credentials)` 必须保存 digest 而不是原始 key，认证先于 store/cache。GET 用 tenant-aware cache-aside，cache 失败降级；ack 先提交 Store，再 Delete，Store 失败不能失效。`Run(ctx, concurrency, deliver)` 用固定数量 worker 消费 `Next`，传递取消并等待退出。
