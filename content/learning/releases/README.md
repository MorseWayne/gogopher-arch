# Learning releases

每个 `<release-id>/` 都是由 `cmd/learning-content release` 生成的 immutable runtime bundle。Gateway 通过 `content/learning/current-release.json` 为新 Attempt 选择 release；已创建 Attempt 始终保留自己的 `release_id`。

## 规则

1. `release-id` 必须唯一，禁止覆盖或编辑已有 release 目录。
2. 更新 pointer 前必须运行 `learning-content verify`。
3. 部署与回滚必须保留所有被历史 Attempt 引用的 release。
4. 删除旧 release 会让冻结 Attempt 无法恢复或重试。
5. 此目录的 `go.mod` 仅隔离归档 Go test asset，不属于任何 release，也不进入 bundle hash。

## Verify

```bash
npm run build --prefix web
go run ./cmd/learning-content verify \
  --release-dir content/learning/releases/m1-first-slice-v18 \
  --web-dist web/dist
```

## Content rollback

先关闭 `LEARNING_SLICE_ENABLED`，再把 `current-release.json` 指向一个仍在仓库中且已验证的旧 release。Pointer rollback 只影响新 Attempt，不得修改历史 bundle、database record 或 Evidence。
