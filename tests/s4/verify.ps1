param(
    [Parameter(Mandatory=$true)][string]$Phase,
    [string]$EvidenceName = 'actual-first',
    [string]$SdkPath = '',
    [string]$NodePath = 'node',
    [string]$EvidenceRoot = ''
)

$ErrorActionPreference = 'Stop'
$RepositoryRoot = (Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
if ($EvidenceRoot -eq '') { $EvidenceRoot = Join-Path $RepositoryRoot ".verification/s4/$EvidenceName" }
$EvidenceRoot = [IO.Path]::GetFullPath($EvidenceRoot)
$LocalEvidence = [IO.Path]::GetFullPath((Join-Path $RepositoryRoot '.verification/s4/'))
$AssetEvidence = [IO.Path]::GetFullPath('D:/Astrocyte-agent-assets/20261015/relay/S4/')
if (-not $EvidenceRoot.StartsWith($LocalEvidence, [StringComparison]::OrdinalIgnoreCase) -and -not $EvidenceRoot.StartsWith($AssetEvidence, [StringComparison]::OrdinalIgnoreCase)) { throw 'Evidence must stay inside an owned S4 directory.' }
if ($SdkPath -eq '') { $SdkPath = Join-Path (& npm root -g) '@colbymchenry/codegraph' }
$S4Arguments = @(
    'run', './tests/s4/verify', '-phase', $Phase,
    '-a', (Join-Path $RepositoryRoot 'dist/s4/worktree-a'),
    '-b', (Join-Path $RepositoryRoot 'dist/s4/worktree-b'),
    '-evidence', $EvidenceRoot,
    '-sdk', $SdkPath,
    '-node', $NodePath,
    '-script', (Join-Path $RepositoryRoot 'internal/adapters/repositorycontext/codegraph.mjs'),
    '-aoci', (Join-Path $RepositoryRoot 'dist/s4/tools/aoci.exe')
)
$RunLogName = "$Phase-$([DateTime]::UtcNow.ToString('yyyyMMddTHHmmssfffffff')).log"
& go @S4Arguments 2>&1 | Tee-Object -FilePath (Join-Path $EvidenceRoot $RunLogName)
if ($LASTEXITCODE -ne 0) { throw "S4 verification failed: $Phase" }
