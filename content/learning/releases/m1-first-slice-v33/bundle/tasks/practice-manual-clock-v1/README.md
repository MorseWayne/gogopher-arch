# 练习：实现确定性手动时钟

补全 `testkit/clock.go`：

- `NewManualClock` 保存明确的起点；
- `Now` 返回当前测试时间；
- `Advance` 只接受非负 duration，并原子地推进时间；
- 并发调用 `Now` 与 `Advance` 不产生 data race。

这个 test fake 让服务测试不依赖 `time.Sleep`、墙上时钟或机器负载。
