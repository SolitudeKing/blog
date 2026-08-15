# 可改进项（P2 · 10 项）

> 范围：一致性、可维护性、长期扩展。属于排期项，不阻塞当前发布。
> 估算工时合计：6–8 人天。每项独立可实施。

---

## §4.1 移除 Service 层 `s.db == nil` 内存分支

**问题位置**：所有 `server/internal/service/*.go`

**现状**：
启动器 `bootstrap.NewApp` 总是传入 `resources.DB`（除非 `Open` 失败），
内存分支实际不可达，但每个 Service 都保留 `if s.db != nil { ... } else { ... }` 双路径。

**建议方案**：

1. 抽取 `internal/repository` 包，定义接口（如 `ArticleRepository`、`TopicRepository`）；
2. dev / test 用内存实现（`memarticle.New()`），生产用 GORM 实现（`gormarticle.New(db)`）；
3. Service 只依赖接口；移除所有 `s.db == nil` 分支；
4. 内存实现可被 service 测试覆盖，无需绕开 SQL；
5. `bootstrap.NewApp` 始终实例化 GORM 仓储，dev 模式仅替换为 `mem`。

**验收**：

- 全部 Service 不再出现 `s.db` / `if s.db != nil`；
- `service/*_test.go` 不需要 `NewArticleService(nil, nil)` 这种"伪无 DB"模式；
- 引入新接口不会改变响应契约。

---

## §4.2 `c.GetUint64("user_id")` 改为 `MustGet` 防 0 兜底

**问题位置**：[server/internal/handler/article_handler.go:51](server/internal/handler/article_handler.go#L51) · 所有写接口

**现状**：
`gin.Context.GetUint64` 在键不存在时返回 `0`。
当前 Service 收到 `actorID == 0` 会返回 `Unauthorized`，与"必须登录"的语义有歧义。

**建议方案**：

1. `internal/middleware/auth.go` 在 `AuthRequired` 末尾用 `c.Set`；
2. Service 入参把 `actorID uint64` 改为 `actorID *uint64`：
   - 由 middleware 强制注入；
   - Service 入口 `if actorID == nil { return Unauthorized }`；
3. 或者保留 `uint64` 但在中间件 `panic`（仅 dev 模式），保证 prod 不会绕过；
4. 文档明确"未携带 Authorization 头 = 401 而非 403"。

**验收**：

- 任何写接口都不会"通过" 0 号用户；
- dev 自检：去掉 `AuthRequired` 时立即 panic（仅 dev 配置下生效）。

---

## §4.3 404 路由走统一响应壳

**问题位置**：[server/internal/router/router.go](server/internal/router/router.go)

**现状**：
Gin 默认 404 返回空 body + 404，破坏统一响应结构。

**建议方案**：

1. 在 `router.Register` 末尾追加：
   ```go
   engine.NoRoute(func(c *gin.Context) {
       response.Error(c, apperrors.New(apperrors.CodeResourceNotFound))
   })
   engine.NoMethod(func(c *gin.Context) {
       response.Error(c, apperrors.New(apperrors.CodeResourceNotFound))
   })
   ```
2. 文档在 `03-data-api-design.md` §"HTTP 状态码" 明确：404 + 业务码 30000。

**验收**：

- 不存在的路径返回 `{code:30000, message:"resource not found"}`；
- 不影响正常 404 业务接口。

---

## §4.4 `article/info/{id}` 与 `article/detail/{slug}` 数据契约分离

**问题位置**：[server/internal/service/article_service.go:168](server/internal/service/article_service.go#L168)

**现状**：
`Info` 是后台编辑器入口，但当前返回完整 `ArticleDetail`（含 `content_md`、`view_count`、`topic` 完整对象）。
Markdown 可能 MB 级，编辑器不需要阅读量与相邻文章。

**建议方案**：

1. 新增 `ArticleEditItem` 结构，仅保留编辑器需要的字段：
   ```go
   type ArticleEditItem struct {
       ID, Title, Slug, Summary, CoverURL, ContentMD,
       Status string, TopicID uint64, TagIDs []uint64,
       PublishedAt *time.Time, UpdatedAt time.Time,
   }
   ```
2. `Info` 返回 `ArticleEditItem`；
3. `Detail`（公开）保留 `ArticleDetail`，并已包含 `Prev/Next`；
4. 文档在 `03-data-api-design.md` §"文章管理" 与 §"文章详情" 区分两种契约。

**验收**：

- 后台 `Info` 响应 < 100KB（含 markdown）；
- 前台 `Detail` 仍然拿到 `Prev/Next` 与 `view_count`。

---

## §4.5 `topic/list`、`tag/list` 改为统一 `ListBody` 响应壳

**问题位置**：[server/internal/handler/topic_handler.go:25](server/internal/handler/topic_handler.go#L25) · [server/internal/handler/tag_handler.go:25](server/internal/handler/tag_handler.go#L25)

**现状**：
`topic/list` 与 `tag/list` 返回裸数组 `[{...}, {...}]`，与增长型列表响应不一致。
前端若有通用 `useList` 会拿到 `undefined.count`。

**建议方案**：

方案 A（推荐）：
- 维持裸数组，但加 `"version": "v1"` 元字段；
- 文档明确这是"字典列表"，与分页列表区分。

方案 B（破坏性变更）：
- 改为 `ListBody[TopicItem]`，`total=len(items)`；
- 同时升级 `tag/list`；
- 前端配合更新。

**验收**：

- 选定方案后给前端一个迁移窗口；
- 文档在 `03-data-api-design.md` §"响应约定" 增加字典列表定义。

---

## §4.6 API 版本头大小写容忍

**问题位置**：[server/internal/middleware/api_version.go:23-28](server/internal/middleware/api_version.go#L23-L28)

**现状**：
`if version != expected` 大小写敏感；客户端 `V1` 会失败。

**建议方案**：

1. `version := strings.TrimSpace(c.GetHeader("X-API-Version"))`；
2. `if !strings.EqualFold(version, expected) { ... }`；
3. 文档在 `03-data-api-design.md` §"路径与版本" 增加大小写说明。

**验收**：

- `V1` / ` v1 ` / `V1` 均通过；
- 空头仍返回 `20004`。

---

## §4.7 接入访问日志与业务关键路径审计日志

**问题位置**：[server/internal/bootstrap/app.go](server/internal/bootstrap/app.go) · `internal/middleware/access_log.go`（待新增）

**现状**：
只有 panic recovery 才有 `slog.Error`；登录、刷新、上传、删除等关键操作无审计。

**建议方案**：

1. 新增 `middleware.AccessLog()`：
   - 字段：`method / path / status / duration_ms / user_id / request_id / remote_ip`；
   - 输出到 `slog.Info`，按 `path` 采样（如 `/healthz` 跳过）；
2. 业务关键路径单独写 `slog.Info`：
   - `auth/login` 成功 / 失败；
   - `auth/logout` 撤销数量；
   - `article.create / update / delete`；
   - `asset.upload / delete`；
   - `setting.update`；
3. 在 `12-maintenance-guide.md` 标注日志级别 / 字段 / 采样规则；
4. 评估是否把日志通过 filebeat / fluentd 输出到聚合系统。

**验收**：

- 任何请求都有一条结构化 access 日志；
- 关键写操作可在日志中通过 `user_id` 串联完整轨迹。

---

## §4.8 缓存键命名收敛

**问题位置**：[server/internal/cache/keys.go:9-12](server/internal/cache/keys.go#L9-L12)

**现状**：
`prefix = "blog:v1"` 与 `articlePrefix = "blog:v3"` 同进程并存，运营难以理解。

**建议方案**：

1. 全部统一为 `blog:v1`；
2. 文章缓存键保留 `article:list:*` / `article:detail:*` 区分；
3. 在 `cache/keys.go` 顶部加注释说明 `v1` 是 API 版本，不是数据 schema 版本；
4. 提供 `cache.Migrate()` 函数（如有需要）：
   - `RENAME blog:v3:article:list:* blog:v1:article:list:*`；
   - 部署时由运维手动执行一次；
5. 文档在 `03-data-api-design.md` §"缓存与失效" 同步更新。

**验收**：

- `redis-cli --scan --pattern 'blog:*' | grep -v '^blog:v1:'` 输出为空；
- 旧 key 清理脚本可跑通。

---

## §4.9 `ArticleCreateRequest` 与 `ArticleUpdateRequest` 拆分

**问题位置**：[server/internal/service/article_service.go](server/internal/service/article_service.go)

**现状**：
`type ArticleUpdateRequest = ArticleCreateRequest` 类型别名，未来需要差异字段（如禁止 `update` 改 `author_id`）时迁移成本高。

**建议方案**：

1. 拆为两个独立 struct：
   ```go
   type ArticleCreateRequest struct {
       Title, Slug, Summary, CoverURL, ContentMD string
       TopicID uint64
       TagIDs []uint64
       Status string
   }
   type ArticleUpdateRequest struct {
       Title, Slug, Summary, CoverURL, ContentMD string
       TopicID uint64
       TagIDs []uint64
       Status string
   }
   ```
2. 共享校验逻辑放到私有函数 `validateArticleFields(req)`；
3. 文档在 `03-data-api-design.md` §"文章管理" 标注两个接口字段一致。

**验收**：

- 类型不再别名；
- 校验函数被两边共用，无重复 if/else。

---

## §4.10 `healthz` 响应壳与业务码解耦

**问题位置**：[server/internal/handler/health_handler.go:48-56](server/internal/handler/health_handler.go#L48-L56)

**现状**：
`healthz` 把 `data` 当作 `{mysql, redis}` 状态，但同时返回 `code`（业务码）。
Kubernetes / LB 等探活器通常按 HTTP 状态判断，对业务码敏感时容易误判。

**建议方案**：

1. `healthz` 响应固定为：
   ```json
   { "api": "ok", "mysql": "ok|disabled|error", "redis": "ok|disabled|error", "time": "..." }
   ```
2. 不再包含 `code` / `message` 字段；
3. 失败时只通过 HTTP 503 表达（业务码 0 + data.error）；
4. 文档在 `08-deployment-runbook.md` §"健康检查" 明确探活规则。

**验收**：

- 探活器读取 `data.mysql == "ok"`；
- HTTP 503 时不再含 `code: 50001`，仅 `data.mysql == "error"`。
