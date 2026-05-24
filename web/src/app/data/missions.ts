export type MissionStatus = "locked" | "available" | "in-progress" | "completed";

export type Mission = {
  slug: string;
  day: number;
  chapter: string;
  title: string;
  duration: string;
  difficulty: string;
  prerequisite: string[];
  status: MissionStatus;
  background: string[];
  objectives: string[];
  hints: string[];
  knowledge: string[];
  criteria: string[];
  starterCode: string;
};

export const missions: Mission[] = [
  {
    slug: "slice-memory-leak",
    day: 1,
    chapter: "入职第一周",
    title: "修复 Slice 内存泄露",
    duration: "约 30 分钟",
    difficulty: "初级",
    prerequisite: ["Go 基础", "slice 与底层数组"],
    status: "in-progress",
    background: [
      "你刚加入 Go 后端团队，第一张工单来自线上稳定性告警。某个批处理任务在高峰期持续拉高内存，并伴随 Goroutine 数量异常增长。",
      "资深同事已经把最小复现代码交给你：程序只保留了很小的数据片段，却始终无法释放底层大数组，同时有一批 Goroutine 永久阻塞。",
      "你的任务是在不改变业务输出的前提下，让数据处理函数及时释放内存，并确保后台任务可以正常退出。",
    ],
    objectives: [
      "识别 slice 引用底层数组导致的隐性内存保留问题。",
      "理解 Goroutine 永久阻塞会如何放大资源泄露。",
      "使用更安全的数据拷贝和退出控制方式完成修复。",
    ],
    hints: [
      "切片得到的小 slice 仍可能引用原始大数组，垃圾回收器会认为整块数组仍然可达。",
      "空的 select 会永久阻塞当前 Goroutine，真实服务中通常需要 context 或 channel 管理生命周期。",
    ],
    knowledge: ["slice header", "copy 函数", "Goroutine 生命周期", "runtime.NumGoroutine"],
    criteria: [
      "修复 processData 中的底层数组保留问题。",
      "Goroutine 能够在任务结束后退出。",
      "运行后内存和 Goroutine 数量保持稳定。",
      "所有现有测试通过。",
    ],
    starterCode: `package main

import (
    "fmt"
    "runtime"
    "time"
)

func main() {
    fmt.Println("任务开始：处理样本数据...")
    for i := 0; i < 20; i++ {
        go processData()
    }

    time.Sleep(200 * time.Millisecond)
    fmt.Printf("当前 Goroutine 数量: %d\n", runtime.NumGoroutine())
}

func processData() {
    largeData := make([]byte, 64*1024)
    _ = largeData[:10]
    select {}
}`,
  },
  {
    slug: "map-concurrent-write",
    day: 2,
    chapter: "入职第一周",
    title: "Map 并发读写崩溃",
    duration: "约 35 分钟",
    difficulty: "初级",
    prerequisite: ["Go 基础", "并发入门"],
    status: "locked",
    background: [
      "团队的统计接口在压测时偶发崩溃，日志里出现 concurrent map writes。这个问题平时难以复现，但在高并发请求下会快速暴露。",
      "你需要阅读一个共享 map 的计数逻辑，判断哪些访问路径没有同步保护，并选择适合当前任务规模的修复方式。",
    ],
    objectives: [
      "理解原生 map 不支持并发写入的原因。",
      "练习使用 mutex 或 sync.Map 保护共享状态。",
      "通过表征压测验证修复后的稳定性。",
    ],
    hints: [
      "如果所有读写都在同一个临界区内完成，普通 map 依然可以安全使用。",
      "sync.Map 适合读多写少且 key 稳定的场景，不一定是所有任务的首选。",
    ],
    knowledge: ["sync.Mutex", "sync.RWMutex", "sync.Map", "go test -race"],
    criteria: [
      "并发写入不再触发 runtime panic。",
      "计数结果在重复运行时保持一致。",
      "race detector 不报告数据竞争。",
      "保留现有接口签名。",
    ],
    starterCode: `package main

import (
    "fmt"
    "sync"
)

var counters = map[string]int{}

func main() {
    var wg sync.WaitGroup
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counters["request"]++
        }()
    }
    wg.Wait()
    fmt.Println(counters["request"])
}`,
  },
  {
    slug: "defer-order",
    day: 3,
    chapter: "入职第一周",
    title: "Defer 执行顺序迷局",
    duration: "约 25 分钟",
    difficulty: "初级",
    prerequisite: ["Go 函数", "错误处理"],
    status: "locked",
    background: [
      "一个文件处理函数在错误场景下返回了意外结果，日志显示清理逻辑确实执行了，但最终返回值和预期不一致。",
      "你需要定位 defer、命名返回值和错误覆盖之间的关系，修复函数的返回行为，并保留必要的资源清理。",
    ],
    objectives: [
      "掌握 defer 的后进先出执行顺序。",
      "理解 defer 闭包如何读取和修改命名返回值。",
      "避免清理逻辑覆盖真正的业务错误。",
    ],
    hints: [
      "defer 在 return 赋值之后、函数真正返回之前执行。",
      "多个 defer 会按后注册先执行的顺序运行。",
    ],
    knowledge: ["defer", "named return values", "error wrapping", "resource cleanup"],
    criteria: [
      "错误路径返回原始业务错误。",
      "资源清理仍然会被执行。",
      "defer 顺序在测试中被覆盖。",
      "代码可读性适合评审。",
    ],
    starterCode: `package main

import "fmt"

func main() {
    if err := handleFile(); err != nil {
        fmt.Println("failed:", err)
    }
}

func handleFile() (err error) {
    defer func() {
        err = closeFile()
    }()

    return readFile()
}

func readFile() error {
    return fmt.Errorf("read failed")
}

func closeFile() error {
    return nil
}`,
  },
];

export const missionCatalog = Object.fromEntries(missions.map((mission) => [mission.slug, mission])) as Record<string, Mission>;

export const statusMeta: Record<MissionStatus, { label: string; className: string }> = {
  locked: { label: "未解锁", className: "bg-neutral-800 text-neutral-500 border-neutral-700" },
  available: { label: "可开始", className: "bg-[#00ADD8]/10 text-[#00ADD8] border-[#00ADD8]/30" },
  "in-progress": { label: "进行中", className: "bg-yellow-500/10 text-yellow-400 border-yellow-500/30" },
  completed: { label: "已完成", className: "bg-green-500/10 text-green-400 border-green-500/30" },
};

export const goKeywords = new Set([
  "break",
  "case",
  "chan",
  "const",
  "continue",
  "default",
  "defer",
  "else",
  "fallthrough",
  "for",
  "func",
  "go",
  "if",
  "import",
  "interface",
  "map",
  "package",
  "range",
  "return",
  "select",
  "struct",
  "switch",
  "type",
  "var",
]);

export const goFunctions = new Set(["Add", "Done", "NumGoroutine", "Println", "Printf", "Sleep", "Wait", "closeFile", "handleFile", "main", "make", "processData", "readFile"]);

export function getMissionBySlug(slug?: string | null) {
  return slug ? missionCatalog[slug] ?? missions[0] : missions[0];
}
