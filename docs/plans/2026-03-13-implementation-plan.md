# GoGopher Arch: 实施计划 (Implementation Plan)

- **版本**：v1.0.0
- **日期**：2026-03-13
- **目标**：在 4-8 周内交付一个包含核心闭环的 MVP 学习平台。

---

## 1. 核心技术选型 (Tech Stack)

### 1.1 后端 (Go & Cloud Native)
- **语言**：Go 1.22+ (利用最新的 Go Workspace 和性能优化)。
- **Web 框架**：Gin (轻量高效) 或 Kratos (微服务化，适合演示分布式架构)。
- **沙盒运行**：Docker + `os/exec` (容器隔离执行)。
- **监控指标**：OpenTelemetry + Prometheus (实时采集用户代码运行数据)。
- **数据存储**：PostgreSQL (基础数据) + Redis (压测指标实时缓存)。

### 1.2 前端 (Modern Web)
- **框架**：React + TypeScript (强大的组件化能力，适合复杂的 Dashboard)。
- **状态管理**：Zustand (轻量响应式)。
- **可视化动画**：Three.js 或 Canvas API (实现“大爆炸”崩溃动画)。
- **代码编辑器**：Monaco Editor (VS Code 同款，体验极佳)。

### 1.3 AI 集成
- **模型**：Gemini 1.5 Pro/Flash API (快速响应、长上下文，适合代码分析)。
- **提示词工程**：基于 RAG 的 Go 架构师知识库，提供精准的 Code Review。

---

## 2. 阶段性路线图 (Roadmap)

### 2.1 第一阶段：基础设施与 MVP 闭环 (2-4 周)
- [ ] **Task 1: 项目脚手架搭建**
    - [ ] 初始化 Go Monorepo 结构。
    - [ ] 配置 Docker Compose 本地开发环境。
- [ ] **Task 2: Gopher Sandbox 引擎实现**
    - [ ] 实现基础的代码构建与容器化执行逻辑。
    - [ ] 实现执行时长、内存配额限制。
- [ ] **Task 3: Dashboard 核心界面**
    - [ ] 开发 Monaco Editor 集成组件。
    - [ ] 实现简单的任务引导系统原型。
- [ ] **Task 4: Lv.1 实习生任务包**
    - [ ] 开发 3 个关于 Slice/Map 并发不安全、Defer 常见错误的基础挑战。
    - [ ] 实现基于 Unit Test 的自动评测逻辑。

### 2.2 第二阶段：深度反馈与可视化 (4 周)
- [ ] **Task 5: 实时监控系统集成**
    - [ ] 利用 OpenTelemetry 采集容器内的 Goroutine 数。
    - [ ] 前端实现动态波形图显示压测 QPS。
- [ ] **Task 6: “大爆炸”崩溃动画**
    - [ ] 捕捉 Runtime Panic/OOM 异常。
    - [ ] 实现 Canvas/Three.js 驱动的故障动画反馈。
- [ ] **Task 7: AI CTO 初版上线**
    - [ ] 对接 Gemini API 进行初步代码 Review。
    - [ ] 实现针对 Go 常见陷阱的自动化提示系统。

### 2.3 第三阶段：全栈进阶与职业线 (长期迭代)
- [ ] **Task 8: 高性能 IM 挑战 (Lv.2)**
    - [ ] 开发 Mock Socket 服务端。
    - [ ] 引导用户学习 Netpoll 和异步 IO。
- [ ] **Task 9: 云原生实战 (Lv.3)**
    - [ ] 集成模拟的 K8s API 或服务发现演示。
- [ ] **Task 10: 剧情系统与社交化**
    - [ ] 加入虚拟职场邮件通知。
    - [ ] 实现全球排行榜与方案分享。

---

## 3. 核心挑战与应对策略
- **安全性 (Security)**：用户提交的代码必须在严密的 Docker 隔离中运行，防止恶意代码破坏系统。
- **高并发处理 (Concurrency)**：模拟压测时会消耗大量服务器资源，需建立任务队列（如使用 NATS/Redis 队列）进行限流排队。
- **反馈即时性 (Latency)**：代码提交后需在 3-5 秒内给出结果，关键链路需高度优化。
