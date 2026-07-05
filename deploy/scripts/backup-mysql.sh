#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT_DIR/deploy/docker-compose.yml}"
read_env_key() {
  [ -f "$ROOT_DIR/.env" ] || return 0
  awk -v key="$1" 'index($0, key "=") == 1 { print substr($0, length(key) + 2); exit }' "$ROOT_DIR/.env"
}

BACKUP_DIR="${BACKUP_DIR:-$ROOT_DIR/backups/mysql}"
MYSQL_DATABASE="${MYSQL_DATABASE:-$(read_env_key MYSQL_DATABASE)}"
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-$(read_env_key MYSQL_ROOT_PASSWORD)}"
MYSQL_DATABASE="${MYSQL_DATABASE:-blog}"
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-root}"
TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
TARGET="$BACKUP_DIR/${MYSQL_DATABASE}_${TIMESTAMP}.sql.gz"

mkdir -p "$BACKUP_DIR"

docker compose -f "$COMPOSE_FILE" exec -T mysql \
  mysqldump -uroot -p"$MYSQL_ROOT_PASSWORD" \
  --single-transaction \
  --routines \
  --triggers \
  --default-character-set=utf8mb4 \
  "$MYSQL_DATABASE" | gzip > "$TARGET"

printf '%s\n' "$TARGET"
