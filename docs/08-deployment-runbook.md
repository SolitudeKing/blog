# 部署与备份运行手册

本文说明当前 Compose 部署方式。仓库内 Nginx 只监听 HTTP；生产 HTTPS、域名和证书由实际边缘代理或托管平台负责。

**核心约定**：MySQL / Redis 均由**外部托管服务**提供，不在本仓库的 Compose 内启动。本仓库的 Compose 仅打包 `api` 与 `nginx` 两个容器；外部数据库 / 缓存必须由运维在部署机之外提供。

## 1. 前置条件（外部依赖清单）

部署前必须确认以下条件全部满足：

- 外部 MySQL 实例已就绪，账号具备库级 `SELECT / INSERT / UPDATE / DELETE` 权限；
  - 若未在托管方手动 `CREATE DATABASE`，账号需额外具备 `CREATE` 权限（应用启动时会尝试 `CREATE DATABASE IF NOT EXISTS`，托管账号通常只授予库级权限，所以**强烈建议先在托管方建库**）
- 外部 Redis 6+ 实例已就绪，建议使用 ACL 用户（如 `app`）并配置所需命令与键空间权限；
- API 容器能通过网络访问外部 MySQL / Redis（生产推荐内网或 VPC 私网，避免数据库直接暴露公网）；
- 若使用公网 / 跨区域访问，请按 §5 配置 `MYSQL_TLS` / `REDIS_TLS`。

## 2. 准备环境变量

```bash
cp .env.example .env
```

必须填写 MySQL / Redis / JWT / 管理员凭据，并将 `APP_ENV` 设为 `production`，把 `SITE_BASE_URL` 改为真实 HTTPS 站点地址。
`.env.example` 的 `SITE_BASE_URL=http://localhost` 对应完整 Compose 的 Nginx 入口；只有在宿主机运行 Vite 开发服务器时才改为 `http://localhost:5173`。

数据库与缓存使用原子变量，Go 配置层会据此组合完整的 MySQL DSN 与 Redis 地址，**不在 `.env` 中人工拼接或同步组合字段**：

```env
# MySQL（外部托管）
MYSQL_HOST=<REPLACE_WITH_MYSQL_HOST>
MYSQL_PORT=3306
MYSQL_DATABASE=solitude_blog
MYSQL_USER=blog
MYSQL_PASSWORD=<强密码>
MYSQL_TLS=false

# Redis（外部托管，6+ ACL）
REDIS_HOST=<REPLACE_WITH_REDIS_HOST>
REDIS_PORT=6379
REDIS_USER=app
REDIS_PASSWORD=<强密码>
REDIS_TLS=false
```

从旧配置升级时，必须先将 `MYSQL_DSN`、`REDIS_ADDR` 拆成上述原子变量并移除旧变量。Go 加载器暂时兼容只提供旧变量的非 Compose 环境，但新旧变量同时存在会拒绝启动；当前 Compose 部署只接受原子变量。

不要提交实际 `.env`。占位符不能直接用于生产。

## 3. 展开并启动

先检查最终配置：

```bash
docker compose --env-file .env -f deploy/docker-compose.yml config --quiet
docker compose --env-file .env -f deploy/docker-compose.yml config --services
```

第二条命令应仅输出 `api` 与 `nginx` 两个服务。如出现 `mysql` 或 `redis`，说明 Compose 仍包含旧服务定义，请检查 `deploy/docker-compose.yml` 是否已切到当前外部化版本。

再启动服务：

```bash
docker compose --env-file .env -f deploy/docker-compose.yml up -d --build
```

启动顺序：API 容器等待自身健康检查通过后启动 Nginx；API 直接通过 env 注入的外部地址连接 MySQL / Redis，**假设外部服务已就绪且网络可达**。宿主机的 API 调试端口（`APP_PORT`，默认 8080）仅绑定 `127.0.0.1`；外部流量只进入 Nginx。

## 4. 健康检查

```bash
sh deploy/scripts/healthcheck.sh
docker compose --env-file .env -f deploy/docker-compose.yml exec nginx nginx -t
```

验收点：

- `api`、`nginx` 均为 running/healthy。
- `GET /healthz` 返回 API、外部 MySQL 与外部 Redis 的健康结果；任一持久化依赖异常时返回非 2xx，使脚本和容器健康检查失败。
- 首页、登录、公开文章列表、RSS 与 sitemap 可访问。
- 一次受控媒体上传未触发代理层 1 MB 默认限制。

## 5. 传输加密（TLS）

外部托管服务跨网络访问时，必须在传输层加密。仓库内代码通过两个开关启用：

| 变量 | 取值 | 含义 |
| --- | --- | --- |
| `MYSQL_TLS` | `false` / `true` / `skip-verify` | 默认 `false`（明文）。`true` 启用 TLS 并校验服务端证书；`skip-verify` 启用 TLS 但跳过证书校验（仅自签名证书场景） |
| `REDIS_TLS` | `false` / `true` / `skip-verify` | 同上 |

约束：

- 当前实现仅支持内置名字（`true` / `skip-verify`），不接受自定义 TLS Profile 名字。若需 `client cert` 或自定义 CA，请走后续 P1（`MYSQL_TLS=<name>` 注册自定义 `tls.Config`）。
- `skip-verify` 等价于不做证书校验，**不应在生产环境使用**，仅供临时调试或自签名证书测试。
- 当 `MYSQL_TLS=true` 且外部 MySQL 使用自签名证书时，请使用 `skip-verify` 或在外部 MySQL 侧配置受信 CA。
- 启用 TLS 后仍建议把数据库放在内网，TLS 不能替代网络层防护。

## 6. 备份与恢复

### 责任边界

MySQL / Redis 均由外部托管服务提供，**备份责任由托管方承担**（快照、PITR、异地副本、保留周期等通常由托管方按 SLA 提供）。本仓库不持有备份 / 恢复脚本；不要在生产文档里再引用 `backup-mysql.sh` / `restore-mysql.sh`。

### 推荐做法

- 与托管方约定自动备份频率与保留窗口（建议每日全量 + 短期 binlog/PITR）。
- 备份恢复演练每季度至少一次，结果记入运维记录。
- 手动导出辅助：本机安装 `mysqldump` / `redis-cli` 后，从 `.env` 读取连接信息即可对外部实例做一次性导出。
- 上传文件：`api-storage` 命名卷仍由本仓库持有，请额外配置卷或对象存储备份（参考 §7）。

### 恢复流程

恢复会覆盖现有数据。**先在一次性或隔离环境演练，不要直接把生产库作为测试目标**。

```bash
docker compose --env-file .env -f deploy/docker-compose.yml stop api
# 由托管方控制台触发 PITR / 还原快照；
# 或在本机使用 mysqldump 导入：
mysql -h "$MYSQL_HOST" -P "$MYSQL_PORT" -u "$MYSQL_USER" -p"$MYSQL_PASSWORD" "$MYSQL_DATABASE" < backup.sql
docker compose --env-file .env -f deploy/docker-compose.yml up -d api nginx
sh deploy/scripts/healthcheck.sh
```

## 7. 持久化与日志

| 数据 | 位置 | 备份责任 |
| --- | --- | --- |
| MySQL | **外部托管实例** | 托管方（PITR / 快照 / 异地副本） |
| Redis | **外部托管实例** | 托管方；本仓库仅作为可重建缓存，不作为业务事实来源 |
| 上传文件 | `api-storage` volume | **本仓库负责**：需补充独立备份（与外部数据库的恢复点无强对应关系；丢失后通过内容运营恢复） |
| 容器日志 | Docker `json-file` | 单文件 10 MB，最多 5 个文件 |

Compose 使用 `restart: unless-stopped`。部署机仍需监控磁盘、上传卷备份结果、证书、容器健康和恢复演练时间。

## 8. 回滚原则

- 应用回滚前确认新版本是否已执行不可逆数据变更。
- 不直接删除命名卷解决启动问题。
- 数据恢复前保存当前数据库与上传卷快照。
- 回滚完成后核对三专题、文章数量、媒体引用、站点设置、RSS 和 sitemap。

## 9. 后续 P1（不在本仓库持有）

- `ensureDatabase` 创建数据库权限：当前应用启动会执行 `CREATE DATABASE IF NOT EXISTS`。若托管账号无 `CREATE` 权限，请先在托管方建库；后续将引入 `MYSQL_SKIP_CREATE_DB` 开关作为代码层补项。
- Redis 静默默认值：`config.go` 在无 `REDIS_ADDR` 时回落 `localhost:6379`，外部化场景下可能掩盖配置遗漏。建议保持 `REDIS_HOST` 为必填，代码侧收紧作为后续。
- TLS 自定义名字：当前仅支持内置 `true` / `skip-verify`；自定义 `tls.Config`（客户端证书 / 自定义 CA 池）作为后续 P1。
- 显式迁移版本：当前 `database.go` 使用 GORM AutoMigrate，建议逐步迁移到版本化迁移工具（golang-migrate 等）。