# 能力证据纵向切片 Plan D：前端闭环与端到端验证

> 状态：Complete — D1–D10 accepted
> 日期：2026-07-13
> 上游：Plan A–C 已验收的 Learning API 和状态语义

## 目标

交付围绕能力证据闭环重新设计的前端：匿名同源 session、多文件编辑、服务端草稿、固定动作、提示记录、提交结果、Evidence/Snapshot 变化和真实下一活动入口，并用端到端测试验证完整纵向切片。

## 体验边界

- 重新设计学习区信息架构、视觉系统和组件边界；旧 `LearningLayout`、shadcn 组合、CodeMirror 包装和课程页面只在证明更优时复用。
- 新页面只展示服务端 Evidence/Snapshot；客户端不能自行判断“已掌握”。
- localStorage 只是未同步草稿备份，服务端 workspace 是恢复和提交事实来源。
- 前端不接触 held-out test、内部命令、session token 或 learner UUID 所有权参数。
- 删除或替换 `CourseExercisePanel`、Mission 页面、静态 Dashboard 和旧 sandbox 快捷运行，不提供旧路由兼容层。

## 预计文件区域

```text
web/src/api/learning/{client,types,session,activities,attempts}.ts
web/src/app/pages/CapabilityActivity.tsx
web/src/app/components/learning/
├── ActivityHeader.tsx
├── WorkspaceExplorer.tsx
├── MultiFileEditor.tsx
├── ActionBar.tsx
├── ExecutionPanel.tsx
├── AssistancePanel.tsx
└── EvidenceSummary.tsx
web/src/app/hooks/{useLearningSession,useAttemptWorkspace}.ts
web/src/app/routes.tsx
web/src/app/pages/Dashboard.tsx
web/e2e/learning-slice.spec.ts
```

## 实施任务

### D1 — 前端测试基础与 Learning client

- [x] 增加 Vitest、Testing Library、MSW 和 Playwright；配置 unit/component/e2e scripts。
- [x] 定义与服务端 JSON DTO 一致的 TypeScript 类型，不复制服务端内部 model。
- [x] 建立 `/api/v1/learning` client，所有请求使用 `credentials: "same-origin"`。
- [x] 保留 HTTP status、domain error code 和 conflict payload，不把所有错误压成 message string。
- [x] 为 401 session bootstrap、404 owner isolation、409 revision/idempotency conflict 和 422 validation 写 client tests。

完成条件：UI 可以根据 domain error 做恢复选择，而不是解析错误文案。

### D2 — 匿名 session bootstrap

- [x] 在 Learning route 进入时调用 `POST /learning/session`，已有 cookie 时复用。
- [x] session 建立失败显示可重试错误，不降级成伪本地进度。
- [x] 不把原始 token、learner ID 或 cookie 写 localStorage。
- [x] 刷新页面验证 HttpOnly cookie 自动携带，Attempt 所有权保持。
- [x] 增加 cookie 丢失后得到新 Learner、旧 Attempt 返回 404 的测试和用户提示。

完成条件：浏览器代码从未读取 token，所有权完全由同源 cookie 建立。

### D3 — Activity route 与工作台骨架

- [x] 新增 `/learning/activities/:activityId` route，读取 Activity 公开定义。
- [x] 展示能力节点、活动 mode、本次证据目标、assistance policy 和 task README 摘要。
- [x] 开始活动时创建或恢复 Attempt；刷新直接 GET Attempt。
- [x] 按 loading/active/submitted/infra_failed/completed 渲染明确状态。
- [x] 只在 feature gate 开启且 API 可用时显示入口。

完成条件：从 Activity 定义进入 Attempt，不使用前端硬编码 starter 或验收规则。

### D4 — 多文件 workspace 与草稿恢复

- [x] 实现文件树和单文件 CodeMirror 编辑器，只允许选择 Task 暴露的 editable path。
- [x] readonly 文件可查看但不可编辑；held-out 文件完全不存在于客户端数据。
- [x] 保存完整 file map + base revision，并在成功后更新 revision/hash。
- [x] localStorage 以 Attempt/revision 为 key 保存未同步 backup；服务端保存成功后清理对应 backup。
- [x] 409 时同时展示服务端版本和本地未同步草稿，让用户显式选择重新载入或覆盖式再编辑，不自动 merge。
- [x] 为刷新恢复、保存失败、双标签 revision conflict 和超限文件写 component tests。

完成条件：任何“已保存”提示都对应服务端新 revision；刷新不会丢失已确认保存内容。

### D5 — Build/Test/Vet 操作和 Execution 反馈

- [x] ActionBar 只渲染 Activity 公开允许的 action。
- [x] 每次动作创建 request key，并随 workspace revision/hash 提交；响应丢失重试复用同 key。
- [x] 展示 queued/running/succeeded/user_failed/infra_failed 和轮询恢复。
- [x] 分开展示编译、visible test、vet、held-out 摘要、RuleResult 和 truncation。
- [x] timeout 文案区分用户代码超时与基础设施失败；不能把 infra failure 显示为任务失败。
- [x] 页面刷新后从 GET Attempt 恢复公开 Execution history。

完成条件：UI 状态来自服务端 Execution，不根据 stdout/exit code自行重算规则。

### D6 — Assistance 与独立性提示

- [x] 提示分级展示；点击 reveal 先请求服务端记录，再显示正文。
- [x] reference/solution/AI 声明使用 event key，重复点击不重复降级。
- [x] 显示当前平台观察到的 assistance level，并明确无法检测外部帮助。
- [x] Submit 后禁用所有帮助动作；409 时刷新 Attempt 状态。
- [x] component tests 覆盖记录失败时不泄露提示正文、重复事件和 submit 并发。

完成条件：前端不会在 AssistanceEvent 持久化前展示受控内容。

### D7 — Submit、retry、Evidence 和 Snapshot

- [x] Submit 前确保最新 workspace 已保存，使用稳定 submission key/revision/hash。
- [x] 双击或响应丢失重试复用 key；409 展示已有 Submission 并进入恢复轮询。
- [x] user_failed 展示规则级结果；infra_failed 提供对冻结 Submission 的显式 retry。
- [x] completed 后展示 Evidence result、type、independence、context 和对应 Capability。
- [x] 展示 Snapshot 前后变化以及“平台观察证据”说明，不使用防作弊/认证文案。
- [x] review Attempt 显示每个 Capability item 的 passed/failed/not_evaluated 后续安排。

完成条件：重复提交/刷新不产生视觉上的第二次 Submission，Evidence 数量与服务端一致。

### D8 — Dashboard 真实下一活动入口

- [x] 用 `/learning/next` 替换“今日建议”中的一个静态条目。
- [x] 区分首次学习、到期 review、claimed review 和暂无建议。
- [x] 链接到 Activity 或已有 Attempt；review 入口先调用领取 API。
- [x] 删除静态进度和演示建议；产品状态只来自 Learning API。
- [x] API/feature gate 不可用时隐藏真实入口或显示明确关闭状态，不伪造 fallback progress。

完成条件：Dashboard 至少一个建议来自真实服务端事实，且来源标签准确。

### D9 — E2E 场景

- [x] 建立本地专用 Compose/profile：Web、Gateway、Sandbox、PostgreSQL，只有 Web 绑定宿主 `127.0.0.1`。
- [x] 测试匿名 session → guided activity → 保存/刷新恢复。
- [x] 测试提示 reveal → assessment submit → Evidence independence 降级。
- [x] 测试独立 assessment + held-out pass → verified Snapshot。
- [x] 测试重复 submit 不重复执行/Evidence。
- [x] 使用测试 clock 推进到 review due，失败后变 rusty 并安排 remediation。
- [x] 注入 Sandbox infra failure，显式 retry 后只产生一个 EvaluationBatch。
- [x] 扫描 Web 构建和网络响应，断言不存在 held-out source/content fingerprint。

完成条件：上述纵向 E2E 规格均在全新数据库上可重复通过。

### D10 — 发布、回滚与文档

- [x] 更新 README 本地启用方式、migration、release 内容目录和安全限制。
- [x] 更新 `.env.example`，默认关闭 Learning slice。
- [x] 增加宿主绑定/Compose 端口断言，禁止 Gateway/Sandbox 对外发布。
- [x] 验证 feature flag 关闭时显示明确 unavailable 状态，不回退旧课程、Mission 或 sandbox。
- [x] 编写回滚步骤：关闭 flag，保留表、Evidence 和 release，不删除历史。
- [x] 人工检查所有产品文案，不得声称匿名 session 是认证、held-out 是防作弊、裸 Sandbox 是生产安全隔离。

## 验证

```bash
go test ./...
npm test --prefix web -- --run
npm run build --prefix web
npm run e2e --prefix web
docker compose config
git diff --check
```

额外进行桌面与移动端人工检查：文件树、编辑器、action feedback、提示、Evidence、Dashboard 入口、light/dark/system 主题和刷新恢复。

## 工程审查重点

- localStorage 和服务端 workspace 的优先级是否清晰。
- session cookie 是否只通过同源请求使用。
- UI 是否保留精确的 server status/error code，而不是自行推断。
- held-out 信息是否在 DTO、bundle、source map、日志和错误中均不可见。
- 旧课程练习、Mission 路径和静态状态是否已移除，避免双重信息架构。

## 停止与最终验收

若必须把服务端规则复制到前端才能完成页面，或 E2E 依赖公网暴露 Sandbox，停止发布。

最终验收必须从全新 clone/数据库按 README 启动，完成 assessment 和 review 闭环，再关闭 flag 验证 unavailable 状态。完成后才可规划其余内容迁入能力节点；旧页面可直接删除，不属于迁移成功标准。
