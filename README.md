# GoGopher Arch: Go 后端实习成长平台

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/Frontend-React%20%2B%20Tailwind-61DAFB?style=flat&logo=react)](https://reactjs.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**GoGopher Arch** 是一个面向 Go 学习者的实战成长平台。它通过虚拟职场任务，帮助用户从 Go 基础入门，成长到具备 Go 后端实习岗位能力，并进一步进阶为掌握 Go 技术栈与 AI 应用工程能力的新型全栈程序员。

---

## 核心定位

GoGopher Arch 不把学习过程设计成单纯的课程目录，而是把知识点放进真实工作任务里：

- 初学者先完成入职前训练营，补齐 Go 基础。
- 准实习生和实习生通过任务卡练习修 Bug、补接口、写测试、读日志和处理评审意见。
- 有经验的 Go 工程师继续进入数据库、缓存、并发、部署、可观测性等工程能力训练。
- 进阶用户可以沿着 RAG、Agent、LLM 应用工程和 AI 产品评测路线，转向 AI 时代的新型全栈开发。

---

## 核心特性

- **Go 基础训练营**：GoGopher Arch 完整内置 Go 1.24+ 基础课程，按 13 章组织语法、数据结构、并发、测试和工程实践，每章配一个可运行的 Go 练习。
- **实习生任务线**：以虚拟 Go 后端团队的入职第一周为主线，围绕任务卡、验收标准和导师反馈推进学习。
- **任务前小课**：每个任务前只讲完成当前任务必须用到的 Go 知识，降低上手挫败感。
- **交互式沙盒**：在浏览器中编写 Go 代码，运行程序或测试，立即看到输出、错误和任务反馈。
- **任务后复盘**：完成任务后沉淀知识点、真实工作场景、常见坑和面试追问。
- **成长路线图**：从 Go 基础、后端实习、工程进阶一路延伸到 RAG、Agent 和 AI 应用工程。

---

## 技术栈

| 模块 | 技术实现 |
| :--- | :--- |
| 后端 | Go 1.22+, Gateway, Sandbox Engine |
| 前端 | React, TypeScript, Tailwind CSS, shadcn/ui |
| 沙盒 | Docker, `os/exec`, 执行超时控制 |
| 反馈 | 编译结果、控制台输出、任务检查、导师提示 |
| AI 路线 | LLM API、RAG、Agent、结构化输出、评测与安全 |

---

## 路线图

### 第零阶段：Go 基础训练营

- [x] 建立 GoGopher Arch 完整内置的 13 章 Go 基础学习路径
- [x] 每章提供课程正文、学习目标、现代 Go 说明、工程实践、常见坑、复盘问题和 sandbox 练习
- [x] 课程总览页和章节详情页接入前端导航

### 第一阶段：Go 后端实习生入职第一周

- [x] 项目定位重构规格确认
- [x] README、设计文档和实施计划统一为新定位
- [x] 前端首屏改为实习生工作台
- [x] Day 0：Go 基础自检和第一次沙盒运行
- [x] Day 1：修复 slice、map 和指针相关 Bug
- [x] Day 2：补全一个 HTTP API handler
- [x] Day 3：增加参数校验和错误处理
- [x] Day 4：编写表驱动测试
- [x] Day 5：修复一个简单并发问题或 context 超时问题

### 第二阶段：Go 工程能力进阶

- [x] 数据库和事务任务
- [x] 缓存和并发任务
- [x] 日志、配置和可观测性任务
- [x] 部署和服务可靠性任务

### 第三阶段：AI 时代全栈工程路线

- [ ] LLM API 调用和 Prompt 设计
- [ ] 结构化输出和工具调用
- [ ] RAG：文档切分、Embedding、向量检索和重排
- [ ] Agent：规划、工具使用、记忆、上下文管理和评估
- [ ] AI 产品的成本控制、安全边界和评测集

---

## 快速开始

```bash
git clone https://github.com/MorseWayne/gogopher-arch.git
cd gogopher-arch
cp .env.example .env # 可选：需要改端口或默认连接串时再执行
./scripts/dev.sh docker
```

前端默认运行在 `http://localhost:3000`，Gateway 默认运行在 `http://localhost:8080`。前端在容器模式下通过 Nginx 将 `/api` 反向代理到 Gateway，本地 Vite 开发模式也会将 `/api` 代理到 `localhost:8080`。

---

## 部署说明

### 环境要求

- [Docker](https://docs.docker.com/get-docker/) 20.10+
- [Docker Compose](https://docs.docker.com/compose/install/) v2+

### 服务架构

本项目通过 Docker Compose 编排 Web、Learning Gateway、versioned multi-file Sandbox 和 PostgreSQL。Web 是唯一默认发布的应用入口；Gateway 与 Sandbox 只在 Compose 内部网络监听。

| 服务 | 说明 | 镜像 / 构建方式 | 默认访问 |
| :--- | :--- | :--- | :--- |
| `web` | React + Tailwind 前端（Nginx 托管，代理 `/api`） | `web/Dockerfile` | `http://localhost:3000` |
| `gateway` | Learning API 与应用 wiring | `cmd/gateway/Dockerfile` | Compose internal `gateway:8080` |
| `sandbox-engine` | versioned multi-file Go runner | `cmd/sandbox/Dockerfile` | Compose internal `sandbox-engine:8081` |
| `postgres` | 数据库 | `postgres:15-alpine` | `localhost:5432` |

### 部署步骤

1. **克隆仓库**

   ```bash
   git clone https://github.com/MorseWayne/gogopher-arch.git
   cd gogopher-arch
   ```

2. **按需覆盖本地默认配置**

   ```bash
   cp .env.example .env
   ```

   如果默认端口没有冲突，可以跳过这一步。Docker Compose 会使用 `.env.example` 中展示的默认值。

3. **启动所有服务**

   ```bash
   # 前台运行（便于查看日志）
   ./scripts/dev.sh docker

   # 后台运行
   ./scripts/dev.sh docker -d
   ```

4. **查看服务状态和健康检查**

   ```bash
   ./scripts/dev.sh status
   curl http://localhost:3000/api/v1/learning/session -X POST
   docker compose exec gateway wget --spider -q http://localhost:8080/health
   docker compose exec sandbox-engine wget --spider -q http://localhost:8081/health
   ```

5. **查看日志**

   ```bash
   # 查看所有服务日志
   ./scripts/dev.sh logs

   # 查看指定服务日志
   ./scripts/dev.sh logs gateway
   ```

6. **停止服务**

   ```bash
   ./scripts/dev.sh docker:down
   ```

### 启动脚本

项目提供 `scripts/dev.sh` 封装常用启动场景，避免反复手动输入环境变量和 `docker compose` 命令。

```bash
./scripts/dev.sh help
```

| 命令 | 适用场景 |
| :--- | :--- |
| `./scripts/dev.sh docker` | 完整 Docker 部署验证，会执行 `docker compose up --build` |
| `./scripts/dev.sh docker:up` | 使用已有镜像启动完整 Docker 环境，不主动重新构建 |
| `./scripts/dev.sh backend` | 用 Docker 启动 Gateway、Sandbox、migration 和 PostgreSQL，适合前端本地热开发 |
| `./scripts/dev.sh deps` | 只启动 PostgreSQL，适合 Go service 本地运行 |
| `./scripts/dev.sh sandbox` | 本地启动 versioned multi-file Sandbox |
| `./scripts/dev.sh gateway` | 执行 migration 并本地启动 Learning Gateway |
| `./scripts/dev.sh web` | 启动本地 Vite 前端开发服务，访问 `http://localhost:5173` |
| `./scripts/dev.sh local` | 启动 PostgreSQL，并提示本地热开发需要开的 3 个终端 |
| `./scripts/dev.sh status` | 查看 Docker Compose 服务状态 |
| `./scripts/dev.sh logs [service]` | 跟随查看所有服务或指定服务日志 |

如果只是修改前端 UI，推荐不要每次重建 Web 镜像：

```bash
./scripts/dev.sh backend
./scripts/dev.sh web
```

如果前后端都要本地热开发，分别开终端运行：

```bash
./scripts/dev.sh deps
./scripts/dev.sh sandbox
./scripts/dev.sh gateway
./scripts/dev.sh web
```

### 本地开发（混合模式）

如果你希望在本地运行上层应用代码（便于断点调试和热重载），同时用 Docker 只启动 PostgreSQL，可按以下步骤操作。

#### 本地开发环境要求

- [Docker](https://docs.docker.com/get-docker/) 20.10+
- [Node.js](https://nodejs.org/) 20+
- [Go](https://go.dev/dl/) 1.25+

#### 1. 启动基础组件

```bash
./scripts/dev.sh deps
```

#### 2. 启动沙盒引擎（Go）

```bash
./scripts/dev.sh sandbox
```

沙盒引擎默认监听 `http://localhost:8081`，健康检查地址为 `http://localhost:8081/health`。

#### 3. 启动 API 网关（Go）

```bash
./scripts/dev.sh gateway
```

API 网关默认监听 `http://localhost:8080`，健康检查地址为 `http://localhost:8080/health`。

#### 4. 启动前端（React + Vite）

```bash
./scripts/dev.sh web
```

前端开发服务器默认运行在 [http://localhost:5173](http://localhost:5173)。开发模式下，Vite 会将 `/api` 请求代理到 `http://localhost:8080`；如需绕过代理，可设置 `VITE_API_BASE_URL`。

#### 本地开发环境变量速查

| 变量 | 值（本地混合模式） | 说明 |
| :--- | :--- | :--- |
| `DATABASE_URL` | `postgres://user:pass@localhost:5432/gogopher?sslmode=disable` | Learning Gateway 连接本地 PostgreSQL |
| `LEARNING_SLICE_ENABLED` | `true` | 仅在 `APP_ENV=local` 时启用 Learning API |
| `LEARNING_CONTENT_DIR` | `content/learning` | release、schema 与 current pointer 根目录 |
| `SANDBOX_LISTEN_ADDRESS` | `127.0.0.1:8081` | 本地 Sandbox 监听地址；Compose 内显式覆盖为容器接口 |
| `VITE_API_BASE_URL` | `/api/v1` | 前端 API 基址；默认走相对路径和代理 |

### 环境变量

以下变量可在 `.env` 中覆盖；默认值见 `.env.example`。

| 变量 | 所在服务 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `WEB_PORT` | host | `3000` | Web 容器映射到宿主机的端口 |
| `POSTGRES_PORT` | host | `5432` | PostgreSQL 映射到宿主机的端口 |
| `DATABASE_URL` | `gateway`, `migrate` | `postgres://user:pass@postgres:5432/gogopher?sslmode=disable` | PostgreSQL 连接串 |
| `LEARNING_SESSION_TTL` | `gateway` | `720h` | 匿名 Learner session 有效期 |
| `SANDBOX_ENDPOINT` | `gateway` | `http://127.0.0.1:8081/v1/executions` | versioned Sandbox 内部 endpoint；Compose 会覆盖为 service 地址 |
| `SANDBOX_RPC_DEADLINE` | `gateway` | `35s` | Gateway 等待 Sandbox 完整响应的上限 |
| `EXECUTION_WORKER_LEASE` | `gateway` | `45s` | PostgreSQL execution worker lease 时长 |
| `EXECUTION_MAX_CLAIMS` | `gateway` | `3` | worker crash/lease expiry 后允许的最大 claim 次数 |
| `POSTGRES_USER` | `postgres` | `user` | 数据库用户名 |
| `POSTGRES_PASSWORD` | `postgres` | `pass` | 数据库密码 |
| `POSTGRES_DB` | `postgres` | `gogopher` | 数据库名 |

### 访问地址

服务启动后，可通过以下地址访问：

- **前端界面**：[http://localhost:3000](http://localhost:3000)
- **Learning API**：通过 Web 的 [http://localhost:3000/api/v1/learning](http://localhost:3000/api/v1/learning) 反向代理访问
- **PostgreSQL**：`localhost:5432`

### 生产环境注意事项

- 请将默认的数据库密码 (`pass`) 修改为强密码，并通过环境变量或 Docker Secrets 注入。
- 前端 Nginx 已在本地容器模式下代理 `/api` 到 Gateway；生产环境可根据需要补充 HTTPS、Gzip 压缩、缓存策略和外层反向代理。
- Compose 已为核心服务配置健康检查和 `restart: unless-stopped`，但还没有覆盖滚动发布、备份恢复和集中式日志。
- 当前 Sandbox 只适合本地可信学习环境；`network=none` 在响应中明确标记为 `policy_only`，不代表进程已被网络、CPU 或内存隔离。
- held-out source 在测试 binary 生成后、运行用户代码前删除，但开源内容和同一进程信任域意味着该机制不能抵抗恶意逆向，也不能作为认证防作弊边界。
- 如需暴露到公网，请在网关前添加反向代理（如 Nginx、Traefik）并配置 SSL 证书。

---

## 开源协议与课程来源

本项目采用 [MIT License](LICENSE) 协议。

Go 基础训练营是 GoGopher Arch 的完整内置课程：知识点组织、课程讲解、sandbox 练习、验收标准和复盘问题由本项目重新整理生成，课程页面不依赖外部教程正文或章节阅读入口。

外部 Go 教程与社区资料仅作为历史参考和灵感来源保留在项目说明中。其中《Go 语言圣经中文版》项目 [gopl-zh/gopl-zh.github.com](https://github.com/gopl-zh/gopl-zh.github.com) 的仓库 LICENSE 为 [BSD 3-Clause](https://github.com/gopl-zh/gopl-zh.github.com/blob/master/LICENSE)；其[授权附录](https://gopl-zh.github.io/appendix/appendix-c-cpoyright.html)说明正文采用 CC-BY 3.0，代码遵循 Go 项目的 BSD 协议。
