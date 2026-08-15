# 后端 API 优化 Backlog（2026-08-15 起）

本目录记录 2026-08-15 后端 API 审查后尚未实施的两批优化项，供后续排期使用。
代码已修复的 P0 必修项见 `../04-operations/reviews/01-project-review.md` 的 `2026-08-15 后端 API 修复` 章节。

## 子文档

| 文件 | 内容 | 估算工时 |
| --- | --- | --- |
| [`p1-strongly-recommended.md`](./p1-strongly-recommended.md) | 强烈建议（P1）8 项 | ~5–7 人天 |
| [`p2-improvements.md`](./p2-improvements.md) | 可改进（P2）10 项 | ~6–8 人天 |

## 排期建议

| 里程碑 | 范围 | 触发条件 |
| --- | --- | --- |
| M1（两周内） | P1 §3.1、§3.2、§3.3、§3.7 | 任意生产事故或慢日志告警 |
| M2（一月内） | P1 剩余 4 项 | 完成 M1 验收 |
| M3（季度内） | P2 §4.7、§4.8、§4.1、§4.4 | 服务稳定运行 |
| M4（待评估） | P2 剩余 6 项 | 与新需求对齐 |

## 验证基线

每项落地都需要补充或更新：

1. **单元测试**：handler / service / model 各自冒烟；
2. **集成测试**：至少覆盖一条 happy path + 一条失败 path；
3. **可观测性**：日志 / 指标 / 缓存命中率（若引入新缓存）；
4. **文档同步**：`../01-decisions/lld/01-data-and-api-design.md` / `../04-operations/reviews/01-project-review.md` 同步更新。

任何修改前必须先在独立 MySQL/Redis 实例上验证，**禁止直接对生产数据库执行变更**。
