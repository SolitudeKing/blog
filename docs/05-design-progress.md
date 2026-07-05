# 设计进度追踪

本文档用于记录新个人博客系统的设计推进状态。后续每次完成架构、交互、接口、数据、迁移、部署等设计变更时，都应同步更新本文件，避免设计决策散落在多个文档里。

## 当前状态

| 项目 | 内容 |
| --- | --- |
| 项目名称 | 新个人博客系统 |
| 设计阶段 | 方案设计中 |
| 当前版本 | `design-v0.8.0` |
| 开始日期 | 2026-07-03 |
| 最近更新 | 2026-07-05 |
| 设计目标 | 基于旧博客功能结构，设计 Vue3 + Vite + TS + Sass + Go + Gin + GORM + JWT + Redis + Celery + MySQL 的新版博客系统 |

## 状态说明

| 状态 | 含义 |
| --- | --- |
| `完成` | 已形成可作为开发依据的设计文档，后续只做小范围修订 |
| `进行中` | 已有方向或草案，还需要补充细节、样例或验收标准 |
| `待开始` | 尚未展开设计 |
| `阻塞` | 需要外部信息或关键决策才能继续 |

## 总体进度

| 模块 | 状态 | 进度 | 对应文档 | 说明 |
| --- | --- | --- | --- | --- |
| 旧项目评审 | 完成 | 100% | [00-current-project-review.md](./00-current-project-review.md) | 已梳理旧项目前台、后台、接口、数据模型、技术债和可复用资产 |
| 总体架构设计 | 完成 | 90% | [01-new-blog-architecture.md](./01-new-blog-architecture.md) | 已确定前后端、Redis、Celery、存储、部署与目录结构；后续可补充更细的模块依赖图 |
| 产品与交互设计 | 进行中 | 75% | [02-product-and-interaction-design.md](./02-product-and-interaction-design.md) | 已覆盖公开站点、后台管理、自研 UI、CreamyUI 风格；待补充页面线框与关键流程图 |
| 数据模型与 API 设计 | 完成 | 85% | [03-data-api-design.md](./03-data-api-design.md) | 已确定表设计、无 `/api` 前缀、Header 版本、模块一级路径、统一响应、游标分页、错误码；待补充字段级请求/响应 DTO |
| 实施路线图 | 完成 | 90% | [04-implementation-roadmap.md](./04-implementation-roadmap.md) | 已拆分 Phase 0-8、里程碑、验收和风险；后续可根据真实开发进度滚动调整 |
| UI 主题与组件规范 | 进行中 | 55% | [02-product-and-interaction-design.md](./02-product-and-interaction-design.md) | 已确定参考 CreamyUI；待输出本项目 token 文件结构、组件状态矩阵和首批组件 API |
| 数据迁移方案 | 完成 | 100% | [03-data-api-design.md](./03-data-api-design.md)、[07-migration-runbook.md](./07-migration-runbook.md) | 已完成旧 SQLite、Markdown、PicBed Base64、图片文件导出与 Go 导入链路，并输出导出/导入报告 |
| 部署与运维设计 | 完成 | 90% | [01-new-blog-architecture.md](./01-new-blog-architecture.md)、[04-implementation-roadmap.md](./04-implementation-roadmap.md)、[08-deployment-runbook.md](./08-deployment-runbook.md) | 已补充 Compose 健康检查、Redis AOF、上传卷、日志限制、MySQL 备份/恢复脚本和部署运行手册；本机缺少 Docker CLI，Compose 实机校验待部署机执行 |
| 编码设计 | 完成 | 100% | [06-coding-design.md](./06-coding-design.md) | 已定义工程目录、前端 API client、后端分层、Worker 任务、配置和测试策略 |
| 编码实现骨架 | 进行中 | 99% | [../README.md](../README.md) | 已创建 `web`、`server`、`worker`、`deploy` 骨架；Go API 已接入 `.env` 加载、MySQL/GORM 自动迁移、Redis 健康检查、管理员初始化、JWT 登录鉴权、文章数据库优先 CRUD、站点配置持久化、公告管理、后台摘要统计、媒体资源管理、前台内容缓存、RSS、sitemap、文章搜索、版本记录；后台已接入文章管理、分类标签管理、文章分类标签选择、站点设置页、公告管理页、增强仪表盘、媒体库和编辑器自动保存；M4 迁移导入、部署健康检查、备份恢复已完成代码收口；Compose 尚未实机校验 |
| M2 可写可读文章 | 完成 | 100% | [04-implementation-roadmap.md](./04-implementation-roadmap.md) | 登录、创建文章、发布文章、前台文章列表、文章详情、Markdown 渲染、目录导航、列表分页、空状态和错误状态已完成；已通过前端类型检查和构建 |
| M3 可日常管理 | 完成 | 100% | [04-implementation-roadmap.md](./04-implementation-roadmap.md) | 分类标签 CRUD 和文章编辑器选择器作为 M3 前置工作已提前完成；站点配置、公告管理、后台仪表盘和媒体资源管理已完成数据库持久化、后台维护页、摘要聚合、上传管理和前台读取链路 |
| M4 可迁移上线 | 完成 | 95% | [04-implementation-roadmap.md](./04-implementation-roadmap.md)、[07-migration-runbook.md](./07-migration-runbook.md)、[08-deployment-runbook.md](./08-deployment-runbook.md) | 旧数据迁移、媒体库、Redis 缓存、部署配置、备份恢复脚本和健康检查已完成；剩余 5% 为部署机 Docker Compose 实机验证 |
| M5 体验增强 | 完成 | 100% | [04-implementation-roadmap.md](./04-implementation-roadmap.md) | 搜索增强、RSS、sitemap、统计面板、版本记录和自动保存均已完成 |

## 已确认设计决策

- 新项目基于旧博客功能重建，不直接在旧代码上改造。
- 前端使用 Vue3 + Vite + `<script setup>` + TypeScript + Sass。
- 前端不使用 UI 组件库，基础组件全部自研。
- 自研组件 UI 风格参考 `C:\Users\XiaoMeng\.cc-switch\skills\creamy-ui\SKILL.md` 中的 CreamyUI。
- CreamyUI 默认建议使用 `forest + light` 作为第一阶段主视觉，保留 `strawberry` 和 `dark` 主题扩展能力。
- 后端使用 Go + Gin + GORM + JWT。
- 数据库使用 MySQL，缓存和任务基础设施使用 Redis。
- Celery 作为独立异步任务服务，负责图片处理、索引、统计聚合、备份、迁移等任务。
- API 不使用全局 `/api` 前缀。
- API 不在 URL 中放版本号，版本号放在 `X-API-Version` 请求头。
- API 按业务模块使用一级路径前缀，例如 `auth/login`、`user/info`、`setting/lobby`。
- 所有列表接口使用游标分页，响应中列表数据放在 `data` 数组，分页信息放在同级 `page` 字段。
- 新系统采用 `web`、`server`、`worker`、`deploy` 四个新工程目录，旧项目仅作为参考和迁移来源。
- Go 后端依赖方向固定为 `router -> handler -> service -> repository -> model`，service 负责事务、缓存失效和任务投递。
- 前端 API client 统一注入 `X-API-Version` 和 JWT，并解析统一响应与游标分页响应。
- 第一轮编码实现从可运行骨架开始，先完成 health、auth、user、setting、article 示例链路；当前 M2、M3、M4、M5 已完成代码收口。分类标签管理已作为 M3 前置工作提前完成，站点配置、公告管理、后台仪表盘和媒体资源管理已完成数据库持久化、后台界面、摘要聚合、上传管理和前台读取链路；旧数据迁移、Redis 缓存、部署健康检查和备份恢复已完成；搜索增强、RSS、sitemap、统计面板、版本记录和自动保存已完成，下一步转向部署机实机验证与体验细节打磨。

## 近期设计任务

| 优先级 | 任务 | 状态 | 产出文档 |
| --- | --- | --- | --- |
| P0 | 补充 UI token 文件结构与主题切换实现方案 | 待开始 | `02-product-and-interaction-design.md`、`01-new-blog-architecture.md` |
| P0 | 细化首批自研组件 API：Button、Input、Textarea、Select、Modal、Toast、Table、CursorPagination、Upload | 待开始 | `02-product-and-interaction-design.md` |
| P0 | 为核心 API 增加请求/响应 DTO 示例 | 待开始 | `03-data-api-design.md` |
| P0 | 按编码设计创建 `web`、`server`、`worker`、`deploy` 工程脚手架 | 完成 | `06-coding-design.md` |
| P0 | 接入 Go 数据库层、GORM 模型和 migration | 完成 | `06-coding-design.md`、`03-data-api-design.md` |
| P0 | 将 auth 示例 token 替换为真实 JWT 签发与刷新 | 完成 | `06-coding-design.md`、`03-data-api-design.md` |
| P0 | 完善后台文章列表、创建、编辑与删除页面 | 完成 | `06-coding-design.md`、`02-product-and-interaction-design.md` |
| P0 | 完善分类标签 CRUD 与后台选择器 | 完成 | `06-coding-design.md`、`03-data-api-design.md`；M3 前置工作 |
| P0 | 完成站点配置持久化与后台设置页 | 完成 | `03-data-api-design.md`、`02-product-and-interaction-design.md`；M3 主要任务 |
| P0 | 完成公告管理与前台当前公告展示 | 完成 | `03-data-api-design.md`、`02-product-and-interaction-design.md`；M3 主要任务 |
| P0 | 完成后台仪表盘摘要统计 | 完成 | `03-data-api-design.md`、`02-product-and-interaction-design.md`；M3 主要任务 |
| P0 | 完成媒体资源上传与后台媒体库 | 完成 | `03-data-api-design.md`、`02-product-and-interaction-design.md`；M3 主要任务 |
| P1 | 完善 Markdown 渲染、目录导航与文章阅读体验 | 完成 | `02-product-and-interaction-design.md` |
| P1 | 完成搜索增强 | 完成 | `04-implementation-roadmap.md` |
| P1 | 完成 RSS 与 sitemap | 完成 | `04-implementation-roadmap.md` |
| P1 | 完成增强统计面板 | 完成 | `04-implementation-roadmap.md` |
| P1 | 完成文章版本记录 | 完成 | `03-data-api-design.md`、`04-implementation-roadmap.md` |
| P1 | 完成编辑器自动保存 | 完成 | `02-product-and-interaction-design.md`、`04-implementation-roadmap.md` |
| P1 | 绘制公开站点页面线框：首页、文章详情、归档、搜索、关于 | 待开始 | `02-product-and-interaction-design.md` |
| P1 | 绘制后台页面线框：登录、仪表盘、文章列表、编辑器、媒体库、配置 | 待开始 | `02-product-and-interaction-design.md` |
| P1 | 补充旧数据迁移校验规则和报告格式 | 完成 | `07-migration-runbook.md` |
| P2 | 补充 Docker Compose、Nginx、环境变量、备份与日志设计 | 完成 | `08-deployment-runbook.md` |

## 设计日志

| 日期 | 版本 | 内容 |
| --- | --- | --- |
| 2026-07-05 | `design-v0.8.0` | 完成 M5 体验增强：新增公开文章搜索页和 `search/article` 接口，后台仪表盘补充热门文章与分类分布，文章保存写入版本快照并在编辑器展示历史版本，编辑器支持本地自动保存与恢复草稿；M5 主要任务全部完成 |
| 2026-07-05 | `design-v0.7.1` | 启动 M5 体验增强：新增公开 `/rss.xml` 与 `/sitemap.xml`，RSS 输出最近已发布文章，sitemap 输出首页、归档页和已发布文章 URL；支持 `SITE_BASE_URL` 配置与反向代理 Host 推断 |
| 2026-07-05 | `design-v0.7.0` | 完成 M4 可迁移上线：新增旧博客导出器和 Go 导入器，支持旧 SQLite、Markdown、PicBed Base64 与图片文件迁移并生成报告；前台文章列表、文章详情和站点配置接入 Redis 缓存；Compose 补充健康检查、Redis AOF、上传卷、日志限制，新增 MySQL 备份/恢复与部署运行手册；本机缺少 Docker CLI，Compose 实机校验待部署机执行 |
| 2026-07-05 | `design-v0.6.4` | 完成 M3 媒体资源管理：新增 `assets` 数据模型、本地上传存储、媒体列表/筛选/编辑/删除接口和后台媒体库页面；`/uploads` 支持静态访问；M3 主要任务全部完成，下一步进入 M4 |
| 2026-07-05 | `design-v0.6.3` | 完成 M3 后台仪表盘：新增 `dashboard/summary` 摘要接口，聚合文章状态、阅读量、分类标签、公告数量、最近文章和当前公告；后台首页替换骨架页并提供快捷维护入口；M3 下一步转向媒体资源管理 |
| 2026-07-04 | `design-v0.6.2` | 完成 M3 公告管理：新增 `notices` 数据模型、公开当前公告接口、后台公告列表/创建/编辑/删除入口，并在首页展示当前启用公告；M3 下一步转向后台仪表盘和媒体资源管理 |
| 2026-07-04 | `design-v0.6.1` | 启动 M3 日常管理阶段：站点配置从硬编码改为 `site_settings` 持久化；后台新增站点设置页，支持站点名、作者、签名、主题模式和社交链接维护；前台 `setting/lobby` 继续作为公开读取入口 |
| 2026-07-04 | `design-v0.6.0` | 完成 M2 收口：前台文章列表支持游标分页、空状态和错误重试；文章详情支持 Markdown 渲染、目录导航和基础阅读样式；完成前端类型检查和前端构建 |
| 2026-07-04 | `design-v0.5.3` | 明确区分里程碑主要任务和前置工作：M2 当前仍处于收口阶段，分类标签管理记录为 M3 前置工作提前完成，M3 不在 M2 正式验收前标记为进行中 |
| 2026-07-04 | `design-v0.5.2` | 新增分类与标签后端 CRUD、后台分类标签管理页、文章编辑器分类下拉与标签勾选；修复软删除文章后标签引用误判；完成 Go 测试、前端类型检查、前端构建和分类标签关联 API 冒烟 |
| 2026-07-04 | `design-v0.5.1` | 后台文章管理接入真实 API，完成文章列表、搜索筛选、新建、编辑、删除和保存发布入口；新增文章编辑页和基础 Textarea 组件；完成 Go 测试、前端类型检查、前端构建和本地 API 冒烟 |
| 2026-07-04 | `design-v0.5` | Go API 接入 `.env` 自动加载、MySQL/GORM 自动迁移、Redis 健康检查、管理员初始化、JWT 登录与鉴权；文章 service 改为数据库优先的列表、详情、创建、更新和删除，并保留无数据库时的内存降级；完成 Go 测试 |
| 2026-07-03 | `design-v0.4` | 根据文档创建新系统编码骨架：Go API、Vue Web、Celery Worker、Docker Compose、Nginx、环境变量和根 README；完成 Go 测试、前端类型检查、前端构建和 Worker Python 编译；Docker CLI 缺失，Compose 配置待后续校验 |
| 2026-07-03 | `design-v0.3` | 新增编码设计文档，明确工程目录、前端编码规则、后端分层、Worker 任务、配置、测试和脚手架顺序 |
| 2026-07-03 | `design-v0.2` | 新增设计进度追踪文档，记录当前设计模块状态、已确认决策和近期任务 |
| 2026-07-03 | `design-v0.1` | 完成旧项目评审、总体架构、产品交互、数据/API、实施路线图初稿 |
| 2026-07-03 | `design-v0.1.1` | API 规范调整为无 `/api` 前缀、Header 版本、业务模块一级前缀、统一响应、游标分页和错误码 |
| 2026-07-03 | `design-v0.1.2` | 自研组件 UI 风格同步为参考 CreamyUI，补充主题、token、组件状态与可访问性要求 |

## 下一次更新规则

- 每次修改设计文档后，更新“总体进度”中的状态和进度。
- 每次确认关键技术或产品决策后，追加到“已确认设计决策”。
- 每次开始一项设计任务，将状态从 `待开始` 改为 `进行中`。
- 每次完成一项设计任务，追加一条“设计日志”。
- 如果某项设计需要用户确认，标记为 `阻塞` 并写明阻塞原因。
