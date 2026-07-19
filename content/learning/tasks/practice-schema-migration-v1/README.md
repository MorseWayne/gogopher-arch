# 练习：建立首个 schema migration

完成 project 表 migration：使用 UUID 主键，owner_id、name、created_at 非空，active 非空且默认 true；用 CHECK 拒绝空白 name，并保证同一 owner 内 name 唯一。再为列表查询建立 `(owner_id, active, created_at DESC)` 索引。
