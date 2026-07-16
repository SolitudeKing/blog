#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT_DIR/deploy/docker-compose.yml}"
ENV_FILE="${ENV_FILE:-$ROOT_DIR/.env}"
read_env_key() {
  [ -f "$ENV_FILE" ] || return 0
  awk -v key="$1" 'index($0, key "=") == 1 { print substr($0, length(key) + 2); exit }' "$ENV_FILE"
}

WEB_PORT="${WEB_PORT:-$(read_env_key WEB_PORT)}"
WEB_PORT="${WEB_PORT:-80}"
HEALTH_URL="${HEALTH_URL:-http://localhost:$WEB_PORT/healthz}"

SERVICES="mysql redis api nginx"
docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps $SERVICES

# compose ps 默认只展示状态，不会因服务退出而失败；逐项检查可避免把不完整的服务组误报为健康。
RUNNING_SERVICES="$(docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" ps --services --status running $SERVICES)"
for service in $SERVICES; do
  if ! printf '%s\n' "$RUNNING_SERVICES" | grep -qx "$service"; then
    printf 'service is not running: %s\n' "$service" >&2
    exit 1
  fi
done

# /healthz 在 MySQL 或 Redis 异常时返回非 2xx；-f/非零退出码会阻止成功提示。
if command -v curl >/dev/null 2>&1; then
  curl -fsS "$HEALTH_URL"
else
  wget -qO- "$HEALTH_URL"
fi

printf '\nhealthcheck ok: %s\n' "$HEALTH_URL"
