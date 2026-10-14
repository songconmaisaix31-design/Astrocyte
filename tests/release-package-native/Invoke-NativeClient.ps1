param(
    [Parameter(Mandatory = $true)][string]$Client,
    [Parameter(Mandatory = $true)][int]$ServerProcessId,
    [Parameter(Mandatory = $true)][string]$Executable,
    [Parameter(Mandatory = $true)][string]$Zip,
    [Parameter(Mandatory = $true)][string]$PackageSource,
    [Parameter(Mandatory = $true)][string]$Base,
    [Parameter(Mandatory = $true)][string]$VerificationRoot,
    [Parameter(Mandatory = $true)][string]$Output,
    [Parameter(Mandatory = $true)][ValidateSet('baseline', 'registry', 'request', 'events', 'idle')][string]$Mode,
    [string]$Method = 'GET', [string]$Path = '', [string]$BodyFile = '',
    [int]$ExpectStatus = 200, [string]$Database = '', [string]$RelayDatabase = '', [string]$Task = '', [string]$Project = '',
    [ValidateRange(1, 100000)][int]$MinimumEvents = 1
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
$expected = [IO.Path]::GetFullPath($Executable)
$server = Get-Process -Id $ServerProcessId -ErrorAction Stop
if ($server.HasExited -or $server.MainModule.FileName -ine $expected) { throw 'Actual process executable does not match the extracted ZIP executable.' }
$uri = [Uri]$Base
if ($uri.Scheme -ne 'http' -or $uri.Host -ne '127.0.0.1') { throw 'Dedicated loopback URL required.' }
$listeners = @(Get-NetTCPConnection -State Listen -LocalPort $uri.Port -ErrorAction Stop)
if ($listeners.Count -ne 1 -or $listeners[0].OwningProcess -ne $ServerProcessId) { throw 'Actual HTTP listener does not belong to the packaged process.' }
$arguments = @('--base', $Base, '--mode', $Mode, '--output', $Output, '--verification-root', $VerificationRoot, '--zip', $Zip, '--executable', $expected, '--package-source', $PackageSource, '--method', $Method, '--expect-status', [string]$ExpectStatus, '--minimum-events', [string]$MinimumEvents)
foreach ($pair in @(@('--path', $Path), @('--body-file', $BodyFile), @('--db', $Database), @('--relay-db', $RelayDatabase), @('--task', $Task), @('--project', $Project))) {
    if ($pair[1] -ne '') { $arguments += $pair }
}
# 验收客户端只执行一个动作，所有写入失败均需检查原回执。
& $Client @arguments
if ($LASTEXITCODE -ne 0) { throw 'Native package client failed; inspect local summary and original response without resending writes.' }
