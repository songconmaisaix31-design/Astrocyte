$ErrorActionPreference = 'Stop'
$runtime = Join-Path $env:LOCALAPPDATA 'astrocyte-workbench/v0.1'
if (!(Test-Path -LiteralPath (Join-Path $runtime 'config.json'))) { throw 'Run start.ps1 first.' }
if ((Get-TimeZone).Id -ne 'China Standard Time') { throw '00:30 schedule requires Shanghai system timezone.' }
$user = [System.Security.Principal.WindowsIdentity]::GetCurrent().Name
$principal = New-ScheduledTaskPrincipal -UserId $user -LogonType Interactive -RunLevel Limited
$settings = New-ScheduledTaskSettingsSet -StartWhenAvailable -MultipleInstances IgnoreNew -ExecutionTimeLimit (New-TimeSpan -Minutes 5)
$refresh = Join-Path $runtime 'scripts/refresh.ps1'
$action = New-ScheduledTaskAction -Execute 'powershell.exe' -Argument "-NoProfile -NonInteractive -ExecutionPolicy Bypass -WindowStyle Hidden -File `"$refresh`"" -WorkingDirectory $runtime
$daily = New-ScheduledTaskTrigger -Daily -At '00:30'
Register-ScheduledTask -TaskName 'Astrocyte-Workbench-v01-Snapshot' -Action $action -Trigger $daily -Principal $principal -Settings $settings -Force | Select-Object TaskName,State
$resume = Join-Path $runtime 'scripts/resume.ps1'
$action = New-ScheduledTaskAction -Execute 'powershell.exe' -Argument "-NoProfile -NonInteractive -ExecutionPolicy Bypass -WindowStyle Hidden -File `"$resume`"" -WorkingDirectory $runtime
$login = New-ScheduledTaskTrigger -AtLogOn -User $user
Register-ScheduledTask -TaskName 'Astrocyte-Workbench-v01-Start' -Action $action -Trigger $login -Principal $principal -Settings $settings -Force | Select-Object TaskName,State
