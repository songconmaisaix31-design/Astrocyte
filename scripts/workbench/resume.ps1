$ErrorActionPreference = 'Stop'
$runtime = Join-Path $env:LOCALAPPDATA 'astrocyte-workbench/v0.1'
$config = Get-Content -LiteralPath (Join-Path $runtime 'config.json') -Raw | ConvertFrom-Json
$exe = Join-Path $runtime 'bin/astrocyte-workbench.exe'
$existing = Get-CimInstance Win32_Process -Filter "Name = 'astrocyte-workbench.exe'" | Where-Object { $_.ExecutablePath -eq $exe }
if (!$existing) {
    if (Get-NetTCPConnection -State Listen -LocalPort $config.port -ErrorAction SilentlyContinue) { throw 'Configured port occupied.' }
    $env:ASTROCYTE_WORKBENCH = '1'
    $env:ASTROCYTE_DATA_DIR = Join-Path $runtime 'app-state'
    $env:ASTROCYTE_WEB_DIR = $config.web
    $env:ASTROCYTE_PORT = [string]$config.port
    $env:ASTROCYTE_ENABLE_CODEX_DISTILLATION = 'false'
    $env:ASTROCYTE_NODE = $config.node
    $env:ASTROCYTE_SUMMARIZE_CLI = $config.summarize
    $p = Start-Process -FilePath $exe -WorkingDirectory $runtime -WindowStyle Hidden -RedirectStandardOutput (Join-Path $runtime 'logs/server.log') -RedirectStandardError (Join-Path $runtime 'logs/server-error.log') -PassThru
    $p.Id | Set-Content -LiteralPath (Join-Path $runtime 'server.pid')
}
$healthy = $false
$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession
for ($i=0; $i -lt 120; $i++) {
    if ($p) { $p.Refresh(); if ($p.HasExited) { break } }
    try {
        Invoke-RestMethod "http://127.0.0.1:$($config.port)/api/v1/auth/session" -WebSession $session -TimeoutSec 2 | Out-Null
        $health = Invoke-RestMethod "http://127.0.0.1:$($config.port)/api/v1/cognition/health" -WebSession $session -TimeoutSec 2
        if ($health.status -eq 'ok') { $healthy = $true; break }
    } catch { Start-Sleep -Milliseconds 250 }
}
if (!$healthy) { throw 'Workbench session health check failed; inspect owned runtime logs.' }
Write-Output "Workbench ready: http://127.0.0.1:$($config.port)/cognition/today"
