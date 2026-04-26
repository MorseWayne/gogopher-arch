# 前端任务线 MVP 实施计划

> **给 agentic workers：** 必须使用 `superpowers:subagent-driven-development`（推荐）或 `superpowers:executing-plans` 按任务逐步实施本计划。步骤使用 checkbox（`- [ ]`）语法跟踪进度。

**目标：** 构建一个仅前端实现的 Day 0 到 Day 5 实习任务线，让学习者可以切换任务、运行起始 Go 代码、获得任务专属检查，并阅读导师指导。

**架构：** 在 `web/src/tasks.ts` 新增静态任务目录，在 `web/src/taskFeedback.ts` 新增纯反馈评估器，并重构 `web/src/App.tsx`，让工作台根据任务数据渲染。保持现有 Gateway 和 Sandbox Engine 协议不变。

**技术栈：** React、TypeScript、Vite、Monaco Editor、Axios、Lucide React、Vitest、Go 沙盒 API。

---

## 范围检查

本计划只实现一个子系统：前端任务线 MVP。不新增后端任务 API、持久化进度、认证、AI 反馈或真正的 `go test` 沙盒模式。

## 文件结构

- 新建 `web/src/taskFeedback.ts`：用于根据任务检查评估沙盒输出的纯类型和函数。
- 新建 `web/src/taskFeedback.test.ts`：用 Vitest 覆盖待运行、连接错误、通过和失败反馈状态。
- 新建 `web/src/tasks.ts`：静态 Day 0 到 Day 5 任务目录。
- 新建 `web/src/tasks.test.ts`：用 Vitest 覆盖任务目录完整性和任务 id 唯一性。
- 修改 `web/package.json`：增加 `test` 脚本和 `vitest` 开发依赖。
- 修改 `web/package-lock.json`：通过 `npm install` 更新。
- 修改 `web/src/App.tsx`：渲染任务驱动的工作台。
- 修改 `web/src/App.css`：增加任务导航、状态和复盘样式。
- 修改 `README.md`：标记第一阶段文档、工作台和任务线条目已完成。

## 任务 1：新增反馈评估器和测试

**文件：**
- 修改： `web/package.json`
- 修改： `web/package-lock.json`
- 新建： `web/src/taskFeedback.test.ts`
- 新建： `web/src/taskFeedback.ts`

- [ ] **步骤 1：安装 Vitest**

运行：

```bash
cd web
npm install -D vitest
```

预期：`package.json` 和 `package-lock.json` 发生变化，`node_modules` 仍保持未跟踪。

- [ ] **步骤 2：增加测试脚本**

在 `web/package.json` 中更新 `scripts` 对象，使其包含：

```json
{
  "dev": "vite",
  "build": "tsc -b && vite build",
  "lint": "eslint .",
  "preview": "vite preview",
  "test": "vitest"
}
```

- [ ] **步骤 3：为反馈评估写失败测试**

创建 `web/src/taskFeedback.test.ts`：

```ts
import { describe, expect, it } from 'vitest';
import {
  didPassTask,
  evaluateTaskChecks,
  type SandboxResponse,
  type TaskCheck,
} from './taskFeedback';

const checks: TaskCheck[] = [
  {
    type: 'stdoutIncludes',
    label: 'stdout phrase',
    passDetail: 'stdout contains the expected phrase.',
    failDetail: 'stdout does not contain the expected phrase.',
    value: 'hello intern',
  },
  {
    type: 'stderrExcludes',
    label: 'no panic',
    passDetail: 'stderr does not include panic.',
    failDetail: 'stderr still includes panic.',
    value: 'panic:',
  },
];

function sandbox(overrides: Partial<SandboxResponse> = {}): SandboxResponse {
  return {
    stdout: 'hello intern\n',
    stderr: '',
    status: 'success',
    duration: 1200000,
    exit_code: 0,
    ...overrides,
  };
}

describe('evaluateTaskChecks', () => {
  it('returns idle feedback before the first run', () => {
    const feedback = evaluateTaskChecks(null, null, checks);

    expect(feedback.map((item) => item.state)).toEqual(['idle', 'idle', 'idle', 'idle']);
    expect(feedback[0].label).toBe('连接 Gateway');
    expect(feedback[2].label).toBe('stdout phrase');
  });

  it('returns connection failure feedback when the gateway request fails', () => {
    const feedback = evaluateTaskChecks(null, 'network error', checks);

    expect(feedback[0]).toMatchObject({
      label: '连接 Gateway',
      state: 'fail',
    });
    expect(feedback[2].state).toBe('idle');
  });

  it('passes task checks when the sandbox succeeds and all checks match', () => {
    const feedback = evaluateTaskChecks(sandbox(), null, checks);

    expect(feedback.map((item) => item.state)).toEqual(['pass', 'pass', 'pass', 'pass']);
    expect(didPassTask(sandbox(), null, checks)).toBe(true);
  });

  it('fails task checks when stdout or stderr does not match', () => {
    const output = sandbox({
      stdout: 'wrong output\n',
      stderr: 'panic: assignment to entry in nil map',
      status: 'error',
      exit_code: 1,
    });

    const feedback = evaluateTaskChecks(output, null, checks);

    expect(feedback.map((item) => item.state)).toEqual(['pass', 'fail', 'fail', 'fail']);
    expect(didPassTask(output, null, checks)).toBe(false);
  });
});
```

- [ ] **步骤 4：运行测试，确认测试失败**

运行：

```bash
cd web
npm test -- --run src/taskFeedback.test.ts
```

预期：FAIL，因为 `./taskFeedback` 尚不存在。

- [ ] **步骤 5：实现反馈评估器**

创建 `web/src/taskFeedback.ts`：

```ts
export interface SandboxResponse {
  stdout: string;
  stderr: string;
  status: string;
  duration: number;
  exit_code: number;
}

export type FeedbackState = 'idle' | 'pass' | 'fail';

export interface FeedbackItem {
  label: string;
  detail: string;
  state: FeedbackState;
}

export type TaskCheck =
  | {
      type: 'exitSuccess';
      label: string;
      passDetail: string;
      failDetail: string;
    }
  | {
      type: 'stdoutIncludes';
      label: string;
      passDetail: string;
      failDetail: string;
      value: string;
    }
  | {
      type: 'stdoutRegex';
      label: string;
      passDetail: string;
      failDetail: string;
      pattern: string;
      flags?: string;
    }
  | {
      type: 'stderrExcludes';
      label: string;
      passDetail: string;
      failDetail: string;
      value: string;
    };

function didSandboxSucceed(output: SandboxResponse): boolean {
  return output.status === 'success' && output.exit_code === 0;
}

function idleCheckFeedback(check: TaskCheck): FeedbackItem {
  return {
    label: check.label,
    detail: '运行当前任务后查看检查结果。',
    state: 'idle',
  };
}

function evaluateSingleCheck(check: TaskCheck, output: SandboxResponse): FeedbackItem {
  let passed = false;

  if (check.type === 'exitSuccess') {
    passed = didSandboxSucceed(output);
  }

  if (check.type === 'stdoutIncludes') {
    passed = output.stdout.includes(check.value);
  }

  if (check.type === 'stdoutRegex') {
    passed = new RegExp(check.pattern, check.flags).test(output.stdout);
  }

  if (check.type === 'stderrExcludes') {
    passed = !output.stderr.includes(check.value);
  }

  return {
    label: check.label,
    detail: passed ? check.passDetail : check.failDetail,
    state: passed ? 'pass' : 'fail',
  };
}

export function evaluateTaskChecks(
  output: SandboxResponse | null,
  error: string | null,
  checks: TaskCheck[],
): FeedbackItem[] {
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
      ...checks.map(idleCheckFeedback),
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
      ...checks.map(idleCheckFeedback),
    ];
  }

  const runSucceeded = didSandboxSucceed(output);

  return [
    {
      label: '连接 Gateway',
      detail: '已收到沙盒执行结果。',
      state: 'pass',
    },
    {
      label: '运行结果',
      detail: runSucceeded ? '程序正常退出。' : '程序未正常退出，请查看 stderr。',
      state: runSucceeded ? 'pass' : 'fail',
    },
    ...checks.map((check) => evaluateSingleCheck(check, output)),
  ];
}

export function didPassTask(
  output: SandboxResponse | null,
  error: string | null,
  checks: TaskCheck[],
): boolean {
  if (error || !output || !didSandboxSucceed(output)) {
    return false;
  }

  return checks.every((check) => evaluateSingleCheck(check, output).state === 'pass');
}
```

- [ ] **步骤 6：运行反馈测试**

运行：

```bash
cd web
npm test -- --run src/taskFeedback.test.ts
```

预期：PASS。

- [ ] **步骤 7：提交反馈评估器**

运行：

```bash
git add web/package.json web/package-lock.json web/src/taskFeedback.ts web/src/taskFeedback.test.ts
git commit -m "test: add task feedback evaluator"
```

## 任务 2：新增静态 Day 0 到 Day 5 任务目录

**文件：**
- 新建： `web/src/tasks.test.ts`
- 新建： `web/src/tasks.ts`

- [ ] **步骤 1：为任务目录写失败测试**

创建 `web/src/tasks.test.ts`：

```ts
import { describe, expect, it } from 'vitest';
import { defaultTaskId, findTaskById, internshipTasks } from './tasks';

describe('internshipTasks', () => {
  it('contains exactly Day 0 through Day 5', () => {
    expect(internshipTasks.map((task) => task.day)).toEqual([0, 1, 2, 3, 4, 5]);
  });

  it('uses stable unique ids and complete task content', () => {
    const ids = internshipTasks.map((task) => task.id);

    expect(new Set(ids).size).toBe(ids.length);

    for (const task of internshipTasks) {
      expect(task.title.length).toBeGreaterThan(4);
      expect(task.summary.length).toBeGreaterThan(8);
      expect(task.background.length).toBeGreaterThan(20);
      expect(task.objective.length).toBeGreaterThan(10);
      expect(task.starterCode).toContain('package main');
      expect(task.criteria.length).toBeGreaterThanOrEqual(3);
      expect(task.lesson.length).toBeGreaterThanOrEqual(3);
      expect(task.mentorHints.length).toBeGreaterThanOrEqual(3);
      expect(task.review.length).toBeGreaterThanOrEqual(3);
      expect(task.checks.length).toBeGreaterThanOrEqual(2);
    }
  });

  it('finds tasks by id and falls back to the default task', () => {
    expect(defaultTaskId).toBe('day-0-first-run');
    expect(findTaskById('day-3-validation').day).toBe(3);
    expect(findTaskById('missing-task').id).toBe(defaultTaskId);
  });
});
```

- [ ] **步骤 2：运行测试，确认测试失败**

运行：

```bash
cd web
npm test -- --run src/tasks.test.ts
```

预期：FAIL，因为 `./tasks` 尚不存在。

- [ ] **步骤 3：实现任务目录**

创建 `web/src/tasks.ts`：

```ts
import type { TaskCheck } from './taskFeedback';

export interface InternTask {
  id: string;
  day: number;
  title: string;
  track: string;
  summary: string;
  background: string;
  objective: string;
  starterCode: string;
  criteria: string[];
  lesson: string[];
  mentorHints: string[];
  review: string[];
  checks: TaskCheck[];
}

export const internshipTasks: InternTask[] = [
  {
    id: 'day-0-first-run',
    day: 0,
    title: '第一次运行 Go 代码',
    track: '入职前训练营',
    summary: '确认沙盒可以运行 Go 程序，并学会观察 stdout。',
    background:
      '你刚加入虚拟 Go 后端团队。导师让你先跑通第一段 Go 程序，确认本地 Gateway、Sandbox 和前端工作台都能连起来。',
    objective: '运行代码并看到 GoGopher Day 0 的欢迎输出。',
    starterCode: `package main

import "fmt"

func main() {
\tfmt.Println("GoGopher Day 0 sandbox ready")
}
`,
    criteria: [
      '程序可以成功运行，退出码为 0。',
      '控制台输出 GoGopher Day 0 sandbox ready。',
      'stderr 中没有 panic 或编译错误。',
    ],
    lesson: [
      'Go 程序从 package main 和 main 函数开始执行。',
      'fmt.Println 会把内容写到 stdout，也就是控制台标准输出。',
      '先跑通最小程序，再逐步修改，是排查问题最稳的方式。',
    ],
    mentorHints: [
      '先不要改结构，直接运行一次观察输出。',
      '如果连接失败，先确认 docker compose 或本地 Gateway 是否启动。',
      '如果出现编译错误，优先看 stderr 的第一行。',
    ],
    review: [
      '你完成了第一轮沙盒运行。',
      '真实工作中，先复现和观察输出比直接改代码更重要。',
      '下一步会开始处理 Go 零值和 map 初始化问题。',
    ],
    checks: [
      {
        type: 'exitSuccess',
        label: '程序退出状态',
        passDetail: '程序已正常退出。',
        failDetail: '程序还没有正常退出，请查看 stderr。',
      },
      {
        type: 'stdoutIncludes',
        label: '欢迎输出',
        passDetail: 'stdout 中出现了 Day 0 欢迎语。',
        failDetail: 'stdout 中还没有看到 GoGopher Day 0 sandbox ready。',
        value: 'GoGopher Day 0 sandbox ready',
      },
      {
        type: 'stderrExcludes',
        label: '无 panic',
        passDetail: 'stderr 中没有 panic。',
        failDetail: 'stderr 中仍然包含 panic。',
        value: 'panic:',
      },
    ],
  },
  {
    id: 'day-1-nil-map',
    day: 1,
    title: '修复 nil map 写入',
    track: 'Go 基础 Bug 修复',
    summary: '定位 nil map 写入 panic，并用 make 初始化 map。',
    background:
      '导师把一个用户分数统计函数交给你。当前代码会在运行时 panic，你需要定位原因并完成修复。',
    objective: '让 buildScoreMap 返回包含所有用户分数的 map。',
    starterCode: `package main

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
`,
    criteria: [
      '程序可以成功运行，不再出现 nil map 写入 panic。',
      'buildScoreMap 返回包含所有用户分数的 map。',
      '不要修改 main 函数里的输入数据和输出语句。',
    ],
    lesson: [
      'map 在写入前必须完成初始化。',
      'var scores map[string]int 声明的是 nil map，只能读，不能写。',
      'make(map[string]int, len(users)) 可以创建可写 map，并预留容量。',
    ],
    mentorHints: [
      '先定位 panic 行，再判断这个变量是否已经初始化。',
      '这类问题在实习任务里很常见：看起来类型对了，但零值不能直接写入。',
      '修复后再运行一次，确认 stdout 中出现 Ming 的分数。',
    ],
    review: [
      'nil map 是 Go 初学者最常见的运行时错误之一。',
      '声明变量和初始化底层数据结构不是一回事。',
      '真实项目里看到 assignment to entry in nil map，优先检查 map 创建路径。',
    ],
    checks: [
      {
        type: 'exitSuccess',
        label: '程序退出状态',
        passDetail: '程序已正常退出。',
        failDetail: '程序未正常退出，请检查 panic 或编译错误。',
      },
      {
        type: 'stdoutRegex',
        label: '目标分数输出',
        passDetail: 'stdout 中出现了 Ming 的正确分数。',
        failDetail: 'stdout 中还没有看到 Ming 的分数: 86。',
        pattern: 'Ming 的分数:\\s*86\\b',
      },
      {
        type: 'stderrExcludes',
        label: '无 nil map panic',
        passDetail: 'stderr 中没有 nil map 写入 panic。',
        failDetail: 'stderr 中仍然包含 nil map 写入 panic。',
        value: 'assignment to entry in nil map',
      },
    ],
  },
  {
    id: 'day-2-json-response',
    day: 2,
    title: '补全 JSON 响应',
    track: 'HTTP Handler 基础',
    summary: '用结构体和 json.Marshal 生成稳定的接口响应。',
    background:
      '团队的用户接口需要返回 JSON。当前函数只返回空对象，你需要补全响应编码逻辑，让输出包含用户 id 和 name。',
    objective: '让 encodeUser 返回包含 id 和 name 字段的 JSON 字符串。',
    starterCode: `package main

import "fmt"

type UserResponse struct {
\tID   int    \`json:"id"\`
\tName string \`json:"name"\`
}

func encodeUser(user UserResponse) string {
\treturn "{}"
}

func main() {
\tuser := UserResponse{ID: 7, Name: "Ming"}
\tfmt.Println(encodeUser(user))
}
`,
    criteria: [
      '程序可以成功运行。',
      '输出 JSON 中包含 id: 7。',
      '输出 JSON 中包含 name: "Ming"。',
    ],
    lesson: [
      'Go 的 encoding/json 可以把结构体编码为 JSON。',
      '结构体字段需要导出，也就是首字母大写，json.Marshal 才能读取。',
      'json tag 可以控制输出字段名，例如 json:"name"。',
    ],
    mentorHints: [
      '你需要引入 encoding/json。',
      'json.Marshal 返回 []byte 和 error，先处理 error，再转成 string。',
      '不要手写拼接 JSON，真实接口里很容易漏转义。',
    ],
    review: [
      '你完成了一个简化版 handler 响应编码任务。',
      '真实后端接口常见工作就是把内部结构转成稳定的响应结构。',
      '下一步会在响应逻辑里加入参数校验和错误处理。',
    ],
    checks: [
      {
        type: 'exitSuccess',
        label: '程序退出状态',
        passDetail: '程序已正常退出。',
        failDetail: '程序还没有正常退出，请查看 stderr。',
      },
      {
        type: 'stdoutRegex',
        label: '包含用户 ID',
        passDetail: 'JSON 输出中包含 id: 7。',
        failDetail: 'JSON 输出中还没有看到 id: 7。',
        pattern: '"id"\\s*:\\s*7',
      },
      {
        type: 'stdoutRegex',
        label: '包含用户名称',
        passDetail: 'JSON 输出中包含 name: Ming。',
        failDetail: 'JSON 输出中还没有看到 name: Ming。',
        pattern: '"name"\\s*:\\s*"Ming"',
      },
    ],
  },
  {
    id: 'day-3-validation',
    day: 3,
    title: '增加参数校验',
    track: '错误处理',
    summary: '为空输入返回明确错误，同时保持正常输入可用。',
    background:
      '昨天的响应函数可以输出正常数据，但接口还没有处理无效输入。导师要求你补上最基础的参数校验。',
    objective: '当 name 为空时返回 validation error，name 有值时仍然创建成功。',
    starterCode: `package main

import (
\t"fmt"
\t"strings"
)

func createUser(name string) (string, error) {
\treturn "created user: " + strings.TrimSpace(name), nil
}

func main() {
\tfor _, name := range []string{"", "Ming"} {
\t\tresult, err := createUser(name)
\t\tif err != nil {
\t\t\tfmt.Println("validation error:", err)
\t\t\tcontinue
\t\t}
\t\tfmt.Println(result)
\t}
}
`,
    criteria: [
      '程序可以成功运行。',
      '空 name 会输出 validation error。',
      '有效 name 仍然输出 created user: Ming。',
    ],
    lesson: [
      'Go 通常用返回 error 表达可恢复的业务错误。',
      'strings.TrimSpace 可以避免只输入空格时绕过校验。',
      '调用方应该先判断 err，再使用正常结果。',
    ],
    mentorHints: [
      '校验应该放在 createUser 内部，而不是只改 main。',
      '可以用 errors.New 创建简单错误。',
      '注意空字符串和全空格字符串都应该被拦住。',
    ],
    review: [
      '你补上了接口处理中最常见的输入防线。',
      '真实业务里，错误信息既要能帮助调用方定位，也不能泄露内部细节。',
      '下一步会用表驱动方式把多个输入案例组织起来。',
    ],
    checks: [
      {
        type: 'exitSuccess',
        label: '程序退出状态',
        passDetail: '程序已正常退出。',
        failDetail: '程序还没有正常退出，请查看 stderr。',
      },
      {
        type: 'stdoutIncludes',
        label: '无效输入错误',
        passDetail: '空 name 已输出 validation error。',
        failDetail: '空 name 还没有输出 validation error。',
        value: 'validation error:',
      },
      {
        type: 'stdoutIncludes',
        label: '有效输入成功',
        passDetail: '有效 name 仍然创建成功。',
        failDetail: '还没有看到 created user: Ming。',
        value: 'created user: Ming',
      },
    ],
  },
  {
    id: 'day-4-table-driven',
    day: 4,
    title: '补全表驱动检查',
    track: '测试思维',
    summary: '用一组 case 验证边界输入，而不是只测一个 happy path。',
    background:
      '导师希望你把分数归一化逻辑的边界情况补齐。当前程序用一个小循环模拟表驱动测试，但有些 case 还不能通过。',
    objective: '让 lower bound、upper bound 和 normal score 三个 case 都输出 PASS。',
    starterCode: `package main

import "fmt"

func normalizeScore(score int) int {
\treturn score
}

func main() {
\tcases := []struct {
\t\tname  string
\t\tinput int
\t\twant  int
\t}{
\t\t{name: "lower bound", input: -3, want: 0},
\t\t{name: "upper bound", input: 120, want: 100},
\t\t{name: "normal score", input: 86, want: 86},
\t}

\tfor _, tc := range cases {
\t\tgot := normalizeScore(tc.input)
\t\tif got != tc.want {
\t\t\tfmt.Printf("FAIL %s: got %d want %d\\n", tc.name, got, tc.want)
\t\t\tcontinue
\t\t}
\t\tfmt.Println("PASS", tc.name)
\t}
}
`,
    criteria: [
      '程序可以成功运行。',
      'lower bound case 输出 PASS。',
      'upper bound case 输出 PASS。',
      'normal score case 输出 PASS。',
    ],
    lesson: [
      '表驱动测试把输入、期望和案例名称放在同一张表里。',
      '边界值比普通值更容易暴露隐藏 Bug。',
      '当前沙盒还运行 go run，所以这里先用 main 函数模拟测试输出。',
    ],
    mentorHints: [
      '先看 FAIL 行，它会告诉你 got 和 want 的差异。',
      'normalizeScore 应该把小于 0 的值收敛到 0。',
      'normalizeScore 也应该把大于 100 的值收敛到 100。',
    ],
    review: [
      '你用多个 case 验证了同一个函数的边界行为。',
      '真实 Go 项目里会把这种结构放进 _test.go 并运行 go test。',
      '下一步会处理带 timeout 的任务，练习 context 取消。',
    ],
    checks: [
      {
        type: 'exitSuccess',
        label: '程序退出状态',
        passDetail: '程序已正常退出。',
        failDetail: '程序还没有正常退出，请查看 stderr。',
      },
      {
        type: 'stdoutIncludes',
        label: 'lower bound 通过',
        passDetail: 'lower bound case 已通过。',
        failDetail: 'lower bound case 还没有通过。',
        value: 'PASS lower bound',
      },
      {
        type: 'stdoutIncludes',
        label: 'upper bound 通过',
        passDetail: 'upper bound case 已通过。',
        failDetail: 'upper bound case 还没有通过。',
        value: 'PASS upper bound',
      },
      {
        type: 'stdoutIncludes',
        label: 'normal score 通过',
        passDetail: 'normal score case 已通过。',
        failDetail: 'normal score case 还没有通过。',
        value: 'PASS normal score',
      },
    ],
  },
  {
    id: 'day-5-context-timeout',
    day: 5,
    title: '尊重 context 超时',
    track: '并发与超时',
    summary: '让慢操作监听 ctx.Done，避免请求超时后还继续工作。',
    background:
      '一个报表查询函数可能很慢。导师要求你让它尊重 context timeout，这样上游请求取消后，函数能尽快返回。',
    objective: '让 fetchReport 在 context 超时时输出 timeout respected。',
    starterCode: `package main

import (
\t"context"
\t"fmt"
\t"time"
)

func fetchReport(ctx context.Context) string {
\ttime.Sleep(200 * time.Millisecond)
\treturn "report ready"
}

func main() {
\tctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
\tdefer cancel()

\tfmt.Println(fetchReport(ctx))
}
`,
    criteria: [
      '程序可以成功运行。',
      'fetchReport 会监听 ctx.Done。',
      'stdout 输出 timeout respected。',
    ],
    lesson: [
      'context.Context 用来在调用链中传递取消和超时信号。',
      'select 可以同时等待业务结果和 ctx.Done。',
      '后台工作不尊重取消，会浪费资源并拖慢服务关闭。',
    ],
    mentorHints: [
      '不要只把 sleep 时间改短，重点是响应 ctx.Done。',
      '可以用 select 在 time.After 和 ctx.Done 之间选择。',
      'ctx.Err() 可以告诉你取消原因，例如 deadline exceeded。',
    ],
    review: [
      '你完成了入职第一周里最轻量的并发取消练习。',
      '真实接口里，数据库查询、RPC 调用和外部 API 都应该接收 context。',
      '后续工程进阶路线会继续展开并发、超时、重试和可观测性。',
    ],
    checks: [
      {
        type: 'exitSuccess',
        label: '程序退出状态',
        passDetail: '程序已正常退出。',
        failDetail: '程序还没有正常退出，请查看 stderr。',
      },
      {
        type: 'stdoutIncludes',
        label: '超时路径',
        passDetail: 'stdout 中出现了 timeout respected。',
        failDetail: 'stdout 中还没有看到 timeout respected。',
        value: 'timeout respected',
      },
      {
        type: 'stderrExcludes',
        label: '无 panic',
        passDetail: 'stderr 中没有 panic。',
        failDetail: 'stderr 中仍然包含 panic。',
        value: 'panic:',
      },
    ],
  },
];

export const defaultTaskId = internshipTasks[0].id;

export function findTaskById(id: string): InternTask {
  return internshipTasks.find((task) => task.id === id) ?? internshipTasks[0];
}
```

- [ ] **步骤 4：运行任务目录测试**

运行：

```bash
cd web
npm test -- --run src/tasks.test.ts
```

预期：PASS。

- [ ] **步骤 5：提交任务目录**

运行：

```bash
git add web/src/tasks.ts web/src/tasks.test.ts
git commit -m "feat: add first week task catalog"
```

## 任务 3：重构 App 以渲染任务线

**文件：**
- 修改： `web/src/App.tsx`

- [ ] **步骤 1：替换 imports 并移除硬编码任务内容**

在 `web/src/App.tsx` 中，用新模块的 imports 替换当前 imports 和本地任务数组：

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
  RotateCcw,
  Terminal,
} from 'lucide-react';
import './App.css';
import { defaultTaskId, findTaskById, internshipTasks } from './tasks';
import {
  didPassTask,
  evaluateTaskChecks,
  type SandboxResponse,
} from './taskFeedback';
```

从 `App.tsx` 中移除本地的 `SandboxResponse`、`FeedbackState`、`FeedbackItem`、`DEFAULT_CODE`、`taskCriteria`、`lessonPoints`、`mentorHints` 和 `getFeedback` 定义。

- [ ] **步骤 2：更新组件状态和任务选择逻辑**

在 `App` 内，将初始状态代码块替换为：

```tsx
function App() {
  const [selectedTaskId, setSelectedTaskId] = useState(defaultTaskId);
  const selectedTask = findTaskById(selectedTaskId);
  const [code, setCode] = useState(selectedTask.starterCode);
  const [output, setOutput] = useState<SandboxResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [taskResults, setTaskResults] = useState<Record<string, 'pass' | 'fail'>>({});

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
```

- [ ] **步骤 3：更新 `handleRun`**

将现有 `handleRun` 替换为：

```tsx
  const handleRun = async () => {
    setLoading(true);
    setError(null);
    setOutput(null);

    try {
      const response = await axios.post<SandboxResponse>('http://localhost:8080/api/v1/execute', {
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
```

- [ ] **步骤 4：替换 JSX return**

将 `return` 主体替换为：

```tsx
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
        <div className="topbar-actions">
          <button className="ghost-button" onClick={handleResetCode} disabled={loading}>
            <RotateCcw size={16} />
            重置代码
          </button>
          <button className="run-button" onClick={handleRun} disabled={loading}>
            <Play size={17} fill="currentColor" />
            {loading ? '运行中' : '运行代码'}
          </button>
        </div>
      </header>

      <main className="workbench">
        <aside className="task-panel" aria-label="任务卡">
          <section className="panel-section task-nav-section">
            <div className="section-title">
              <ClipboardCheck size={16} />
              <span>任务列表</span>
            </div>
            <div className="task-list">
              {internshipTasks.map((task) => {
                const result = taskResults[task.id];
                const isSelected = task.id === selectedTask.id;

                return (
                  <button
                    className={`task-list-item ${isSelected ? 'selected' : ''} ${result || ''}`}
                    key={task.id}
                    onClick={() => handleSelectTask(task.id)}
                    type="button"
                  >
                    <span className="task-day">Day {task.day}</span>
                    <span className="task-list-title">{task.title}</span>
                    <span className="task-list-track">{task.track}</span>
                  </button>
                );
              })}
            </div>
          </section>

          <section className="panel-section hero-section">
            <div className="section-title">
              <ClipboardCheck size={16} />
              <span>任务卡</span>
            </div>
            <h2>
              Day {selectedTask.day}：{selectedTask.title}
            </h2>
            <p>{selectedTask.background}</p>
            <p className="objective">{selectedTask.objective}</p>
          </section>

          <section className="panel-section">
            <div className="section-title">
              <CheckCircle2 size={16} />
              <span>验收标准</span>
            </div>
            <ul className="check-list">
              {selectedTask.criteria.map((item) => (
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
              {selectedTask.lesson.map((item) => (
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
            <span className="file-badge">{selectedTask.track}</span>
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
            <div className="feedback-summary">
              {currentTaskPassed ? '本任务已通过。' : '运行代码后查看任务检查。'}
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
              {selectedTask.mentorHints.map((hint) => (
                <li key={hint}>{hint}</li>
              ))}
            </ul>
          </section>

          <section className="panel-section">
            <div className="section-title">
              <BookOpen size={16} />
              <span>任务后复盘</span>
            </div>
            <ul className="review-list">
              {selectedTask.review.map((item) => (
                <li key={item}>{item}</li>
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
```

保留 `App` 上方现有的 `formatDuration` helper，不做修改。

- [ ] **步骤 5：运行 TypeScript 构建**

运行：

```bash
cd web
npm run build
```

预期：PASS。

- [ ] **步骤 6：运行前端测试**

运行：

```bash
cd web
npm test -- --run
```

预期：PASS。

- [ ] **步骤 7：提交 App 重构**

运行：

```bash
git add web/src/App.tsx
git commit -m "feat: render first week task line"
```

## 任务 4：新增任务导航和复盘样式

**文件：**
- 修改： `web/src/App.css`

- [ ] **步骤 1：增加 topbar 操作区样式**

在 `web/src/App.css` 中，将以下样式添加到 `.brand h1` 后面：

```css
.topbar-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.ghost-button {
  height: 40px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 0 14px;
  border: 1px solid #31414a;
  border-radius: 6px;
  background: #10171a;
  color: #dce5e2;
  font-weight: 700;
  cursor: pointer;
}

.ghost-button:disabled {
  cursor: wait;
  opacity: 0.62;
}
```

- [ ] **步骤 2：增加任务列表样式**

将以下样式添加到 `.panel-section` 后面：

```css
.task-nav-section {
  padding-bottom: 14px;
}

.task-list {
  display: grid;
  gap: 8px;
  margin-top: 12px;
}

.task-list-item {
  width: 100%;
  display: grid;
  grid-template-columns: auto 1fr;
  gap: 4px 10px;
  align-items: center;
  padding: 10px 12px;
  border: 1px solid #253038;
  border-radius: 8px;
  background: #10171a;
  color: #dce5e2;
  text-align: left;
  cursor: pointer;
}

.task-list-item.selected {
  border-color: #f0b429;
  background: #1d231d;
}

.task-list-item.pass {
  border-color: #2c8a62;
}

.task-list-item.fail {
  border-color: #8a3b3b;
}

.task-day {
  grid-row: span 2;
  min-width: 48px;
  color: #f0b429;
  font-size: 12px;
  font-weight: 800;
}

.task-list-title {
  color: #f8fffb;
  font-weight: 700;
}

.task-list-track {
  color: #8fb7aa;
  font-size: 12px;
}
```

- [ ] **步骤 3：增加目标、摘要和复盘样式**

将以下样式添加到现有列表样式附近：

```css
.objective {
  margin-top: 12px !important;
  color: #f0d58a !important;
}

.feedback-summary {
  margin-top: 12px;
  padding: 10px 12px;
  border: 1px solid #31414a;
  border-radius: 8px;
  background: #10171a;
  color: #f8fffb;
  font-weight: 700;
}

.review-list {
  margin: 12px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 10px;
}

.review-list li {
  padding: 10px 12px;
  border-left: 3px solid #f0b429;
  background: #10171a;
  color: #dce5e2;
}
```

- [ ] **步骤 4：更新移动端 topbar 行为**

在现有 `@media (max-width: 760px)` 块中，将 `.run-button` 规则替换为：

```css
  .topbar-actions {
    width: 100%;
    display: grid;
    grid-template-columns: 1fr 1fr;
  }

  .run-button,
  .ghost-button {
    width: 100%;
  }
```

- [ ] **步骤 5：运行构建**

运行：

```bash
cd web
npm run build
```

预期：PASS。

- [ ] **步骤 6：提交样式**

运行：

```bash
git add web/src/App.css
git commit -m "style: add task line workbench states"
```

## 任务 5：更新路线图文档

**文件：**
- 修改： `README.md`

- [ ] **步骤 1：标记第一阶段已完成条目**

在 `README.md` 中，将第一阶段清单更新为：

```markdown
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
```

- [ ] **步骤 2：验证 README 路线图**

运行：

```bash
rg -n "Day 0|Day 1|Day 2|Day 3|Day 4|Day 5|\\[x\\] 前端首屏" README.md
```

预期：输出包含 Day 0 到 Day 5，以及已完成的工作台条目。

- [ ] **步骤 3：提交 README 更新**

运行：

```bash
git add README.md
git commit -m "docs: mark first week task line progress"
```

## 任务 6：全量验证

**文件：**
- 验证： `web/src/taskFeedback.ts`
- 验证： `web/src/tasks.ts`
- 验证： `web/src/App.tsx`
- 验证： `web/src/App.css`
- 验证： `README.md`

- [ ] **步骤 1：运行 Go 测试**

运行：

```bash
go test ./...
```

预期：PASS。

- [ ] **步骤 2：运行前端测试**

运行：

```bash
cd web
npm test -- --run
```

预期：PASS。

- [ ] **步骤 3：运行前端构建**

运行：

```bash
cd web
npm run build
```

预期：PASS。

- [ ] **步骤 4：验证任务线内容存在**

运行：

```bash
rg -n "day-0-first-run|day-1-nil-map|day-2-json-response|day-3-validation|day-4-table-driven|day-5-context-timeout" web/src/tasks.ts
```

预期：输出包含全部 6 个任务 id。

运行：

```bash
rg -n "任务列表|任务后复盘|导师提示|Day \\{selectedTask.day\\}" web/src/App.tsx
```

预期：输出包含任务导航、复盘、导师提示和选中日期渲染标签。

- [ ] **步骤 5：验证旧第一屏定位仍已移除**

运行：

```bash
rg -n "架构师进化之路|双十一|高性能 IM|去中心化交易系统|10W QPS" README.md docs/specs/2026-03-13-gogopher-arch-design.md docs/plans/2026-03-13-implementation-plan.md web/src
```

预期：无输出，退出码为 1。

- [ ] **步骤 6：检查 git 状态和最近提交**

运行：

```bash
git status --short
git log --oneline --max-count=8
```

预期：只剩无关的既有未跟踪文件，并且最近提交包含：

```text
docs: mark first week task line progress
style: add task line workbench states
feat: render first week task line
feat: add first week task catalog
test: add task feedback evaluator
docs: design frontend task line mvp
```

## 自检

### 规格覆盖

- Day 0 到 Day 5 任务线：任务 2。
- 静态前端任务数据：任务 2。
- 数据驱动工作台渲染：任务 3。
- 任务专属反馈检查：任务 1 和任务 3。
- 导师提示和任务复盘：任务 2 和任务 3。
- 保留现有沙盒协议：任务 3 使用当前 `/api/v1/execute` 请求形态。
- 不做完整重设计的前端样式：任务 4。
- 验证和旧定位扫描：任务 6。

### 占位符扫描

本计划刻意避免开放式占位符。每个新增文件都有具体代码，每个验证步骤都有明确命令和预期结果。

### 类型一致性

- `SandboxResponse` 只在 `taskFeedback.ts` 中定义一次，并由 `App.tsx` 导入。
- `TaskCheck` 在 `taskFeedback.ts` 中定义，并由 `tasks.ts` 导入。
- `InternTask.checks` 使用 `TaskCheck[]`。
- `findTaskById`、`defaultTaskId` 和 `internshipTasks` 由 `App.tsx` 导入。
