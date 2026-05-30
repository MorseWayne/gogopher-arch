# Workflow Ledger

Claude Code 开发工作的轻量级可恢复台账。

## Active

None.

## Backlog / Future

- [ ] 若课程数据继续膨胀，后续可考虑将 13 章拆分为独立章节文件；当前计划明确首版保留单文件。

## Completed

### WF-2026-05-30-001 — Go 基础训练营完整内置课程重制
Completed: 2026-05-30
Level: 3

Close summary:
- Outcome: 已按计划将 Go 基础训练营改为 GoGopher Arch 完整内置课程；重写课程数据模型和 13 章内容，改造课程总览页、章节详情页、Landing 文案和 README 来源边界；课程页面不再依赖外部教程正文链接。
- Validation: 运行文本检查，确认课程数据无旧 source 字段/旧概念模型，课程页面无外部原文入口文案；通过 esbuild bundle 调用 `validateGoBasicsCourse()`，结果 `[]`；本地 `go run` 验证 13 个 exercise starterCode 输出匹配；通过 `npm run build --prefix web`。
- Gaps: 未启动浏览器逐页人工抽样；未验证 sandbox 服务实际运行按钮，因为本次未启动 Gateway/Sandbox Engine。

