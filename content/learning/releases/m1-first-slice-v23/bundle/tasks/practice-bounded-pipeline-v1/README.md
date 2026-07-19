# 受控并发映射

实现 `Map`：使用不超过 `workers` 个 goroutine 处理全部输入，并按输入顺序返回结果。

约束：

- `workers <= 0` 或输入为空时返回空切片；
- 不能为每个输入启动一个 goroutine；
- 返回前所有 worker 都必须退出；
- `transform` 可能耗时，但不会 panic。
