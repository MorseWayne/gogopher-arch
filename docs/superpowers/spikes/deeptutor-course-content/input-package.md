# DeepTutor 输入包 — ch10 包和工具

Date: 2026-06-03
Spike: DeepTutor 离线课程内容工作流
Target chapter: `ch10-packages-tools`
Target MDX: `web/src/content/go-basics/ch10-packages-tools.mdx`
Metadata source: `web/src/content/go-basics/courseChapters.ts`

## 使用方式

把本输入包交给 DeepTutor，要求它作为“GoGopher Arch 离线课程内容研究员”工作：先开放检索高质量资料，再在站内课程约束下生成一版教程级 MDX 正文草稿，并同时输出完整来源审计材料。

本包不是让 DeepTutor 自动发布内容。DeepTutor 的输出必须经过审计、人工质量评估和构建验证，才能替换目标章节并保留。

## 任务目标

为 GoGopher Arch 的 Go 基础训练营重写 `ch10-packages-tools` 章节正文草稿。

核心目标：

- 生成一版明显优于当前原章节的教程级正文。
- 把 `package`、`import path`、`module`、`workspace`、`toolchain`、Go 命令、构建约束、文档注释、`internal` 和 `go list` 组织成一条清晰的后端工程学习路径。
- 让章节更像真实后端项目从“单文件原型”演进到“可维护服务”的过程，而不是 Go 命令清单或外链资料摘要。
- 保留和现有 warmup/core/challenge 练习的衔接，不默认修改练习 ID、expected output 或 metadata。

## 目标学习者

目标学习者是“后端新手到实习”：

- 已经学过 Go 基础语法、数据类型、函数、方法、接口、并发和测试入门。
- 需要理解真实项目里为什么要划分包、管理依赖、写文档注释、固定工具链和把命令放进 CI。
- 对命令和目录结构容易停留在“照着敲”，需要通过场景理解工程判断。

## 课程风格契约

DeepTutor 生成正文时必须遵守 GoGopher Arch 的课程风格：

1. **具体场景引入**：从真实后端项目问题开始，例如把单文件原型整理成可维护服务、让本地和 CI 复用同一套 Go 命令。
2. **基础概念逐步讲解**：按学习路径解释概念，不在开头堆密集概念地图。
3. **最小示例**：每个关键概念用最小 Go 示例或命令示例说明。
4. **工程化示例**：解释在服务仓库、CI、内部包、文档和依赖图中的使用方式。
5. **常见坑**：覆盖循环依赖、`util/common` 垃圾桶包、匿名导入副作用、workspace 掩盖版本问题、只在本机构建通过等问题。
6. **工程实践**：给出可被代码评审使用的 checklist。
7. **概念回看**：概念表格放在讲解之后，而不是章节开头。
8. **练习衔接**：用 `PracticeBridge` 或正文说明连接现有 warmup/core/challenge 练习。

必须避免：

- 把正文写成 Go 命令备忘录。
- 把外部链接列表当成课程正文。
- 大段搬运外部教程表达。
- 引入当前练习无法支撑的大范围新内容。
- 默认修改 metadata、exercise ID、expected output 或路由。

## 开放检索要求

允许 DeepTutor 开放网页检索，但资料只作为知识来源层。最终正文必须转化为 GoGopher Arch 自己的站内教程。

优先检索和校准来源：

- Go 官方文档：
  - Go Modules Reference
  - go command 文档
  - workspace / `go work`
  - toolchain 指令
  - build constraints
  - documentation comments
  - internal packages
- 本地/站内既有来源：
  - gopl-zh 第 10 章：包和工具主线。
  - GoGopher Arch 当前 ch10 原文和 metadata。
- 质量校准来源：
  - Effective Go
  - Go Code Review Comments
  - 高质量 Go 工程教程或官方博客。

检索输出必须可审计：主要段落应能追溯到一个或多个来源；无法验证或存在争议的技术判断要列入人工审校 checklist。

## 输出要求

DeepTutor 应输出两类内容。

### 1. MDX 正文草稿

要求：

- 输出可直接放入 `web/src/content/go-basics/ch10-packages-tools.mdx` 的 MDX 正文。
- 保持现有组件风格，可使用：
  - `SourceNote`
  - `CompareNote`
  - `ExamplePair`
  - `PitfallCard`
  - `PracticeBridge`
  - `DeepDive`（如确有必要）
- 保留或重新建立现有练习 ID 的连接：
  - `ch10-toolchain-env`
  - `ch10-import-summary`
  - `ch10-internal-rule`
- 不默认修改 metadata 或练习。
- 如果认为必须修改练习或 metadata，单独写“建议修改”，不要直接混入正文。

建议正文结构：

1. 场景引入：从单文件原型整理成可维护后端服务。
2. 你会解决什么问题：把包边界、依赖身份、工具反馈和 CI 联系起来。
3. Package：编译、命名空间和封装单元。
4. 包名和调用点：职责边界与命名判断。
5. Import path：依赖身份和 API 稳定性。
6. Package declaration / main 包 / 外部测试包。
7. Import declaration：分组、别名、匿名导入和副作用。
8. Module / workspace / toolchain：现代 Go 协作边界。
9. Go 工具箱：本地与 CI 的统一反馈回路。
10. 构建、交叉编译和构建约束。
11. 文档注释：导出 API 的第一份契约。
12. `internal`：仓库内部边界。
13. `go list`：让脚本理解 Go 包图。
14. 概念回看。
15. 本章工程视角和评审 checklist。

### 2. 完整审计材料

要求输出以下内容：

- 来源清单：URL、标题、用途、可信度说明。
- 段落级来源映射：正文主要段落或小节对应的来源。
- 搬运风险：指出是否存在与外部资料近似的大段表达；风险等级为 low / medium / high。
- 与原章节差异：说明新增、重组、删除或弱化的内容。
- 不可验证声明：列出需要人工确认的技术判断。
- 人工审校 checklist：帮助 GoGopher Arch 决定保留、修改后保留或回滚。
- 最终建议：保留 / 人工修改后保留 / 回滚。

## 质量 Rubric

生成结果应按以下维度自评：

| 维度 | 合格标准 |
|---|---|
| 技术准确性 | package、module、workspace、toolchain、internal、go list、build tag 等解释符合 Go 官方语义。 |
| 概念深度 | 不只告诉用户命令怎么敲，还解释为什么这些边界影响维护、版本和协作。 |
| 工程场景 | 能自然服务后端实习场景：项目拆包、CI、依赖分析、API 文档、内部实现保护。 |
| 示例解释 | Go 代码、命令和目录结构示例最小、清楚、可解释。 |
| 常见坑 | 覆盖循环依赖、垃圾桶包、匿名导入、副作用、workspace 掩盖版本问题、平台差异。 |
| 练习衔接 | 与 `ch10-toolchain-env`、`ch10-import-summary`、`ch10-internal-rule` 有自然连接。 |
| 非拼贴 | 外部资料被重组为站内教程表达，不出现大段近似搬运。 |
| 可维护性 | 正文结构清晰，后续人工编辑者能继续维护。 |

## 当前章节结构摘要

当前 `ch10-packages-tools.mdx` 的标题结构：

```text
# 包和工具：用边界组织代码，用命令统一反馈
## 你会解决什么问题
## Package：编译、命名空间和封装单元
## 包名和调用点要一起读
## Import path：依赖身份，不只是路径字符串
## Package declaration：目录名、包名和 main 包
## Import declaration：分组、别名和匿名导入
## Module：现代 Go 的版本和依赖边界
## Go 工具箱：统一反馈回路
## 构建、交叉编译和构建约束
## 文档注释：导出 API 的第一份契约
## internal：给仓库内部使用的边界
## go list：让脚本理解 Go 的包图
## 包和工具概念回看
## 本章工程视角
```

当前问题摘要：

- 原章节覆盖面完整，但工程主线可以更强：需要把“单文件原型 → 包边界 → module/workspace → CI 命令反馈 → 文档和 internal 边界 → go list 自动化”串成更明确的迁移故事。
- 当前练习衔接分散：warmup 是工具链环境，core 是 import path 分类，challenge 是 internal 规则。新正文要让这三者成为同一条工程路径的一部分。
- 开放资料不应让章节变长为命令大全；应优先增强判断标准、真实协作场景和常见错误。

## 现有 metadata 条目

```ts
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
    id: "ch10-toolchain-env",
    kind: "run",
    difficulty: "warmup",
    concepts: ["runtime.GOOS", "runtime.GOARCH", "toolchain"],
    estimatedMinutes: 8,
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
    solutionOutline: ["导入 runtime 包。", "读取 runtime.GOOS 和 runtime.GOARCH。", "使用稳定前缀输出，允许不同平台差异。"],
  },
  exercises: [
    {
      id: "ch10-toolchain-env",
      kind: "run",
      difficulty: "warmup",
      concepts: ["runtime.GOOS", "runtime.GOARCH", "toolchain"],
      estimatedMinutes: 8,
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
      solutionOutline: ["导入 runtime 包。", "读取 runtime.GOOS 和 runtime.GOARCH。", "使用稳定前缀输出，允许不同平台差异。"],
    },
    {
      id: "ch10-import-summary",
      kind: "edit",
      difficulty: "core",
      concepts: ["import path", "sort", "strings", "standard library"],
      estimatedMinutes: 18,
      title: "整理导入路径摘要",
      prompt: "补全 ImportSummary：把导入路径分成 standard 和 external，并在每组内按字典序输出。",
      context: "工具脚本和评审机器人常需要分析依赖清单。真实项目应优先用 go list，这里先用字符串规则建立 import path 身份意识。",
      starterCode: `package main

import (
    "fmt"
    "sort"
    "strings"
)

func ImportSummary(paths []string) string {
    standard := []string{}
    external := []string{}
    for _, path := range paths {
        if strings.Contains(path, ".") {
            // TODO: 第三方或组织域名路径放入 external
        } else {
            // TODO: 标准库路径放入 standard
        }
    }
    sort.Strings(standard)
    sort.Strings(external)
    return fmt.Sprintf("standard=%s external=%s", strings.Join(standard, ","), strings.Join(external, ","))
}

func main() {
    paths := []string{"github.com/lib/pq", "fmt", "net/http", "golang.org/x/net/html"}
    fmt.Println(ImportSummary(paths))
}`,
      expectedOutput: "standard=fmt,net/http external=github.com/lib/pq,golang.org/x/net/html",
      outputMatch: "trimmed-exact",
      hints: ["这里用是否包含点号粗略区分外部域名路径。", "append 的结果要重新赋值给 slice。", "输出前分别 sort.Strings。"],
      solutionOutline: ["遍历 paths。", "包含点号的路径 append 到 external，否则 append 到 standard。", "分别排序后用 strings.Join 拼接。"],
    },
    {
      id: "ch10-internal-rule",
      kind: "debug",
      difficulty: "challenge",
      concepts: ["internal package", "strings", "import boundary"],
      estimatedMinutes: 20,
      title: "判断 internal 包导入边界",
      prompt: "当前 CanImportInternal 过于宽松，允许仓库外部路径导入 internal 包。修复判断逻辑。",
      context: "Go 的 internal 规则能防止不稳定内部实现被外部项目依赖。理解这个规则有助于设计大型仓库边界。",
      starterCode: `package main

import (
    "fmt"
    "strings"
)

func CanImportInternal(importer, target string) bool {
    marker := "/internal/"
    index := strings.Index(target, marker)
    if index == -1 {
        return true
    }
    parent := target[:index]
    // TODO: 只有 importer 等于 parent 或在 parent 子树下时才允许
    _ = parent
    return true
}

func main() {
    tests := []struct {
        importer string
        target   string
    }{
        {"example.com/service/cmd/server", "example.com/service/internal/cache"},
        {"example.com/other/app", "example.com/service/internal/cache"},
    }

    allowed := 0
    for _, tt := range tests {
        if CanImportInternal(tt.importer, tt.target) {
            allowed++
        }
    }
    fmt.Printf("allowed=%d/%d\\n", allowed, len(tests))
}`,
      expectedOutput: "allowed=1/2",
      outputMatch: "trimmed-exact",
      hints: ["target 中 /internal/ 前面的部分是允许导入的父目录树。", "importer == parent 应允许。", "strings.HasPrefix(importer, parent+\"/\") 可判断子树。"],
      solutionOutline: ["找到 /internal/ 的位置。", "取 internal 父路径 parent。", "允许 importer == parent 或 strings.HasPrefix(importer, parent+\"/\")。", "其他路径返回 false。"],
    },
  ],
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
    primary: "本章基于 gopl-zh.github.com/ch10 全章精简重组。",
    references: ["The Go Programming Language 第 10 章", "Go Modules Reference", "go command 官方文档", "Effective Go", "Go Code Review Comments"],
    license: "gopl-zh.github.com 使用 BSD 风格许可；后续批量迁移应保留项目级参考与许可说明。",
  },
}
```

## 原章节 MDX 全文

```mdx
# 包和工具：用边界组织代码，用命令统一反馈

Go 项目能长期维护，靠的不只是语法简单。包决定代码边界，模块决定版本和依赖身份，工具链决定团队如何格式化、构建、测试、审查和发布。一个服务刚开始可以只有几个文件，但随着命令、内部库、测试、生成代码和 CI 增加，包和工具的设计会直接影响协作效率。

这一章围绕一个小场景展开：**你要把一个后端服务从单文件原型整理成可维护项目，并让本地和 CI 使用同一套 Go 命令反馈**。你会学习 `package`、导出规则、import path、匿名导入、module、workspace、`internal`、build tag、`go list`、`go doc` 和常用 go 子命令。

<SourceNote
  source="gopl-zh"
  title="第 10 章：包和工具"
  note="本章主干基于本地 gopl-zh 第 10 章重组：保留包系统、导入路径、包声明、导入声明、匿名导入、工具箱、构建、文档、internal 和 go list 的核心脉络，并将 GOPATH 时代内容改写为 module 时代的工程实践。"
/>

<CompareNote
  title="本章如何整合外部优秀资料"
  points={[
    "gopl-zh 负责包和工具主线：包命名空间、封装、import path、导入声明、工具箱、文档、internal 和 go list。",
    "Go Modules Reference 用于更新 module、go.mod、go.sum、workspace 和 toolchain 的现代协作语境。",
    "Effective Go 和 Go Code Review Comments 的命名建议被转化为包名、导出 API、util/common 反模式和文档注释检查清单。",
    "go command 官方文档被整合为日常反馈回路：fmt、test、vet、build、run、env、doc、list 各自解决不同问题。"
  ]}
/>

## 你会解决什么问题

我们先从一个最小工具链观察开始：程序输出当前编译目标平台。

```text
goos=linux goarch=amd64
```

不同 sandbox 或本地机器可能输出不同值，所以练习只要求包含 `goos=`。这件小事会牵出本章大多数知识：

| 任务 | 需要的包和工具能力 | 工程关注点 |
|---|---|---|
| 组织源文件 | `package` 声明 | 同一目录通常是同一包 |
| 引用标准库和内部库 | import path | 路径是依赖身份，不只是文件路径 |
| 设计公开 API | 导出规则和文档注释 | 大写名字会成为包外契约 |
| 统一本地反馈 | `go fmt`、`go test`、`go vet`、`go build` | CI 应能复现本地命令 |
| 管理依赖版本 | `go.mod`、`go.sum`、`go work` | module 是现代协作边界 |
| 限制内部实现 | `internal` | 不稳定 API 不应被外部依赖 |
| 查询包图 | `go list` | 脚本应理解 Go 构建规则 |

<PracticeBridge
  exercise="ch10-toolchain-env"
  text="先完成热身练习：读取 runtime.GOOS 和 runtime.GOARCH，观察构建目标如何影响程序行为。"
/>

## Package：编译、命名空间和封装单元

每个 Go 文件都以 package 声明开始：

```go
package billing
```

同一目录下的 `.go` 文件通常属于同一个包。编译时，一个包作为整体编译；包内文件共享同一个命名空间，因此可以互相访问未导出的名字。

包的第一作用是命名空间：不同包可以各自定义 `Client`、`Config`、`Server`，调用方通过包名区分：

```go
http.Client
sql.DB
json.Decoder
```

包的第二作用是封装。小写开头的名字只在包内可见，大写开头的名字会导出给其他包。这个规则很简单，但影响很大：导出的类型、函数、方法和字段都会成为调用方依赖的 API。

```go
type Config struct {
    Endpoint string // 导出字段，包外可读写
    token    string // 未导出字段，只能包内维护
}
```

<PitfallCard
  title="过早导出内部字段"
  symptom="调用方开始直接修改字段，包内无法再保证配置合法或状态一致。"
  fix="先用未导出字段和构造函数保护不变量；确实稳定且需要包外设置时再导出。"
/>

## 包名和调用点要一起读

包名应该短小、清楚、稳定。调用点应该自然：

```go
strings.NewReader
json.Marshal
http.Get
flag.String
```

因此成员名里通常不重复包名。不要写成 `json.JSONMarshal` 或 `strings.StringReader`，包名已经提供上下文。

`util`、`common`、`helper` 这类包名很容易变成垃圾桶：什么都能放，最后所有包都依赖它。更好的名字应该表达领域或能力，例如 `authz`、`billing`、`retry`、`clock`、`cache`。

<ExamplePair
  title="包命名：垃圾桶 vs 职责边界"
  leftTitle="边界越来越模糊"
  rightTitle="调用点表达能力"
  left={`package util

func Retry(...)
func Now(...)
func CheckPermission(...)`}
  right={`retry.Do(...)
clock.Now(...)
authz.Check(...)`}
/>

## Import path：依赖身份，不只是路径字符串

`import` 语句里写的是导入路径：

```go
import (
    "encoding/json"
    "fmt"

    "github.com/go-sql-driver/mysql"
    "golang.org/x/net/html"
)
```

标准库路径短小稳定。第三方包通常使用域名或托管平台路径作为前缀，保证全局唯一，也方便版本管理。

对库作者来说，import path 是 API 的一部分。调用方写下 `example.com/acme/billing` 后，这个路径就进入代码、文档、构建缓存和依赖图。随意改模块路径或包路径，几乎等于要求所有调用方改代码。

module 时代，`go.mod` 让依赖版本更明确，但 import path 的身份意义没有消失。

<PracticeBridge
  exercise="ch10-import-summary"
  text="核心练习会让你把 import path 分成标准库和外部依赖，并按稳定顺序输出，模拟工具脚本处理依赖清单。"
/>

## Package declaration：目录名、包名和 main 包

通常包名是导入路径最后一段：

```go
import "encoding/json"

json.Marshal(v)
```

但也有例外：

1. 可执行程序使用 `package main`，导入路径最后一段不决定调用名；
2. 外部测试包可以使用 `package xxx_test`，让测试从包外视角验证导出 API；
3. 某些路径带版本后缀，但包名不带后缀，例如旧生态里的 `yaml.v2` 通常包名仍是 `yaml`。

工程上应尽量让目录名、包名和职责一致。频繁依赖 import alias 才能读懂的包，通常说明命名或边界需要调整。

## Import declaration：分组、别名和匿名导入

多个 import 通常用分组形式：

```go
import (
    "crypto/rand"
    "encoding/hex"
    "fmt"

    mrand "math/rand"
)
```

空行通常分隔标准库、第三方包和内部包。`gofmt` 会排序每组内部顺序，`goimports` 还能自动增删 import。

当两个包默认包名冲突时，可以给其中一个起别名：

```go
import (
    "crypto/rand"
    mrand "math/rand"
)
```

别名只影响当前文件，不改变包自身名称。不要滥用别名；如果必须使用，名字应帮助阅读，而不是制造新的缩写谜题。

匿名导入使用 `_`：

```go
import _ "image/png" // register PNG decoder
```

它表示导入这个包只是为了执行初始化副作用。常见场景包括图片解码器、数据库驱动、插件注册。匿名导入要克制，并用注释说明注册目的；如果一个包导入后会连接网络、读取生产配置或启动 goroutine，那通常不是好设计。

<PitfallCard
  title="匿名导入隐藏副作用"
  symptom="删除一行看似未使用的 import 后，图片解码或数据库驱动突然失效。"
  fix="只在注册型场景使用 `_` import，并在同一行注释说明副作用。"
/>

## Module：现代 Go 的版本和依赖边界

现代 Go 项目通常使用 module：

```bash
go mod init example.com/acme/service
go mod tidy
go test ./...
```

`go.mod` 描述模块路径、Go 版本和依赖版本。`go.sum` 记录依赖校验和。多数应用仓库从单 module 开始最简单；只有发布边界、版本节奏或依赖图确实需要分离时，才考虑多 module。

`go work` 适合本地同时开发多个 module。例如你同时修改一个服务和一个内部库，可以用 workspace 把它们放在一起调试。但发布和 CI 仍应验证每个 module 独立可构建，避免 workspace 掩盖版本问题。

现代工具链还支持 `toolchain` 指令表达期望 Go 工具链版本。即便如此，README、CI 镜像和本地脚本仍应明确版本，避免“我机器上能跑”。

## Go 工具箱：统一反馈回路

Go 命令是一组工具的入口：

```bash
go fmt ./...
go test ./...
go vet ./...
go build ./...
go env
go list ./...
go doc time.Since
```

常见命令职责：

| 命令 | 作用 | 典型时机 |
|---|---|---|
| `go fmt` | 格式化源代码 | 保存或提交前 |
| `go test` | 运行测试、示例、基准测试 | 本地和 CI |
| `go vet` | 检查可疑代码 | CI 基础检查 |
| `go build` | 验证包或命令可编译 | 发布前、容器构建前 |
| `go run` | 临时构建并运行命令 | 小程序和本地调试 |
| `go install` | 安装命令到 `$GOBIN` | 安装开发工具 |
| `go env` | 查看工具链环境 | 排查平台差异 |
| `go list` | 查询包元数据 | 自动化脚本和依赖分析 |
| `go doc` | 查看包、类型、函数文档 | 学习 API 和评审文档 |

团队项目应该把这些命令固化进 Makefile、脚本、Taskfile 或 CI，而不是依赖每个人记忆。最小 Go CI 通常至少包含：

```bash
go test ./...
go vet ./...
go build ./...
```

如果项目还包含前端、生成代码、数据库迁移或容器镜像，也要把对应验证命令写清楚。

## 构建、交叉编译和构建约束

`go build` 可以编译当前目录、导入路径或相对路径：

```bash
go build ./cmd/server
go build ./...
```

对于 `package main`，构建会生成可执行程序；对于库包，构建主要用于验证编译通过。

Go 交叉编译很直接：

```bash
GOOS=linux GOARCH=amd64 go build ./cmd/server
```

但“能交叉编译”不等于“跨平台行为一定正确”。涉及文件路径、系统调用、cgo、时区、证书、换行符或大小写敏感文件系统时，仍需要在目标平台验证。

某些文件只应在特定平台构建，可以用文件名后缀或 build tag：

```go
//go:build linux || darwin
```

现代代码优先使用 `//go:build`。构建约束适合隔离平台差异，不适合当普通业务功能开关滥用。

## 文档注释：导出 API 的第一份契约

Go 鼓励文档和代码放在一起。导出的包、类型、函数和方法都应该有说明，注释通常以被注释的名字开头：

```go
// ParseConfig reads configuration from r and returns a validated Config.
func ParseConfig(r io.Reader) (Config, error) {
    // ...
}
```

好的文档不复述实现，而说明用途、约束、错误行为和并发安全性。包文档较长时，通常放在 `doc.go`。

查看文档可以用：

```bash
go doc time
go doc time.Since
go doc json.Decoder.Decode
```

如果一个导出函数要求调用方先初始化、不能并发调用、必须关闭返回值或错误可用 `errors.Is` 判断，文档就应该说清楚。

## internal：给仓库内部使用的边界

Go 的导出规则只有包内和包外两级。有时你希望某些包只给仓库内部某个子树使用，而不是暴露给所有调用方。`internal` 目录提供了这种边界：

```text
service/
  internal/cache
  internal/auth
  cmd/server
```

`service/internal/cache` 只能被 `service` 目录树下面的包导入，外部项目不能直接 import。它适合放内部实现、实验性接口、不稳定工具包和基础设施适配。

不要把 `internal` 当作隐藏混乱代码的地方。它仍然应该有清晰职责，只是可见范围更小。

<PracticeBridge
  exercise="ch10-internal-rule"
  text="挑战练习会让你模拟 internal 包导入规则：只有 internal 父目录树下的包可以导入它。"
/>

## go list：让脚本理解 Go 的包图

`go list` 可以查询包路径和元数据：

```bash
go list ./...
go list -json ./cmd/server
go list -f '{{.ImportPath}} -> {{join .Imports " "}}' ./...
```

它比手写 `find` 更懂 Go 的构建规则，包括 build tag、测试文件、module、vendor 和工作区语义。

常见用途：

- 找出仓库内所有包；
- 查询某个包的直接 import；
- 区分普通源文件、内部测试和外部测试；
- 为 lint、生成代码或 CI 分组提供输入；
- 分析依赖图和循环依赖苗头。

工具脚本应该优先使用 `go list` 这类结构化信息，而不是自己猜目录结构。

## 包和工具概念回看

| 概念 | 最小定义 | 工程判断 |
|---|---|---|
| package | 编译、命名空间和封装单元 | 职责是否清晰 |
| 导出名字 | 大写开头，包外可见 | 是否已经成为 API 契约 |
| 包名 | 调用时使用的限定名 | 与成员名一起读是否自然 |
| import path | 依赖身份 | 路径是否稳定 |
| import alias | 当前文件内重命名包 | 是否只为冲突或清晰性使用 |
| 匿名导入 | 只执行初始化副作用 | 是否有注册目的注释 |
| module | 版本和依赖边界 | go.mod、go.sum 是否进入 CI |
| go work | 本地多 module 联调 | 是否避免掩盖发布问题 |
| toolchain | 期望工具链版本 | README/CI/容器是否一致 |
| build tag | 按平台或条件选择文件 | 是否只用于构建边界 |
| internal | 限制导入范围 | 内部 API 是否避免外泄 |
| go list | 查询包元数据 | 脚本是否理解 Go 构建规则 |

## 本章工程视角

包和工具是 Go 工程的骨架。包边界决定依赖方向，import path 决定依赖身份，文档注解决定调用方能否正确使用 API，工具链决定团队反馈是否一致。

评审 Go 项目结构时可以用这组问题自查：

- 包名、目录名和职责是否匹配？
- 导出的 API 是否稳定，文档是否说明约束和错误行为？
- 是否出现 util/common/helper 垃圾桶包？
- import 关系是否保持单向，是否有循环依赖苗头？
- 匿名导入是否有明确注册目的？
- internal 是否用于保护不稳定实现，而不是掩盖混乱？
- 本地脚本和 CI 是否运行同一组 go 命令？
- 构建问题能否通过 go env、go list、go test、go build 复现？

能回答这些问题，包和工具就不只是目录和命令，而是让团队协作、重构和发布变得可控的工程系统。
```

## DeepTutor 最终提示模板

你是 GoGopher Arch 的离线课程内容研究员。请基于上面的输入包，为 `ch10-packages-tools` 生成一版可审校的教程级 MDX 正文草稿，并附完整审计材料。

请遵守：

1. 可以开放网页检索，但外部资料只能作为知识来源层，不能形成外链合集，也不能拼贴外部教程。
2. 保持 GoGopher Arch 风格：场景引入、基础概念、最小示例、工程示例、常见坑、工程实践、概念回看、练习衔接。
3. 不默认修改 metadata、exercise ID、expected output 或路由。
4. 保留与 `ch10-toolchain-env`、`ch10-import-summary`、`ch10-internal-rule` 的衔接。
5. 输出必须分成两部分：`MDX_DRAFT` 和 `AUDIT_REPORT`。
6. `AUDIT_REPORT` 必须包含来源清单、段落级来源映射、搬运风险、与原章节差异、不可验证声明、人工审校 checklist 和最终建议。
7. 如果资料之间存在技术口径冲突，请不要自行掩盖，必须列入审计报告。
