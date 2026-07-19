# GoGopher Arch M1–M2 能力图谱需求

- 日期：2026-07-12
- 状态：产品方向已确认，节点证据模型与内容细节待继续设计
- 适用范围：M1 Go 编程基础、M2 Go 后端开发
- 关联文档：`requirements-gogopher-arch-product-positioning.md`
- 项目训练：`requirements-gogopher-arch-miniflux-training-pack.md`

## 一、设计目标

本图谱服务于具有其他语言基础、了解 Go 语法但缺少独立开发经验的用户。它不把教材章节直接当作能力，也不要求所有人按相同课时线性学习。

图谱需要做到：

- 清楚表达从语法理解到完整程序、再到后端服务的硬前置关系。
- 把“看懂、模仿、独立实现、诊断、迁移、长期保持”分开记录。
- 允许诊断后跳过教学材料，但不能跳过关键独立证据。
- 让现有课程、官方资料、书籍、练习和开源项目共同服务于能力节点。
- 让 Go 基础能力在 M2 项目中反复出现并提高复杂度。

## 二、能力节点模型

每个能力节点至少包含：

- `id`、名称、所属阶段与能力领域。
- 硬前置、推荐前置和后续节点。
- 用户需要理解的概念与边界。
- 引导示范、补全练习、独立手写、故障实验和项目任务。
- 无 AI 独立证据与 AI 协作证据。
- 代码、测试、解释、诊断和迁移验收标准。
- 开放资源、许可证、版本和阅读目标。
- 最近验证时间、薄弱模式和维护任务类型。

关键节点不能仅凭选择题或阅读完成标记为掌握。

## 三、M1 Go 编程基础

### 阶段目标

用户能够从空目录独立完成一个结构清晰、可以测试、构建和运行的 Go 程序，并能够处理错误、并发、超时和资源关闭。

### 节点定义

| ID | 能力节点 | 硬前置 | 关键毕业证据 |
|---|---|---|---|
| M1-01 | 运行模型、工具链与 module | 无 | 从空目录初始化 module，完成 run、build、fmt、vet，并解释 package、main、stdout、stderr 和 exit code |
| M1-02 | 值、类型、控制流与 Go 语义 | M1-01 | 独立实现包含零值、命名类型、转换、Unicode、分支和循环的数据处理功能 |
| M1-03 | 函数、错误契约、defer 与资源清理 | M1-02 | 设计函数返回值，正确包装和判断错误，保证文件或响应体在所有路径关闭 |
| M1-04 | slice、map、数组与内存语义 | M1-02、M1-03 | 修复 alias、扩容、底层数组保留或 map 使用错误，并解释原因 |
| M1-05 | struct、方法、组合与领域建模 | M1-03、M1-04 | 用类型和方法表达小型领域模型，避免 C++ 式继承和无意义 getter/setter |
| M1-06 | interface、动态类型与泛型基础 | M1-05 | 在使用方定义最小接口，完成替身测试；能读写简单泛型函数但不滥用抽象 |
| M1-07 | 文件、流、JSON 与 CLI 契约 | M1-03、M1-04 | 完成配置读取、流式处理、JSON 编解码、命令行参数和稳定退出码 |
| M1-08 | 包、模块、API 与文档设计 | M1-05、M1-06、M1-07 | 把程序拆成多个职责清晰的包，控制导出 API、依赖方向并编写有效 doc comment |
| M1-09 | 自动化测试与测试设计 | M1-03、M1-05、M1-08 | 编写表格驱动测试、子测试、临时目录和失败用例，能够解释测试边界 |
| M1-10 | 调试、静态工具与故障定位 | M1-01、M1-09 | 根据编译错误、失败测试、日志和调试器定位预置缺陷，使用 vet 和基础 profile 信息 |
| M1-11 | goroutine、channel 与协作式并发 | M1-03、M1-04 | 实现有界并发 pipeline，能够关闭 channel、等待任务并避免 goroutine 泄漏 |
| M1-12 | 共享状态、同步与 race detector | M1-11 | 复现并修复 data race，比较 mutex、atomic、channel 所有权的适用边界 |
| M1-13 | context、超时、取消与生命周期 | M1-03、M1-11、M1-12 | 让一组并发任务在超时、上游取消和程序退出时及时释放资源 |
| M1-14 | 完整程序交付 | M1-07、M1-08、M1-09、M1-10、M1-13 | 从空目录完成 `gocheck`，通过独立、迁移和延迟复验 |

### 依赖结构

```text
M1-01 工具链
   ↓
M1-02 核心语义
   ├─────────────→ M1-04 集合与内存 ─→ M1-05 类型建模 ─→ M1-06 接口
   ↓                                      │                 │
M1-03 函数与错误 ─→ M1-07 I/O 与 CLI ────┴────→ M1-08 包与 API
   │                                                    ↓
   │                                               M1-09 测试
   │                                                    ↓
   │                                               M1-10 调试
   ↓
M1-11 并发协作 → M1-12 共享状态 → M1-13 context 与生命周期
   └───────────────────────────────┬─────────────────────┘
                                   ↓
                         M1-14 完整程序交付
```

### M1 主线作品：gocheck

`gocheck` 是一个服务状态检查命令行工具，不依赖公网完成测试。最低功能包括：

- 从 JSON 配置读取多个检查目标。
- 使用命令行参数控制超时、并发度和输出格式。
- 使用 `net/http` 和 `context` 执行检查。
- 使用有界并发并保证任务可取消、可等待、无泄漏。
- 输出稳定的人类可读报告和 JSON 报告。
- 使用退出码区分全部成功、部分失败和配置错误。
- 使用 `httptest`、表格驱动测试、临时目录和 race detector 验证。
- 提供 README、构建命令和示例配置。

学习过程先提供局部示范和补全任务，最终要求用户在不同题材下从空目录实现一个变式工具，形成强独立证据。

## 四、M2 Go 后端开发

### 阶段目标

用户能够独立交付一个包含 HTTP API、PostgreSQL、后台任务、认证、测试和容器化运行方式的后端服务，并能在 Miniflux 中识别和修改对应工程结构。

### 节点定义

| ID | 能力节点 | 硬前置 | 关键毕业证据 |
|---|---|---|---|
| M2-01 | HTTP server、路由、handler 与 middleware | M1-14 | 使用标准库构建服务，解释请求生命周期、超时和 server shutdown |
| M2-02 | API 契约、JSON、校验与错误映射 | M2-01、M1-03 | 设计一致的请求响应、状态码和错误协议，覆盖非法输入与边界测试 |
| M2-03 | 业务建模、包边界与依赖组装 | M2-02、M1-06、M1-08 | 把 transport、use case 和 storage 职责分开，接口由使用方定义且依赖方向清晰 |
| M2-04 | database/sql、连接池与查询资源 | M2-03、M1-13 | 使用 Context 执行查询，正确处理 Rows、Scan、连接池和取消 |
| M2-05 | 关系建模、约束、索引与 migration | M2-04 | 设计 schema 和 migration，验证升级、数据约束和查询计划基础 |
| M2-06 | 事务、并发更新与幂等性 | M2-05 | 在真实竞争场景下保证一致性，解释事务边界、锁和幂等键 |
| M2-07 | 外部 HTTP client、超时与失败边界 | M2-02、M1-13 | 使用受控 client 调用依赖，处理取消、响应体、限流和明确的重试边界 |
| M2-08 | 认证、授权、密钥与输入安全 | M2-02、M2-03 | 实现基础身份认证和资源级授权，避免日志泄密与常见输入风险 |
| M2-09 | 配置、依赖初始化与 graceful shutdown | M2-03、M2-04、M1-13 | 在 main 中显式组装依赖，验证启动失败和信号关闭的资源顺序 |
| M2-10 | 后台任务、调度、背压与恢复 | M2-06、M2-07、M2-09、M1-12 | 实现可停止 worker，处理重复任务、积压、部分失败和进程重启 |
| M2-11 | 缓存与数据一致性基础 | M2-06、M2-10 | 为读路径增加缓存，明确失效、穿透、故障降级和数据真相来源 |
| M2-12 | 日志、指标、健康检查与请求关联 | M2-02、M2-09 | 使用结构化日志和低基数指标定位一次请求，区分 liveness 与 readiness |
| M2-13 | 单元、handler、数据库与集成测试 | M2-02、M2-04、M2-06 | 建立分层测试，使用 fixture、真实 PostgreSQL 和确定性时钟或服务替身 |
| M2-14 | 构建、容器、CI 与依赖安全 | M2-09、M2-13 | 生成最小运行镜像，建立 fmt、vet、test、race、vuln 和 migration 检查 |
| M2-15 | Miniflux 导览与真实修改 | M2-07、M2-10、M2-12、M2-13、M2-14 | 追踪并修改 API 或后台链路，提交可测试、可解释、可回滚的变更 |
| M2-16 | 后端服务完整交付 | M2-08、M2-10、M2-11、M2-12、M2-14、M2-15 | 独立交付 `gocheck-hub`，通过无 AI、AI 协作、迁移和延迟复验 |

### 依赖结构

```text
请求主线：M2-01 HTTP → M2-02 API 契约 → M2-03 业务与包边界
                                           ├→ M2-07 外部依赖
数据主线：                    M2-04 SQL → M2-05 Schema → M2-06 事务
                                           │                    │
运行主线：                         M2-09 生命周期 → M2-10 后台任务 → M2-11 缓存
                                           │                    │
安全主线：                         M2-08 认证授权                │
质量主线：                         M2-12 可观测 → M2-13 测试 → M2-14 交付
                                                                    ↓
                                                        M2-15 Miniflux
                                                                    ↓
                                                     M2-16 完整后端服务
```

### M2 主线作品：gocheck-hub

M2 不另起一个毫无关系的玩具项目，而是把 M1 的检查工具演进为后端服务：

- 管理检查目标、执行计划和检查结果。
- 提供版本化 HTTP API 和稳定错误协议。
- 使用 PostgreSQL 保存配置、任务和结果。
- 使用后台 worker 执行检查，并处理取消、积压和重复任务。
- 提供基础认证、资源授权和审计信息。
- 提供结构化日志、指标、liveness 和 readiness。
- 提供单元、handler、数据库和端到端测试。
- 提供 migration、Dockerfile、Compose 和 CI 检查。

随后用户在 Miniflux 中寻找同类结构，完成一次受控修改。这样自建项目用于从零交付，开源项目用于陌生代码迁移，两种证据互相补充。

## 五、开放资源映射

### 权威主资源

- M1-01 至 M1-02：[Go 官方 Tutorials](https://go.dev/doc/tutorial/)、Create a Module、A Tour of Go 和语言规范。
- M1-03：[Working with Errors](https://go.dev/blog/go1.13-errors) 与标准库错误文档。
- M1-05 至 M1-08：[Effective Go](https://go.dev/doc/effective_go)、[Go Code Review Comments](https://go.dev/wiki/CodeReviewComments)、[Go Doc Comments](https://go.dev/doc/comment)；使用 Effective Go 时必须提示其未覆盖现代 module、testing 和 generics 生态。
- M1-09：标准库 [`testing`](https://pkg.go.dev/testing) 文档、[Go fuzzing](https://go.dev/doc/security/fuzz/) 文档和 Learn Go with Tests 的练习方法。
- M1-11 至 M1-13：Go 官方并发资料、race detector、[`context`](https://go.dev/blog/context) 文档和标准库源码。
- M2-01 至 M2-03：标准库 [`net/http`](https://pkg.go.dev/net/http) 文档与源码，框架只作为后续对照，不先替代 HTTP 心智模型。
- M2-04 至 M2-06：[Go database/sql 指南](https://go.dev/doc/database/)，包括查询、事务、连接池和取消。
- M2-08：[Go 安全最佳实践](https://go.dev/doc/security/best-practices)和 `govulncheck`。
- M2-15：固定 Miniflux v2.3.2 源码、官方开发文档和贡献约束。

### 补充资源

- 本地 gopl-zh：概念解释、经典示例和标准库脉络。
- Go 101：语言语义、类型系统和运行时细节深挖。
- Learn Go with Tests：从行为和测试驱动小步练习。
- 标准库源码：为接口、HTTP、context、testing 和同步原语提供真实实现锚点。

开放资源只提供知识和实例，平台仍需负责前置关系、阅读目标、原创任务、反馈和掌握验证。

## 六、现有 13 章课程迁移

现有章节不直接删除，也不再承担唯一学习顺序。它们被拆分为能力节点的讲解、示例和练习资源。

| 现有章节 | 主要迁移目标 | 调整 |
|---|---|---|
| ch1 入门 | M1-01 | 保留最小程序，补 module、退出码和真实工具链 |
| ch2 程序结构 | M1-01、M1-08 | 作用域保留；包初始化与 API 边界进入工程节点 |
| ch3 基础数据类型 | M1-02 | 保留类型、Unicode、零值和转换 |
| ch4 复合数据类型 | M1-04、M1-05 | 拆分集合内存语义与领域建模 |
| ch5 函数 | M1-03 | 把错误契约、defer 和资源清理提升为核心 |
| ch6 方法 | M1-05 | 与 struct、组合和领域模型合并 |
| ch7 接口 | M1-06 | 加入使用方接口、替身测试和泛型边界 |
| ch8 Goroutines 和 Channels | M1-11、M1-13 | 增加取消、泄漏和有界并发 |
| ch9 共享变量并发 | M1-12、M1-13 | 与 race detector、生命周期和关闭协议结合 |
| ch10 包和工具 | M1-01、M1-08、M1-10 | 拆到学习早期，不再等并发之后才学习 |
| ch11 测试 | M1-09、M2-13 | M1 建立基本测试，M2 深化数据库与集成测试 |
| ch12 反射 | M1 可选深化、M3/M4 | 不作为 M1 毕业硬要求，只在真实需求中学习 |
| ch13 底层编程 | M3/M4 | unsafe 不应作为基础阶段主线，转入运行时和性能专题 |

现有内容明显缺失、需要新增的核心节点包括：M1-03 错误契约、M1-07 真实 I/O 与 CLI、M1-10 调试、M1-13 context 生命周期、M1-14 完整交付，以及完整的 M2 后端链路。

## 七、诊断与跳过规则

- 用户自评“学过”不能跳过节点。
- 能通过独立挑战的节点可以跳过讲解和示范，但保留后续维护任务。
- 能看懂但写不出的节点进入补全到手写的渐隐路线。
- AI 辅助完成只能减少重复讲解，不能替代独立验证。
- M1-14、M2-15、M2-16 是综合门槛，不允许仅凭单点测试免修。

## 八、能力维护规则

- 语法和 API 使用短检索与代码补全维护。
- 错误、I/O、测试和并发使用从空白重建与故障修复维护。
- 包边界、事务和缓存使用方案比较与代码评审维护。
- M1 能力必须在 M2 任务中螺旋复现；M2 能力必须在 Miniflux 和后续 M3 生产任务中复现。
- 独立证据、AI 协作证据和延迟复验证据分别记录。

## 九、第一版内容生产顺序

第一版不按现有 ch1 至 ch13 顺序逐章重写，而按最短交付闭环生产：

1. M1-01、M1-03、M1-07、M1-09：先打通最小完整程序所需能力。
2. M1-02、M1-04、M1-05、M1-08：补齐数据和程序设计。
3. M1-11、M1-12、M1-13：完成并发与生命周期。
4. M1-06、M1-10：已在真实代码中补齐接口、泛型和证据化调试闭环。
5. M1-14：已完成 `gocheck` 独立交付与 `endpointaudit` 延迟变式的可执行内容。
6. M2-01：已完成标准库 HTTP 请求切片、`gocheck-hub` 独立服务评估与 `jobwatch` 延迟变式。
7. M2-02：已完成严格 JSON API 契约练习、`gocheck-hub` 独立评估与 `alert rules` 延迟变式。
8. M2-03：已完成 `gocheck-hub` use case/storage 练习、checks 分层独立评估与 alert 分层延迟变式。
9. M2-04：已完成 SQL pool 练习、checks SQL storage 独立评估与 alert SQL 延迟变式。
10. M2-05、M2-06：下一批继续打通 schema、migration、事务与一致性。
11. M2-07 至 M2-14：补外部依赖、运行、安全、质量和交付。
12. M2-15、M2-16：完成 Miniflux 迁移和 `gocheck-hub` 毕业项目。

## 十、已确认决策与后续细化

- M1 泛型要求能够阅读、解释并完成简单修改，不强制从零设计复杂泛型 API。
- `gocheck` 与 `gocheck-hub` 先作为内容原型代号，试学验证后再决定产品名称。
- M2 必须掌握缓存判断、失效和一致性，但不以必须使用 Redis 作为毕业条件；Redis 作为真实扩展任务。
- 能力证据采用多维标签，不压缩为单一总分；具体标签与判定规则继续设计。

## 十一、当前实现状态

截至 `m1-first-slice-v19`，运行时已经发布 M1-01 至 M1-14、M2-01 至 M2-04 共十八个能力节点。M1-01 已调整为从亲手完成第一个 Go 程序开始的引导、练习、评估、复习闭环；M2 后端阶段已经从标准库 HTTP、稳定 API 契约和业务分层进入 SQL 查询资源管理：

| 能力 | 首次习得 | 独立评估 | 异题复验 |
|---|---|---|---|
| M1-02 | 名称分类与 Unicode 语义引导 | 文本记录分类器 | 温度区间分类 |
| M1-04 | slice 快照与 map 初始化练习 | 事件批次快照 | 字节帧 payload |
| M1-05 | 检查状态模型练习 | 检查目标领域模型 | 预算状态模型 |
| M1-06 | 通知服务最小接口练习 | 交付服务接口与泛型索引 | 审计服务接口与泛型分组 |
| M1-08 | 报告 package 边界练习 | 状态报告 module | 输出格式 module |
| M1-10 | 从失败栈帧定位边界错误 | 报告生成故障诊断 | 指标生成故障诊断 |
| M1-11 | 受控并发映射练习 | 有界 worker pool | 缩略图流水线 |
| M1-12 | 共享计数器修复 | 并发安全注册表与 race detector | 并发账本 |
| M1-13 | 可取消并发映射 | 可取消检查器 | 可取消批量抓取 |
| M1-14 | gocheck 跨 package 切片 | 从空目录交付 gocheck | 从空目录交付 endpointaudit |
| M2-01 | handler、ServeMux 与 request ID 切片 | gocheck-hub HTTP 入口、timeout 与 graceful shutdown | jobwatch admin server 与 trace ID 迁移 |
| M2-02 | 严格 JSON、DTO 与错误包络 | gocheck-hub checks API | alert rules API 迁移 |
| M2-03 | checks use case 与 memory storage | gocheck-hub checks 分层与显式组装 | gocheck-hub alert 分层迁移 |
| M2-04 | SQL connection pool 配置 | gocheck-hub checks SQL storage | gocheck-hub alert SQL storage 迁移 |

每个独立评估只产生单一能力的 Evidence，且有同能力、不同题材的 review Activity；首次验证后可直接进入延迟复习调度。M1-06 通过 AST、隐藏替身和跨类型隐藏测试共同验证使用方最小接口与实用泛型，避免只靠测试偶然通过。M1-08 v2 将 M1-06 加入硬前置，并继续使用可执行的导出 API 消费者测试和 doc comment AST 验收。M1-10 要求把失败测试、breakpoint、Vet 与 `alloc_space` profile 连接成诊断小结，不能只提交修复代码。M1-12 增加只在 Submit 装载的私有 `race_test` 资产与 `race` 评测阶段，普通功能测试通过不能替代真实 `go test -race` 证据；私有源码和事件详情继续由服务端隔离。M1-14 增加受策略约束的新建/删除文件能力，并以 `new_project` 证据上下文要求学习者从空工作区交付完整 module；gocheck 独立评估和 endpointaudit 延迟变式都经过真实 sandbox、隐藏测试、Vet、Race Detector 与证据生成验证。M2-01 从 `httptest` 内存请求切片渐进到真实 loopback listener：隐藏测试分别验证 method-aware route、middleware context 传播、四类 server timeout 和活动请求期间的 `Shutdown` 等待；AST 规则要求学习者补充至少三个命名 case 的表格测试，解释规则要求串起 `ServeMux`、`context`、`timeout` 与 `Shutdown`。M2-02 继续沿用同一项目上下文，但把重点收敛到传输契约：严格拒绝未知字段与多余 JSON，验证 URL 和数值边界，把非法输入、领域冲突与内部故障稳定映射为 400、409 和 500，同时禁止内部错误泄漏；7 天后用不同 DTO、领域错误和 URL 约束的 alert API 复验迁移能力。M2-03 在同一项目中把 transport、use case 和 storage 拆成可替换边界：业务消费方拥有最小 storage 接口，transport 拥有调用业务所需接口，composition root 通过 constructor 选择 concrete 实现；隐藏测试验证每层行为和整链组装，AST 同时验证接口归属、constructor 调用与 learner 表格测试，解释规则要求说明依赖方向。延迟复验把相同原则迁移到 alert 业务切片。

为保持已发布 Attempt 的冻结定义可重放，早期 release 仍保留其原始前置关系和评测定义。`m1-first-slice-v9` 中的 M1-08 v1 未被回写；v10 通过 M1-08 v2 显式加入 M1-06 硬前置；v11 新增 M1-14；v13 以 Activity/Task v2 前向增加综合提交的冷缓存预算至 45 秒；v14 新增 M2-01；v15、v16 前向调整初学入口；v17 新增 M2-02；v18 新增 M2-03；v19 新增 M2-04 的三个 Activity/Task、标准库私有 SQL driver 测试与连接池 AST 规则，整个过程均未回写任何历史 release。下一内容批次进入 M2-05、M2-06 的 schema、migration、事务与一致性主线。
