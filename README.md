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
docker compose up --build
```

前端默认运行在 `http://localhost:3000`，Gateway 默认运行在 `http://localhost:8080`。

---

## 部署说明

### 环境要求

- [Docker](https://docs.docker.com/get-docker/) 20.10+
- [Docker Compose](https://docs.docker.com/compose/install/) v2+

### 服务架构

本项目通过 Docker Compose 编排以下服务：

| 服务 | 说明 | 镜像 / 构建方式 | 暴露端口 |
| :--- | :--- | :--- | :--- |
| `web` | React + Tailwind 前端（Nginx 托管） | `web/Dockerfile` | `3000` → `80` |
| `gateway` | Go API 网关 | `src/services/gateway/Dockerfile` | `8080` |
| `sandbox-engine` | Go 沙盒执行引擎 | `src/services/sandbox-engine/Dockerfile` | `8081` |
| `postgres` | 数据库 | `postgres:15-alpine` | `5432` |
| `redis` | 缓存 | `redis:7-alpine` | `6379` |

### 部署步骤

1. **克隆仓库**

   ```bash
   git clone https://github.com/MorseWayne/gogopher-arch.git
   cd gogopher-arch
   ```

2. **启动所有服务**

   ```bash
   # 前台运行（便于查看日志）
   docker compose up --build

   # 后台运行
   docker compose up --build -d
   ```

3. **查看服务状态**

   ```bash
   docker compose ps
   ```

4. **查看日志**

   ```bash
   # 查看所有服务日志
   docker compose logs -f

   # 查看指定服务日志
   docker compose logs -f gateway
   ```

5. **停止服务**

   ```bash
   docker compose down
   ```

### 本地开发（混合模式）

如果你希望在本地运行上层应用代码（便于断点调试和热重载），同时用 Docker 只启动基础依赖（PostgreSQL、Redis），可按以下步骤操作。

#### 本地开发环境要求

- [Docker](https://docs.docker.com/get-docker/) 20.10+
- [Node.js](https://nodejs.org/) 20+
- [Go](https://go.dev/dl/) 1.24+

#### 1. 启动基础组件

```bash
docker compose up postgres redis -d
```

#### 2. 启动沙盒引擎（Go）

```bash
go run ./src/services/sandbox-engine/main.go
```

沙盒引擎默认监听 `http://localhost:8081`。

#### 3. 启动 API 网关（Go）

```bash
export DB_URL="postgres://user:pass@localhost:5432/gogopher?sslmode=disable"
export REDIS_URL="localhost:6379"
go run ./src/services/gateway/main.go
```

API 网关默认监听 `http://localhost:8080`。

#### 4. 启动前端（React + Vite）

```bash
cd web
npm install
npm run dev
```

前端开发服务器默认运行在 [http://localhost:5173](http://localhost:5173)。

#### 本地开发环境变量速查

| 变量 | 值（本地混合模式） | 说明 |
| :--- | :--- | :--- |
| `DB_URL` | `postgres://user:pass@localhost:5432/gogopher?sslmode=disable` | 连接本地 Docker PostgreSQL |
| `REDIS_URL` | `localhost:6379` | 连接本地 Docker Redis |
| `SANDBOX_URL` | `http://localhost:8081/execute` | Gateway 连接本地沙盒引擎（已硬编码为默认值，一般无需手动设置） |

### 环境变量

以下环境变量在 `docker-compose.yml` 中已预配置，如需自定义可修改该文件：

| 变量 | 所在服务 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `SANDBOX_URL` | `gateway` | `http://sandbox-engine:8081/execute` | 沙盒引擎地址 |
| `DB_URL` | `gateway`, `sandbox-engine` | `postgres://user:pass@postgres:5432/gogopher?sslmode=disable` | PostgreSQL 连接串 |
| `REDIS_URL` | `gateway`, `sandbox-engine` | `redis:6379` | Redis 地址 |
| `POSTGRES_USER` | `postgres` | `user` | 数据库用户名 |
| `POSTGRES_PASSWORD` | `postgres` | `pass` | 数据库密码 |
| `POSTGRES_DB` | `postgres` | `gogopher` | 数据库名 |

### 访问地址

服务启动后，可通过以下地址访问：

- **前端界面**：[http://localhost:3000](http://localhost:3000)
- **API 网关**：[http://localhost:8080](http://localhost:8080)
- **沙盒引擎**（内部调用）：[http://localhost:8081](http://localhost:8081)
- **PostgreSQL**：`localhost:5432`
- **Redis**：`localhost:6379`

### 生产环境注意事项

- 请将默认的数据库密码 (`pass`) 修改为强密码，并通过环境变量或 Docker Secrets 注入。
- 前端 Nginx 配置可根据需要启用 HTTPS、Gzip 压缩和反向代理。
- 建议为 Go 服务添加健康检查（`healthcheck`）和重启策略（`restart: unless-stopped`）。
- 如需暴露到公网，请在网关前添加反向代理（如 Nginx、Traefik）并配置 SSL 证书。

---

## 开源协议与课程来源

本项目采用 [MIT License](LICENSE) 协议。

Go 基础训练营是 GoGopher Arch 的完整内置课程：知识点组织、课程讲解、sandbox 练习、验收标准和复盘问题由本项目重新整理生成，课程页面不依赖外部教程正文或章节阅读入口。

外部 Go 教程与社区资料仅作为历史参考和灵感来源保留在项目说明中。其中《Go 语言圣经中文版》项目 [gopl-zh/gopl-zh.github.com](https://github.com/gopl-zh/gopl-zh.github.com) 的仓库 LICENSE 为 [BSD 3-Clause](https://github.com/gopl-zh/gopl-zh.github.com/blob/master/LICENSE)；其[授权附录](https://gopl-zh.github.io/appendix/appendix-c-cpoyright.html)说明正文采用 CC-BY 3.0，代码遵循 Go 项目的 BSD 协议。
