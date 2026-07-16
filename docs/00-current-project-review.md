# 当前项目评审

> 术语说明：本文记录旧博客原貌，因此保留旧系统的“分类”名称；新系统将该概念迁移为“专题（Topic）”，不再沿用 Category 作为当前领域模型。

## 项目结构概览

当前项目由两个主要目录组成：

- `blog-mini-serve`：Flask + SQLite 后端，同时包含 Jinja2 后台管理模板、静态资源、文章 Markdown 文件与图片文件。
- `blog-mini-v3`：Vue CLI + Vue 3 前台项目，使用 Element Plus、Axios、Vue Router、Vuex、md-editor-v3。

旧项目已经完成了一个个人博客系统的雏形：前台展示文章和个人信息，后台管理文章、分类、标签、图床和站点配置。它的价值在于功能边界比较明确，后续重构时可以直接复用这些业务概念。

## 旧项目功能结构

### 公开前台

旧前台的核心页面包括：

- 首页 `/home`
  - 文章列表。
  - 客户端分页。
  - 右侧作者卡片、公告卡片、词云卡片。
- 文章详情 `/article?id=...`
  - Markdown 内容渲染。
  - 右侧目录导航。
  - 阅读量自增。
- 归档 `/archives`
  - 按时间线展示文章。
- 关于 `/about`
  - 从站点配置读取 `aboutContent` 并以 Markdown 渲染。
- 全局布局
  - Header 菜单。
  - Footer 占位。
  - 回到顶部。

### 后台管理

旧后台基于 Flask 模板，主要页面包括：

- 登录页 `/login`
  - 上传密钥文件，校验通过后写入 session。
- 后台首页 `/`
  - 系统信息。
  - 文章数量、分类数量、标签数量、图片数量。
  - 站点签名。
- 文章管理 `/articlelist`
  - 文章列表。
  - 跳转新增/编辑文章。
- 文章编辑 `/articleeditor`
  - Markdown 编辑。
  - 选择分类、标签、状态。
  - 支持新增与修改。
- 分类管理 `/typesmanage`
  - 分类增删改查。
- 标签管理 `/labelsmanage`
  - 标签增删改查。
- 图床管理 `/picbedmanage`
  - 图片上传、列表、删除、重命名。
- 博客配置 `/blogconfig`
  - 作者、签名、社交链接、收款码、关于内容。

### 后端接口模块

Flask 通过 Blueprint 拆分为：

- `Article`
  - `POST /article/addarticle`
  - `POST /article/delarticle`
  - `POST /article/updatearticle`
  - `POST /article/getarticle`
  - `GET /article/articletotals`
  - `GET /article/articlecontent`
- `ArticleTypes`
  - `POST /articletypes/addtype`
  - `POST /articletypes/deltype`
  - `POST /articletypes/updatetype`
  - `GET|POST /articletypes/gettypes`
- `ArticleLabels`
  - `POST /articlelabels/addlabel`
  - `POST /articlelabels/dellabel`
  - `POST /articlelabels/updatelabel`
  - `GET|POST /articlelabels/getlabels`
- `PicBed`
  - `POST /picbed/`
  - `POST /picbed/delpic`
  - `POST /picbed/renamepic`
  - `GET /picbed/<file_id>`
  - `GET /picbed/piclist`
- `BlogConfig`
  - `GET /blogconfig/getconfig`
  - `POST /blogconfig/setconfig`
- `Other`
  - `GET /other/getosinfo`

## 旧数据模型

SQLite schema 中包含以下表：

- `Articles`
  - 文章标题、内容、分类、标签、创建时间、更新时间、浏览量、状态。
- `PicBed`
  - 图片 ID、名称、后缀、Base64、创建时间、状态。
- `BlogConfig`
  - 作者、签名、访问量、社交链接、收款码、关于内容。
- `BlogNotice`
  - 公告标题、公告内容、创建时间、更新时间。
- `ArticleTypes`
  - 分类名称、创建时间、更新时间。
- `ArticleLabels`
  - 标签名称、创建时间、更新时间。

旧系统中文章内容同时存在于数据库和 `articles/*.md` 文件中，图片同时存在于数据库 Base64 和 `pics/` 文件中。这种“双写”思路便于早期开发，但在新系统中应该改为单一可信来源。

## 可以保留的业务资产

- 文章、标签、归档、关于、公告、图床、站点配置这些业务模块值得保留；旧分类数据迁移为新系统专题。
- 已有 Markdown 文章可以迁移到新系统。
- `article_status` 中的草稿、公开、私有概念值得保留并扩展。
- 作者卡片、公告卡片、标签云、文章目录这些前台组件方向值得保留。
- 后台首页统计面板、文章编辑器、图床管理、站点配置是新版管理台的核心入口。

## 主要技术债

### 前端

- 当前项目使用 Vue CLI，不符合新技术栈中的 Vite。
- 大量组件依赖 Element Plus，与“新版不使用 UI 组件库”的目标冲突。
- 主要代码仍是普通 `setup()` 写法，未系统使用 `<script setup lang="ts">`。
- 缺少 TypeScript 类型约束，请求响应、文章对象、配置对象都没有明确类型。
- 页面布局固定宽度较多，移动端适配不足。
- API loading 与错误提示放在 Axios 全局拦截器中，容易造成列表、局部刷新、静默请求体验不佳。
- `BlogConfig` 依赖 localStorage 缓存，缺少过期策略和降级处理。
- 部分分页、过滤在客户端完成，不利于数据增长。

### 后端

- 使用字符串拼接 SQL，存在 SQL 注入风险。
- 后台密钥校验为硬编码 `123456`，且通过上传文件登录，安全性弱。
- 使用 Flask session，无刷新令牌、无访问令牌过期治理、无登出黑名单。
- CORS 使用通配配置，缺少按环境区分。
- SQLite 不适合后续多连接、统计、全文搜索和部署扩展。
- 无 ORM、无 migration、无统一错误码、无参数校验层。
- 文章内容 DB 与文件双写，图片 DB Base64 与文件双写，容易出现不一致。
- 图片 Base64 入库会放大存储体积，也不利于 CDN、缓存和范围读取。
- 缺少结构化日志、链路追踪、测试与健康检查。

### 业务逻辑

- 文章查询缺少服务端分页、排序、搜索。
- 分类和标签使用字符串 ID 串联，不利于关系查询。
- 阅读量每次详情请求直接写库，容易在访问量变大时形成写入热点。
- 公告表已存在，但前台公告仍是组件内静态文本。
- 图床没有文件大小限制、MIME 校验、缩略图、引用关系统计。
- 后台路由和公开 API 混杂，权限边界不够清晰。

## 新系统重构方向

新版不建议在旧项目上直接修改，而是以旧项目为功能原型重新搭建：

- 前端新建 `web` 应用，使用 Vue3 + Vite + TypeScript + Sass。
- 后端新建 `server` 应用，使用 Go + Gin + GORM + JWT。
- 异步任务新建 `worker` 应用，使用 Python + Celery，Redis 做 broker 和 result backend。
- 数据从 SQLite 和 Markdown 文件迁移到 MySQL。
- 图片文件以本地对象存储目录起步，预留 S3/MinIO/CDN 抽象。
- 公开站点和后台管理可以共用一个前端项目，但路由、布局、状态和权限要分组。
