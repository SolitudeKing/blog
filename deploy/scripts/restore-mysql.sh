#!/usr/bin/env sh
set -eu

if [ "$#" -ne 1 ]; then
  printf 'usage: %s <backup.sql|backup.sql.gz>\n' "$0" >&2
  exit 2
fi

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
COMPOSE_FILE="${COMPOSE_FILE:-$ROOT_DIR/deploy/docker-compose.yml}"
read_env_key() {
  [ -f "$ROOT_DIR/.env" ] || return 0
  awk -v key="$1" 'index($0, key "=") == 1 { print substr($0, length(key) + 2); exit }' "$ROOT_DIR/.env"
}

MYSQL_DATABASE="${MYSQL_DATABASE:-$(read_env_key MYSQL_DATABASE)}"
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-$(read_env_key MYSQL_ROOT_PASSWORD)}"
MYSQL_DATABASE="${MYSQL_DATABASE:-blog}"
MYSQL_ROOT_PASSWORD="${MYSQL_ROOT_PASSWORD:-root}"
BACKUP_FILE="$1"

if [ ! -f "$BACKUP_FILE" ]; then
  printf 'backup file not found: %s\n' "$BACKUP_FILE" >&2
  exit 1
fi

case "$BACKUP_FILE" in
  *.gz)
    gzip -dc "$BACKUP_FILE" | docker compose -f "$COMPOSE_FILE" exec -T mysql \
      mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE"
    ;;
  *)
    docker compose -f "$COMPOSE_FILE" exec -T mysql \
      mysql -uroot -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" < "$BACKUP_FILE"
    ;;
esac

printf 'restored %s into %s\n' "$BACKUP_FILE" "$MYSQL_DATABASE"
