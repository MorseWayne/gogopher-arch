# Workflow Ledger

Claude Code 开发工作的轻量级可恢复台账。

## Active

### WF-2026-05-31-002 — Go 基础课程 React + MDX 内容系统
Status: Active
Level: 2
Started: 2026-05-31
Updated: 2026-06-01
Current phase: P17 — Batch 2 ch05-ch07 顺序章节补强。

Intent: 将 Go 基础课程内容从前端 TypeScript 硬编码数据逐步迁移到 Markdown/MDX 管理，保留现有 React 课程页、sandbox 练习和任务衔接体验，并为后续基于 gopl-zh 与开源教程资料的内容改编打基础。

Plan:
- [done] P1 — 引入 MDX/Markdown 依赖与 Vite 配置，建立章节内容加载层。
- [done] P2 — 迁移首个样章到 MDX，验证页面仍能渲染课程目录、正文、练习和校验信息。
- [done] P3 — 跑前端构建和空白检查，记录迁移边界与后续批量内容改编路径。
- [done] P4 — 批量迁移第 2-4 章为 MDX 内容，基于 gopl-zh 对应章节精简重组。
- [done] P5 — 批量迁移第 5-7 章为 MDX 内容，基于 gopl-zh 函数、方法、接口章节精简重组。
- [done] P6 — 批量迁移第 8-9 章为 MDX 内容，基于 gopl-zh 并发章节精简重组。
- [done] P7 — 优化 MDX 加载：metadata 独立为 TypeScript 数据，章节正文通过动态 import 按需加载。
- [done] P8 — 批量迁移第 10-13 章为 MDX 内容，基于 gopl-zh 包/工具、测试、反射、底层编程章节精简重组。
- [done] P9 — 练习系统 v2 竖切：多练习、可编辑代码、草稿保存、运行用户代码，并用 ch4/ch11 做样板。
- [done] P10 — CodeMirror Go 编辑器竖切：练习区替换 Textarea，支持 Go 高亮、行号、括号匹配和课程定制补全，并按需懒加载编辑器 chunk。
- [done] P11 — 第 4 章优秀资料整合样板：新增 MDX 教学组件，改造复合类型章节为“场景 → 来源 → 对照示例 → 常见坑 → 练习衔接”结构。
- [done] P12 — 第 4 章基础概念补强：根据反馈补充数组、slice、map、struct 的定义形式、初始化方式、内存/语义模型和常见误区。
- [done] P13 — 第 11 章测试样板改造：按“基础概念 → 场景 → 对照示例 → 常见坑 → 练习”的结构重写测试章节，整合 gopl-zh、Go 官方 testing/fuzz/pprof 文档和 Learn Go with Tests 思路。
- [done] P14 — 概念地图呈现顺序调整：根据反馈将第 4/11 章概念地图从文章前部移动到具体讲解后的回看/总结位置，避免开篇过于晦涩。
- [done] P15 — 固化课程设计方法：项目 `CLAUDE.md` 增加 Go 课程设计原则，新增 `.claude/skills/go-course-chapter-redesign/SKILL.md` 固化章节改造流程。
- [done] P16 — Batch 1：按教程级标准补强 ch01-ch03，并同步检查 metadata/exercises。
- [doing] P17 — Batch 2：按教程级标准补强 ch05-ch07，并同步检查 metadata/exercises。
- [todo] P18 — Batch 3：按教程级标准补强 ch08-ch10，并同步检查 metadata/exercises。
- [todo] P19 — Batch 4：按教程级标准补强 ch12-ch13，并同步检查 metadata/exercises。

Current todo:
- [ ] P17 — Batch 2：按教程级标准补强 ch05-ch07，并同步检查 metadata/exercises；正文与练习已改，构建、空白检查和 starter code 抽查通过，待用户确认体验验证结果后 checkpoint。

Changes:
- P17 ch05-ch07 正文与 metadata/exercises 已完成，自动验证通过，待用户确认体验验证结果后再切换到 P18。

History so far:
- Intent: 将 Go 基础课程从 TypeScript 硬编码迁移到 MDX 内容系统，同时保留 React 课程页、sandbox 练习和任务衔接体验。
- Completed milestones:
  - [done] P1-P8 — MDX 基础设施、metadata/content 拆分、13 章内容迁移和章节懒加载已完成。
  - [done] P9-P10 — 练习系统 v2 与 lazy CodeMirror Go 编辑器竖切已完成，并以 ch4/ch11 作为样板。
  - [done] P11-P15 — 第 4/11 章教程级样板、概念地图顺序调整，以及课程改造原则/Skill 固化已完成。
  - [done] P16 — Batch 1 ch01-ch03 教程级补强、练习分层和 metadata/exercises 同步已完成。
- Key changes:
  - 课程章节现在使用 TypeScript metadata + 动态 MDX 正文；练习支持多题、可编辑草稿、重置和运行反馈。
  - 章节改造标准已明确为场景引入、基础概念逐步讲解、对照示例、讲解后概念回看和分层练习。
- Validation:
  - 多轮 `npm run build --prefix web` 与 `git diff --check` 通过；P15 通过 `git diff --check` 和 CLAUDE.md/Skill 文件可读性检查。
  - 用户已确认第 4/11 章样板体验稳定。
  - P16 验证：`npm run build --prefix web` 与 `git diff --check` 通过；ch01/ch02/ch03 warmup starter code 抽查输出匹配；Vite >500kB chunk 警告已消失；用户已确认体验验证通过。
  - P17 当前验证：`npm run build --prefix web` 与 `git diff --check` 通过；ch05/ch06/ch07 共 9 个 starter code 抽查可运行，warmup 输出匹配。
- Deferred / gaps:
  - P17 还需用户确认体验验证结果；后续已明确按 P18 ch08-ch10、P19 ch12-ch13 顺序推进，全部批次完成后做最终构建/空白检查与归档。

Prerequisites:
- 用户确认采用 React + MDX 方案，不引入 VuePress 作为主课程框架。

Resume next: 向用户请求 P17 手动体验验证结果；若通过，checkpoint：P17 -> done、P18 -> doing，并继续 ch08-ch10。

### WF-2026-05-31-001 — 全站 shadcn 视觉系统重构
Status: Active
Level: 3
Started: 2026-05-31
Updated: 2026-05-31
Current phase: P6 — 补充本地开发启动脚本。

Intent: 将前端重构为 shadcn/ui 驱动的统一视觉系统，并优化首页、导航、课程阅读和任务工作台的信息架构。

Plan:
- [done] P1 — 完成需求澄清、视觉方向和信息架构设计确认。
- [done] P2 — 编写并评审设计规格文档。
- [done] P3 — 制定实施计划。
- [done] P4 — 分阶段实施全站 shadcn 重构并验证。
- [done] P5 — 增加全站 light/dark/system 主题支持。
- [done] P6 — 补充本地开发启动脚本，封装 Docker、混合开发和本地服务启动场景。

Current todo:
- [x] P6 — 已新增 `scripts/dev.sh` 启动助手，README 已补充常用场景；脚本语法、help 输出和空白检查通过。

Changes:
- 用户选择方案 B：保持核心功能，但允许优化首页、导航、课程阅读布局和 Dashboard 信息层级。
- 用户选择混合视觉方向：首页/Dashboard 保留平台感，课程页偏文档站，任务页偏工作台，并用 shadcn token 与 Go 蓝统一。
- 用户确认导航骨架采用分区混合壳：首页为公开营销入口，学习区使用统一 App Shell/侧栏。
- 用户确认视觉语言采用自适应混合：课程正文浅色阅读，沙盒/任务反馈局部深色终端。
- 用户确认学习区 Sidebar 采用“工作区 + 学习路径 + 项目”分组，支持 Go 基础、后端实习、工程进阶和 AI 全栈路线。
- 用户确认未实现路线在 UI 中保留入口并标记“即将开放”，不做假功能。
- 用户确认核心页面采用职责重排：首页负责品牌和路径入口，Dashboard 负责下一步行动，课程页负责系统学习，任务页负责动手实战。
- 用户确认采用方案 B：分区 App Shell + 页面职责重排，作为后续设计主方案。
- 用户确认设计第 1 部分：架构与路由分区，公开区使用 PublicLayout，学习区使用 LearningLayout/App Shell。
- 用户确认设计第 2 部分：shadcn 组件系统与 Design Token，采用 Go 蓝 primary、浅色阅读 surface 和局部深色终端 surface。
- 用户确认设计第 3 部分：核心页面布局，包含 Landing、Dashboard、课程总览、章节详情和任务详情的页面职责。
- 用户确认设计第 4 部分：数据流、状态与错误处理，课程/任务数据保持本地，沙盒运行沿用现有 API，状态用 shadcn 组件表达。
- 用户确认设计第 5 部分：验证与实施切片策略。
- 设计规格文档已写入 docs/superpowers/specs/2026-05-31-shadcn-visual-system-redesign.md，并在第三轮独立评审通过。
- 已完成全站 shadcn 视觉系统重构：布局拆分、学习区 Sidebar、首页、Dashboard、课程页、章节页、任务页和练习面板均已迁移；`npm run build --prefix web` 与 `git diff --check` 通过。
- 用户追加确认全站支持深色模式：默认跟随系统，并提供 light/dark/system 手动切换；采用 next-themes。
- 已实现全站 light/dark/system 主题支持：App 接入 next-themes，公开区与学习区顶部加入 ThemeToggle，dark token 保持 Go 蓝 primary，并修正 Landing/课程页硬编码浅色样式；`npm run build --prefix web` 与 `git diff --check` 通过。
- 修复 ThemeToggle 交互：为避免 Radix DropdownMenu trigger 与直接切换点击互相干扰，已改为本地受控主题菜单；按钮点击会打开 light/dark/system 选项，选择后调用 next-themes 并关闭菜单；构建和空白检查通过。
- 用户反馈频繁 `docker compose up --build` 影响开发效率；已新增 `scripts/dev.sh` 封装 full Docker、Docker 后端 + 本地 Vite、本地 Go 服务 + Docker 依赖等启动场景，并更新 README。

History so far:
- Intent: 将前端重构为 shadcn/ui 视觉系统，并优化首页、学习区、课程阅读和任务工作台的信息架构。
- Completed milestones:
  - [done] P1-P3 — 需求澄清、设计规格和实施计划已完成。
  - [done] P4-P5 — 全站 shadcn 重构和 light/dark/system 主题支持已完成。
  - [done] P6 — 本地开发启动脚本与 README 使用场景补充已完成。
- Key changes:
  - 用户选择混合视觉方向、分区 App Shell 和未实现路线保留“即将开放”入口。
  - 开发脚本支持 full Docker、Docker 后端 + 本地 Vite、本地 Go 服务 + Docker 依赖等场景。
- Validation:
  - 视觉/主题阶段 `npm run build --prefix web` 与 `git diff --check` 通过；P6 脚本语法、help 输出和空白检查通过。
- Deferred / gaps:
  - 仍需浏览器 smoke light/dark/system、首页、Dashboard、课程、章节、任务、沙盒锚点和移动端 Sidebar，并可按需实跑开发脚本。

Prerequisites:
- 规格已由用户批准，可开始实施；提交仍需用户明确要求。

Resume next: 如需收尾，进行浏览器人工 smoke（light/dark/system、首页、Dashboard、课程、章节、任务、沙盒锚点和移动端 Sidebar），按需实跑 `./scripts/dev.sh backend`/`web` 检查开发启动体验，然后按用户要求提交。

## Backlog / Future

- [ ] 若课程数据继续膨胀，后续可考虑将 13 章拆分为独立章节文件；当前计划明确首版保留单文件。

## Completed

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

