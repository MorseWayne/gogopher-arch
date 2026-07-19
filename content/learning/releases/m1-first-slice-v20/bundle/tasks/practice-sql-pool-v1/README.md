# 练习：配置 SQL 连接池

完成 `sqlpool.Configure`。拒绝 nil DB、非正数配置和 `MaxIdle > MaxOpen`，然后调用 `SetMaxOpenConns`、`SetMaxIdleConns`、`SetConnMaxLifetime`、`SetConnMaxIdleTime`。成功返回 nil。
