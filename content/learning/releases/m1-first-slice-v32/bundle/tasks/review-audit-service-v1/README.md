# 变式复习：审计事件服务

完成审计事件服务，并重新证明两条边界：

- 使用方 `Service` 只要求 `Sink.Write(Event) error`，替身无需实现刷新或关闭；
- `GroupBy` 是可复用的泛型函数，接受任意元素类型和 comparable key，重复 key 对应的元素都应保留。

`Append` 按顺序写入事件，遇到首个错误停止并保留错误。
