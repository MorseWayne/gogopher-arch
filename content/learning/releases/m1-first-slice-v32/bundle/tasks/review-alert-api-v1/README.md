# 迁移实现 alert rules API

不使用提示，完成 `alertapi.NewHandler(Creator)` 的 `POST /rules`。请求为 `destination` 和 `threshold`；destination 必须是 HTTP(S) URL，threshold 为 1–100。成功返回 `201`；输入错误、`ErrRuleExists` 和未知错误分别映射为 `400 invalid_request`、`409 rule_exists`、`500 internal_error`，内部错误不可泄漏。补充至少三个命名表格测试。
