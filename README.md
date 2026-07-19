# GoGopher Arch

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/Frontend-React%20%2B%20TypeScript-61DAFB?style=flat&logo=react)](https://react.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

GoGopher Arch 是一个以 Capability、Evidence 和 Review 为核心的 Go 学习系统。它把 versioned 学习定义、冻结的 Attempt、多文件执行、服务端评估、能力投影和间隔 review 连成一条可复现的纵向闭环。

## 产品事实模型

| 对象 | 作用 |
| :--- | :--- |
| `Capability` | versioned 能力节点与 prerequisite graph |
| `Activity` / `Task` | 公开学习目标、可用 action、workspace 和评估策略 |
| `Attempt` | 固定 release、定义引用、workspace 与 rule set hash |
| `Submission` | 幂等冻结一次待评估 workspace；基础设施失败可复用原 Submission retry |
| `Evidence` | 记录规则结果、证据类型、independence 与 context |
| `CapabilitySnapshot` | 由服务端 projection 从 Evidence 派生能力状态 |
| `ReviewItem` | 按 Snapshot 与 due time 安排后续练习 |

浏览器不计算掌握状态，也不生成静态 progress。Dashboard 的下一项只来自 `GET /api/v1/learning/next`。
成长路线页的能力定义、Snapshot、保持状态和前置条件只来自 `GET /api/v1/learning/roadmap`，前端只负责分阶段展示，不自行判定掌握。
文字小结会随 Submission 冻结并保存为 Artifact；当前切片不按字数自动生成 Evidence，也不参与能力投影。

## 当前纵向切片

- anonymous same-origin session 建立 Learner ownership，并通过 HttpOnly cookie 恢复；
- immutable content release 固定 Capability、Activity、Task 和文件 hash；
- 多文件 workspace 支持 revision conflict、服务端保存与刷新恢复；
- Build、Test、Vet、hint reveal、Submit、infra retry 均有显式状态；
- held-out 与 race evaluation 只在服务端运行，对外只返回公开摘要和 RuleResult；
- Evidence projection 更新 Snapshot，并生成 acquisition、due review 或 remediation；
- Playwright 在全新 PostgreSQL 上验证 guided → assessment → review 闭环。

当前 `m1-first-slice-v19` 已发布 M1-01 至 M1-14、M2-01 至 M2-04。学习主链从“亲手完成第一个 Go 程序”开始，经练习、独立评估和变式复习后进入类型语义、工程能力、HTTP 服务、稳定 API 契约、显式业务分层与 SQL 查询资源管理。

## 快速开始

要求 Docker 20.10+、Docker Compose v2。应用配置在未提供环境变量时默认关闭 Learning slice；仓库提供的 `.env.example` 和 local Compose 为一键本地体验显式开启它：

```bash
git clone https://github.com/MorseWayne/gogopher-arch.git
cd gogopher-arch
cp .env.example .env
```

如需手动配置，确认 `.env` 包含：

```dotenv
APP_ENV=local
LEARNING_SLICE_ENABLED=true
```

然后启动：

```bash
./scripts/dev.sh docker -d
```

访问 [http://localhost:3000](http://localhost:3000)。Web 通过同源 `/api` 反向代理到 Gateway。

构建会自动把当前 shell 或 `.env` 中的 `HTTP_PROXY`、`HTTPS_PROXY`、`ALL_PROXY`、`NO_PROXY`
传给 Docker build，但不会把这些值固化到最终镜像。Go module、Go build 和 npm 下载使用
BuildKit cache；Web 依赖严格按 `package-lock.json` 通过 `npm ci` 安装。首次构建仍需下载依赖，
后续构建会复用缓存。构建网络默认使用 `host`，使 `127.0.0.1` 形式的宿主机代理在 BuildKit
步骤中仍然可达；无需 loopback 代理的环境可以设置 `BUILD_NETWORK=default` 恢复默认隔离。
需要代理时可在启动前导出标准变量，例如：

```bash
export HTTP_PROXY=http://127.0.0.1:10808
export HTTPS_PROXY=http://127.0.0.1:10808
export ALL_PROXY=socks5://127.0.0.1:10808
export NO_PROXY=localhost,127.0.0.1
./scripts/dev.sh docker -d
```

如果不启用 feature flag，首页仍可访问，但 Dashboard 和直接 Activity route 会显示明确 unavailable 状态，不会回退到旧 Course、Mission 或 Sandbox 页面。

## Compose 拓扑与端口

基础 [docker-compose.yml](docker-compose.yml) 按 production-style topology 运行：

| service | host exposure | internal endpoint |
| :--- | :--- | :--- |
| `web` | `127.0.0.1:3000` | `web:80` |
| `gateway` | none | `gateway:8080` |
| `sandbox-engine` | none | `sandbox-engine:8081` |
| `postgres` | none | `postgres:5432` |
| `migrate` | none | one-shot process |

Gateway、Sandbox 和 PostgreSQL 不应由基础 Compose 发布到 host。以下断言同时检查基础 topology 与 loopback-only development overlay：

```bash
./scripts/check-compose-exposure.sh
```

[docker-compose.dev.yml](docker-compose.dev.yml) 只用于本地热开发，把 Gateway、Sandbox 和 PostgreSQL 映射到 `127.0.0.1`。不要把这个 overlay 当作公网部署配置。

## 本地热开发

只改前端：

```bash
# .env 中先设置 LEARNING_SLICE_ENABLED=true
./scripts/dev.sh backend
./scripts/dev.sh web
```

`backend` 使用 development overlay；Vite 在 [http://localhost:5173](http://localhost:5173) 启动，并把 `/api` 代理到 loopback Gateway。

本地运行 Go service：

```bash
./scripts/dev.sh deps

# 再分别开三个终端
./scripts/dev.sh sandbox
./scripts/dev.sh gateway
./scripts/dev.sh web
```

`gateway` command 会执行 migration，并在 `APP_ENV=local` 下显式启用 Learning。可用 `LOCAL_DATABASE_URL` 覆盖它连接的 loopback PostgreSQL。

## Database migration

完整 Compose 启动时，one-shot `migrate` service 先执行所有 up migration；只有成功后 Gateway 才启动。手动执行和检查：

```bash
docker compose run --rm migrate up
docker compose run --rm migrate status
```

本地 Go process：

```bash
DATABASE_URL=postgres://user:pass@localhost:5432/gogopher?sslmode=disable \
  go run ./cmd/migrate status
```

Migration 只向前运行。发布回滚不得执行 down migration 或删除 Learning table。

## Content release

Draft 与 runtime release 分离：

```text
content/learning/
├── capabilities/          # draft Capability
├── activities/            # draft Activity
├── tasks/                 # draft Task assets
├── schemas/               # release validation schemas
├── current-release.json   # 新 Attempt 使用的 release pointer
└── releases/
    └── <release-id>/
        ├── manifest.json
        └── bundle/        # immutable runtime content
```

验证 draft：

```bash
go run ./cmd/learning-content validate --activity-set m1-first-slice
```

创建新 release 时使用全新 `release-id`，禁止覆盖已有目录：

```bash
go run ./cmd/learning-content release \
  --activity-set m1-first-slice \
  --release-id m1-first-slice-v19 \
  --created-at 2026-07-19T00:00:00Z
```

在更新 `current-release.json` 前验证 manifest、文件 hash 和 frontend bundle：

```bash
npm run build --prefix web
go run ./cmd/learning-content verify \
  --release-dir content/learning/releases/m1-first-slice-v19 \
  --web-dist web/dist
```

部署必须保留所有被历史 Attempt 引用的 release 目录。Pointer 只决定新 Attempt 使用哪个 release，不能改写历史。

## 发布与回滚

发布顺序：

1. 运行 unit/component test、Go test、build、release verify 和 Compose exposure check。
2. 先执行 up migration。
3. 部署 immutable release 和 Web/Gateway/Sandbox image。
4. 更新 `current-release.json` 后重新验证目标 release。
5. 在 `APP_ENV=local` 或 `test` 环境显式设置 `LEARNING_SLICE_ENABLED=true`。

发生应用、执行链或内容问题时：

1. 将 `LEARNING_SLICE_ENABLED=false` 并重新部署 Gateway；Dashboard 会进入明确关闭状态。
2. 保留所有 database table、Attempt、Submission、Evidence、Snapshot、ReviewItem 和 outbox record。
3. 保留当前及历史 `content/learning/releases/<release-id>`；不要修改 immutable bundle。
4. 内容 release 有缺陷时，在 flag 关闭期间把 `current-release.json` 指向已验证的旧 release，仅影响后续新 Attempt。
5. 修复或选择 release 后重新运行 migration status、release verify、smoke test，再开启 flag。

关闭 flag 是可逆的流量止损，不是数据回滚。删除表、Evidence 或 release 会破坏审计链和冻结 Attempt 的可恢复性。

## 验证

```bash
go test ./...
go vet ./...
npm test --prefix web -- --run
npm run build --prefix web
./scripts/check-compose-exposure.sh
go run ./cmd/learning-content verify \
  --release-dir content/learning/releases/m1-first-slice-v19 \
  --web-dist web/dist
npm run e2e:compose --prefix web
git diff --check
```

`e2e:compose` 使用独立 Compose project、临时 PostgreSQL volume、`APP_ENV=test` 和显式 feature flag，结束后自动清理。

## 关键环境变量

默认值见 [.env.example](.env.example)。

| variable | default | purpose |
| :--- | :--- | :--- |
| `APP_ENV` | `local` | Learning 只允许在 `local` / `test` 开启 |
| `LEARNING_SLICE_ENABLED` | 应用默认 `false`；Compose / `.env.example` 为 `true` | 本地 Learning feature gate；生产环境仍禁止启用 |
| `LEARNING_CONTENT_DIR` | `content/learning` | schema、pointer 与 release root |
| `LEARNING_SESSION_TTL` | `720h` | anonymous session lifetime |
| `DATABASE_URL` | Compose internal URL | Gateway / migrate database connection |
| `LOCAL_DATABASE_URL` | loopback URL | `scripts/dev.sh gateway` database connection |
| `SANDBOX_ENDPOINT` | loopback URL | Gateway execution endpoint；Compose 内覆盖为 service URL |
| `SANDBOX_RPC_DEADLINE` | `50s` | execution RPC deadline（覆盖最长 45s 综合提交与响应余量） |
| `EXECUTION_WORKER_LEASE` | `60s` | execution claim lease |
| `PROJECTION_WORKER_MAX_ATTEMPTS` | `5` | projection request failure threshold |
| `WEB_PORT` | `3000` | base Compose 唯一 published application port |
| `GATEWAY_PORT` / `SANDBOX_PORT` / `POSTGRES_PORT` | `8080` / `8081` / `5432` | development overlay loopback ports |

## 安全边界

- anonymous same-origin session 用于本地 Learner ownership continuity，不是账号认证或授权系统。Cookie 丢失会创建新 Learner，旧 Attempt 不会跨 owner 暴露。
- held-out source 在 test binary 生成后、执行用户代码前删除；它减少正常 UI/API 的意外泄漏，但开源内容与同一进程 trust domain 意味着它不是防作弊边界。
- Sandbox 当前只适合本地可信学习环境。响应中的 `network=none` / `policy_only` 是策略声明，不代表网络、CPU、内存或 process 已获得生产级隔离。
- 基础 Compose 只把 Web 绑定到 host loopback。公网部署仍需要独立的 authentication、authorization、TLS、rate limit、secret 管理和 hardened execution isolation。
- 默认 database credential 只适合本地开发；部署前必须替换并通过 secret mechanism 注入。

## License 与 attribution

项目使用 [MIT License](LICENSE)。

仓库中保留的历史 Go course source 仅作为未接入当前 route 的参考资产。《Go 语言圣经中文版》
[gopl-zh/gopl-zh.github.com](https://github.com/gopl-zh/gopl-zh.github.com) 的仓库代码采用
[BSD 3-Clause](https://github.com/gopl-zh/gopl-zh.github.com/blob/master/LICENSE)，正文授权说明为 CC-BY 3.0。
