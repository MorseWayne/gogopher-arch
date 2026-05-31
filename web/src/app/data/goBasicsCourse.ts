import type { ComponentType } from "react";
import { mdxGoBasicsChapterMetadata } from "../../content/go-basics/courseChapters";
import { missionCatalog, type Mission } from "./missions";

export type GoCourseDifficulty = "入门" | "基础" | "进阶" | "高级";
export type GoCourseOutputMatch = "trimmed-exact" | "contains";
export type GoCourseExerciseKind = "run" | "edit" | "test" | "debug" | "project" | "review";
export type GoCourseExerciseDifficulty = "warmup" | "core" | "challenge";

export type GoCourseLesson = {
  title: string;
  body: string[];
  code?: string;
};

export type GoCourseModernNote = {
  title: string;
  body: string;
};

export type GoCoursePitfall = {
  title: string;
  symptom: string;
  fix: string;
};

export type GoCourseExercise = {
  id?: string;
  title: string;
  kind?: GoCourseExerciseKind;
  difficulty?: GoCourseExerciseDifficulty;
  concepts?: string[];
  estimatedMinutes?: number;
  prompt: string;
  context?: string;
  starterCode: string;
  expectedOutput: string;
  outputMatch: GoCourseOutputMatch;
  hints: string[];
  solutionOutline?: string[];
};

export type GoCourseContentSource = {
  primary: string;
  references: string[];
  license?: string;
};

export type GoCourseChapter = {
  slug: string;
  order: number;
  title: string;
  duration: string;
  difficulty: GoCourseDifficulty;
  summary: string;
  goals: string[];
  lessons: GoCourseLesson[];
  lessonCount?: number;
  modernNotes: GoCourseModernNote[];
  engineeringPractices: string[];
  pitfalls: GoCoursePitfall[];
  exercise: GoCourseExercise;
  exercises?: GoCourseExercise[];
  checklist: string[];
  reviewQuestions: string[];
  nextMissionSlugs: string[];
  loadContent?: () => Promise<ComponentType>;
  contentKind?: "structured" | "mdx";
  contentSource?: GoCourseContentSource;
};

export type GoBasicsMdxChapter = Omit<GoCourseChapter, "loadContent" | "contentKind">;

type GoBasicsMdxModule = {
  default: ComponentType;
};

const mdxChapterContentLoaders = import.meta.glob<GoBasicsMdxModule>("../../content/go-basics/*.mdx");

function getMdxChapterContentLoader(chapter: GoBasicsMdxChapter) {
  const slugSuffix = chapter.slug.replace(/^ch\d+-/, "");
  const paddedPath = `../../content/go-basics/ch${String(chapter.order).padStart(2, "0")}-${slugSuffix}.mdx`;
  const plainPath = `../../content/go-basics/${chapter.slug}.mdx`;

  return mdxChapterContentLoaders[paddedPath] ?? mdxChapterContentLoaders[plainPath];
}

const baseGoBasicsChapters: GoCourseChapter[] = [
  {
    slug: "ch1-getting-started",
    order: 1,
    title: "入门",
    duration: "约 45 分钟",
    difficulty: "入门",
    summary: "从第一段可运行的 Go 程序开始，建立命令行、标准输出、错误检查和本地工具链的基础手感。",
    goals: [
      "能解释 package main、func main 与可执行程序入口的关系。",
      "能使用 go run 在 sandbox 中运行最小程序并阅读 stdout。",
      "能区分程序输入、输出、错误和退出状态。",
      "能把小程序拆成清晰的变量、分支和输出步骤。",
    ],
    lessons: [
      {
        title: "从可执行程序入口开始",
        body: [
          "Go 的可执行程序从 main 包中的 main 函数启动。你可以把 main 函数看成程序和操作系统之间的握手点：运行时完成初始化后，会把控制权交给这里。",
          "学习阶段不要只记语法，而要形成运行模型：源码经过编译，依赖包被链接，main 函数执行，stdout/stderr 和 exit code 成为观察程序行为的第一批信号。",
        ],
        code: `package main

import "fmt"

func main() {
    fmt.Println("hello, gopher")
}`,
      },
      {
        title: "用标准输出建立反馈回路",
        body: [
          "fmt.Print、fmt.Println 和 fmt.Printf 是最早使用的调试工具。它们适合展示变量值、分支结果和函数返回值，让你在没有复杂调试器时也能定位问题。",
          "真实后端服务最终会用结构化日志和指标系统，但基础阶段先把 stdout 看清楚：输出是否存在、格式是否稳定、是否多了空格或换行，都会影响自动验收。",
        ],
      },
      {
        title: "错误优先而不是异常优先",
        body: [
          "Go 倾向于把错误作为普通返回值显式处理。调用可能失败的函数后，下一行通常先检查 err，而不是假设一切成功。",
          "这种风格一开始显得啰嗦，但在工程代码中很可贵：失败路径就在成功路径旁边，评审者能直接看到你是否考虑了文件不存在、网络超时和输入非法。",
        ],
      },
    ],
    modernNotes: [
      { title: "Go 1.24+ 仍重视简单启动路径", body: "现代 Go 增加了工具链、标准库和性能改进，但一个服务能否被快速运行、快速观察、快速失败，仍取决于 main 包入口和清晰的初始化流程。" },
      { title: "sandbox 与本地环境要分开理解", body: "浏览器 sandbox 适合练习语法和最小复现；真实项目还会受环境变量、网络、文件系统和容器镜像影响。课程练习刻意保持可移植，避免依赖特定机器。" },
    ],
    engineeringPractices: [
      "先运行最小程序，再逐步增加输入、分支和依赖，避免一次写太多导致错误来源不清。",
      "把示例输出写成稳定文本，便于 CI、评审机器人或课程验收做字符串匹配。",
      "遇到运行失败时同时查看 stdout、stderr、exit code 和耗时，而不是只看最后一行报错。",
    ],
    pitfalls: [
      { title: "文件能编译不代表能运行", symptom: "库包没有 main 函数，go run 时提示不是可执行命令。", fix: "需要运行程序时使用 package main，并提供 func main；可复用逻辑再拆到普通包。" },
      { title: "忽略 err 会隐藏真正原因", symptom: "后续代码出现空值、默认值或 panic，但最初失败点已经被跳过。", fix: "每次调用可能失败的函数后立刻检查 err，并输出足够的上下文。" },
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
    checklist: ["程序能成功编译。", "stdout 中包含 welcome, Gopher。", "能解释 main 函数何时执行。"],
    reviewQuestions: [
      "package main 和普通业务包的差异是什么？",
      "stdout、stderr 和 exit code 分别适合表达什么信息？",
      "为什么 Go 代码中经常在调用后立即检查 err？",
    ],
    nextMissionSlugs: [],
  },
  {
    slug: "ch2-program-structure",
    order: 2,
    title: "程序结构",
    duration: "约 50 分钟",
    difficulty: "入门",
    summary: "理解声明、作用域、包级变量、短变量声明和初始化顺序，写出可读、可维护的小型 Go 程序。",
    goals: [
      "能区分包级声明和函数内声明。",
      "能解释短变量声明何时创建新变量、何时复用旧变量。",
      "能用 const 和 iota 表达稳定的枚举式配置。",
      "能发现变量遮蔽导致的状态错误。",
    ],
    lessons: [
      {
        title: "声明决定名字的生命周期",
        body: [
          "Go 程序由声明组织起来：包声明、import 声明、类型声明、变量声明、常量声明和函数声明共同定义一个包的对外边界。",
          "工程代码中，名字的生命周期越短越容易维护。能放在函数内部的临时变量就不要提升到包级；包级状态应当有明确所有者和并发策略。",
        ],
      },
      {
        title: "短变量声明和遮蔽",
        body: [
          "短变量声明 := 很方便，但它至少需要引入一个新变量。若在内层作用域重新声明同名变量，外层变量不会被更新，这就是常见的遮蔽问题。",
          "遮蔽在 err 变量上尤其危险：你以为修改了外层状态，实际只在 if 或 for 的小范围里创建了新名字。评审时要特别关注冒号是否应该改成等号。",
        ],
        code: `level := 1
if passed {
    level := 2 // 新变量，只在 if 内有效
    _ = level
}`,
      },
      {
        title: "初始化顺序要可预测",
        body: [
          "Go 会先初始化导入包，再初始化当前包的常量、变量，最后执行 init 函数和 main 函数。这个顺序是确定的，但过度依赖 init 会让代码难测试。",
          "后端项目更推荐显式初始化：把配置读取、连接创建和依赖注入放在 main 或启动函数中，让测试可以传入替身对象，而不是被包级副作用绑死。",
        ],
      },
    ],
    modernNotes: [
      { title: "现代 Go 项目更偏向显式依赖", body: "即使工具链支持复杂初始化，服务端代码也更倾向于在 main 中组装依赖，在 handler、service、repository 层传递接口或结构体。" },
      { title: "go vet 与编辑器能发现部分遮蔽", body: "静态工具可以提示可疑写法，但不能替你判断业务状态是否被正确更新。理解作用域仍是阅读代码的基础能力。" },
    ],
    engineeringPractices: [
      "把临时变量限制在最小作用域，减少后续维护者需要追踪的状态。",
      "遇到 if value, err := call(); err != nil 这种写法时确认 value 不需要在外层继续使用。",
      "包级变量只存放真正共享且不可轻易改变的对象，例如只读配置、常量表或受保护的缓存。",
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
    checklist: ["能解释 := 和 = 的差异。", "能指出变量的有效作用域。", "能避免把临时状态提升到包级。"],
    reviewQuestions: [
      "短变量声明至少需要满足什么条件？",
      "为什么 err 遮蔽会导致排障困难？",
      "哪些初始化逻辑不适合放进 init 函数？",
    ],
    nextMissionSlugs: [],
  },
  {
    slug: "ch3-basic-data-types",
    order: 3,
    title: "基础数据类型",
    duration: "约 55 分钟",
    difficulty: "基础",
    summary: "掌握数字、字符串、布尔值、rune 和零值，理解类型选择如何影响接口、存储和国际化输入。",
    goals: [
      "能解释零值如何简化结构体初始化。",
      "能区分 byte 长度和 rune 数量。",
      "能根据业务含义选择 int、int64、float64 或 decimal 替代方案。",
      "能避免字符串切片破坏 UTF-8 字符。",
    ],
    lessons: [
      {
        title: "零值是可用状态的一部分",
        body: [
          "Go 为每种类型定义零值：数字为 0，布尔为 false，字符串为空串，指针、slice、map、channel、函数和接口为 nil。",
          "好的结构体设计会让零值尽量安全可用。例如 bytes.Buffer 的零值就能直接写入；而需要外部资源的类型则应通过构造函数表达依赖。",
        ],
      },
      {
        title: "字符串是只读字节序列",
        body: [
          "Go 字符串底层是一段只读字节，通常保存 UTF-8 文本。len(string) 返回字节数，不是用户看到的字符数。",
          "处理中文昵称、emoji 或多语言输入时，要根据需求选择 []rune、utf8 包或更完整的文本分段库。简单把字符串按字节截断，可能得到非法 UTF-8。",
        ],
        code: `name := "Go语言"
fmt.Println(len(name), len([]rune(name)))`,
      },
      {
        title: "数值类型要表达业务边界",
        body: [
          "int 适合循环计数和本机内存索引；int64 常用于数据库 ID、时间戳和跨平台协议；float64 适合科学计算但不适合金额精确计算。",
          "类型转换在 Go 中是显式的，这能减少隐式溢出和精度损失。写服务接口时，应提前明确字段单位、范围和序列化格式。",
        ],
      },
    ],
    modernNotes: [
      { title: "UTF-8 仍是 Web 服务默认文本假设", body: "现代 API、日志和数据库通常以 UTF-8 为默认编码，但输入可能来自不同终端、浏览器和第三方系统。边界层需要校验和规范化。" },
      { title: "泛型不替代基础类型判断", body: "Go 1.18+ 的泛型能抽象容器和算法，但金额、ID、状态码等业务类型仍需要清晰命名和边界校验。" },
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
    checklist: ["能解释字符串长度的两种常见含义。", "能说出基础类型的零值。", "能为业务字段选择合适数值类型。"],
    reviewQuestions: [
      "为什么字符串按字节截断可能产生乱码？",
      "什么时候 int64 比 int 更适合作为字段类型？",
      "零值设计对配置结构体有什么帮助和风险？",
    ],
    nextMissionSlugs: [],
  },
  {
    slug: "ch4-composite-types",
    order: 4,
    title: "复合数据类型",
    duration: "约 70 分钟",
    difficulty: "基础",
    summary: "学习数组、slice、map、struct 和 JSON 建模，理解底层数组、引用语义和后端数据结构设计。",
    goals: [
      "能解释 slice header 和底层数组的关系。",
      "能使用 map 做计数、去重和索引。",
      "能设计清晰的 struct 表达请求、响应和领域对象。",
      "能识别 slice 保留大数组导致的内存问题。",
    ],
    lessons: [
      {
        title: "slice 是窗口，不是数组本身",
        body: [
          "slice 由指向底层数组的指针、长度和容量组成。多个 slice 可以共享同一块底层数组，因此修改元素可能影响另一个视图。",
          "后端批处理常见问题是从大切片中截取小片段并长期保存，导致整块底层数组无法释放。需要保留少量数据时，主动 copy 到新切片更安全。",
        ],
      },
      {
        title: "map 适合表达索引和集合",
        body: [
          "map 的零值为 nil，可以读取但不能写入；写入前需要 make。它很适合状态计数、ID 到对象的索引、以及集合式存在性判断。",
          "map 的遍历顺序是故意不稳定的。任何依赖输出顺序的逻辑都应显式取出 key 并排序，尤其是生成签名、快照和测试输出时。",
        ],
        code: `counts := map[string]int{}
for _, status := range statuses {
    counts[status]++
}`,
      },
      {
        title: "struct 是 API 和领域模型的契约",
        body: [
          "struct 字段名、类型和标签共同决定 JSON、数据库和日志中的形状。命名不清会把歧义扩散到多个系统边界。",
          "不要把所有字段都塞进一个万能结构体。请求参数、内部领域对象、响应 DTO 可以分开定义，让校验、默认值和敏感字段控制更明确。",
        ],
      },
    ],
    modernNotes: [
      { title: "slices 包提供了更多通用工具", body: "现代 Go 标准库提供 slices、maps 等辅助包，排序、克隆、比较等操作更直接。但理解底层共享关系仍然是避免内存和并发 bug 的关键。" },
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
    },
    checklist: ["能解释 slice 的 len 和 cap。", "能用 map 完成计数。", "能说明结构体标签的作用。"],
    reviewQuestions: [
      "为什么小 slice 可能保留大数组？",
      "map 的遍历顺序为什么不能作为业务逻辑依赖？",
      "请求结构体和数据库结构体混用有什么风险？",
    ],
    nextMissionSlugs: ["slice-memory-leak"],
  },
  {
    slug: "ch5-functions",
    order: 5,
    title: "函数",
    duration: "约 65 分钟",
    difficulty: "基础",
    summary: "理解函数签名、多返回值、错误返回、defer 和闭包，把业务步骤拆成可测试的小单元。",
    goals: [
      "能用函数签名表达输入、输出和错误。",
      "能合理使用 defer 释放资源。",
      "能识别闭包捕获变量带来的可读性和并发风险。",
      "能把解析、校验和执行拆成可测试函数。",
    ],
    lessons: [
      {
        title: "函数签名是最小接口",
        body: [
          "函数签名告诉调用者需要提供什么、会得到什么、失败时如何表达。多返回值让 Go 可以自然返回结果和错误。",
          "写业务函数时，不要让函数偷偷读取全局状态或环境变量。把依赖作为参数传入，函数会更容易测试，也更容易在评审中看懂影响范围。",
        ],
      },
      {
        title: "defer 管理资源生命周期",
        body: [
          "defer 会在当前函数返回前逆序执行，适合关闭文件、释放锁、记录耗时和恢复 panic。它让清理逻辑靠近资源获取位置。",
          "defer 不是异步执行，也不会跨函数延迟。循环中大量 defer 可能推迟资源释放，处理长列表文件或连接时要注意作用域。",
        ],
        code: `f, err := os.Open(path)
if err != nil {
    return err
}
defer f.Close()`,
      },
      {
        title: "闭包让行为和状态靠得更近",
        body: [
          "闭包可以捕获外层变量，适合构造小型回调、统计器或中间件。它能减少参数传递，但也会隐藏状态变化。",
          "并发场景中闭包捕获循环变量曾经是高频坑。现代 Go 已改善常见 range 捕获行为，但你仍应让 goroutine 参数显式，避免读者猜测变量何时变化。",
        ],
      },
    ],
    modernNotes: [
      { title: "range 变量捕获语义已有改进", body: "新版本 Go 对常见循环变量捕获问题做了语言层改进，但老代码、索引变量和手写复用变量仍需要审查。课程建议继续显式传参，保持意图清楚。" },
      { title: "错误包装是服务排障基础", body: "fmt.Errorf 使用 %w 包装错误后，上层可以用 errors.Is/As 判断原因，同时保留上下文。基础阶段要养成返回带上下文错误的习惯。" },
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
    checklist: ["能写出返回 error 的函数。", "能解释 defer 的执行时机。", "能识别闭包捕获状态。"],
    reviewQuestions: [
      "函数签名如何帮助调用者理解失败路径？",
      "defer 为什么适合资源释放，但不适合无限制放在大循环里？",
      "闭包捕获变量时，什么时候应该改成显式参数？",
    ],
    nextMissionSlugs: ["defer-order"],
  },
  {
    slug: "ch6-methods",
    order: 6,
    title: "方法",
    duration: "约 55 分钟",
    difficulty: "基础",
    summary: "学习方法接收者、值语义、指针语义和封装边界，用小结构体承载业务行为。",
    goals: [
      "能为结构体定义方法并选择值接收者或指针接收者。",
      "能解释方法集对接口实现的影响。",
      "能把数据和行为组织成内聚对象。",
      "能避免在方法中制造隐藏的全局副作用。",
    ],
    lessons: [
      {
        title: "方法让行为靠近数据",
        body: [
          "方法是带接收者的函数。它让某类数据的行为更容易发现，例如 Account.AddPoints 比 AddPoints(account, n) 更像领域语言。",
          "不要把方法当作面向对象继承的替代品。Go 更强调组合、清晰接收者和小接口，复杂层级通常会降低可读性。",
        ],
      },
      {
        title: "值接收者和指针接收者",
        body: [
          "值接收者会复制接收者，适合小型不可变值；指针接收者可以修改原对象，也能避免复制大型结构体。",
          "同一类型的方法最好保持接收者风格一致。若部分方法需要指针接收者，通常全部使用指针接收者更少让调用者困惑。",
        ],
        code: `func (a *Account) Add(points int) {
    a.Points += points
}`,
      },
      {
        title: "方法集影响接口满足关系",
        body: [
          "一个类型是否实现接口，取决于它的方法集。指针接收者方法属于 *T 的方法集，不一定属于 T 的方法集。",
          "当接口赋值失败时，不要只看方法名是否存在，还要检查接收者类型、参数和返回值是否完全一致。",
        ],
      },
    ],
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
    checklist: ["能定义带接收者的方法。", "能说明值接收者和指针接收者差异。", "能解释方法集对接口的影响。"],
    reviewQuestions: [
      "什么时候值接收者比指针接收者更合适？",
      "为什么同一类型的方法接收者风格最好保持一致？",
      "方法和普通函数在组织业务代码时各有什么优势？",
    ],
    nextMissionSlugs: [],
  },
  {
    slug: "ch7-interfaces",
    order: 7,
    title: "接口",
    duration: "约 75 分钟",
    difficulty: "进阶",
    summary: "用小接口表达能力边界，理解隐式实现、nil 接口和依赖倒置，让代码更容易测试和替换。",
    goals: [
      "能定义只包含必要方法的小接口。",
      "能解释 Go 的接口是隐式实现。",
      "能识别 nil 接口和带类型 nil 值的差异。",
      "能用接口替换外部依赖以便单元测试。",
    ],
    lessons: [
      {
        title: "接口描述能力而不是家族",
        body: [
          "Go 接口通常很小，描述调用者真正需要的能力。例如 io.Reader 只要求 Read，这让文件、网络连接、压缩流都能被统一处理。",
          "设计接口时从使用方出发，而不是从实现方出发。调用者只需要发送通知，就定义 Notifier；不要把数据库、缓存和日志方法都塞进一个巨大接口。",
        ],
      },
      {
        title: "隐式实现降低耦合",
        body: [
          "类型不需要显式声明实现了某接口，只要方法集匹配即可。这让第三方类型也能自然满足你的接口。",
          "隐式实现要求你关注方法签名的精确匹配。参数、返回值、接收者和包路径不同都会导致接口不满足。",
        ],
        code: `type Notifier interface {
    Notify(message string) string
}`,
      },
      {
        title: "nil 接口要谨慎",
        body: [
          "接口值由动态类型和动态值组成。一个接口只在类型和值都为空时才等于 nil；如果动态类型存在但动态值是 nil 指针，接口本身并不为 nil。",
          "这类问题常出现在返回自定义错误、可选依赖或 mock 对象时。返回接口类型时，尽量避免把 nil 指针装进接口。",
        ],
      },
    ],
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
    checklist: ["能写出小接口。", "能说明隐式实现。", "能描述 nil 接口问题。"],
    reviewQuestions: [
      "为什么接口应尽量由消费方定义？",
      "接口值在什么情况下等于 nil？",
      "泛型和接口分别适合解决什么问题？",
    ],
    nextMissionSlugs: [],
  },
  {
    slug: "ch8-goroutines-channels",
    order: 8,
    title: "Goroutines 和 Channels",
    duration: "约 80 分钟",
    difficulty: "进阶",
    summary: "建立 goroutine 生命周期、channel 通信、select、取消和超时的工程直觉，写出不会泄露的并发代码。",
    goals: [
      "能启动 goroutine 并用 channel 收集结果。",
      "能解释 buffered 和 unbuffered channel 的差异。",
      "能使用 select 处理超时、取消和多路事件。",
      "能识别 goroutine 泄露的常见模式。",
    ],
    lessons: [
      {
        title: "goroutine 不是免费线程",
        body: [
          "goroutine 很轻量，但不是没有成本。每个 goroutine 都需要栈、调度和生命周期管理；失控创建会造成内存、调度和下游依赖压力。",
          "后端并发代码首先要回答三个问题：谁启动它，谁等待它，谁在请求取消或服务关闭时通知它退出。回答不清，就可能泄露。",
        ],
      },
      {
        title: "channel 用通信表达同步",
        body: [
          "channel 可以传递值，也可以传递完成信号。无缓冲 channel 要求发送和接收同时准备好；有缓冲 channel 可以吸收有限突发。",
          "不要把 channel 当成万能队列。容量大小、关闭时机、发送方数量和接收方退出条件都要明确，否则容易出现永久阻塞或向已关闭 channel 发送。",
        ],
        code: `results := make(chan string, 2)
go func() { results <- "cache" }()
go func() { results <- "db" }()`,
      },
      {
        title: "select 是并发控制台",
        body: [
          "select 允许同时等待多个 channel 事件，常用于结果、取消、超时和心跳。它让并发逻辑不必拆成多个互相竞争的阻塞点。",
          "真实服务中，context cancellation 应贯穿请求链。goroutine 如果只等待自己的结果 channel，却不监听 ctx.Done，调用方超时后它仍可能继续占用资源。",
        ],
      },
    ],
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
    checklist: ["能用 channel 收集 goroutine 结果。", "能说明 channel 关闭所有权。", "能描述 goroutine 泄露的触发条件。"],
    reviewQuestions: [
      "为什么启动 goroutine 时必须同时考虑退出条件？",
      "buffered channel 能解决哪些问题，又会掩盖哪些问题？",
      "select 中监听 ctx.Done 的意义是什么？",
    ],
    nextMissionSlugs: ["slice-memory-leak"],
  },
  {
    slug: "ch9-shared-variable-concurrency",
    order: 9,
    title: "基于共享变量的并发",
    duration: "约 75 分钟",
    difficulty: "进阶",
    summary: "理解数据竞争、互斥锁、读写锁、原子操作和并发安全数据结构，避免共享状态把服务拖垮。",
    goals: [
      "能解释 data race 和逻辑竞争的区别。",
      "能使用 sync.Mutex 保护共享变量。",
      "能判断何时使用 channel、锁或 atomic。",
      "能为并发代码设计可复现的测试。",
    ],
    lessons: [
      {
        title: "共享变量需要明确保护策略",
        body: [
          "多个 goroutine 同时访问同一变量，只要至少一个是写操作且没有同步，就可能产生数据竞争。结果可能丢失更新、读到中间状态，甚至触发运行时崩溃。",
          "保护策略要写进代码结构：这个字段由哪把锁保护，哪些方法持锁访问，哪些快照可以无锁读取。只靠口头约定很难长期维护。",
        ],
      },
      {
        title: "Mutex 保护的是不变量",
        body: [
          "锁不是为了保护某一行代码，而是保护一组状态之间的不变量。例如库存数量和预留数量必须一起更新，就应在同一把锁内完成。",
          "持锁范围太小会破坏不变量，太大又会降低并发度。先保证正确性，再通过测量决定是否需要优化锁粒度。",
        ],
        code: `mu.Lock()
total++
mu.Unlock()`,
      },
      {
        title: "atomic 适合简单数值状态",
        body: [
          "原子操作适合计数器、开关和指标这类简单状态。它比锁更轻，但不适合维护多个字段之间的关系。",
          "如果代码里出现多次 atomic 操作才能表达一个业务动作，就要警惕：你可能需要锁、channel 或重新设计数据所有权。",
        ],
      },
    ],
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
    checklist: ["能解释数据竞争。", "能用 Mutex 保护共享变量。", "能说明 race detector 的作用。"],
    reviewQuestions: [
      "数据竞争和业务层面的竞态条件有什么不同？",
      "什么时候锁比 channel 更直接？",
      "为什么含 Mutex 的结构体通常不应复制？",
    ],
    nextMissionSlugs: ["map-concurrent-write"],
  },
  {
    slug: "ch10-packages-tools",
    order: 10,
    title: "包和工具",
    duration: "约 60 分钟",
    difficulty: "基础",
    summary: "掌握 package、module、go fmt、go test、go vet 和构建标签，理解工具链如何支撑团队协作。",
    goals: [
      "能解释 module、package 和 import path 的关系。",
      "能使用 gofmt 保持代码风格一致。",
      "能区分 go run、go test、go build 的使用场景。",
      "能理解工具链环境对构建结果的影响。",
    ],
    lessons: [
      {
        title: "module 是版本和依赖边界",
        body: [
          "go.mod 定义模块路径、Go 版本和依赖版本。一个仓库可以有一个或多个 module，但多数后端服务会从单 module 开始。",
          "import path 不是随意字符串，它代表依赖的身份。修改模块路径会影响所有导入方，因此项目创建早期就应确定清晰路径。",
        ],
      },
      {
        title: "package 是编译和封装单元",
        body: [
          "同一目录下的 Go 文件通常属于同一个 package。导出的标识符以大写字母开头，构成其他包可见的 API。",
          "包应围绕稳定职责组织，而不是按技术动作拆碎。过细会导致导入网复杂，过粗会让内部边界不清。",
        ],
      },
      {
        title: "工具链是团队约定的执行者",
        body: [
          "gofmt 统一格式，go test 运行测试，go vet 检查可疑代码，go build 验证编译产物。它们让团队不用在风格和基础错误上反复争论。",
          "CI 中应固定 Go 版本并缓存依赖，但不要让本地环境和 CI 完全脱节。开发者需要知道如何在本地复现构建失败。",
        ],
        code: `go test ./...
go vet ./...
go build ./...`,
      },
    ],
    modernNotes: [
      { title: "toolchain 指令改善版本协作", body: "现代 Go module 可以通过 go/toolchain 信息表达期望工具链。团队仍应在 README、CI 和容器镜像中明确使用版本。" },
      { title: "工作区适合多模块联调", body: "go work 能把多个 module 放进同一工作区，适合本地同时修改服务和内部库。但发布和 CI 仍要验证独立 module 行为。" },
    ],
    engineeringPractices: [
      "提交前运行 gofmt 或配置编辑器保存时自动格式化。",
      "CI 至少执行 go test ./... 和关键构建命令，避免只在本机可用。",
      "包名短小且表达职责，避免 util、common 这类不断膨胀的垃圾桶包。",
    ],
    pitfalls: [
      { title: "循环导入", symptom: "两个包互相 import，编译失败。", fix: "上移共同抽象或重新划分依赖方向，保持单向依赖。" },
      { title: "包名和目录名混乱", symptom: "导入后使用的名字与路径不一致，阅读成本高。", fix: "尽量让目录名和包名一致，必要别名要有充分理由。" },
      { title: "只在 IDE 里能跑", symptom: "命令行或 CI 构建失败。", fix: "以 go 命令为准，把环境变量和启动参数写入文档或脚本。" },
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
      hints: ["不同 sandbox 可能运行在不同 GOOS/GOARCH 上。", "contains 匹配允许保留平台差异。"],
    },
    checklist: ["能解释 module 和 package。", "能说出 gofmt/go test/go build 的职责。", "能识别循环导入问题。"],
    reviewQuestions: [
      "为什么 import path 是项目 API 的一部分？",
      "什么时候应该拆包，什么时候不该拆？",
      "本地和 CI 的 Go 版本不一致可能造成什么问题？",
    ],
    nextMissionSlugs: [],
  },
  {
    slug: "ch11-testing",
    order: 11,
    title: "测试",
    duration: "约 75 分钟",
    difficulty: "进阶",
    summary: "用表驱动测试、子测试、测试替身和基准测试建立质量反馈，写出能保护重构的 Go 代码。",
    goals: [
      "能组织表驱动测试覆盖多个输入输出组合。",
      "能使用 t.Run 表达子场景。",
      "能通过接口注入替身隔离外部依赖。",
      "能理解单元测试、集成测试和基准测试的边界。",
    ],
    lessons: [
      {
        title: "测试是可执行规格",
        body: [
          "好的测试不仅验证当前实现，还记录业务规则。后来者阅读失败用例，应该能理解输入、预期输出和边界条件。",
          "不要只测试快乐路径。解析失败、空输入、重复数据、超时和权限不足才是后端服务最容易出事故的地方。",
        ],
      },
      {
        title: "表驱动测试降低重复",
        body: [
          "Go 社区常用结构体切片列出测试用例，再循环执行。每个用例包含 name、input、want 和可选 wantErr。",
          "表驱动不是为了炫技，而是让新增边界条件变成新增一行数据。测试结构稳定后，评审更容易聚焦业务差异。",
        ],
        code: `tests := []struct {
    name string
    input int
    want int
}{...}`,
      },
      {
        title: "测试替身隔离外部世界",
        body: [
          "单元测试不应依赖真实支付、邮件或第三方 API。通过小接口注入 fake，可以验证业务逻辑如何处理成功、失败和重试。",
          "集成测试仍然重要，但它们更慢、更脆弱，适合覆盖关键链路。团队应明确哪些测试在每次提交运行，哪些在夜间或发布前运行。",
        ],
      },
    ],
    modernNotes: [
      { title: "fuzz test 适合输入解析", body: "Go 内置 fuzzing 能自动生成输入，适合解析器、编码器和安全敏感函数。基础阶段先掌握表驱动，再把复杂输入函数纳入 fuzz。" },
      { title: "测试缓存和并行要理解", body: "go test 会缓存成功结果，t.Parallel 可提升速度但会暴露共享状态污染。遇到奇怪结果时，先确认缓存、并行和外部资源隔离。" },
    ],
    engineeringPractices: [
      "每个 bug 修复尽量补一个失败优先的回归测试。",
      "测试名称表达业务场景，而不是只写 case1、case2。",
      "对依赖时间、随机数和外部服务的代码注入可控替身。",
    ],
    pitfalls: [
      { title: "测试只覆盖实现细节", symptom: "重构内部结构后大量测试失败，但行为没有变化。", fix: "优先测试外部可观察行为，少断言私有中间步骤。" },
      { title: "共享状态污染", symptom: "单独运行通过，整包运行失败。", fix: "清理全局状态，使用 t.Cleanup，并避免测试间共享可变对象。" },
      { title: "忽略错误分支", symptom: "生产事故来自测试从未覆盖的失败路径。", fix: "表格中加入非法输入、依赖失败和边界值。" },
    ],
    exercise: {
      title: "运行表驱动思路",
      prompt: "用表格描述两个加法用例，模拟测试统计输出。",
      starterCode: `package main

import "fmt"

func add(a, b int) int {
    return a + b
}

func main() {
    tests := []struct {
        a, b int
        want int
    }{
        {1, 2, 3},
        {-1, 1, 0},
    }

    passed := 0
    for _, tt := range tests {
        if add(tt.a, tt.b) == tt.want {
            passed++
        }
    }
    fmt.Printf("passed=%d/%d\\n", passed, len(tests))
}`,
      expectedOutput: "passed=2/2",
      outputMatch: "trimmed-exact",
      hints: ["真实项目中这些逻辑会放在 *_test.go 里。", "表驱动的关键是让新增用例足够低成本。"],
    },
    checklist: ["能描述表驱动测试结构。", "能区分单元测试和集成测试。", "能说明测试替身的价值。"],
    reviewQuestions: [
      "为什么测试名称应该描述业务场景？",
      "什么时候应该为 bug 补回归测试？",
      "并行测试为什么容易暴露共享状态问题？",
    ],
    nextMissionSlugs: [],
  },
  {
    slug: "ch12-reflection",
    order: 12,
    title: "反射",
    duration: "约 70 分钟",
    difficulty: "高级",
    summary: "理解 reflect 的能力和代价，知道 JSON、ORM、验证器如何读取类型信息，并学会在必要时克制使用。",
    goals: [
      "能说明反射在运行时检查类型和值的基本机制。",
      "能读取结构体字段标签。",
      "能理解反射带来的性能、可读性和安全成本。",
      "能判断何时用接口、泛型或代码生成替代反射。",
    ],
    lessons: [
      {
        title: "反射让程序观察自身",
        body: [
          "reflect 包可以在运行时获取类型、字段、方法和值。JSON 编解码、ORM 映射和验证器都依赖类似能力理解结构体。",
          "这种能力很强，也会绕开编译期类型检查的一部分保护。普通业务逻辑如果大量依赖反射，调试和重构成本会显著上升。",
        ],
      },
      {
        title: "Value 和 Type 要一起看",
        body: [
          "reflect.Type 描述类型信息，reflect.Value 描述具体值。读取字段、调用方法和设置值时，需要同时考虑可寻址性和导出规则。",
          "未导出字段、不可设置值和 nil 值都会让反射代码变复杂。写库时要把这些边界转成清晰错误，而不是让 panic 泄漏给调用方。",
        ],
        code: `field := reflect.TypeOf(User{}).Field(0)
jsonName := field.Tag.Get("json")`,
      },
      {
        title: "反射适合框架边界",
        body: [
          "反射最适合出现在通用库和框架边界，例如根据标签序列化字段、自动绑定配置或实现通用验证。",
          "在业务核心中，接口和泛型通常更清楚。你应该先问：能不能通过明确类型表达？能不能生成代码？只有通用性收益足够大时再使用反射。",
        ],
      },
    ],
    modernNotes: [
      { title: "泛型减少了一部分反射需求", body: "容器、集合和通用算法可以用泛型表达，避免 interface{} 加反射。但 JSON/ORM 这类运行时结构映射仍然会用到反射。" },
      { title: "代码生成仍是高性能选择", body: "对性能敏感的序列化、RPC 和数据库访问常通过生成代码减少反射开销。现代工具链让生成步骤更容易接入 CI。" },
    ],
    engineeringPractices: [
      "把反射代码集中在少数基础设施包中，并提供普通类型安全 API 给业务层。",
      "反射读取标签时定义清楚默认值、空标签和非法标签的处理策略。",
      "对反射路径写边界测试，覆盖 nil、未导出字段、缺失标签和错误类型。",
    ],
    pitfalls: [
      { title: "CanSet 判断缺失", symptom: "反射设置字段时 panic。", fix: "设置前检查 Value 是否可设置，并确保传入的是指针。" },
      { title: "把业务逻辑写成字符串字段名", symptom: "重命名字段后运行时才失败。", fix: "业务代码优先使用强类型访问，反射只放在边界层。" },
      { title: "忽略反射性能成本", symptom: "热点路径 CPU 占用异常。", fix: "用 benchmark 量化，再考虑缓存元数据、泛型或代码生成。" },
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
    field, _ := reflect.TypeOf(User{}).FieldByName("Name")
    fmt.Printf("json=%s\\n", field.Tag.Get("json"))
}`,
      expectedOutput: "json=name",
      outputMatch: "trimmed-exact",
      hints: ["结构体标签是反射可读取的元数据。", "FieldByName 会返回字段和是否存在的布尔值。"],
    },
    checklist: ["能读取结构体标签。", "能说明反射适用边界。", "能说出反射替代方案。"],
    reviewQuestions: [
      "为什么反射代码更难重构？",
      "泛型能替代哪些反射场景，不能替代哪些？",
      "反射库应如何把 panic 转化为可诊断错误？",
    ],
    nextMissionSlugs: [],
  },
  {
    slug: "ch13-low-level-programming",
    order: 13,
    title: "底层编程",
    duration: "约 60 分钟",
    difficulty: "高级",
    summary: "认识 unsafe、内存布局、对齐、cgo 和底层优化边界，知道何时应该远离以及如何隔离风险。",
    goals: [
      "能解释 unsafe 的能力和危险。",
      "能观察结构体大小和字段对齐。",
      "能理解底层优化必须以基准测试为依据。",
      "能把不安全代码隔离在小范围并补充测试。",
    ],
    lessons: [
      {
        title: "unsafe 是逃生门不是日常工具",
        body: [
          "unsafe 允许绕过 Go 类型系统做指针转换、地址运算和布局观察。它能连接底层系统，也能破坏内存安全和未来兼容性。",
          "业务服务中绝大多数性能问题不需要 unsafe。先优化算法、数据结构、分配次数、IO 批量和缓存策略，再考虑更底层的手段。",
        ],
      },
      {
        title: "内存布局影响大小和缓存",
        body: [
          "结构体字段会按对齐规则排列，字段顺序可能影响整体大小。理解布局有助于处理大规模对象、二进制协议和性能热点。",
          "不要为了节省几个字节牺牲可读性，除非对象数量巨大且有测量证据。优化应由 benchmark 和 profile 驱动，而不是猜测。",
        ],
        code: `type Metric struct {
    Count int64
    Healthy bool
}`,
      },
      {
        title: "隔离底层边界",
        body: [
          "如果确实需要 unsafe 或 cgo，把它们封装在小包中，提供普通 Go API 给其他代码使用。这样风险边界更清楚，测试也更集中。",
          "底层代码还要关注不同 GOOS、GOARCH、编译器版本和内存模型差异。跨平台项目尤其不能只在本机验证一次。",
        ],
      },
    ],
    modernNotes: [
      { title: "标准库持续减少 unsafe 必要性", body: "随着 slices、maps、unsafe 相关文档和 runtime 优化演进，很多过去需要手写底层技巧的场景已有更安全方案。优先选择标准库和清晰代码。" },
      { title: "性能优化要结合 pprof", body: "现代 Go 服务通常先用 benchmark、pprof、trace 找瓶颈。没有数据支撑的 unsafe 优化，很可能增加风险却没有收益。" },
    ],
    engineeringPractices: [
      "底层优化前先写 benchmark，记录优化前后的 CPU、内存和分配数据。",
      "把 unsafe 代码集中封装，并在包注释中说明不变量和平台假设。",
      "跨平台敏感代码在 CI 中覆盖目标 GOOS/GOARCH，避免只在开发机通过。",
    ],
    pitfalls: [
      { title: "把 unsafe 当性能捷径", symptom: "代码更难懂，但 profile 中瓶颈并不在这里。", fix: "先测量，再优化；优先选择算法和分配优化。" },
      { title: "依赖结构体布局不写说明", symptom: "字段调整后协议或二进制解析悄悄损坏。", fix: "为布局假设写测试和注释，必要时使用显式编码。" },
      { title: "忽略平台差异", symptom: "本机正常，换架构后大小、对齐或 cgo 行为不同。", fix: "在目标平台构建测试，并避免不必要的平台假设。" },
    ],
    exercise: {
      title: "观察结构体大小",
      prompt: "使用 unsafe.Sizeof 观察结构体布局带来的大小，建立对对齐的直觉。",
      starterCode: `package main

import (
    "fmt"
    "unsafe"
)

type Metric struct {
    Count int64
    Healthy bool
}

func main() {
    fmt.Printf("size=%d\\n", unsafe.Sizeof(Metric{}))
}`,
      expectedOutput: "size=16",
      outputMatch: "trimmed-exact",
      hints: ["int64 通常需要 8 字节对齐。", "bool 只有 1 字节，但结构体整体会包含填充。"],
    },
    checklist: ["能说明 unsafe 的风险。", "能观察结构体大小。", "能用 benchmark 支撑优化决定。"],
    reviewQuestions: [
      "为什么 unsafe 代码应该集中封装？",
      "字段顺序如何影响结构体大小？",
      "没有 profile 数据时进行底层优化有什么风险？",
    ],
    nextMissionSlugs: [],
  },
];

const mdxChapterOverrides = Object.fromEntries(
  mdxGoBasicsChapterMetadata.map((chapter) => {
    const loadModule = getMdxChapterContentLoader(chapter);

    return [
      chapter.slug,
      {
        ...chapter,
        loadContent: async () => {
          if (!loadModule) {
            throw new Error(`missing MDX content loader for ${chapter.slug}`);
          }

          const module = await loadModule();
          return module.default;
        },
        contentKind: "mdx" as const,
      },
    ];
  }),
) as Record<string, GoCourseChapter>;

const baseGoBasicsChapterSlugs = new Set(baseGoBasicsChapters.map((chapter) => chapter.slug));

export const goBasicsChapters: GoCourseChapter[] = [
  ...baseGoBasicsChapters.map((chapter) => mdxChapterOverrides[chapter.slug] ?? chapter),
  ...Object.values(mdxChapterOverrides).filter((chapter) => !baseGoBasicsChapterSlugs.has(chapter.slug)),
].sort((left, right) => left.order - right.order);

export function getGoBasicsLessonCount(chapter: GoCourseChapter) {
  return chapter.lessonCount ?? chapter.lessons.length;
}

export function getGoBasicsExerciseCount(chapter: GoCourseChapter) {
  return chapter.exercises?.length ?? (chapter.exercise ? 1 : 0);
}

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
  const externalSourceKeys = ["source" + "Path", "source" + "Url"];

  if (goBasicsChapters.length !== 13) {
    errors.push(`expected 13 chapters, got ${goBasicsChapters.length}`);
  }

  for (const chapter of goBasicsChapters) {
    if (slugs.has(chapter.slug)) {
      errors.push(`duplicate chapter slug: ${chapter.slug}`);
    }
    slugs.add(chapter.slug);

    for (const key of externalSourceKeys) {
      if (Object.prototype.hasOwnProperty.call(chapter, key)) {
        errors.push(`unexpected external source field in ${chapter.slug}`);
      }
    }

    if (!chapter.slug || !chapter.order || !chapter.title || !chapter.duration || !chapter.difficulty || !chapter.summary) {
      errors.push(`missing required chapter metadata: ${chapter.slug || "unknown"}`);
    }

    if (chapter.goals.length === 0 || chapter.checklist.length === 0) {
      errors.push(`missing goals or checklist: ${chapter.slug}`);
    }

    if (getGoBasicsLessonCount(chapter) < 3) {
      errors.push(`expected at least 3 lessons in ${chapter.slug}`);
    }

    if (chapter.lessons.length > 0) {
      for (const lesson of chapter.lessons) {
        if (!lesson.title || lesson.body.filter((paragraph) => paragraph.trim()).length < 2) {
          errors.push(`lesson needs a title and at least 2 paragraphs in ${chapter.slug}`);
        }
      }
    }

    if (chapter.modernNotes.length < 2) {
      errors.push(`expected at least 2 modern notes in ${chapter.slug}`);
    }

    if (chapter.engineeringPractices.length < 3) {
      errors.push(`expected at least 3 engineering practices in ${chapter.slug}`);
    }

    if (chapter.pitfalls.length < 3) {
      errors.push(`expected at least 3 pitfalls in ${chapter.slug}`);
    }

    for (const pitfall of chapter.pitfalls) {
      if (!pitfall.title || !pitfall.symptom || !pitfall.fix) {
        errors.push(`pitfall needs title, symptom and fix in ${chapter.slug}`);
      }
    }

    if (chapter.reviewQuestions.length < 3) {
      errors.push(`expected at least 3 review questions in ${chapter.slug}`);
    }

    const exercises = chapter.exercises ?? [chapter.exercise];
    if (exercises.length === 0) {
      errors.push(`missing required exercise data: ${chapter.slug}`);
    }

    const exerciseIds = new Set<string>();
    for (const exercise of exercises) {
      if (!exercise.title || !exercise.prompt || !exercise.starterCode || !exercise.expectedOutput || !exercise.outputMatch) {
        errors.push(`missing required exercise data: ${chapter.slug}`);
      }

      if (exercise.id) {
        if (exerciseIds.has(exercise.id)) {
          errors.push(`duplicate exercise id ${exercise.id} in ${chapter.slug}`);
        }
        exerciseIds.add(exercise.id);
      }
    }

    for (const missionSlug of chapter.nextMissionSlugs) {
      if (!missionCatalog[missionSlug]) {
        errors.push(`unknown mission slug ${missionSlug} in ${chapter.slug}`);
      }
    }
  }

  return errors;
}
