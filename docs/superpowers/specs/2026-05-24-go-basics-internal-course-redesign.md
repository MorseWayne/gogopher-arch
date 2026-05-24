# Go 基础训练营完整内置课程重制设计

## 背景

Go 基础训练营已经作为独立课程入口接入 GoGopher Arch，并提供 13 章路径、章节详情页和 sandbox 练习。当前版本仍将《Go 语言圣经中文版》作为明显来源入口：章节页展示原文链接，总览页展示外部项目和授权按钮。

用户希望下一阶段不再把课程做成外部链接导向的教程页，而是由 GoGopher Arch 直接整理、重新生成并内置完整知识点内容。课程需要结合 Go 1.24+ 的最新特性和当前 Go 后端生态现状，形成自洽的项目内课程；外部教程只作为历史参考/灵感来源保留在 README 中。

## 目标

- 将 Go 基础训练营升级为完整内置课程。
- 以 Go 1.24+ 为课程基线，补充现代 Go 和当前生态实践。
- 保留现有 13 章路径和每章 sandbox 练习。
- 每章新增完整知识点正文、现代 Go 说明、工程实践、常见坑和复盘问题。
- 章节页不再展示外部教程链接或“阅读原文”入口。
- README 保留来源/灵感说明，但不把外部教程作为课程依赖入口。

## 非目标

- 不新增 30-40 个小节路由。
- 不做进度持久化。
- 不做复杂在线编辑器。
- 不做 AI 批改。
- 不改 sandbox API。
- 不改现有实习任务线。
- 不复制外部教程正文。

## 内容策略

课程内容由 GoGopher Arch 重制生成，而不是外部教程的摘要或镜像。

每个知识点都应回答：

- 它解决什么问题？
- 在 Go 后端开发中什么时候会遇到？
- 现代 Go 推荐怎么写？
- 常见错误是什么？
- 如何用练习验证？

内容风格面向 Go 后端学习者，不写成语言规范翻译，也不写成百科全书式罗列。

## 来源与版权边界

课程页面不展示外部教程链接，也不把外部教程作为必读材料。README 必须保留清晰的来源边界说明，避免用户误解为外部教程镜像。

README 使用以下说明模板：

```md
## 课程来源与版权边界

Go 基础训练营是 GoGopher Arch 面向 Go 后端学习者重新制作的内置课程，内容以 Go 1.24+ 和当前后端工程实践为基线组织。课程知识点、讲解、练习、验收标准和复盘问题由本项目重新整理生成，不复制或镜像外部教程正文。

本课程的历史参考和灵感来源包括《Go 语言圣经中文版》项目 gopl-zh/gopl-zh.github.com。该项目的原始内容和代码遵循其自身授权说明；本项目仅在 README 中提供来源说明，不在课程页面中提供外部教程跳转作为学习依赖。
```

README 可以保留原项目仓库链接和授权说明链接，但这些链接只出现在 README 的来源说明区，不出现在课程总览页或章节详情页。

## Go 版本基线

课程基线为 **Go 1.24+**。

课程正文应结合现代 Go 实践，包括但不限于：

- Go Modules、workspace 和 toolchain 管理。
- Go 1.22+ range 变量语义变化。
- 泛型的实际使用边界。
- `errors.Is`、`errors.As`、`errors.Join`。
- `context` 取消、超时和请求生命周期。
- `log/slog`、结构化日志和可观测性基础。
- `testing` 的现代能力：`t.Cleanup`、`t.TempDir`、`t.Setenv`、fuzz testing、parallel tests。
- `sync/atomic` 类型、`sync.Map` 使用边界和 race detector。
- 反射与泛型的边界。
- `unsafe`、pprof、benchmark 和性能证据优先原则。

练习代码应尽量保持 sandbox 可运行，不强依赖当前 sandbox 可能尚未支持的过新语法。涉及版本差异时，在 `modernNotes` 中说明。

## 课程结构

保留 13 章路径：

1. 入门
2. 程序结构
3. 基础数据类型
4. 复合数据类型
5. 函数
6. 方法
7. 接口
8. Goroutines 和 Channels
9. 基于共享变量的并发
10. 包和工具
11. 测试
12. 反射
13. 底层编程

每章内容量建议：

- 3-6 个 `lessons`
- 每个 lesson 2-4 段正文
- 每章 0-2 个短代码示例
- 2-4 条 `modernNotes`
- 4-6 条 `engineeringPractices`
- 3-5 个 `pitfalls`
- 3-5 个 `reviewQuestions`
- 1 个 sandbox 练习

## 数据模型

在现有 `GoCourseChapter` 基础上升级字段。

继续保留：

- `slug`
- `order`
- `title`
- `duration`
- `difficulty`
- `summary`
- `goals`
- `exercise`
- `checklist`
- `nextMissionSlugs`

停止在页面中使用，并应从课程数据模型中删除：

- `sourcePath`
- `sourceUrl`

新增字段：

```ts
type GoCourseLesson = {
  title: string;
  body: string[];
  example?: string;
};

type GoCourseModernNote = {
  title: string;
  body: string;
};

type GoCoursePitfall = {
  title: string;
  body: string;
};

type GoCourseChapter = {
  slug: string;
  order: number;
  title: string;
  duration: string;
  difficulty: "入门" | "基础" | "进阶" | "高级";
  summary: string;
  goals: string[];
  lessons: GoCourseLesson[];
  modernNotes: GoCourseModernNote[];
  engineeringPractices: string[];
  pitfalls: GoCoursePitfall[];
  exercise: GoCourseExercise;
  checklist: string[];
  reviewQuestions: string[];
  nextMissionSlugs: string[];
};
```

`validateGoBasicsCourse` 应更新为检查：

- 章节数量为 13。
- `slug` 唯一。
- 每章有 `summary`、`goals`、`lessons`、`modernNotes`、`engineeringPractices`、`pitfalls`、`reviewQuestions`、`exercise` 和 `checklist`。
- 每章至少 3 个 lesson。
- 每个 lesson 至少 2 段正文；正文段落不能为空。
- 每章至少 2 条 modern note。
- 每章至少 3 条 engineering practice。
- 每章至少 3 个 pitfall。
- 每章至少 3 个 review question。
- 每章有 1 个 sandbox 练习，且 `starterCode`、`expectedOutput` 和 `outputMatch` 存在。
- `nextMissionSlugs` 不生成死链接。
- 不再要求 `sourceUrl`，且课程数据源不再包含 `sourcePath` 或 `sourceUrl` 字段。
- `sourceUrl`、`sourcePath`、`gopl-zh.github.io` 不应出现在课程页面渲染依赖中。

## 页面设计

### 课程总览页

更新 `web/src/app/pages/GoBasicsCourse.tsx`：

- 文案改为“GoGopher Arch 内置课程，结合 Go 1.24+ 和后端工程现状重制”。
- 移除“查看原项目”“LICENSE”“译文授权”按钮。
- 移除总览页外部来源卡片。
- 保留 13 章课程卡片。
- 可在 Hero 或说明区展示：基线 Go 1.24+、内置课程、每章 sandbox 练习、衔接实习任务。

### 章节详情页

更新 `web/src/app/pages/GoBasicsChapter.tsx`：

移除：

- 顶部“阅读原文”链接。
- “打开原教程章节”按钮。
- 右侧“来源”卡片。
- 页面主体中的“参考并改编自……”文案。
- 对 `sourcePath` 和 `sourceUrl` 的渲染依赖。

保留：

- 章节头部。
- 上一章 / 下一章导航。
- 学习目标。
- sandbox 练习。
- 验收 checklist。
- 衔接实习任务。
- 未找到章节状态。

新增：

- `lessons`：完整课程正文主区域。
- `modernNotes`：Go 1.24+ / 当前生态现状卡片。
- `engineeringPractices`：后端工程实践建议。
- `pitfalls`：常见坑独立区域。
- `reviewQuestions`：复盘问题区域。

## 13 章重制重点

### 1. 入门

覆盖 Go 程序结构、`go run`、模块初始化、命令行程序、最小 HTTP 服务。现代重点是 Go Modules 已是默认工作流，不再把 GOPATH 作为主路径。

### 2. 程序结构

覆盖声明、变量、常量、作用域、包级状态、命名。现代重点是 `go.mod`、`toolchain`、包边界和小包设计。

### 3. 基础数据类型

覆盖数值、字符串、rune、byte、布尔、常量。现代重点是 UTF-8 文本处理、配置值、日志字段和避免过度类型转换。

### 4. 复合数据类型

覆盖 array、slice、map、struct、JSON。现代重点是 slice 内存保留、map 并发安全、JSON tag、`encoding/json` 的边界。

### 5. 函数

覆盖函数签名、多返回值、错误处理、defer、panic/recover。现代重点是 `errors.Is`、`errors.As`、`errors.Join`、context 传递和清晰错误边界。

### 6. 方法

覆盖接收者、值/指针方法、嵌入、封装。现代重点是 API 设计、不可变倾向、避免贫血模型或过度 OOP。

### 7. 接口

覆盖小接口、隐式实现、接口值、类型断言。现代重点是接口由调用方定义、泛型出现后接口的使用边界、`io.Reader` 风格组合。

### 8. Goroutines 和 Channels

覆盖 goroutine、channel、select、超时、退出控制。现代重点是 context cancellation、worker pool、不要用 channel 解决所有同步问题。

### 9. 基于共享变量的并发

覆盖 race、mutex、RWMutex、sync.Once、atomic、race detector。现代重点是 `sync/atomic` 类型、`sync.Map` 使用边界、并发安全 API 设计。

### 10. 包和工具

覆盖包组织、导入、gofmt、go test、go list、go env。现代重点是 modules、workspaces、toolchain、`go vet`、`go mod tidy`、CI 中的工具链管理。

### 11. 测试

覆盖单元测试、表驱动测试、覆盖率、benchmark、example。现代重点是 fuzz testing、`t.Cleanup`、`t.TempDir`、`t.Setenv`、并行测试和测试可维护性。

### 12. 反射

覆盖 `reflect.Type`、`reflect.Value`、tag、可设置性。现代重点是反射在 JSON/ORM/DI 中的地位，泛型出现后什么时候可以不用反射。

### 13. 底层编程

覆盖 `unsafe`、内存布局、对齐、cgo。现代重点是何时远离 unsafe、性能优化证据、pprof/benchmark 先于底层技巧。

## 实现策略

一次性完成 13 章内容重制，但按内部顺序实施：

1. 模型升级
   - 更新类型和校验函数。
   - 页面兼容新字段。
2. 内容重制
   - 逐章生成完整内置内容。
   - 保证每章结构一致。
   - 控制每章内容可读，不写成巨型论文。
3. 页面升级
   - 章节页改为完整课程页。
   - 总览页移除外部链接按钮和来源卡片。
   - README 调整来源文案。
4. 验证与修正
   - 构建。
   - 浏览器验证。
   - 修复布局、类型或内容遗漏。

首版继续把 13 章课程数据放在 `web/src/app/data/goBasicsCourse.ts`。如果后续维护困难，再拆成 `web/src/app/data/goBasicsCourse/ch1.ts` 这类章节文件。本轮不做文件拆分，避免过度重构。

## 错误处理

- 未找到章节时继续展示课程级“未找到章节”状态。
- sandbox 调用失败时继续在练习区显示服务连接错误。
- 代码运行失败继续展示 stdout、stderr 和 exit code。
- 课程正文渲染不依赖外部网络，因此外部资料不可访问不影响学习体验。

## 验收标准

实现完成后应满足：

- `/courses/go-basics` 不再展示外部项目按钮或授权按钮。
- 任一章节页不再出现：
  - “阅读原文”
  - “打开原教程章节”
  - `gopl-zh.github.io`
- 所有 13 章详情页都能显示：
  - 至少 3 个完整课程正文小节
  - Go 1.24+ / 现代 Go 说明
  - 工程实践
  - 常见坑
  - 复盘问题
  - sandbox 练习
- 第 1、5、8、11、13 章作为人工抽样重点页面，需额外确认内容排版和练习反馈。
- 13 章都通过课程数据校验。
- README 保留来源说明，但表述为“历史参考/灵感来源”，不是课程依赖入口。
- `npm run build --prefix web` 通过。
- GitNexus detect changes 风险不超出预期。
- `/dashboard` 和 `/missions/slice-memory-leak` 无回归。

## 测试策略

### 静态验证

- TypeScript 构建通过。
- `validateGoBasicsCourse` 不返回错误。
- 静态文本检查确认课程页不渲染外部教程 URL。
- 静态文本检查确认 `GoBasicsCourse.tsx` 和 `GoBasicsChapter.tsx` 不包含 `gopl-zh.github.io`、`阅读原文`、`打开原教程章节`、`查看原项目`、`译文授权`。
- 静态文本检查确认课程数据中 13 章均具备 `lessons`、`modernNotes`、`engineeringPractices`、`pitfalls` 和 `reviewQuestions`。

### 浏览器验证

启动前端后验证：

- `/courses/go-basics`
- `/courses/go-basics/ch1-getting-started`
- `/courses/go-basics/ch5-functions`
- `/courses/go-basics/ch8-goroutines-channels`
- `/courses/go-basics/ch11-testing`
- `/courses/go-basics/ch13-low-level-programming`
- `/courses/go-basics/not-found`
- `/dashboard`
- `/missions/slice-memory-leak`

### 练习验证

至少点击第 1、5、8、11、13 章的运行按钮，确认练习区能展示成功、失败或服务不可用反馈。

## 风险

- 内容体量较大，单文件会变长。首版接受单文件，后续再评估拆分。
- Go 1.24+ 内容必须与 sandbox 运行能力区分：课程可以讲现代特性，练习避免强依赖 sandbox 尚未支持的语法。
- 移除章节外链后，README 的来源说明必须足够清楚，避免误导用户认为课程复制自外部教程。
