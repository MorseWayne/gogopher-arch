# gocheck-hub 持久化后台 worker

完成 `internal/checkworker` 与表格测试。worker 从持久化 Store 领取带 lease 的检查任务，固定数量 goroutine 形成背压；成功才 Ack，临时失败按确定性策略 Retry，永久失败或耗尽次数 Fail。Store 会把已完成幂等键标成 `Duplicate`，worker 必须确认但不得再次执行副作用。

`Run` 必须响应 context 取消并等待全部 goroutine 退出。进程重启后，Store 可以重新发放 lease 已过期的任务；worker 不得用进程内 channel 或 map 代替这个持久化边界。解释中说明 backpressure、idempotency、lease、retry 与 context 如何共同约束故障恢复。
