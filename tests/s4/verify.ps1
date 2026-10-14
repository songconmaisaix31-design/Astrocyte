param(
    [Parameter(Mandatory=$true)][string]$Phase,
    [string]$EvidenceName = 'actual-first',
    [string]$SdkPath = ''
)

$ErrorActionPreference = 'Stop'
$RepositoryRoot = (Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
$EvidenceRoot = Join-Path $RepositoryRoot ".verification/s4/$EvidenceName"
if ($SdkPath -eq '') { $SdkPath = Join-Path (& npm root -g) '@colbymchenry/codegraph' }
$S4Arguments = @(
    'run', './tests/s4/verify', '-phase', $Phase,
    '-a', (Join-Path $RepositoryRoot 'dist/s4/worktree-a'),
    '-b', (Join-Path $RepositoryRoot 'dist/s4/worktree-b'),
    '-evidence', $EvidenceRoot,
    '-sdk', $SdkPath,
    '-script', (Join-Path $RepositoryRoot 'internal/adapters/repositorycontext/codegraph.mjs'),
    '-aoci', (Join-Path $RepositoryRoot 'dist/s4/tools/aoci.exe')
)
$RunLogName = "$Phase-$([DateTime]::UtcNow.ToString('yyyyMMddTHHmmssfffffff')).log"
& go @S4Arguments 2>&1 | Tee-Object -FilePath (Join-Path $EvidenceRoot $RunLogName)
if ($LASTEXITCODE -ne 0) { throw "S4 verification failed: $Phase" }
