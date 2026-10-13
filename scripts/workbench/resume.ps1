$ErrorActionPreference = 'Stop'
$runtime = Join-Path $env:LOCALAPPDATA 'astrocyte-workbench/v0.1'
$exe = Join-Path $runtime 'workbench.exe'
$existing = Get-CimInstance Win32_Process -Filter "Name = 'workbench.exe'" | Where-Object { $_.ExecutablePath -eq $exe }
if (!$existing) {
    $p = Start-Process -FilePath $exe -WorkingDirectory $runtime -WindowStyle Hidden -RedirectStandardOutput (Join-Path $runtime 'logs/server.log') -RedirectStandardError (Join-Path $runtime 'logs/server-error.log') -PassThru
    $p.Id | Set-Content -LiteralPath (Join-Path $runtime 'server.pid')
}
$config = Get-Content -LiteralPath (Join-Path $runtime 'config.json') -Raw | ConvertFrom-Json
$healthy = $false
for ($i=0; $i -lt 30; $i++) {
    try { $health = Invoke-RestMethod "http://127.0.0.1:$($config.port)/api/health"; if ($health.status -eq 'ok') { $healthy = $true; break } } catch { Start-Sleep -Milliseconds 200 }
}
if (!$healthy) { throw 'Workbench health check failed.' }
Write-Output "Workbench ready: http://127.0.0.1:$($config.port)"
