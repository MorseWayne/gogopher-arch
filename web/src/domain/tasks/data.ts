import type { InternTask } from './types';

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
  {
    id: 'day-6-db-query',
    day: 6,
    title: '数据库连接与查询',
    track: '数据库基础',
    summary: '连接 PostgreSQL 并查询员工数据，学习 database/sql 基础用法。',
    background:
      '团队的用户服务需要查询数据库。导师给你一个 employees 表，里面有员工的姓名、部门和薪资。你需要用 Go 的标准库 database/sql 连接数据库，查出所有在职员工并按薪资排序。',
    objective: '连接数据库，查询 employees 表中所有 active=true 的员工，按 salary 降序输出。',
    starterCode: `package main

import (
\t"database/sql"
\t"fmt"
\t"log"
\t"os"

\t_ "github.com/lib/pq"
)

func main() {
\tdbURL := os.Getenv("DATABASE_URL")
\tif dbURL == "" {
\t\tlog.Fatal("DATABASE_URL not set")
\t}

\tdb, err := sql.Open("postgres", dbURL)
\tif err != nil {
\t\tlog.Fatalf("connect: %v", err)
\t}
\tdefer db.Close()

\tif err := db.Ping(); err != nil {
\t\tlog.Fatalf("ping: %v", err)
\t}

\tfmt.Println("connected")
}
`,
    criteria: [
      '程序可以成功运行，退出码为 0。',
      '输出中包含 Zhang 的信息（Engineering 部门 130k）。',
      '输出中包含 Ming 的信息（Engineering 部门 120k）。',
      '输出中包含 Li 和 Wang 的信息。',
    ],
    lesson: [
      'database/sql 是 Go 标准库中的数据库抽象层，不绑定具体数据库。',
      'sql.Open 只是验证参数，不建立连接——要用 db.Ping 确认连通性。',
      '第三方驱动用 _ "github.com/lib/pq" 匿名导入，通过 init 注册。',
      'defer rows.Close() 必须写在 err 检查之后，否则 nil pointer panic。',
      'Query 返回的 rows 需要用 Next 遍历，再用 Scan 读取每一列。',
    ],
    mentorHints: [
      '先确认 db.Ping 成功（容器内 postgres:5432 可访问）。',
      'WHERE active = true 过滤在职员工。',
      'ORDER BY salary DESC 实现降序排列。',
      '不要忘记 defer rows.Close()。',
      '如果 DATABASE_URL 读不到，检查环境变量是否正确传入。',
    ],
    review: [
      '你完成了第一个数据库查询任务。',
      'database/sql 是 Go 后端开发的核心技能，大多数后端服务都离不开它。',
      'sql.Open + db.Ping 是标准的连接验证模式。',
      '接下来要学习事务处理，确保多步操作的原子性。',
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
        label: '查询到 Zhang',
        passDetail: 'stdout 中出现了 Zhang。',
        failDetail: 'stdout 中还没有看到 Zhang。',
        value: 'Zhang',
      },
      {
        type: 'stdoutIncludes',
        label: '查询到 Ming',
        passDetail: 'stdout 中出现了 Ming。',
        failDetail: 'stdout 中还没有看到 Ming。',
        value: 'Ming',
      },
      {
        type: 'stdoutIncludes',
        label: '查询到 Li',
        passDetail: 'stdout 中出现了 Li。',
        failDetail: 'stdout 中还没有看到 Li。',
        value: 'Li',
      },
      {
        type: 'stderrExcludes',
        label: '无连接错误',
        passDetail: 'stderr 中没有数据库连接错误。',
        failDetail: 'stderr 中包含数据库连接错误。',
        value: 'connect:',
      },
    ],
  },
  {
    id: 'day-7-transaction',
    day: 7,
    title: '事务回滚',
    track: '事务处理',
    summary: '用数据库事务实现转账操作，余额不足时自动回滚。',
    background:
      '团队正在开发一个转账功能。导师给了你一个 accounts 表，包含 Ming (余额 5000) 和 Yan (余额 3000)。你需要用事务实现转账：从 Ming 转 4000 到 Yan。如果转出方余额不足，事务要自动回滚；否则提交。',
    objective: '用 BEGIN/COMMIT/ROLLBACK 实现安全的转账事务，余额不足时输出 rollback，成功时输出 committed。',
    starterCode: `package main

import (
\t"context"
\t"database/sql"
\t"fmt"
\t"log"
\t"os"

\t_ "github.com/lib/pq"
)

func main() {
\tdbURL := os.Getenv("DATABASE_URL")
\tif dbURL == "" {
\t\tlog.Fatal("DATABASE_URL not set")
\t}

\tdb, err := sql.Open("postgres", dbURL)
\tif err != nil {
\t\tlog.Fatalf("open: %v", err)
\t}
\tdefer db.Close()

\tif err := db.Ping(); err != nil {
\t\tlog.Fatalf("ping: %v", err)
\t}

\tctx := context.Background()
\tamount := 4000

\tfmt.Println("transfer attempted")
}
`,
    criteria: [
      '程序可以成功运行，退出码为 0。',
      'Ming 余额 5000，转出 4000 后余额充足，事务应提交。',
      'stdout 中包含 committed（提交成功）。',
      '金额超过余额时，stdout 中包含 rollback（回滚成功）。',
    ],
    lesson: [
      '数据库事务用 BEGIN 开始，COMMIT 提交，ROLLBACK 回滚。',
      'Go 中 db.BeginTx 返回一个 *sql.Tx，提供 Begin/Commit/Rollback 方法。',
      '事务的错误处理：任何一个操作失败，都应该调用 Rollback。',
      'defer 不能直接 defer tx.Rollback()——Commit 后再 Rollback 会报错，要用 defer + 条件判断。',
      'CHECK (balance >= 0) 约束可以在数据库层面防止余额变负。',
    ],
    mentorHints: [
      '先查 from 和 to 的余额：SELECT balance FROM accounts WHERE holder = $1。',
      '更新用 UPDATE accounts SET balance = balance - $1 WHERE holder = $2。',
      '余额不足时调用 Rollback 并输出 rollback。',
      '所有操作成功后再调用 Commit 并输出 committed。',
      '注意：余额字段是 INT，不是字符串。',
    ],
    review: [
      '你完成了第一个数据库事务练习。',
      '真实系统中，转账、订单、库存扣减都要用事务保证一致性。',
      '事务的 ACID 特性（原子性、一致性、隔离性、持久性）是后端工程师面试高频考点。',
      '下一阶段将学习缓存和并发编程，进一步提升工程能力。',
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
        label: '提交或回滚判断',
        passDetail: 'stdout 中出现了 committed。',
        failDetail: 'stdout 中还没有看到 committed。',
        value: 'committed',
      },
      {
        type: 'stderrExcludes',
        label: '无数据库错误',
        passDetail: 'stderr 中没有数据库错误。',
        failDetail: 'stderr 中包含数据库错误。',
        value: 'sql:',
      },
    ],
  },
  {
    id: 'day-8-redis-cache',
    day: 8,
    title: 'Redis缓存基础',
    track: '缓存与性能',
    summary: '连接 Redis 并使用 SET/GET/DEL，学习缓存读写模式和 TTL。',
    background:
      '团队的用户服务经常查询同一批员工数据，每次都查数据库太慢了。导师要求你在数据库前面加一层 Redis 缓存：先查缓存，缓存未命中再查数据库，最后把结果写回缓存。',
    objective: '用 Redis 缓存查询结果，第二次查询同一 key 时命中缓存，输出 cached。',
    starterCode: `package main

import (
\t"context"
\t"fmt"
\t"log"
\t"os"
\t"time"

\t"github.com/redis/go-redis/v9"
)

func main() {
\tredisURL := os.Getenv("REDIS_URL")
\tif redisURL == "" {
\t\tlog.Fatal("REDIS_URL not set")
\t}

\trdb := redis.NewClient(&redis.Options{
\t\tAddr: redisURL,
\t})
\tctx := context.Background()

\tif err := rdb.Ping(ctx).Err(); err != nil {
\t\tlog.Fatalf("redis ping: %v", err)
\t}

\tkey := "employee:count"

\t// TODO: 实现缓存穿透模式
\t// 第一次查询：SET key, EXPIRE 10s，输出 "miss"
\t// 第二次查询同一 key：GET key，输出 "cached"

\tfmt.Println("redis ready")
}
`,
    criteria: [
      '程序可以成功运行，退出码为 0。',
      '第一次查询输出 miss（缓存未命中）。',
      '第二次查询同一 key 输出 cached（缓存命中）。',
      '使用 EXPIRE 设置了 TTL。',
    ],
    lesson: [
      'Redis 是高性能内存数据库，常用于缓存、会话存储和消息队列。',
      'go-redis 是 Go 社区使用最广的 Redis 客户端。',
      '缓存穿透模式：先查缓存 → 未命中查源 → 写入缓存。',
      'EXPIRE 设置过期时间，防止缓存雪崩和内存泄漏。',
      'Ping 是确认 Redis 连接可用的标准方法，类似 sql.DB.Ping。',
    ],
    mentorHints: [
      '先确认 rdb.Ping 成功（容器内 redis:6379 可访问）。',
      'rdb.Set(ctx, key, value, ttl) 设置值的同时可以指定过期时间。',
      'rdb.Get(ctx, key) 返回一个 string 和 error。',
      'redis.Nil 表示 key 不存在，不是连接错误。',
      '如果 REDIS_URL 读不到，检查 docker-compose 环境变量。',
    ],
    review: [
      '你完成了第一个 Redis 缓存练习。',
      '真实后端系统中，Redis 缓存可以将数据库 QPS 降低 10-100 倍。',
      '缓存穿透、缓存击穿、缓存雪崩是后端面试高频考点。',
      '下一步将学习并发编程，用 goroutine 和 channel 加速批量任务。',
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
        label: '缓存未命中',
        passDetail: 'stdout 中出现了 miss。',
        failDetail: 'stdout 中还没有看到 miss。',
        value: 'miss',
      },
      {
        type: 'stdoutIncludes',
        label: '缓存命中',
        passDetail: 'stdout 中出现了 cached。',
        failDetail: 'stdout 中还没有看到 cached。',
        value: 'cached',
      },
      {
        type: 'stdoutRegex',
        label: '缓存读写顺序',
        passDetail: '"miss" 出现在 "cached" 之前，说明先查缓存未命中再写入。',
        failDetail: '请确保先输出 miss 再输出 cached（先查 → 未命中 → 写入 → 命中）。',
        pattern: '(?s)miss.*cached',
      },
      {
        type: 'stderrExcludes',
        label: '无 Redis 错误',
        passDetail: 'stderr 中没有 Redis 连接错误。',
        failDetail: 'stderr 中包含 Redis 连接错误。',
        value: 'redis',
      },
    ],
  },
  {
    id: 'day-9-concurrency',
    day: 9,
    title: '并发编程基础',
    track: '并发编程',
    summary: '用 goroutine 和 channel 加速批量任务，学习并发编程核心模式。',
    background:
      '团队有一个批量处理任务：计算一组数字的平方。目前是串行执行，速度太慢。导师要求你用 goroutine 并发计算，用 channel 收集结果，最后用 WaitGroup 等待所有 worker 完成。',
    objective: '用 3 个 goroutine worker 并发生成平方数，通过 channel 收集结果，WaitGroup 等待完成，最终输出结果数量。',
    starterCode: `package main

import (
\t"fmt"
\t"sync"
)

func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
\tdefer wg.Done()
\tfor n := range jobs {
\t\tresults <- n * n
\t}
}

func main() {
\tnumbers := []int{2, 4, 6, 8, 10, 12, 14, 16, 18}

\tjobs := make(chan int, len(numbers))
\tresults := make(chan int, len(numbers))
\tvar wg sync.WaitGroup

\t// TODO: 启动 3 个 worker goroutine
\t// TODO: 发送所有数字到 jobs channel，然后关闭
\t// TODO: 等待所有 worker 完成
\t// TODO: 关闭 results channel
\t// TODO: 输出结果数量

\tfmt.Println("workers done")
}
`,
    criteria: [
      '程序可以成功运行，退出码为 0，无 deadlock 或 panic。',
      '输出结果数量应为 9（与输入数字数量一致）。',
      '使用了 sync.WaitGroup 等待所有 goroutine 完成。',
      '正确关闭了 jobs 和 results channel。',
    ],
    lesson: [
      'goroutine 是 Go 的轻量级并发单元，用 go func() 启动。',
      'channel 用于 goroutine 间通信：make(chan T) 无缓冲，make(chan T, n) 有缓冲。',
      'sync.WaitGroup 用于等待一组 goroutine 完成：Add/Done/Wait。',
      '关闭 channel 用 close(ch)，range 遍历会在 channel 关闭时自动退出。',
      '未关闭的 channel 可能导致 goroutine 泄漏或 deadlock。',
    ],
    mentorHints: [
      '先启动 worker goroutine，再发送任务——否则 jobs 满了会死锁（除非用有缓冲 channel）。',
      '发送完所有任务后 close(jobs) 通知 worker 没有更多任务。',
      'wg.Wait() 等待所有 worker 完成后，再 close(results)。',
      '可以用 for result := range results 收集结果。',
      '如果遇到 fatal error: all goroutines are asleep - deadlock，检查 channel 关闭顺序。',
    ],
    review: [
      '你完成了第一个并发编程练习。',
      'goroutine + channel + WaitGroup 是 Go 并发编程的三大核心原语。',
      '实际项目中，并发模式还有 worker pool、fan-out/fan-in、pipeline 等变体。',
      '下一步将学习日志、配置和可观测性，为生产环境做准备。',
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
        label: '输出结果数量',
        passDetail: '结果数量为 9。',
        failDetail: '结果数量不是 9，请检查并发逻辑。',
        pattern: '\\b9\\b',
      },
      {
        type: 'stderrExcludes',
        label: '无 deadlock',
        passDetail: 'stderr 中没有 deadlock 或 panic。',
        failDetail: 'stderr 中仍然包含 deadlock 或 panic。',
        value: 'fatal error:',
      },
    ],
  },
  {
    id: 'day-10-slog-logging',
    day: 10,
    title: '结构化日志',
    track: '日志与可观测性',
    summary: '用 log/slog 输出结构化 JSON 日志，替换 fmt.Println 打印。',
    background:
      '团队的服务已经上线，但排查线上问题时发现所有日志都是用 fmt.Println 打印的，没有时间戳、没有日志级别、无法用日志系统检索。导师要求你改用 Go 1.21+ 标准库的 log/slog 输出结构化 JSON 日志，包括 DEBUG、INFO、WARN、ERROR 四个级别。',
    objective: '用 slog JSON handler 输出结构化日志，至少包含 INFO、WARN、ERROR 三个级别的日志。',
    starterCode: `package main

import (
\t"log/slog"
\t"os"
)

type Record struct {
\tName  string
\tAge   int
\tEmail string
}

func process(records []Record) {
\t// TODO: 用 slog 替换下面的 fmt.Println
\t// - 每条记录处理时输出 DEBUG 级别日志
\t// - 所有记录处理完输出 INFO 级别日志，包含总数
\t// - 年龄 < 0 或 > 150 的记录输出 WARN 级别日志
\t// - Email 为空的记录输出 ERROR 级别日志

\t_ = records
}

func main() {
\tlogger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
\t\tLevel: slog.LevelDebug,
\t}))
\tslog.SetDefault(logger)

\trecords := []Record{
\t\t{"Alice", 30, "alice@example.com"},
\t\t{"Bob", 200, "bob@example.com"},
\t\t{"Charlie", 25, ""},
\t\t{"Diana", 35, "diana@example.com"},
\t}

\tprocess(records)
}
`,
    criteria: [
      '程序可以成功运行，退出码为 0。',
      '日志输出为 JSON 格式，可通过 jq 或 grep 解析。',
      '至少输出 INFO、WARN、ERROR 三个级别的日志。',
      'ERROR 日志用于 Email 为空的记录，WARN 用于年龄异常的记录。',
    ],
    lesson: [
      'log/slog 是 Go 1.21 引入的结构化日志标准库，向后兼容 log 包。',
      'slog.JSONHandler 输出 JSON 格式日志，可直接接入 ELK、Loki 等日志系统。',
      '日志级别从低到高：DEBUG < INFO < WARN < ERROR。生产环境通常设置 INFO。',
      '结构化的 key=value 字段比 fmt.Sprintf 拼接字符串更易于检索和分析。',
      'slog.SetDefault 设置默认 logger，后续 slog.Info/Warn/Error 会使用该 logger。',
    ],
    mentorHints: [
      'slog 支持结构化字段：slog.Info("msg", "key", value)。',
      'JSON Handler 输出形如 {"time":"...","level":"INFO","msg":"..."}。',
      'slog.Warn 和 slog.Error 的用法与 slog.Info 相同。',
      '使用 slog.Debug 记录详细调试信息，生产环境可关闭 DEBUG 级别。',
      '检查 stderr 确保没有 panic 或未捕获的错误。',
    ],
    review: [
      '你完成了结构化日志的任务。',
      '生产环境中，结构化日志是排障的第一道防线——因为时间和日志级别是可检索的。',
      'slog 还支持 TextHandler（可读文本）、自定义 Handler（写入文件/远程）。',
      '下一步将学习函数中间件模式，用 middleware 给处理函数添加日志和计时。',
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
        label: 'INFO 日志',
        passDetail: 'stdout 中出现了 INFO 级别的日志。',
        failDetail: 'stdout 中还没有看到 INFO 日志。',
        value: '"level":"INFO"',
      },
      {
        type: 'stdoutIncludes',
        label: 'WARN 日志',
        passDetail: 'stdout 中出现了 WARN 级别的日志。',
        failDetail: 'stdout 中还没有看到 WARN 日志。',
        value: '"level":"WARN"',
      },
      {
        type: 'stdoutIncludes',
        label: 'ERROR 日志',
        passDetail: 'stdout 中出现了 ERROR 级别的日志。',
        failDetail: 'stdout 中还没有看到 ERROR 日志。',
        value: '"level":"ERROR"',
      },
      {
        type: 'stderrExcludes',
        label: '无 panic',
        passDetail: 'stderr 中没有 panic。',
        failDetail: 'stderr 中包含 panic，请检查错误处理。',
        value: 'panic:',
      },
    ],
  },
  {
    id: 'day-11-middleware',
    day: 11,
    title: '函数中间件模式',
    track: '日志与可观测性',
    summary: '用 middleware 模式给函数添加日志和耗时统计，学习函数式组合。',
    background:
      '导师对你昨天的结构化日志很满意，现在要求你更进一步：别让业务代码直接写日志，而是用 middleware（中间件）模式把日志和计时逻辑从业务函数中分离。这样每个 middleware 只做一件事，组合起来就能给任何函数加上日志、计时、错误恢复等功能。',
    objective: '实现 logging middleware 和 timing middleware，组合后处理业务函数，输出日志和耗时。',
    starterCode: `package main

import (
\t"fmt"
\t"time"
)

type Processor func(data string) string

func LoggingMiddleware(name string, next Processor) Processor {
\treturn func(data string) string {
\t\t// TODO: 在处理前输出 "start: <name> <data>"
\t\t// 调用 next(data)
\t\t// 在处理后输出 "done: <name> <data>"
\t\treturn next(data)
\t}
}

func TimingMiddleware(next Processor) Processor {
\treturn func(data string) string {
\t\t// TODO: 记录开始时间
\t\t// 调用 next(data)
\t\t// 辕出 "duration: <elapsed>ms"
\t\treturn next(data)
\t}
}

func BusinessLogic(data string) string {
\ttime.Sleep(50 * time.Millisecond)
\treturn "processed: " + data
}

func main() {
\t// TODO: 组合 middleware（Logging → Timing → BusinessLogic）
\tpipeline := LoggingMiddleware("pipeline", TimingMiddleware(BusinessLogic))
\tresult := pipeline("hello")
\tfmt.Println("result:", result)
}
`,
    criteria: [
      '程序可以成功运行，退出码为 0。',
      'LoggingMiddleware 在处理前输出 "start:" 和处理后输出 "done:"。',
      'TimingMiddleware 输出耗时信息（包含 duration）。',
      '中间件组合后业务逻辑能得到正确结果。',
    ],
    lesson: [
      'Middleware 模式的核心：函数接受一个函数，返回一个增强后的函数（高阶函数）。',
      '每个 middleware 只做一件事（单一职责），可以自由组合成 pipeline。',
      'Go 中 middleware 常用在 HTTP handler（net/http）、gRPC interceptor、数据库访问层。',
      '函数签名统一（如 Processor）是实现 middleware 组合的关键。',
      '这个模式让你在不修改业务代码的情况下，给函数增加日志、计时、重试、限流等功能。',
    ],
    mentorHints: [
      'LoggingMiddleware 先执行日志逻辑，再调用 next(data)，最后打印完成日志。',
      'TimingMiddleware 用 time.Since(start) 计算耗时，输出为毫秒。',
      '组合顺序：LoggingMiddleware("pipeline", TimingMiddleware(BusinessLogic)) — 先 logging 后 timing。',
      '确保 LoggingMiddleware 打印时带上 name 参数以区分不同的 pipeline。',
      '如果 next 返回值和原数据有关，日志中可以体现出变化。',
    ],
    review: [
      '你完成了函数中间件模式的任务。',
      'middleware 模式是 Go 后端系统的核心设计模式之一，尤其在 HTTP 和 gRPC 中无处不在。',
      '生产环境中，典型的 middleware 链包括：日志 → 认证 → 限流 → 超时 → 业务逻辑。',
      '下一步将学习 context 超时与取消传播，为优雅关闭和可靠部署打下基础。',
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
        label: '开始日志',
        passDetail: 'stdout 中出现了 start 日志。',
        failDetail: 'stdout 中还没有看到 start 日志。',
        value: 'start:',
      },
      {
        type: 'stdoutIncludes',
        label: '完成日志',
        passDetail: 'stdout 中出现了 done 日志。',
        failDetail: 'stdout 中还没有看到 done 日志。',
        value: 'done:',
      },
      {
        type: 'stdoutIncludes',
        label: '耗时统计',
        passDetail: 'stdout 中出现了 duration 信息。',
        failDetail: 'stdout 中还没有看到耗时信息。',
        value: 'duration:',
      },
      {
        type: 'stdoutIncludes',
        label: '正确结果',
        passDetail: 'result 包含 processed。',
        failDetail: 'result 中没有 processed，请检查 pipeline 组合。',
        value: 'processed:',
      },
      {
        type: 'stderrExcludes',
        label: '无 panic',
        passDetail: 'stderr 中没有 panic。',
        failDetail: 'stderr 中包含 panic，请检查错误处理。',
        value: 'panic:',
      },
    ],
  },
  {
    id: 'day-12-context-timeout',
    day: 12,
    title: 'Context超时与取消',
    track: '部署与可靠性',
    summary: '用 context.WithTimeout 控制操作超时，实现取消传播。',
    background:
      '团队的外部调用偶尔会卡住很久不返回，导致整个请求超时。导师要求你实现一个任务管道，每一层都检查 context 是否已经超时或取消，如果超时就立即退出并输出清理日志。这就是 Go 的 context 超时传播机制。',
    objective: '用 context.WithTimeout 创建超时 context，在多层操作中检查 ctx.Done()，超时后输出超时日志。',
    starterCode: `package main

import (
\t"context"
\t"fmt"
\t"time"
)

func step1(ctx context.Context, name string) {
\t// TODO: 用 select 检查 ctx.Done()，如果取消输出 "step1 cancelled" 并返回
\ttime.Sleep(100 * time.Millisecond)
\tfmt.Println("step1 done:", name)
}

func step2(ctx context.Context, name string) {
\t// TODO: 用 select 检查 ctx.Done()，如果取消输出 "step2 cancelled" 并返回
\ttime.Sleep(150 * time.Millisecond)
\tfmt.Println("step2 done:", name)
}

func pipeline(ctx context.Context, name string) {
\tstep1(ctx, name)
\tstep2(ctx, name)
}

func main() {
\t// TODO: 用 context.WithTimeout 创建 80ms 超时的 context
\t// 超时时间短于 step1(100ms)+step2(150ms)=250ms，所以必然超时
\tctx := context.Background()

\t// TODO: 调用 pipeline(ctx, "task-1")

\t// TODO: 检查 ctx.Err() 是否为 context.DeadlineExceeded
\t// 如果超时，输出 "task timeout"

\ttime.Sleep(500 * time.Millisecond)
\tfmt.Println("program exit")
}
`,
    criteria: [
      '程序可以成功运行，退出码为 0。',
      '80ms 超时导致 pipeline 被取消，输出 task timeout。',
      'step1 或 step2 输出 cancelled（表示检测到了超时）。',
      '没有 goroutine 泄漏或未关闭的资源。',
    ],
    lesson: [
      'context.Context 是 Go 中传递取消信号、超时和请求范围值的标准方式。',
      'context.WithTimeout 创建一个在指定时间后自动取消的 context。',
      'ctx.Done() 返回一个 channel，当 context 被取消时该 channel 会关闭。',
      'select 语句用于同时监听多个 channel：select { case <-ctx.Done(): return }。',
      'context 超时是防止系统雪崩的关键机制——没有超时，慢服务可能拖垮整个系统。',
    ],
    mentorHints: [
      '用 select 配合 ctx.Done() 检查取消：select { case <-ctx.Done(): ... default: }。',
      'context.WithTimeout(parent, 80*time.Millisecond) 创建 80ms 超时 context。',
      '调用完 pipeline 后检查 ctx.Err() 判断是否超时（返回 context.DeadlineExceeded）。',
      'defer cancel() 确保 context 资源释放（即使不使用 cancelFunc 也要 defer）。',
      'step1 100ms + step2 150ms = 250ms > 80ms 超时，至少一个步骤会被取消。',
    ],
    review: [
      '你完成了 context 超时与取消的任务。',
      'context 超时是 Go 服务可靠性的基石——所有网络调用、数据库查询都应带超时。',
      '生产环境中，context 的超时通常设置为 P99 延迟的 2-3 倍。',
      '下一步将学习错误处理进阶：自定义错误类型、错误包装和 errors.Is/As。',
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
        label: '超时提示',
        passDetail: 'task timeout 已输出。',
        failDetail: '请确认创建了 80ms 超时 context 并检查了 ctx.Err()。',
        value: 'task timeout',
      },
      {
        type: 'stdoutIncludes',
        label: '步骤取消',
        passDetail: 'cancelled 已输出。',
        failDetail: '请确认在 step1 或 step2 中检查了 ctx.Done()。',
        value: 'cancelled',
      },
      {
        type: 'stderrExcludes',
        label: '无 panic',
        passDetail: 'stderr 中没有 panic。',
        failDetail: 'stderr 中包含 panic。',
        value: 'panic:',
      },
    ],
  },
  {
    id: 'day-13-error-handling',
    day: 13,
    title: '错误处理进阶',
    track: '部署与可靠性',
    summary: '用自定义错误类型、%w 包装和 errors.Is/As 实现业务错误链。',
    background:
      '团队的 API 经常返回 "internal error"，完全看不到出错原因。导师要求你重构错误处理：定义业务错误类型，用 fmt.Errorf("%w") 包装底层错误，在上层用 errors.As 判断错误类型，给用户返回不同的错误信息。',
    objective: '定义 ValidationError 和 NotFoundError，用 %w 包装底层错误，用 errors.As 判断错误类型并输出对应的处理信息。',
    starterCode: `package main

import (
\t"errors"
\t"fmt"
)

// TODO: 实现 ValidationError 的 Error() 方法
type ValidationError struct {
\tField string
\tMsg   string
}

func (e *ValidationError) Error() string {
\treturn ""
}

// TODO: 实现 NotFoundError 的 Error() 方法
type NotFoundError struct {
\tResource string
\tID       string
}

func (e *NotFoundError) Error() string {
\treturn ""
}

// TODO: 返回合适的错误
func validateAge(age int) error {
\t// age < 0: return &ValidationError{Field: "age", Msg: "must not be negative"}
\t// age > 150: return &ValidationError{Field: "age", Msg: "too large"}
\treturn nil
}

// TODO: 返回合适的错误
func findUser(id string) error {
\treturn nil
}

// TODO: 用 %w 包装底层错误
func processRequest(id string, age int) error {
\tif err := validateAge(age); err != nil {
\t\treturn nil
\t}
\tif err := findUser(id); err != nil {
\t\treturn nil
\t}
\treturn nil
}

func main() {
\terr := processRequest("user-1", -1)
\t// TODO: 用 errors.As 判断 ValidationError，输出 "validation failed: <field> <msg>"

\terr = processRequest("user-999", 30)
\t// TODO: 用 errors.As 判断 NotFoundError，输出 "not found: <resource> <id>"

\terr = processRequest("user-1", 30)
\tif err != nil {
\t\tfmt.Println("unexpected error:", err)
\t} else {
\t\tfmt.Println("all good")
\t}

\t_ = err
}
`,
    criteria: [
      '程序可以成功运行，退出码为 0。',
      'ValidationError 被正确识别：输出 "validation failed"。',
      'NotFoundError 被正确识别：输出 "not found"。',
      '使用了 errors.As 判断错误类型（而不是类型断言）。',
      '正常参数时输出 "all good"（无错误）。',
    ],
    lesson: [
      'Go 的错误只是实现了 Error() 方法的普通类型，不是异常。',
      'fmt.Errorf("%w", err) 包装错误，保留原始错误的类型信息（wrapping）。',
      'errors.Is(err, target) 判断错误链中是否包含特定 Sentinel Error。',
      'errors.As(err, &target) 提取错误链中某个类型的错误实例。',
      '自定义错误类型比字符串比较更语义化、更安全（不会被重名误导）。',
    ],
    mentorHints: [
      'ValidationError 需要 Error() 方法才能实现 error 接口。',
      '在 processRequest 中用 fmt.Errorf("validate failed: %w", err) 包装错误。',
      'errors.As 使用方式：var ve *ValidationError; if errors.As(err, &ve) { ... }。',
      '注意 errors.As 的第二个参数是指向指针的指针（&ve 其中 ve 是 *ValidationError）。',
      '如果要同时输出底层错误信息，可以用 fmt.Println(err) 打印完整错误链。',
    ],
    review: [
      '你完成了错误处理进阶的任务。',
      '良好的错误处理是生产级 Go 服务的标志。',
      'Go 1.13+ 的 %w 和 errors.Is/As 让你在不丢失类型信息的情况下添加上下文。',
      '恭喜完成第二阶段全部 8 个工程进阶任务！',
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
        label: 'Validation 错误识别',
        passDetail: 'validation failed 被正确输出。',
        failDetail: '请检查 errors.As 判断 ValidationError。',
        value: 'validation failed:',
      },
      {
        type: 'stdoutIncludes',
        label: 'Not Found 错误识别',
        passDetail: 'not found 被正确输出。',
        failDetail: '请检查 errors.As 判断 NotFoundError。',
        value: 'not found:',
      },
      {
        type: 'stdoutIncludes',
        label: '正常路径',
        passDetail: 'all good 已输出。',
        failDetail: '请确认正常参数不返回错误。',
        value: 'all good',
      },
      {
        type: 'stderrExcludes',
        label: '无 panic',
        passDetail: 'stderr 中没有 panic。',
        failDetail: 'stderr 中包含 panic。',
        value: 'panic:',
      },
    ],
  },
];

export const defaultTaskId = internshipTasks[0].id;

export function findTaskById(id: string): InternTask {
  return internshipTasks.find((task) => task.id === id) ?? internshipTasks[0];
}