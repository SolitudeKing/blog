# 在 Windows 主机上把制品推到部署服务器并重启容器，不触发 docker build。
#
# 用法：
#   .\deploy\scripts\deploy.ps1 -Host ubuntu@1.2.3.4
#   .\deploy\scripts\deploy.ps1 -Host ubuntu@1.2.3.4 -RemoteRoot /srv/blog
#   .\deploy\scripts\deploy.ps1 -Host ubuntu@1.2.3.4 -Fallback    # 走源码 build 的回退路径
#   .\deploy\scripts\deploy.ps1 -Host ubuntu@1.2.3.4 -SkipBuild   # 制品已构建好，仅推送
#
# 默认流程：build-artifacts → scp 制品 → ssh 触发 docker compose up -d → healthcheck。
# -Fallback 切换到 deploy/docker-compose.yml + --build 的回退流程（用于紧急修复）。

[CmdletBinding()]
param(
  [Parameter(Mandatory)][string]$Host,
  [string]$RemoteRoot = '/opt/solitude-blog',
  [switch]$Fallback,
  [switch]$SkipBuild
)

$ErrorActionPreference = 'Stop'
$root = (Resolve-Path "$PSScriptRoot/../..").Path
$runtime = Join-Path $root 'deploy/runtime'

# 1. 制品构建（可选跳过）。
if (-not $SkipBuild) {
  & "$PSScriptRoot\build-artifacts.ps1"
  if ($LASTEXITCODE -ne 0) { throw "build-artifacts.ps1 exited $LASTEXITCODE" }
}

$composeFile = if ($Fallback) { 'deploy/docker-compose.yml' } else { 'deploy/docker-compose.runtime.yml' }
$remote = "${Host}:${RemoteRoot}"

# 2. scp 制品到服务器；先确保远端目录存在。
Write-Host "[deploy] ensure remote dirs on $Host"
ssh $Host "mkdir -p $RemoteRoot/deploy/runtime/api $RemoteRoot/deploy/runtime/web"

Write-Host "[deploy] copy API binary to $remote"
scp "$runtime\api\api" "${remote}/deploy/runtime/api/api"

Write-Host "[deploy] copy web dist to $remote"
scp -r "$runtime\web" "${remote}/deploy/runtime"

# 3. .env 是可选的（很多服务器已经手工放好）；存在则同步一份。
$envLocal = Join-Path $root '.env'
if (Test-Path $envLocal) {
  Write-Host "[deploy] copy .env to $remote"
  scp $envLocal "${remote}/.env"
}

# 4. 远端重启容器；带 --build 的回退路径在 -Fallback 时生效。
$composeFlag = if ($Fallback) { '-f deploy/docker-compose.yml up -d --build' } else { '-f deploy/docker-compose.runtime.yml up -d' }
$remoteCmd = "set -e; cd $RemoteRoot && docker compose --env-file .env $composeFlag"
Write-Host "[deploy] $Host`: $remoteCmd"
ssh $Host $remoteCmd

# 5. 远端 healthcheck；COMPOSE_FILE 透传给 healthcheck.sh，让它指向 runtime compose。
$hcCmd = "set -e; cd $RemoteRoot && COMPOSE_FILE=$RemoteRoot/$composeFile sh deploy/scripts/healthcheck.sh"
Write-Host "[deploy] healthcheck on $Host"
ssh $Host $hcCmd