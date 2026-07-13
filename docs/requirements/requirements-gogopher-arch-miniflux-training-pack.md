# Miniflux 开源项目训练包选型与验证需求

- 日期：2026-07-12
- 状态：候选项目已完成本地可行性验证，建议锁定为第一版主训练项目
- 关联文档：`requirements-gogopher-arch-product-positioning.md`
- 上游仓库：<https://github.com/miniflux/v2>

## 一、选型结论

第一版开源项目训练包选择 Miniflux，贯穿 M2“Go 后端开发”和 M3“生产工程能力”。训练目标不是要求用户立即向上游贡献，而是借助真实代码、真实依赖和真实运行约束形成可验证的 Go 工程能力。

建议固定基线：

- Release：`v2.3.2`
- Commit：`51f2e0d8199ea8fa305081f6e175bba64b0ef94b`
- 上游许可证：Apache-2.0
- Go 工具链：Go 1.26.0
- 数据库基线：PostgreSQL 15，后续可以增加兼容版本验证

训练包必须同时记录 tag 和完整 commit。不得使用 `main`、`latest` 或只依赖浮动 Docker tag。

## 二、真实验证结果

验证环境为 Linux amd64、宿主 Go 1.25.6、Docker 29.5.3。Miniflux 通过 Go toolchain 机制使用显式 Go 1.26.0。

### 仓库规模

- 411 个 Go 文件。
- 79 个 Go 测试文件。
- 86 个 Go package。
- 约 78,451 行 Go 代码。
- 主要学习区域包括 `internal/api`、`internal/cli`、`internal/config`、`internal/http`、`internal/model`、`internal/reader`、`internal/storage` 和 `internal/worker`。

该规模足以承载大型仓库阅读和复杂系统训练，但不适合让 M2 用户无引导地从仓库根目录自由探索。每个任务必须限定阅读切片、入口符号和验证目标。

### 构建与测试

在依赖已经下载的验证环境中：

- `make miniflux`：成功，约 13 秒。
- `GOTOOLCHAIN=go1.26.0 make test`：全量 race 单元测试成功，约 22 秒。
- 隔离 PostgreSQL 下的 API 冒烟测试：健康检查、版本、分类创建和 feed 创建均成功，约 12 秒。
- 对 v2.3.1 执行的完整 API 集成测试成功，约 166 秒；2.3.2 冒烟测试验证其主链路仍然成立。

首次直接使用宿主 Go 1.25.6 和自动工具链运行 `make test` 时，曾出现 Go 1.25.6 与 Go 1.26.0 编译产物混用。显式设置 `GOTOOLCHAIN=go1.26.0` 后测试稳定通过。因此正式训练环境必须直接提供匹配的 Go 1.26 工具链，不能把自动下载和宿主缓存状态当作可靠前提。

完整 API 集成测试会访问公网 feed，并产生大量抓取与解析日志。必修训练任务不得依赖公网资源，需要由本地 fixture server 提供 RSS、Atom、JSON Feed、超时、限流、无效内容和大响应等确定性场景。

## 三、核心架构与执行链

### 启动与关闭链路

```text
main.main
  → cli.Parse
  → 解析并校验配置
  → 生成静态资源 bundle
  → 创建数据库连接池与 Storage
  → 执行或校验数据库迁移
  → startDaemon
      ├── 创建 worker pool
      ├── 启动 feed 与 cleanup scheduler
      ├── 启动 HTTP server
      ├── 启动 metrics collector
      └── 等待信号并按顺序关闭 HTTP、metrics 与 worker
```

主要源码切片：

- `main.go`
- `internal/cli/cli.go`
- `internal/cli/daemon.go`
- `internal/http/server/server.go`
- `internal/http/server/routes.go`
- `internal/database/database.go`

该链路适合训练配置、依赖初始化、数据库连接、服务生命周期、goroutine、信号处理和 graceful shutdown。

### 创建 Feed 的同步请求链路

```text
POST /v1/feeds
  → API 认证中间件
  → api.createFeedHandler
  → 请求解码与 validator
  → reader/handler.CreateFeed
  → HTTP fetcher 与 response handler
  → feed parser
  → entry processor
  → storage.CreateFeed
  → PostgreSQL 事务与 entry 写入
  → icon discovery
  → 201 JSON response
```

主要源码切片：

- `internal/api/api.go`
- `internal/api/middleware.go`
- `internal/api/feed_handlers.go`
- `internal/validator/feed.go`
- `internal/reader/handler/handler.go`
- `internal/reader/fetcher/`
- `internal/reader/parser/`
- `internal/storage/feed.go`

该链路适合训练 HTTP API、认证、校验、错误映射、外部 I/O、解析、事务和端到端测试。

### 后台刷新链路

```text
feedScheduler
  → storage.BatchBuilder 查询到期任务
  → worker.Pool.Push
  → worker.Run
  → reader/handler.RefreshFeed
      ├── 读取 feed 状态
      ├── ETag / Last-Modified / Retry-After 处理
      ├── 下载、解析与处理 entries
      ├── storage.RefreshFeedEntries
      ├── 可选推送第三方集成
      └── 更新下次检查时间与错误状态
```

主要源码切片：

- `internal/cli/scheduler.go`
- `internal/storage/batch.go`
- `internal/worker/pool.go`
- `internal/worker/worker.go`
- `internal/reader/handler/handler.go`
- `internal/storage/entry.go`

该链路适合训练 channel、worker pool、调度、HTTP 缓存、限流、重试、事务粒度、指标、故障恢复和资源关闭。

## 四、训练包结构

训练包不复制一套与上游脱节的教学项目，而由以下部分组成：

```text
训练包元数据
├── 上游仓库、tag、commit、许可证与校验值
├── 固定 Go 与 PostgreSQL 环境
├── 本地 feed fixture server
├── 数据库种子与可重置状态
├── 任务清单与能力节点映射
├── 阅读切片、入口符号和调用链提示
├── 可见测试、隐藏验证和故障注入
├── 分级提示与参考解释
└── 上游版本升级审校记录
```

上游源码保持原始状态。任务所需的初始变更、测试和故障注入通过可重置的 patch、训练分支或外部 harness 管理。必修任务不能依赖某个上游 Issue 长期保持开放。

## 五、首批任务梯度

### A. 建立仓库心智模型

1. 构建、运行并调用健康检查和版本 API。
2. 从 `main` 追踪到 HTTP server、worker pool 和数据库。
3. 修改一个配置项并解释其解析、校验和使用路径。
4. 观察 SIGTERM 下的 graceful shutdown，定位各组件的关闭顺序。

### B. M2 后端开发

5. 追踪分类创建请求，补充输入校验和 API 测试。
6. 为只读 API 增加一个字段，完成 model、storage、handler 和测试变更。
7. 增加一个数据库字段与 migration，验证升级和回滚边界。
8. 使用本地 feed fixture 追踪 feed 创建全过程并修复一个错误映射问题。
9. 为外部抓取增加明确的超时行为和失败测试。
10. 从需求说明独立完成一个跨 API、storage 和数据库的小功能。

### C. M3 生产工程

11. 为后台刷新链路增加或校准指标，并解释 label 基数风险。
12. 注入 feed 超时、429、无效内容和 PostgreSQL 短暂不可用故障。
13. 使用 race detector、pprof 或 trace 定位一个预置问题。
14. 分析 worker pool 的吞吐、背压和关闭行为，提交改进方案与实验。
15. 完成压测，定位 API 或数据库热点并验证优化效果。
16. 执行部署、数据库迁移、版本升级、回滚和故障复盘演练。

### D. 迁移验证

17. 在用户自建 M2 服务中复现 Miniflux 的一个工程模式。
18. 在陌生模块中完成同类型修改，避免只记住源码位置。
19. 间隔复验启动链、API 链和后台链，并重新完成一项限制 AI 的任务。

## 六、测试分层与反馈速度

训练环境需要提供四层验证，避免每次练习都运行完整测试：

| 层级 | 目标 | 期望反馈时间 |
|---|---|---:|
| Task | 当前能力点的可见测试和隐藏测试 | 10 秒内 |
| Package | 受影响 package 与 race 测试 | 30 秒内 |
| Repository | 全量单元、race 和静态检查 | 1 分钟内 |
| Integration | PostgreSQL、HTTP fixture 和端到端场景 | 3 分钟内 |

公开测试用于帮助用户理解契约；隐藏测试用于验证边界、迁移和防止只针对示例编码。隐藏测试不能引入任务描述中没有说明的需求。

## 七、能力证据要求

每项工程任务至少保存：

- 用户首次尝试和最终实现的 diff。
- 测试、日志或实验输出。
- 使用过的提示等级和 AI 协作记录。
- 用户对执行链、错误边界和方案取舍的解释。
- 对上游代码风格和现有设计的遵循情况。
- 后续迁移任务与延迟复验结果。

只让测试通过不能自动证明掌握。关键任务需要同时通过代码、测试、解释和迁移验证。

## 八、风险与约束

- **仓库规模过大**：每个 M2 任务限制推荐阅读面，逐步扩大仓库范围。
- **Go 版本较新**：训练镜像固定 Go 1.26，并在启动前验证工具链，不依赖宿主自动下载。
- **PostgreSQL 依赖**：提供一键启动、健康检查、种子数据和任务级重置。
- **公网测试不稳定**：必修任务统一使用本地 fixture；公网只用于可选观察实验。
- **上游持续演进**：训练包固定 commit；升级作为独立审校流程，不自动漂移。
- **贡献预期错误**：明确区分训练 patch 与可向上游提交的变更，贡献是可选出口。
- **许可证与归属**：保留 Apache-2.0 LICENSE、NOTICE、上游来源和修改说明。

## 九、进入实现前的验收条件

- 固定环境能够从零启动并在支持的平台上重复通过。
- 必修任务不依赖公网，数据库和 fixture 可以安全重置。
- 至少完成启动、API、后台任务三条源码导览。
- 首批任务均映射到明确能力节点、前置条件和证据标准。
- Task、Package、Repository、Integration 四层测试入口可独立运行。
- 至少由一个目标用户实际完成前三个任务并记录阻塞点。
