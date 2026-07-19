# Miniflux category validation patch

基线：Miniflux v2.3.2，commit `51f2e0d8199ea8fa305081f6e175bba64b0ef94b`，Go 1.26.0。

编辑 `internal/validator/category.go` 与同包测试。create 和 update 必须共用同一规范化规则：`strings.TrimSpace`，空标题返回 `error.title_required`，超过 100 个 Unicode rune 返回 `error.title_too_long`，重复检查与 request mutation 都使用规范化标题。`nil` update 不应查询 store。

提交说明必须包含 fixed commit、API chain、training patch、test boundary 和 rollback。
