# 资源生命周期栈

完成 `lifecycle.Stack`：`Push` 注册关闭函数，`Close` 以 LIFO 顺序传入同一 Context 执行，不因单个错误停止，并使用 `errors.Join` 返回全部错误。重复 Close 不得再执行 closer。
