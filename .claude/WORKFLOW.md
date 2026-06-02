# Workflow Ledger

Claude Code 开发工作的轻量级可恢复台账。

## Active

### WF-2026-06-02-001 — Go 课程质量样板设计
Status: Active
Level: 3
Started: 2026-06-02
Updated: 2026-06-02
Current phase: P5 — 验证 ch07 内容、练习和构建。

Intent: 建立 Go 课程质量升级样板：完整升级 ch07 Interfaces 为样板章，并为 ch11 Testing 形成审计与改造蓝图。

Plan:
- [done] D1 — 确认样板路线、课程质量标准、ch07 设计和 ch11 蓝图边界。
- [done] D2 — 写入并评审设计规格文档。
- [done] D3 — 制定实施计划。
- [done] P1 — 审计当前 ch07/ch11 与来源映射。
- [done] P2 — 改造 ch07 MDX 正文为接口样板章。
- [done] P3 — 更新 ch07 metadata 与 warmup/core/challenge 练习。
- [done] P4 — 落实 ch11 审计与改造蓝图。
- [doing] P5 — 验证 ch07 内容、练习和构建。
- [todo] P6 — 更新持久状态并准备关闭或进入下一轮。

Current todo:
- [ ] P5 — 做最终验证：前端构建、空白检查、ch07 练习抽查和内容质量人工检查。

Changes:
- 用户选择方案 D：先做样板和标准；样板章选择 ch07 Interfaces + ch11 Testing；产物边界选择完整改造 ch07、ch11 只做审计和蓝图。
- 用户确认目标学习者为“后端新手到实习”。
- 用户以“实施计划”确认设计规格可进入实现计划；设计规格状态已更新为 Approved for implementation planning。
- 用户确认实施计划，开始执行 P1 审计。
- P1 审计已记录在实施计划：ch07 当前语义覆盖强但需从构建通知切换到订单通知主线；ch11 当前 NormalizeName 主线保持不重写，只记录未来迁移蓝图。

- P2 已完成 ch07 MDX 正文改造：从构建通知切换为订单通知主线，保留并重组接口方法集、隐式实现、接口值、nil 陷阱、any、类型断言/分支、error、小接口和工程 checklist；`npm run build --prefix web` 与 `git diff --check` 通过。
- P3 已完成 ch07 metadata 和练习同步：summary/goals/notes/practices/pitfalls/checklist/reviewQuestions/exercises 全部切到订单通知主线；`npm run build --prefix web`、`git diff --check` 通过，warmup starter 输出匹配，core/challenge starter 可运行且参考解法输出匹配。
- P4 已轻量落实 ch11 蓝图：保留当前 NormalizeName 主线，新增与 ch07 Interfaces 的衔接说明、测试替身桥接和下一轮订单通知测试样板蓝图；未完整重写 ch11，`npm run build --prefix web` 通过。

Prerequisites:
- None.

Resume next: 执行 P5：运行最终 `npm run build --prefix web`、`git diff --check`、ch07 warmup/core/challenge starter/参考解法抽查，并人工检查 ch07 rubric 与 ch11 蓝图范围。

## Backlog / Future

- [ ] 若课程数据继续膨胀，后续可考虑将 13 章拆分为独立章节文件；当前计划明确首版保留单文件。

## Completed

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

