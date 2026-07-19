# alert rule cache 延迟复验

把 cache-aside 契约迁移到 alert rule。Repository 是 source of truth；fresh positive/negative hit、miss 装载、cache outage 降级、并发 miss 合并和 context 取消必须保持。Save 先提交 Repository，再失效缓存，分别暴露 truth 写失败和失效失败。补表格测试并解释 source of truth、negative cache、invalidation 与 degradation。
