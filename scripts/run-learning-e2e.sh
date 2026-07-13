#!/bin/sh
set -eu

root="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"
project="${E2E_COMPOSE_PROJECT:-gogopher-learning-e2e}"
port="${E2E_WEB_PORT:-4173}"

compose() {
  docker compose \
    --project-name "$project" \
    --file "$root/docker-compose.yml" \
    --file "$root/docker-compose.e2e.yml" \
    "$@"
}

cleanup() {
  compose down --volumes --remove-orphans
}
trap cleanup EXIT INT TERM

compose down --volumes --remove-orphans
WEB_PORT="$port" compose up --detach --build --wait

cd "$root/web"
E2E_BASE_URL="http://127.0.0.1:$port" \
E2E_COMPOSE_PROJECT="$project" \
E2E_COMPOSE_ROOT="$root" \
npx playwright test e2e/learning-slice.spec.ts
