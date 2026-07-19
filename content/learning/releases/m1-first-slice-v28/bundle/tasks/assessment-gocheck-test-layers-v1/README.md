# 独立评估：建立 gocheck-hub 分层测试

生产代码已冻结。补全三个可编辑测试文件：

- `service_test.go`：用确定性 clock、ID source 与 store fake 覆盖至少 6 个命名 unit cases，不使用 `time.Sleep`；
- `handler_test.go`：用 `httptest.NewRequest`、`httptest.NewRecorder` 和 service fake 覆盖 HTTP contract；
- `postgres_integration_test.go`：读取 `TEST_DATABASE_DRIVER` 与 `TEST_DATABASE_URL`，没有配置时明确 Skip；配置存在时用 `sql.Open`、`PingContext`、`t.Cleanup` 和 `ExecContext` 建立隔离 fixture。CI 中这条测试必须连接真实 PostgreSQL。

`testdata/checks.json` 是跨层 fixture。Sandbox 运行 unit/handler 与隐藏测试，并用 AST 验证真实 PostgreSQL 测试入口；项目 Compose 回归负责提供真实 PostgreSQL 环境。

提交说明需包含 `unit test`、`httptest`、`fixture`、`PostgreSQL` 与 `deterministic clock`。
