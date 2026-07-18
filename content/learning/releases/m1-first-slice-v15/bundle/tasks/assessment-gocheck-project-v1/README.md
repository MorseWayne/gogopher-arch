# M1 独立交付：从空白工作区完成 gocheck

工作区只提供任务契约、公开测试和示例输入。请使用“新建文件”自行建立 module、包、命令和测试，不得依赖公网或第三方库。

## 项目结构契约

- 根目录 `go.mod` 的 module path 必须是 `gocheck`；
- `internal/config`：定义带 JSON tag 的 `Target{Name, URL}`、`Config{Targets}`，实现 `Load(path) (Config, error)`；拒绝空列表、空名称、重复名称和非 HTTP(S) URL，文件错误保留原始错误链；
- `internal/check`：定义 `Client`（只含 `Do(*http.Request)`）、`Target`、带 JSON tag 的 `Result{Name, URL, StatusCode, Error}`；实现 `All(ctx, client, targets, workers, timeout) ([]Result, error)`；保持输入顺序、有界并发、逐请求超时、关闭 Body，上游取消后等待 worker 退出并返回 `ctx.Err()`；
- `internal/report`：实现 `Text([]check.Result) string` 与 `JSON([]check.Result) (string, error)`；文本每行是 `<name>\t<ok|fail|error>\t<status-code>\n`，JSON 是 Result 数组加结尾换行；
- `internal/app`：定义 `Dependencies{Client check.Client}`，实现 `Run(ctx, args, stdout, stderr, dependencies) int`；支持 `-config`、`-timeout`、`-concurrency`、`-format=text|json`；全部成功返回 0，检查失败返回 1，参数或配置错误返回 2；
- `cmd/gocheck`：组装真实依赖并把 `Run` 的结果交给 `os.Exit`；
- 自行添加至少一份包含 3 个命名 case 和 `t.Run` 的表格驱动测试；
- 添加项目 `README.md`、构建命令和 `examples/targets.json`。

HTTP 2xx/3xx 视为成功，其他状态视为检查失败；请求错误的状态码为 0。提交时会运行 Build、Vet、公开测试、隐藏契约测试和 Race Detector。

完成小结不少于 120 个字符，并包含 `package`、`context`、`httptest`、`exit code`，说明你的依赖方向、取消路径、测试边界和进程契约。
