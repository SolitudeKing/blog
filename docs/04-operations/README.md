# 运维与事故

本目录记录"如何稳定运行"与"出过什么问题、怎么修的"。

- `runbook/` 运维手册：部署、启动、健康检查、备份、恢复、回滚。
- `reviews/` 审查与复盘：周期性审查、已修复项、验证状态、剩余风险。
- `incidents/` 事故报告：预留事故发生时的第一时间记录与影响面评估。
- `rca/` 根因分析报告：预留对事故的根因分析与改进措施。

## 当前内容

- [Runbook · 端到端部署指南](./runbook/00-deploy-from-scratch.md)：从零到对外可访问的部署 playbook（`git clone` → 配 `deploy/.env.production` → 启动 → 验收）。
- [Runbook · 部署与备份](./runbook/01-deployment-and-backup.md)：Compose、Nginx、健康检查、备份恢复、TLS、回滚原则（参考手册）。
- [Runbook · 密钥轮换](./runbook/02-secret-rotation.md)：JWT / MySQL / Redis / Admin / S3 五个密钥的轮换步骤、影响面与回滚。
- [复盘 · 项目审查](./reviews/01-project-review.md)：本轮已修复项与剩余 P1 / P2 风险。

## 新增文档的落点

- **事故发生 24 小时内**：在 `incidents/` 落事故时间线、影响面、止血措施，即使根因未明也先记录现象。
- **根因明确后**：把完整分析、长期改进写入 `rca/`，并把引用关系写回事故报告。
- **周期性审查**：每轮审查结束后更新 `reviews/` 下的对应文件，文件名带日期前缀（如 `01-project-review.md` / `02-xxx.md`），并在 `README.md` 文档地图中追加链接。

## 与其他目录的协作

- 复盘中识别的演进项登记到 [02-requirements/roadmap/](../02-requirements/roadmap/01-implementation-roadmap.md) 或 [backend-optimization/](../backend-optimization/)。
- 复盘中识别的设计层取舍与变更，应回写 [01-decisions/adr/](../01-decisions/adr/)。
- 维护命令与目录边界以 [03-knowledge/wiki/02-maintenance-guide.md](../03-knowledge/wiki/02-maintenance-guide.md) 为入口。
