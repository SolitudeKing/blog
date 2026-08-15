# 0. 端到端部署指南

> **本指南面向接手生产部署的运维**：从一台刚开机的 Linux 服务器出发，到站点对外可访问、管理员可登录后台、媒体可上传。所有命令默认从 `/opt/solitude-blog` 目录执行，使用 `deploy/.env.production` 作为唯一生产环境配置；密钥由运维现场生成后填入。已有部署的滚动升级见 §12；备份与回滚的深度说明见 [`01-deployment-and-backup.md`](./01-deployment-and-backup.md)；密钥轮换见 [`02-secret-rotation.md`](./02-secret-rotation.md)。

## 1. 适用对象 & 范围

- **谁用**：接手生产部署的运维；本文假设你已有 Docker 与 Linux 基础。
- **适用**：基于 `git clone` 拉取最新代码的全新部署。
- **不适用**：本机开发（见仓库根 [README.md](../../../README.md) §"本机开发"）；自建 MySQL/Redis（外部依赖由 §4 单独准备，本仓库不托管）。

## 2. 架构速览

两个容器由本仓库的 `deploy/docker-compose.yml` 编排：

| 容器 | 镜像来源 | 端口 | 作用 |
| --- | --- | --- | --- |
| `api` | `server/Dockerfile`（Go 1.26 + Alpine） | `127.0.0.1:8080`（仅本机） | 业务 API、JWT 鉴权、读写 MySQL、缓存写 Redis、对象存 S3 |
| `nginx` | `web/Dockerfile`（Node 22 构建 + nginx:alpine） | `0.0.0.0:80`（公网） | 提供 Vue SPA 静态资源，反代 `/auth/user/...` 等 API 路径到 `api` |

外部依赖（**部署机之外**，由托管方提供）：

- 外部 MySQL：业务数据唯一可信来源。
- 外部 Redis 6+：公开文章、站点配置等的读取缓存。
- 外部对象存储（MinIO / S3 兼容）：上传的媒体文件。
- 边缘代理（CDN / 反代 / 平台 LB）：终止 TLS，把公网 80/443 流量转发到本机 TCP/80。

```
浏览器 ── HTTPS ──▶ 边缘代理 ── HTTP ──▶ 部署机 :80 (nginx) ──▶ api :8080
                                              │
                                              ├──▶ 外部 MySQL
                                              ├──▶ 外部 Redis
                                              └──▶ 外部 S3 / MinIO
```

## 3. 服务器准备

推荐 Ubuntu 22.04 LTS 或 Debian 12；任何支持 Docker 24+ 的现代发行版均可。

```bash
# 1) 安装 Docker 与 Compose v2
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker "$USER"   # 重新登录后生效
docker --version                  # 期望：Docker 24+
docker compose version            # 期望：v2.x

# 2) 安装常用工具
sudo apt-get update && sudo apt-get install -y git curl wget openssl ca-certificates python3
# python3 或 jq 二选一即可；healthcheck.sh 优先用 python3
```

防火墙放行：TCP 22（运维 SSH）、TCP 80/443（外部流量，由边缘代理或直连决定）。**数据库端口（3306/6379/9000）绝不能从公网可达**。

```bash
sudo timedatectl set-timezone Asia/Shanghai
```

## 4. 外部依赖清单

部署前请向托管方拿到以下信息并记录下来。

### 4.1 外部 MySQL

- 库名：建议 `solitude_blog`（先在托管方手工 `CREATE DATABASE`；应用启动会尝试 `CREATE DATABASE IF NOT EXISTS`，但多数托管账号无 `CREATE` 权限）
- 账号：至少 `SELECT/INSERT/UPDATE/DELETE` 权限
- 记录：`MYSQL_HOST` / `MYSQL_PORT`（默认 3306）/ `MYSQL_USER` / `MYSQL_PASSWORD` / `MYSQL_TLS`（`false` / `true` / `skip-verify`）

### 4.2 外部 Redis 6+

- 建议使用 ACL 用户（如 `app`）
- 记录：`REDIS_HOST` / `REDIS_PORT`（默认 6379）/ `REDIS_USER` / `REDIS_PASSWORD` / `REDIS_TLS`

### 4.3 外部对象存储（MinIO / S3 兼容）

- bucket：在托管方手工建好（如 `blog`）
- 一对 access key / secret key
- 记录：`STORAGE_S3_ENDPOINT`（`https://host:port`）/ `STORAGE_S3_BUCKET` / `STORAGE_S3_REGION` / `STORAGE_S3_USE_SSL` / `STORAGE_S3_PUBLIC_URL`（浏览器拉取对象的 URL，**必须含 `/bucket` 路径**，如 `https://cdn.example.com/blog`）

### 4.4 边缘代理

- DNS 把域名解析到本机公网 IP
- 代理层终止 TLS
- 80/443 流量转发到本机 TCP/80

## 5. 选择部署目录并拉取代码

```bash
sudo mkdir -p /opt/solitude-blog
sudo chown "$USER" /opt/solitude-blog
cd /opt/solitude-blog
git clone <repo-url> .
```

> 不要用 `~` 或 `/root`，避免日后切到非 root 账号时遇到权限问题。后续升级在同目录 `git fetch && git checkout <tag>`，永远保留上一版本以便回滚。

```bash
git log -1 --oneline   # 与仓库管理员确认 tag/commit hash 一致
```

## 6. 生成 5 个密钥

在部署机上执行，输出直接复制到 §7 的 `.env.production`：

```bash
openssl rand -base64 64   # 写入 JWT_ACCESS_SECRET
openssl rand -base64 64   # 写入 JWT_REFRESH_SECRET（必须与上一行不同）
openssl rand -base64 24   # 写入 MYSQL_PASSWORD（先在 MySQL 托管方设置这个值，再回填 .env）
openssl rand -base64 24   # 写入 REDIS_PASSWORD（先在 Redis 托管方设置，再回填 .env）
openssl rand -base64 24   # 写入 ADMIN_PASSWORD
```

- 不要复用任何旧 secret。
- 不要把生成结果落盘到 `~/secrets.txt` 这类易被备份系统抓取的位置；用完即弃。
- 规则约束：JWT secret ≥32 字符、互不相同、不以 `dev-` 开头、不含 `change-me`；`APP_ENV=production` 下 `Validate()` 会拒绝 `admin` 与含 `change-me` 的管理员口令。

## 7. 编辑 `deploy/.env.production`

`deploy/.env.production` 已经是生产模板（含占位符），直接用编辑器修改即可，**不需要 `cp` 一份新的 `.env`**。Compose 启动时通过 `--env-file deploy/.env.production` 显式指向。

必填替换：

| 变量 | 替换为 |
| --- | --- |
| `SITE_BASE_URL` | 真实的 HTTPS 站点 URL（如 `https://www.solitude.love`），影响 RSS / sitemap 内的链接 |
| `WEB_PORT` | `80`（默认）；仅当边缘代理把流量转到非 80 端口时才需要改 |
| `MYSQL_HOST` | §4.1 拿到的 host |
| `MYSQL_DATABASE` | `solitude_blog`（与托管方建库名一致） |
| `MYSQL_USER` / `MYSQL_PASSWORD` | §4.1 的账号 / §6 生成的口令 |
| `MYSQL_TLS` | 同 VPC 内网 `false`；跨公网 `true`（自签名证书用 `skip-verify`） |
| `REDIS_HOST` / `REDIS_PORT` / `REDIS_USER` / `REDIS_PASSWORD` | §4.2 的值 |
| `REDIS_TLS` | 与 `MYSQL_TLS` 同规则 |
| `REDIS_DB` | 多数情况保持 `0`；只有托管方明确划分了逻辑库时才改 |
| `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` | §6 生成的两段 |
| `ADMIN_USERNAME` | 保持 `admin` 或改成期望名 |
| `ADMIN_PASSWORD` | §6 生成的口令 |
| `STORAGE_S3_*` | §4.3 拿到的 MinIO/S3 凭据 |

完成后**强制校验零占位符**：

```bash
grep -n 'REPLACE_WITH' deploy/.env.production
# 期望：无任何输出
```

## 8. 预检 Compose 配置

```bash
cd /opt/solitude-blog
docker compose --env-file deploy/.env.production -f deploy/docker-compose.yml config --quiet
docker compose --env-file deploy/.env.production -f deploy/docker-compose.yml config --services
```

- `config --quiet` 静默退出表示配置无语法错误。
- `config --services` 应**仅**输出 `api` 与 `nginx`。如出现 `mysql` / `redis` 等其他服务，说明仓库还在旧版本或 `deploy/.env.production` 引用了不存在的变量——回到 §7 重新检查。

## 9. 首次启动

```bash
cd /opt/solitude-blog
docker compose --env-file deploy/.env.production -f deploy/docker-compose.yml up -d --build
```

- 首次构建镜像约 5–10 分钟（Go 模块 + `npm ci` 拉取）；后续 `--build` 走本地缓存，秒级完成。
- 启动顺序：`api` 先自检通过 → `nginx` 才启动（`depends_on: condition: service_healthy`）。

## 10. 验收

每条都必须通过：

```bash
# 1. 容器状态
docker compose --env-file deploy/.env.production -f deploy/docker-compose.yml ps
# 期望：api 与 nginx 都是 running，State 列 healthy（start_period 20s/10s 后）

# 2. 健康检查（脚本会校验 JSON 响应 + 清洗 CRLF）
sh deploy/scripts/healthcheck.sh
# 期望末行：healthcheck ok: http://localhost:80/healthz

# 3. nginx 配置语法
docker compose --env-file deploy/.env.production -f deploy/docker-compose.yml exec nginx nginx -t

# 4. AutoMigrate 日志（首次启动应能看到建表输出）
docker compose --env-file deploy/.env.production -f deploy/docker-compose.yml logs api | grep -i migrate
# 期望：能看到 "ALTER TABLE" / "created table" 之类输出；若完全没有，
# 说明启动在更早阶段就失败了，结合 §13 排查。
```

人工验收：

- 浏览器打开 `https://<域名>/`（或经过边缘代理后的 URL）应看到首页。
- `/admin/login` 用 `admin` + §6 生成的 `ADMIN_PASSWORD` 登录成功。
- 后台做一次"新建公告 + 上传一张图"冒烟，确认 S3 凭据有效。

## 11. 首次登录后立即改管理员密码

进入管理后台 → 个人中心 → 修改密码。详见 [`01-deployment-and-backup.md` §1.5](./01-deployment-and-backup.md)。**生产中无任何强制约束，必须人工执行**。若遗忘且无法登录，唯一恢复路径是直接修改 `users` 表的 `password_hash`（用 `bcrypt` 重新哈希新口令）。

## 12. 日常运维命令速查

```bash
# 看日志（最近 100 行 + 持续跟踪）
docker compose --env-file deploy/.env.production -f deploy/docker-compose.yml logs -f --tail=100 api
docker compose --env-file deploy/.env.production -f deploy/docker-compose.yml logs -f --tail=100 nginx

# 滚动重启 api（不影响数据）
docker compose --env-file deploy/.env.production -f deploy/docker-compose.yml up -d --no-deps --force-recreate api

# 拉取新版本
cd /opt/solitude-blog
git fetch
git checkout <new-tag>
docker compose --env-file deploy/.env.production -f deploy/docker-compose.yml up -d --build

# 紧急回滚到上一个 tag
git checkout <previous-tag>
docker compose --env-file deploy/.env.production -f deploy/docker-compose.yml up -d --build

# 看卷使用情况（上传文件）
docker system df -v | grep api-storage

# 备份当前 .env
cp deploy/.env.production deploy/.env.production.bak.$(date +%Y%m%d-%H%M%S)
```

## 13. 常见故障排查

| 现象 | 原因 / 处置 |
| --- | --- |
| `healthcheck.sh` 失败，日志含 `REDIS_HOST is required` | §7 没替换 `REDIS_HOST` 为真实值；或从旧 `.env` 迁移时漏字段 |
| 容器 healthy 但 `curl /healthz` 返回 503 | DB 抖动；按 [`01-deployment-and-backup.md` §4](./01-deployment-and-backup.md) "健康检查"段诊断 |
| Nginx 502 + `docker compose restart api` 后仍 502 | 不应再发生：nginx 现在每次新建连接都重新解析 `api` 域名；若仍出现，看 `docker compose logs nginx | grep "no live upstreams"` |
| 上传图片失败 + 日志 `x509: certificate signed by unknown authority` | MinIO 启用了 TLS 但 api 容器缺 CA；先 `docker compose build --no-cache api` 重打镜像（基础镜像已带 `ca-certificates`） |
| 容器启动后立即退出，日志 `permission denied` on `/app/storage/uploads` | 旧版 `api-storage` 卷是 root 拥有的，UID 10001 写不进去；执行 `docker compose run --rm -u 0 api chown -R 10001:10001 /app/storage/uploads` |
| AutoMigrate 没有日志 | 启动在更早阶段失败；`docker compose logs api | head -50` 排查，多数是 MySQL/Redis 不可达 |
| 健康检查全绿但页面 502 | 检查边缘代理是否正确转发到本机 80；`docker compose exec nginx wget -qO- http://api:8080/healthz` 看 nginx 容器能否解析 api |

## 14. 安全与合规须知

- **不要把含真实凭据的 `deploy/.env.production` 提交到 git**。本仓库的 `deploy/.env.production` 是占位符模板（已纳入版本库），运维应在仓库根 `.gitignore` 排除已填值的本地副本（`*.local` 已忽略）；或者直接编辑 `deploy/.env.production` 然后不提交该文件变更。
- **定期轮换密钥**：见 [`02-secret-rotation.md`](./02-secret-rotation.md)。建议每 90 天一次。
- **上传卷 `api-storage` 备份由本仓库负责**，需运维侧另外配置（详见 [`01-deployment-and-backup.md` §7](./01-deployment-and-backup.md)）。
- **数据库与对象存储的备份责任在托管方**——快照、PITR、保留周期等由托管方按 SLA 提供。
- **单 owner 账号**是当前设计约束，没有"忘记密码"邮件通道，所有管理操作都走这一个 owner；后续若引入多管理员或邀请流程，会在 `reviews/` 跟踪。

## 15. 延伸阅读

- [`runbook/01-deployment-and-backup.md`](./01-deployment-and-backup.md)：详细参考手册（备份、TLS、回滚原则、P1 列表）
- [`runbook/02-secret-rotation.md`](./02-secret-rotation.md)：5 类密钥的轮换步骤
- [`docs/12-maintenance-guide.md`](../../12-maintenance-guide.md)：仓库维护入口
- 仓库根 [`README.md`](../../../README.md)：本机开发与项目结构
