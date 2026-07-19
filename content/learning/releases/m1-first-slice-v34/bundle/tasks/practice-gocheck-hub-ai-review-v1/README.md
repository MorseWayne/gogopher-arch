# AI 协作交付审查

修改 `deliveryreview/review.go` 和对应测试。`Review` 返回全部问题代码；安全方案返回空切片。问题代码必须稳定、去重，覆盖认证顺序、数据真相、缓存降级、worker 并发与重试、forward-only migration、完整 gate、非 root runtime 和前向回滚。提交解释必须说明 AI 产出由谁验证。
