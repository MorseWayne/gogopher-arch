# 用函数和分支生成状态消息

补全 `readiness` 和 `report`：

- `readiness(true)` 返回 `ready`；
- `readiness(false)` 返回 `retry`；
- `report("build", true)` 返回 `build: ready`。

不要把 `build` 写死。先运行公开测试，根据失败信息逐步修正，再运行 Build 和 Vet。
