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
REDIS_DB=0
REDIS_TLS=false
```

从旧配置升级时，必须先将 `MYSQL_DSN`、`REDIS_ADDR` 拆成上述原子变量并移除旧变量。Go 加载器暂时兼容只提供旧变量的非 Compose 环境，但新旧变量同时存在会拒绝启动；当前 Compose 部署只接受原子变量。

不要提交实际 `.env`。占位符不能直接用于生产。

## 1.5 首次部署 · 管理员引导

应用启动时会以 `ADMIN_USERNAME` / `ADMIN_PASSWORD` / `ADMIN_NICKNAME` 创建唯一一个 owner 账号。`APP_ENV=production` 下 `Validate()` 会拒绝默认 `admin/admin` 与任何含 `change-me` 的口令；`APP_ENV=development` 则会静默放行默认值，**仅供本地开发使用**。

首次部署推荐流程：

1. 在 `cp .env.example .env` 之后立刻设置 `ADMIN_PASSWORD=<强口令>`，至少 12 位、含大小写字母与数字。
2. 启动后用 `ADMIN_USERNAME` / `ADMIN_PASSWORD` 登录 `/admin/login`。
3. **首次登录后立即在管理后台修改密码**——这一步在生产中无任何强制约束，依赖人工执行；轮换步骤见 [runbook/02-secret-rotation.md](./02-secret-rotation.md)。
4. 媒体走 S3（详见 §10）；owner 在后台改密不影响已上传对象的公开 URL（公开 URL 走 `STORAGE_S3_PUBLIC_URL`，与签名密钥无关）。
5. 若遗忘了 owner 密码且无法登录，唯一恢复路径是直接修改 `users` 表的 `password_hash`（应用启动时不会重建 owner）。

> 单一 owner 账号是当前设计约束：没有"忘记密码"邮件通道，没有自助注册，所有管理操作都走这一个账号。后续若引入多管理员或邀请流程，会作为 P1 在 `reviews/` 跟踪。

## 3. 展开并启动

仓库提供两条等价路径，按运维场景二选一：

| 路径 | 适用场景 | 启动命令 |
| --- | --- | --- |
| **制品挂载（默认）** | 日常升级，运行时镜像已构建 | 开发机执行 `.\deploy\scripts\deploy.ps1 -Host <user>@<host>`；服务器端 `docker compose --env-file .env -f deploy/docker-compose.runtime.yml up -d` |
| **源码构建（回退）** | 紧急修复 / 运行时镜像缺失 | 服务器端 `docker compose --env-file .env -f deploy/docker-compose.yml up -d --build` |

预检配置（任意一条路径都建议先做）：

```bash
docker compose --env-file .env -f deploy/docker-compose.yml config --quiet
docker compose --env-file .env -f deploy/docker-compose.yml config --services
```

第二条命令应仅输出 `api` 与 `nginx` 两个服务。如出现 `mysql` 或 `redis`，说明 Compose 仍包含旧服务定义，请检查 `deploy/docker-compose.yml` 是否已切到当前外部化版本。

启动顺序：API 容器等待自身健康检查通过后启动 Nginx；API 直接通过 env 注入的外部地址连接 MySQL / Redis，**假设外部服务已就绪且网络可达**。宿主机的 API 调试端口（`APP_PORT`，默认 8080）仅绑定 `127.0.0.1`；外部流量只进入 Nginx。

制品路径的细节（制品目录、scp / rsync、`-Fallback` 标志、何时重建运行时镜像、故障排查）见 [`03-artifact-based-deploy.md`](./03-artifact-based-deploy.md)。

## 4. 健康检查

```bash
sh deploy/scripts/healthcheck.sh
docker compose --env-file .env -f deploy/docker-compose.yml exec nginx nginx -t
docker compose --env-file .env -f deploy/docker-compose.yml logs api | grep -i migrate
```

验收点：

- `api`、`nginx` 均为 running/healthy。
- `GET /healthz` 返回 API、外部 MySQL 与外部 Redis 的健康结果；任一持久化依赖异常时返回非 2xx，使脚本和容器健康检查失败。
- `logs api` 中能看到 GORM AutoMigrate 的输出（首次启动会建表；后续启动若 schema 未变更则只输出"nothing to migrate"）。**若完全没有 AutoMigrate 相关日志，说明启动过程在更早的阶段就失败了**，需结合 `docker compose logs api` 排查。
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
- 上传文件：生产走 S3 / MinIO（§10），不再依赖 `api-storage` 命名卷；该卷仅在开发或临时回退到 local 驱动时使用。

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
| 上传文件 | **外部 S3 / MinIO**：`STORAGE_S3_ENDPOINT` / `STORAGE_S3_BUCKET` / `STORAGE_S3_PUBLIC_URL` 在 `deploy/.env.production` 配齐；S3 凭据轮换见 [runbook/02-secret-rotation.md §5](./02-secret-rotation.md#5-storage_s3_access_key--storage_s3_secret_key) | **托管方**（与 MySQL 同口径） |
| 容器日志 | Docker `json-file` | 单文件 10 MB，最多 5 个文件 |

> 生产环境**不要**依赖 `api-storage` 命名卷。该卷在 `STORAGE_DRIVER=local` 或 `STORAGE_S3_*` 缺失时才会被实际写入；保留它是为了本地开发与紧急回退方便，**生产部署中应当是空的**。如果发现该卷在生产中占用磁盘，先确认 `STORAGE_DRIVER=s3` 与全部 `STORAGE_S3_*` 已生效，再 `docker volume rm <project>_api-storage`。

Compose 使用 `restart: unless-stopped`。部署机仍需监控磁盘、S3 桶配额 / 访问日志、证书、容器健康和恢复演练时间。

## 8. 回滚原则

- 应用回滚前确认新版本是否已执行不可逆数据变更。
- 不直接删除命名卷解决启动问题；S3 模式下命名卷本来就是空的。
- 数据恢复前保存当前数据库与 S3 桶的访问凭据快照（`STORAGE_S3_*` 在 `.env` 中的版本）。
- 回滚完成后核对三专题、文章数量、媒体引用（按公开 URL 抽查几张图）、站点设置、RSS 和 sitemap。

## 9. 后续 P1（不在本仓库持有）

- `ensureDatabase` 创建数据库权限：当前应用启动会执行 `CREATE DATABASE IF NOT EXISTS`。若托管账号无 `CREATE` 权限，请先在托管方建库；后续将引入 `MYSQL_SKIP_CREATE_DB` 开关作为代码层补项。
- TLS 自定义名字：当前仅支持内置 `true` / `skip-verify`；自定义 `tls.Config`（客户端证书 / 自定义 CA 池）作为后续 P1。
- 显式迁移版本：当前 `database.go` 使用 GORM AutoMigrate，建议逐步迁移到版本化迁移工具（golang-migrate 等）。
- S3 健康检查纳入 `/healthz`：当前 `health_handler.go` 只检查 MySQL / Redis，S3 桶不可达时 healthcheck 仍会通过；后续加 `HeadBucket` 探针。
- 资源限制与编排：当前 `docker-compose.yml` 未声明 `deploy.resources` / `security_opt`，仅给 `api` 显式了 `user: 10001:10001`；进一步在部署环境规格明确后再评估。

## 10. S3 / 对象存储运维

生产上传路径走 S3 / MinIO；`api-storage` 命名卷在 S3 模式下不被写入。

### 10.1 桶存在性与权限

应用启动时会调用 `BucketExists`（`server/internal/storage/factory.go`），失败立即终止启动。常见原因：

- 桶未在 MinIO / S3 托管方先建好；`STORAGE_S3_BUCKET` 与托管方桶名不一致。
- access key 缺少 `s3:ListBucket` / `s3:GetObject` / `s3:PutObject` / `s3:DeleteObject` 中至少前 3 项。
- `STORAGE_S3_ENDPOINT` 协议写错（`http` vs `https`），或者 endpoint 写成了管理控制台地址而非 API 地址。
- 桶 region 与 `STORAGE_S3_REGION` 不一致（MinIO 通常忽略 region，但 AWS S3 会强制校验）。

### 10.2 公开 URL 与签名 URL 的差异

- `STORAGE_S3_ENDPOINT`：后端写入 / 删除 / 列出对象用的 API 地址。**不要**把它配到 `STORAGE_S3_PUBLIC_URL`。
- `STORAGE_S3_PUBLIC_URL`：浏览器直拉的只读 URL，**必须**含 `/bucket` 路径（如 `https://cdn.example.com/blog`）。后端在生成图片 `<img src>` 时填的就是这个值。

公开 URL 通常通过 CDN（Cloudflare / 七牛 / 阿里云 CDN）走，可以加缓存与图片处理；本仓库不强制 CDN，但 `STORAGE_S3_PUBLIC_URL` 直接指向 MinIO 公开桶也能用。

桶策略需要允许匿名 `s3:GetObject`；否则浏览器会拿到 403 / `AccessDenied`。

### 10.3 MinIO 自签名证书

`STORAGE_S3_USE_SSL=true` 但 endpoint 证书是自签的：

- 临时方案：把 `STORAGE_S3_USE_SSL` 改为 `false`，或换受信证书（Let's Encrypt 等）。
- 根治：自定义 CA 池支持——当前实现**不支持**在 `.env` 里配置额外 CA 路径（详见 §9 TLS 自定义名字 P1）。

基础镜像已经预装 `ca-certificates`（见 `server/Dockerfile`），所以使用受信 CA 签发的证书不会出 `x509: certificate signed by unknown authority`；如果仍报，先 `docker compose build --no-cache api` 拉取新基础镜像。

### 10.4 凭据轮换

明确指向 [runbook/02-secret-rotation.md §5](./02-secret-rotation.md#5-storage_s3_access_key--storage_s3_secret_key)。注意：

- 旧 AccessKey 在 S3 模式下**不要**立即删除：保留 24 小时供仍在跑旧实例的容器完成最后一次上传。
- 轮换期间已分发的公开 URL 仍可访问（对象 URL 与 AccessKey 无关），所以**轮换 S3 凭据不需要批量改 HTML**。

### 10.5 监控与排错

```bash
# 拉取应用最近 5 分钟日志，过滤 S3 相关报错
docker compose --env-file deploy/.env.production -f deploy/docker-compose.yml logs --since=5m api \
  | grep -iE 's3|presign|putobject|getobject|no such bucket|access denied|signaturedoesnotmatch'
```

更深一步可在 MinIO 控制台 / CloudWatch 看桶的请求日志（按 `s3:GetObject` / `s3:PutObject` 状态码筛 4xx / 5xx）。

### 10.6 紧急回退到 local 驱动

如果 S3 在短期内无法恢复（证书过期 / 桶被误删 / 凭据被吊销），可以临时把 `STORAGE_DRIVER` 切回 `local`：

```bash
# 1. 编辑 deploy/.env.production：STORAGE_DRIVER=local
# 2. api-storage 卷会立即接管；旧上传对象因为留在 S3 上而无法访问，
#    本地驱动只接管"切换之后的"上传。
# 3. 滚动重启 api。
# 4. 恢复 S3 后切回 s3 即可，期间上传的新文件可由脚本迁回（仓库暂不提供迁回脚本）。
```

切回 `local` 是临时手段，**不是**备份策略；切回后请同步通知所有依赖"媒体走 S3 URL"的运营与第三方集成。