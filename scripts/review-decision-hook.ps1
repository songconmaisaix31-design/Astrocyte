param(
    [ValidatePattern('^run_[a-z0-9]+$')][string]$RunId,
    [ValidatePattern('^task_[a-z0-9]+$')][string]$TaskId,
    [ValidatePattern('^ctx_[a-z0-9]+$')][string]$DispatchId,
    [ValidatePattern('^term_[a-f0-9-]+$')][string]$FromHandle,
    [ValidatePattern('^[a-f0-9]{40}$')][string]$CandidateSHA,
    [ValidateSet('accept', 'revise', 'insufficient_evidence')][string]$Decision,
    [ValidateNotNullOrEmpty()][string]$ReportPath,
    [ValidateNotNullOrEmpty()][string]$Reason
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

foreach ($requiredName in @('RunId', 'TaskId', 'DispatchId', 'FromHandle', 'CandidateSHA', 'Decision', 'ReportPath', 'Reason')) {
    if (-not $PSBoundParameters.ContainsKey($requiredName)) {
        throw "缺少必填参数 $requiredName。"
    }
}
if ([string]::IsNullOrWhiteSpace($Reason)) {
    throw '审查依据不能为空。'
}
if (-not [System.IO.Path]::IsPathRooted($ReportPath) -or $ReportPath -match '^[a-zA-Z]:[^\\/]' -or $ReportPath -match '^[\\/][^\\/]') {
    throw '审查报告必须使用实际绝对路径。'
}
$report = Get-Item -LiteralPath $ReportPath -ErrorAction Stop
if ($report.PSIsContainer -or $report.Length -eq 0) {
    throw '审查报告必须是已有的非空文件。'
}

$orcaExecutable = (Get-Command orca.exe -CommandType Application -ErrorAction Stop).Source
$body = @(
    "Decision: $Decision"
    "Candidate SHA: $CandidateSHA"
    "Run: $RunId"
    "Reviewer Task: $TaskId"
    "Reviewer Dispatch: $DispatchId"
    "Reviewer terminal: $FromHandle"
    "Report: $($report.FullName)"
    "Reason: $Reason"
) -join [Environment]::NewLine
$sendArguments = @(
    'orchestration', 'send', '--from', $FromHandle,
    '--to', "run:$RunId", '--type', 'handoff', '--subject', 'ReviewDecision',
    '--body', $body, '--task-id', $TaskId, '--dispatch-id', $DispatchId,
    '--report-path', $report.FullName, '--json'
)

# 单次调用保留原生回执；未知送达由主控核查，不在此处重发。
$receiptText = & $orcaExecutable @sendArguments
$sendExitCode = $LASTEXITCODE
$receiptText | Write-Output
if ($sendExitCode -ne 0) {
    throw "Orca 审查通知失败，退出码为 $sendExitCode；请保留本次回执并核查送达状态。"
}
$receipt = ($receiptText -join [Environment]::NewLine) | ConvertFrom-Json -ErrorAction Stop
if ($receipt.ok -ne $true) {
    throw 'Orca 未确认审查通知入队；请保留原回执核查，禁止自动重发。'
}
Write-Output '审查通知已持久入队；主控收到并回复后才能确认唤起。'
