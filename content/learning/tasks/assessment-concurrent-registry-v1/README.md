# 并发安全的状态注册表

实现 `Registry`：

- `Record(key, delta)` 原子地累加指定 key；
- `Snapshot()` 返回调用时刻的独立副本，调用方修改副本不能影响注册表；
- 两个方法可被多个 goroutine 并发调用；
- 不要暴露内部 map，也不要用 sleep 组织同步。

完成代码后，在小结中说明为什么这里选择 mutex、atomic 或 channel ownership，并至少比较一种未采用方案的代价。
