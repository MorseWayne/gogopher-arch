# GoGopher Arch: Go 后端实习成长平台

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/Frontend-React-61DAFB?style=flat&logo=react)](https://reactjs.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

**GoGopher Arch** 是一个面向 Go 学习者的实战成长平台。它通过虚拟职场任务，帮助用户从 Go 基础入门，成长到具备 Go 后端实习岗位能力，并进一步进阶为掌握 Go 技术栈与 AI 应用工程能力的新型全栈程序员。

---

## 核心定位

GoGopher Arch 不把学习过程设计成单纯的课程目录，而是把知识点放进真实工作任务里：

- 初学者先完成入职前训练营，补齐 Go 基础。
- 准实习生和实习生通过任务卡练习修 Bug、补接口、写测试、读日志和处理评审意见。
- 有经验的 Go 工程师继续进入数据库、缓存、并发、部署、可观测性等工程能力训练。
- 进阶用户可以沿着 RAG、Agent、LLM 应用工程和 AI 产品评测路线，转向 AI 时代的新型全栈开发。

---

## 核心特性

- **实习生任务线**：以虚拟 Go 后端团队的入职第一周为主线，围绕任务卡、验收标准和导师反馈推进学习。
- **任务前小课**：每个任务前只讲完成当前任务必须用到的 Go 知识，降低上手挫败感。
- **交互式沙盒**：在浏览器中编写 Go 代码，运行程序或测试，立即看到输出、错误和任务反馈。
- **任务后复盘**：完成任务后沉淀知识点、真实工作场景、常见坑和面试追问。
- **成长路线图**：从 Go 基础、后端实习、工程进阶一路延伸到 RAG、Agent 和 AI 应用工程。

---

## 技术栈

| 模块 | 技术实现 |
| :--- | :--- |
| 后端 | Go 1.22+, Gateway, Sandbox Engine |
| 前端 | React, TypeScript, Monaco Editor |
| 沙盒 | Docker, `os/exec`, 执行超时控制 |
| 反馈 | 编译结果、控制台输出、任务检查、导师提示 |
| AI 路线 | LLM API、RAG、Agent、结构化输出、评测与安全 |

---

## 路线图

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

### 第二阶段：Go 工程能力进阶

- [x] 数据库和事务任务
- [x] 缓存和并发任务
- [x] 日志、配置和可观测性任务
- [x] 部署和服务可靠性任务

### 第三阶段：AI 时代全栈工程路线

- [ ] LLM API 调用和 Prompt 设计
- [ ] 结构化输出和工具调用
- [ ] RAG：文档切分、Embedding、向量检索和重排
- [ ] Agent：规划、工具使用、记忆、上下文管理和评估
- [ ] AI 产品的成本控制、安全边界和评测集

---

## 快速开始

```bash
git clone https://github.com/MorseWayne/gogopher-arch.git
cd gogopher-arch
docker compose up --build
```

前端默认运行在 `http://localhost:3000`，Gateway 默认运行在 `http://localhost:8080`。

---

## 开源协议

本项目采用 [MIT License](LICENSE) 协议。
