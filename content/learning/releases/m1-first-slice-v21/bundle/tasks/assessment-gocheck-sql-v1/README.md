# 独立评估：gocheck-hub SQL storage

完成 `internal/checks/postgres/repository.go`，只使用标准库 `database/sql`。

- `ConfigurePool` 拒绝 nil DB、非正数配置和 MaxIdle 大于 MaxOpen，并设置四项 pool 参数。
- `NewRepository` 拒绝 nil DB。
- `Create` 使用 `ExecContext` 和占位符参数写入 `checks(id,target)`。
- `List` 使用 `QueryContext` 读取 `id,target`，立即 defer `Rows.Close`，逐行 Scan，循环后返回 `Rows.Err`。
- `Find` 使用 `QueryRowContext` 按 ID 读取并 Scan；只把 `sql.ErrNoRows` 映射为 `checks.ErrNotFound`。
- 所有 driver/Scan/Context 错误保持可用 `errors.Is` 识别。

在 `repository_test.go` 写至少三个命名 case 和 `t.Run`。提交解释不少于 120 字，并包含 `connection pool`、`QueryContext`、`Scan`、`Close`。
