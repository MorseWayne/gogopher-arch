# Workflow Ledger

Claude Code 开发工作的轻量级可恢复台账。

## Active

### WF-2026-05-31-002 — Go 基础课程 React + MDX 内容系统
Status: Active
Level: 2
Started: 2026-05-31
Updated: 2026-05-31
Current phase: P15 — Go 课程设计原则与章节改造 Skill 固化完成。

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

Current todo:
- [x] 已完成 React + MDX 基础设施：Vite 接入 MDX/GFM/heading slug，课程数据层支持 MDX 覆盖硬编码章节，第 1 章已迁移为 `web/src/content/go-basics/ch01-getting-started.mdx` 样章。
- [x] 验证通过：`npm run build --prefix web` 与 `git diff --check`。
- [x] 后续批量迁移第 2-13 章时，应先阅读 gopl-zh 对应章节和优秀开源教程，再做精简重组与工程化补充。
- [x] 已迁移第 2-4 章：`ch02-program-structure.mdx`、`ch03-basic-data-types.mdx`、`ch04-composite-types.mdx`，均基于本地 gopl-zh 对应章节重组。
- [x] 验证通过：`npm run build --prefix web` 与 `git diff --check`；构建出现 Vite chunk >500kB 警告，属于内容打包体积提示，未阻断构建。
- [x] 下一批迁移第 5-7 章：函数、方法、接口。
- [x] 已迁移第 5-7 章：`ch05-functions.mdx`、`ch06-methods.mdx`、`ch07-interfaces.mdx`，均基于本地 gopl-zh 对应章节重组。
- [x] 验证通过：`npm run build --prefix web` 与 `git diff --check`；构建仍有 Vite chunk >500kB 体积警告，随着 MDX eager 打包章节增加更明显，后续应考虑章节懒加载。
- [x] 下一批迁移第 8-9 章：Goroutines/Channels 与共享变量并发。
- [x] 已迁移第 8-9 章：`ch08-goroutines-channels.mdx`、`ch09-shared-variable-concurrency.mdx`，均基于本地 gopl-zh 对应章节重组。
- [x] 验证通过：`npm run build --prefix web` 与 `git diff --check`；构建仍有 Vite chunk >500kB 体积警告，当前主 JS 约 633.52 kB，后续迁移剩余章节前建议优先做 MDX 正文懒加载。
- [x] 已优化 MDX 章节正文懒加载：`web/src/content/go-basics/courseChapters.ts` 保存 metadata，MDX 文件只保留正文；`goBasicsCourse.ts` 只 eager 加载 metadata，章节页通过 `loadContent()` 动态 import 对应 MDX。
- [x] 验证通过：`npm run build --prefix web` 与 `git diff --check`；MDX 正文已分出 per-chapter chunks，主 JS 从约 633.52 kB 降至约 517.22 kB（gzip 约 168.76 kB），仍有 >500kB 警告但已不再随正文线性增长。
- [x] 已迁移第 10-13 章：`ch10-packages-tools.mdx`、`ch11-testing.mdx`、`ch12-reflection.mdx`、`ch13-low-level-programming.mdx`，metadata 已写入 `courseChapters.ts`，正文继续按章节动态加载。
- [x] 验证通过：新 MDX 章节单独编译检查通过；`npm run build --prefix web` 与 `git diff --check` 通过；构建仍有 Vite chunk >500kB 警告，主 JS 约 533.86 kB（gzip 约 174.59 kB），第 10-13 章各自拆成 lazy chunks。
- [x] 已完成练习系统 v2 竖切：数据模型支持 `exercises[]` 和练习类型，练习面板支持多练习切换、可编辑代码、localStorage 草稿、重置、运行用户代码；ch4/ch11 已改为样板章节。
- [x] 验证通过：`npm run build --prefix web` 与 `git diff --check`；构建仍有 Vite chunk >500kB 警告，主 JS 约 546.67 kB（gzip 约 178.44 kB）。
- [x] 已完成 CodeMirror Go 编辑器竖切：新增 `GoCodeEditor`，练习区从 Textarea 改为 CodeMirror，支持 Go 语法高亮、行号、括号匹配、基础补全和基于练习 concepts 的课程 snippets；编辑器通过 React lazy/Suspense 按需加载。
- [x] 验证通过：`npm run build --prefix web` 与 `git diff --check`；CodeMirror 被拆成 `GoCodeEditor` lazy chunk 约 460.12 kB（gzip 约 152.99 kB），主 JS 约 546.28 kB（gzip 约 178.40 kB），仍有 Vite chunk >500kB 警告。
- [x] 已完成第 4 章优秀资料整合样板：`CourseMdxContent` 新增 `SourceNote`、`CompareNote`、`ExamplePair`、`DeepDive`、`PitfallCard`、`PracticeBridge` 等 MDX 组件；`ch04-composite-types.mdx` 改为订单状态统计与 API 响应边界场景，整合 gopl-zh 主干、Go by Example 短例子风格、Effective Go 工程实践和练习衔接。
- [x] 验证通过：`npm run build --prefix web` 与 `git diff --check`；第 4 章 lazy chunk 约 32.43 kB（gzip 约 9.31 kB），主 JS 约 549.39 kB（gzip 约 179.00 kB），仍有 Vite chunk >500kB 警告。
- [x] 已根据反馈补强第 4 章基础概念：新增“基础概念地图”，并扩展数组、slice、map、struct 的定义形式、初始化方式、内存/语义模型、值复制/引用共享、可比较性、nil/空值、`make`/字面量、指针传参和常见误区说明。
- [x] 验证通过：`npm run build --prefix web` 与 `git diff --check`；第 4 章 lazy chunk 约 50.05 kB（gzip 约 12.25 kB），主 JS 约 549.39 kB（gzip 约 179.00 kB），仍有 Vite chunk >500kB 警告。
- [x] 已完成第 11 章测试样板改造：`ch11-testing.mdx` 按“基础概念 → 回归场景 → go test 约定 → 测试函数/表驱动/子测试 → 随机与 fuzzing → 命令测试/替身/覆盖率 → benchmark/pprof/example → 常见坑与练习衔接”重写；整合 gopl-zh 第 11 章、Go 官方 testing/fuzz/pprof 文档和 Learn Go with Tests 的先测试后实现思路。
- [x] 验证通过：`npm run build --prefix web` 与 `git diff --check`；第 11 章 lazy chunk 约 39.97 kB（gzip 约 10.61 kB），主 JS 约 549.39 kB（gzip 约 179.00 kB），仍有 Vite chunk >500kB 警告。
- [x] 已根据反馈调整概念地图位置：第 4 章“基础概念地图”改为讲解后的“概念回看：四类复合类型怎么区分”；第 11 章“基础概念地图”改为讲解后的“测试概念回看：go test 工具链怎么串起来”，避免文章开头出现晦涩表格。
- [x] 验证通过：`npm run build --prefix web` 与 `git diff --check`；第 4 章 lazy chunk 约 50.18 kB（gzip 约 12.34 kB），第 11 章 lazy chunk 约 40.07 kB（gzip 约 10.64 kB），主 JS 约 549.39 kB（gzip 约 179.00 kB），仍有 Vite chunk >500kB 警告。
- [x] 已固化课程设计方法：`CLAUDE.md` 增加 Go 课程设计原则；新增 `.claude/skills/go-course-chapter-redesign/SKILL.md`，沉淀章节改造流程、来源使用、MDX 组件使用、练习分层、验证要求和反模式。
- [x] 验证通过：`git diff --check`；并用脚本检查 `CLAUDE.md` 与 Skill 文件非空可读。
- [ ] 下一步建议：浏览器抽样 smoke 第 4/11 章样板组件展示、PracticeBridge 跳转、ch4/ch11 练习切换、CodeMirror 高亮/补全、草稿保存和运行反馈；若体验稳定，再批量按同一标准补强其他章节。

Prerequisites:
- 用户确认采用 React + MDX 方案，不引入 VuePress 作为主课程框架。

Resume next: 浏览器抽样 smoke 第 4/11 章样板组件展示、PracticeBridge 跳转、ch4/ch11 练习切换、CodeMirror 高亮/补全、草稿保存和运行反馈；若体验稳定，再使用 `go-course-chapter-redesign` Skill 批量按“场景引入 → 基础概念逐步讲解 → 对照示例 → 概念回看/总结 → 练习”的标准补强其他章节，后续可继续扩展 exercise kind 的 `test` 模式（main.go + main_test.go / go test）、接入更完整的 gopls/LSP 智能提示，或批量丰富其他章节练习。

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

