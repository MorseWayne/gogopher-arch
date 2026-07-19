# alert worker composition root 与 graceful shutdown

完成 `cmd/alertworker/main.go` 与表格测试。`run` 先校验 DSN、Concurrency 和 ShutdownTimeout，再 OpenStore、NewWorker 并 Run。初始化失败、Run 失败和 Context 取消都必须逆序释放资源。

取消时使用独立有界 Context 先 Worker.Shutdown，再 Store.Close，以 `errors.Join` 保留两个错误。main 使用 `signal.NotifyContext` 监听 SIGINT 和 SIGTERM。提交说明 configuration、dependency injection、SIGTERM、Shutdown 与 reverse order。
