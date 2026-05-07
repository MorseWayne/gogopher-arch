# 前端 UI 布局深入优化设计文档

- 日期：2026-04-28
- 项目名：GoGopher Arch
- 版本：v1.2.0
- 目标：在保持现有三栏工作台骨架的前提下，通过组件拆分、响应式增强、视觉系统化和交互精细化，将前端布局从"可用"提升到"好用"。

---

## 1. 现状分析

当前前端采用 React + TypeScript + Vite 技术栈，核心依赖 `@monaco-editor/react` 提供代码编辑能力。

### 1.1 当前结构

- `App.tsx`（~288 行）：包含全部 UI 逻辑和状态管理
- `App.css`（~406 行）：包含全部样式，无模块化
- `index.css`：全局 reset + 基础变量

### 1.2 主要问题

1. **文件过大**：`App.tsx` 和 `App.css` 承担过多职责，难以维护
2. **响应式不足**：仅 2 个断点（1120px / 760px），中小屏幕体验粗糙
3. **缺乏交互**：面板宽度固定，无法折叠，运行状态无视觉反馈
4. **视觉不统一**：间距、字体大小零散，无系统化 design token
5. **编辑器区域简陋**：无文件标签栏、状态栏信息不足

---

## 2. 设计方案

### 2.1 策略选择：渐进式精细化

保持三栏骨架（任务面板 + 编辑器 + 反馈面板），在现有结构上做深度打磨。相比 IDE 式重构风险更低，用户心智模型不变，为后续演进保留空间。

### 2.2 组件架构

#### 目录结构

```
web/src/
├── main.tsx
├── App.tsx                  # orchestrator，~60 行，只负责核心状态和组合
├── index.css                # design tokens + CSS reset
│
├── components/
│   ├── TopBar/
│   │   ├── TopBar.tsx
│   │   ├── TaskProgress.tsx  # 横向 Day 进度条
│   │   └── TopBar.module.css
│   ├── TaskPanel/
│   │   ├── TaskPanel.tsx     # 任务详情面板（背景/目标/标准/小课）
│   │   ├── TaskContent.tsx   # 当前任务的完整阅读内容
│   │   └── TaskPanel.module.css
│   ├── EditorPanel/
│   │   ├── EditorPanel.tsx
│   │   ├── EditorToolbar.tsx
│   │   └── EditorPanel.module.css
│   ├── FeedbackPanel/
│   │   ├── FeedbackPanel.tsx
│   │   ├── FeedbackList.tsx
│   │   ├── Console.tsx
│   │   └── FeedbackPanel.module.css
│   ├── ResizableSplit/
│   │   ├── ResizableSplit.tsx
│   │   └── ResizableSplit.module.css
│   └── common/
│       └── SectionTitle.tsx
│
├── hooks/
│   └── useResizable.ts      # 拖拽调整宽度的核心 hook
│
├── types/
│   └── workbench.ts         # 面板相关的 TypeScript 类型
│
└── lib/
    └── formatDuration.ts    # 工具函数
```

#### 状态管理

- `App.tsx` 保留核心状态：`selectedTaskId`、`code`、`output`、`error`、`loading`、`taskResults`
- 通过 props 向下传递，不引入全局状态库（Zustand/Redux）
- 理由：当前状态简单，props drilling 层级不深（3 层），引入状态库反而增加复杂度

#### CSS Modules 选择理由

- 避免类名冲突，组件自包含
- Vite 开箱即用，零配置
- tree-shaking 友好
- 与现有工具链无缝集成

---

### 2.2 任务信息层级优化

#### 问题

当前左侧面板将"任务列表"（导航）放在顶部，纵向卡片占用了大量垂直空间。真正需要阅读的任务内容（背景、目标、验收标准、小课）被挤到下方，阅读顺序被打断。

#### 方案：横向进度条 + 精简任务详情面板

- **TopBar 下方增加 `TaskProgress` 组件**：横向 Day 进度条，紧凑不占高度
  - 显示所有 Day（Day 0 → Day 1 → Day 2...）
  - 每个 Day 显示完成状态（绿色勾选 / 当前黄色高亮 / 未开始灰色）
  - 点击可跳转对应任务
- **左侧面板（`TaskPanel`）完全留给阅读内容**：
  - 任务标题（`text-heading`）
  - 背景描述（`text-body`）
  - 目标（`color-accent` 高亮）
  - 验收标准
  - 任务前小课
  - （可选）关联知识点
- **不再在左栏重复显示纵向任务列表卡片**

#### 阅读流

```
[TopBar: 品牌 + 操作按钮]
[TaskProgress: Day 0 ✅ → Day 1 ▶ → Day 2 ○ → Day 3 ○]
┌─────────────────┬──────────────────────┬─────────────────┐
│ 任务详情         │ 代码编辑器            │ 任务反馈         │
│ - 背景           │                      │ - 检查项        │
│ - 目标           │                      │ - 导师提示      │
│ - 验收标准       │                      │ - 复盘          │
│ - 小课           │                      │ - 控制台        │
└─────────────────┴──────────────────────┴─────────────────┘
```

用户打开页面后的自然阅读顺序：
1. 扫一眼顶部进度条，知道自己在哪个阶段
2. 阅读左栏任务详情，理解要做什么
3. 在中间编辑器编码
4. 在右栏查看反馈

---

### 2.3 响应式策略

采用 4 个断点，从宽屏到移动端逐步简化：

| 断点 | 名称 | 布局 | 面板行为 |
|------|------|------|----------|
| `≥1440px` | Desktop Wide | 三栏（300px / 1fr / 340px） | 三栏全开，空间充裕 |
| `≥1120px` | Desktop | 三栏（minmax(260px, 300px) / 1fr / minmax(280px, 320px)） | 三栏自适应 |
| `≥960px` | Tablet | 两栏（260px / 1fr） | TaskPanel 固定，FeedbackPanel 可折叠为悬浮按钮或底部抽屉 |
| `<960px` | Mobile | 单栏 | 底部 Tab Bar 切换三个面板（任务 / 编辑 / 反馈） |

#### 面板折叠行为（Tablet）

- FeedbackPanel 默认收起，右侧边缘显示悬浮按钮
- 点击按钮展开为覆盖层（overlay）或底部抽屉（bottom sheet）
- 运行代码后自动展开 FeedbackPanel 显示结果

#### 移动端 Tab Bar（Mobile）

```
┌─────────────────────┐
│                     │
│    当前面板内容      │
│                     │
├─────────────────────┤
│  📋任务  ✏️编辑  💬反馈 │
└─────────────────────┘
```

---

### 2.4 视觉设计系统（Design Tokens）

#### Spacing（8px base）

| Token | 值 | 用途 |
|-------|------|------|
| `space-xs` | 4px | 图标与文字间距、紧凑内边距 |
| `space-sm` | 8px | 组件内部 gap、列表项间距 |
| `space-md` | 16px | 面板 padding、section 间距 |
| `space-lg` | 24px | 大模块间距 |
| `space-xl` | 32px | 页面级间距 |

#### Typography

| Token | 大小 | 字重 | 颜色 | 用途 |
|-------|------|------|------|------|
| `text-display` | 20px | 700 | `#f8fffb` | 品牌标题（GoGopher Arch） |
| `text-heading` | 18px | 700 | `#f8fffb` | 面板标题（Day 1: 修复 Bug） |
| `text-body` | 14px | 400 | `#dce5e2` | 正文内容 |
| `text-caption` | 12px | 400 | `#8fb7aa` | 辅助信息、标签 |
| `text-mono` | 13px | 400 | `#94f7b2` | 控制台输出、代码 |

#### Colors（语义化）

| Token | 值 | 用途 |
|-------|------|------|
| `color-bg-primary` | `#101417` | 页面背景 |
| `color-bg-surface` | `#151b1f` | 面板背景 |
| `color-bg-elevated` | `#1a2227` | 工具栏、标题栏 |
| `color-bg-editor` | `#0d1114` | 编辑器区域 |
| `color-bg-console` | `#070a0c` | 控制台 |
| `color-border-default` | `#253038` | 默认边框 |
| `color-border-active` | `#31414a` | 悬停/激活边框 |
| `color-accent` | `#f0b429` | 主色调（按钮、选中态） |
| `color-success` | `#42d392` | 通过状态 |
| `color-error` | `#ff6b6b` | 错误状态 |
| `color-text-primary` | `#f8fffb` | 主文本 |
| `color-text-body` | `#dce5e2` | 正文 |
| `color-text-muted` | `#8fb7aa` | 辅助文本 |

---

### 2.5 交互设计

#### 1. 拖拽调整面板宽度

- 使用自定义 Hook `useResizable` 实现
- 拖拽条位于面板之间，宽度 4px，鼠标悬停变为 `col-resize`
- 最小宽度限制：
  - TaskPanel: 200px
  - EditorPanel: 320px
  - FeedbackPanel: 240px
- 拖拽状态实时更新，释放后写入 `localStorage` 记住用户偏好

#### 2. 面板折叠按钮

- 每个面板顶部工具栏增加折叠按钮（`‹` / `›`）
- 折叠后面板收缩为 32px 宽的边缘条，显示图标
- 点击边缘条恢复面板
- 折叠状态写入 `localStorage`

#### 3. 运行状态反馈动画

| 状态 | 视觉表现 |
|------|----------|
| 运行中 | 按钮显示旋转 spinner + "运行中" 文字 |
| 成功 | 控制台输出淡入（fade-in 200ms），反馈项逐条高亮 |
| 失败 | 错误输出红色闪烁一次（flash），滚动到错误位置 |
| 通过 | 任务列表项边框变绿色，显示 checkmark 动画 |

---

### 2.6 编辑器区域优化

#### 工具栏增强

```
┌──────────────────────────────────────────────┐
│ [main.go] [main_test.go]          [slice] Go │
├──────────────────────────────────────────────┤
│                                              │
│           Monaco Editor Area                 │
│                                              │
└──────────────────────────────────────────────┘
```

- 增加文件标签栏（当前只读，为后续多文件支持预留）
- 右侧显示：track badge + 语言模式 + 编码信息
- 主题：保持 `vs-dark`，后续可选自定义主题匹配平台色调

#### Monaco Editor 配置优化

```typescript
options: {
  fontSize: 14,
  fontFamily: "'JetBrains Mono', 'Fira Code', ui-monospace, monospace",
  minimap: { enabled: true, scale: 1 },  // 开启小地图
  padding: { top: 18 },
  scrollBeyondLastLine: false,
  automaticLayout: true,                  // 自动适应容器大小
  tabSize: 2,
  insertSpaces: true,
}
```

---

## 3. 技术实现要点

### 3.1 useResizable Hook

```typescript
// 核心逻辑：监听 mousedown → mousemove → mouseup
// 计算 deltaX，应用到 panel width
// 限制在 min/max 范围内
// 可选：持久化到 localStorage
```

### 3.2 ResizableSplit 组件

- 接收 `left`、`center`、`right` 三个 children
- 管理两个拖拽条的位置
- 处理折叠/展开状态
- 内部使用 CSS Grid：`grid-template-columns: var(--left-width) 4px var(--center-width) 4px var(--right-width)`

### 3.3 CSS 变量策略

```css
:root {
  --space-xs: 4px;
  --space-sm: 8px;
  --space-md: 16px;
  --space-lg: 24px;
  --space-xl: 32px;

  --color-bg-primary: #101417;
  --color-bg-surface: #151b1f;
  /* ... 其他 token ... */
}
```

CSS Modules 中通过 `composes` 或直接引用变量。

### 3.4 移动端 Tab Bar

使用 CSS Grid + `position: fixed` 底部导航。通过 React state 控制当前激活面板，仅渲染一个面板内容。

---

## 4. 边界情况与错误处理

### 4.1 拖拽极端情况

- 窗口缩小时，自动调整面板宽度不低于最小值
- 如果总宽度不足以容纳三个面板最小宽度之和，优先保证 EditorPanel，两侧面板自动折叠

### 4.2 localStorage 损坏

- 读取 `localStorage` 时加 `try/catch`，失败则回退到默认值
- 不存储敏感信息

### 4.3 Monaco Editor 自适应

- 使用 `automaticLayout: true` 让 Monaco 自动适应容器
- 面板折叠/展开时，触发 `editor.layout()`

### 4.4 运行时错误

- 组件级 Error Boundary：防止单个面板崩溃导致整个应用不可用
- 建议实现一个简单的 `ErrorFallback` 组件

---

## 5. 测试策略

### 5.1 单元测试

- `useResizable`：测试拖拽逻辑、边界限制、持久化
- `formatDuration`：已有测试，保持不变
- 各 Panel 组件：测试渲染和交互（点击、折叠）

### 5.2 视觉回归

- 由于涉及大量 CSS 改动，建议在不同断点下截图对比
- 重点关注：任务列表选中态、反馈项颜色、按钮状态

### 5.3 交互测试

- 拖拽调整宽度后刷新页面，验证持久化
- 折叠面板后运行代码，验证自动展开
- 移动端 Tab 切换

---

## 6. 实现顺序建议

1. **基础准备**：创建目录结构、迁移 `formatDuration`、创建 types
2. **Design Tokens**：提取 CSS 变量到 `index.css`
3. **组件拆分**：按面板逐个拆分（TaskPanel → FeedbackPanel → EditorPanel → TopBar）
4. **响应式增强**：添加新断点、实现面板折叠
5. **拖拽交互**：实现 `useResizable` 和 `ResizableSplit`
6. **编辑器优化**：文件标签栏、状态栏、Monaco 配置
7. **动画与细节**：运行状态动画、错误闪烁、过渡效果
8. **测试与验证**：单元测试、视觉回归、多端检查

---

## 7. 风险与回退

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 组件拆分引入 bug | 中 | 逐组件迁移，每步验证功能 |
| 拖拽交互性能问题 | 低 | 使用 `requestAnimationFrame`，节流 mousemove |
| CSS Modules 与现有样式冲突 | 低 | 彻底删除 `App.css`，避免并存 |
| 移动端体验不佳 | 中 | 单独测试移动端，必要时简化功能 |

---

## 8. 成功标准

- `App.tsx` 行数从 288 行减少到 60 行以内
- `App.css` 完全删除，样式分散到各组件 CSS Modules
- 响应式支持 4 个断点，移动端可用
- 面板支持拖拽调整和折叠
- 运行状态有清晰的视觉反馈动画
- 任务信息层级优化完成：横向进度条在 TopBar 下方，左栏只展示任务阅读内容（背景/目标/验收标准/小课）
- 所有现有测试通过，新增组件有对应测试
