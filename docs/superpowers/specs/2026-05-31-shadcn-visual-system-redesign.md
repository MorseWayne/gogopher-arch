# 全站 shadcn 视觉系统重构设计

状态：已由用户确认  
日期：2026-05-31  
范围：GoGopher Arch 前端全站视觉系统与信息架构优化

## 1. 背景

GoGopher Arch 当前前端是 React + Vite + Tailwind 应用，已经包含首页、学习总览、Go 基础课程、章节详情、任务详情和沙盒执行链路。用户希望参考 VuePress Theme Hope 的清晰文档体验，同时使用 shadcn/ui 重构全站视觉系统。

本设计选择保留现有 React/Vite 技术栈，不迁移到 VuePress。VuePress Theme Hope 作为信息架构和阅读体验参考；shadcn/ui 作为组件系统和设计规范基础。

## 2. 目标

- 将全站 UI 收敛到 shadcn/ui 组件体系，减少一次性自定义 UI。
- 优化信息架构：公开首页负责介绍项目价值，学习区负责课程、任务和沙盒操作。
- 提升课程页阅读体验，使 Go 基础训练营更接近高质量文档站。
- 提升任务页动手体验，使任务 Brief、验收标准、沙盒运行和导师反馈路径更清晰。
- 保留现有沙盒 API 链路，不因 UI 重构导致执行功能回退。
- 对未实现路线使用“即将开放”标识，不制造假功能。

## 3. 非目标

- 不迁移到 VuePress 或 VuePress Theme Hope。
- 不新增登录、账号体系、真实学习进度持久化。
- 不实现 AI 全栈路线、工程进阶路线的真实练习功能。
- 不重写后端 API、Gateway 或 Sandbox Engine。
- 不做与本轮 UI 重构无关的大规模课程内容改写。

## 4. 已选方案

采用 **方案 B：分区 App Shell + 页面职责重排**。

- 公开区使用 `PublicLayout`。
- 学习区使用 `LearningLayout`，包含 shadcn `Sidebar`、顶部栏、`Breadcrumb` 和学习状态入口。
- 首页保持品牌入口和路径介绍。
- Dashboard 负责告诉用户下一步该做什么。
- 课程页负责系统学习和章节阅读。
- 任务页负责动手实战和沙盒反馈。

不采用的方案：

- 保守换肤：风险低，但信息架构提升不足。
- 完整学习平台重设计：产品感强，但会引入大量当前没有的数据和功能。

## 5. 架构与路由分区

### 5.1 PublicLayout

适用路由：

- `/`

职责：

- 展示 GoGopher Arch 的定位和价值。
- 展示学习路径卡片。
- 引导用户进入 Go 基础训练营或学习总览。
- 展示工程进阶、AI 全栈路线等未实现方向，但标记为“即将开放”。

结构：

- 顶部公开导航：Logo、首页、Go 基础训练营、GitHub。
- Hero：项目定位、核心 CTA。
- 学习路径卡：Go 基础、后端实习任务、工程能力进阶、AI 全栈路线。
- 路线说明和当前推荐入口。

### 5.2 LearningLayout

适用路由：

- `/dashboard`
- `/courses/go-basics`
- `/courses/go-basics/:id`
- `/missions/:id`

职责：

- 提供学习区统一导航。
- 统一页面容器、面包屑、移动端菜单行为。
- 保持课程、任务、沙盒入口的一致访问路径。

Sidebar 分组：

- 工作区
  - 总览
  - 沙盒快捷入口
- 学习路径
  - Go 基础训练营
  - 后端实习任务线
  - 工程能力进阶（即将开放）
  - AI 全栈路线（即将开放）
- 项目
  - GitHub / README

沙盒快捷入口不新增独立 `/sandbox` 页面。本轮将它定义为上下文跳转：

- 如果当前页面已有练习或任务沙盒，跳转到当前页面的 `#exercise` 或 `#sandbox` 区域。
- 如果当前页面没有沙盒，跳转到默认可运行入口，例如 Go 基础训练营第一个可运行章节的 `#exercise`。
- 如果未来新增独立沙盒页，再把该入口升级为真实路由；本轮不得留下悬挂导航。

沙盒快捷入口解析规则必须可测试：

```ts
type SandboxTarget =
  | { kind: "anchor"; href: string; label: string }
  | { kind: "fallback"; href: string; label: string; reason: string }
  | { kind: "unavailable"; message: string };
```

解析顺序：

1. 当前路由是章节页且页面存在 `#exercise`：返回当前章节 `#exercise`。
2. 当前路由是任务页且页面存在 `#sandbox`：返回当前任务 `#sandbox`。
3. 当前路由同时存在多个候选锚点时，优先级为 `#sandbox` 高于 `#exercise`，因为任务执行区比课程练习更接近“沙盒”语义。
4. 当前页面没有可用锚点时，返回第一个可运行 Go 基础章节的 `#exercise` 作为 `fallback`。
5. 默认章节目标不存在、课程数据为空或锚点检查失败时，不跳转到不存在页面；显示非阻塞提示，并提供进入课程总览的按钮。

验收必须覆盖：当前页有锚点、当前页无锚点、多个候选锚点、默认目标缺失、目标锚点不存在。

移动端：

- Sidebar 折叠进 `Sheet` 或等价侧边抽屉。
- 顶部保留菜单按钮和当前页面标题。
- 章节右侧目录在窄屏隐藏或折叠。

## 6. 视觉语言与 Design Token

整体采用 **自适应混合风格**：

- 课程正文和主要页面使用浅色阅读 surface。
- 沙盒、终端输出、运行结果和任务重点状态使用局部深色 surface。
- 全站用 Go 蓝作为品牌主色。
- 避免回到整页大面积深色、一次性自定义 class 的旧风格。

Token 原则：

- `primary`：Go 蓝。
- `background` / `card` / `muted` / `border`：使用 shadcn 语义 token。
- `destructive`：用于错误和失败状态。
- 局部终端 surface 使用深色容器组件封装，不散落在页面中。

组件原则：

- 使用 shadcn 组件表达结构和状态。
- 优先使用组件 variant，不手写颜色和状态样式。
- `className` 主要用于布局，不用于覆盖组件语义样式。
- 使用 `gap-*`，避免 `space-x-*` / `space-y-*`。
- 使用 `Badge` 表示状态，不手写状态 pill。
- 使用 `Alert` 表示错误、提示和学习反馈。
- 使用 `Separator` 替代手写分隔线。

## 7. 组件分层

### 7.1 UI primitives

来自 shadcn/ui 或现有 `ui` 组件目录：

- `Button`
- `Badge`
- `Card`
- `Tabs`
- `Accordion`
- `Alert`
- `Progress`
- `Sidebar`
- `Breadcrumb`
- `Sheet`
- `ScrollArea`
- `Separator`
- `Skeleton`
- `Tooltip`
- `Avatar`

实施前需要检查现有 `web/src/app/components/ui`，避免重复添加或覆盖组件。

### 7.2 App composites

建议新增或整理的应用级组合组件：

- `PublicLayout`
- `LearningLayout`
- `LearningSidebar`
- `PageHeader`
- `PathCard`
- `ComingSoonBadge`
- `TerminalPanel`
- `ExerciseRunnerPanel`
- `ChapterNav`
- `ChapterToc`
- `MissionBriefCard`
- `AcceptanceCriteriaCard`

这些组件封装 GoGopher 业务语义，页面只负责组合和传入数据。

### 7.3 最小组件契约

关键组合组件需要使用明确的输入输出，避免页面间行为漂移。实施计划可在不改变语义的前提下细化类型名。

```ts
type Availability = "available" | "coming-soon";

type LearningPathItem = {
  id: string;
  title: string;
  description: string;
  href?: string;
  availability: Availability;
};

type ChapterSummary = {
  id: string;
  title: string;
  summary: string;
  objectives: string[];
  hasExercise: boolean;
};

type ExerciseSpec = {
  instructions: string;
  starterCode: string;
  expectedOutput?: string;
};

type MissionSummary = {
  id: string;
  title: string;
  brief: string;
  objectives: string[];
  acceptanceCriteria: string[];
  starterCode: string;
  lesson?: string;
};

type SandboxRunState = "idle" | "running" | "success" | "error" | "timeout";

type SandboxRunResult = {
  state: SandboxRunState;
  stdout?: string;
  stderr?: string;
  exitCode?: number;
  message?: string;
};

type MetricSource = "staticMock" | "derivedFromContent" | "localSession";

type DisplayMetric<T> = {
  value: T;
  source: MetricSource;
  label: string;
  helperText?: string;
};
```

Dashboard 中的路径进度、今日任务和继续学习必须绑定 `DisplayMetric` 来源：

- `staticMock`：静态演示值，文案需标注“访客演示”。
- `derivedFromContent`：从现有课程/任务内容数量推导，例如 13 章、Day 0-5。
- `localSession`：当前浏览器会话内的运行结果或临时交互状态；刷新后可丢失。

不得把这些指标描述为账号级真实进度或跨设备同步状态。

组件契约：

- `LearningSidebar` 接收当前路径、学习路径条目和沙盒快捷入口目标；不得渲染无目标的可点击项。
- `PathCard` 接收 `LearningPathItem`；`coming-soon` 项展示 Badge，并禁用真实跳转。
- `ChapterNav` 接收 `ChapterSummary[]`、当前章节 ID 和 `onSelectChapter(id)`；点击只做路由跳转或调用回调，不修改章节数据；内容为空时显示可恢复空态。
- `ChapterToc` 接收页面 headings 和 `onSelectHeading(id)`；无 headings 时隐藏或显示空态，不阻塞正文。
- `MissionBriefCard` 接收 `MissionSummary`，只展示任务语义，不发起 API。
- `AcceptanceCriteriaCard` 接收验收标准字符串列表和可选完成状态；空列表显示“暂无验收标准”的 `Alert`，不白屏；点击或勾选行为仅限本地 UI 状态，不宣称持久化。
- `TerminalPanel` 接收 `SandboxRunResult` 和可选 `onRetry()`；只负责展示 stdout、stderr、exit code、timeout 和可读说明；不得自行发起 API。
- `ExerciseRunnerPanel` 接收 `ExerciseSpec`、当前代码、运行状态、`onCodeChange(code)`、`onRun(code)` 和 `onRetry()`；`onRun` 返回 `Promise<SandboxRunResult>`；运行失败或超时后必须保留用户代码。
- 所有组合组件的错误回传通过显式 props 或回调完成，不在组件内部吞掉异常。

## 8. 核心页面布局

### 8.1 Landing

职责：公开首页，讲清项目是什么、用户从哪里开始。

布局：

- Hero：项目定位、主 CTA、次 CTA。
- 学习路径卡：Go 基础、后端实习任务、工程进阶、AI 全栈路线。
- 当前推荐入口：继续 Go 基础训练营。
- 路线图摘要：明确哪些路线已可用，哪些即将开放。

成功标准：

- 新用户能在 5 秒内理解平台价值。
- 能清晰找到“开始 Go 基础训练营”和“进入学习总览”。
- 未实现路线不会误导用户。

### 8.2 Dashboard

职责：学习总览，减少选择成本，告诉用户下一步做什么。

布局：

- 顶部 PageHeader：当前学习模式、路径进度、快捷入口。
- 下一步 Card：继续学习章节或任务。
- 今日任务 Checklist：阅读、运行、完成任务。
- 路径进度 Card：Go 基础、后端实习、工程进阶、AI 路线。
- 最近运行 / 导师提示摘要。

数据：

- 本轮不新增真实后端进度，也不展示登录身份。
- “当前学习模式”使用访客/本地学习模式文案。
- 进度和下一步使用现有课程/任务数据派生或静态展示。
- 所有 Dashboard 指标必须在实现层标记来源：`staticMock`、`derivedFromContent` 或 `localSession`。
- 静态演示指标在 UI 文案中使用“访客演示”“本地会话”或等价说明，避免被理解为真实账号进度。
- 最近运行仅来自当前浏览器会话内的沙盒运行结果；没有运行记录时显示空态。
- 导师提示摘要来自课程/任务已有静态提示或通用学习建议，不调用 LLM，也不新增后端依赖。

### 8.3 GoBasicsCourse

职责：课程总览。

布局：

- 课程 Header：目标、章节数、学习建议。
- 章节卡列表：标题、学习目标、练习状态入口。
- 课程说明：为什么先学 Go 基础，如何进入任务线。

组件：

- `Card`
- `Badge`
- `Progress`
- `Button`
- `Alert`

### 8.4 GoBasicsChapter

职责：文档站式章节阅读和章节练习。

桌面布局：

- 左侧：章节目录。
- 中间：正文、学习目标、现代 Go 说明、工程实践、常见坑、复盘问题。
- 右侧：本页目录。
- 底部或正文后：练习沙盒。

移动端布局：

- 章节目录折叠。
- 本页目录隐藏或折叠。
- 正文优先，沙盒在正文后展示。

组件：

- `ScrollArea`
- `Accordion`
- `Alert`
- `Card`
- `Badge`
- `Button`
- `TerminalPanel`

### 8.5 MissionDetail

职责：任务工作台。

桌面布局：

- 左侧：任务 Brief、背景、目标、限制条件、验收标准、任务前小课。
- 右侧：代码输入、运行按钮、终端结果、导师反馈。

导师反馈在本轮中是静态/规则化反馈：来自任务数据中的提示、验收标准和沙盒运行结果组合，不接入真实 AI 导师或远程评审服务。

移动端布局：

- Brief → 验收标准 → 小课 → 沙盒 → 反馈。

组件：

- `Card`
- `Badge`
- `Alert`
- `Tabs`
- `Button`
- `TerminalPanel`
- `AcceptanceCriteriaCard`

## 9. 数据流与状态

### 9.1 本地数据

继续使用现有 TypeScript 数据文件承载：

- Go 基础课程正文、章节、练习 starter code。
- 实习任务线任务描述、验收标准和初始代码。

本轮不引入 CMS、数据库课程读取或远程内容服务。

### 9.2 会话内状态与静态提示

Dashboard 和任务页中容易被误解为“动态服务”的内容必须限定来源：

- 最近运行：仅保存当前浏览器会话内最后一次沙盒运行结果；刷新页面后可以丢失。
- 导师提示：来自课程/任务数据中的静态提示、验收标准、常见坑或通用学习建议。
- 任务反馈：由沙盒运行结果、验收标准和静态提示组合展示。
- 本轮不使用 localStorage 作为产品承诺；如实施时为了体验临时缓存，必须在计划中明确且不得宣称跨设备同步。

### 9.3 API 数据

只有沙盒运行走 API：

```text
用户代码 → executeCode() → /api/v1/execute → Gateway → Sandbox Engine → 运行结果
```

UI 重构不得改变现有 API 语义。

### 9.4 沙盒状态机

状态：

- `idle`：等待运行。
- `running`：请求中，运行按钮禁用，显示 loading。
- `success`：展示 stdout、exit code 和成功提示。
- `error`：展示可读错误和原始 stderr。
- `timeout`：作为学习反馈展示，不当作页面崩溃。

原则：

- 用户代码在错误后保留。
- 网络错误提供重试。
- 编译失败和测试失败属于正常学习反馈。
- Toast 只用于轻量反馈，不替代结果面板。

## 10. 错误处理

- 章节 ID 不存在：显示 Not Found 状态，并提供返回课程总览按钮。
- 任务 ID 不存在：显示 Not Found 状态，并提供返回 Dashboard 或任务列表入口。
- 沙盒网络错误：`Alert` 显示可读说明，保留代码，提供重试按钮。
- 沙盒编译失败：终端面板展示 stderr，提示这是练习反馈。
- Coming soon 路线：使用 disabled nav 或点击后轻量提示，不跳转到假功能页。
- 章节或任务内容为空：显示空态 Card/Alert，提供返回课程总览、Dashboard 或可用章节入口。
- 章节目录或本页目录无法生成：隐藏目录区域或显示非阻塞空态，正文仍可阅读。
- 沙盒超时：在终端面板中以 `timeout` 状态展示，保留代码，提供重试按钮。
- 数据字段缺失导致无法渲染关键区域：页面展示可读错误和返回入口，不白屏。

## 11. 实施切片

### Slice 1：基础设施

- 检查现有 shadcn/ui 组件和导入路径。
- 规范全局 token 和主题变量。
- 建立或整理 app composites 的目录边界。

### Slice 2：App Shell

- 拆分 `PublicLayout` 和 `LearningLayout`。
- 实现学习区 Sidebar 分组。
- 实现沙盒快捷入口的上下文跳转或默认练习跳转，不新增悬挂 `/sandbox` 路由。
- 实现移动端菜单行为。
- 保持现有 URL 可访问。

### Slice 3：Landing

- 重构公开首页 Hero、路径卡和 CTA。
- 标记未实现路线为“即将开放”。
- 删除或弱化过度赛博/炫技但不服务学习路径的视觉噪声。

### Slice 4：学习页面

- 重构 Dashboard。
- 重构 Go 基础课程总览。
- 重构章节详情为文档站式阅读布局。
- 重构任务详情为任务工作台布局。
- 保留沙盒执行能力。

### Slice 5：验证收口

- 前端 build。
- 路由 smoke test。
- 沙盒状态 smoke test。
- 移动端布局抽查。
- Coming soon 无假功能检查。

## 12. 验证策略

### 自动验证

- `npm run build --prefix web`
- 现有 Go 后端测试如涉及沙盒链路变更则运行；本轮 UI 重构原则上不改后端。

### 手动 smoke

检查以下路径：

- `/`
- `/dashboard`
- `/courses/go-basics`
- `/courses/go-basics/:id`
- `/missions/:id`

检查以下状态：

- 学习区 Sidebar 桌面/移动端行为。
- Coming soon 项不能误导用户进入假功能。
- 沙盒运行 loading/success/error/timeout 状态仍可见。
- Sidebar 沙盒快捷入口不会指向不存在页面：有上下文时跳到当前沙盒区域，无上下文时跳到默认可运行练习。
- 沙盒快捷入口边界用例可复现：当前页无锚点、多个候选锚点、默认目标缺失、目标锚点不存在。
- Dashboard 最近运行为空、会话清空或刷新后，显示明确空态，不展示假历史记录。
- Dashboard 静态演示或派生指标标明来源，不暗示真实账号级进度。
- 章节/任务不存在时不白屏。
- 章节/任务内容为空或目录无法生成时显示空态或降级展示。

### 边界用例断言

实施计划应把以下场景纳入可复现检查，可用轻量单元测试、组件测试或手动 smoke 覆盖：

- 章节 ID 不存在时出现返回课程总览入口。
- 任务 ID 不存在时出现返回 Dashboard 或可用任务入口。
- 章节内容为空时显示空态，不影响页面框架。
- 验收标准为空时 `AcceptanceCriteriaCard` 显示 Alert。
- 本页目录无法生成时正文仍可阅读。
- 沙盒快捷入口在当前页无锚点、多个候选锚点、默认目标缺失、目标锚点不存在时都有明确兜底。
- 沙盒 timeout、网络失败、编译失败分别展示不同状态，且保留用户代码。
- 最近运行为空或会话状态清空后显示空态。
- Coming soon 项不会进入假功能页。

### 体验验收

- 首页能快速说明项目价值和起点。
- Dashboard 能明确告诉用户下一步。
- 课程章节阅读更清晰，不被深色视觉干扰。
- 任务页能顺畅完成“阅读任务 → 修改代码 → 运行 → 查看反馈”的闭环。

## 13. 风险与缓解

| 风险 | 缓解 |
| --- | --- |
| 已有 shadcn/ui 组件和新组件重复 | 实施前检查现有 `ui` 目录，优先复用，必要时再用 CLI 更新。 |
| 视觉重构破坏沙盒功能 | 沙盒面板作为独立组合组件迁移，每个切片做运行状态 smoke。 |
| Coming soon 误导用户 | 使用 Badge 和 disabled 行为，不创建假交互。 |
| 沙盒导航入口悬挂 | 本轮不新增 `/sandbox`；Sidebar 沙盒入口必须跳到现有练习/任务沙盒锚点或默认可运行章节。 |
| 最近运行和导师提示被误解为后端功能 | 明确限定为会话内运行结果和静态提示，不宣称持久化或 AI 服务。 |
| 页面一次性重写导致难以验证 | 按实施切片推进，每片 build/smoke 后再继续。 |
| 移动端学习区复杂 | Sidebar 使用 Sheet/Drawer，章节右侧目录折叠或隐藏。 |

## 14. 用户确认记录

- 选择“视觉 + 信息架构优化”，不是单纯换肤。
- 选择混合风格：首页和 Dashboard 保留平台感，课程页偏文档站，任务页偏工作台。
- 选择分区混合壳：首页为公开入口，学习区使用 App Shell。
- 选择自适应混合视觉：浅色阅读 + 局部深色终端。
- 选择学习区 Sidebar 分组：工作区、学习路径、项目。
- 选择未实现路线保留入口并标记“即将开放”。
- 选择页面职责重排。
- 确认方案 B：分区 App Shell + 页面职责重排。
- 已确认设计第 1-5 部分。

## 15. 后续步骤

1. 对本规格进行独立评审并修正问题。
2. 用户复核规格文档。
3. 用户确认后，进入实施计划阶段。
4. 按切片执行全站 shadcn 重构。
