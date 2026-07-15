# Solitude Blog

新个人博客系统工程骨架。旧项目保留在 `blog-mini-serve` 与 `blog-mini-v3`，新实现按文档拆分为：

- `web`：Vue3 + Vite + TypeScript + Sass 前端。
- `server`：Go + Gin API。
- `worker`：Celery 异步任务。
- `deploy`：Docker Compose 与 Nginx 配置。
- `docs`：设计与进度文档。

## 当前阶段

当前已完成 M4 可迁移上线阶段的代码收口，已完成：

- Go API 基础路由、统一响应、错误码、API 版本头、游标分页结构。
- Go API 自动读取 `.env`，并接入 MySQL/GORM 自动迁移、Redis 健康检查、管理员初始化、JWT 登录与鉴权。
- 文章接口已支持数据库优先的列表、详情、创建、更新、删除；数据库不可用时保留内存示例数据降级。
- 分类与标签已支持后台 CRUD，并可在文章编辑器中选择。
- 前台文章列表已支持游标分页、空状态和错误重试；文章详情已支持 Markdown 渲染和目录导航。
- 前台文章列表、文章详情和站点配置已接入 Redis 缓存与写入失效。
- 站点配置、公告管理、后台仪表盘和媒体资源管理已完成后台日常维护能力。
- 旧博客迁移链路已支持 SQLite、Markdown、PicBed Base64 和图片文件导出/导入，并生成迁移报告。
- Vue 前端 API client、路由、Pinia，以及 Mist UI 海盐/青森二维主题 token 与自研组件。
- Celery worker 基础配置与任务命名。
- Docker Compose、Nginx、健康检查、Redis AOF、上传文件卷、日志限制、MySQL 备份与恢复脚本。
- M5 已完成，公开站点已支持文章搜索、`/rss.xml` 和 `/sitemap.xml`；后台已支持增强统计面板、文章版本记录和编辑器自动保存。

本机当前缺少 Docker CLI，Compose 实机展开和容器级健康检查需要在部署机继续执行。

## 本地开发

依赖安装需要联网：

```bash
cd server
go mod tidy

cd ../web
npm install

cd ../worker
pip install -e .
```

启动顺序建议：

```bash
cd deploy
docker compose up -d mysql redis

cd ../server
go run ./cmd/api

cd ../web
npm run dev
```

API 请求必须携带：

```http
X-API-Version: v1
```

首次连接 MySQL 时，服务会自动迁移基础表，并按 `.env` 中的 `ADMIN_USERNAME`、`ADMIN_PASSWORD`、`ADMIN_NICKNAME` 初始化管理员账号。
