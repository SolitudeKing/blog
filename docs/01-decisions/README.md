# 决策与设计

本目录收录所有"已经做出、约束后续实现"的设计契约，分为三层：

- `hld/` 概要设计（HLD）：跨模块、跨端的架构边界、运行形态与部署模型。
- `lld/` 详细设计（LLD）：可被前端 / 后端直接消费的字段、token、组件、布局、配置规则。
- `adr/` 架构决策记录（ADR）：单点决策的"问题—选项—决定—后果"四段式，新增时落地于此。

## 阅读顺序

1. 先读 [概要设计 · 架构](./hld/01-architecture.md) 建立系统边界。
2. 按关注点进入 LLD：
   - 数据与 API：[01-data-and-api-design.md](./lld/01-data-and-api-design.md)
   - 编码组织：[02-coding-design.md](./lld/02-coding-design.md)
   - UI 与组件：[03-ui-design-system.md](./lld/03-ui-design-system.md)
   - 布局与响应式：[04-layout-patterns.md](./lld/04-layout-patterns.md)
   - 主题与色调：[05-theme-color-system.md](./lld/05-theme-color-system.md)
3. 对单点决策（例如三专题契约、二维主题矩阵）可在 `adr/` 检索或新增 ADR。

## 与其他目录的协作

- 产品侧的"目标体验"在 [02-requirements/prd/](../02-requirements/prd/01-product-and-interaction-design.md)；当 PRD 与 LLD 冲突时，以 LLD 为实现契约。
- 待完成的演进项登记在 [02-requirements/roadmap/](../02-requirements/roadmap/01-implementation-roadmap.md) 与 [backend-optimization/](../backend-optimization/)。
- 设计文档的实现偏差清单在对应 LLD 的"实现偏差"小节；状态以 [03-knowledge/wiki/01-design-progress.md](../03-knowledge/wiki/01-design-progress.md) 为准。
- 每次审查后，对设计层的取舍与变更应回写 `adr/`。
