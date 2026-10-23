[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$SourceRoot,
    [Parameter(Mandatory)][string]$Runtime,
    [Parameter(Mandatory)][string]$Node,
    [Parameter(Mandatory)][string]$Work,
    [Parameter(Mandatory)][string]$OutputFile,
    [string]$Python = 'python'
)
$ErrorActionPreference = 'Stop'
$phase = 'validate'
function Assert-Plain([string]$Path) {
    $current = [IO.Path]::GetFullPath($Path)
    while ($current) {
        $item = Get-Item -LiteralPath $current -Force
        if ($item.Attributes -band [IO.FileAttributes]::ReparsePoint) { throw 'A diagnostic input is redirected' }
        $parent = [IO.Directory]::GetParent($current)
        $current = if ($null -eq $parent) { $null } else { $parent.FullName }
    }
}
function Invoke-Original([string]$Stage, [string]$Executable, [string[]]$Arguments) {
    $script:phase = $Stage
    & $Executable @Arguments *> (Join-Path $Work ($Stage + '.raw.log'))
    if ($LASTEXITCODE -ne 0) { throw ('Original component operation failed: ' + $Stage) }
}
try {
    if (-not $IsWindows -or $PSVersionTable.PSVersion.Major -lt 7) { throw 'Windows PowerShell 7 is required' }
    foreach ($path in @($SourceRoot, $Runtime, $Node, $Work)) {
        if (-not [IO.Path]::IsPathFullyQualified($path)) { throw 'Absolute component inputs are required' }
        Assert-Plain $path
    }
    $checkout = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
    $composition = [IO.Path]::GetFullPath((Join-Path $env:RUNNER_TEMP 'swa-controller-open-development-7ff3a7d7'))
    foreach ($pair in @(
        @($SourceRoot, (Join-Path $composition 'source-7ff3a7d7')),
        @($Runtime, (Join-Path $composition 'development-7ff3a7d7/runtime')),
        @($Node, (Join-Path $composition 'development-7ff3a7d7/node/node.exe')),
        @($OutputFile, (Join-Path $composition 'private/diagnostic-native-preparation.json'))
    )) {
        if (-not [IO.Path]::GetFullPath($pair[0]).Equals([IO.Path]::GetFullPath($pair[1]), [StringComparison]::OrdinalIgnoreCase)) { throw 'The diagnostic source or output escaped the exclusive development composition' }
    }
    $buildRoot = Join-Path $checkout '.verification/swa-cli-headless-20261020/build'
    if (-not [IO.Directory]::GetParent($Work).FullName.Equals($buildRoot, [StringComparison]::OrdinalIgnoreCase) -or [IO.Path]::GetFileName($Work) -cnotmatch '^candidate-[A-Za-z0-9]{6}$') { throw 'The guarded exclusive native work directory is required' }
    if (-not [IO.Path]::IsPathFullyQualified($OutputFile)) { throw 'An absolute preparation result is required' }
    Assert-Plain ([IO.Path]::GetDirectoryName($OutputFile))
    if (Test-Path -LiteralPath $OutputFile) { throw 'The diagnostic preparation result already exists' }
    $pinJSON = & $Node --input-type=module -e 'import{pathToFileURL}from"node:url";const m=await import(pathToFileURL(process.argv[1]));const p=m.NODE_RUNTIME_PIN,a=m.NODE_RUNTIME_ASSETS["win32-x64"],l=p.windowsImportLibs["win32-x64"];console.log(JSON.stringify({version:p.version,asset:a,headers:p.headers,library:l,headersUrl:m.nodeRuntimeReleaseUrl("official",p.headers.file),libraryUrl:m.nodeRuntimeReleaseUrl("official",l.file)}))' (Join-Path $SourceRoot 'third_party/orca-v1.4.230/src/shared/node-runtime-pin.ts')
    if ($LASTEXITCODE -ne 0) { throw 'The fixed runtime pin is unavailable' }
    $pin = $pinJSON | ConvertFrom-Json
    if ($pin.version -cne '24.21.0' -or (Get-FileHash -LiteralPath $Node -Algorithm SHA256).Hash.ToLowerInvariant() -cne $pin.asset.executableSha256) { throw 'The retained Node differs from the original pin' }
    $pythonJSON = & $Node --input-type=module -e 'import{pathToFileURL}from"node:url";const m=await import(pathToFileURL(process.argv[1]));console.log(JSON.stringify(await m.buildPython(process.argv[2])))' (Join-Path $SourceRoot 'scripts/swa-package-config.mjs') $Python
    if ($LASTEXITCODE -ne 0) { throw 'The original Python preflight failed' }
    $pythonInfo = $pythonJSON | ConvertFrom-Json
    $Python = $pythonInfo.executable
    $download = Join-Path $checkout 'scripts/download-swa-node.ps1'
    if ((Get-FileHash -LiteralPath $download -Algorithm SHA256).Hash -cne (Get-FileHash -LiteralPath (Join-Path $SourceRoot 'scripts/download-swa-node.ps1') -Algorithm SHA256).Hash) { throw 'The original guarded downloader differs' }
    $powershell = Join-Path $PSHOME 'pwsh.exe'
    $headersArchive = Join-Path $Work $pin.headers.file
    Invoke-Original 'headers-download' $powershell @('-NoProfile', '-NonInteractive', '-File', $download, '-Uri', $pin.headersUrl, '-WorkDirectory', $Work, '-OutputFile', $headersArchive)
    if ((Get-FileHash -LiteralPath $headersArchive -Algorithm SHA256).Hash.ToLowerInvariant() -cne $pin.headers.sha256) { throw 'The original headers checksum differs' }
    $headersRoot = Join-Path $Work 'headers'
    New-Item -ItemType Directory -Path $headersRoot | Out-Null
    Invoke-Original 'headers-extract' (Join-Path $env:SystemRoot 'System32/tar.exe') @('-xzf', $headersArchive, '-C', $headersRoot)
    $nodeDir = Join-Path $headersRoot 'node-v24.21.0'
    New-Item -ItemType Directory -Path (Join-Path $nodeDir 'Release') | Out-Null
    $nodeLibrary = Join-Path $nodeDir 'Release/node.lib'
    Invoke-Original 'node-library-download' $powershell @('-NoProfile', '-NonInteractive', '-File', $download, '-Uri', $pin.libraryUrl, '-WorkDirectory', $Work, '-OutputFile', $nodeLibrary)
    if ((Get-FileHash -LiteralPath $nodeLibrary -Algorithm SHA256).Hash.ToLowerInvariant() -cne $pin.library.sha256) { throw 'The original import library checksum differs' }
    $toolRuntime = Join-Path $SourceRoot 'runtime/swa-orca'
    $gyp = (& $Node -e 'const r=require("node:module").createRequire(process.argv[1]);if(r("node-gyp/package.json").version!=="12.3.0")process.exit(2);process.stdout.write(r.resolve("node-gyp/bin/node-gyp.js"))' (Join-Path $toolRuntime 'package.json')).Trim()
    if ($LASTEXITCODE -ne 0) { throw 'The original locked node-gyp differs' }
    $phase = 'official-msvc-toolchain'
    $vswhere = Join-Path ${env:ProgramFiles(x86)} 'Microsoft Visual Studio/Installer/vswhere.exe'
    $vsJSON = & $vswhere -products '*' -version '[18.0,19.0)' -requires 'Microsoft.VisualStudio.Component.VC.Tools.x86.x64' -format json -utf8 -nologo
    if ($LASTEXITCODE -ne 0) { throw 'Official VS2026 discovery failed' }
    $installations = @($vsJSON | ConvertFrom-Json | Sort-Object { [version]$_.installationVersion } -Descending)
    if ($installations.Count -eq 0 -or -not $installations[0].isComplete) { throw 'A complete VS2026 installation is required' }
    $selected = $installations[0]
    & (Join-Path $selected.installationPath 'Common7/Tools/Launch-VsDevShell.ps1') -Arch amd64 -HostArch amd64 -SkipAutomaticLocation *> (Join-Path $Work 'developer-shell.raw.log')
    if (-not $env:VCINSTALLDIR -or ([version]$env:VSCMD_VER).Major -ne 18 -or -not [IO.Path]::GetFullPath((Join-Path $env:VCINSTALLDIR '..')).Equals([IO.Path]::GetFullPath($selected.installationPath), [StringComparison]::OrdinalIgnoreCase)) { throw 'The actual DevShell differs from the selected VS' }
    if (-not (Test-Path -LiteralPath (Join-Path $selected.installationPath 'MSBuild/Current/Bin/MSBuild.exe') -PathType Leaf) -or -not $env:WindowsSdkDir -or -not $env:WindowsSDKVersion -or -not (Test-Path -LiteralPath (Join-Path $env:WindowsSdkDir ('Include/' + $env:WindowsSDKVersion)) -PathType Container)) { throw 'The actual MSBuild or SDK is unavailable' }
    $compiler = (Get-Command cl -CommandType Application | Select-Object -First 1).Source
    if (-not [IO.Path]::GetFullPath($compiler).StartsWith(([IO.Path]::GetFullPath($selected.installationPath).TrimEnd('\') + '\VC\Tools\MSVC\'), [StringComparison]::OrdinalIgnoreCase) -or $compiler -notmatch '\\bin\\Hostx64\\x64\\cl\.exe$') { throw 'The actual x64 compiler differs' }
    $env:PATH = (Split-Path -Parent $Node) + ';' + $env:PATH
    $env:npm_config_msvs_version = '2026'
    $env:npm_config_devdir = Join-Path $Work 'gyp-cache'
    $effective = & $Node -e 'const g=require("node:module").createRequire(process.argv[1])("node-gyp/lib/node-gyp.js")();g.parseArgv([process.execPath,process.argv[2],"rebuild"]);console.log(JSON.stringify({msvs_version:g.opts["msvs-version"]}))' (Join-Path $toolRuntime 'package.json') $gyp
    if ($LASTEXITCODE -ne 0 -or ($effective | ConvertFrom-Json).msvs_version -cne '2026') { throw 'The effective original node-gyp selection conflicts' }
    $configuration = [ordered]@{ name = '@vscode/windows-process-tree'; index = 1; runtime = $Runtime; work = $Work; node = $Node; nodeDir = $nodeDir; python = $Python; gyp = $gyp } | ConvertTo-Json -Compress
    Invoke-Original 'single-diagnostic-process-tree-build' $Node @('--input-type=module', '-e', 'import{pathToFileURL}from"node:url";const m=await import(pathToFileURL(process.argv[1]));await m.buildNativePackage({...JSON.parse(process.argv[2]),environment:{...process.env}})', (Join-Path $SourceRoot 'scripts/build-swa-native.mjs'), $configuration)
    $preparation = [ordered]@{ status = 'PASS'; native_component = '@vscode/windows-process-tree'; native_component_index = 1; diagnostic_only = $true; other_native_components_built = $false; original_inputs_modified = $false; node_version = $pin.version; node_sha256 = $pin.asset.executableSha256; headers_sha256 = $pin.headers.sha256; import_library_sha256 = $pin.library.sha256; python_version = $pythonInfo.version; compiler_year = '2026'; actual_vs_version = $selected.installationVersion; actual_cl_version = [Diagnostics.FileVersionInfo]::GetVersionInfo($compiler).FileVersion }
    $stream = [IO.FileStream]::new($OutputFile, [IO.FileMode]::CreateNew, [IO.FileAccess]::Write, [IO.FileShare]::None)
    try {
        $bytes = [Text.UTF8Encoding]::new($false).GetBytes(($preparation | ConvertTo-Json -Depth 6))
        $stream.Write($bytes, 0, $bytes.Length)
    } finally { $stream.Dispose() }
    exit 0
} catch {
    [Console]::Error.WriteLine('native-diagnostic phase=' + $phase + '; kind=' + $_.Exception.GetType().Name)
    exit 1
}
