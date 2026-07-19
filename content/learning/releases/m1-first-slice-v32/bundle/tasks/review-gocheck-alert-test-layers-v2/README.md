# 延迟复验：迁移 alert API 分层测试

在没有提示的情况下，为冻结的 alert slice 补齐三层测试：确定性 unit test、`httptest` handler contract、以及由 `TEST_DATABASE_URL` 驱动的真实 PostgreSQL integration test。使用 `testdata/alerts.json` 作为跨层 fixture，数据库测试必须通过 `sql.Open("pgx", ...)` 连接，执行 `PingContext`、注册 `t.Cleanup` 并用 `ExecContext` 建立隔离数据；CI 以 `go test -tags=integration ./...` 加载只读 pgx driver 注册文件。

提交说明需包含 `unit test`、`httptest`、`fixture`、`PostgreSQL` 与 `deterministic clock`。
