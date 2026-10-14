param(
    [ValidateRange(1, 65535)]
    [int]$Port = 8787,
    [string]$DataDir = (Join-Path $PSScriptRoot 'data'),
    [string]$SummarizeCli = '',
    [string]$NodeExecutable = '',
    [string]$YtDlp = '',
    [string]$FFmpeg = '',
    [string]$WhisperBinary = '',
    [string]$WhisperModel = '',
    [string]$ListingPython = ''
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$binary = Join-Path $PSScriptRoot 'astrocyte.exe'
$web = Join-Path $PSScriptRoot 'web'
if (-not (Test-Path -LiteralPath $binary -PathType Leaf) -or -not (Test-Path -LiteralPath (Join-Path $web 'index.html') -PathType Leaf)) {
    throw 'Extract the complete Windows ZIP before starting.'
}
if ([Environment]::OSVersion.Platform -ne 'Win32NT' -or -not [Environment]::Is64BitOperatingSystem) { throw 'Windows x64 required.' }
if (($SummarizeCli -eq '') -ne ($NodeExecutable -eq '')) { throw 'SummarizeCli and NodeExecutable must be supplied together.' }
$data = [IO.Path]::GetFullPath($DataDir)
if ($data -eq [IO.Path]::GetPathRoot($data)) { throw 'DataDir must be a dedicated application directory.' }
$saved = @{}
Get-ChildItem Env: | Where-Object { $_.Name -like 'ASTROCYTE_*' } | ForEach-Object { $saved[$_.Name] = $_.Value }
$saved['LOCALAPPDATA'] = $env:LOCALAPPDATA
try {
    # 当前应用只接受启动参数，不继承宿主的正式实验和自动处理配置。
    foreach ($name in @($saved.Keys)) { if ($name -like 'ASTROCYTE_*') { [Environment]::SetEnvironmentVariable($name, $null, 'Process') } }
    $env:LOCALAPPDATA = Join-Path $data 'local'
    $env:ASTROCYTE_DATA_DIR = Join-Path $env:LOCALAPPDATA 'astrocyte-workbench\v0.2-access\app-state'
    $env:ASTROCYTE_WEB_DIR = $web
    $env:ASTROCYTE_PORT = [string]$Port
    $env:ASTROCYTE_WEB_PORT = [string]$Port
    $env:ASTROCYTE_ALLOWED_ORIGINS = 'http://127.0.0.1:' + $Port + ',http://localhost:' + $Port
    $env:ASTROCYTE_RELAY = '1'
    $env:ASTROCYTE_ENABLE_SUMMARIZE = 'false'
    $env:ASTROCYTE_ENABLE_CODEX_DISTILLATION = 'false'
    if ($SummarizeCli -ne '') {
        foreach ($path in @($SummarizeCli, $NodeExecutable)) {
            if (-not [IO.Path]::IsPathRooted($path) -or -not (Test-Path -LiteralPath $path -PathType Leaf)) { throw 'SummarizeCli and NodeExecutable require existing absolute files.' }
        }
        $env:ASTROCYTE_ENABLE_SUMMARIZE = 'true'
        $env:ASTROCYTE_SUMMARIZE_CLI = $SummarizeCli
        $env:ASTROCYTE_NODE = $NodeExecutable
    }
    $optional = @{ ASTROCYTE_YT_DLP_PATH = $YtDlp; ASTROCYTE_FFMPEG_PATH = $FFmpeg; ASTROCYTE_WHISPER_BINARY = $WhisperBinary; ASTROCYTE_WHISPER_MODEL = $WhisperModel; ASTROCYTE_LISTING_PYTHON = $ListingPython }
    foreach ($name in $optional.Keys) {
        $path = $optional[$name]
        if ($path -ne '') {
            if (-not [IO.Path]::IsPathRooted($path) -or -not (Test-Path -LiteralPath $path -PathType Leaf)) { throw ($name + ' requires an existing absolute file.') }
            [Environment]::SetEnvironmentVariable($name, $path, 'Process')
        }
    }
    New-Item -ItemType Directory -Path $env:LOCALAPPDATA -Force | Out-Null
    Write-Host ('Astrocyte: http://127.0.0.1:' + $Port)
    Write-Host ('SQLite: ' + (Join-Path $env:ASTROCYTE_DATA_DIR 'state.sqlite'))
    Write-Host 'Keep this console open. Press Ctrl+C to stop.'
    Push-Location $PSScriptRoot
    try {
        & $binary
        if ($LASTEXITCODE -ne 0) { throw ('Astrocyte exited with code ' + $LASTEXITCODE) }
    } finally { Pop-Location }
} finally {
    Get-ChildItem Env: | Where-Object { $_.Name -like 'ASTROCYTE_*' } | ForEach-Object { [Environment]::SetEnvironmentVariable($_.Name, $null, 'Process') }
    foreach ($name in $saved.Keys) { [Environment]::SetEnvironmentVariable($name, $saved[$name], 'Process') }
}
