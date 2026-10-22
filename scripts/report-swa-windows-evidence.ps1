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
