# 需求与计划

本目录描述"目标是什么"和"接下来做什么"。与 [01-decisions](../01-decisions/) 的"如何实现"分工，避免文档错位。

- `prd/` 产品需求文档：目标用户、信息架构、关键页面与能力，不描述实现细节。
- `roadmap/` 路线图与 Backlog：尚未完成的工作，按 P0 / P1 / P2 排期，含"完成定义"。

## 当前内容

- [PRD · 产品与交互设计](./prd/01-product-and-interaction-design.md)：公开站点与后台管理台的产品定位、页面结构与设计决策背景。
- [路线图 · 实施路线图](./roadmap/01-implementation-roadmap.md)：本仓库 P0 / P1 / P2 任务与完成定义。
- 后端 API 优化 Backlog：[backend-optimization/](../backend-optimization/)（本轮未参与归类调整，独立维护排期）。

## 写作约束

- PRD 描述目标体验，不写真实运行的接口或表结构；实现细节回查 [01-decisions/lld/](../01-decisions/lld/)。
- 路线图只记录"基线之后仍未完成"的工作；已落地的能力不再保留为待办。
- 新增 Backlog 时使用统一模板：问题位置、现状、建议方案、验收标准。
- 路线图条目从"待办"变为"已修复"时，必须同步在 [04-operations/reviews/](../04-operations/reviews/01-project-review.md) 留下复盘记录。
