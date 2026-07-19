# 状态报告 module

独立完成 `statusmodule`。`health.Summarize` 按输入顺序保留名称并统计失败项；`Summary.ExitCode` 在全部成功时返回 0，有失败时返回 1；`cmd/status` 只能依赖 `health` 的导出 API。

`health/report.go` 中每个导出类型、函数和方法都要有以声明名开头的 doc comment，文件还要有以 `Package health` 开头的 package comment。不要导出实现细节或让 `health` 依赖 `cmd/status`。
