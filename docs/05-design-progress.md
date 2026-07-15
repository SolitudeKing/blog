# 设计进度追踪

本文档用于记录新个人博客系统的设计推进状态。后续每次完成架构、交互、接口、数据、迁移、部署等设计变更时，都应同步更新本文件，避免设计决策散落在多个文档里。

## 当前状态

| 项目 | 内容 |
| --- | --- |
| 项目名称 | 新个人博客系统 |
| 当前阶段 | demo 视觉迁移收口与部署验证 |
| 当前版本 | `design-v0.12.0` |
| 开始日期 | 2026-07-03 |
| 最近更新 | 2026-07-15 |
| 设计目标 | 基于旧博客功能结构，设计 Vue3 + Vite + TS + Sass + Go + Gin + GORM + JWT + Redis + Celery + MySQL 的新版博客系统 |

## 状态说明

| 状态 | 含义 |
| --- | --- |
| `完成` | 已形成可作为开发依据的设计文档，后续只做小范围修订 |
| `进行中` | 已有方向或草案，还需要补充细节、样例或验收标准 |
| `复验中` | 主要代码已存在，但仍有明确缺陷或验收项未关闭 |
| `待开始` | 尚未展开设计 |
| `阻塞` | 需要外部信息或关键决策才能继续 |

## 设计基线进度

| 模块 | 状态 | 进度 | 对应文档 | 说明 |
| --- | --- | --- | --- | --- |
| 旧项目评审 | 完成 | 100% | [00-current-project-review.md](./00-current-project-review.md) | 已梳理旧项目前台、后台、接口、数据模型、技术债和可复用资产 |
| 总体架构设计 | 完成 | 90% | [01-new-blog-architecture.md](./01-new-blog-architecture.md) | 已确定前后端、Redis、Celery、存储、部署与目录结构；后续可补充更细的模块依赖图 |
| 产品与交互设计 | 进行中 | 95% | [02-product-and-interaction-design.md](./02-product-and-interaction-design.md)、[10-layout-patterns.md](./10-layout-patterns.md) | demo 的编辑式前台与创作工作台轮廓已迁入 Vue；后续按真实浏览器与使用反馈细化 |
| 数据模型与 API 设计 | 完成 | 85% | [03-data-api-design.md](./03-data-api-design.md) | 已确定表设计、无 `/api` 前缀、Header 版本、模块一级路径、统一响应、游标分页、错误码；待补充字段级请求/响应 DTO |
| 实施路线图 | 完成 | 90% | [04-implementation-roadmap.md](./04-implementation-roadmap.md) | 已拆分 Phase 0-8、里程碑、验收和风险；后续可根据真实开发进度滚动调整 |
| UI 主题与组件规范 | 复验中 | 100% | [09-ui-design-system.md](./09-ui-design-system.md)、[11-theme-color-system.md](./11-theme-color-system.md)、[10-layout-patterns.md](./10-layout-patterns.md) | 二维契约、源码迁移、四组 token 对齐和构建已完成；关键页面的四组合浏览器截图基线待补 |
| 数据迁移方案 | 完成 | 100% | [03-data-api-design.md](./03-data-api-design.md)、[07-migration-runbook.md](./07-migration-runbook.md) | 已完成旧 SQLite、Markdown、PicBed Base64、图片文件导出与 Go 导入链路，并输出导出/导入报告 |
| 部署与运维设计 | 完成 | 100% | [01-new-blog-architecture.md](./01-new-blog-architecture.md)、[04-implementation-roadmap.md](./04-implementation-roadmap.md)、[08-deployment-runbook.md](./08-deployment-runbook.md) | Compose、健康检查、持久化、日志、备份与恢复设计已成文；实机结果在下方实现验收表跟踪 |
| 编码设计 | 完成 | 100% | [06-coding-design.md](./06-coding-design.md) | 已定义工程目录、前端 API client、后端分层、Worker 任务、配置和测试策略 |

上表的百分比只表示设计文档完整度，不代表源码完成度。

## 实现与验收进度

| 范围 | 状态 | 已有实现 | 未关闭验收 |
| --- | --- | --- | --- |
| 整体编码实现 | 进行中 | `web`、`server`、`worker`、`deploy` 均已建立；前台/后台已按 demo 重构并通过构建 | Worker 多数任务仍为占位实现；前端缺测试体系；repository/DTO 分层、浏览器 UI 基线与部署实机验证未收口 |
| M2 可写可读文章 | 复验中 | 登录、文章创建/发布、编辑式列表/详情、Markdown、当前章节目录、分页和移动 TOC 已实现 | 相邻文章仍由额外列表推导；真实浏览器的阅读宽度、键盘和读屏抽查待完成 |
| M3 可日常管理 | 复验中 | 管理页、可折叠/抽屉 Shell、创作 Dashboard、全局文章搜索及二维外观设置已实现 | 其余表单 ARIA、窄屏表格和移动键盘仍需浏览器复验 |
| M4 可迁移上线 | 进行中 | 迁移导入、缓存、部署配置、健康检查与备份恢复脚本已完成代码和文档收口 | 部署机 Docker Compose 展开、容器健康检查和备份恢复演练尚未完成 |
| M5 体验增强 | 复验中 | 搜索工作台、About、RSS、sitemap、统计面板、版本记录、自动保存和二维主题状态已有实现 | Category/Tag 无独立公开页；Mist UI 四组合浏览器视觉回归尚未关闭 |

## 已确认设计决策

- 新项目基于旧博客功能重建，不直接在旧代码上改造。
- 前端使用 Vue3 + Vite + `<script setup>` + TypeScript + Sass。
- 前端不使用 UI 组件库，基础组件全部自研。
- 自研组件使用 Mist UI 的语义 token、雾面氛围和交互规范，具体色值以 [雾境主题色调系统](./11-theme-color-system.md) 为准。
- 支持 `mist-sea-salt`、`mist-forest` 两个主题和 `light`、`dark` 两种模式，默认组合为 `mist-sea-salt + light`。
- 主题风格与明暗模式是两个独立维度：后台全局控制 theme 与访客默认 mode，公开前台只提供 mode 控件，组件不得感知具体主题名称。
- 浏览器合法 `blog:mode` 优先于服务端默认 mode；主题变化不得清除或覆盖该偏好。
- 前台布局采用编辑式 7/5 舞台、阅读流和连续时间线；后台采用可折叠 Shell 与数据工作台。统一使用 Grid/Flex、流式尺寸、移动抽屉和语义 token，不使用 Float、固定大画布或默认卡片墙。
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
- 第一轮编码从 health、auth、user、setting、article 主链路展开，M2–M5 的主要页面与服务已有实现；“已有代码”不等于“通过验收”，当前以部署实机验证、Worker 真实任务、测试体系、[UI 偏差](./09-ui-design-system.md#实现偏差清单) 与 [布局偏差](./10-layout-patterns.md#16-当前实现偏差清单) 收口为主。

## 近期设计任务

| 优先级 | 任务 | 状态 | 产出文档 |
| --- | --- | --- | --- |
| P0 | 重构 Mist UI 双主题 token、氛围入口与四组合验收方案 | 复验中 | 两主题四组合均包含 86 个一致 token，前端构建通过；待补关键页面浏览器截图基线 |
| P0 | 将 `mist-violet` 迁移为 `mist-sea-salt/mist-forest × light/dark` | 完成 | 旧运行时文件已退出，历史数据显式映射 `forest` → `mist-forest`，其余旧值回退海盐 |
| P0 | 细化首批自研组件 API 与状态矩阵 | 完成 | `09-ui-design-system.md` |
| P0 | 收敛后台全局主题、访客默认模式与 `blog:mode` 本地偏好的优先级 | 完成 | 服务端 theme 始终同步，本地 mode 优先；首屏缓存、跨标签同步和旧复合键迁移已落地 |
| P0 | 修复文章 slug 重载、搜索请求竞态与三类移动抽屉可访问性 | 复验中 | 源码状态机与构建已完成；待浏览器键盘/读屏复验 |
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
| P1 | 完善 Markdown 渲染、目录导航与文章阅读体验 | 复验中 | `02-product-and-interaction-design.md`、`10-layout-patterns.md` |
| P1 | 完成搜索增强 | 复验中 | `04-implementation-roadmap.md`、`10-layout-patterns.md` |
| P1 | 完成 RSS 与 sitemap | 完成 | `04-implementation-roadmap.md` |
| P1 | 完成增强统计面板 | 完成 | `04-implementation-roadmap.md` |
| P1 | 完成文章版本记录 | 完成 | `03-data-api-design.md`、`04-implementation-roadmap.md` |
| P1 | 完成编辑器自动保存 | 完成 | `02-product-and-interaction-design.md`、`04-implementation-roadmap.md` |
| P1 | 落地公开站点核心页面：首页、文章详情、归档、搜索、About | 复验中 | demo 构图已迁入 Vue 并通过构建；Category/Tag 仍使用搜索入口，视觉基线待补 |
| P1 | 绘制后台壳层与登录布局，定义仪表盘、列表、编辑器、媒体和设置的通用模式 | 完成 | `10-layout-patterns.md` |
| P1 | 将 Worker 占位任务替换为可重试、可观测的真实实现 | 待开始 | `06-coding-design.md` |
| P1 | 建立前端 API client、主题、路由与基础组件测试 | 待开始 | `06-coding-design.md` |
| P1 | 补充旧数据迁移校验规则和报告格式 | 完成 | `07-migration-runbook.md` |
| P2 | 补充 Docker Compose、Nginx、环境变量、备份与日志设计 | 完成 | `08-deployment-runbook.md` |

## 设计日志

日志保留当时的判断；后续复验发现回归或欠账时，不改写历史记录，以本文件顶部“实现与验收进度”为当前结论。

| 日期 | 版本 | 内容 |
| --- | --- | --- |
| 2026-07-15 | `design-v0.12.0` | 按 `demo/` 落地页语言重构 Vue 前台与后台：公共壳层改为全宽薄边导航和编辑式页脚；首页采用作者 7/5 舞台、一主两辅文章目录和连续文章流；文章、归档、搜索、About 迁入对应阅读/时间线/工作台轮廓；后台改为 quiet Shell 与非卡片墙 Dashboard。保留海盐/青森 × light/dark 二维主题，补齐 demo 语义 token；前端构建通过，真实浏览器截图与读屏抽查待补 |
| 2026-07-15 | `design-v0.11.1` | 完成 Mist UI 二维主题源码迁移：新增雾境海盐/青森四组完整 token、五层雾境氛围与后台主题卡预览；拆分服务端 theme、默认 mode、`blog:mode` 和首屏缓存权威；服务端统一旧值归一化与严格写入校验。前端构建、Go 测试、Mist UI 技能校验和四组 token 一致性检查通过，浏览器截图基线待补 |
| 2026-07-15 | `design-v0.11.0` | 启动 Mist UI 二维主题重构：目标矩阵改为雾境海盐/雾境青森 × light/dark；后台全局管理 theme 与访客默认 mode，前台只切换并本地保存 mode。当前先整理设计、架构、API、编码与验收契约，源码重构和四组合回归尚在进行中 |
| 2026-07-15 | `design-v0.10.2` | 按雾境紫与 CreamyUI 表面层级重构前端视觉：完整启用八层环境光斑与双 Blob，首页收敛为三栏玻璃主壳并简化文章卡层级；同步优化悬浮导航、阅读排版、归档时间线、搜索结果、后台工作台与分栏登录页，继续只保留 light/dark 模式 |
| 2026-07-15 | `design-v0.10.1` | 按雾境紫色调系统完成前后端主题重构：运行时只加载 `mist-violet`，界面仅保留 light/dark 切换；同步首屏防闪烁、本地偏好迁移、站点默认值、数据库旧值归一化与高对比控件边界 |
| 2026-07-15 | `design-v0.10.0` | 从外部文档色板提取并审计规范颜色，按 CreamyUI 二维主题与语义 token 格式整理为自主命名的 `mist-violet × light/dark` 色调系统；记录来源哈希、冲突裁决、完整 CSS 草案、派生规则、对比度和后续迁移清单，不采用 CreamyUI 内置主题配色 |
| 2026-07-13 | `design-v0.9.0` | 结合 CreamyUI 重构自主设计的文档体系；新增 UI 设计系统与布局模式，明确双维主题、语义 token、组件状态、公共/后台壳层、响应式断点和实现偏差，并重构文档索引与交叉引用；基于源码复核重新打开 M2、M3、M5 体验复验和 M4 部署验收 |
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
