param(
    [Parameter(Mandatory=$true)][string]$EvidenceRoot,
    [ValidateSet('tests', 'http', 'checks')][string]$Phase = 'tests'
)

$ErrorActionPreference = 'Stop'
$RepositoryRoot = (Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
if (-not [IO.Path]::IsPathRooted($EvidenceRoot)) { throw 'EvidenceRoot must be an explicit absolute owned directory' }
$EvidenceRoot = [IO.Path]::GetFullPath($EvidenceRoot)
$AssetRoot = 'D:/Astrocyte-agent-assets/20261015/judgment-s5'
if (-not $EvidenceRoot.StartsWith([IO.Path]::GetFullPath($AssetRoot) + [IO.Path]::DirectorySeparatorChar, [StringComparison]::OrdinalIgnoreCase)) {
    throw 'EvidenceRoot must stay within the authorized S5 D drive directory'
}
$null = New-Item -ItemType Directory -Path $EvidenceRoot -Force
$TemporaryRoot = Join-Path $EvidenceRoot 'temp'
$RunRoot = Join-Path $EvidenceRoot 'runs'
$null = New-Item -ItemType Directory -Path $TemporaryRoot, $RunRoot -Force
$ProcessEnvironment = @{
    GOCACHE = 'D:/Astrocyte-agent-assets/20261015/go-build'
    GOTMPDIR = $TemporaryRoot
    TEMP = $TemporaryRoot
    TMP = $TemporaryRoot
    ASTROCYTE_S5_LIVE = '1'
    ASTROCYTE_S5_RUN_ROOT = $RunRoot
}
$PreviousEnvironment = @{}
$LogSuffix = [DateTime]::UtcNow.ToString('yyyyMMddTHHmmssfffffff')
Push-Location $RepositoryRoot
try {
    foreach ($Name in $ProcessEnvironment.Keys) {
        $PreviousEnvironment[$Name] = [Environment]::GetEnvironmentVariable($Name, 'Process')
        [Environment]::SetEnvironmentVariable($Name, $ProcessEnvironment[$Name], 'Process')
    }
    $ProcessEnvironment | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $EvidenceRoot "environment-$LogSuffix.json") -Encoding utf8
    if ($Phase -eq 'tests') {
        $GoArguments = @('test', '-mod=readonly', '-p', '1', './tests/s5', '-run', '^TestActual(RulesVersionsAndFeedbackRestart|GrantRevocationAndAtomicRollback|PublicWorktreeIsolationAndUncommittedChange|CompleteSourcesInvalidateCacheAndAdmission|AutomaticSourcesAndExternalPermission|DelegationBindingsAndAdmission|GoalAndRankingSources|DelegationSourceSnapshot)$', '-count=1', '-v', '-timeout', '8m')
    } elseif ($Phase -eq 'http') {
        $GoArguments = @('test', '-mod=readonly', './internal/adapters/httpapi', '-run', '^TestJudgmentActualHTTPRulesFeedbackAndRestart$', '-count=1', '-v', '-timeout', '2m')
    } else {
        $GoArguments = @('vet', '-mod=readonly', './internal/judgment', './internal/adapters/judgment', './internal/adapters/sqlite', './internal/adapters/httpapi', './tests/s5')
    }
    $GoOutput = Join-Path $EvidenceRoot "$Phase-$LogSuffix.stdout.log"
    $GoErrors = Join-Path $EvidenceRoot "$Phase-$LogSuffix.stderr.log"
    $GoProcess = Start-Process -FilePath 'go' -ArgumentList $GoArguments -WorkingDirectory $RepositoryRoot -WindowStyle Hidden -Wait -PassThru -RedirectStandardOutput $GoOutput -RedirectStandardError $GoErrors
    Get-Content -LiteralPath $GoOutput
    Get-Content -LiteralPath $GoErrors
    if ($GoProcess.ExitCode -ne 0) { throw "S5 $Phase failed with exit code $($GoProcess.ExitCode)" }
    if ($Phase -eq 'checks') {
        & node --check internal/adapters/judgment/laya-worker.mjs
        if ($LASTEXITCODE -ne 0) { throw 'S5 Node syntax check failed' }
    }
} finally {
    foreach ($Name in $PreviousEnvironment.Keys) {
        [Environment]::SetEnvironmentVariable($Name, $PreviousEnvironment[$Name], 'Process')
    }
    Pop-Location
}
