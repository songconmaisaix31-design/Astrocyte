param(
    [Parameter(Mandatory = $true)]
    [string]$Zip,
    [ValidateRange(1, 65535)]
    [int]$Port = 17887,
    [switch]$RequireCombined
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
Add-Type -AssemblyName System.Net.Http
Add-Type -AssemblyName System.IO.Compression.FileSystem
$repo = Split-Path $PSScriptRoot -Parent
$output = Join-Path $repo ('.verification\verify-' + (Get-Date -Format 'yyyyMMdd-HHmmss'))
if (Test-Path -LiteralPath $output) { throw 'Verification directory already exists.' }
New-Item -ItemType Directory -Path $output | Out-Null
Start-Transcript -Path (Join-Path $output 'verify.log') | Out-Null
$process = $null
$client = $null
$archive = $null
$base = 'http://127.0.0.1:' + $Port
$package = Join-Path $output 'extracted\Astrocyte-v0.1-windows-x64'
$checks = [Collections.Generic.List[object]]::new()
function Assert-Release([bool]$condition, [string]$name) {
    if (-not $condition) { throw ('FAIL: ' + $name) }
    $checks.Add(@{ check = $name; result = 'PASS' })
    Write-Output ('PASS: ' + $name)
}
function New-Client {
    $handler = [Net.Http.HttpClientHandler]::new()
    $handler.UseProxy = $false
    $handler.CookieContainer = [Net.CookieContainer]::new()
    $script:client = [Net.Http.HttpClient]::new($handler)
    $script:client.Timeout = [TimeSpan]::FromSeconds(10)
}
function Request-Release([string]$method, [string]$path, [string]$body = '', [hashtable]$headers = @{}) {
    $request = [Net.Http.HttpRequestMessage]::new([Net.Http.HttpMethod]::new($method), ($base + $path))
    foreach ($key in $headers.Keys) { $request.Headers.Add($key, [string]$headers[$key]) }
    if ($body -ne '') { $request.Content = [Net.Http.StringContent]::new($body, [Text.Encoding]::UTF8, 'application/json') }
    try {
        $response = $client.SendAsync($request).GetAwaiter().GetResult()
        try { return @{ status = [int]$response.StatusCode; body = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult() } }
        finally { $response.Dispose() }
    } finally { $request.Dispose() }
}
function Start-Release([string]$cycle) {
    $info = [Diagnostics.ProcessStartInfo]::new()
    $info.FileName = Join-Path $env:SystemRoot 'System32\WindowsPowerShell\v1.0\powershell.exe'
    $info.Arguments = '-NoLogo -NoProfile -ExecutionPolicy Bypass -File "' + (Join-Path $package 'Start-Astrocyte.ps1') + '" -Port ' + $Port
    $info.WorkingDirectory = $package
    $info.UseShellExecute = $false
    $info.CreateNoWindow = $true
    $info.RedirectStandardOutput = $true
    $info.RedirectStandardError = $true
    # 使用空的验收身份目录与系统 PATH，禁止发现宿主原会话或开发依赖。
    $homeDir = Join-Path $output 'isolated-user'
    New-Item -ItemType Directory -Path $homeDir -Force | Out-Null
    $info.EnvironmentVariables['PATH'] = (Join-Path $env:SystemRoot 'System32') + ';' + $env:SystemRoot
    foreach ($name in @('USERPROFILE', 'HOME', 'APPDATA', 'LOCALAPPDATA', 'CODEX_HOME', 'CLAUDE_CONFIG_DIR', 'PI_CODING_AGENT_DIR')) {
        $info.EnvironmentVariables[$name] = Join-Path $homeDir $name
        New-Item -ItemType Directory -Path $info.EnvironmentVariables[$name] -Force | Out-Null
    }
    $script:process = [Diagnostics.Process]::Start($info)
    $script:stdout = $process.StandardOutput.ReadToEndAsync()
    $script:stderr = $process.StandardError.ReadToEndAsync()
    $deadline = [DateTime]::UtcNow.AddSeconds(45)
    $ready = $false
    while ([DateTime]::UtcNow -lt $deadline) {
        if ($process.HasExited) { throw ('Application exited before health: ' + $process.ExitCode) }
        # 只等待真实 HTTP 监听建立；应用错误响应立即失败。
        $probe = $null
        try { $probe = Request-Release 'GET' '/api/v1/health' }
        catch [Net.Http.HttpRequestException] { Start-Sleep -Milliseconds 250; continue }
        if ($probe.status -ne 200) { throw ('Health returned ' + $probe.status) }
        $ready = $true
        break
    }
    Assert-Release $ready ($cycle + ' real HTTP health')
}
function Stop-Release([string]$cycle) {
    if ($null -eq $process) { return }
    if (-not $process.HasExited) {
        & (Join-Path $env:SystemRoot 'System32\taskkill.exe') /PID $process.Id /T /F | Out-Null
        if ($LASTEXITCODE -ne 0) { throw 'Owned process tree termination failed.' }
        if (-not $process.WaitForExit(10000)) { throw 'Owned process tree exit unconfirmed.' }
    }
    [IO.File]::WriteAllText((Join-Path $output ($cycle + '-stdout.log')), $stdout.GetAwaiter().GetResult())
    [IO.File]::WriteAllText((Join-Path $output ($cycle + '-stderr.log')), $stderr.GetAwaiter().GetResult())
    $process.Dispose()
    $script:process = $null
    Assert-Release (@(Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction SilentlyContinue).Count -eq 0) ($cycle + ' owned listener stopped')
}
try {
    Assert-Release (@(Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction SilentlyContinue).Count -eq 0) 'isolated port initially free'
    $archive = [IO.Compression.ZipFile]::OpenRead([IO.Path]::GetFullPath($Zip))
    $count = 0
    foreach ($entry in $archive.Entries) {
        $name = $entry.FullName.Replace('\', '/')
        if ($name.EndsWith('/')) { continue }
        $count++
        $allowed = $name -match '^Astrocyte-v0\.1-windows-x64/(astrocyte(-mcp|-sleep)?\.exe|Start-Astrocyte\.(ps1|cmd)|README\.md|LICENSE|web/[^:]+)$'
        $unsafe = $name -match '(^|/)(\.\.|node_modules|data|test-results|\.git)(/|$)|\.(sqlite|db|log|jsonl)(-|$|\.)'
        if (-not $allowed -or $unsafe) { throw ('Unexpected ZIP member: ' + $name) }
    }
    Assert-Release ($count -gt 8) 'ZIP production allowlist excludes data and development dependencies'
    $archive.Dispose()
    $archive = $null
    Expand-Archive -LiteralPath $Zip -DestinationPath (Join-Path $output 'extracted')
    Assert-Release (-not (Test-Path -LiteralPath (Join-Path $package 'data'))) 'fresh extraction has no database or user data'
    New-Client
    Start-Release 'initial'
    foreach ($route in @('/', '/attention', '/workspace', '/swarm', '/today', '/focus', '/agents', '/events')) {
        $page = Request-Release 'GET' $route
        Assert-Release ($page.status -eq 200 -and $page.body.Contains('<div id="root">')) ('static SPA ' + $route)
    }
    foreach ($file in @(Get-ChildItem -LiteralPath (Join-Path $package 'web\assets') -File)) {
        $asset = Request-Release 'GET' ('/assets/' + $file.Name)
        Assert-Release ($asset.status -eq 200) ('static asset ' + $file.Name)
    }
    $denied = Request-Release 'GET' '/api/v1/material-domains'
    Assert-Release ($denied.status -eq 403) 'protected read requires session'
    $session = Request-Release 'GET' '/api/v1/auth/session'
    Assert-Release ($session.status -eq 200) 'real session bootstrap'
    $identity = $session.body | ConvertFrom-Json
    Assert-Release ($identity.actor_kind -eq 'human' -and $identity.csrf_token.Length -gt 0) 'real human and CSRF binding'
    $command = @{ schema_version = 1; request_id = [guid]::NewGuid().ToString(); expected_version = 1; title = 'Windows release verification'; description = 'Created through the real application API in the isolated extracted package.' } | ConvertTo-Json -Compress
    $headers = @{ Origin = $base; 'Idempotency-Key' = [guid]::NewGuid().ToString() }
    $deniedWrite = Request-Release 'POST' '/api/v1/material-domains' $command $headers
    Assert-Release ($deniedWrite.status -eq 403) 'real write without CSRF refused'
    $headers['X-CSRF-Token'] = $identity.csrf_token
    $write = Request-Release 'POST' '/api/v1/material-domains' $command $headers
    Assert-Release ($write.status -eq 201) 'real authenticated SQLite write'
    $created = $write.body | ConvertFrom-Json
    $list = Request-Release 'GET' '/api/v1/material-domains'
    Assert-Release ($list.status -eq 200 -and $list.body.Contains($created.domain.id)) 'real protected read observes committed domain'
    $headers['Origin'] = 'https://example.invalid'
    $cross = Request-Release 'POST' '/api/v1/material-domains' $command $headers
    Assert-Release ($cross.status -eq 403) 'cross-origin write refused'
    if ($RequireCombined) {
        foreach ($path in @('/api/v1/cognition/health', '/api/v1/cognition/today')) {
            $snapshot = Request-Release 'GET' $path
            Assert-Release ($snapshot.status -eq 503 -and $snapshot.body.Contains('provider_unavailable')) ('missing snapshot remains unavailable ' + $path)
        }
        foreach ($path in @('/api/v1/relay/agents', '/api/v1/relay/tasks')) {
            $relay = Request-Release 'GET' $path
            Assert-Release ($relay.status -eq 200) ('unconfigured Relay real state ' + $path)
            [IO.File]::WriteAllText((Join-Path $output (($path.Split('/')[-1]) + '-safe.json')), $relay.body)
        }
        $events = Request-Release 'GET' '/api/v1/relay/session-events'
        Assert-Release ($events.status -eq 200 -or ($events.status -eq 503 -and $events.body.Contains('provider_unavailable'))) 'Relay session events preserve empty or unavailable state'
        [IO.File]::WriteAllText((Join-Path $output 'session-events-safe.json'), $events.body)
    }
    $db = Join-Path $package 'data\local\astrocyte-workbench\v0.2-access\app-state\state.sqlite'
    Assert-Release ((Test-Path -LiteralPath $db -PathType Leaf) -and (Get-Item -LiteralPath $db).Length -gt 0) 'real independent SQLite created in package data'
    Stop-Release 'initial'
    $client.Dispose()
    New-Client
    Start-Release 'restart'
    $expired = Request-Release 'GET' '/api/v1/material-domains'
    Assert-Release ($expired.status -eq 403) 'restart requires new session'
    $renewed = Request-Release 'GET' '/api/v1/auth/session'
    Assert-Release ($renewed.status -eq 200) 'restart session bootstrap'
    $persisted = Request-Release 'GET' '/api/v1/material-domains'
    Assert-Release ($persisted.status -eq 200 -and $persisted.body.Contains($created.domain.id)) 'SQLite domain survives actual process restart'
    Stop-Release 'restart'
    $summary = @{ result = 'PASS'; scope = $(if ($RequireCombined) { 'combined-package-runtime' } else { 'candidate-package-runtime' }); zip = [IO.Path]::GetFullPath($Zip); output = $output; checks = $checks; limitations = @('No default browser or native controller acceptance is claimed.', 'Optional summarize/media dependencies were not invoked.', 'Restart exercised owned Windows process termination and SQLite recovery.') }
    $summary | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath (Join-Path $output 'result.json') -Encoding UTF8
    Write-Output ('VERIFICATION=' + $output)
} finally {
    if ($null -ne $archive) { $archive.Dispose() }
    if ($null -ne $process) { Stop-Release 'failed' }
    if ($null -ne $client) { $client.Dispose() }
    Stop-Transcript | Out-Null
}
