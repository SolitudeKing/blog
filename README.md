# Solitude Blog

以 Vue 3 与 Go 构建的个人博客系统。`web`、`server` 和 `deploy` 构成运行基线；`demo` 作为博客视觉与交互蓝图长期保留，但不参与生产构建和运行。

## 工程结构

| 目录 | 职责 |
| --- | --- |
| `web/` | Vue 3、Vite、TypeScript 与 Sass 前端，包含公开站点和管理后台 |
| `server/` | Go、Gin、GORM API，负责认证、文章、专题、标签、媒体和站点配置 |
| `deploy/` | Docker Compose、Nginx、健康检查与 MySQL 备份恢复脚本 |
| `docs/` | 当前架构、产品、数据、设计系统、部署与维护文档 |
| `demo/` | 博客落地页与后台的静态设计蓝图，用于后续视觉、布局和交互演进参考 |

MySQL 是业务数据的唯一可信来源，Redis 用于公开文章、站点配置等读取缓存。当前系统不依赖异步任务服务。

## 固定专题

初始化数据库时会确保以下三个专题存在；`label` 和 `slug` 是代码、接口与链接使用的稳定标识，不应随展示文案调整。

| 名称 | Label | Slug | 内容方向 |
| --- | --- | --- | --- |
| 雾里拾笺 | `NODES` | `nodes` | 学习笔记、知识整理与问题记录 |
| 微光造物 | `CODE` | `code` | 编程实践、创意实现与作品复盘 |
| 风过留痕 | `JOTTING` | `jotting` | 随笔、感受与生活片段 |

早期数据库中的完整默认 `Notes / Notes / notes` 会在启动迁移中安全转换为“雾里拾笺”；中间版本的 `categories/category_id` 兼容逻辑仍保留，用于已有数据库升级。

## 使用 Docker Compose 启动

1. 复制 `.env.example` 为 `.env`。
2. 填写 MySQL、Redis、JWT 和管理员必填凭据。
3. 保证 `MYSQL_DSN` 中的用户名、密码与 `MYSQL_USER`、`MYSQL_PASSWORD` 一致。
4. 启动服务：

```bash
docker compose --env-file .env -f deploy/docker-compose.yml up -d --build
sh deploy/scripts/healthcheck.sh
```

默认只有 Nginx 的 `${WEB_PORT:-80}` 面向外部监听。MySQL、Redis 和 API 的调试端口仅绑定 `127.0.0.1`。仓库内 Nginx 配置不终止 TLS，生产环境应由边缘代理或平台负责 HTTPS。

## 本机开发

先通过 Compose 启动依赖：

```bash
docker compose --env-file .env -f deploy/docker-compose.yml up -d mysql redis
```

如果 API 在宿主机运行，需要在本机环境中将连接地址改为 `127.0.0.1`，不要直接沿用容器网络中的 `mysql`、`redis` 主机名。

```bash
cd server
go run ./cmd/api

cd ../web
npm ci
npm run dev
```

除 `/healthz`、`/rss.xml` 和 `/sitemap.xml` 外，API 请求需携带版本头：

```http
X-API-Version: v1
```

## 验证

```bash
cd server
go test ./...

cd ../web
npm run build

cd ..
sh -n deploy/scripts/backup-mysql.sh deploy/scripts/healthcheck.sh deploy/scripts/restore-mysql.sh
docker compose --env-file .env -f deploy/docker-compose.yml config --quiet
```

当前机器若没有 Docker CLI，只能完成源码与脚本级验证；Compose 展开、容器健康检查和备份恢复必须在部署环境继续验收。

## 维护入口

- [文档索引](./docs/README.md)
- [维护指南](./docs/12-maintenance-guide.md)
- [本轮项目审查](./docs/13-project-review.md)
- [部署与备份手册](./docs/08-deployment-runbook.md)

不要提交 `.env`、数据库备份、上传文件、构建产物或本地迁移产物。清理数据前先备份，并通过稳定 slug、引用关系和实际环境查询确认对象，不能仅凭本机历史 ID 执行删除。
