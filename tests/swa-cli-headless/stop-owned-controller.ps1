#Requires -Version 7.2
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$OwnerFile,
    [Parameter(Mandatory)][string]$ProtectedProcessIds,
    [Parameter(Mandatory)][string]$OutputFile
)
$ErrorActionPreference = 'Stop'
[Diagnostics.Process]::GetCurrentProcess().PriorityClass = 'BelowNormal'
$swaOwner = (Get-Content -LiteralPath $OwnerFile -Raw | ConvertFrom-Json).result
$swaProtected = @($ProtectedProcessIds.Split(',') | ForEach-Object { [int]$_ })
if ($swaProtected.Count -lt 1 -or [string]::IsNullOrWhiteSpace($swaOwner.controller)) { throw '真实控制端或原进程保护列表缺失' }
$swaIdentities = @($swaOwner.process, $swaOwner.helper) | Sort-Object -Property pid -Unique
$swaHeld = @()
$swaObservations = @()
try {
    foreach ($swaIdentity in $swaIdentities) {
        $swaPid = [int]$swaIdentity.pid
        if ($swaPid -le 0 -or $swaPid -eq $PID -or $swaProtected -contains $swaPid) { throw '实际故障目标与受保护进程重合' }
        $swaProcess = [Diagnostics.Process]::GetProcessById($swaPid)
        $swaHeld += $swaProcess
        [void]$swaProcess.SafeHandle
        $swaCreated = [long][Math]::Floor(($swaProcess.StartTime.ToUniversalTime() - [DateTime]::UnixEpoch).TotalMilliseconds)
        if ($swaCreated -ne [long]$swaIdentity.creationTimeMs -or ([IO.Path]::GetFileName($swaProcess.MainModule.FileName)) -ine [string]$swaIdentity.name) { throw '实际控制端创建身份或程序发生变化' }
        $swaObservations += @{ pid = $swaPid; creation_time_ms = $swaCreated; executable = $swaProcess.MainModule.FileName; actual_process_handle_held = $true }
    }
    foreach ($swaProcess in $swaHeld) { if (-not $swaProcess.HasExited) { $swaProcess.Kill() } }
    foreach ($swaProcess in $swaHeld) { if (-not $swaProcess.WaitForExit(10000)) { throw '实际控制端故障退出未确认' } }
    [IO.File]::WriteAllText($OutputFile, (@{ controller = $swaOwner.controller; actual_owned_processes = $swaObservations; actual_exit = $true; protected_process_ids = $swaProtected } | ConvertTo-Json -Depth 12))
} finally {
    foreach ($swaProcess in $swaHeld) { $swaProcess.Dispose() }
}
