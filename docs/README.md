# Solitude Blog · 文档索引

本目录记录当前博客系统的产品、架构、数据、界面、部署与维护约定。源码描述当前事实，文档描述已经确认的契约和仍待完成的工作；计划能力不得写成已实现能力。

## 当前基线

| 领域 | 当前约定 |
| --- | --- |
| 前端 | Vue 3 + Vite + TypeScript + Sass，自研基础组件 |
| 后端 | Go + Gin + GORM + JWT |
| 数据 | MySQL 是业务数据唯一可信来源，Redis 只承载可重建的读取缓存 |
| 部署 | Docker Compose + Nginx；当前仓库不包含 TLS 终止 |
| 主题 | `mist-sea-salt`、`mist-forest` × `light`、`dark` |
| 专题 | 雾里拾笺 `NODES/nodes`、微光造物 `CODE/code`、风过留痕 `JOTTING/jotting` |

## 推荐阅读

| 目标 | 首选文档 | 补充文档 |
| --- | --- | --- |
| 了解系统边界 | [架构设计](./01-new-blog-architecture.md) | [编码设计](./06-coding-design.md) |
| 理解页面与交互 | [产品与交互设计](./02-product-and-interaction-design.md) | [布局模式](./10-layout-patterns.md) |
| 维护数据与 API | [数据模型与 API 设计](./03-data-api-design.md) | [维护指南](./12-maintenance-guide.md) |
| 维护主题与组件 | [UI 设计系统](./09-ui-design-system.md) | [主题色调系统](./11-theme-color-system.md) |
| 查看当前任务 | [实施路线图](./04-implementation-roadmap.md) | [设计进度](./05-design-progress.md) |
| 部署、备份与恢复 | [部署与备份手册](./08-deployment-runbook.md) | [维护指南](./12-maintenance-guide.md) |
| 查看本轮审查结果 | [项目审查](./13-project-review.md) | [实施路线图](./04-implementation-roadmap.md) |

## 文档地图

1. [架构设计](./01-new-blog-architecture.md)：当前服务边界、依赖关系与部署形态。
2. [产品与交互设计](./02-product-and-interaction-design.md)：公开站点和管理后台能力。
3. [数据模型与 API 设计](./03-data-api-design.md)：数据不变量、接口与缓存约定。
4. [实施路线图](./04-implementation-roadmap.md)：尚未完成的 P0、P1、P2 工作。
5. [设计进度](./05-design-progress.md)：当前可验证状态和最近变更。
6. [编码设计](./06-coding-design.md)：目录、分层、命名、配置与测试规则。
7. [部署与备份手册](./08-deployment-runbook.md)：启动、健康检查、备份和恢复。
8. [UI 设计系统](./09-ui-design-system.md)：语义 token、组件状态和可访问性。
9. [布局模式](./10-layout-patterns.md)：公共页面、后台与响应式布局。
10. [主题色调系统](./11-theme-color-system.md)：主题与明暗模式的颜色契约。
11. [维护指南](./12-maintenance-guide.md)：日常维护入口和关键不变量。
12. [项目审查](./13-project-review.md)：本轮已修复项与后续风险。

## 文档维护规则

- README 和运行手册只陈述当前可运行能力；目标方案放入路线图并标记优先级。
- 路由、环境变量、专题标识或目录职责变化时，同步更新对应源码旁注释和本索引。
- UI 组件只消费语义 token；新增颜色先更新主题契约，再更新实现。
- 示例命令必须可复制，并明确是在宿主机还是容器内执行。
- 删除文件后使用 `rg` 检查路径、链接和术语；`demo/` 是受保护的设计蓝图，不得按旧工程或迁移脚手架清理。
- 注释解释约束、边界和原因，不逐行翻译代码。
