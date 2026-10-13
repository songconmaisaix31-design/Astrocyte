param([string]$WebDir = '', [int]$Port = 8787, [switch]$SkipSnapshot)
$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
$runtime = Join-Path $env:LOCALAPPDATA 'astrocyte-workbench/v0.1'
New-Item -ItemType Directory -Force $runtime, (Join-Path $runtime 'scripts'), (Join-Path $runtime 'logs'), (Join-Path $runtime 'bin') | Out-Null
$buildWeb = !$WebDir
if ($buildWeb) { $WebDir = Join-Path $repo 'web/dist' }
$WebDir = [System.IO.Path]::GetFullPath($WebDir)
$exe = Join-Path $runtime 'bin/astrocyte-workbench.exe'
$existing = Get-CimInstance Win32_Process -Filter "Name = 'astrocyte-workbench.exe'" | Where-Object { $_.ExecutablePath -eq $exe }
if ($existing) { throw 'Owned workbench already running; stop it before rebuilding.' }
# Check before building or changing configuration; never stop another instance.
$listener = Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction SilentlyContinue
if ($listener) { throw 'Requested port occupied; use -Port 17321 or another free port.' }
$configPath = Join-Path $runtime 'config.json'
$config = [ordered]@{
    root = $runtime; web = $WebDir; port = $Port
    python = (Get-Command python.exe -ErrorAction Stop).Source
    node = (Get-Command node.exe -ErrorAction Stop).Source
    summarize = (Resolve-Path (Join-Path $repo 'node_modules/@steipete/summarize/dist/cli.js')).Path
}
$config | ConvertTo-Json | Set-Content -LiteralPath $configPath -Encoding UTF8
foreach ($name in @('snapshot.py','refresh.ps1','resume.ps1','audit.py','monitor.ps1','monitored-refresh.ps1')) { Copy-Item -LiteralPath (Join-Path $PSScriptRoot $name) -Destination (Join-Path $runtime "scripts/$name") -Force }
Push-Location $repo
try {
    go build -mod=readonly -trimpath -o $exe ./cmd/server
    if ($LASTEXITCODE -ne 0) { throw 'Server build failed.' }
    if ($buildWeb) {
        pnpm --dir web build
        if ($LASTEXITCODE -ne 0) { throw 'Web build failed.' }
    }
    if (!(Test-Path -LiteralPath (Join-Path $WebDir 'index.html'))) { throw 'Built web index unavailable.' }
} finally { Pop-Location }
if (!$SkipSnapshot) {
    & powershell.exe -NoProfile -NonInteractive -ExecutionPolicy Bypass -File (Join-Path $runtime 'scripts/refresh.ps1')
    if ($LASTEXITCODE -ne 0) { throw 'Snapshot failed.' }
}
& (Join-Path $runtime 'scripts/resume.ps1')
& (Join-Path $PSScriptRoot 'install-tasks.ps1')
