param(
    [Parameter(Mandatory = $true)]
    [ValidatePattern('^[0-9a-f]{40}$')]
    [string]$Source,
    [ValidatePattern('^[a-zA-Z0-9-]+$')]
    [string]$Label = 'candidate'
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
if (-not [Environment]::Is64BitOperatingSystem -or [Environment]::OSVersion.Platform -ne 'Win32NT') {
    throw 'Windows x64 build host required.'
}
$repo = Split-Path $PSScriptRoot -Parent
$output = Join-Path $repo ('.verification\package-' + $Label + '-' + (Get-Date -Format 'yyyyMMdd-HHmmss'))
if (Test-Path -LiteralPath $output) { throw 'Output directory already exists.' }
New-Item -ItemType Directory -Path $output | Out-Null
Start-Transcript -Path (Join-Path $output 'build.log') | Out-Null
$priorGo = @{}
foreach ($name in @('GOOS', 'GOARCH', 'CGO_ENABLED', 'GOFLAGS', 'GOMAXPROCS', 'NODE_OPTIONS')) { $priorGo[$name] = [Environment]::GetEnvironmentVariable($name, 'Process') }
try {
    Push-Location $repo
    try {
        $resolved = & git rev-parse --verify ($Source + '^{commit}')
        if ($LASTEXITCODE -ne 0 -or $resolved -ne $Source) { throw 'Exact source commit unavailable.' }
        & git archive --format=zip ('--output=' + (Join-Path $output 'source.zip')) $Source
        if ($LASTEXITCODE -ne 0) { throw 'git archive failed.' }
    } finally { Pop-Location }
    $sourceDir = Join-Path $output 'source'
    Expand-Archive -LiteralPath (Join-Path $output 'source.zip') -DestinationPath $sourceDir
    $env:GOOS = 'windows'
    $env:GOARCH = 'amd64'
    $env:CGO_ENABLED = '0'
    $env:GOFLAGS = '-p=1'
    $env:GOMAXPROCS = '2'
    $env:NODE_OPTIONS = '--max-old-space-size=512'
    Push-Location $sourceDir
    try {
        & go version
        if ($LASTEXITCODE -ne 0) { throw 'Go unavailable.' }
        & node --version
        if ($LASTEXITCODE -ne 0) { throw 'Node unavailable.' }
        & pnpm --version
        if ($LASTEXITCODE -ne 0) { throw 'pnpm unavailable.' }
        & pnpm install --frozen-lockfile
        if ($LASTEXITCODE -ne 0) { throw 'Frozen dependency installation failed.' }
        & pnpm build
        if ($LASTEXITCODE -ne 0) { throw 'Existing production build failed.' }
        & go build -mod=readonly -trimpath -o (Join-Path $sourceDir 'dist\workbench-snapshot.exe') ./cmd/workbench-snapshot
        if ($LASTEXITCODE -ne 0) { throw 'Portable Workbench snapshot build failed.' }
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
    $zip = Join-Path $output ('Astrocyte-v0.1-windows-x64-' + $Source.Substring(0, 12) + '.zip')
    Compress-Archive -LiteralPath $package -DestinationPath $zip -CompressionLevel Optimal
    Write-Output ('SOURCE=' + $Source)
    Write-Output ('ZIP=' + $zip)
} finally {
    foreach ($name in $priorGo.Keys) { [Environment]::SetEnvironmentVariable($name, $priorGo[$name], 'Process') }
    Stop-Transcript | Out-Null
}
