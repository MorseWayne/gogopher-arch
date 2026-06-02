# Go 课程质量样板实施计划

Date: 2026-06-02
Spec: `docs/superpowers/specs/2026-06-02-go-course-quality-sample-design.md`

## Objective

按已批准的“双轨样板”设计，完整升级 `ch07 Interfaces` 为课程质量样板章，并为 `ch11 Testing` 产出可执行审计与改造蓝图。实现结果应沉淀一套可复用课程质量标准，但本轮不批量改造 13 章，也不完整重写 ch11。

目标学习者是“后端新手到实习”：内容需要同时讲清 Go 基础语义、后端工程场景、测试衔接和代码评审判断。

## Scope

### In scope

- `web/src/content/go-basics/ch07-interfaces.mdx`
- `web/src/content/go-basics/ch11-testing.mdx` 的审计与蓝图记录，不做完整正文改写
- `web/src/content/go-basics/courseChapters.ts` 中 ch07 metadata/exercises
- 必要时微调 `docs/superpowers/specs/2026-06-02-go-course-quality-sample-design.md` 状态或补充说明
- 必要时更新 `.claude/skills/go-course-chapter-redesign/SKILL.md` 或项目 `CLAUDE.md` 中已固化的课程质量标准；只有当实现中发现现有规则不足时才修改
- `.claude/WORKFLOW.md` 恢复状态和验证记录

### Out of scope

- 批量重写全部 13 章
- 完整重写 `ch11-testing.mdx`
- 新增测试运行型 sandbox 或 exercise kind
- 新增复杂 MDX 组件
- 把外部教程做成用户阅读外链合集
- 将接口章节扩展成泛型约束、反射或类型系统高级专题

## Plan

### P1 — 审计当前 ch07/ch11 与来源映射

范围：

- `web/src/content/go-basics/ch07-interfaces.mdx`
- `web/src/content/go-basics/ch11-testing.mdx`
- `web/src/content/go-basics/courseChapters.ts` 中 ch07/ch11 metadata 与 exercises
- 本地 `/home/wayne/source/open/gopl-zh.github.com/ch7/` 和 `/home/wayne/source/open/gopl-zh.github.com/ch11/`
- 必要的 Go 官方文档、Effective Go / Go Code Review Comments、Learn Go with Tests 参考

执行：

1. 列出 ch07 当前结构与设计 spec 的差距：场景主线、基础概念、最小示例、工程示例、常见坑、PracticeBridge、概念回看、工程 checklist、练习锚点。
2. 列出 ch11 当前结构与蓝图差距，但只记录审计结论和后续改造蓝图，不完整改写正文。
3. 为 ch07 确认“主源 + 辅助源”映射：gopl-zh、Effective Go / Go Code Review Comments、Go 官方文档、Learn Go with Tests。
4. 为 ch11 确认后续蓝图映射：gopl-zh、Go 官方 testing 文档、Learn Go with Tests、测试风格参考。
5. 如果官方文档、gopl-zh 和既有正文出现教学口径冲突，暂停并向用户确认。

验收：

- 能明确说明 ch07 要保留、重组、替换和新增的内容。
- 能明确说明 ch11 当前差距和下一轮落地方向。
- 没有把外部来源变成外链阅读合集。

### P2 — 改造 ch07 MDX 正文为接口样板章

范围：

- `web/src/content/go-basics/ch07-interfaces.mdx`

执行：

1. 将主线从当前“构建结果通知”收敛到 spec 中的“订单通知”场景。
2. 按样板结构重组正文：
   - 场景引入：订单通知为什么会变难。
   - 本章问题：业务逻辑如何依赖行为边界而不是具体实现。
   - 基础概念：interface 方法集合、隐式实现、方法集、接口值、`any`、类型断言/type switch。
   - 最小示例：`Notifier` / `EmailSender` / `SendWelcome`。
   - 对照示例：大接口/实现方接口 vs 使用方最小接口。
   - 工程示例：`OrderService`、构造函数注入、`EmailNotifier`/`SMSNotifier`/`LogNotifier`/`SpyNotifier`。
   - 常见坑：过早抽象、接口太大、`any` 滥用、nil interface、接口放错包。
   - PracticeBridge：对应 warmup/core/challenge。
   - 概念回看和工程 checklist。
3. 保留并重组必要的 gopl-zh 经典内容：标准库小接口、接口值模型、nil 接口、类型断言、error 相关语义。
4. 避免把反射、泛型约束接口、复杂类型集作为主线；如确需提及，放入 `DeepDive`。
5. 使用现有 MDX 组件：`SourceNote`、`CompareNote`、`ExamplePair`、`PitfallCard`、`DeepDive`、`PracticeBridge`。

验收：

- ch07 不再只是接口概念串讲，而是形成“耦合问题 → 行为边界 → 最小语义 → 工程应用 → 常见坑 → 练习”的完整递进。
- 概念图不前置，只作为讲解后的回看。
- 章节内容能直接作为后续“概念 + 工程抽象类章节”样板。

### P3 — 更新 ch07 metadata 与 warmup/core/challenge 练习

范围：

- `web/src/content/go-basics/courseChapters.ts` 中 `ch7-interfaces` 条目

执行：

1. 更新 ch07 summary、goals、modernNotes、engineeringPractices、pitfalls、checklist、reviewQuestions，使其和订单通知主线一致。
2. 将 ch07 exercises 调整为三层：
   - `warmup`：定义最小 `Notifier` 接口，实现 `ConsoleNotifier`，调用 `SendWelcome` 输出欢迎消息。
   - `core`：把直接依赖 `EmailSender` 的订单通知服务重构为依赖 `Notifier`，保持输出稳定，并增加替换实现。
   - `challenge`：实现 `SpyNotifier` / failing notifier，观察通知行为和错误传播。
3. 确保 exercise `concepts` 覆盖 `interface`、`implicit implementation`、`dependency injection`、`small interface`、`test seam`、`spy` 等关键词。
4. 确保 starter code 是完整可运行 Go 程序，expected output 稳定，hints 渐进，solutionOutline 检查推理而不只是给答案。
5. 维护旧单个 `exercise` 字段与 `exercises` 数组的一致性：`exercise` 指向 warmup 或与现有课程约定保持一致。

验收：

- ch07 三个练习都能在 sandbox 里运行。
- 输出格式稳定，适合自动匹配。
- 练习和正文 PracticeBridge 一一对应。

### P4 — 落实 ch11 审计与改造蓝图

范围：

- `web/src/content/go-basics/ch11-testing.mdx`，只做轻量蓝图/衔接记录，避免完整改写
- 必要时更新 `courseChapters.ts` 中 ch11 的 non-blocking metadata 提示；若会造成范围膨胀则不改

执行：

1. 按 spec rubric 审计当前 ch11：Concept、Progression、Practice、Source、Backend relevance。
2. 记录当前 ch11 与下一轮蓝图差距：现有 NormalizeName 主线如何迁移或改造为订单通知主线。
3. 在 ch11 正文中只做最小必要的衔接说明或 `DeepDive`/`SourceNote` 补充；不完整重写。
4. 如正文改动会触发大范围重构，改为在 implementation notes 或 Workflow Ledger 中记录后续任务，不扩大本轮范围。
5. 明确下一轮 ch11 落地目标：`FormatOrderMessage`、`OrderService.PlaceOrder`、`SpyNotifier`/`FailingNotifier`、回归测试与 checklist。

验收：

- ch11 的后续改造方向清晰。
- 本轮没有将 ch11 变成完整正文重写。
- ch07 与 ch11 的教学链路在文档或正文衔接中可被读者理解。

### P5 — 验证 ch07 内容、练习和构建

执行：

1. 运行 `npm run build --prefix web`。
2. 运行 `git diff --check`。
3. 抽查 ch07 warmup/core/challenge starter code：用本地 `go run` 或临时文件运行，确认 expected output 匹配。
4. 人工检查 ch07 MDX：来源说明、章节结构、概念回看位置、PracticeBridge、PitfallCard、DeepDive、工程 checklist。
5. 人工检查 ch11 蓝图：确认只做审计/蓝图，不误进入完整改写。

验收：

- 构建通过。
- 空白检查通过。
- ch07 三个 starter code 可运行且输出稳定。
- 内容质量符合 spec rubric。

### P6 — 更新持久状态并准备关闭或进入下一轮

范围：

- `.claude/WORKFLOW.md`
- 必要时 `docs/superpowers/plans/2026-06-02-go-course-quality-sample-implementation.md`
- 必要时课程规则文档或 skill

执行：

1. 记录验证结果、已完成范围和未做事项。
2. 若 ch07 样板通过，将 ch11 完整改造、其他章节推广作为 Backlog/Future 或下一轮计划。
3. 提交本轮实现变更。
4. 向用户汇报 outcome、validation、gaps、下一步建议。

验收：

- Workflow Ledger 能恢复当前状态。
- 本轮已提交或明确说明未提交原因。
- 后续 ch11 或批量推广有清晰入口。

## P1 audit results

Completed: 2026-06-02

### ch07 current state

Current files:

- `web/src/content/go-basics/ch07-interfaces.mdx`
- `web/src/content/go-basics/courseChapters.ts` `ch7-interfaces` entry

Findings:

- Current ch07 already has strong coverage of core interface semantics: method sets, implicit implementation, interface values, typed nil, `any`, type assertions, type switches, `error`, and standard library interfaces.
- Current structure is close to the preferred flow, but the main scenario is “build result notification”; the approved sample design requires a more backend-product-oriented “order notification” scenario that can connect directly to ch11 testing.
- Current exercise set is coherent but not aligned with the new sample route:
  - warmup uses `EmailNotifier` and `build passed`.
  - core uses `PublishBuildResult` and a fake notifier.
  - challenge focuses on typed nil `io.Writer`.
- P2 should retain the strong interface semantics but reorganize them around the order notification story. Standard library examples (`Stringer`, `flag.Value`, `http.Handler`, `sort.Interface`) should support the concept rather than interrupt the main route; `sort.Interface`, interface comparability, type switch, and deeper error classification can move to `DeepDive` or later sections.
- P3 should replace or substantially revise ch07 exercises so all three levels anchor to the order notification route: minimal `Notifier`, `OrderService` refactor, and `SpyNotifier` / failing notifier for test-seam preparation.

### ch11 current state

Current files:

- `web/src/content/go-basics/ch11-testing.mdx`
- `web/src/content/go-basics/courseChapters.ts` `ch11-testing` entry

Findings:

- Current ch11 is already tutorial-grade and covers `go test`, test package modes, failure messages, table-driven tests, random/fuzz testing, test doubles, coverage, benchmark, pprof, examples, concept recap, fragile tests, and engineering review.
- Current main route is `NormalizeName`, not order notification. It is internally coherent, so P4 must not rewrite it wholesale in this slice.
- Current exercises are useful but not aligned with the approved future ch11 route:
  - warmup simulates table-driven add cases.
  - core fixes `NormalizeName`.
  - challenge compares `strings.Builder` behavior before benchmarking.
- P4 should record the delta from the future route rather than replace current content: future ch11 should introduce `FormatOrderMessage(userID, item string)`, table-driven tests for order behavior, `SpyNotifier` / `FailingNotifier`, regression tests, and an order-notification checklist.
- A minimal ch11 touch, if any, should only add a forward/backward bridge to ch07 and future blueprint notes. If this becomes a full rewrite, stop and ask before expanding scope.

### Source mapping

ch07 primary sources:

- `gopl-zh.github.com/ch7/ch7-01.md`: interface as a contract, `io.Writer`, replaceability.
- `gopl-zh.github.com/ch7/ch7-02.md`: interface type, small `Reader`/`Closer`, interface embedding.
- `gopl-zh.github.com/ch7/ch7-03.md`: implementation conditions, method sets, `interface{}` / `any`, compile-time satisfaction assertions.
- `gopl-zh.github.com/ch7/ch7-05.md`: interface value model, dynamic type/value, nil-interface trap, comparability caveat.
- `gopl-zh.github.com/ch7/ch7-10.md`: type assertions and `value, ok` form.
- `gopl-zh.github.com/ch7/ch7-11.md`: structured error classification; modern implementation should prefer `errors.Is` / `errors.As` framing.
- `gopl-zh.github.com/ch7/ch7-13.md`: type switch as a discriminated-union style use of interface values.
- `gopl-zh.github.com/ch7/ch7-15.md`: avoid premature one-implementation interfaces; ask only for what you need.

ch07 auxiliary sources:

- Effective Go / Go Code Review Comments: small interfaces, naming by behavior, consumer-side interfaces, avoid unnecessary abstraction.
- Go official docs: `any`, `io.Reader` / `io.Writer`, `errors.Is` / `errors.As` semantics.
- Learn Go with Tests: interface as a test seam; use only as teaching-method inspiration.

ch11 primary sources:

- `gopl-zh.github.com/ch11/ch11-01.md`: `go test` discovery and temporary test main model.
- `gopl-zh.github.com/ch11/ch11-02.md`: `TestXxx`, failure messages, regression from bug reports, table-driven tests, `t.Fatal` / `t.Error` distinction.
- `gopl-zh.github.com/ch11/ch11-03.md`: coverage as a heuristic, not a proof of correctness.
- `gopl-zh.github.com/ch11/ch11-04.md`: benchmark structure, `b.N`, `-benchmem`, compare before optimizing.
- `gopl-zh.github.com/ch11/ch11-05.md`: pprof and measurement-before-optimization framing.
- `gopl-zh.github.com/ch11/ch11-06.md`: examples as executable documentation.

ch11 auxiliary sources:

- Go official `testing` documentation: subtests, `t.Helper`, `t.Cleanup`, fuzzing, examples, benchmarks.
- Learn Go with Tests: behavior-first testing and using tests as design feedback.
- Effective Go / Go Code Review Comments: readable assertions, stable behavior over implementation details, avoid over-mocking.

### P2/P3 execution implications

- P2 should rewrite ch07 in one coherent pass rather than patching isolated sections, because the scenario switch from build notification to order notification affects introduction, examples, PracticeBridge, and checklist.
- P3 must update ch07 metadata and exercise output together with P2; otherwise the chapter body and exercise panel will diverge.
- P4 should be deliberately small and may be represented as a blueprint section rather than a full ch11 rewrite.

### P4 results

Completed: 2026-06-02

Implemented lightweight ch11 blueprint notes without rewriting the chapter:

- Added a top-level `CompareNote` explaining that current ch11 keeps the `NormalizeName` route, while the future sample route will extend ch07's order notification interface boundary.
- Added a bridge in the “测试替身” section: ch07 `Notifier` maps directly to ch11 fake/spy/stub testing, especially `SpyNotifier` and `FailingNotifier` for behavior and error propagation.
- Added “下一轮蓝图：把订单通知写成测试样板” before the concept recap, specifying the future sequence:
  - `FormatOrderMessage(userID, item string)` minimal test.
  - table-driven order-message tests with `t.Run`.
  - `SpyNotifier` for `OrderService.PlaceOrder` behavior.
  - `FailingNotifier` for error propagation.
  - regression tests for message-format and invalid-order bugs.

Validation:

- `npm run build --prefix web` passed after the ch11 blueprint update.

Scope control:

- ch11 remains on the existing `NormalizeName` route in this slice.
- No exercise, sandbox, or full chapter rewrite was introduced.

## Validation commands

至少运行：

```bash
npm run build --prefix web
git diff --check
```

如果更新 ch07 starter code 或 expected output，额外抽查每个练习：

```bash
go run /tmp/ch07-warmup.go
go run /tmp/ch07-core.go
go run /tmp/ch07-challenge.go
```

实际临时文件名可以不同，但需要逐个确认输出与 metadata 中 `expectedOutput` 匹配。

## Stop conditions

遇到以下情况暂停并请用户确认：

- ch07 需要新增 exercise kind、测试运行型 sandbox 或 UI 交互。
- ch11 蓝图无法轻量记录，必须完整重写正文才能自洽。
- 外部来源与 gopl-zh/既有正文存在明显冲突，需要选择教学口径。
- starter code 需要多文件 module、第三方依赖或非 sandbox 默认能力。
- 构建或练习验证失败且无法快速定位。

## Checkpoint protocol

每完成一个 P 项后：

1. 更新 `.claude/WORKFLOW.md` 中对应 Plan 状态。
2. 运行该 P 项所需的最小验证。
3. 若有文件变更，和 ledger 更新一起提交，除非用户明确要求不提交。
4. 汇报本步完成、验证、gaps、提交状态和下一步，并询问是否继续。
