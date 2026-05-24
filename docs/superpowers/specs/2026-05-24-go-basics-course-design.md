# Go 基础训练营设计

## 背景

GoGopher Arch 当前定位为面向 Go 学习者的实战成长平台，通过虚拟职场任务、任务前小课、交互式沙盒和任务后复盘帮助用户从 Go 基础成长到后端实习岗位能力。现有产品已经有实习任务线和 sandbox 运行能力，但 Go 基础学习内容仍分散在任务前置说明中，不适合系统性学习。

本设计将 `gopl-zh/gopl-zh.github.com` 的《Go 语言圣经中文版》作为参考来源，融合为项目内独立的 Go 基础学习课程。课程不镜像原教程正文，而是以原书 13 章主题为骨架，重写为 GoGopher Arch 风格的基础训练营，并为每章配一个可运行的 Go sandbox 练习。

## 目标

- 新增独立的 **Go 基础训练营** 产品入口。
- 按《Go 语言圣经中文版》的 13 章主题组织完整学习路径。
- 每章提供重写导读、学习目标、核心概念、常见坑、原教程阅读链接和 sandbox 练习。
- 课程数据与现有实习任务数据解耦，避免把系统课程和职场任务混在同一模型中。
- 明确来源、改编方式和许可证说明，不复制原教程正文。

## 非目标

- 不镜像 `gopl-zh/gopl-zh.github.com` 的全文内容。
- 不在首版实现复杂在线编辑器、进度持久化、AI 批改或多文件项目练习。
- 不改造现有实习任务线的数据模型。
- 不实现完整自动判题系统；首版只做轻量输出匹配和人工 checklist。

## 信息架构

新增独立课程入口，不混入现有实习任务线。

新增路径：

- `/courses/go-basics`：Go 基础训练营总览页。
- `/courses/go-basics/:chapterSlug`：单章课程详情页。

现有路径保持不变：

- `/dashboard`
- `/missions/:slug`

入口位置：

- 首页新增一个明显入口，引导用户进入 Go 基础训练营。
- 顶层布局导航也应提供课程入口，避免训练营只能从首页进入。

课程路由使用明确的未找到状态。访问不存在的课程章节时展示课程级“未找到章节”页面，不自动 fallback 到第一章；访问非课程未知路径时保持应用现有路由行为，不在本阶段新增全站 404 体系。

## 课程结构

训练营共 13 章，对应原教程主题：

1. 入门
2. 程序结构
3. 基础数据类型
4. 复合数据类型
5. 函数
6. 方法
7. 接口
8. Goroutines 和 Channels
9. 基于共享变量的并发
10. 包和工具
11. 测试
12. 反射
13. 底层编程

每章包含：

- 学习目标
- 本章导读
- 核心概念
- 常见坑
- 原教程阅读链接
- sandbox 练习
- 验收 checklist
- 和后续实习任务的衔接说明

## 单章页面体验

单章页面采用“学习内容 + 练习面板”的结构。

### 章节头部

展示：

- 章序号
- 标题
- 预计时长
- 难度
- 来源说明
- 原文阅读链接

### 学习内容

每章正文由项目重写，强调知识点在 Go 后端实战中的用途。页面展示：

- 3-5 条学习目标
- 一段简短导读
- 4-6 个核心概念卡片
- 每个核心概念包含名称、解释和常见误区

### 练习区

每章包含一个可运行 Go 练习。练习区复用现有 `executeCode` API，向 sandbox 提交 Go 代码并展示运行结果。

练习区展示：

- 练习标题
- 任务说明
- starter code
- 运行按钮
- stdout
- stderr
- exit code
- expected output
- hints
- 验收 checklist

首版可以先运行固定 starter code，不引入复杂代码编辑器。若后续需要编辑体验，再单独引入编辑器能力。

## 数据模型

新增独立课程数据模型，不复用现有 `Mission` 类型。

建议类型：

```ts
type GoCourseChapter = {
  slug: string;
  order: number;
  title: string;
  sourcePath: string;
  sourceUrl: string;
  duration: string;
  difficulty: "入门" | "基础" | "进阶" | "高级";
  summary: string;
  goals: string[];
  concepts: {
    name: string;
    explanation: string;
    pitfall: string;
  }[];
  exercise: {
    title: string;
    prompt: string;
    starterCode: string;
    expectedOutput: string;
    outputMatch: "trimmed-exact" | "contains";
    hints: string[];
  };
  checklist: string[];
  nextMissionSlugs: string[];
};
```

字段说明：

- `slug`：章节路由标识。
- `order`：章节顺序。
- `sourcePath`：原教程对应章节路径，例如 `ch4/ch4.md`。
- `sourceUrl`：原教程可访问链接，按 `https://gopl-zh.github.io/{sourcePath}` 生成并可被页面直接渲染，协议和根域名固定为 `https://gopl-zh.github.io`。
- `summary`、`concepts`、`exercise`：项目内重写内容。
- `exercise.outputMatch`：首版只支持去除首尾空白后的精确匹配，或包含匹配；不做复杂判题。
- `nextMissionSlugs`：连接到现有实习任务线的相关任务，渲染前应过滤不存在的任务 slug，避免死链接。

课程数据实现时应导出全部章节数组、按 slug 查询的 helper，以及开发期校验函数。校验函数至少检查 13 章数量、slug 唯一、必填字段存在、`sourceUrl` 存在、`sourceUrl` 使用固定根域名、`nextMissionSlugs` 不生成死链接。

当某章没有有效的 `nextMissionSlugs` 时，页面显示“暂无绑定实习任务”的占位文案，而不是隐藏整个衔接区域。

## 文件组织

建议新增：

- `web/src/app/data/goBasicsCourse.ts`：13 章课程数据和 helper，例如 `getGoCourseChapterBySlug`。
- `web/src/app/pages/GoBasicsCourse.tsx`：训练营总览页，展示 13 章路径、完成状态占位和每章 CTA。
- `web/src/app/pages/GoBasicsChapter.tsx`：单章详情页，展示课程内容和练习。
- `web/src/app/components/CourseExercisePanel.tsx`：课程练习运行面板，复用现有 sandbox API。

需要修改：

- `web/src/app/routes.tsx`：增加课程路由。
- `web/src/app/pages/Landing.tsx`：新增课程入口。
- `web/src/app/components/Layout.tsx`：在顶层导航增加课程入口。
- `README.md`：更新课程来源与改编说明。

## 错误处理

### 未知章节 slug

访问不存在的章节时，显示“未找到章节”，并提供返回训练营总览的按钮。不自动 fallback 到第一章，避免错误链接被误认为有效。

### sandbox 调用失败

如果 Gateway 或 sandbox-engine 不可用，练习区显示明确错误：

- 无法连接到代码运行服务。
- 请确认本地 Gateway 和 Sandbox Engine 已启动。

课程阅读内容不受 sandbox 错误影响。

### 代码运行失败

展示 stdout、stderr 和 exit code。编译失败或运行失败本身是学习反馈，不隐藏为通用错误。

## 来源与许可证

课程采用“重写为主”：

- 不复制原教程正文。
- 每章保留原教程阅读链接，链接由 `sourceUrl` 提供。
- README 和课程总览页注明课程参考并改编自《Go 语言圣经中文版》。
- README 和课程总览页注明原项目为 `gopl-zh/gopl-zh.github.com`，并链接到原仓库。
- README 和课程总览页注明原教程授权信息：仓库 LICENSE 为 BSD 3-Clause；其附录 C 说明正文采用 CC-BY 3.0，代码遵循 Go 项目的 BSD 协议。实现时应链接到原仓库 LICENSE 和附录 C，不在本项目内复制大段许可证正文。
- 每章页面展示短来源文案：`本章为 GoGopher Arch 改编课程，参考《Go 语言圣经中文版》对应章节。`
- 练习代码使用本项目新场景，不直接搬运原教程示例。

## 验收标准

首版完成后应满足：

- 首页能进入 Go 基础训练营。
- `/courses/go-basics` 展示完整 13 章路径。
- 每章详情页展示学习目标、重写导读、核心概念、原文链接、练习和验收 checklist。
- 每章练习都能调用 sandbox。
- sandbox 成功、编译失败、服务不可用都有明确反馈。
- README 包含来源、改编方式和许可证说明，并链接到原仓库 LICENSE 与附录 C。
- 每章数据包含可直接渲染的 `sourceUrl`。
- 不复制原教程正文。
- TypeScript 构建通过。
- 人工验证至少覆盖第 1、4、11 章练习。

## 测试策略

### 静态验证

- TypeScript 类型检查通过。
- 13 章课程数据字段完整。
- 每章 `slug` 唯一。
- 每章都有 `sourcePath` 和 `sourceUrl`。
- `sourceUrl` 能直接打开到原教程对应章节。
- `nextMissionSlugs` 不生成死链接。

### 功能验证

- `/courses/go-basics` 可访问。
- 至少 3 个章节详情页可访问。
- “运行代码”按钮能触发 sandbox 请求。
- 成功、编译失败、服务不可用三种状态都有可读反馈。
- 练习输出匹配按 `exercise.outputMatch` 执行：`trimmed-exact` 去除首尾空白后精确匹配，`contains` 检查 stdout 是否包含期望片段。
- 不存在的章节 slug 展示课程级“未找到章节”状态，并提供返回训练营总览的入口。

### 人工 UI 验证

启动前端后实际浏览：

- 首页入口。
- 训练营总览页。
- 第 1 章详情页。
- 第 4 章详情页。
- 第 11 章详情页。

至少运行第 1、4、11 章练习，并检查窄屏布局不破坏阅读体验。

## 落地策略

虽然目标是完整 13 章，但实现可以分成三个内部步骤，一次交付完整首版：

1. 课程数据层
   - 建立 13 章课程数据。
   - 每章包含标题、目标、导读、概念、练习、链接和 checklist。
2. 页面和路由层
   - 新增课程总览页和章节详情页。
   - 首页增加课程入口。
   - 练习面板复用现有 sandbox API。
3. 验收和来源说明
   - README 更新来源与改编说明。
   - 运行类型检查和构建。
   - 启动前端进行人工 UI 验证。

## 后续扩展

首版之后可以逐步加入：

- 在线编辑器。
- 进度持久化。
- 自动判题。
- AI 导师反馈。
- 每章多个练习。
- 与实习任务线的解锁关系。
