# 延迟变式：从空白工作区交付 endpointaudit

不要复制 gocheck 的包名和数据结构。请重新从契约拆解输入、并发执行、输出和命令组装。

- module path：`endpointaudit`；
- `internal/spec`：`Check{Name, URL, AcceptStatus}`、`Document{Checks}` 和 `Load(path)`；拒绝空列表、重复/空名称、非 HTTP(S) URL、非 100–599 的期望状态码，保留文件错误链；JSON 字段为 `name`、`url`、`accept_status`、`checks`；
- `internal/probe`：最小 `Client`、`Check`、带 JSON tag 的 `Result{Name, Expected, Actual, Error}`，以及 `All(ctx, client, checks, workers, timeout) ([]Result, error)`；要求稳定顺序、有界并发、逐请求超时、关闭 Body、取消后等待退出；
- `internal/output`：`Text([]probe.Result)` 与 `JSON([]probe.Result)`；文本行是 `<name>\t<pass|fail|error>\t<expected>\t<actual>\n`；
- `internal/app`：`Dependencies{Client probe.Client}` 和 `Run(ctx, args, stdout, stderr, dependencies) int`；flags 为 `-spec`、`-deadline`、`-workers`、`-output=text|json`；全部符合期望返回 0，探测不符或请求失败返回 1，参数/配置错误返回 2；
- `cmd/endpointaudit` 负责真实依赖和退出码；自行添加含 3 个命名 case 的表格驱动测试、项目 README 与 examples 配置。

提交小结不少于 120 个字符，并包含 `package`、`context`、`httptest`、`exit code`。本题不提供公开契约测试，提交时使用隐藏集成测试与 Race Detector 验证迁移能力。
