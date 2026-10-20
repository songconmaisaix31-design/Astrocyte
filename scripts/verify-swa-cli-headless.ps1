#Requires -Version 7.2
[CmdletBinding()]
param(
    [Parameter(Mandatory)][ValidateSet('T01','T09')][string]$Case,
    [Parameter(Mandatory)][string]$SwaPath,
    [Parameter(Mandatory)][string]$InstallRoot,
    [Parameter(Mandatory)][string]$DataRoot,
    [Parameter(Mandatory)][string]$CaseRoot
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
if (-not (Test-Path -LiteralPath $swaNode -PathType Leaf)) { throw '同候选 Node 不存在' }
$swaEvidenceParent = Join-Path $CaseRoot '.verification/swa-cli-headless'
$swaEvidence = Join-Path $swaEvidenceParent ($Case.ToLowerInvariant() + '-' + [Guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $swaEvidence -Force | Out-Null
$env:TEMP = Join-Path $swaEvidence 'temp'
$env:TMP = $env:TEMP
New-Item -ItemType Directory -Path $env:TEMP | Out-Null
$swaGate = Join-Path (Split-Path $PSScriptRoot -Parent) '.verification/swa-cli-headless-20261020/resource-gate.py'
& C:/Python313/python.exe $swaGate 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'resources.log')
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
if ($Case -eq 'T01') {
    & $swaNode --max-old-space-size=1024 (Join-Path $PSScriptRoot 'check-swa-runtime.mjs') --candidate $InstallRoot 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'closure.log')
} else {
    $swaFixture = Join-Path (Split-Path $PSScriptRoot -Parent) 'tests/swa-cli-headless/provider-job.test.mjs'
    & $swaNode --max-old-space-size=1024 $swaFixture --candidate $InstallRoot --output (Join-Path $swaEvidence 'ordinary-processes.json') 2>&1 | Tee-Object -FilePath (Join-Path $swaEvidence 'ordinary-processes.log')
}
$swaExit = $LASTEXITCODE
[IO.File]::WriteAllText((Join-Path $swaEvidence 'exit.txt'), [string]$swaExit)
if ($Case -eq 'T09' -and $swaExit -eq 0) {
    $swaObserved = Get-Content -LiteralPath (Join-Path $swaEvidence 'ordinary-processes.json') -Raw | ConvertFrom-Json
    if ($swaObserved.remainingCases.Count -gt 0) {
        Write-Warning ('T09 仍有必要分支未运行: ' + ($swaObserved.remainingCases -join ', '))
        exit 6
    }
}
exit $swaExit
