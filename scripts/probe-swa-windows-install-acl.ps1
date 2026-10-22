#Requires -Version 7.2
[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $false
$env:SWA_ACL_PROBE_PHASE = 'preflight'
if (-not $IsWindows -or [Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -ne 'X64') { throw 'This local observation requires actual Windows x64' }
$revision = 'c7ab4a7479182f29f990cd649accac193514d169'
$checkout = (Get-Item -LiteralPath $env:GITHUB_WORKSPACE).FullName
Set-Location -LiteralPath $checkout
$developmentNode = (Get-Command node -CommandType Application | Select-Object -First 1).Source
if ((& $developmentNode --version).Trim() -cne 'v24.16.0') { throw 'The pinned development Node differs' }
$evidence = Join-Path $checkout '.verification/swa-ci/install-acl-probe'
$work = Join-Path $env:RUNNER_TEMP ('swa-install-acl-probe-' + $env:GITHUB_SHA)
foreach ($directory in @($evidence, $work)) {
    if (-not [IO.Path]::IsPathFullyQualified($directory) -or (Test-Path -LiteralPath $directory)) { throw 'A new absolute local observation directory is required' }
    New-Item -ItemType Directory -Path $directory | Out-Null
}
$env:SWA_ACL_PROBE_PHASE = 'installer-source'
$resolved = (& git rev-parse ($revision + '^{commit}')).Trim()
if ($LASTEXITCODE -ne 0 -or $resolved -cne $revision) { throw 'The fixed failed installer commit is unavailable' }
$sourceLines = & git show ($revision + ':scripts/install-swa.ps1')
if ($LASTEXITCODE -ne 0) { throw 'The original failed installer source is unavailable' }
$source = $sourceLines -join "`n"
$tokens = $null
$parseErrors = $null
$ast = [Management.Automation.Language.Parser]::ParseInput($source, [ref]$tokens, [ref]$parseErrors)
if ($parseErrors.Count -ne 0) { throw 'The original installer source did not parse' }
$expected = [ordered]@{
    create = 'New-Item -ItemType Directory -Path $destination | Out-Null'
    identity = '$sid = [System.Security.Principal.WindowsIdentity]::GetCurrent().User.Value'
    acl = '& (Join-Path $env:SystemRoot ''System32/icacls.exe'') $destination ''/inheritance:r'' ''/grant:r'' "*$($sid):(OI)(CI)F"'
    acl_exit = 'if ($LASTEXITCODE -ne 0) { throw ''安装根私有 ACL 设置失败'' }'
    copy_source = '$copy = "require(''node:fs'').cpSync(process.argv[1],process.argv[2],{recursive:true,dereference:true,errorOnExist:true,force:false})"'
    copy = '& $bundledNode --max-old-space-size=1024 -e $copy $package $destination'
    copy_exit = 'if ($LASTEXITCODE -ne 0) { throw ''复制安装闭包失败，保留实际目标供检查'' }'
}
$operations = @{}
foreach ($name in $expected.Keys) {
    $selectedOperations = @($ast.EndBlock.Statements | Where-Object { $_.Extent.Text -ceq $expected[$name] })
    if ($selectedOperations.Count -ne 1) { throw 'The fixed installer operation has an unexpected AST shape' }
    $operations[$name] = [scriptblock]::Create($selectedOperations[0].Extent.Text)
}
$env:SWA_ACL_PROBE_PHASE = 'node-pin'
$assetJSON = & $developmentNode --input-type=module -e 'import{NODE_RUNTIME_ASSETS,nodeRuntimeReleaseUrl}from"./third_party/orca-v1.4.230/src/shared/node-runtime-pin.ts";let a=NODE_RUNTIME_ASSETS["win32-x64"];console.log(JSON.stringify({...a,url:nodeRuntimeReleaseUrl("official",a.archive)}));'
if ($LASTEXITCODE -ne 0) { throw 'The original Node asset pin is unavailable' }
$asset = $assetJSON | ConvertFrom-Json
$buildRoot = Join-Path $checkout '.verification/swa-cli-headless-20261020/build'
New-Item -ItemType Directory -Path $buildRoot -Force | Out-Null
$downloadWork = Join-Path $buildRoot ('candidate-' + [Convert]::ToHexString([Security.Cryptography.RandomNumberGenerator]::GetBytes(3)))
if (Test-Path -LiteralPath $downloadWork) { throw 'The exclusive download directory already exists' }
New-Item -ItemType Directory -Path $downloadWork | Out-Null
$env:SWA_ACL_PROBE_PHASE = 'node-download'
$archive = Join-Path $downloadWork $asset.archive
& (Join-Path $PSHOME 'pwsh.exe') -NoProfile -NonInteractive -File (Join-Path $checkout 'scripts/download-swa-node.ps1') -Uri $asset.url -WorkDirectory $downloadWork -OutputFile $archive *> (Join-Path $evidence 'node-download.log')
if ($LASTEXITCODE -ne 0) { throw 'The original bounded pinned Node download failed' }
$env:SWA_ACL_PROBE_PHASE = 'node-verification'
if ((Get-FileHash -LiteralPath $archive -Algorithm SHA256).Hash.ToLowerInvariant() -cne $asset.archiveSha256) { throw 'The fixed Node archive checksum differs' }
$nodeRoot = Join-Path $downloadWork 'node'
[IO.Compression.ZipFile]::ExtractToDirectory($archive, $nodeRoot, $false)
$bundledNode = Join-Path $nodeRoot 'node-v24.21.0-win-x64/node.exe'
if ((Get-Item -LiteralPath $bundledNode).Length -ne $asset.executableSize -or (Get-FileHash -LiteralPath $bundledNode -Algorithm SHA256).Hash.ToLowerInvariant() -cne $asset.executableSha256 -or (& $bundledNode --version).Trim() -cne 'v24.21.0') { throw 'The actual fixed Node executable differs' }
$env:SWA_ACL_PROBE_PHASE = 'ordinary-source'
$package = Join-Path $work 'ordinary-source'
New-Item -ItemType Directory -Path $package | Out-Null
$license = Join-Path $checkout 'LICENSE'
if ((Get-Item -LiteralPath $license).Attributes -band [IO.FileAttributes]::ReparsePoint) { throw 'The public source component is redirected' }
[IO.File]::Copy($license, (Join-Path $package 'LICENSE'), $false)
$licenseHash = (Get-FileHash -LiteralPath $license -Algorithm SHA256).Hash
$destination = Join-Path $work 'ordinary-copy'
$userSID = [Security.Principal.WindowsIdentity]::GetCurrent().User.Value
$observations = [Collections.Generic.List[object]]::new()
function Observe-LocalACL([string]$Stage) {
    $item = Get-Item -LiteralPath $destination -Force
    if (-not $item.PSIsContainer -or ($item.Attributes -band [IO.FileAttributes]::ReparsePoint)) { throw 'The selected observation directory is redirected' }
    $acl = Get-Acl -LiteralPath $destination
    $owner = $acl.GetOwner([Security.Principal.SecurityIdentifier]).Value
    $rules = @($acl.GetAccessRules($true, $true, [Security.Principal.SecurityIdentifier]) | ForEach-Object {
        $trustee = $_.IdentityReference.Value
        $kind = if ($trustee -ceq $userSID) { 'current-user' } elseif ($trustee -ceq 'S-1-5-18') { 'system' } elseif ($trustee -ceq 'S-1-5-32-544') { 'administrators' } else { 'other' }
        [ordered]@{ trustee_kind = $kind; access_type = $_.AccessControlType.ToString(); inherited = $_.IsInherited; full_control = $_.FileSystemRights -eq [Security.AccessControl.FileSystemRights]::FullControl; inheritance = $_.InheritanceFlags.ToString(); propagation = $_.PropagationFlags.ToString() }
    })
    $observation = [ordered]@{ stage = $Stage; access_rules_protected = $acl.AreAccessRulesProtected; owner_matches_current_user = $owner -ceq $userSID; owner_is_administrators = $owner -ceq 'S-1-5-32-544'; owner_is_system = $owner -ceq 'S-1-5-18'; rules = $rules }
    $observations.Add($observation)
    $observations | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath (Join-Path $evidence 'acl-observations.json')
    $message = $observation | ConvertTo-Json -Depth 6 -Compress
    if ($message.Length -gt 8000) { $message = $message.Substring(0, 8000) + ' [bounded observation; original JSON retained]' }
    $escaped = $message.Replace('%', '%25').Replace("`r", '%0D').Replace("`n", '%0A')
    Write-Host ('::warning title=Original installer local ACL observation::' + $escaped)
}
$env:SWA_ACL_PROBE_PHASE = 'original-create'
. $operations['create']
Observe-LocalACL 'after-original-create'
$env:SWA_ACL_PROBE_PHASE = 'original-acl'
. $operations['identity']
. $operations['acl'] *> (Join-Path $evidence 'original-icacls.log')
. $operations['acl_exit']
Observe-LocalACL 'after-original-icacls'
$env:SWA_ACL_PROBE_PHASE = 'original-copy'
. $operations['copy_source']
. $operations['copy'] *> (Join-Path $evidence 'ordinary-copy.log')
. $operations['copy_exit']
Observe-LocalACL 'after-original-copy'
$env:SWA_ACL_PROBE_PHASE = 'copy-verification'
$copied = Get-Item -LiteralPath (Join-Path $destination 'LICENSE') -Force
if ($copied.PSIsContainer -or ($copied.Attributes -band [IO.FileAttributes]::ReparsePoint) -or (Get-FileHash -LiteralPath $copied.FullName -Algorithm SHA256).Hash -cne $licenseHash) { throw 'The ordinary public component copy differs' }
$result = [ordered]@{ diagnostic_only = $true; installer_source_revision = $revision; observation_source_revision = $env:GITHUB_SHA; bundled_node_version = '24.21.0'; original_operations = @($expected.Keys); ordinary_component_bytes_equal = $true; observations = $observations; candidate_built = $false; installation_acceptance = 'NOT_RUN'; service_started = $false; model_calls = 0 }
$result | ConvertTo-Json -Depth 10 | Set-Content -LiteralPath (Join-Path $evidence 'result.json')
Write-Output 'The original narrow ACL operations were observed; no candidate, installation acceptance, native build, service or model was run.'
exit 0
