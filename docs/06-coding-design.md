# 编码设计

## 目标

- 让目录职责和依赖方向可预测。
- 让数据不变量集中在领域层，不散落在页面和 Handler。
- 通过清晰命名、类型和必要中文注释降低后续维护成本。
- 当前源码与文档保持一致，不用空脚手架表达未来计划。

## 顶层结构

```text
.
├── web/                 # Vue 前端
├── server/              # Go API
├── deploy/              # Compose、Nginx 与运维脚本
├── docs/                # 当前契约和维护记录
├── .env.example         # 可提交的配置模板
├── .dockerignore        # Web 构建上下文边界
└── README.md            # 启动与维护入口
```

本地 `.env`、上传文件、备份、迁移产物、依赖和构建输出不得提交。

## 通用规则

- 一个文件聚焦一个明确职责；重复逻辑先判断是否属于共享领域规则，再决定抽取位置。
- 稳定标识使用常量统一定义。三专题 label/slug 不在页面中重复硬编码。
- 错误在最接近业务语义的层转换；日志保留上下文但不输出密钥、Token 或正文隐私数据。
- 配置错误应在启动时失败，不通过不安全默认值悄悄运行。
- 删除兼容代码前先列出所有持久化环境的升级证据和回滚方案。

## 前端

```text
web/src/
├── api/                 # HTTP client 与接口模块
├── components/          # base、blog、admin 组件
├── composables/         # 可复用状态与副作用
├── config/              # 稳定展示/领域映射
├── layouts/             # 公开与后台壳层
├── pages/               # 路由页面
├── router/              # 路由和守卫
├── stores/              # Pinia 状态
├── styles/              # token、主题与页面样式
├── types/               # API 与领域类型
└── utils/               # 无状态工具
```

约定：

- Vue 组件使用 `<script setup lang="ts">`。
- 页面负责流程编排，基础组件不请求业务 API。
- API 响应和表单 payload 提供 TypeScript 类型。
- 请求竞态由取消、序号或明确状态机处理，旧响应不能覆盖新查询。
- 主题颜色只来自语义 token；页面样式不复制具体主题色值。
- 外部链接先经过协议白名单处理，再设置 `target`/`rel`。

## 后端

```text
server/
├── cmd/api/             # 进程入口
└── internal/
    ├── appearance/      # 主题领域规则
    ├── bootstrap/       # 资源装配和日志
    ├── cache/           # 缓存键
    ├── config/          # 配置加载与启动校验
    ├── database/        # 连接、升级和初始数据
    ├── errors/          # 业务错误
    ├── handler/         # HTTP 适配
    ├── middleware/      # 请求中间件
    ├── model/           # GORM 模型与领域常量
    ├── pagination/      # 分页契约
    ├── response/        # 统一响应
    ├── router/          # 路由注册
    └── service/         # 业务规则
```

依赖方向：

```text
router -> handler -> service -> model/database/cache
```

- Handler 不直接写 SQL。
- Service 控制业务校验、关联处理、事务边界和缓存失效。
- Model 不依赖 HTTP 请求或响应语义。
- Database 中的当前版本兼容升级保持幂等，并用注释说明保留原因。
- GORM 错误统一转换为稳定业务错误，不把原始数据库错误直接返回前端。

## 配置

配置模板以根 `.env.example` 为准。当前关键变量：

```env
APP_ENV=development
APP_PORT=8080
APP_API_VERSION=v1
# Compose 使用 http://localhost；宿主机 Vite 开发时使用 http://localhost:5173。
SITE_BASE_URL=http://localhost
MYSQL_HOST=127.0.0.1
MYSQL_PORT=3306
MYSQL_DATABASE=blog
MYSQL_USER=blog
MYSQL_PASSWORD=...
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=...
JWT_ACCESS_SECRET=...
JWT_REFRESH_SECRET=...
ADMIN_USERNAME=admin
ADMIN_PASSWORD=...
STORAGE_LOCAL_ROOT=./storage/uploads
```

`.env` 只记录原子连接字段，Go 配置层负责组合 `MySQLDSN` 和 `RedisAddr`，调用方仍使用统一的完整配置。Compose 为 API 容器显式覆盖 `MYSQL_HOST=mysql`、`MYSQL_PORT=3306`、`REDIS_HOST=redis`、`REDIS_PORT=6379`；API 在宿主机运行时使用 `127.0.0.1` 和 Compose 映射到宿主机的端口。

当前没有 Celery 或其他异步任务服务，因此 `.env` 不记录 `CELERY_BROKER_URL`、`CELERY_RESULT_BACKEND` 等无消费者配置。实际 `.env` 不得写入文档或提交。

## 注释约定

- 优先使用中文解释“为什么”和“什么条件下可移除”。
- 事务顺序、兼容迁移、安全校验、缓存失效和降级边界必须有必要说明。
- 不逐行翻译代码，不给显而易见的 getter、赋值或循环添加噪声注释。
- TODO 写清问题与验收条件；长期设想放入路线图。
- 行为改变时同步修改注释和文档。

## 异步能力边界

异步任务不在当前运行基线。未来如需图片转换、批量索引或统计聚合，应按实际瓶颈重新设计，并至少具备：

- 明确生产者和版本化 payload。
- 幂等消费者、超时、有限重试和死信处理。
- 任务状态、结构化日志、指标和告警。
- 部署健康检查与失败恢复手册。
- 与同步事务之间清晰的一致性策略。

不得用返回固定占位状态的任务或仅能 ping 的进程代替真实能力。

## 测试

### Go

- 领域归一化、配置校验和迁移兼容使用单元测试。
- Service 覆盖成功、未找到、冲突、无效关联和数据库故障。
- 涉及结构升级时在旧 schema 副本上执行迁移测试。

### Web

- 当前最低检查为 `npm run typecheck` 和 `npm run build`。
- 后续补充工具函数、stores、关键组件和路由流程测试。
- 登录、写作、发布、删除与媒体管理建立端到端测试。

### 部署

- Shell 脚本先执行 `sh -n`。
- Compose 在有 Docker 的环境执行 `config --quiet`、构建、启动和健康检查。
- 备份必须通过完整性检查，并在隔离环境完成恢复演练。

完整命令见 [维护指南](./12-maintenance-guide.md)。
