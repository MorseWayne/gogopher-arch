# 追踪 Miniflux 调用链

基线固定为 Miniflux v2.3.2 commit `51f2e0d8199ea8fa305081f6e175bba64b0ef94b`。补全 `trace/trace.go`，按执行顺序返回 category creation API 与后台 feed refresh 的真实文件、符号和职责。不要写 commit 中不存在的中间层。
