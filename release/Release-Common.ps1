$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

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
