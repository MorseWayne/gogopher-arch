# DeepTutor 课程内容 Spike — P1 章节选择

Date: 2026-06-03
Plan: `docs/superpowers/plans/2026-06-03-deeptutor-course-content-spike-implementation.md`

## 结论

推荐本轮 DeepTutor 离线内容工作流 Spike 选择：

**`ch10-packages-tools` — 包和工具**
目标文件：`web/src/content/go-basics/ch10-packages-tools.mdx`

选择理由：

- 它不是当前 ch07/ch11 样板链路的一部分，替换风险低于接口和测试章节。
- 包、module、workspace、go command、CI 反馈、`internal`、文档注释和 `go list` 都高度依赖外部资料校准，适合验证 DeepTutor 的开放网页检索和资料综合能力。
- 当前章节已经可用，但仍有明显提升空间：可以把“从单文件原型到可维护服务”的工程主线讲得更像真实后端项目迁移，并强化 package boundary、module/workspace、toolchain/CI 和 API 文档之间的关系。
- 练习链路相对稳定，不需要新增 sandbox 能力；即使正文替换失败，也容易回滚。

本轮不选择 `ch07-interfaces` 和 `ch11-testing`：

- `ch07-interfaces` 已作为课程质量样板章完成正文、metadata 和练习同步，当前应作为人工样板基线，而不是被 DeepTutor 草稿覆盖。
- `ch11-testing` 在当前样板任务中明确只保留 `NormalizeName` 主线并加入轻量蓝图，不做完整重写；本轮选择它会与既有范围控制冲突。

## 评分方法

每章按四个维度粗评，每项 1-5 分：

- **内容深度**：是否讲清定义、语义、初始化/使用方式、常见误区。
- **资料接入价值**：是否适合通过 Go 官方文档、gopl-zh、优秀教程和开放网页检索进一步校准。
- **工程场景**：是否自然服务“后端新手到实习”的真实任务场景。
- **练习衔接**：正文是否能支撑 warmup/core/challenge 练习。

分数不是唯一选择依据。当前 13 章整体已多轮补强，很多章节并不“薄弱”；本轮优先选择“DeepTutor 有机会显著改善、替换风险可控、且不与当前样板任务冲突”的章节。

## 13 章评分表

| 章节 | 内容深度 | 资料接入价值 | 工程场景 | 练习衔接 | 总分 | 证据与判断 |
|---|---:|---:|---:|---:|---:|---|
| `ch1-getting-started` 入门 | 4 | 4 | 4 | 4 | 16 | 已围绕 `package main`、`go run`、`go build`、stdout/stderr/exit code 建立第一反馈模型，并有 `PracticeBridge` 串联练习。改写空间主要是入门讲法，不适合作为开放检索能力的强验证。 |
| `ch2-program-structure` 程序结构 | 4 | 4 | 4 | 4 | 16 | 覆盖命名、声明、变量、短变量声明、指针、类型、初始化和作用域，常见坑明确。可补现代工程边界案例，但主题偏基础语义，DeepTutor 开放检索优势不明显。 |
| `ch3-basic-data-types` 基础数据类型 | 4 | 4 | 4 | 4 | 16 | 整数、浮点、字符串、UTF-8、常量等语义和金额/文本处理场景较完整。可改进示例密度，但替换收益不如工具链或并发章节。 |
| `ch4-composite-types` 复合数据类型 | 5 | 5 | 5 | 5 | 20 | 订单状态统计、slice/map/struct、JSON DTO、稳定输出和复盘链路都很成熟。虽然外部资料价值高，但当前质量已经强，DeepTutor 草稿很难“明显优于原章”。 |
| `ch5-functions` 函数 | 5 | 4 | 5 | 4 | 18 | 函数签名、错误处理、闭包、defer、panic/recover 与工程场景衔接较好。可补现代错误分类和组织模式，但正文结构已稳定。 |
| `ch6-methods` 方法 | 4 | 4 | 4 | 4 | 16 | 值/指针接收者、方法集、嵌入、nil 接收者和封装都有覆盖。替换可提升工程主线，但开放检索价值不如 ch10/ch08。 |
| `ch7-interfaces` 接口 | 5 | 5 | 5 | 5 | 20 | 已升级为订单通知样板章，包含接口语义、隐式实现、nil 陷阱、`any`、类型断言、错误接口、小接口和测试接缝。作为人工样板基线，应避开。 |
| `ch8-goroutines-channels` Goroutines 和 Channels | 5 | 5 | 5 | 4 | 19 | 并发生命周期、channel、pipeline、取消、泄露和并发上限都适合外部资料校准。候选价值高，但并发生成稿技术风险大，死锁/取消语义需要强审校，不适合作为第一轮直接替换目标。 |
| `ch9-shared-variable-concurrency` 基于共享变量的并发 | 4 | 4 | 5 | 4 | 17 | data race、逻辑竞态、Mutex/RWMutex/Once/atomic/race detector 与工程场景强相关。可改善深度，但也有并发语义风险，且比 ch10 更难审计。 |
| `ch10-packages-tools` 包和工具 | 4 | 5 | 5 | 3 | 17 | 当前已覆盖 package、import path、module、go work、toolchain、go command、build tag、文档注释、`internal`、`go list`，但练习衔接和工程主线仍可明显增强。它最适合验证 DeepTutor 是否能把开放资料转化为站内工程教程。 |
| `ch11-testing` 测试 | 5 | 5 | 5 | 5 | 20 | 已覆盖 `go test`、测试包名、表驱动、fuzz、测试替身、coverage、benchmark、pprof、示例函数，并有 ch07 桥接蓝图。当前计划明确不完整重写，应避开。 |
| `ch12-reflection` 反射 | 4 | 5 | 4 | 4 | 17 | 反射的 Type/Value/Kind/tag/CanSet/替代方案都有覆盖，外部资料价值高。可作为后续候选，但反射容易生成过深或过宽内容，第一轮直接替换风险高于 ch10。 |
| `ch13-low-level-programming` 低层编程 | 4 | 4 | 4 | 4 | 16 | `unsafe`、`uintptr`、对齐、cgo、系统调用和 build tag 风险提示到位。主题更容易出现技术和安全表述风险，不适合作为第一个开放检索替换样例。 |

## Top 3 候选

### 1. `ch10-packages-tools` — 推荐目标

优势：

- 开放资料价值高：Go Modules Reference、go command 文档、Effective Go、Go Code Review Comments、workspace/toolchain 文档都可作为校准来源。
- 工程场景强：包边界、导出 API、CI 反馈、内部包和命令契约都贴近后端实习工作。
- 替换风险可控：不涉及并发/unsafe 的高风险运行语义，也不影响 ch07/ch11 样板链路。
- DeepTutor 有明显发挥空间：可以把现有知识点重组为“原型服务 → 包边界 → module/workspace → CI 命令反馈 → 文档和 internal 边界”的完整故事。

风险：

- 需要避免生成“Go 工具命令清单”式外链资料汇编。
- 需要保持现有 `PracticeBridge` 和练习 ID 的衔接，除非后续明确允许改 metadata。

### 2. `ch8-goroutines-channels`

优势：

- 并发生命周期是 Go 学习高价值主题，DeepTutor 可通过 Go blog、context 文档和优秀教程补强。
- 工程场景天然强：取消、泄露、并发上限和结果收集都适合后端任务。

风险：

- 生成内容必须严格校验死锁、channel 关闭所有权、取消和 goroutine 泄露语义。
- 作为第一轮直接替换样例，审计成本偏高。

### 3. `ch12-reflection`

优势：

- 反射资料接入价值高，适合结合官方 `reflect` 文档、`encoding/json` 文档和框架边界案例。
- 可以强化“反射只放在基础设施边界”的工程判断。

风险：

- 主题容易过度扩展到泛型、ORM、validator、序列化框架细节。
- 错误示例或不可验证声明风险比 ch10 高。

## 为什么不选总分最低章节

`ch1`、`ch2`、`ch3`、`ch6`、`ch13` 的粗评分都不高于 `ch10`，但它们不一定更适合本轮 Spike：

- `ch1`-`ch3` 属于基础入门语义，开放检索能提供的增益有限，容易把章节写得过长而不一定更好。
- `ch6` 可以提升，但资料接入价值和工程增量不如 ch10。
- `ch13` 涉及 `unsafe`、cgo、系统调用和底层优化，开放检索风险更高，第一轮直接替换不够稳妥。

因此本轮采用“综合评分 + DeepTutor 改善机会 + 替换风险 + 样板冲突”共同决策，选择 `ch10-packages-tools`。

## P2 输入包需要包含的信息

针对 `ch10-packages-tools`，P2 输入包应包含：

1. **原章节正文**
   - `web/src/content/go-basics/ch10-packages-tools.mdx` 全文。
   - 章节标题、H2/H3 结构、现有 `SourceNote`、`CompareNote`、`ExamplePair`、`PitfallCard`、`PracticeBridge` 位置。

2. **metadata 与练习摘要**
   - `courseChapters.ts` 中 `ch10-packages-tools` 条目的 `summary`、`goals`、`modernNotes`、`engineeringPractices`、`pitfalls`、`checklist`、`reviewQuestions`。
   - `exercise` 和 `exercises` 数组，尤其是 `ch10-toolchain-env`、`ch10-import-summary`、`ch10-internal-rule` 的 id、kind、expectedOutput、outputMatch、hints 和 solutionOutline。

3. **课程风格契约**
   - 场景引入在前，密集概念回看在后。
   - 必须讲清 package/import/module/toolchain/internal/go list 的定义、使用方式、工程语义和常见误区。
   - 外部资料只能作为知识来源层，不得变成外链阅读清单。
   - 输出保持 GoGopher Arch 教程风格，而不是普通博客或命令备忘录。

4. **开放检索建议来源**
   - Go 官方文档：Go Modules Reference、go command、workspace、toolchain、internal packages、build constraints、doc comments。
   - Effective Go / Go Code Review Comments：包命名、导出 API、文档注释、避免 util/common。
   - gopl-zh 第 10 章：包、导入路径、工具、文档、internal、go list 的主线。
   - 可少量参考高质量教程，但必须在审计包中标注，不得拼贴表达。

5. **输出限制**
   - 默认不改 metadata 和练习。
   - 保留或重新建立 `PracticeBridge` 与现有练习 ID 的关系。
   - 若 DeepTutor 认为必须改练习或 expected output，必须单独说明理由，不直接执行。

## P1 验收结果

- 13 章均已按四个维度评分并记录证据。
- 目标章节选择为 `ch10-packages-tools`，理由清晰。
- 已明确避开 ch07/ch11，避免与当前课程质量样板任务冲突。
- 已列出 P2 输入包必须包含的信息。
