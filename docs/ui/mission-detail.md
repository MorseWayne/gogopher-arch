# 任务详情页 `/missions/:slug`

> Figma Make 设计规格 · 状态：待实现
> 设计系统：沿用 web/ 现有设计（Tailwind v4, shadcn/ui, `#00ADD8` 品牌蓝）

---

## 页面定位

学习平台的核心入口。用户从路线图或 Dashboard 侧边栏进入，查看任务背景、目标、验收标准，然后开始挑战。

---

## 布局 (Desktop, ≥1024px)

```
┌─────────────────────────────────────────────────────┐
│  ← 返回路线图                    [← 上一关] [下一关 →] │
├─────────────────────────────────────────────────────┤
│  🏷️ Day 1 · 入职第一周                    📍 进行中    │
│                                                       │
│  修复 Slice 内存泄露                                   │
│  ⏱ 约 30 分钟   ⭐ 初级   📋 前置: Go 基础             │
├────────────────────────┬────────────────────────────┤
│                         │                            │
│  📖 任务背景             │                            │
│                         │    📝 验收标准              │
│  🎯 任务目标             │    ☐ 修复 processData     │
│                         │    ☐ Goroutine 正确退出    │
│  💡 关键提示             │    ☐ 无内存泄露            │
│                         │    ☐ 通过所有测试           │
│  📚 前置知识             │                            │
│                         │    [    开始挑战    ]       │
│                         │                            │
├────────────────────────┴────────────────────────────┤
│                                                       │
│  🧪 初始代码                                          │
│  ┌──────────────────────────────────────────────────┐ │
│  │  1  package main                                 │ │
│  │  2                                               │ │
│  │  3  func main() {                                │ │
│  │  4      // TODO: 修复这里的代码                    │ │
│  │  5  }                                            │ │
│  └──────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────┘
```

---

## 响应式行为

| 断点 | 布局 |
|------|------|
| Desktop (≥1024px) | 左右两栏（60% / 40%），代码预览在下方 |
| Tablet (768-1023px) | 单栏，验收面板移到内容下方 |
| Mobile (<768px) | 单栏全宽，顶部导航精简为 `←` + 标题 |

---

## 逐区域规格

### 1. 顶部导航条

| 元素 | 样式 | 行为 |
|------|------|------|
| `← 返回路线图` | `text-neutral-400 hover:text-white`，左对齐 | 导航至 `/roadmap` |
| `← 上一关` | 右对齐，不可用时 `text-neutral-700 cursor-not-allowed` | 导航至上一条任务 |
| `下一关 →` | 右对齐，不可用时 `text-neutral-700 cursor-not-allowed` | 导航至下一条任务 |

### 2. 标签行

| 元素 | 组件 | 变体 |
|------|------|------|
| 章节标签 | `Badge` | `bg-neutral-800 text-neutral-300`，如 "Day 1 · 入职第一周" |
| 状态标签 | `Badge` | 进行中: `bg-yellow-500/10 text-yellow-400`；已完成: `bg-green-500/10 text-green-400`；未解锁: `bg-neutral-800 text-neutral-500` |

### 3. 标题与元信息

| 字段 | 样式 |
|------|------|
| 任务标题 | `text-3xl md:text-4xl font-bold text-white` |
| 预计耗时 | 图标 + `text-neutral-400` |
| 难度 | 图标 + `text-neutral-400` |
| 前置要求 | 图标 + `text-neutral-400`，可点击跳转至对应知识点 |

### 4. 左侧内容区 — Accordion

使用 `Accordion` 组件，默认展开「任务背景」和「任务目标」，其余折叠。

| 面板 | 默认 | 内容格式 |
|------|------|----------|
| 📖 **任务背景** | 展开 | 故事化叙述（2-3 段），描述 Bug 工单场景 |
| 🎯 **任务目标** | 展开 | 无序列表，`CheckCircle2` 图标，说明学习目标 |
| 💡 **关键提示** | 折叠 | 黄色警告卡片（`border-yellow-500/30 bg-yellow-500/5`），提示坑点但不给答案 |
| 📚 **前置知识** | 折叠 | 链接列表，每项可跳转 |

### 5. 右侧面板 — 验收标准卡片

固定卡片 `border border-neutral-800 rounded-2xl p-6 bg-neutral-900`。

| 元素 | 说明 |
|------|------|
| 标题 | `text-lg font-bold` "验收标准" |
| 验收项 | 4-5 条 `text-sm text-neutral-300`，前缀图标：未完成 `Square`，已完成 `CheckSquare` |
| 分隔 | `Separator` |
| CTA 按钮 | 全宽，`py-4 rounded-xl font-bold` |
| 按钮文案 | 按状态变体（见下方状态表） |

### 6. 底部初始代码预览

| 元素 | 说明 |
|------|------|
| 标题 | `text-sm font-semibold text-neutral-400` "🧪 初始代码" |
| 代码容器 | `bg-[#0d0d0d] border border-neutral-800 rounded-xl p-6 overflow-x-auto` |
| 代码 | 只读语法高亮，`font-mono text-sm leading-relaxed` |
| 配色 | 关键字 `text-pink-500`，字符串 `text-green-400`，函数 `text-blue-400`，注释 `text-neutral-500 italic`（与 Dashboard 代码区一致） |
| 行号 | 左侧 `text-neutral-600 text-right select-none pr-4 border-r border-neutral-800` |

---

## 四种状态变体

同一页面根据用户进度渲染不同状态。

| 状态 | CTA 按钮 | 验收标准 | 代码区 | 状态标签 |
|------|----------|----------|--------|----------|
| **未解锁** 🔒 | `bg-neutral-800 text-neutral-500 cursor-not-allowed` "🔒 完成前置任务解锁" | 全部未勾 | `blur-sm` 模糊遮罩 | `text-neutral-500` |
| **可开始** 🟢 | `bg-[#00ADD8] text-neutral-950 hover:bg-[#00ADD8]/90` "开始挑战" | 全部未勾 | 正常显示 | `text-neutral-400` |
| **进行中** 🟡 | `bg-[#00ADD8] text-neutral-950` "继续挑战" | 部分勾选 | 正常显示 | `text-yellow-400` |
| **已完成** ✅ | `border border-white/20 text-white hover:bg-white/5` "查看复盘" | 全部勾选 | 正常显示 | `text-green-400` |

---

## 交互行为

| 触发 | 行为 |
|------|------|
| 点击 "返回路线图" | 导航到 `/roadmap` |
| 点击 "上一关" / "下一关" | 导航到相邻任务详情 |
| 点击前置知识链接 | 外部跳转新窗口；内部待定 |
| 点击 "开始挑战" / "继续挑战" | 导航到 `/dashboard?mission=:slug`，加载该任务代码 |
| 点击 "查看复盘" | 导航到 `/missions/:slug/review` |
| Accordion 展开/收起 | 组件自带动画 |

---

## 数据模型 (概念)

```
任务 {
  slug:        string        // "slice-memory-leak"
  day:         number        // 1
  chapter:     string        // "入职第一周"
  title:       string        // "修复 Slice 内存泄露"
  duration:    string        // "约 30 分钟"
  difficulty:  "beginner" | "intermediate" | "advanced"
  prerequisite: string[]     // ["Go 基础"]
  status:      "locked" | "available" | "in-progress" | "completed"
  background:  string[]      // 3 段故事化叙述
  objectives:  string[]      // 学习目标
  hints:       string[]      // 关键提示
  knowledge:   { label: string, url?: string }[]
  criteria:    string[]      // 验收标准（4-5 条）
  starterCode: string        // 初始代码（Go 源码）
}
```

---

## 使用的现有组件

| 组件 | 路径 | 用途 |
|------|------|------|
| `Layout` | `src/app/components/Layout.tsx` | 页面外壳（导航栏+页脚） |
| `Badge` | `src/app/components/ui/badge.tsx` | 章节标签、状态标签 |
| `Button` | `src/app/components/ui/button.tsx` | CTA、导航按钮 |
| `Card` | `src/app/components/ui/card.tsx` | 验收面板 |
| `Accordion` | `src/app/components/ui/accordion.tsx` | 左侧内容折叠 |
| `Separator` | `src/app/components/ui/separator.tsx` | 面板内分隔 |
| `Progress` | `src/app/components/ui/progress.tsx` | 可选：页面 mini 进度条 |

---

## 新建文件

```
web/src/app/pages/MissionDetail.tsx
```

路由注册在 `routes.tsx` 中新增：
```tsx
{ path: "missions/:slug", Component: MissionDetail }
```
