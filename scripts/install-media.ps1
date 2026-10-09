param(
    [string]$PythonExecutable,
    [string]$MediaRoot = $env:ASTROCYTE_MEDIA_DIR
)
$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'
if (-not $IsWindows -or [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture -ne 'X64') {
    throw 'This installer supports Windows x64. On other platforms configure the documented media executable paths.'
}
if (-not $MediaRoot) { $MediaRoot = Join-Path $env:LOCALAPPDATA 'Astrocyte/media' }
if (-not [System.IO.Path]::IsPathFullyQualified($MediaRoot)) { throw 'MediaRoot must be absolute' }
$MediaRoot = [System.IO.Path]::GetFullPath($MediaRoot)
New-Item -ItemType Directory -Path $MediaRoot -Force | Out-Null
$pins = (Get-Content -LiteralPath (Join-Path $PSScriptRoot '../dependencies.lock.json') -Raw | ConvertFrom-Json).external_media
function MediaPath([string]$Relative) {
    $target = [System.IO.Path]::GetFullPath((Join-Path $MediaRoot $Relative))
    if (-not $target.StartsWith($MediaRoot.TrimEnd('\', '/') + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "Media path escapes installation directory: $Relative"
    }
    return $target
}
function RunTool([string]$Executable, [string[]]$Arguments) {
    & $Executable @Arguments
    if ($LASTEXITCODE -ne 0) { throw "Media tool failed ($LASTEXITCODE): $Executable" }
}
function Download([string]$URL, [string]$Target) {
    if (Test-Path -LiteralPath $Target -PathType Leaf) { return }
    New-Item -ItemType Directory -Path (Split-Path -Parent $Target) -Force | Out-Null
    Invoke-WebRequest -Uri $URL -OutFile ($Target + '.partial')
    Move-Item -LiteralPath ($Target + '.partial') -Destination $Target
}
if (-not $PythonExecutable) { $PythonExecutable = (Get-Command python -CommandType Application -ErrorAction Stop).Source }
if (-not [System.IO.Path]::IsPathFullyQualified($PythonExecutable) -or -not (Test-Path -LiteralPath $PythonExecutable -PathType Leaf)) { throw 'PythonExecutable must be an installed absolute executable path' }
RunTool $PythonExecutable @('-c', 'import sys; assert sys.version_info >= (3,10), "yt-dlp needs Python >=3.10"; print(sys.executable); print(sys.version)')
$ytRoot = MediaPath $pins.yt_dlp.directory
$venvPython = Join-Path $ytRoot 'Scripts/python.exe'
if (-not (Test-Path -LiteralPath $venvPython -PathType Leaf)) { RunTool $PythonExecutable @('-m', 'venv', $ytRoot) }
RunTool $venvPython @('-m', 'pip', '--isolated', 'install', '--index-url', 'https://pypi.org/simple', '--disable-pip-version-check', ('yt-dlp==' + $pins.yt_dlp.pip_version))
RunTool $venvPython @('-m', 'pip', '--isolated', 'check')
$ytExecutable = MediaPath ($pins.yt_dlp.directory + '/' + $pins.yt_dlp.executable)
$ytVersion = (RunTool $ytExecutable @('--version') | Out-String).Trim()
if ($ytVersion -ne $pins.yt_dlp.version) { throw "yt-dlp version mismatch: $ytVersion" }
foreach ($item in @($pins.ffmpeg, $pins.whisper)) {
    $destination = MediaPath $item.directory
    $archive = MediaPath ($item.directory + '.zip')
    Download $item.url $archive
    if (-not (Test-Path -LiteralPath (Join-Path $destination $item.executable) -PathType Leaf)) {
        Expand-Archive -LiteralPath $archive -DestinationPath $destination -Force
    }
}
$ffmpeg = MediaPath ($pins.ffmpeg.directory + '/' + $pins.ffmpeg.executable)
$ffprobe = MediaPath ($pins.ffmpeg.directory + '/' + $pins.ffmpeg.probe)
foreach ($executable in @($ffmpeg, $ffprobe)) {
    $version = RunTool $executable @('-version')
    if (($version | Out-String) -notmatch ('version ' + [regex]::Escape($pins.ffmpeg.version))) { throw "FFmpeg version mismatch: $executable" }
    $version | Select-Object -First 1
}
$whisper = MediaPath ($pins.whisper.directory + '/' + $pins.whisper.executable)
# whisper.cpp prints help to stderr; successful native loading is checked here.
$whisperHelp = RunTool $whisper @('--help') 2>&1
$whisperHelp | Select-Object -First 4
$model = MediaPath $pins.model.file
Download $pins.model.url $model
if ((Get-Item -LiteralPath $model).Length -ne $pins.model.size_bytes) { throw 'Whisper model size mismatch; remove only this model file and rerun installation' }
Write-Output "Installed pinned media dependencies in $MediaRoot"
Write-Output "yt-dlp $ytVersion; whisper.cpp release $($pins.whisper.version); multilingual model $model"
Write-Output 'Installation/native loading is not proof that a selected video can be downloaded or transcribed.'
