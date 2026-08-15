# 密钥轮换手册

本文描述仓库 5 个核心密钥的轮换步骤与影响面。原则：先在外部系统（数据库 / 对象存储）新建凭据并验证可用，再回填到 `.env` 重启 api 容器；避免"先改应用、再改外部"导致两段时间窗内应用无法连接。

> 通用前置：
> - 当前部署的 .env 在部署机，路径随运维约定。修改前先 `cp .env .env.bak.$(date +%Y%m%d-%H%M%S)`。
> - 每次轮换在 `reviews/` 下记录一次审查（即便只是 "按计划轮换，无异常"），便于审计。
> - 容器内 API 监听 8080；只有 `127.0.0.1:APP_PORT` 暴露给宿主机调试，外部流量全部走 Nginx。

## 1. `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET`

**影响面**：轮换会让所有当前已签发的 access token 与 refresh token 立即失效。前端表现为"突然需要重新登录"；后端表现为 refresh token 校验失败的 401。

**步骤**：

1. 在部署机生成两个新 secret（必须互不相同、≥32 字符）：
   ```bash
   openssl rand -base64 64   # 写入 JWT_ACCESS_SECRET
   openssl rand -base64 64   # 写入 JWT_REFRESH_SECRET
   ```
2. 编辑 `.env` 替换两行；保留旧值另存 `secrets-rotation.log` 一行记录替换时间。
3. **滚动重启 api**（避免整站瞬时全部下线）：
   ```bash
   docker compose --env-file .env -f deploy/docker-compose.yml up -d --no-deps --force-recreate api
   docker compose --env-file .env -f deploy/docker-compose.yml exec api wget -qO- http://localhost:8080/healthz
   ```
   `nginx` 容器无需重启。
4. 通知用户：约在重启时刻之后登录过的会话需要重新登录。`/healthz` 不依赖 JWT，所以脚本失败只可能是配置或镜像问题。

**回滚**：把 `.env` 恢复为旧值，再 `up -d --force-recreate api`。已签发的旧 token 不会因为新 secret 部署而被撤销，只要回滚即可继续生效。

## 2. `MYSQL_PASSWORD`

**影响面**：轮换期间会切断数据库连接；`/healthz` 在切换瞬间返回 503，但 nginx 自检 `/nginx-health` 仍 200，所以 LB 不会立刻把全站摘除。

**步骤**：

1. 在 MySQL 托管方控制台**先**新建一个新账号（或者用 `ALTER USER ... IDENTIFIED BY`），确认能 `mysql -h ... -u <user> -p<new>` 登录。
2. 验证新账号的权限（库级 `SELECT / INSERT / UPDATE / DELETE`）与原账号一致：
   ```bash
   mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "<new_user>" -p<new_pw> -e "SHOW GRANTS FOR CURRENT_USER();" "$MYSQL_DATABASE"
   ```
3. 改 `.env` 的 `MYSQL_USER` / `MYSQL_PASSWORD`，删除旧账号或保留作冷备（建议保留 7 天）。
4. 滚动重启 api（同上）。
5. 跑 `sh deploy/scripts/healthcheck.sh` 确认 `mysql: ok`。

**回滚**：恢复旧账号权限或把 `.env` 改回旧值。

## 3. `REDIS_PASSWORD`

**影响面**：Redis 是缓存层，丢缓存不会丢数据，但缓存击穿期间 MySQL 压力会上升，仪表盘可能出现慢查询。

**步骤**：

1. 在 Redis 6+ 控制台给 ACL 用户 `app` 设置新密码（`ACL SETUSER app on '>newpassword'`），或者用托管方提供的"重置密码"功能。
2. 验证新密码：
   ```bash
   redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" --user app --pass <new_pw> PING
   ```
3. 改 `.env` 的 `REDIS_PASSWORD`，滚动重启 api。
4. `sh deploy/scripts/healthcheck.sh` 确认 `redis: ok`。缓存会自然重建，不需要手动预热。

**回滚**：恢复旧 ACL 密码 / 改回 `.env` 重启。

## 4. `ADMIN_PASSWORD`

**影响面**：单一 owner 账号的口令。轮换不会让会话失效（JWT 独立于 admin 密码），但会改变下次登录所需的口令。

**步骤**：

1. **不要通过改 `.env` 来"轮换"**。`admin` 用户的密码哈希存在 `users` 表里，`.env` 里的 `ADMIN_PASSWORD` 只在 owner 账号不存在时被 bootstrap 一次。
2. 登录 `/admin/login`，进入管理后台修改密码。
3. 旧的 `ADMIN_PASSWORD` 保留在 `.env` 中不再被使用；如要清理，可在确认 owner 已经能正常登录后把 `.env` 里那行改为占位符并加注释。
4. 若遗忘了口令且无法登录：**唯一恢复路径是直接修改 `users` 表的 `password_hash`**（用 `bcrypt` 重新哈希新口令）。`bcrypt.GenerateFromPassword([]byte(newPwd), bcrypt.DefaultCost)`，然后 `UPDATE users SET password_hash = ? WHERE username = 'admin';`。

## 5. `STORAGE_S3_ACCESS_KEY` / `STORAGE_S3_SECRET_KEY`

**影响面**：轮换期间对象存储的所有上传 / 删除 / 签名 URL 请求会失败；公开读取（走 `STORAGE_S3_PUBLIC_URL` 的浏览器请求）不受影响。已通过 `STORAGE_S3_PUBLIC_URL` 分发的外链 URL 在 secret 轮换后**仍可继续访问**（S3/MinIO 的对象 URL 默认长期有效），所以**轮换 S3 凭据不需要批量改 HTML**。

**步骤**：

1. 在 MinIO / S3 控制台创建一个新 AccessKey + SecretKey。
2. 验证新凭据：使用 `mc alias set` 或 `aws s3 ls --endpoint-url ... s3://blog/` 测试读写。
3. 改 `.env` 的 `STORAGE_S3_ACCESS_KEY` / `STORAGE_S3_SECRET_KEY`，**保留旧凭据 24 小时不删**（以防旧实例还在用旧 key 上传）。
4. 滚动重启 api。
5. 做一次"受控上传"冒烟测试（管理后台新建一个 asset），确认 200。

**回滚**：恢复旧 AccessKey，删除新 AccessKey。

## 6. 轮换前后检查清单

```bash
# 1. 修改 .env 之前
cp .env .env.bak.$(date +%Y%m%d-%H%M%S)

# 2. 修改 .env
$EDITOR .env

# 3. 校验配置语法
docker compose --env-file .env -f deploy/docker-compose.yml config --quiet

# 4. 滚动重启 api
docker compose --env-file .env -f deploy/docker-compose.yml up -d --no-deps --force-recreate api

# 5. 等 ~15s 让 healthcheck 走完
sleep 15

# 6. 验收
sh deploy/scripts/healthcheck.sh

# 7. 观察错误率
docker compose --env-file .env -f deploy/docker-compose.yml logs --since=5m api | grep -iE 'unauthor|forbid|denied|panic'
# 期望：无新增 unauthorized / 401-风暴
```
