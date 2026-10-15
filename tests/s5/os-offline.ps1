param([switch]$Execute)
$ErrorActionPreference = 'Stop'
$workspace = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
$ownedRoot = [IO.Path]::GetFullPath((Join-Path $workspace '.verification/s5/os-offline'))
if (-not $ownedRoot.StartsWith($workspace + [IO.Path]::DirectorySeparatorChar, [StringComparison]::OrdinalIgnoreCase)) { throw 'Owned directory escaped the S5 workspace' }
$sharedNode = (Get-Command node.exe -CommandType Application).Source
$ownedNode = Join-Path $ownedRoot 'node.exe'
$dispatchId = 'ctx_0a7797464f02'
$ruleName = "Astrocyte-S5-$dispatchId-owned-node-outbound"
$executionId = [Guid]::NewGuid().ToString('N')
$ruleGroup = "Astrocyte-S5-$dispatchId-$executionId"
$evidencePrefix = Join-Path $workspace ".verification/s5/os-offline-$dispatchId-$executionId"
if (-not (Test-Path -LiteralPath $ownedRoot)) { New-Item -ItemType Directory -Path $ownedRoot | Out-Null }
$rootItem = Get-Item -LiteralPath $ownedRoot
if (($rootItem.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'Owned directory is a reparse point' }
$unexpected = @(Get-ChildItem -LiteralPath $ownedRoot -Force | Where-Object { $_.Name -ne 'node.exe' })
if ($unexpected.Count -ne 0) { throw 'Owned directory contains unexplained files' }
$sourceHash = (Get-FileHash -LiteralPath $sharedNode -Algorithm SHA256).Hash
if (Test-Path -LiteralPath $ownedNode) {
    if (((Get-Item -LiteralPath $ownedNode).Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'Owned node.exe is a reparse point' }
    if ((Get-FileHash -LiteralPath $ownedNode -Algorithm SHA256).Hash -ne $sourceHash) { throw 'An existing owned node.exe differs from the configured runtime' }
} else {
    Copy-Item -LiteralPath $sharedNode -Destination $ownedNode
}
if ((Get-FileHash -LiteralPath $ownedNode -Algorithm SHA256).Hash -ne $sourceHash) { throw 'Owned runtime copy differs' }
if ($ownedNode -eq $sharedNode) { throw 'Shared Node cannot be isolated' }
$runtimeVersion = & $ownedNode --version
if ($LASTEXITCODE -ne 0) { throw "Owned runtime version check failed with exit $LASTEXITCODE" }
$plan = [ordered]@{
    task_id = 'task_1d3e15a66b2a'; dispatch_id = $dispatchId
    program = $ownedNode; shared_program = $sharedNode; program_sha256 = $sourceHash
    runtime_version = $runtimeVersion; rule_name = $ruleName; rule_group = $ruleGroup
    scope = 'Outbound traffic of this dedicated node.exe only; shared Node and Orca remain outside this scope'
    create = 'New-NetFirewallRule -Name $ruleName -DisplayName $ruleName -Group $ruleGroup -PolicyStore PersistentStore -Direction Outbound -Action Block -Program $ownedNode -Profile Any -Protocol Any -Enabled True'
    verify = 'Read the actual rule and ApplicationFilter, run the same dedicated Node HTTP refusal followed by local CPU inference and restart'
    cleanup = 'Nested finally independently attempts every dedicated process termination, removal of the exact rule matching name, unique group, program, direction and action, and every saved process environment restoration; read back each final state and throw all execution or cleanup failures'
    execution = 'NOT_RUN; reviewed explicit administrator action is required'
}
$planPath = "$evidencePrefix-plan.json"
$plan | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $planPath -Encoding utf8
if (-not $Execute) { $plan | ConvertTo-Json -Depth 8; return }

$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = [Security.Principal.WindowsPrincipal]::new($identity)
if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) { throw 'Administrator identity is required; no rule was created' }
$existing = @(Get-NetFirewallRule -PolicyStore PersistentStore | Where-Object { $_.Name -eq $ruleName })
if ($existing.Count -ne 0) { throw 'The exact rule already exists; it will not be changed or removed' }
if (@(Get-NetFirewallProfile | Where-Object { -not $_.Enabled }).Count -ne 0) { throw 'All firewall profiles must already be enabled; profiles will not be changed' }
if (@(Get-CimInstance Win32_Process | Where-Object { $_.ExecutablePath -eq $ownedNode }).Count -ne 0) { throw 'The owned runtime is already active' }
$created = $false
$creationAttempted = $false
$operationFailure = $null
$savedEnvironment = @{}
$environmentNames = @('ASTROCYTE_S5_NODE','ASTROCYTE_S5_NETWORK_MODE','ASTROCYTE_S5_ASSET_ROOT','HTTP_PROXY','HTTPS_PROXY','ALL_PROXY','NO_PROXY','NODE_USE_ENV_PROXY')
foreach ($name in $environmentNames) { $savedEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process') }
try {
    $creationAttempted = $true
    New-NetFirewallRule -Name $ruleName -DisplayName $ruleName -Group $ruleGroup -PolicyStore PersistentStore -Direction Outbound -Action Block -Program $ownedNode -Profile Any -Protocol Any -Enabled True | Out-Null
    $created = $true
    $actualRule = Get-NetFirewallRule -PolicyStore PersistentStore -Name $ruleName
    $actualProgram = ($actualRule | Get-NetFirewallApplicationFilter).Program
    if ($actualProgram -ne $ownedNode -or $actualRule.Group -ne $ruleGroup -or $actualRule.Action -ne 'Block' -or $actualRule.Direction -ne 'Outbound' -or $actualRule.Enabled -ne 'True') { throw 'The actual rule differs from the approved program scope' }
    [ordered]@{ rule = $actualRule | Select-Object Name,Direction,Action,Enabled,Profile; program = $actualProgram; sha256 = $sourceHash; observed_at = [DateTime]::UtcNow.ToString('o') } | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath "$evidencePrefix-rule.json" -Encoding utf8
    foreach ($name in @('HTTP_PROXY','HTTPS_PROXY','ALL_PROXY','NO_PROXY','NODE_USE_ENV_PROXY')) { [Environment]::SetEnvironmentVariable($name, $null, 'Process') }
    $env:ASTROCYTE_S5_NODE = $ownedNode
    $env:ASTROCYTE_S5_NETWORK_MODE = 'os-program-block'
    $env:ASTROCYTE_S5_ASSET_ROOT = 'D:/Astrocyte-agent-assets/20261015/judgment-s5'
    & $ownedNode (Join-Path $workspace 'tests/s5/offline-live.mjs') 2>&1 | Tee-Object -FilePath "$evidencePrefix-first.log"
    if ($LASTEXITCODE -ne 0) { throw "Actual isolated inference failed with exit $LASTEXITCODE" }
} catch {
    $operationFailure = $_.Exception
    throw
} finally {
    $cleanupFailures = [Collections.Generic.List[Exception]]::new()
    $ownedProcessesAbsent = $false
    $ruleAbsent = $false
    $environmentRestored = $false
    try {
        try {
            $ownedProcesses = @(Get-CimInstance Win32_Process | Where-Object { $_.ExecutablePath -eq $ownedNode })
            foreach ($process in $ownedProcesses) {
                try {
                    $current = @(Get-CimInstance Win32_Process -Filter "ProcessId = $($process.ProcessId)")
                    if ($current.Count -eq 0) { continue }
                    if ($current[0].ExecutablePath -ne $ownedNode -or $current[0].CreationDate -ne $process.CreationDate) { throw 'Process ownership changed; termination refused' }
                    $running = Get-Process -Id $process.ProcessId -ErrorAction Stop
                    Stop-Process -InputObject $running -ErrorAction Stop
                    if (-not $running.WaitForExit(10000)) { throw 'A dedicated process did not exit within the cleanup deadline' }
                }
                catch { $cleanupFailures.Add($_.Exception) }
            }
            if (@(Get-CimInstance Win32_Process | Where-Object { $_.ExecutablePath -eq $ownedNode }).Count -ne 0) { throw 'An owned runtime process remained after cleanup' }
            $ownedProcessesAbsent = $true
        } catch { $cleanupFailures.Add($_.Exception) }
    } finally {
        try {
            if ($creationAttempted) {
                $remainingRules = @(Get-NetFirewallRule -PolicyStore PersistentStore | Where-Object { $_.Name -eq $ruleName })
                foreach ($remainingRule in $remainingRules) {
                    if ($remainingRule.Group -ne $ruleGroup -or ($remainingRule | Get-NetFirewallApplicationFilter).Program -ne $ownedNode -or $remainingRule.Direction -ne 'Outbound' -or $remainingRule.Action -ne 'Block') { throw 'Rule ownership changed; automatic removal refused' }
                    Remove-NetFirewallRule -InputObject $remainingRule -ErrorAction Stop
                }
                if (@(Get-NetFirewallRule -PolicyStore PersistentStore | Where-Object { $_.Name -eq $ruleName }).Count -ne 0) { throw 'The dedicated rule remained after cleanup' }
                $ruleAbsent = $true
            }
        } catch { $cleanupFailures.Add($_.Exception) }
        finally {
            foreach ($name in $environmentNames) {
                try {
                    [Environment]::SetEnvironmentVariable($name, $savedEnvironment[$name], 'Process')
                    if ([Environment]::GetEnvironmentVariable($name, 'Process') -cne $savedEnvironment[$name]) { throw "Process environment restoration failed for $name" }
                }
                catch { $cleanupFailures.Add($_.Exception) }
            }
            $environmentRestored = @($environmentNames | Where-Object { [Environment]::GetEnvironmentVariable($_, 'Process') -cne $savedEnvironment[$_] }).Count -eq 0
        }
    }
    try {
        [ordered]@{ rule_name = $ruleName; rule_group = $ruleGroup; creation_completed = $created; rule_absent = $ruleAbsent; owned_processes_absent = $ownedProcessesAbsent; environment_restored = $environmentRestored; operation_failure = $(if ($null -ne $operationFailure) { $operationFailure.ToString() }); cleanup_failures = @($cleanupFailures | ForEach-Object { $_.ToString() }); observed_at = [DateTime]::UtcNow.ToString('o') } | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath "$evidencePrefix-cleanup.json" -Encoding utf8
    } catch { $cleanupFailures.Add($_.Exception) }
    if ($cleanupFailures.Count -ne 0) {
        if ($null -ne $operationFailure) { $cleanupFailures.Insert(0, $operationFailure) }
        throw [AggregateException]::new('S5 execution or required cleanup failed', $cleanupFailures.ToArray())
    }
}
