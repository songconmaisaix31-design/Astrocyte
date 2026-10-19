param(
    [string]$Pandoc = 'pandoc'
)
$ErrorActionPreference = 'Stop'
(Get-Process -Id $PID).PriorityClass = 'BelowNormal'
$env:NODE_OPTIONS = '--max-old-space-size=512'
$samples = @()
foreach ($index in 1..2) {
    $cpu = (Get-CimInstance Win32_PerfFormattedData_PerfOS_Processor -Filter "Name='_Total'").PercentProcessorTime
    $available = (Get-CimInstance Win32_OperatingSystem).FreePhysicalMemory * 1KB / 1GB
    $samples += [pscustomobject]@{utc=(Get-Date).ToUniversalTime().ToString('o');cpu=$cpu;availableGiB=$available}
    if ($cpu -ge 70 -or $available -lt 4) { throw "Resource guard rejected: CPU=$cpu availableGiB=$available" }
    if ($index -eq 1) { Start-Sleep -Seconds 3 }
}
$outputDirectory = Join-Path $PSScriptRoot 'build'
New-Item -ItemType Directory -Force -Path $outputDirectory | Out-Null
$samples | ConvertTo-Json | Set-Content -Encoding utf8 (Join-Path $outputDirectory 'resources.json')
Push-Location $PSScriptRoot
try {
    & $Pandoc --version
    if ($LASTEXITCODE -ne 0) { throw 'Pandoc version check failed' }
    & $Pandoc manuscript.md --from=markdown --to=html5 --standalone --citeproc --bibliography=references.bib --mathml --number-sections --fail-if-warnings --output=build/manuscript.html
    if ($LASTEXITCODE -ne 0) { throw 'HTML build failed' }
    & $Pandoc manuscript.md --from=markdown --to=latex --standalone --citeproc --bibliography=references.bib --number-sections --fail-if-warnings --output=build/manuscript.tex
    if ($LASTEXITCODE -ne 0) { throw 'LaTeX source generation failed' }
    & $Pandoc manuscript.md --from=markdown --to=json --output=build/manuscript.json
    if ($LASTEXITCODE -ne 0) { throw 'Document AST generation failed' }
}
finally {
    Pop-Location
}
Write-Output 'HTML, LaTeX source, and document AST built successfully. PDF compilation is a separate check.'
