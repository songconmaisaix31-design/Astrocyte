param(
    [Parameter(Mandatory = $true)][string]$Output,
    [Parameter(Mandatory = $true)][string]$EnglishAssets,
    [Parameter(Mandatory = $true)][string]$MultilingualAssets,
    [Parameter(Mandatory = $true)][string]$ModelRoot,
    [string]$NodeArchive = '',
    [string]$NodeChecksums = '',
    [string]$NpmCache = '',
    [string]$DirectMLPackage = '',
    [string]$RuntimeSource = ''
)

. (Join-Path $PSScriptRoot 'Release-Common.ps1')
Add-Type -AssemblyName System.IO.Compression.FileSystem
Assert-ReleaseAssetPath $Output
Assert-ReleaseSpace
if (Test-Path -LiteralPath $Output) { throw 'Use a new runtime preparation directory; failed attempts are retained.' }
New-Item -ItemType Directory -Path $Output | Out-Null
Start-Transcript -Path (Join-Path $Output 'prepare.log') | Out-Null
$saved = @{}
foreach ($name in @('PATH', 'npm_config_cache', 'npm_config_registry', 'GOCACHE', 'GOTMPDIR')) { $saved[$name] = [Environment]::GetEnvironmentVariable($name, 'Process') }
try {
    $env:GOCACHE = 'D:\Astrocyte-agent-assets\20261015\go-build\backend'
    $env:GOTMPDIR = 'D:\Astrocyte-agent-assets\20261015\go-temp\release-package'
    New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null
    if ($NpmCache -eq '') { $NpmCache = Join-Path $Output 'npm-cache' }
    Assert-ReleaseAssetPath $NpmCache
    $env:npm_config_cache = $NpmCache
    $env:npm_config_registry = 'https://registry.npmjs.org'
    $download = Join-Path $Output 'downloads'
    $runtime = Join-Path $Output 'runtime'
    $legal = Join-Path $Output 'legal'
    New-Item -ItemType Directory -Path $download, $runtime, $legal | Out-Null
    function Get-Official([string]$URL, [string]$Destination) {
        & curl.exe --silent --show-error --fail --location --proto '=https' --proto-redir '=https' --tlsv1.2 --connect-timeout 30 --max-time 600 --output $Destination $URL
        if ($LASTEXITCODE -ne 0) { throw ('Official download failed: ' + $URL) }
        Assert-ReleaseSpace
    }
    if ($RuntimeSource -ne '') {
        Assert-ReleaseAssetPath $RuntimeSource
        foreach ($file in @('runtime\node\node.exe', 'runtime\sdk\package-lock.json', 'legal\npm-runtime.cdx.json')) {
            if (-not (Test-Path -LiteralPath (Join-Path $RuntimeSource $file) -PathType Leaf)) { throw ('Explicit prepared runtime source missing: ' + $file) }
        }
        foreach ($dir in @('node', 'sdk')) { Copy-Item -LiteralPath (Join-Path $RuntimeSource ('runtime\' + $dir)) -Destination $runtime -Recurse }
        foreach ($item in Get-ChildItem -LiteralPath (Join-Path $RuntimeSource 'legal') -Force) { Copy-Item -LiteralPath $item.FullName -Destination $legal -Recurse }
        $node = Join-Path $runtime 'node\node.exe'
        Assert-ReleaseSpace
    } else {
    if (($NodeArchive -eq '') -ne ($NodeChecksums -eq '')) { throw 'Explicit NodeArchive and official NodeChecksums must be supplied together.' }
    $nodeZip = $NodeArchive
    $sums = $NodeChecksums
    if ($nodeZip -eq '') {
        $nodeZip = Join-Path $download 'node-v24.16.0-win-x64.zip'
        Get-Official 'https://nodejs.org/dist/v24.16.0/node-v24.16.0-win-x64.zip' $nodeZip
        $sums = Join-Path $download 'SHASUMS256.txt'
        Get-Official 'https://nodejs.org/dist/v24.16.0/SHASUMS256.txt' $sums
    } else {
        Assert-ReleaseAssetPath $nodeZip
        Assert-ReleaseAssetPath $sums
    }
    $expected = @(Select-String -LiteralPath $sums -Pattern '^[0-9a-f]{64}  node-v24\.16\.0-win-x64\.zip$')
    if ($expected.Count -ne 1 -or (Get-FileHash -LiteralPath $nodeZip -Algorithm SHA256).Hash -ine ($expected[0].Line.Split(' ')[0])) { throw 'Official Node archive checksum mismatch.' }
    [IO.Compression.ZipFile]::ExtractToDirectory($nodeZip, (Join-Path $Output 'node-distribution'))
    Assert-ReleaseSpace
    $nodeHome = Join-Path $Output 'node-distribution\node-v24.16.0-win-x64'
    $node = Join-Path $nodeHome 'node.exe'
    $npm = Join-Path $nodeHome 'node_modules\npm\bin\npm-cli.js'
    if ((& $node --version) -ne 'v24.16.0') { throw 'Actual Node version differs.' }
    $env:PATH = $nodeHome + ';' + $env:PATH
    $nodeOut = Join-Path $runtime 'node'
    New-Item -ItemType Directory -Path $nodeOut | Out-Null
    Copy-Item -LiteralPath $node, (Join-Path $nodeHome 'LICENSE') -Destination $nodeOut
    Assert-ReleaseSpace

    $laya = Join-Path $Output 'laya-source'
    & git clone --no-checkout https://github.com/NandhaKishorM/laya.git $laya
    if ($LASTEXITCODE -ne 0) { throw 'Official Laya clone failed.' }
    & git -C $laya -c advice.detachedHead=false checkout --detach 432312b36714d8889eba405507b37eda25257d9c
    if ($LASTEXITCODE -ne 0 -or (& git -C $laya rev-parse HEAD) -ne '432312b36714d8889eba405507b37eda25257d9c') { throw 'Laya source revision differs.' }
    Assert-ReleaseSpace
    $sdk = Join-Path $runtime 'sdk'
    $packages = Join-Path $runtime 'packages'
    New-Item -ItemType Directory -Path $sdk, $packages | Out-Null
    Push-Location (Join-Path $laya 'laya-ts')
    try {
        & $node $npm ci --ignore-scripts --registry=https://registry.npmjs.org --no-audit --no-fund
        if ($LASTEXITCODE -ne 0) { throw 'Fixed official Laya build dependencies failed.' }
        Assert-ReleaseSpace
        & $node $npm run build
        if ($LASTEXITCODE -ne 0) { throw 'Actual official laya-ts subdirectory build failed.' }
        Assert-ReleaseSpace
        & $node $npm pack --ignore-scripts --pack-destination $packages
        if ($LASTEXITCODE -ne 0) { throw 'Standard npm pack of built Laya failed.' }
    } finally { Pop-Location }
    $codegraph = Join-Path $Output 'codegraph-source'
    & git clone --no-checkout https://github.com/colbymchenry/codegraph.git $codegraph
    if ($LASTEXITCODE -ne 0) { throw 'Official CodeGraph clone failed.' }
    & git -C $codegraph -c advice.detachedHead=false checkout --detach a1489f77a6d69538bfe29020b8366ee034f90860
    if ($LASTEXITCODE -ne 0 -or (& git -C $codegraph rev-parse HEAD) -ne 'a1489f77a6d69538bfe29020b8366ee034f90860') { throw 'CodeGraph source revision differs.' }
    Assert-ReleaseSpace
    Push-Location $codegraph
    try {
        & $node $npm ci --ignore-scripts --registry=https://registry.npmjs.org --no-audit --no-fund
        if ($LASTEXITCODE -ne 0) { throw 'Fixed official CodeGraph build dependencies failed.' }
        Assert-ReleaseSpace
        & $node $npm run build
        if ($LASTEXITCODE -ne 0) { throw 'Actual official CodeGraph SDK build failed.' }
        & $node $npm pack --ignore-scripts --pack-destination $packages
        if ($LASTEXITCODE -ne 0) { throw 'Standard npm pack of built CodeGraph failed.' }
    } finally { Pop-Location }
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'runtime.package.json') -Destination (Join-Path $sdk 'package.json')
    Push-Location $sdk
    try {
        & $node $npm install --omit=dev --omit=optional --ignore-scripts --registry=https://registry.npmjs.org --no-audit --no-fund
        if ($LASTEXITCODE -ne 0) { throw 'Explicit production runtime installation failed.' }
        Assert-ReleaseSpace
        & $node $npm ls --omit=dev --omit=optional --all
        if ($LASTEXITCODE -ne 0) { throw 'Production dependency graph is invalid.' }
        $sbom = & $node $npm sbom --sbom-format cyclonedx --omit=dev --omit=optional
        if ($LASTEXITCODE -ne 0) { throw 'Standard production npm SBOM failed.' }
        [IO.File]::WriteAllText((Join-Path $legal 'npm-runtime.cdx.json'), ($sbom -join [Environment]::NewLine), [Text.UTF8Encoding]::new($false))
        $graphRevision = & $node $npm view '@colbymchenry/codegraph@1.0.1' gitHead --registry=https://registry.npmjs.org
        if ($LASTEXITCODE -ne 0 -or $graphRevision -ne 'a1489f77a6d69538bfe29020b8366ee034f90860') { throw 'Official CodeGraph npm source revision differs.' }
    } finally { Pop-Location }
    Copy-Item -LiteralPath (Join-Path $laya 'LICENSE') -Destination (Join-Path $legal 'laya-LICENSE')
    Get-Official 'https://huggingface.co/convaiinnovations/laya/raw/7b928d828b7b0e022f929d9bd2e44165aa270148/README.md' (Join-Path $legal 'laya-model-card.md')
    Get-Official 'https://raw.githubusercontent.com/sourcegraph/jsonrpc2/2cc94179e136b565cb9df7d9d0bff9fc06912ac9/LICENSE' (Join-Path $legal 'go-jsonrpc2-LICENSE')
    Get-Official 'https://raw.githubusercontent.com/colbymchenry/codegraph/a1489f77a6d69538bfe29020b8366ee034f90860/LICENSE' (Join-Path $legal 'codegraph-LICENSE')
    & (Join-Path $PSScriptRoot 'Prepare-OnnxLegal.ps1') -Output $legal -Downloads $download -DirectMLPackage $DirectMLPackage
    }
    foreach ($checkpoint in @('english', 'multilingual')) {
        $descriptor = if ($checkpoint -eq 'english') { $EnglishAssets } else { $MultilingualAssets }
        $metadata = Get-Content -LiteralPath $descriptor -Raw -Encoding UTF8 | ConvertFrom-Json
        if ($metadata.repo -ne 'convaiinnovations/laya' -or $metadata.model_revision -ne '7b928d828b7b0e022f929d9bd2e44165aa270148' -or $metadata.source_revision -ne '432312b36714d8889eba405507b37eda25257d9c' -or $metadata.device -ne 'cpu' -or $metadata.checkpoint -ne $checkpoint) { throw 'Actual S5 asset metadata differs.' }
        $names = @('encoder.onnx', 'encoder.onnx.data', 'head.onnx', 'head.onnx.data', 'tokenizer.json', 'rl_agent_config.json')
        if (@($metadata.files.PSObject.Properties).Count -ne $names.Count) { throw 'S5 asset file set differs.' }
        $target = Join-Path $runtime ('models\' + $checkpoint)
        Assert-ReleaseSpace
        New-Item -ItemType Directory -Path $target | Out-Null
        foreach ($name in $names) {
            $file = Join-Path $ModelRoot ($checkpoint + '\' + $name)
            if ((Get-FileHash -LiteralPath $file -Algorithm SHA256).Hash -ine $metadata.files.$name) { throw ('Actual S5 asset differs: ' + $checkpoint + '/' + $name) }
            Copy-Item -LiteralPath $file -Destination $target
            if ((Get-FileHash -LiteralPath (Join-Path $target $name) -Algorithm SHA256).Hash -ine $metadata.files.$name) { throw 'Copied S5 asset differs.' }
        }
        Copy-Item -LiteralPath $descriptor -Destination (Join-Path $target 'assets.json')
        Assert-ReleaseSpace
    }
    Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'NOTICE.md') -Destination $Output
    & $node (Join-Path $PSScriptRoot 'Check-Runtime.mjs') $runtime
    if ($LASTEXITCODE -ne 0) { throw 'Actual packaged runtime import or model asset check failed.' }
    & $node (Join-Path $PSScriptRoot 'Inspect-Package.mjs') $runtime
    if ($LASTEXITCODE -ne 0) { throw 'Prepared runtime privacy check failed; retain assets and hand off to their owner.' }
    Assert-ReleaseSpace
    Write-Output ('PREPARED_RUNTIME=' + $runtime)
} finally {
    foreach ($name in $saved.Keys) { [Environment]::SetEnvironmentVariable($name, $saved[$name], 'Process') }
    Stop-Transcript | Out-Null
}
