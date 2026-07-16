# 能力节点与证据纵向切片设计

- 日期：2026-07-12
- 状态：独立规格评审通过，待用户审阅
- 范围：M1-01、M1-03、M1-07、M1-09 第一条端到端纵向切片
- 关联需求：
  - `docs/requirements/requirements-gogopher-arch-product-positioning.md`
  - `docs/requirements/requirements-gogopher-arch-m1-m2-capability-graph.md`
  - `docs/requirements/requirements-gogopher-arch-miniflux-training-pack.md`

## 1. 背景

当前 GoGopher Arch 已具备静态课程、静态任务、浏览器代码编辑器、Gateway 和 Go Sandbox。当前链路适合演示和单文件练习，但不能形成可信能力证据：

- 课程与任务定义保存在前端 TypeScript 中，章节顺序就是主要学习顺序。
- 草稿保存在浏览器 `localStorage`，刷新或换设备后没有服务端学习记录。
- 练习是否通过由客户端比较 stdout，无法作为可信毕业证据。
- Gateway 只透传原始执行请求，不知道能力、活动、尝试或提示使用情况。
- Sandbox 只创建一个 `main.go` 并执行 `go run`，不能验证 module、多文件包、测试、vet 或 held-out 用例。
- PostgreSQL 只有课程任务需要的样例业务表，没有学习者、尝试、证据和维护队列。
- Dashboard 显示的是静态或本地派生进度，不是账号级真实能力状态。

本设计用一个小而完整的纵向切片验证新的产品核心：用户完成任务后，系统保存可审计尝试，由服务端验证产生不可变证据，派生当前能力状态，并安排后续维护任务。

## 2. 目标

### 2.1 用户目标

用户能够围绕一个多文件 Go 配置处理程序，完成以下闭环：

1. 理解任务对应的四个能力节点。
2. 在多文件 workspace 中编辑 Go 代码和测试。
3. 运行、测试并根据反馈修正实现。
4. 使用分级提示时留下帮助记录。
5. 最终提交后由服务端运行可见和 held-out 验证。
6. 查看本次尝试产生的能力证据，而不是只看到“输出匹配”。
7. 在后续变式任务中重新验证相同能力。

### 2.2 产品目标

- 验证仓库内版本化定义与服务端运行状态分离是否可维护。
- 验证 Attempt、Execution、Evidence、CapabilitySnapshot 和 ReviewItem 的最小闭环。
- 验证一个活动能同时为多个能力节点产生不同证据。
- 验证独立完成、使用提示和 AI 协作可以被区分。
- 验证能力状态能够由历史证据重新计算，而不是手工写死。
- 为后续 M1 图谱、`gocheck`、M2 和 Miniflux 训练提供稳定扩展边界。

### 2.3 成功标准

- 新学习者能够开始任务、保存多文件草稿、运行测试并最终提交。
- 服务端能够重放该学习者的尝试过程和帮助使用记录。
- 客户端无法通过修改本地通过状态直接生成 Evidence。
- 相同 Attempt 重复提交不会产生重复证据。
- 更新活动定义后，已有 Attempt 仍引用原始版本并可以解释当时的验收规则。
- 四个能力节点能够显示获得、独立、迁移和保持状态。
- 独立验证通过后能够生成后续 ReviewItem；到期本身不直接把能力标记为生疏。
- 关键定义、API、证据投影和多文件执行均有自动化测试。

## 3. 非目标

本规格明确不包含：

- M1/M2 完整能力图谱页面。
- 正式账号、登录、付费、组织或导师系统。
- AI 自动评分、AI 对话导师或自然语言解释的最终掌握判定。
- 完整自适应间隔算法或机器学习推荐。
- Miniflux 仓库检出、patch 管理和开源贡献工作流。
- 完整 `gocheck` 或 `gocheck-hub` 毕业项目。
- 通用课程 CMS、可视化内容编辑器或管理员后台。
- 多语言代码执行。
- 面向公网的不可信代码执行。
- 具有防作弊、监考或对外认证效力的高风险能力证书。

## 4. 已确认方案

采用“版本化定义 + 服务端证据”的混合架构。

### 4.1 定义层

能力、前置关系、活动、任务、资源和掌握门槛以版本化 JSON 文件保存在仓库。JSON 是第一版的语言中立交换格式，使用 JSON Schema 和 CI 校验，不引入课程 CMS 或自定义编译器。

课程正文继续使用 MDX。定义文件只保存正文引用、活动编排、评估规则和版本信息，不复制大段正文。

### 4.2 运行层

PostgreSQL 保存学习者、尝试、运行、帮助事件、产物、证据、能力快照和维护任务。运行记录引用定义的 `id`、`version` 和 `content_hash`。

### 4.3 信任边界

- 浏览器负责编辑体验、草稿缓存和展示，不决定任务是否通过。
- Gateway 负责加载定义、创建 Attempt、选择允许的执行动作、记录帮助事件和触发评估。
- Sandbox 只执行 Gateway 生成的内部 ExecutionSpec，不接受浏览器提供的任意 shell 命令。
- Evidence 只能由服务端 Evaluator 根据已持久化 Execution 和规则生成。
- CapabilitySnapshot 和 ReviewItem 是 Evidence 的派生投影，可以重建。
- 第一切片的“独立完成”只表示平台观察到的帮助使用情况。系统无法识别用户在外部设备或工具中使用 AI，不提供防作弊或认证承诺。

### 4.4 Breaking-change 实施原则

第一条切片是新产品定位的替代实现，不承担旧 API、旧数据库、旧前端路由、旧页面结构或旧 Sandbox 协议的兼容义务。实施时可以直接删除、迁移或替换现有实现，以当前规格的领域边界、可信证据和用户闭环为唯一产品契约。

这项授权不放宽定义不可变性、Attempt/Evidence 可重放性、匿名会话所有权和本地 Sandbox 安全限制。这里的“不兼容”针对重设计之前的原型；新系统一旦写入 release、Attempt 或 Evidence，仍必须遵守本规格的历史事实与版本回放规则。

## 5. 第一条学习切片

### 5.1 覆盖能力

| 能力 | 本切片证据 |
|---|---|
| M1-01 运行模型、工具链与 module | 初始化和理解 module；运行 build、test、vet；解释命令失败信号 |
| M1-03 函数、错误契约、defer 与资源清理 | 返回带上下文的错误；保留错误链；正确关闭文件或响应资源 |
| M1-07 文件、流、JSON 与 CLI 契约 | 从文件读取 JSON；校验输入；稳定输出；使用非零退出码表达失败 |
| M1-09 自动化测试与测试设计 | 表格驱动测试；临时目录；成功和失败用例；held-out 边界验证 |

### 5.2 切片任务

任务代号：`check-config-normalizer-v1`。

用户实现一个配置规范化命令：读取检查目标 JSON，验证字段并输出稳定的规范化 JSON。

初始 workspace：

```text
go.mod
cmd/checkcfg/main.go
internal/config/config.go
internal/config/config_test.go
testdata/valid-targets.json
README.md
```

任务要求：

- `Config` 包含多个检查目标，每个目标至少包含名称、URL 和超时。
- 读取函数接收文件路径并返回结构化配置和错误。
- 文件打开、读取、JSON 解码和校验错误需要增加操作上下文。
- 对需要暴露给调用方判断的错误保留 error chain；不能无条件把底层实现错误变成公开 API。
- URL 仅允许 `http` 和 `https`，名称不能为空，超时必须大于零。
- 输出按目标名称稳定排序，避免测试和自动化结果随机变化。
- CLI 使用 `-config` 指定文件；成功输出规范化 JSON，失败写 stderr 并返回非零退出码。
- 用户至少新增一个成功用例和两个失败用例。
- 测试使用 `t.TempDir` 或等价隔离，不读取开发者机器上的固定路径。

### 5.3 活动梯度

1. `guided-run-model`：观察 workspace、运行 build/test 并解释反馈。
2. `practice-error-contract`：修复一个丢失错误链或资源关闭的实现。
3. `practice-json-io`：补全 JSON 读取、校验和稳定输出。
4. `practice-table-tests`：补全并扩展表格驱动测试。
5. `assessment-check-config`：在限制提示的 workspace 中完成综合提交。
6. `review-check-config-variant`：到期后使用字段和错误场景不同的变式任务复验。

## 6. 定义模型

定义文件位于 `content/learning/`，第一版目录约定：

```text
content/learning/
├── schemas/
│   ├── capability.schema.json
│   ├── activity.schema.json
│   └── task.schema.json
├── capabilities/m1/
│   ├── m1-01.json
│   ├── m1-03.json
│   ├── m1-07.json
│   └── m1-09.json
├── activities/m1-first-slice/
│   ├── guided-run-model.json
│   ├── practice-error-contract.json
│   ├── practice-json-io.json
│   ├── practice-table-tests.json
│   ├── assessment-check-config.json
│   └── review-check-config-variant.json
├── tasks/
│   ├── guided-run-model-v1/
│   ├── practice-error-contract-v1/
│   ├── practice-json-io-v1/
│   ├── practice-table-tests-v1/
│   ├── assessment-check-config-v1/
│   └── review-check-config-variant-v1/
└── releases/
    └── m1-first-slice-v1/
        ├── manifest.json
        └── bundle/
            ├── capabilities/
            ├── activities/
            └── tasks/
```

六个 Activity 各自引用一个明确版本的 TaskDefinition。它们可以通过文件内容哈希复用相同 starter 资产，但第一版不支持 Activity 对 Task 做隐式覆盖。引导、分项练习、综合 assessment 和 review variant 的 starter、可编辑范围与测试集合都由各自 TaskDefinition 完整声明。

评估用例称为 held-out tests：它们不通过 API 或前端产物提供，但项目本身是开源的，因此不把它们当作防作弊秘密。它们只用于减少“针对页面示例编码”，不能支撑高风险认证。生产前端构建产物不得包含 held-out test 文件。

### 6.1 发布包与不可变哈希

作者编辑 capabilities、activities 和 tasks 草稿目录。每次内容发布把解析后的完整定义和所有引用资产复制到新的 `releases/<release_id>/bundle/`，并生成不可变 ReleaseManifest。Gateway 只加载 release bundle，不直接加载草稿目录。

ReleaseManifest 包含：

- `release_id`、创建时间和 schema 版本。
- 每个 Capability、Activity、Task 的 `id`、`version`、规范化定义哈希。
- Task 引用的 starter、可见测试、held-out tests、README 和 fixture 的相对路径与 SHA-256。
- 每个 Activity 的 `rule_set_hash`，覆盖 Activity evidence rules、Capability refs、Task assessment rules 和所有评估资产哈希。
- 按相对路径排序后的完整 bundle hash。

JSON 使用 RFC 8785 JSON Canonicalization Scheme 产生规范化字节，再计算 SHA-256。文件资产直接对原始字节计算 SHA-256。Task 的最终 `bundle_hash` 对“规范化 task JSON + 按路径排序的所有引用文件路径和哈希”再次计算 SHA-256，因此修改任何 starter 或测试都必须提升 Task 版本并创建新 release。

Gateway 运行时使用只读 `LEARNING_CONTENT_DIR` 加载 release。Docker 构建显式把该目录复制到 Gateway 最终镜像；Web 的构建上下文仍只包含 `web/`，并增加断言确保评估资产没有进入前端产物。

PostgreSQL 新增 `definition_releases` 与 `definition_versions`：首次发布时记录 release manifest、每个定义版本和 bundle hash。相同 `kind + id + version` 已存在但哈希不同时，发布和服务启动均失败。CI 同时校验当前文件与提交的 ReleaseManifest；数据库记录是已经发布版本的最终历史基准。

旧 release 目录和数据库版本记录不得原地修改或删除。Gateway 启动时检查数据库中仍被 Attempt 引用的 release bundle 均存在且哈希匹配，缺失时拒绝启用 Learning API。Docker 最终镜像复制全部仍被支持的 release 目录。已有 Attempt 通过 `release_id` 从归档 bundle 加载原始 JSON、starter、fixture 和测试资产，真正回放当时规则，而不是只比较哈希。

### 6.2 CapabilityDefinition

必填字段：

- `id`：稳定能力 ID，例如 `M1-03`。
- `version`：正整数定义版本。
- `content_hash`：由 release 生成，不由作者手填。
- `name`、`description`、`milestone`、`domain`。
- `prerequisites.hard` 与 `prerequisites.recommended`。
- `required_evidence`：证据类型、独立性、上下文和复验要求。
- `review_policy`：首次复验、成功后间隔和失败后动作的启发式策略。
- `resource_refs`：指向资源登记表或 MDX 内容。
- `supersedes`：可选旧版本引用。第一切片不自动合并不同版本 Evidence；未来若允许迁移，必须由新版本显式列出兼容的旧证据规则。

### 6.3 ActivityDefinition

必填字段：

- `id`、`version`、`title`、`kind`。
- `capability_refs`：每项包含 Capability `id + version`，release 解析后附加哈希。
- `mode`：`guided`、`practice`、`assessment` 或 `review`。
- `content_ref` 或包含 `id + version` 的 `task_ref`，release 解析后附加 bundle hash。
- `assistance_policy`：是否允许资料、提示或 AI。
- `evidence_rules`：哪些结果可以产生何种证据。

### 6.4 TaskDefinition

必填字段：

- `id`、`version`、`language`。
- `starter_files`、`editable_paths`、`readonly_paths` 和文件数量/大小上限。
- `actions`：允许的 `build`、`test`、`vet`、`submit` 动作。
- 每个动作的内部命令模板、超时、输出上限和网络策略。
- 可见测试与 held-out tests 路径。
- `assessment_rules`：每条规则包含稳定 `rule_id`、执行阶段、结构化结果选择器、关联 Capability 版本、Evidence 类型和通过条件。
- `artifact_rules`：保存完整文件还是 diff、保存哪些日志。

TaskDefinition 必须显式声明其全部文件资产；运行时发现未登记文件或哈希不一致时拒绝执行。

## 7. 运行时数据模型

### 7.1 Learner

第一版使用匿名学习者与匿名会话：

- `POST /api/v1/learning/session` 创建 Learner 和一个高熵随机 session token。
- 服务端只保存 token hash、learner ID、创建时间、过期时间和最后使用时间。
- 原始 token 通过 `HttpOnly`、`SameSite=Lax` cookie 返回；本地 HTTP 环境不设置 `Secure`，未来 HTTPS 环境必须设置。
- 其余 Learning API 都从 cookie 解析 Learner，不接受客户端直接传 learner UUID。
- cookie 缺失、过期或无效返回 `401`；访问不属于该 Learner 的 Attempt 或 ReviewItem 统一返回 `404`。
- cookie 丢失会创建新的匿名学习者，第一切片不提供恢复、合并或跨设备同步。

匿名 session 只解决本地 Attempt 所有权隔离，不等同正式账号认证。

第一切片只支持 Web 与 Gateway 同源访问：开发环境由 Vite 把 `/api` 代理到 Gateway，打包环境由同源反向代理转发 `/api`。Web 请求显式使用 `credentials: "same-origin"`，session cookie 的 path 覆盖 `/api/v1/learning`。不支持浏览器直接跨域调用 Learning API，也不为该场景开放带凭据 CORS；未来若拆分域名，需要另行设计 cookie domain、`Secure`、CSRF 和允许来源策略。

### 7.2 Attempt

一次活动的工作会话。核心字段：

- `id`、`learner_id`。
- `release_id`。
- `activity_id`、`activity_version`、`activity_hash`。
- `task_id`、`task_version`、`task_hash`。
- 本次活动引用的 Capability `id + version + hash` 列表。
- `mode`、`status`。
- `workspace`：小型文件映射或其持久化引用。
- `workspace_revision` 与 `workspace_hash`。
- 通过 `attempt_review_items` 关联的零个或多个 ReviewItem。
- `started_at`、`submitted_at`、`completed_at`。

状态机：

```text
active → submitted → completed
   │         └────→ submit_infra_failed → submitted（显式重跑）

submitted、submit_infra_failed、completed 不允许继续修改 workspace。
```

最终提交创建独立 Submission 记录。Submission 保存 `submission_key`、冻结的 workspace revision/hash、`rule_set_hash`、状态和当前 Execution ID。每个 Attempt 只允许一个最终 Submission；第一次成功冻结者获胜。

Submission 还保存 `assistance_cutoff_seq`。每个 Attempt 的 AssistanceEvent 使用在 Attempt 行锁内递增的序号；提交事务在同一把 Attempt 行锁下把状态从 active 改为 submitted，并把当时最后一个事件序号冻结为 cutoff。Evaluator 只读取 `event_seq <= assistance_cutoff_seq` 的事件，因此提交与提示请求无论谁先获得锁，都有唯一且可重放的 independence 结果。

Submission 状态机：

```text
frozen → executing → evaluated
             └────→ infra_failed → executing（显式 retry）
```

用户代码或测试失败仍进入 evaluated；只有基础设施失败进入 infra_failed。

### 7.3 Execution

每次 build、test、vet 或 submit 都产生一条 Execution：

- `id`、`attempt_id`、可选 `submission_id`、`action`、`sequence`。
- `request_key`：客户端普通动作或服务端 submit 重跑的幂等键。
- `request_fingerprint`：对 action、workspace revision/hash 和定义动作参数计算的哈希。
- 使用的定义版本和 workspace hash。
- 各阶段 exit code、stdout、stderr、持续时间和截断标记。
- `status`：`queued`、`running`、`succeeded`、`user_failed`、`infra_failed`。
- 可见测试结果、held-out test 汇总、RuleResult 列表和 Sandbox 状态。
- worker lease、开始时间、结束时间和允许的重跑次数。
- `created_at`。

held-out test 名称和失败细节只返回 TaskDefinition 允许公开的摘要，不把测试源码发送给浏览器。

普通 build/test/vet 使用 `attempt_id + request_key` 唯一约束，响应丢失后以相同 fingerprint 重试同 key 返回原 Execution；同 key 但 action、revision 或 hash 不同返回 `409 idempotency_conflict`。`infra_failed` 不自动重跑，客户端可以为普通动作发送新 request key。最终 submit 的 HTTP 重试始终返回同一 Submission，不重新执行。只有 `POST /submissions/{id}/retry` 可以在同一冻结 workspace 上创建新的 submit Execution。

Execution 状态机：

```text
queued → running → succeeded
              ├→ user_failed
              └→ infra_failed
```

只有持有有效 lease 的 worker 可以完成 running Execution；终态不可原地重开。worker 被关闭或 lease 丢失时不新增终态，任务在 lease 过期后回到可领取的 queued 语义。用户代码超过 Task action timeout 记为 `user_failed` RuleResult；Gateway-Sandbox RPC deadline、worker lease 或服务进程故障记为 `infra_failed`。

三层时限必须满足严格顺序：`Gateway RPC deadline > Task action timeout + Sandbox 清理/结果序列化余量`，worker lease 必须覆盖完整 RPC deadline 和结果落库余量，并在长任务期间持续续租。Gateway 启动时校验所有 Task action timeout 与运行配置，不满足顺序就拒绝启用 Learning API。这样正常到达 action timeout 的用户进程必定先由 Sandbox 返回结构化 timeout 结果；只有超过该结果返回上界或 worker/服务失联才归类为 `infra_failed`。

### 7.4 AssistanceEvent

记录 `hint_revealed`、`reference_opened`、`solution_viewed`、`ai_declared` 等事件。每个事件包含客户端生成的 `event_key` 并以 `attempt_id + event_key` 去重，同时保存 Attempt 内单调递增的 `event_seq`。事件写入事务锁定 Attempt 行，只允许 active Attempt 写入；提示或解答内容必须在事件提交成功后才返回。Submission 已冻结时，新的帮助请求返回 `409 attempt_already_submitted`，不能改变该 Submission 的 independence。

平台只能记录自身提供的帮助和用户主动声明的 AI 使用，不能检测外部帮助。最终独立性标签表示“平台观察到的独立性”，不是防作弊结论。

### 7.5 Artifact

保存最终 workspace、关键 diff、用户解释和测试报告。第一版设置严格大小上限并存入 PostgreSQL JSONB；超限提交返回 `422`，本规格不实现对象存储。

### 7.6 Evidence

Evidence 是追加写入的历史事实：

- `learner_id`、`capability_id`、`capability_version`。
- `capability_hash`。
- `attempt_id`、`activity_id`、`artifact_id`。
- `evaluation_batch_id`、`evidence_rule_id`。
- `evidence_type`：第一切片只使用 `implement`、`test`、`diagnose`。
- `result`：第一切片按单条 evidence rule 记录 `failed` 或 `passed`。
- `independence`：`guided`、`hinted`、`referenced`、`ai_assisted`、`independent`。
- `context_level`：第一切片只使用 `same_context`、`variant`。
- `evaluator`：第一切片固定为 `deterministic`。
- `occurred_at`。
- `rule_version` 和生成原因。

Evidence 不直接更新或删除。任务被撤销或评估规则存在错误时，追加 `superseded` 关系并重新投影，不改写历史。

文字解释保存为 Artifact，但第一切片不自动评分、不创建正式 Evidence，也不参与投影。

一个最终 Submission 对应一个 EvaluationBatch。数据库唯一约束为 `submission_id + rule_set_hash`；单条 Evidence 唯一约束为 `evaluation_batch_id + capability_id + capability_version + evidence_rule_id + evidence_type`。EvidenceBatch、全部 Evidence、Submission evaluated 标记和投影 outbox 在同一事务中写入，避免部分成功。

### 7.7 CapabilitySnapshot

按 `learner_id + capability_id + capability_version` 保存可重建投影。不同 Capability 版本的 Evidence 第一切片不自动合并：

- `acquisition_state`：`not_started`、`exploring`、`practiced`、`verified`、`stable`。
- `independence_state`：`unverified`、`guided`、`ai_assisted`、`hinted`、`referenced`、`independent`。
- `transfer_state`：`unverified`、`same_context`、`variant`、`new_project`。
- `retention_base_state`：持久化 `fresh` 或 `rusty`；API 对外的 `retention_state` 额外包含派生的 `due`。
- `last_evidence_at`、`last_independent_at`、派生的 `next_review_at`。
- `projection_version`。

`next_review_at` 取该版本最早的 active ReviewItem `due_at`。API 读取时传入明确的 `as_of` 时间：`retention_base_state=fresh` 且存在到期 active ReviewItem 时，对外派生 `due`，不需要定时任务改写 Snapshot。只有关联 review 的失败 Evidence 才把持久化 base state 变为 `rusty`。

`independence_state` 表示历史上已通过证据达到的最高平台观察自治级别，固定排序为：`guided < ai_assisted < hinted < referenced < independent`。较低自治的新尝试不会抹掉较高自治的历史证据；UI 另行展示最近一次 Evidence 的实际 independence。

### 7.8 ReviewItem

保存待执行的维护任务：

- `learner_id`、`capability_id`。
- `capability_version`、`source_evidence_id`。
- `release_id`、`activity_id`、`activity_version`、`activity_hash`。
- `review_group_key`、`due_at`、`priority`、`reason`。
- `status`：`open`、`claimed`、`completed`、`replaced`；`open` 且 `due_at <= as_of` 时对外视为 available。
- 可选 `claimed_attempt_id` 和完成该项的 `evaluation_batch_id`。
- 创建该任务的 Evidence 和策略版本。

第一版调度使用节点定义中的透明启发式规则，不使用机器学习：assessment 首次独立通过后 3 天安排 variant review；variant review 独立通过后 14 天安排下一次维护；review 失败时完成当前 ReviewItem，状态派生为 rusty，并在 1 天内安排针对性 practice。测试环境可以覆盖时钟和间隔，但生产定义必须保存这些策略值。

同一 Capability 版本和策略只允许一个未完成 ReviewItem。一次多能力 assessment 为各 Capability 创建独立 ReviewItem，但共享 `review_group_key` 和精确 Activity/release 版本。

从任一组内 ReviewItem 创建 review Attempt 时，事务查找同 learner、group key、Activity 版本且已到期的全部 open 项，把它们共同标记 claimed，并写入 `attempt_review_items` 关联表。重复领取组内任一项返回同一个 Attempt。这样一次四能力 variant review 可以完成四个 ReviewItem，而不是要求用户重复运行相同任务。

review 评估完成后，EvidenceBatch outbox 按关联 Capability 分别处理：

- 某 Capability 的 review rule 通过：完成其原 ReviewItem，base state 变 fresh，创建 14 天后的常规 ReviewItem。
- 某 Capability 的 review rule 失败：完成其原 ReviewItem，base state 变 rusty，创建 1 天内到期的 remediation practice ReviewItem。
- remediation practice 通过：完成 remediation，保持 rusty，创建 3 天后的 variant review；只有该 variant review 通过才恢复 fresh。
- remediation practice 失败：完成当前项，并按相同策略创建 1 天后的新 remediation；每次都是新 ReviewItem，历史不覆盖。
- 某 Capability 的规则因更早阶段失败而 `not_evaluated`：完成原 ReviewItem，但不改变该 Capability 的 retention base state，也不生成 Evidence；创建 1 天内到期、同类型同 Activity 版本的新 ReviewItem，`reason=review_incomplete`。其他已经得到 passed/failed RuleResult 的 Capability 仍分别正常流转。

### 7.9 第一切片数据库对象

第一切片明确创建：

- `definition_releases`、`definition_versions`。
- `learners`、`learner_sessions`。
- `learning_attempts`、`attempt_submissions`、`attempt_executions`、`assistance_events`、`artifacts`。
- `evaluation_batches`、`evidence_records`、`capability_snapshots`。
- `review_items`、`attempt_review_items`、`learning_outbox`。

不创建课程 CMS、账号权限、对象存储或通用事件总线表。

## 8. 服务边界

### 8.1 DefinitionRegistry

职责：加载、校验和提供不可变定义。

接口语义：

- 按 `release_id + id + version` 获取 Capability、Activity 和 Task。
- 获取当前发布版本。
- 返回规范化 `content_hash`。
- 启动时校验 ReleaseManifest、引用、前置环、文件路径、资产哈希和数据库历史版本不可变性。

DefinitionRegistry 不读取用户状态，不执行任务。

### 8.2 AttemptService

职责：创建 Attempt、保存 workspace、记录 AssistanceEvent 和处理普通 build/test/vet 请求幂等。最终冻结与提交由 SubmissionWorkflow 独占。

AttemptService 不判断能力是否掌握。

### 8.3 ExecutionService

职责：把受信任 TaskDefinition 和 workspace 转为内部 ExecutionSpec，调用 Sandbox 并保存结构化结果。

ExecutionService 不接受客户端命令字符串，不生成 Evidence。

### 8.4 SubmissionWorkflow

职责：作为最终提交的唯一编排者，原子冻结 Attempt workspace、创建 Submission 和 queued Execution；执行完成后根据状态调用 EvidenceEvaluator，并推进 Submission/Attempt 状态。

编排规则：

1. 数据库事务中锁定 Attempt，验证 owner、active 状态、workspace revision/hash 和 submission key。
2. 第一次有效提交冻结 workspace，并原子创建唯一 Submission 与 queued submit Execution。
3. 事务提交后由 worker 执行，不持有长事务。
4. `user_failed` 仍进入确定性评估：已执行规则按规则级产生 `failed` 或 `passed` Evidence，未执行规则保留 `not_evaluated` RuleResult 但不生成 Evidence；`infra_failed` 不产生 EvaluationBatch 或 Evidence，Attempt 进入 submit_infra_failed。
5. retry 复用同一 Submission 和冻结 workspace，只创建新 Execution。
6. EvaluationBatch 成功写入后 Attempt 进入 completed。

HTTP handler 只调用 SubmissionWorkflow，不自行组合冻结、执行、评估或投影逻辑。

### 8.5 EvidenceEvaluator

职责：读取冻结 Attempt、Execution、Submission cutoff 内的 AssistanceEvent 和 Activity evidence rules，追加 Evidence。EvaluationBatch 同时保存全部 RuleResult，包括不会生成 Evidence 的 `not_evaluated`，供复习调度确定性处理。

第一版只让结构化 RuleResult 直接产生确定性 Evidence。用户文字解释保存为未审 Artifact，不创建 Evidence，也不自动满足独立毕业门槛。

### 8.6 CapabilityProjector

职责：从 Evidence 重建 CapabilitySnapshot，并在状态变化时写入幂等调度 outbox；不直接在投影事务中调用 ReviewScheduler。

### 8.7 ReviewScheduler

职责：根据节点 review policy、最新投影和 EvidenceBatch outbox 创建、领取、替换或完成 ReviewItem。它不能删除历史 Evidence。

### 8.8 RuleResult 映射

Sandbox 以 `go test -json` 等结构化输出为基础，ExecutionService 生成稳定 RuleResult：

- `rule_id`、`status`、`stage`。
- 匹配的 package、测试名或静态检查器。
- 公开摘要和原始 Execution 引用。

TaskDefinition 的每条 assessment rule 显式把 RuleResult 映射到一个或多个 Capability 版本和 Evidence 类型。例如：

- `module-builds` → M1-01 `implement`。
- `error-chain-preserved` → M1-03 `implement`。
- `invalid-input-rejected` → M1-07 `implement`。
- `learner-tests-present` → M1-09 `test`。

`learner-tests-present` 使用冻结 starter 与提交 workspace 的 diff 加 Go AST 检查：至少修改一个 `_test.go` 文件，存在表格驱动结构且包含至少三个命名 case；随后这些测试必须在普通 `go test` 中通过。该规则只作为第一切片的确定性学习证据，不宣称能够防止刻意规避。

## 9. API 设计

所有新 API 位于 `/api/v1/learning`。

浏览器始终通过当前页面同源的 `/api/v1/learning` 访问这些接口并携带 `credentials: "same-origin"`。第一切片不支持跨域直连 Gateway。

### 9.1 定义和状态

- `POST /session`：创建匿名 Learner/session，设置 HttpOnly cookie；已有有效 cookie 时幂等返回当前 Learner。
- `GET /capabilities/{id}`：返回当前发布定义和学习者快照。
- `GET /activities/{id}`：返回允许暴露给客户端的活动和任务信息，不包含 held-out tests。
- `GET /next`：返回第一切片中的下一活动和到期 ReviewItem。

除创建 session 外，所有接口要求有效匿名 session cookie。缺少或无效 cookie 返回 `401`。

### 9.2 Attempt

- `POST /attempts`：根据 `activity_id + activity_version` 创建 Attempt 和初始 workspace。
- `POST /review-items/{id}/attempts`：领取 ReviewItem 并创建或返回其唯一 Attempt。
- `GET /attempts/{id}`：返回 Attempt、workspace、公开 Execution 和 Evidence 摘要。
- `PUT /attempts/{id}/workspace`：提交 `base_revision + files`，在 active 状态保存完整受限文件映射；成功返回新 revision/hash，旧 revision 返回 `409`。
- `POST /attempts/{id}/hints/{hintID}/reveal`：仅 active Attempt 可调用；锁定 Attempt、记录 AssistanceEvent 后再返回对应提示。
- `POST /attempts/{id}/assistance-events`：仅 active Attempt 可调用；以 event key 记录 reference、solution 或 AI 声明，不能伪装为平台可以检测外部帮助。
- `POST /attempts/{id}/execute`：提交 `request_key + action + workspace_revision + workspace_hash`；动作必须存在于 TaskDefinition，同 key 重试返回原 Execution。
- `POST /attempts/{id}/submit`：提交 `submission_key + workspace_revision + workspace_hash`，冻结 workspace 并异步运行 submit。
- `POST /submissions/{id}/retry`：只在最近 submit Execution 为 infra_failed 时，对相同冻结 workspace 创建新 Execution。

### 9.3 响应原则

- 版本冲突返回 `409`，响应包含当前 Attempt 或定义版本信息。
- 非 active Attempt 的 workspace 修改返回 `409`。
- 非法路径、超限文件或未允许动作返回 `422`。
- Sandbox 超时和基础设施失败返回结构化 Execution，不能生成 passed Evidence。
- 同一 submission key、revision 和 hash 的 HTTP 重试返回第一次 Submission，不重复执行或生成证据。
- 同一 submission key 携带不同 revision/hash 返回 `409 idempotency_conflict`。
- 两个不同 key 并发提交时，数据库锁下第一个有效冻结者成功；另一个返回 `409 attempt_already_submitted` 和现有 Submission ID。
- 不属于当前 Learner 的 Attempt、Submission 或 ReviewItem 统一返回 `404`。

## 10. Sandbox 多文件执行契约

### 10.1 内部 ExecutionSpec

Gateway 生成并发送给 Sandbox：

- `execution_id`。
- 文件映射及每个文件的只读/可编辑来源。
- 由 TaskDefinition 解析出的动作枚举和参数。
- 可见或 held-out test 文件及其资产哈希。
- 工作目录、超时、最大输出、文件数量和总大小。
- 网络策略，第一切片默认 `none`。

浏览器不能传递命令、环境变量、挂载路径或 held-out tests。

### 10.2 支持动作

第一版仅支持：

- `build` → `go build ./...`
- `test` → `go test ./...`
- `vet` → `go vet ./...`
- `submit` → 先运行 visible tests，再在受信任构建目录加入 held-out tests；对 TaskDefinition 声明的每个被评估 package 使用 `go test -c` 生成测试二进制，随后只把二进制和允许的 runtime fixture 放入干净运行目录，并通过 `go tool test2json` 执行和采集结构化事件

第一切片不运行 race detector。race 进入 M1-12 的后续切片，避免扩大当前镜像、超时和验收范围。

held-out tests 不进入用户编辑 workspace，评估运行目录也不包含其源码。由于项目开源且用户代码与评估二进制仍在同一进程信任域，这只降低意外窥视，不能抵抗恶意逆向或支撑认证防作弊。

### 10.3 安全约束

现有 Sandbox 基于宿主 `os/exec`，不能安全承载公网不可信代码，也不能真正强制 `network=none`、运行时只读文件、CPU 或内存隔离。ExecutionSpec 中这些字段在第一切片属于策略声明和未来兼容接口，不得在产品文案中声称已经形成安全边界。

第一切片启用时必须同时满足本地模式硬约束：

- `LEARNING_SLICE_ENABLED=true` 且 `APP_ENV=local`，否则 Gateway 启动失败。
- 宿主原生运行 Gateway、Vite 或反向代理时，各进程只监听 `127.0.0.1`。
- Compose 容器内的 Gateway 和 Web/反向代理监听容器网络接口，以便容器间转发；Web 是唯一必须发布的宿主入口，并使用 `127.0.0.1:<host_port>:<container_port>`。Gateway 默认不发布宿主端口；开发时确需发布也只能使用同样的宿主侧回环绑定。
- 第一切片不接入其他公网或局域网 ingress。安全约束针对宿主发布面，不要求容器内进程监听容器 loopback。
- Compose 中 Sandbox 不发布宿主端口，只允许 Gateway 通过内部网络访问。
- Sandbox 拒绝客户端环境变量、命令字符串和绝对路径；每次执行都从数据库冻结 workspace 重建临时目录，用户进程的文件修改不回写 Attempt。
- 任务不需要网络才能完成；测试不会把“声明 network none”等同实际断网。

这些约束不能替代容器/微虚机隔离。生产或公网启用前必须完成独立安全设计与验证。

多文件协议必须预先防止：

- `../`、绝对路径、符号链接和路径覆盖。
- 超出白名单的文件修改。
- 任意命令、环境变量和 shell 插值。
- 超量文件、输出和运行时间。
- 把未强制的网络策略误报为已经隔离。

生产公开运行必须先完成独立的 Sandbox 隔离设计和安全验收；该安全强化不是本规格的实现范围，但属于部署硬门槛。

## 11. Evidence 判定规则

### 11.1 结果

Evidence 结果按 `evidence_rule_id` 单独计算，不把整个 Activity 压成一个结果：

- RuleResult 通过时，为该规则映射的 Capability 生成 `passed` Evidence。
- RuleResult 被执行但未满足条件时，生成 `failed` Evidence。
- 由于前置构建失败而无法执行的规则标记 `not_evaluated`，不生成 Evidence；Execution 仍为 `user_failed`。
- 同一 Submission 可以同时为 M1-01 生成 passed、为 M1-03 生成 failed，并让其他规则 not_evaluated。
- Capability 只有在其 `required_evidence` 中全部必需规则存在符合 independence/context 要求的 passed Evidence 时才能进入 verified。

第一切片不使用 `partial` Evidence；Activity 的 UI 可以根据 RuleResult 汇总显示“部分规则通过”，但该汇总不是能力事实。

### 11.2 独立性

按一次 Attempt 中使用过的最强帮助降级：

```text
solution_viewed → guided
ai_declared     → ai_assisted
hint_revealed   → hinted
reference_opened→ referenced
无已记录帮助    → independent
```

同一 Attempt 出现多类事件时采用上述从上到下的第一项。引导活动始终产生 `guided` Evidence；assessment 或 review 活动才可能产生 `independent` Evidence。这个标签只描述平台记录，不证明用户没有使用外部工具。

### 11.3 能力投影门槛

第一切片使用以下透明规则：

- 完成引导活动：`acquisition_state=exploring`。
- 练习活动通过：最高到 `practiced`。
- assessment 的确定性规则独立通过：相应能力到 `verified`。
- 至少在首次独立 assessment 3 天后完成 variant review：`transfer_state=variant`，能力进入 `stable`，当前 ReviewItem 完成并创建 14 天后的下一项。
- 只使用 AI 完成的 assessment 不提升 `independence_state=independent`。
- 存在 `due_at <= as_of` 的 active ReviewItem：读取时派生 `retention_state=due`。
- review 失败：完成当前 ReviewItem，保留既有历史并追加失败 Evidence，`retention_state=rusty`，创建 1 天内到期的针对性练习。
- review 成功：`retention_state=fresh`，下一次时间只来自新 ReviewItem；Evidence 本身不保存过期时间。

节点可以要求不同证据组合；投影器不得把所有节点写死为同一规则。

## 12. 前端体验

第一切片可以重建现有学习区的信息架构、视觉语言和交互组件，不要求复用旧课程页面、Dashboard、CodeMirror 包装组件或 shadcn 组合方式。前端以“下一活动 → 多文件实作 → 服务端反馈 → Evidence/Snapshot”闭环为主线；是否复用资产只由实现质量和维护成本决定。

工作台最小区域：

- 能力节点、活动模式和本次证据目标。
- 多文件列表与编辑器。
- Build、Test、Vet、Submit 操作。
- 分级提示；每次展开由服务端记录。
- 结构化编译、测试和 held-out 验证摘要。
- 当前 Attempt 的帮助级别。
- 提交后的 Evidence 和 CapabilitySnapshot 变化。

浏览器 localStorage 仍可作为未同步草稿备份，但服务端 Attempt workspace 是恢复和提交的来源。客户端不得自行显示“已掌握”；只能展示服务端返回的 Evidence 和 Snapshot。

Dashboard 只需要把静态“今日建议”替换为 `/learning/next` 的一个真实入口，不在本规格中实现完整能力图谱。

## 13. 错误处理与边界情况

- 定义 JSON 无效、引用不存在、出现硬前置环、ReleaseManifest 不匹配、资产哈希不一致或数据库中相同版本哈希不同：CI 失败，Gateway 拒绝加载该发布集。
- 用户打开旧活动版本：已有 Attempt 继续使用固定版本；新 Attempt 使用当前发布版本。
- 新 Capability 版本默认不继承旧版本 Snapshot；只有未来显式兼容规则可以迁移历史 Evidence，第一切片不实现自动迁移。
- 两个标签页同时保存：使用 workspace revision，旧 revision 返回 `409` 和服务端当前值。
- 保存失败：客户端保留本地草稿并提示未同步，禁止假装已经服务端保存。
- 用户程序在 Sandbox 已正常接收任务后超过 TaskDefinition 的 action timeout：Execution 标记 `user_failed`，生成 timeout RuleResult；已执行规则参与评估，其他规则为 `not_evaluated`，最终 Submission 不允许按基础设施失败重跑。
- Gateway 到 Sandbox 的 RPC deadline、worker 心跳/lease、进程监督器或 Sandbox 服务本身超时、崩溃或不可达：Execution 标记 `infra_failed`，允许对冻结 Submission 显式重试，不生成 EvaluationBatch 或 Evidence。
- stdout/stderr 超限：截断并记录标记，评估不得依赖被截断内容之外的信息。
- held-out test 失败：返回规则级摘要，不返回源码或具体输入；开源用户仍可能从仓库研究评估资产，系统不宣称防作弊。
- 提交后刷新：GET Attempt 可以恢复冻结 workspace、Execution、Evidence 和状态变化。
- Evidence 投影失败：EvaluationBatch 事务已经写入 outbox，由幂等 worker 重试；不得重复追加 Evidence。
- ReviewItem 重复生成：使用学习者、Capability 版本、来源 Evidence 和策略版本的唯一约束保证幂等。
- 同一 submission key 内容变化、不同 key 并发提交和非 owner 访问按第 9 节确定返回，不允许 handler 自行选择行为。
- 定义被撤回：追加发布状态或 superseding 定义，不删除已有 Attempt 和 Evidence。

## 14. 数据一致性

- SubmissionWorkflow 在一个短事务中锁定 Attempt、校验 revision/hash、冻结 workspace，并创建唯一 Submission、queued Execution 和 outbox job；该事务不调用 Sandbox。
- Sandbox 调用不能持有长数据库事务；先记录 queued Execution，执行后再原子写回结果。
- Execution worker 使用 lease 从 queued 进入 running；进程崩溃后过期 lease 可重领。完成写回只能从 running 进入终态。
- `infra_failed` 保留 Submission 和冻结 workspace，不创建 EvaluationBatch；显式 retry 创建新的 queued Execution。
- EvidenceEvaluator 使用唯一 EvaluationBatch 和单条 Evidence identity 作为幂等边界；一个事务写入完整批次、Evidence、Submission/Attempt 完成状态和投影 outbox。
- CapabilityProjector 按 `learner + capability version + as_of` 从 Evidence 与 active ReviewItem 全量重建，Snapshot 不是事实来源。
- ReviewScheduler 根据 EvaluationBatch outbox 中的 RuleResult 和 Evidence 完成被领取的 ReviewItem 并创建下一项；`not_evaluated` 按 7.8 节创建 `review_incomplete` 替代项。ReviewItem 是调度事实，能力是否 due/rusty 仍由 Evidence、ReviewItem 和 as_of 派生。

### 14.1 唯一约束矩阵

| 对象 | 唯一或并发边界 |
|---|---|
| Anonymous session | 原始随机 token 唯一，数据库只存 token hash |
| Definition version | `kind + id + version` 唯一且 bundle hash 不可变 |
| Workspace save | `attempt_id + workspace_revision` 乐观并发 |
| Normal execution | `attempt_id + request_key` 唯一 |
| Final submission | 每个 Attempt 最多一个；submission key 与冻结 revision/hash 绑定 |
| Submit execution | `submission_id + retry_sequence` 唯一 |
| Evaluation batch | `submission_id + rule_set_hash` 唯一 |
| Evidence | `batch + capability version + rule_id + evidence_type` 唯一 |
| Active review | 每个 learner + capability version + policy 只允许一个未完成项 |
| Claimed review | 每个 ReviewItem 最多一个 claimed Attempt |
| Attempt-review link | `attempt_id + review_item_id` 唯一，且每个 ReviewItem 最多关联一个 Attempt |

## 15. 测试策略

### 15.1 定义层

- JSON Schema 校验。
- ID、版本、哈希不可变性和引用完整性。
- ReleaseManifest 对 starter、visible tests、held-out tests 和 fixture 的完整哈希覆盖。
- 已发布数据库哈希与本地相同 id/version 冲突时拒绝启动。
- 硬前置环检测。
- held-out tests 不进入前端产物的构建断言。

### 15.2 服务层

- Attempt 状态机和 workspace revision 单元测试。
- AssistanceEvent 先写后读、event key 幂等和独立性降级测试。
- AssistanceEvent 与 submit 并发时的 Attempt 行锁、cutoff sequence 和冻结后 `409` 测试。
- 相同 key 响应丢失重试、同 key 不同内容冲突、不同 key 并发提交和最终 workspace hash 测试。
- Submission infra failure 后显式重跑和 user failure 进入评估测试。
- EvaluationBatch 原子写入、重复 Evidence 防护和投影 outbox 重试测试。
- Evidence 规则与 Snapshot 投影表格驱动测试。
- ReviewItem 创建、替换、完成和失败重排测试。
- 多能力 review 中 passed、failed 与 not_evaluated 混合时逐能力流转和 `review_incomplete` 重排测试。

### 15.3 Sandbox

- 多文件 module build/test/vet fixture。
- 可见测试、held-out test 编译和干净运行目录测试。
- 路径穿越、绝对路径、符号链接、超限文件和任意命令拒绝测试。
- 输出截断、用户进程失败、Gateway-Sandbox RPC deadline 和 Sandbox 不可达测试。
- 用户进程 action timeout 与 Sandbox RPC/worker timeout 的分类测试。
- action timeout、Sandbox 结果余量、RPC deadline、worker lease 顺序不合法时的启动拒绝测试。
- 用户代码尝试枚举运行目录时无法读取 held-out test 源码的测试，同时文档明确这不是恶意逆向隔离。

### 15.4 API 与数据库

- PostgreSQL migration 测试。
- Attempt create/save/execute/submit/restore 集成测试。
- definition version conflict 和 workspace revision conflict 测试。
- 基础设施失败不生成 passed Evidence 的测试。
- anonymous session 缺失、过期、伪造、cookie 丢失和跨 Learner 所有权隔离测试。
- 旧 release Attempt 在发布新版本后仍使用原始 bundle 和 Capability 版本的回放测试。
- 宿主原生进程监听回环、Compose 仅把 Web 宿主侧发布到回环、Gateway/Sandbox 不发布宿主端口的配置断言。

### 15.5 端到端

至少覆盖：

1. 匿名学习者开始 `guided-run-model` 并恢复草稿。
2. 用户展开提示，最终 Evidence 被标记为 hinted。
3. 用户独立完成 assessment，服务端 held-out tests 通过并生成 verified Snapshot。
4. 重复提交不重复执行、不重复生成 Evidence。
5. 到期 review 进入 `/learning/next`，失败后变为 rusty 并安排针对性练习。
6. 基础设施失败后对冻结 Submission 显式重跑，最终只产生一个 EvaluationBatch。

## 16. 可观测性

第一版记录：

- Attempt 创建、提交、完成和基础设施失败数量。
- Execution 按 action、status 的数量和耗时。
- 用户 action timeout、Gateway-Sandbox RPC deadline、其他基础设施失败和输出截断数量。
- Evidence 按类型、独立性和结果的数量。
- Projection 和 ReviewScheduler 重试次数。

日志包含 attempt、execution、activity、task 和 learner 的内部 ID，但不能记录完整用户代码、held-out tests、密钥或 session token。

## 17. 迁移策略

本切片采用 breaking migration：旧章节、练习、Mission、单文件 Sandbox API 和静态 Dashboard 不构成兼容边界，可以被删除、替换或暂时下线。课程正文仍可作为内容素材，但不能限制新领域模型和信息架构。

数据库 schema 使用 `db/migrations/` 下单调递增的 SQL 文件和 `schema_migrations` 表。新增 repo-local Go migration command 提供 `up` 与 `status`；Compose 在 Gateway 启动前运行 `up`。migration 必须幂等检测已应用版本，但已应用 SQL 文件不得原地修改。

- 为四个能力节点新增版本化定义。
- `check-config-normalizer-v1` 使用新的 Activity/Task 定义，删除 `GoCourseExercise` 事实模型。
- 仍有价值的 MDX 只能作为 `content_ref` 迁入新定义；未迁入内容不保证继续暴露路由。
- 能力活动工作台直接替代旧 `CourseExercisePanel`、Mission 工作台和单文件执行体验。
- Dashboard 改为服务端 `/learning/next` 驱动的真实入口，不保留静态进度作为产品状态。

切片验证通过后，再制定 13 章内容向能力节点迁移的独立计划。

## 18. 发布与回滚

应用配置未显式提供 `LEARNING_SLICE_ENABLED` 时默认关闭。仓库提供的 `.env.example` 和 local Compose 可以为一键本地体验显式设为 `true`；非 `local` / `test` 环境仍禁止启用。

发布前检查：

- 定义校验、Go 测试、前端构建和数据库 migration 通过。
- Learning API 未开启时只提供明确的“本地学习切片未启用”状态，不承诺回退到旧课程或旧 Sandbox 行为。
- 开启后可以完成端到端任务并重建 Snapshot。
- 宿主原生模式下 Gateway 与 Vite/Web/同源反向代理只监听回环地址；Compose 模式下只有 Web 发布宿主端口，且宿主侧绑定 `127.0.0.1`，Gateway 和 Sandbox 不发布宿主端口。
- 没有其他 ingress 把学习入口暴露到公网或局域网，相关启动检查与 Compose 配置断言通过。
- 产品文案没有把 held-out tests、匿名 session 或裸 `os/exec` 描述成防作弊或生产安全能力。

回滚时关闭功能开关并保留新表和历史 Evidence，不删除学习记录。定义发布可以回退当前版本指针，但旧 Attempt 始终保留其固定版本引用。

## 19. 实施边界建议

本文是纵向切片的总设计，不应展开成一个超大实施计划。批准后必须拆成以下四份彼此独立、依次验收的实施计划：

1. **Plan A — 定义、会话与 Attempt**：release/schema、四个 Capability、六个 Activity/Task、migration、anonymous session、DefinitionRegistry、Attempt workspace 持久化与并发控制。
2. **Plan B — 多文件执行、提交与证据**：固定动作协议、Execution worker、SubmissionWorkflow、RuleResult、EvaluationBatch、Evidence 幂等与本地执行边界验证。
3. **Plan C — 投影与复习调度**：CapabilityProjector、Snapshot 重建、ReviewItem 分组领取、复验与 remediation 状态流转、`/learning/next` 查询。
4. **Plan D — 前端闭环与端到端验证**：能力活动工作台、同源 session 接入、Dashboard 下一活动入口、草稿恢复、Evidence 展示和完整 E2E。

每份计划都必须单独接受工程审查、拥有自己的测试与完成条件，并允许在不启动下一份计划的情况下停下。Plan B 依赖 Plan A 的稳定契约，Plan C 依赖 Plan B 的 Evidence 事实，Plan D 最后集成前三者；后续计划不得借机反向扩大已验收计划的范围。

四份计划可以直接移除现有单文件课程练习、旧 API 和旧页面。Learning 执行仍必须通过功能开关限制在本地环境；公开不可信代码执行需要另立安全设计和实施计划，不能随本切片默认上线。
