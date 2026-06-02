# Go 课程质量样板设计：ch07 Interfaces 与 ch11 Testing 蓝图

Date: 2026-06-02
Status: Approved for implementation planning

## 背景

GoGopher Arch 已完成 Go 基础课程的 React + MDX 内容系统、13 章内置正文迁移、分层练习与视觉系统重构。当前重点从“把课程做出来”转向“把课程做得更像高质量训练营”。

本轮不追求批量重写 13 章，而是采用双轨样板策略：

1. 建立可复用课程质量标准。
2. 完整升级 `ch07 Interfaces` 为样板章。
3. 为 `ch11 Testing` 产出审计与改造蓝图，下一轮再落地。

目标学习者是“后端新手到实习”：已经能写基础 Go 程序，但需要理解如何把语法放进后端工程、测试、重构和代码评审语境中。

## 目标

- 建立一套可复用的 Go 课程内容质量标准，指导后续章节迭代。
- 用 `ch07 Interfaces` 打样“概念 + 工程抽象”类章节。
- 用 `ch11 Testing` 蓝图打样“验证 + 实践反馈”类章节。
- 让 `ch07` 与 `ch11` 形成连续教学链路：先设计可替换行为边界，再用测试保护边界行为。
- 保持课程内容以内置 MDX 正文为主，把优秀教程作为知识来源层，而不是外链合集。

## 非目标

本轮不做：

- 批量改造全部 13 章。
- 完整重写 `ch11 Testing` 正文。
- 新增测试运行型 sandbox。
- 新增复杂 MDX 组件，除非现有组件明显不足。
- 把外部教程直接作为用户阅读入口合集。
- 为了覆盖所有高级主题而牺牲主线清晰度。

## 课程质量标准

### 来源吸收标准

每章需要说明不同来源承担的角色，避免无差别搬运：

- `gopl-zh`：章节主脉络、经典概念、语义深度和原理背景。
- Go 官方文档：现代 Go 行为、标准库 API、工具链、`testing` 等准确性校准。
- Effective Go / Go Code Review Comments / Uber Go Style Guide：工程风格、命名、接口边界、可维护性。
- Go by Example：短小可运行示例的表达方式。
- Learn Go with Tests：从行为和测试反推设计的教学方法，尤其适合 `ch07`/`ch11`。
- Exercism / Gophercises：练习灵感，需要改造成本站后端场景，不能直接复制题目。

执行时采用“主源 + 辅助源”规则：每章选择 2–3 个主源承担核心内容校准，再选择最多 2 个辅助源补充示例风格或练习灵感。若参考池超过 5 个来源，需要在章节设计中明确哪些来源本轮保留、改写或剔除，避免执行时无边界扩张。

每章至少应明确 2–4 个来源如何被吸收，并通过 `SourceNote`、metadata 或正文说明体现“来源如何转化为 GoGopher Arch 的学习路径”。

### 章节结构标准

每章默认采用以下顺序：

1. 场景引入：用后端新手能理解的真实任务开场。
2. 本章要解决的问题：把学习目标变成具体任务。
3. 基础概念：讲清语法、定义形式、使用方式和语义模型。
4. 最小示例：能独立运行，验证单个概念。
5. 工程化示例：放进后端语境，如 handler、repository、service、任务调度、测试替身。
6. 常见坑和反例：讲症状、原因和修正方式。
7. PracticeBridge：明确正文知识点如何进入练习。
8. 概念回看：在学习者已有上下文之后再总结概念地图。
9. 工程 checklist：帮助学习者判断真实项目中是否写得合理。

### 示例质量标准

示例要形成递进链，而不是散点代码。每章至少包含：

- 一个最小示例：只讲一个概念。
- 一个对照示例：bad vs good、具体类型 vs 接口、重复代码 vs 抽象边界等。
- 一个工程示例：贴近后端工作，如订单通知、配置读取、存储适配、测试替身。
- 一个失败或误用示例：展示初学者常犯错误。

每个示例都应能回答：这个代码为什么存在、它展示哪个概念、真实项目中哪里会遇到。

### 练习质量标准

练习继续采用三层：

- `warmup`：确认基础概念能跑通。
- `core`：完成一个有业务语境的小任务。
- `challenge`：训练调试、重构、测试、边界处理或设计判断。

新增要求：每个练习都要有明确正文锚点。练习不能只是“另一个题目”，而要能回指本章某个概念、示例、坑或 checklist。

### 审计 rubric

后续每章按五个维度审计：

- **Concept**：基础定义、语义模型、常见误区是否讲清。
- **Progression**：是否从最小示例自然走到工程示例。
- **Practice**：练习是否分层，并和正文强绑定。
- **Source**：是否正确吸收优秀教程，而不是复制或外链替代。
- **Backend relevance**：是否服务“后端新手到实习”的学习目标。

## ch07 Interfaces 样板章设计

### 样板定位

`ch07 Interfaces` 不只讲“interface 是一组方法”，而是帮助后端新手理解：

> 接口是 Go 中表达行为边界的工具。好的接口不是为了炫技抽象，而是为了让业务代码依赖稳定行为、隐藏可替换细节，并让测试更容易。

本章作为“概念 + 工程抽象”类章节样板。

### 主线场景

本章用订单通知流程串联全文：

1. 第一版订单服务直接调用 `EmailSender`。
2. 后续需要增加短信、日志、测试替身或失败处理。
3. 业务逻辑和具体发送方式耦合过紧。
4. 抽出最小 `Notifier` 接口，让订单服务依赖行为而不是具体实现。
5. 用不同实现替换通知方式，并为 `ch11 Testing` 铺垫 fake/spy/stub。

这条主线表达一个核心判断：接口应来自调用方真实需要，而不是实现方提前设计的一大套能力。

### 推荐正文结构

#### 1. 场景引入：订单通知为什么会变难

开场使用小型后端场景：用户下单后需要发送通知。第一版直接调用 `EmailSender`，后来产品要加短信，测试要替换发送器，失败要记录日志，耦合问题由此出现。

#### 2. 基础概念：interface 到底定义了什么

需要讲清：

- `type Notifier interface { Notify(userID string, message string) error }`
- interface 描述行为集合，不描述字段。
- Go 是隐式实现：方法集匹配即可实现接口。
- 接口值包含动态类型和动态值。
- `any` 是空接口别名，适合特定边界，不是默认设计工具。
- 类型断言和 type switch 用于运行时恢复具体能力；过度使用可能说明接口设计不清。

来源映射：

- 主源 `gopl-zh`：接口值、动态类型、动态分派、类型断言。
- 主源 Effective Go / Go Code Review Comments：接口命名、行为抽象、小接口和避免不必要抽象。
- 辅助源 Go 官方文档：`any`、标准库接口如 `io.Reader` / `io.Writer`，用于现代语义校准。
- 辅助源 Learn Go with Tests：只吸收“接口作为测试接缝”的教学方法，为 `ch11` 衔接服务。

#### 3. 最小示例：从具体类型到接口

示例主题：`EmailSender` 隐式实现 `Notifier`，`SendWelcome` 只依赖 `Notifier`。

教学重点：

- `EmailSender` 不需要声明 `implements Notifier`。
- 方法签名匹配即可传给依赖接口的函数。
- 使用方只关心 `Notify` 行为，不关心底层实现。

#### 4. 对照示例：接口不是越早越好

使用 `ExamplePair` 展示：

- Bad：实现方定义很大的 `NotificationService` 接口，包含调用方不需要的方法。
- Good：使用方附近定义最小 `Notifier` 接口，只描述当前业务真正需要的行为。

核心观点：接口应该来自调用方的需要，而不是实现方的野心。

#### 5. 工程示例：订单服务依赖行为边界

引入：

- `OrderService` 持有 `Notifier`。
- `NewOrderService(notifier Notifier)` 通过构造函数注入依赖。
- `PlaceOrder` 完成业务逻辑后调用通知器。
- 可替换实现包括 `EmailNotifier`、`SMSNotifier`、`LogNotifier`、`FakeNotifier` 或 `SpyNotifier`。

这一节明确为 `ch11 Testing` 铺垫：接口让测试不必依赖真实外部系统。

#### 6. 常见坑

至少覆盖五个 `PitfallCard`：

1. 过早抽象：只有一个实现、没有替换需求，却先定义巨大接口。
2. 接口太大：调用方为了一个方法，被迫实现很多无关方法。
3. 把 `any` 当万能设计：函数接收 `any`，内部到处类型断言。
4. nil interface 陷阱：接口值的动态类型和动态值导致 nil 判断意外。
5. 接口放错包：实现方定义一堆接口，使用方被迫接受。

#### 7. PracticeBridge

练习与正文主线绑定：

- `warmup`：定义最小接口并验证隐式实现。
- `core`：把订单通知从具体实现重构为接口依赖。
- `challenge`：用 spy/failing notifier 观察通知行为和错误传播。

#### 8. 概念回看

在文章后半段整理概念地图：

- interface type
- concrete type
- implicit satisfaction
- dynamic type / dynamic value
- type assertion / type switch
- `any`
- small interface
- consumer-defined interface
- test seam

概念图只能作为回看，不放在文章开头。

#### 9. 工程 checklist

结尾提供判断清单：

- 这个接口有没有明确调用方？
- 调用方真的需要接口里的每个方法吗？
- 是否已经出现多个实现、测试替身或外部依赖边界？
- 接口名是否表达行为，如 `Reader`、`Writer`、`Notifier`？
- 是否可以先用具体类型，等需求稳定后再抽象？
- 测试是否因此更简单，而不是更复杂？

### ch07 练习设计

#### warmup：定义一个最小接口

任务：

- 定义 `Notifier` 接口。
- 实现 `ConsoleNotifier`。
- 调用 `SendWelcome` 输出欢迎消息。

考察点：interface 方法集合、隐式实现、函数参数依赖接口而不是具体类型。

#### core：重构订单通知服务

任务：

- 给定直接依赖 `EmailSender` 的 `OrderService`。
- 将其改为依赖 `Notifier`。
- 保持订单确认输出稳定。
- 增加 `SMSNotifier` 或 `LogNotifier` 实现。

考察点：使用方定义小接口、构造函数注入依赖、业务逻辑和外部通知实现解耦。

#### challenge：用测试替身观察通知行为

任务：

- 实现 `SpyNotifier`，记录收到的 `userID` 和 `message`。
- 当通知失败时，让 `PlaceOrder` 返回错误。
- 编写或补全检查逻辑，确认订单服务调用了通知器。

如果当前练习系统暂不运行测试，可用可运行程序模拟 spy 输出：

```text
notified user=u-1001 message="order placed: book"
error propagated=true
```

考察点：接口作为 test seam、fake/spy 的基本思路、错误路径不要吞掉、为 `ch11 Testing` 铺垫。

### ch07 样板价值

`ch07` 完整改造后应展示一套可复用教学范式：

1. 先让学习者遇到耦合问题。
2. 再引出接口作为行为边界。
3. 用最小示例讲清语法和语义。
4. 用工程示例展示为什么值得用。
5. 用坑和 checklist 防止过度抽象。
6. 用练习把“会写”推进到“会判断”。

反射、复杂类型集、泛型约束接口等高阶主题不进入主线，可放入 `DeepDive` 或后续章节关联。

## ch11 Testing 审计与改造蓝图

### 蓝图定位

`ch11 Testing` 本轮不完整改写，而是用同一质量标准做审计与蓝图。它和 `ch07` 的关系是：

> ch07 让学习者知道如何设计可替换的行为边界；ch11 让学习者知道如何用测试保护这些边界的行为。

本章核心句：测试不是为了追求覆盖率数字，而是为了把重要行为固定下来，让你敢改代码。

### 来源吸收方向

来源映射：

- 主源 `gopl-zh`：testing 包基础、测试函数命名、benchmark、example、coverage 等经典主题。
- 主源 Go 官方文档：`testing` 标准库、`go test`、subtest、`t.Helper`、`t.Fatal` / `t.Error`、examples、benchmarks、fuzzing 的现代行为校准。
- 主源 Learn Go with Tests：从行为需求开始，用测试推动设计，写有用的错误信息。
- 辅助源 Effective Go / Go Code Review Comments：测试可读性、命名、断言、不过度 mock。
- 辅助源 Exercism / Gophercises：小而真实的验证任务灵感，需改造成本站后端场景。

### 推荐结构蓝图

#### 1. 场景引入：你敢改订单通知代码吗

延续 `ch07`：已经抽出 `Notifier`，现在要修改 `OrderService` 的消息格式、失败传播或重试逻辑。没有测试时只能靠手工运行；测试的价值由此出现。

本章保持单一主线：订单通知。若需要最小函数作为 warmup，使用从主线中抽出的 `FormatOrderMessage(userID, item string) string`，它负责生成 `OrderService.PlaceOrder` 将要发送的通知文本。这样 warmup、core 和 challenge 都围绕同一个业务语境展开。

#### 2. 基础概念：Go 测试的最小模型

讲清：

- `*_test.go`
- `func TestXxx(t *testing.T)`
- `go test ./...`
- `t.Error` vs `t.Fatal`
- 测试验证行为，不复述实现细节
- Arrange / Act / Assert 结构

#### 3. 最小示例：测试订单消息格式

用 `FormatOrderMessage("u-1001", "book")` 建立最小测试模型。这个函数不是孤立玩具，而是订单通知主线中的纯函数边界：它把 userID 和 item 转成稳定消息文本，后续 `OrderService.PlaceOrder` 会通过 `Notifier` 发送这条消息。

最小示例需要给出明确输入输出，例如：

```text
user u-1001 ordered book
```

#### 4. 表驱动测试

围绕 `FormatOrderMessage` 和订单业务规则覆盖多组输入和边界，例如不同商品名、空 userID、空 item 或特殊字符。强调表驱动测试适合“同一行为，多组输入输出”，不是所有测试都必须表驱动。

#### 5. subtest 与错误信息

引入 `t.Run`、case name、got/want 和业务语境，让失败信息可定位。示例应展示可读失败信息，而不是只输出 `failed`。

#### 6. fake/spy/stub：从 ch07 接口衔接

使用 `Notifier` 作为测试接缝：

- `SpyNotifier` 检查订单服务是否发送正确通知。
- `FailingNotifier` 检查错误是否向上传递。
- 不需要真实邮件服务参与单元测试。

同时区分 fake、spy、stub、mock。初学阶段优先 fake/spy/stub，不引入复杂 mock 框架。

#### 7. 回归测试

设计小 bug：例如通知失败被吞掉、消息格式漏掉 userID、空商品名也能下单。教学顺序为：描述 bug → 写失败测试 → 修代码 → 测试变绿 → 解释防回归价值。

#### 8. DeepDive：coverage / benchmark / fuzzing / examples

这些主题可讲，但不作为主线：

- coverage：帮助发现未测试路径，不等价于质量。
- benchmark：比较性能，不证明业务正确。
- fuzzing：适合输入空间复杂函数。
- examples：适合文档化 API 用法。

#### 9. 常见坑

至少覆盖：

1. 只测 happy path。
2. 测试实现细节。
3. 错误信息不可读。
4. 过度 mock。
5. 覆盖率崇拜。

#### 10. PracticeBridge

练习与 `ch07` 串联，并全部锚定订单通知主线：

- `warmup`：给 `FormatOrderMessage(userID, item string)` 写最小测试，验证消息格式稳定。
- `core`：为 `OrderService.PlaceOrder` 写表驱动测试，覆盖成功、边界和失败输入。
- `challenge`：用 `SpyNotifier` / `FailingNotifier` 验证通知行为和错误传播。

#### 11. 概念回看

在章节后半段整理测试概念地图：

- test file / test function
- `testing.T`
- Arrange / Act / Assert
- table-driven test
- subtest
- fake / spy / stub / mock
- regression test
- coverage / benchmark / fuzzing / examples

概念图只能作为回看，不放在文章开头。

#### 12. 工程 checklist

结尾提供判断清单：

- 这个测试保护的是用户可观察行为，还是脆弱实现细节？
- 成功、边界、失败路径是否至少各有代表性 case？
- 失败信息是否包含 case 名、got/want 和业务语境？
- 外部依赖是否通过接口替身隔离，而不是调用真实服务？
- 表驱动测试是否提升可读性，而不是把逻辑藏进复杂表格？
- coverage 数字是否服务风险判断，而不是替代质量判断？

### ch11 审计 rubric

- **Concept**：是否讲清 `*_test.go`、`TestXxx`、`testing.T`、`go test`、`t.Fatal` / `t.Error`。
- **Progression**：是否从最小测试进入表驱动、subtest、fake/spy/stub。
- **Practice**：练习是否覆盖成功、边界、失败和回归。
- **Source**：是否吸收 Learn Go with Tests 的行为驱动讲法，并用官方文档校准 API。
- **Backend relevance**：是否围绕后端业务规则、外部依赖、错误传播、重构保护，而不是工具命令清单。

### ch11 本轮不做

- 不完整重写 MDX。
- 不全量替换练习。
- 不新增测试运行型 sandbox。
- 不深挖 coverage/fuzzing 实战。
- 不建立跨章节项目测试线。

## 执行顺序建议

后续实现计划建议拆为：

1. 建立 Workflow Ledger Active 任务，记录课程样板升级范围和恢复点。
2. 审计当前 `ch07`/`ch11` MDX 与 `courseChapters.ts` metadata/exercises。
3. 读取本地 `gopl-zh.github.com` 对应章节，并用 Go 官方文档和优秀教程校准现代行为。
4. 改造 `ch07` MDX 正文。
5. 更新 `ch07` metadata 与 warmup/core/challenge 练习。
6. 将 `ch11` 审计结果和改造蓝图落实为后续计划或文档化条目。
7. 运行验证：`npm run build --prefix web`、`git diff --check`、ch07 starter code 抽查。
8. 若 ch07 质量通过，将标准推广到后续章节迭代。

## 验证策略

设计阶段：

- 设计文档通过 spec review。
- 用户确认设计边界。
- 在用户确认前不进入 ch07/ch11 正文实现。

实现阶段：

- `npm run build --prefix web`
- `git diff --check`
- ch07 starter code 抽查：warmup/core/challenge 可运行、输出稳定。
- 内容质量人工检查：ch07 是否满足 rubric，ch11 蓝图是否足以指导下一轮。

## 后续推广方式

- `ch07` 完成后作为“概念 + 工程抽象类章节”模板，可推广到 `ch05 Functions`、`ch06 Methods`、`ch08 Goroutines/Channels` 等。
- `ch11` 落地后作为“验证 + 实践反馈类章节”模板，可推广到调试、并发、工具链和后端实战课程。
- 两类模板合起来支撑 Go 基础 13 章质量升级，也可迁移到后续后端工程课程。
