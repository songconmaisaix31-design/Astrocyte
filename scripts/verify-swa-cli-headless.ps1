#Requires -Version 7.2
[CmdletBinding()]
param(
    [Parameter(Mandatory)][ValidateSet('T00','T01','T02','T03','T06','T07','T08','T09','T13','T15')][string]$Case,
    [Parameter(Mandatory)][string]$SwaPath,
    [Parameter(Mandatory)][string]$InstallRoot,
    [Parameter(Mandatory)][string]$DataRoot,
    [Parameter(Mandatory)][string]$CaseRoot,
    [string]$ProjectRoot,
    [string]$CodexAuthFile,
    [string]$PackageRoot,
    [string[]]$ProtectedRoots,
    [string]$OrcaPath,
    [string]$OrcaMetadataFile,
    [string[]]$ProtectedFiles,
    [int[]]$OriginalProcessIds,
    [int]$Port,
    [string]$ProjectId,
    [string]$OwnedProcessEvidence,
    [string]$FlowEvidence,
    [string]$GoTestDirectory,
    [ValidateSet('idempotency','disconnect')][string]$NativeCase = 'idempotency'
)

$ErrorActionPreference = 'Stop'
[System.Diagnostics.Process]::GetCurrentProcess().PriorityClass = 'BelowNormal'
$env:NODE_OPTIONS = '--max-old-space-size=1024'
$env:GOMAXPROCS = '4'
$env:GOMEMLIMIT = '512MiB'
foreach ($path in @($SwaPath,$InstallRoot,$DataRoot,$CaseRoot)) {
    if (-not [System.IO.Path]::IsPathFullyQualified($path)) { throw '验证路径必须是明确的绝对路径' }
}
$swaExpectedPath = [System.IO.Path]::GetFullPath((Join-Path $InstallRoot 'swa.exe'))
if ([System.IO.Path]::GetFullPath($SwaPath) -ne $swaExpectedPath) { throw 'CLI 必须来自同一安装候选' }
$swaNode = Join-Path $InstallRoot 'node/node.exe'
if ($Case -eq 'T02') {
    if (-not [IO.Path]::IsPathFullyQualified($PackageRoot)) { throw '安装检查需要明确的真实候选目录' }
    $swaNode = Join-Path $PackageRoot 'node/node.exe'
}
if (-not (Test-Path -LiteralPath $swaNode -PathType Leaf)) { throw '同候选 Node 不存在' }
if ($Case -eq 'T03') {
    if (-not [IO.Path]::IsPathFullyQualified($GoTestDirectory)) { throw '真实T03需要原编译输出的绝对路径' }
    $swaSourceRoot = Split-Path $PSScriptRoot -Parent
    $swaSourceRevision = (& git -C $swaSourceRoot rev-parse HEAD).Trim()
    if ($LASTEXITCODE -ne 0 -or $swaSourceRevision -notmatch '^[0-9a-f]{40}$') { throw '真实T03缺少准确源码提交' }
    $swaSourceDirty = & git -C $swaSourceRoot status --porcelain
    if ($LASTEXITCODE -ne 0 -or $swaSourceDirty) { throw '真实T03需要准确干净源码' }
    $swaCompiledRevision = (Get-Content -LiteralPath (Join-Path $GoTestDirectory 'SOURCE') -Raw).Trim()
    $swaInstalledRevision = (Get-Content -LiteralPath (Join-Path $InstallRoot 'VERSION') -Raw).Trim()
    if ($swaSourceRevision -cne $swaCompiledRevision -or $swaSourceRevision -cne $swaInstalledRevision) { throw '候选VERSION、编译SOURCE与准确源码必须相同' }
}
$swaEvidenceParent = Join-Path $CaseRoot '.verification/swa-cli-headless'
$swaEvidence = Join-Path $swaEvidenceParent ($Case.ToLowerInvariant() + '-' + [Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $swaEvidence -Force | Out-Null
$env:TEMP = Join-Path $swaEvidence 'temp'
$env:TMP = $env:TEMP
New-Item -ItemType Directory -Path $env:TEMP | Out-Null
$swaGate = Join-Path (Split-Path $PSScriptRoot -Parent) 'tests/swa-cli-headless/resource-gate.py'
& C:/Python313/python.exe $swaGate 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'resources.log')
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
if ($Case -in @('T00','T02','T03','T06','T08','T13','T15')) {
    $swaTestDirectory = Join-Path (Split-Path $PSScriptRoot -Parent) 'tests/swa-cli-headless'
    Import-Module (Join-Path $swaTestDirectory 'OrcaPreservation.psm1') -Force
    $swaPreservation = New-OrcaPreservation -OrcaPath $OrcaPath -MetadataFile $OrcaMetadataFile -ProtectedFiles $ProtectedFiles -OriginalProcessIds $OriginalProcessIds -ProbeDirectory $swaEvidence
    [IO.File]::WriteAllText((Join-Path $swaEvidence 'notification-probe.json'), ($swaPreservation.notification_probe | ConvertTo-Json))
    $swaService = $null
    $swaServiceStarted = $false
    $swaExit = 7
    try {
        Assert-OrcaPreservation -Observation $swaPreservation -OutputFile (Join-Path $swaEvidence 'orca-before.json') | Out-Null
        if ($Case -eq 'T00') {
            & $swaNode --max-old-space-size=1024 (Join-Path $swaTestDirectory 'public-case.mjs') --root $CaseRoot 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'public-source.log')
            $swaExit = $LASTEXITCODE
        } elseif ($Case -eq 'T02') {
            & pwsh -NoProfile -File (Join-Path $PackageRoot 'install-swa.ps1') -PackageRoot $PackageRoot -InstallRoot $InstallRoot -ProtectedRoots $ProtectedRoots 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'install.log')
            $swaExit = $LASTEXITCODE
            if ($swaExit -eq 0) {
                & (Join-Path $InstallRoot 'node/node.exe') --max-old-space-size=1024 (Join-Path $PSScriptRoot 'check-swa-runtime.mjs') --candidate $InstallRoot 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'installed-closure.log')
                $swaExit = $LASTEXITCODE
            }
            if ($swaExit -eq 0) {
                if ($Port -lt 1 -or $Port -gt 65535 -or -not [IO.Path]::IsPathFullyQualified($CodexAuthFile) -or $ProtectedRoots.Count -lt 1) { throw '安装隔离负向检查需要明确端口、只读认证及实际保护目录' }
                $swaOccupied = [Net.Sockets.TcpListener]::new([Net.IPAddress]::Loopback,$Port)
                $swaOccupied.Start()
                try {
                    & $SwaPath serve --data-dir $DataRoot --port $Port --codex-auth-file $CodexAuthFile --protected-roots ($ProtectedRoots -join ';') --json 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'actual-port-conflict.log')
                    if ($LASTEXITCODE -ne 5) { throw '真实占用端口没有按 EXIT5 拒绝' }
                } finally { $swaOccupied.Stop() }
                & $SwaPath serve --data-dir $ProtectedRoots[0] --port $Port --codex-auth-file $CodexAuthFile --protected-roots ($ProtectedRoots -join ';') --json 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'actual-protected-root.log')
                if ($LASTEXITCODE -ne 3) { throw '真实保护目录没有在写入前拒绝' }
                $swaJunction = Join-Path $swaEvidence 'owned-protected-root-junction'
                New-Item -ItemType Junction -Path $swaJunction -Target $ProtectedRoots[0] | Out-Null
                try {
                    & $SwaPath serve --data-dir $swaJunction --port $Port --codex-auth-file $CodexAuthFile --protected-roots ($ProtectedRoots -join ';') --json 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'actual-protected-junction.log')
                    if ($LASTEXITCODE -ne 3) { throw '实际本项 junction 没有在目标写入前拒绝' }
                } finally {
                    $swaLinkAbsolute = [IO.Path]::GetFullPath($swaJunction)
                    $swaEvidenceAbsolute = [IO.Path]::GetFullPath($swaEvidence).TrimEnd('\') + '\'
                    if (-not $swaLinkAbsolute.StartsWith($swaEvidenceAbsolute,[StringComparison]::OrdinalIgnoreCase) -or -not ((Get-Item -LiteralPath $swaLinkAbsolute -Force).Attributes -band [IO.FileAttributes]::ReparsePoint)) { throw '本项 junction 清理目标无法核实' }
                    [IO.Directory]::Delete($swaLinkAbsolute)
                }
                [IO.File]::WriteAllText((Join-Path $swaEvidence 'installation-scope.json'), (@{ installed_candidate = $InstallRoot; actual_port_conflict = 'rejected'; actual_protected_root = 'rejected'; actual_owned_junction = 'rejected'; actual_private_service_consumers = 'requires_T03_T13'; model_calls = 0 } | ConvertTo-Json))
                $swaExit = 6
            }
        } elseif ($Case -in @('T03','T06','T08','T13')) {
            if (-not [IO.Path]::IsPathFullyQualified($CodexAuthFile) -or -not [IO.Path]::IsPathFullyQualified($ProjectRoot) -or $Port -lt 1 -or $Port -gt 65535 -or $ProtectedRoots.Count -lt 1) { throw '真实服务需要明确认证/公开项目/端口/保护目录' }
            $swaStart = [Diagnostics.ProcessStartInfo]::new($SwaPath)
            $swaStart.UseShellExecute = $false
            $swaStart.CreateNoWindow = $true
            $swaStart.RedirectStandardInput = $true
            $swaStart.RedirectStandardOutput = $true
            $swaStart.RedirectStandardError = $true
            foreach ($argument in @('serve','--data-dir',$DataRoot,'--port',[string]$Port,'--codex-auth-file',$CodexAuthFile,'--protected-roots',($ProtectedRoots -join ';'),'--json')) { $swaStart.ArgumentList.Add($argument) }
            $swaService = [Diagnostics.Process]::new()
            $swaService.StartInfo = $swaStart
            if (-not $swaService.Start()) { throw '实际安装服务不能启动' }
            $swaServiceStarted = $true
            $swaServiceError = $swaService.StandardError.ReadToEndAsync()
            $swaReadyLine = $swaService.StandardOutput.ReadLineAsync()
            if (-not $swaReadyLine.Wait(60000)) { throw '真实安装服务 readiness 超时，保留实际所属状态' }
            $swaReady = $swaReadyLine.Result | ConvertFrom-Json
            if (-not $swaReady.ok) { throw '真实安装服务没有取得 readiness' }
            [IO.File]::WriteAllText((Join-Path $swaEvidence 'service-readiness.json'), ($swaReady | ConvertTo-Json -Depth 40))
            Assert-OrcaPreservation -Observation $swaPreservation -OutputFile (Join-Path $swaEvidence 'orca-after-private-start.json') | Out-Null
            if ([string]::IsNullOrWhiteSpace($env:SWA_PROJECT_TOKEN) -or [string]::IsNullOrWhiteSpace($ProjectId)) {
                [IO.File]::WriteAllText((Join-Path $swaEvidence 'approval-required.json'), (@{ ok = $false; required_action = 'human_register_public_project_review_settings_grant_agent_issue_project_token'; endpoint = "http://127.0.0.1:$Port"; project_root = $ProjectRoot; instructions = (Join-Path $InstallRoot 'USAGE.md'); automatic_human_session = $false } | ConvertTo-Json))
                $swaExit = 3
            } elseif ($Case -eq 'T03') {
                $env:SWA_ENDPOINT = "http://127.0.0.1:$Port"
                $env:SWA_PROJECT_ID = $ProjectId
                $swaStatusText = & $SwaPath status --json
                if ($LASTEXITCODE -ne 0) { throw '实际项目批准未通过状态查询' }
                $swaStatus = ($swaStatusText -join "`n") | ConvertFrom-Json
                $swaControllerText = & $SwaPath controller open --project $ProjectId --workdir $ProjectRoot --request-id ([Guid]::NewGuid().ToString()) --json
                if ($LASTEXITCODE -ne 0) { throw '实际 Go 检查控制端创建失败，保留原请求' }
                $swaGoController = ($swaControllerText -join "`n") | ConvertFrom-Json
                [IO.File]::WriteAllText((Join-Path $swaEvidence 'go-controller.json'), ($swaGoController | ConvertTo-Json -Depth 40))
                $env:SWA_REAL_ENABLED = '1'
                $env:SWA_REAL_INSTALL_ROOT = $InstallRoot
                $env:SWA_REAL_DATA_ROOT = $DataRoot
                $env:SWA_REAL_CONTROLLER = $swaGoController.controller
                $env:SWA_REAL_RUNTIME_ID = $swaStatus.runtime_id
                $env:SWA_REAL_CODEX_AUTH_FILE = $CodexAuthFile
                $env:SWA_REAL_PROTECTED_ROOTS = $ProtectedRoots -join ';'
                try {
                    $swaTests = Get-Content -LiteralPath (Join-Path $GoTestDirectory 'tests.json') -Raw | ConvertFrom-Json
                    if ($swaTests.Count -ne 4) { throw '四个真实 Go 检查条目不完整' }
                    foreach ($swaTest in $swaTests) {
                        $swaTestLog = Join-Path $swaEvidence ($swaTest.binary + '.log')
                        & (Join-Path $GoTestDirectory $swaTest.binary) '-test.v' '-test.run' '^TestOrcaCLIReal' 2>&1 | Tee-Object -FilePath $swaTestLog
                        if ($LASTEXITCODE -ne 0) { throw '实际 Go 检查失败' }
                        $swaActualTestText = Get-Content -LiteralPath $swaTestLog -Raw
                        if ($swaActualTestText -notmatch ('--- PASS: ' + [Regex]::Escape($swaTest.test) + ' ') -or $swaActualTestText -match '--- SKIP:') { throw '真实 Go 检查未执行匹配函数，或存在跳过' }
                    }
                    [IO.File]::WriteAllText((Join-Path $swaEvidence 'go-checks.json'), (@{ actual_checks = $swaTests; model_calls = 0; remaining_T03 = @('real_human_safety_stop_after_revoke','different_real_project_run','old_controller_generation') } | ConvertTo-Json -Depth 15))
                    $swaExit = 6
                } finally {
                    & $SwaPath controller close --controller $swaGoController.controller --request-id ([Guid]::NewGuid().ToString()) --json 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'go-controller-close.log')
                    if ($LASTEXITCODE -ne 0) { throw '实际 Go 检查控制端退出未确认，保留原请求和资源' }
                }
            } elseif ($Case -in @('T06','T08')) {
                $swaNativeCase = if ($Case -eq 'T08') { 'storage' } else { $NativeCase }
                & $swaNode --max-old-space-size=1024 (Join-Path $swaTestDirectory 'native-idempotency.mjs') --case $swaNativeCase --candidate $InstallRoot --data-root $DataRoot --project $ProjectRoot --project-id $ProjectId --endpoint "http://127.0.0.1:$Port" --output (Join-Path $swaEvidence 'native-idempotency.json') 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'native-idempotency.log')
                $swaExit = $LASTEXITCODE
                if ($swaExit -eq 0) {
                    $swaIdempotency = Get-Content -LiteralPath (Join-Path $swaEvidence 'native-idempotency.json') -Raw | ConvertFrom-Json
                    if ($swaIdempotency.remainingCases.Count -gt 0) { $swaExit = 6 }
                }
            } else {
                $swaObservationConfig = Join-Path $swaEvidence 'original-observation.json'
                [IO.File]::WriteAllText($swaObservationConfig, (@{ orca_path = $OrcaPath; metadata_file = $OrcaMetadataFile; protected_files = $ProtectedFiles; original_process_ids = $OriginalProcessIds } | ConvertTo-Json -Depth 10))
                & $swaNode --max-old-space-size=1024 (Join-Path $swaTestDirectory 'installed-flow.mjs') --candidate $InstallRoot --data-root $DataRoot --project $ProjectRoot --project-id $ProjectId --endpoint "http://127.0.0.1:$Port" --orca-observation $swaObservationConfig --output (Join-Path $swaEvidence 'installed-flow.json') 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'installed-flow.log')
                $swaExit = $LASTEXITCODE
            }
        } else {
            if (-not [IO.Path]::IsPathFullyQualified($OwnedProcessEvidence)) { throw 'T15 需要先前真实运行保存的 owned 进程身份原件' }
            $swaOwned = Get-Content -LiteralPath $OwnedProcessEvidence -Raw | ConvertFrom-Json
            if ($swaOwned.ownedProcesses.Count -lt 1) { throw '实际 owned 进程身份列表缺失' }
            foreach ($identity in $swaOwned.ownedProcesses) {
                $swaCurrent = Get-CimInstance Win32_Process -Filter "ProcessId = $([int]$identity.pid)"
                if ($swaCurrent -and $swaCurrent.CreationDate.ToUniversalTime().ToString('o') -eq $identity.creation_utc) { throw "先前 owned 进程仍存在: $($identity.pid)" }
            }
            $swaExit = 0
        }
    } finally {
        if ($swaServiceStarted) {
            $swaAllProcesses = @(Get-CimInstance Win32_Process)
            $swaOwnedPids = [Collections.Generic.HashSet[int]]::new()
            $swaOwnedPids.Add($swaService.Id) | Out-Null
            do {
                $swaAdded = $false
                foreach ($swaRow in $swaAllProcesses) {
                    if ($swaOwnedPids.Contains([int]$swaRow.ParentProcessId) -and $swaOwnedPids.Add([int]$swaRow.ProcessId)) { $swaAdded = $true }
                }
            } while ($swaAdded)
            $swaOwnedIdentities = @($swaAllProcesses | Where-Object { $swaOwnedPids.Contains([int]$_.ProcessId) } | ForEach-Object { @{ pid = [int]$_.ProcessId; creation_utc = $_.CreationDate.ToUniversalTime().ToString('o'); executable = $_.ExecutablePath } })
            [IO.File]::WriteAllText((Join-Path $swaEvidence 'owned-processes.json'), (@{ ownedProcesses = $swaOwnedIdentities; observed_before_owner_EOF = $true } | ConvertTo-Json -Depth 10))
            $swaService.StandardInput.Close()
            if (-not $swaService.WaitForExit(60000)) {
                [IO.File]::WriteAllText((Join-Path $swaEvidence 'owned-service-unknown.json'), (@{ pid = $swaService.Id; root_exit = 'unknown'; automatic_pid_kill = $false } | ConvertTo-Json))
                $swaExit = 6
            } else {
                [IO.File]::WriteAllText((Join-Path $swaEvidence 'owned-service.log'), $swaServiceError.GetAwaiter().GetResult())
                if ($swaService.ExitCode -ne 0 -and $swaExit -eq 0) { $swaExit = 6 }
            }
            $swaExitObservations = @()
            for ($swaIndex = 0; $swaIndex -lt 2; $swaIndex++) {
                if ($swaIndex) { Start-Sleep -Milliseconds 75 }
                $swaCurrentProcesses = @(Get-CimInstance Win32_Process)
                $swaRetained = @($swaCurrentProcesses | Where-Object {
                    $swaCurrentRow = $_
                    @($swaOwnedIdentities | Where-Object { $_.pid -eq $swaCurrentRow.ProcessId -and $_.creation_utc -eq $swaCurrentRow.CreationDate.ToUniversalTime().ToString('o') }).Count -gt 0 -or ($swaCurrentRow.ExecutablePath -and $swaCurrentRow.ExecutablePath.StartsWith($InstallRoot.TrimEnd('\') + '\',[StringComparison]::OrdinalIgnoreCase))
                } | Select-Object ProcessId,ParentProcessId,CreationDate,ExecutablePath)
                $swaExitObservations += @{ utc = [DateTime]::UtcNow.ToString('o'); retained = $swaRetained }
                if ($swaRetained.Count -gt 0) { $swaExit = 6 }
            }
            [IO.File]::WriteAllText((Join-Path $swaEvidence 'owned-process-exit.json'), ($swaExitObservations | ConvertTo-Json -Depth 15))
        }
        try { Assert-OrcaPreservation -Observation $swaPreservation -OutputFile (Join-Path $swaEvidence 'orca-after.json') | Out-Null } finally { Close-OrcaPreservation -Observation $swaPreservation }
    }
    [IO.File]::WriteAllText((Join-Path $swaEvidence 'exit.txt'), [string]$swaExit)
    exit $swaExit
} elseif ($Case -eq 'T07') {
    if (-not [IO.Path]::IsPathFullyQualified($FlowEvidence)) { throw 'T07 需要同一安装候选的真实双Worker结果原件' }
    & $swaNode --max-old-space-size=512 (Join-Path (Split-Path $PSScriptRoot -Parent) 'tests/swa-cli-headless/result-sources.test.mjs') --candidate $InstallRoot --flow-evidence $FlowEvidence --output (Join-Path $swaEvidence 'result-sources.json') 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'result-sources.log')
} elseif ($Case -eq 'T01') {
    & $swaNode --max-old-space-size=1024 (Join-Path $PSScriptRoot 'check-swa-runtime.mjs') --candidate $InstallRoot 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'closure.log')
} else {
    $swaFixture = Join-Path (Split-Path $PSScriptRoot -Parent) 'tests/swa-cli-headless/provider-job.test.mjs'
    & $swaNode --max-old-space-size=1024 $swaFixture --candidate $InstallRoot --output (Join-Path $swaEvidence 'ordinary-processes.json') 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'ordinary-processes.log')
}
$swaExit = $LASTEXITCODE
[IO.File]::WriteAllText((Join-Path $swaEvidence 'exit.txt'), [string]$swaExit)
if ($Case -eq 'T09' -and $swaExit -eq 0) {
    $swaObserved = Get-Content -LiteralPath (Join-Path $swaEvidence 'ordinary-processes.json') -Raw | ConvertFrom-Json
    if (-not [IO.Path]::IsPathFullyQualified($ProjectRoot) -or -not [IO.Path]::IsPathFullyQualified($CodexAuthFile)) { throw '真实SDK检查需要明确公开项目和现有只读Codex认证路径' }
    & C:/Python313/python.exe $swaGate 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'sdk-resources.log')
    if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
    & $swaNode --max-old-space-size=1024 (Join-Path (Split-Path $PSScriptRoot -Parent) 'tests/swa-cli-headless/sdk-session.test.mjs') --candidate $InstallRoot --project $ProjectRoot --auth-file $CodexAuthFile --output (Join-Path $swaEvidence 'sdk-session.json') 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'sdk-session.log')
    $swaExit = $LASTEXITCODE
    [IO.File]::WriteAllText((Join-Path $swaEvidence 'sdk-exit.txt'), [string]$swaExit)
    if ($swaExit -ne 0) { exit $swaExit }
    $swaSDK = Get-Content -LiteralPath (Join-Path $swaEvidence 'sdk-session.json') -Raw | ConvertFrom-Json
    if ($swaSDK.remainingCases.Count -gt 0 -or @($swaObserved.remainingCases | Where-Object { $_ -ne 'actual_sdk_abort_then_close' }).Count -gt 0) { throw 'T09 仍有必要分支未运行' }
}
exit $swaExit
