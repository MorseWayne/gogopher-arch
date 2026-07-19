# 独立评估：checks schema 与 migration

完成两步 forward-only migration。0001 创建 checks：UUID 主键、owner 外键、target 与 schedule、enabled 默认值、时间戳、非空白 target CHECK，以及 owner 内 target 唯一。0002 只增加 `(owner_id, enabled, created_at DESC)` 索引。

生产查询按 owner 和 enabled 过滤、按 created_at 倒序并限制条数；EXPLAIN 查询必须保持同一查询形状。不得使用 DROP、DELETE、TRUNCATE 或 down migration。补充至少三个 migration 契约测试。
