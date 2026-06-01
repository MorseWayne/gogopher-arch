# Go 基础课程顺序章节补强设计

Date: 2026-06-01
Status: Draft for implementation planning

## Goal

按章节顺序把 Go 基础课程中尚未样板化的章节补强到教程级标准，延续 ch04 与 ch11 已验证的内容结构、MDX 组件使用方式和练习衔接体验。

## Scope

本轮设计覆盖 `web/src/content/go-basics` 下的 Go 基础课程章节内容与对应 metadata/exercises 检查。

改造批次按章节顺序推进，每轮最多 3 章：

1. Batch 1: `ch01-getting-started.mdx`、`ch02-program-structure.mdx`、`ch03-basic-data-types.mdx`
2. Batch 2: `ch05-functions.mdx`、`ch06-methods.mdx`、`ch07-interfaces.mdx`
3. Batch 3: `ch08-goroutines-channels.mdx`、`ch09-shared-variable-concurrency.mdx`、`ch10-packages-tools.mdx`
4. Batch 4: `ch12-reflection.mdx`、`ch13-low-level-programming.mdx`

`ch04-composite-types.mdx` 与 `ch11-testing.mdx` 已作为样板章节，不做重复大改；每个相邻批次经过时只做一致性检查，避免样板与新章节结构漂移。

## Non-goals

- 不引入 VuePress 或新的课程框架。
- 不把外部链接合集替代站内正文。
- 不在这一轮新增完整 LSP/gopls 能力。
- 不优先扩展新的 exercise kind，除非某章无法用现有练习模型表达核心目标。
- 不重写已稳定的 React 课程页、CodeMirror 编辑器或 sandbox API。

## Chapter content standard

每章按同一套教程级结构补强，但允许根据主题轻微调整顺序：

1. 场景引入：用一个具体工程问题说明本章要解决什么。
2. 基础概念逐步讲解：定义形式、初始化/使用方式、语义模型、常见误区。
3. 最小示例：展示概念最小可运行形态。
4. 工程化示例：把概念放入更接近真实项目的代码边界。
5. 常见坑：说明容易误用、隐藏成本或边界条件。
6. 概念回看：在讲解之后总结概念地图，避免开篇堆密集表格。
7. 练习衔接：用 `PracticeBridge` 或正文说明把 warmup/core/challenge 与对应知识点连接起来。

## Source policy

内容来源层遵循项目既有规则：

- 优先基于本地 gopl-zh 原文精简、重组和润色。
- 结合 Go 官方文档校准语言语义、工具行为和边界条件。
- 可参考优秀开源教程的讲解方式，但最终正文必须转化为 GoGopher Arch 自己的学习路径。
- 外部资料只作为知识来源，不作为用户直接阅读层。

## MDX component use

延续 ch04/ch11 的组件模式：

- `SourceNote`：说明本节知识来源和改编边界。
- `CompareNote`：对照相近概念或不同实践取舍。
- `ExamplePair`：并列展示最小示例与工程化示例。
- `DeepDive`：解释隐藏语义、内存模型、工具行为或性能边界。
- `PitfallCard`：突出常见坑。
- `PracticeBridge`：连接章节正文与练习任务。

组件服务于阅读节奏，不要求每章机械使用全部组件。

## Exercises and metadata

每轮章节改造同步检查 `courseChapters.ts`：

- 保持章节 metadata 与正文标题、学习目标、关键概念一致。
- 优先让每章练习覆盖 warmup/core/challenge 三层目标。
- starter code 只在能明显提升学习闭环时更新。
- 练习变更必须能通过现有运行反馈解释清楚，不引入无法验证的任务。

## Validation

每个批次完成后运行：

- MDX/TypeScript 构建：`npm run build --prefix web`
- 空白检查：`git diff --check`
- 若修改练习 starter code，则抽查运行输出与 expected output 的一致性。
- 对本批次章节做浏览器 smoke：正文渲染、教学组件展示、PracticeBridge、练习切换、编辑器、草稿保存和运行反馈。

## Rollout and review

每个批次作为一个可独立 review 的改造单元：

- Batch 1 完成后先评估阅读结构和练习衔接是否稳定。
- 若 Batch 1 质量稳定，后续批次按同一标准顺序推进。
- 如果某章主题复杂度明显超过批次容量，可以把该章单独切成一轮，但不改变总体章节顺序。

## Risks

- 批量改写内容容易出现风格不一致；用 ch04/ch11 和本设计作为基准。
- 高级章节可能需要更多语义校准；优先核对 Go 官方文档，再组织正文。
- 练习扩展过度会拖慢内容改造；本阶段只做与章节理解直接相关的练习补强。
