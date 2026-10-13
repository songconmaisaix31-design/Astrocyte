$ErrorActionPreference = 'Stop'
$runtime = Join-Path $env:LOCALAPPDATA 'astrocyte-workbench/v0.1'
try {
    $configPath = Join-Path $runtime 'config.json'
    $config = Get-Content -LiteralPath $configPath -Raw | ConvertFrom-Json
    & $config.python (Join-Path $runtime 'scripts/snapshot.py') --config $configPath
    $result = $LASTEXITCODE
    ('{0} snapshot_exit={1}' -f [DateTimeOffset]::Now.ToString('o'), $result) | Add-Content -LiteralPath (Join-Path $runtime 'logs/snapshot.log')
    exit $result
} catch {
    Write-Output '{"status":"error","code":"SNAPSHOT_UNAVAILABLE"}'
    exit 1
}
