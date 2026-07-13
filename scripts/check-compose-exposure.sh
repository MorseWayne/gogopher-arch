#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ASSERTION="$ROOT_DIR/scripts/assert-compose-exposure.mjs"

docker compose \
  --project-directory "$ROOT_DIR" \
  -f "$ROOT_DIR/docker-compose.yml" \
  config --format json \
  | node "$ASSERTION" base

docker compose \
  --project-directory "$ROOT_DIR" \
  -f "$ROOT_DIR/docker-compose.yml" \
  -f "$ROOT_DIR/docker-compose.dev.yml" \
  config --format json \
  | node "$ASSERTION" dev
