#Requires -Version 7.2
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$EvidenceRoot,
    [Parameter(Mandatory)][ValidatePattern('^[0-9a-f]{40}$')][string]$ExpectedRevision,
    [ValidateSet('Short', 'Closure')][string]$ControllerOpenSection
)

$ErrorActionPreference = 'Stop'
if (-not [IO.Path]::IsPathFullyQualified($EvidenceRoot)) { throw 'An absolute downloaded evidence directory is required' }
if ($ExpectedRevision -ceq '1b8828d6add9eaab6ed59f445c2ee2c11c4f6511') {
    if ($ControllerOpenSection -cnotin @('Short', 'Closure')) { throw 'The fixed D10 section is required' }
    $file = Get-Item -LiteralPath (Join-Path $EvidenceRoot 'result.json') -Force
    if ($file -isnot [IO.FileInfo] -or ($file.Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0 -or $file.Length -gt 1048576) { throw 'The saved D10 result must be a bounded ordinary file' }
    $saved = Get-Content -LiteralPath $file.FullName -Raw | ConvertFrom-Json
    $expected = @{
        schema = 'swa.windows-controller-open-development.v1'; status = 'FAIL'; phase = 'short-identify-then-controller-open'
        scope = 'development_composition_not_candidate'; observation_revision = $ExpectedRevision; go_source_revision = $ExpectedRevision
        source_revision = '7ff3a7d7d7e029c2fc9e2ae1fb6b81da8e0d4f8f'; runtime_source_revision = '7ff3a7d7d7e029c2fc9e2ae1fb6b81da8e0d4f8f'
        input_revision = '1d3006fe0df0cb04b7f0de1e4563264bd0bfd9da'; native_input_revision = '1d3006fe0df0cb04b7f0de1e4563264bd0bfd9da'; task_create = 'NOT_RUN'
    }
    foreach ($name in $expected.Keys) { if ($saved.$name -cne $expected[$name]) { throw ('The fixed D10 source or result differs: ' + $name) } }
    if (($saved.model_calls -isnot [int] -and $saved.model_calls -isnot [long]) -or ($saved.worker_calls -isnot [int] -and $saved.worker_calls -isnot [long]) -or $saved.model_calls -cne 0 -or $saved.worker_calls -cne 0) { throw 'The fixed D10 no-model scope differs' }
    function Get-D10Value($Object, [string]$Field, [string]$Type, [bool]$Nullable = $false, [string[]]$Allowed = @()) {
        if ($null -eq $Object -or $Object.PSObject.Properties.Name -notcontains $Field) { return 'UNKNOWN' }
        $value = $Object.$Field
        if ($null -eq $value) { if ($Nullable) { return $null }; return 'UNKNOWN' }
        if ($Type -ceq 'Boolean') { if ($value -is [bool]) { return $value }; return 'UNKNOWN' }
        if ($Type -ceq 'Tag') { if ($value -is [string] -and $value -cmatch '^[A-Za-z][A-Za-z0-9_.-]{0,63}$') { return $value }; return 'UNKNOWN' }
        if ($Type -ceq 'Enum') { if ($value -is [string] -and $Allowed -ccontains $value) { return $value }; return 'UNKNOWN' }
        if ($value -isnot [int] -and $value -isnot [long] -and $value -isnot [double] -and $value -isnot [decimal]) { return 'UNKNOWN' }
        if (-not [double]::IsFinite([double]$value) -or [math]::Abs([double]$value) -gt 9007199254740991) { return 'UNKNOWN' }
        if ($Type -ceq 'Number') { return $value }
        if ([math]::Floor([double]$value) -ne $value) { return 'UNKNOWN' }
        if ($Type -ceq 'Integer') { return $value }
        if ($value -lt 0 -or ($Type -ceq 'PID' -and $value -gt [uint32]::MaxValue)) { return 'UNKNOWN' }
        return $value
    }
    function Get-D10Event($Event) {
        if ($null -eq $Event) { return $null }
        if ($Event -isnot [pscustomobject]) { return 'UNKNOWN' }
        $out = [ordered]@{}
        $out.lane = Get-D10Value $Event lane Enum $false @('tty', 'pipe')
        $out.stage = Get-D10Value $Event stage Enum $false @('create', 'spawn_identity', 'native_inputs', 'write', 'reply', 'identity', 'eof', 'absence', 'cleanup', 'exit')
        $out.boundary = Get-D10Value $Event boundary Enum $false @('start', 'complete', 'observed')
        foreach ($field in @('packet_index', 'packet_bytes')) { $out[$field] = Get-D10Value $Event $field Count $true }
        $out.observed_at_ms = Get-D10Value $Event observed_at_ms Count
        foreach ($field in @('monotonic_ms', 'elapsed_ms', 'deadline_remaining_ms')) { $out[$field] = Get-D10Value $Event $field Number }
        return $out
    }
    function Get-D10Window($Window) {
        if ($Window -isnot [pscustomobject]) { return 'UNKNOWN' }
        $out = [ordered]@{}
        foreach ($field in @('start_at_ms', 'end_at_ms')) { $out[$field] = Get-D10Value $Window $field Count }
        foreach ($field in @('start_monotonic_ms', 'end_monotonic_ms')) { $out[$field] = Get-D10Value $Window $field Number }
        foreach ($field in @('root_exit_observed_before_query', 'root_exit_observed_after_query', 'root_close_observed_before_query', 'root_close_observed_after_query')) { $out[$field] = Get-D10Value $Window $field Boolean }
        return $out
    }
    function Get-D10OptionalEvent($Object, [string]$Field) {
        if ($null -eq $Object -or $Object.PSObject.Properties.Name -notcontains $Field) { return 'UNKNOWN' }
        return (Get-D10Event $Object.$Field)
    }
    function Get-D10ProcessEvent($Object, [string]$Field) {
        if ($null -eq $Object -or $Object.PSObject.Properties.Name -notcontains $Field) { return 'UNKNOWN' }
        $event = $Object.$Field
        if ($null -eq $event) { return $null }
        return [ordered]@{ exit = Get-D10Value $event exit Integer $true; signal = Get-D10Value $event signal Tag $true
            observed_at_ms = Get-D10Value $event observed_at_ms Count; monotonic_ms = Get-D10Value $event monotonic_ms Number }
    }
    function Write-D10Observation([string]$Title, $Observation) {
        $text = $Observation | ConvertTo-Json -Depth 12 -Compress
        if ([Text.Encoding]::UTF8.GetByteCount($text) -gt 3500) { throw 'The typed retained D10 observation exceeds its fixed bound' }
        Write-Host ('::warning title=' + $Title + '::' + $text.Replace('%', '%25').Replace("`r", '%0D').Replace("`n", '%0A'))
    }
    $identity = [ordered]@{ diagnostic_only = $true; original_run = '39952528075'; artifact_id = '12500479586'; original_revision = $ExpectedRevision
        section = $ControllerOpenSection; original_status = 'FAIL'; complete_service_acceptance = 'FAIL'; model_calls = 0; worker_calls = 0 }
    $short = $saved.short_transport_diagnostic
    if ($ControllerOpenSection -ceq 'Short') {
        $lanes = [ordered]@{}
        foreach ($name in @('tty', 'pipe')) {
            $lane = $short.lanes.$name
            if ($null -eq $lane) { $lanes[$name] = if ($null -ne $short.lanes -and $short.lanes.PSObject.Properties.Name -contains $name) { $null } else { 'UNKNOWN' }; continue }
            $row = [ordered]@{ status = Get-D10Value $lane status Enum $false @('PASS', 'FAIL'); natural_eof = Get-D10Value $lane natural_eof Boolean
                forced_cleanup = Get-D10Value $lane forced_cleanup Boolean; reply_count = Get-D10Value $lane reply_count Count
                raw_output_bytes = if ($lane.raw_output_bytes -ceq 'NOT_RECORDED') { 'NOT_RECORDED' } else { Get-D10Value $lane raw_output_bytes Count }; failure_phase = Get-D10OptionalEvent $lane failure_phase
                exit = if ($null -ne $lane.exit) { [ordered]@{ code = Get-D10Value $lane.exit code Integer $true; signal = Get-D10Value $lane.exit signal Tag $true } } elseif ($lane.PSObject.Properties.Name -contains 'exit') { $null } else { 'UNKNOWN' }
                framing = if ($lane.framing -ceq 'NOT_RECORDED') { 'NOT_RECORDED' } else { [ordered]@{ valid_reply_records = Get-D10Value $lane.framing valid_reply_records Count; schema_rejected_records = Get-D10Value $lane.framing schema_rejected_records Count } }
                framer_rejections = 'NOT_RECORDED' }
            if ($lane.framer_rejections -is [array]) {
                if ($lane.framer_rejections.Count -gt 3) { throw 'The saved D10 framer count exceeds its source bound' }
                $row.framer_rejections = @($lane.framer_rejections | ForEach-Object { [ordered]@{ kind = Get-D10Value $_ kind Enum $false @('invalid-json', 'line-too-long', 'UNKNOWN'); count = Get-D10Value $_ count Count } })
            }
            $lanes[$name] = $row
        }
        $raw = [ordered]@{ status = 'UNKNOWN' }
        $summary = $saved.control.transport
        if ($summary.coverage -ceq 'explicit_short_identify_only' -and $summary.long_input -ceq 'NOT_RUN' -and $summary.rpc_mutations -ceq 'NOT_RUN' -and $summary.model_execution -ceq 'NOT_RUN') {
            $raw.status = Get-D10Value $short.raw status Enum $false @('OBSERVED', 'NOT_CAPTURED', 'UNKNOWN')
            if ($raw.status -ceq 'OBSERVED') {
                if ($short.raw.source -cne 'explicit_short_identify_only' -or $short.raw.head -isnot [string] -or $short.raw.tail -isnot [string]) { throw 'The saved D10 raw fragment source differs' }
                $bytes = [Text.Encoding]::UTF8.GetByteCount($short.raw.head) + [Text.Encoding]::UTF8.GetByteCount($short.raw.tail)
                if ($bytes -gt 512 -or $short.raw.fragment_limit_bytes -cne 512 -or $short.raw.fragment_bytes -cne $bytes) { throw 'The saved D10 raw fragments exceed their source bound' }
                $raw.source = 'explicit_short_identify_only'
                foreach ($field in @('original_bytes', 'projected_bytes', 'fragment_limit_bytes', 'fragment_bytes')) { $raw[$field] = Get-D10Value $short.raw $field Count }
                $raw.head = $short.raw.head; $raw.tail = $short.raw.tail; $raw.truncated = Get-D10Value $short.raw truncated Boolean
            }
        }
        Write-D10Observation 'Retained D10 short transport' ([ordered]@{ identity = $identity; status = Get-D10Value $short status Enum $false @('PASS', 'FAIL')
            cleanup_unknown = Get-D10Value $short cleanup_unknown Boolean; lanes = $lanes; raw_status = $raw.status; last_event = Get-D10OptionalEvent $short last_event
            trace_total = Get-D10Value $short trace_total Count; trace_truncated = Get-D10Value $short trace_truncated Boolean
            ready_event_total = if ($short.ready_event_total -ceq 'NOT_RECORDED') { 'NOT_RECORDED' } else { Get-D10Value $short ready_event_total Count }
            ready_events_status = if ($short.ready_events -is [array]) { 'OBSERVED' } elseif ($short.ready_events -ceq 'NOT_RECORDED') { 'NOT_RECORDED' } else { 'UNKNOWN' }
            ready_events_truncated = Get-D10Value $short ready_events_truncated Boolean })
        Write-D10Observation 'Retained D10 short raw fragment' ([ordered]@{ original_revision = $ExpectedRevision; raw = $raw })
        if ($short.ready_events -is [array]) {
            if ($short.ready_events.Count -gt 32) { throw 'The saved D10 ready events exceed their source bound' }
            for ($offset = 0; $offset -lt $short.ready_events.Count; $offset += 8) {
                $rows = @()
                foreach ($event in $short.ready_events[$offset..([math]::Min($offset + 7, $short.ready_events.Count - 1))]) {
                    $row = [ordered]@{ event = Get-D10Value $event event Enum $false @('startup-command-delivery', 'startup-command-stage-failed', 'shell-ready-timeout', 'shell-ready-wrapper-replaced')
                        observed_at_ms = Get-D10Value $event observed_at_ms Count; monotonic_ms = Get-D10Value $event monotonic_ms Number }
                    if ($row.event -ceq 'startup-command-delivery') {
                        foreach ($field in @('written', 'has_command', 'via_shell_args', 'queued_by_shell_ready_barrier')) { $row[$field] = Get-D10Value $event $field Boolean }
                        $row.command_length = Get-D10Value $event command_length Count
                    } elseif ($row.event -ceq 'shell-ready-timeout') { $row.timeout_ms = Get-D10Value $event timeout_ms Count }
                    elseif ($row.event -ceq 'startup-command-stage-failed') { $row.reason = Get-D10Value $event reason Tag }
                    $rows += $row
                }
                Write-D10Observation 'Retained D10 ready events' ([ordered]@{ original_revision = $ExpectedRevision; offset = $offset; retained_count = $short.ready_events.Count
                    total_count = Get-D10Value $short ready_event_total Count; source_truncated = Get-D10Value $short ready_events_truncated Boolean; events = $rows })
            }
        }
        return
    }
    $unknown = $saved.unknown_process_observation_projection
    $reasons = @('invalid_native_process_row', 'current_root_identity_unverified', 'root_in_descendant_branch', 'parent_incarnation_unverified', 'process_identity_not_unique', 'creation_time_unavailable', 'parent_incarnation_changed', 'UNKNOWN')
    $reasonCounts = 'UNKNOWN'; $queryCounts = 'UNKNOWN'
    if ($unknown.reason_counts -is [array]) {
        if ($unknown.reason_counts.Count -gt 8) { throw 'The saved D10 reason count exceeds its source bound' }
        $reasonCounts = @($unknown.reason_counts | ForEach-Object { [ordered]@{ reason = Get-D10Value $_ reason Enum $false $reasons; count = Get-D10Value $_ count Count } })
    }
    function Get-D10Query($Query) {
        if ($Query -isnot [pscustomobject]) { return 'UNKNOWN' }
        return [ordered]@{ version = Get-D10Value $Query version Count; api_stage = Get-D10Value $Query api_stage Enum $false @('not_requested', 'OpenProcess', 'GetProcessTimes', 'invalid_filetime', 'complete')
            win32_error = Get-D10Value $Query win32_error Count; close_error = Get-D10Value $Query close_error Count }
    }
    if ($unknown.native_query_counts -is [array]) {
        if ($unknown.native_query_counts.Count -gt 8) { throw 'The saved D10 query count exceeds its source bound' }
        $queryCounts = @($unknown.native_query_counts | ForEach-Object { [ordered]@{ query = Get-D10Query $_.query; count = Get-D10Value $_ count Count } })
    }
    $remaining = 'UNKNOWN'
    if ($saved.service.native_remaining -is [array]) {
        if ($saved.service.native_remaining.Count -gt 2) { throw 'The saved D10 native exit observation count differs' }
        $remaining = @($saved.service.native_remaining | ForEach-Object { Get-D10Value ([pscustomobject]@{ count = $_ }) count Count })
    }
    Write-D10Observation 'Retained D10 closure counts' ([ordered]@{ identity = $identity; original_fixture_exit = Get-D10Value $saved fixture_exit Integer
        ownership_closure_incomplete = Get-D10Value $unknown ownership_closure_incomplete Boolean
        serve_root = [ordered]@{ pid = Get-D10Value $unknown.serve_root pid PID; creation_time_ms = Get-D10Value $unknown.serve_root creation_time_ms Count }
        root_process_events = [ordered]@{ exit = Get-D10ProcessEvent $unknown.root_process_events exit; close = Get-D10ProcessEvent $unknown.root_process_events close }
        latch_inputs = [ordered]@{ root_creation_time_readable = Get-D10Value $unknown.latch_inputs root_creation_time_readable Boolean
            controller_input_ownership_incomplete = Get-D10Value $unknown.latch_inputs controller_input_ownership_incomplete Boolean
            producer_ownership_incomplete = Get-D10Value $unknown.latch_inputs producer_ownership_incomplete Boolean
            readiness_check_error = if ($null -ne $unknown.latch_inputs.readiness_check_error) { [ordered]@{ name = Get-D10Value $unknown.latch_inputs.readiness_check_error name Tag $true; code = Get-D10Value $unknown.latch_inputs.readiness_check_error code Tag $true } } elseif ($null -ne $unknown.latch_inputs -and $unknown.latch_inputs.PSObject.Properties.Name -contains 'readiness_check_error') { $null } else { 'UNKNOWN' } }
        unknown_item_count = Get-D10Value $unknown unknown_item_count Count; reason_counts = $reasonCounts
        native_query_type_count = Get-D10Value $unknown native_query_type_count Count; native_query_counts_truncated = Get-D10Value $unknown native_query_counts_truncated Boolean
        identity_group_count = Get-D10Value $unknown identity_group_count Count; samples_truncated = Get-D10Value $unknown samples_truncated Boolean
        identity_history_is_not_a_single_snapshot_chain = Get-D10Value $unknown identity_history_is_not_a_single_snapshot_chain Boolean
        saved_projection_utf8_bytes = if ($null -ne $unknown) { [Text.Encoding]::UTF8.GetByteCount(($unknown | ConvertTo-Json -Depth 16 -Compress)) } else { 'UNKNOWN' }; original_annotation_bound = 3500
        service = [ordered]@{ exit = Get-D10Value $saved.service exit Integer $true; signal = Get-D10Value $saved.service signal Tag $true
            unknown_process_observation_count = Get-D10Value $saved.service unknown_process_observation_count Count; native_remaining = $remaining
            cim_exit = Get-D10Value $saved.service cim_exit Integer $true; cim_all_exited = Get-D10Value $saved.service cim_all_exited Boolean } })
    Write-D10Observation 'Retained D10 native query counts' ([ordered]@{ original_revision = $ExpectedRevision
        total_type_count = Get-D10Value $unknown native_query_type_count Count; source_truncated = Get-D10Value $unknown native_query_counts_truncated Boolean; counts = $queryCounts })
    if ($unknown.samples -is [array]) {
        if ($unknown.samples.Count -gt 8) { throw 'The saved D10 samples exceed their source bound' }
        for ($index = 0; $index -lt $unknown.samples.Count; $index++) {
            $sample = $unknown.samples[$index]
            $row = [ordered]@{ original_revision = $ExpectedRevision; index = $index; retained_count = $unknown.samples.Count
                total_group_count = Get-D10Value $unknown identity_group_count Count; source_truncated = Get-D10Value $unknown samples_truncated Boolean
                reason = Get-D10Value $sample reason Enum $false $reasons; pid = Get-D10Value $sample pid PID; ppid = Get-D10Value $sample ppid PID
                creation_time_ms = Get-D10Value $sample creation_time_ms Count; count = Get-D10Value $sample count Count
                source_anchor = [ordered]@{ kind = Get-D10Value $sample.source_anchor kind Enum $false @('serve', 'additional'); pid = Get-D10Value $sample.source_anchor pid PID; creation_time_ms = Get-D10Value $sample.source_anchor creation_time_ms Count }
                diagnostic_query = Get-D10Query $sample.diagnostic_query
                actual_role_history = Get-D10Value $sample actual_role_history Enum $false @('tty_root', 'tty_helper', 'pipe_helper', 'short_checker', 'UNKNOWN')
                source_anchor_role_history = Get-D10Value $sample source_anchor_role_history Enum $false @('tty_root', 'tty_helper', 'pipe_helper', 'short_checker', 'UNKNOWN')
                first_query_window = Get-D10Window $sample.first_query_window; last_query_window = Get-D10Window $sample.last_query_window
                matches_recorded_identity_history = Get-D10Value $sample matches_recorded_identity_history Boolean
                recorded_parent_identity_count = Get-D10Value $sample recorded_parent_identity_count Count
                recorded_parent_identities_truncated = Get-D10Value $sample recorded_parent_identities_truncated Boolean; recorded_parent_identities_history = 'UNKNOWN'
                identity_history_is_not_a_single_snapshot_chain = $true }
            foreach ($field in @('first_observed_at_ms', 'last_observed_at_ms', 'first_sample_sequence', 'last_sample_sequence')) { $row[$field] = Get-D10Value $sample $field Count }
            if ($sample.recorded_parent_identities_history -is [array]) {
                if ($sample.recorded_parent_identities_history.Count -gt 2) { throw 'The saved D10 parent history exceeds its source bound' }
                $row.recorded_parent_identities_history = @($sample.recorded_parent_identities_history | ForEach-Object { [ordered]@{ pid = Get-D10Value $_ pid PID; creation_time_ms = Get-D10Value $_ creation_time_ms Count } })
            }
            Write-D10Observation 'Retained D10 closure identity group' $row
        }
    }
    return
}
if ($PSBoundParameters.ContainsKey('ControllerOpenSection')) { throw 'The section parameter belongs only to the fixed D10 artifact' }
$source = (Get-Content -LiteralPath (Join-Path $EvidenceRoot 'SOURCE') -Raw).Trim()
if ($source -cne $ExpectedRevision) { throw 'The prior artifact source revision differs' }
$result = Get-Content -LiteralPath (Join-Path $EvidenceRoot 'result.json') -Raw | ConvertFrom-Json
if ($result.status -cne 'FAIL') { throw 'This is not an original failed artifact' }
if ($source -ceq '11717767f82a8a4070351ff74674aa0d268f79b2' -or $source -ceq '4aef552c63dd1ab5cd787fd2514d92a7815b943c') {
    $priorRun = if ($source -ceq '11717767f82a8a4070351ff74674aa0d268f79b2') { '39778043411' } else { '39846687508' }
    if ($result.failed_stage -cne 'service-smoke') { throw 'The original service candidate failed at a different stage' }
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
        diagnostic_only = $true; prior_run = $priorRun; source_revision = $source; original_status = 'FAIL'; failed_stage = 'service-smoke'; recorded_preceding_exits = $exits
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
    if ($source -ceq '4aef552c63dd1ab5cd787fd2514d92a7815b943c') {
        function Get-RetainedNumber($Value) {
            if (($Value -is [int] -or $Value -is [long] -or $Value -is [double] -or $Value -is [decimal]) -and [double]::IsFinite([double]$Value) -and $Value -ge 0 -and $Value -le 9007199254740991) { return $Value }
            return 'UNKNOWN'
        }
        function Get-RetainedBoolean($Value) {
            if ($Value -is [bool]) { return $Value }
            return 'UNKNOWN'
        }
        function Get-RetainedNullableText($Object, [string]$Field) {
            if ($null -eq $Object -or $Object.PSObject.Properties.Name -notcontains $Field) { return 'UNKNOWN' }
            if ($null -eq $Object.$Field) { return $null }
            if ($Object.$Field -is [string]) { return (Format-ServiceError $Object.$Field) }
            return 'UNKNOWN'
        }
        function Get-RetainedSnapshotChain($Snapshot, $RootPid, $RootCreation, $TargetPid, $TargetCreation) {
            $projection = [ordered]@{
                snapshot_source = 'latest-actual-observation.json/current_owned_processes'
                snapshot_scope = 'single-raw-ppid-closure-subset'
                snapshot_observed_at_ms = Get-RetainedNumber $Snapshot.observed_at_ms
                snapshot_retained_row_count = if ($Snapshot.current_owned_processes -is [array]) { $Snapshot.current_owned_processes.Count } else { 'UNKNOWN' }
                root_metadata_matches_actual_spawn = 'UNKNOWN'
                status = 'UNKNOWN'; stop = 'SNAPSHOT_UNAVAILABLE'; chain_order = 'ancestor-to-target'; chain = @(); chain_truncated = 'UNKNOWN'
            }
            $snapshotPid = Get-RetainedNumber $Snapshot.pid
            $snapshotCreation = Get-RetainedNumber $Snapshot.creation_time_ms
            if ($Snapshot.current_owned_processes -isnot [array] -or $projection.snapshot_observed_at_ms -eq 'UNKNOWN' -or $projection.snapshot_observed_at_ms -lt 1 -or [math]::Floor($projection.snapshot_observed_at_ms) -ne $projection.snapshot_observed_at_ms -or $RootPid -eq 'UNKNOWN' -or $RootCreation -eq 'UNKNOWN' -or $snapshotPid -eq 'UNKNOWN' -or $snapshotCreation -eq 'UNKNOWN') { return $projection }
            $projection.root_metadata_matches_actual_spawn = $snapshotPid -ceq $RootPid -and $snapshotCreation -ceq $RootCreation
            if (-not $projection.root_metadata_matches_actual_spawn) { $projection.stop = 'SNAPSHOT_ROOT_MISMATCH'; return $projection }
            if ($TargetPid -eq 'UNKNOWN' -or $null -eq $TargetCreation -or $TargetCreation -eq 'UNKNOWN' -or $TargetPid -lt 1 -or $TargetPid -gt [uint32]::MaxValue -or [math]::Floor($TargetPid) -ne $TargetPid -or $TargetCreation -lt 1 -or [math]::Floor($TargetCreation) -ne $TargetCreation) { $projection.stop = 'INVALID_TARGET_IDENTITY'; return $projection }
            $projection.chain_truncated = $false
            $seen = [Collections.Generic.HashSet[long]]::new()
            $currentPid = $TargetPid
            for ($hop = 0; $hop -lt 8; $hop++) {
                if (-not $seen.Add([long]$currentPid)) { $projection.stop = 'CYCLE'; break }
                $rows = @($Snapshot.current_owned_processes | Where-Object { (Get-RetainedNumber $_.pid) -ceq $currentPid })
                if ($rows.Count -eq 0) { $projection.stop = 'MISSING_SNAPSHOT_NODE'; break }
                if ($rows.Count -ne 1) { $projection.stop = 'DUPLICATE_SNAPSHOT_PID'; break }
                $rowPid = Get-RetainedNumber $rows[0].pid
                $parentPid = Get-RetainedNumber $rows[0].ppid
                $created = Get-RetainedNumber $rows[0].creationTimeMs
                if ($rowPid -eq 'UNKNOWN' -or $parentPid -eq 'UNKNOWN' -or $created -eq 'UNKNOWN' -or $rowPid -lt 1 -or $rowPid -gt [uint32]::MaxValue -or [math]::Floor($rowPid) -ne $rowPid -or $parentPid -gt [uint32]::MaxValue -or [math]::Floor($parentPid) -ne $parentPid -or $created -lt 1 -or [math]::Floor($created) -ne $created) { $projection.stop = 'INVALID_SNAPSHOT_IDENTITY'; break }
                $projection.chain += [ordered]@{ pid = $rowPid; ppid = $parentPid; creationTimeMs = $created }
                if ($hop -eq 0 -and $created -cne $TargetCreation) { $projection.stop = 'TARGET_INCARNATION_MISMATCH'; break }
                if ($rowPid -ceq $RootPid) {
                    if ($created -ceq $RootCreation) { $projection.status = 'OBSERVED'; $projection.stop = 'ACTUAL_SPAWN_ROOT_REACHED' } else { $projection.stop = 'ROOT_INCARNATION_MISMATCH' }
                    break
                }
                if ($parentPid -eq 0) { $projection.stop = 'PARENT_ZERO_BEFORE_ROOT'; break }
                $currentPid = $parentPid
                if ($hop -eq 7) { $projection.chain_truncated = $true; $projection.stop = 'NODE_LIMIT' }
            }
            $chain = @($projection.chain)
            [array]::Reverse($chain)
            $projection.chain = $chain
            return $projection
        }
        $cimExit = $actual.actual_cim_process_exit
        $cimReport = $actual.actual_cim_report
        $rootPid = Get-RetainedNumber $actual.pid
        if ($rootPid -ne 'UNKNOWN' -and ($rootPid -lt 1 -or $rootPid -gt [uint32]::MaxValue -or [math]::Floor($rootPid) -ne $rootPid)) { $rootPid = 'UNKNOWN' }
        $rootCreation = 'UNKNOWN'
        if ($rootPid -ne 'UNKNOWN' -and $actual.events -is [array]) {
            $spawnEvents = @($actual.events | Where-Object { $_.type -ceq 'actual-spawn' -and (Get-RetainedNumber $_.pid) -ceq $rootPid })
            if ($spawnEvents.Count -eq 1) { $rootCreation = Get-RetainedNumber $spawnEvents[0].creation_time_ms }
            if ($rootCreation -ne 'UNKNOWN' -and ($rootCreation -lt 1 -or [math]::Floor($rootCreation) -ne $rootCreation)) { $rootCreation = 'UNKNOWN' }
        }
        $readyObservation = $null
        if ($cimReport.observations -is [array] -and @($cimReport.observations | Where-Object { $_.known_identities_exited -is [bool] -and $_.known_identities_exited -eq $false }).Count -gt 0) {
            $readyFile = Join-Path $EvidenceRoot 'smoke/readiness-eof/ready-service-observation.json'
            if (Test-Path -LiteralPath $readyFile -PathType Leaf) {
                $readyObservation = Get-Content -LiteralPath $readyFile -Raw | ConvertFrom-Json
                if ($readyObservation.source_revision -cne $source) { throw 'The retained ready observation source revision differs' }
            }
        }
        $termination = 'UNKNOWN'
        if ($null -ne $cimExit -and $cimExit.PSObject.Properties.Name -contains 'actual_helper_termination') {
            $termination = if ($null -eq $cimExit.actual_helper_termination) { $null } elseif ($cimExit.actual_helper_termination -is [pscustomobject]) { [ordered]@{ signal_sent = Get-RetainedBoolean $cimExit.actual_helper_termination.signal_sent } } else { 'UNKNOWN' }
        }
        $queryFailure = 'UNKNOWN'
        if ($null -ne $cimReport -and $cimReport.PSObject.Properties.Name -contains 'query_failure') {
            if ($null -eq $cimReport.query_failure) { $queryFailure = $null } elseif ($cimReport.query_failure -is [pscustomobject]) {
                $failure = $cimReport.query_failure
                $queryFailure = [ordered]@{
                    query_index = Get-RetainedNumber $failure.query_index
                    query_succeeded = Get-RetainedBoolean $failure.query_succeeded
                    monotonic_ms = Get-RetainedNumber $failure.monotonic_ms
                    exception_type = if ($failure.exception_type -is [string] -and $failure.exception_type -cmatch '^[A-Za-z_][A-Za-z0-9_.+]{0,159}$') { $failure.exception_type } else { 'UNKNOWN' }
                    message = Get-RetainedNullableText $failure 'message'
                }
            }
        }
        $cimRows = 'UNKNOWN'
        if ($cimReport.observations -is [array]) {
            $cimRows = @($cimReport.observations | Select-Object -First 2 | ForEach-Object {
                $row = $_
                $projected = [ordered]@{
                    query_succeeded = Get-RetainedBoolean $row.query_succeeded
                    complete_result_count = Get-RetainedNumber $row.complete_result_count
                    invalid_pid_record_count = Get-RetainedNumber $row.invalid_pid_record_count
                    required_identity_fields_complete = Get-RetainedBoolean $row.required_identity_fields_complete
                    known_identities_exited = Get-RetainedBoolean $row.known_identities_exited
                    unknown_pids_absent = Get-RetainedBoolean $row.unknown_pids_absent
                    monotonic_ms = Get-RetainedNumber $row.monotonic_ms
                    actual_match_count = if ($row.actual_matches -is [array]) { $row.actual_matches.Count } else { 'UNKNOWN' }
                }
                if ($row.known_identities_exited -is [bool] -and $row.known_identities_exited -eq $false) {
                    $projected.actual_matches = 'UNKNOWN'
                    $projected.actual_matches_truncated = 'UNKNOWN'
                    $projected.observed_processes_same_pid = 'UNKNOWN'
                    $projected.observed_processes_same_pid_count = 'UNKNOWN'
                    $projected.observed_processes_same_pid_truncated = 'UNKNOWN'
                    if ($row.actual_matches -is [array]) {
                        $matches = @($row.actual_matches | Select-Object -First 8)
                        $projected.actual_matches = @($matches | ForEach-Object {
                            $matchPid = Get-RetainedNumber $_.pid
                            $matchCreation = if ($_.PSObject.Properties.Name -contains 'creation_time_ms' -and $null -eq $_.creation_time_ms) { $null } else { Get-RetainedNumber $_.creation_time_ms }
                            $identity = [ordered]@{ pid = $matchPid; creation_time_ms = $matchCreation; ready_identity_match_count = 'UNKNOWN'; ready_observed_role = 'UNKNOWN'; ready_observed_parent_pid = 'UNKNOWN'; ready_observed_named_job_membership = 'UNKNOWN' }
                            $identity.single_snapshot_raw_ppid_chain = Get-RetainedSnapshotChain $latest $rootPid $rootCreation $matchPid $matchCreation
                            if ($matchPid -ne 'UNKNOWN' -and $matchPid -gt 0 -and $matchPid -le [uint32]::MaxValue -and [math]::Floor($matchPid) -eq $matchPid -and $null -ne $matchCreation -and $matchCreation -ne 'UNKNOWN' -and $matchCreation -gt 0 -and [math]::Floor($matchCreation) -eq $matchCreation -and $readyObservation.members -is [array]) {
                                $readyMatches = @($readyObservation.members | Where-Object { (Get-RetainedNumber $_.pid) -ceq $matchPid -and (Get-RetainedNumber $_.creation_time_ms) -ceq $matchCreation })
                                $identity.ready_identity_match_count = $readyMatches.Count
                                if ($readyMatches.Count -eq 1) {
                                    $member = $readyMatches[0]
                                    $identity.ready_observed_role = if ($member.role -is [string] -and $member.role -cin @('serve', 'runtime', 'server', 'daemon')) { $member.role } else { 'UNKNOWN' }
                                    $identity.ready_observed_parent_pid = Get-RetainedNumber $member.parent_pid
                                    $identity.ready_observed_named_job_membership = Get-RetainedBoolean $member.actual_named_job_membership
                                }
                            }
                            $identity
                        })
                        $projected.actual_matches_truncated = $row.actual_matches.Count -gt 8
                        if ($actual.observed_processes -is [array]) {
                            $matchPids = @($matches | ForEach-Object { Get-RetainedNumber $_.pid } | Where-Object { $_ -ne 'UNKNOWN' })
                            $known = @($actual.observed_processes | Where-Object { $_.pid -in $matchPids })
                            $projected.observed_processes_same_pid = @($known | Select-Object -First 8 | ForEach-Object { [ordered]@{ pid = Get-RetainedNumber $_.pid; creationTimeMs = Get-RetainedNumber $_.creationTimeMs; ppid = Get-RetainedNumber $_.ppid; parent_pid = Get-RetainedNumber $_.parent_pid } })
                            $projected.observed_processes_same_pid_count = $known.Count
                            $projected.observed_processes_same_pid_truncated = $known.Count -gt 8
                        }
                    }
                }
                $projected
            })
        }
        $nativeRows = 'UNKNOWN'
        if ($actual.exit_observations -is [array]) {
            $nativeRows = @($actual.exit_observations | Select-Object -First 2 | ForEach-Object { [ordered]@{ observed_at_ms = Get-RetainedNumber $_.observed_at_ms; remaining_count = if ($_.remaining -is [array]) { $_.remaining.Count } else { 'UNKNOWN' } } })
        }
        $details = [ordered]@{
            diagnostic_only = $true; prior_run = $priorRun; source_revision = $source; complete_service_acceptance = 'FAIL'
            actual_pid = $rootPid
            root_creation_time_ms = $rootCreation
            actual_cim_process_exit = [ordered]@{
                exit = if ($null -ne $cimExit -and $cimExit.PSObject.Properties.Name -contains 'exit' -and $null -eq $cimExit.exit) { $null } else { Get-RetainedNumber $cimExit.exit }
                signal = Get-RetainedNullableText $cimExit 'signal'
                query_timed_out = Get-RetainedBoolean $cimExit.query_timed_out
                query_deadline_ms = Get-RetainedNumber $cimExit.query_deadline_ms
                completion_unknown = Get-RetainedBoolean $cimExit.completion_unknown
                actual_helper_termination = $termination
            }
            actual_cim_error = Get-RetainedNullableText $actual 'actual_cim_error'
            actual_cim_report = [ordered]@{
                schema_version = Get-RetainedNumber $cimReport.schema_version
                query_failure = $queryFailure
                observation_count = if ($cimReport.observations -is [array]) { $cimReport.observations.Count } else { 'UNKNOWN' }
                observations_truncated = if ($cimReport.observations -is [array]) { $cimReport.observations.Count -gt 2 } else { 'UNKNOWN' }
                observations = $cimRows
            }
            exit_observation_count = if ($actual.exit_observations -is [array]) { $actual.exit_observations.Count } else { 'UNKNOWN' }
            exit_observations_truncated = if ($actual.exit_observations -is [array]) { $actual.exit_observations.Count -gt 2 } else { 'UNKNOWN' }
            exit_observations = $nativeRows
        }
        $detailText = $details | ConvertTo-Json -Depth 12 -Compress
        if ($detailText.Length -gt 8000) { throw 'The retained CIM field projection exceeds its output limit' }
        $escapedDetails = $detailText.Replace('%', '%25').Replace("`r", '%0D').Replace("`n", '%0A')
        Write-Host ('::warning title=Retained Windows CIM exit predicates::' + $escapedDetails)
    }
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
    $message = 'Prior run' + $priorRun + '; source=' + $source + '; filtered original service errors:' + "`n" + ($errorLines -join "`n")
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
if ($source -ceq '50d099ddf9787a283573a477fb5f29017aef6c89') {
    if ($result.source_revision -cne $source -or $result.failed_stage -cne 'build' -or $result.platform -cne 'win32-x64' -or $result.edition -cne 'combined' -or ($result.model_calls -isnot [int] -and $result.model_calls -isnot [long]) -or $result.model_calls -ne 0) { throw 'The retained N46 artifact does not match the original combined build failure' }
    $exits = [ordered]@{}
    foreach ($stage in @('frozen-install', 'go-download', 'go-verify')) {
        $code = (Get-Content -LiteralPath (Join-Path $EvidenceRoot ('logs/' + $stage + '.exit.txt')) -Raw).Trim()
        if ($code -cne '0') { throw 'An original preceding build operation did not succeed' }
        $exits[$stage] = [int]$code
    }
    $buildExit = (Get-Content -LiteralPath (Join-Path $EvidenceRoot 'logs/build.exit.txt') -Raw).Trim()
    if ($buildExit -cne '1') { throw 'The retained original build exit differs' }
    $observation = [ordered]@{ diagnostic_only = $true; prior_run = '39900064207'; source_revision = $source; original_status = 'FAIL'; failed_stage = 'build'; platform = $result.platform; edition = $result.edition; recorded_preceding_exits = $exits; recorded_build_exit = [int]$buildExit; complete_candidate_acceptance = 'FAIL'; model_calls = $result.model_calls }
    $facts = ($observation | ConvertTo-Json -Depth 6 -Compress).Replace('%', '%25').Replace("`r", '%0D').Replace("`n", '%0A')
    Write-Host ('::warning title=Retained Windows N46 build failure facts::' + $facts)
} elseif ($source -cne '82e1b09503486c14a9dececb5e3673827894c120' -or $result.failed_stage -cne 'build') { throw 'This is not the expected original failed build artifact' }
$priorBuildRun = if ($source -ceq '50d099ddf9787a283573a477fb5f29017aef6c89') { '39900064207' } else { '39769388709' }
$buildLog = Join-Path $EvidenceRoot 'logs/build.log'
if (-not (Test-Path -LiteralPath $buildLog -PathType Leaf)) { throw 'The original build log is missing' }
if ($source -ceq '50d099ddf9787a283573a477fb5f29017aef6c89') {
    $webTools = @(Get-Content -LiteralPath $buildLog | Where-Object { $_ -match '(?i)(\berror\b|\bfatal\b|ERR_[A-Z_]+|\bMSB[0-9]+\b|\bLNK[0-9]+\b|\.go:[0-9]+:|scripts[\\/]build-swa-(?:package|native)\.mjs:[0-9]+:[0-9]+|\bELIFECYCLE\b|not recognized as an internal or external command|command not found|Cannot find module|\bMODULE_NOT_FOUND\b|\bENOENT\b|^\s*>\s*(?:@swa/web@\S+\s+build(?:\s|$)|vite\s+build(?:\s|$)))' } | ForEach-Object { '[logs/build.log] ' + $_ })
    $frozenTools = @(Get-Content -LiteralPath (Join-Path $EvidenceRoot 'logs/frozen-install.log') | Where-Object { $_ -match '(?i)(No projects matched the filters|Local package\.json exists, but node_modules missing)' } | ForEach-Object { '[logs/frozen-install.log] ' + $_ })
    $selected = @(@($webTools + $frozenTools) | Select-Object -Last 100)
} else {
    $selected = @(Get-Content -LiteralPath $buildLog | Where-Object { $_ -match '(?i)(\berror\b|\bfatal\b|ERR_[A-Z_]+|\bMSB[0-9]+\b|\bLNK[0-9]+\b|\.go:[0-9]+:|scripts[\\/]build-swa-(?:package|native)\.mjs:[0-9]+:[0-9]+)' } | Select-Object -Last 100)
}
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
$message = 'Prior run' + $priorBuildRun + '; source=' + $source + '; phase=build; filtered original errors:' + "`n" + ($safe -join "`n")
if (-not $safe.Count) { $message += '[no whitelisted error line found; root cause remains unknown]' }
if ($message.Length -gt 8000) { $message = $message.Substring(0, 8000) + ' [original artifact retained]' }
$escaped = $message.Replace('%', '%25').Replace("`r", '%0D').Replace("`n", '%0A')
Write-Host ('::warning title=Prior Windows build diagnostic only::' + $escaped)
Write-Output 'Read-only artifact diagnosis completed; no build, installation, service or model was run.'
