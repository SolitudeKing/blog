#!/bin/sh
# API 容器入口脚本：把宿主机挂载进来的二进制拷贝到镜像内可写路径，
# 修正属主与权限后降权到 UID 10001 执行。
#
# 为什么不在挂载点直接执行：
#   1. 宿主机经 scp 落盘的二进制 UID 与容器内的 app 用户（10001）往往不一致，
#      在 :ro bind mount 上无法 chown；
#   2. 不同环境（Windows scp、root 同步、CI tar）落盘的属主差异大；
#   3. 把拷贝 + chown 放在 entrypoint，部署脚本与运维不再关心宿主机文件属主。
#
# 通过 API_BIN_SRC / API_BIN_DST 环境变量可覆盖默认路径，默认与
# deploy/docker-compose.runtime.yml 的 bind mount 对齐。

set -eu

SRC="${API_BIN_SRC:-/app/api}"
DST="${API_BIN_DST:-/app/runtime/bin/api}"

if [ ! -f "$SRC" ]; then
  echo "api-entrypoint: artifact not found at $SRC" >&2
  echo "  bind-mount the artifact directory and retry." >&2
  exit 1
fi

cp -f "$SRC" "$DST"
chown app:app "$DST"
chmod 0755 "$DST"

# setpriv 来自 util-linux（在 server/Dockerfile.runtime 安装）；
# --init-groups 复制目标用户的附加组（这里没有，但保持对称）。
exec setpriv --reuid=10001 --regid=10001 --init-groups -- "$DST"