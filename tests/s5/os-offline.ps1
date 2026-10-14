param([switch]$Execute)
$ErrorActionPreference = 'Stop'
$workspace = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
$ownedRoot = [IO.Path]::GetFullPath((Join-Path $workspace '.verification/s5/os-offline'))
if (-not $ownedRoot.StartsWith($workspace + [IO.Path]::DirectorySeparatorChar, [StringComparison]::OrdinalIgnoreCase)) { throw 'Owned directory escaped the S5 workspace' }
$sharedNode = (Get-Command node.exe -CommandType Application).Source
$ownedNode = Join-Path $ownedRoot 'node.exe'
$ruleName = 'Astrocyte-S5-ctx_f403dc7813bc-owned-node-outbound'
if (-not (Test-Path -LiteralPath $ownedRoot)) { New-Item -ItemType Directory -Path $ownedRoot | Out-Null }
$rootItem = Get-Item -LiteralPath $ownedRoot
if (($rootItem.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'Owned directory is a reparse point' }
$unexpected = @(Get-ChildItem -LiteralPath $ownedRoot | Where-Object { $_.Name -ne 'node.exe' })
if ($unexpected.Count -ne 0) { throw 'Owned directory contains unexplained files' }
$sourceHash = (Get-FileHash -LiteralPath $sharedNode -Algorithm SHA256).Hash
if (Test-Path -LiteralPath $ownedNode) {
    if ((Get-FileHash -LiteralPath $ownedNode -Algorithm SHA256).Hash -ne $sourceHash) { throw 'An existing owned node.exe differs from the configured runtime' }
} else {
    Copy-Item -LiteralPath $sharedNode -Destination $ownedNode
}
if ((Get-FileHash -LiteralPath $ownedNode -Algorithm SHA256).Hash -ne $sourceHash) { throw 'Owned runtime copy differs' }
if ($ownedNode -eq $sharedNode) { throw 'Shared Node cannot be isolated' }
$plan = [ordered]@{
    task_id = 'task_1d3e15a66b2a'; dispatch_id = 'ctx_f403dc7813bc'
    program = $ownedNode; shared_program = $sharedNode; program_sha256 = $sourceHash
    runtime_version = (& $ownedNode --version); rule_name = $ruleName
    scope = 'Outbound traffic of this dedicated node.exe only; shared Node and Orca remain outside this scope'
    create = 'New-NetFirewallRule -Name $ruleName -DisplayName $ruleName -Direction Outbound -Action Block -Program $ownedNode -Profile Any -Protocol Any -Enabled True'
    verify = 'Read the actual rule and ApplicationFilter, run the same dedicated Node HTTP refusal followed by local CPU inference and restart'
    cleanup = 'Finally terminate any surviving process at exactly $ownedNode and remove only this exact new rule; confirm absence'
    execution = 'NOT_RUN; reviewed explicit administrator action is required'
}
$planPath = Join-Path $workspace '.verification/s5/os-offline-plan.json'
$plan | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $planPath -Encoding utf8
if (-not $Execute) { $plan | ConvertTo-Json -Depth 8; return }

# 只有批准后的显式执行模式才创建规则，不自动提权。
$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = [Security.Principal.WindowsPrincipal]::new($identity)
if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) { throw 'Administrator identity is required; no rule was created' }
$existing = @(Get-NetFirewallRule | Where-Object { $_.Name -eq $ruleName })
if ($existing.Count -ne 0) { throw 'The exact rule already exists; it will not be changed or removed' }
if (@(Get-NetFirewallProfile | Where-Object { -not $_.Enabled }).Count -ne 0) { throw 'All firewall profiles must already be enabled; profiles will not be changed' }
if (@(Get-CimInstance Win32_Process | Where-Object { $_.ExecutablePath -eq $ownedNode }).Count -ne 0) { throw 'The owned runtime is already active' }
$created = $false
$savedEnvironment = @{}
$environmentNames = @('ASTROCYTE_S5_NODE','ASTROCYTE_S5_NETWORK_MODE','ASTROCYTE_S5_ASSET_ROOT','HTTP_PROXY','HTTPS_PROXY','ALL_PROXY','NO_PROXY','http_proxy','https_proxy','no_proxy','NODE_USE_ENV_PROXY')
foreach ($name in $environmentNames) { $savedEnvironment[$name] = [Environment]::GetEnvironmentVariable($name, 'Process') }
try {
    New-NetFirewallRule -Name $ruleName -DisplayName $ruleName -Direction Outbound -Action Block -Program $ownedNode -Profile Any -Protocol Any -Enabled True | Out-Null
    $created = $true
    $actualRule = Get-NetFirewallRule -Name $ruleName
    $actualProgram = ($actualRule | Get-NetFirewallApplicationFilter).Program
    if ($actualProgram -ne $ownedNode -or $actualRule.Action -ne 'Block' -or $actualRule.Direction -ne 'Outbound' -or $actualRule.Enabled -ne 'True') { throw 'The actual rule differs from the approved program scope' }
    [ordered]@{ rule = $actualRule | Select-Object Name,Direction,Action,Enabled,Profile; program = $actualProgram; sha256 = $sourceHash; observed_at = [DateTime]::UtcNow.ToString('o') } | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath (Join-Path $workspace '.verification/s5/os-offline-rule.json') -Encoding utf8
    foreach ($name in @('HTTP_PROXY','HTTPS_PROXY','ALL_PROXY','NO_PROXY','http_proxy','https_proxy','no_proxy','NODE_USE_ENV_PROXY')) { [Environment]::SetEnvironmentVariable($name, $null, 'Process') }
    $env:ASTROCYTE_S5_NODE = $ownedNode
    $env:ASTROCYTE_S5_NETWORK_MODE = 'os-program-block'
    $env:ASTROCYTE_S5_ASSET_ROOT = 'D:/Astrocyte-agent-assets/20261015/judgment-s5'
    & $sharedNode (Join-Path $workspace 'tests/s5/offline-live.mjs') 2>&1 | Tee-Object -FilePath (Join-Path $workspace '.verification/s5/offline-os-first.log')
    if ($LASTEXITCODE -ne 0) { throw "Actual isolated inference failed with exit $LASTEXITCODE" }
} finally {
    foreach ($process in @(Get-CimInstance Win32_Process | Where-Object { $_.ExecutablePath -eq $ownedNode })) { Stop-Process -Id $process.ProcessId -ErrorAction Stop }
    if ($created) {
        $remainingRule = Get-NetFirewallRule -Name $ruleName
        if (($remainingRule | Get-NetFirewallApplicationFilter).Program -ne $ownedNode) { throw 'Rule ownership changed; automatic removal refused' }
        Remove-NetFirewallRule -Name $ruleName
        if (@(Get-NetFirewallRule | Where-Object { $_.Name -eq $ruleName }).Count -ne 0) { throw 'The dedicated rule remained after cleanup' }
        [ordered]@{ rule_name = $ruleName; removed = $true; observed_at = [DateTime]::UtcNow.ToString('o') } | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $workspace '.verification/s5/os-offline-cleanup.json') -Encoding utf8
    }
    foreach ($name in $environmentNames) { [Environment]::SetEnvironmentVariable($name, $savedEnvironment[$name], 'Process') }
}
