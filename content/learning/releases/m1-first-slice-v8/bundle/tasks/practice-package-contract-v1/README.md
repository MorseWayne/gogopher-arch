# 拆出可复用的报告包

补全 `internal/report` 的 `Summarize`，按输入顺序保留名称并统计失败项。`cmd/status` 只能通过 `report` 的导出 API 使用它，不要把命令行职责反向放进领域包。

运行 Build 确认整个 module 的依赖方向可编译，运行 Test 验证包契约。
