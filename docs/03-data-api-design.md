# 数据模型与 API 设计

## 数据建模原则

- MySQL 是文章、专题、标签、媒体、配置的唯一可信数据源。
- Markdown 正文存入 MySQL `longtext`，文件导出作为备份能力，不再作为运行时主数据源。
- 图片文件不再以 Base64 存入数据库，数据库只保存元数据和存储 key。
- 专题使用 `articles.topic_id` 建立一对多关系，标签使用 `article_tags` 多对多关系，不再用字符串拼接。
- 所有表统一保留 `created_at`、`updated_at`，需要软删除的表增加 `deleted_at`。
- 公开 URL 使用 `slug`，内部关系使用自增 ID 或雪花 ID。

## 核心实体

```mermaid
erDiagram
  users ||--o{ articles : writes
  topics ||--o{ articles : contains
  articles ||--o{ article_tags : has
  tags ||--o{ article_tags : belongs
  assets ||--o{ articles : cover
  articles ||--o{ article_versions : records
  site_settings ||--|| users : maintained_by
  notices ||--|| users : maintained_by

  users {
    bigint id
    string username
    string password_hash
    string role
    string status
  }

  articles {
    bigint id
    string title
    string slug
    longtext content_md
    text summary
    string status
    bigint topic_id
    bigint author_id
    bigint cover_asset_id
  }

  topics {
    bigint id
    string name
    string label
    string slug
    string description
    string cover_url
    int sort_order
  }

  tags {
    bigint id
    string name
    string slug
  }

  assets {
    bigint id
    string storage_key
    string url
    string mime_type
  }
```

## 表设计建议

### users

管理员用户表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | bigint unsigned | 主键 |
| username | varchar(64) | 登录名，唯一 |
| nickname | varchar(64) | 昵称 |
| email | varchar(128) | 邮箱，可选 |
| password_hash | varchar(255) | 密码哈希 |
| role | varchar(32) | `owner` / `editor` / `viewer` |
| status | varchar(32) | `active` / `disabled` |
| last_login_at | datetime | 最近登录时间 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### articles

文章表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | bigint unsigned | 主键 |
| title | varchar(200) | 标题 |
| slug | varchar(220) | 公开 URL，唯一 |
| summary | text | 摘要 |
| content_md | longtext | Markdown 正文 |
| content_html | longtext | 可选，异步渲染后的 HTML |
| toc_json | json | 可选，异步提取目录 |
| status | varchar(32) | `draft` / `published` / `private` / `archived` |
| topic_id | bigint unsigned | 专题 ID |
| author_id | bigint unsigned | 作者 ID |
| cover_asset_id | bigint unsigned | 封面图 ID，可空 |
| view_count | bigint unsigned | 阅读量 |
| word_count | int unsigned | 字数 |
| reading_minutes | int unsigned | 预计阅读分钟 |
| is_top | tinyint(1) | 是否置顶 |
| is_featured | tinyint(1) | 是否精选 |
| published_at | datetime | 发布时间 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |
| deleted_at | datetime | 软删除 |

索引：

- `uk_articles_slug`
- `idx_articles_status_published_at`
- `idx_articles_topic_status`
- `idx_articles_author`

### topics

专题表。专题用于把文章组织成有独立名称、短 Label 与视觉封面的编辑集合；一篇文章归属一个专题。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | bigint unsigned | 主键 |
| name | varchar(80) | 专题名称 |
| label | varchar(32) | 专题短 Label，用于导航、卡片等紧凑场景 |
| slug | varchar(120) | 唯一 slug |
| description | varchar(500) | 专题描述 |
| cover_url | varchar(500) | 专题封面 URL，可空 |
| sort_order | int | 排序 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |
| deleted_at | datetime | 软删除 |

### tags

标签表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | bigint unsigned | 主键 |
| name | varchar(64) | 标签名 |
| slug | varchar(80) | 唯一 slug |
| description | varchar(255) | 描述 |
| color | varchar(32) | 展示颜色 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |
| deleted_at | datetime | 软删除 |

### article_tags

文章标签关系表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| article_id | bigint unsigned | 文章 ID |
| tag_id | bigint unsigned | 标签 ID |
| created_at | datetime | 创建时间 |

联合唯一：

- `uk_article_tag(article_id, tag_id)`

### assets

媒体资源表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | bigint unsigned | 主键 |
| original_name | varchar(255) | 原始文件名 |
| display_name | varchar(255) | 展示名称 |
| storage_key | varchar(500) | 存储 key |
| url | varchar(500) | 访问 URL |
| thumb_url | varchar(500) | 缩略图 URL |
| mime_type | varchar(100) | MIME |
| ext | varchar(20) | 后缀 |
| size | bigint unsigned | 文件大小 |
| width | int unsigned | 图片宽 |
| height | int unsigned | 图片高 |
| sha256 | char(64) | 内容 hash |
| status | varchar(32) | `ready` / `processing` / `failed` |
| ref_count | int unsigned | 引用数量 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |
| deleted_at | datetime | 软删除 |

### notices

公告表。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | bigint unsigned | 主键 |
| title | varchar(120) | 标题 |
| content | text | 内容 |
| enabled | tinyint(1) | 是否启用 |
| sort_order | int | 排序 |
| starts_at | datetime | 生效时间 |
| ends_at | datetime | 失效时间 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### site_settings

站点配置采用单行列式表，当前固定读取主键 `id = 1`，字段与实际 GORM model 保持一致。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | bigint unsigned | 主键；当前站点配置固定为 `1` |
| site_name | varchar(120) | 站点名称，非空 |
| author | varchar(80) | 作者名称，非空 |
| essay | varchar(500) | 站点签名 |
| theme | varchar(32) | 全局主题 ID，非空 |
| mode | varchar(32) | 访客默认明暗模式，非空 |
| theme_elements_json | json | 按主题 ID 保存的主题元素对象 |
| social_links_json | json | 社交链接对象 |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

外观字段的 API 结构固定为：

```json
{
  "theme": "mist-sea-salt",
  "mode": "light",
  "theme_elements": {
    "mist-sea-salt": {
      "home_latest_empty_description": "第一篇文章正在潮汐之外酝酿。",
      "home_latest_end_text": "已经读到潮汐尽头"
    },
    "mist-forest": {
      "home_latest_empty_description": "第一篇文章正在林雾之间酝酿。",
      "home_latest_end_text": "已经走到林径尽头"
    }
  }
}
```

- `theme` 是站点级主题，只允许 `mist-sea-salt`、`mist-forest`，默认 `mist-sea-salt`。
- `mode` 是无个人偏好访客的默认模式，只允许 `light`、`dark`，默认 `light`。
- `theme_elements` 按主题 ID 持久化；切换 `theme` 只改变当前使用的元素组，不得覆盖另一主题的配置。
- 当前主题元素只包含首页“最近发布”区的两个纯文本字段：`home_latest_empty_description` 最多 160 个 Unicode 字符，`home_latest_end_text` 最多 80 个 Unicode 字符。
- `theme_elements` 不承载 HTML、SVG、CSS、颜色、布局或任意组件配置；前端必须使用文本插值渲染，不能将字段解释为标记或样式。
- 加载、错误、重试、按钮、表单提示与无障碍名称等功能性文案保持全局中性，不随主题变化。
- 服务端不保存单个访客的明暗偏好；访客主动选择由浏览器 `blog:mode` 保存。
- 写接口收到枚举外的值时返回参数错误。服务端历史数据迁移采用显式映射：旧 `forest` → `mist-forest`；旧 `mist-violet`、`strawberry`、空值和其他未知值 → `mist-sea-salt`。该映射只迁移数据，不表示可以复用旧 `_forest.scss`。

### article_versions

文章版本表，用于编辑器回滚。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | bigint unsigned | 主键 |
| article_id | bigint unsigned | 文章 ID |
| title | varchar(200) | 标题快照 |
| content_md | longtext | 正文快照 |
| summary | text | 摘要快照 |
| created_by | bigint unsigned | 操作人 |
| created_at | datetime | 创建时间 |

## API 通用规范

### 路径与版本

接口路径遵循以下规则：

- 不使用全局 `/api` 前缀。
- 不在 URL 中放版本号。
- 版本号统一放到请求头中。
- 按业务模块使用一级前缀，例如 `auth/login`、`user/info`、`setting/lobby`。
- 一级前缀使用单数业务名，例如 `article`、`topic`、`tag`、`asset`。
- 列表接口统一以 `list` 或 `*-list` 结尾；文章、归档等增长型列表使用游标分页，`topic/list`、`tag/list` 作为小规模完整字典列表直接返回数组。

版本请求头：

```http
X-API-Version: v1
```

示例：

```text
POST auth/login
GET user/info
GET setting/lobby
GET article/list?cursor=&limit=20
GET article/detail/my-post
```

兼容策略：

- 除 `/healthz`、`/rss.xml`、`/sitemap.xml` 和 `/uploads/*` 外，客户端必须携带 `X-API-Version`。
- 服务端不支持该版本时返回 HTTP `400`，业务码 `20004`。
- 版本升级时保持旧版本 header 可用一段兼容期，避免前端和后端必须同时发布。

### 统一响应

成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

错误响应：

```json
{
  "code": 10001,
  "message": "token expired",
  "data": null
}
```

响应约定：

- `code` 为业务状态码，`0` 表示成功。
- `message` 为稳定、简短、可读的英文信息，不作为前端逻辑判断依据。
- `data` 成功时为对象、数组或具体值；错误时固定为 `null`。
- 游标分页列表额外返回与 `data` 同级的 `page` 字段。
- 非分页接口（包括 `topic/list`、`tag/list`）不返回 `page` 字段。

分页响应：

增长型列表固定使用游标分页。列表数据固定放在 `data` 数组中，分页信息固定放在与 `data` 同级的 `page` 字段中；专题和标签完整字典列表是明确例外。

```json
{
  "code": 0,
  "message": "ok",
  "data": [],
  "page": {
    "cursor": "",
    "next_cursor": "cursor-value",
    "limit": 20,
    "has_more": true
  }
}
```

游标分页规则：

- 请求参数统一使用 `cursor` 和 `limit`。
- `cursor` 为空表示第一页。
- `limit` 默认 `20`，最大 `100`。
- `next_cursor` 为空且 `has_more=false` 表示没有下一页。
- `cursor` 是服务端生成的不透明字符串，前端只保存和回传，不解析其中内容。
- 不返回 `total`、`page`、`pageSize`、`totalPages`，避免大表计数拖慢列表接口。
- 排序必须稳定，推荐使用 `created_at DESC, id DESC` 或业务明确指定的组合键生成游标。
- 游标失效或格式错误时返回 HTTP `400`，业务码 `20003`。

### HTTP 状态码

HTTP 状态码表达协议层语义，业务码表达业务层语义。错误响应仍保持统一 JSON 结构。

| HTTP 状态码 | 使用场景 |
| --- | --- |
| 200 | 请求成功，包含查询、更新、删除、业务动作成功 |
| 201 | 创建成功，例如上传资源或创建文章 |
| 400 | 参数格式错误、缺少版本头、游标非法、JSON 解析失败 |
| 401 | 未登录、token 无效、token 过期 |
| 403 | 已登录但权限不足 |
| 404 | 资源不存在或路由不存在 |
| 409 | 资源冲突，例如 slug 重复、状态冲突 |
| 413 | 请求体或上传文件超过限制 |
| 415 | 不支持的 Content-Type 或文件类型 |
| 422 | 参数格式正确但业务校验失败 |
| 429 | 触发限流 |
| 500 | 服务内部错误 |
| 502 | 上游依赖异常，例如反向代理后的依赖服务不可用 |
| 503 | 服务不可用，例如数据库或 Redis 不可用 |

### 业务错误码

错误码按范围划分：

- `0`：成功。
- `10xxx`：认证、授权、账号相关。
- `20xxx`：请求参数、协议、分页、版本相关。
- `30xxx`：资源与业务状态相关。
- `40xxx`：限流、安全策略、上传策略相关。
- `50xxx`：服务端和外部依赖错误。

| code | 含义 |
| --- | --- |
| 0 | ok |
| 10000 | unauthorized |
| 10001 | token expired |
| 10002 | invalid token |
| 10003 | refresh token expired |
| 10004 | forbidden |
| 10005 | invalid credentials |
| 10006 | account disabled |
| 20000 | invalid request |
| 20001 | missing required field |
| 20002 | invalid parameter |
| 20003 | invalid cursor |
| 20004 | unsupported api version |
| 20005 | malformed json body |
| 30000 | resource not found |
| 30001 | resource conflict |
| 30002 | duplicate slug |
| 30003 | invalid resource state |
| 30004 | referenced resource exists |
| 40000 | rate limited |
| 40001 | upload file too large |
| 40002 | unsupported file type |
| 40003 | operation too frequent |
| 50000 | internal server error |
| 50001 | database unavailable |
| 50002 | cache unavailable |
| 50004 | storage unavailable |

## 公开 API

### 健康、订阅与站点地图

```text
GET healthz
GET rss.xml
GET sitemap.xml
```

`healthz` 返回 API、MySQL 与 Redis 状态，并在 MySQL 或 Redis 异常时使用非 2xx 状态；RSS 与 sitemap 只包含已发布且满足公开约束的文章。

### 站点配置

```text
GET setting/lobby
```

返回：

- 站点名称。
- 作者。
- 签名。
- 头像。
- 社交链接。
- SEO 默认配置。
- 全局主题 `theme`：`mist-sea-salt` 或 `mist-forest`。
- 访客默认模式 `mode`：`light` 或 `dark`。
- 完整主题元素映射 `theme_elements`，键为受支持的主题 ID；前台根据响应中的 `theme` 读取对应元素组。

`theme` 与 `mode` 必须始终返回合法值，`theme_elements` 必须始终返回海盐和青森两个完整元素组。前端以 lobby 的 `theme` 作为全局主题，并从 `theme_elements[theme]` 读取首页最近发布区文案；`mode` 只在浏览器不存在合法 `blog:mode` 时作为默认值。公开接口只读，不提供主题元素写入口。

### 首页文章列表

```text
GET article/list?cursor=&limit=20&topic=distributed-systems&tag=go&keyword=gin
```

只返回 `published` 状态文章。

### 文章详情

```text
GET article/detail/{slug}
```

返回：

- 文章标题。
- Markdown 或 HTML。
- TOC。
- 专题轻量引用（只包含 `id`、`name`、`label`、`slug`）。
- 标签。
- 上一篇/下一篇。
- SEO 信息。

### 归档

```text
GET article/list?cursor=&limit=50
```

当前没有独立的 `archive/list` 接口。归档页复用公开文章游标列表，并在前端按 `published_at` 的年份、月份分组；需要更多数据时继续使用 `next_cursor` 加载。

### 专题与标签

```text
GET topic/list
GET tag/list
GET article/list?cursor=&limit=20&topic={topic_slug}&tag={tag_slug}
```

专题列表响应额外返回按已发布文章计算的 `article_count`，并按 `sort_order ASC, created_at DESC, id DESC` 稳定排序；专题筛选复用文章列表接口且只返回 `published` 状态文章。

### 公告

```text
GET notice/active
```

### 搜索

```text
GET search/article?keyword=redis&cursor=&limit=20
```

搜索范围包含文章标题、摘要、专题名称/Label 与标签名称；只返回 `published` 状态文章。

## 后台 API

### 认证

```text
POST auth/login
POST auth/refresh
POST auth/logout
GET  user/info
```

### 仪表盘

```text
GET dashboard/summary
```

摘要响应包含文章状态、阅读量、专题与标签数量、最近文章、热门文章、公告状态，以及按专题聚合的文章分布。当前没有独立的最近文章列表或异步任务列表接口。

### 文章管理

```text
GET    article/manage-list?cursor=&limit=20&status=&topic=&tag=&keyword=
POST   article/create
GET    article/info/{id}
PUT    article/update/{id}
DELETE article/delete/{id}
GET    article/version-list/{id}
```

文章列表过滤参数 `topic`、`tag` 均使用 slug；创建与更新请求使用 `topic_id` 关联专题，不再接受 `category_id`。发布、撤回与归档通过 `article/update/{id}` 修改 `status`，创建和更新会在同一事务中自动记录版本；当前只提供最近 20 条版本的只读列表，不提供独立发布、手动创建版本或恢复版本接口。公开与后台文章响应统一返回 `topic_id`，以及字段为 `topic` 的轻量引用 `{id,name,label,slug}`。

### 专题与标签

```text
GET    topic/list
POST   topic/create
PUT    topic/update/{id}
DELETE topic/delete/{id}

GET    tag/list
POST   tag/create
PUT    tag/update/{id}
DELETE tag/delete/{id}
```

专题创建与更新使用同一字段契约：

```json
{
  "name": "分布式系统",
  "label": "DISTRIBUTED",
  "slug": "distributed-systems",
  "description": "关于一致性、消息与系统边界的系列文章",
  "cover_url": "/uploads/topics/distributed-systems.webp",
  "sort_order": 10
}
```

`name`、`label`、`slug` 必填；`label` 最多 32 个 Unicode 字符，`slug` 全局唯一，`cover_url` 可为空，`sort_order` 默认为 `0`。`article_count` 是列表响应的聚合字段，不写入 `topics`。固定三个专题允许维护展示名称、说明、封面与排序，但修改其稳定 `label/slug` 或删除专题会返回 `30001 resource conflict`；其他专题删除前必须检查文章引用，存在引用时返回 `30004 referenced resource used`。

### 媒体库

```text
GET    asset/list?cursor=&limit=40&keyword=&mime=
POST   asset/upload
PUT    asset/update/{id}
DELETE asset/delete/{id}
GET    asset/reference-list/{id}?cursor=&limit=20
```

### 公告

```text
GET    notice/manage-list?cursor=&limit=20
POST   notice/create
PUT    notice/update/{id}
DELETE notice/delete/{id}
```

### 站点配置

```text
GET setting/detail
PUT setting/update
```

`setting/lobby`、`setting/detail` 和 `setting/update` 的外观字段使用同一响应契约；后台更新请求也使用同样的 `theme_elements` 结构：

```json
{
  "theme": "mist-sea-salt",
  "mode": "light",
  "theme_elements": {
    "mist-sea-salt": {
      "home_latest_empty_description": "第一篇文章正在潮汐之外酝酿。",
      "home_latest_end_text": "已经读到潮汐尽头"
    },
    "mist-forest": {
      "home_latest_empty_description": "第一篇文章正在林雾之间酝酿。",
      "home_latest_end_text": "已经走到林径尽头"
    }
  }
}
```

- 后台可修改 `theme` 与访客默认 `mode`，公开 API 没有主题写入口。
- `theme` 只接受 `mist-sea-salt`、`mist-forest`；`mode` 只接受 `light`、`dark`。
- `theme_elements` 以主题 ID 为隔离边界：保存一个主题的文案不得覆盖另一主题已经持久化的文案。
- 为兼容旧客户端，请求省略 `theme_elements` 或传 `null` 时保留数据库中的当前映射；只有数据库也没有有效映射时才使用内置默认值。
- 显式提交局部映射时，未出现的主题键保留当前值；一旦提交某个主题对象，该对象内缺失、`null` 或去除首尾空白后为空的字段回退到该主题内置默认值。
- 非空自定义值必须是字符串；`home_latest_empty_description` 最多 160 个 Unicode 字符，`home_latest_end_text` 最多 80 个 Unicode 字符。超限返回参数错误，非字符串类型作为畸形 JSON 请求拒绝。
- 主题元素始终按纯文本保存和插值渲染，即使内容包含类似 HTML、SVG 或 CSS 的字符也不得解释为标记或视觉代码；颜色、图形和布局仍由受版本控制的主题 token 与组件负责。
- 更新成功后必须失效 `blog:v1:site:settings`，响应返回归一化后的完整站点配置。
- 修改全局主题不得删除或覆盖任何浏览器中的 `blog:mode`；访客本地偏好仍具有更高优先级。

## JWT 设计

Access Token claims：

```json
{
  "sub": "1",
  "username": "admin",
  "role": "owner",
  "jti": "uuid",
  "exp": 1710000000,
  "iat": 1709996400
}
```

当前实现：

- Access Token 与 Refresh Token 的有效期分别由环境变量配置。
- 刷新接口校验 Refresh Token 后签发新令牌。
- 当前没有完整的服务端 Refresh Token 持久化、轮换和撤销记录；登出不能被视为终止全部刷新会话。
- 服务端撤销、密码变更失效和多设备会话审计列为 P1，见 [项目审查](./13-project-review.md)。

## 缓存与失效

| 缓存键 | 内容 | TTL | 失效时机 |
| --- | --- | --- | --- |
| `blog:v1:site:settings` | 公开站点配置，含完整 `theme_elements` | 10 分钟 | 修改站点配置或主题元素 |
| `blog:v1:article:detail:{slug}` | 文章详情 | 5 分钟 | 修改/发布/下线文章 |
| `blog:v1:article:list:*` | 文章列表 | 5 分钟 | 修改文章、专题、标签 |

读取不含 `theme_elements` 的旧站点设置缓存时，服务端必须按海盐、青森默认值补齐完整映射；包含部分合法字段的旧缓存保留已有值并补齐缺项。任一站点设置更新都会删除该缓存，后续写回统一使用完整结构。

## 当前数据库升级兼容

当前启动迁移仍需要处理已经运行过早期版本的数据库：

1. 若存在中间版本 `categories` 表和 `articles.category_id`，按分类 slug 幂等写入 `topics`，并回填 `articles.topic_id`。
2. 若存在完整默认专题 `Notes / Notes / notes`，转换为“雾里拾笺 / NODES / nodes”；匹配不完整时不擅自修改用户同名数据。
3. 补齐固定的三个专题，并恢复被软删除的固定专题。
4. 对仍没有专题的文章回填“雾里拾笺”，再建立文章到专题的外键结构。

上述逻辑是当前系统版本升级能力，不属于已删除的一次性旧博客迁移工具。它必须保持幂等，并在所有持久化环境建立显式迁移基线前继续保留。
