param(
    [Parameter(Mandatory = $true)][ValidatePattern('^[0-9a-f]{40}$')][string]$Source,
    [Parameter(Mandatory = $true)][ValidateSet('repositorycontext', 'judgment')][string]$Component,
    [Parameter(Mandatory = $true)][string]$Runtime,
    [Parameter(Mandatory = $true)][string]$Output
)

. (Join-Path $PSScriptRoot 'Release-Common.ps1')
Add-Type -AssemblyName System.IO.Compression.FileSystem
Assert-ReleaseAssetPath $Runtime
Assert-ReleaseAssetPath $Output
Assert-ReleaseSpace
if (Test-Path -LiteralPath $Output) { throw 'Use a new adapter preparation directory.' }
New-Item -ItemType Directory -Path $Output | Out-Null
$relative = if ($Component -eq 'repositorycontext') { 'internal/adapters/repositorycontext/codegraph.mjs' } else { 'internal/adapters/judgment/laya-worker.mjs' }
$target = Join-Path $Runtime ('runtime\adapters\' + [IO.Path]::GetFileName($relative))
if (Test-Path -LiteralPath $target) { throw 'Adapter destination must be new; retain existing source-specific assets.' }
Push-Location (Split-Path $PSScriptRoot -Parent)
try {
    $resolved = & git rev-parse --verify ($Source + '^{commit}')
    if ($LASTEXITCODE -ne 0 -or $resolved -ne $Source) { throw 'Exact adapter source commit unavailable.' }
    $blob = & git rev-parse ($Source + ':' + $relative)
    if ($LASTEXITCODE -ne 0) { throw 'Exact adapter source blob unavailable.' }
    & git archive --format=zip ('--output=' + (Join-Path $Output 'adapter-source.zip')) $Source -- $relative
    if ($LASTEXITCODE -ne 0) { throw 'Single committed adapter archive failed.' }
} finally { Pop-Location }
[IO.Compression.ZipFile]::ExtractToDirectory((Join-Path $Output 'adapter-source.zip'), (Join-Path $Output 'source'))
New-Item -ItemType Directory -Force -Path (Split-Path $target -Parent) | Out-Null
Copy-Item -LiteralPath (Join-Path (Join-Path $Output 'source') $relative) -Destination $target
$actual = & git hash-object -- $target
if ($LASTEXITCODE -ne 0 -or $actual -ne $blob) { throw 'Prepared adapter differs from exact committed Git blob.' }
& (Join-Path $Runtime 'runtime\node\node.exe') --check $target
if ($LASTEXITCODE -ne 0) { throw 'Committed adapter syntax check failed.' }
Assert-ReleaseSpace
Write-Output ('ADAPTER_SOURCE=' + $Source)
Write-Output ('ADAPTER_BLOB=' + $blob)
Write-Output ('ADAPTER_PATH=' + $target)
