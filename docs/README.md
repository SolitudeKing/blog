# Solitude Blog · 设计与交付文档

本目录是新版个人博客的设计与交付基线，覆盖产品、架构、数据、前端体验、实施、迁移和部署。文档以旧博客能力为业务来源，以当前 `web`、`server`、`worker`、`deploy` 工程为实现对象。

## 当前设计基线

| 项目 | 约定 |
| --- | --- |
| 前端 | Vue 3 + Vite + TypeScript + Sass，自研基础组件 |
| 后端 | Go + Gin + GORM + JWT |
| 数据与任务 | MySQL + Redis + Celery |
| 默认视觉 | `mist-sea-salt + light`；支持 `mist-sea-salt/mist-forest × light/dark` 四种组合 |
| UI 规范 | Mist UI 雾境海盐与雾境青森色调 + 语义 token、雾面玻璃、完整状态与可访问性 |
| 布局策略 | CSS Grid 负责页面骨架，Flex 负责组件排列，移动端使用抽屉或单列 |
| 当前进度 | 以 [设计进度追踪](./05-design-progress.md) 为准 |

## 按目标阅读

| 我想了解 | 从这里开始 | 继续阅读 |
| --- | --- | --- |
| 为什么重建 | [当前项目评审](./00-current-project-review.md) | [新系统架构设计](./01-new-blog-architecture.md) |
| 产品有哪些页面与流程 | [产品与交互设计](./02-product-and-interaction-design.md) | [布局模式](./10-layout-patterns.md) |
| 数据库和接口如何设计 | [数据模型与 API 设计](./03-data-api-design.md) | [编码设计](./06-coding-design.md) |
| 主题色从哪里来、如何映射 | [雾境主题色调系统](./11-theme-color-system.md) | [UI 设计系统](./09-ui-design-system.md) |
| 组件和布局如何落地 | [UI 设计系统](./09-ui-design-system.md) | [布局模式](./10-layout-patterns.md) |
| 下一步做什么 | [设计进度追踪](./05-design-progress.md) | [实施路线图](./04-implementation-roadmap.md) |
| 如何迁移旧数据 | [旧博客迁移运行手册](./07-migration-runbook.md) | [数据模型与 API 设计](./03-data-api-design.md) |
| 如何部署、备份和恢复 | [部署与备份运行手册](./08-deployment-runbook.md) | [新系统架构设计](./01-new-blog-architecture.md) |

## 文档地图

### 背景与系统设计

1. [当前项目评审](./00-current-project-review.md)
   - 旧项目前后台、接口、数据模型、可复用资产与技术债。
2. [新系统架构设计](./01-new-blog-architecture.md)
   - 服务边界、前后端分层、鉴权、缓存、存储和部署架构。
3. [产品与交互设计](./02-product-and-interaction-design.md)
   - 公开站点与后台管理的信息架构、页面能力和关键交互。
4. [数据模型与 API 设计](./03-data-api-design.md)
   - MySQL 表、统一响应、错误码、鉴权、缓存键与 API 清单。

### 体验设计

5. [UI 设计系统](./09-ui-design-system.md)
   - 语义 token 契约、字体、圆角、阴影、动效、组件状态、主题控制器和可访问性。
6. [雾境主题色调系统](./11-theme-color-system.md)
   - 雾境海盐与雾境青森的 light/dark 精确映射、氛围 token、派生边界和对比度基线。
7. [布局模式](./10-layout-patterns.md)
   - 公共壳层、首页、文章、归档、搜索、后台、登录和响应式模式。

### 实施与跟踪

8. [实施路线图](./04-implementation-roadmap.md)
   - Phase、里程碑、验收标准、风险控制和 Definition of Done。
9. [设计进度追踪](./05-design-progress.md)
   - 当前状态、已确认决策、近期任务与变更日志。
10. [编码设计](./06-coding-design.md)
   - 工程目录、编码约定、前后端分层、Worker、配置与测试。

### 运行手册

11. [旧博客迁移运行手册](./07-migration-runbook.md)
    - 导出、导入、映射、校验和失败处理。
12. [部署与备份运行手册](./08-deployment-runbook.md)
    - 启动、健康检查、MySQL 备份恢复、日志与持久化。

## 推荐阅读路径

### 前端开发

`02 产品与交互 → 09 UI 设计系统 → 11 色调系统 → 10 布局模式 → 06 编码设计 → 03 API`

### 后端开发

`01 架构 → 03 数据与 API → 06 编码设计 → 04 实施路线图`

### 部署与迁移

`01 部署架构 → 07 迁移手册 → 08 部署手册 → 05 进度追踪`

## 设计原则与取舍

- 业务能力来自仓库中的旧博客项目，旧代码只作为行为与迁移依据。
- 主题采用 token 化组织，布局使用玻璃内容壳、可折叠导航和自适应卡片网格，形成项目自身的视觉语言。
- 不在组件中硬编码品牌配色，不使用 Float 主布局、固定大尺寸、单一粗断点和重阴影；具体规则由 [色调系统](./11-theme-color-system.md)、[UI 设计系统](./09-ui-design-system.md) 与 [布局模式](./10-layout-patterns.md) 统一定义。
- 色调与氛围采用 Mist UI 的雾境海盐、雾境青森规范，并按本项目运行时需求扩展为二维控制面：后台全局设置主题，访客在前台切换明暗模式。
- 主题与模式具有独立状态源：服务端 `theme` 只允许 `mist-sea-salt`、`mist-forest`；服务端 `mode` 是访客默认值，浏览器 `blog:mode` 合法本地偏好优先。

## 目标契约与当前实现

设计规范描述目标，源码描述当前事实；两者的差距必须显式记录，不能用任一方悄悄覆盖另一方。

| 内容 | 目标契约 | 当前实现位置 |
| --- | --- | --- |
| 设计状态 | [设计进度追踪](./05-design-progress.md) | 各文档的“实现偏差”或“当前进度”段落 |
| 主题色值与来源映射 | [雾境主题色调系统](./11-theme-color-system.md) | `web/src/styles/tokens/`、`web/src/styles/themes/` |
| UI token 与组件 | [UI 设计系统](./09-ui-design-system.md) | `web/src/styles/tokens/`、`web/src/styles/components/`、`web/src/components/base/` |
| 页面布局 | [布局模式](./10-layout-patterns.md) | `web/src/layouts/`、`web/src/pages/`、`web/src/styles/components/` |
| 主题与明暗模式 | [UI 设计系统](./09-ui-design-system.md#主题与明暗模式规范) | `web/src/composables/useTheme.ts`、`web/src/stores/setting.ts` |
| API 与数据 | [数据模型与 API 设计](./03-data-api-design.md) | `server/internal/handler/`、`server/internal/service/`、`server/internal/model/` |
| 工程组织 | [编码设计](./06-coding-design.md) | `web/`、`server/`、`worker/`、`deploy/` |
| 环境变量 | `.env.example` | 各服务配置加载代码与部署配置 |
| 部署与运维 | [部署与备份运行手册](./08-deployment-runbook.md) | `deploy/` 与 `scripts/` |

规范与实现冲突时，先判断是“设计尚未落地”还是“实现已改变”。若设计仍有效，把差异加入对应偏差清单；若实现代表新决策，先修订目标契约。完成调整后同步 [设计进度追踪](./05-design-progress.md)。

## 文档维护约定

- 新增能力时同时更新产品、接口、编码和验收文档中的相关部分。
- UI 组件只使用语义 token；新增或调整颜色时先更新 [色调系统](./11-theme-color-system.md)，再修改主题映射，不复制组件样式。
- 页面布局变化先更新 [布局模式](./10-layout-patterns.md)，再更新页面实现。
- 示例代码必须给出语言标识；流程和层级复杂时优先使用 Mermaid、表格或结构图。
- 相对链接必须可解析；命令、路径、环境变量和接口名必须可复制。
- 进度文档只记录可验证状态，不用计划状态替代实现结果。
