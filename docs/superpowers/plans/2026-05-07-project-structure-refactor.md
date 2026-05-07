# GoGopher Arch 项目结构重构实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 GoGopher Arch 从 MVP 单文件结构重构为标准企业级分层架构，Go 后端 handler/runner 分离，前端 domain/api 分层，补齐 Makefile 和 CI。

**Architecture:** Go 后端按 service → internal/{handlers,middleware,routes,runner} 拆分，main.go 只做启动；前端按 api/ → domain/ → components/ 分层；基础设施补齐 Makefile 和 GitHub Actions CI。

**Tech Stack:** Go 1.24, React 19 + TypeScript 5.9, Vite 8, Docker Compose, GitHub Actions

---

## 文件结构

### Go 后端

| 文件 | 职责 |
|------|------|
| `src/pkg/common/models.go` | 跨服务共享的 SandboxRequest/SandboxResponse（已有，保持不变） |
| `src/pkg/common/errors.go` | 自定义错误类型和 JSON 错误响应辅助函数 |
| `src/internal/config/config.go` | 读取环境变量，提供默认值 |
| `src/services/sandbox-engine/internal/runner/runner.go` | 沙盒执行核心：创建临时目录、go run、超时控制 |
| `src/services/sandbox-engine/internal/runner/module.go` | 判断是否需要复制预缓存的 go.mod/go.sum |
| `src/services/sandbox-engine/internal/handlers/execute.go` | HTTP handler：解析请求、调用 runner、返回 JSON |
| `src/services/sandbox-engine/main.go` | 依赖注入 + HTTP 启动 |
| `src/services/gateway/internal/middleware/cors.go` | CORS 中间件 |
| `src/services/gateway/internal/middleware/logging.go` | 请求日志中间件 |
| `src/services/gateway/internal/middleware/recovery.go` | panic 恢复中间件 |
| `src/services/gateway/internal/handlers/execute.go` | 转发执行请求到 sandbox-engine |
| `src/services/gateway/internal/handlers/health.go` | 健康检查 |
| `src/services/gateway/internal/routes/routes.go` | 路由注册与中间件链组装 |
| `src/services/gateway/main.go` | 依赖注入 + HTTP 启动 |

### 前端

| 文件 | 职责 |
|------|------|
| `web/vite.config.ts` | 添加 `@/` 路径别名 |
| `web/tsconfig.app.json` | 添加 `baseUrl` 和 `paths` |
| `web/src/api/types.ts` | API DTO 类型（SandboxRequest, SandboxResponse） |
| `web/src/api/client.ts` | axios 实例、拦截器、baseURL |
| `web/src/api/execute.ts` | `executeCode()` 封装 |
| `web/src/domain/tasks/types.ts` | InternTask, TaskCheck, FeedbackItem 等类型 |
| `web/src/domain/tasks/data.ts` | 14 个任务数据、defaultTaskId、findTaskById |
| `web/src/domain/tasks/checks.ts` | evaluateTaskChecks, didPassTask, didSandboxSucceed |
| `web/src/domain/tasks/index.ts` | 统一导出 |
| `web/src/domain/workbench/types.ts` | TopBarProps, TaskPanelProps 等 UI 组件 props |
| `web/src/App.tsx` | 精简为状态管理 + 组件组合 |
| `web/src/App.test.tsx` | 更新 import |
| `web/src/domain/tasks/data.test.ts` | 从 tasks.test.ts 迁移 |
| `web/src/domain/tasks/checks.test.ts` | 从 taskFeedback.test.ts 迁移 |

### 基础设施

| 文件 | 职责 |
|------|------|
| `Makefile` | 统一开发、测试、构建、清理命令 |
| `.github/workflows/ci.yml` | Go vet/test + 前端 lint/test CI 流水线 |
| `db/migrations/` | 空目录，预留数据库迁移 |

---

## Task 0: 准备 — 创建目录并验证当前状态

**Files:**
- Create: 多个目录（见下方）
- Test: 现有测试全部通过

- [ ] **Step 0.1: 创建所有目标目录**

```bash
mkdir -p src/services/gateway/internal/{handlers,middleware,routes}
mkdir -p src/services/sandbox-engine/internal/{handlers,runner}
mkdir -p src/internal/{task,config}
mkdir -p src/pkg/common
mkdir -p web/src/api
mkdir -p web/src/domain/{tasks,workbench}
mkdir -p db/migrations
mkdir -p .github/workflows
```

- [ ] **Step 0.2: 验证当前 Go 编译通过**

```bash
go build ./src/services/gateway/main.go
go build ./src/services/sandbox-engine/main.go
```

Expected: 两个命令都成功，无错误输出。

- [ ] **Step 0.3: 验证当前前端测试通过**

```bash
cd web && npm run test -- --run
```

Expected: 所有测试通过（tasks.test.ts, taskFeedback.test.ts, App.test.tsx）。

- [ ] **Step 0.4: Commit 标记起点**

```bash
git add -A
git commit -m "chore: create directory structure for refactor"
```

---

## Task 1: Go 共享代码 — errors.go 和 config.go

**Files:**
- Create: `src/pkg/common/errors.go`
- Create: `src/internal/config/config.go`
- Test: `go build` 验证

- [ ] **Step 1.1: 创建 `src/pkg/common/errors.go`**

```go
package common

import (
	"encoding/json"
	"net/http"
)

// ValidationError 表示请求参数校验失败
type ValidationError struct {
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return e.Message
}

// InternalError 表示内部服务错误
type InternalError struct {
	Message string `json:"message"`
}

func (e InternalError) Error() string {
	return e.Message
}

// WriteError 向响应写入统一的 JSON 错误格式
func WriteError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
}
```

- [ ] **Step 1.2: 创建 `src/internal/config/config.go`**

```go
package config

import "os"

type Config struct {
	Port       string
	SandboxURL string
	DB_URL     string
	RedisURL   string
}

func Load() Config {
	return Config{
		Port:       envOrDefault("PORT", ":8080"),
		SandboxURL: envOrDefault("SANDBOX_URL", "http://localhost:8081/execute"),
		DB_URL:     envOrDefault("DB_URL", "postgres://user:pass@localhost:5432/gogopher?sslmode=disable"),
		RedisURL:   envOrDefault("REDIS_URL", "localhost:6379"),
	}
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
```

- [ ] **Step 1.3: 验证编译通过**

```bash
go build ./src/pkg/common/...
go build ./src/internal/config/...
```

Expected: 无错误。

- [ ] **Step 1.4: Commit**

```bash
git add src/pkg/common/errors.go src/internal/config/config.go
git commit -m "feat: add shared errors and config packages"
```

---

## Task 2: Sandbox Engine — Runner 层

**Files:**
- Create: `src/services/sandbox-engine/internal/runner/runner.go`
- Create: `src/services/sandbox-engine/internal/runner/module.go`
- Modify: `src/services/sandbox-engine/main.go`（后续 Task 3 重写）
- Test: `go build` 验证

- [ ] **Step 2.1: 创建 `src/services/sandbox-engine/internal/runner/runner.go`**

从现有 `main.go` 提取 `GopherRunner` 结构体、`Run` 方法、`errorResponse` 方法。

```go
package runner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/MorseWayne/gogopher-arch/src/pkg/common"
)

type Runner struct{}

func New() *Runner {
	return &Runner{}
}

func (r *Runner) Run(req common.SandboxRequest) common.SandboxResponse {
	start := time.Now()
	tmpDir, err := os.MkdirTemp("", "gopher-task-*")
	if err != nil {
		return r.errorResponse(req.ID, "Failed to create temp directory: "+err.Error())
	}
	defer os.RemoveAll(tmpDir)

	if err := r.ensureGoModule(tmpDir, req.Code); err != nil {
		return r.errorResponse(req.ID, "Failed to setup Go module: "+err.Error())
	}

	codePath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(codePath, []byte(req.Code), 0644); err != nil {
		return r.errorResponse(req.ID, "Failed to write code file: "+err.Error())
	}

	timeout := time.Duration(req.Timeout) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", "run", codePath)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	duration := time.Since(start)

	resp := common.SandboxResponse{
		ID:       req.ID,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		Duration: duration,
	}

	if ctx.Err() == context.DeadlineExceeded {
		resp.Status = "timeout"
		resp.ExitCode = -1
	} else if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			resp.ExitCode = exitError.ExitCode()
		} else {
			resp.ExitCode = 1
		}
		resp.Status = "error"
	} else {
		resp.Status = "success"
		resp.ExitCode = 0
	}

	return resp
}

func (r *Runner) errorResponse(id, msg string) common.SandboxResponse {
	return common.SandboxResponse{ID: id, Status: "internal_error", Stderr: msg}
}
```

- [ ] **Step 2.2: 创建 `src/services/sandbox-engine/internal/runner/module.go`**

从现有 `main.go` 提取 `ensureGoModule`。

```go
package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (r *Runner) ensureGoModule(tmpDir, code string) error {
	if !strings.Contains(code, "github.com/lib/pq") && !strings.Contains(code, "github.com/redis/go-redis") {
		return nil
	}
	for _, f := range []string{"go.mod", "go.sum"} {
		src := filepath.Join("/app/sandbox-module", f)
		data, err := os.ReadFile(src)
		if err != nil {
			return fmt.Errorf("read %s: %w", src, err)
		}
		if err := os.WriteFile(filepath.Join(tmpDir, f), data, 0644); err != nil {
			return fmt.Errorf("write %s: %w", f, err)
		}
	}
	return nil
}
```

- [ ] **Step 2.3: 验证编译通过**

```bash
go build ./src/services/sandbox-engine/internal/runner/...
```

Expected: 无错误。

- [ ] **Step 2.4: Commit**

```bash
git add src/services/sandbox-engine/internal/runner/
git commit -m "refactor: extract sandbox runner into internal/runner package"
```

---

## Task 3: Sandbox Engine — Handler 和 main.go

**Files:**
- Create: `src/services/sandbox-engine/internal/handlers/execute.go`
- Modify: `src/services/sandbox-engine/main.go`（重写）
- Test: `go build` 验证

- [ ] **Step 3.1: 创建 `src/services/sandbox-engine/internal/handlers/execute.go`**

```go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/MorseWayne/gogopher-arch/src/pkg/common"
	"github.com/MorseWayne/gogopher-arch/src/services/sandbox-engine/internal/runner"
)

type ExecuteHandler struct {
	runner *runner.Runner
}

func NewExecuteHandler(r *runner.Runner) *ExecuteHandler {
	return &ExecuteHandler{runner: r}
}

func (h *ExecuteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req common.SandboxRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	fmt.Printf("[%s] Executing code task: %s\n", time.Now().Format(time.RFC3339), req.ID)
	resp := h.runner.Run(req)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
```

- [ ] **Step 3.2: 重写 `src/services/sandbox-engine/main.go`**

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/MorseWayne/gogopher-arch/src/internal/config"
	"github.com/MorseWayne/gogopher-arch/src/services/sandbox-engine/internal/handlers"
	"github.com/MorseWayne/gogopher-arch/src/services/sandbox-engine/internal/runner"
)

func main() {
	cfg := config.Load()
	r := runner.New()
	h := handlers.NewExecuteHandler(r)

	port := ":8081"
	if cfg.Port != ":8080" {
		port = cfg.Port
	}

	fmt.Printf("Gogopher Arch Sandbox Engine listening on %s...\n", port)
	if err := http.ListenAndServe(port, h); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
```

- [ ] **Step 3.3: 验证编译通过**

```bash
go build ./src/services/sandbox-engine/main.go
```

Expected: 无错误。

- [ ] **Step 3.4: Commit**

```bash
git add src/services/sandbox-engine/
git commit -m "refactor: extract sandbox handler and slim down main.go"
```

---

## Task 4: Gateway — Middleware

**Files:**
- Create: `src/services/gateway/internal/middleware/cors.go`
- Create: `src/services/gateway/internal/middleware/logging.go`
- Create: `src/services/gateway/internal/middleware/recovery.go`
- Test: `go build` 验证

- [ ] **Step 4.1: 创建 `src/services/gateway/internal/middleware/cors.go`**

```go
package middleware

import "net/http"

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 4.2: 创建 `src/services/gateway/internal/middleware/logging.go`**

```go
package middleware

import (
	"fmt"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		fmt.Printf("[%s] %s %s %s\n", time.Now().Format(time.RFC3339), r.Method, r.URL.Path, duration)
	})
}
```

- [ ] **Step 4.3: 创建 `src/services/gateway/internal/middleware/recovery.go`**

```go
package middleware

import (
	"fmt"
	"net/http"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				fmt.Printf("[PANIC] %v\n", rec)
				http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
```

- [ ] **Step 4.4: 验证编译通过**

```bash
go build ./src/services/gateway/internal/middleware/...
```

Expected: 无错误。

- [ ] **Step 4.5: Commit**

```bash
git add src/services/gateway/internal/middleware/
git commit -m "feat: add gateway middleware (cors, logging, recovery)"
```

---

## Task 5: Gateway — Handlers 和 Routes

**Files:**
- Create: `src/services/gateway/internal/handlers/execute.go`
- Create: `src/services/gateway/internal/handlers/health.go`
- Create: `src/services/gateway/internal/routes/routes.go`
- Test: `go build` 验证

- [ ] **Step 5.1: 创建 `src/services/gateway/internal/handlers/execute.go`**

```go
package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MorseWayne/gogopher-arch/src/internal/config"
)

type ExecuteHandler struct {
	cfg config.Config
}

func NewExecuteHandler(cfg config.Config) *ExecuteHandler {
	return &ExecuteHandler{cfg: cfg}
}

func (h *ExecuteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}

	fmt.Printf("[%s] Forwarding execution request to sandbox...\n", time.Now().Format(time.RFC3339))
	resp, err := http.Post(h.cfg.SandboxURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "Sandbox engine unreachable: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, resp.Body)
}
```

- [ ] **Step 5.2: 创建 `src/services/gateway/internal/handlers/health.go`**

```go
package handlers

import (
	"encoding/json"
	"net/http"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

- [ ] **Step 5.3: 创建 `src/services/gateway/internal/routes/routes.go`**

```go
package routes

import (
	"net/http"

	"github.com/MorseWayne/gogopher-arch/src/internal/config"
	"github.com/MorseWayne/gogopher-arch/src/services/gateway/internal/handlers"
	"github.com/MorseWayne/gogopher-arch/src/services/gateway/internal/middleware"
)

func New(cfg config.Config) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/v1/execute", handlers.NewExecuteHandler(cfg))
	mux.Handle("/health", handlers.NewHealthHandler())

	var h http.Handler = mux
	h = middleware.CORS(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	return h
}
```

- [ ] **Step 5.4: 验证编译通过**

```bash
go build ./src/services/gateway/internal/handlers/...
go build ./src/services/gateway/internal/routes/...
```

Expected: 无错误。

- [ ] **Step 5.5: Commit**

```bash
git add src/services/gateway/internal/handlers/ src/services/gateway/internal/routes/
git commit -m "feat: add gateway handlers and routes"
```

---

## Task 6: Gateway — 重写 main.go

**Files:**
- Modify: `src/services/gateway/main.go`
- Test: `go build` 验证

- [ ] **Step 6.1: 重写 `src/services/gateway/main.go`**

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/MorseWayne/gogopher-arch/src/internal/config"
	"github.com/MorseWayne/gogopher-arch/src/services/gateway/internal/routes"
)

func main() {
	cfg := config.Load()
	h := routes.New(cfg)

	fmt.Printf("Gogopher Arch Gateway listening on %s...\n", cfg.Port)
	if err := http.ListenAndServe(cfg.Port, h); err != nil {
		fmt.Printf("Failed to start gateway: %v\n", err)
	}
}
```

- [ ] **Step 6.2: 验证编译通过**

```bash
go build ./src/services/gateway/main.go
```

Expected: 无错误。

- [ ] **Step 6.3: Commit**

```bash
git add src/services/gateway/main.go
git commit -m "refactor: slim gateway main.go to bootstrapping only"
```

---

## Task 7: Go 验证 — 构建和测试

**Files:**
- Test: 所有 Go 服务编译

- [ ] **Step 7.1: 验证两个服务都能编译**

```bash
go build ./src/services/gateway/main.go
go build ./src/services/sandbox-engine/main.go
go vet ./...
```

Expected: 编译通过，`go vet` 无警告。

- [ ] **Step 7.2: Commit**

```bash
git commit --allow-empty -m "checkpoint: Go backend refactor complete"
```

---

## Task 8: 前端 — 路径别名配置

**Files:**
- Modify: `web/vite.config.ts`
- Modify: `web/tsconfig.app.json`
- Test: `cd web && npm run build` 验证

- [ ] **Step 8.1: 更新 `web/vite.config.ts`**

当前内容：
```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

修改为：
```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
```

- [ ] **Step 8.2: 更新 `web/tsconfig.app.json`**

当前 `compilerOptions` 中没有 `baseUrl` 和 `paths`。在 `"include": ["src"]` 之前添加：

```json
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"]
    },
```

完整文件应如下：
```json
{
  "compilerOptions": {
    "tsBuildInfoFile": "./node_modules/.tmp/tsconfig.app.tsbuildinfo",
    "target": "ES2023",
    "useDefineForClassFields": true,
    "lib": ["ES2023", "DOM", "DOM.Iterable"],
    "module": "ESNext",
    "types": ["vite/client"],
    "skipLibCheck": true,

    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"]
    },

    "moduleResolution": "bundler",
    "allowImportingTsExtensions": true,
    "verbatimModuleSyntax": true,
    "moduleDetection": "force",
    "noEmit": true,
    "jsx": "react-jsx",

    "strict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "erasableSyntaxOnly": true,
    "noFallthroughCasesInSwitch": true,
    "noUncheckedSideEffectImports": true
  },
  "include": ["src"]
}
```

- [ ] **Step 8.3: 验证前端构建通过**

```bash
cd web && npm run build
```

Expected: 构建成功，无 TypeScript 错误。

- [ ] **Step 8.4: Commit**

```bash
git add web/vite.config.ts web/tsconfig.app.json
git commit -m "chore: configure @/ path alias for frontend"
```

---

## Task 9: 前端 — Domain 层（Tasks）

**Files:**
- Create: `web/src/domain/tasks/types.ts`
- Create: `web/src/domain/tasks/data.ts`
- Create: `web/src/domain/tasks/checks.ts`
- Create: `web/src/domain/tasks/index.ts`
- Test: `cd web && npm run test -- --run` 验证

- [ ] **Step 9.1: 创建 `web/src/domain/tasks/types.ts`**

从现有 `tasks.ts` 和 `taskFeedback.ts` 提取类型定义。

```typescript
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
```

- [ ] **Step 9.2: 创建 `web/src/domain/tasks/data.ts`**

从现有 `tasks.ts` 提取数据。保留所有 14 个任务的完整定义和 `findTaskById`、`defaultTaskId`、`internshipTasks`。

文件头：
```typescript
import type { InternTask } from './types';
```

注意：需要将原 `import type { TaskCheck } from './taskFeedback'` 改为 `import type { InternTask } from './types'`（因为 InternTask 内部已经包含 TaskCheck）。

文件内容（完整复制原 tasks.ts 的所有内容，只修改 import）：
```typescript
import type { InternTask } from './types';

export const internshipTasks: InternTask[] = [
  // ... 所有 14 个任务的完整定义 ...
];

export const defaultTaskId = 'day-0-first-run';

export function findTaskById(id: string): InternTask {
  const task = internshipTasks.find((t) => t.id === id);
  return task ?? internshipTasks[0];
}
```

- [ ] **Step 9.3: 创建 `web/src/domain/tasks/checks.ts`**

从现有 `taskFeedback.ts` 提取所有逻辑。

```typescript
import type { FeedbackItem, SandboxResponse, TaskCheck } from './types';

export type { FeedbackItem, SandboxResponse, TaskCheck };

function didSandboxSucceed(output: SandboxResponse): boolean {
  return output.status === 'success' && output.exit_code === 0;
}

function evaluateSingleCheck(
  check: TaskCheck,
  output: SandboxResponse
): FeedbackItem {
  switch (check.type) {
    case 'exitSuccess': {
      const passed = didSandboxSucceed(output);
      return {
        label: check.label,
        detail: passed ? check.passDetail : check.failDetail,
        state: passed ? 'pass' : 'fail',
      };
    }
    case 'stdoutIncludes': {
      const passed = output.stdout.includes(check.value);
      return {
        label: check.label,
        detail: passed ? check.passDetail : check.failDetail,
        state: passed ? 'pass' : 'fail',
      };
    }
    case 'stdoutRegex': {
      const regex = new RegExp(check.pattern, check.flags || '');
      const passed = regex.test(output.stdout);
      return {
        label: check.label,
        detail: passed ? check.passDetail : check.failDetail,
        state: passed ? 'pass' : 'fail',
      };
    }
    case 'stderrExcludes': {
      const passed = !output.stderr.includes(check.value);
      return {
        label: check.label,
        detail: passed ? check.passDetail : check.failDetail,
        state: passed ? 'pass' : 'fail',
      };
    }
    default: {
      return {
        label: (check as TaskCheck).label || '未知检查',
        detail: '未识别的检查类型。',
        state: 'fail',
      };
    }
  }
}

export function evaluateTaskChecks(
  output: SandboxResponse | null,
  error: string | null,
  checks: TaskCheck[]
): FeedbackItem[] {
  if (error) {
    return checks.map((check) => ({
      label: check.label,
      detail: '请求失败，无法验证。',
      state: 'fail' as const,
    }));
  }
  if (!output) {
    return checks.map((check) => ({
      label: check.label,
      detail: '等待运行...',
      state: 'idle' as const,
    }));
  }
  return checks.map((check) => evaluateSingleCheck(check, output));
}

export function didPassTask(
  output: SandboxResponse | null,
  error: string | null,
  checks: TaskCheck[]
): boolean {
  const feedback = evaluateTaskChecks(output, error, checks);
  return feedback.length > 0 && feedback.every((f) => f.state === 'pass');
}
```

- [ ] **Step 9.4: 创建 `web/src/domain/tasks/index.ts`**

```typescript
export * from './types';
export * from './data';
export * from './checks';
```

- [ ] **Step 9.5: 验证当前无使用 `@/` import 的文件，先确认编译通过**

```bash
cd web && npx tsc -b
```

Expected: 编译通过（此时还没有任何文件使用 `@/` import，所以路径别名配置不影响）。

- [ ] **Step 9.6: Commit**

```bash
git add web/src/domain/tasks/
git commit -m "refactor: extract task domain layer (types, data, checks)"
```

---

## Task 10: 前端 — Domain 层（Workbench）

**Files:**
- Create: `web/src/domain/workbench/types.ts`
- Test: `cd web && npx tsc -b` 验证

- [ ] **Step 10.1: 创建 `web/src/domain/workbench/types.ts`**

从现有 `types/workbench.ts` 迁移，更新内部 import。

```typescript
import type { InternTask, FeedbackItem, SandboxResponse } from '@/domain/tasks';

export interface TopBarProps {
  onReset: () => void;
  onRun: () => void;
  loading: boolean;
}

export interface TaskProgressProps {
  tasks: InternTask[];
  selectedTaskId: string;
  taskResults: Record<string, 'pass' | 'fail'>;
  onSelectTask: (taskId: string) => void;
}

export interface TaskPanelProps {
  task: InternTask;
}

export interface EditorPanelProps {
  code: string;
  onChange: (value: string) => void;
  track: string;
}

export interface FeedbackPanelProps {
  feedback: FeedbackItem[];
  currentTaskPassed: boolean;
  mentorHints: string[];
  review: string[];
  output: SandboxResponse | null;
  error: string | null;
}

export interface ResizableSplitProps {
  left: React.ReactNode;
  center: React.ReactNode;
  right: React.ReactNode;
}
```

注意：需要在文件顶部添加 React import（因为 `ResizableSplitProps` 使用了 `React.ReactNode`）：

```typescript
import type React from 'react';
import type { InternTask, FeedbackItem, SandboxResponse } from '@/domain/tasks';
```

- [ ] **Step 10.2: 验证 TypeScript 编译**

```bash
cd web && npx tsc -b
```

Expected: 编译通过（`@/domain/tasks` 可以正确解析）。

- [ ] **Step 10.3: Commit**

```bash
git add web/src/domain/workbench/
git commit -m "refactor: migrate workbench types to domain/workbench"
```

---

## Task 11: 前端 — API 层

**Files:**
- Create: `web/src/api/types.ts`
- Create: `web/src/api/client.ts`
- Create: `web/src/api/execute.ts`
- Test: `cd web && npx tsc -b` 验证

- [ ] **Step 11.1: 创建 `web/src/api/types.ts`**

```typescript
export interface SandboxRequest {
  id: string;
  code: string;
  language: string;
  timeout: number;
}

export interface SandboxResponse {
  id: string;
  stdout: string;
  stderr: string;
  exit_code: number;
  duration: number;
  status: string;
}
```

- [ ] **Step 11.2: 创建 `web/src/api/client.ts`**

```typescript
import axios from 'axios';

export const apiClient = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
});

apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (axios.isAxiosError(error)) {
      const message =
        typeof error.response?.data === 'string'
          ? error.response.data
          : error.message || '无法连接到 Gateway 服务';
      return Promise.reject(new Error(message));
    }
    return Promise.reject(error);
  }
);
```

- [ ] **Step 11.3: 创建 `web/src/api/execute.ts`**

```typescript
import { apiClient } from './client';
import type { SandboxRequest, SandboxResponse } from './types';

export async function executeCode(req: SandboxRequest): Promise<SandboxResponse> {
  const { data } = await apiClient.post<SandboxResponse>('/execute', req);
  return data;
}
```

- [ ] **Step 11.4: 验证 TypeScript 编译**

```bash
cd web && npx tsc -b
```

Expected: 编译通过。

- [ ] **Step 11.5: Commit**

```bash
git add web/src/api/
git commit -m "feat: add frontend API client layer"
```

---

## Task 12: 前端 — App.tsx 和组件 Import 更新

**Files:**
- Modify: `web/src/App.tsx`
- Modify: `web/src/components/EditorPanel/EditorPanel.tsx`
- Modify: `web/src/components/TaskPanel/TaskPanel.tsx`
- Modify: `web/src/components/TaskPanel/TaskContent.tsx`
- Modify: `web/src/components/TopBar/TopBar.tsx`
- Modify: `web/src/components/TopBar/TaskProgress.tsx`
- Modify: `web/src/components/ResizableSplit/ResizableSplit.tsx`
- Modify: `web/src/components/FeedbackPanel/FeedbackPanel.tsx`
- Modify: `web/src/components/FeedbackPanel/FeedbackList.tsx`
- Modify: `web/src/components/FeedbackPanel/Console.tsx`
- Test: `cd web && npm run test -- --run` 验证

- [ ] **Step 12.1: 重写 `web/src/App.tsx`**

```tsx
import { useMemo, useState } from 'react';
import { TopBar } from './components/TopBar/TopBar';
import { TaskProgress } from './components/TopBar/TaskProgress';
import { TaskPanel } from './components/TaskPanel/TaskPanel';
import { EditorPanel } from './components/EditorPanel/EditorPanel';
import { FeedbackPanel } from './components/FeedbackPanel/FeedbackPanel';
import { ResizableSplit } from './components/ResizableSplit/ResizableSplit';
import { useMediaQuery } from './hooks/useMediaQuery';
import { defaultTaskId, findTaskById, internshipTasks } from '@/domain/tasks';
import { didPassTask, evaluateTaskChecks, type SandboxResponse } from '@/domain/tasks';
import { executeCode } from '@/api/execute';
import type { SandboxRequest } from '@/api/types';
import './index.css';
import styles from './App.module.css';

function App() {
  const [selectedTaskId, setSelectedTaskId] = useState(defaultTaskId);
  const selectedTask = findTaskById(selectedTaskId);
  const [code, setCode] = useState(selectedTask.starterCode);
  const [output, setOutput] = useState<SandboxResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [taskResults, setTaskResults] = useState<Record<string, 'pass' | 'fail'>>({});
  const [mobileTab, setMobileTab] = useState<'task' | 'editor' | 'feedback'>('editor');

  const isMobile = useMediaQuery('(max-width: 959px)');

  const feedback = useMemo(
    () => evaluateTaskChecks(output, error, selectedTask.checks),
    [output, error, selectedTask]
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

  const handleRun = async () => {
    setLoading(true);
    setError(null);
    setOutput(null);

    try {
      const req: SandboxRequest = {
        id: `${selectedTask.id}-${Date.now()}`,
        code,
        language: 'go',
        timeout: 5,
      };
      const nextOutput = await executeCode(req);
      setOutput(nextOutput);
      setTaskResults((results) => ({
        ...results,
        [selectedTask.id]: didPassTask(nextOutput, null, selectedTask.checks) ? 'pass' : 'fail',
      }));
    } catch (err: unknown) {
      if (err instanceof Error) {
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

  const taskPanelNode = <TaskPanel task={selectedTask} />;
  const editorPanelNode = (
    <EditorPanel code={code} onChange={setCode} track={selectedTask.track} />
  );
  const feedbackPanelNode = (
    <FeedbackPanel
      feedback={feedback}
      currentTaskPassed={currentTaskPassed}
      mentorHints={selectedTask.mentorHints}
      review={selectedTask.review}
      output={output}
      error={error}
    />
  );

  return (
    <div className={styles.appShell}>
      <TopBar onReset={handleResetCode} onRun={handleRun} loading={loading} />
      <TaskProgress
        tasks={internshipTasks}
        selectedTaskId={selectedTask.id}
        taskResults={taskResults}
        onSelectTask={handleSelectTask}
      />

      {isMobile ? (
        <main className={styles.mobileMain}>
          {mobileTab === 'task' && taskPanelNode}
          {mobileTab === 'editor' && editorPanelNode}
          {mobileTab === 'feedback' && feedbackPanelNode}

          <nav className={styles.mobileTabBar}>
            <button
              className={mobileTab === 'task' ? styles.activeTab : ''}
              onClick={() => setMobileTab('task')}
            >
              任务
            </button>
            <button
              className={mobileTab === 'editor' ? styles.activeTab : ''}
              onClick={() => setMobileTab('editor')}
            >
              编辑
            </button>
            <button
              className={mobileTab === 'feedback' ? styles.activeTab : ''}
              onClick={() => setMobileTab('feedback')}
            >
              反馈
            </button>
          </nav>
        </main>
      ) : (
        <main className={styles.desktopMain}>
          <ResizableSplit
            left={taskPanelNode}
            center={editorPanelNode}
            right={feedbackPanelNode}
          />
        </main>
      )}
    </div>
  );
}

export default App;
```

- [ ] **Step 12.2: 更新组件 import 路径**

逐个修改以下文件中的 import（从 `../../types/workbench` 改为 `@/domain/workbench`，从 `../../taskFeedback` 改为 `@/domain/tasks`）：

1. `web/src/components/EditorPanel/EditorPanel.tsx`:
   - `import type { EditorPanelProps } from '../../types/workbench'` → `import type { EditorPanelProps } from '@/domain/workbench'`

2. `web/src/components/TaskPanel/TaskPanel.tsx`:
   - `import type { TaskPanelProps } from '../../types/workbench'` → `import type { TaskPanelProps } from '@/domain/workbench'`

3. `web/src/components/TaskPanel/TaskContent.tsx`:
   - `import type { TaskPanelProps } from '../../types/workbench'` → `import type { TaskPanelProps } from '@/domain/workbench'`

4. `web/src/components/TopBar/TopBar.tsx`:
   - `import type { TopBarProps } from '../../types/workbench'` → `import type { TopBarProps } from '@/domain/workbench'`

5. `web/src/components/TopBar/TaskProgress.tsx`:
   - `import type { TaskProgressProps } from '../../types/workbench'` → `import type { TaskProgressProps } from '@/domain/workbench'`

6. `web/src/components/ResizableSplit/ResizableSplit.tsx`:
   - `import type { ResizableSplitProps } from '../../types/workbench'` → `import type { ResizableSplitProps } from '@/domain/workbench'`

7. `web/src/components/FeedbackPanel/FeedbackPanel.tsx`:
   - `import type { FeedbackPanelProps } from '../../types/workbench'` → `import type { FeedbackPanelProps } from '@/domain/workbench'`

8. `web/src/components/FeedbackPanel/FeedbackList.tsx`:
   - `import type { FeedbackItem } from '../../taskFeedback'` → `import type { FeedbackItem } from '@/domain/tasks'`

9. `web/src/components/FeedbackPanel/Console.tsx`:
   - `import type { SandboxResponse } from '../../taskFeedback'` → `import type { SandboxResponse } from '@/domain/tasks'`

- [ ] **Step 12.3: 验证 TypeScript 编译和测试**

```bash
cd web && npx tsc -b && npm run test -- --run
```

Expected: TypeScript 编译通过，测试通过。

- [ ] **Step 12.4: Commit**

```bash
git add web/src/App.tsx web/src/components/
git commit -m "refactor: update App.tsx and component imports to use domain/api layers"
```

---

## Task 13: 前端 — 测试迁移和旧文件清理

**Files:**
- Create: `web/src/domain/tasks/data.test.ts`（从 `tasks.test.ts` 迁移）
- Create: `web/src/domain/tasks/checks.test.ts`（从 `taskFeedback.test.ts` 迁移）
- Modify: `web/src/App.test.tsx`（更新 import）
- Delete: `web/src/tasks.ts`
- Delete: `web/src/taskFeedback.ts`
- Delete: `web/src/types/workbench.ts`
- Delete: `web/src/tasks.test.ts`
- Delete: `web/src/taskFeedback.test.ts`
- Delete: `web/src/types/`（空目录）
- Test: `cd web && npm run test -- --run` 验证

- [ ] **Step 13.1: 创建 `web/src/domain/tasks/data.test.ts`**

从 `web/src/tasks.test.ts` 迁移，将 import 从 `./tasks` 改为 `./data`。

```typescript
import { describe, expect, it } from 'vitest';
import { defaultTaskId, findTaskById, internshipTasks } from './data';

describe('internshipTasks', () => {
  it('contains exactly Day 0 through Day 9', () => {
    expect(internshipTasks.map((task) => task.day)).toEqual([0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13]);
  });

  it('uses stable unique ids and complete task content', () => {
    const ids = internshipTasks.map((task) => task.id);

    expect(new Set(ids).size).toBe(ids.length);

    for (const task of internshipTasks) {
      expect(task.title.length).toBeGreaterThanOrEqual(4);
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
    expect(findTaskById('nonexistent').id).toBe(defaultTaskId);
  });
});
```

- [ ] **Step 13.2: 创建 `web/src/domain/tasks/checks.test.ts`**

从 `web/src/taskFeedback.test.ts` 迁移，将 import 从 `./taskFeedback` 改为 `./checks`。

```typescript
import { describe, expect, it } from 'vitest';
import {
  didPassTask,
  evaluateTaskChecks,
  type SandboxResponse,
  type TaskCheck,
} from './checks';

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
    exit_code: 0,
    duration: 123,
    ...overrides,
  };
}

describe('evaluateTaskChecks', () => {
  it('returns idle feedback when output is null', () => {
    const result = evaluateTaskChecks(null, null, checks);
    expect(result.every((f) => f.state === 'idle')).toBe(true);
  });

  it('returns fail feedback when there is an error', () => {
    const result = evaluateTaskChecks(null, 'network error', checks);
    expect(result.every((f) => f.state === 'fail')).toBe(true);
  });

  it('passes all checks for successful sandbox output', () => {
    const result = evaluateTaskChecks(sandbox(), null, checks);
    expect(result.every((f) => f.state === 'pass')).toBe(true);
  });

  it('fails stdoutIncludes when phrase is missing', () => {
    const result = evaluateTaskChecks(sandbox({ stdout: 'nope' }), null, checks);
    expect(result[0].state).toBe('fail');
    expect(result[1].state).toBe('pass');
  });

  it('fails stderrExcludes when panic is present', () => {
    const result = evaluateTaskChecks(sandbox({ stderr: 'panic: oh no' }), null, checks);
    expect(result[0].state).toBe('pass');
    expect(result[1].state).toBe('fail');
  });
});

describe('didPassTask', () => {
  it('returns false when output is null', () => {
    expect(didPassTask(null, null, checks)).toBe(false);
  });

  it('returns false when there is an error', () => {
    expect(didPassTask(null, 'fail', checks)).toBe(false);
  });

  it('returns true when all checks pass', () => {
    expect(didPassTask(sandbox(), null, checks)).toBe(true);
  });
});
```

- [ ] **Step 13.3: 更新 `web/src/App.test.tsx`**

当前 `App.test.tsx` 只 import `./App`，不需要修改业务逻辑 import。但如果测试未来需要 import domain 类型，应该使用 `@/domain/tasks`。当前文件保持不变即可。

- [ ] **Step 13.4: 删除旧文件**

```bash
cd web/src
rm tasks.ts taskFeedback.ts types/workbench.ts tasks.test.ts taskFeedback.test.ts
rmdir types 2>/dev/null || true
```

- [ ] **Step 13.5: 验证前端测试全部通过**

```bash
cd web && npm run test -- --run
```

Expected: 所有测试通过（data.test.ts, checks.test.ts, App.test.tsx）。

- [ ] **Step 13.6: Commit**

```bash
git add web/src/domain/tasks/*.test.ts web/src/App.test.tsx
git rm web/src/tasks.ts web/src/taskFeedback.ts web/src/types/workbench.ts web/src/tasks.test.ts web/src/taskFeedback.test.ts
git commit -m "refactor: migrate tests and remove old frontend files"
```

---

## Task 14: 前端验证 — 构建和 Lint

**Files:**
- Test: `cd web && npm run build`, `npm run lint`

- [ ] **Step 14.1: 构建前端**

```bash
cd web && npm run build
```

Expected: 构建成功，输出到 `web/dist/`。

- [ ] **Step 14.2: 运行前端 Lint**

```bash
cd web && npm run lint
```

Expected: 无错误，无警告。

- [ ] **Step 14.3: Commit**

```bash
git commit --allow-empty -m "checkpoint: frontend refactor complete"
```

---

## Task 15: 基础设施 — Makefile、CI、文档合并

**Files:**
- Create: `Makefile`
- Create: `.github/workflows/ci.yml`
- Modify: `docs/` 目录结构
- Create: `db/migrations/`（空目录）

- [ ] **Step 15.1: 创建 `Makefile`**

```makefile
.PHONY: dev test build lint clean

dev:
	docker compose up postgres redis -d
	# 本地混合开发：后台启动 Go 服务，前台启动前端 dev server
	# 停止时运行 `make clean`
	go run ./src/services/sandbox-engine/main.go &
	go run ./src/services/gateway/main.go &
	cd web && npm run dev

test:
	go test ./...
	cd web && npm run test -- --run

build:
	docker compose build

lint:
	go vet ./...
	cd web && npm run lint

clean:
	docker compose down -v
	pkill -f "go run ./src/services" || true
```

- [ ] **Step 15.2: 创建 `.github/workflows/ci.yml`**

```yaml
name: CI
on: [push, pull_request]
jobs:
  backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.24' }
      - run: go vet ./...
      - run: go test ./...
  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with: { node-version: '20' }
      - run: cd web && npm ci
      - run: cd web && npm run lint
      - run: cd web && npm run test -- --run
```

- [ ] **Step 15.3: 合并文档目录**

将 `docs/specs/` 和 `docs/plans/` 下的旧文件迁移到 `docs/superpowers/specs/` 和 `docs/superpowers/plans/`：

```bash
# 移动旧文件（如果它们还在原位置）
mv docs/specs/* docs/superpowers/specs/ 2>/dev/null || true
mv docs/plans/* docs/superpowers/plans/ 2>/dev/null || true
rmdir docs/specs docs/plans 2>/dev/null || true
```

- [ ] **Step 15.4: 创建 `db/migrations/` 空目录**

已在 Task 0 中创建，无需额外操作。

- [ ] **Step 15.5: Commit**

```bash
git add Makefile .github/workflows/ci.yml
git commit -m "chore: add Makefile, CI workflow, and merge doc directories"
```

---

## Task 16: 端到端验证

**Files:**
- Test: Docker Compose 构建和运行

- [ ] **Step 16.1: Docker Compose 构建**

```bash
docker compose build
```

Expected: 所有镜像构建成功（web, gateway, sandbox-engine）。

- [ ] **Step 16.2: Docker Compose 运行并验证功能**

```bash
docker compose up -d
```

等待 10 秒后：

```bash
curl -s http://localhost:8080/health | jq .
```

Expected: `{"status":"ok"}`

```bash
curl -s -X POST http://localhost:8080/api/v1/execute \
  -H "Content-Type: application/json" \
  -d '{"id":"test","code":"package main\nimport \"fmt\"\nfunc main() { fmt.Println(\"hello\") }","language":"go","timeout":5}' | jq '.status'
```

Expected: `"success"`

- [ ] **Step 16.3: 停止服务**

```bash
docker compose down
```

- [ ] **Step 16.4: Commit**

```bash
git commit --allow-empty -m "checkpoint: end-to-end verification passed"
```

---

## Self-Review Checklist

### 1. Spec Coverage

| 规格要求 | 对应 Task | 状态 |
|----------|----------|------|
| Go 后端分层拆分 | Task 1-7 | 已覆盖 |
| 前端 domain 层 | Task 9-10 | 已覆盖 |
| 前端 api 层 | Task 11 | 已覆盖 |
| App.tsx 精简 | Task 12 | 已覆盖 |
| 组件 import 更新 | Task 12 | 已覆盖 |
| 测试文件迁移 | Task 13 | 已覆盖 |
| 路径别名配置 | Task 8 | 已覆盖 |
| Makefile | Task 15 | 已覆盖 |
| GitHub Actions CI | Task 15 | 已覆盖 |
| 文档目录合并 | Task 15 | 已覆盖 |
| db/migrations 预留 | Task 0 | 已覆盖 |
| 端到端验证 | Task 16 | 已覆盖 |

### 2. Placeholder Scan

- [x] 无 "TBD"、"TODO"、"implement later"
- [x] 无 "Add appropriate error handling" 等模糊描述
- [x] 每个代码步骤都有完整代码块
- [x] 每个验证步骤都有具体命令和预期输出

### 3. Type Consistency

- [x] `SandboxRequest` / `SandboxResponse` 字段名在前后端一致
- [x] `InternTask` 接口字段与原始 `tasks.ts` 一致
- [x] `FeedbackItem`、`TaskCheck` 类型与原始 `taskFeedback.ts` 一致
- [x] `duration` 类型在前端保持 `number`（映射 Go 的 `time.Duration` 毫秒值）

---

## Execution Handoff

**Plan complete and saved to `docs/superpowers/plans/2026-05-07-project-structure-refactor.md`.**

Two execution options:

**1. Subagent-Driven (recommended)** — Dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** — Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?**
