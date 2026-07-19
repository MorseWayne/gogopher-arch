# M2 毕业项目：gocheck-hub

从空白工作区创建 `module gocheckhub.local/service`。公开与隐藏测试约定 `internal/hub` 暴露以下边界：

- `Project{ID,TenantID,Name}`、`Job{ID,TenantID,Target}`；
- `Store`：`CreateProject`、`Project`、`Ready`、`Claim`、`Complete`；
- `Cache`：`Get`、`Set`、`Delete`，cache key 必须包含 tenant；
- `NewService(store, cache, credentials, logger)`、`Handler()`、`RunWorker(ctx, concurrency, probe)`；
- 稳定错误 `ErrNotFound`、`ErrConflict`、`ErrNoJob`。

HTTP 契约：`POST /v1/projects`、`GET /v1/projects/{id}` 使用 `X-API-Key` 映射租户；严格 JSON、认证先于 lookup、跨租户统一 404。`/livez` 不依赖存储，`/readyz` 检查 source of truth，`/metrics` 只输出低基数 route。每个响应返回或透传 `X-Request-ID`。

Worker 固定并发、继承取消、等待所有 goroutine 退出；无任务时有界等待，不能忙循环。另需提供 `SQLStore` 的 `database/sql` adapter、两步 forward-only migration、README、非 root multi-stage Dockerfile、Makefile 与 CI。Sandbox 不提供 PostgreSQL driver，运行时默认可使用内存 adapter；SQL adapter 由集成环境注入已注册 driver 的 `*sql.DB`。
