# 变式复习：缩略图流水线

实现 `Generate`，用固定数量的 worker 调用 `render`，并通过 channel 返回所有缩略图。

要求与生产任务一致：每个路径处理一次，并发数不超过 `workers`，发送完成后关闭 channel，返回前后不遗留无法退出的 goroutine。`workers <= 0` 或空输入返回关闭的 channel。
