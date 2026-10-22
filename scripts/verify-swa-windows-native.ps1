#Requires -Version 7.2
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$PnpmCli,
    [Parameter(Mandatory)][string]$Python,
    [ValidateSet('2022', '2026')][string]$MsvsVersion = '2026'
)

$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $false
$env:SWA_CI_PHASE = 'preflight-platform'
if (-not $IsWindows -or [System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -ne 'X64') { throw 'Native Windows x64 is required' }
$inputs = [ordered]@{ pnpm = $PnpmCli; python = $Python; checkout = $env:GITHUB_WORKSPACE; temporary = $env:RUNNER_TEMP; program_files = $env:ProgramFiles }
foreach ($role in $inputs.Keys) {
    $env:SWA_CI_PHASE = 'preflight-input-' + $role
    $path = $inputs[$role]
    if (-not [IO.Path]::IsPathFullyQualified($path) -or -not (Test-Path -LiteralPath $path)) { throw ('An actual absolute CI input is missing: ' + $role) }
}
$env:SWA_CI_PHASE = 'preflight-source'
$checkout = (Get-Item -LiteralPath $env:GITHUB_WORKSPACE).FullName
Set-Location -LiteralPath $checkout
$revision = (& git rev-parse HEAD).Trim()
if ($LASTEXITCODE -ne 0 -or $revision -cnotmatch '^[0-9a-f]{40}$' -or $revision -cne $env:GITHUB_SHA) { throw 'The exact CI source revision differs' }
$sourceStatus = & git status --porcelain --untracked-files=normal
if ($LASTEXITCODE -ne 0 -or $sourceStatus) { throw 'The candidate source is not clean' }
$evidence = Join-Path $checkout '.verification/swa-ci/windows'
$logs = Join-Path $evidence 'logs'
if (Test-Path -LiteralPath $evidence) { throw 'This evidence directory already exists' }
New-Item -ItemType Directory -Path $logs | Out-Null
Set-Content -LiteralPath (Join-Path $evidence 'SOURCE') -Value $revision
$node = (Get-Command node -CommandType Application | Select-Object -First 1).Source
$go = (Get-Command go -CommandType Application | Select-Object -First 1).Source
$powershell = Join-Path $env:ProgramFiles 'PowerShell/7/pwsh.exe'
$env:SWA_PNPM_CLI = (Get-Item -LiteralPath $PnpmCli).FullName
$env:SWA_BUILD_PYTHON = (Get-Item -LiteralPath $Python).FullName
$env:ASTROCYTE_GO = $go
$env:GOTOOLCHAIN = 'local'
$env:SWA_CI_PHASE = 'preflight-toolchain'
$result = [ordered]@{ source_revision = $revision; platform = 'win32-x64'; edition = 'cli'; status = 'NOT_RUN'; model_calls = 0; agent_acceptance = 'NOT_RUN' }

function Invoke-Logged([string]$Name, [string]$Program, [string[]]$Arguments) {
    $env:SWA_CI_PHASE = $Name
    $log = Join-Path $logs ($Name + '.log')
    & $Program @Arguments *> $log
    $code = $LASTEXITCODE
    Set-Content -LiteralPath (Join-Path $logs ($Name + '.exit.txt')) -Value $code
    if ($code -ne 0) {
        Get-Content -LiteralPath $log -Tail 60
        throw "$Name exited $code; its original output and partial files are retained"
    }
    Write-Output "$Name exit=0"
}

try {
    if ((& $node --version).Trim() -cne 'v24.16.0' -or (& $node $PnpmCli --version).Trim() -cne '11.27.0') { throw 'The development Node or pnpm version differs' }
    if ((& $go version).Trim() -cne 'go version go1.27.2 windows/amd64') { throw 'The Go version differs' }
    Invoke-Logged 'frozen-install' $node @($PnpmCli, '--filter', 'astrocyte', '--filter', '@astrocyte/web', '--filter', '@swa/orca-runtime', '--filter', '@orca/process-host', 'install', '--frozen-lockfile')
    Invoke-Logged 'go-download' $go @('mod', 'download')
    Invoke-Logged 'go-verify' $go @('mod', 'verify')
    $env:SWA_CI_PHASE = 'msvc-toolchain'
    $major = if ($MsvsVersion -ceq '2026') { 18 } else { 17 }
    $toolset = if ($MsvsVersion -ceq '2026') { 'v145' } else { 'v143' }
    $vswhere = Join-Path ${env:ProgramFiles(x86)} 'Microsoft Visual Studio/Installer/vswhere.exe'
    $metadataJSON = & $vswhere -products '*' -version ('[' + $major + '.0,' + ($major + 1) + '.0)') -requires 'Microsoft.VisualStudio.Component.VC.Tools.x86.x64' -format json -utf8 -nologo
    if ($LASTEXITCODE -ne 0) { throw 'The official selected Visual Studio discovery failed' }
    $installations = @($metadataJSON | ConvertFrom-Json | Sort-Object { [version]$_.installationVersion } -Descending)
    if ($installations.Count -eq 0) { throw 'The explicitly selected Visual Studio x64 toolchain is not installed' }
    $selected = $installations[0]
    if (-not $selected.isComplete -or ([version]$selected.installationVersion).Major -ne $major) { throw 'The selected Visual Studio installation is incomplete or has the wrong version' }
    & (Join-Path $selected.installationPath 'Common7/Tools/Launch-VsDevShell.ps1') -Arch amd64 -HostArch amd64 -SkipAutomaticLocation *> (Join-Path $logs 'toolchain-devshell.log')
    if (-not $env:VCINSTALLDIR -or -not $env:VSCMD_VER -or ([version]$env:VSCMD_VER).Major -ne $major) { throw 'The official developer environment did not establish the selected compiler version' }
    $actualVS = [IO.Path]::GetFullPath((Join-Path $env:VCINSTALLDIR '..'))
    if (-not $actualVS.Equals([IO.Path]::GetFullPath($selected.installationPath), [StringComparison]::OrdinalIgnoreCase)) { throw 'The developer environment belongs to a different Visual Studio installation' }
    $msbuild = Join-Path $selected.installationPath 'MSBuild/Current/Bin/MSBuild.exe'
    if (-not (Test-Path -LiteralPath $msbuild -PathType Leaf) -or -not $env:WindowsSdkDir -or -not $env:WindowsSDKVersion -or -not (Test-Path -LiteralPath (Join-Path $env:WindowsSdkDir ('Include/' + $env:WindowsSDKVersion)) -PathType Container)) { throw 'The selected actual MSBuild or Windows SDK is missing' }
    $compiler = (Get-Command cl -CommandType Application | Select-Object -First 1).Source
    if (-not [IO.Path]::GetFullPath($compiler).StartsWith(([IO.Path]::GetFullPath($selected.installationPath).TrimEnd('\') + '\VC\Tools\MSVC\'), [StringComparison]::OrdinalIgnoreCase) -or $compiler -notmatch '\\bin\\Hostx64\\x64\\cl\.exe$') { throw 'The actual x64 compiler is outside the selected installation' }
    $compilerVersion = [Diagnostics.FileVersionInfo]::GetVersionInfo($compiler)
    $env:npm_config_msvs_version = $MsvsVersion
    $gypOptions = & $node -e 'const{createRequire}=require("node:module");const r=createRequire(process.cwd()+"/runtime/swa-orca/package.json");const gyp=r("node-gyp/lib/node-gyp.js")();gyp.parseArgv([process.execPath,r.resolve("node-gyp/bin/node-gyp.js"),"rebuild"]);console.log(JSON.stringify({msvs_version:gyp.opts["msvs-version"]}))'
    if ($LASTEXITCODE -ne 0 -or ($gypOptions | ConvertFrom-Json).msvs_version -cne $MsvsVersion) { throw 'The effective node-gyp compiler selection conflicts with the explicit CI selection' }
    [ordered]@{ selected_msvs_version = $MsvsVersion; actual_vs_path = $actualVS; actual_vs_version = $selected.installationVersion; developer_environment_version = $env:VSCMD_VER; expected_gyp_toolset = $toolset; actual_msbuild = $msbuild; actual_sdk_version = $env:WindowsSDKVersion; actual_cl_path = $compiler; actual_cl_file_version = $compilerVersion.FileVersion; actual_cl_product_version = $compilerVersion.ProductVersion; effective_gyp_msvs_version = ($gypOptions | ConvertFrom-Json).msvs_version; native_compilation = 'NOT_RUN' } | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $logs 'toolchain-preflight.log')
    $name = 'swa-windows-cli-' + $revision
    $candidate = Join-Path $checkout ('dist/' + $name)
    Invoke-Logged 'build' $node @($PnpmCli, 'build:swa', '--output', $candidate, '--edition', 'cli', '--python', $Python, '--msvs-version', $MsvsVersion)
    $artifacts = Join-Path $env:RUNNER_TEMP 'swa-artifacts'
    $packages = Join-Path $env:RUNNER_TEMP 'swa-packages'
    $installs = Join-Path $env:RUNNER_TEMP 'swa-installs'
    foreach ($path in @($artifacts, $packages, $installs)) {
        if (Test-Path -LiteralPath $path) { throw 'A candidate output parent already exists' }
        New-Item -ItemType Directory -Path $path | Out-Null
    }
    $env:SWA_CI_PHASE = 'archive'
    $archive = Join-Path $artifacts ($name + '.zip')
    [IO.Compression.ZipFile]::CreateFromDirectory($candidate, $archive, [IO.Compression.CompressionLevel]::Optimal, $true)
    $hash = (Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant()
    Set-Content -LiteralPath ($archive + '.sha256') -Value ($hash + '  ' + [IO.Path]::GetFileName($archive))
    if ((Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant() -cne $hash) { throw 'The archive checksum differs' }
    $env:SWA_CI_PHASE = 'extract'
    [IO.Compression.ZipFile]::ExtractToDirectory($archive, $packages, $false)
    Set-Content -LiteralPath (Join-Path $logs 'archive-check.log') -Value ('ZIP SHA256=' + $hash + '; ordinary .NET extraction completed without overwrite')
    $package = Join-Path $packages $name
    $install = Join-Path $installs $name
    $checker = Join-Path $package 'scripts/check-swa-runtime.mjs'
    Invoke-Logged 'extracted-check' (Join-Path $package 'node/node.exe') @($checker, '--candidate', $package)
    $env:SWA_CI_PHASE = 'install'
    try {
        & (Join-Path $package 'install-swa.ps1') -PackageRoot $package -InstallRoot $install -ProtectedRoots @($checkout, $env:ProgramFiles) *> (Join-Path $logs 'install.log')
        Set-Content -LiteralPath (Join-Path $logs 'install.exit.txt') -Value 0
    } catch {
        Set-Content -LiteralPath (Join-Path $logs 'install.exit.txt') -Value 1
        throw
    }
    Invoke-Logged 'installed-check' (Join-Path $install 'node/node.exe') @((Join-Path $install 'scripts/check-swa-runtime.mjs'), '--candidate', $install)
    $actual = (& (Join-Path $install 'swa.exe') --version).Trim()
    if ($LASTEXITCODE -ne 0 -or $actual -cne $revision -or (Get-Content -LiteralPath (Join-Path $install 'VERSION') -Raw).Trim() -cne $revision) { throw 'The installed source version differs' }
    $acl = Get-Acl -LiteralPath $install
    $env:SWA_CI_PHASE = 'installed-acl'
    $sid = [System.Security.Principal.WindowsIdentity]::GetCurrent().User.Value
    if (-not $acl.AreAccessRulesProtected -or $acl.GetOwner([System.Security.Principal.SecurityIdentifier]).Value -ne $sid) { throw 'The actual installation root is not private to its owner' }
    $seenTrustees = [Collections.Generic.HashSet[string]]::new([StringComparer]::Ordinal)
    $expectedInheritance = [Security.AccessControl.InheritanceFlags]::ContainerInherit -bor [Security.AccessControl.InheritanceFlags]::ObjectInherit
    foreach ($rule in $acl.GetAccessRules($true, $true, [Security.Principal.SecurityIdentifier])) {
        $trustee = $rule.IdentityReference.Value
        if ($rule.IsInherited -or $rule.AccessControlType -ne [Security.AccessControl.AccessControlType]::Allow -or $trustee -notin @($sid, 'S-1-5-18') -or -not $seenTrustees.Add($trustee) -or $rule.FileSystemRights -ne [Security.AccessControl.FileSystemRights]::FullControl -or $rule.InheritanceFlags -ne $expectedInheritance -or $rule.PropagationFlags -ne [Security.AccessControl.PropagationFlags]::None) { throw 'The installation root has unexpected access rules' }
    }
    if (-not $seenTrustees.Contains($sid)) { throw 'The installation root lacks the exact owner access rule' }
    $env:SWA_CI_PHASE = 'service-smoke'
    $private = Join-Path $env:RUNNER_TEMP ('swa-data-' + $revision)
    if (Test-Path -LiteralPath $private) { throw 'The private CI data parent already exists' }
    New-Item -ItemType Directory -Path $private | Out-Null
    $privateHome = Join-Path $private 'discovery-home'
    $roaming = Join-Path $private 'appdata'
    $local = Join-Path $private 'localappdata'
    $temporary = Join-Path $private 'temp'
    $dataParent = Join-Path $private 'product'
    foreach ($path in @($privateHome, $roaming, $local, $temporary, $dataParent)) { New-Item -ItemType Directory -Path $path | Out-Null }
    $start = [Diagnostics.ProcessStartInfo]::new()
    $start.FileName = Join-Path $install 'node/node.exe'
    $start.WorkingDirectory = $install
    $start.UseShellExecute = $false
    $start.RedirectStandardOutput = $true
    $start.RedirectStandardError = $true
    $start.Environment.Clear()
    foreach ($key in @('SystemRoot', 'WINDIR', 'ComSpec', 'ProgramFiles', 'ProgramFiles(x86)', 'ProgramW6432', 'PATHEXT')) {
        $value = [Environment]::GetEnvironmentVariable($key)
        if ($value) { $start.Environment[$key] = $value }
    }
    $git = (Get-Command git -CommandType Application | Select-Object -First 1).Source
    $start.Environment['PATH'] = ((Join-Path $install 'node'), (Join-Path $env:SystemRoot 'System32'), (Split-Path -Parent $powershell), (Split-Path -Parent $git)) -join ';'
    $start.Environment['USERPROFILE'] = $privateHome
    $start.Environment['HOME'] = $privateHome
    $start.Environment['APPDATA'] = $roaming
    $start.Environment['LOCALAPPDATA'] = $local
    $start.Environment['TEMP'] = $temporary
    $start.Environment['TMP'] = $temporary
    foreach ($argument in @('--max-old-space-size=1024', (Join-Path $checkout 'tests/swa-cli-headless/serve-lifecycle.mjs'), '--candidate', $install, '--source-revision', $revision, '--evidence', (Join-Path $evidence 'smoke'), '--data-parent', $dataParent, '--protected-roots', ($checkout + ';' + $env:ProgramFiles), '--port', '18793', '--mode', 'readiness-eof')) { $start.ArgumentList.Add($argument) }
    $process = [Diagnostics.Process]::new()
    $process.StartInfo = $start
    if (-not $process.Start()) { throw 'The actual installed service check did not start' }
    $output = $process.StandardOutput.ReadToEndAsync()
    $errors = $process.StandardError.ReadToEndAsync()
    if (-not $process.WaitForExit(240000)) { throw 'The service check is unknown after its deadline; no whole-tree exit is claimed' }
    Set-Content -LiteralPath (Join-Path $logs 'service-smoke.log') -Value (($output.GetAwaiter().GetResult()) + ($errors.GetAwaiter().GetResult()))
    Set-Content -LiteralPath (Join-Path $logs 'service-smoke.exit.txt') -Value $process.ExitCode
    if ($process.ExitCode -ne 0) { throw 'The actual service/Job shutdown check failed or remains unknown' }
    $env:SWA_CI_PHASE = 'final-source'
    $finalRevision = (& git rev-parse HEAD).Trim()
    if ($LASTEXITCODE -ne 0 -or $finalRevision -cne $revision) { throw 'The source version changed during verification' }
    $finalStatus = & git status --porcelain --untracked-files=normal
    if ($LASTEXITCODE -ne 0 -or $finalStatus) { throw 'The source changed during verification' }
    $result.status = 'PASS'
    $result.archive_sha256 = $hash
    $result.installed_version = $actual
    $result.service = 'PASS'
    $result.whole_service_job_shutdown = 'PASS'
} catch {
    $result.status = 'FAIL'
    $result.failed_stage = $env:SWA_CI_PHASE
    $result.failure = $_.Exception.Message
    throw
} finally {
    $result | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath (Join-Path $evidence 'result.json')
}
