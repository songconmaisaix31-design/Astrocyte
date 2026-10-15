param(
    [Parameter(Mandatory=$true)][string]$EvidenceRoot,
    [ValidateSet('Preview','Prepare','Execute')][string]$Phase = 'Preview',
    [switch]$PromptForKey
)

$ErrorActionPreference = 'Stop'
$RepositoryRoot = (Resolve-Path -LiteralPath (Join-Path $PSScriptRoot '../..')).Path
$AuthorizedRoot = [IO.Path]::GetFullPath('D:/Astrocyte-agent-assets/20261015/judgment-s5')
if (-not [IO.Path]::IsPathRooted($EvidenceRoot)) { throw 'EvidenceRoot must be absolute' }
$EvidenceRoot = [IO.Path]::GetFullPath($EvidenceRoot)
if (-not $EvidenceRoot.StartsWith($AuthorizedRoot + [IO.Path]::DirectorySeparatorChar, [StringComparison]::OrdinalIgnoreCase)) {
    throw 'EvidenceRoot must stay in the authorized S5 asset directory'
}
if ($PromptForKey -and $Phase -ne 'Execute') { throw 'PromptForKey is only valid for an explicitly authorized Execute' }
$SourceRevision = '432312b36714d8889eba405507b37eda25257d9c'
$SourceRoot = Join-Path $RepositoryRoot '.verification/s5/upstream/laya'
$SourceURL = "https://github.com/NandhaKishorM/laya/blob/$SourceRevision/laya-ts/README.md"
$Plan = [ordered]@{
    phase = $Phase; model = 'jev-1.13.0'; endpoint = 'https://api.typesafe.ai/v1/systemone'
    source = $SourceURL; language = 'en'; question_types = @('choice','score','truth')
    manual_build_tag = 'astrocyte_jev_live'
    planned_provider_requests = 3; request_timeout_ms = 60000; max_input_bytes = 131072
    max_input_tokens = 8192; max_output_tokens = 256; automatic_retries = 0
    evidence_root = $EvidenceRoot
    key_scope = 'TYPESAFE_API_KEY only in the fixed task child environment; absent from compiler children'
    receipt_scope = 'Public request bodies and actual response bodies, status, model, usage and nullable confidence; no request headers or key'
    authorization = 'Preview and Prepare send no provider requests; Execute requires prior user authorization'
}
if ($Phase -eq 'Preview') { $Plan | ConvertTo-Json -Depth 5; return }
if (Test-Path -LiteralPath $EvidenceRoot) { throw 'EvidenceRoot already exists; first receipts must remain unchanged' }
$CheckPath = $EvidenceRoot
while ($CheckPath.StartsWith($AuthorizedRoot, [StringComparison]::OrdinalIgnoreCase)) {
    if ((Test-Path -LiteralPath $CheckPath) -and (((Get-Item -LiteralPath $CheckPath).Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0)) {
        throw 'An evidence path ancestor is redirected'
    }
    if ($CheckPath -eq $AuthorizedRoot) { break }
    $CheckPath = [IO.Path]::GetDirectoryName($CheckPath)
}
$null = New-Item -ItemType Directory -Path $EvidenceRoot
if (((Get-Item -LiteralPath $EvidenceRoot).Attributes -band [IO.FileAttributes]::ReparsePoint) -ne 0) { throw 'EvidenceRoot is redirected' }
$TemporaryRoot = Join-Path $EvidenceRoot 'temp'
$ReceiptRoot = Join-Path $EvidenceRoot 'receipts'
$null = New-Item -ItemType Directory -Path $TemporaryRoot, $ReceiptRoot
$Plan | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath (Join-Path $EvidenceRoot 'public-task.json') -Encoding utf8

function Invoke-JevTaskProcess {
    param([string]$Program, [string]$Arguments, [string]$WorkingDirectory, [string]$Name, [string]$Secret = '')
    $StartInfo = New-Object Diagnostics.ProcessStartInfo
    $StartInfo.FileName = $Program
    $StartInfo.Arguments = $Arguments
    $StartInfo.WorkingDirectory = $WorkingDirectory
    $StartInfo.UseShellExecute = $false
    $StartInfo.CreateNoWindow = $true
    $StartInfo.RedirectStandardOutput = $true
    $StartInfo.RedirectStandardError = $true
    $StartInfo.EnvironmentVariables.Remove('TYPESAFE_API_KEY')
    $StartInfo.EnvironmentVariables.Remove('ASTROCYTE_S5_JEV_LIVE')
    $StartInfo.EnvironmentVariables['GOCACHE'] = 'D:/Astrocyte-agent-assets/20261015/go-build/backend'
    $StartInfo.EnvironmentVariables['GOTMPDIR'] = $TemporaryRoot
    $StartInfo.EnvironmentVariables['TEMP'] = $TemporaryRoot
    $StartInfo.EnvironmentVariables['TMP'] = $TemporaryRoot
    if ($Secret -ne '') {
        $StartInfo.EnvironmentVariables['TYPESAFE_API_KEY'] = $Secret
        $StartInfo.EnvironmentVariables['ASTROCYTE_S5_JEV_LIVE'] = '1'
        $StartInfo.EnvironmentVariables['ASTROCYTE_S5_JEV_RUN_ROOT'] = $ReceiptRoot
    }
    $Child = New-Object Diagnostics.Process
    $Child.StartInfo = $StartInfo
    try {
        if (-not $Child.Start()) { throw 'The owned child could not be started' }
        $StartInfo.EnvironmentVariables.Remove('TYPESAFE_API_KEY')
        $OutputTask = $Child.StandardOutput.ReadToEndAsync()
        $ErrorsTask = $Child.StandardError.ReadToEndAsync()
        $Child.WaitForExit()
        $Output = $OutputTask.GetAwaiter().GetResult()
        $Errors = $ErrorsTask.GetAwaiter().GetResult()
        if ($Secret -ne '' -and ($Output.Contains($Secret) -or $Errors.Contains($Secret))) {
            throw 'Child output contained the credential; output was not printed or persisted'
        }
        [IO.File]::WriteAllText((Join-Path $EvidenceRoot "$Name.stdout.log"), $Output)
        [IO.File]::WriteAllText((Join-Path $EvidenceRoot "$Name.stderr.log"), $Errors)
        if ($Child.ExitCode -ne 0) { throw "The owned $Name child failed with exit $($Child.ExitCode); no automatic retry" }
        return $Output
    } finally {
        $StartInfo.EnvironmentVariables.Remove('TYPESAFE_API_KEY')
        $Child.Dispose()
    }
}

$ActualRevision = Invoke-JevTaskProcess -Program 'git.exe' -Arguments 'rev-parse HEAD' -WorkingDirectory $SourceRoot -Name 'public-source-revision'
if ($ActualRevision.Trim() -ne $SourceRevision) { throw 'The public source revision differs from the fixed task' }
$null = Invoke-JevTaskProcess -Program 'git.exe' -Arguments "diff --quiet $SourceRevision -- laya-ts/README.md" -WorkingDirectory $SourceRoot -Name 'public-source-content'
$TestProgram = Join-Path $EvidenceRoot 'jev-public.test.exe'
$null = Invoke-JevTaskProcess -Program 'go.exe' -Arguments "test -c -tags=astrocyte_jev_live -trimpath -mod=readonly -o `"$TestProgram`" ./internal/adapters/judgment" -WorkingDirectory $RepositoryRoot -Name 'compile'
if ($Phase -eq 'Prepare') {
    [ordered]@{ phase = 'prepared'; provider_requests = 0; source_revision = $SourceRevision; test_program = $TestProgram; receipts = $ReceiptRoot } | ConvertTo-Json
    return
}

$Secret = $null
$SecureKey = $null
$KeyPointer = [IntPtr]::Zero
try {
    if ($PromptForKey) {
        $SecureKey = Read-Host 'One-time TYPESAFE_API_KEY for this fixed public task' -AsSecureString
        $KeyPointer = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($SecureKey)
        $Secret = [Runtime.InteropServices.Marshal]::PtrToStringBSTR($KeyPointer)
    } else {
        $Secret = [Environment]::GetEnvironmentVariable('TYPESAFE_API_KEY', 'Process')
    }
    if ([string]::IsNullOrWhiteSpace($Secret)) { throw 'TYPESAFE_API_KEY is absent; no provider request was sent' }
    $Output = Invoke-JevTaskProcess -Program $TestProgram -Arguments '-test.run=^TestJevActualPinnedPublicThreeTypes$ -test.v -test.timeout=4m' -WorkingDirectory (Join-Path $RepositoryRoot 'internal/adapters/judgment') -Name 'actual' -Secret $Secret
    Write-Output $Output
    [ordered]@{ phase = 'actual_public_task_passed'; model = 'jev-1.13.0'; question_types = @('choice','score','truth'); receipts = $ReceiptRoot } | ConvertTo-Json
} finally {
    if ($KeyPointer -ne [IntPtr]::Zero) { [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($KeyPointer) }
    if ($null -ne $SecureKey) { $SecureKey.Dispose() }
    $Secret = $null
}
