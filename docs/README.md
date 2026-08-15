# Solitude Blog · 文档索引

本目录记录当前博客系统的产品、架构、数据、界面、部署与维护约定。源码描述当前事实，文档描述已经确认的契约和仍待完成的工作；计划能力不得写成已实现能力。

文档按四类组织：

- [01-decisions](./01-decisions/)：决策与设计（HLD / LLD / ADR）
- [02-requirements](./02-requirements/)：需求与计划（PRD / 路线图 / Backlog）
- [03-knowledge](./03-knowledge/)：知识沉淀（Wiki / API 规范）
- [04-operations](./04-operations/)：运维与事故（Runbook / 审查复盘 / 事故 / RCA）

`backend-optimization/` 子目录为后端 P1/P2 优化 Backlog，本轮不参与归类调整。

## 当前基线

| 领域 | 当前约定 |
| --- | --- |
| 前端 | Vue 3 + Vite + TypeScript + Sass，自研基础组件 |
| 后端 | Go + Gin + GORM + JWT |
| 数据 | MySQL 是业务数据唯一可信来源，Redis 只承载可重建的读取缓存 |
| 部署 | Docker Compose + Nginx，仅打包 `api` 与 `nginx` 两个服务；MySQL / Redis 由外部托管服务提供，备份由托管方负责；当前仓库不包含 TLS 终止 |
| 主题 | `mist-sea-salt`、`mist-forest` × `light`、`dark` |
| 专题 | 雾里拾笺 `NODES/nodes`、微光造物 `CODE/code`、风过留痕 `JOTTING/jotting` |

## 推荐阅读

| 目标 | 首选文档 | 补充文档 |
| --- | --- | --- |
| 了解系统边界 | [概要设计 · 架构](./01-decisions/hld/01-architecture.md) | [详细设计 · 编码设计](./01-decisions/lld/02-coding-design.md) |
| 理解页面与交互 | [PRD · 产品与交互设计](./02-requirements/prd/01-product-and-interaction-design.md) | [详细设计 · 布局模式](./01-decisions/lld/04-layout-patterns.md) |
| 维护数据与 API | [详细设计 · 数据与 API 设计](./01-decisions/lld/01-data-and-api-design.md) | [知识沉淀 · 维护指南](./03-knowledge/wiki/02-maintenance-guide.md) |
| 维护主题与组件 | [详细设计 · UI 设计系统](./01-decisions/lld/03-ui-design-system.md) | [详细设计 · 主题色调系统](./01-decisions/lld/05-theme-color-system.md) |
| 查看当前任务 | [路线图 · 实施路线图](./02-requirements/roadmap/01-implementation-roadmap.md) | [知识沉淀 · 设计进度](./03-knowledge/wiki/01-design-progress.md) |
| 部署、备份与恢复 | [Runbook · 部署与备份](./04-operations/runbook/01-deployment-and-backup.md) | [知识沉淀 · 维护指南](./03-knowledge/wiki/02-maintenance-guide.md) |
| 查看本轮审查结果 | [复盘 · 项目审查](./04-operations/reviews/01-project-review.md) | [Backlog · 后端 API 优化](./backend-optimization/README.md) |
| 排期后端 P1/P2 优化 | [Backlog · 后端 API 优化](./backend-optimization/README.md) | [复盘 · 项目审查](./04-operations/reviews/01-project-review.md) |

## 文档地图

### 01 决策与设计

- [概要设计 · 架构](./01-decisions/hld/01-architecture.md)：当前服务边界、依赖关系与部署形态。
- [详细设计 · 数据与 API 设计](./01-decisions/lld/01-data-and-api-design.md)：数据不变量、接口与缓存约定。
- [详细设计 · 编码设计](./01-decisions/lld/02-coding-design.md)：目录、分层、命名、配置与测试规则。
- [详细设计 · UI 设计系统](./01-decisions/lld/03-ui-design-system.md)：语义 token、组件状态和可访问性。
- [详细设计 · 布局模式](./01-decisions/lld/04-layout-patterns.md)：公共页面、后台管理与响应式布局。
- [详细设计 · 主题色调系统](./01-decisions/lld/05-theme-color-system.md)：主题与明暗模式的颜色契约。
- [ADR 目录](./01-decisions/adr/)：架构决策记录（当前为空，新增 ADR 后落地于此）。

### 02 需求与计划

- [PRD · 产品与交互设计](./02-requirements/prd/01-product-and-interaction-design.md)：公开站点和管理后台能力。
- [路线图 · 实施路线图](./02-requirements/roadmap/01-implementation-roadmap.md)：尚未完成的 P0、P1、P2 工作。
- [Backlog · 后端 API 优化](./backend-optimization/README.md)：强烈建议（P1）与可改进（P2）的具体方案与排期建议（本轮不归类）。

### 03 知识沉淀

- [Wiki · 设计进度](./03-knowledge/wiki/01-design-progress.md)：当前可验证状态和最近变更。
- [Wiki · 维护指南](./03-knowledge/wiki/02-maintenance-guide.md)：日常维护入口和关键不变量。
- [API 规范目录](./03-knowledge/api/)：预留对外 API 文档落点；当前 API 契约以 [数据与 API 设计](./01-decisions/lld/01-data-and-api-design.md) 为唯一事实来源。

### 04 运维与事故

- [Runbook · 部署与备份](./04-operations/runbook/01-deployment-and-backup.md)：启动、健康检查、备份和恢复。
- [复盘 · 项目审查](./04-operations/reviews/01-project-review.md)：本轮已修复项与后续风险。
- [事故报告目录](./04-operations/incidents/)：预留事故记录落点。
- [RCA 目录](./04-operations/rca/)：预留根因分析报告落点。

## 文档维护规则

- README 和运行手册只陈述当前可运行能力；目标方案放入路线图并标记优先级。
- 路由、环境变量、专题标识或目录职责变化时，同步更新对应源码旁注释和本索引。
- UI 组件只消费语义 token；新增颜色先更新主题契约，再更新实现。
- 示例命令必须可复制，并明确是在宿主机还是容器内执行。
- 删除文件后使用 `rg` 检查路径、链接和术语；`demo/` 是受保护的设计蓝图，不得按旧工程或迁移脚手架清理。
- 跨分类引用使用相对路径；新增 ADR / API 规范 / 事故报告 / RCA 时分别落到 `01-decisions/adr/`、`03-knowledge/api/`、`04-operations/incidents/`、`04-operations/rca/`。
- 注释解释约束、边界和原因，不逐行翻译代码。
