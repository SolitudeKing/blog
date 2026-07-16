# 项目维护指南

本文是后续维护 Solitude Blog 的首要入口，记录稳定目录边界、数据不变量和常用检查方法。实现细节以源码为准；发现文档与源码不一致时，应在同一变更中修正文档。

## 目录职责

| 目录 | 维护边界 |
| --- | --- |
| `web/` | 公开站点、管理后台、路由、API client、状态管理、主题与组件样式 |
| `server/cmd/api/` | API 进程入口，不放业务逻辑 |
| `server/internal/handler/` | HTTP 参数解析、鉴权上下文和响应转换 |
| `server/internal/service/` | 业务规则、事务编排、缓存失效和领域校验 |
| `server/internal/model/` | GORM 模型与跨层稳定领域常量 |
| `server/internal/database/` | 数据库连接、结构升级和初始化数据 |
| `deploy/` | Compose、Nginx、健康检查与数据库运维脚本 |
| `docs/` | 当前契约、维护方法、待办和验收记录 |
| `demo/` | 不参与运行的静态设计蓝图；用于校准正式前端的视觉、布局与交互方向 |

仓库根目录的 `.env`、`migration-output/`、备份、上传文件、依赖目录和构建产物都是本地状态，不得进入版本历史或 Docker 构建上下文。
`demo/` 虽然被排除在 Docker 构建上下文之外，但属于需要纳入版本历史的设计资产，清理旧工程时不得删除。

## 三专题契约

| 名称 | Label | Slug | 排序 |
| --- | --- | --- | --- |
| 雾里拾笺 | `NODES` | `nodes` | 1 |
| 微光造物 | `CODE` | `code` | 2 |
| 风过留痕 | `JOTTING` | `jotting` | 3 |

维护规则：

- `label` 与 `slug` 是跨前后端稳定标识；只修改展示名称时不得连带修改它们。
- 一篇文章必须且只能关联一个有效专题。
- 初始化逻辑按 slug 幂等补齐三个专题，不创建示例文章和示例标签。
- 早期完整默认值 `Notes / Notes / notes` 可迁移为 `NODES/nodes`；`categories/category_id` 到 `topics/topic_id` 的升级兼容逻辑仍是当前数据库升级链路的一部分，不属于可随旧工程删除的代码。
- 若未来确需新增或删除专题，必须同时评估数据库迁移、文章归属、前端目录、缓存键、RSS/sitemap 和已有链接，不能只改界面常量。

## 关键数据不变量

- MySQL 是文章、专题、标签、媒体元数据和站点配置的唯一可信来源；数据库不可用时应明确报错，不应返回内存示例数据。
- 公开列表、搜索、RSS 与 sitemap 只能暴露满足公开状态约束的文章。
- 文章 slug、专题 slug 和媒体 `storage_key` 必须保持唯一；修改稳定标识前先建立重定向或迁移策略。
- 固定三个专题不能经由后台删除或修改 `label/slug`；删除其他专题前必须确认没有文章引用，删除标签前应先解除文章关联。
- 删除媒体前必须检查文章正文和媒体引用关系；数据库记录与磁盘文件的处理需要保持可恢复。
- 缓存不是事实来源。写入成功后应失效相关缓存，排障时先区分数据库事实与缓存副本。
- 实际 `.env`、JWT 密钥、数据库口令和备份文件不得写入日志、文档或提交记录。

## 常用验证命令

### 后端

```bash
cd server
gofmt -w <本次修改的 Go 文件>
go test ./...
```

### 前端

```bash
cd web
npm ci
npm run typecheck
npm run build
```

### 部署与脚本

```bash
sh -n deploy/scripts/backup-mysql.sh deploy/scripts/healthcheck.sh deploy/scripts/restore-mysql.sh
docker compose --env-file .env -f deploy/docker-compose.yml config --quiet
docker compose --env-file .env -f deploy/docker-compose.yml up -d --build
docker compose --env-file .env -f deploy/docker-compose.yml exec nginx nginx -t
sh deploy/scripts/healthcheck.sh
```

### 引用与仓库状态

```bash
git status --short
git diff --check
```

## 测试数据清理原则

1. 先备份数据库，并记录清理前数量。
2. 用稳定 slug、标题、创建时间和关联关系确认测试对象，不跨环境照搬历史自增 ID。
3. 按“文章关联 → 文章 → 无引用标签/专题 → 无引用媒体文件”的顺序清理。
4. 清理专题、标签或媒体前再次查询引用；名称相似不构成删除依据。
5. 检查站点配置是否被测试导入覆盖，并清理 Redis 中相关缓存。
6. 清理后验证公开列表、后台列表、RSS、sitemap、文章详情和媒体访问。
7. 迁移报告、SQL 临时文件和数据库备份保持在忽略目录中；确认不再需要后再删除。

## 中文注释约定

- 注释说明设计原因、兼容边界、事务顺序、安全假设和不直观的失败处理。
- 命名已经清楚表达的赋值、循环和条件不重复注释。
- 兼容代码要注明保留条件与可移除前提，避免后续把当前升级能力误当旧代码删除。
- TODO 必须包含明确问题或验收条件；纯设想写入路线图，不散落在业务代码中。
- 修改行为时同步修改失效注释，错误注释比缺少注释更危险。

## 未来扩展入口

- 新公开页面：`web/src/router/public.routes.ts`、`web/src/pages/public/` 与对应页面样式。
- 新后台能力：管理路由、后台页面、API module，以及服务端 handler/service/router 同步落地。
- 新数据字段：先明确迁移与兼容策略，再修改 model、service、API 类型和表单。
- 新主题：先扩充语义 token 和主题映射，再实现组件状态与可访问性验收。
- 异步任务：只有出现可量化的耗时或可靠性需求时再引入；必须同时提供任务生产者、幂等消费者、重试、监控和失败恢复，不能恢复空占位服务。
- 对象存储或 CDN：保持媒体 `storage_key` 语义稳定，通过存储抽象迁移，不在文章正文中写环境相关磁盘路径。
