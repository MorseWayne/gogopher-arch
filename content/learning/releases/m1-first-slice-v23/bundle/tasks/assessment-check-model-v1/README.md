# 检查目标领域模型

独立完成 `Target`：`NewTarget` 拒绝空白名称和小于 1 的失败上限，并保存去除两端空白后的名称。`RecordFailure` 增加连续失败次数，`RecordSuccess` 清零；达到上限时 `State()` 返回 `StateOpen`，否则返回 `StateReady`。`Label()` 输出 `<name>:<state>`。

字段保持私有，由构造函数和行为方法维护合法状态。
