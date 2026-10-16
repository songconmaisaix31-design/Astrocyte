param(
    [Parameter(Mandatory = $true)][string]$Output,
    [Parameter(Mandatory = $true)][string]$Downloads,
    [string]$DirectMLPackage = ''
)

. (Join-Path $PSScriptRoot 'Release-Common.ps1')
Add-Type -AssemblyName System.IO.Compression.FileSystem
Assert-ReleaseAssetPath $Output
Assert-ReleaseAssetPath $Downloads
Assert-ReleaseSpace
New-Item -ItemType Directory -Path $Output, $Downloads -Force | Out-Null
function Get-Legal([string]$URL, [string]$Destination) {
    if (Test-Path -LiteralPath $Destination) { throw 'License download destination must be new.' }
    & curl.exe --silent --show-error --fail --location --proto '=https' --proto-redir '=https' --connect-timeout 30 --max-time 600 --output $Destination $URL
    if ($LASTEXITCODE -ne 0) { throw ('Fixed official license download failed: ' + $URL) }
    Assert-ReleaseSpace
}
Get-Legal 'https://raw.githubusercontent.com/microsoft/onnxruntime/v1.30.0/LICENSE' (Join-Path $Output 'onnxruntime-LICENSE')
Get-Legal 'https://raw.githubusercontent.com/microsoft/onnxruntime/v1.30.0/ThirdPartyNotices.txt' (Join-Path $Output 'onnxruntime-ThirdPartyNotices.txt')
Get-Legal 'https://raw.githubusercontent.com/microsoft/DirectXShaderCompiler/b4711839eb9a87da7c3436d9b212e0492359fbbd/LICENSE.TXT' (Join-Path $Output 'DirectXShaderCompiler-LICENSE.txt')
$nupkg = $DirectMLPackage
if ($nupkg -eq '') {
    $nupkg = Join-Path $Downloads 'microsoft.ai.directml.1.15.4.nupkg'
    Get-Legal 'https://api.nuget.org/v3-flatcontainer/microsoft.ai.directml/1.15.4/microsoft.ai.directml.1.15.4.nupkg' $nupkg
} else { Assert-ReleaseAssetPath $nupkg }
$zip = [IO.Compression.ZipFile]::OpenRead($nupkg)
try {
    $nuspec = $zip.GetEntry('Microsoft.AI.DirectML.nuspec')
    if ($null -eq $nuspec) { throw 'Official DirectML package metadata missing.' }
    $reader = [IO.StreamReader]::new($nuspec.Open())
    try { [xml]$metadata = $reader.ReadToEnd() } finally { $reader.Dispose() }
    if ($metadata.package.metadata.id -ne 'Microsoft.AI.DirectML' -or $metadata.package.metadata.version -ne '1.15.4' -or $metadata.package.metadata.license.type -ne 'file') { throw 'Official DirectML package/license identity differs.' }
    $licenseName = [string]$metadata.package.metadata.license.'#text'
    if ([IO.Path]::GetFileName($licenseName) -ne $licenseName) { throw 'DirectML license member must be a plain file.' }
    $entry = $zip.GetEntry($licenseName)
    if ($null -eq $entry) { throw 'DirectML redistributable license file missing.' }
    [IO.Compression.ZipFileExtensions]::ExtractToFile($entry, (Join-Path $Output 'DirectML-LICENSE.txt'))
    foreach ($name in @('LICENSE-CODE.txt', 'ThirdPartyNotices.txt')) {
        $notice = $zip.GetEntry($name)
        if ($null -eq $notice) { throw ('Fixed DirectML notice missing: ' + $name) }
        [IO.Compression.ZipFileExtensions]::ExtractToFile($notice, (Join-Path $Output ('DirectML-' + $name)))
    }
} finally { $zip.Dispose() }
Assert-ReleaseSpace
Write-Output 'PASS: pinned ONNX, DirectML 1.15.4 and DXC b4711839 full license files prepared.'
