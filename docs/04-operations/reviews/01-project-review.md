# 2026-07-16 项目审查

## 审查范围

本轮审查覆盖仓库基线、旧资产边界、文章专题、测试数据、前后端可维护性、部署安全、备份恢复和文档可信度。结论按“已修复”和“待后续”区分；没有完成实机或浏览器验证的事项不会记为完成。

## 本轮已修复

- 运行基线收敛为 `web`、`server`、`deploy`，文档由 `docs` 维护；`demo` 明确保留为设计蓝图，仅移除无调用的异步占位代码和一次性旧博客迁移脚手架。
- 根 README 与文档索引改为当前系统口径，补充固定三专题契约和维护入口。
- Compose 不再启动空异步进程；API 调试端口仅绑定 `127.0.0.1`，敏感口令不再提供可直接运行的弱默认回退，并为常驻服务增加重启策略。MySQL / Redis 已切到外部托管服务，不再由本仓库 Compose 启动。
- 新增根 `.dockerignore`，避免 `.env`、Git 元数据、旧工程、迁移产物和本机依赖进入 Web 构建上下文。
- Nginx 增加媒体上传大小限制和不依赖 TLS 的基础安全响应头；是否启用 HTTPS 仍由真实部署环境决定。
- Server 镜像依赖层同时复制 `go.mod` 与 `go.sum`，减少依赖缓存与校验不一致。
- MySQL 备份使用最小文件权限、临时文件、失败即停、`gzip -t` 完整性检查和同目录原子落盘；恢复要求显式确认，并在导入压缩包前检查 gzip 完整性。（注：本轮已删除备份恢复脚本，外部托管后由托管方承担备份责任。）
- 保留当前数据库从 `categories/category_id` 升级到 `topics/topic_id` 的兼容逻辑，并明确其不是旧博客运行代码。
- 已按 slug、标题、删除状态和关联指纹清除导入测试与 smoke 数据；数据库只保留三个正式专题，现有站点外观设置未重置。
- 媒体上传拒绝未净化 SVG，并按检测到的 MIME 生成安全扩展名；写文件或落库失败时清理孤儿文件。

## 验证状态

- 已通过：`go test ./...`。
- 已通过：`npm run build`（包含 `vue-tsc --noEmit`）。
- 已通过：Markdown 相对链接扫描、旧路径引用扫描、Shell `sh -n` 与 `git diff --check`。
- 已通过：修复后 API 使用现有 MySQL/Redis 在独立端口启动，`/healthz`、三专题列表、空文章与空标签响应均符合预期。
- 当前机器没有 Docker CLI；Compose 配置展开、全容器启动、Nginx 配置实机检查、健康检查和备份恢复演练仍待部署环境完成。
- 待真实浏览器：四种主题组合、响应式布局、键盘路径、读屏语义和截图基线。

## 2026-07-28 配置跟进

- MySQL 与 Redis 连接配置已收敛为原子环境变量，由 Go 配置层组合 DSN 和地址；环境模板不再保留无人消费的 Celery URL。
- API 容器改为显式注入所需变量，不再通过 `env_file` 接收 MySQL root 密码等无关配置。

## 2026-08-13 外部托管化

- 部署文档与 Compose 切到"外部托管"口径：Compose 仅保留 `api` 与 `nginx` 两个服务，删除自带 `mysql` / `redis` 服务、命名卷与 healthcheck；`.env.production` 与根 `.env.example` 同步去除 `MYSQL_ROOT_PASSWORD` 等无关字段，补 `MYSQL_TLS` / `REDIS_TLS` / `REDIS_USER` 透传。
- 删除 `backup-mysql.sh` / `restore-mysql.sh`，把备份责任显式划归数据库托管方；`healthcheck.sh` 调整为仅校验本仓库持有的容器，外部依赖的健康通过 API `/healthz` 端点间接观察。
- 后端新增 `MYSQL_TLS` / `REDIS_TLS` 字段（取值 `false` / `true` / `skip-verify`），分别注入 MySQL DSN 与 Redis client 的 TLS Config；并补齐配套单元测试。

## 2026-08-15 后端 API 修复

本轮针对 2026-08-15 后端 API 审查识别的 7 项必修（P0）实施修复；强烈建议（P1）与可改进（P2）见 `../../backend-optimization/` 子目录。

### 本轮已修复

- **公开文章列表排序**：当 `onlyPublished=true` 时 `ORDER BY articles.published_at IS NULL, articles.published_at DESC, articles.id DESC`，后台管理列表继续按 `created_at DESC, id DESC`（编辑视角）。修复了首页 / 归档在草稿被发布后顺序错乱的问题。`sortArticles` / `articleListOrderClause` 集中维护两套排序键。
- **文章详情补 prev/next**：`ArticleDetail` 新增 `prev` / `next` 字段，按同专题 + `published_at` 在已发布集合中选取 `LIMIT 1`。邻居使用 `ArticleNeighbor` 轻量引用，避免冗余 `topic` / `tags` 数据。修复了前端拿不到上一篇/下一篇、与设计文档 §"文章详情" 契约不一致的问题。
- **PublishedAt 改指针类型**：`ArticleItem.PublishedAt` 由 `time.Time` 改为 `*time.Time`，未发布文章（草稿/归档）序列化为 `null`，避免前端 `Invalid Date`。
- **auth/login 内存模式 bcrypt 缓存**：把无 DB 模式下的 bcrypt hash 计算移到 `AuthService.memoryAdminUser`，启动时一次生成，登录时直接比对，不再每次请求消耗 50–150ms。
- **Refresh Token 撤销机制**：新增 `RevokedRefreshToken` 模型 + AutoMigrate；TokenClaims 增加 `JTI`；`Refresh` 校验撤销表并在签发新对时把旧 jti 写入；`Logout` 把当前 access / refresh jti 一并写入。文档 [01-data-and-api-design.md §JWT 设计](../../01-decisions/lld/01-data-and-api-design.md#jwt-设计) 的 P1 要求已落地主体能力，剩余完善见 [`p1-strongly-recommended.md §3.4`](../../backend-optimization/p1-strongly-recommended.md)。
- **Dashboard 文章计数合并为单查询**：`fillArticleCounts` 由 6 次独立的 COUNT / SUM 合并为一次 `SUM(CASE WHEN status = ...)` 聚合，文章表 ≥10K 行时仪表盘加载时延显著下降。
- **Topic List Join 抽取共享常量**：`TopicService.List` 把 SELECT 列表 / GROUP BY / JOIN 条件抽到 `topicListSelectColumns` / `topicListGroupColumns` / `joinPublishedArticlesOnTopic`，避免日后 `article_count` 在两侧漂移。`Dashboard.fillTopicStats` 显式声明独立的 join（计数范围含草稿），通过注释解释为何不复用同一常量。

### 验证状态

- 已通过：`go build ./...`、`go test ./...`（含新增的 `TestAuthService*` / `TestArticleService*` 用例）。
- 新增 `server/internal/model/revoked_refresh_token.go` 与 `server/internal/service/auth_jti_test.go`、`article_detail_test.go`，覆盖 jti 唯一性、prev/next 内存行为、排序键、logout 无 DB 行为。
- 待真实数据：在独立 MySQL/Redis 上验证 `RevokedRefreshToken` 表 AutoMigrate、Refresh 撤销查询、Topic/Dashboard join 行为；浏览器验证 `published_at = null` 在前台 article 列表的展示。

## P1：上线前应继续完成

### 认证与会话

- Refresh Token 需要服务端持久化、轮换和撤销机制；登出和密码变更应能终止既有刷新会话。
- 复核 Cookie、跨域、代理头和限流策略，并在真实域名与 HTTPS 边界下验收。

### 媒体安全

- 删除媒体前建立可靠的文章引用检查，明确正文链接、封面、缩略图和磁盘文件之间的处理顺序。
- 为图片增加完整解码、像素上限与压缩炸弹防护，并规划缩略图生成策略；当前已有大小、MIME 和安全扩展名边界。

### 备份与恢复（外部托管边界）

- MySQL / Redis 备份责任由托管方承担（快照 / PITR / 异地副本 / 保留周期）。与托管方约定自动备份频率与季度恢复演练。
- `api-storage` 上传卷仍由本仓库持有，需补充独立备份（卷快照或对象存储镜像），并与数据库恢复点无强对应关系。
- 恢复演练：先在一次性或隔离环境完成，再应用到生产；恢复前先停 API、保存当前数据库与上传卷快照。
- 当前应用启动会执行 `CREATE DATABASE IF NOT EXISTS`，托管账号通常无 `CREATE` 权限。建议先在托管方建库；后续将引入 `MYSQL_SKIP_CREATE_DB` 开关。

### 部署验收

- 在具备 Docker 的目标机器执行 Compose 展开与启动，确认密码特殊字符、Redis 认证、健康状态、持久化卷和重启策略。
- 明确生产 TLS 终止、域名、证书续期、真实客户端 IP 和安全响应头的最终责任层。

## P2：持续质量建设

- 为前端补充单元测试、组件测试和关键流程端到端测试；当前构建与类型检查不能替代行为回归。
- 在隔离 MySQL/Redis 上补充文章事务、标签关联、软删除 slug 冲突与旧 schema 升级测试；现有内存单测和真实启动冒烟不能替代数据库集成回归。
- 建立首页、文章、归档、搜索、About 与后台关键页面的真实浏览器截图基线。
- 完成键盘全流程、焦点管理、对比度和读屏抽查，记录浏览器与辅助技术版本。
- 逐步用显式数据库迁移替代只依赖 AutoMigrate 的生产结构变更，并记录可回滚边界。
- 增加结构化日志、关键指标、错误告警和备份任务监控。
- 仅在同步链路出现明确瓶颈后评估异步任务系统；引入时应作为完整能力实现，不恢复占位模块。

## 后续验收原则

每个待办只有在源码、自动化检查和对应运行环境验证都完成后才能改为“已修复”。涉及真实数据的清理、迁移和恢复必须先备份，并保留可核对的数量、时间和结果记录。
