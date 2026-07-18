# 事件批次快照

实现 `SnapshotWindow` 与 `IndexLatest`。窗口边界非法时返回 nil；合法时返回独立 slice，调用方修改输入或输出都不能相互影响。`IndexLatest` 按事件 `Key` 建索引，同名事件以后出现者覆盖先出现者；空输入返回非 nil 空 map。

请按所有权和返回值契约实现，不要针对公开样例硬编码。
