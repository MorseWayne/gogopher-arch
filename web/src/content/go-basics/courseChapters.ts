import type { GoBasicsMdxChapter } from "../../app/data/goBasicsCourse";

export const mdxGoBasicsChapterMetadata: GoBasicsMdxChapter[] = [
{
  slug: "ch1-getting-started",
  order: 1,
  title: "入门",
  duration: "约 70 分钟",
  difficulty: "入门",
  summary: "基于《Go 程序设计语言》第 1 章重组：从 Hello World、go run/go build、package main、import、stdout 与基础格式化习惯建立第一批 Go 程序运行模型。",
  goals: [
    "能解释 package main、func main 与可执行程序入口的关系。",
    "能使用 go run 和 go build 理解编译、链接、运行之间的关系。",
    "能说明 import、fmt.Println、Unicode 输出和 gofmt 的基础作用。",
    "能把 stdout、stderr 和 exit code 当成观察程序行为的第一批信号。",
  ],
  lessonCount: 6,
  lessons: [],
  modernNotes: [
    { title: "从命令行反馈回路开始", body: "即使未来使用 IDE、容器和 CI，Go 初学者仍应先掌握 go run、go build、stdout、stderr 和 exit code。它们是排查服务启动失败、脚本异常和 CI 构建失败的共同语言。" },
    { title: "保留教材脉络，但面向当前工程实践重组", body: "本章以 gopl-zh 第 1 章的 Hello World 和工具链说明为底稿，删去过时的 GOPATH 获取流程，把重点放在 module 时代仍稳定的 package、import、main、gofmt 和可观察输出。" },
  ],
  engineeringPractices: [
    "先运行最小程序，再逐步增加输入、分支和依赖，避免一次写太多导致错误来源不清。",
    "把示例输出写成稳定文本，便于 CI、评审机器人或课程验收做字符串匹配。",
    "提交前使用 gofmt 或编辑器自动格式化，减少无意义的格式争论。",
  ],
  pitfalls: [
    { title: "文件能编译不代表能作为命令运行", symptom: "普通库包没有 main 函数，go run 时不能生成独立命令。", fix: "需要运行程序时使用 package main，并提供 func main；可复用逻辑再拆到普通包。" },
    { title: "忘记必要 import 或保留未使用 import", symptom: "代码看起来完整，但编译器直接拒绝通过。", fix: "让 import 精确反映当前文件实际使用的包，并用 gofmt/goimports 辅助维护。" },
    { title: "输出格式不稳定", symptom: "人工看起来正确，自动匹配却失败。", fix: "明确换行、空格和字段顺序；需要格式化时优先使用 fmt.Printf。" },
  ],
  exercise: {
    title: "打印入职欢迎信息",
    prompt: "运行一段最小 Go 程序，确认你能通过 sandbox 得到稳定输出。",
    starterCode: `package main

import "fmt"

func main() {
    name := "Gopher"
    fmt.Printf("welcome, %s\\n", name)
}`,
    expectedOutput: "welcome, Gopher",
    outputMatch: "contains",
    hints: ["程序必须声明 package main。", "fmt.Printf 不会自动换行，必要时写入 \\n。"],
  },
  checklist: [
    "能说出 package main 和普通业务包的差异。",
    "能解释 go run 与 go build 的不同反馈节奏。",
    "能指出 import、fmt.Println 和 gofmt 在最小程序中的作用。",
  ],
  reviewQuestions: [
    "为什么 Go 要求源文件先声明 package，再声明 import？",
    "stdout、stderr 和 exit code 分别适合表达什么信息？",
    "为什么统一格式化对团队协作比个人格式偏好更重要？",
  ],
  nextMissionSlugs: [],
  contentSource: {
    primary: "本章基于 /home/wayne/source/open/gopl-zh.github.com/ch1/ch1.md 与 ch1/ch1-01.md 精简重组。",
    references: ["The Go Programming Language 第 1 章", "Go 官方文档 go.dev/doc", "Go by Example"],
    license: "gopl-zh.github.com 使用 BSD 风格许可；后续批量迁移应保留项目级参考与许可说明。",
  },
},
{
  slug: "ch2-program-structure",
  order: 2,
  title: "程序结构",
  duration: "约 80 分钟",
  difficulty: "入门",
  summary: "基于《Go 程序设计语言》第 2 章重组：理解命名、声明、变量、短变量声明、指针、赋值、类型、包初始化和作用域，写出可读、可维护的小型 Go 程序。",
  goals: [
    "能区分包级声明、文件级 import 和函数内局部声明的作用域。",
    "能解释 var、const、type、func 各自定义了什么程序实体。",
    "能判断什么时候使用短变量声明，什么时候必须使用普通赋值。",
    "能识别变量遮蔽、指针别名和 init 过度使用带来的维护风险。",
  ],
  lessonCount: 8,
  lessons: [],
  modernNotes: [
    { title: "module 时代仍然依赖清晰的包边界", body: "Go module 改变了依赖获取和版本管理方式，但包、文件、导出名字和 import 的基本规则没有变。写服务时，包边界仍是控制耦合和复用的第一层结构。" },
    { title: "显式初始化比隐式副作用更利于测试", body: "教材讲解了包级变量和 init 的初始化顺序。现代后端项目仍应尽量把配置读取、连接创建和依赖组装放在 main 或构造函数中，让测试可以显式传入替身对象。" },
  ],
  engineeringPractices: [
    "名字的作用域越大，语义就应该越明确；局部循环变量可以短，包级 API 应清楚表达业务含义。",
    "需要更新已有变量时使用 =，不要让 := 在内层作用域创建意外的新变量。",
    "包级变量只存放真正共享且初始化清楚的对象，避免导入包就触发网络连接、文件读取或生产副作用。",
  ],
  pitfalls: [
    { title: "误用 := 导致外层状态不变", symptom: "日志显示 if 内部值正确，函数返回值却仍是旧值。", fix: "需要更新已有变量时使用 =，并把变量声明放在共同作用域。" },
    { title: "把可变状态放到包级", symptom: "单测互相污染，或者并发请求读写同一份数据。", fix: "将状态封装进结构体实例，通过构造函数传递；并发访问时加锁或使用 channel。" },
    { title: "init 函数做太多事", symptom: "导入包就发起网络连接或读取生产配置，测试难以隔离。", fix: "把有副作用的初始化改为显式函数调用，并返回错误。" },
  ],
  exercise: {
    title: "修正变量作用域",
    prompt: "观察短变量声明如何影响外层变量，练习写出稳定的状态更新。",
    starterCode: `package main

import "fmt"

func main() {
    level := 1
    passed := true
    if passed {
        level = 2
    }
    fmt.Printf("level=%d\\n", level)
}`,
    expectedOutput: "level=2",
    outputMatch: "trimmed-exact",
    hints: ["如果写成 level := 2，会创建一个只在 if 内有效的新变量。", "观察输出是否反映外层变量的最终值。"],
  },
  checklist: [
    "能解释 := 和 = 的差异。",
    "能指出变量的有效作用域。",
    "能避免把临时状态提升到包级。",
  ],
  reviewQuestions: [
    "短变量声明至少需要满足什么条件？",
    "为什么 err 遮蔽会导致排障困难？",
    "哪些初始化逻辑不适合放进 init 函数？",
  ],
  nextMissionSlugs: [],
  contentSource: {
    primary: "本章基于 /home/wayne/source/open/gopl-zh.github.com/ch2 全章精简重组。",
    references: ["The Go Programming Language 第 2 章", "Effective Go", "Go Code Review Comments"],
    license: "gopl-zh.github.com 使用 BSD 风格许可；后续批量迁移应保留项目级参考与许可说明。",
  },
},
{
  slug: "ch3-basic-data-types",
  order: 3,
  title: "基础数据类型",
  duration: "约 85 分钟",
  difficulty: "基础",
  summary: "基于《Go 程序设计语言》第 3 章重组：掌握整数、浮点数、布尔值、字符串、Unicode/UTF-8、常量和 iota，理解类型选择如何影响接口、存储和国际化输入。",
  goals: [
    "能根据业务含义选择 int、int64、float64、bool、string、byte 或 rune。",
    "能解释整数溢出、浮点精度、NaN 和显式类型转换的基本风险。",
    "能区分字符串的字节长度、rune 数量和用户可见字符。",
    "能使用常量和 iota 表达稳定枚举、单位和位标志。",
  ],
  lessonCount: 8,
  lessons: [],
  modernNotes: [
    { title: "UTF-8 仍是 Web 服务默认文本假设", body: "现代 API、日志和数据库通常以 UTF-8 为默认编码，但输入可能来自不同终端、浏览器和第三方系统。边界层需要校验、规范化并明确按字节还是按字符处理。" },
    { title: "泛型不替代基础类型判断", body: "Go 1.18+ 的泛型能抽象容器和算法，但金额、ID、状态码、耗时和容量单位仍需要清晰命名、类型选择和边界校验。" },
  ],
  engineeringPractices: [
    "为金额、计数、耗时等字段写清单位，例如 milliseconds、cents 或 bytes。",
    "处理用户可见字符串时优先按 rune 或更高层文本规则处理，不按字节随意截断。",
    "利用零值简化默认配置，但对必须填写的字段做显式校验。",
  ],
  pitfalls: [
    { title: "把 len 当字符数", symptom: "中文或 emoji 昵称长度判断异常。", fix: "明确需要字节数还是 rune 数；用户可见长度通常不能直接用 len(string)。" },
    { title: "用 float 表示金额", symptom: "订单金额出现 0.30000000000000004 一类误差。", fix: "使用整数最小货币单位或专门 decimal 类型。" },
    { title: "忽略零值语义", symptom: "0、空串和未填写状态混在一起。", fix: "必要时使用指针、额外布尔字段或显式枚举区分未知和零值。" },
  ],
  exercise: {
    title: "统计中文标题长度",
    prompt: "比较字符串的字节长度和 rune 数量，建立处理多语言文本的直觉。",
    starterCode: `package main

import "fmt"

func main() {
    title := "Go语言"
    fmt.Printf("bytes=%d runes=%d\\n", len(title), len([]rune(title)))
}`,
    expectedOutput: "bytes=8 runes=4",
    outputMatch: "trimmed-exact",
    hints: ["len(string) 返回底层字节数。", "[]rune 会按 Unicode 码点展开字符串。"],
  },
  checklist: [
    "能解释字符串长度的两种常见含义。",
    "能说出基础类型的零值。",
    "能为业务字段选择合适数值类型。",
  ],
  reviewQuestions: [
    "为什么字符串按字节截断可能产生乱码？",
    "什么时候 int64 比 int 更适合作为字段类型？",
    "零值设计对配置结构体有什么帮助和风险？",
  ],
  nextMissionSlugs: [],
  contentSource: {
    primary: "本章基于 /home/wayne/source/open/gopl-zh.github.com/ch3 全章精简重组。",
    references: ["The Go Programming Language 第 3 章", "Go by Example", "Go 官方 strings/unicode/utf8 文档"],
    license: "gopl-zh.github.com 使用 BSD 风格许可；后续批量迁移应保留项目级参考与许可说明。",
  },
},
{
  slug: "ch4-composite-types",
  order: 4,
  title: "复合数据类型",
  duration: "约 95 分钟",
  difficulty: "基础",
  summary: "基于《Go 程序设计语言》第 4 章重组：学习数组、slice、map、struct 和 JSON 建模，理解底层数组、引用语义、稳定输出和后端数据结构设计。",
  goals: [
    "能解释数组、slice、map 和 struct 的适用边界。",
    "能说明 slice header、底层数组、len、cap 和 append 的关系。",
    "能使用 map 做计数、去重和索引，并处理稳定输出。",
    "能设计清晰的 struct 与 JSON tag 表达请求、响应和领域对象。",
  ],
  lessonCount: 9,
  lessons: [],
  modernNotes: [
    { title: "slices 和 maps 包补充了通用工具", body: "现代 Go 标准库提供 slices、maps 等辅助包，排序、克隆、比较等操作更直接。但理解底层共享关系仍然是避免内存和并发 bug 的关键。" },
    { title: "结构体标签是跨系统边界", body: "JSON、ORM、验证器和日志库都可能读取标签。标签不是注释，修改它会影响外部协议，应像修改接口一样谨慎。" },
  ],
  engineeringPractices: [
    "长期保存从大数据中截取的小 slice 时，使用 append([]T(nil), part...) 或 copy 断开引用。",
    "需要稳定输出 map 内容时，对 key 排序后再遍历。",
    "为请求、响应和内部模型分别定义结构体，避免把数据库字段直接暴露给前端。",
  ],
  pitfalls: [
    { title: "nil map 写入 panic", symptom: "读取 map 正常，第一次赋值时崩溃。", fix: "写入前使用 make 初始化，或通过构造函数创建包含 map 的结构体。" },
    { title: "slice 共享底层数组", symptom: "修改一个切片后，另一个看似独立的切片内容也变了。", fix: "需要独立数据时显式 copy，并注意 append 可能复用容量。" },
    { title: "依赖 map 遍历顺序", symptom: "本地测试偶尔通过，CI 输出顺序不同。", fix: "将 key 收集后排序，再按排序结果访问 map。" },
  ],
  exercise: {
    id: "ch4-status-counts",
    kind: "run",
    difficulty: "warmup",
    concepts: ["slice", "map", "stable output"],
    estimatedMinutes: 8,
    title: "统计订单状态",
    prompt: "用 slice 保存事件流，用 map 聚合状态数量，输出一个稳定的统计结果。",
    starterCode: `package main

import "fmt"

func main() {
    statuses := []string{"paid", "pending", "paid", "failed", "paid", "pending"}
    counts := map[string]int{}
    for _, status := range statuses {
        counts[status]++
    }
    fmt.Printf("pending=%d paid=%d failed=%d\\n", counts["pending"], counts["paid"], counts["failed"])
}`,
    expectedOutput: "pending=2 paid=3 failed=1",
    outputMatch: "trimmed-exact",
    hints: ["map 的零值不能直接写入，字面量或 make 都可以初始化。", "为了输出稳定，不要直接 range map。"],
    solutionOutline: ["初始化 map 后遍历 statuses。", "使用 counts[status]++ 统计。", "按固定字段顺序输出，不直接 range map。"],
  },
  exercises: [
    {
      id: "ch4-status-counts",
      kind: "run",
      difficulty: "warmup",
      concepts: ["slice", "map", "stable output"],
      estimatedMinutes: 8,
      title: "统计订单状态",
      prompt: "用 slice 保存事件流，用 map 聚合状态数量，输出一个稳定的统计结果。",
      starterCode: `package main

import "fmt"

func main() {
    statuses := []string{"paid", "pending", "paid", "failed", "paid", "pending"}
    counts := map[string]int{}
    for _, status := range statuses {
        counts[status]++
    }
    fmt.Printf("pending=%d paid=%d failed=%d\\n", counts["pending"], counts["paid"], counts["failed"])
}`,
      expectedOutput: "pending=2 paid=3 failed=1",
      outputMatch: "trimmed-exact",
      hints: ["map 写入前要初始化。", "读取不存在的 key 会得到 int 零值。", "输出时按固定 key 顺序读取，避免 map 遍历随机顺序。"],
      solutionOutline: ["遍历事件 slice。", "用 map 计数。", "固定输出 pending、paid、failed。"],
    },
    {
      id: "ch4-stable-counts",
      kind: "edit",
      difficulty: "core",
      concepts: ["map", "sort", "strings.Builder", "deterministic output"],
      estimatedMinutes: 18,
      title: "实现稳定统计输出",
      prompt: "补全 StableCounts：统计任意状态出现次数，并按 key 字典序输出，避免依赖 map 的随机遍历顺序。",
      context: "真实后端里的日志、导出文件、签名字符串和测试快照都需要稳定顺序。",
      starterCode: `package main

import (
    "fmt"
    "sort"
    "strings"
)

func StableCounts(statuses []string) string {
    counts := map[string]int{}
    for _, status := range statuses {
        // TODO: 统计每个 status
    }

    keys := make([]string, 0, len(counts))
    // TODO: 收集并排序 keys

    parts := make([]string, 0, len(keys))
    for _, key := range keys {
        parts = append(parts, fmt.Sprintf("%s=%d", key, counts[key]))
    }
    return strings.Join(parts, " ")
}

func main() {
    statuses := []string{"paid", "pending", "paid", "failed", "paid", "pending"}
    fmt.Println(StableCounts(statuses))
}`,
      expectedOutput: "failed=1 paid=3 pending=2",
      outputMatch: "trimmed-exact",
      hints: ["先用 counts[status]++ 完成计数。", "用 for key := range counts 收集 key。", "sort.Strings(keys) 后再拼接输出。"],
      solutionOutline: ["统计 map。", "收集 key 到 slice。", "排序 key。", "按排序后的 key 拼接结果。"],
    },
    {
      id: "ch4-safe-response",
      kind: "debug",
      difficulty: "challenge",
      concepts: ["struct", "json", "DTO", "data boundary"],
      estimatedMinutes: 22,
      title: "修复响应结构体泄漏",
      prompt: "当前 JSON 响应会暴露 Email 和 PasswordHash。改成明确的响应 DTO，只输出 id 和 name。",
      context: "数据库模型、内部领域对象和 API 响应不应随意混用。",
      starterCode: `package main

import (
    "encoding/json"
    "fmt"
)

type User struct {
    ID           int64  \`json:"id"\`
    Name         string \`json:"name"\`
    Email        string \`json:"email"\`
    PasswordHash string \`json:"password_hash"\`
}

func PublicJSON(user User) string {
    data, _ := json.Marshal(user)
    return string(data)
}

func main() {
    user := User{ID: 7, Name: "Gopher", Email: "gopher@example.com", PasswordHash: "secret"}
    fmt.Println(PublicJSON(user))
}`,
      expectedOutput: `{"id":7,"name":"Gopher"}`,
      outputMatch: "trimmed-exact",
      hints: ["新增 UserResponse 结构体，只包含 ID 和 Name。", "在 PublicJSON 内把 User 转成 UserResponse。", "不要依赖 omitempty 隐藏敏感字段；边界上应该明确建模。"],
      solutionOutline: ["定义响应 DTO。", "从内部 User 映射到响应结构。", "只 marshal 响应结构。"],
    },
  ],
  checklist: [
    "能解释 slice 的 len 和 cap。",
    "能用 map 完成计数。",
    "能说明结构体标签的作用。",
  ],
  reviewQuestions: [
    "为什么小 slice 可能保留大数组？",
    "map 的遍历顺序为什么不能作为业务逻辑依赖？",
    "请求结构体和数据库结构体混用有什么风险？",
  ],
  nextMissionSlugs: ["slice-memory-leak"],
  contentSource: {
    primary: "本章基于 /home/wayne/source/open/gopl-zh.github.com/ch4/ch4.md 与 ch4-01 至 ch4-05 精简重组。",
    references: ["The Go Programming Language 第 4 章", "Go by Example", "encoding/json 官方文档", "Go Code Review Comments"],
    license: "gopl-zh.github.com 使用 BSD 风格许可；后续批量迁移应保留项目级参考与许可说明。",
  },
},
{
  slug: "ch5-functions",
  order: 5,
  title: "函数",
  duration: "约 90 分钟",
  difficulty: "基础",
  summary: "基于《Go 程序设计语言》第 5 章重组：理解函数签名、多返回值、错误返回、递归、函数值、闭包、defer 和 panic，把业务步骤拆成可测试的小单元。",
  goals: [
    "能用函数签名表达输入、输出和错误。",
    "能合理使用 error 返回、错误包装、重试、记录和忽略策略。",
    "能识别函数值、匿名函数和闭包捕获状态带来的好处与风险。",
    "能使用 defer 管理资源生命周期，并判断 panic 的适用边界。",
  ],
  lessonCount: 9,
  lessons: [],
  modernNotes: [
    { title: "错误包装是服务排障基础", body: "现代 Go 中 fmt.Errorf 使用 %w 包装错误后，上层可以用 errors.Is/As 判断原因，同时保留上下文。基础阶段要养成返回带上下文错误的习惯。" },
    { title: "range 变量捕获语义已有改进", body: "新版本 Go 对常见循环变量捕获问题做了语言层改进，但老代码、索引变量和手写复用变量仍需要审查。课程建议继续显式传参，保持意图清楚。" },
  ],
  engineeringPractices: [
    "把解析、校验、执行拆成函数，让每一步能单独写表驱动测试。",
    "资源获取成功后尽快写 defer，并确认循环中的 defer 不会拖延释放。",
    "返回错误时补充业务上下文，例如任务 ID、文件名或参数名。",
  ],
  pitfalls: [
    { title: "命名返回值被误改", symptom: "函数很长，多个 defer 或分支修改返回值，最终结果不直观。", fix: "短函数可使用命名返回值；复杂函数优先显式 return 局部变量。" },
    { title: "循环里 defer 过多", symptom: "处理大量文件时句柄迟迟不释放。", fix: "把每次循环体拆成小函数，或在本轮处理结束时显式关闭。" },
    { title: "闭包隐藏共享状态", symptom: "回调执行后外层变量被改变，测试顺序影响结果。", fix: "减少可变捕获；需要状态时封装结构体并提供方法。" },
  ],
  exercise: {
    title: "解析任务耗时",
    prompt: "把字符串中的分钟数解析为整数，并输出业务需要的耗时格式。",
    starterCode: `package main

import (
    "fmt"
    "strconv"
    "strings"
)

func parseMinutes(input string) (int, error) {
    parts := strings.Split(input, "m")
    return strconv.Atoi(parts[0])
}

func main() {
    minutes, err := parseMinutes("45m")
    if err != nil {
        fmt.Println(err)
        return
    }
    fmt.Printf("duration=%dmin\\n", minutes)
}`,
    expectedOutput: "duration=45min",
    outputMatch: "trimmed-exact",
    hints: ["函数可以返回结果和 error。", "main 中先处理 err，再使用结果。"],
  },
  checklist: [
    "能写出返回 error 的函数。",
    "能解释 defer 的执行时机。",
    "能识别闭包捕获状态。",
  ],
  reviewQuestions: [
    "函数签名如何帮助调用者理解失败路径？",
    "defer 为什么适合资源释放，但不适合无限制放在大循环里？",
    "闭包捕获变量时，什么时候应该改成显式参数？",
  ],
  nextMissionSlugs: ["defer-order"],
  contentSource: {
    primary: "本章基于 /home/wayne/source/open/gopl-zh.github.com/ch5 全章精简重组。",
    references: ["The Go Programming Language 第 5 章", "Effective Go", "Go blog: errors are values", "Go by Example"],
    license: "gopl-zh.github.com 使用 BSD 风格许可；后续批量迁移应保留项目级参考与许可说明。",
  },
},
{
  slug: "ch6-methods",
  order: 6,
  title: "方法",
  duration: "约 75 分钟",
  difficulty: "基础",
  summary: "基于《Go 程序设计语言》第 6 章重组：学习方法接收者、值语义、指针语义、嵌入、方法值、方法表达式和封装，用小结构体承载业务行为。",
  goals: [
    "能为命名类型定义方法并选择值接收者或指针接收者。",
    "能解释方法集、自动取址/解引用和接口实现之间的关系。",
    "能使用结构体嵌入组合能力，而不是误解为继承。",
    "能通过封装保护对象不变量，避免外部随意修改内部状态。",
  ],
  lessonCount: 8,
  lessons: [],
  modernNotes: [
    { title: "组合仍是 Go 的主线", body: "即使项目越来越大，Go 代码仍普遍使用结构体组合和小接口表达能力，而不是深继承树。方法应服务于清晰的领域行为。" },
    { title: "泛型与方法各有边界", body: "泛型适合抽象算法和容器；方法适合绑定具体领域行为。不要为了复用而让业务对象变成难懂的类型参数集合。" },
  ],
  engineeringPractices: [
    "需要修改接收者状态时使用指针接收者，并在方法名中体现业务动作。",
    "避免方法内部直接读取全局配置；把依赖放在结构体字段中由构造函数注入。",
    "为有状态结构体写最小行为测试，覆盖初始状态、变更和边界输入。",
  ],
  pitfalls: [
    { title: "值接收者修改无效", symptom: "方法内部字段变了，调用后原对象仍是旧值。", fix: "需要修改原对象时使用指针接收者。" },
    { title: "接收者命名过长或过短", symptom: "方法体里读不出当前对象含义，或和包名混淆。", fix: "使用一两个字母但有辨识度的接收者名，如 account 用 a。" },
    { title: "接口实现卡在指针和值", symptom: "T 不能赋给接口，但 *T 可以。", fix: "检查方法接收者，必要时统一使用指针实例。" },
  ],
  exercise: {
    title: "给积分账户增加方法",
    prompt: "为结构体定义指针接收者方法，更新账户积分并输出最终值。",
    starterCode: `package main

import "fmt"

type Account struct {
    Points int
}

func (a *Account) Add(points int) {
    a.Points += points
}

func main() {
    account := &Account{Points: 10}
    account.Add(15)
    fmt.Printf("points=%d\\n", account.Points)
}`,
    expectedOutput: "points=25",
    outputMatch: "trimmed-exact",
    hints: ["需要修改原结构体时，接收者应是 *Account。", "方法调用语法会自动处理部分取址，但语义仍要清楚。"],
  },
  checklist: [
    "能定义带接收者的方法。",
    "能说明值接收者和指针接收者差异。",
    "能解释方法集对接口的影响。",
  ],
  reviewQuestions: [
    "什么时候值接收者比指针接收者更合适？",
    "为什么同一类型的方法接收者风格最好保持一致？",
    "方法和普通函数在组织业务代码时各有什么优势？",
  ],
  nextMissionSlugs: [],
  contentSource: {
    primary: "本章基于 /home/wayne/source/open/gopl-zh.github.com/ch6 全章精简重组。",
    references: ["The Go Programming Language 第 6 章", "Effective Go", "Go Code Review Comments"],
    license: "gopl-zh.github.com 使用 BSD 风格许可；后续批量迁移应保留项目级参考与许可说明。",
  },
},
{
  slug: "ch7-interfaces",
  order: 7,
  title: "接口",
  duration: "约 95 分钟",
  difficulty: "进阶",
  summary: "基于《Go 程序设计语言》第 7 章重组：用小接口表达能力边界，理解隐式实现、接口值、nil 接口、类型断言和错误分类，让代码更容易测试和替换。",
  goals: [
    "能定义只包含必要方法的小接口。",
    "能解释 Go 的接口是隐式实现，以及方法集如何影响接口满足关系。",
    "能识别 nil 接口和带类型 nil 值的差异。",
    "能用接口替换外部依赖，并在必要时使用类型断言或类型分支。",
  ],
  lessonCount: 10,
  lessons: [],
  modernNotes: [
    { title: "泛型没有淘汰接口", body: "泛型适合表达同构算法，接口适合表达运行时行为能力。后端代码中二者会并存：接口隔离依赖，泛型减少容器样板。" },
    { title: "context.Context 是接口协作的典型边界", body: "现代服务几乎每条请求链都会传递 context，用于取消、超时和少量请求级元数据。它体现了小接口和约定式协作的价值。" },
  ],
  engineeringPractices: [
    "在消费方定义接口，只暴露当前函数实际需要的方法。",
    "为外部系统依赖定义小接口，测试时注入 fake 或 stub。",
    "返回 error 接口时确认没有把 nil 指针包装成非 nil 接口。",
  ],
  pitfalls: [
    { title: "接口过大", symptom: "测试 fake 必须实现一堆用不到的方法。", fix: "按调用场景拆成多个小接口。" },
    { title: "nil 判断失效", symptom: "看起来返回 nil，调用方 if err != nil 却成立。", fix: "避免返回带类型的 nil 指针；直接返回 nil 接口值。" },
    { title: "在实现方过早抽象", symptom: "只有一个实现却先定义复杂接口层。", fix: "等出现真实替换需求，或从测试隔离需求出发定义接口。" },
  ],
  exercise: {
    title: "用接口统一通知渠道",
    prompt: "定义一个小接口，让发送逻辑不关心具体通知实现。",
    starterCode: `package main

import "fmt"

type Notifier interface {
    Notify(message string) string
}

type EmailNotifier struct{}

func (EmailNotifier) Notify(message string) string {
    return "email:" + message
}

func send(n Notifier, message string) string {
    return n.Notify(message)
}

func main() {
    fmt.Println(send(EmailNotifier{}, "build passed"))
}`,
    expectedOutput: "email:build passed",
    outputMatch: "trimmed-exact",
    hints: ["接口只需要包含调用方真正使用的方法。", "EmailNotifier 不需要显式声明 implements。"],
  },
  checklist: [
    "能写出小接口。",
    "能说明隐式实现。",
    "能描述 nil 接口问题。",
  ],
  reviewQuestions: [
    "为什么接口应尽量由消费方定义？",
    "接口值在什么情况下等于 nil？",
    "泛型和接口分别适合解决什么问题？",
  ],
  nextMissionSlugs: [],
  contentSource: {
    primary: "本章基于 /home/wayne/source/open/gopl-zh.github.com/ch7 重点章节精简重组。",
    references: ["The Go Programming Language 第 7 章", "Effective Go", "Go Code Review Comments", "Go 官方 errors 文档"],
    license: "gopl-zh.github.com 使用 BSD 风格许可；后续批量迁移应保留项目级参考与许可说明。",
  },
},
{
  slug: "ch8-goroutines-channels",
  order: 8,
  title: "Goroutines 和 Channels",
  duration: "约 105 分钟",
  difficulty: "进阶",
  summary: "基于《Go 程序设计语言》第 8 章重组：建立 goroutine 生命周期、channel 通信、pipeline、select、取消、并发限制和 goroutine 泄露的工程直觉。",
  goals: [
    "能启动 goroutine 并用 channel 收集结果。",
    "能解释 unbuffered 和 buffered channel 的差异。",
    "能使用 select 处理超时、取消和多路事件。",
    "能识别 goroutine 泄露、过度并发和 channel 关闭所有权问题。",
  ],
  lessonCount: 10,
  lessons: [],
  modernNotes: [
    { title: "context 是现代 Go 服务的默认参数", body: "HTTP、数据库、RPC 和队列客户端普遍支持 context。只要函数可能阻塞、访问外部资源或跨 goroutine 工作，就应考虑接收 context。" },
    { title: "结构化并发越来越重要", body: "Go 生态常用 errgroup 管理一组 goroutine 的错误传播和取消。即使基础练习只用 channel，也要形成成组启动、成组等待的习惯。" },
  ],
  engineeringPractices: [
    "启动 goroutine 的函数负责提供退出条件，常见方式是 context、done channel 或关闭输入 channel。",
    "对外部服务并发调用设置上限，避免把本服务压力转嫁成下游雪崩。",
    "在测试中使用超时，防止并发 bug 让测试进程永久挂起。",
  ],
  pitfalls: [
    { title: "main 提前退出", symptom: "goroutine 里的输出偶尔看不到。", fix: "使用 channel、WaitGroup 或 errgroup 等待工作完成。" },
    { title: "无人接收导致阻塞", symptom: "发送方卡住，程序没有报错但不再前进。", fix: "确认接收方存在；必要时使用缓冲、select 超时或取消信号。" },
    { title: "关闭 channel 的所有权混乱", symptom: "panic: send on closed channel。", fix: "通常由发送方关闭 channel，并确保只有一个关闭所有者。" },
  ],
  exercise: {
    title: "并发收集两个任务结果",
    prompt: "启动两个 goroutine，把结果发送到 channel，再按稳定顺序输出。",
    starterCode: `package main

import (
    "fmt"
    "sort"
    "strings"
)

func main() {
    results := make(chan string, 2)
    go func() { results <- "cache" }()
    go func() { results <- "db" }()

    collected := []string{<-results, <-results}
    sort.Strings(collected)
    fmt.Printf("results=%s\\n", strings.Join(collected, ","))
}`,
    expectedOutput: "results=cache,db",
    outputMatch: "trimmed-exact",
    hints: ["带缓冲 channel 可以容纳两个结果。", "并发返回顺序不稳定，输出前排序。"],
  },
  checklist: [
    "能用 channel 收集 goroutine 结果。",
    "能说明 channel 关闭所有权。",
    "能描述 goroutine 泄露的触发条件。",
  ],
  reviewQuestions: [
    "为什么启动 goroutine 时必须同时考虑退出条件？",
    "buffered channel 能解决哪些问题，又会掩盖哪些问题？",
    "select 中监听 ctx.Done 的意义是什么？",
  ],
  nextMissionSlugs: ["slice-memory-leak"],
  contentSource: {
    primary: "本章基于 /home/wayne/source/open/gopl-zh.github.com/ch8 全章精简重组。",
    references: ["The Go Programming Language 第 8 章", "Go blog: Pipelines and cancellation", "context 官方文档", "Go by Example"],
    license: "gopl-zh.github.com 使用 BSD 风格许可；后续批量迁移应保留项目级参考与许可说明。",
  },
},
{
  slug: "ch9-shared-variable-concurrency",
  order: 9,
  title: "基于共享变量的并发",
  duration: "约 95 分钟",
  difficulty: "进阶",
  summary: "基于《Go 程序设计语言》第 9 章重组：理解数据竞争、互斥锁、读写锁、内存同步、sync.Once、race detector 和并发安全缓存，避免共享状态把服务拖垮。",
  goals: [
    "能解释 data race 和逻辑竞争的区别。",
    "能使用 sync.Mutex 保护共享变量和不变量。",
    "能判断何时使用 channel、锁、RWMutex、sync.Once 或 atomic。",
    "能使用 race detector 发现运行时数据竞争，并为并发代码设计可复现测试。",
  ],
  lessonCount: 10,
  lessons: [],
  modernNotes: [
    { title: "race detector 是必备工具", body: "go test -race 能发现许多运行时数据竞争。它不能证明完全无并发 bug，但应成为并发代码合并前的常规检查。" },
    { title: "sync/atomic 类型化 API 更安全", body: "现代 Go 提供 atomic.Int64、atomic.Bool 等类型，减少手动传指针和类型转换错误。基础课程仍先用 Mutex 建立正确性模型。" },
  ],
  engineeringPractices: [
    "在结构体注释或字段旁说明共享状态由哪把锁保护。",
    "先写串行正确版本，再引入并发和同步，方便对比结果。",
    "并发测试加入超时、重复运行和 race detector，提升复现概率。",
  ],
  pitfalls: [
    { title: "map 并发写崩溃", symptom: "fatal error: concurrent map writes。", fix: "用 Mutex/RWMutex 保护 map，或使用 sync.Map 处理特定读多写少场景。" },
    { title: "复制含锁结构体", symptom: "锁保护失效，数据表现异常。", fix: "含 Mutex 的结构体不要复制；通常通过指针传递。" },
    { title: "defer 解锁放错位置", symptom: "提前返回忘记解锁或锁持有时间过长。", fix: "小函数中 Lock 后立即 defer Unlock；大函数拆分临界区。" },
  ],
  exercise: {
    title: "安全累加请求数",
    prompt: "用 Mutex 保护共享计数器，确保 100 个 goroutine 的更新都被保留。",
    starterCode: `package main

import (
    "fmt"
    "sync"
)

func main() {
    var wg sync.WaitGroup
    var mu sync.Mutex
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
    fmt.Printf("total=%d\\n", total)
}`,
    expectedOutput: "total=100",
    outputMatch: "trimmed-exact",
    hints: ["WaitGroup 等待所有 goroutine 完成。", "total++ 不是原子操作，需要同步保护。"],
  },
  checklist: [
    "能解释数据竞争。",
    "能用 Mutex 保护共享变量。",
    "能说明 race detector 的作用。",
  ],
  reviewQuestions: [
    "数据竞争和业务层面的竞态条件有什么不同？",
    "什么时候锁比 channel 更直接？",
    "为什么含 Mutex 的结构体通常不应复制？",
  ],
  nextMissionSlugs: ["map-concurrent-write"],
  contentSource: {
    primary: "本章基于 /home/wayne/source/open/gopl-zh.github.com/ch9 全章精简重组。",
    references: ["The Go Programming Language 第 9 章", "Go Memory Model", "Data Race Detector 官方文档", "sync 与 sync/atomic 官方文档"],
    license: "gopl-zh.github.com 使用 BSD 风格许可；后续批量迁移应保留项目级参考与许可说明。",
  },
},
{
  slug: "ch10-packages-tools",
  order: 10,
  title: "包和工具",
  duration: "约 100 分钟",
  difficulty: "基础",
  summary: "基于《Go 程序设计语言》第 10 章重组：理解 package、import path、匿名导入、internal、Go 工具箱、module/workspace 与构建查询工具，建立可维护的包边界和团队反馈回路。",
  goals: [
    "能解释 package、module、import path 和包名之间的关系。",
    "能根据职责设计包边界，避免循环依赖和 util 垃圾桶包。",
    "能使用 go fmt、go test、go vet、go build、go doc、go list 等工具完成日常反馈。",
    "能理解匿名导入、internal 包、构建标签和交叉编译的适用边界。",
  ],
  lessonCount: 10,
  lessons: [],
  modernNotes: [
    { title: "module 取代 GOPATH 作为主要协作边界", body: "教材重点讲 GOPATH 工作区；现代 Go 项目通常以 go.mod 描述模块路径、Go 版本和依赖版本。GOPATH 仍存在，但日常协作应围绕 module、go.sum 和 CI 工具链展开。" },
    { title: "go work 适合本地多模块联调", body: "当服务和内部库分属不同 module 时，go work 可以把它们临时放入同一工作区。发布和 CI 仍应验证每个 module 独立可构建，避免 workspace 掩盖版本问题。" },
    { title: "工具链版本也是项目契约", body: "toolchain 指令、CI 镜像、README 和开发脚本应共同说明 Go 版本。构建失败时，先确认 go env、GOOS、GOARCH、CGO_ENABLED 和 module 状态。" },
  ],
  engineeringPractices: [
    "包名短小但要表达职责，调用点应像 json.Marshal、http.Client 一样自然。",
    "CI 至少运行 go test ./...、go vet ./... 和 go build ./...，不要只依赖 IDE 的绿色提示。",
    "匿名导入必须写注释说明注册目的，例如数据库驱动或图片解码器。",
    "内部实现优先放入 internal 子树，避免不稳定 API 被外部调用方依赖。",
  ],
  pitfalls: [
    { title: "循环依赖", symptom: "两个包互相 import，go build 直接失败。", fix: "重新划分依赖方向，把共同接口或数据类型上移到底层包，避免业务上层反向依赖。" },
    { title: "util/common 垃圾桶包", symptom: "所有功能都能放进去，最后任何包都依赖它。", fix: "按领域或能力命名包，例如 retry、clock、billing、authz，让职责可被评审。" },
    { title: "匿名导入隐藏副作用", symptom: "删除一行看似未使用的 import 后，图片解码或数据库驱动失效。", fix: "仅在注册型场景使用 `_` import，并用注释说明副作用。" },
    { title: "只在本机能构建", symptom: "开发机通过，CI 或容器里 go env 不同导致失败。", fix: "固定工具链版本，记录 GOOS/GOARCH/CGO_ENABLED，并用脚本复现 CI 命令。" },
  ],
  exercise: {
    title: "查看工具链环境",
    prompt: "读取 runtime 提供的目标系统信息，理解构建环境会影响程序行为。",
    starterCode: `package main

import (
    "fmt"
    "runtime"
)

func main() {
    fmt.Printf("goos=%s goarch=%s\\n", runtime.GOOS, runtime.GOARCH)
}`,
    expectedOutput: "goos=",
    outputMatch: "contains",
    hints: ["不同 sandbox 可能运行在不同 GOOS/GOARCH 上。", "contains 匹配允许保留平台差异。", "真实项目还可以用 go env 查看更完整工具链环境。"],
  },
  checklist: [
    "能解释 package 和 module 的差异。",
    "能说出 import path 为什么是依赖身份。",
    "能使用 go list 或 go doc 查询包信息。",
    "能识别循环依赖和匿名导入风险。",
  ],
  reviewQuestions: [
    "为什么 Go 禁止循环 import？",
    "包名和成员名应该如何配合，避免重复和歧义？",
    "什么时候应该使用 internal 包？",
    "go list 相比手写 find 脚本有什么优势？",
  ],
  nextMissionSlugs: [],
  contentSource: {
    primary: "本章基于 /home/wayne/source/open/gopl-zh.github.com/ch10 全章精简重组。",
    references: ["The Go Programming Language 第 10 章", "Go Modules Reference", "go command 官方文档", "Effective Go", "Go Code Review Comments"],
    license: "gopl-zh.github.com 使用 BSD 风格许可；后续批量迁移应保留项目级参考与许可说明。",
  },
},
{
  slug: "ch11-testing",
  order: 11,
  title: "测试",
  duration: "约 110 分钟",
  difficulty: "进阶",
  summary: "基于《Go 程序设计语言》第 11 章重组：掌握 go test、测试函数、表驱动测试、随机/fuzz 测试、测试替身、外部测试包、覆盖率、基准测试、pprof 和示例函数。",
  goals: [
    "能编写清晰的 Test 函数，并输出可诊断的失败信息。",
    "能使用表驱动测试、子测试和测试替身覆盖正常路径与错误路径。",
    "能区分黑盒测试、白盒测试、外部测试包和集成测试。",
    "能使用 coverage、benchmark、benchmem 和 pprof 为质量与性能提供数据反馈。",
  ],
  lessonCount: 11,
  lessons: [],
  modernNotes: [
    { title: "fuzzing 已进入标准工具链", body: "Go 内置 fuzzing 能自动探索输入空间，适合解析器、编码器、协议处理和安全敏感函数。它不是表驱动测试的替代，而是补充手写边界用例的工具。" },
    { title: "t.Cleanup 改善测试资源恢复", body: "替换全局变量、创建临时文件、启动测试服务后，应使用 t.Cleanup 或 defer 恢复状态，避免测试间互相污染。" },
    { title: "性能问题先 profile 再优化", body: "Benchmark 适合比较固定操作，pprof 适合寻找真实热点。没有 profile 数据就重写代码，往往会增加复杂度却没有收益。" },
  ],
  engineeringPractices: [
    "每个 bug 修复尽量先补失败用例，再改实现，确保回归不会重复出现。",
    "测试失败信息包含函数、输入、got 和 want，让维护者不看源码也能定位问题。",
    "对外部依赖使用 fake、stub 或接口注入，单元测试不直接请求生产服务。",
    "覆盖率用于发现盲区，不把百分比当作唯一目标；关键错误路径比数字更重要。",
  ],
  pitfalls: [
    { title: "断言信息无上下文", symptom: "测试失败只显示 not equal 或 failed。", fix: "输出调用、关键输入、实际值和期望值，例如 f(x)=got, want y。" },
    { title: "共享状态污染", symptom: "单独运行通过，整包运行失败或顺序相关。", fix: "避免包级可变状态；必须替换时使用 t.Cleanup 恢复。" },
    { title: "脆弱测试", symptom: "内部重构后大量测试失败，但外部行为没变。", fix: "优先断言稳定的可观察行为，少检查私有中间步骤和完整大字符串。" },
    { title: "未隔离真实外部依赖", symptom: "测试偶发超时、扣费或污染真实数据。", fix: "用接口注入 fake；把真实服务交互放到受控集成测试。" },
  ],
  exercise: {
    id: "ch11-table-driven-summary",
    kind: "run",
    difficulty: "warmup",
    concepts: ["table-driven tests", "test cases", "failure messages"],
    estimatedMinutes: 8,
    title: "模拟表驱动测试统计",
    prompt: "用表格描述多个加法用例，模拟 go test 中表驱动测试的核心思想。",
    starterCode: `package main

import "fmt"

func add(a, b int) int {
    return a + b
}

func main() {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 1, 2, 3},
        {"zero", -1, 1, 0},
        {"negative", -2, -3, -5},
    }

    passed := 0
    for _, tt := range tests {
        if add(tt.a, tt.b) == tt.want {
            passed++
        }
    }
    fmt.Printf("passed=%d/%d\\n", passed, len(tests))
}`,
    expectedOutput: "passed=3/3",
    outputMatch: "trimmed-exact",
    hints: ["真实测试会写在 *_test.go 中，并使用 t.Errorf 报告失败。", "表驱动的关键是新增边界条件足够低成本。", "name 字段能让失败输出更容易定位。"],
    solutionOutline: ["把输入和期望写进结构体切片。", "循环执行同一套检查逻辑。", "统计或报告每个 case 的结果。"],
  },
  exercises: [
    {
      id: "ch11-table-driven-summary",
      kind: "run",
      difficulty: "warmup",
      concepts: ["table-driven tests", "test cases", "failure messages"],
      estimatedMinutes: 8,
      title: "模拟表驱动测试统计",
      prompt: "用表格描述多个加法用例，模拟 go test 中表驱动测试的核心思想。",
      starterCode: `package main

import "fmt"

func add(a, b int) int {
    return a + b
}

func main() {
    tests := []struct {
        name string
        a, b int
        want int
    }{
        {"positive", 1, 2, 3},
        {"zero", -1, 1, 0},
        {"negative", -2, -3, -5},
    }

    passed := 0
    for _, tt := range tests {
        if add(tt.a, tt.b) == tt.want {
            passed++
        }
    }
    fmt.Printf("passed=%d/%d\\n", passed, len(tests))
}`,
      expectedOutput: "passed=3/3",
      outputMatch: "trimmed-exact",
      hints: ["真实测试会写在 *_test.go 中。", "表驱动的关键是新增用例只要新增一行数据。", "name 字段能让失败输出更容易定位。"],
      solutionOutline: ["定义测试表。", "循环执行函数。", "比较 got 和 want。"],
    },
    {
      id: "ch11-normalize-name",
      kind: "debug",
      difficulty: "core",
      concepts: ["regression test", "strings", "edge cases"],
      estimatedMinutes: 18,
      title: "修复名称规范化边界",
      prompt: "当前 NormalizeName 没有处理前后空格和多余空格。修复函数，让输出稳定通过三个场景。",
      context: "先把用户报告转成可复现输入，再修实现，是回归测试的基本节奏。",
      starterCode: `package main

import (
    "fmt"
    "strings"
)

func NormalizeName(input string) string {
    return strings.ToLower(input)
}

func main() {
    tests := []struct {
        input string
        want  string
    }{
        {"Gopher", "gopher"},
        {"  Gopher  ", "gopher"},
        {"Go   Developer", "go developer"},
    }

    passed := 0
    for _, tt := range tests {
        if got := NormalizeName(tt.input); got == tt.want {
            passed++
        } else {
            fmt.Printf("NormalizeName(%q) = %q, want %q\\n", tt.input, got, tt.want)
        }
    }
    fmt.Printf("passed=%d/%d\\n", passed, len(tests))
}`,
      expectedOutput: "passed=3/3",
      outputMatch: "contains",
      hints: ["strings.TrimSpace 可以去掉前后空白。", "strings.Fields 可以把连续空白拆成字段。", "strings.Join(fields, \" \") 可以恢复为单空格分隔。"],
      solutionOutline: ["TrimSpace 或 Fields 处理空白。", "ToLower 统一大小写。", "用 Join 保证中间只有一个空格。"],
    },
    {
      id: "ch11-benchmark-comparison",
      kind: "project",
      difficulty: "challenge",
      concepts: ["benchmark", "allocation", "strings.Builder"],
      estimatedMinutes: 25,
      title: "比较两种字符串拼接策略",
      prompt: "补全 builderJoin，让它和 slowJoin 输出一致。这个练习模拟 benchmark 前先保证行为一致。",
      context: "真正优化前应先写基准测试；sandbox 里先用相同输入输出建立正确性基线。",
      starterCode: `package main

import (
    "fmt"
    "strings"
)

func slowJoin(parts []string) string {
    result := ""
    for i, part := range parts {
        if i > 0 {
            result += ","
        }
        result += part
    }
    return result
}

func builderJoin(parts []string) string {
    var b strings.Builder
    // TODO: 使用 Builder 拼出和 slowJoin 相同的结果
    return b.String()
}

func main() {
    parts := []string{"cache", "db", "queue"}
    fmt.Printf("same=%v output=%s\\n", slowJoin(parts) == builderJoin(parts), builderJoin(parts))
}`,
      expectedOutput: "same=true output=cache,db,queue",
      outputMatch: "trimmed-exact",
      hints: ["循环 parts，索引大于 0 时先写逗号。", "Builder 使用 WriteString。", "性能优化前先保证两个实现行为一致。"],
      solutionOutline: ["用 strings.Builder 累积字符串。", "处理分隔符位置。", "比较 slowJoin 与 builderJoin 输出。"],
    },
  ],
  checklist: [
    "能写出 TestXxx 的基本结构。",
    "能用表驱动测试组织多个输入输出。",
    "能说明 fake 和真实外部依赖的区别。",
    "能运行 go test -cover 和 go test -bench。",
  ],
  reviewQuestions: [
    "为什么好的测试失败信息要包含输入、got 和 want？",
    "黑盒测试和白盒测试分别适合发现什么问题？",
    "覆盖率为什么不能证明程序没有 bug？",
    "什么时候应该写 benchmark，什么时候需要 pprof？",
  ],
  nextMissionSlugs: [],
  contentSource: {
    primary: "本章基于 /home/wayne/source/open/gopl-zh.github.com/ch11 全章精简重组。",
    references: ["The Go Programming Language 第 11 章", "testing 包官方文档", "Go fuzzing 官方教程", "Go blog: Profiling Go Programs"],
    license: "gopl-zh.github.com 使用 BSD 风格许可；后续批量迁移应保留项目级参考与许可说明。",
  },
},
{
  slug: "ch12-reflection",
  order: 12,
  title: "反射",
  duration: "约 100 分钟",
  difficulty: "高级",
  summary: "基于《Go 程序设计语言》第 12 章重组：理解 reflect.Type、reflect.Value、Kind、递归值遍历、反射设置值、结构体标签、方法集以及反射在 JSON/ORM/框架中的边界和代价。",
  goals: [
    "能解释反射为何存在，以及它解决的是未知类型集合问题。",
    "能区分 reflect.Type、reflect.Value 和 Kind，并知道哪些操作可能 panic。",
    "能读取结构体字段标签，并理解 JSON、参数绑定和 ORM 的基本实现思路。",
    "能判断何时使用接口、泛型或代码生成替代反射。",
  ],
  lessonCount: 10,
  lessons: [],
  modernNotes: [
    { title: "泛型减少但没有消灭反射", body: "Go 泛型适合容器、集合和算法，能替代许多 interface{} 加反射的写法。但 JSON、ORM、配置绑定这类运行时结构映射仍然需要读取类型和标签。" },
    { title: "代码生成是高性能框架常用选择", body: "对性能敏感的序列化、RPC、数据库访问常用生成代码减少反射成本。生成代码牺牲一些灵活性，换来静态类型和更低运行时开销。" },
    { title: "反射错误应转成可诊断 error", body: "基础设施库不应让 CanSet、Kind 不匹配或 nil 引发的 panic 泄漏给业务层。检查输入并返回清晰错误，才是生产库的基本要求。" },
  ],
  engineeringPractices: [
    "把反射集中在基础设施包中，对外提供普通类型安全 API。",
    "设置值前检查传入是否为非 nil 指针，目标 Value 是否 CanSet，类型是否可赋值。",
    "结构体标签是协议的一部分，修改 tag 要有测试覆盖。",
    "热点路径上的反射要用 benchmark 量化，必要时缓存元数据或使用代码生成。",
  ],
  pitfalls: [
    { title: "Kind 检查缺失", symptom: "反射代码在少数输入上 panic。", fix: "调用 Int、Field、Index、Set 等方法前先检查 Kind、IsNil、CanSet。" },
    { title: "业务逻辑依赖字符串字段名", symptom: "字段重命名后编译通过，运行时才失败。", fix: "业务层优先使用强类型访问；反射仅放在框架边界。" },
    { title: "把 reflect.Value 暴露给调用方", symptom: "API 难懂且容易被误用。", fix: "在包内部消化反射，导出普通函数、结构体和 error。" },
    { title: "忽略性能成本", symptom: "请求热点路径 CPU 和分配异常。", fix: "用 benchmark/pprof 验证，考虑缓存字段信息、泛型或代码生成。" },
  ],
  exercise: {
    title: "读取 JSON 标签",
    prompt: "使用 reflect 读取结构体字段上的 json 标签，理解框架如何识别序列化名称。",
    starterCode: `package main

import (
    "fmt"
    "reflect"
)

type User struct {
    Name string \`json:"name"\`
}

func main() {
    field, ok := reflect.TypeOf(User{}).FieldByName("Name")
    if !ok {
        panic("missing field")
    }
    fmt.Printf("json=%s\\n", field.Tag.Get("json"))
}`,
    expectedOutput: "json=name",
    outputMatch: "trimmed-exact",
    hints: ["结构体标签是反射可读取的元数据。", "FieldByName 返回字段和是否存在的布尔值。", "真实 JSON tag 还可能包含 omitempty 等选项。"],
  },
  checklist: [
    "能用 TypeOf 和 ValueOf 观察动态类型和值。",
    "能读取结构体标签。",
    "能解释 CanSet 和传指针的重要性。",
    "能说出反射的替代方案和成本。",
  ],
  reviewQuestions: [
    "reflect.Type 和 reflect.Kind 的区别是什么？",
    "为什么 json.Unmarshal 需要传入指针？",
    "反射为什么会降低自动化重构的可靠性？",
    "泛型、接口、代码生成分别能替代哪些反射场景？",
  ],
  nextMissionSlugs: [],
  contentSource: {
    primary: "本章基于 /home/wayne/source/open/gopl-zh.github.com/ch12 全章精简重组。",
    references: ["The Go Programming Language 第 12 章", "reflect 包官方文档", "encoding/json 官方文档", "Go generics 官方教程"],
    license: "gopl-zh.github.com 使用 BSD 风格许可；后续批量迁移应保留项目级参考与许可说明。",
  },
},
{
  slug: "ch13-low-level-programming",
  order: 13,
  title: "底层编程",
  duration: "约 90 分钟",
  difficulty: "高级",
  summary: "基于《Go 程序设计语言》第 13 章重组：认识 unsafe、Sizeof/Alignof/Offsetof、unsafe.Pointer、uintptr、深度相等、cgo、系统边界和底层优化的风险隔离原则。",
  goals: [
    "能解释 Go 为什么默认隐藏底层内存和运行时细节。",
    "能使用 unsafe.Sizeof、Alignof、Offsetof 观察结构体大小和字段对齐。",
    "能说明 unsafe.Pointer 与 uintptr 转换的危险，避免长期保存地址整数。",
    "能判断 cgo、系统调用和 unsafe 是否真的必要，并把风险隔离在小范围。",
  ],
  lessonCount: 9,
  lessons: [],
  modernNotes: [
    { title: "unsafe 规则随着运行时演进更需谨慎", body: "即使当前实现没有通用移动 GC，栈增长、逃逸分析和未来运行时变化都可能让错误 unsafe 代码暴露。不要依赖未承诺的实现细节。" },
    { title: "标准库和泛型减少底层技巧需求", body: "现代 Go 提供更多标准库工具和泛型能力，许多过去需要 unsafe 或反射技巧的场景可以用更安全方式表达。" },
    { title: "cgo 是构建和部署边界", body: "引入 cgo 后，C 编译器、系统头文件、动态库、交叉编译和指针规则都会进入项目维护范围。它不只是代码调用方式的变化。" },
  ],
  engineeringPractices: [
    "底层优化前先写 benchmark，并用 pprof 确认真正热点。",
    "unsafe 或 cgo 代码集中封装在小包内，对外提供普通 Go API。",
    "在注释中写清平台假设、内存布局假设、对象生命周期和不可变约束。",
    "跨平台敏感代码使用 build tag 隔离，并在目标 GOOS/GOARCH 上验证。",
  ],
  pitfalls: [
    { title: "把 unsafe 当性能捷径", symptom: "代码更复杂，但 profile 中瓶颈并不在这里。", fix: "先测量，再优化；优先改算法、分配、批量 I/O 和缓存策略。" },
    { title: "uintptr 保存地址", symptom: "偶发崩溃或数据损坏，难以复现。", fix: "不要把指针地址长期存成整数；必要转换尽量保持在单个表达式内。" },
    { title: "把结构体内存当协议", symptom: "换平台、改字段或升级 Go 后二进制解析损坏。", fix: "使用显式编码格式，写清字节序、字段宽度和版本。" },
    { title: "低估 cgo 部署成本", symptom: "本机能构建，CI 或容器缺头文件/动态库失败。", fix: "明确 C 依赖安装方式，提供构建镜像或纯 Go fallback。" },
  ],
  exercise: {
    title: "观察结构体大小",
    prompt: "使用 unsafe.Sizeof 观察结构体布局带来的大小，建立字段对齐直觉。",
    starterCode: `package main

import (
    "fmt"
    "unsafe"
)

type Metric struct {
    Count   int64
    Healthy bool
}

func main() {
    fmt.Printf("size=%d\\n", unsafe.Sizeof(Metric{}))
}`,
    expectedOutput: "size=16",
    outputMatch: "trimmed-exact",
    hints: ["int64 通常需要 8 字节对齐。", "bool 只有 1 字节，但结构体整体会包含填充。", "不同架构上大小可能不同，课程 sandbox 通常是 64 位环境。"],
  },
  checklist: [
    "能说明 unsafe 的主要风险。",
    "能观察结构体大小和字段对齐。",
    "能解释 unsafe.Pointer 与 uintptr 的区别。",
    "能说出 cgo 引入的构建和部署成本。",
  ],
  reviewQuestions: [
    "为什么 Go 默认不暴露结构体和运行时的全部底层细节？",
    "什么时候调整字段顺序值得做，什么时候不值得？",
    "uintptr 为什么不能当作长期保存的指针？",
    "引入 cgo 前应该评估哪些维护成本？",
  ],
  nextMissionSlugs: [],
  contentSource: {
    primary: "本章基于 /home/wayne/source/open/gopl-zh.github.com/ch13 全章精简重组。",
    references: ["The Go Programming Language 第 13 章", "unsafe 包官方文档", "cgo 官方文档", "Go blog: Profiling Go Programs"],
    license: "gopl-zh.github.com 使用 BSD 风格许可；后续批量迁移应保留项目级参考与许可说明。",
  },
}
];
