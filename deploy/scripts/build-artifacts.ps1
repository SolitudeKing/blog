# 构建 Go 二进制与 Vue SPA，把制品输出到 deploy/runtime/{api,web}。
# 在仓库根执行；前置依赖：go 1.26+、node 22+、npm 10+。
#
# 行为：
#   - Go 产出 deploy/runtime/api/api（与 docker-compose.runtime.yml 的 bind mount 对齐）。
#   - Web 若 web/node_modules 已存在则跳过 npm ci（重复构建提速）；删除该目录可强制重装。
#   - 用 robocopy /MIR 把 web/dist 镜像到 deploy/runtime/web，确保旧 hashed 资产被清理。
#
# 与 deploy/scripts/deploy.ps1 配合使用；单独运行也可作为本机烟雾测试。

$ErrorActionPreference = 'Stop'
$root = (Resolve-Path "$PSScriptRoot/../..").Path
$runtime = Join-Path $root 'deploy/runtime'
$apiOut = Join-Path $runtime 'api'
$webOut = Join-Path $runtime 'web'
$webDir = Join-Path $root 'web'

New-Item -ItemType Directory -Force -Path $apiOut | Out-Null
New-Item -ItemType Directory -Force -Path $webOut | Out-Null

Write-Host '[1/2] Building Go API ->' $apiOut
Push-Location $root
try {
  $env:CGO_ENABLED = '0'
  $env:GOOS = 'linux'
  $env:GOFLAGS = '-trimpath'
  # 与 server/Dockerfile 第 11 行保持一致：trimpath + ldflags -s -w -buildid=
  go build -trimpath -ldflags='-s -w -buildid=' -o "$apiOut/api" ./cmd/api
  if ($LASTEXITCODE -ne 0) { throw 'go build failed' }
} finally { Pop-Location }

Write-Host '[2/2] Building Vue SPA ->' $webOut
Push-Location $webDir
try {
  if (-not (Test-Path 'node_modules')) {
    npm ci
    if ($LASTEXITCODE -ne 0) { throw 'npm ci failed' }
  } else {
    Write-Host '  node_modules present; skipping npm ci (delete it to force refresh).'
  }
  npm run build
  if ($LASTEXITCODE -ne 0) { throw 'npm run build failed' }
} finally { Pop-Location }

# robocopy /MIR 镜像同步，删除旧的 hashed 资产；返回码 0-7 都算成功。
$srcDist = Join-Path $webDir 'dist'
robocopy $srcDist $webOut /MIR /NFL /NDL /NJH /NJS /NP | Out-Null
if ($LASTEXITCODE -ge 8) { throw "robocopy failed: $LASTEXITCODE" }

# nginx 工作进程以 nginx 用户运行；显式放宽权限，避免宿主机 0700/0640 阻断 worker 读。
# Windows 不区分 owner，但 scp 到 Linux 后这些 mode 会保留，提前 a+rX 更安全。
icacls $webOut /grant Everyone:(OI)(CI)RX /T | Out-Null

Write-Host 'Artifacts ready in' $runtime
Get-ChildItem -Recurse $runtime | Select-Object -First 20