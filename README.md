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

