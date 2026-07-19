# gocheck-hub project read-through cache

完成 `internal/projectcache` 与表格测试。Source 是唯一 source of truth：fresh positive/negative hit 不访问 Source；miss 从 Source 装载并使用不同 TTL；Cache Get/Set 故障时读路径降级到 Source。并发 cold miss 必须合并为一次 Source Get，等待者仍响应各自 context。

Update 必须先写 Source，成功后再 Delete cache；Source 失败不能失效，失效失败必须明确返回“truth 已更新但 cache 未失效”的错误。解释 cache-aside、source of truth、negative cache、invalidation 与 degradation 的一致性边界。
