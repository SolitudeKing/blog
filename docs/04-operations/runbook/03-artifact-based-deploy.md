# 制品挂载部署（Artifact-based deploy）

> 本附录说明仓库当前推荐的**默认**部署路径：`deploy/docker-compose.runtime.yml` + 预构建的运行时镜像 + 宿主机 bind mount 的制品。源码构建路径（`deploy/docker-compose.yml` + `--build`）保留作为**回退**，详见 §5。

## 1. 为什么改

源码构建路径（`docker compose up -d --build`）每次升级都重做以下高耗时工作：

- `golang:1.26-alpine` 拉 Go module 缓存 + 完整重编译 API 二进制；
- `node:22-alpine` 跑 `npm ci` + `vue-tsc` + Vite 重新打包 SPA。

即便 Docker layer cache 命中，源代码层任一文件变动都会让 `COPY web/` / `COPY ./internal` 之上的缓存失效，导致每次冷启动仍是数分钟。

**核心观察**：运行环境（alpine + ca-certificates；nginx + 配置）几乎不变，只有应用代码在变。把两者解耦：

1. **运行时镜像**：仅在 alpine / nginx / 系统包升级时重建一次（数月一次）。
2. **应用制品**：每次代码变动时在开发机本地构建，再 scp / rsync 到部署机（秒级）。
3. **容器启动**：bind mount 制品 + entrypoint 修正属主 / 权限后启动，无需 `docker build`。

## 2. 仓库内文件

```
deploy/
├── docker-compose.yml              # 源码构建路径（回退，--build）
├── docker-compose.runtime.yml      # 制品挂载路径（默认，image）
├── runtime/                        # 宿主机端制品目录（不入版本库）
│   ├── api/                        #   └─ Go 二进制（容器 bind mount 进 /app/api）
│   └── web/                        #   └─ SPA dist（容器 bind mount 进 /usr/share/nginx/html）
├── scripts/
│   ├── api-entrypoint.sh           # API 容器入口（拷贝 + chown + setpriv 降权）
│   ├── build-artifacts.ps1         # Windows 主机构建脚本
│   ├── build-artifacts.sh          # POSIX 等价脚本
│   ├── deploy.ps1                  # Windows 主机推送 + 重启
│   ├── deploy.sh                   # POSIX 等价推送 + 重启
│   └── healthcheck.sh              # 兼容 COMPOSE_FILE 覆盖（line 11）
└── nginx/default.conf              # 同时被 web/Dockerfile 与 web/Dockerfile.runtime 使用

server/
├── Dockerfile                      # 源码构建（回退）
└── Dockerfile.runtime              # 运行时镜像（默认）

web/
├── Dockerfile                      # 源码构建（回退）
└── Dockerfile.runtime              # 运行时镜像（默认）
```

`deploy/runtime/api/` 与 `deploy/runtime/web/` 由 `.gitignore` 排除，不入版本库。

## 3. 一次性服务器初始化

仅在「首次升级到新流程」或「运行时镜像需要重建」时执行：

```bash
cd /opt/solitude-blog

# API 运行时：alpine + ca-certificates + tzdata + util-linux + 非 root app + entrypoint
docker build -t solitude-blog-api-runtime:1.0.0 -f server/Dockerfile.runtime server/

# SPA 运行时：nginx:alpine + deploy/nginx/default.conf（默认 html 已被清空）
docker build -t solitude-blog-nginx-runtime:1.0.0 -f web/Dockerfile.runtime .

docker images | grep solitude-blog-runtime   # 两个 tag 都应出现
```

升级镜像版本后，记得同步修改 `deploy/docker-compose.runtime.yml` 中对应的 `image:` 行。

## 4. 日常升级流程

### 4.1 Windows 主机（推荐）

```powershell
# 在仓库根目录
.\deploy\scripts\deploy.ps1 -Host ubuntu@1.2.3.4

# 制品已构建好、只想推送 + 重启：
.\deploy\scripts\deploy.ps1 -Host ubuntu@1.2.3.4 -SkipBuild

# 紧急修复回退到源码构建：
.\deploy\scripts\deploy.ps1 -Host ubuntu@1.2.3.4 -Fallback
```

### 4.2 Linux / macOS 主机

```bash
deploy/scripts/deploy.sh ubuntu@1.2.3.4
deploy/scripts/deploy.sh ubuntu@1.2.3.4 /srv/blog --skip-build
deploy/scripts/deploy.sh ubuntu@1.2.3.4 --fallback
```

脚本串起：

1. 调 `build-artifacts.{ps1,sh}` 本机构建（除非 `-SkipBuild`）。
2. `scp` / `rsync` `deploy/runtime/api/`、`deploy/runtime/web/`、`.env` 到服务器。
3. ssh 到服务器执行 `docker compose --env-file .env -f deploy/docker-compose.runtime.yml up -d`（**不带 `--build`**）。
4. 透传 `COMPOSE_FILE` 给 `healthcheck.sh` 验证容器状态与 `/healthz`。

### 4.3 仅服务器侧手工执行

```bash
# 服务器上手动 scp 完制品后：
cd /opt/solitude-blog
docker compose --env-file .env -f deploy/docker-compose.runtime.yml up -d
sh deploy/scripts/healthcheck.sh
```

## 5. 何时需要重建运行时镜像

| 变更 | 是否重建？ |
| --- | --- |
| 升级 `alpine` / `nginx` 基础镜像 | 是 |
| 给 api 镜像新增系统包（如 `curl`） | 是 |
| 修改 `deploy/nginx/default.conf` | 是（nginx 运行时） |
| 修改 `deploy/scripts/api-entrypoint.sh` | 是（api 运行时） |
| 改 Go / Vue 源代码 | **否**，仅重建制品 |
| 改 `.env` | **否**，仅 scp `.env` |

## 6. 为什么 entrypoint 要拷贝二进制

`deploy/runtime/api-entrypoint.sh` 不直接在 `/app/api`（bind mount 路径）执行，而是：

```sh
cp -f "$SRC" "$DST"      # /app/api -> /app/runtime/bin/api（镜像内可写路径）
chown app:app "$DST"     # 修正属主为 10001:10001
chmod 0755 "$DST"
exec setpriv --reuid=10001 --regid=10001 --init-groups -- "$DST"
```

原因是三种现实场景会在 mount 点直接执行时崩：

1. **Windows scp 落盘的属主是 root（UID 0）或 nobody**：在 `:ro` bind mount 上无法 `chown`。
2. **不同主机/CI 的默认 UID 差异**：用拷贝代替就地 chown 后，所有权差异在 entrypoint 一次性消化掉。
3. **临时调试或回退**：用拷贝路径后，部署脚本不再要求宿主机二进制是特定属主。

`setpriv` 来自 `util-linux`（在 `server/Dockerfile.runtime` 通过 `apk add util-linux` 安装），比 `gosu` / `su-exec` 更轻量。

## 7. 故障排查

| 现象 | 原因 / 处置 |
| --- | --- |
| 容器退出，日志 `api-entrypoint: artifact not found at /app/api` | 服务器上 `deploy/runtime/api/api` 不存在；`scp` / `rsync` 没成功（看权限或路径）；或 `docker-compose.runtime.yml` 的 bind mount 源路径写错 |
| SPA 加载正常，但页面是「Welcome to nginx」 | 运行时镜像在构建时没执行 `rm -rf /usr/share/nginx/html/*`；或 bind mount 缺失（容器没有 `runtime/web` 源）—— 重建镜像或检查卷 |
| SPA 静态资源 403 | `deploy/runtime/web/` 在服务器上权限过紧（`0700` 等）。nginx worker 以 `nginx` 用户运行；执行 `chmod -R a+rX deploy/runtime/web`；或在 `build-artifacts.{ps1,sh}` 末尾固定 `chmod -R a+rX`（PS1 已经调用 `icacls ... /grant Everyone:RX /T`） |
| `setpriv: command not found` | 运行时镜像在构建时缺 `apk add util-linux`；按 §3 重建 |
| 容器 `running` 但 `healthy` 始终 false，healthz 502 | 不是制品挂载的回归：沿用 [01-deployment-and-backup.md §4](./01-deployment-and-backup.md) "健康检查"段诊断（外部 MySQL/Redis 不可达等） |
| 想验证回退路径仍可用 | `docker compose --env-file .env -f deploy/docker-compose.yml up -d --build` 后跑 `healthcheck.sh`；通过即可视为回退路径完好 |

## 8. 与其他文档的关系

- 端到端首次部署：[`00-deploy-from-scratch.md`](./00-deploy-from-scratch.md)，特别是 §8.5 与 §12。
- Compose / 健康检查 / 备份 / 回滚：[`01-deployment-and-backup.md`](./01-deployment-and-backup.md)。
- 密钥轮换：[`02-secret-rotation.md`](./02-secret-rotation.md)。