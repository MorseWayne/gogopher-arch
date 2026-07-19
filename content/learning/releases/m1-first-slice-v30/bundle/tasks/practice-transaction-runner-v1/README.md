# 练习：事务执行边界

实现 `WithinTx`。使用 `BeginTx` 和调用方 Context；回调成功后 Commit，回调返回错误时 Rollback，回调 panic 时先 Rollback 再继续 panic。nil DB 或 nil callback 必须返回错误。
