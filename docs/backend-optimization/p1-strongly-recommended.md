# 强烈建议优化项（P1 · 8 项）

> 范围：影响生产质量（性能 / 安全 / 规范）但暂不阻塞发布的 8 项。
> 估算工时合计：5–7 人天。每项独立可实施，互不阻塞。

---

## §3.1 公开文章列表缓存失效改用版本号

**问题位置**：[server/internal/cache/keys.go:22](server/internal/cache/keys.go#L22) · [server/internal/service/article_service.go:740-759](server/internal/service/article_service.go#L740-L759)

**现状**：
`invalidateArticleCaches` 通过 `SCAN MATCH blog:v3:article:list:*` 收集全部列表 key 后
整体 `DEL`。在大流量或长 TTL 场景下，SCAN 会遍历整个键空间并阻塞 Redis 线程。

**建议方案**：

1. 在 Redis 引入 `blog:v1:article:list:version`（整数计数器）作为缓存 key 的前缀；
2. `ArticleListKey(page, pageSize, filterHash)` 改写为：
   ```text
   blog:v1:article:list:{version}:{page}:{pageSize}:{filterHash}
   ```
3. 失效时仅 `INCR blog:v1:article:list:version`，旧 key 自然过期（保留 5 分钟 TTL）；
4. `TopicService` 与 `TagService` 中的 `invalidateArticleCaches` 同样改为版本号方案；
5. 注意：`feed`、`search` 暂不走缓存，不受影响。

**验收**：

- 失效耗时从 O(总 key 数) 降到 O(1)；
- 旧 key 在 TTL 过期前仍可被读，但下一个写入会自动覆盖；
- 在 100K key 的 Redis 实例上 `INCR` < 1ms，SCAN 通常 > 100ms。

---

## §3.2 媒体删除采用 "软标记 + 对象清理" 两阶段

**问题位置**：[server/internal/service/asset_service.go:241-247](server/internal/service/asset_service.go#L241-L247)

**现状**：
Delete 路径先 `db.Delete(&row)` 再 `_ = s.store.Delete(...)`。一旦对象删除失败（403、网络抖动），DB
记录已消失，孤儿文件无法再被定位。

**建议方案**：

1. `assets` 表新增 `storage_state` 字段：`ready / orphan / missing`；
2. Delete 流程：
   - Step 1：`UPDATE assets SET storage_state='orphan'` 在事务内执行；
   - Step 2：调用 `s.store.Delete(...)`；
   - Step 3：成功 → `DELETE FROM assets`；失败 → `storage_state='orphan'` 留待清理；
3. 增加 `asset/purge-orphans` 后台接口（管理端可手动触发 / 定时任务）；
4. 后台 `asset/list` 增加按 `storage_state` 过滤的能力；
5. 文档在 `03-data-api-design.md` §"媒体库" 标注 `storage_state` 字段。

**验收**：

- DB 删除失败时对象不删除（与原行为一致）；
- 对象删除失败时 DB 保留记录，至少不会静默丢文件；
- `purge-orphans` 接口幂等，连续执行不会重复写对象存储。

---

## §3.3 媒体引用列表 `reference-list` 实现引用关系

**问题位置**：[server/internal/service/asset_service.go:264-276](server/internal/service/asset_service.go#L264-L276) · [docs/01-decisions/lld/01-data-and-api-design.md:582](../01-decisions/lld/01-data-and-api-design.md#L582)

**现状**：
API 文档已经定义了 `asset/reference-list/{id}`，但 Service 只返回空数组。
前端表格渲染时会显示"暂无引用"造成误判。

**建议方案**：

1. 新增表 `article_assets(article_id, asset_id, role)`：
   - `role` 取值 `cover` / `inline` / `topic_cover` / `avatar`；
   - 联合唯一 `(article_id, asset_id, role)`。
2. 文章保存时扫描 `content_md` 中的 `![alt](/uploads/...)` 与 `cover_url`，
   同步写入 `article_assets`，并 `UPDATE assets SET ref_count = ...`；
3. `asset/reference-list` 实现：按 `asset_id` JOIN `articles` 取标题 / 状态 / 链接；
4. 引入 `article_assets` 后的删除检查：
   - `assets.delete` 必须先查 `article_assets` 中是否存在记录，否则报 `30004 referenced resource used`；
5. 给现有 `assets.ref_count` 字段一个定时校正任务（与文章保存事件双写）。

**验收**：

- 上传一张图片 → 创建文章并在正文中引用 → `reference-list` 至少能返回 1 条；
- 删除被引用的图片 → `Delete` 返回 `30004`；
- 修改正文不再引用 → 重新保存后引用列表自动消失。

---

## §3.4 Refresh Token 轮换收紧（同 §2.5 拓展）

**问题位置**：[server/internal/service/auth_service.go](server/internal/service/auth_service.go)

**现状**：
P0 已经把撤销表与 jti 接上，并让 `Refresh` 撤销旧 jti。但仍存在两个缺口：
1. access token 没有撤销路径；
2. 撤销表没有 TTL 清理任务，长期累积。

**建议方案**：

1. 撤销表 `revoked_refresh_tokens` 增加定期清理：
   - 增加一个 `auth/purge-expired-revocations` 管理端接口；
   - 启动时由 cron-style goroutine 每小时执行一次 `DELETE WHERE expires_at < now()`；
2. access token 撤销：复用现有 `LogoutRevoke`，但同时校验 `expires_at`；
3. 引入"单设备最近一次"语义：
   - 撤销同 user_id 的所有旧 refresh jti（保留最新一次）；
   - 在 `Refresh` 时先查同 user_id 的 jti 是否已有 N（默认 5）个有效记录，超过则撤销最旧的；
4. 文档更新 `../04-operations/reviews/01-project-review.md` 的 P1 / P0 状态。

**验收**：

- 撤销表行数 < `活跃用户数 × 最大设备数`；
- 单设备重复刷新不会无限签发新 token；
- access token 撤销后立即失效（中间件加查询或短 TTL 黑名单）。

---

## §3.5 `Notice.ManageList` 排序参数化

**问题位置**：[server/internal/service/notice_service.go:227-233](server/internal/service/notice_service.go#L227-L233)

**现状**：
后台公告管理列表固定按 `created_at DESC` 排序，前台 `Notice.Active` 按 `sort_order ASC, created_at DESC`。
编辑完一条公告后，前台展示顺序与后台列表不一致；管理端也无法按 `sort_order` 排序。

**建议方案**：

1. `NoticeListQuery` 增加 `OrderBy string`（取值 `sort_order` / `updated_at` / `created_at`）；
2. handler 接受 `order_by` query 参数，默认 `created_at`，非法值回退到默认；
3. 写接口创建公告时，若 `sort_order=0` 自动追加到当前最大值 + 1，避免新公告被默认置顶；
4. 文档在 `03-data-api-design.md` §"公告" 增加 `order_by` 说明；
5. 索引评估：为 `notices(sort_order, created_at, id)` 加联合索引，配合 ORDER BY 使用。

**验收**：

- 后台按 `sort_order` 排序后能看到预期顺序；
- 前台 Active 顺序与公开 spec 一致；
- 新建公告不会"意外"置顶。

---

## §3.6 `search/article` 全文搜索优化

**问题位置**：[server/internal/service/search_service.go:52-73](server/internal/service/search_service.go#L52-L73)

**现状**：
SQL 用 `LIKE '%keyword%'` 命中 `articles.content_md LONGTEXT`，加 EXISTS 子查询；
文章量超过 1K 后单次查询就会扫全表。

**建议方案**：

1. 短期（必做）：
   - `keyword` 长度 < 2 时返回空；
   - `total` 字段上限 1000（截断），并在响应里加 `truncated: true` 提示；
   - 把 `title`、`summary`、`topic.name/label/slug`、`tag.name/slug` 这几列长度限制在 200–500；
2. 中期（推荐）：
   - 加 `article_search_index(article_id, source, content)` 表，仅镜像 title + summary + topic + tag 字符串；
   - 上 `idx_source_keyword` 前缀索引（MySQL 8.0+ 函数索引），支持 `LIKE 'keyword%'`；
3. 长期（评估）：
   - 评估接入 Meilisearch / Typesense / SQLite FTS5 等轻量方案；
   - 写接口 `article.create / update` 触发同步索引；
   - 读路径始终走索引，回退到 SQL 只在索引缺失时。

**验收**：

- 1 万篇文章下 LIKE 查询 P95 < 200ms；
- 索引覆盖度 ≥ 95%；
- 搜索相关性可配置（标题权重 > 标签权重）。

---

## §3.7 限流中间件（`/auth/*` + `/asset/upload`）

**问题位置**：[server/internal/router/router.go](server/internal/router/router.go) · 文档未定义

**现状**：
`/auth/login`、`/auth/refresh`、`/asset/upload` 三个高敏感接口没有任何限流；
攻击者可无限次尝试。`40000 rate limited` 业务码已定义但无实际使用。

**建议方案**：

1. 新增 `internal/middleware/ratelimit.go`：
   - 基于 `golang.org/x/time/rate` token bucket；
   - 支持 `(IP × 路由)` 与 `(user_id × 路由)` 两维度；
2. 配置：
   - `RATE_LIMIT_LOGIN_IP` 默认 `5/min`；
   - `RATE_LIMIT_LOGIN_USER` 默认 `3/min`；
   - `RATE_LIMIT_UPLOAD_USER` 默认 `30/min`；
   - `RATE_LIMIT_REFRESH_IP` 默认 `20/min`；
3. 触发时返回 `40000`，HTTP 429；
4. 文档在 `03-data-api-design.md` §"HTTP 状态码" 与 §"业务错误码" 补充。

**验收**：

- 同 IP 5 秒内连续 6 次登录 → 第 6 次返回 429；
- 高频上传在 1 分钟内超过 30 次 → 返回 429，错误码 `40000`；
- 正常用户无感知。

---

## §3.8 CORS 中间件支持 Cookie 场景

**问题位置**：[server/internal/middleware/cors.go:21-24](server/internal/middleware/cors.go#L21-L24)

**现状**：
dev 模式固定 `Access-Control-Allow-Origin: *` + 不带 `Allow-Credentials`。
未来若引入 cookie session（如主题模式持久化或后台刷新会话），浏览器仍会拒绝。

**建议方案**：

1. 引入 `CORS_ALLOWED_ORIGINS`（逗号分隔白名单）；
2. dev 模式下也允许通过配置指定；
3. 当 `Origin` 在白名单内时回显具体 origin 而非 `*`，并补 `Access-Control-Allow-Credentials: true`；
4. OPTIONS 预检仍直接 204 放行；
5. 文档在 `08-deployment-runbook.md` §"反向代理" 增加 CORS 边界说明。

**验收**：

- 同源 Cookie 场景下浏览器不再拒绝；
- 非白名单 origin 不会拿到 `Access-Control-Allow-Credentials` 头；
- 不影响现有 dev 模式 Vite proxy 调用。
