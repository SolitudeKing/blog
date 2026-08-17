#!/usr/bin/env bash
# POSIX 版本：在 Linux / macOS 开发机上推送制品并重启容器。
# 等价于 deploy/scripts/deploy.ps1。
#
# 用法：
#   deploy.sh user@host [remote-root] [--fallback] [--skip-build]
#
# 默认流程：build-artifacts → rsync 制品 → ssh 触发 docker compose up -d → healthcheck。
# --fallback 切换到源码 build 的回退路径。

set -euo pipefail
HOST="${1:?usage: deploy.sh user@host [remote-root] [--fallback] [--skip-build]}"
REMOTE_ROOT="${2:-/opt/solitude-blog}"
shift 2 || true
FALLBACK=0
SKIP_BUILD=0
for arg in "$@"; do
  case "$arg" in
    --fallback) FALLBACK=1 ;;
    --skip-build) SKIP_BUILD=1 ;;
    *) echo "unknown arg: $arg" >&2; exit 1 ;;
  esac
done

LOCAL_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
COMPOSE_FILE="deploy/docker-compose.runtime.yml"
if [ "$FALLBACK" = "1" ]; then
  COMPOSE_FILE="deploy/docker-compose.yml"
fi

if [ "$SKIP_BUILD" = "0" ]; then
  "$LOCAL_ROOT/deploy/scripts/build-artifacts.sh"
fi

echo "[deploy] ensure remote dirs on $HOST"
ssh "$HOST" "mkdir -p $REMOTE_ROOT/deploy/runtime/api $REMOTE_ROOT/deploy/runtime/web"

echo "[deploy] rsync artifacts to $HOST:$REMOTE_ROOT"
rsync -az --delete "$LOCAL_ROOT/deploy/runtime/api/" "$HOST:$REMOTE_ROOT/deploy/runtime/api/"
rsync -az --delete "$LOCAL_ROOT/deploy/runtime/web/" "$HOST:$REMOTE_ROOT/deploy/runtime/web/"
[ -f "$LOCAL_ROOT/.env" ] && rsync -az "$LOCAL_ROOT/.env" "$HOST:$REMOTE_ROOT/.env"

UP_FLAGS="-f $COMPOSE_FILE up -d"
[ "$FALLBACK" = "1" ] && UP_FLAGS="-f deploy/docker-compose.yml up -d --build"

echo "[deploy] restarting containers on $HOST"
ssh "$HOST" "cd $REMOTE_ROOT && docker compose --env-file .env $UP_FLAGS"

echo "[deploy] running healthcheck on $HOST"
ssh "$HOST" "cd $REMOTE_ROOT && COMPOSE_FILE=$REMOTE_ROOT/$COMPOSE_FILE sh deploy/scripts/healthcheck.sh"