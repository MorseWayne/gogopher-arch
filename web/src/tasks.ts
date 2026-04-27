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
