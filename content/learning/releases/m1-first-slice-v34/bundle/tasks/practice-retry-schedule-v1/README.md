# 练习有界重试调度

实现 `Policy.Next`。首次失败等待 `Base`，之后指数退避但不得超过 `Max`；达到 `MaxAttempts` 后停止重试。无效策略和小于 1 的 attempt 都应返回 `ok=false`，计算不能因 duration 溢出而变成负数。
