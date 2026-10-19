param(
    [string]$Pandoc = 'pandoc',
    [string]$Tectonic = 'tectonic'
)
$ErrorActionPreference = 'Stop'
(Get-Process -Id $PID).PriorityClass = 'BelowNormal'
$env:NODE_OPTIONS = '--max-old-space-size=1024'
$samples = @()
$sampleClock = [System.Diagnostics.Stopwatch]::StartNew()
$previousSampleElapsed = $null
foreach ($index in 1..2) {
    $cpu = (Get-CimInstance Win32_PerfFormattedData_PerfOS_Processor -Filter "Name='_Total'").PercentProcessorTime
    $available = (Get-CimInstance Win32_OperatingSystem).FreePhysicalMemory * 1KB / 1GB
    $sampleTime = (Get-Date).ToUniversalTime()
    $sampleElapsed = $sampleClock.Elapsed.TotalSeconds
    if ($index -eq 2) {
        $elapsedSeconds = $sampleElapsed - $previousSampleElapsed
        if ($elapsedSeconds -lt 2 -or $elapsedSeconds -gt 5) { throw "Resource sample interval outside 2–5 seconds: $elapsedSeconds" }
    }
    $samples += [pscustomobject]@{utc=$sampleTime.ToString('o');cpu=$cpu;availableGiB=$available}
    $previousSampleElapsed = $sampleElapsed
    if ($cpu -ge 85 -or $available -lt 4) { throw "Resource guard rejected: CPU=$cpu availableGiB=$available" }
    if ($index -eq 1) { Start-Sleep -Seconds 2 }
}
$outputDirectory = Join-Path $PSScriptRoot 'build'
New-Item -ItemType Directory -Force -Path $outputDirectory | Out-Null
$env:TECTONIC_CACHE_DIR = Join-Path $outputDirectory 'tectonic-cache'
$env:FONTCONFIG_FILE = Join-Path $PSScriptRoot 'fontconfig.conf'
$env:XDG_CACHE_HOME = Join-Path $outputDirectory 'cache'
$env:TEMP = Join-Path $outputDirectory 'temp'
$env:TMP = $env:TEMP
New-Item -ItemType Directory -Force -Path $env:TEMP | Out-Null
$samples | ConvertTo-Json | Set-Content -Encoding utf8 (Join-Path $outputDirectory 'resources.json')
Push-Location $PSScriptRoot
try {
    & $Pandoc --version
    if ($LASTEXITCODE -ne 0) { throw 'Pandoc version check failed' }
    & $Tectonic --version
    if ($LASTEXITCODE -ne 0) { throw 'Tectonic version check failed' }
    & $Pandoc manuscript.md --from=markdown --to=html5 --standalone --citeproc --bibliography=references.bib --mathml --number-sections --fail-if-warnings --output=build/manuscript.html
    if ($LASTEXITCODE -ne 0) { throw 'HTML build failed' }
    & $Pandoc manuscript.md --from=markdown --to=latex --standalone --citeproc --bibliography=references.bib --number-sections --variable=geometry:margin=1in --variable=fontsize:11pt --fail-if-warnings --output=build/manuscript.tex
    if ($LASTEXITCODE -ne 0) { throw 'LaTeX source generation failed' }
    & $Pandoc manuscript.md --from=markdown --to=json --output=build/manuscript.json
    if ($LASTEXITCODE -ne 0) { throw 'Document AST generation failed' }
    & $Tectonic -X compile --untrusted --keep-logs --keep-intermediates --outdir=build build/manuscript.tex 2>&1 | Tee-Object -FilePath build/tectonic-console.log
    if ($LASTEXITCODE -ne 0) { throw 'TeX compilation failed; inspect build/tectonic-console.log and build/manuscript.log' }
    if (-not (Test-Path build/manuscript.pdf -PathType Leaf)) { throw 'TeX produced no PDF' }
    New-Item -ItemType Directory -Force -Path preprint | Out-Null
    Copy-Item build/manuscript.tex preprint/manuscript.tex
    Copy-Item build/manuscript.pdf preprint/manuscript.pdf
}
finally {
    Pop-Location
}
Write-Output 'Built HTML, document AST, complete LaTeX source, and PDF. Read the retained TeX diagnostics and run verify.py before delivery.'
