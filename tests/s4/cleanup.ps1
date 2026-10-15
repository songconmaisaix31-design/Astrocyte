param(
    [Parameter(Mandatory=$true)][string]$WorkRoot,
    [Parameter(Mandatory=$true)][string]$SdkPath,
    [Parameter(Mandatory=$true)][string]$NodePath
)

$ErrorActionPreference = 'Stop'
$AllowedRoot = [IO.Path]::GetFullPath('D:/Astrocyte-agent-assets/20261015/relay/S4/')
$WorkRoot = [IO.Path]::GetFullPath($WorkRoot)
if (-not $WorkRoot.StartsWith($AllowedRoot, [StringComparison]::OrdinalIgnoreCase)) { throw 'Cleanup evidence must stay inside the owned S4 asset directory.' }
if (Test-Path -LiteralPath $WorkRoot) { throw 'Use a new cleanup evidence directory; previous attempts are retained.' }
New-Item -ItemType Directory -Path $WorkRoot | Out-Null
$Saved = @{}
foreach ($Name in @('GOCACHE', 'GOTMPDIR', 'TEMP', 'TMP')) { $Saved[$Name] = [Environment]::GetEnvironmentVariable($Name, 'Process') }
try {
    $env:GOCACHE = 'D:/Astrocyte-agent-assets/20261015/go-build/backend'
    $env:GOTMPDIR = 'D:/Astrocyte-agent-assets/20261015/go-temp/S4'
    $env:TEMP = $env:GOTMPDIR
    $env:TMP = $env:GOTMPDIR
    foreach ($Phase in @('initial', 'cleanup', 'cleanup-restart', 'cleanup-worker', 'cleanup-worker-restart', 'cleanup-sql-boundary')) {
        & (Join-Path $PSScriptRoot 'verify.ps1') -Phase $Phase -EvidenceRoot $WorkRoot -SdkPath $SdkPath -NodePath $NodePath
    }
} finally {
    foreach ($Name in $Saved.Keys) { [Environment]::SetEnvironmentVariable($Name, $Saved[$Name], 'Process') }
}
