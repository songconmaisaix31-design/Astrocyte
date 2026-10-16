param(
    [Parameter(Mandatory=$true)][string]$Repository,
    [Parameter(Mandatory=$true)][string]$PublicRef,
    [Parameter(Mandatory=$true)][string]$ActualScope,
    [Parameter(Mandatory=$true)][string]$SleepDay,
    [Parameter(Mandatory=$true)][string]$Timezone,
    [Parameter(Mandatory=$true)][double]$Strength,
    [Parameter(Mandatory=$true)][double]$Decay,
    [Parameter(Mandatory=$true)][double]$InitialConfidence,
    [Parameter(Mandatory=$true)][switch]$AuthorizedPublicReplay
)

$ErrorActionPreference = 'Stop'
$workspace = (Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
$verification = Join-Path $workspace '.verification'
$owned = Join-Path $verification ('public-cognition-' + [guid]::NewGuid().ToString('N'))
$local = Join-Path $owned 'local'
$root = Join-Path $local 'astrocyte-workbench/v0.1'
$source = Join-Path $owned 'app-state/state.sqlite'
New-Item -ItemType Directory -Path $owned -Force | Out-Null
$serverBinary = Join-Path $owned 'server.exe'
$snapshotBinary = Join-Path $owned 'workbench-snapshot.exe'
$collectorBinary = Join-Path $owned 'actual-public-cognition.exe'
$sleepBinary = Join-Path $owned 'sleep.exe'
$processes = @()
$savedEnvironment = @{}

function Check-Exit([string]$operation) {
    if ($LASTEXITCODE -ne 0) { throw "$operation failed with exit $LASTEXITCODE" }
}
function Wait-File([string]$path, $process) {
    $deadline = [DateTime]::UtcNow.AddMinutes(5)
    while (-not (Test-Path -LiteralPath $path)) {
        $process.Refresh()
        if ($process.HasExited) { throw "Actual process exited before $path" }
        if ([DateTime]::UtcNow -gt $deadline) { throw 'Actual writer readiness timed out' }
        Start-Sleep -Milliseconds 100
    }
}

try {
    Push-Location $workspace
    go build -mod=readonly -o $serverBinary ./cmd/server
    Check-Exit 'server build'
    go build -mod=readonly -o $snapshotBinary ./cmd/workbench-snapshot
    Check-Exit 'snapshot build'
    go build -mod=readonly -o $collectorBinary ./cmd/workbench-snapshot/verify
    Check-Exit 'collector build'
    go build -mod=readonly -o $sleepBinary ./cmd/sleep
    Check-Exit 'Sleep build'

    # 当前验证进程显式隔离所有产品配置，子进程不继承私人范围或模型授权。
    Get-ChildItem Env: | Where-Object { $_.Name.StartsWith('ASTROCYTE_') -or $_.Name -eq 'LOCALAPPDATA' } | ForEach-Object {
        $savedEnvironment[$_.Name] = $_.Value
        [Environment]::SetEnvironmentVariable($_.Name, $null, 'Process')
    }
    $listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, 0)
    $listener.Start()
    $port = $listener.LocalEndpoint.Port
    $listener.Stop()
    $env:LOCALAPPDATA = $local
    $env:ASTROCYTE_DATA_DIR = Split-Path -Parent $source
    $env:ASTROCYTE_PORT = [string]$port
    $env:ASTROCYTE_PROJECT_DISCOVERY_ROOTS = '[]'
    $env:ASTROCYTE_ENABLE_CODEX_DISTILLATION = 'false'
    $env:ASTROCYTE_ENABLE_SUMMARIZE = 'false'
    $env:ASTROCYTE_WORKBENCH = '1'
    $server = Start-Process -FilePath $serverBinary -PassThru -WindowStyle Hidden -RedirectStandardOutput (Join-Path $owned 'server.stdout.log') -RedirectStandardError (Join-Path $owned 'server.stderr.log')
    $processes += $server
    $deadline = [DateTime]::UtcNow.AddSeconds(30)
    while ($true) {
        $server.Refresh()
        if ($server.HasExited) { throw 'Actual server exited before initialization' }
        try {
            $response = Invoke-WebRequest "http://127.0.0.1:$port/api/v1/health" -TimeoutSec 1
            if ($response.StatusCode -eq 200) { break }
        } catch {
            if ([DateTime]::UtcNow -gt $deadline) { throw }
        }
        Start-Sleep -Milliseconds 100
    }
    & $snapshotBinary --source $source --root $root --timeout 120s > (Join-Path $owned 'empty-publication.json')
    Check-Exit 'actual empty snapshot'
    Stop-Process -Id $server.Id
    $server.WaitForExit()

    $writerDirectory = Join-Path $owned 'writer'
    New-Item -ItemType Directory -Path $writerDirectory | Out-Null
    $writerArguments = @('--mode','collect','--source',$source,'--repository',$Repository,'--ref',$PublicRef,'--scope',$ActualScope,'--output',$writerDirectory,'--strength',[string]$Strength,'--decay',[string]$Decay,'--initial-confidence',[string]$InitialConfidence,'--authorized-public-replay','--pace','25ms')
    # ArgumentList 中路径逐一引用，验证目录不依赖开发 node_modules。
    $quotedArguments = $writerArguments | ForEach-Object { '"' + $_.Replace('"','\"') + '"' }
    $writer = Start-Process -FilePath $collectorBinary -ArgumentList $quotedArguments -PassThru -WindowStyle Hidden -RedirectStandardOutput (Join-Path $owned 'writer.stdout.log') -RedirectStandardError (Join-Path $owned 'writer.stderr.log')
    $processes += $writer
    Wait-File (Join-Path $writerDirectory 'writer-started.json') $writer
    $progressPath = Join-Path $writerDirectory 'progress.json'
    Wait-File $progressPath $writer
    Start-Sleep -Milliseconds 1000
    $wal = Get-Item -LiteralPath ($source + '-wal')
    if ($wal.Length -le 0) { throw 'Actual source has no committed WAL' }
    $before = Get-Content -LiteralPath $progressPath -Raw | ConvertFrom-Json
    $writer.Refresh()
    if ($writer.HasExited) { throw 'Writer must be active during online backup' }
    & $snapshotBinary --source $source --root $root --timeout 120s > (Join-Path $owned 'concurrent-publication.json')
    Check-Exit 'multiprocess WAL snapshot'
    $after = Get-Content -LiteralPath $progressPath -Raw | ConvertFrom-Json
    @{ writer_pid=$writer.Id; source_wal_bytes=$wal.Length; recorded_before=$before.recorded; recorded_after=$after.recorded; actual_total=$after.actual_commits } | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $owned 'concurrent-safe.json') -Encoding utf8
    $writer.WaitForExit()
    if ($writer.ExitCode -ne 0) { throw 'Actual public writer failed' }
    & $snapshotBinary --source $source --root $root --timeout 120s > (Join-Path $owned 'complete-publication.json')
    Check-Exit 'complete public snapshot'
    & $sleepBinary --data-dir (Split-Path -Parent $source) --day $SleepDay --timezone $Timezone --decay $Decay --initial-confidence $InitialConfidence --fact-projection none > (Join-Path $owned 'sleep-result-local.json')
    Check-Exit 'actual public Sleep'
    & $snapshotBinary --source $source --root $root --timeout 120s > (Join-Path $owned 'sleep-publication.json')
    Check-Exit 'post Sleep snapshot'
    & $collectorBinary --mode inspect --source $source --root $root --output (Join-Path $owned 'safe')
    Check-Exit 'actual SQL and production Reader'
    @{ source=$source; snapshot_root=$root; localappdata=$local; safe=(Join-Path $owned 'safe'); public_ref=$PublicRef; actual_scope=$ActualScope; provenance='actual-public-Git-production-replay'; formal_experiment=$false } | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $owned 'result.json') -Encoding utf8
    Write-Output $owned
} finally {
    foreach ($process in $processes) {
        $process.Refresh()
        if (-not $process.HasExited) { Stop-Process -Id $process.Id; $process.WaitForExit() }
    }
    Get-ChildItem Env: | Where-Object { $_.Name.StartsWith('ASTROCYTE_') -or $_.Name -eq 'LOCALAPPDATA' } | ForEach-Object { [Environment]::SetEnvironmentVariable($_.Name, $null, 'Process') }
    foreach ($entry in $savedEnvironment.GetEnumerator()) { [Environment]::SetEnvironmentVariable($entry.Key, $entry.Value, 'Process') }
    Pop-Location
}
