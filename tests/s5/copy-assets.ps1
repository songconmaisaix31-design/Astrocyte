$ErrorActionPreference = 'Stop'
$s5Workspace = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '../..')).Path
$s5OwnedRoot = 'D:\Astrocyte-agent-assets\20261015\judgment-s5'
if ((Resolve-Path -LiteralPath $s5OwnedRoot).Path -ne $s5OwnedRoot) { throw 'S5 asset root differs from the authorized directory' }
if ((Get-Item -LiteralPath $s5OwnedRoot).Attributes -band [IO.FileAttributes]::ReparsePoint) { throw 'S5 asset root is redirected' }
$s5Items = @(
    @{ Checkpoint = 'english'; Source = '.verification/s5/assets/onnx' },
    @{ Checkpoint = 'multilingual'; Source = '.verification/s5/assets-multilingual/onnx' }
)
$s5Copied = @()
foreach ($s5Item in $s5Items) {
    $s5Source = (Resolve-Path -LiteralPath (Join-Path $s5Workspace $s5Item.Source)).Path
    if (-not $s5Source.StartsWith($s5Workspace + '\.verification\s5\', [StringComparison]::OrdinalIgnoreCase)) { throw 'Source escapes the owned S5 directory' }
    $s5Destination = [IO.Path]::GetFullPath((Join-Path $s5OwnedRoot $s5Item.Checkpoint))
    if (-not $s5Destination.StartsWith($s5OwnedRoot + '\', [StringComparison]::OrdinalIgnoreCase)) { throw 'Destination escapes the authorized root' }
    if (Test-Path -LiteralPath $s5Destination) { throw 'Destination already exists; existing data must be reviewed before copying' }
    New-Item -ItemType Directory -Path $s5Destination | Out-Null
    $s5Metadata = Get-Content -LiteralPath (Join-Path $s5Workspace ('.verification/s5/laya-' + $s5Item.Checkpoint + '-assets.json')) -Raw | ConvertFrom-Json
    foreach ($s5Property in $s5Metadata.files.PSObject.Properties) {
        if ([IO.Path]::GetFileName($s5Property.Name) -ne $s5Property.Name) { throw 'Unsafe asset filename' }
        $s5Original = Join-Path $s5Source $s5Property.Name
        $s5Target = Join-Path $s5Destination $s5Property.Name
        $s5OriginalDigest = (Get-FileHash -LiteralPath $s5Original -Algorithm SHA256).Hash.ToLowerInvariant()
        if ($s5OriginalDigest -ne $s5Property.Value) { throw 'Source digest differs from the verified export' }
        Copy-Item -LiteralPath $s5Original -Destination $s5Target
        $s5TargetDigest = (Get-FileHash -LiteralPath $s5Target -Algorithm SHA256).Hash.ToLowerInvariant()
        if ($s5TargetDigest -ne $s5OriginalDigest) { throw 'Copied asset digest differs' }
        $s5Copied += [pscustomobject]@{ Checkpoint = $s5Item.Checkpoint; Name = $s5Property.Name; Bytes = (Get-Item -LiteralPath $s5Target).Length; SHA256 = $s5TargetDigest }
    }
}
# 仅复制和核验；D 实际加载与重启通过后才单独删除可重建的 C 副本。
$s5Copied | ConvertTo-Json -Depth 4 | Set-Content -LiteralPath (Join-Path $s5Workspace '.verification/s5/assets-copy-d-safe.json') -Encoding utf8
$s5Copied | Measure-Object Bytes -Sum | Select-Object Count, Sum
