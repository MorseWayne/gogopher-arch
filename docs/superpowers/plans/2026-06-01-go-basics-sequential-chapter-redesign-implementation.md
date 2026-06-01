# Go 基础课程顺序章节补强实施计划

Date: 2026-06-01
Spec: `docs/superpowers/specs/2026-06-01-go-basics-sequential-chapter-redesign.md`

## Objective

按每轮最多 3 章、从 ch01 到 ch13 的顺序，逐批把 Go 基础课程未样板化章节补强到 ch04/ch11 的教程级标准。

## Plan

### P16 — Batch 1: ch01-ch03

范围：

- `web/src/content/go-basics/ch01-getting-started.mdx`
- `web/src/content/go-basics/ch02-program-structure.mdx`
- `web/src/content/go-basics/ch03-basic-data-types.mdx`
- `web/src/content/go-basics/courseChapters.ts` 中对应 metadata/exercises

执行：

1. 对照 ch04/ch11 样板，梳理三章当前结构缺口。
2. 基于本地 gopl-zh 与 Go 官方文档校准概念边界。
3. 按“场景引入 → 基础概念 → 最小示例 → 工程化示例 → 常见坑 → 概念回看 → 练习衔接”重写或补强正文。
4. 同步检查三章 exercises 的 warmup/core/challenge 分层和 PracticeBridge 衔接。
5. 运行验证并做浏览器 smoke。

验收：

- 三章不再只是概念串讲，而有清晰场景、概念递进、坑点和练习衔接。
- `npm run build --prefix web` 通过。
- `git diff --check` 通过。
- 章节页面、组件渲染、练习切换、编辑器、草稿保存和运行反馈 smoke 通过。

### P17 — Batch 2: ch05-ch07

范围：

- `ch05-functions.mdx`
- `ch06-methods.mdx`
- `ch07-interfaces.mdx`
- 对应 metadata/exercises

执行重点：

- 函数：错误返回、闭包、defer/panic/recover 的工程边界。
- 方法：值/指针接收者、方法集、嵌入和封装。
- 接口：隐式实现、nil 接口陷阱、小接口、测试替身。

验收同 P16。

### P18 — Batch 3: ch08-ch10

范围：

- `ch08-goroutines-channels.mdx`
- `ch09-shared-variable-concurrency.mdx`
- `ch10-packages-tools.mdx`
- 对应 metadata/exercises

执行重点：

- 并发：goroutine 生命周期、channel 同步、泄露、数据竞争、锁保护不变量。
- 工具：package/module/import/go test/go vet/go fmt 等工具链反馈回路。

验收同 P16；并发章节额外关注示例是否避免误导性 race 或 goroutine 泄露写法。

### P19 — Batch 4: ch12-ch13

范围：

- `ch12-reflection.mdx`
- `ch13-low-level-programming.mdx`
- 对应 metadata/exercises

执行重点：

- 反射：Type/Value/Kind、可寻址/可设置、结构体标签、适用边界。
- 底层编程：unsafe、cgo、syscall、内存布局、优化前先测量。

验收同 P16；高级章节额外核对 Go 官方文档，避免过度鼓励 unsafe/reflect。

## Batch workflow

每个批次按以下固定顺序执行：

1. 读取当前章节与样板章节，列出结构缺口。
2. 查阅本地来源材料和官方文档，确定本批每章的场景主线。
3. 修改 MDX 正文和必要的 metadata/exercises。
4. 运行构建和空白检查。
5. 浏览器 smoke 本批章节。
6. 更新 Workflow Ledger：完成当前 P 项，切到下一批或准备关闭任务。
7. 向用户汇报本批结果、验证、风险和下一批建议。

## Validation commands

每批至少运行：

```bash
npm run build --prefix web
git diff --check
```

如果更新 starter code 或 expected output，则增加针对性运行抽查，确认练习反馈与预期一致。

## Stop conditions

遇到以下情况时暂停并向用户确认：

- 某章需要新增 exercise kind 或改动 sandbox 执行模型。
- 某章内容需要超出本批范围的大型 UI/交互改造。
- 官方文档与既有正文或 gopl-zh 表述存在明显冲突，需要选择教学口径。
- 构建、练习运行或浏览器 smoke 出现无法快速定位的问题。
