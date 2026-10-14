param(
    [Parameter(Mandatory = $true)][string]$Zip,
    [Parameter(Mandatory = $true)][ValidatePattern('^[0-9a-f]{40}$')][string]$PackageSource,
    [ValidateRange(1, 65535)][int]$Port = 17896
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
if ($Port -in @(17321, 17342, 17343, 17344, 17345)) { throw 'Reserved port.' }
$repo = Split-Path (Split-Path $PSScriptRoot -Parent) -Parent
$verification = Join-Path $repo '.verification'
$output = Join-Path $verification ('basic-http-' + (Get-Date -Format 'yyyyMMdd-HHmmss'))
if (Test-Path -LiteralPath $output) { throw 'New output directory required.' }
New-Item -ItemType Directory -Path $output | Out-Null
$process = $null
$serverId = $null
$result = @{ result = 'FAIL'; scope = 'early-package-basic-http-client'; package_source = $PackageSource; zip = [IO.Path]::GetFullPath($Zip); relay = 'NOT_RUN'; native = 'NOT_RUN'; registry = 'NOT_RUN'; DOM = 'NOT_RUN'; output = $output }
function Assert-Basic([bool]$Condition, [string]$Code) {
    if (-not $Condition) { $script:result.error_code = $Code; throw $Code }
}
function Port-Free {
    return @([Net.NetworkInformation.IPGlobalProperties]::GetIPGlobalProperties().GetActiveTcpListeners() | Where-Object { $_.Port -eq $Port }).Count -eq 0
}
try {
    Assert-Basic (Port-Free) 'dedicated_port_not_free'
    Expand-Archive -LiteralPath $Zip -DestinationPath (Join-Path $output 'extracted')
    $package = Join-Path $output 'extracted\Astrocyte-v0.1-windows-x64'
    $executable = Join-Path $package 'astrocyte.exe'
    Assert-Basic (Test-Path -LiteralPath $executable -PathType Leaf) 'zip_executable_missing'
    $result.executable = $executable
    $homeParts = ([Environment]::GetFolderPath([Environment+SpecialFolder]::UserProfile) -split '[\\/]') | Where-Object { $_ -ne '' }
    $homePattern = (($homeParts | ForEach-Object { [Regex]::Escape($_) }) -join '[\\/]+') + '[\\/]+'
    $pattern = $homePattern + '|astrocyte-validation[\\/]+v1-20261012|formal-20261013-first|v1r-w1[\\/]+replay-002'
    $matchedFiles = 0
    $matchCount = 0
    foreach ($file in @(Get-ChildItem -LiteralPath $package -Recurse -File)) {
        $contents = [Text.Encoding]::UTF8.GetString([IO.File]::ReadAllBytes($file.FullName))
        $count = [Regex]::Matches($contents, $pattern, [Text.RegularExpressions.RegexOptions]::IgnoreCase).Count
        if ($count -gt 0) { $matchedFiles++; $matchCount += $count }
    }
    $result.private_path_matched_files = $matchedFiles
    $result.private_path_match_count = $matchCount
    Assert-Basic ($matchedFiles -eq 0) 'private_path_scan_failed'
    $start = Join-Path $package 'Start-Astrocyte.ps1'
    Assert-Basic (Test-Path -LiteralPath $start -PathType Leaf) 'original_package_start_missing'
    $data = Join-Path $output 'data'
    $info = [Diagnostics.ProcessStartInfo]::new()
    $info.FileName = Join-Path $env:SystemRoot 'System32\WindowsPowerShell\v1.0\powershell.exe'
    $info.Arguments = '-NoLogo -NoProfile -ExecutionPolicy Bypass -File "' + $start + '" -Port ' + $Port + ' -DataDir "' + $data + '"'
    $info.WorkingDirectory = $package
    $info.UseShellExecute = $false
    $info.CreateNoWindow = $true
    $info.RedirectStandardOutput = $true
    $info.RedirectStandardError = $true
    $info.EnvironmentVariables['PATH'] = (Join-Path $env:SystemRoot 'System32') + ';' + $env:SystemRoot
    foreach ($name in @('USERPROFILE', 'HOME', 'APPDATA', 'LOCALAPPDATA', 'CODEX_HOME', 'CLAUDE_CONFIG_DIR', 'PI_CODING_AGENT_DIR')) {
        $identity = Join-Path $output ('empty-identity\' + $name)
        New-Item -ItemType Directory -Path $identity -Force | Out-Null
        $info.EnvironmentVariables[$name] = $identity
    }
    foreach ($name in @($info.EnvironmentVariables.Keys)) {
        if ([string]$name -like 'ASTROCYTE_*') { $info.EnvironmentVariables.Remove($name) }
    }
    $process = [Diagnostics.Process]::Start($info)
    $stdout = $process.StandardOutput.ReadToEndAsync()
    $stderr = $process.StandardError.ReadToEndAsync()
    $deadline = [DateTime]::UtcNow.AddSeconds(45)
    while ((Port-Free) -and [DateTime]::UtcNow -lt $deadline) {
        Assert-Basic (-not $process.HasExited) 'original_package_exited_before_listen'
        Start-Sleep -Milliseconds 250
    }
    Assert-Basic (-not (Port-Free)) 'original_package_listen_timeout'
    $children = @(Get-CimInstance Win32_Process -Filter ('ParentProcessId = ' + $process.Id) | Where-Object { $_.ExecutablePath -ieq $executable })
    Assert-Basic ($children.Count -eq 1) 'owned_packaged_process_not_identified'
    $serverId = [int]$children[0].ProcessId
    $result.server_process_id = $serverId
    $invoke = @{ Client = (Join-Path $verification 'native-client.exe'); ServerProcessId = $serverId; Executable = $executable; Zip = $Zip; PackageSource = $PackageSource; Base = ('http://127.0.0.1:' + $Port); VerificationRoot = $verification; Output = (Join-Path $output 'protected-read'); Mode = 'request'; Method = 'GET'; Path = '/api/v1/local-projects'; ExpectStatus = 200 }
    & (Join-Path $PSScriptRoot 'Invoke-NativeClient.ps1') @invoke
    $result.result = 'PASS'
    $result.health_bootstrap_protected_read = 'PASS'
} finally {
    if ($null -ne $process) {
        if (-not $process.HasExited) {
            & (Join-Path $env:SystemRoot 'System32\taskkill.exe') /PID $process.Id /T /F | Out-Null
            Assert-Basic ($LASTEXITCODE -eq 0) 'owned_process_termination_failed'
            Assert-Basic ($process.WaitForExit(10000)) 'owned_process_exit_unconfirmed'
        }
        [IO.File]::WriteAllText((Join-Path $output 'server-stdout-local.log'), $stdout.GetAwaiter().GetResult())
        [IO.File]::WriteAllText((Join-Path $output 'server-stderr-local.log'), $stderr.GetAwaiter().GetResult())
        $process.Dispose()
        Assert-Basic (Port-Free) 'owned_port_not_closed'
        $result.owned_process_and_port_closed = 'PASS'
    }
    $result | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath (Join-Path $output 'basic-safe.json') -Encoding UTF8
    Write-Output ('LOCAL_EVIDENCE=' + $output)
}
