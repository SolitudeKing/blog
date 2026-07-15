# 编码设计

本文档定义新个人博客系统的编码层目标，用于约束现有 `web`、`server`、`worker`、`deploy` 工程继续演进。它承接架构、产品、API 与路线图文档，重点回答“代码应该如何组织、如何命名、如何分层、如何协作”。

前端视觉值与布局尺寸不在本文重复定义：具体颜色与来源映射以 [雾境主题色调系统](./11-theme-color-system.md) 为准，token 契约和组件以 [UI 设计系统](./09-ui-design-system.md) 为准，页面结构与断点以 [布局模式](./10-layout-patterns.md) 为准。

本文中的目录树和代码片段是目标结构，不代表所有目录与能力已经落地。当前实现状态以源码和 [设计进度追踪](./05-design-progress.md) 为准；尚未存在的 repository、DTO、测试或任务实现必须作为迁移欠账跟踪。

## 编码目标

- 让前端、后端、异步任务和部署脚本可以独立开发、独立测试、清晰协作。
- 避免旧项目中视图、SQL、模板、配置混杂的问题。
- 所有接口、状态、模型、组件 props 和响应结构尽量类型化。
- 所有业务模块遵循统一的目录、命名、错误处理和测试方式。
- 第一阶段优先完成可运行骨架和文章主链路，避免过早抽象。

## 顶层工程结构

建议在仓库根目录创建新系统目录，旧项目保留用于参考和迁移：

```text
Blog/
  blog-mini-serve/       # 旧 Flask 后端，保留参考
  blog-mini-v3/          # 旧 Vue 前端，保留参考
  docs/                  # 设计文档
  web/                   # 新 Vue3 + Vite 前端
  server/                # 新 Go + Gin API
  worker/                # Celery 异步任务
  deploy/                # Docker Compose、Nginx、脚本
  storage/               # 本地开发上传文件目录，生产用挂载卷
```

第一阶段只创建必要骨架：

```text
web/
server/
worker/
deploy/
```

## 代码风格基线

### 通用规则

- 配置通过 `.env` 或配置文件注入，不在代码中写死环境地址、密钥和账号。
- 对外暴露的数据结构必须使用 DTO，不直接返回数据库模型。
- 错误必须走统一错误码，不在业务代码中随手返回字符串。
- 日志必须包含 request id 或 task id。
- 业务时间统一使用 UTC 存储，前端按用户时区展示。
- 文件名和目录名使用小写、短横线或下划线，避免中英文混排。

### Git 与提交

建议提交粒度：

- `docs:` 文档。
- `feat:` 前端、后端或异步任务功能；提交主题只聚焦一个模块。
- `fix:` 修复。
- `refactor:` 重构。
- `test:` 测试。
- `chore:` 构建、配置、脚本。

## 前端编码设计

### 目录结构

```text
web/
  index.html
  package.json
  vite.config.ts
  tsconfig.json
  src/
    app/
      App.vue
      main.ts
      providers.ts
    router/
      index.ts
      public.routes.ts
      admin.routes.ts
      guards.ts
    layouts/
      public/
        PublicLayout.vue
      admin/
        AdminLayout.vue
    pages/
      public/
      admin/
    components/
      base/
      blog/
      admin/
    api/
      http.ts
      types.ts
      modules/
    stores/
    composables/
    styles/
      index.scss
      reset.scss
      tokens/
        _base.scss
        _mist-sea-salt.scss
        _mist-forest.scss
      themes/
        mist.scss
      components/
    types/
    utils/
```

### 路由约定

- 公开站点页面放在 `pages/public`。
- 后台管理页面放在 `pages/admin`。
- 后台路由统一走 auth guard。
- 页面路由和 API 路径不需要保持同名；页面关注用户语义，API 关注业务模块。

示例：

```text
/                     -> public home
/articles/:slug       -> public article detail
/archives             -> public archives
/admin/login          -> admin login
/admin/articles       -> admin article list
/admin/articles/:id   -> admin editor
```

### API client 设计

`api/http.ts` 负责：

- 注入 `X-API-Version: v1`。
- 注入 `Authorization: Bearer <token>`。
- 解析统一响应。
- 处理 `10001 token expired` 并触发刷新。
- 识别游标分页响应。
- 将业务错误转换为前端可识别的 `ApiError`。

当前实现已覆盖版本号与 access token 注入、统一响应解析、单次 refresh、并发刷新复用、原请求重放，以及刷新失败后的凭证清理与登录跳转。尚缺的是这条故障链的前端自动化测试，以及服务端登出后的 token 主动失效能力，见本文“当前实现偏差与迁移清单”。

核心类型：

```ts
export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export interface CursorPage {
  cursor: string
  next_cursor: string
  limit: number
  has_more: boolean
}

export interface ApiListResponse<T> {
  code: number
  message: string
  data: T[]
  page: CursorPage
}
```

### 类型组织

- `types/article.ts`：文章领域类型。
- `types/auth.ts`：登录、token、用户信息。
- `types/setting.ts`：站点配置。
- `types/asset.ts`：媒体资源。
- `api/modules/*.ts`：只写请求函数，不写页面状态。
- `stores/*.ts`：只放跨页面状态。
- 页面内部状态优先放在页面组件或 composable 中。

### 自研组件编码规则

组件命名：

- Vue 文件使用 `BaseButton.vue`、`BaseInput.vue`、`BaseModal.vue`。
- 基础组件样式使用 Mist UI 的 `mist-*` 前缀，例如 `mist-button`、`mist-card`；业务组件使用 `blog-navbar`、`article-card` 等业务语义类。
- 类名结构采用 BEM：元素使用 `__`，变体使用 `--`，瞬时状态使用 `is-*`。
- 业务组件使用 `ArticleCard.vue`、`EditorToolbar.vue`。

组件 API：

- 表单组件统一支持 `modelValue` 和 `update:modelValue`。
- 交互组件必须支持 `disabled`、`loading`、`error` 等状态。
- 可替换内容使用 slot。
- 复杂行为抽成 composable，例如 `useFocusTrap`、`useClickOutside`、`useToast`。

样式规则：

- 组件只消费语义 token，不直接写死主题色。
- 默认使用 `mist-sea-salt + light`，具体色值只允许出现在 `_mist-sea-salt.scss`、`_mist-forest.scss` 的主题映射中。
- 保留 `mist-sea-salt/mist-forest × light/dark` 四种完整 token 组合，主题和明暗模式不得合并为一个布尔值。
- 两套主题必须使用主题选择器隔离，禁止直接同时导入 Mist UI 原始的裸 `[data-mode]` 映射导致后加载主题覆盖前者，也禁止缺失 token 时跨主题回退。
- 焦点态必须清晰，使用 `--focus-ring`。
- 页面组件不得重复声明全局间距、圆角、阴影和 z-index；需要新增值时先更新 [UI 设计系统](./09-ui-design-system.md)。
- 页面骨架优先使用 CSS Grid，组件内部优先使用 Flex；禁止为了主布局新增 Float 或依赖固定视口高度。

### 前端状态管理

Pinia store 建议：

```text
stores/
  auth.ts       # token、用户信息、登录状态
  setting.ts    # 服务端站点配置、全局主题、访客默认模式
  editor.ts     # 编辑器草稿、自动保存状态
```

主题状态规则：

- `setting.ts` 保存 lobby 返回的 `theme` 与默认 `mode`，其中 `theme` 只接受 `mist-sea-salt | mist-forest`。
- `useTheme.ts` 负责把服务端 theme 与最终 mode 映射到 DOM；对公开前台只暴露 `setMode()`、`cycleMode()`，不暴露主题轮换 API。
- 最终 mode 的解析顺序是合法 `blog:mode` > 合法服务端默认 mode > `light`。前台主动切换只写 `blog:mode`，不能回写站点配置。
- 后台保存主题或默认模式后以服务端归一化响应更新当前页面；改变 theme 时保持当前最终 mode 不变。
- 旧 `blog:theme` 浏览器复合对象只迁移其中合法的 mode 一次，其 theme 一律丢弃，不能覆盖后台全局主题。服务端历史数据单独按 `forest` → `mist-forest`、`mist-violet|strawberry|空值|未知值` → `mist-sea-salt` 显式归一，但不得复用旧主题样式文件。

不进入 store 的状态：

- 列表筛选条件。
- 表单局部输入。
- Modal 打开关闭状态。
- 一次性请求 loading。

## 后端编码设计

### 目录结构

```text
server/
  go.mod
  cmd/
    api/
      main.go
  internal/
    bootstrap/
    config/
    router/
    middleware/
    handler/
    service/
    repository/
    model/
    dto/
      request/
      response/
    cache/
    task/
    storage/
    validator/
    errors/
    pagination/
    logger/
  migrations/
  tests/
```

### 依赖方向

后端依赖必须单向流动：

```text
router -> handler -> service -> repository -> model
                    -> cache
                    -> storage
                    -> task
```

约束：

- `handler` 不直接访问数据库。
- `repository` 不处理 Gin context。
- `model` 不依赖 DTO。
- `service` 是事务边界和缓存失效边界。
- `middleware` 不写业务逻辑。

### 模块划分

第一阶段模块：

```text
auth       # 登录、刷新、登出
user       # 当前用户信息
setting    # 站点配置
article    # 文章
category   # 分类
tag        # 标签
asset      # 媒体
notice     # 公告
task       # 异步任务触发和列表
dashboard  # 后台首页统计
```

每个模块建议包含：

```text
handler/{module}_handler.go
service/{module}_service.go
repository/{module}_repository.go
dto/request/{module}_request.go
dto/response/{module}_response.go
```

### 路由编码规则

路由遵循 API 文档：

- 不加 `/api`。
- 不加 URL 版本号。
- 版本号通过 `X-API-Version` middleware 校验。
- 使用业务模块一级前缀。

示例：

```go
r.POST("auth/login", authHandler.Login)
r.GET("user/info", authMiddleware, userHandler.Info)
r.GET("article/list", articleHandler.PublicList)
r.POST("article/create", authMiddleware, articleHandler.Create)
```

### 请求与响应

统一响应类型：

```go
type Response[T any] struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Data    T      `json:"data"`
}

type ListResponse[T any] struct {
    Code    int        `json:"code"`
    Message string     `json:"message"`
    Data    []T        `json:"data"`
    Page    CursorPage `json:"page"`
}
```

错误响应中 `data` 固定为 `null`。

### 错误处理

建议定义：

```text
errors/
  codes.go       # 业务错误码
  app_error.go   # AppError
  mapper.go      # HTTP 状态映射
```

业务层只返回 `AppError` 或普通 error，由统一响应层转换成：

```json
{
  "code": 10001,
  "message": "token expired",
  "data": null
}
```

### 游标分页编码规则

游标分页封装在 `pagination` 包：

```text
pagination/
  cursor.go
  page.go
```

游标内容：

- 排序字段值，例如 `created_at`。
- tie-breaker，例如 `id`。
- 筛选条件 hash，可选。

游标只由服务端生成和解析，前端不理解游标内部结构。

### GORM 模型规则

- 数据库模型只表达表结构和关联关系。
- 字段显式声明 column、index、size。
- 软删除字段只在需要恢复或审计的表中使用。
- 复杂查询写在 repository，不放到 model method。
- 写操作优先使用事务。

### 缓存规则

缓存 key 集中在 `cache/keys.go`：

```go
func ArticleDetailKey(slug string) string
func ArticleListKey(cursor string, limit int, filterHash string) string
func SiteSettingsKey() string
```

缓存失效在 service 层完成：

- 创建/更新/发布/下线文章后失效文章详情、列表、归档、分类标签相关缓存。
- 修改站点配置后失效 `setting/lobby` 缓存。
- 修改公告后失效当前公告缓存。

## Worker 编码设计

### 目录结构

```text
worker/
  pyproject.toml
  app/
    celery_app.py
    config.py
    tasks/
      article_tasks.py
      asset_tasks.py
      search_tasks.py
      sitemap_tasks.py
      backup_tasks.py
      migration_tasks.py
    services/
    repositories/
    utils/
```

### 任务命名

任务名使用稳定字符串：

```text
article.extract_metadata
asset.generate_thumbnail
search.rebuild_index
sitemap.generate
backup.create
migration.import_legacy_blog
```

### Worker 约束

- 任务必须幂等。
- 任务入参只传 ID 或存储 key，不传大对象。
- 任务开始、成功、失败都写日志。
- 长任务定期更新进度。
- 失败任务保留错误原因，供后台任务列表查看。

## 当前实现偏差与迁移清单

本文的目标分层尚未全部落地，以下项目不得因目录骨架存在而标记为完成功能。

| 优先级 | 当前实现 | 目标 |
| --- | --- | --- |
| P0 | `server/internal/` 主要由 handler、service、model 组成，repository 与 request/response DTO 尚未独立 | 按业务风险逐模块迁移，禁止为追求目录完整一次性空拆层 |
| P0 | 多数 `worker/app/tasks/*_tasks.py` 仍返回 `pending-implementation` | 为真实任务补幂等实现、日志、重试、进度和失败原因后再计入里程碑 |
| P1 | `web/package.json` 没有 test 脚本和测试依赖；会话刷新、并发 401、失败回登录页仍只通过类型检查和接口联调验证 | 先覆盖 API client、会话存储、主题控制器、游标分页、路由守卫与关键 Base 组件 |
| P1 | 前端自动刷新链路已闭环，但服务端登出尚未维护 access token 黑名单或 refresh token 撤销状态 | 按 [数据模型与 API 设计](./03-data-api-design.md) 增加 token 主动失效与对应测试 |
| P1 | UI、布局和可访问性仍有目标/实现差距 | 分别按 [UI 设计系统](./09-ui-design-system.md#实现偏差清单) 与 [布局模式](./10-layout-patterns.md#16-当前实现偏差清单) 收敛 |

迁移时以可验证业务目标拆分，不把“创建目录”和“功能完成”视为同一件事。

## 配置与环境变量

建议 `.env.example` 分组：

```text
APP_ENV=development
APP_PORT=8080
APP_API_VERSION=v1

MYSQL_DSN=
REDIS_ADDR=
REDIS_PASSWORD=

JWT_ACCESS_SECRET=
JWT_REFRESH_SECRET=
JWT_ACCESS_TTL_MINUTES=30
JWT_REFRESH_TTL_DAYS=14

STORAGE_DRIVER=local
STORAGE_LOCAL_ROOT=./storage/uploads

CELERY_BROKER_URL=redis://redis:6379/1
CELERY_RESULT_BACKEND=redis://redis:6379/2
```

## 测试设计

### 前端测试

第一阶段最低要求：

- API client 响应解析测试。
- 游标分页 composable 测试。
- 基础组件状态快照或交互测试。
- 路由守卫测试。

### 后端测试

第一阶段最低要求：

- 错误码映射测试。
- JWT 生成、刷新、失效测试。
- 游标编码/解码测试。
- Article service 创建、发布、下线测试。
- Repository 使用测试数据库跑关键查询。

### Worker 测试

第一阶段最低要求：

- 任务 payload 校验。
- 图片缩略图任务幂等性。
- 文章 metadata 提取。
- 失败重试策略。

## 脚手架顺序

建议按以下顺序开始编码：

1. 创建 `server`，完成 config、logger、response、error、health。
2. 创建 `web`，完成 Vite、路由、Sass token、API client。
3. 创建 `deploy/docker-compose.yml`，跑 MySQL 和 Redis。
4. 实现 server migration 和 users/auth。
5. 实现 article/category/tag 模型和 CRUD。
6. 实现 web 登录、公开首页、文章详情。
7. 创建 `worker`，跑通 Celery health task。
8. 接入图片处理、阅读量聚合、索引任务。

## 编码阶段完成标准

每轮编码收口都应满足；未满足项必须进入上方偏差清单或进度文档：

- 工程目录可按本文档直接创建。
- API client 与 Go response 类型对齐。
- 错误码、HTTP 状态码、游标分页有明确实现位置。
- 雾境海盐、雾境青森按 Mist UI 语义结构有独立 Sass 落点，并覆盖四种主题/模式组合。
- 前端、后端、worker 的第一批测试目标明确。
