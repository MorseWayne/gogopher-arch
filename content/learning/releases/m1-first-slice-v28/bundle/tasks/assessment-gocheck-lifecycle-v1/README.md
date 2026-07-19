# gocheck-hub composition root 与 graceful shutdown

完成 `cmd/gocheckhub/main.go` 与表格测试。`run` 必须先校验 Address、DSN 和 ShutdownTimeout，再依次 OpenDatabase、BuildHandler、NewServer 并 Serve。初始化失败、Serve 失败和 Context 取消都必须逆序释放已取得的资源。

取消时使用 `context.WithTimeout(context.Background(), config.ShutdownTimeout)` 先调用 Server.Shutdown，再 Database.Close；两者错误用 `errors.Join` 保留。main 使用 `signal.NotifyContext` 监听 SIGINT 和 SIGTERM。提交说明 configuration、dependency injection、SIGTERM、Shutdown 与 reverse order。
