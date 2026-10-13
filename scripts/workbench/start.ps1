param([string]$SpaDir = '', [int]$Port = 17321, [switch]$SkipSnapshot)
$ErrorActionPreference = 'Stop'
$repo = (Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
$runtime = Join-Path $env:LOCALAPPDATA 'astrocyte-workbench/v0.1'
New-Item -ItemType Directory -Force $runtime, (Join-Path $runtime 'scripts'), (Join-Path $runtime 'logs') | Out-Null
if (!$SpaDir) { $SpaDir = Join-Path $repo 'workbench/frontend/dist' }
$SpaDir = [System.IO.Path]::GetFullPath($SpaDir)
$exe = Join-Path $runtime 'workbench.exe'
$existing = Get-CimInstance Win32_Process -Filter "Name = 'workbench.exe'" | Where-Object { $_.ExecutablePath -eq $exe }
if ($existing) { throw 'Workbench already running; stop the owned process before rebuilding.' }
$configPath = Join-Path $runtime 'config.json'
$config = [ordered]@{
    root = $runtime; spa = $SpaDir; port = $Port
    source = 'C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v4-w1/formal-20261013-first/state.sqlite'
    history = 'C:/Users/DW/AppData/Local/astrocyte-validation/v1-20261012/v1r-w1/replay-002/topic-timeseries-safe.json'
    python = (Get-Command python.exe -ErrorAction Stop).Source
}
$config | ConvertTo-Json | Set-Content -LiteralPath $configPath -Encoding UTF8
Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'snapshot.py') -Destination (Join-Path $runtime 'scripts/snapshot.py') -Force
Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'refresh.ps1') -Destination (Join-Path $runtime 'scripts/refresh.ps1') -Force
Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'resume.ps1') -Destination (Join-Path $runtime 'scripts/resume.ps1') -Force
Push-Location (Join-Path $repo 'workbench/backend')
try { go build -trimpath -o $exe .; if ($LASTEXITCODE -ne 0) { throw 'Build failed.' } } finally { Pop-Location }
if (!$SkipSnapshot) { & (Join-Path $runtime 'scripts/refresh.ps1'); if ($LASTEXITCODE -ne 0) { throw 'Snapshot failed.' } }
& (Join-Path $runtime 'scripts/resume.ps1')
