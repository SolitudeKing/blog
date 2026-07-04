# 新个人博客系统设计文档索引

本文档集基于当前旧项目 `blog-mini-serve` 与 `blog-mini-v3` 的功能结构整理而来，目标是为后续使用 Vue3 + Vite + TypeScript + Sass 与 Go + Gin + ORM + JWT + Cache + Celery + MySQL + Redis 重建个人博客系统提供可执行的设计依据。

## 阅读顺序

1. [当前项目评审](./00-current-project-review.md)
   - 复盘旧项目的功能结构、数据模型、可复用资产与主要技术债。
2. [新系统架构设计](./01-new-blog-architecture.md)
   - 定义前端、后端、缓存、异步任务、存储、部署与目录结构。
3. [产品与交互设计](./02-product-and-interaction-design.md)
   - 设计新版博客的公开站点、后台管理台、视觉系统和核心交互。
4. [数据模型与 API 设计](./03-data-api-design.md)
   - 给出 MySQL 表结构、接口规范、鉴权流程、缓存键和迁移方案。
5. [实施路线图](./04-implementation-roadmap.md)
   - 拆分开发阶段、验收标准、风险控制和推荐迭代节奏。
6. [设计进度追踪](./05-design-progress.md)
   - 记录当前设计状态、已确认决策、近期任务和设计日志。
7. [编码设计](./06-coding-design.md)
   - 定义新系统工程目录、编码规范、前后端分层、Worker 任务、配置和测试策略。

## 设计原则

- 保留旧项目中已经成型的博客能力：文章、分类、标签、归档、关于、公告、图床、站点配置、后台管理。
- 新系统前后台统一采用现代 SPA 体验，但后台能力与公开站点在路由、权限、接口层面清晰隔离。
- 前端不使用 UI 组件库，所有按钮、表单、弹窗、表格、导航、标签、分页等都沉淀为项目内自研基础组件。
- 自研基础组件的视觉风格参考 `C:\Users\XiaoMeng\.cc-switch\skills\creamy-ui\SKILL.md` 中的 CreamyUI 设计系统，优先使用其语义 token、主题模式、圆润层级、完整状态与可访问性规则。
- 后端从脚本式 Flask 视图升级为 Go 分层服务：Router、Handler、Service、Repository、Model、DTO、Middleware。
- MySQL 作为长期数据源，Redis 同时承担缓存、限流、JWT 黑名单、异步任务 broker/result backend。
- Celery 作为独立异步任务服务接入，负责内容索引、图片处理、站点地图、统计聚合、备份等非同步请求链路任务。
