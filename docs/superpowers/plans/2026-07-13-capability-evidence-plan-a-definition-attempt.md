# 能力证据纵向切片 Plan A：定义、会话与 Attempt

> 状态：Approved for breaking-change implementation
> 日期：2026-07-13
> 上游规格：`docs/superpowers/specs/2026-07-12-capability-evidence-vertical-slice-design.md`

## 目标

交付第一条能力切片的可信定义入口和可恢复工作区：Gateway 能从不可变 release bundle 加载四个 Capability、六个 Activity/Task，为匿名学习者创建会话和 Attempt，并使用 revision/hash 安全保存多文件 workspace。

本计划完成后，系统还不能执行或提交新任务。Plan B 只能依赖本计划已经验收的定义、数据库和 Attempt 契约，不能在实现执行链路时反向修改这些语义。

## 架构与布局决定

- 保持单一 Go module，但把后端迁到标准根目录布局，不继承 `src/services` 和 `src/pkg/common`。
- 使用模块化单体 Gateway + 独立 Sandbox process：entry point 放在 `cmd/`，业务代码放在根 `internal/`，跨进程协议放在 `api/`。
- 使用手工 constructor injection 和显式 `App` wiring；当前规模不引入 DI framework。
- 新增 `cmd/migrate/main.go` 作为一次性 admin command；业务逻辑放入 `internal/platform/database`，`main` 只解析 `up/status` 并调用它。
- Learning 领域代码位于 `internal/learning/`，按 `definition`、`session`、`attempt`、`httpapi` 拆分。
- 定义草稿和 release bundle 都保存在 `content/learning/`。Gateway 只加载 release，不从草稿目录提供运行时数据。
- 当前裸 `os/exec` 的 Sandbox 不在本计划中改动；Learning API 默认关闭。

## 范围

### 包含

- migration runner、schema history 和 Plan A 所需表。
- JSON Schema、四个 Capability、六个 Activity/Task 及完整资产声明。
- release 构建/校验 command、RFC 8785 canonical JSON 和 SHA-256 manifest。
- DefinitionRegistry 启动校验和数据库版本不可变性校验。
- 匿名 session cookie、Learner 所有权解析。
- Attempt 创建、读取、workspace 保存、revision/hash 并发控制。
- 创建 session、读取 Capability/Activity、创建/读取 Attempt、保存 workspace 的 API。
- feature gate、本地环境限制和 Compose 内容挂载/复制基础。

### 不包含

- build/test/vet/submit 执行、worker lease、Evidence 或 Snapshot。
- ReviewItem、`/learning/next` 和任何前端页面。
- 账号注册、跨设备恢复、跨域 cookie 或生产认证。
- 生产 Sandbox 隔离。

## 交付文件

计划实施时预计新增或修改以下区域；工程审查可以调整具体文件拆分，但不得改变边界语义。

```text
content/learning/
├── schemas/{capability,activity,task,release-manifest}.schema.json
├── capabilities/m1/{m1-01,m1-03,m1-07,m1-09}.json
├── activities/m1-first-slice/*.json
├── tasks/*/{task.json,README.md,starter,tests,testdata}/...
└── releases/m1-first-slice-v1/{manifest.json,bundle/...}

db/migrations/
├── 000001_learning_definition_session_attempt.up.sql
└── 000001_learning_definition_session_attempt.down.sql

cmd/migrate/main.go
cmd/learning-content/main.go
cmd/gateway/main.go
internal/platform/database/{database.go,migrate.go,migrate_test.go}
internal/learning/
├── definition/{model.go,registry.go,release.go,validate.go,*_test.go}
├── session/{model.go,service.go,repository.go,*_test.go}
├── attempt/{model.go,service.go,repository.go,workspace.go,*_test.go}
└── httpapi/{handler.go,session.go,definition.go,attempt.go,*_test.go}
```

## 数据库边界

本计划只创建：

- `schema_migrations`
- `definition_releases`
- `definition_versions`
- `learners`
- `learner_sessions`
- `learning_attempts`

`learning_attempts` 必须包含冻结后续流程所需的稳定字段：release/activity/task/Capability 版本与 hash、mode、status、workspace JSONB、workspace revision/hash、时间戳。Plan A 只允许 `active` 状态，但 schema 预留总规格定义的完整状态枚举。

Plan B/C 的表由各自 migration 新增，不在本计划提前创建空壳表。

## 实施任务

### A1 — 建立 migration 基础

- [x] 为 `internal/platform/database` 写 migration state、checksum 和顺序校验测试。
- [x] 实现 PostgreSQL 连接和 `schema_migrations` bootstrap。
- [x] 实现 `cmd/migrate` 的 `up`、`status`；已应用文件 checksum 变化必须失败。
- [x] 新增 Plan A migration，包含外键、唯一约束、状态约束和必要索引。
- [x] 增加临时 PostgreSQL 集成测试，验证首次执行、重复执行、乱序/篡改拒绝。
- [x] 更新 Compose，在 Gateway 启动前执行 migration；失败时 Gateway 不启动。

完成条件：空数据库可一次建成；重复 `up` 无副作用；修改已应用 SQL 会被明确拒绝。

### A2 — 定义 schema 与第一 release 内容

- [x] 先写 schema 失败用例，覆盖缺字段、非法版本、非法引用、非法路径和未登记资产。
- [x] 实现 Capability、Activity、Task JSON Schema；ReleaseManifest schema 在 A3 随 release contract 实现。
- [x] 写入 M1-01、M1-03、M1-07、M1-09 version 1 定义。
- [x] 写入六个 Activity version 1，并让每个 Activity 显式引用唯一 Task version。
- [x] 为六个 Task 准备 starter、可见测试、held-out tests、README 和 fixture；每个资产必须在 TaskDefinition 中声明。
- [x] 对 `assessment-check-config-v1` 和 `review-check-config-variant-v1` 使用不同字段/错误场景，避免 review 只是重复原题。

完成条件：草稿定义通过 schema 和引用校验；所有 Task 可单独还原完整 workspace，且没有隐式覆盖。

### A3 — 构建不可变 release

- [ ] 为 canonical JSON、文件 hash、task bundle hash、rule set hash 和完整 bundle hash 写 golden tests（canonical/task bundle 已完成，rule set/full bundle 待完成）。
- [ ] 实现 `cmd/learning-content` 的 `validate`、`release`、`verify`，输入草稿目录并输出新 release bundle；禁止覆盖已有 release。
- [ ] 复制解析后的定义及全部资产，生成按路径排序的 manifest。
- [ ] 验证 hard prerequisite 无环、所有引用存在、路径为相对 clean path、资产 hash 匹配。
- [ ] 提交 `m1-first-slice-v1` release，并增加“重新生成无 diff”的确定性检查。
- [ ] 增加前端构建断言，确认 `web/dist` 不包含 held-out test 文件名或内容指纹。

完成条件：相同输入跨两次构建产生相同 hash；任意资产改变都会改变 task/bundle hash 并要求新版本。

### A4 — 实现 DefinitionRegistry

- [ ] 先写加载有效 release 和拒绝损坏 release 的测试。
- [ ] 实现按 `release_id + kind + id + version` 的只读查询。
- [ ] 启动时校验 manifest、引用、schema、资产、旧 release 存在性和运行配置。
- [ ] 在事务中登记 `definition_releases/definition_versions`；相同 `kind + id + version` 的 hash 冲突必须失败。
- [ ] 读取当前 release 指针，但保留按旧 release 回放定义和 starter 的能力。
- [ ] 生成面向客户端的 Activity/Task view，明确剔除 held-out tests、内部命令和资产路径。

完成条件：Gateway 不能在定义不完整、hash 冲突或被 Attempt 引用的旧 release 缺失时启用 Learning API。

### A5 — 匿名 session

- [ ] 为新建、复用、过期、伪造和 cookie 丢失写服务/API 测试。
- [ ] 使用 CSPRNG 生成高熵 token，只持久化 token hash。
- [ ] 设置 `HttpOnly`、`SameSite=Lax` 和 `/api/v1/learning` path；本地 HTTP 不设置 `Secure`。
- [ ] 实现 `POST /api/v1/learning/session` 幂等返回当前 Learner。
- [ ] 实现 session middleware；其余 Learning API 无效 session 返回 `401`。
- [ ] 日志禁止输出原始 token 或 cookie。

完成条件：数据库泄露不直接暴露可用 token；不同 Learner 不能通过 UUID 参数绕过 cookie 所有权。

### A6 — Attempt 和 workspace 并发控制

- [ ] 为 starter 创建、owner 隔离、路径白名单、大小限制、revision 冲突和 hash 稳定性写测试。
- [ ] 实现 `POST /attempts`，只接受 Activity id/version，从固定 release 创建完整 starter workspace。
- [ ] 实现 `GET /attempts/{id}`，非 owner 统一返回 `404`。
- [ ] 实现 `PUT /attempts/{id}/workspace`，接收 `base_revision + files`，只保存完整文件映射。
- [ ] 在单个数据库事务中锁定 Attempt、校验 active/revision、校验路径与限额、更新 workspace 和 hash。
- [ ] 对旧 revision 返回 `409`，附当前 revision/hash，不静默合并。
- [ ] 预留 Attempt 读取 DTO 中的 execution/evidence 摘要字段，但 Plan A 返回空集合并明确 API version。

完成条件：两个并发保存只有一个成功；刷新后可从服务端恢复完全相同的 workspace。

### A7 — 路由、配置和本地启用边界

- [ ] 扩展 config：`APP_ENV`、`LEARNING_SLICE_ENABLED`、`LEARNING_CONTENT_DIR`、session TTL 和 DB pool 参数。
- [ ] 仅在 `LEARNING_SLICE_ENABLED=true && APP_ENV=local` 注册 Learning API；关闭时返回明确的 unavailable 状态，不回退旧 API。
- [ ] 注册 Plan A 路由：`POST /session`、`GET /capabilities/{id}`、`GET /activities/{id}`、`POST /attempts`、`GET /attempts/{id}`、`PUT /attempts/{id}/workspace`；Capability 在 Plan C 前返回空 Snapshot。
- [ ] Gateway 原生本地模式默认监听 `127.0.0.1`；Compose 下使用显式容器监听配置。
- [ ] Docker Gateway 镜像复制所有受支持 release；启动时使用只读内容目录。
- [ ] 调整 Compose 基础发布面：Web 宿主端口绑定 `127.0.0.1`，Gateway/Sandbox 默认不发布宿主端口。
- [ ] 删除旧 Gateway `/api/v1/execute`、`src/services/gateway` 和不再使用的 shared request model；不提供兼容 adapter。

完成条件：feature gate 关闭时返回明确 unavailable 状态；错误环境组合会在启动阶段失败而不是降级运行。

## 验证命令

```bash
go test ./...
go run ./cmd/migrate status
go run ./cmd/migrate up
npm run build --prefix web
docker compose config
git diff --check
```

还需运行 Plan A 专属集成测试：定义确定性发布、PostgreSQL migration、session cookie、跨 Learner 404、workspace 并发保存和旧 release 回放。

## 工程审查重点

- release command 是否真正覆盖 Task 所有资产，而不只 hash JSON。
- migration checksum 与 definition version hash 是否使用不同且清晰的不可变边界。
- handler 是否只处理 HTTP 映射，所有权与并发规则是否集中在 service/repository transaction。
- `internal/learning` 的 package 划分是否足够小而没有抽象空壳。
- Compose 调整是否保留 Web→Gateway→Sandbox 的容器网络可达性。

## 工程审查结论

2026-07-13 已按 breaking-change 决策重新审查，Plan A 可以进入实施：

- 使用标准 `cmd/`、`internal/`、`api/` 根布局和手工 constructor injection，不继承 `src/services` 目录结构。
- migration、content release 均使用独立 `cmd` 入口，`main` 保持最小化。
- Plan A 只创建 definition/session/Attempt 表，后续事实表由 Plan B/C 的新增 migration 管理。
- Plan A API 明确返回空 Snapshot，不临时伪造掌握状态。
- Compose 必须以一次性 migration service 形成启动门；旧 service 名称、端口和容器契约都可以直接替换。

审查未发现需要修改总规格的阻塞项。实现中若具体依赖库要求改变 canonical JSON、migration checksum 或 transaction 边界，必须重新进入工程审查。

## 停止条件

出现以下任一情况时停在 Plan A，不启动 Plan B：

- 无法确定性重建 release 或 manifest 不能覆盖全部评估资产。
- 旧 release 无法从 Attempt 固定引用回放。
- session/Attempt 所有权测试或 workspace 并发测试不稳定。
- 为继续实现而需要改变总规格的信任边界、Evidence 语义或 API 幂等约定。

## 验收结果

Plan A 只有在全部测试通过、工程审查批准、ledger 记录验证结果后才算完成。完成时应提供一条可演示流程：创建 session → 查看 Activity → 创建 Attempt → 修改多文件 workspace → 刷新并恢复；不得把这条流程描述为已经可评估或已掌握能力。
