import { missionCatalog, type Mission } from "./missions";

export type GoCourseDifficulty = "入门" | "基础" | "进阶" | "高级";
export type GoCourseOutputMatch = "trimmed-exact" | "contains";

export type GoCourseConcept = {
  name: string;
  explanation: string;
  pitfall: string;
};

export type GoCourseExercise = {
  title: string;
  prompt: string;
  starterCode: string;
  expectedOutput: string;
  outputMatch: GoCourseOutputMatch;
  hints: string[];
};

export type GoCourseChapter = {
  slug: string;
  order: number;
  title: string;
  sourcePath: string;
  sourceUrl: string;
  duration: string;
  difficulty: GoCourseDifficulty;
  summary: string;
  goals: string[];
  concepts: GoCourseConcept[];
  exercise: GoCourseExercise;
  checklist: string[];
  nextMissionSlugs: string[];
};

const sourceUrl = (sourcePath: string) => `https://gopl-zh.github.io/${sourcePath}`;

export const goBasicsChapters: GoCourseChapter[] = [
  {
    slug: "ch1-getting-started",
    order: 1,
    title: "入门",
    sourcePath: "ch1/ch1.md",
    sourceUrl: sourceUrl("ch1/ch1.md"),
    duration: "约 45 分钟",
    difficulty: "入门",
    summary: "从第一段 Go 程序开始，建立命令行、输入输出、HTTP 请求和简单并发的直觉，为后续任务准备运行 Go 代码的基本手感。",
    goals: [
      "能写出并运行一个最小 Go 程序。",
      "能区分命令行参数、标准输入和网络响应。",
      "能解释 go run、package main 和 func main 的关系。",
      "能读懂简单循环、条件判断和 fmt 输出。",
    ],
    concepts: [
      {
        name: "package main",
        explanation: "可执行程序从 main 包中的 main 函数启动。",
        pitfall: "文件能编译不代表能直接运行；库包通常没有 main 入口。",
      },
      {
        name: "标准输入输出",
        explanation: "fmt、os.Stdin 和命令行参数是最早接触程序边界的方式。",
        pitfall: "调试时不要把真实业务输入硬编码进函数内部。",
      },
      {
        name: "错误优先",
        explanation: "Go 鼓励显式检查错误，让失败路径和成功路径一样清楚。",
        pitfall: "忽略 err 会把真实问题延迟到更难定位的位置。",
      },
      {
        name: "轻量并发直觉",
        explanation: "goroutine 让函数可以并发执行，但启动并发不等于自动等待完成。",
        pitfall: "没有同步手段时，main 函数退出会让后台 goroutine 来不及完成。",
      },
    ],
    exercise: {
      title: "打印入职欢迎信息",
      prompt: "运行一段最小 Go 程序，确认你能通过 sandbox 得到稳定输出。",
      starterCode: `package main

import "fmt"

func main() {
    name := "Gopher"
    fmt.Printf("welcome, %s\n", name)
}`,
      expectedOutput: "welcome, Gopher",
      outputMatch: "contains",
      hints: ["程序必须声明 package main。", "fmt.Printf 不会自动换行，必要时写入 \\n。"],
    },
    checklist: ["程序能成功编译。", "stdout 中包含 welcome, Gopher。", "能解释 main 函数何时执行。"],
    nextMissionSlugs: [],
  },
  {
    slug: "ch2-program-structure",
    order: 2,
    title: "程序结构",
    sourcePath: "ch2/ch2.md",
    sourceUrl: sourceUrl("ch2/ch2.md"),
    duration: "约 50 分钟",
    difficulty: "入门",
    summary: "学习声明、变量、赋值、包和作用域，理解 Go 代码如何被组织成可读、可维护的程序单元。",
    goals: [
      "能区分 var、const、type 和 func 声明。",
      "能说明短变量声明适合出现在哪里。",
      "能解释包级作用域和函数级作用域。",
      "能把命名改得更符合 Go 代码习惯。",
    ],
    concepts: [
      {
        name: "声明",
        explanation: "声明把名字绑定到变量、常量、类型或函数。",
        pitfall: "把所有东西都放到包级变量会增加隐藏状态和测试难度。",
      },
      {
        name: "短变量声明",
        explanation: ":= 适合在函数内部声明局部变量。",
        pitfall: "在嵌套作用域中误用 := 可能创建新变量，而不是更新外层变量。",
      },
      {
        name: "零值",
        explanation: "每种类型都有可用的默认值，减少了未初始化状态。",
        pitfall: "零值可用不代表零值一定符合业务语义。",
      },
      {
        name: "作用域",
        explanation: "名字只在其声明可见范围内有效，越小的作用域越容易推理。",
        pitfall: "遮蔽外层变量会让错误处理和返回值变得难读。",
      },
    ],
    exercise: {
      title: "修正变量作用域",
      prompt: "运行代码并观察局部变量如何影响最终输出。",
      starterCode: `package main

import "fmt"

func main() {
    level := 1
    passed := true
    if passed {
        level = level + 1
    }
    fmt.Printf("level=%d\n", level)
}`,
      expectedOutput: "level=2",
      outputMatch: "contains",
      hints: ["如果在 if 中使用 :=，可能会创建新的局部变量。", "本练习希望更新外层 level。"],
    },
    checklist: ["能解释 level 为什么最终是 2。", "能指出 := 和 = 的使用差异。", "能避免不必要的包级变量。"],
    nextMissionSlugs: [],
  },
  {
    slug: "ch3-basic-data-types",
    order: 3,
    title: "基础数据类型",
    sourcePath: "ch3/ch3.md",
    sourceUrl: sourceUrl("ch3/ch3.md"),
    duration: "约 55 分钟",
    difficulty: "基础",
    summary: "掌握数字、布尔、字符串和常量的基本行为，建立处理接口参数、日志文本和配置值时需要的类型意识。",
    goals: [
      "能选择合适的整数和浮点类型。",
      "能解释字符串、字节和 rune 的关系。",
      "能使用常量表达固定配置或枚举语义。",
      "能识别常见类型转换错误。",
    ],
    concepts: [
      {
        name: "整数与浮点数",
        explanation: "数值类型决定范围、精度和计算行为。",
        pitfall: "把金额、计数和比例都用 float64 会制造精度和比较问题。",
      },
      {
        name: "字符串",
        explanation: "字符串是只读字节序列，常用于存储 UTF-8 文本。",
        pitfall: "用 len 统计中文字符数量会得到字节长度，不是字符数量。",
      },
      {
        name: "rune",
        explanation: "rune 表示 Unicode 码点，适合按字符遍历文本。",
        pitfall: "按 byte 切分 UTF-8 字符串可能截断多字节字符。",
      },
      {
        name: "常量",
        explanation: "常量在编译期确定，适合表达稳定的业务含义。",
        pitfall: "把会变化的配置写成常量会让部署调整变困难。",
      },
    ],
    exercise: {
      title: "统计中文昵称长度",
      prompt: "比较字节长度和字符数量，理解字符串与 rune 的差异。",
      starterCode: `package main

import "fmt"

func main() {
    nickname := "小Gopher"
    fmt.Printf("bytes=%d runes=%d\n", len(nickname), len([]rune(nickname)))
}`,
      expectedOutput: "bytes=10 runes=7",
      outputMatch: "contains",
      hints: ["中文字符在 UTF-8 中通常占多个字节。", "[]rune 可以按 Unicode 码点统计。"],
    },
    checklist: ["能解释 bytes 与 runes 的差异。", "能说出什么时候需要 []rune。", "能避免按字节截断用户文本。"],
    nextMissionSlugs: [],
  },
  {
    slug: "ch4-composite-types",
    order: 4,
    title: "复合数据类型",
    sourcePath: "ch4/ch4.md",
    sourceUrl: sourceUrl("ch4/ch4.md"),
    duration: "约 70 分钟",
    difficulty: "基础",
    summary: "学习数组、slice、map、struct 和 JSON，掌握后端服务中最常见的数据建模与集合处理方式。",
    goals: [
      "能解释数组和 slice 的差异。",
      "能用 map 完成计数和索引。",
      "能用 struct 表达业务数据。",
      "能识别 slice 共享底层数组带来的资源保留风险。",
    ],
    concepts: [
      {
        name: "slice",
        explanation: "slice 是对底层数组片段的描述，包含指针、长度和容量。",
        pitfall: "小 slice 可能仍引用大数组，导致内存不能及时释放。",
      },
      {
        name: "map",
        explanation: "map 适合按 key 快速查询、计数和去重。",
        pitfall: "原生 map 不能在无同步保护下并发写入。",
      },
      {
        name: "struct",
        explanation: "struct 把相关字段组织成一个明确的数据模型。",
        pitfall: "字段命名和导出规则会影响 JSON 编码和跨包访问。",
      },
      {
        name: "JSON",
        explanation: "JSON 是后端接口最常见的数据交换格式之一。",
        pitfall: "忽略编码/解码错误会让脏数据悄悄进入业务流程。",
      },
    ],
    exercise: {
      title: "统计订单状态",
      prompt: "用 map 统计订单状态数量，模拟接口聚合数据的基础逻辑。",
      starterCode: `package main

import "fmt"

func main() {
    statuses := []string{"paid", "pending", "paid", "failed", "paid", "pending"}
    counts := map[string]int{}
    for _, status := range statuses {
        counts[status]++
    }
    fmt.Printf("pending=%d paid=%d failed=%d\n", counts["pending"], counts["paid"], counts["failed"])
}`,
      expectedOutput: "pending=2 paid=3 failed=1",
      outputMatch: "contains",
      hints: ["map 的零值读取会返回对应值类型的零值。", "统计时可以直接对 counts[status]++。"],
    },
    checklist: ["使用 map 完成计数。", "输出包含 pending=2 paid=3 failed=1。", "能解释 slice 遍历中的 index/value。"],
    nextMissionSlugs: ["slice-memory-leak"],
  },
  {
    slug: "ch5-functions",
    order: 5,
    title: "函数",
    sourcePath: "ch5/ch5.md",
    sourceUrl: sourceUrl("ch5/ch5.md"),
    duration: "约 65 分钟",
    difficulty: "基础",
    summary: "学习函数声明、多返回值、错误处理、匿名函数、可变参数、defer、panic 和 recover，建立可组合的业务逻辑单元。",
    goals: [
      "能写出清晰的函数签名。",
      "能用多返回值表达结果和错误。",
      "能解释 defer 的执行时机。",
      "能避免滥用 panic 处理普通业务错误。",
    ],
    concepts: [
      {
        name: "多返回值",
        explanation: "Go 常用多个返回值同时表达结果和错误。",
        pitfall: "返回值过多会让调用方难以理解，必要时应引入结构体。",
      },
      {
        name: "error",
        explanation: "error 是普通值，调用方需要显式判断。",
        pitfall: "吞掉错误会让上层无法区分成功和失败。",
      },
      {
        name: "defer",
        explanation: "defer 在函数返回前按后进先出顺序执行。",
        pitfall: "defer 修改命名返回值可能覆盖真正的业务错误。",
      },
      {
        name: "匿名函数",
        explanation: "匿名函数适合封装局部逻辑或闭包状态。",
        pitfall: "闭包捕获循环变量时需要确认每次迭代的变量语义。",
      },
    ],
    exercise: {
      title: "解析任务耗时",
      prompt: "用多返回值返回解析结果和错误，练习显式错误处理。",
      starterCode: `package main

import (
    "fmt"
    "strconv"
)

func main() {
    minutes, err := parseMinutes("45")
    if err != nil {
        fmt.Println("invalid duration")
        return
    }
    fmt.Printf("duration=%dmin\n", minutes)
}

func parseMinutes(input string) (int, error) {
    return strconv.Atoi(input)
}`,
      expectedOutput: "duration=45min",
      outputMatch: "contains",
      hints: ["strconv.Atoi 会返回 int 和 error。", "调用方应先判断 err。"],
    },
    checklist: ["函数返回结果和 error。", "调用方显式检查错误。", "能说明 defer 与 return 的执行顺序。"],
    nextMissionSlugs: ["defer-order"],
  },
  {
    slug: "ch6-methods",
    order: 6,
    title: "方法",
    sourcePath: "ch6/ch6.md",
    sourceUrl: sourceUrl("ch6/ch6.md"),
    duration: "约 55 分钟",
    difficulty: "基础",
    summary: "理解方法、接收者、指针方法、嵌入和封装，学会把行为放到合适的数据类型旁边。",
    goals: [
      "能为自定义类型定义方法。",
      "能区分值接收者和指针接收者。",
      "能解释嵌入结构体带来的方法提升。",
      "能通过未导出字段保护类型不变量。",
    ],
    concepts: [
      {
        name: "方法接收者",
        explanation: "方法通过接收者绑定到某个类型。",
        pitfall: "接收者命名过长或过短都会降低可读性。",
      },
      {
        name: "指针接收者",
        explanation: "需要修改对象或避免复制大对象时通常使用指针接收者。",
        pitfall: "混用值接收者和指针接收者会让方法集更难推理。",
      },
      {
        name: "嵌入",
        explanation: "嵌入让类型组合更轻量，常用于复用字段和方法。",
        pitfall: "过度嵌入会让字段来源不清晰。",
      },
      {
        name: "封装",
        explanation: "通过导出和未导出名字控制跨包访问边界。",
        pitfall: "把所有字段都导出会让外部代码绕过校验逻辑。",
      },
    ],
    exercise: {
      title: "给积分账户增加方法",
      prompt: "为结构体定义方法，练习指针接收者修改内部状态。",
      starterCode: `package main

import "fmt"

type Account struct {
    Points int
}

func (a *Account) Add(points int) {
    a.Points += points
}

func main() {
    account := Account{Points: 10}
    account.Add(15)
    fmt.Printf("points=%d\n", account.Points)
}`,
      expectedOutput: "points=25",
      outputMatch: "contains",
      hints: ["Add 需要修改 Account，因此使用指针接收者。", "Go 会在可取地址的值上自动调用指针方法。"],
    },
    checklist: ["方法绑定到 Account 类型。", "Add 使用指针接收者。", "输出 points=25。"],
    nextMissionSlugs: [],
  },
  {
    slug: "ch7-interfaces",
    order: 7,
    title: "接口",
    sourcePath: "ch7/ch7.md",
    sourceUrl: sourceUrl("ch7/ch7.md"),
    duration: "约 75 分钟",
    difficulty: "进阶",
    summary: "学习接口作为合约的用法，理解隐式实现、接口值、类型断言和常见标准库接口。",
    goals: [
      "能定义小而稳定的接口。",
      "能解释 Go 的隐式接口实现。",
      "能判断接口值是否为 nil。",
      "能用接口隔离调用方和具体实现。",
    ],
    concepts: [
      {
        name: "接口是合约",
        explanation: "接口描述行为，具体类型只要拥有对应方法就满足接口。",
        pitfall: "提前设计大接口会让实现者被迫依赖不需要的方法。",
      },
      {
        name: "隐式实现",
        explanation: "类型不需要声明 implements，只要方法集匹配即可。",
        pitfall: "方法接收者不同可能导致值类型和指针类型满足接口的结果不同。",
      },
      {
        name: "接口值",
        explanation: "接口值包含动态类型和动态值。",
        pitfall: "带有动态类型的 nil 指针接口值并不等于 nil。",
      },
      {
        name: "类型断言",
        explanation: "类型断言用于从接口值中取回具体能力或类型。",
        pitfall: "单返回值断言失败会 panic，业务代码通常使用双返回值形式。",
      },
    ],
    exercise: {
      title: "用接口统一通知渠道",
      prompt: "定义一个小接口，让不同通知实现可以被同一个函数调用。",
      starterCode: `package main

import "fmt"

type Notifier interface {
    Notify(message string) string
}

type EmailNotifier struct{}

func (EmailNotifier) Notify(message string) string {
    return "email:" + message
}

func send(n Notifier, message string) {
    fmt.Println(n.Notify(message))
}

func main() {
    send(EmailNotifier{}, "build passed")
}`,
      expectedOutput: "email:build passed",
      outputMatch: "contains",
      hints: ["接口只需要描述调用方真正依赖的方法。", "EmailNotifier 不需要显式声明实现了 Notifier。"],
    },
    checklist: ["接口只包含 Notify。", "EmailNotifier 隐式满足接口。", "send 依赖接口而非具体类型。"],
    nextMissionSlugs: [],
  },
  {
    slug: "ch8-goroutines-channels",
    order: 8,
    title: "Goroutines 和 Channels",
    sourcePath: "ch8/ch8.md",
    sourceUrl: sourceUrl("ch8/ch8.md"),
    duration: "约 80 分钟",
    difficulty: "进阶",
    summary: "掌握 goroutine、channel、select、退出控制和并发循环，为服务端任务调度和 I/O 并发打基础。",
    goals: [
      "能启动 goroutine 并等待结果。",
      "能用 channel 传递数据和关闭信号。",
      "能解释无缓冲和有缓冲 channel 的差异。",
      "能避免 goroutine 泄漏。",
    ],
    concepts: [
      {
        name: "goroutine",
        explanation: "goroutine 是 Go 的轻量并发执行单元。",
        pitfall: "启动 goroutine 后不管理生命周期会造成泄漏。",
      },
      {
        name: "channel",
        explanation: "channel 用于 goroutine 之间通信和同步。",
        pitfall: "向无人接收的无缓冲 channel 发送会永久阻塞。",
      },
      {
        name: "select",
        explanation: "select 可以等待多个 channel 操作。",
        pitfall: "空 select 会永久阻塞当前 goroutine。",
      },
      {
        name: "退出控制",
        explanation: "并发程序需要明确的取消、关闭或超时机制。",
        pitfall: "只关注启动，不关注退出，是很多资源泄漏的根源。",
      },
    ],
    exercise: {
      title: "并发收集两个任务结果",
      prompt: "用 channel 等待两个 goroutine 的输出，理解并发结果汇聚。",
      starterCode: `package main

import "fmt"

func main() {
    results := make(chan string, 2)
    go func() { results <- "cache" }()
    go func() { results <- "db" }()

    first := <-results
    second := <-results
    fmt.Printf("results=%s,%s\n", first, second)
}`,
      expectedOutput: "results=",
      outputMatch: "contains",
      hints: ["两个 goroutine 的完成顺序不固定。", "缓冲为 2 的 channel 可以容纳两个结果。"],
    },
    checklist: ["启动两个 goroutine。", "用 channel 收集两个结果。", "能解释输出顺序为什么可能变化。"],
    nextMissionSlugs: ["slice-memory-leak"],
  },
  {
    slug: "ch9-shared-variable-concurrency",
    order: 9,
    title: "基于共享变量的并发",
    sourcePath: "ch9/ch9.md",
    sourceUrl: sourceUrl("ch9/ch9.md"),
    duration: "约 75 分钟",
    difficulty: "进阶",
    summary: "理解竞争条件、互斥锁、读写锁、内存同步和 race detector，学会保护共享状态。",
    goals: [
      "能解释什么是数据竞争。",
      "能用 mutex 保护共享变量。",
      "能判断何时需要 RWMutex 或 sync.Once。",
      "能用 race detector 验证并发安全。",
    ],
    concepts: [
      {
        name: "竞争条件",
        explanation: "多个 goroutine 访问同一数据且至少一个写入时，需要同步保护。",
        pitfall: "偶尔运行正确并不代表没有数据竞争。",
      },
      {
        name: "sync.Mutex",
        explanation: "互斥锁让临界区一次只被一个 goroutine 执行。",
        pitfall: "忘记 Unlock 或锁住太大范围都会影响正确性和性能。",
      },
      {
        name: "sync.RWMutex",
        explanation: "读写锁适合读多写少的共享数据。",
        pitfall: "写操作频繁时 RWMutex 不一定比 Mutex 更好。",
      },
      {
        name: "race detector",
        explanation: "go test -race 可以帮助发现数据竞争。",
        pitfall: "race detector 是验证工具，不是同步机制本身。",
      },
    ],
    exercise: {
      title: "安全累加请求数",
      prompt: "用 mutex 保护共享计数器，避免并发写入造成结果不稳定。",
      starterCode: `package main

import (
    "fmt"
    "sync"
)

func main() {
    var mu sync.Mutex
    var wg sync.WaitGroup
    total := 0

    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            mu.Lock()
            total++
            mu.Unlock()
        }()
    }

    wg.Wait()
    fmt.Printf("total=%d\n", total)
}`,
      expectedOutput: "total=100",
      outputMatch: "contains",
      hints: ["共享变量 total 的写入需要放在临界区。", "defer wg.Done() 只负责等待，不负责保护数据。"],
    },
    checklist: ["使用 mutex 保护 total++。", "使用 WaitGroup 等待所有 goroutine。", "输出 total=100。"],
    nextMissionSlugs: ["map-concurrent-write"],
  },
  {
    slug: "ch10-packages-tools",
    order: 10,
    title: "包和工具",
    sourcePath: "ch10/ch10.md",
    sourceUrl: sourceUrl("ch10/ch10.md"),
    duration: "约 60 分钟",
    difficulty: "基础",
    summary: "理解包、导入路径、命名和 Go 工具链，知道如何把代码组织成可被团队协作维护的模块。",
    goals: [
      "能解释包名和导入路径的区别。",
      "能使用 gofmt、go test 和 go list 等工具。",
      "能设计简洁的包名。",
      "能识别匿名导入的适用场景。",
    ],
    concepts: [
      {
        name: "包",
        explanation: "包是 Go 代码复用和封装的基本单元。",
        pitfall: "包名不应重复上层路径语义，例如 util、common 容易变成垃圾桶。",
      },
      {
        name: "导入路径",
        explanation: "导入路径定位包的位置，包名是代码中引用的名字。",
        pitfall: "导入路径和 package 名不一致时需要读者额外记忆。",
      },
      {
        name: "gofmt",
        explanation: "gofmt 统一格式，减少代码评审中的风格争论。",
        pitfall: "手工对齐格式通常会被 gofmt 覆盖。",
      },
      {
        name: "工具链",
        explanation: "go run、go build、go test、go env 是日常开发的基础工具。",
        pitfall: "不了解工具输出会让构建和测试问题变得难排查。",
      },
    ],
    exercise: {
      title: "查看工具链环境",
      prompt: "运行一个小程序输出当前平台信息，理解工具链和运行环境相关性。",
      starterCode: `package main

import (
    "fmt"
    "runtime"
)

func main() {
    fmt.Printf("goos=%s goarch=%s\n", runtime.GOOS, runtime.GOARCH)
}`,
      expectedOutput: "goos=",
      outputMatch: "contains",
      hints: ["runtime 包能读取当前 Go 运行环境信息。", "不同 sandbox 平台输出可能不同，所以这里只检查前缀。"],
    },
    checklist: ["程序能输出 goos 和 goarch。", "能说明包名和导入路径的区别。", "能列举至少三个常用 go 命令。"],
    nextMissionSlugs: [],
  },
  {
    slug: "ch11-testing",
    order: 11,
    title: "测试",
    sourcePath: "ch11/ch11.md",
    sourceUrl: sourceUrl("ch11/ch11.md"),
    duration: "约 75 分钟",
    difficulty: "进阶",
    summary: "学习 go test、测试函数、覆盖率、基准测试和示例函数，建立用测试保护行为的习惯。",
    goals: [
      "能写出基本测试函数。",
      "能理解表驱动测试的价值。",
      "能解释覆盖率和基准测试适用场景。",
      "能把失败信息写得便于定位问题。",
    ],
    concepts: [
      {
        name: "测试函数",
        explanation: "以 Test 开头并接收 *testing.T 的函数会被 go test 识别。",
        pitfall: "只测试 happy path 会遗漏真实服务中更常见的错误分支。",
      },
      {
        name: "表驱动测试",
        explanation: "用一组测试用例驱动同一段测试逻辑，适合覆盖边界。",
        pitfall: "测试表字段命名不清会让失败信息难读。",
      },
      {
        name: "覆盖率",
        explanation: "覆盖率显示哪些代码被测试执行过。",
        pitfall: "高覆盖率不等于高质量测试。",
      },
      {
        name: "基准测试",
        explanation: "Benchmark 用于观察性能变化。",
        pitfall: "没有稳定输入和环境时，基准数字容易误导判断。",
      },
    ],
    exercise: {
      title: "运行表驱动思路",
      prompt: "在 main 中模拟一组测试用例，理解表驱动测试的结构。",
      starterCode: `package main

import "fmt"

func add(a, b int) int {
    return a + b
}

func main() {
    cases := []struct {
        a, b int
        want int
    }{
        {1, 2, 3},
        {2, 3, 5},
    }

    passed := 0
    for _, tc := range cases {
        if add(tc.a, tc.b) == tc.want {
            passed++
        }
    }
    fmt.Printf("passed=%d/%d\n", passed, len(cases))
}`,
      expectedOutput: "passed=2/2",
      outputMatch: "contains",
      hints: ["真实项目中这些 case 通常放在 *_test.go 中。", "本练习在单文件 main 中模拟测试结构，方便 sandbox 运行。"],
    },
    checklist: ["能识别测试表字段。", "输出 passed=2/2。", "能说明真实测试应由 go test 运行。"],
    nextMissionSlugs: [],
  },
  {
    slug: "ch12-reflection",
    order: 12,
    title: "反射",
    sourcePath: "ch12/ch12.md",
    sourceUrl: sourceUrl("ch12/ch12.md"),
    duration: "约 70 分钟",
    difficulty: "高级",
    summary: "理解 reflect.Type、reflect.Value、结构体标签和反射修改值的边界，知道框架如何在运行时处理未知类型。",
    goals: [
      "能解释为什么需要反射。",
      "能区分 reflect.Type 和 reflect.Value。",
      "能读取结构体字段标签。",
      "能判断反射是否值得引入。",
    ],
    concepts: [
      {
        name: "reflect.Type",
        explanation: "Type 描述值的静态或动态类型信息。",
        pitfall: "依赖字符串形式的类型名做业务判断通常很脆弱。",
      },
      {
        name: "reflect.Value",
        explanation: "Value 表示运行时值，可用于读取或在满足条件时修改。",
        pitfall: "不可设置的 Value 调用 Set 会 panic。",
      },
      {
        name: "结构体标签",
        explanation: "标签常用于 JSON、数据库和校验框架的字段映射。",
        pitfall: "标签写错通常不会被编译器发现。",
      },
      {
        name: "反射成本",
        explanation: "反射提高通用性，但牺牲可读性、类型安全和部分性能。",
        pitfall: "能用接口和普通函数解决的问题，不必优先使用反射。",
      },
    ],
    exercise: {
      title: "读取 JSON 标签",
      prompt: "用反射读取结构体字段标签，理解框架如何发现字段元数据。",
      starterCode: `package main

import (
    "fmt"
    "reflect"
)

type User struct {
    ID   int    \`json:"id"\`
    Name string \`json:"name"\`
}

func main() {
    t := reflect.TypeOf(User{})
    field, _ := t.FieldByName("Name")
    fmt.Printf("json=%s\n", field.Tag.Get("json"))
}`,
      expectedOutput: "json=name",
      outputMatch: "contains",
      hints: ["reflect.TypeOf(User{}) 可以得到结构体类型。", "StructTag.Get 用于按 key 读取标签值。"],
    },
    checklist: ["能读取 Name 字段的 json 标签。", "输出 json=name。", "能说出反射的代价。"],
    nextMissionSlugs: [],
  },
  {
    slug: "ch13-low-level-programming",
    order: 13,
    title: "底层编程",
    sourcePath: "ch13/ch13.md",
    sourceUrl: sourceUrl("ch13/ch13.md"),
    duration: "约 60 分钟",
    difficulty: "高级",
    summary: "了解 unsafe、内存布局、cgo 和底层能力的边界，知道什么时候应该远离这些锋利工具。",
    goals: [
      "能说明 unsafe 的主要用途和风险。",
      "能读取类型大小和对齐信息。",
      "能理解 cgo 带来的构建与运行成本。",
      "能判断底层优化是否值得。",
    ],
    concepts: [
      {
        name: "unsafe.Sizeof",
        explanation: "Sizeof 可以观察值在内存中的大小。",
        pitfall: "内存大小与序列化大小、业务数据大小不是同一个概念。",
      },
      {
        name: "unsafe.Pointer",
        explanation: "unsafe.Pointer 允许绕过类型系统进行底层转换。",
        pitfall: "错误使用可能破坏内存安全和垃圾回收假设。",
      },
      {
        name: "内存对齐",
        explanation: "字段顺序会影响结构体占用空间。",
        pitfall: "为了省几个字节牺牲可读性通常不值得。",
      },
      {
        name: "cgo",
        explanation: "cgo 允许调用 C 代码。",
        pitfall: "cgo 会增加构建、跨平台和调试复杂度。",
      },
    ],
    exercise: {
      title: "观察结构体大小",
      prompt: "用 unsafe.Sizeof 观察结构体占用空间，建立内存布局直觉。",
      starterCode: `package main

import (
    "fmt"
    "unsafe"
)

type Metric struct {
    Active bool
    Count  int64
}

func main() {
    fmt.Printf("size=%d\n", unsafe.Sizeof(Metric{}))
}`,
      expectedOutput: "size=16",
      outputMatch: "contains",
      hints: ["结构体大小会受到字段对齐影响。", "在常见 64 位平台上该结构体通常为 16 字节。"],
    },
    checklist: ["能运行 unsafe.Sizeof。", "能解释字段对齐会影响大小。", "能说明 unsafe 不应作为日常首选工具。"],
    nextMissionSlugs: [],
  },
];

export const goBasicsCatalog = Object.fromEntries(goBasicsChapters.map((chapter) => [chapter.slug, chapter])) as Record<string, GoCourseChapter>;

export function getGoBasicsChapterBySlug(slug?: string | null) {
  return slug ? goBasicsCatalog[slug] : undefined;
}

export function getRelatedMissions(chapter: GoCourseChapter): Mission[] {
  return chapter.nextMissionSlugs.map((slug) => missionCatalog[slug]).filter((mission): mission is Mission => Boolean(mission));
}

export function validateGoBasicsCourse() {
  const slugs = new Set<string>();
  const errors: string[] = [];

  if (goBasicsChapters.length !== 13) {
    errors.push(`expected 13 chapters, got ${goBasicsChapters.length}`);
  }

  for (const chapter of goBasicsChapters) {
    if (slugs.has(chapter.slug)) {
      errors.push(`duplicate chapter slug: ${chapter.slug}`);
    }
    slugs.add(chapter.slug);

    if (!chapter.sourceUrl.startsWith("https://gopl-zh.github.io/")) {
      errors.push(`invalid sourceUrl root: ${chapter.slug}`);
    }

    if (!chapter.title || !chapter.summary || !chapter.exercise.starterCode || !chapter.exercise.expectedOutput) {
      errors.push(`missing required chapter data: ${chapter.slug}`);
    }

    for (const missionSlug of chapter.nextMissionSlugs) {
      if (!missionCatalog[missionSlug]) {
        errors.push(`unknown mission slug ${missionSlug} in ${chapter.slug}`);
      }
    }
  }

  return errors;
}
