# DeepTutor 离线课程内容工作流 Spike 实施计划

Date: 2026-06-03
Spec: `docs/superpowers/specs/2026-06-03-deeptutor-course-content-spike-design.md`

## Objective

按已批准的 DeepTutor Spike 设计，验证 DeepTutor 能否作为离线课程内容研究工具，利用开放网页检索与站内课程约束，生成一版明显优于原章节的教程级 Go 基础章节正文草稿。

本轮以内容质量为第一成功标准。DeepTutor 不接入线上运行时，不作为生产依赖；AI 导师能力仅保留后续扩展经验。

## Scope

### In scope

- 审查 `web/src/content/go-basics/` 下 13 个 MDX 章节。
- 审查 `web/src/content/go-basics/courseChapters.ts` 中对应 metadata、练习和复盘信息。
- 选择一个最适合 Spike 的薄弱章节。
- 使用 DeepTutor 做开放资料检索、资料综合和章节正文草稿生成。
- 将通过初筛的草稿直接替换目标 `web/src/content/go-basics/chXX-*.mdx`。
- 产出完整审计包到 `docs/superpowers/spikes/deeptutor-course-content/`。
- 运行构建、空白检查和内容质量评估。
- 输出最终结论：保留、人工修改后保留或回滚。

### Out of scope

- 不接入线上 AI Tutor UI。
- 不将 DeepTutor 加入 Docker Compose 或生产服务链路。
- 不批量自动重写 13 章。
- 不自动修改目标章节 metadata、练习或路由，除非执行中发现正文必须同步最小修正并经用户确认。
- 不新增 MDX 组件或 sandbox exercise kind。
- 不把开放网页资料作为站内课程外链合集。

## Plan

### P1 — 审查 13 章并选择目标章节

范围：

- `web/src/content/go-basics/ch01-getting-started.mdx` 至 `ch13-low-level-programming.mdx`
- `web/src/content/go-basics/courseChapters.ts`
- 已固化课程原则：项目 `CLAUDE.md` 与 Go 课程设计原则

执行：

1. 为每个章节建立简短评分记录，维度为：
   - 内容深度：定义、语义、初始化/使用方式、常见误区是否讲清。
   - 资料接入价值：该主题是否适合通过官方文档、gopl-zh 和开放资料补强。
   - 工程场景：是否自然服务“后端新手到实习”任务场景。
   - 练习衔接：正文是否支撑 warmup/core/challenge 练习。
2. 每项按 1-5 分粗评，并记录 1-2 句证据。
3. 选择一个最适合 Spike 的章节，而不是必然选择总分最低章节；优先选择“DeepTutor 有机会显著改善”的章节。
4. 将评分表写入审计包目录的 `chapter-selection.md`。

验收：

- 13 章均有评分与选择证据。
- 目标章节选择理由清晰。
- 选择不与当前正在进行的课程质量样板任务冲突；如冲突，说明为什么仍适合。

### P2 — 准备 DeepTutor 输入包与课程风格契约

范围：

- 目标章节 MDX 原文。
- 目标章节 metadata、练习摘要、review questions。
- 课程风格契约和输出格式说明。
- DeepTutor 使用提示模板。

执行：

1. 创建 `docs/superpowers/spikes/deeptutor-course-content/input-package.md`。
2. 写入目标章节信息：slug、title、当前问题摘要、原章节结构、练习摘要。
3. 写入 Course Style Contract：
   - 具体场景引入；
   - 基础概念逐步讲解；
   - 最小示例；
   - 工程化示例；
   - 常见坑；
   - 工程实践；
   - 概念回看；
   - 与分层练习的衔接说明。
4. 写入开放检索要求：允许网页检索，但必须优先权威来源，并为主要段落保留来源依据。
5. 写入 DeepTutor 输出格式要求：
   - MDX 正文草稿；
   - 来源清单；
   - 段落级来源映射；
   - 搬运风险；
   - 与原章节差异；
   - 人工审校 checklist。

验收：

- 输入包足以让 DeepTutor 独立理解目标章节、课程风格和输出格式。
- 输入包明确禁止拼贴外部教程与外链替代。
- 输入包可供后续 AI 导师方向复用。

### P3 — 安装/运行 DeepTutor 并生成章节正文草稿

范围：

- DeepTutor 本地 Web/交互式运行；优先方案 A。
- 可选 DeepTutor CLI；仅作加分验证。
- `docs/superpowers/spikes/deeptutor-course-content/deeptutor-run-notes.md`

执行：

1. 确认 DeepTutor 推荐安装与运行方式。
2. 本地启动 DeepTutor 或确认无法运行的阻塞原因。
3. 将 P2 输入包交给 DeepTutor，要求开放检索和课程正文生成。
4. 保留运行记录：使用方式、关键提示、是否用到 CLI、遇到的限制。
5. 如果 CLI 支持稳定 JSON/Markdown 输出，记录命令模板；否则只记录 Web/交互式过程。
6. 将生成草稿暂存为 `generated-draft.mdx`，将来源和审计初稿暂存为 `source-audit-draft.md`。

验收：

- 若 DeepTutor 可运行：得到可阅读的章节正文草稿和来源记录。
- 若 DeepTutor 不可运行：记录可复现阻塞，且不修改课程章节。
- CLI 可用性有明确结论：可用、不可用或未验证。

### P4 — 审计草稿并替换目标 MDX

范围：

- DeepTutor 生成草稿。
- 目标章节 `web/src/content/go-basics/chXX-*.mdx`。
- 审计包目录。

执行：

1. 对草稿进行初筛：结构是否符合 Course Style Contract，是否明显优于原章，是否存在显著幻觉或拼贴风险。
2. 完成审计包 `audit-report.md`：
   - 来源清单；
   - 段落级来源映射；
   - 搬运风险；
   - 与原章节差异；
   - 不可验证声明；
   - 人工审校 checklist；
   - 初步保留/修改/回滚建议。
3. 若初筛通过，将草稿替换目标 MDX 正文。
4. 若初筛不通过，不替换课程正文；记录失败原因和 DeepTutor 能力结论。
5. 替换时优先保持课程文件已有导入、MDX 组件使用方式和 frontmatter/文件结构不被破坏；如原文件没有 frontmatter，不新增不必要 frontmatter。

验收：

- 审计包完整可读。
- 替换后的目标 MDX 保持项目现有 MDX 风格。
- 任何高风险来源或不可验证技术声明都有明确处理结论。

### P5 — 验证内容质量、构建和回滚决策

范围：

- 替换后的目标 MDX。
- 审计包。
- DeepTutor 能力评估结论。

执行：

1. 运行：

   ```bash
   npm run build --prefix web
   git diff --check
   ```

2. 按内容质量 Rubric 做人工评分：
   - 技术准确性；
   - 概念深度；
   - 场景引入质量；
   - 示例解释质量；
   - 工程实践价值；
   - 常见坑覆盖；
   - 概念回看位置；
   - 与练习和复盘衔接。
3. 对比原章节与 DeepTutor 草稿，判断是否明显更好。
4. 根据结果选择：
   - 保留；
   - 人工修改后保留；
   - 回滚。
5. 若选择“人工修改后保留”，只做必要编辑，不扩大为完整人工重写；如果需要大改，暂停并向用户确认。

验收：

- 构建通过或失败原因明确。
- 空白检查通过。
- 内容质量结论明确。
- 最终课程状态与审计结论一致。

### P6 — 记录结论、更新 ledger 并提交

范围：

- `docs/superpowers/spikes/deeptutor-course-content/`
- 目标章节 MDX（如果保留或修改后保留）
- `.claude/WORKFLOW.md`
- 必要时本实施计划

执行：

1. 在审计包中写入最终 `decision.md` 或在 `audit-report.md` 中补齐 Final decision。
2. 更新 `.claude/WORKFLOW.md`：记录 DeepTutor 是否可运行、目标章节、内容质量结论、验证结果和 gaps。
3. 如果保留或修改后保留，提交目标章节、审计包、计划/ledger 更新。
4. 如果回滚，提交审计包、结论和 ledger 更新，不保留失败正文。
5. 汇报 outcome、validation、gaps、commit status 和下一步建议。

验收：

- Workflow Ledger 能恢复本 Spike 状态。
- 结论能够回答“DeepTutor 值不值得继续用于课程内容生产”。
- 变更已提交，或明确说明未提交原因。

## Stop conditions

遇到以下情况暂停并请用户确认：

- DeepTutor 安装/运行需要较重外部依赖、账号、付费 API 或长时间环境配置。
- DeepTutor 输出明显涉及大段外部教程近似表达，无法通过重写消除风险。
- 目标章节草稿需要同步大改 metadata、练习、路由或新增组件才自洽。
- 开放网页来源与 Go 官方文档、gopl-zh 或站内课程出现关键技术口径冲突。
- 生成草稿质量不达标，但需要人工完整重写才可能保留。
- `npm run build --prefix web` 失败且无法快速定位为 MDX 格式问题。

## Validation commands

基础验证：

```bash
npm run build --prefix web
git diff --check
```

可选人工验证：

```bash
./scripts/dev.sh web
```

仅当需要检查 MDX 页面渲染、交互或布局时启动前端人工抽查。

## Checkpoint protocol

每完成一个 P 项后：

1. 更新 `.claude/WORKFLOW.md` 中对应 Plan 状态。
2. 运行该 P 项需要的最小验证。
3. 若有文件变更，和 ledger 更新一起提交，除非用户明确要求不提交。
4. 汇报本步完成、验证、gaps、提交状态和下一步。
5. 等用户确认后继续下一 P 项。
