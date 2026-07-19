# 延迟复验：迁移 alert API 可观测边界

不依赖提示，把 M2-12 的约束迁移到 gocheck-hub alert API：

- 用 request ID 关联 response、context 与 structured log；
- 指标只记录 method、`/api/v1/alerts/{id}` route template、status class 和 duration；
- 捕获隐式/显式状态与响应字节数，不把 alert ID、query、密钥或依赖错误放进 label；
- `/livez` 不探测依赖，`/readyz` 才检查 alert delivery 依赖并返回稳定结果；
- 用至少 8 个命名场景记录迁移后的边界。

提交说明必须包含 `structured log`、`request ID`、`low cardinality`、`liveness` 和 `readiness`。
