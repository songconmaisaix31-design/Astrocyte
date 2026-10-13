$ErrorActionPreference = 'Stop'
$runtime = Join-Path $env:LOCALAPPDATA 'astrocyte-workbench/v0.1'
$dir = Join-Path $runtime 'logs/filesystem'
New-Item -ItemType Directory -Force $dir | Out-Null
$configPath = Join-Path $runtime 'config.json'
$config = Get-Content -LiteralPath $configPath -Raw | ConvertFrom-Json
$serverPid = [int](Get-Content -LiteralPath (Join-Path $runtime 'server.pid'))
$p = Start-Process -FilePath $config.python -ArgumentList @("`"$(Join-Path $runtime 'scripts/snapshot.py')`"",'--config',"`"$configPath`"") -WindowStyle Hidden -RedirectStandardOutput (Join-Path $dir 'refresh-output.log') -RedirectStandardError (Join-Path $dir 'refresh-error.log') -PassThru
@($serverPid,$p.Id) | Set-Content -LiteralPath (Join-Path $dir 'owned-pids.txt')
$p.WaitForExit()
if ($p.ExitCode -ne 0) { throw 'Monitored refresh failed.' }
& $config.python (Join-Path $runtime 'scripts/audit.py')
if ($LASTEXITCODE -ne 0) { throw 'Monitored API audit failed.' }
