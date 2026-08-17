#!/usr/bin/env bash
# POSIX 版本：在 Linux / macOS 开发机或 CI 上构建制品。
# 等价于 deploy/scripts/build-artifacts.ps1。

set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
RUNTIME="$ROOT/deploy/runtime"
mkdir -p "$RUNTIME/api" "$RUNTIME/web"

echo "[1/2] Building Go API -> $RUNTIME/api"
(
  cd "$ROOT"
  # 与 server/Dockerfile 第 11 行保持一致：trimpath + ldflags -s -w -buildid=
  CGO_ENABLED=0 GOOS=linux GOFLAGS=-trimpath \
    go build -trimpath -ldflags="-s -w -buildid=" -o "$RUNTIME/api/api" ./cmd/api
)

echo "[2/2] Building Vue SPA -> $RUNTIME/web"
(
  cd "$ROOT/web"
  if [ ! -d node_modules ]; then
    npm ci
  else
    echo "  node_modules present; skipping npm ci"
  fi
  npm run build
)

# 镜像同步 web/dist -> deploy/runtime/web，自动清理旧 hashed 资产。
rm -rf "$RUNTIME/web"/*
cp -R "$ROOT/web/dist/." "$RUNTIME/web/"

# nginx worker 以 nginx 用户运行；放宽权限避免宿主机收紧 mode 阻断读取。
chmod -R a+rX "$RUNTIME/web"

echo "Artifacts ready in $RUNTIME"
find "$RUNTIME" -maxdepth 3 -print