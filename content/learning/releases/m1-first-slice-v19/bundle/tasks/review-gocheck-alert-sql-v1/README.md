# 异题复验：alert SQL storage

完成 `internal/alerts/sqlstore/repository.go`。配置四项 pool 参数；用 `ExecContext` 保存 `alerts(id,destination)`；用 `QueryContext` 列出并正确 Close/Scan/Err；用 `QueryRowContext` 按 ID 查找，只把 `sql.ErrNoRows` 映射为 `alerts.ErrNotFound`。所有取消和 driver 错误原样传播。

补充至少三个命名 case 的表格测试。解释不少于 120 字，包含 `connection pool`、`QueryContext`、`Scan`、`Close`。
