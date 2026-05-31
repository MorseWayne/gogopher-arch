## Workflow Ledger

使用 `workflow-ledger` 跟踪可恢复的开发工作。

- Workflow Ledger 只有在项目初始化后才生效。如果缺少 `.claude/WORKFLOW.md`，不要把普通开发任务套用到此工作流；先运行 `npx workflow-ledger init`。
- 执行前先分级：Level 0 问答/只读/发布 tag，Level 1 轻量编辑，Level 2 标准代码工作，Level 3 复杂工作。
- Level 2/3 任务，以及用户希望跨会话跟踪的任务，都维护在 `.claude/WORKFLOW.md` 中。
- 被跟踪的工作只记录恢复状态：`Intent`、可变的 `Current todo`、`Prerequisites`、可选的 `Blocked by`，以及一个具体的 `Resume next`。
- 关闭工作前，把任务移到 `Completed`，写简短 `Close summary`：outcome、validation、gaps；验证失败时任务保持 In Progress 或 Blocked。
- 记录依赖和发现的未来任务；先完成阻塞当前工作的前置条件，把非阻塞发现延后到 Backlog/Future。
- 当前会话执行用 TodoWrite；里程碑历史和恢复点用 `.claude/WORKFLOW.md`。
- 不要创建附件或额外 spec 文件，除非 Level 3 工作确实需要，或用户明确要求。

不要找理由跳过 ledger：

- “这个很小”仍然需要分级；Level 2/3 工作必须跟踪。
- “我之后再更新”不安全；在重要待办/范围变化、阻塞、验证结果和交接点更新。
- TodoWrite 是会话本地状态；`.claude/WORKFLOW.md` 是持久恢复状态。
- 保持核心字段稳定，让 `workflow-ledger doctor` 能检查 ledger。

## Go 课程设计原则

这些规则适用于 Go 基础训练营和后续 Go 课程内容改造。

- 课程内容以内置正文为主，不用外链合集替代站内教程。
- 不要凭空编写教程内容；优先基于本地 `/home/wayne/source/open/gopl-zh.github.com` 原文精简、重组、润色，并结合 Go 官方文档和优秀开源教程校准。
- 外部资料是知识来源层，不是用户直接阅读层；最终要转化成 GoGopher Arch 自己的学习路径。
- 章节应先用具体场景引入，再逐步讲基础概念、最小示例、工程化示例、常见坑、工程实践、概念回看和分层练习。
- 基础概念必须讲清定义形式、初始化/使用方式、语义或内存模型、常见误区，不能只做工程化包装。
- 不要把密集“基础概念地图”放在文章开头；应在具体讲解之后作为“概念回看/总结”出现。
- 练习按 warmup / core / challenge 分层，并用 PracticeBridge 或正文说明把练习和对应知识点连接起来。
- 改造课程章节时优先使用 `.claude/skills/go-course-chapter-redesign/SKILL.md` 中的流程；只有需要大量独立调研或评审时再考虑 custom agents。
