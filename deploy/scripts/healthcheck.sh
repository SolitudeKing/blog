#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT_DIR/deploy/docker-compose.yml}"
read_env_key() {
  [ -f "$ROOT_DIR/.env" ] || return 0
  awk -v key="$1" 'index($0, key "=") == 1 { print substr($0, length(key) + 2); exit }' "$ROOT_DIR/.env"
}

WEB_PORT="${WEB_PORT:-$(read_env_key WEB_PORT)}"
WEB_PORT="${WEB_PORT:-80}"
HEALTH_URL="${HEALTH_URL:-http://localhost:$WEB_PORT/healthz}"

docker compose -f "$COMPOSE_FILE" ps

if command -v curl >/dev/null 2>&1; then
  curl -fsS "$HEALTH_URL"
else
  wget -qO- "$HEALTH_URL"
fi

printf '\nhealthcheck ok: %s\n' "$HEALTH_URL"
