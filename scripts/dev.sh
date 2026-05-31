#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WEB_DIR="$ROOT_DIR/web"

LOCAL_DB_URL="${DB_URL:-postgres://user:pass@localhost:5432/gogopher?sslmode=disable}"
LOCAL_REDIS_URL="${REDIS_URL:-localhost:6379}"
LOCAL_SANDBOX_URL="${SANDBOX_URL:-http://localhost:8081/execute}"

usage() {
  cat <<'EOF'
GoGopher Arch 启动助手

用法:
  ./scripts/dev.sh <command>

常用命令:
  help             显示帮助
  docker           全 Docker 生产式启动，会重新构建镜像
  docker:up        全 Docker 生产式启动，不主动重新构建镜像
  docker:down      停止 Docker Compose 服务
  deps             只启动 Postgres 和 Redis
  backend          用 Docker 启动 gateway、sandbox-engine、Postgres、Redis
  web              启动本地 Vite 前端开发服务，访问 http://localhost:5173
  sandbox          本地启动 Sandbox Engine，自动注入本地 DB/Redis 环境变量
  gateway          本地启动 Gateway，自动注入本地 DB/Redis/Sandbox 环境变量
  local            启动 Postgres/Redis，并提示分别运行 sandbox、gateway、web
  status           查看 Docker Compose 服务状态
  logs [service]   跟随查看 Docker Compose 日志，可选 service 名

推荐场景:
  只改前端 UI:
    ./scripts/dev.sh backend
    ./scripts/dev.sh web

  前后端都要热开发:
    ./scripts/dev.sh deps
    ./scripts/dev.sh sandbox
    ./scripts/dev.sh gateway
    ./scripts/dev.sh web

  验证完整容器部署:
    ./scripts/dev.sh docker
EOF
}

ensure_web_deps() {
  if [[ ! -d "$WEB_DIR/node_modules" ]]; then
    echo "未发现 web/node_modules，正在执行 npm install..."
    npm install --prefix "$WEB_DIR"
  fi
}

run_compose() {
  docker compose --project-directory "$ROOT_DIR" "$@"
}

command="${1:-help}"
shift || true

case "$command" in
  help|-h|--help)
    usage
    ;;
  docker)
    run_compose up --build "$@"
    ;;
  docker:up)
    run_compose up "$@"
    ;;
  docker:down|down)
    run_compose down "$@"
    ;;
  deps)
    run_compose up -d postgres redis
    ;;
  backend)
    run_compose up -d gateway
    ;;
  web)
    ensure_web_deps
    npm run dev --prefix "$WEB_DIR"
    ;;
  sandbox)
    export DB_URL="$LOCAL_DB_URL"
    export REDIS_URL="$LOCAL_REDIS_URL"
    echo "DB_URL=$DB_URL"
    echo "REDIS_URL=$REDIS_URL"
    go run "$ROOT_DIR/src/services/sandbox-engine/main.go"
    ;;
  gateway)
    export DB_URL="$LOCAL_DB_URL"
    export REDIS_URL="$LOCAL_REDIS_URL"
    export SANDBOX_URL="$LOCAL_SANDBOX_URL"
    echo "DB_URL=$DB_URL"
    echo "REDIS_URL=$REDIS_URL"
    echo "SANDBOX_URL=$SANDBOX_URL"
    go run "$ROOT_DIR/src/services/gateway/main.go"
    ;;
  local)
    run_compose up -d postgres redis
    cat <<EOF
Postgres 和 Redis 已启动。

请分别开 3 个终端运行：
  ./scripts/dev.sh sandbox
  ./scripts/dev.sh gateway
  ./scripts/dev.sh web

访问：
  http://localhost:5173
EOF
    ;;
  status|ps)
    run_compose ps "$@"
    ;;
  logs)
    run_compose logs -f "$@"
    ;;
  *)
    echo "未知命令：$command" >&2
    echo >&2
    usage >&2
    exit 1
    ;;
esac
