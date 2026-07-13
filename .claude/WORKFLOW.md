# Workflow Ledger

Claude Code 开发工作的轻量级可恢复台账。

## Active

### WF-2026-07-12-001 — 能力节点与证据纵向切片设计
Status: Active
Level: 3
Started: 2026-07-12
Updated: 2026-07-13
Priority: Product redesign specification; independent from the active DeepTutor spike.
Current phase: Plan C C3 已完成，开始首次 review 调度。

Intent: 为 M1-01、M1-03、M1-07、M1-09 建立第一条端到端能力训练切片，验证版本化能力定义、服务端 Attempt/Evidence、派生能力状态、多文件 Go 任务和维护调度闭环。

Plan:
- [done] D1 — 明确产品定位、能力图谱、能力维护和第一版范围。
- [done] D2 — 核对当前课程、任务、Dashboard、Gateway 和 Sandbox 数据流。
- [done] D3 — 比较数据方案并确认“版本化定义 + 服务端证据”方案。
- [done] D4 — 写入设计规格并完成五轮独立规格评审。
- [done] D5 — 用户批准继续规格工作并进入分阶段实施计划。
- [done] D6 — 按定义/Attempt、执行/证据、投影/复习、前端/E2E 拆分四份计划。
- [done] P0 — 审查 Plan A 的工程边界，确认 migration、release command、Gateway package 和 Compose 边界。
- [done] A1 — 建立 migration runner、schema history、Plan A 初始 migration 和 Compose 启动门。
- [done] A2 — 建立 JSON Schema、四个 Capability、六个 Activity/Task 及完整资产。
- [done] A3 — 实现 canonical hash、release manifest、bundle 构建和 verify command。
- [done] A4 — 实现只读 DefinitionRegistry、启动校验和数据库不可变版本登记。
- [done] A5 — 实现匿名 Learner session、token hash 和 cookie 所有权边界。
- [done] A6 — 实现 Attempt 创建、所有权读取、workspace revision/hash 与并发保存。
- [done] A7 — 实现 Gateway wiring、Learning feature gate、路由和旧 API 删除。

Current todo:
- [ ] Plan C C4 — assessment 独立通过后幂等创建首次 variant ReviewItem。

Changes:
- 用户确认 M1 14 节点、M2 16 节点、`gocheck`/`gocheck-hub` 内容原型和 Miniflux v2.3.2 训练项目方向。
- 用户确认能力与活动使用仓库内版本化定义，用户尝试、证据、能力状态和复习队列使用服务端持久化。
- 第一条实现规格只覆盖 M1-01、M1-03、M1-07、M1-09 和一个多文件完整程序任务。
- 正式规格已通过独立审查，关闭发布包回放、幂等提交、多能力 ReviewItem、复习补救、帮助事件 cutoff、timeout 分类和本地网络边界等问题。
- 已新增四份独立实施计划：Plan A 定义/会话/Attempt，Plan B 多文件执行/提交/Evidence，Plan C 投影/复习，Plan D 前端/E2E。
- 四份计划按顺序设置验收门：后续计划只能依赖已验收契约，不得反向扩大前一计划范围。
- Plan A 已按 breaking-change 边界重新审查：使用单 Go module、标准根目录布局和手工 DI，Plan A 只负责 definition/session/Attempt 表与 API。
- 用户明确授权 breaking change：不兼容旧 API、旧数据库、旧页面、旧路由和旧视觉；前后端按新产品闭环重新设计。
- Plan A 改用标准根目录 `cmd/`、`internal/`、`api/` 布局；旧 `src/services` 和单文件执行协议不再是兼容边界。
- A1 已完成：使用 `pgx/v5`、Go 1.25、advisory lock、逐 migration transaction 和 SHA-256 drift 检测；真实 PostgreSQL 与 migrate Docker image 验证通过。
- A2 已完成 Draft 2020-12 Capability/Activity/Task schema、Go validator、四个 Capability 和六个 Activity；仓库定义全目录校验通过。
- 六个 TaskDefinition 已声明全部真实资产和 SHA-256；assessment 与 review 使用不同 module、字段和错误场景，starter/测试均通过编译基线检查。
- A3 已完成 RFC 8785 canonical JSON、task/rule-set/full bundle 分层 hash、ReleaseManifest schema，以及确定性 `validate`/`release`/`verify` command。
- `m1-first-slice-v1` 已归档 4 个 Capability、6 个 Activity、6 个 Task 和 32 个声明资产；完整 bundle hash 为 `ed9ae605b819bfc3d75cf3142a00021c3ac3bfd32c50070cbb2fc4f6e3d985cf`。
- release verifier 会拒绝路径逃逸、symlink、引用缺失、hard prerequisite 环、规则错配、遗漏或多余文件及任意 hash 漂移；前端断言会扫描 held-out 文件名和内容指纹。
- A4 已实现 `current-release.json`、多 release 只读 Registry、历史 release 启动门、公开 Activity/Task DTO 和不含 held-out 的 public workspace。
- `ReleaseStore` 使用 serializable transaction 与 advisory lock 登记 release/definition history；真实 PostgreSQL 已验证幂等登记、current 指针、Attempt 引用查询和 version hash 冲突回滚。
- A5 已实现 256-bit CSPRNG session token、SHA-256-only persistence、匿名 Learner transaction、过期/伪造替换语义和 owner context middleware。
- session cookie 固定 `HttpOnly`、`SameSite=Lax` 与 `/api/v1/learning` path；真实 PostgreSQL 已验证 hash-only persistence、复用、过期拒绝和新 Learner 替换。
- A6 已实现冻结 release/activity/task hash 的 Attempt、完整 public workspace、readonly/大小限制与 length-prefixed SHA-256 workspace hash。
- workspace 保存使用 `SELECT ... FOR UPDATE` 和 revision CAS；真实 PostgreSQL 并发测试确认两个同 revision 保存恰好一个成功，另一个返回当前 revision/hash。
- A7 已切换到 `cmd/gateway` 与显式 App wiring；Learning feature gate 仅允许 local，关闭时返回 `503 learning_disabled`，旧 `/api/v1/execute` 返回 `404`。
- 隔离 Compose smoke test 已通过 Web proxy 完成 session → Activity → Attempt → workspace revision 1；Gateway/Sandbox 无宿主发布端口，Web 仅绑定 `127.0.0.1`。
- Plan A 已以 `ee32072` 完成提交和推送；A1–A7 的全仓测试、Web build、release verify 与真实 Compose smoke 均通过。
- B1 已固化无 command/env/mount 的 versioned ExecutionSpec；SpecBuilder 只允许冻结 TaskDefinition 生成 action policy 与 release asset。
- B2/B3 已用 `cmd/sandbox` 替换旧单文件 service，支持多文件 build/test/vet 与 visible → held-out submit，响应明确记录 `network=none` 仅为 `policy_only`。
- 非 root Sandbox 容器已完成真实 HTTP build smoke；旧 `/execute` 返回 `404`，临时目录与 held-out source 在执行后清理。
- B4 已新增 `000002_learning_execution_evidence` migration，并为 normal Execution 固定 request fingerprint、queued snapshot 与 `attempt_id + request_key` 幂等边界。
- PostgreSQL lease 使用 `FOR UPDATE SKIP LOCKED`、heartbeat 和 owner-only terminal update；过期 job 可重领，达到 claim 上限后落为 `infra_failed`。
- 真实 PostgreSQL → worker → HTTP Sandbox process → terminal write 集成测试通过；完整 Compose 已应用两份 migration，Gateway 启动时校验 task/RPC/lease 时限顺序。
- B5 已实现 `attempt_id + event_key` 幂等 AssistanceEvent、Attempt 行锁内单调 `event_seq`、active-only 写入和先记录后返回提示内容。
- 真实 PostgreSQL 竞争测试验证 AssistanceEvent 与 Submission cutoff 共用 Attempt 行锁；`event_seq <= cutoff` 可重复计算 guided/hinted/referenced/ai_assisted/independent。
- B6 已实现唯一 Submission 冻结事务，原子保存 workspace revision/hash、rule_set_hash、assistance cutoff 并排入 submit Execution。
- 相同 submission key 可安全回放，不同 fingerprint/key 返回带已有 Submission ID 的 domain conflict；真实 PostgreSQL 验证不同 key 竞争只有一个赢家。
- submit Execution 的 RPC/lease `infra_failed` 会原子推进 Submission/Attempt；显式 retry 使用冻结 workspace 并按 sequence 追加幂等 Execution。
- B7 已把 submit 固定为 build → vet → visible → held-out pipeline，并在共享 output/timeout budget 下保留阶段短路事实。
- RuleResult generator 使用冻结 TaskDefinition 匹配 package/test，使用 Go AST 验证 deferred Close 与修改后的三组命名 table case，并保留原始 Execution ID。
- 真实 assessment solution 已走完整 Sandbox 产生 10 条 passed RuleResult；人工 terminal fixture 验证 targeted pass、stage failure 和 not_evaluated 可同时出现。
- B8 已在 terminal submit 落库事务中写入 durable evaluation request，并用可重领 worker 消费，避免依赖进程内 callback。
- EvidenceEvaluator 校验冻结 activity/rule-set/task、latest terminal Execution 和 Assistance cutoff；passed/failed 生成 Evidence，not_evaluated 只保留在 batch。
- EvaluationBatch 原子保存 workspace、diff、explanation、test_report Artifact 和关联 Evidence；每个 Artifact 的 canonical JSON 上限为 4 MiB。
- `infra_failed` 不创建 evaluation request、EvaluationBatch、Artifact 或 Evidence，Submission/Attempt 保持明确的失败状态。
- 真实 PostgreSQL 已验证 EvaluationBatch、Evidence、Submission/Attempt completed 与 projection outbox 原子提交、非法 Evidence 全量回滚和 replay 不重复。
- B9 已暴露 execute、submit、retry、hint reveal 和 assistance event endpoint，并统一映射 owner、validation、state 与 idempotency error。
- GET Attempt 通过 owner-scoped read model 恢复 Submission、Execution、RuleResult 和 Evidence；held-out output、test name 与 package 不进入公开 DTO。
- `/metrics` 暴露 bounded label 的 Attempt、Execution duration/status/failure/truncation 和 Evidence counters；worker 只记录 ID、enum、duration 与计数。
- observability regression test 使用带 secret marker 的 workspace、output、failure message 和 Evidence reason，确认 structured logs 不包含这些 payload。
- Plan B Compose acceptance 已通过真实 session → assessment Attempt → test → reference assistance → submit → Evaluation → GET Attempt 主链。
- starter workspace 的 test/submit 均稳定落为 `user_failed`；Evaluation 保存 10 条 RuleResult 和 6 条 referenced Evidence，重复 submit 不增加 Execution/Evidence。
- 第二个 Learner 读取该 Attempt 返回 `404 attempt_not_found`；metrics 与 Gateway logs 仅包含固定维度和内部 ID，Compose 服务均保持 healthy。
- C1 已定义 acquisition、independence、transfer、retention base/due 状态与合法转换，并固定自治级别排序。
- Registry 按 Capability version 返回各自 required_evidence 和 review_policy；projection 使用显式 as_of，输入 Evidence 顺序不改变 JSON result。
- C2 migration 已新增 capability_snapshots、review_items、attempt_review_items 和 outbox consumer metadata/index。
- CapabilityProjector 从完整 Evidence 与 active/completed ReviewItem 重建 Snapshot；unlinked review failure 不会变 rusty，active due 只在显式 as_of 下派生。
- Snapshot 使用 projection_version 幂等 upsert；删除后重建得到相同状态，scheduler outbox 依赖 projection payload hash 去重。
- C3 已为 projection 与 review scheduler event 增加显式 event version，并把 projection worker 接入 Gateway 生命周期。
- projection request 使用 `FOR UPDATE SKIP LOCKED` claim、owner lease 和 expired recovery；成功记录 `capability_projector@1`。
- 失败按有上限指数 backoff 重试并记录 bounded `last_error`、retry metric；达到上限进入带 `failed_at` 的终态。
- 真实 PostgreSQL 已验证 EvaluationBatch 提交后 projection owner 崩溃、replacement owner 重领、幂等 Snapshot 重建和 poison event exhaustion。

Prerequisites:
- 当前 Sandbox 仅支持单个 `main.go` 且缺少生产级隔离；规格必须限制为本地可信环境，并明确公开运行前的安全门槛。
- 现有 DeepTutor Spike 和用户对其 ledger/spec 的修改必须保持不变。

Resume next: 实施 C4 review scheduler consumer、首次 3 天 variant ReviewItem 和多 Capability grouping/去重。

### WF-2026-06-02-002 — DeepTutor 离线课程内容工作流 Spike
Status: Active
Level: 3
Started: 2026-06-02
Updated: 2026-06-03
Priority: Current implementation planning; WF-2026-06-02-001 remains paused.
Current phase: P3 — 安装/运行 DeepTutor 并生成章节正文草稿。

Intent: 验证 DeepTutor 能否作为离线课程内容研究工具，利用开放网页检索与站内课程约束生成明显优于原章节的教程级 Go 课程正文草稿。

Plan:
- [done] D1-D5 — 需求澄清、方案选择、设计规格、规格评审和用户审阅已完成。
- [done] P1 — 审查 13 章并选择目标章节。
- [done] P2 — 准备 DeepTutor 输入包与课程风格契约。
- [doing] P3 — 安装/运行 DeepTutor 并生成章节正文草稿。
- [todo] P4 — 审计草稿并替换目标 MDX。
- [todo] P5 — 验证内容质量、构建和回滚决策。
- [todo] P6 — 记录结论、更新 ledger 并提交。

Current todo:
- [ ] P3 — 确认 DeepTutor 安装/运行方式，将 `input-package.md` 交给 DeepTutor 生成 ch10 正文草稿和来源记录。

Changes:
- 用户选择研究型集成 Spike，而不是立即产品化或直接深度集成。
- 用户明确优先方向为离线内容工作流；AI 导师作为后续方向。
- 用户选择开放网页检索、完整审计包、内容质量优先，以及“方案 A 起步，方案 B 作为加分验证”。
- 设计规格已写入 `docs/superpowers/specs/2026-06-03-deeptutor-course-content-spike-design.md` 并通过规格评审；规格提交为 `78c31c7`。
- 实施计划已写入 `docs/superpowers/plans/2026-06-03-deeptutor-course-content-spike-implementation.md`；规格状态更新为 Approved for implementation planning。
- P1 已完成 13 章评分，选择 `ch10-packages-tools` 作为 DeepTutor Spike 目标章节；结果写入 `docs/superpowers/spikes/deeptutor-course-content/chapter-selection.md`。
- P2 已完成 DeepTutor 输入包，包含 ch10 原文、metadata/练习摘要、课程风格契约、开放检索要求、输出格式和最终提示模板；文件为 `docs/superpowers/spikes/deeptutor-course-content/input-package.md`。

Prerequisites:
- DeepTutor 安装/运行可能需要外部依赖、账号或模型配置；若成本较重，按 Stop condition 暂停确认。
- 课程正文仍需遵守站内正文优先、外部资料只作来源层和禁止拼贴的项目原则。

Resume next: 执行 P3：确认 DeepTutor 安装/运行方式，并尝试用 `input-package.md` 生成 ch10 正文草稿和来源记录。

### WF-2026-06-02-001 — Go 课程质量样板设计
Completed: 2026-06-02
Level: 3

Close summary:
- Outcome: 已完成 Go 课程质量升级样板：ch07 Interfaces 已完整升级为订单通知接口样板章，ch07 metadata/exercises 已同步，ch11 Testing 已补充与 ch07 的衔接说明和后续订单通知测试样板蓝图。
- Validation: `npm run build --prefix web`、`git diff --check`、ch07 三个 starter 可运行、三份参考解法输出匹配；ch07 rubric 人工检查通过，ch11 保持蓝图范围。
- Gaps: ch11 完整订单通知测试样板、更多章节推广已进入 Backlog / Future；本轮无阻塞 gap。

Archived execution:
- Intent: 建立 Go 课程质量升级样板：完整升级 ch07 Interfaces，并为 ch11 Testing 形成审计与改造蓝图。
- Plan:
  - [done] D1-D3 — 确认样板路线、写入并评审设计规格、制定实施计划。
  - [done] P1 — 审计当前 ch07/ch11 与来源映射。
  - [done] P2 — 改造 ch07 MDX 正文为接口样板章。
  - [done] P3 — 更新 ch07 metadata 与 warmup/core/challenge 练习。
  - [done] P4 — 落实 ch11 审计与改造蓝图。
  - [done] P5-P6 — 完成最终验证、记录后续任务并关闭。
- Key changes:
  - ch07 从构建通知主线切换为订单通知主线，保留并重组接口方法集、隐式实现、接口值、nil 陷阱、any、类型断言/分支、error、小接口和工程 checklist。
  - ch07 summary/goals/notes/practices/pitfalls/checklist/reviewQuestions/exercises 全部同步到订单通知、Notifier、SpyNotifier 和错误传播主线。
  - ch11 保留当前 NormalizeName 主线，仅新增与 ch07 Interfaces 的衔接说明、测试替身桥接和下一轮订单通知测试样板蓝图。
- Validation:
  - 多轮 `npm run build --prefix web` 与 `git diff --check` 通过。
  - ch07 warmup starter 输出匹配；core/challenge starter 可运行且参考解法输出匹配。
  - P5 最终 rubric 检查确认 ch07 满足 Concept、Progression、Practice、Source、Backend relevance；ch11 未超出蓝图范围。
- Deferred / gaps:
  - ch11 完整订单通知测试样板后续单独实施。
  - ch07 样板推广到其他章节后续按批次规划。

### WF-2026-05-31-001 — 全站 shadcn 视觉系统重构
Completed: 2026-06-02
Level: 3

Close summary:
- Outcome: 已完成全站 shadcn/ui 视觉系统重构、分区 App Shell、核心页面信息架构优化、light/dark/system 主题支持，以及本地开发启动脚本补充。
- Validation: 视觉/主题阶段 `npm run build --prefix web` 与 `git diff --check` 通过；P6 脚本语法、help 输出和空白检查通过；用户已手工验证 light/dark/system、首页、Dashboard、课程、章节、任务、沙盒锚点和移动端 Sidebar。
- Gaps: None.

Archived execution:
- Intent: 将前端重构为 shadcn/ui 视觉系统，并优化首页、学习区、课程阅读和任务工作台的信息架构。
- Plan:
  - [done] P1-P3 — 需求澄清、设计规格和实施计划已完成。
  - [done] P4 — 全站 shadcn 重构已完成。
  - [done] P5 — light/dark/system 主题支持已完成。
  - [done] P6 — 本地开发启动脚本与 README 使用场景补充已完成。
- Key changes:
  - 用户选择混合视觉方向、分区 App Shell、页面职责重排，并保留未实现路线的“即将开放”入口。
  - 公开区、学习区、Dashboard、课程页、章节页、任务页和练习面板迁移到统一 shadcn token 与 Go 蓝视觉系统。
  - 主题支持默认跟随系统并提供 light/dark/system 手动切换；开发脚本支持 full Docker、Docker 后端 + 本地 Vite、本地 Go 服务 + Docker 依赖等场景。
- Validation:
  - `npm run build --prefix web` 与 `git diff --check` 通过。
  - P6 脚本语法、help 输出和空白检查通过。
  - 用户手工验证 light/dark/system、首页、Dashboard、课程、章节、任务、沙盒锚点和移动端 Sidebar 通过。
- Deferred / gaps:
  - None.

### WF-2026-05-31-002 — Go 基础课程 React + MDX 内容系统
Completed: 2026-06-01
Level: 2

Close summary:
- Outcome: 已完成 Go 基础课程 React + MDX 内容系统：课程章节使用 metadata + 动态 MDX 正文，保留 React 课程页、sandbox 练习和任务衔接体验，并完成全部 13 章迁移、练习系统 v2、CodeMirror Go 编辑器、章节教程级补强和课程改造流程固化。
- Validation: 多轮 `npm run build --prefix web`、`git diff --check` 和 starter code 抽查通过；用户已确认第 4/11 章样板、P16、P17、P18 和 P19 体验验证通过。
- Gaps: None.

Archived execution:
- Intent: 将 Go 基础课程从 TypeScript 硬编码迁移到 MDX 内容系统，同时保留 React 课程页、sandbox 练习和任务衔接体验。
- Plan:
  - [done] P1-P8 — MDX 基础设施、metadata/content 拆分、13 章内容迁移和章节懒加载已完成。
  - [done] P9-P10 — 练习系统 v2 与 lazy CodeMirror Go 编辑器竖切已完成，并以 ch4/ch11 作为样板。
  - [done] P11-P15 — 第 4/11 章教程级样板、概念地图顺序调整，以及课程改造原则/Skill 固化已完成。
  - [done] P16-P19 — ch01-ch03、ch05-ch10、ch12-ch13 按教程级标准补强，并同步 metadata/exercises。
- Key changes:
  - 课程章节现在使用 TypeScript metadata + 动态 MDX 正文；练习支持多题、可编辑草稿、重置、运行反馈和 CodeMirror Go 编辑器。
  - 章节改造标准已明确为场景引入、基础概念逐步讲解、对照示例、讲解后概念回看和 warmup/core/challenge 分层练习。
  - `.claude/skills/go-course-chapter-redesign/SKILL.md` 与项目 CLAUDE.md 固化了 Go 课程设计原则。
- Validation:
  - 多轮 `npm run build --prefix web` 与 `git diff --check` 通过。
  - P16 验证：ch01/ch02/ch03 warmup starter code 抽查输出匹配，用户确认体验通过。
  - P17 验证：ch05/ch06/ch07 共 9 个 starter code 抽查可运行，warmup 输出匹配，用户确认体验通过。
  - P18 验证：ch08/ch09/ch10 共 9 个 starter code 抽查可运行，warmup 输出匹配，用户确认可推进。
  - P19 验证：ch12/ch13 共 6 个 starter code 抽查可运行，warmup 输出匹配，用户确认体验通过。
- Deferred / gaps:
  - None.

### WF-2026-05-30-002 — 本地部署 P0/P1 稳定化
Completed: 2026-05-30
Level: 2

Close summary:
- Outcome: 已完成本地部署 P0/P1 稳定化：前端改用相对 API，Vite/Nginx 代理 `/api`，Gateway 保留 Sandbox status code，Sandbox 增加 `/health`，Compose 默认保留 Postgres/Redis 并加入 healthcheck、restart、`.env.example` 和 README 说明。
- Validation: 通过 `go test ./...`、`npm run build --prefix web`、`docker compose config`、`git diff --check`、`docker compose build web gateway sandbox-engine`；实际 `docker compose up -d` 后 gateway/sandbox/postgres/redis healthy，Web `/api/v1/execute` smoke test 返回成功。
- Gaps: P2 Sandbox 安全隔离和资源限制明确暂缓；前端依赖审计中仍有 1 个 high severity vulnerability，未纳入本轮部署稳定化范围。

Archived execution:
- Intent: 让默认本地部署和混合开发模式更稳定，同时暂不进入 Sandbox P2 安全强化。
- Plan:
  - [done] P1 — 落地前端相对 API、Vite/Nginx 代理、Gateway status code 转发和 Sandbox `/health`。
  - [done] P2 — 更新 Docker Compose healthcheck、restart、`.env.example` 和 README 本地部署说明。
  - [done] P3 — 运行针对性验证并记录结果。
- Key changes:
  - 用户确认 Postgres/Redis 保留在默认 Compose 中，P2 Sandbox 安全边界暂缓。
  - 请求链路改为浏览器相对 `/api/v1`，开发由 Vite proxy，容器由 Nginx proxy，Gateway 透传 Sandbox HTTP status。
  - Compose 增加健康检查、healthy 依赖、restart 策略和 `.env` 默认值；README 同步全量 Docker 与混合开发说明。
  - 验证发现 sandbox-engine 构建期 `go get` 依赖外网会导致镜像构建不稳定，已改为运行期按需写入 `go.mod`，避免构建阶段额外下载。
- Validation:
  - Go 测试、前端构建、Compose 配置渲染、空白检查和应用镜像构建均通过。
  - Compose 全量启动后，gateway、sandbox-engine、postgres、redis 均为 healthy，Web 容器 `/api/v1/execute` smoke test 成功。
- Deferred / gaps:
  - Sandbox 执行隔离、资源限制、网络限制和公网暴露防护留到 P2。
  - 前端依赖审计漏洞治理留到单独依赖维护任务。

### WF-2026-05-30-001 — Go 基础训练营完整内置课程重制
Completed: 2026-05-30
Level: 3

Close summary:
- Outcome: 已按计划将 Go 基础训练营改为 GoGopher Arch 完整内置课程；重写课程数据模型和 13 章内容，改造课程总览页、章节详情页、Landing 文案和 README 来源边界；课程页面不再依赖外部教程正文链接。
- Validation: 运行文本检查，确认课程数据无旧 source 字段/旧概念模型，课程页面无外部原文入口文案；通过 esbuild bundle 调用 `validateGoBasicsCourse()`，结果 `[]`；本地 `go run` 验证 13 个 exercise starterCode 输出匹配；通过 `npm run build --prefix web`。
- Gaps: 未启动浏览器逐页人工抽样；未验证 sandbox 服务实际运行按钮，因为本次未启动 Gateway/Sandbox Engine。
