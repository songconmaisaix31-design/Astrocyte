param(
    [Parameter(Mandatory = $true)][string]$Runtime,
    [Parameter(Mandatory = $true)][string]$Output
)

. (Join-Path $PSScriptRoot 'Release-Common.ps1')
Assert-ReleaseAssetPath $Output
Assert-ReleaseAssetPath $Runtime
Assert-ReleaseSpace
if (Test-Path -LiteralPath $Output) { throw 'Use a new probe output directory.' }
New-Item -ItemType Directory -Path $Output | Out-Null
Start-Transcript -Path (Join-Path $Output 'probe.log') | Out-Null
try {
    $node = Join-Path $Runtime 'runtime\node\node.exe'
    & $node (Join-Path $PSScriptRoot 'Check-Runtime.mjs') (Join-Path $Runtime 'runtime')
    if ($LASTEXITCODE -ne 0) { throw 'Actual runtime dependency/asset check failed.' }
    foreach ($file in Get-ChildItem -LiteralPath $PSScriptRoot -Filter '*.ps1') {
        $tokens = $null
        $errors = $null
        [Management.Automation.Language.Parser]::ParseFile($file.FullName, [ref]$tokens, [ref]$errors) | Out-Null
        if ($errors.Count -gt 0) { throw ('PowerShell parser failed: ' + $file.Name + ': ' + $errors[0].Message) }
    }
    $package = Join-Path $Output 'Astrocyte-v0.1-windows-x64'
    Assert-ReleaseSpace
    $runtimeOut = Join-Path $package 'runtime'
    New-Item -ItemType Directory -Path $runtimeOut | Out-Null
    # 此归档仅验证实际依赖的打包路径，不创建应用候选包。
    foreach ($dir in @('node', 'sdk')) { Copy-Item -LiteralPath (Join-Path $Runtime ('runtime\' + $dir)) -Destination $runtimeOut -Recurse }
    if (-not (Test-Path -LiteralPath (Join-Path $runtimeOut 'node\node.exe') -PathType Leaf) -or -not (Test-Path -LiteralPath (Join-Path $runtimeOut 'sdk\package-lock.json') -PathType Leaf)) { throw 'Actual runtime copy is incomplete.' }
    Copy-Item -LiteralPath (Join-Path $Runtime 'legal') -Destination $package -Recurse
    Copy-Item -LiteralPath (Join-Path $Runtime 'NOTICE.md') -Destination $package
    Assert-ReleaseSpace
    foreach ($file in Get-ChildItem -LiteralPath $package -Recurse -File -Force) {
        Assert-ReleaseMember ('Astrocyte-v0.1-windows-x64/' + $file.FullName.Substring($package.Length + 1).Replace('\', '/'))
    }
    & $node (Join-Path $PSScriptRoot 'Inspect-Package.mjs') $package
    if ($LASTEXITCODE -ne 0) { throw 'Actual dependency privacy scan failed.' }
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $zip = Join-Path $Output 'runtime-dependencies-probe.zip'
    New-ReleaseZip $package $zip
    Assert-ReleaseSpace
    $archive = [IO.Compression.ZipFile]::OpenRead($zip)
    try {
        foreach ($entry in $archive.Entries) { if (-not $entry.FullName.EndsWith('/')) { Assert-ReleaseMember $entry.FullName } }
    } finally { $archive.Dispose() }
    $extract = Join-Path $Output 'extracted'
    [IO.Compression.ZipFile]::ExtractToDirectory($zip, $extract)
    Assert-ReleaseSpace
    $extracted = Join-Path $extract 'Astrocyte-v0.1-windows-x64'
    $extractedNode = Join-Path $extracted 'runtime\node\node.exe'
    if ((& $extractedNode --version) -ne 'v24.16.0') { throw 'Extracted Node cannot execute.' }
    & $extractedNode (Join-Path $PSScriptRoot 'Check-CodeGraph.mjs') (Join-Path $extracted 'runtime\sdk\node_modules\@colbymchenry\codegraph')
    if ($LASTEXITCODE -ne 0) { throw 'Extracted real S4 SDK API check failed.' }
    & $extractedNode (Join-Path $PSScriptRoot 'Inspect-Package.mjs') $extracted
    if ($LASTEXITCODE -ne 0) { throw 'Extracted dependency privacy scan failed.' }
    $files = @(Get-ChildItem -LiteralPath $package -Recurse -File -Force)
    $restored = @(Get-ChildItem -LiteralPath $extracted -Recurse -File -Force)
    if ($files.Count -ne $restored.Count) { throw 'Standard ZIP roundtrip file count differs.' }
    foreach ($file in $files) {
        $relative = $file.FullName.Substring($package.Length + 1)
        if ((Get-FileHash -LiteralPath $file.FullName -Algorithm SHA256).Hash -ne (Get-FileHash -LiteralPath (Join-Path $extracted $relative) -Algorithm SHA256).Hash) { throw ('ZIP roundtrip changed actual file: ' + $relative) }
    }
    Assert-ReleaseSpace
    Write-Output ('PASS: actual dependency ZIP and extraction; files=' + $files.Count + '; bytes=' + (Get-Item -LiteralPath $zip).Length + '; PowerShell=' + $PSVersionTable.PSVersion)
} finally { Stop-Transcript | Out-Null }
