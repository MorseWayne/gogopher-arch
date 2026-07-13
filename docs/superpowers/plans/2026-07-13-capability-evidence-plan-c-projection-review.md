# 能力证据纵向切片 Plan C：投影与复习调度

> 状态：Blocked until Plan B is accepted
> 日期：2026-07-13
> 上游：Plan B 已验收的 Evidence、EvaluationBatch、RuleResult 和 outbox 契约

## 目标

把追加写入的 Evidence 转为可重建的 CapabilitySnapshot，并用透明 review policy 创建、领取、完成和替换 ReviewItem，使“已验证、到期、复验稳定、失败生锈、补救再练”成为可查询的真实闭环。

本计划不开发完整能力图谱 UI；只提供稳定 API 和服务端状态语义，供 Plan D 集成。

## 数据库增量

新增 migration 创建：

- `capability_snapshots`
- `review_items`
- `attempt_review_items`

同时为 Plan B 的 `learning_outbox` 增加本计划 consumer 所需的 event type/claim/lease 索引。所有 migration 只新增，不修改已应用文件。

## 实施任务

### C1 — 固化投影规则模型

- [ ] 为 acquisition、independence、transfer、retention base state 写类型和合法转换表。
- [ ] 从 CapabilityDefinition 读取 `required_evidence` 与 `review_policy`，禁止为所有节点硬编码同一规则。
- [ ] 定义稳定 independence 排序：guided < ai_assisted < hinted < referenced < independent。
- [ ] 固化 `as_of` 参数；所有时间相关测试使用注入 clock。
- [ ] 为 guided/practice/assessment/review 与不同 evidence combination 写表格驱动测试。

完成条件：给定同一 Evidence/ReviewItem 集合和 `as_of`，总能得到字节等价的 projection result。

### C2 — CapabilityProjector 全量重建

- [ ] 按 `learner + capability id + version` 查询全部有效 Evidence 和 active ReviewItem。
- [ ] 从事实全量计算 Snapshot，不依赖旧 Snapshot 增量值。
- [ ] 分别计算 acquisition、历史最高 independence、transfer、retention base、last evidence 和 next review。
- [ ] `retention_state=due` 只在 API 读取时由 `as_of` 派生，不用定时任务改写。
- [ ] 只由关联 review 的失败 Evidence 把 base state 改为 rusty。
- [ ] 以 projection version 幂等 upsert，并在状态变化时写 scheduler outbox，不直接调用 scheduler。

完成条件：删除 Snapshot 后可从历史事实完全重建；重放不会产生额外 Evidence 或 ReviewItem。

### C3 — Outbox worker

- [ ] 为 projection 和 review scheduler 定义 versioned outbox payload。
- [ ] 使用数据库 claim/lease 领取事件；处理成功后记录 consumer/version。
- [ ] worker 崩溃后可重领，重复处理必须安全。
- [ ] 失败使用有上限 backoff，并记录 retry metric 和最后错误摘要。
- [ ] 增加“EvaluationBatch 已提交但投影进程崩溃”的恢复测试。

完成条件：事实事务和派生处理解耦，任何进程边界失败都不需要人工补写 Snapshot。

### C4 — 首次 review 调度

- [ ] assessment 首次独立通过 required evidence 后，为每个相关 Capability 生成 3 天后 variant ReviewItem。
- [ ] 相同 assessment 中多个 Capability 共享一个 `review_group_key` 和 review Activity，但保留逐 Capability ReviewItem。
- [ ] 使用 learner/capability version/source evidence/policy version 唯一约束去重。
- [ ] 已存在未完成同 policy ReviewItem 时执行明确 replace/no-op 规则，不能静默重复。
- [ ] `next_review_at` 取最早 active ReviewItem due time。

完成条件：一次多能力 assessment 产生四个可审计 ReviewItem，并能作为一组领取。

### C5 — ReviewItem 领取和 Attempt 绑定

- [ ] 实现 `POST /review-items/{id}/attempts` 的数据库锁和 owner 校验。
- [ ] 领取同组仍未完成项，创建一个固定 review Activity/Task version 的 Attempt。
- [ ] 创建 `attempt_review_items` links；每个 ReviewItem 最多关联一个 Attempt。
- [ ] 相同 Learner 重试领取返回现有 Attempt；其他 Learner 一律 `404`。
- [ ] claimed item 不因刷新或重复 `/next` 被另一个 Attempt 再领取。

完成条件：并发领取只有一个 Attempt，且所有被领取 Capability links 可重放。

### C6 — Review 成功、失败和未评估流转

- [ ] variant review 独立通过：完成当前 item，transfer=variant、acquisition=stable、retention=fresh，并安排 14 天后维护。
- [ ] review failed：完成当前 item，保留历史 verified/stable 事实，base retention=rusty，并在 1 天内创建针对性 practice。
- [ ] 某 Capability 的规则 not_evaluated：不把它视为 failed，替换为 `review_incomplete` item，保留原 due 和 retry semantics。
- [ ] 同一 review 对多 Capability 出现 passed/failed/not_evaluated 时逐节点处理，不以整个 Attempt 粗粒度完成。
- [ ] remediation practice 完成后按 policy 决定重新安排 variant review，不直接伪造 stable。

完成条件：混合结果不会让未运行节点变 rusty，也不会因其他节点通过而错误完成。

### C7 — `/capabilities` 和 `/learning/next`

- [ ] GET Capability 合并当前 release 定义、当前版本 Snapshot、最近 Evidence 和派生 retention。
- [ ] GET `/learning/next?as_of=` 在测试环境支持显式 as_of；生产忽略客户端时间并使用 server clock。
- [ ] 优先返回已到期/claimed review，其次返回 acquisition path 中可开始的下一 Activity。
- [ ] hard prerequisite 未满足时不推荐被阻塞 Activity；recommended prerequisite 只作为说明。
- [ ] 返回真实 source metadata，避免把静态演示状态标为服务端进度。
- [ ] 对无 Snapshot、新版本未继承旧 Evidence、无可用活动写 API 测试。

完成条件：同一 `as_of` 下 next query 稳定；新 Capability version 不自动继承旧 Snapshot。

### C8 — 重建 command 与可观测性

- [ ] 新增 repo-local admin command，用于按 learner/Capability 或全量重建 Snapshot。
- [ ] command 先 dry-run 输出差异，再显式 apply；不修改 Evidence。
- [ ] 增加 projection lag、outbox retry、review created/claimed/completed/replaced 和 due count metrics。
- [ ] 日志使用内部 ID，不包含用户代码或 session token。
- [ ] 为重复重建、并发 scheduler 和旧 release review 写集成测试。

## 验证

```bash
go test ./...
go test ./src/services/gateway/internal/learning/... -count=10
go run ./src/cmd/learning-rebuild --dry-run
git diff --check
```

使用可控 clock 的集成场景必须覆盖：assessment 独立通过 → 3 天到期 → variant review 通过 → 14 天维护；以及 review failed → rusty → 1 天 remediation。

## 工程审查重点

- Snapshot 是否始终可重建，而不是新的事实来源。
- due 是否按读时 `as_of` 派生，避免后台时钟更新竞争。
- 多能力 ReviewItem 是否分项存储、分组领取、逐项完成。
- not_evaluated 是否与 failed 严格区分。
- outbox claim/lease 和唯一约束是否共同保证幂等。

## 停止与验收

若任何状态只能通过直接手改 Snapshot 修复，或重放 outbox 会重复 ReviewItem，必须停在 Plan C。

验收演示：独立 assessment 生成 verified Snapshot 和四个 review item；推进 clock 后 `/learning/next` 返回 variant review；提交混合结果后分别展示 stable、rusty 和 review_incomplete，并可删除/重建 Snapshot 得到相同状态。
