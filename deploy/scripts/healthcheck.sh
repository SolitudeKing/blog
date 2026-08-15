#!/usr/bin/env sh
set -eu

# 本脚本仅校验本仓库持有的容器（api / nginx）。
# MySQL / Redis 由外部托管服务提供，不在 Compose 内，
# 它们对 API 的健康影响通过 /healthz 端点间接观察：
#   - handler/health_handler.go 在 MySQL / Redis 异常时返回非 2xx
#   - 因此 api 容器 healthy + /healthz 200 等价于"外部依赖可达"

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT_DIR/deploy/docker-compose.yml}"
ENV_FILE="${ENV_FILE:-$ROOT_DIR/.env}"
read_env_key() {
  [ -f "$ENV_FILE" ] || return 0
  # 清洗 CRLF：Windows 提交 .env 时行尾带 \r，拼接 URL 会变成 .../healthz\r
  # 触发 -fsS 的非 2xx 误报。
  awk -v key="$1" 'index($0, key "=") == 1 { print substr($0, length(key) + 2); exit }' "$ENV_FILE" \
    | tr -d '\r'
}

WEB_PORT="${WEB_PORT:-$(read_env_key WEB_PORT)}"
WEB_PORT="${WEB_PORT:-80}"
HEALTH_URL="${HEALTH_URL:-http://localhost:$WEB_PORT/healthz}"

SERVICES="api nginx"
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps $SERVICES

# compose ps 默认只展示状态，不会因服务退出而失败；逐项检查可避免把不完整的服务组误报为健康。
RUNNING_SERVICES="$(docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps --services --status running $SERVICES)"
for service in $SERVICES; do
  if ! printf '%s\n' "$RUNNING_SERVICES" | grep -qx "$service"; then
    printf 'service is not running: %s\n' "$service" >&2
    exit 1
  fi
done

# /healthz 在外部 MySQL 或 Redis 异常时返回非 2xx；-f/非零退出码会阻止成功提示。
# 进一步断言返回的是合法 JSON 且 data.api == "ok"，
# 防止 nginx 把 SPA 的 index.html 当成 200 误判通过。
fetch_health_body() {
  if command -v curl >/dev/null 2>&1; then
    curl -fsS "$HEALTH_URL"
  else
    wget -qO- "$HEALTH_URL"
  fi
}

HEALTH_BODY="$(fetch_health_body)"
if command -v python3 >/dev/null 2>&1; then
  printf '%s' "$HEALTH_BODY" | python3 -c '
import json, sys
try:
    payload = json.load(sys.stdin)
except json.JSONDecodeError as exc:
    print(f"healthz response is not valid JSON: {exc}", file=sys.stderr)
    sys.exit(1)
data = payload.get("data", payload)
api_status = data.get("api") if isinstance(data, dict) else None
if api_status != "ok":
    print(f"healthz reports api={api_status!r}, want ok", file=sys.stderr)
    sys.exit(1)
'
elif command -v jq >/dev/null 2>&1; then
  printf '%s' "$HEALTH_BODY" | jq -e '.data.api == "ok"' >/dev/null
else
  printf 'neither python3 nor jq available; cannot validate healthz payload\n' >&2
  exit 1
fi

printf '\nhealthcheck ok: %s\n' "$HEALTH_URL"
