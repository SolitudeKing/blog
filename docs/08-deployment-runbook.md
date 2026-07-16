# 部署与备份运行手册

本文说明当前 Compose 部署方式。仓库内 Nginx 只监听 HTTP；生产 HTTPS、域名和证书由实际边缘代理或托管平台负责。

## 1. 准备环境变量

```bash
cp .env.example .env
```

必须填写 MySQL、Redis、JWT 和管理员凭据。Compose 不再为数据库和 Redis 提供弱密码回退。
生产部署还必须将 `APP_ENV` 设为 `production`，并把 `SITE_BASE_URL` 改为真实 HTTPS 站点地址。
`.env.example` 的 `SITE_BASE_URL=http://localhost` 对应完整 Compose 的 Nginx 入口；只有在宿主机运行 Vite 开发服务器时才改为 `http://localhost:5173`。

`MYSQL_DSN` 包含数据库账号与密码，必须与 `MYSQL_USER`、`MYSQL_PASSWORD` 同步：

```env
MYSQL_DSN=blog:<MYSQL_PASSWORD>@tcp(mysql:3306)/blog?charset=utf8mb4&parseTime=True&loc=UTC
MYSQL_DATABASE=blog
MYSQL_USER=blog
MYSQL_PASSWORD=<同一密码>
REDIS_ADDR=redis:6379
REDIS_PASSWORD=<强密码>
STORAGE_LOCAL_ROOT=/app/storage/uploads
```

不要提交实际 `.env`。占位符不能直接用于生产。

## 2. 展开并启动

先检查最终配置：

```bash
docker compose --env-file .env -f deploy/docker-compose.yml config --quiet
```

再启动服务：

```bash
docker compose --env-file .env -f deploy/docker-compose.yml up -d --build
```

服务启动关系：MySQL、Redis 健康后启动 API，API 健康后启动 Nginx。MySQL、Redis、API 的宿主机端口仅绑定 `127.0.0.1`，外部流量只进入 Nginx。

## 3. 健康检查

```bash
sh deploy/scripts/healthcheck.sh
docker compose --env-file .env -f deploy/docker-compose.yml exec nginx nginx -t
```

验收点：

- `mysql`、`redis`、`api`、`nginx` 均为 running/healthy。
- `GET /healthz` 返回 API、MySQL 与 Redis 的健康结果；任一持久化依赖异常时返回非 2xx，使脚本和容器健康检查失败。
- 首页、登录、公开文章列表、RSS 与 sitemap 可访问。
- 一次受控媒体上传未触发代理层 1 MB 默认限制。

## 4. 备份 MySQL

```bash
sh deploy/scripts/backup-mysql.sh
```

脚本从根 `.env` 读取数据库名与 root 密码，以 `umask 077` 限制临时文件和备份文件权限，先生成临时 SQL，再压缩和执行 `gzip -t`，成功后原子移动为 `backups/mysql/{database}_{UTC时间}.sql.gz`。

当前脚本只备份 MySQL，不包含 `api-storage` 中的上传文件，因此它不是完整站点备份。上线前必须补充上传卷备份，并让数据库与媒体恢复点能够对应。

## 5. 恢复 MySQL

恢复会覆盖现有数据。先备份当前状态并停止 API 写入：

```bash
docker compose --env-file .env -f deploy/docker-compose.yml stop api
```

显式确认后恢复：

```bash
CONFIRM_RESTORE=1 sh deploy/scripts/restore-mysql.sh backups/mysql/blog_20260716T120000Z.sql.gz
```

重新启动并验收：

```bash
docker compose --env-file .env -f deploy/docker-compose.yml up -d api nginx
sh deploy/scripts/healthcheck.sh
```

压缩包会在导入前执行 `gzip -t`，但首次恢复仍必须在一次性数据库或隔离环境演练，不能直接把生产库作为测试目标。

## 6. 持久化与日志

| 数据 | 位置 | 当前保护 |
| --- | --- | --- |
| MySQL | `mysql-data` volume | 数据库备份脚本 |
| Redis | `redis-data` volume，AOF | 主要承载可重建缓存，不作为业务事实来源 |
| 上传文件 | `api-storage` volume | 尚需补充独立备份 |
| 容器日志 | Docker `json-file` | 单文件 10 MB，最多 5 个文件 |

Compose 使用 `restart: unless-stopped`。部署机仍需监控磁盘、备份结果、证书、容器健康和恢复演练时间。

## 7. 回滚原则

- 应用回滚前确认新版本是否已执行不可逆数据变更。
- 不直接删除命名卷解决启动问题。
- 数据恢复前保存当前数据库与上传卷快照。
- 回滚完成后核对三专题、文章数量、媒体引用、站点设置、RSS 和 sitemap。
