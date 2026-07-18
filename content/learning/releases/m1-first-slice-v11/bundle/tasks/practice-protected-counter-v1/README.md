# 修复共享计数器

`Counter` 会被多个 goroutine 同时调用。修复它，使 `Add` 与 `Value` 形成清晰的同步边界，并在并发累加后得到准确结果。

可以选择 mutex、atomic 或单 owner goroutine，但要让选择与状态形态匹配：这里是一项整数累加职责，不要依赖 sleep 或扩大 channel buffer 掩盖问题。
