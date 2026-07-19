# 构建有界 worker pool

实现 `Process`。它返回一个结果 channel，由固定数量的 worker 消费任务。

公开契约：

- 每个输入恰好处理一次，`Result.Index` 对应原始下标；
- 同时执行的 `transform` 不超过 `workers`；
- 所有结果发送完成后关闭返回 channel；
- `workers <= 0` 或空输入应返回已经关闭的 channel；
- 调用方消费完 channel 时，不应留下 worker、producer 或 closer goroutine。
