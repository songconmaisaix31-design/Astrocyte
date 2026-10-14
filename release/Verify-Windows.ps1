param(
    [Parameter(Mandatory = $true)]
    [string]$Zip,
    [ValidateRange(1, 65535)]
    [int]$Port = 17887,
    [switch]$RequireCombined,
    [string]$Source = '',
    [string]$CognitionSource = '',
    [string]$CognitionHistory = '',
    [string]$CognitionPublicationRoot = '',
    [string]$CognitionSafeDir = '',
    [string]$CognitionSourceSha = ''
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
Add-Type -AssemblyName System.Net.Http
Add-Type -AssemblyName System.IO.Compression.FileSystem
$repo = Split-Path $PSScriptRoot -Parent
if ($CognitionSource -ne '') { $RequireCombined = $true }
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
function Test-PackagePrivatePaths {
    $segments = ([Environment]::GetFolderPath([Environment+SpecialFolder]::UserProfile) -split '[\\/]') | Where-Object { $_ -ne '' }
    $homePattern = (($segments | ForEach-Object { [Regex]::Escape($_) }) -join '[\\/]+') + '[\\/]+'
    $pattern = $homePattern + '|astrocyte-validation[\\/]+v1-20261012|formal-20261013-first|v1r-w1[\\/]+replay-002'
    $findings = @()
    # 仅扫描生产文件中的字符串，不解析可执行格式，不输出实际匹配内容。
    foreach ($file in @(Get-ChildItem -LiteralPath $package -Recurse -File)) {
        $contents = [Text.Encoding]::UTF8.GetString([IO.File]::ReadAllBytes($file.FullName))
        $count = [Regex]::Matches($contents, $pattern, [Text.RegularExpressions.RegexOptions]::IgnoreCase).Count
        if ($count -gt 0) {
            $relative = $file.FullName.Substring($package.Length + 1)
            $findings += @{ file = $relative; matches = $count }
            Write-Output ('PRIVATE_PATH_MATCHES: ' + $relative + ' count=' + $count)
        }
    }
    @{ scope = 'UTF8 production strings; current user home and known formal source constants'; files_with_matches = $findings.Count; findings = $findings } | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath (Join-Path $output 'private-path-scan.json') -Encoding UTF8
    Assert-Release ($findings.Count -eq 0) 'production package excludes current personal paths and known formal source constants'
}
function New-Client {
    $handler = [Net.Http.HttpClientHandler]::new()
    $handler.UseProxy = $false
    $handler.CookieContainer = [Net.CookieContainer]::new()
    $script:client = [Net.Http.HttpClient]::new($handler)
    $script:client.Timeout = [TimeSpan]::FromSeconds(10)
}
function Test-PortFree {
    return @([Net.NetworkInformation.IPGlobalProperties]::GetIPGlobalProperties().GetActiveTcpListeners() | Where-Object { $_.Port -eq $Port }).Count -eq 0
}
function Request-Release([string]$method, [string]$path, [string]$body = '', [hashtable]$headers = @{}) {
    $request = [Net.Http.HttpRequestMessage]::new([Net.Http.HttpMethod]::new($method), ($base + $path))
    foreach ($key in $headers.Keys) { $request.Headers.Add($key, [string]$headers[$key]) }
    if ($body -ne '') { $request.Content = [Net.Http.StringContent]::new($body, [Text.Encoding]::UTF8, 'application/json') }
    try {
        $response = $client.SendAsync($request).GetAwaiter().GetResult()
        try {
            $snapshotAt = ''
            if ($response.Headers.Contains('X-Snapshot-At')) { $snapshotAt = [string]::Join(',', [string[]]$response.Headers.GetValues('X-Snapshot-At')) }
            return @{ status = [int]$response.StatusCode; body = $response.Content.ReadAsStringAsync().GetAwaiter().GetResult(); snapshot_at = $snapshotAt }
        }
        finally { $response.Dispose() }
    } finally { $request.Dispose() }
}
function Test-CognitionInput {
    foreach ($path in @($CognitionSource, $CognitionHistory, $CognitionPublicationRoot, $CognitionSafeDir)) {
        Assert-Release ([IO.Path]::IsPathRooted($path) -and (Test-Path -LiteralPath $path)) 'explicit actual cognition input exists'
    }
    Assert-Release ($Source -match '^[0-9a-f]{40}$' -and $CognitionSourceSha -match '^[0-9a-f]{40}$') 'cognition verification records exact package and Q source commits'
    Assert-Release ([IO.Path]::GetFileName($Zip).EndsWith('-' + $Source.Substring(0, 12) + '.zip')) 'ZIP name matches explicit package source'
    Push-Location $repo
    try {
        & git merge-base --is-ancestor $CognitionSourceSha $Source
        Assert-Release ($LASTEXITCODE -eq 0) 'exact Q source is an ancestor of package source'
    } finally { Pop-Location }
    $script:cognitionExpected = Get-Content -LiteralPath (Join-Path $CognitionSafeDir 'sql-expected.json') -Raw -Encoding UTF8 | ConvertFrom-Json
    $script:cognitionEvents = @(Get-Content -LiteralPath (Join-Path $CognitionSafeDir 'events-six-fields.json') -Raw -Encoding UTF8 | ConvertFrom-Json)
    $script:cognitionAdmission = Get-Content -LiteralPath (Join-Path $CognitionSafeDir 'admission-safe.json') -Raw -Encoding UTF8 | ConvertFrom-Json
    $publication = Get-Content -LiteralPath (Join-Path $CognitionPublicationRoot 'current.json') -Raw -Encoding UTF8 | ConvertFrom-Json
    Assert-Release ($cognitionAdmission.state_phase -eq 'after_sleep' -and $cognitionAdmission.actual_sleep_runs -gt 0) 'actual cognition input is the admitted after-Sleep state'
    Assert-Release ($publication.snapshot_at -ceq $cognitionExpected.snapshot_at) 'SQL expected values bind to current admitted publication'
    $expectedHistory = Join-Path $CognitionPublicationRoot ('snapshots\' + $publication.generation + '\history.json')
    Assert-Release ([IO.Path]::GetFullPath($CognitionHistory) -ieq [IO.Path]::GetFullPath($expectedHistory)) 'explicit actual history belongs to admitted publication'
    $script:cognitionHistoryExpected = Get-Content -LiteralPath $CognitionHistory -Raw -Encoding UTF8 | ConvertFrom-Json
    $zone = [TimeZoneInfo]::FindSystemTimeZoneById('China Standard Time')
    $day = [TimeZoneInfo]::ConvertTime([DateTimeOffset]::UtcNow, $zone).ToString('yyyy-MM-dd')
    Assert-Release ($cognitionExpected.day -ceq $day) 'daily SQL expected values match actual Shanghai day'
    Assert-Release (@($cognitionExpected.values).Count -eq 10 -and $cognitionEvents.Count -gt 0 -and $cognitionEvents.Count -eq $cognitionAdmission.six_field_sql_matches) 'actual SQL10 and nonempty six-field source are complete'
}
function Test-ActualCognition {
    $publisher = Join-Path $package 'Publish-WorkbenchSnapshot.ps1'
    & (Join-Path $env:SystemRoot 'System32\WindowsPowerShell\v1.0\powershell.exe') -NoProfile -ExecutionPolicy Bypass -File $publisher -Source $CognitionSource -History $CognitionHistory
    Assert-Release ($LASTEXITCODE -eq 0) 'packaged snapshot publishes explicit actual public after-Sleep source'
    $publication = Get-Content -LiteralPath (Join-Path $package 'data\local\astrocyte-workbench\v0.1\current.json') -Raw -Encoding UTF8 | ConvertFrom-Json
    $views = @{}
    foreach ($name in @('health', 'today', 'focus', 'agents', 'topics/evidence', 'filters')) {
        $view = Request-Release 'GET' ('/api/v1/cognition/' + $name)
        Assert-Release ($view.status -eq 200 -and $view.snapshot_at -ceq $publication.snapshot_at) ('actual public cognition view ' + $name)
        $views[$name] = $view.body | ConvertFrom-Json
    }
    foreach ($metric in $cognitionExpected.values) {
        $actual = $null
        switch ($metric.metric) {
            'daily_count' {
                $rows = @($views['today'].daily_counts | Where-Object { $_.day -ceq $metric.day })
                Assert-Release ($rows.Count -eq 1) 'actual SQL day has one HTTP count'
                $actual = $rows[0].count
            }
            { $_ -in @('attention', 'momentum', 'confidence') } {
                $rows = @($views['focus'].topics | Where-Object { $_.topic -ceq $metric.topic })
                Assert-Release ($rows.Count -eq 1) 'actual SQL topic has one HTTP state'
                $actual = $rows[0].($metric.metric)
            }
            { $_ -in @('today', 'days7', 'days30') } {
                $rows = @($views['agents'].groups | Where-Object { $_.source -ceq $metric.source })
                Assert-Release ($rows.Count -eq 1) 'actual SQL source has one HTTP group'
                $actual = $rows[0].($metric.metric)
            }
            'total' { $actual = $views['topics/evidence'].total }
            default { throw 'Unsupported SQL expected metric.' }
        }
        Assert-Release ($null -ne $actual -and $actual -is [ValueType] -and $actual -isnot [bool] -and [Math]::Abs([double]$actual - [double]$metric.value) -le 1e-12) ('actual independent SQL value ' + $metric.metric)
    }
    Assert-Release ($views['focus'].topics[0].topic -ceq $cognitionAdmission.highest_attention_topic) 'actual highest attention topic retains SQL rank'
    foreach ($topic in $cognitionAdmission.history_points.PSObject.Properties) {
        $series = @($views['focus'].series | Where-Object { $_.topic -ceq $topic.Name })
        Assert-Release ($series.Count -eq 1 -and @($series[0].points).Count -eq $topic.Value) 'actual history retains its known point count'
        $points = @($cognitionHistoryExpected.($topic.Name))
        for ($index = 0; $index -lt $points.Count; $index++) {
            $actual = $series[0].points[$index]
            $stamp = [DateTimeOffset]::Parse($actual.ts, [Globalization.CultureInfo]::InvariantCulture).ToUnixTimeMilliseconds()
            if ($stamp -ne $points[$index].ts -or [Math]::Abs([double]$actual.attention - [double]$points[$index].attention) -gt 1e-12) { throw 'Actual history point differs from explicitly admitted history.' }
        }
    }
    $offset = 0
    for ($page = 1; $offset -lt $cognitionEvents.Count; $page++) {
        $response = Request-Release 'GET' ('/api/v1/cognition/events?page=' + $page + '&page_size=100')
        Assert-Release ($response.status -eq 200 -and $response.snapshot_at -ceq $publication.snapshot_at) ('actual six-field page ' + $page)
        $result = $response.body | ConvertFrom-Json
        $wanted = [Math]::Min(100, $cognitionEvents.Count - $offset)
        Assert-Release ($result.total -eq $cognitionEvents.Count -and @($result.events).Count -eq $wanted) ('actual six-field page count ' + $page)
        foreach ($row in $result.events) {
            $expected = $cognitionEvents[$offset]
            $names = ($row.PSObject.Properties.Name | Sort-Object) -join ','
            if ($names -cne 'event_id,kind,scope,tokens,topic,ts') { throw ('Unexpected public event fields at row ' + $offset) }
            foreach ($field in @('event_id', 'kind', 'topic', 'tokens', 'scope', 'ts')) {
                if ($row.$field -cne $expected.$field) { throw ('SQL event field mismatch at row ' + $offset + ' field=' + $field) }
            }
            $offset++
        }
    }
    Assert-Release ($offset -eq $cognitionEvents.Count) 'all actual public events match independent SQL six fields'
    foreach ($size in $cognitionAdmission.page_sizes_verified) {
        $topic = [Uri]::EscapeDataString($cognitionAdmission.pagination_topic)
        $response = Request-Release 'GET' ('/api/v1/cognition/topics/evidence?topic=' + $topic + '&page_size=' + $size + '&page=2')
        Assert-Release ($response.status -eq 200 -and $response.snapshot_at -ceq $publication.snapshot_at) ('actual same-topic pagination ' + $size)
        $result = $response.body | ConvertFrom-Json
        Assert-Release ($result.total -eq $cognitionAdmission.pagination_events -and @($result.events).Count -eq $size) ('actual same-topic page preserves SQL count ' + $size)
    }
    [IO.File]::WriteAllText((Join-Path $output 'cognition-input-safe.json'), (@{ source_commit = $CognitionSourceSha; admitted_phase = $cognitionAdmission.state_phase; original_snapshot_at = $cognitionExpected.snapshot_at; published_snapshot_at = $publication.snapshot_at; sql_values = 10; six_field_events = $offset } | ConvertTo-Json -Compress))
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
    Assert-Release (Test-PortFree) ($cycle + ' owned listener stopped')
}
try {
    if ($CognitionSource -ne '') { Test-CognitionInput }
    Assert-Release (Test-PortFree) 'isolated port initially free'
    $archive = [IO.Compression.ZipFile]::OpenRead([IO.Path]::GetFullPath($Zip))
    $count = 0
    foreach ($entry in $archive.Entries) {
        $name = $entry.FullName.Replace('\', '/')
        if ($name.EndsWith('/')) { continue }
        $count++
        $allowed = $name -match '^Astrocyte-v0\.1-windows-x64/(astrocyte(-mcp|-sleep)?\.exe|workbench-snapshot\.exe|Start-Astrocyte\.(ps1|cmd)|Publish-WorkbenchSnapshot\.ps1|README\.md|LICENSE|web/[^:]+)$'
        $unsafe = $name -match '(^|/)(\.\.|node_modules|data|test-results|\.git)(/|$)|\.(sqlite|db|log|jsonl)(-|$|\.)'
        if (-not $allowed -or $unsafe) { throw ('Unexpected ZIP member: ' + $name) }
    }
    Assert-Release ($count -gt 8) 'ZIP production allowlist excludes data and development dependencies'
    $archive.Dispose()
    $archive = $null
    Expand-Archive -LiteralPath $Zip -DestinationPath (Join-Path $output 'extracted')
    Assert-Release (-not (Test-Path -LiteralPath (Join-Path $package 'data'))) 'fresh extraction has no database or user data'
    Test-PackagePrivatePaths
    New-Client
    Start-Release 'initial'
    $routes = @('/', '/attention', '/workspace', '/swarm')
    if ($RequireCombined) { $routes += @('/cognition/today', '/cognition/focus', '/cognition/agents', '/cognition/events', '/workspace/works') }
    foreach ($route in $routes) {
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
            $state = $relay.body | ConvertFrom-Json
            Assert-Release ($state.schema_version -eq 1) ('Relay state schema ' + $path)
            if ($path.EndsWith('/agents')) {
                Assert-Release ($null -ne $state.agents -and @($state.agents).Count -eq 0) 'fresh Relay registry contains no inferred Agent'
            } else {
                Assert-Release ($null -ne $state.tasks -and @($state.tasks).Count -eq 0 -and @($state.dispatched_work_ids).Count -eq 0) 'fresh Relay has no task or inferred dispatch'
            }
            [IO.File]::WriteAllText((Join-Path $output (($path.Split('/')[-1]) + '-safe.json')), $relay.body)
        }
        $events = Request-Release 'GET' '/api/v1/relay/session-events'
        Assert-Release ($events.status -eq 200 -or ($events.status -eq 503 -and $events.body.Contains('provider_unavailable'))) 'Relay session events preserve empty or unavailable state'
        if ($events.status -eq 200) {
            $admission = $events.body | ConvertFrom-Json
            Assert-Release (@($admission.events).Count -eq 0 -and $admission.total -eq 0) 'fresh Relay has no invented session event'
            foreach ($coverage in $admission.coverage) {
                Assert-Release ($coverage.status -in @('unknown', 'partial') -and -not $coverage.session_complete) 'unconfigured native coverage preserves uncertainty'
            }
        }
        [IO.File]::WriteAllText((Join-Path $output 'session-events-safe.json'), $events.body)
    }
    $db = Join-Path $package 'data\local\astrocyte-workbench\v0.2-access\app-state\state.sqlite'
    Assert-Release ((Test-Path -LiteralPath $db -PathType Leaf) -and (Get-Item -LiteralPath $db).Length -gt 0) 'real independent SQLite created in package data'
    if ($RequireCombined) {
        $publisher = Join-Path $package 'Publish-WorkbenchSnapshot.ps1'
        & (Join-Path $env:SystemRoot 'System32\WindowsPowerShell\v1.0\powershell.exe') -NoProfile -ExecutionPolicy Bypass -File $publisher -Source $db
        Assert-Release ($LASTEXITCODE -eq 0) 'packaged manual snapshot publishes actual live SQLite without development dependencies'
        $health = Request-Release 'GET' '/api/v1/cognition/health'
        Assert-Release ($health.status -eq 200) 'real published snapshot health'
        $snapshotHealth = $health.body | ConvertFrom-Json
        Assert-Release ($snapshotHealth.status -eq 'ok' -and $snapshotHealth.snapshot_at.Length -gt 0) 'actual snapshot health identifies its publication'
        $today = Request-Release 'GET' '/api/v1/cognition/today'
        Assert-Release ($today.status -eq 200) 'actual published cognition today read'
        $unchanged = Request-Release 'GET' '/api/v1/material-domains'
        Assert-Release ($unchanged.status -eq 200 -and $unchanged.body.Contains($created.domain.id)) 'read-only snapshot leaves actual application domain available'
    }
    if ($CognitionSource -ne '') { Test-ActualCognition }
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
    if ($RequireCombined) {
        $restoredSnapshot = Request-Release 'GET' '/api/v1/cognition/health'
        Assert-Release ($restoredSnapshot.status -eq 200) 'actual published snapshot remains readable after application restart'
    }
    if ($CognitionSource -ne '') {
        $restoredFocus = Request-Release 'GET' '/api/v1/cognition/focus'
        Assert-Release ($restoredFocus.status -eq 200) 'actual public focus survives process restart'
        $topics = ($restoredFocus.body | ConvertFrom-Json).topics
        foreach ($metric in @($cognitionExpected.values | Where-Object { $_.metric -in @('attention', 'momentum', 'confidence') })) {
            $rows = @($topics | Where-Object { $_.topic -ceq $metric.topic })
            Assert-Release ($rows.Count -eq 1 -and [Math]::Abs([double]$rows[0].($metric.metric) - [double]$metric.value) -le 1e-12) ('actual SQL state survives restart ' + $metric.metric)
        }
    }
    Stop-Release 'restart'
    $summary = @{ result = 'PASS'; source_commit = $(if ($Source -eq '') { $null } else { $Source }); scope = $(if ($CognitionSource -ne '') { 'candidate-package-cognition-runtime' } elseif ($RequireCombined) { 'combined-package-runtime' } else { 'candidate-package-runtime' }); zip = [IO.Path]::GetFullPath($Zip); output = $output; checks = $checks; limitations = @('No default browser or native controller acceptance is claimed.', 'Optional summarize/media dependencies were not invoked.', 'Restart exercised owned Windows process termination and SQLite recovery.') }
    $summary | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath (Join-Path $output 'result.json') -Encoding UTF8
    Write-Output ('VERIFICATION=' + $output)
} finally {
    if ($null -ne $archive) { $archive.Dispose() }
    if ($null -ne $process) { Stop-Release 'failed' }
    if ($null -ne $client) { $client.Dispose() }
    Stop-Transcript | Out-Null
}
