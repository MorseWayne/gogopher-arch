# 可取消的并发检查器

实现 `CheckAll`：使用不超过 `workers` 个 worker 对 targets 执行 `check`。

- 全部成功返回 `nil`；
- 父 context 取消或超时时，停止派发并返回 `ctx.Err()`；
- 任一检查失败时取消同批其他检查，返回的错误保留该错误；
- 函数返回前，producer、worker、closer 以及已经进入 `check` 的调用都必须退出；
- `workers <= 0` 返回明确错误，空输入返回 `nil`。

不要用 sleep 等待 goroutine，也不要在错误路径提前返回而放弃清理。
