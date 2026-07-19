# 练习：拆开 checks 与 memory

完成 `internal/checks/service.go` 和 `internal/checks/memory/repository.go`。

- `checks.Repository` 是 use case 消费的最小接口，保留在 `checks` 包。
- `NewService` 拒绝 nil repository 或 nil `nextID`。
- `Service.Create` 去除 target 首尾空白，空值返回 `ErrInvalidTarget`；有效输入使用 `nextID` 生成 ID，再把完整 `Check` 交给 repository。repository 错误原样返回。
- memory repository 的 constructor 返回可用实例；`Create` 响应 context 取消，并以忽略大小写的 target 判断重复，重复时返回 `checks.ErrCheckExists`。实现必须可被并发调用。

不要在 use case 中导入具体 storage，也不要用包级可变变量藏依赖。
