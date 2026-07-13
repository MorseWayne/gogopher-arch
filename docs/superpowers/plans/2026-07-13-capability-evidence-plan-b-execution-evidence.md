# 能力证据纵向切片 Plan B：多文件执行、提交与证据

> 状态：In progress — B1/B2/B3 accepted
> 日期：2026-07-13
> 上游：Plan A 已验收的 definition、session、Attempt 和 workspace 契约

## 目标

把 Plan A 的冻结定义与 workspace 接入可信执行链：Gateway 只根据 TaskDefinition 生成固定 ExecutionSpec，Sandbox 执行多文件 Go module，最终 Submission 以幂等方式冻结、执行和评估，并追加可重放的 Evidence。

本计划完成后，API 可以产生 Evidence，但还不计算 CapabilitySnapshot、ReviewItem 或 `/learning/next`；这些属于 Plan C。

## 固定边界

- 删除旧 `/api/v1/execute`、单文件 `SandboxRequest` 和 `go run main.go` runner；只保留 versioned 多文件 ExecutionSpec。
- 浏览器只提交 action enum、workspace revision/hash 和 idempotency key，永远不能提交命令、环境变量、挂载路径或 held-out tests。
- worker 第一版运行在 Gateway 进程中，使用 PostgreSQL queue + lease；不新增 Redis queue 或独立 worker service。
- Sandbox 仍是本地可信开发能力，不宣称生产隔离或真正强制断网。
- `user_failed` 可以进入确定性评估；`infra_failed` 不能产生 EvaluationBatch/Evidence。

## 数据库增量

新增下一份单调 migration，创建：

- `attempt_submissions`
- `attempt_executions`
- `assistance_events`
- `artifacts`
- `evaluation_batches`
- `evidence_records`
- `learning_outbox`

必须落实总规格 14.1 的唯一约束：普通 execution request key、每 Attempt 唯一 Submission、submit retry sequence、EvaluationBatch identity 和 Evidence identity。

## 实施任务

### B1 — 固化内部 ExecutionSpec 协议

- [x] 在共享 Go package 定义 versioned `ExecutionSpec`、file asset、action policy、stage result、RuleResult 和 response。
- [x] 为 build/test/vet/submit 写 JSON round-trip 与协议不变量测试；不测试旧 payload 兼容。
- [x] 只允许 `build`、`test`、`vet`、`submit` 四种动作，命令由 Sandbox 内部映射。
- [x] 校验 clean 相对路径、可编辑/只读来源、文件数量/总大小、timeout、输出上限和 asset hash。
- [x] 未实现的 `network=none` 必须在响应中标记为 policy-only，日志和 UI 不能宣称已隔离。

完成条件：无法用协议表达任意 shell、env、absolute path 或宿主 mount。

### B2 — Sandbox 多文件 runner

- [x] 先用 fixture 写 build/test/vet、path traversal、symlink、超限和输出截断失败测试。
- [x] 每次执行从 ExecutionSpec 在新临时目录重建 workspace，不接受已有目录。
- [x] 对 readonly/held-out asset 在写入前复核 SHA-256；拒绝未登记文件和路径覆盖。
- [x] 把动作映射为直接 argv 调用，不经过 shell 插值。
- [x] 捕获 stdout/stderr、exit code、duration、timeout 和 truncation，区分用户失败与 runner 基础设施失败。
- [x] 用户进程文件变更不回写 Attempt；执行结束清理临时目录。

完成条件：普通 module 可稳定执行；所有恶意路径 fixture 在启动 `go` 前被拒绝。

### B3 — submit 的 visible/held-out 阶段

- [x] 为 submit 编写分阶段执行 fixture 和结构化 `go test -json` parser 测试。
- [x] 先运行 visible tests；只有定义允许的后续阶段才注入 held-out tests。
- [x] 对每个 assessment package 使用 `go test -c`，再把二进制和 runtime fixture 放入干净目录。
- [x] 使用 `go tool test2json` 收集 package/test 事件，生成稳定 stage result。
- [x] 返回 TaskDefinition 允许公开的 held-out 摘要，不返回源码、绝对路径或具体隐藏输入。
- [x] 增加用户代码枚举目录时读不到 held-out 源码的回归测试，并保留“不抗恶意逆向”的文档声明。

完成条件：评估依赖结构化结果，不通过 stdout 字符串猜测通过状态。

### B4 — Execution queue、lease 与失败分类

- [x] 实现普通 execution 创建的 `attempt_id + request_key` 幂等和 fingerprint 冲突检测。
- [x] 在事务中创建 queued Execution；HTTP 请求不持有执行期间的数据库事务。
- [x] worker 以 lease 领取 queued item，持续续租，只有有效 lease owner 可以写终态。
- [x] 实现 `queued → running → succeeded|user_failed|infra_failed`，终态不可重开。
- [x] lease 过期恢复为可领取语义，不伪造终态；限制最大基础设施重试次数。
- [x] 启动时校验 `task timeout < sandbox response bound < RPC deadline < worker lease` 并预留落库余量。
- [x] 为 action timeout、RPC deadline、Sandbox 不可达、worker 中断和输出截断写分类测试。

完成条件：用户死循环稳定得到 `user_failed` timeout；Sandbox/RPC/worker 故障稳定得到 `infra_failed`。

### B5 — AssistanceEvent 与 independence cutoff

- [ ] 创建 `assistance_events`，实现 `attempt_id + event_key` 幂等和 Attempt 内单调 `event_seq`。
- [ ] 提示 reveal 必须先在事务中记录事件，提交成功后才返回提示内容。
- [ ] `hint_revealed`、`reference_opened`、`solution_viewed`、`ai_declared` 使用固定 enum。
- [ ] 只允许 active Attempt 写事件；Submission 冻结后返回 `409 attempt_already_submitted`。
- [ ] 为 assistance 与 submit 并发写测试，验证同一 Attempt 行锁决定唯一 cutoff。
- [ ] 实现基于 `event_seq <= assistance_cutoff_seq` 的 independence 计算。

完成条件：无论提示与 submit 谁先获得锁，都可以从持久化事实重复得到同一 independence。

### B6 — SubmissionWorkflow

- [ ] 为相同 key 重试、同 key 内容变化、不同 key 并发和非 owner 访问写测试。
- [ ] 在短事务中锁定 Attempt，校验 owner/active/revision/hash，冻结 workspace，创建唯一 Submission 和 submit Execution。
- [ ] 保存 `rule_set_hash`、workspace hash/revision 和 assistance cutoff；事务内不调用 Sandbox。
- [ ] 重复相同 submission key 返回同一 Submission；不同 fingerprint 返回 `409 idempotency_conflict`。
- [ ] 第一个不同 key 冻结成功后，其他 key 返回 `409 attempt_already_submitted` 和已有 Submission ID。
- [ ] `infra_failed` 把 Attempt 置为 `submit_infra_failed`；显式 retry 只复用冻结 workspace 并创建新 retry sequence。

完成条件：任何并发/响应丢失场景都只存在一个冻结 Submission，且每次 retry 都有独立不可变 Execution。

### B7 — RuleResult 生成

- [ ] 按 TaskDefinition assessment rules 把 stage result 转为 `passed|failed|not_evaluated` RuleResult。
- [ ] 实现 module build、error chain、invalid input、stable output 和测试相关规则的 fixture。
- [ ] `learner-tests-present` 使用 starter diff + Go AST：修改 `_test.go`、表格结构、至少三个命名 case，并要求普通测试通过。
- [ ] RuleResult 保存匹配 package/test/analyzer、公开摘要和原始 Execution 引用。
- [ ] 前置阶段失败时只把实际运行规则标为 failed，后续规则标为 not_evaluated。

完成条件：同一 Submission 可以对不同 Capability 产生混合结果，且每条结果有稳定 `rule_id`。

### B8 — EvidenceEvaluator 和原子批次

- [ ] 根据冻结 release/rule set、终态 Execution、cutoff AssistanceEvent 和 Activity mode 计算 Evidence。
- [ ] `passed/failed` 每规则追加 Evidence；`not_evaluated` 只保留在 EvaluationBatch。
- [ ] 在同一事务写 EvaluationBatch、全部 Evidence、Submission/Attempt completed 和 projection outbox。
- [ ] 使用唯一约束保证相同 Submission/rule set 重放不重复追加 Evidence。
- [ ] 保存最终 workspace、关键 diff、解释和测试报告 Artifact，并执行 JSONB 大小上限。
- [ ] infra failure 路径断言没有 EvaluationBatch、Evidence 或 passed 状态。

完成条件：重放 evaluator 或重复 HTTP submit 不改变 Evidence 数量；完整批次要么全部提交，要么全部回滚。

### B9 — Learning API 与可观测性

- [ ] 实现 execute、submit、submission retry、hints reveal 和 assistance event endpoints。
- [ ] GET Attempt 返回公开 Execution、RuleResult、Evidence 摘要，但不泄露内部命令/held-out 内容。
- [ ] 统一映射 `401/404/409/422` 和结构化 domain error code。
- [ ] 增加 Attempt/Execution/Submission/Evidence 计数、耗时、timeout 和 truncation metrics/log fields。
- [ ] 日志禁止完整用户代码、held-out tests、session token 和 cookie。

## 验证

```bash
go test ./...
go test ./internal/sandbox/... -count=10
go test ./internal/learning/execution/... -count=10
docker compose config
git diff --check
```

集成测试必须使用真实 PostgreSQL 和 Sandbox process，覆盖普通动作、独立 submit、提示降级、用户失败、基础设施失败、显式 retry 和重复提交。

## 工程审查重点

- ExecutionSpec 是否完全由受信任定义生成。
- timeout/RPC/lease 是否能被测试证明分类无歧义。
- EvaluationBatch 事务是否包含 Evidence、状态推进和 outbox，而非由 handler 拼接。
- held-out 源码是否没有进入公开 DTO、日志和 Web 构建。
- 旧单文件 Sandbox API 和 runner 是否已完全删除，避免形成第二套执行事实。

## 停止与验收

若 Plan A 契约需要语义修改、失败分类无法稳定、Evidence 幂等依赖人工清理，必须停下重新审查，不启动 Plan C。

验收演示：创建 Attempt → 保存 workspace → Test → 记录提示 → Submit → 查看结构化 RuleResult 和 hinted Evidence；随后模拟 Sandbox 故障，对冻结 Submission retry，最终仍只有一个 EvaluationBatch。
