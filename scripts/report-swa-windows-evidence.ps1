#Requires -Version 7.2
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$EvidenceRoot,
    [Parameter(Mandatory)][ValidatePattern('^[0-9a-f]{40}$')][string]$ExpectedRevision
)

$ErrorActionPreference = 'Stop'
if (-not [IO.Path]::IsPathFullyQualified($EvidenceRoot)) { throw 'An absolute downloaded evidence directory is required' }
$source = (Get-Content -LiteralPath (Join-Path $EvidenceRoot 'SOURCE') -Raw).Trim()
if ($source -cne $ExpectedRevision) { throw 'The prior artifact source revision differs' }
$result = Get-Content -LiteralPath (Join-Path $EvidenceRoot 'result.json') -Raw | ConvertFrom-Json
if ($result.status -cne 'FAIL') { throw 'This is not an original failed artifact' }
if ($source -ceq '11717767f82a8a4070351ff74674aa0d268f79b2') {
    if ($result.failed_stage -cne 'service-smoke') { throw 'The original 117 candidate failed at a different stage' }
    function Format-ServiceError([string]$Text) {
        if ($Text -match '(?i)(authorization|\bbearer\b|\bbasic\s+[A-Za-z0-9+/=]+|token|nonce|api[-_]?key|secret|password|cookie|credential)') { return '[sensitive log line omitted]' }
        $safeText = [regex]::Replace($Text, '(?i)https?://\S+', '[URL omitted]')
        $safeText = [regex]::Replace($safeText, '(?i)[a-z]:[\\/]Users[\\/][^\\/\s"'']+', '[CI user home]')
        $safeText = [regex]::Replace($safeText, '/(?:home|Users)/[^/\s"'']+', '[user home]')
        $safeText = [regex]::Replace($safeText, '(?i)\bS-\d-(?:\d+-){1,14}\d+\b', '[SID omitted]')
        $safeText = [regex]::Replace($safeText, '(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b', '[UUID omitted]')
        $safeText = [regex]::Replace($safeText, '(?i)\b[0-9a-f]{64}\b', '[digest or secret omitted]')
        $safeText = [regex]::Replace($safeText, '[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]', '')
        if ($safeText.Length -gt 500) { $safeText = $safeText.Substring(0, 500) + ' [line truncated]' }
        return $safeText
    }
    $exits = [ordered]@{}
    foreach ($stage in @('frozen-install', 'go-download', 'go-verify', 'build', 'extracted-check', 'install', 'installed-check')) {
        $code = (Get-Content -LiteralPath (Join-Path $EvidenceRoot ('logs/' + $stage + '.exit.txt')) -Raw).Trim()
        if ($code -cne '0') { throw 'An original preceding service operation did not succeed' }
        $exits[$stage] = 0
    }
    $acl = Get-Content -LiteralPath (Join-Path $EvidenceRoot 'logs/installed-acl.json') -Raw | ConvertFrom-Json
    if ($acl.owner_matches_current_user -isnot [bool] -or $acl.access_rules_protected -isnot [bool]) { throw 'The actual original ACL booleans are unavailable' }
    $helperExit = 'UNKNOWN'
    $helperExitFile = Join-Path $EvidenceRoot 'logs/service-smoke.exit.txt'
    if (Test-Path -LiteralPath $helperExitFile -PathType Leaf) {
        $helperExitText = (Get-Content -LiteralPath $helperExitFile -Raw).Trim()
        if ($helperExitText -cnotmatch '^-?[0-9]{1,9}$') { throw 'The actual service helper exit is invalid' }
        $helperExit = [int]$helperExitText
    }
    $mode = $null
    $modeFile = Join-Path $EvidenceRoot 'smoke/readiness-eof/result.json'
    if (Test-Path -LiteralPath $modeFile -PathType Leaf) { $mode = Get-Content -LiteralPath $modeFile -Raw | ConvertFrom-Json }
    $actual = $null
    $actualFile = Join-Path $EvidenceRoot 'smoke/readiness-eof/actual-observation.json'
    if (Test-Path -LiteralPath $actualFile -PathType Leaf) { $actual = Get-Content -LiteralPath $actualFile -Raw | ConvertFrom-Json }
    $latest = $null
    $latestFile = Join-Path $EvidenceRoot 'smoke/readiness-eof/latest-actual-observation.json'
    if (Test-Path -LiteralPath $latestFile -PathType Leaf) { $latest = Get-Content -LiteralPath $latestFile -Raw | ConvertFrom-Json }
    $observation = [ordered]@{
        diagnostic_only = $true; prior_run = '39778043411'; source_revision = $source; original_status = 'FAIL'; failed_stage = 'service-smoke'; recorded_preceding_exits = $exits
        actual_installed_acl = [ordered]@{ owner_matches_current_user = $acl.owner_matches_current_user; access_rules_protected = $acl.access_rules_protected }
        service_helper_exit = $helperExit
        mode_result_present = $null -ne $mode
        mode_status = if ($mode.status -cin @('UNKNOWN', 'OBSERVED', 'INCONCLUSIVE', 'AWAITING_ASSERTIONS')) { $mode.status } else { 'UNKNOWN' }
        serve_exit = if ($null -ne $actual -and $actual.PSObject.Properties.Name -contains 'exit') { $actual.exit } else { 'UNKNOWN' }
        unknown_process_observation_count = if ($actual.unknown_process_observations -is [array]) { $actual.unknown_process_observations.Count } elseif ($latest.unknown_process_observations -is [array]) { $latest.unknown_process_observations.Count } else { 'UNKNOWN' }
        exit_observation_count = if ($actual.exit_observations -is [array]) { $actual.exit_observations.Count } else { 'UNKNOWN' }
        cim_process_exit = if ($null -ne $actual.actual_cim_process_exit -and $actual.actual_cim_process_exit.PSObject.Properties.Name -contains 'exit') { $actual.actual_cim_process_exit.exit } else { 'UNKNOWN' }
        cim_all_exited = if ($actual.actual_cim_report.all_exited -is [bool]) { $actual.actual_cim_report.all_exited } else { 'UNKNOWN' }
        ready_service_observation_present = Test-Path -LiteralPath (Join-Path $EvidenceRoot 'smoke/readiness-eof/ready-service-observation.json') -PathType Leaf
        unsettled_observation_present = Test-Path -LiteralPath (Join-Path $EvidenceRoot 'smoke/readiness-eof/unsettled.json') -PathType Leaf
        complete_service_acceptance = 'FAIL'; model_calls = 0
    }
    $facts = ($observation | ConvertTo-Json -Depth 7 -Compress).Replace('%', '%25').Replace("`r", '%0D').Replace("`n", '%0A')
    Write-Host ('::warning title=Retained Windows service failure facts::' + $facts)
    $errorLines = [Collections.Generic.List[string]]::new()
    foreach ($relative in @('logs/service-smoke.log', 'smoke/readiness-eof/stderr.log')) {
        $log = Join-Path $EvidenceRoot $relative
        if (Test-Path -LiteralPath $log -PathType Leaf) {
            foreach ($line in (Get-Content -LiteralPath $log | Where-Object { $_ -match '(?i)(\berror\b|\bfatal\b|ERR_[A-Z_]+|失败|异常|拒绝|\bunknown\b|\bactual:|\bexpected:|\boperator:|tests[\\/]swa-cli-headless[\\/]serve-lifecycle\.mjs:[0-9]+|scripts[\\/][A-Za-z0-9-]+\.(?:mjs|ps1):[0-9]+)' } | Select-Object -Last 60)) { $errorLines.Add((Format-ServiceError $line)) }
        }
    }
    if ($latest.stderr -is [string]) {
        foreach ($line in ($latest.stderr -split '\r?\n' | Where-Object { $_ -match '(?i)(\berror\b|\bfatal\b|ERR_[A-Z_]+|失败|异常|拒绝|\bunknown\b)' } | Select-Object -Last 20)) { $errorLines.Add((Format-ServiceError $line)) }
    }
    $stdoutFile = Join-Path $EvidenceRoot 'smoke/readiness-eof/stdout.json'
    if (Test-Path -LiteralPath $stdoutFile -PathType Leaf) {
        $stdout = Get-Content -LiteralPath $stdoutFile -Raw
        if (Test-Json -Json $stdout -ErrorAction SilentlyContinue) {
            $envelope = $stdout | ConvertFrom-Json
            if ($envelope.error.code -is [string] -and $envelope.error.code -cmatch '^[a-z_]{1,80}$') { $errorLines.Add('Original serve error code: ' + $envelope.error.code) }
            if ($envelope.error.message -is [string]) { $errorLines.Add((Format-ServiceError $envelope.error.message)) }
        } else { $errorLines.Add('[original stdout is not a complete JSON envelope; no raw output exported]') }
    }
    $message = 'Prior run39778043411; source=' + $source + '; filtered original service errors:' + "`n" + ($errorLines -join "`n")
    if ($errorLines.Count -eq 0) { $message += '[no whitelisted error line found; root cause remains unknown]' }
    if ($message.Length -gt 8000) { $message = $message.Substring(0, 8000) + ' [original artifact retained]' }
    $escaped = $message.Replace('%', '%25').Replace("`r", '%0D').Replace("`n", '%0A')
    Write-Host ('::warning title=Retained Windows filtered service errors::' + $escaped)
    Write-Output 'Only the retained artifact was read; no candidate, native build, service or model was run.'
    return
}
if ($source -ceq 'c7ab4a7479182f29f990cd649accac193514d169') {
    if ($result.failed_stage -cne 'installed-acl') { throw 'The original c7 installation failed at a different stage' }
    $exits = [ordered]@{}
    foreach ($stage in @('frozen-install', 'go-download', 'go-verify', 'build', 'extracted-check', 'install', 'installed-check')) {
        $code = (Get-Content -LiteralPath (Join-Path $EvidenceRoot ('logs/' + $stage + '.exit.txt')) -Raw).Trim()
        if ($code -cne '0') { throw 'An original preceding installation operation did not succeed' }
        $exits[$stage] = 0
    }
    $observation = [ordered]@{ diagnostic_only = $true; prior_run = '39773729265'; source_revision = $source; original_status = 'FAIL'; failed_stage = 'installed-acl'; recorded_preceding_exits = $exits; original_owner_match = 'UNKNOWN'; original_access_rules_protected = 'UNKNOWN'; original_acl_values_recorded = $false; complete_installation_acceptance = 'FAIL'; model_calls = 0 }
    $message = ($observation | ConvertTo-Json -Depth 6 -Compress).Replace('%', '%25').Replace("`r", '%0D').Replace("`n", '%0A')
    Write-Host ('::warning title=Retained Windows installed ACL failure facts::' + $message)
    Write-Output 'The original artifact has no owner/protection observation; the failed compound assertion remains unresolved.'
    return
}
if ($source -cne '82e1b09503486c14a9dececb5e3673827894c120' -or $result.failed_stage -cne 'build') { throw 'This is not the expected original failed build artifact' }
$buildLog = Join-Path $EvidenceRoot 'logs/build.log'
if (-not (Test-Path -LiteralPath $buildLog -PathType Leaf)) { throw 'The original build log is missing' }
$selected = @(Get-Content -LiteralPath $buildLog | Where-Object { $_ -match '(?i)(\berror\b|\bfatal\b|ERR_[A-Z_]+|\bMSB[0-9]+\b|\bLNK[0-9]+\b|\.go:[0-9]+:|scripts[\\/]build-swa-(?:package|native)\.mjs:[0-9]+:[0-9]+)' } | Select-Object -Last 100)
$safe = [Collections.Generic.List[string]]::new()
foreach ($line in $selected) {
    if ($line -match '(?i)(authorization|\bbearer\b|\bbasic\s+[A-Za-z0-9+/=]+|token|api[-_]?key|client[-_]?secret|password|cookie|credential)') {
        $safe.Add('[sensitive log line omitted]')
        continue
    }
    $sanitized = [regex]::Replace($line, '(?i)https?://\S+', '[URL omitted]')
    $sanitized = [regex]::Replace($sanitized, '(?i)[a-z]:[\\/]Users[\\/][^\\/\s"'']+', '[CI user home]')
    $sanitized = [regex]::Replace($sanitized, '/(?:home|Users)/[^/\s"'']+', '[user home]')
    $sanitized = [regex]::Replace($sanitized, '[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]', '')
    if ($sanitized.Length -gt 500) { $sanitized = $sanitized.Substring(0, 500) + ' [line truncated]' }
    $safe.Add($sanitized)
}
$message = 'Prior run39769388709; source=' + $source + '; phase=build; filtered original errors:' + "`n" + ($safe -join "`n")
if (-not $safe.Count) { $message += '[no whitelisted error line found; root cause remains unknown]' }
if ($message.Length -gt 8000) { $message = $message.Substring(0, 8000) + ' [original artifact retained]' }
$escaped = $message.Replace('%', '%25').Replace("`r", '%0D').Replace("`n", '%0A')
Write-Host ('::warning title=Prior Windows build diagnostic only::' + $escaped)
Write-Output 'Read-only artifact diagnosis completed; no build, installation, service or model was run.'
