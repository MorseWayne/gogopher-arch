# 为并发任务接入取消信号

实现 `Map`：最多使用 `workers` 个 worker 调用 `transform`，结果保持输入顺序。

当父 `ctx` 取消或任一 `transform` 返回错误时，必须停止派发新任务、通知其他 worker，并在所有已启动 goroutine 退出后返回。返回错误应保留原始错误，便于 `errors.Is` 判断。
