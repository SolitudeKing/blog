#!/usr/bin/env sh
set -eu

# 备份包含完整业务数据，临时 SQL 与最终压缩包只允许当前用户读写。
umask 077

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT_DIR/deploy/docker-compose.yml}"
ENV_FILE="${ENV_FILE:-$ROOT_DIR/.env}"
read_env_key() {
  [ -f "$ENV_FILE" ] || return 0
  awk -v key="$1" 'index($0, key "=") == 1 { print substr($0, length(key) + 2); exit }' "$ENV_FILE"
}

BACKUP_DIR="${BACKUP_DIR:-$ROOT_DIR/backups/mysql}"
MYSQL_DATABASE="${MYSQL_DATABASE:-$(read_env_key MYSQL_DATABASE)}"
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-$(read_env_key MYSQL_ROOT_PASSWORD)}"
MYSQL_DATABASE="${MYSQL_DATABASE:-blog}"
: "${MYSQL_ROOT_PASSWORD:?MYSQL_ROOT_PASSWORD must be set in the environment or .env}"
TIMESTAMP="$(date -u +%Y%m%dT%H%M%SZ)"
TARGET="$BACKUP_DIR/${MYSQL_DATABASE}_${TIMESTAMP}.sql.gz"
TEMP_SQL="$TARGET.sql.tmp"
TEMP_GZIP="$TARGET.tmp"

cleanup() {
  rm -f "$TEMP_SQL" "$TEMP_GZIP"
}
trap cleanup EXIT HUP INT TERM

mkdir -p "$BACKUP_DIR"

docker compose --env-file "$ENV_FILE" -f "$COMPOSE_FILE" exec -T mysql \
  mysqldump -uroot -p"$MYSQL_ROOT_PASSWORD" \
  --single-transaction \
  --routines \
  --triggers \
  --default-character-set=utf8mb4 \
  "$MYSQL_DATABASE" > "$TEMP_SQL"

gzip -c "$TEMP_SQL" > "$TEMP_GZIP"
gzip -t "$TEMP_GZIP"
mv "$TEMP_GZIP" "$TARGET"

printf '%s\n' "$TARGET"
