# Miniflux Retry-After training patch

基线：Miniflux v2.3.2，commit `51f2e0d8199ea8fa305081f6e175bba64b0ef94b`，Go 1.26.0。

编辑 fetcher 与 handler 两个文件。`ParseRetryDelay(now, maximum)` 接受正整数秒或 RFC1123 HTTP-date；无效值、过去时间、非正秒数和非正 maximum 返回零，合法结果上限为 maximum。`RateLimitDelay` 只在 429 上调用 parser。为两层补齐至少六个命名测试 case。

提交说明必须包含 fixed commit、background chain、training patch、test boundary 和 rollback。
