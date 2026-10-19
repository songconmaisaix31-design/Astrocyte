param(
    [Parameter(Mandatory = $true)]
    [ValidatePattern('^[0-9a-f]{40}$')]
    [string]$Source,
    [ValidatePattern('^[a-zA-Z0-9-]+$')]
    [string]$Label = 'candidate',
    [Parameter(Mandatory = $true)]
    [string]$Runtime,
    [string]$Output = ''
)

. (Join-Path $PSScriptRoot 'Release-Common.ps1')
Add-Type -AssemblyName System.IO.Compression.FileSystem
if (-not [Environment]::Is64BitOperatingSystem -or [Environment]::OSVersion.Platform -ne 'Win32NT') {
    throw 'Windows x64 build host required.'
}
$repo = Split-Path $PSScriptRoot -Parent
if ($Output -eq '') { $Output = Join-Path ('D:\Astrocyte-agent-assets\20261015\release-package\today-' + (Get-Date -Format 'yyyyMMdd')) ('package-' + $Label + '-' + (Get-Date -Format 'yyyyMMdd-HHmmss')) }
Assert-ReleaseAssetPath $Output
Assert-ReleaseAssetPath $Runtime
Assert-ReleaseSpace
$output = [IO.Path]::GetFullPath($Output)
foreach ($path in @('runtime\node\node.exe', 'runtime\sdk\package-lock.json', 'legal\laya-LICENSE', 'legal\laya-model-card.md', 'legal\go-jsonrpc2-LICENSE', 'legal\codegraph-LICENSE', 'legal\onnxruntime-LICENSE', 'legal\onnxruntime-ThirdPartyNotices.txt', 'legal\DirectML-LICENSE.txt', 'legal\DirectML-LICENSE-CODE.txt', 'legal\DirectML-ThirdPartyNotices.txt', 'legal\DirectXShaderCompiler-LICENSE.txt', 'legal\npm-runtime.cdx.json', 'NOTICE.md')) {
    if (-not (Test-Path -LiteralPath (Join-Path $Runtime $path) -PathType Leaf)) { throw ('Prepared runtime input missing: ' + $path) }
}
if (Test-Path -LiteralPath $output) { throw 'Output directory already exists.' }
New-Item -ItemType Directory -Path $output | Out-Null
Start-Transcript -Path (Join-Path $output 'build.log') | Out-Null
$priorGo = @{}
$buildProcess = [Diagnostics.Process]::GetCurrentProcess()
$priorPriority = $buildProcess.PriorityClass
foreach ($name in @('GOOS', 'GOARCH', 'CGO_ENABLED', 'GOFLAGS', 'GOMAXPROCS', 'GOMEMLIMIT', 'NODE_OPTIONS', 'GOCACHE', 'GOTMPDIR', 'TEMP', 'TMP', 'npm_config_cache', 'NODE_COMPILE_CACHE', 'ASTROCYTE_TEST_OUTPUT_ROOT', 'PATH')) { $priorGo[$name] = [Environment]::GetEnvironmentVariable($name, 'Process') }
try {
    $buildProcess.PriorityClass = 'BelowNormal'
    $env:GOOS = 'windows'
    $env:GOARCH = 'amd64'
    $env:CGO_ENABLED = '0'
    $env:GOFLAGS = '-p=4'
    $env:GOMAXPROCS = '4'
    $env:GOMEMLIMIT = '512MiB'
    $env:NODE_OPTIONS = '--max-old-space-size=1024'
    Write-Output 'BUILD_LIMITS=BelowNormal GOMAXPROCS=4 GOFLAGS=-p=4 GOMEMLIMIT=512MiB NODE_HEAP_MIB=1024'
    $env:GOCACHE = 'D:\Astrocyte-agent-assets\20261015\go-build\backend'
    $env:GOTMPDIR = Join-Path $output 'go-temp'
    $env:TEMP = Join-Path $output 'temp'
    $env:TMP = $env:TEMP
    $env:npm_config_cache = Join-Path $output 'npm-cache'
    $env:NODE_COMPILE_CACHE = Join-Path $output 'node-compile-cache'
    $env:ASTROCYTE_TEST_OUTPUT_ROOT = Join-Path $output 'test-output'
    foreach ($dir in @($env:GOTMPDIR, $env:TEMP, $env:npm_config_cache, $env:NODE_COMPILE_CACHE)) { New-Item -ItemType Directory -Force -Path $dir | Out-Null }
    $env:PATH = (Join-Path $Runtime 'runtime\node') + ';' + $env:PATH
    # Node 运行时和隐私检查在 D 环境作用域内、source archive 前执行。
    Assert-ReleaseResources 'runtime-check' $output
    & (Join-Path $Runtime 'runtime\node\node.exe') (Join-Path $PSScriptRoot 'Check-Runtime.mjs') (Join-Path $Runtime 'runtime')
    if ($LASTEXITCODE -ne 0) { throw 'Prepared real runtime is invalid.' }
    Assert-ReleaseResources 'runtime-privacy' $output
    & (Join-Path $Runtime 'runtime\node\node.exe') (Join-Path $PSScriptRoot 'Inspect-Package.mjs') (Join-Path $Runtime 'runtime')
    if ($LASTEXITCODE -ne 0) { throw 'Prepared runtime privacy check failed before source archive.' }
    Push-Location $repo
    try {
        $resolved = & git rev-parse --verify ($Source + '^{commit}')
        if ($LASTEXITCODE -ne 0 -or $resolved -ne $Source) { throw 'Exact source commit unavailable.' }
        Assert-ReleaseResources 'source-archive' $output
        & git archive --format=zip ('--output=' + (Join-Path $output 'source.zip')) $Source
        if ($LASTEXITCODE -ne 0) { throw 'git archive failed.' }
        Assert-ReleaseSpace
    } finally { Pop-Location }
    $sourceDir = Join-Path $output 'source'
    Assert-ReleaseResources 'source-extract' $output
    [IO.Compression.ZipFile]::ExtractToDirectory((Join-Path $output 'source.zip'), $sourceDir)
    Assert-ReleaseSpace
    Push-Location $sourceDir
    try {
        & go version
        if ($LASTEXITCODE -ne 0) { throw 'Go unavailable.' }
        & node --version
        if ($LASTEXITCODE -ne 0) { throw 'Node unavailable.' }
        & pnpm --version
        if ($LASTEXITCODE -ne 0) { throw 'pnpm unavailable.' }
        $pnpmConfig = @('--config.registry=https://registry.npmjs.org', ('--config.cache-dir=' + (Join-Path $output 'pnpm-cache')), ('--config.state-dir=' + (Join-Path $output 'pnpm-state')))
        $registry = & pnpm @pnpmConfig config get registry
        if ($LASTEXITCODE -ne 0 -or ([string]$registry).TrimEnd('/') -cne 'https://registry.npmjs.org') { throw 'Effective pnpm registry differs from the selected official registry.' }
        Write-Output ('PNPM_REGISTRY=' + $registry)
        Assert-ReleaseResources 'dependency-install' $output
        & pnpm @pnpmConfig install --frozen-lockfile --store-dir (Join-Path $output 'pnpm-store')
        if ($LASTEXITCODE -ne 0) { throw 'Frozen dependency installation failed.' }
        Assert-ReleaseSpace
        Assert-ReleaseResources 'frontend-lint' $output
        & pnpm @pnpmConfig --dir web lint
        if ($LASTEXITCODE -ne 0) { throw 'Existing frontend lint failed.' }
        Assert-ReleaseResources 'production-build' $output
        & pnpm @pnpmConfig build
        if ($LASTEXITCODE -ne 0) { throw 'Existing production build failed.' }
        Assert-ReleaseResources 'snapshot-build' $output
        & go build -mod=readonly -trimpath -o (Join-Path $sourceDir 'dist\workbench-snapshot.exe') ./cmd/workbench-snapshot
        if ($LASTEXITCODE -ne 0) { throw 'Portable Workbench snapshot build failed.' }
        Assert-ReleaseSpace
    } finally { Pop-Location }
    $package = Join-Path $output 'Astrocyte-v0.1-windows-x64'
    New-Item -ItemType Directory -Path $package | Out-Null
    # 只复制指定的生产产物；开发源码、数据库和本机配置不进入包。
    foreach ($file in @('astrocyte.exe', 'astrocyte-mcp.exe', 'astrocyte-sleep.exe', 'workbench-snapshot.exe')) {
        Copy-Item -LiteralPath (Join-Path $sourceDir ('dist\' + $file)) -Destination $package
    }
    Copy-Item -LiteralPath (Join-Path $sourceDir 'web\dist') -Destination (Join-Path $package 'web') -Recurse
    foreach ($file in @('Start-Astrocyte.ps1', 'Start-Astrocyte.cmd', 'Publish-WorkbenchSnapshot.ps1')) {
        Copy-Item -LiteralPath (Join-Path $sourceDir ('release\' + $file)) -Destination $package
    }
    Copy-Item -LiteralPath (Join-Path $sourceDir 'docs\release\windows.md') -Destination (Join-Path $package 'README.md')
    Copy-Item -LiteralPath (Join-Path $sourceDir 'LICENSE') -Destination $package
    Copy-Item -LiteralPath (Join-Path $Runtime 'legal') -Destination $package -Recurse
    Copy-Item -LiteralPath (Join-Path $sourceDir 'release\NOTICE.md') -Destination $package
    $runtimeOut = Join-Path $package 'runtime'
    New-Item -ItemType Directory -Path $runtimeOut | Out-Null
    foreach ($dir in @('node', 'sdk', 'models')) {
        Assert-ReleaseResources ('copy-' + $dir) $output
        Copy-Item -LiteralPath (Join-Path $Runtime ('runtime\' + $dir)) -Destination $runtimeOut -Recurse
    }
    Assert-ReleaseSpace
    # 生产适配脚本必须来自本次准确应用源码。
    New-Item -ItemType Directory -Path (Join-Path $runtimeOut 'adapters') | Out-Null
    foreach ($script in @('judgment\laya-worker.mjs', 'repositorycontext\codegraph.mjs')) {
        Copy-Item -LiteralPath (Join-Path $sourceDir ('internal\adapters\' + $script)) -Destination (Join-Path $runtimeOut 'adapters')
    }
    Assert-ReleaseResources 'package-privacy' $output
    & (Join-Path $runtimeOut 'node\node.exe') (Join-Path $PSScriptRoot 'Inspect-Package.mjs') $package
    if ($LASTEXITCODE -ne 0) { throw 'Production privacy scan failed.' }
    foreach ($file in Get-ChildItem -LiteralPath $package -Recurse -File -Force) {
        Assert-ReleaseMember ('Astrocyte-v0.1-windows-x64/' + $file.FullName.Substring($package.Length + 1).Replace('\', '/'))
    }
    $zip = Join-Path $output ('Astrocyte-v0.1-windows-x64-' + $Source.Substring(0, 12) + '.zip')
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    Assert-ReleaseResources 'package-zip' $output
    New-ReleaseZip $package $zip
    Assert-ReleaseSpace
    Write-Output ('SOURCE=' + $Source)
    Write-Output ('ZIP=' + $zip)
} finally {
    foreach ($name in $priorGo.Keys) { [Environment]::SetEnvironmentVariable($name, $priorGo[$name], 'Process') }
    $buildProcess.PriorityClass = $priorPriority
    Stop-Transcript | Out-Null
}
