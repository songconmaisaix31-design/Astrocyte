param(
    [Parameter(Mandatory=$true)][string]$Source,
    [Parameter(Mandatory=$true)][string]$WorkRoot,
    [Parameter(Mandatory=$true)][string]$SdkPath,
    [Parameter(Mandatory=$true)][string]$NodePath
)

$ErrorActionPreference = 'Stop'
$RepositoryRoot = (Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
$AllowedRoot = [IO.Path]::GetFullPath('D:/Astrocyte-agent-assets/20261015/relay/S4/')
$WorkRoot = [IO.Path]::GetFullPath($WorkRoot)
if (-not $WorkRoot.StartsWith($AllowedRoot, [StringComparison]::OrdinalIgnoreCase)) { throw 'Production evidence must stay inside the owned S4 asset directory.' }
if (Test-Path -LiteralPath $WorkRoot) { throw 'Use a new production evidence directory; previous attempts are retained.' }
if ($Source -notmatch '^[0-9a-f]{40}$') { throw 'An exact committed source is required.' }
foreach ($Ancestor in @('91609e5d467d2227f32dfdd8717b773340eaf1da', 'aa0a5d310cbfc5c07063bb22fdc951e6ca549195')) {
    & git -C $RepositoryRoot merge-base --is-ancestor $Ancestor $Source
    if ($LASTEXITCODE -ne 0) { throw 'The exact source must contain the committed S4 provider and S2 transaction implementation.' }
}
New-Item -ItemType Directory -Path $WorkRoot | Out-Null
$SourceRoot = Join-Path $WorkRoot 'source'
$EvidenceRoot = Join-Path $WorkRoot 'evidence'
$TemporaryRoot = Join-Path $WorkRoot 'temp'
New-Item -ItemType Directory -Path $SourceRoot, $EvidenceRoot, $TemporaryRoot | Out-Null
$Archive = Join-Path $WorkRoot 'source.zip'

$Environment = @{
    GOMAXPROCS = '2'
    GOFLAGS = '-p=1'
    GOCACHE = 'D:/Astrocyte-agent-assets/20261015/go-build/backend'
    GOMODCACHE = 'D:/Astrocyte-agent-assets/20261015/relay/S4/go-mod-cache'
    GOTMPDIR = $TemporaryRoot
    TEMP = $TemporaryRoot
    TMP = $TemporaryRoot
    ASTROCYTE_TEST_OUTPUT_ROOT = $EvidenceRoot
    ASTROCYTE_S4_SOURCE_ROOT = $SourceRoot
    ASTROCYTE_S4_ROOT_A = Join-Path $RepositoryRoot 'dist/s4/worktree-a'
    ASTROCYTE_S4_ROOT_B = Join-Path $RepositoryRoot 'dist/s4/worktree-b'
    ASTROCYTE_CODEGRAPH_NODE = $NodePath
    ASTROCYTE_CODEGRAPH_SDK = $SdkPath
    ASTROCYTE_AOCI_EXECUTABLE = Join-Path $RepositoryRoot 'dist/s4/tools/aoci.exe'
}
$Saved = @{}
foreach ($Name in $Environment.Keys) {
    $Saved[$Name] = [Environment]::GetEnvironmentVariable($Name, 'Process')
    [Environment]::SetEnvironmentVariable($Name, $Environment[$Name], 'Process')
}
try {
    & git -C $RepositoryRoot archive --format=zip ('--output=' + $Archive) $Source
    if ($LASTEXITCODE -ne 0) { throw 'The committed source archive could not be created.' }
    Expand-Archive -LiteralPath $Archive -DestinationPath $SourceRoot
    $Log = Join-Path $WorkRoot 'production-first.log'
    & $NodePath --test (Join-Path $PSScriptRoot 'production.test.mjs') *> $Log
    $ExitCode = $LASTEXITCODE
    Write-Output ('SOURCE=' + $Source)
    Write-Output ('LOG=' + $Log)
    Write-Output ('EXIT=' + $ExitCode)
    if ($ExitCode -ne 0) { throw 'Actual production S4 verification failed; inspect the retained local log.' }
} finally {
    foreach ($Name in $Saved.Keys) { [Environment]::SetEnvironmentVariable($Name, $Saved[$Name], 'Process') }
}
