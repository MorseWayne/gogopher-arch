# GoGopher Arch 项目定位重构实施计划

> **给 agentic workers：** 实施本计划时必须使用 `superpowers:subagent-driven-development`（推荐）或 `superpowers:executing-plans`。每个步骤使用 checkbox（`- [ ]`）跟踪执行状态。

**目标：** 将项目从“架构师进阶”叙事落地为“Go 后端实习成长平台”，并在文档和前端首屏中呈现 Go 基础、实习任务线、Go 工程进阶、AI 时代全栈路线四阶段成长路径。

**架构：** 本轮不改后端服务协议，优先统一产品叙事和前端第一屏体验。文档侧更新 README、旧设计文档、旧实施计划；前端侧把当前 Runtime 指标演示改成静态任务卡 + Monaco 编辑器 + 任务反馈的实习生工作台，继续复用现有 `/api/v1/execute` 沙盒执行接口。

**技术栈：** Markdown、React、TypeScript、Vite、Monaco Editor、Axios、Lucide React、Go 沙盒接口。

---

## 范围检查

规格文档覆盖了多个长期方向，但本实施计划只落地第一阶段：

- README 和旧文档定位统一
- 前端首屏改为实习生工作台
- 保留 AI 时代全栈路线为路线图，不实现 RAG、Agent 或 AI 导师后端
- 保留现有 gateway、sandbox-engine 和 `/api/v1/execute` 接口

不在本轮实施：

- 新增后端任务模型
- 新增数据库
- 新增 AI API 调用
- 新增 RAG 或 Agent 课程内容
- 新增端到端浏览器测试框架

## 文件结构

本计划会修改或创建以下文件：

- 修改 `README.md`：项目门面，从“架构师进化之路”改成“Go 后端实习成长平台”。
- 修改 `docs/specs/2026-03-13-gogopher-arch-design.md`：旧设计文档改为新定位下的设计基线。
- 修改 `docs/plans/2026-03-13-implementation-plan.md`：旧实施计划改为第一阶段可执行路线。
- 修改 `web/src/App.tsx`：前端主体验从 Runtime 指标演示改为实习生工作台。
- 修改 `web/src/App.css`：移除 Vite 模板样式，承载工作台布局和控件样式。
- 修改 `web/src/index.css`：移除模板居中容器和单色紫色主题，设置全屏应用基础样式。

## Task 1: 更新 README 产品门面

**Files:**
- Modify: `README.md`

- [ ] **Step 1: 确认 README 仍包含旧定位词**

Run:

```bash
rg -n "架构师进化之路|双十一|高并发 IM|AI CTO|10W QPS|区块链" README.md
```

Expected: 输出包含旧定位词所在行。

- [ ] **Step 2: 用下面内容替换 `README.md`**

```markdown
# GoGopher Arch: Go 后端实习成长平台

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/Frontend-React-61DAFB?style=flat&logo=react)](https://reactjs.org)
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
| 前端 | React, TypeScript, Monaco Editor |
| 沙盒 | Docker, `os/exec`, 执行超时控制 |
| 反馈 | 编译结果、控制台输出、任务检查、导师提示 |
| AI 路线 | LLM API、RAG、Agent、结构化输出、评测与安全 |

---

## 路线图

### 第一阶段：Go 后端实习生入职第一周

- [x] 项目定位重构规格确认
- [ ] README、设计文档和实施计划统一为新定位
- [ ] 前端首屏改为实习生工作台
- [ ] Day 0：Go 基础自检和第一次沙盒运行
- [ ] Day 1：修复 slice、map 和指针相关 Bug
- [ ] Day 2：补全一个 HTTP API handler
- [ ] Day 3：增加参数校验和错误处理
- [ ] Day 4：编写表驱动测试
- [ ] Day 5：修复一个简单并发问题或 context 超时问题

### 第二阶段：Go 工程能力进阶

- [ ] 数据库和事务任务
- [ ] 缓存和并发任务
- [ ] 日志、配置和可观测性任务
- [ ] 部署和服务可靠性任务

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

前端默认运行在 `http://localhost:5173`，Gateway 默认运行在 `http://localhost:8080`。

---

## 开源协议

本项目采用 [MIT License](LICENSE) 协议。
```

- [ ] **Step 3: 验证 README 新旧定位**

Run:

```bash
rg -n "Go 后端实习成长平台|AI 时代全栈|实习生任务线" README.md
```

Expected: 输出至少 3 行，包含新定位关键词。

Run:

```bash
rg -n "架构师进化之路|双十一|高并发 IM|10W QPS|AI CTO" README.md
```

Expected: 无输出，命令退出码为 1。

- [ ] **Step 4: 提交 README 修改**

```bash
git add README.md
git commit -m "docs: update readme positioning"
```

## Task 2: 更新旧设计文档

**Files:**
- Modify: `docs/specs/2026-03-13-gogopher-arch-design.md`

- [ ] **Step 1: 用下面内容替换旧设计文档**

```markdown
# GoGopher Arch: Go 后端实习成长平台设计文档

- 日期：2026-03-13
- 更新日期：2026-04-26
- 项目名：GoGopher Arch
- 版本：v1.1.0
- 目标：通过虚拟职场任务，帮助 Go 学习者从基础入门成长到具备 Go 后端实习岗位能力，并进一步走向 Go 工程能力和 AI 应用工程能力。

---

## 1. 项目愿景

GoGopher Arch 希望打破“先看完课程再做项目”的学习方式，把 Go 基础、后端工程实践和 AI 时代全栈能力放进连续的职场任务中。

用户不是一开始就扮演架构师，而是先从一个准备入职或刚入职的 Go 后端实习生开始：读任务、理解验收标准、修 Bug、补接口、写测试、看日志、处理评审意见。随着能力提升，用户再进入 Go 工程能力进阶和 AI 应用工程路线。

## 2. 用户成长路径

| 阶段 | 用户状态 | 核心目标 | 主要内容 |
| :--- | :--- | :--- | :--- |
| 入职前训练营 | Go 初学者 | 掌握完成任务所需的 Go 基础 | 语法、类型、函数、结构体、接口、错误处理、slice、map、指针、defer、HTTP 和 JSON |
| 实习生任务线 | 准实习生 / 实习生 | 具备 Go 后端实习岗位的基本工作能力 | 任务卡、Bug 修复、HTTP handler、参数校验、错误处理、表驱动测试、日志、简单并发和 context |
| Go 工程能力进阶 | Go 工程师 | 建立更成熟的工程判断 | 数据库、事务、缓存、并发设计、性能分析、部署、可观测性、服务可靠性 |
| AI 时代全栈路线 | 进阶工程师 | 将 Go 工程能力与 AI 应用工程结合 | LLM API、Prompt、结构化输出、工具调用、RAG、Agent、流式响应、评测、安全和成本控制 |

## 3. 核心学习模式

平台采用“实习模拟器 + 项目制学习 + 知识地图”的组合模式。

### 3.1 实习模拟器

用户以 Go 后端实习生身份加入一个虚拟团队。平台通过任务卡提供背景、目标、验收标准和关联知识点。用户需要像真实工作一样完成代码修改并通过检查。

### 3.2 项目制学习

任务不是彼此孤立的练习，而是围绕同一个后端项目持续推进。每个任务都新增、修复、测试或改进项目的一部分，让用户逐步积累工程上下文。

### 3.3 知识地图

知识地图整理任务中出现过的概念，帮助用户复习薄弱点。它不是第一屏主体验，而是辅助查漏补缺的结构化索引。

## 4. 核心学习循环

1. **收到任务卡**：用户看到职场化背景、目标、验收标准和关联知识点。
2. **任务前小课**：平台提供 5-10 分钟的知识补给，只讲完成当前任务所需的概念。
3. **动手编码**：用户在 Monaco 编辑器中修改 Go 代码。
4. **运行沙盒**：平台通过沙盒运行代码、测试或任务检查。
5. **获得反馈**：平台展示编译结果、测试结果、任务检查、控制台输出和导师提示。
6. **任务后复盘**：平台总结本次任务涉及的知识点、真实工作场景和常见追问。
7. **更新成长记录**：平台记录用户完成的任务和已覆盖知识点。

## 5. MVP 范围

第一版 MVP 聚焦“Go 后端实习生：入职第一周”。

推荐任务线：

1. Day 0：Go 基础自检和第一次沙盒运行
2. Day 1：修复 slice、map 和指针相关 Bug
3. Day 2：补全一个 HTTP API handler
4. Day 3：增加参数校验和错误处理
5. Day 4：编写表驱动测试
6. Day 5：修复一个简单并发问题或 context 超时问题

第一版不实现云原生、区块链、大规模压测、高级 Runtime 可视化、完整 AI 导师、完整 RAG 课程或完整 Agent 课程。

## 6. 技术方案架构

### 6.1 Gateway

- 技术栈：Go
- 职责：提供前端访问入口，接收代码执行请求，并转发给沙盒引擎。

### 6.2 Sandbox Engine

- 技术栈：Go、Docker、`os/exec`
- 职责：隔离执行用户提交的 Go 代码，提供执行超时和输出捕获能力。

### 6.3 Intern Workbench

- 技术栈：React、TypeScript、Monaco Editor
- 职责：提供任务卡、编辑器、运行按钮、任务反馈和控制台输出。

### 6.4 Mentor Feedback

- 初期职责：基于运行结果和常见 Go 错误提供静态导师提示。
- 进阶职责：接入 LLM、RAG 和 Agent 能力，形成代码审查、知识问答和任务规划反馈。

## 7. 成功衡量标准

- Go 初学者能理解自己如何开始学习。
- 准实习生能通过任务线练习 Go 后端实习工作方式。
- 用户能在浏览器中完成至少一条可运行、可反馈的实习任务。
- 文档、路线图和前端首屏不再把高级架构主题作为第一承诺。
- 项目长期路线能清楚表达 Go 工程能力和 AI 应用工程能力的结合。
```

- [ ] **Step 2: 验证旧设计文档定位**

Run:

```bash
rg -n "Go 后端实习成长平台|AI 时代全栈路线|实习模拟器|入职第一周" docs/specs/2026-03-13-gogopher-arch-design.md
```

Expected: 输出包含新设计关键词。

Run:

```bash
rg -n "双十一|高性能 IM|去中心化交易系统|10W QPS" docs/specs/2026-03-13-gogopher-arch-design.md
```

Expected: 无输出，命令退出码为 1。

- [ ] **Step 3: 提交设计文档修改**

```bash
git add docs/specs/2026-03-13-gogopher-arch-design.md
git commit -m "docs: refresh design spec positioning"
```

## Task 3: 更新旧实施计划

**Files:**
- Modify: `docs/plans/2026-03-13-implementation-plan.md`

- [ ] **Step 1: 用下面内容替换旧实施计划**

```markdown
# GoGopher Arch: Go 后端实习成长平台实施计划

- 版本：v1.1.0
- 日期：2026-03-13
- 更新日期：2026-04-26
- 目标：优先交付“Go 后端实习生入职第一周”的可运行学习闭环。

---

## 1. 当前阶段目标

第一阶段不追求完整架构训练，也不提前实现复杂 AI 能力。目标是让用户打开产品后，立即理解自己正在完成 Go 后端实习任务，并能通过浏览器编辑、运行和查看反馈。

第一阶段完成后，项目应该具备：

- 清晰的 Go 后端实习成长定位
- 一条入职第一周任务线
- 前端实习生工作台
- 可复用的沙盒执行能力
- 面向后续任务检查和导师提示的反馈结构

## 2. 技术选型

### 2.1 后端

- 语言：Go 1.22+
- Gateway：接收前端请求并转发到沙盒引擎
- Sandbox Engine：隔离运行用户代码，返回 stdout、stderr、状态、耗时和退出码
- 容器：Docker 和 Docker Compose

### 2.2 前端

- 框架：React + TypeScript
- 构建：Vite
- 编辑器：Monaco Editor
- 请求：Axios
- 图标：Lucide React

### 2.3 AI 路线

第一阶段只在路线图中表达 AI 时代全栈方向，不接入真实模型。

后续 AI 能力包括：

- LLM API 调用
- Prompt 设计
- 结构化输出
- 工具调用
- RAG 知识库问答
- Agent 任务规划和工具使用
- 评测、安全和成本控制

## 3. 第一阶段路线图

### Task 1：统一项目文档定位

- 更新 README 首屏定位。
- 更新旧设计文档，替换“架构师进阶”叙事。
- 更新旧实施计划，聚焦实习生任务线。

### Task 2：前端首屏改为实习生工作台

- 顶部显示产品名、当前阶段和运行按钮。
- 左侧显示任务卡、验收标准和任务前小课。
- 中间保留 Monaco 编辑器。
- 右侧显示任务反馈、导师提示和控制台输出。
- 默认代码使用 nil map 写入这一类基础实习任务。

### Task 3：完善 Day 0 到 Day 1 任务内容

- Day 0：Go 基础自检和第一次沙盒运行。
- Day 1：修复 slice、map 和指针相关 Bug。
- 每个任务包含任务背景、验收标准、关联知识点和复盘内容。

### Task 4：扩展沙盒反馈结构

- 保留 stdout、stderr、status、duration、exit_code。
- 前端基于执行状态推导编译、运行和任务检查状态。
- 后续再把任务检查从前端推导迁移到后端。

## 4. 第二阶段路线图

- Day 2：补全 HTTP API handler。
- Day 3：增加参数校验和错误处理。
- Day 4：编写表驱动测试。
- Day 5：修复简单并发问题或 context 超时问题。
- 增加任务后复盘页面。
- 增加成长记录。

## 5. 第三阶段路线图

- 数据库和事务任务。
- 缓存和并发任务。
- 日志、配置和可观测性任务。
- 部署和服务可靠性任务。
- LLM API、RAG、Agent 和 AI 产品工程任务。

## 6. 验证方式

每次改动至少执行：

```bash
git status --short
```

前端改动执行：

```bash
cd web
npm install
npm run build
```

文档定位改动执行：

```bash
rg -n "Go 后端实习成长平台|实习生工作台|AI 时代全栈" README.md docs
rg -n "双十一|10W QPS|去中心化交易系统|高性能 IM" README.md docs
```

第二条命令应无输出，表示第一屏定位不再使用旧的高级架构承诺。
```

- [ ] **Step 2: 验证旧实施计划定位**

Run:

```bash
rg -n "Go 后端实习成长平台|实习生工作台|AI 时代全栈" docs/plans/2026-03-13-implementation-plan.md
```

Expected: 输出包含新计划关键词。

Run:

```bash
rg -n "高性能 IM|全球排行榜|AI CTO|压测 QPS" docs/plans/2026-03-13-implementation-plan.md
```

Expected: 无输出，命令退出码为 1。

- [ ] **Step 3: 提交实施计划修改**

```bash
git add docs/plans/2026-03-13-implementation-plan.md
git commit -m "docs: refocus implementation plan"
```

## Task 4: 前端改为实习生工作台

**Files:**
- Modify: `web/src/App.tsx`
- Modify: `web/src/App.css`
- Modify: `web/src/index.css`

- [ ] **Step 1: 确认前端依赖可安装**

Run:

```bash
cd web
npm install
```

Expected: 安装完成，`node_modules` 可用。不要提交 `node_modules`。

- [ ] **Step 2: 用下面内容替换 `web/src/App.tsx`**

```tsx
import { useMemo, useState } from 'react';
import Editor from '@monaco-editor/react';
import axios from 'axios';
import {
  AlertCircle,
  BookOpen,
  CheckCircle2,
  ClipboardCheck,
  Code2,
  GraduationCap,
  Play,
  Terminal,
} from 'lucide-react';
import './App.css';

interface SandboxResponse {
  stdout: string;
  stderr: string;
  status: string;
  duration: number;
  exit_code: number;
}

type FeedbackState = 'idle' | 'pass' | 'fail';

interface FeedbackItem {
  label: string;
  detail: string;
  state: FeedbackState;
}

const DEFAULT_CODE = `package main

import "fmt"

type User struct {
\tName  string
\tScore int
}

func buildScoreMap(users []User) map[string]int {
\tvar scores map[string]int
\tfor _, user := range users {
\t\tscores[user.Name] = user.Score
\t}
\treturn scores
}

func main() {
\tusers := []User{
\t\t{Name: "Ming", Score: 86},
\t\t{Name: "Yan", Score: 91},
\t}

\tscores := buildScoreMap(users)
\tfmt.Println("Ming 的分数:", scores["Ming"])
}
`;

const taskCriteria = [
  '程序可以成功运行，不再出现 nil map 写入 panic。',
  'buildScoreMap 返回包含所有用户分数的 map。',
  '不要修改 main 函数里的输入数据和输出语句。',
];

const lessonPoints = [
  'map 在写入前必须完成初始化。',
  'var scores map[string]int 声明的是 nil map，只能读，不能写。',
  'make(map[string]int, len(users)) 可以创建可写 map，并预留容量。',
];

const mentorHints = [
  '先定位 panic 行，再判断这个变量是否已经初始化。',
  '这类问题在实习任务里很常见：看起来类型对了，但零值不能直接写入。',
  '修复后再运行一次，确认 stdout 中出现 Ming 的分数。',
];

function getFeedback(output: SandboxResponse | null, error: string | null): FeedbackItem[] {
  if (error) {
    return [
      {
        label: '连接 Gateway',
        detail: '前端无法连接到本地 Gateway，请确认后端服务已启动。',
        state: 'fail',
      },
      {
        label: '运行结果',
        detail: '等待 Gateway 恢复后重新运行。',
        state: 'idle',
      },
      {
        label: '任务检查',
        detail: '任务检查需要基于沙盒运行结果判断。',
        state: 'idle',
      },
    ];
  }

  if (!output) {
    return [
      {
        label: '连接 Gateway',
        detail: '等待第一次运行。',
        state: 'idle',
      },
      {
        label: '运行结果',
        detail: '点击运行代码后查看 stdout 和 stderr。',
        state: 'idle',
      },
      {
        label: '任务检查',
        detail: '修复 nil map 后，程序应输出 Ming 的分数。',
        state: 'idle',
      },
    ];
  }

  const succeeded = output.status === 'success' && output.exit_code === 0;
  const hasExpectedOutput = output.stdout.includes('Ming 的分数:');

  return [
    {
      label: '连接 Gateway',
      detail: '已收到沙盒执行结果。',
      state: 'pass',
    },
    {
      label: '运行结果',
      detail: succeeded ? '程序正常退出。' : '程序未正常退出，请查看 stderr。',
      state: succeeded ? 'pass' : 'fail',
    },
    {
      label: '任务检查',
      detail: hasExpectedOutput ? '已输出目标用户分数。' : '还没有看到预期输出。',
      state: hasExpectedOutput ? 'pass' : 'fail',
    },
  ];
}

function formatDuration(duration: number): string {
  if (duration <= 0) {
    return '--';
  }

  return `${(duration / 1_000_000).toFixed(2)}ms`;
}

function App() {
  const [code, setCode] = useState(DEFAULT_CODE);
  const [output, setOutput] = useState<SandboxResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const feedback = useMemo(() => getFeedback(output, error), [output, error]);

  const handleRun = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await axios.post<SandboxResponse>('http://localhost:8080/api/v1/execute', {
        id: `task-${Date.now()}`,
        code,
        language: 'go',
        timeout: 5,
      });
      setOutput(response.data);
    } catch (err: unknown) {
      if (axios.isAxiosError(err)) {
        const message =
          typeof err.response?.data === 'string'
            ? err.response.data
            : err.message || '无法连接到 Gateway 服务';
        setError(message);
      } else if (err instanceof Error) {
        setError(err.message);
      } else {
        setError('无法连接到 Gateway 服务');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="app-shell">
      <header className="topbar">
        <div className="brand">
          <div className="brand-icon">
            <GraduationCap size={22} />
          </div>
          <div>
            <p className="eyebrow">Go 后端实习生 · 入职第一周</p>
            <h1>GoGopher Arch</h1>
          </div>
        </div>
        <button className="run-button" onClick={handleRun} disabled={loading}>
          <Play size={17} fill="currentColor" />
          {loading ? '运行中' : '运行代码'}
        </button>
      </header>

      <main className="workbench">
        <aside className="task-panel" aria-label="任务卡">
          <section className="panel-section hero-section">
            <div className="section-title">
              <ClipboardCheck size={16} />
              <span>任务卡</span>
            </div>
            <h2>Day 1：修复 nil map 写入</h2>
            <p>
              你的导师把一个用户分数统计函数交给你。当前代码会在运行时 panic，
              请定位原因并完成修复。
            </p>
          </section>

          <section className="panel-section">
            <div className="section-title">
              <CheckCircle2 size={16} />
              <span>验收标准</span>
            </div>
            <ul className="check-list">
              {taskCriteria.map((item) => (
                <li key={item}>{item}</li>
              ))}
            </ul>
          </section>

          <section className="panel-section">
            <div className="section-title">
              <BookOpen size={16} />
              <span>任务前小课</span>
            </div>
            <ul className="lesson-list">
              {lessonPoints.map((item) => (
                <li key={item}>{item}</li>
              ))}
            </ul>
          </section>
        </aside>

        <section className="editor-panel" aria-label="代码编辑器">
          <div className="panel-toolbar">
            <div className="section-title">
              <Code2 size={16} />
              <span>main.go</span>
            </div>
            <span className="file-badge">Go 基础 Bug 修复</span>
          </div>
          <Editor
            height="100%"
            theme="vs-dark"
            defaultLanguage="go"
            value={code}
            onChange={(value) => setCode(value || '')}
            options={{
              fontSize: 14,
              minimap: { enabled: false },
              padding: { top: 18 },
              fontFamily: "'JetBrains Mono', 'Fira Code', ui-monospace, monospace",
              scrollBeyondLastLine: false,
            }}
          />
        </section>

        <aside className="feedback-panel" aria-label="任务反馈">
          <section className="panel-section">
            <div className="section-title">
              <ClipboardCheck size={16} />
              <span>任务反馈</span>
            </div>
            <div className="feedback-list">
              {feedback.map((item) => (
                <div className={`feedback-item ${item.state}`} key={item.label}>
                  <span className="feedback-dot" />
                  <div>
                    <strong>{item.label}</strong>
                    <p>{item.detail}</p>
                  </div>
                </div>
              ))}
            </div>
          </section>

          <section className="panel-section">
            <div className="section-title">
              <AlertCircle size={16} />
              <span>导师提示</span>
            </div>
            <ul className="hint-list">
              {mentorHints.map((hint) => (
                <li key={hint}>{hint}</li>
              ))}
            </ul>
          </section>

          <section className="console-section">
            <div className="console-header">
              <div className="section-title">
                <Terminal size={16} />
                <span>控制台</span>
              </div>
              <span>{output ? formatDuration(output.duration) : '--'}</span>
            </div>
            <div className="console-body">
              {error && <pre className="console-error">{error}</pre>}
              {output ? (
                <>
                  {output.stdout && <pre>{output.stdout}</pre>}
                  {output.stderr && <pre className="console-error">{output.stderr}</pre>}
                  <p className="console-meta">
                    退出码：{output.exit_code} · 状态：{output.status.toUpperCase()}
                  </p>
                </>
              ) : (
                <p className="console-placeholder">点击运行代码，查看沙盒输出。</p>
              )}
            </div>
          </section>
        </aside>
      </main>
    </div>
  );
}

export default App;
```

- [ ] **Step 3: 用下面内容替换 `web/src/App.css`**

```css
.app-shell {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #101417;
  color: #dce5e2;
}

.topbar {
  min-height: 68px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 24px;
  padding: 12px 20px;
  border-bottom: 1px solid #253038;
  background: #151b1f;
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.brand-icon {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  border-radius: 8px;
  background: #1f8a70;
  color: #f8fffb;
}

.eyebrow {
  margin: 0 0 2px;
  font-size: 12px;
  color: #8fb7aa;
}

.brand h1 {
  margin: 0;
  font-size: 20px;
  line-height: 1.1;
  font-weight: 700;
  color: #f8fffb;
}

.run-button {
  height: 40px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 0 16px;
  border: 0;
  border-radius: 6px;
  background: #f0b429;
  color: #15110a;
  font-weight: 700;
  cursor: pointer;
}

.run-button:disabled {
  cursor: wait;
  opacity: 0.62;
}

.workbench {
  flex: 1;
  min-height: 0;
  display: grid;
  grid-template-columns: minmax(280px, 340px) minmax(420px, 1fr) minmax(300px, 380px);
}

.task-panel,
.feedback-panel {
  min-width: 0;
  overflow-y: auto;
  background: #151b1f;
}

.task-panel {
  border-right: 1px solid #253038;
}

.feedback-panel {
  border-left: 1px solid #253038;
}

.editor-panel {
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: #0d1114;
}

.panel-toolbar,
.console-header {
  height: 42px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0 14px;
  border-bottom: 1px solid #253038;
  background: #1a2227;
}

.panel-section {
  padding: 18px;
  border-bottom: 1px solid #253038;
}

.hero-section {
  background: #18211f;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #8fb7aa;
  font-size: 12px;
  font-weight: 700;
}

.panel-section h2 {
  margin: 12px 0 10px;
  font-size: 22px;
  line-height: 1.2;
  color: #f8fffb;
}

.panel-section p {
  margin: 0;
  color: #b6c5c0;
}

.check-list,
.lesson-list,
.hint-list {
  margin: 12px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 10px;
}

.check-list li,
.lesson-list li,
.hint-list li {
  padding: 10px 12px;
  border-radius: 8px;
  background: #10171a;
  border: 1px solid #253038;
  color: #dce5e2;
}

.file-badge {
  white-space: nowrap;
  border-radius: 999px;
  padding: 3px 9px;
  color: #101417;
  background: #8fb7aa;
  font-size: 12px;
  font-weight: 700;
}

.feedback-list {
  display: grid;
  gap: 10px;
  margin-top: 12px;
}

.feedback-item {
  display: grid;
  grid-template-columns: 10px 1fr;
  gap: 10px;
  align-items: start;
  padding: 12px;
  border: 1px solid #253038;
  border-radius: 8px;
  background: #10171a;
}

.feedback-dot {
  width: 10px;
  height: 10px;
  margin-top: 5px;
  border-radius: 999px;
  background: #66737a;
}

.feedback-item.pass .feedback-dot {
  background: #42d392;
}

.feedback-item.fail .feedback-dot {
  background: #ff6b6b;
}

.feedback-item strong {
  display: block;
  margin-bottom: 3px;
  color: #f8fffb;
}

.feedback-item p {
  margin: 0;
  color: #aebbb7;
}

.console-section {
  display: flex;
  min-height: 260px;
  flex-direction: column;
}

.console-body {
  flex: 1;
  min-height: 180px;
  overflow: auto;
  padding: 14px;
  background: #070a0c;
  color: #94f7b2;
}

.console-body pre {
  margin: 0 0 10px;
  white-space: pre-wrap;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
  line-height: 1.5;
}

.console-error {
  color: #ff9b9b;
}

.console-meta,
.console-placeholder {
  margin: 0;
  color: #7b878b;
  font-size: 13px;
}

@media (max-width: 1120px) {
  .workbench {
    grid-template-columns: 300px 1fr;
  }

  .feedback-panel {
    grid-column: 1 / -1;
    border-left: 0;
    border-top: 1px solid #253038;
  }
}

@media (max-width: 760px) {
  .topbar {
    align-items: stretch;
    flex-direction: column;
  }

  .run-button {
    width: 100%;
  }

  .workbench {
    grid-template-columns: 1fr;
  }

  .task-panel,
  .feedback-panel {
    border: 0;
    border-bottom: 1px solid #253038;
  }

  .editor-panel {
    min-height: 520px;
  }
}
```

- [ ] **Step 4: 用下面内容替换 `web/src/index.css`**

```css
:root {
  font-family:
    Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  color: #dce5e2;
  background: #101417;
  font-synthesis: none;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

* {
  box-sizing: border-box;
}

html,
body,
#root {
  width: 100%;
  min-width: 320px;
  min-height: 100%;
  margin: 0;
}

body {
  min-height: 100vh;
}

button,
input,
textarea {
  font: inherit;
}
```

- [ ] **Step 5: 验证 TypeScript 构建**

Run:

```bash
cd web
npm run build
```

Expected: `tsc -b && vite build` 成功，命令退出码为 0。

- [ ] **Step 6: 验证旧前端叙事已移除**

Run:

```bash
rg -n "Goroutine 泄露|实时指标|架构师|CTO|双十一|10W QPS" web/src/App.tsx web/src/App.css web/src/index.css
```

Expected: 无输出，命令退出码为 1。

Run:

```bash
rg -n "任务卡|验收标准|任务前小课|导师提示|Go 后端实习生" web/src/App.tsx web/src/App.css web/src/index.css
```

Expected: 输出包含新工作台关键词。

- [ ] **Step 7: 提交前端工作台修改**

```bash
git add web/src/App.tsx web/src/App.css web/src/index.css
git commit -m "feat: present intern workbench"
```

## Task 5: 全量验证和收尾

**Files:**
- Verify: `README.md`
- Verify: `docs/specs/2026-03-13-gogopher-arch-design.md`
- Verify: `docs/plans/2026-03-13-implementation-plan.md`
- Verify: `web/src/App.tsx`
- Verify: `web/src/App.css`
- Verify: `web/src/index.css`

- [ ] **Step 1: 验证工作区状态**

Run:

```bash
git status --short
```

Expected: 如果前面每个任务都已提交，此命令无输出。

- [ ] **Step 2: 验证新定位覆盖 README 和 docs**

Run:

```bash
rg -n "Go 后端实习成长平台|实习生工作台|AI 时代全栈|RAG|Agent" README.md docs
```

Expected: 输出覆盖 README、旧设计文档、旧实施计划和中文定位规格。

- [ ] **Step 3: 验证第一屏旧承诺不再出现**

Run:

```bash
rg -n "架构师进化之路|双十一|高性能 IM|去中心化交易系统|10W QPS" README.md docs/specs/2026-03-13-gogopher-arch-design.md docs/plans/2026-03-13-implementation-plan.md web/src
```

Expected: 无输出，命令退出码为 1。

- [ ] **Step 4: 验证前端构建**

Run:

```bash
cd web
npm run build
```

Expected: 构建成功，命令退出码为 0。

- [ ] **Step 5: 查看提交历史**

Run:

```bash
git log --oneline -5
```

Expected: 最近提交包含：

```text
feat: present intern workbench
docs: refocus implementation plan
docs: refresh design spec positioning
docs: update readme positioning
docs: localize positioning redesign spec
```

## 自检记录

### 规格覆盖

- 中文定位规格中的新定位由 Task 1、Task 2、Task 3 覆盖。
- 用户分层由 Task 1 和 Task 2 覆盖。
- 核心学习循环由 Task 2 和 Task 4 覆盖。
- MVP 入职第一周由 Task 1、Task 2、Task 3 和 Task 4 覆盖。
- AI 时代全栈路线由 Task 1、Task 2 和 Task 3 作为路线图覆盖。

### 占位符扫描

占位符扫描已通过。所有修改步骤都给出明确文件路径、替换内容、验证命令和期望结果。

### 类型一致性

前端计划中定义的 `SandboxResponse` 字段与当前 `web/src/App.tsx` 使用的字段一致：`stdout`、`stderr`、`status`、`duration`、`exit_code`。`getFeedback`、`formatDuration` 和 `FeedbackItem` 都在同一个任务中定义并使用。
