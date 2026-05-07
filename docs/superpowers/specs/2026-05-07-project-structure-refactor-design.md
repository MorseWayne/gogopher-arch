# GoGopher Arch 项目结构重构设计

**日期**: 2026-05-07
**作者**: Claude Code
**状态**: 草案

---

## 1. 背景与目标

### 1.1 当前问题

GoGopher Arch 当前处于 MVP 阶段，代码量快速增长，但目录结构仍停留在单文件原型状态：

- **Go 后端扁平化严重**: `gateway/main.go` 和 `sandbox-engine/main.go` 各自只有一个文件，handler 逻辑、业务逻辑、HTTP 启动逻辑全部混在一起。随着 API 增加和工程进阶任务扩展，这种模式会很快失控。
- **前端业务数据膨胀**: `web/src/tasks.ts` 已达 1300+ 行，包含 14 个完整任务的数据定义，与 `App.tsx` 平级。`taskFeedback.ts` 混入了类型定义、检查逻辑和反馈生成，职责不单一。
- **缺少 API 封装层**: 前端直接调用 `axios.post('/api/v1/execute', ...)`，没有统一客户端、拦截器或错误处理。
- **基础设施缺失**: 没有 Makefile 统一命令、没有 CI 流水线、没有数据库迁移目录。

### 1.2 重构目标

1. **Go 后端分层拆分**: transport (handler) 与核心逻辑（runner）分离，main.go 只做依赖注入和启动，让代码职责清晰、测试可分层
2. **前端业务域拆分**: 将膨胀的任务数据和检查逻辑按 domain 拆分，建立 API 客户端层
3. **工程化补齐**: 引入 Makefile、GitHub Actions CI、数据库迁移目录预留
4. **为未来扩展预留空间**: 第三阶段（LLM API、RAG、Agent）可以直接在现有分层上扩展

---

## 2. 总体架构

```
gogopher-arch/
├── Makefile                          # 统一命令入口
├── docker-compose.yml
├── go.mod
├── .github/workflows/ci.yml          # GitHub Actions CI
├── docs/superpowers/                 # 统一文档目录
│   ├── specs/
│   └── plans/
├── db/
│   ├── migrations/                   # 数据库迁移文件（预留）
│   └── seed.sql
├── src/
│   ├── pkg/common/                   # 跨服务共享的契约
│   │   ├── models.go                 # SandboxRequest, SandboxResponse
│   │   └── errors.go                 # 自定义错误类型
│   ├── internal/                     # 私有代码，禁止外部导入
│   │   ├── task/                     # 任务域（预留，未来持久化）
│   │   │   ├── service.go
│   │   │   └── store.go
│   │   └── config/                   # 配置管理
│   │       └── config.go             # 环境变量、默认值
│   └── services/
│       ├── gateway/                  # API 网关
│       │   ├── main.go               # 仅负责依赖注入和启动
│       │   ├── Dockerfile
│       │   └── internal/
│       │       ├── handlers/         # HTTP handler（transport 层）
│       │       │   ├── execute.go    # POST /api/v1/execute
│       │       │   └── health.go     # GET /health
│       │       ├── middleware/       # 可复用中间件
│       │       │   ├── cors.go
│       │       │   ├── logging.go
│       │       │   └── recovery.go
│       │       └── routes/           # 路由注册与中间件链组装
│       │           └── routes.go
│       └── sandbox-engine/           # 沙盒执行引擎
│           ├── main.go               # 仅负责依赖注入和启动
│           ├── Dockerfile
│           └── internal/
│               ├── handlers/
│               │   └── execute.go    # POST /execute
│               └── runner/
│                   ├── runner.go     # 从 main.go 提取的执行逻辑
│                   └── module.go     # ensureGoModule 逻辑
├── web/
│   ├── public/
│   ├── src/
│   │   ├── api/                      # API 客户端层
│   │   │   ├── client.ts             # axios 实例、拦截器、baseURL
│   │   │   ├── execute.ts            # 沙盒执行 API 封装
│   │   │   └── types.ts              # API DTO 类型
│   │   ├── domain/                   # 业务核心（领域层）
│   │   │   ├── tasks/
│   │   │   │   ├── data.ts           # 任务数据（原 tasks.ts）
│   │   │   │   ├── checks.ts         # 检查逻辑（原 taskFeedback.ts）
│   │   │   │   ├── types.ts          # 任务领域类型
│   │   │   │   └── index.ts          # 统一导出
│   │   │   └── workbench/
│   │   │       └── types.ts          # 工作台类型（原 types/workbench.ts）
│   │   ├── components/               # UI 组件（保持现有结构）
│   │   │   ├── common/
│   │   │   ├── EditorPanel/
│   │   │   ├── FeedbackPanel/
│   │   │   ├── ResizableSplit/
│   │   │   ├── TaskPanel/
│   │   │   └── TopBar/
│   │   ├── hooks/
│   │   ├── lib/                      # 通用工具函数
│   │   ├── App.tsx                   # 精简为状态管理与组件组合
│   │   └── main.tsx
│   └── index.html
```

### 2.1 设计原则

- **main.go 只做三件事**: 读配置、组装依赖、启动服务器。所有业务逻辑必须下沉到 `internal/` 包。
- **Go `internal/` 包机制**: 利用 Go 编译器对 `internal/` 目录的保护，确保业务逻辑不会被外部服务意外导入，强制通过显式公开的 API 交互。
- **前端 `domain/` 层**: 所有业务数据、领域类型、业务规则集中在这里。`App.tsx` 只负责 UI 状态和组件组合。
- **前后端类型镜像**: `web/src/api/types.ts` 与 `src/pkg/common/models.go` 保持字段一一对应，减少通信时的理解成本。

---

## 3. Go 后端重构设计

### 3.1 Gateway 服务

#### 当前状态

`src/services/gateway/main.go` 共 62 行，包含：
- `getSandboxURL()` 环境变量读取
- `/api/v1/execute` handler（含 CORS、请求转发、错误处理）
- `http.ListenAndServe` 启动

#### 目标状态

**`src/services/gateway/main.go`**

```go
package main

func main() {
    cfg := config.Load()
    r := routes.New(cfg)
    log.Printf("Gateway listening on %s\n", cfg.Port)
    if err := http.ListenAndServe(cfg.Port, r); err != nil {
        log.Fatalf("Failed to start gateway: %v\n", err)
    }
}
```

**`src/services/gateway/internal/handlers/execute.go`**

处理 `POST /api/v1/execute`：
- 解析请求体为 `common.SandboxRequest`
- 转发到 sandbox-engine
- 将响应写回客户端
- 统一错误处理（返回 JSON 格式错误）

**`src/services/gateway/internal/handlers/health.go`**

处理 `GET /health`：
- 返回 `{"status":"ok"}`
- 供 Docker healthcheck 和负载均衡使用

**`src/services/gateway/internal/middleware/cors.go`**

提取当前写死在 handler 里的 CORS 逻辑，封装为可复用的 `CORS()` 中间件。

**`src/services/gateway/internal/middleware/logging.go`**

请求日志中间件，记录：
- 请求方法、路径
- 响应状态码
- 处理耗时
- 输出到 stdout（容器化环境友好）

**`src/services/gateway/internal/middleware/recovery.go`**

panic 恢复中间件：
- 捕获 handler 中的 panic
- 返回 500 Internal Server Error（JSON 格式）
- 将 panic 信息记录到 stderr

**`src/services/gateway/internal/routes/routes.go`**

集中注册路由：
- 应用中间件链：`Recovery → Logging → CORS`
- 注册 `/api/v1/execute` → `handlers.Execute`
- 注册 `/health` → `handlers.Health`

### 3.2 Sandbox Engine 服务

#### 当前状态

`src/services/sandbox-engine/main.go` 共 132 行，包含：
- `GopherRunner` 结构体
- `ensureGoModule()` 外部包判断
- `Run()` 沙盒执行核心逻辑（写临时文件、go run、超时控制）
- `errorResponse()` 辅助函数
- `/execute` HTTP handler

#### 目标状态

**`src/services/sandbox-engine/main.go`**

```go
package main

func main() {
    cfg := config.Load()
    runner := runner.New(cfg)
    h := handlers.New(runner)
    r := routes.New(h)
    log.Printf("Sandbox Engine listening on %s\n", cfg.Port)
    if err := http.ListenAndServe(cfg.Port, r); err != nil {
        log.Fatalf("Failed to start sandbox engine: %v\n", err)
    }
}
```

**`src/services/sandbox-engine/internal/handlers/execute.go`**

处理 `POST /execute`：
- 解析 JSON 请求为 `common.SandboxRequest`
- 调用 `runner.Run(req)`
- 编码 `common.SandboxResponse` 返回
- 输入校验（如 timeout 范围）

**`src/services/sandbox-engine/internal/runner/runner.go`**

从原 main.go 提取的 `GopherRunner`：
- `New(cfg) *Runner`
- `Run(req common.SandboxRequest) common.SandboxResponse`
- 临时目录创建与清理
- `go run` 执行与超时控制（`context.WithTimeout`）
- stdout/stderr 捕获
- 状态判断（success / error / timeout）

**`src/services/sandbox-engine/internal/runner/module.go`**

从原 runner.go 提取的 `ensureGoModule`：
- 判断用户代码是否包含 `github.com/lib/pq` 或 `github.com/redis/go-redis`
- 从 `/app/sandbox-module/` 复制预缓存的 go.mod/go.sum

### 3.3 共享代码

**`src/pkg/common/models.go`**（保持现有内容）

```go
package common

type SandboxRequest struct {
    ID       string `json:"id"`
    Code     string `json:"code"`
    Language string `json:"language"`
    Timeout  int    `json:"timeout"`
}

type SandboxResponse struct {
    ID       string        `json:"id"`
    Stdout   string        `json:"stdout"`
    Stderr   string        `json:"stderr"`
    ExitCode int           `json:"exit_code"`
    Duration time.Duration `json:"duration"`
    Status   string        `json:"status"`
}
```

**`src/pkg/common/errors.go`**（新增）

定义自定义错误类型，供 handler 统一返回：
- `ValidationError` — 请求参数校验失败
- `InternalError` — 内部服务错误
- 提供 `WriteError(w, err)` 辅助函数，统一输出 JSON 错误响应

**`src/internal/config/config.go`**（新增）

```go
package config

type Config struct {
    Port       string // 默认 :8080 (gateway), :8081 (sandbox-engine)
    SandboxURL string // gateway 专用
    DB_URL     string // 两个服务共用
    RedisURL   string // 两个服务共用
}

func Load() Config
```

读取环境变量，提供合理的默认值（兼容当前 docker-compose.yml 和本地混合开发模式）。

---

## 4. 前端重构设计

### 4.1 API 客户端层（新增 `src/api/`）

**`src/api/client.ts`**

创建 axios 实例：
- `baseURL`: 开发环境 `/api/v1`，生产环境同域名
- 请求拦截器：统一 Content-Type
- 响应拦截器：统一错误格式转换
- 导出 `apiClient`

**`src/api/execute.ts`**

```typescript
import { apiClient } from './client';
import type { SandboxRequest, SandboxResponse } from './types';

export async function executeCode(req: SandboxRequest): Promise<SandboxResponse> {
  const { data } = await apiClient.post<SandboxResponse>('/execute', req);
  return data;
}
```

**`src/api/types.ts`**

与后端 `models.go` 字段一一对应的 TypeScript 类型：
- `SandboxRequest`
- `SandboxResponse`

### 4.2 业务域拆分（新增 `src/domain/`）

**`src/domain/tasks/types.ts`**

从原 `tasks.ts` 提取：
```typescript
export interface InternTask {
  id: string;
  day: number;
  title: string;
  track: string;
  // ... 其他字段
}
```

**`src/domain/tasks/data.ts`**

从原 `tasks.ts` 提取的 14 个任务数据：
- 按天定义常量：`export const day0Task`、`export const day1Task` ...
- 汇总数组：`export const internshipTasks: InternTask[] = [day0Task, day1Task, ...]`
- `export const defaultTaskId = day0Task.id`
- `export function findTaskById(id: string): InternTask`

**`src/domain/tasks/checks.ts`**

从原 `taskFeedback.ts` 提取的所有检查逻辑：
- `evaluateTaskChecks()`
- `didPassTask()`
- `evaluateSingleCheck()`
- `didSandboxSucceed()`

**`src/domain/tasks/index.ts`**

统一导出，外部只 import 这个文件：
```typescript
export * from './types';
export * from './data';
export * from './checks';
```

**`src/domain/workbench/types.ts`**

从 `src/types/workbench.ts` 迁移，保持不变。

### 4.3 App.tsx 瘦身

重构后 `App.tsx` 的职责：
1. 管理 UI 状态（`selectedTaskId`、`code`、`output`、`loading`、`error`、`taskResults`、`mobileTab`）
2. 组合子组件（`TopBar`、`TaskProgress`、`TaskPanel`、`EditorPanel`、`FeedbackPanel`、`ResizableSplit`）
3. 调用 `api/execute.ts` 中的 `executeCode()`

所有业务逻辑（任务查找、检查评估）通过 import `domain/tasks/` 消费。

---

## 5. 基础设施与工程化

### 5.1 Makefile（新增）

```makefile
.PHONY: dev test build lint clean

dev:
	docker compose up postgres redis -d
	# 本地混合开发：后台启动 Go 服务，前台启动前端 dev server
	# 停止时运行 `make clean`
	go run ./src/services/sandbox-engine/main.go &
	go run ./src/services/gateway/main.go &
	cd web && npm run dev

test:
	go test ./...
	cd web && npm run test -- --run

build:
	docker compose build

lint:
	go vet ./...
	cd web && npm run lint

clean:
	docker compose down -v
	pkill -f "go run ./src/services" || true
```

### 5.2 GitHub Actions CI（新增 `.github/workflows/ci.yml`）

```yaml
name: CI
on: [push, pull_request]
jobs:
  backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.24' }
      - run: go vet ./...
      - run: go test ./...
  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: '20' }
      - run: cd web && npm ci
      - run: cd web && npm run lint
      - run: cd web && npm run test -- --run
```

### 5.3 文档目录合并

将 `docs/specs/` 和 `docs/plans/` 下的旧文件迁移到 `docs/superpowers/specs/` 和 `docs/superpowers/plans/`，删除旧目录，保持统一入口。

### 5.4 数据库迁移目录（预留）

新增 `db/migrations/`。当前 `seed.sql` 保留在 `db/` 根目录。后续 schema 变更时，迁移文件按 `YYYYMMDD_HHMMSS_description.sql` 命名。

---

## 6. 迁移步骤

### Step 1: 创建新目录

```bash
mkdir -p src/services/gateway/internal/{handlers,middleware,routes}
mkdir -p src/services/sandbox-engine/internal/{handlers,runner}
mkdir -p src/internal/{task,config}
mkdir -p src/pkg/common
mkdir -p web/src/api
mkdir -p web/src/domain/{tasks,workbench}
mkdir -p db/migrations
mkdir -p .github/workflows
```

### Step 2: 迁移共享代码

1. `src/pkg/common/models.go` — 保持不变（注意：`Duration` 字段保持 `time.Duration` 类型，前端 `formatDuration.ts` 中的除法逻辑无需修改）
2. 新增 `src/pkg/common/errors.go`
3. 新增 `src/internal/config/config.go`

### Step 3: 重构 Sandbox Engine

1. 新增 `src/services/sandbox-engine/internal/runner/runner.go` — 从 main.go 提取 `GopherRunner` 和 `Run()`
2. 新增 `src/services/sandbox-engine/internal/runner/module.go` — 从 main.go 提取 `ensureGoModule()`
3. 新增 `src/services/sandbox-engine/internal/handlers/execute.go` — 从 main.go 提取 `/execute` handler
4. 重写 `src/services/sandbox-engine/main.go` — 精简为依赖注入 + 启动

### Step 4: 重构 Gateway

1. 新增 `src/services/gateway/internal/handlers/execute.go` — 提取 `/api/v1/execute` handler
2. 新增 `src/services/gateway/internal/handlers/health.go` — 新增健康检查
3. 新增 `src/services/gateway/internal/middleware/cors.go` — 提取 CORS 逻辑
4. 新增 `src/services/gateway/internal/middleware/logging.go`
5. 新增 `src/services/gateway/internal/middleware/recovery.go`
6. 新增 `src/services/gateway/internal/routes/routes.go` — 路由注册
7. 重写 `src/services/gateway/main.go` — 精简为依赖注入 + 启动

### Step 5: 重构前端

1. 新增 `web/src/api/types.ts`
2. 新增 `web/src/api/client.ts`
3. 新增 `web/src/api/execute.ts`
4. 新增 `web/src/domain/tasks/types.ts` — 从 `tasks.ts` 提取
5. 新增 `web/src/domain/tasks/data.ts` — 从 `tasks.ts` 提取
6. 新增 `web/src/domain/tasks/checks.ts` — 从 `taskFeedback.ts` 提取
7. 新增 `web/src/domain/tasks/index.ts`
8. 新增 `web/src/domain/workbench/types.ts` — 从 `types/workbench.ts` 迁移
9. 重写 `web/src/App.tsx` — 使用新的 domain 和 api 模块
10. 更新测试文件 import 路径：
    - `web/src/App.test.tsx` — 更新 `tasks` 和 `taskFeedback` 的 import 为 `domain/tasks`
    - `web/src/tasks.test.ts` → 重命名为 `web/src/domain/tasks/data.test.ts`，更新 import
    - `web/src/taskFeedback.test.ts` → 重命名为 `web/src/domain/tasks/checks.test.ts`，更新 import
11. 删除旧文件：`web/src/tasks.ts`、`web/src/taskFeedback.ts`、`web/src/types/workbench.ts`

### Step 6: 新增基础设施

1. 新增 `Makefile`
2. 新增 `.github/workflows/ci.yml`
3. 合并文档目录
4. 新增 `db/migrations/`

### Step 7: 验证

```bash
make test        # 运行所有测试
make build       # 构建 Docker 镜像
docker compose up --build  # 端到端验证
```

---

## 7. 新旧文件映射

| 旧文件 | 新文件 | 操作 |
|--------|--------|------|
| `src/services/gateway/main.go` | `src/services/gateway/main.go` | 重写（精简） |
| — | `src/services/gateway/internal/handlers/execute.go` | 新增 |
| — | `src/services/gateway/internal/handlers/health.go` | 新增 |
| — | `src/services/gateway/internal/middleware/cors.go` | 新增 |
| — | `src/services/gateway/internal/middleware/logging.go` | 新增 |
| — | `src/services/gateway/internal/middleware/recovery.go` | 新增 |
| — | `src/services/gateway/internal/routes/routes.go` | 新增 |
| `src/services/sandbox-engine/main.go` | `src/services/sandbox-engine/main.go` | 重写（精简） |
| — | `src/services/sandbox-engine/internal/handlers/execute.go` | 新增 |
| — | `src/services/sandbox-engine/internal/runner/runner.go` | 新增 |
| — | `src/services/sandbox-engine/internal/runner/module.go` | 新增 |
| `src/pkg/common/models.go` | `src/pkg/common/models.go` | 保持 |
| — | `src/pkg/common/errors.go` | 新增 |
| — | `src/internal/config/config.go` | 新增 |
| `web/src/tasks.ts` | `web/src/domain/tasks/data.ts` | 迁移 |
| `web/src/tasks.ts` | `web/src/domain/tasks/types.ts` | 提取 |
| `web/src/taskFeedback.ts` | `web/src/domain/tasks/checks.ts` | 迁移 |
| `web/src/types/workbench.ts` | `web/src/domain/workbench/types.ts` | 迁移 |
| `web/src/App.tsx` | `web/src/App.tsx` | 重写（使用新模块） |
| — | `web/src/api/client.ts` | 新增 |
| — | `web/src/api/execute.ts` | 新增 |
| — | `web/src/api/types.ts` | 新增 |
| — | `web/src/domain/tasks/index.ts` | 新增 |
| — | `Makefile` | 新增 |
| — | `.github/workflows/ci.yml` | 新增 |
| `web/src/App.test.tsx` | `web/src/App.test.tsx` | 更新 import 路径 |
| `web/src/tasks.test.ts` | `web/src/domain/tasks/data.test.ts` | 迁移并更新 import |
| `web/src/taskFeedback.test.ts` | `web/src/domain/tasks/checks.test.ts` | 迁移并更新 import |
| `docs/specs/*` | `docs/superpowers/specs/*` | 迁移 |
| `docs/plans/*` | `docs/superpowers/plans/*` | 迁移 |
| `db/seed.sql` | `db/seed.sql` | 保持 |
| — | `db/migrations/` | 新增（空目录） |

---

## 8. 风险与回滚策略

### 8.1 主要风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 文件移动导致 import 路径断裂 | 编译失败 | 每次移动后立刻运行 `go build` 和 `tsc -b`，分步提交 |
| 前端路径别名配置未同步 | 构建失败 | Vite `resolve.alias` 已配置 `@/`，确保新目录在别名范围内 |
| Docker 构建上下文变化 | 镜像构建失败 | 逐步验证每个 Dockerfile 的 COPY 路径 |
| 测试引用旧路径 | 测试失败 | 同步更新 `*.test.ts` 和 `*_test.go` 中的 import（已在 Step 5 中列为独立子步骤） |
| `make dev` 后台进程泄漏 | 端口占用 / 僵尸进程 | `make clean` 使用 `pkill` 清理；长期方案建议统一用 `docker compose up` 或引入进程管理器 |
| models.go Duration 类型变更 | 前端 duration 显示错误 | 保持 `time.Duration` 不变，不修改序列化行为 |

### 8.2 回滚策略

- 整个重构在一个独立分支 `refactor/project-structure` 上进行
- 每个 Step 作为一个独立的 commit，便于 `git bisect` 定位问题
- 如果重构过程中发现阻塞性问题，可以立即切回 `main` 分支，废弃重构分支

---

## 9. 成功标准

1. `make test` 通过（Go 测试 + 前端 vitest）
2. `make build` 成功构建所有 Docker 镜像
3. `docker compose up --build` 后端、前端、沙盒全部正常运行
4. 前端可以正常选择任务、编辑代码、运行、查看反馈
5. CI 流水线通过
6. `go vet ./...` 无警告，`cd web && npm run lint` 无错误

---

## 10. 后续可扩展项（不在本次重构范围）

- 用户系统与认证（JWT / Session）
- 任务进度持久化到 PostgreSQL
- LLM API 集成（第三阶段）
- 前端路由（React Router）
- 单元测试覆盖率提升
