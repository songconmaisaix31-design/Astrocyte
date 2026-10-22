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
if ($result.status -cne 'FAIL' -or $result.failed_stage -cne 'build') { throw 'This is not the expected original failed build artifact' }
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
