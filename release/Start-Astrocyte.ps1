param(
    [ValidateRange(1, 65535)]
    [int]$Port = 8787,
    [string]$DataDir = '',
    [string]$SummarizeCli = '',
    [string]$NodeExecutable = '',
    [string]$YtDlp = '',
    [string]$FFmpeg = '',
    [string]$WhisperBinary = '',
    [string]$WhisperModel = '',
    [string]$ListingPython = '',
    [string]$UVXExecutable = '',
    [ValidateSet('archive', 'native_notifications')]
    [string]$RelaySessionSource = 'archive',
    [ValidateRange(1, 86400)]
    [Nullable[int]]$RelayLiveSeconds = $null,
    # S4 CodeGraph/Repository context — all three must be supplied together, or none.
    [string]$CodeGraphNode = '',
    [string]$CodeGraphSDK = '',
    [string]$CodeGraphScript = '',
    [string]$RepositoryEnvironmentID = '',
    [string]$AOCIExecutable = '',
    # S5 Laya judgment — all required together when any is supplied.
    [string]$LayaNode = '',
    [string]$LayaWorker = '',
    [string]$LayaModelDir = '',
    [string]$LayaAssets = '',
    [string]$LayaSDKEntry = '',
    [string]$LayaRPCEntry = '',
    [string]$LayaLanguage = '',
    [string]$LayaStderrDir = '',
    [Nullable[int]]$LayaStartupTimeoutMS = $null,
    [Nullable[int]]$LayaShutdownTimeoutMS = $null,
    # Jev cloud REST — only the key, no local model.
    [string]$TypesafeApiKey = ''
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
# S4 CodeGraph/Repository — incomplete group fails immediately.
$codeGraphAny = $CodeGraphNode -ne '' -or $CodeGraphSDK -ne '' -or $CodeGraphScript -ne ''
if ($codeGraphAny -and ($CodeGraphNode -eq '' -or $CodeGraphSDK -eq '' -or $CodeGraphScript -eq '')) { throw 'CodeGraphNode, CodeGraphSDK and CodeGraphScript must be supplied together.' }
if ($codeGraphAny) {
    foreach ($pair in @(@('CodeGraphNode', $CodeGraphNode), @('CodeGraphScript', $CodeGraphScript))) {
        if (-not [IO.Path]::IsPathRooted($pair[1]) -or -not (Test-Path -LiteralPath $pair[1] -PathType Leaf)) { throw ($pair[0] + ' must be an existing absolute file.') }
    }
    if (-not [IO.Path]::IsPathRooted($CodeGraphSDK) -or -not (Test-Path -LiteralPath $CodeGraphSDK -PathType Container)) { throw 'CodeGraphSDK must be an existing absolute directory.' }
    if (-not (Test-Path -LiteralPath (Join-Path $CodeGraphSDK 'package.json') -PathType Leaf)) { throw 'CodeGraphSDK must contain package.json.' }
}
if ($AOCIExecutable -ne '') {
    if (-not [IO.Path]::IsPathRooted($AOCIExecutable) -or -not (Test-Path -LiteralPath $AOCIExecutable -PathType Leaf)) { throw 'AOCIExecutable must be an existing absolute file.' }
}
# S5 Laya — all paths, language, stderr directory and lifecycle budgets are required together.
$layaPaths = @($LayaNode, $LayaWorker, $LayaModelDir, $LayaAssets, $LayaSDKEntry, $LayaRPCEntry, $LayaStderrDir)
$layaAny = ($layaPaths | Where-Object { $_ -ne '' }).Count -gt 0 -or $LayaLanguage -ne '' -or $null -ne $LayaStartupTimeoutMS -or $null -ne $LayaShutdownTimeoutMS
if ($layaAny) {
    foreach ($value in $layaPaths) { if ($value -eq '') { throw 'Laya requires all explicit runtime paths, language, private stderr directory and lifecycle budgets.' } }
    if ($LayaLanguage -eq '') { throw 'Laya requires an explicit language.' }
    if ($null -eq $LayaStartupTimeoutMS -or $LayaStartupTimeoutMS -lt 1) { throw 'Laya requires an explicit positive startup timeout in milliseconds.' }
    if ($null -eq $LayaShutdownTimeoutMS -or $LayaShutdownTimeoutMS -lt 1) { throw 'Laya requires an explicit positive shutdown timeout in milliseconds.' }
    foreach ($pair in @(@('LayaNode', $LayaNode), @('LayaWorker', $LayaWorker), @('LayaAssets', $LayaAssets), @('LayaSDKEntry', $LayaSDKEntry), @('LayaRPCEntry', $LayaRPCEntry))) {
        if (-not [IO.Path]::IsPathRooted($pair[1]) -or -not (Test-Path -LiteralPath $pair[1] -PathType Leaf)) { throw ($pair[0] + ' must be an existing absolute file.') }
    }
    if (-not [IO.Path]::IsPathRooted($LayaModelDir) -or -not (Test-Path -LiteralPath $LayaModelDir -PathType Container)) { throw 'LayaModelDir must be an existing absolute directory.' }
    if (-not [IO.Path]::IsPathRooted($LayaStderrDir)) { throw 'LayaStderrDir must be an absolute directory.' }
    if (-not (Test-Path -LiteralPath $LayaStderrDir -PathType Container)) { New-Item -ItemType Directory -Path $LayaStderrDir -Force | Out-Null }
}
if ($DataDir -eq '') { $DataDir = Join-Path $PSScriptRoot 'data' }
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
    $env:ASTROCYTE_RELAY_SESSION_SOURCE = $RelaySessionSource
    if ($null -ne $RelayLiveSeconds) { $env:ASTROCYTE_RELAY_LIVE_SECONDS = [string]$RelayLiveSeconds }
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
    $optional = @{ ASTROCYTE_YT_DLP_PATH = $YtDlp; ASTROCYTE_FFMPEG_PATH = $FFmpeg; ASTROCYTE_WHISPER_BINARY = $WhisperBinary; ASTROCYTE_WHISPER_MODEL = $WhisperModel; ASTROCYTE_LISTING_PYTHON = $ListingPython; ASTROCYTE_UVX_PATH = $UVXExecutable }
    foreach ($name in $optional.Keys) {
        $path = $optional[$name]
        if ($path -ne '') {
            if (-not [IO.Path]::IsPathRooted($path) -or -not (Test-Path -LiteralPath $path -PathType Leaf)) { throw ($name + ' requires an existing absolute file.') }
            [Environment]::SetEnvironmentVariable($name, $path, 'Process')
        }
    }
    # S4 CodeGraph/Repository context — explicit absolute paths, no defaults.
    if ($codeGraphAny) {
        $env:ASTROCYTE_CODEGRAPH_NODE = $CodeGraphNode
        $env:ASTROCYTE_CODEGRAPH_SDK = $CodeGraphSDK
        $env:ASTROCYTE_CODEGRAPH_SCRIPT = $CodeGraphScript
    }
    if ($RepositoryEnvironmentID -ne '') { $env:ASTROCYTE_REPOSITORY_ENVIRONMENT_ID = $RepositoryEnvironmentID }
    if ($AOCIExecutable -ne '') { $env:ASTROCYTE_AOCI_EXECUTABLE = $AOCIExecutable }
    # S5 Laya judgment — all paths, language, stderr and lifecycle budgets required together.
    if ($layaAny) {
        $env:ASTROCYTE_LAYA_NODE = $LayaNode
        $env:ASTROCYTE_LAYA_WORKER = $LayaWorker
        $env:ASTROCYTE_LAYA_MODEL_DIR = $LayaModelDir
        $env:ASTROCYTE_LAYA_ASSETS = $LayaAssets
        $env:ASTROCYTE_LAYA_SDK_ENTRY = $LayaSDKEntry
        $env:ASTROCYTE_LAYA_RPC_ENTRY = $LayaRPCEntry
        $env:ASTROCYTE_LAYA_LANGUAGE = $LayaLanguage
        $env:ASTROCYTE_LAYA_STDERR_DIR = $LayaStderrDir
        $env:ASTROCYTE_LAYA_STARTUP_TIMEOUT_MS = [string]$LayaStartupTimeoutMS
        $env:ASTROCYTE_LAYA_SHUTDOWN_TIMEOUT_MS = [string]$LayaShutdownTimeoutMS
    }
    # Jev cloud REST — only the key, no local model weights.
    if ($TypesafeApiKey -ne '') { $env:TYPESAFE_API_KEY = $TypesafeApiKey }
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
