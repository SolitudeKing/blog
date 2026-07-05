# M4 部署与备份运行手册

本文档用于 M4 阶段完成新个人博客系统的上线前部署验证。

## 环境变量

1. 复制根目录 `.env.example` 为 `.env`。
2. 修改 MySQL、Redis、JWT、管理员密码和端口配置。
3. Docker Compose 环境中建议保持：

   ```env
   MYSQL_DSN=blog:blog@tcp(mysql:3306)/blog?charset=utf8mb4&parseTime=True&loc=UTC
   REDIS_ADDR=redis:6379
   STORAGE_LOCAL_ROOT=/app/storage/uploads
   ```

## 启动服务

```bash
docker compose -f deploy/docker-compose.yml up -d --build
```

Compose 会等待 MySQL 和 Redis 健康后再启动 API，Nginx 会等待 API 健康后再对外服务。

## 健康检查

```bash
sh deploy/scripts/healthcheck.sh
```

验收点：

- `mysql` 状态为 healthy。
- `redis` 状态为 healthy。
- `api` 状态为 healthy。
- `worker` 状态为 healthy。
- `nginx` 状态为 healthy。
- `GET /healthz` 返回 `api/mysql/redis` 状态。

## 备份 MySQL

```bash
sh deploy/scripts/backup-mysql.sh
```

备份文件默认写入 `backups/mysql/`，文件名格式为 `{database}_{utc timestamp}.sql.gz`。

## 恢复 MySQL

```bash
sh deploy/scripts/restore-mysql.sh backups/mysql/blog_20260705T120000Z.sql.gz
```

恢复前建议先停止 API 和 Worker，避免恢复期间继续写入：

```bash
docker compose -f deploy/docker-compose.yml stop api worker
```

恢复完成后重新启动：

```bash
docker compose -f deploy/docker-compose.yml up -d api worker
```

## 日志与持久化

- MySQL 数据保存在 `mysql-data` volume。
- Redis 开启 AOF，数据保存在 `redis-data` volume。
- 上传文件保存在 `api-storage` volume，并同时挂载到 API 与 Worker。
- 服务日志使用 Docker `json-file`，单文件 10MB，最多保留 5 个文件。
