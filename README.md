# Solitude Blog

新个人博客系统工程骨架。旧项目保留在 `blog-mini-serve` 与 `blog-mini-v3`，新实现按文档拆分为：

- `web`：Vue3 + Vite + TypeScript + Sass 前端。
- `server`：Go + Gin API。
- `worker`：Celery 异步任务。
- `deploy`：Docker Compose 与 Nginx 配置。
- `docs`：设计与进度文档。

## 当前阶段

当前处于第一轮编码骨架阶段，已完成：

- Go API 基础路由、统一响应、错误码、API 版本头、游标分页结构。
- Go API 自动读取 `.env`，并接入 MySQL/GORM 自动迁移、Redis 健康检查、管理员初始化、JWT 登录与鉴权。
- 文章接口已支持数据库优先的列表、详情、创建、更新、删除；数据库不可用时保留内存示例数据降级。
- 分类与标签已支持后台 CRUD，并可在文章编辑器中选择。
- 前台文章列表已支持游标分页、空状态和错误重试；文章详情已支持 Markdown 渲染和目录导航。
- Vue 前端 API client、路由、Pinia、基础 CreamyUI token 和示例页面。
- Celery worker 基础配置与任务命名。
- Docker Compose 与 Nginx 初稿。

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
