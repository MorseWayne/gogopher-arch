# 变式复验：alerts schema

完成两步 forward-only migration。0001 创建 alert_rules：UUID 主键、tenant 外键、destination、severity、active 默认值、时间戳、destination 非空白 CHECK，以及 tenant 内 destination 唯一。0002 只增加 `(tenant_id, active, severity, created_at DESC)` 索引。

生产查询按 tenant、active 和 severity 过滤、按 created_at 倒序并限制条数；EXPLAIN 查询保持同一形状。不得使用破坏性 SQL，并补充至少三个契约测试。
