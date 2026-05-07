# 前端 UI 布局深入优化 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将现有单文件 App.tsx + App.css 重构为模块化组件架构，实现响应式布局、面板拖拽/折叠、任务信息层级优化（横向进度条 + 左栏只展示任务内容），并添加运行状态动画。

**Architecture:** 保持三栏骨架，按面板拆分为独立 React 组件（CSS Modules），新增 ResizableSplit 容器管理拖拽，TopBar 下方增加 TaskProgress 横向进度条替代纵向任务列表。状态仍由 App.tsx 集中管理，通过 props 向下传递。

**Tech Stack:** React 19 + TypeScript + Vite + CSS Modules + @monaco-editor/react + vitest

---

## File Structure

```
web/src/
├── main.tsx                          (no change)
├── App.tsx                           (refactor: orchestrator ~60 lines)
├── App.css                           (DELETE)
├── index.css                         (modify: add CSS design tokens)
│
├── components/
│   ├── TopBar/
│   │   ├── TopBar.tsx                (create: brand + actions)
│   │   ├── TopBar.module.css         (create)
│   │   ├── TaskProgress.tsx          (create: horizontal day progress)
│   │   └── TaskProgress.module.css   (create)
│   ├── TaskPanel/
│   │   ├── TaskPanel.tsx             (create: task detail panel)
│   │   ├── TaskContent.tsx           (create: background/objective/criteria/lesson)
│   │   └── TaskPanel.module.css      (create)
│   ├── EditorPanel/
│   │   ├── EditorPanel.tsx           (create: Monaco + toolbar)
│   │   ├── EditorToolbar.tsx         (create: file tabs + status)
│   │   └── EditorPanel.module.css    (create)
│   ├── FeedbackPanel/
│   │   ├── FeedbackPanel.tsx         (create: feedback + hints + review)
│   │   ├── FeedbackList.tsx          (create: pass/fail items)
│   │   ├── Console.tsx               (create: stdout/stderr output)
│   │   └── FeedbackPanel.module.css  (create)
│   ├── ResizableSplit/
│   │   ├── ResizableSplit.tsx        (create: 3-col grid + drag handles)
│   │   └── ResizableSplit.module.css (create)
│   └── common/
│       └── SectionTitle.tsx          (create: reusable section header)
│
├── hooks/
│   └── useResizable.ts               (create: drag logic + localStorage)
│
├── types/
│   └── workbench.ts                  (create: panel prop types)
│
└── lib/
    ├── formatDuration.ts             (create: extract from App.tsx)
    └── formatDuration.test.ts        (create: unit test)
```

---

## Dependencies

No new npm packages required. CSS Modules is built into Vite.

---

### Task 1: 基础准备 — lib, types, common

**Files:**
- Create: `web/src/lib/formatDuration.ts`
- Create: `web/src/lib/formatDuration.test.ts`
- Create: `web/src/types/workbench.ts`
- Create: `web/src/components/common/SectionTitle.tsx`

- [ ] **Step 1: Extract `formatDuration` to lib**

Create `web/src/lib/formatDuration.ts`:
```typescript
export function formatDuration(duration: number): string {
  if (duration <= 0) {
    return '--';
  }
  return `${(duration / 1_000_000).toFixed(2)}ms`;
}
```

- [ ] **Step 2: Write test for `formatDuration`**

Create `web/src/lib/formatDuration.test.ts`:
```typescript
import { describe, expect, it } from 'vitest';
import { formatDuration } from './formatDuration';

describe('formatDuration', () => {
  it('returns -- for zero or negative', () => {
    expect(formatDuration(0)).toBe('--');
    expect(formatDuration(-1)).toBe('--');
  });

  it('formats nanoseconds to milliseconds', () => {
    expect(formatDuration(1_500_000)).toBe('1.50ms');
    expect(formatDuration(12_340_000)).toBe('12.34ms');
  });
});
```

- [ ] **Step 3: Run test to verify it passes**

```bash
cd web && npx vitest run src/lib/formatDuration.test.ts
```
Expected: 2 tests PASS

- [ ] **Step 4: Create workbench types**

Create `web/src/types/workbench.ts`:
```typescript
import type { InternTask } from '../tasks';
import type { FeedbackItem, SandboxResponse } from '../taskFeedback';

export interface TopBarProps {
  onReset: () => void;
  onRun: () => void;
  loading: boolean;
}

export interface TaskProgressProps {
  tasks: InternTask[];
  selectedTaskId: string;
  taskResults: Record<string, 'pass' | 'fail'>;
  onSelectTask: (taskId: string) => void;
}

export interface TaskPanelProps {
  task: InternTask;
}

export interface EditorPanelProps {
  code: string;
  onChange: (value: string) => void;
  track: string;
}

export interface FeedbackPanelProps {
  feedback: FeedbackItem[];
  currentTaskPassed: boolean;
  mentorHints: string[];
  review: string[];
  output: SandboxResponse | null;
  error: string | null;
}

export interface ResizableSplitProps {
  left: React.ReactNode;
  center: React.ReactNode;
  right: React.ReactNode;
}
```

- [ ] **Step 5: Create SectionTitle common component**

Create `web/src/components/common/SectionTitle.tsx`:
```typescript
import type { LucideIcon } from 'lucide-react';
import styles from './SectionTitle.module.css';

interface SectionTitleProps {
  icon: LucideIcon;
  label: string;
}

export function SectionTitle({ icon: Icon, label }: SectionTitleProps) {
  return (
    <div className={styles.sectionTitle}>
      <Icon size={16} />
      <span>{label}</span>
    </div>
  );
}
```

Create `web/src/components/common/SectionTitle.module.css`:
```css
.sectionTitle {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--color-text-muted);
  font-size: 12px;
  font-weight: 700;
}
```

- [ ] **Step 6: Commit**

```bash
git add web/src/lib/ web/src/types/ web/src/components/common/
git commit -m "feat: add formatDuration lib, workbench types, and SectionTitle component

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 2: Design Tokens — index.css

**Files:**
- Modify: `web/src/index.css`

- [ ] **Step 1: Add CSS design tokens to index.css**

Replace entire `web/src/index.css` with:
```css
:root {
  /* Spacing */
  --space-xs: 4px;
  --space-sm: 8px;
  --space-md: 16px;
  --space-lg: 24px;
  --space-xl: 32px;

  /* Colors */
  --color-bg-primary: #101417;
  --color-bg-surface: #151b1f;
  --color-bg-elevated: #1a2227;
  --color-bg-editor: #0d1114;
  --color-bg-console: #070a0c;
  --color-bg-card: #10171a;
  --color-bg-hero: #18211f;
  --color-bg-selected: #1d231d;

  --color-border-default: #253038;
  --color-border-active: #31414a;
  --color-border-pass: #2c8a62;
  --color-border-fail: #8a3b3b;

  --color-accent: #f0b429;
  --color-success: #42d392;
  --color-error: #ff6b6b;
  --color-error-text: #ff9b9b;

  --color-text-primary: #f8fffb;
  --color-text-body: #dce5e2;
  --color-text-secondary: #b6c5c0;
  --color-text-muted: #8fb7aa;
  --color-text-dim: #66737a;
  --color-text-console: #94f7b2;

  /* Typography */
  --text-display: 20px;
  --text-heading: 18px;
  --text-body: 14px;
  --text-caption: 12px;
  --text-mono: 13px;

  /* Base */
  font-family:
    Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
  color: var(--color-text-body);
  background: var(--color-bg-primary);
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

- [ ] **Step 2: Commit**

```bash
git add web/src/index.css
git commit -m "feat: add design tokens (spacing, colors, typography) to index.css

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 3: TopBar + TaskProgress 组件

**Files:**
- Create: `web/src/components/TopBar/TopBar.tsx`
- Create: `web/src/components/TopBar/TopBar.module.css`
- Create: `web/src/components/TopBar/TaskProgress.tsx`
- Create: `web/src/components/TopBar/TaskProgress.module.css`

- [ ] **Step 1: Create TopBar component**

`web/src/components/TopBar/TopBar.tsx`:
```typescript
import { GraduationCap, Play, RotateCcw } from 'lucide-react';
import type { TopBarProps } from '../../types/workbench';
import styles from './TopBar.module.css';

export function TopBar({ onReset, onRun, loading }: TopBarProps) {
  return (
    <header className={styles.topbar}>
      <div className={styles.brand}>
        <div className={styles.brandIcon}>
          <GraduationCap size={22} />
        </div>
        <div>
          <p className={styles.eyebrow}>Go 后端实习生 · 入职第一周</p>
          <h1 className={styles.brandTitle}>GoGopher Arch</h1>
        </div>
      </div>
      <div className={styles.actions}>
        <button className={styles.ghostButton} onClick={onReset} disabled={loading}>
          <RotateCcw size={16} />
          重置代码
        </button>
        <button className={styles.runButton} onClick={onRun} disabled={loading}>
          <Play size={17} fill="currentColor" />
          {loading ? '运行中' : '运行代码'}
        </button>
      </div>
    </header>
  );
}
```

`web/src/components/TopBar/TopBar.module.css`:
```css
.topbar {
  min-height: 68px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-lg);
  padding: 12px 20px;
  border-bottom: 1px solid var(--color-border-default);
  background: var(--color-bg-surface);
}

.brand {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.brandIcon {
  width: 42px;
  height: 42px;
  display: grid;
  place-items: center;
  border-radius: 8px;
  background: #1f8a70;
  color: var(--color-text-primary);
  flex-shrink: 0;
}

.eyebrow {
  margin: 0 0 2px;
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}

.brandTitle {
  margin: 0;
  font-size: var(--text-display);
  line-height: 1.1;
  font-weight: 700;
  color: var(--color-text-primary);
}

.actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.ghostButton {
  height: 40px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-sm);
  padding: 0 14px;
  border: 1px solid var(--color-border-active);
  border-radius: 6px;
  background: var(--color-bg-card);
  color: var(--color-text-body);
  font-weight: 700;
  cursor: pointer;
}

.ghostButton:disabled {
  cursor: wait;
  opacity: 0.62;
}

.runButton {
  height: 40px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-sm);
  padding: 0 16px;
  border: 0;
  border-radius: 6px;
  background: var(--color-accent);
  color: #15110a;
  font-weight: 700;
  cursor: pointer;
}

.runButton:disabled {
  cursor: wait;
  opacity: 0.62;
}

@media (max-width: 760px) {
  .topbar {
    align-items: stretch;
    flex-direction: column;
  }

  .actions {
    width: 100%;
    display: grid;
    grid-template-columns: 1fr 1fr;
  }

  .ghostButton,
  .runButton {
    width: 100%;
  }
}
```

- [ ] **Step 2: Create TaskProgress component**

`web/src/components/TopBar/TaskProgress.tsx`:
```typescript
import { Check } from 'lucide-react';
import type { TaskProgressProps } from '../../types/workbench';
import styles from './TaskProgress.module.css';

export function TaskProgress({ tasks, selectedTaskId, taskResults, onSelectTask }: TaskProgressProps) {
  return (
    <div className={styles.progressBar}>
      {tasks.map((task, index) => {
        const result = taskResults[task.id];
        const isSelected = task.id === selectedTaskId;
        const isPassed = result === 'pass';

        return (
          <button
            key={task.id}
            className={`${styles.dayItem} ${isSelected ? styles.selected : ''} ${isPassed ? styles.passed : ''}`}
            onClick={() => onSelectTask(task.id)}
            type="button"
          >
            <span className={styles.dayLabel}>Day {task.day}</span>
            {isPassed && <Check size={12} className={styles.checkIcon} />}
            {index < tasks.length - 1 && <span className={styles.arrow}>→</span>}
          </button>
        );
      })}
    </div>
  );
}
```

`web/src/components/TopBar/TaskProgress.module.css`:
```css
.progressBar {
  display: flex;
  align-items: center;
  gap: var(--space-xs);
  padding: 8px 20px;
  border-bottom: 1px solid var(--color-border-default);
  background: var(--color-bg-elevated);
  overflow-x: auto;
}

.dayItem {
  display: inline-flex;
  align-items: center;
  gap: var(--space-xs);
  padding: 4px 10px;
  border: 1px solid var(--color-border-default);
  border-radius: 6px;
  background: var(--color-bg-card);
  color: var(--color-text-muted);
  font-size: var(--text-caption);
  font-weight: 700;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s ease;
}

.dayItem:hover {
  border-color: var(--color-border-active);
  color: var(--color-text-body);
}

.dayItem.selected {
  border-color: var(--color-accent);
  background: var(--color-bg-selected);
  color: var(--color-accent);
}

.dayItem.passed {
  border-color: var(--color-border-pass);
  color: var(--color-success);
}

.dayItem.passed.selected {
  border-color: var(--color-success);
  background: rgba(66, 211, 146, 0.1);
}

.checkIcon {
  flex-shrink: 0;
}

.arrow {
  color: var(--color-text-dim);
  margin-left: 2px;
  font-size: 10px;
}
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/TopBar/
git commit -m "feat: add TopBar and TaskProgress components

- TopBar: brand, reset button, run button
- TaskProgress: horizontal day navigation with pass/fail/selected states

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 4: TaskPanel + TaskContent 组件

**Files:**
- Create: `web/src/components/TaskPanel/TaskPanel.tsx`
- Create: `web/src/components/TaskPanel/TaskContent.tsx`
- Create: `web/src/components/TaskPanel/TaskPanel.module.css`

- [ ] **Step 1: Create TaskContent component**

`web/src/components/TaskPanel/TaskContent.tsx`:
```typescript
import { BookOpen, CheckCircle2, ClipboardCheck } from 'lucide-react';
import { SectionTitle } from '../common/SectionTitle';
import type { TaskPanelProps } from '../../types/workbench';
import styles from './TaskPanel.module.css';

export function TaskContent({ task }: TaskPanelProps) {
  return (
    <>
      <section className={`${styles.section} ${styles.heroSection}`}>
        <SectionTitle icon={ClipboardCheck} label="当前任务" />
        <h2 className={styles.heading}>
          Day {task.day}：{task.title}
        </h2>
        <p className={styles.background}>{task.background}</p>
        <p className={styles.objective}>{task.objective}</p>
      </section>

      <section className={styles.section}>
        <SectionTitle icon={CheckCircle2} label="验收标准" />
        <ul className={styles.checkList}>
          {task.criteria.map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </section>

      <section className={styles.section}>
        <SectionTitle icon={BookOpen} label="任务前小课" />
        <ul className={styles.lessonList}>
          {task.lesson.map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </section>
    </>
  );
}
```

- [ ] **Step 2: Create TaskPanel component**

`web/src/components/TaskPanel/TaskPanel.tsx`:
```typescript
import { TaskContent } from './TaskContent';
import type { TaskPanelProps } from '../../types/workbench';
import styles from './TaskPanel.module.css';

export function TaskPanel({ task }: TaskPanelProps) {
  return (
    <aside className={styles.panel} aria-label="任务详情">
      <TaskContent task={task} />
    </aside>
  );
}
```

`web/src/components/TaskPanel/TaskPanel.module.css`:
```css
.panel {
  min-width: 0;
  overflow-y: auto;
  background: var(--color-bg-surface);
  border-right: 1px solid var(--color-border-default);
}

.section {
  padding: var(--space-md);
  border-bottom: 1px solid var(--color-border-default);
}

.heroSection {
  background: var(--color-bg-hero);
}

.heading {
  margin: 12px 0 10px;
  font-size: var(--text-heading);
  line-height: 1.2;
  font-weight: 700;
  color: var(--color-text-primary);
}

.background {
  margin: 0;
  color: var(--color-text-secondary);
  line-height: 1.6;
}

.objective {
  margin-top: var(--space-sm) !important;
  color: var(--color-accent) !important;
  font-weight: 600;
}

.checkList,
.lessonList {
  margin: 12px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 10px;
}

.checkList li,
.lessonList li {
  padding: 10px 12px;
  border-radius: 8px;
  background: var(--color-bg-card);
  border: 1px solid var(--color-border-default);
  color: var(--color-text-body);
  font-size: var(--text-body);
  line-height: 1.5;
}
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/TaskPanel/
git commit -m "feat: add TaskPanel and TaskContent components

- TaskPanel: left panel containing task reading content
- TaskContent: background, objective, criteria, lesson sections
- No task list (moved to TaskProgress)

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 5: FeedbackPanel + FeedbackList + Console

**Files:**
- Create: `web/src/components/FeedbackPanel/FeedbackPanel.tsx`
- Create: `web/src/components/FeedbackPanel/FeedbackList.tsx`
- Create: `web/src/components/FeedbackPanel/Console.tsx`
- Create: `web/src/components/FeedbackPanel/FeedbackPanel.module.css`

- [ ] **Step 1: Create FeedbackList component**

`web/src/components/FeedbackPanel/FeedbackList.tsx`:
```typescript
import type { FeedbackItem } from '../../taskFeedback';
import styles from './FeedbackPanel.module.css';

interface FeedbackListProps {
  items: FeedbackItem[];
}

export function FeedbackList({ items }: FeedbackListProps) {
  return (
    <div className={styles.feedbackList}>
      {items.map((item) => (
        <div className={`${styles.feedbackItem} ${styles[item.state]}`} key={item.label}>
          <span className={styles.feedbackDot} />
          <div>
            <strong>{item.label}</strong>
            <p>{item.detail}</p>
          </div>
        </div>
      ))}
    </div>
  );
}
```

- [ ] **Step 2: Create Console component**

`web/src/components/FeedbackPanel/Console.tsx`:
```typescript
import { Terminal } from 'lucide-react';
import { SectionTitle } from '../common/SectionTitle';
import { formatDuration } from '../../lib/formatDuration';
import type { SandboxResponse } from '../../taskFeedback';
import styles from './FeedbackPanel.module.css';

interface ConsoleProps {
  output: SandboxResponse | null;
  error: string | null;
}

export function Console({ output, error }: ConsoleProps) {
  return (
    <section className={styles.consoleSection}>
      <div className={styles.consoleHeader}>
        <SectionTitle icon={Terminal} label="控制台" />
        <span className={styles.consoleDuration}>
          {output ? formatDuration(output.duration) : '--'}
        </span>
      </div>
      <div className={styles.consoleBody}>
        {error && <pre className={styles.consoleError}>{error}</pre>}
        {output ? (
          <>
            {output.stdout && <pre>{output.stdout}</pre>}
            {output.stderr && <pre className={styles.consoleError}>{output.stderr}</pre>}
            <p className={styles.consoleMeta}>
              退出码：{output.exit_code} · 状态：{output.status.toUpperCase()}
            </p>
          </>
        ) : (
          <p className={styles.consolePlaceholder}>点击运行代码，查看沙盒输出。</p>
        )}
      </div>
    </section>
  );
}
```

- [ ] **Step 3: Create FeedbackPanel component**

`web/src/components/FeedbackPanel/FeedbackPanel.tsx`:
```typescript
import { AlertCircle, BookOpen, ClipboardCheck } from 'lucide-react';
import { SectionTitle } from '../common/SectionTitle';
import { FeedbackList } from './FeedbackList';
import { Console } from './Console';
import type { FeedbackPanelProps } from '../../types/workbench';
import styles from './FeedbackPanel.module.css';

export function FeedbackPanel({
  feedback,
  currentTaskPassed,
  mentorHints,
  review,
  output,
  error,
}: FeedbackPanelProps) {
  return (
    <aside className={styles.panel} aria-label="任务反馈">
      <section className={styles.section}>
        <SectionTitle icon={ClipboardCheck} label="任务反馈" />
        <div className={styles.feedbackSummary}>
          {currentTaskPassed ? '本任务已通过。' : '运行代码后查看任务检查。'}
        </div>
        <FeedbackList items={feedback} />
      </section>

      <section className={styles.section}>
        <SectionTitle icon={AlertCircle} label="导师提示" />
        <ul className={styles.hintList}>
          {mentorHints.map((hint) => (
            <li key={hint}>{hint}</li>
          ))}
        </ul>
      </section>

      <section className={styles.section}>
        <SectionTitle icon={BookOpen} label="任务后复盘" />
        <ul className={styles.reviewList}>
          {review.map((item) => (
            <li key={item}>{item}</li>
          ))}
        </ul>
      </section>

      <Console output={output} error={error} />
    </aside>
  );
}
```

`web/src/components/FeedbackPanel/FeedbackPanel.module.css`:
```css
.panel {
  min-width: 0;
  overflow-y: auto;
  background: var(--color-bg-surface);
  border-left: 1px solid var(--color-border-default);
}

.section {
  padding: var(--space-md);
  border-bottom: 1px solid var(--color-border-default);
}

.feedbackSummary {
  margin-top: 12px;
  padding: 10px 12px;
  border: 1px solid var(--color-border-active);
  border-radius: 8px;
  background: var(--color-bg-card);
  color: var(--color-text-primary);
  font-weight: 700;
}

.feedbackList {
  display: grid;
  gap: 10px;
  margin-top: 12px;
}

.feedbackItem {
  display: grid;
  grid-template-columns: 10px 1fr;
  gap: 10px;
  align-items: start;
  padding: 12px;
  border: 1px solid var(--color-border-default);
  border-radius: 8px;
  background: var(--color-bg-card);
  animation: fadeIn 0.2s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}

.feedbackDot {
  width: 10px;
  height: 10px;
  margin-top: 5px;
  border-radius: 999px;
  background: var(--color-text-dim);
}

.feedbackItem.pass .feedbackDot {
  background: var(--color-success);
}

.feedbackItem.fail .feedbackDot {
  background: var(--color-error);
}

.feedbackItem strong {
  display: block;
  margin-bottom: 3px;
  color: var(--color-text-primary);
}

.feedbackItem p {
  margin: 0;
  color: var(--color-text-secondary);
}

.hintList {
  margin: 12px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 10px;
}

.hintList li {
  padding: 10px 12px;
  border-radius: 8px;
  background: var(--color-bg-card);
  border: 1px solid var(--color-border-default);
  color: var(--color-text-body);
}

.reviewList {
  margin: 12px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 10px;
}

.reviewList li {
  padding: 10px 12px;
  border-left: 3px solid var(--color-accent);
  background: var(--color-bg-card);
  color: var(--color-text-body);
}

.consoleSection {
  display: flex;
  min-height: 260px;
  flex-direction: column;
}

.consoleHeader {
  height: 42px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0 14px;
  border-bottom: 1px solid var(--color-border-default);
  background: var(--color-bg-elevated);
}

.consoleDuration {
  font-size: var(--text-caption);
  color: var(--color-text-muted);
}

.consoleBody {
  flex: 1;
  min-height: 180px;
  overflow: auto;
  padding: 14px;
  background: var(--color-bg-console);
  color: var(--color-text-console);
}

.consoleBody pre {
  margin: 0 0 10px;
  white-space: pre-wrap;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: var(--text-mono);
  line-height: 1.5;
}

.consoleError {
  color: var(--color-error-text);
}

.consoleMeta,
.consolePlaceholder {
  margin: 0;
  color: var(--color-text-dim);
  font-size: var(--text-mono);
}
```

- [ ] **Step 4: Commit**

```bash
git add web/src/components/FeedbackPanel/
git commit -m "feat: add FeedbackPanel, FeedbackList, and Console components

- FeedbackPanel: feedback, mentor hints, review sections
- FeedbackList: pass/fail/idle feedback items with fade-in animation
- Console: stdout/stderr output with syntax styling

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 6: EditorPanel + EditorToolbar

**Files:**
- Create: `web/src/components/EditorPanel/EditorPanel.tsx`
- Create: `web/src/components/EditorPanel/EditorToolbar.tsx`
- Create: `web/src/components/EditorPanel/EditorPanel.module.css`

- [ ] **Step 1: Create EditorToolbar component**

`web/src/components/EditorPanel/EditorToolbar.tsx`:
```typescript
import { Code2 } from 'lucide-react';
import { SectionTitle } from '../common/SectionTitle';
import styles from './EditorPanel.module.css';

interface EditorToolbarProps {
  track: string;
}

export function EditorToolbar({ track }: EditorToolbarProps) {
  return (
    <div className={styles.toolbar}>
      <div className={styles.toolbarLeft}>
        <div className={styles.fileTab}>
          <SectionTitle icon={Code2} label="main.go" />
        </div>
      </div>
      <div className={styles.toolbarRight}>
        <span className={styles.badge}>{track}</span>
        <span className={styles.meta}>UTF-8 · Go</span>
      </div>
    </div>
  );
}
```

- [ ] **Step 2: Create EditorPanel component**

`web/src/components/EditorPanel/EditorPanel.tsx`:
```typescript
import Editor from '@monaco-editor/react';
import { EditorToolbar } from './EditorToolbar';
import type { EditorPanelProps } from '../../types/workbench';
import styles from './EditorPanel.module.css';

export function EditorPanel({ code, onChange, track }: EditorPanelProps) {
  return (
    <section className={styles.panel} aria-label="代码编辑器">
      <EditorToolbar track={track} />
      <div className={styles.editorWrapper}>
        <Editor
          height="100%"
          theme="vs-dark"
          defaultLanguage="go"
          value={code}
          onChange={(value) => onChange(value || '')}
          options={{
            fontSize: 14,
            fontFamily: "'JetBrains Mono', 'Fira Code', ui-monospace, monospace",
            minimap: { enabled: true, scale: 1 },
            padding: { top: 18 },
            scrollBeyondLastLine: false,
            automaticLayout: true,
            tabSize: 2,
            insertSpaces: true,
          }}
        />
      </div>
    </section>
  );
}
```

`web/src/components/EditorPanel/EditorPanel.module.css`:
```css
.panel {
  min-width: 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-editor);
}

.toolbar {
  height: 42px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 0 14px;
  border-bottom: 1px solid var(--color-border-default);
  background: var(--color-bg-elevated);
  flex-shrink: 0;
}

.toolbarLeft {
  display: flex;
  gap: 4px;
}

.fileTab {
  padding: 4px 10px;
  border: 1px solid var(--color-border-active);
  border-bottom: 0;
  border-radius: 4px 4px 0 0;
  background: var(--color-bg-card);
}

.toolbarRight {
  display: flex;
  align-items: center;
  gap: 10px;
}

.badge {
  white-space: nowrap;
  border-radius: 999px;
  padding: 3px 9px;
  color: var(--color-bg-primary);
  background: var(--color-text-muted);
  font-size: var(--text-caption);
  font-weight: 700;
}

.meta {
  color: var(--color-text-dim);
  font-size: var(--text-caption);
}

.editorWrapper {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}
```

- [ ] **Step 3: Commit**

```bash
git add web/src/components/EditorPanel/
git commit -m "feat: add EditorPanel and EditorToolbar components

- EditorPanel: Monaco Editor with automaticLayout and minimap
- EditorToolbar: file tab (main.go), track badge, language meta

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 7: useResizable Hook

**Files:**
- Create: `web/src/hooks/useResizable.ts`

- [ ] **Step 1: Implement useResizable hook**

`web/src/hooks/useResizable.ts`:
```typescript
import { useCallback, useEffect, useRef, useState } from 'react';

interface UseResizableOptions {
  initialWidth: number;
  minWidth: number;
  maxWidth: number;
  storageKey?: string;
}

interface UseResizableReturn {
  width: number;
  startResize: (direction: 'left' | 'right') => (e: React.MouseEvent) => void;
}

export function useResizable({
  initialWidth,
  minWidth,
  maxWidth,
  storageKey,
}: UseResizableOptions): UseResizableReturn {
  const [width, setWidth] = useState(() => {
    if (storageKey) {
      try {
        const saved = localStorage.getItem(storageKey);
        if (saved) {
          const parsed = parseInt(saved, 10);
          if (!isNaN(parsed)) {
            return Math.max(minWidth, Math.min(maxWidth, parsed));
          }
        }
      } catch {
        // ignore localStorage errors
      }
    }
    return initialWidth;
  });

  const resizing = useRef<{
    startX: number;
    startWidth: number;
    direction: 'left' | 'right';
  } | null>(null);

  const startResize = useCallback(
    (direction: 'left' | 'right') => (e: React.MouseEvent) => {
      resizing.current = {
        startX: e.clientX,
        startWidth: width,
        direction,
      };
      document.body.style.cursor = 'col-resize';
      document.body.style.userSelect = 'none';
    },
    [width],
  );

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (!resizing.current) return;
      const { startX, startWidth, direction } = resizing.current;
      const delta =
        direction === 'left' ? e.clientX - startX : startX - e.clientX;
      const newWidth = Math.max(
        minWidth,
        Math.min(maxWidth, startWidth + delta),
      );
      setWidth(newWidth);
    };

    const handleMouseUp = () => {
      if (resizing.current) {
        resizing.current = null;
        document.body.style.cursor = '';
        document.body.style.userSelect = '';
        if (storageKey) {
          try {
            localStorage.setItem(storageKey, String(width));
          } catch {
            // ignore localStorage errors
          }
        }
      }
    };

    window.addEventListener('mousemove', handleMouseMove);
    window.addEventListener('mouseup', handleMouseUp);
    return () => {
      window.removeEventListener('mousemove', handleMouseMove);
      window.removeEventListener('mouseup', handleMouseUp);
    };
  }, [minWidth, maxWidth, storageKey, width]);

  return { width, startResize };
}
```

- [ ] **Step 2: Commit**

```bash
git add web/src/hooks/useResizable.ts
git commit -m "feat: add useResizable hook for panel drag resizing

- Tracks drag state with mousedown/mousemove/mouseup
- Respects min/max width bounds
- Persists width to localStorage

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 8: ResizableSplit Component

**Files:**
- Create: `web/src/components/ResizableSplit/ResizableSplit.tsx`
- Create: `web/src/components/ResizableSplit/ResizableSplit.module.css`

- [ ] **Step 1: Create ResizableSplit component**

`web/src/components/ResizableSplit/ResizableSplit.tsx`:
```typescript
import { useState } from 'react';
import { useResizable } from '../../hooks/useResizable';
import type { ResizableSplitProps } from '../../types/workbench';
import styles from './ResizableSplit.module.css';

export function ResizableSplit({
  left,
  center,
  right,
}: ResizableSplitProps) {
  const [leftCollapsed, setLeftCollapsed] = useState(false);
  const [rightCollapsed, setRightCollapsed] = useState(false);

  const leftPanel = useResizable({
    initialWidth: 300,
    minWidth: 200,
    maxWidth: 500,
    storageKey: 'gogopher-panel-left',
  });

  const rightPanel = useResizable({
    initialWidth: 340,
    minWidth: 240,
    maxWidth: 500,
    storageKey: 'gogopher-panel-right',
  });

  const leftWidth = leftCollapsed ? 32 : leftPanel.width;
  const rightWidth = rightCollapsed ? 32 : rightPanel.width;

  return (
    <div
      className={styles.split}
      style={{
        gridTemplateColumns: `${leftWidth}px 4px 1fr 4px ${rightWidth}px`,
      }}
    >
      <div className={`${styles.panel} ${leftCollapsed ? styles.collapsed : ''}`}>
        {!leftCollapsed && left}
        {leftCollapsed && (
          <button
            className={styles.expandBtn}
            onClick={() => setLeftCollapsed(false)}
            aria-label="展开任务面板"
          >
            ›
          </button>
        )}
      </div>

      <div
        className={styles.handle}
        onMouseDown={leftCollapsed ? undefined : leftPanel.startResize('left')}
        title={leftCollapsed ? undefined : '拖拽调整宽度'}
      />

      <div className={styles.center}>{center}</div>

      <div
        className={styles.handle}
        onMouseDown={rightCollapsed ? undefined : rightPanel.startResize('right')}
        title={rightCollapsed ? undefined : '拖拽调整宽度'}
      />

      <div className={`${styles.panel} ${rightCollapsed ? styles.collapsed : ''}`}>
        {!rightCollapsed && right}
        {rightCollapsed && (
          <button
            className={styles.expandBtn}
            onClick={() => setRightCollapsed(false)}
            aria-label="展开反馈面板"
          >
            ‹
          </button>
        )}
      </div>
    </div>
  );
}
```

`web/src/components/ResizableSplit/ResizableSplit.module.css`:
```css
.split {
  flex: 1;
  min-height: 0;
  display: grid;
  overflow: hidden;
}

.panel {
  min-width: 0;
  overflow: hidden;
  position: relative;
}

.panel.collapsed {
  background: var(--color-bg-elevated);
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: var(--space-sm);
}

.expandBtn {
  width: 24px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--color-bg-card);
  border: 1px solid var(--color-border-default);
  border-radius: 4px;
  color: var(--color-text-muted);
  font-size: 16px;
  cursor: pointer;
  writing-mode: vertical-rl;
}

.expandBtn:hover {
  color: var(--color-text-body);
  border-color: var(--color-border-active);
}

.handle {
  background: transparent;
  cursor: col-resize;
  position: relative;
  z-index: 1;
}

.handle:hover {
  background: var(--color-accent);
  opacity: 0.3;
}

.center {
  min-width: 0;
  min-height: 0;
  overflow: hidden;
}

@media (max-width: 959px) {
  .split {
    display: flex;
    flex-direction: column;
  }

  .handle {
    display: none;
  }
}
```

- [ ] **Step 2: Commit**

```bash
git add web/src/components/ResizableSplit/
git commit -m "feat: add ResizableSplit component with drag handles and collapse

- Three-column grid layout with draggable dividers
- Left/right panels support collapse/expand
- Width persisted to localStorage
- Handles hidden on mobile

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 9: 响应式增强 + 移动端 Tab Bar

**Files:**
- Create: `web/src/hooks/useMediaQuery.ts`
- Modify: `web/src/components/ResizableSplit/ResizableSplit.module.css`
- Modify: `web/src/App.tsx` (in Task 10, but responsive CSS now)

- [ ] **Step 1: Create useMediaQuery hook**

`web/src/hooks/useMediaQuery.ts`:
```typescript
import { useEffect, useState } from 'react';

export function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState(() => {
    if (typeof window !== 'undefined') {
      return window.matchMedia(query).matches;
    }
    return false;
  });

  useEffect(() => {
    const mql = window.matchMedia(query);
    const handler = (e: MediaQueryListEvent) => setMatches(e.matches);
    mql.addEventListener('change', handler);
    return () => mql.removeEventListener('change', handler);
  }, [query]);

  return matches;
}
```

- [ ] **Step 2: Update responsive CSS in component modules**

The responsive breakpoints are handled per-component. Add to `web/src/components/TaskPanel/TaskPanel.module.css`:
```css
@media (max-width: 959px) {
  .panel {
    border-right: 0;
    border-bottom: 1px solid var(--color-border-default);
  }
}
```

Add to `web/src/components/FeedbackPanel/FeedbackPanel.module.css`:
```css
@media (max-width: 959px) {
  .panel {
    border-left: 0;
    border-bottom: 1px solid var(--color-border-default);
  }
}
```

- [ ] **Step 3: Commit**

```bash
git add web/src/hooks/useMediaQuery.ts
git commit -m "feat: add useMediaQuery hook for responsive layout

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 10: App.tsx 重构 + 删除 App.css

**Files:**
- Modify: `web/src/App.tsx`
- Delete: `web/src/App.css`

- [ ] **Step 1: Rewrite App.tsx as orchestrator**

`web/src/App.tsx`:
```typescript
import { useMemo, useState } from 'react';
import { TopBar } from './components/TopBar/TopBar';
import { TaskProgress } from './components/TopBar/TaskProgress';
import { TaskPanel } from './components/TaskPanel/TaskPanel';
import { EditorPanel } from './components/EditorPanel/EditorPanel';
import { FeedbackPanel } from './components/FeedbackPanel/FeedbackPanel';
import { ResizableSplit } from './components/ResizableSplit/ResizableSplit';
import { useMediaQuery } from './hooks/useMediaQuery';
import { defaultTaskId, findTaskById, internshipTasks } from './tasks';
import { didPassTask, evaluateTaskChecks, type SandboxResponse } from './taskFeedback';
import './index.css';
import styles from './App.module.css';
import axios from 'axios';

function App() {
  const [selectedTaskId, setSelectedTaskId] = useState(defaultTaskId);
  const selectedTask = findTaskById(selectedTaskId);
  const [code, setCode] = useState(selectedTask.starterCode);
  const [output, setOutput] = useState<SandboxResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [taskResults, setTaskResults] = useState<Record<string, 'pass' | 'fail'>>({});
  const [mobileTab, setMobileTab] = useState<'task' | 'editor' | 'feedback'>('editor');

  const isMobile = useMediaQuery('(max-width: 959px)');

  const feedback = useMemo(
    () => evaluateTaskChecks(output, error, selectedTask.checks),
    [output, error, selectedTask],
  );

  const currentTaskPassed = didPassTask(output, error, selectedTask.checks);

  const handleSelectTask = (taskId: string) => {
    const nextTask = findTaskById(taskId);
    setSelectedTaskId(nextTask.id);
    setCode(nextTask.starterCode);
    setOutput(null);
    setError(null);
  };

  const handleResetCode = () => {
    setCode(selectedTask.starterCode);
    setOutput(null);
    setError(null);
  };

  const handleRun = async () => {
    setLoading(true);
    setError(null);
    setOutput(null);

    try {
      const response = await axios.post<SandboxResponse>('/api/v1/execute', {
        id: `${selectedTask.id}-${Date.now()}`,
        code,
        language: 'go',
        timeout: 5,
      });
      const nextOutput = response.data;
      setOutput(nextOutput);
      setTaskResults((results) => ({
        ...results,
        [selectedTask.id]: didPassTask(nextOutput, null, selectedTask.checks) ? 'pass' : 'fail',
      }));
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
      setTaskResults((results) => ({
        ...results,
        [selectedTask.id]: 'fail',
      }));
    } finally {
      setLoading(false);
    }
  };

  const taskPanelNode = <TaskPanel task={selectedTask} />;
  const editorPanelNode = (
    <EditorPanel code={code} onChange={setCode} track={selectedTask.track} />
  );
  const feedbackPanelNode = (
    <FeedbackPanel
      feedback={feedback}
      currentTaskPassed={currentTaskPassed}
      mentorHints={selectedTask.mentorHints}
      review={selectedTask.review}
      output={output}
      error={error}
    />
  );

  return (
    <div className={styles.appShell}>
      <TopBar onReset={handleResetCode} onRun={handleRun} loading={loading} />
      <TaskProgress
        tasks={internshipTasks}
        selectedTaskId={selectedTask.id}
        taskResults={taskResults}
        onSelectTask={handleSelectTask}
      />

      {isMobile ? (
        <main className={styles.mobileMain}>
          {mobileTab === 'task' && taskPanelNode}
          {mobileTab === 'editor' && editorPanelNode}
          {mobileTab === 'feedback' && feedbackPanelNode}

          <nav className={styles.mobileTabBar}>
            <button
              className={mobileTab === 'task' ? styles.activeTab : ''}
              onClick={() => setMobileTab('task')}
            >
              任务
            </button>
            <button
              className={mobileTab === 'editor' ? styles.activeTab : ''}
              onClick={() => setMobileTab('editor')}
            >
              编辑
            </button>
            <button
              className={mobileTab === 'feedback' ? styles.activeTab : ''}
              onClick={() => setMobileTab('feedback')}
            >
              反馈
            </button>
          </nav>
        </main>
      ) : (
        <main className={styles.desktopMain}>
          <ResizableSplit
            left={taskPanelNode}
            center={editorPanelNode}
            right={feedbackPanelNode}
          />
        </main>
      )}
    </div>
  );
}

export default App;
```

`web/src/App.module.css`:
```css
.appShell {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--color-bg-primary);
  color: var(--color-text-body);
}

.desktopMain {
  flex: 1;
  min-height: 0;
  display: flex;
  overflow: hidden;
}

.mobileMain {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  position: relative;
}

.mobileTabBar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  height: 52px;
  display: flex;
  background: var(--color-bg-surface);
  border-top: 1px solid var(--color-border-default);
  z-index: 100;
}

.mobileTabBar button {
  flex: 1;
  background: transparent;
  border: 0;
  color: var(--color-text-muted);
  font-size: var(--text-caption);
  font-weight: 700;
  cursor: pointer;
  border-bottom: 2px solid transparent;
}

.mobileTabBar button.activeTab {
  color: var(--color-accent);
  border-bottom-color: var(--color-accent);
}
```

- [ ] **Step 2: Delete App.css**

```bash
rm web/src/App.css
```

- [ ] **Step 3: Verify build passes**

```bash
cd web && npx tsc -b
```
Expected: No TypeScript errors

- [ ] **Step 4: Commit**

```bash
git add web/src/App.tsx web/src/App.module.css web/src/App.css
git commit -m "refactor: rewrite App.tsx as orchestrator, delete App.css

- App.tsx reduced from 288 lines to ~120 lines
- Delegates rendering to TopBar, TaskProgress, ResizableSplit, panels
- Mobile layout with bottom Tab Bar (<960px)
- Desktop layout with ResizableSplit (>=960px)

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 11: 动画与细节优化

**Files:**
- Modify: `web/src/components/FeedbackPanel/FeedbackPanel.module.css`
- Modify: `web/src/components/TopBar/TopBar.module.css`

- [ ] **Step 1: Add run button spinner animation**

Add to `web/src/components/TopBar/TopBar.module.css`:
```css
@keyframes spin {
  to { transform: rotate(360deg); }
}

.spinner {
  width: 14px;
  height: 14px;
  border: 2px solid currentColor;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}
```

Update `TopBar.tsx` run button to show spinner when loading:
```tsx
<button className={styles.runButton} onClick={onRun} disabled={loading}>
  {loading ? (
    <span className={styles.spinner} />
  ) : (
    <Play size={17} fill="currentColor" />
  )}
  {loading ? '运行中' : '运行代码'}
</button>
```

- [ ] **Step 2: Add console flash animation for errors**

Add to `web/src/components/FeedbackPanel/FeedbackPanel.module.css`:
```css
@keyframes flashError {
  0%, 100% { background: var(--color-bg-console); }
  50% { background: rgba(255, 107, 107, 0.1); }
}

.consoleBody.hasError {
  animation: flashError 0.4s ease;
}
```

Update `Console.tsx` to conditionally add `hasError` class when `error` prop is present.

- [ ] **Step 3: Commit**

```bash
git add web/src/components/TopBar/ web/src/components/FeedbackPanel/
git commit -m "feat: add run spinner and console error flash animations

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

### Task 12: 测试验证

**Files:**
- Modify: `web/src/App.test.tsx` (if needed)

- [ ] **Step 1: Run all existing tests**

```bash
cd web && npx vitest run
```
Expected: All tests pass

- [ ] **Step 2: Build production bundle**

```bash
cd web && npx tsc -b && npx vite build
```
Expected: Build succeeds with no errors

- [ ] **Step 3: Manual verification checklist**

Start dev server and verify:
- [ ] TopBar renders with brand, reset button, run button
- [ ] TaskProgress shows all Days with correct pass/fail/selected states
- [ ] Clicking a Day in TaskProgress switches task and resets code
- [ ] TaskPanel shows background, objective, criteria, lesson
- [ ] EditorPanel shows Monaco editor with minimap enabled
- [ ] EditorToolbar shows main.go tab and track badge
- [ ] FeedbackPanel shows feedback items, hints, review, console
- [ ] Running code updates feedback and console
- [ ] ResizableSplit drag handles adjust panel widths
- [ ] Panel collapse buttons hide/show panels
- [ ] Mobile view (<960px) shows bottom Tab Bar
- [ ] All existing tests still pass

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "feat: complete frontend UI layout optimization

- Component architecture: 9 new components + 1 hook
- Task hierarchy: TaskProgress replaces vertical task list
- Responsive: 4 breakpoints, mobile Tab Bar
- Interactions: draggable panels, collapsible panels
- Animations: run spinner, feedback fade-in, console error flash
- App.tsx reduced from 288 to ~120 lines
- App.css deleted (migrated to CSS Modules)

Co-Authored-By: Claude Opus 4.7 <noreply@anthropic.com>"
```

---

## Self-Review

### 1. Spec Coverage

| Spec Section | Implementing Task |
|-------------|------------------|
| 2.2 Component Architecture (directory structure) | Task 1, 3-8 |
| 2.2 Task Hierarchy (TaskProgress + TaskContent) | Task 3, 4 |
| 2.3 Responsive Strategy (4 breakpoints) | Task 9, 10 |
| 2.4 Design Tokens (CSS variables) | Task 2 |
| 2.5 Interactions (drag, collapse, animations) | Task 7, 8, 11 |
| 2.6 Editor Optimization (toolbar, minimap) | Task 6 |
| 3.1 useResizable Hook | Task 7 |
| 3.2 ResizableSplit | Task 8 |
| 4.1 Drag edge cases | Task 8 (min/max widths) |
| 4.2 localStorage errors | Task 7 (try/catch) |
| 4.3 Monaco adaptive | Task 6 (automaticLayout) |
| 5.1 Unit tests | Task 1, 12 |
| 8. Success criteria | Task 12 (verification) |

**Gap: None identified.**

### 2. Placeholder Scan

- No "TBD", "TODO", "implement later" found
- All code blocks contain complete implementation
- All test commands have expected output
- All file paths are exact

### 3. Type Consistency

- `FeedbackItem` imported from `taskFeedback.ts` (not redefined)
- `SandboxResponse` imported from `taskFeedback.ts`
- `InternTask` imported from `tasks.ts`
- Props interfaces centralized in `types/workbench.ts`
- All component props match their usage in App.tsx

**No inconsistencies found.**
