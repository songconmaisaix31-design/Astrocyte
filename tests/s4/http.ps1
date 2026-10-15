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
if (-not $WorkRoot.StartsWith($AllowedRoot, [StringComparison]::OrdinalIgnoreCase)) { throw 'HTTP evidence must stay inside the owned S4 asset directory.' }
if (Test-Path -LiteralPath $WorkRoot) { throw 'Use a new HTTP evidence directory; previous attempts are retained.' }
if ($Source -notmatch '^[0-9a-f]{40}$') { throw 'An exact source commit is required.' }
$ActualSource = & git -C $RepositoryRoot rev-parse ($Source + '^{commit}')
if ($LASTEXITCODE -ne 0 -or $ActualSource -cne $Source) { throw 'The requested source commit is unavailable.' }

$OriginalEvidence = Join-Path $RepositoryRoot '.verification/s4/actual-first'
$OriginalDatabase = Join-Path $OriginalEvidence 'state.sqlite'
$OriginalProjections = Join-Path $OriginalEvidence 'projections'
foreach ($Sidecar in @($OriginalDatabase + '-wal', $OriginalDatabase + '-journal')) {
    if ((Test-Path -LiteralPath $Sidecar) -and (Get-Item -LiteralPath $Sidecar).Length -gt 0) { throw 'Close the owned source SQLite before copying its snapshot.' }
}
foreach ($Wal in Get-ChildItem -LiteralPath $OriginalProjections -Recurse -File -Filter '*-wal') {
    if ($Wal.Length -gt 0) { throw 'The owned projection has uncheckpointed writes.' }
}

$SourceRoot = Join-Path $WorkRoot 'source'
$EvidenceRoot = Join-Path $WorkRoot 'evidence'
New-Item -ItemType Directory -Path $WorkRoot, $SourceRoot, $EvidenceRoot | Out-Null
$Archive = Join-Path $WorkRoot 'source.zip'
& git -C $RepositoryRoot archive --format=zip ('--output=' + $Archive) $Source
if ($LASTEXITCODE -ne 0) { throw 'The exact source archive could not be created.' }
Expand-Archive -LiteralPath $Archive -DestinationPath $SourceRoot
Copy-Item -LiteralPath $OriginalDatabase -Destination (Join-Path $EvidenceRoot 'state.sqlite')
Copy-Item -LiteralPath $OriginalProjections -Destination (Join-Path $EvidenceRoot 'projections') -Recurse

# overlay只增加本轨验证入口，生产源码保持原样。
$EntryRoot = Join-Path $SourceRoot 'tests/s4/http'
New-Item -ItemType Directory -Path $EntryRoot | Out-Null
$Replacements = @{}
$Replacements[(Join-Path $EntryRoot 'main.go')] = Join-Path $PSScriptRoot '_http/main.go'
$Overlay = Join-Path $WorkRoot 'overlay.json'
[IO.File]::WriteAllText($Overlay, (@{ Replace = $Replacements } | ConvertTo-Json -Depth 4), [Text.UTF8Encoding]::new($false))
$Log = Join-Path $WorkRoot 'http-first.log'
$Saved = @{}
foreach ($Name in @('GOCACHE', 'GOTMPDIR', 'TEMP', 'TMP')) { $Saved[$Name] = [Environment]::GetEnvironmentVariable($Name, 'Process') }
try {
    $env:GOCACHE = 'D:/Astrocyte-agent-assets/20261015/go-build/backend'
    $env:GOTMPDIR = 'D:/Astrocyte-agent-assets/20261015/go-temp/S4'
    $env:TEMP = $env:GOTMPDIR
    $env:TMP = $env:GOTMPDIR
    $Arguments = @(
        '-C', $SourceRoot, 'run', '-mod=readonly', ('-overlay=' + $Overlay), './tests/s4/http',
        '-evidence', $EvidenceRoot, '-sdk', $SdkPath, '-node', $NodePath,
        '-script', (Join-Path $SourceRoot 'internal/adapters/repositorycontext/codegraph.mjs'),
        '-aoci', (Join-Path $RepositoryRoot 'dist/s4/tools/aoci.exe')
    )
    & go @Arguments *> $Log
    if ($LASTEXITCODE -ne 0) { throw ('Actual HTTP verification failed; inspect ' + $Log) }
    Write-Output 'PASS actual_repository_http'
    Write-Output ('SOURCE=' + $Source)
    Write-Output ('LOG=' + $Log)
} finally {
    foreach ($Name in $Saved.Keys) { [Environment]::SetEnvironmentVariable($Name, $Saved[$Name], 'Process') }
}
