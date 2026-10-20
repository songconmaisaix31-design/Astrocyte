Set-StrictMode -Version Latest

function Read-SharedBytes([string]$Path) {
    $stream = [IO.FileStream]::new($Path,[IO.FileMode]::Open,[IO.FileAccess]::Read,([IO.FileShare]::ReadWrite -bor [IO.FileShare]::Delete))
    $buffer = [IO.MemoryStream]::new()
    try { $stream.CopyTo($buffer); return ,$buffer.ToArray() } finally { $buffer.Dispose(); $stream.Dispose() }
}

function Read-OrcaProcess([int]$ProcessId) {
    $process = Get-CimInstance Win32_Process -Filter "ProcessId = $ProcessId"
    if (-not $process -or -not $process.CreationDate) { throw "原 Orca 进程身份不可读取: $ProcessId" }
    return [ordered]@{ pid = [int]$process.ProcessId; parent_pid = [int]$process.ParentProcessId; creation_utc = $process.CreationDate.ToUniversalTime().ToString('o'); executable = $process.ExecutablePath }
}

function Read-OrcaPublicState([string]$CommandPath,[string]$MetadataFile,[int[]]$ProcessIds) {
    $resolvedCommands = @(Get-Command orca -All | ForEach-Object { [ordered]@{ type = [string]$_.CommandType; source = $_.Source; path = $_.Path } })
    $rawStatus = & $CommandPath status --json
    if ($LASTEXITCODE -ne 0) { throw '原绝对路径 Orca CLI status 查询失败' }
    $status = ($rawStatus -join "`n") | ConvertFrom-Json
    if (-not $status.ok -or -not $status.result.runtime.reachable -or $status.result.runtime.state -ne 'ready') { throw '原 Orca runtime 不是真实 ready' }
    $metadata = [Text.UTF8Encoding]::new($false,$true).GetString((Read-SharedBytes $MetadataFile)) | ConvertFrom-Json
    if ($metadata.runtimeId -ne $status.result.runtime.runtimeId) { throw '原 Orca metadata 和真实公开查询身份不符' }
    $identities = @($ProcessIds + [int]$metadata.pid + [int]$status.result.app.pid | Sort-Object -Unique | ForEach-Object { Read-OrcaProcess $_ })
    $listeners = @(Get-NetTCPConnection -State Listen | Where-Object { $identities.pid -contains $_.OwningProcess } | ForEach-Object { [ordered]@{ address = $_.LocalAddress; port = $_.LocalPort; pid = $_.OwningProcess } } | Sort-Object address,port,pid)
    return [ordered]@{ utc = [DateTime]::UtcNow.ToString('o'); commands = $resolvedCommands; status = $status; metadata = [ordered]@{ runtimeId = $metadata.runtimeId; pid = $metadata.pid; startedAt = $metadata.startedAt; transports = $metadata.transports }; processes = $identities; listeners = $listeners }
}

function New-OrcaPreservation {
    param([Parameter(Mandatory)][string]$OrcaPath,[Parameter(Mandatory)][string]$MetadataFile,[Parameter(Mandatory)][string[]]$ProtectedFiles,[Parameter(Mandatory)][int[]]$OriginalProcessIds)
    foreach ($path in @($OrcaPath,$MetadataFile) + $ProtectedFiles) { if (-not [IO.Path]::IsPathFullyQualified($path)) { throw '保全来源必须是实际绝对路径' } }
    if ($ProtectedFiles.Count -lt 1 -or $OriginalProcessIds.Count -lt 1) { throw '必须提供定向配置/认证文件和原 daemon 实际 PID' }
    $files = @($ProtectedFiles | ForEach-Object {
        $file = Get-Item -LiteralPath $_ -Force
        if ($file.Attributes -band [IO.FileAttributes]::ReparsePoint -or $file.PSIsContainer) { throw '保全文件必须为实际普通文件' }
        [pscustomobject]@{ path = $file.FullName; bytes = (Read-SharedBytes $file.FullName); acl = (Get-Acl -LiteralPath $file.FullName).Sddl }
    })
    $watchers = @($files | Group-Object { [IO.Path]::GetDirectoryName($_.path) } | ForEach-Object {
        $watcher = [IO.FileSystemWatcher]::new($_.Name)
        $watcher.IncludeSubdirectories = $false
        $watcher.NotifyFilter = [IO.NotifyFilters]::FileName -bor [IO.NotifyFilters]::LastWrite -bor [IO.NotifyFilters]::Security
        $watcher.InternalBufferSize = 65536
        $prefix = 'swa-orca-preservation-' + [Guid]::NewGuid().ToString()
        $subscriptions = @('Changed','Created','Deleted','Renamed','Error' | ForEach-Object {
            Register-ObjectEvent -InputObject $watcher -EventName $_ -SourceIdentifier ($prefix + '-' + $_)
        })
        $watcher.EnableRaisingEvents = $true
        [pscustomobject]@{ watcher = $watcher; subscriptions = $subscriptions }
    })
    $public = Read-OrcaPublicState $OrcaPath $MetadataFile $OriginalProcessIds
    return [pscustomobject]@{ command = $OrcaPath; metadata = $MetadataFile; process_ids = $OriginalProcessIds; files = $files; watchers = $watchers; baseline = $public }
}

function Assert-OrcaPreservation {
    param([Parameter(Mandatory)]$Observation,[Parameter(Mandatory)][string]$OutputFile)
    $current = Read-OrcaPublicState $Observation.command $Observation.metadata $Observation.process_ids
    $checks = @($Observation.files | ForEach-Object {
        $now = Read-SharedBytes $_.path
        $unchanged = [Collections.StructuralComparisons]::StructuralEqualityComparer.Equals($now,$_.bytes)
        [ordered]@{ path = $_.path; bytes_unchanged = $unchanged; acl_unchanged = (Get-Acl -LiteralPath $_.path).Sddl -eq $_.acl }
    })
    $notifications = @($Observation.watchers | ForEach-Object {
        foreach ($subscription in $_.subscriptions) {
            foreach ($event in @(Get-Event -SourceIdentifier $subscription.SourceIdentifier -ErrorAction SilentlyContinue)) {
                $argument = $event.SourceEventArgs
                $isError = $subscription.EventName -eq 'Error'
                [ordered]@{ utc = $event.TimeGenerated.ToUniversalTime().ToString('o'); event = $subscription.EventName; path = $(if ($isError) { $null } else { $argument.FullPath }); error = $(if ($isError) { $argument.GetException().Message } else { $null }) }
                Remove-Event -EventIdentifier $event.EventIdentifier
            }
        }
    })
    $equal = $true
    foreach ($key in @('commands','metadata','processes','listeners')) {
        if (($Observation.baseline[$key] | ConvertTo-Json -Depth 30 -Compress) -cne ($current[$key] | ConvertTo-Json -Depth 30 -Compress)) { $equal = $false }
    }
    $result = [ordered]@{ baseline = $Observation.baseline; current = $current; targeted_files = $checks; directory_notifications = $notifications; notification_scope = 'targeted_directories_without_writer_PID'; public_identity_unchanged = $equal; passed = $equal -and -not ($checks | Where-Object { -not $_.bytes_unchanged -or -not $_.acl_unchanged }) -and -not ($notifications | Where-Object error) }
    [IO.File]::WriteAllText($OutputFile, ($result | ConvertTo-Json -Depth 50))
    if (-not $result.passed) { throw "本机 Orca 保全证据不足，原件: $OutputFile" }
    return $result
}

function Close-OrcaPreservation {
    param([Parameter(Mandatory)]$Observation)
    foreach ($entry in $Observation.watchers) {
        $entry.watcher.EnableRaisingEvents = $false
        foreach ($subscription in $entry.subscriptions) {
            Unregister-Event -SourceIdentifier $subscription.SourceIdentifier
            Get-Event -SourceIdentifier $subscription.SourceIdentifier -ErrorAction SilentlyContinue | Remove-Event
        }
        $entry.watcher.Dispose()
    }
    foreach ($file in $Observation.files) { [Array]::Clear($file.bytes) }
}

Export-ModuleMember -Function New-OrcaPreservation,Assert-OrcaPreservation,Close-OrcaPreservation
