# 交付通知服务：最小接口与简单泛型

完成 `delivery` 包。

## 服务契约

- `Sender` 由当前使用方定义，只保留 `Send(Message) error`；
- `Service.Deliver` 按输入顺序发送消息，首个错误立即返回并保留；
- 不要让 `Service` 依赖具体 sender 类型。

## 泛型契约

把 `IndexBy` 改为真正的泛型函数：它应接受任意 `[]T` 和返回 `comparable` key 的函数，生成 `map[K]T`；重复 key 保留最后一个值，空输入返回可写的非 nil map。

评测会使用只实现 `Send` 的替身，并以不同于 `Recipient` 的类型调用 `IndexBy`。
