$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

function Assert-ReleaseResources([string]$Phase, [string]$Directory) {
    if ($Phase -cnotmatch '^[a-z0-9-]+$') { throw 'Resource phase must use a unique lowercase name.' }
    $record = Join-Path $Directory ('resources-' + $Phase + '.json')
    if (Test-Path -LiteralPath $record) { throw 'Resource sampling record already exists.' }
    $samples = @()
    $clock = [Diagnostics.Stopwatch]::StartNew()
    for ($index = 0; $index -lt 2; $index++) {
        $memory = Get-CimInstance Win32_OperatingSystem
        $processor = Get-CimInstance Win32_PerfFormattedData_PerfOS_Processor -Filter "Name='_Total'"
        if ($null -eq $memory -or $null -eq $processor) { throw 'Actual CPU and memory observations are required.' }
        $samples += [pscustomobject]@{
            utc = [DateTimeOffset]::UtcNow.ToString('o')
            elapsed_ms = $clock.Elapsed.TotalMilliseconds
            cpu_percent = [double]$processor.PercentProcessorTime
            available_gib = [double]$memory.FreePhysicalMemory / 1MB
        }
        if ($index -eq 0) { Start-Sleep -Seconds 2 }
    }
    $clock.Stop()
    $interval = $samples[1].elapsed_ms - $samples[0].elapsed_ms
    $priority = [Diagnostics.Process]::GetCurrentProcess().PriorityClass
    $valid = $interval -ge 2000 -and $interval -le 5000 -and $priority -eq 'BelowNormal'
    foreach ($sample in $samples) {
        if ($sample.cpu_percent -lt 0 -or $sample.cpu_percent -ge 85 -or $sample.available_gib -lt 4) { $valid = $false }
    }
    [IO.File]::WriteAllText($record, (@{ phase = $Phase; allowed = $valid; priority = [string]$priority; samples = $samples } | ConvertTo-Json -Depth 4))
    Write-Output ('RESOURCE_PHASE=' + $Phase + ' ALLOWED=' + $valid + ' RECORD=' + $record)
    if (-not $valid) { throw ('CPU, memory, priority or sample interval does not permit phase: ' + $Phase) }
}

function Get-ReleaseOwnedListener([Diagnostics.Process]$Owned, [int]$Port, [switch]$AllowPending) {
    if ($null -eq $Owned -or $Owned.HasExited) { throw 'Owned application root is unavailable or exited.' }
    $root = Get-CimInstance Win32_Process -Filter ('ProcessId=' + $Owned.Id) -Property ProcessId, ParentProcessId, CreationDate
    if ($null -eq $root -or [Math]::Abs(($root.CreationDate.ToUniversalTime() - $Owned.StartTime.ToUniversalTime()).TotalMilliseconds) -ge 1) { throw 'Owned application root identity changed.' }
    # MSFT_NetTCPConnection 的 Listen 枚举值为 2；没有监听时返回空集合。
    $listeners = @(Get-CimInstance -Namespace 'root/StandardCimv2' -ClassName MSFT_NetTCPConnection -Filter ('LocalPort=' + $Port + ' AND State=2') -Property LocalAddress, LocalPort, OwningProcess)
    if ($listeners.Count -eq 0 -and $AllowPending) { return $null }
    if ($listeners.Count -ne 1 -or $listeners[0].LocalAddress -ne '127.0.0.1') { throw 'Exact owned loopback listener is unavailable.' }
    $listenerID = [int]$listeners[0].OwningProcess
    $cursor = $listenerID
    $childStarted = $null
    $chain = [Collections.Generic.List[int]]::new()
    while ($true) {
        if ($chain.Contains($cursor)) { throw 'Application process ancestry is cyclic.' }
        $chain.Add($cursor)
        $item = Get-CimInstance Win32_Process -Filter ('ProcessId=' + $cursor) -Property ProcessId, ParentProcessId, CreationDate
        if ($null -eq $item -or $item.CreationDate -lt $root.CreationDate -or ($null -ne $childStarted -and $item.CreationDate -gt $childStarted)) { throw 'Application listener ancestry is unavailable or changed.' }
        if ($cursor -eq $Owned.Id) {
            if ($item.CreationDate -ne $root.CreationDate) { throw 'Application root was replaced during listener verification.' }
            break
        }
        if ($item.ParentProcessId -eq 0) { throw 'Port belongs to another process tree.' }
        $childStarted = $item.CreationDate
        $cursor = [int]$item.ParentProcessId
    }
    $current = @(Get-CimInstance -Namespace 'root/StandardCimv2' -ClassName MSFT_NetTCPConnection -Filter ('LocalPort=' + $Port + ' AND State=2') -Property LocalAddress, LocalPort, OwningProcess)
    if ($Owned.HasExited -or $current.Count -ne 1 -or $current[0].LocalAddress -ne '127.0.0.1' -or [int]$current[0].OwningProcess -ne $listenerID) { throw 'Owned application listener changed before HTTP access.' }
    return @{ root_pid = $Owned.Id; root_started_utc = $root.CreationDate.ToUniversalTime().ToString('o'); listener_pid = $listenerID; process_chain = $chain.ToArray() }
}

function Assert-ReleaseSpace {
    $drive = [IO.DriveInfo]::new('C:\')
    if ($drive.AvailableFreeSpace -lt 10GB) { throw 'C: requires at least 10 GiB free before and after large operations.' }
    Write-Output ('C_FREE_BYTES=' + $drive.AvailableFreeSpace)
}

function Assert-ReleaseAssetPath([string]$Path) {
    $root = [IO.Path]::GetFullPath('D:\Astrocyte-agent-assets\20261015\release-package\')
    $full = [IO.Path]::GetFullPath($Path)
    if (-not $full.StartsWith($root, [StringComparison]::OrdinalIgnoreCase) -or $full.TrimEnd('\') -eq $root.TrimEnd('\')) {
        throw 'New release assets must be inside the exclusive D: release-package directory.'
    }
    $cursor = $full
    while ($cursor -ne [IO.Path]::GetPathRoot($cursor)) {
        if (Test-Path -LiteralPath $cursor) {
            if ((Get-Item -LiteralPath $cursor -Force).Attributes -band [IO.FileAttributes]::ReparsePoint) { throw 'Release output must not traverse a reparse point.' }
        }
        $cursor = Split-Path $cursor -Parent
    }
}

function Assert-ReleaseMember([string]$Name) {
    if ($Name.Contains('\') -or $Name.Contains(':') -or $Name.StartsWith('/') -or $Name -match '(^|/)\.\.(/|$)') { throw ('Unsafe archive member: ' + $Name) }
    $relative = $Name -replace '^Astrocyte-v0\.1-windows-x64/', ''
    $allowed = $Name.StartsWith('Astrocyte-v0.1-windows-x64/') -and ($relative -match '^(astrocyte(-mcp|-sleep)?\.exe|workbench-snapshot\.exe|Start-Astrocyte\.(ps1|cmd)|Publish-WorkbenchSnapshot\.ps1|README\.md|LICENSE|NOTICE\.md|web/[^:]+|legal/[^:]+|runtime/(node/[^:]+|sdk/[^:]+|adapters/(laya-worker|codegraph)\.mjs|models/(english|multilingual)/[^:]+))$')
    $unsafe = $relative -match '(^|/)(\.git|\.env[^/]*|\.npmrc|\.netrc|auth\.json|credentials[^/]*|test-results|playwright-report|data)(/|$)|\.(sqlite|db|log|jsonl)(-|$|\.)'
    if ($relative -match '(^|/)node_modules(/|$)' -and -not $relative.StartsWith('runtime/sdk/node_modules/')) { $unsafe = $true }
    if (-not $allowed -or $unsafe) { throw ('Unexpected package member: ' + $Name) }
}

function New-ReleaseZip([string]$Directory, [string]$Destination) {
    Add-Type -AssemblyName System.IO.Compression
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    $base = [IO.Path]::GetFullPath($Directory).TrimEnd('\')
    $prefix = [IO.Path]::GetFileName($base) + '/'
    $stream = [IO.File]::Open($Destination, [IO.FileMode]::CreateNew, [IO.FileAccess]::Write)
    try {
        $archive = [IO.Compression.ZipArchive]::new($stream, [IO.Compression.ZipArchiveMode]::Create, $true)
        try {
            foreach ($file in Get-ChildItem -LiteralPath $base -Recurse -File -Force) {
                $name = $prefix + $file.FullName.Substring($base.Length + 1).Replace('\', '/')
                Assert-ReleaseMember $name
                [IO.Compression.ZipFileExtensions]::CreateEntryFromFile($archive, $file.FullName, $name, [IO.Compression.CompressionLevel]::Optimal) | Out-Null
            }
        } finally { $archive.Dispose() }
    } finally { $stream.Dispose() }
}
