# Solitude Blog

以 Vue 3 与 Go 构建的个人博客系统。`web`、`server` 和 `deploy` 构成运行基线；`demo` 作为博客视觉与交互蓝图长期保留，但不参与生产构建和运行。

## 工程结构

| 目录 | 职责 |
| --- | --- |
| `web/` | Vue 3、Vite、TypeScript 与 Sass 前端，包含公开站点和管理后台 |
| `server/` | Go、Gin、GORM API，负责认证、文章、专题、标签、媒体和站点配置 |
| `deploy/` | Docker Compose、Nginx 与健康检查脚本（MySQL / Redis 由外部托管服务提供，不在本仓库内启动；备份由托管方负责） |
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

早期数据库中的完整默认 `Notes / Notes / notes` 会在启动迁移中安全转换为"雾里拾笺"；中间版本的 `categories/category_id` 兼容逻辑仍保留，用于已有数据库升级。

## 使用 Docker Compose 启动

1. 复制 `.env.example` 为 `.env`。
2. 填写 MySQL / Redis / JWT / 管理员必填凭据。MySQL 与 Redis 只需填写 `MYSQL_HOST`、`MYSQL_PORT`、`MYSQL_DATABASE`、`MYSQL_USER`、`MYSQL_PASSWORD`、`MYSQL_TLS` 与 `REDIS_HOST`、`REDIS_PORT`、`REDIS_USER`、`REDIS_PASSWORD`、`REDIS_TLS` 等原子变量，Go 配置层会组合连接参数。**外部 MySQL / Redis 必须由运维在部署机之外提供**（参见 `docs/08-deployment-runbook.md` §1 前置条件）。
3. 启动服务：

```bash
docker compose --env-file .env -f deploy/docker-compose.yml up -d --build
sh deploy/scripts/healthcheck.sh
```

默认只有 Nginx 的 `${WEB_PORT:-80}` 面向外部监听；API 调试端口仅绑定 `127.0.0.1`。仓库内 Nginx 配置不终止 TLS，生产环境应由边缘代理或平台负责 HTTPS。
Compose 仅打包 `api` 与 `nginx` 两个服务；外部 MySQL / Redis 通过 `.env` 注入到 API 容器，由 API 直连。

## 本机开发

本机开发需要本地或团队共用的外部 MySQL / Redis 实例。把 `MYSQL_HOST` / `REDIS_HOST` 指向该实例的 IP / 域名，并把端口、数据库名、密码填入本地 `.env`。**本仓库不再提供 `docker compose up -d mysql redis` 这种自托管命令**，开发数据库请自行准备。

如果 API 在宿主机运行，使用 `MYSQL_HOST=<dev-host>`、`REDIS_HOST=<dev-host>`，**不要使用仅能在容器网络解析的 `mysql` / `redis` 主机名**（这些主机名已经不存在）。

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

### 访问上传后的静态资源

`STORAGE_DRIVER=local` 时，本机开发无需启动 Nginx 即可访问上传文件：

| 入口 | URL 形式 | 适用场景 |
| --- | --- | --- |
| Vite dev server（推荐） | `http://localhost:5173/uploads/...` | 前端页面 `<img>`、fetch、`<a href>` |
| API 进程直打（dev 限定） | `http://localhost:8080/uploads/...` | Postman、浏览器调试、跨域脚本 |

- 后端在 `internal/router/router.go` 已用 `r.Static("/uploads", cfg.StorageLocalRoot)` 直接挂载；dev 模式下 `CorsForDev` 中间件会自动放行所有 origin
- Vite proxy（`web/vite.config.ts`）会把 `/uploads` 转发到 `http://localhost:8080`，前端无需特殊配置
- 文件实际写入位置：`./server/storage/uploads/YYYY/MM/<rand>.<ext>`（与 `STORAGE_LOCAL_ROOT` 同步）
- 上传后图片如果显示损坏，请确认 Commit `fix(upload): 修复 sniff 后 reader 未重置` 已应用（修复前 512 字节 magic header 丢失）

## 验证

```bash
cd server
go test ./...

cd ../web
npm run build

cd ..
sh -n deploy/scripts/healthcheck.sh
docker compose --env-file .env -f deploy/docker-compose.yml config --quiet
docker compose --env-file .env -f deploy/docker-compose.yml config --services
```

`config --services` 应仅输出 `api` 与 `nginx`；如出现 `mysql` / `redis`，说明 `deploy/docker-compose.yml` 仍包含旧服务定义。

当前机器若没有 Docker CLI，只能完成源码与脚本级验证；Compose 展开、容器健康检查与外部依赖连通性必须在部署环境继续验收。

## 维护入口

- [文档索引](./docs/README.md)
- [维护指南](./docs/12-maintenance-guide.md)
- [本轮项目审查](./docs/13-project-review.md)
- [后端 API 优化 Backlog](./docs/backend-optimization/README.md)
- [部署与备份手册](./docs/08-deployment-runbook.md)

不要提交 `.env`、数据库备份、上传文件、构建产物或本地迁移产物。清理数据前先备份，并通过稳定 slug、引用关系和实际环境查询确认对象，不能仅凭本机历史 ID 执行删除。