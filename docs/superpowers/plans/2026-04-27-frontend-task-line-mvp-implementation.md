# Frontend Task Line MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build a frontend-only Day 0 to Day 5 intern task line so learners can switch tasks, run starter Go code, receive task-specific checks, and read mentor guidance.

**Architecture:** Add a static task catalog in `web/src/tasks.ts`, a pure feedback evaluator in `web/src/taskFeedback.ts`, and refactor `web/src/App.tsx` to render the workbench from task data. Keep the existing Gateway and Sandbox Engine contract unchanged.

**Tech Stack:** React, TypeScript, Vite, Monaco Editor, Axios, Lucide React, Vitest, Go sandbox API.

---

## Scope Check

This plan implements one subsystem: the frontend task line MVP. It does not add backend task APIs, persisted progress, authentication, AI feedback, or a real `go test` sandbox mode.

## File Structure

- Create `web/src/taskFeedback.ts`: pure types and functions for evaluating sandbox output against task checks.
- Create `web/src/taskFeedback.test.ts`: Vitest coverage for idle, connection error, pass, and fail feedback states.
- Create `web/src/tasks.ts`: static Day 0 to Day 5 task catalog.
- Create `web/src/tasks.test.ts`: Vitest coverage for catalog completeness and unique task ids.
- Modify `web/package.json`: add `test` script and `vitest` dev dependency.
- Modify `web/package-lock.json`: update via `npm install`.
- Modify `web/src/App.tsx`: render task-driven workbench.
- Modify `web/src/App.css`: add task navigation, status, and review styles.
- Modify `README.md`: mark completed first-stage documentation/workbench/task-line items.

## Task 1: Add Feedback Evaluator With Tests

**Files:**
- Modify: `web/package.json`
- Modify: `web/package-lock.json`
- Create: `web/src/taskFeedback.test.ts`
- Create: `web/src/taskFeedback.ts`

- [ ] **Step 1: Install Vitest**

Run:

```bash
cd web
npm install -D vitest
```

Expected: `package.json` and `package-lock.json` change, and `node_modules` remains untracked.

- [ ] **Step 2: Add the test script**

In `web/package.json`, update the `scripts` object so it contains:

```json
{
  "dev": "vite",
  "build": "tsc -b && vite build",
  "lint": "eslint .",
  "preview": "vite preview",
  "test": "vitest"
}
```

- [ ] **Step 3: Write failing tests for feedback evaluation**

Create `web/src/taskFeedback.test.ts`:

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

- [ ] **Step 4: Run the tests to verify they fail**

Run:

```bash
cd web
npm test -- --run src/taskFeedback.test.ts
```

Expected: FAIL because `./taskFeedback` does not exist.

- [ ] **Step 5: Implement the feedback evaluator**

Create `web/src/taskFeedback.ts`:

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

- [ ] **Step 6: Run feedback tests**

Run:

```bash
cd web
npm test -- --run src/taskFeedback.test.ts
```

Expected: PASS.

- [ ] **Step 7: Commit feedback evaluator**

Run:

```bash
git add web/package.json web/package-lock.json web/src/taskFeedback.ts web/src/taskFeedback.test.ts
git commit -m "test: add task feedback evaluator"
```

## Task 2: Add Static Day 0 To Day 5 Task Catalog

**Files:**
- Create: `web/src/tasks.test.ts`
- Create: `web/src/tasks.ts`

- [ ] **Step 1: Write failing tests for the task catalog**

Create `web/src/tasks.test.ts`:

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

- [ ] **Step 2: Run the tests to verify they fail**

Run:

```bash
cd web
npm test -- --run src/tasks.test.ts
```

Expected: FAIL because `./tasks` does not exist.

- [ ] **Step 3: Implement the task catalog**

Create `web/src/tasks.ts`:

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

- [ ] **Step 4: Run catalog tests**

Run:

```bash
cd web
npm test -- --run src/tasks.test.ts
```

Expected: PASS.

- [ ] **Step 5: Commit task catalog**

Run:

```bash
git add web/src/tasks.ts web/src/tasks.test.ts
git commit -m "feat: add first week task catalog"
```

## Task 3: Refactor App To Render Task Line

**Files:**
- Modify: `web/src/App.tsx`

- [ ] **Step 1: Replace imports and remove hard-coded task content**

In `web/src/App.tsx`, replace the current imports and local task arrays with imports from the new modules:

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

Remove the local `SandboxResponse`, `FeedbackState`, `FeedbackItem`, `DEFAULT_CODE`, `taskCriteria`, `lessonPoints`, `mentorHints`, and `getFeedback` definitions from `App.tsx`.

- [ ] **Step 2: Update component state and task selection**

Inside `App`, replace the initial state block with:

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

- [ ] **Step 3: Update `handleRun`**

Replace the existing `handleRun` with:

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

- [ ] **Step 4: Replace the JSX return**

Replace the `return` body with:

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

Keep the existing `formatDuration` helper unchanged above `App`.

- [ ] **Step 5: Run TypeScript build**

Run:

```bash
cd web
npm run build
```

Expected: PASS.

- [ ] **Step 6: Run frontend tests**

Run:

```bash
cd web
npm test -- --run
```

Expected: PASS.

- [ ] **Step 7: Commit App refactor**

Run:

```bash
git add web/src/App.tsx
git commit -m "feat: render first week task line"
```

## Task 4: Add Task Navigation And Review Styles

**Files:**
- Modify: `web/src/App.css`

- [ ] **Step 1: Add topbar action styles**

In `web/src/App.css`, add these styles after `.brand h1`:

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

- [ ] **Step 2: Add task list styles**

Add these styles after `.panel-section`:

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

- [ ] **Step 3: Add objective, summary, and review styles**

Add these styles near the existing list styles:

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

- [ ] **Step 4: Update mobile topbar behavior**

Inside the existing `@media (max-width: 760px)` block, replace the `.run-button` rule with:

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

- [ ] **Step 5: Run build**

Run:

```bash
cd web
npm run build
```

Expected: PASS.

- [ ] **Step 6: Commit styles**

Run:

```bash
git add web/src/App.css
git commit -m "style: add task line workbench states"
```

## Task 5: Update Roadmap Docs

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Mark completed first-stage items**

In `README.md`, update the first-stage checklist to:

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

- [ ] **Step 2: Verify README roadmap**

Run:

```bash
rg -n "Day 0|Day 1|Day 2|Day 3|Day 4|Day 5|\\[x\\] 前端首屏" README.md
```

Expected: output includes Day 0 through Day 5 and the completed workbench line.

- [ ] **Step 3: Commit README update**

Run:

```bash
git add README.md
git commit -m "docs: mark first week task line progress"
```

## Task 6: Full Verification

**Files:**
- Verify: `web/src/taskFeedback.ts`
- Verify: `web/src/tasks.ts`
- Verify: `web/src/App.tsx`
- Verify: `web/src/App.css`
- Verify: `README.md`

- [ ] **Step 1: Run Go tests**

Run:

```bash
go test ./...
```

Expected: PASS.

- [ ] **Step 2: Run frontend tests**

Run:

```bash
cd web
npm test -- --run
```

Expected: PASS.

- [ ] **Step 3: Run frontend build**

Run:

```bash
cd web
npm run build
```

Expected: PASS.

- [ ] **Step 4: Verify task-line content exists**

Run:

```bash
rg -n "day-0-first-run|day-1-nil-map|day-2-json-response|day-3-validation|day-4-table-driven|day-5-context-timeout" web/src/tasks.ts
```

Expected: output includes all six task ids.

Run:

```bash
rg -n "任务列表|任务后复盘|导师提示|Day \\{selectedTask.day\\}" web/src/App.tsx
```

Expected: output includes task navigation, review, mentor hint, and selected day rendering labels.

- [ ] **Step 5: Verify old first-screen positioning stays removed**

Run:

```bash
rg -n "架构师进化之路|双十一|高性能 IM|去中心化交易系统|10W QPS" README.md docs/specs/2026-03-13-gogopher-arch-design.md docs/plans/2026-03-13-implementation-plan.md web/src
```

Expected: no output and exit code 1.

- [ ] **Step 6: Check git status and recent commits**

Run:

```bash
git status --short
git log --oneline --max-count=8
```

Expected: only unrelated pre-existing untracked files remain, and recent commits include:

```text
docs: mark first week task line progress
style: add task line workbench states
feat: render first week task line
feat: add first week task catalog
test: add task feedback evaluator
docs: design frontend task line mvp
```

## Self-Review

### Spec Coverage

- Day 0 to Day 5 task line: Task 2.
- Static frontend task data: Task 2.
- Data-driven workbench rendering: Task 3.
- Task-specific feedback checks: Task 1 and Task 3.
- Mentor hints and task review: Task 2 and Task 3.
- Existing sandbox contract preserved: Task 3 uses the current `/api/v1/execute` request shape.
- Frontend styles without full redesign: Task 4.
- Verification and old-positioning scan: Task 6.

### Placeholder Scan

The plan intentionally avoids open-ended placeholders. Every file addition has concrete code, and every verification step has an exact command and expected result.

### Type Consistency

- `SandboxResponse` is defined once in `taskFeedback.ts` and imported by `App.tsx`.
- `TaskCheck` is defined in `taskFeedback.ts` and imported by `tasks.ts`.
- `InternTask.checks` uses `TaskCheck[]`.
- `findTaskById`, `defaultTaskId`, and `internshipTasks` are imported by `App.tsx`.
