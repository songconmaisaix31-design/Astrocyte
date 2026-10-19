#requires -Version 7.2
param(
    [Parameter(Mandatory)][ValidateSet('Download', 'VerifyInstaller', 'Install', 'Build', 'Validate')][string]$Phase,
    [ValidatePattern('^[0-9a-f]{40}$')][string]$SourceCommit = 'a664233a1276003a69f3d6850e4ecc04c08f4370',
    [string]$Sha512Sum = 'C:/Program Files/Git/usr/bin/sha512sum.exe',
    [string]$Gpg = 'C:/Program Files/Git/usr/bin/gpg.exe',
    [string]$Python = 'python'
)
$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest
(Get-Process -Id $PID).PriorityClass = 'BelowNormal'
$env:NODE_OPTIONS = '--max-old-space-size=1024'
$workspace = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
$workRoot = Join-Path $workspace '.verification/texlive2025'
$repository = 'https://ftp.math.utah.edu/pub/tex/historic/systems/texlive/2025/tlnet-final/'
$signer = 'C78B82D8C79512F79CC0D7C80D5E5D9106BAB6BC'

function Get-WorkPath([string]$RelativePath) {
    $path = [IO.Path]::GetFullPath((Join-Path $workRoot $RelativePath))
    if (-not $path.StartsWith($workRoot + [IO.Path]::DirectorySeparatorChar, [StringComparison]::OrdinalIgnoreCase)) {
        throw "Path outside owned verification directory: $RelativePath"
    }
    return $path
}

function Invoke-Tool([string]$Program, [string[]]$ToolArguments, [string]$LogName) {
    & $Program @ToolArguments 2>&1 | Tee-Object -FilePath (Get-WorkPath "logs/$LogName")
    if ($LASTEXITCODE -ne 0) { throw "$Program failed with exit code $LASTEXITCODE; inspect logs/$LogName" }
}

function New-EmptyDirectory([string]$Directory) {
    if (Test-Path -LiteralPath $Directory) { throw "Output directory already exists: $Directory" }
    New-Item -ItemType Directory -Path $Directory | Out-Null
}

function Confirm-Installer {
    foreach ($tool in @($Sha512Sum, $Gpg)) {
        if (-not (Test-Path -LiteralPath $tool -PathType Leaf)) { throw "Required mature tool missing: $tool" }
    }
    foreach ($name in @('install-tl.zip', 'install-tl.zip.sha512', 'install-tl.zip.sha512.asc', 'texlive.asc')) {
        if (-not (Test-Path -LiteralPath (Join-Path $downloads $name) -PathType Leaf)) { throw "Original downloaded input missing: $name" }
    }
    Push-Location $downloads
    try { Invoke-Tool $Sha512Sum @('--check', '--strict', 'install-tl.zip.sha512') 'verify-installer-sha512.log' }
    finally { Pop-Location }
    $verificationHome = Get-WorkPath 'verify-installer'
    New-EmptyDirectory $verificationHome
    Copy-Item -LiteralPath (Join-Path $downloads 'texlive.asc') -Destination (Join-Path $verificationHome 'texlive.asc')
    Push-Location $verificationHome
    try {
        Invoke-Tool $Gpg @('--no-options', '--batch', '--no-autostart', '--homedir', $verificationHome, '--output', 'pubring.gpg', '--dearmor', 'texlive.asc') 'verify-installer-dearmor.log'
    }
    finally { Pop-Location }
    Push-Location $downloads
    try {
        Invoke-Tool $Gpg @('--no-options', '--batch', '--no-autostart', '--homedir', $verificationHome, '--no-default-keyring', '--keyring', 'pubring.gpg', '--assert-signer', $signer, '--status-fd', '1', '--verify', 'install-tl.zip.sha512.asc', 'install-tl.zip.sha512') 'verify-installer-signature.log'
    }
    finally { Pop-Location }
    New-EmptyDirectory $installer
    Expand-Archive -LiteralPath (Join-Path $downloads 'install-tl.zip') -DestinationPath $installer
}

# 每次真正的下载、安装、构建或验证均使用相同启动守卫。
foreach ($directory in @($workRoot, (Get-WorkPath 'logs'), (Get-WorkPath 'temp'), (Get-WorkPath 'cache'), (Get-WorkPath 'gnupg'))) {
    New-Item -ItemType Directory -Force -Path $directory | Out-Null
}
$resourceStamp = (Get-Date).ToUniversalTime().ToString('yyyyMMddTHHmmssfffffffZ')
$resourceLog = Get-WorkPath "logs/$Phase-resources-$resourceStamp.json"
$samples = @()
$sampleClock = [Diagnostics.Stopwatch]::StartNew()
$previousElapsed = $null
foreach ($index in 1..2) {
    $cpu = (Get-CimInstance Win32_PerfFormattedData_PerfOS_Processor -Filter "Name='_Total'").PercentProcessorTime
    $available = (Get-CimInstance Win32_OperatingSystem).FreePhysicalMemory * 1KB / 1GB
    $elapsed = $sampleClock.Elapsed.TotalSeconds
    $interval = $null
    $rejections = @()
    if ($index -eq 2) {
        $interval = $elapsed - $previousElapsed
        if ($interval -lt 2 -or $interval -gt 5) { $rejections += "Sample interval outside 2-5 seconds: $interval" }
    }
    if ($cpu -ge 85) { $rejections += "CPU at least 85%: $cpu" }
    if ($available -lt 4) { $rejections += "Available memory below 4 GiB: $available" }
    $samples += [pscustomobject]@{utc=(Get-Date).ToUniversalTime().ToString('o');monotonicSeconds=$elapsed;intervalSeconds=$interval;cpu=$cpu;availableGiB=$available;rejectionReasons=@($rejections)}
    ConvertTo-Json -InputObject @($samples) -Depth 4 | Set-Content -Encoding utf8 -LiteralPath $resourceLog
    if ($rejections.Count -ne 0) { throw ('Resource guard rejected: ' + ($rejections -join '; ')) }
    $previousElapsed = $elapsed
    if ($index -eq 1) { Start-Sleep -Seconds 2 }
}
$env:TEMP = Get-WorkPath 'temp'
$env:TMP = $env:TEMP
$env:TMPDIR = $env:TEMP
$env:XDG_CACHE_HOME = Get-WorkPath 'cache'
$env:GNUPGHOME = Get-WorkPath 'gnupg'
$env:TL_GNUPGHOME = $env:GNUPGHOME
$env:TL_GNUPGARGS = '--batch --no-autostart --no-secmem-warning --no-permission-warning'
$installRoot = Get-WorkPath 'texlive'
$env:TEXMFLOCAL = Join-Path $installRoot 'texmf-local'
$env:TEXMFSYSCONFIG = Join-Path $installRoot 'texmf-config'
$env:TEXMFSYSVAR = Join-Path $installRoot 'texmf-var'
$env:TEXMFHOME = $env:TEXMFLOCAL
$env:TEXMFCONFIG = $env:TEXMFSYSCONFIG
$env:TEXMFVAR = $env:TEXMFSYSVAR
$env:TEXMFCACHE = $env:TEXMFSYSVAR
$env:TEXMFCNF = "$installRoot;$(Join-Path $installRoot 'texmf-dist/web2c')"
$env:TEXMFOUTPUT = $env:TEMP
$env:FONTCONFIG_FILE = Join-Path $PSScriptRoot 'fontconfig.conf'

if ($Phase -in @('Download', 'VerifyInstaller', 'Install')) {
    $driveName = [IO.Path]::GetPathRoot($workRoot).Substring(0, 1)
    $free = (Get-PSDrive -Name $driveName).Free
    if ($free -lt 30GB) { throw "Less than planned 30 GiB free disk space: $free bytes" }
}

$downloads = Get-WorkPath 'downloads'
$installer = Get-WorkPath 'installer'
$sourceZip = Get-WorkPath 'package/arxiv-source.zip'
$buildDirectory = Get-WorkPath 'build'
$pdfPath = Join-Path $buildDirectory 'manuscript.pdf'

switch ($Phase) {
    'Download' {
        New-EmptyDirectory $downloads
        foreach ($name in @('install-tl.zip', 'install-tl.zip.sha512', 'install-tl.zip.sha512.asc')) {
            Invoke-WebRequest -Uri ($repository + $name) -OutFile (Join-Path $downloads $name)
        }
        Invoke-WebRequest -Uri 'https://tug.org/texlive/files/texlive.asc' -OutFile (Join-Path $downloads 'texlive.asc')
        Confirm-Installer
    }
    'VerifyInstaller' { Confirm-Installer }
    'Install' {
        $profile = Get-WorkPath 'texlive.profile'
        if (-not (Test-Path -LiteralPath $profile -PathType Leaf)) { throw 'Create the approved absolute-path profile with a file editing tool first' }
        if (Test-Path -LiteralPath $installRoot) { throw 'Installation target already exists; preserve it and report before any retry' }
        $archive = [IO.Compression.ZipFile]::OpenRead((Join-Path $downloads 'install-tl.zip'))
        try {
            $wrappers = @($archive.Entries | Where-Object { $_.Name -eq 'install-tl-windows.bat' })
            if ($wrappers.Count -ne 1) { throw 'Installer ZIP must contain exactly one Windows wrapper' }
            $wrapper = [IO.Path]::GetFullPath((Join-Path $installer $wrappers[0].FullName))
            if (-not $wrapper.StartsWith($installer + [IO.Path]::DirectorySeparatorChar, [StringComparison]::OrdinalIgnoreCase)) { throw 'Installer entry outside extraction directory' }
        }
        finally { $archive.Dispose() }
        if (-not (Test-Path -LiteralPath $wrapper -PathType Leaf)) { throw 'Verified installer wrapper missing' }
        $installerDirectory = Split-Path -Parent $wrapper
        $perlBin = Join-Path $installerDirectory 'tlpkg/tlperl/bin'
        $perl = Join-Path $perlBin 'perl.exe'
        $perlLibrary = Join-Path $installerDirectory 'tlpkg/tlperl/lib'
        $installerScript = Join-Path $installerDirectory 'install-tl'
        foreach ($inputFile in @($perl, $installerScript)) {
            if (-not (Test-Path -LiteralPath $inputFile -PathType Leaf)) { throw "Verified installer input missing: $inputFile" }
        }
        if (-not (Test-Path -LiteralPath $perlLibrary -PathType Container)) { throw 'Bundled Perl library missing' }
        foreach ($pathEntry in $env:PATH.Split(';', [StringSplitOptions]::RemoveEmptyEntries)) {
            if (Test-Path -LiteralPath (Join-Path $pathEntry.Trim('"') 'pdftex.exe') -PathType Leaf) {
                throw "Another TeX installation is on the process PATH: $pathEntry"
            }
        }
        # 沿官方 wrapper 的环境清单使用包内 Perl，并直接保留原生退出码。
        foreach ($variable in @(
            'PERL5OPT', 'PERLIO', 'PERLIO_DEBUG', 'PERLLIB', 'PERL5DB',
            'PERL5DB_THREADED', 'PERL5SHELL', 'PERL_ALLOW_NON_IFS_LSP',
            'PERL_DEBUG_MSTATS', 'PERL_DESTRUCT_LEVEL', 'PERL_DL_NONLAZY',
            'PERL_ENCODING', 'PERL_HASH_SEED', 'PERL_HASH_SEED_DEBUG',
            'PERL_ROOT', 'PERL_SIGNALS', 'PERL_UNICODE'
        )) {
            [Environment]::SetEnvironmentVariable($variable, $null, 'Process')
        }
        $env:PERL5LIB = $perlLibrary
        if (-not (Test-Path -LiteralPath $Gpg -PathType Leaf)) { throw "Selected GPG program missing: $Gpg" }
        $gpgPath = [IO.Path]::GetFullPath($Gpg)
        $gpgDirectory = Split-Path -Parent $gpgPath
        $gpgName = Split-Path -Leaf $gpgPath
        if (Test-Path -LiteralPath (Join-Path $perlBin $gpgName)) { throw 'Bundled Perl directory shadows the selected GPG program' }
        # 官方 Windows which 只接受程序名，目录由本进程 PATH 明确提供。
        $env:TL_GNUPG = $gpgName
        $env:PATH = $perlBin + [IO.Path]::PathSeparator + $gpgDirectory + [IO.Path]::PathSeparator + $env:PATH
        $env:TL_GNUPGHOME = Join-Path $installerDirectory 'tlpkg/gpg'
        if (-not (Test-Path -LiteralPath (Join-Path $env:TL_GNUPGHOME 'pubring.gpg') -PathType Leaf)) { throw 'Official installer repository keyring missing' }
        Push-Location (Get-WorkPath 'logs')
        try {
            Invoke-Tool $perl @($installerScript, '-no-gui', '-lang', 'en', '-profile', $profile, '-portable', '-non-admin', '-repository', $repository, '-no-continue', '-verify-downloads', '-logfile', (Get-WorkPath "logs/install-tl-$resourceStamp.log")) "installer-console-$resourceStamp.log"
        }
        finally { Pop-Location }
    }
    'Build' {
        Invoke-Tool 'git' @('-C', $workspace, 'diff', '--exit-code', $SourceCommit, '--', 'papers/astrocyte/preprint/manuscript.tex') 'source-diff.log'
        New-EmptyDirectory (Get-WorkPath 'package')
        Copy-Item -LiteralPath (Join-Path $PSScriptRoot 'preprint/manuscript.tex') -Destination (Get-WorkPath 'package/manuscript.tex')
        Compress-Archive -LiteralPath (Get-WorkPath 'package/manuscript.tex') -DestinationPath $sourceZip
        $archive = [IO.Compression.ZipFile]::OpenRead($sourceZip)
        try {
            if ($archive.Entries.Count -ne 1 -or $archive.Entries[0].FullName -cne 'manuscript.tex') { throw 'Source ZIP contains unexpected entries' }
        }
        finally { $archive.Dispose() }
        $cleanSource = Get-WorkPath 'clean-source'
        New-EmptyDirectory $cleanSource
        New-EmptyDirectory $buildDirectory
        Expand-Archive -LiteralPath $sourceZip -DestinationPath $cleanSource
        $sourceBytes = [IO.File]::ReadAllBytes((Join-Path $PSScriptRoot 'preprint/manuscript.tex'))
        $extractedBytes = [IO.File]::ReadAllBytes((Join-Path $cleanSource 'manuscript.tex'))
        if (-not [Collections.StructuralComparisons]::StructuralEqualityComparer.Equals($sourceBytes, $extractedBytes)) { throw 'Extracted source differs from complete review source' }
        $engine = Join-Path $installRoot 'bin/windows/xelatex.exe'
        $version = & $engine --version
        if ($LASTEXITCODE -ne 0) { throw 'XeLaTeX version command failed' }
        $version | Set-Content -Encoding utf8 (Get-WorkPath 'logs/xelatex-version.log')
        if (($version -join "`n") -notmatch 'TeX Live 2025') { throw 'Expected the approved TeX Live 2025 engine' }
        Push-Location $cleanSource
        try {
            foreach ($pass in 1..2) {
                Invoke-Tool $engine @('-no-shell-escape', '-halt-on-error', '-file-line-error', '-interaction=nonstopmode', '-recorder', "-output-directory=$buildDirectory", 'manuscript.tex') "xelatex-pass$pass.log"
            }
        }
        finally { Pop-Location }
        if (-not (Test-Path -LiteralPath $pdfPath -PathType Leaf)) { throw 'XeLaTeX did not produce a PDF' }
        $badDiagnostic = 'Undefined control sequence|Missing character|undefined references|LaTeX Warning: (Reference|Citation) .+ undefined|Overfull'
        if (Select-String -LiteralPath (Join-Path $buildDirectory 'manuscript.log') -Pattern $badDiagnostic -Quiet) { throw 'Unresolved TeX diagnostics; inspect the retained complete log' }
        Copy-Item -LiteralPath $sourceZip -Destination (Join-Path $PSScriptRoot 'preprint/arxiv-source.zip')
    }
    'Validate' {
        $env:PYTHONPATH = Join-Path $workspace '.verification/python'
        Invoke-Tool $Python @('-X', 'utf8', (Join-Path $PSScriptRoot 'verify.py'), '--pdf', $pdfPath) 'pdf-validation.log'
    }
}
Write-Output "$Phase finished; retain diagnostics and report actual scope before the next phase."
