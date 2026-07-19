# alert delivery worker 延迟复验

把持久化 worker 契约迁移到告警投递：固定并发领取 Delivery，重复幂等键只确认不重复发送；成功 Ack，临时发送故障按 `RetryDelay` 重新调度，永久故障或达到 `MaxAttempts` 后 Fail。`Run` 在 context 取消后停止领取并等待发送者退出，过期 lease 由 Queue 在新进程中重新发放。

补充表格测试并解释 backpressure、idempotency、lease、retry 与 context 在新题材中的对应关系。
