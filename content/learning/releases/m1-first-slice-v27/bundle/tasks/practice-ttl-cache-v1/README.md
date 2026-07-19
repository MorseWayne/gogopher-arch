# 练习 TTL 与 negative cache

实现并发安全的内存 `Cache`。`hit` 表示 fresh entry，`found` 表示数据真相中是否存在；因此 negative cache 应返回 `hit=true, found=false`。到期 entry 必须在读取时删除，`Delete` 用于写后失效。所有时间来自注入的 `now`，测试不得依赖真实 sleep。
