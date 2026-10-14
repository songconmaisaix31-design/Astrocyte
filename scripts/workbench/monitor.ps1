param([switch]$StopAndAudit)
$ErrorActionPreference = 'Stop'
$runtime = Join-Path $env:LOCALAPPDATA 'astrocyte-workbench/v0.1'
$tool = Join-Path $runtime 'tools/ProcessMonitor/Procmon64.exe'
$isAdmin = (New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (!$isAdmin) { throw 'NOT_RUN: filesystem capture requires an already elevated administrator shell; this script never requests UAC.' }
if ((Get-AuthenticodeSignature -LiteralPath $tool).Status -ne 'Valid') { throw 'Monitor signature invalid.' }
$monitorDir = Join-Path $runtime 'logs/filesystem'
New-Item -ItemType Directory -Force $monitorDir | Out-Null
$pml = Join-Path $monitorDir 'formal-source.pml'
$csv = Join-Path $monitorDir 'formal-source.csv'
$filter = Join-Path $monitorDir 'formal-source.pmc'
# PMC must be saved interactively with exactly three Path Include rules for
# state.sqlite, state.sqlite-wal, state.sqlite-shm below the fixed formal root.
# Capture ALL processes on these paths; filter PID only during audit, so the
# snapshot child and server cannot escape a prematurely chosen PID filter.
if (!(Test-Path -LiteralPath $filter)) { throw 'Save the three-path Include filter to logs/filesystem/formal-source.pmc first.' }
if (!$StopAndAudit) {
    if (Get-Process Procmon64 -ErrorAction SilentlyContinue) { throw 'An existing Procmon capture must not be disturbed.' }
    $capture = Start-Process -FilePath $tool -ArgumentList @('/AcceptEula','/Quiet','/Minimized','/LoadConfig',"`"$filter`"",'/BackingFile',"`"$pml`"") -WindowStyle Hidden -PassThru
    $capture.Id | Set-Content -LiteralPath (Join-Path $monitorDir 'monitor.pid')
    Write-Output 'Capture started; run monitored refresh and API audit, then monitor.ps1 -StopAndAudit.'
    return
}
$capturePid = [int](Get-Content -LiteralPath (Join-Path $monitorDir 'monitor.pid'))
$capture = Get-CimInstance Win32_Process -Filter "ProcessId = $capturePid"
if (!$capture -or $capture.ExecutablePath -ne $tool) { throw 'Owned monitor identity unavailable; no other monitor will be stopped.' }
if (@(Get-Process Procmon64 -ErrorAction SilentlyContinue).Count -ne 1) { throw 'Unexpected monitor instances; review capture ownership.' }
Start-Process -FilePath $tool -ArgumentList '/Terminate' -WindowStyle Hidden -Wait
Start-Process -FilePath $tool -ArgumentList @('/AcceptEula','/Quiet','/OpenLog',"`"$pml`"",'/SaveAs',"`"$csv`"") -WindowStyle Hidden -Wait
$ownedPids = @(Get-Content -LiteralPath (Join-Path $monitorDir 'owned-pids.txt') | ForEach-Object { [int]$_ })
$source = 'C:\Users\DW\AppData\Local\astrocyte-validation\v1-20261012\v4-w1\formal-20261013-first\state.sqlite'
$paths = @($source,"$source-wal","$source-shm")
$rows = @(Import-Csv -LiteralPath $csv | Where-Object { $paths -contains $_.Path -and $ownedPids -contains [int]$_.PID })
$writeOps = @('WriteFile','SetEndOfFileInformationFile','SetAllocationInformationFile','SetDispositionInformationFile','SetRenameInformationFile','SetBasicInformationFile')
$writes = @($rows | Where-Object { $writeOps -contains $_.Operation })
$writeAccess = @($rows | Where-Object { $_.Operation -in @('CreateFile','CreateFileMapping') -and $_.Detail -match 'Generic Write|Write Data|Append Data|Read/Write|PAGE_READWRITE' })
$report = [ordered]@{ events=$rows.Count; writes=$writes.Count; writable_opens_or_mappings=$writeAccess.Count; owned_pids=$ownedPids; status='REVIEW_REQUIRED' }
if ($rows.Count -gt 0 -and $writes.Count -eq 0 -and $writeAccess.Count -eq 0) { $report.status='ZERO_WRITE_OBSERVED' }
$report | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $monitorDir 'summary.json') -Encoding UTF8
$report | ConvertTo-Json -Compress
