param(
    [Parameter(Mandatory = $true)]
    [string]$Source,
    [string]$DataDir = (Join-Path $PSScriptRoot 'data'),
    [string]$History = ''
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$binary = Join-Path $PSScriptRoot 'workbench-snapshot.exe'
if (-not (Test-Path -LiteralPath $binary -PathType Leaf)) { throw 'Extract the complete Windows ZIP before publishing a snapshot.' }
if (-not [IO.Path]::IsPathRooted($Source) -or -not (Test-Path -LiteralPath $Source -PathType Leaf)) { throw 'Source requires an existing absolute SQLite file explicitly selected by the user.' }
$data = [IO.Path]::GetFullPath($DataDir)
$root = Join-Path $data 'local\astrocyte-workbench\v0.1'
$arguments = @('--source', $Source, '--root', $root, '--timeout', '120s')
if ($History -ne '') {
    if (-not [IO.Path]::IsPathRooted($History) -or -not (Test-Path -LiteralPath $History -PathType Leaf)) { throw 'History requires an existing absolute history file explicitly selected by the user.' }
    $arguments += @('--history', $History)
}
# 只执行用户请求的单次只读备份，不设置自动任务，不填充历史。
& $binary @arguments
if ($LASTEXITCODE -ne 0) { throw ('Workbench snapshot exited with code ' + $LASTEXITCODE) }
