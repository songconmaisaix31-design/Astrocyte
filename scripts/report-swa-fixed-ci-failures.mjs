import { mkdir, writeFile } from 'node:fs/promises';

const owner = 'songconmaisaix31-design';
const repo = 'swarm-anything';
const runId = 40001114124;
const source = '6f40a399af301c05ecb79a5316a812641e05e3f9';
const outputDirectory = '.verification/swa-ci/main-ci-failures';
const maxLogBytes = 16 * 1024 * 1024;
const maxAnnotationBytes = 3500;
const jobs = [
  { id: 120577125594, name: 'swa-foundation (windows-2025)', label: 'windows-2025', step: 12, stepName: 'Exercise CLI validation and authority boundaries' },
  { id: 120577125738, name: 'check (windows-2025)', label: 'windows-2025', step: 8, stepName: 'Run pnpm check' },
  { id: 120577126306, name: 'check (ubuntu-24.04)', label: 'ubuntu-24.04', step: 8, stepName: 'Run pnpm check' },
];
const sensitive = /authorization|\bbearer\b|\bbasic\s+[A-Za-z0-9+/=]+|token|nonce|api[-_]?key|secret|password|cookie|credential/i;

class EvidenceError extends Error {
  constructor(code) { super(code); this.code = code; }
}

function requireEvidence(condition, code) {
  if (!condition) throw new EvidenceError(code);
}

export function validateRunMetadata(run) {
  requireEvidence(run?.id === runId && run.run_attempt === 1 && run.head_sha === source
    && run.path === '.github/workflows/check.yml' && run.name === 'SWA checks'
    && run.event === 'push' && run.status === 'completed' && run.conclusion === 'failure'
    && run.repository?.full_name === `${owner}/${repo}`
    && run.head_repository?.full_name === `${owner}/${repo}`, 'RUN_METADATA_MISMATCH');
}

export function validateJobMetadata(job) {
  const expected = jobs.find(item => item.id === job?.id);
  const failures = job?.steps?.filter(step => step.conclusion === 'failure');
  requireEvidence(expected && job.run_id === runId && job.run_attempt === 1 && job.head_sha === source
    && job.status === 'completed' && job.conclusion === 'failure' && job.name === expected.name
    && Array.isArray(job.labels) && job.labels.length === 1 && job.labels[0] === expected.label
    && failures?.length === 1 && failures[0].number === expected.step
    && failures[0].name === expected.stepName && failures[0].status === 'completed'
    && job.check_run_url === `https://api.github.com/repos/${owner}/${repo}/check-runs/${expected.id}`,
  'JOB_METADATA_MISMATCH');
}

function boundedText(text, limit) {
  const bytes = Buffer.from(text, 'utf8');
  if (bytes.length <= limit) return { text, truncated: false };
  let end = limit;
  while ((bytes[end] & 0xc0) === 0x80) end -= 1;
  return { text: bytes.subarray(0, end).toString('utf8'), truncated: true };
}

function safeMessage(line) {
  if (sensitive.test(line)) return { message_omitted_sensitive: true };
  const sanitized = line
    .replace(/https?:\/\/\S+/gi, '[URL omitted]')
    .replace(/[a-z]:[\\/]Users[\\/][^\\/\s"']+/gi, '[CI user home]')
    .replace(/\/(?:home|Users)\/[^/\s"']+/g, '[user home]')
    .replace(/\bS-\d-(?:\d+-){1,14}\d+\b/gi, '[SID omitted]')
    .replace(/\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b/gi, '[UUID omitted]')
    .replace(/\b[0-9a-f]{64}\b/gi, '[digest or secret omitted]')
    .replace(/[\x00-\x1f\x7f]/g, '');
  const result = boundedText(sanitized.trim(), 400);
  return { message: result.text, message_truncated: result.truncated };
}

function sourceLocations(line, checkStage) {
  const positions = [];
  let truncated = false;
  const pattern = /(?:^|[\s(])([A-Za-z0-9_./\\:-]+\.(?:go|[cm]?js|jsx|tsx?|ps1))(?::(\d+)(?::(\d+))?|\((\d+),(\d+)\))/g;
  for (const match of line.matchAll(pattern)) {
    let path = match[1].replaceAll('\\', '/').replace(/^\.\//, '');
    const lineNumber = Number(match[2] ?? match[4]);
    const column = match[3] ?? match[5];
    if (path.length > 256 || path.startsWith('/') || path.includes(':') || path.split('/').includes('..')
      || !/^[A-Za-z0-9_./-]+$/.test(path) || !Number.isSafeInteger(lineNumber) || lineNumber < 1
      || (column && (!Number.isSafeInteger(Number(column)) || Number(column) < 1))) continue;
    // 固定 process.mjs 在 web 运行这些工具；只保留实际文件名，不补成仓库路径。
    if (path.startsWith('src/') && ['tsc', 'eslint', 'vitest'].includes(checkStage)) path = path.split('/').at(-1);
    const kind = path.includes('/') ? 'repo_relative' : 'basename_only';
    if (kind === 'repo_relative' && !/^(?:cmd|internal|runtime|scripts|tests|web)\//.test(path)) continue;
    if (positions.length === 4) { truncated = true; break; }
    positions.push({ path, kind, line: lineNumber, ...(column ? { column: Number(column) } : {}) });
  }
  return { positions, truncated };
}

function observedCheckStage(line) {
  if (!line.startsWith('> ')) return null;
  // 仅识别固定失败来源中 process.mjs 输出的工具或脚本，不回传命令内容。
  if (/[/\\]go(?:\.exe)? vet -mod=readonly /i.test(line)) return 'go-vet';
  if (/[/\\]go(?:\.exe)? test -mod=readonly -count=1 /i.test(line)) {
    if (line.includes(' ./internal/adapters/orca -run ')) return 'go-test-orca';
    if (line.includes(' ./internal/workspace/app -run ')) return 'go-test-workspace';
    if (line.includes(' -skip ') && line.includes(' ./cmd/swa ./cmd/swa-server ')) return 'go-test-packages';
  }
  if (/[/\\]go(?:\.exe)? mod verify$/i.test(line)) return 'go-mod-verify';
  for (const name of ['check-dependencies', 'check-architecture', 'validate-examples', 'generate-client']) {
    if (line.includes(` scripts/${name}.mjs`) && / scripts\/[a-z-]+\.mjs(?: --(?:swa|check))?$/.test(line)) return name;
  }
  if (/ --test (?:scripts\/stop-process\.test\.mjs|tests\/swa-(?:product-acceptance\/driver-failure|cli-headless\/task-budget)\.test\.mjs)(?: |$)/.test(line)) return 'node-test';
  if (/[\\/]@redocly[\\/]cli[\\/]bin[\\/]cli\.js lint contracts\/openapi\.yaml --config contracts\/redocly\.yaml$/.test(line)) return 'redocly';
  if (/[\\/]typescript[\\/]bin[\\/]tsc -b$/.test(line)) return 'tsc';
  if (/[\\/]eslint[\\/]bin[\\/]eslint\.js \.$/.test(line)) return 'eslint';
  if (/[\\/]vitest[\\/]vitest\.mjs run$/.test(line)) return 'vitest';
  if (line === '> git diff --check') return 'git-diff-check';
  return null;
}

function projectLog(text, expected, common) {
  const candidates = [];
  const counts = { candidate_count: 0, retained_candidate_count: 0, location_count: 0, sensitive_message_omissions: 0, truncated_message_lines: 0, truncated_location_lines: 0 };
  let lastStage = 'unknown';
  let commandGroup = false;
  let failureBoundaryObserved = false;
  const lines = text.split(/\r?\n/);
  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index].replace(/^\d{4}-\d{2}-\d{2}T[\d:.]+Z /, '').replace(/\x1b\[[0-9;]*m/g, '');
    if (line.startsWith('##[group]Run ')) { commandGroup = true; continue; }
    if (line.startsWith('##[endgroup]')) { commandGroup = false; continue; }
    if (commandGroup) continue;
    if (/^##\[error\]Process completed with exit code 1\.$/.test(line)) failureBoundaryObserved = true;
    if (expected.step === 8 && !failureBoundaryObserved) lastStage = observedCheckStage(line) ?? lastStage;
    const test = /^\s*--- FAIL: (Test[A-Za-z0-9_]{0,124})(?=\/|\s|$)/.exec(line)?.[1];
    const { positions, truncated } = sourceLocations(line, lastStage);
    if (!test && positions.length === 0 && !/^\s*(?:##\[error\]|(?:[A-Za-z]*Error|error|fatal|panic):|FAIL(?:\s|$)|not ok\b|error TS\d+\b)/.test(line)) continue;
    counts.candidate_count += 1;
    if (truncated) counts.truncated_location_lines += 1;
    const locator = { log_line: index + 1, ...(test ? { go_test: test } : {}), ...(positions.length ? { source_locations: positions, source_locations_truncated: truncated } : {}) };
    // 动态 Go 子用例只在原日志中保留；输出使用根测试标识。
    const message = safeMessage(test ? `--- FAIL: ${test}` : line);
    if (message.message_omitted_sensitive) counts.sensitive_message_omissions += 1;
    if (message.message_truncated) counts.truncated_message_lines += 1;
    candidates.push({ priority: test ? 0 : positions.length ? 1 : 2, locator, message: { log_line: index + 1, ...message } });
    candidates.sort((a, b) => a.priority - b.priority || a.locator.log_line - b.locator.log_line);
    if (candidates.length > 40) candidates.pop();
  }
  counts.retained_candidate_count = candidates.length;
  const locations = candidates.filter(item => item.priority < 2).map(item => item.locator);
  counts.location_count = locations.reduce((total, item) => total + (item.go_test ? 1 : 0) + (item.source_locations?.length ?? 0), 0);
  const messages = candidates.map(item => item.message);
  const summary = {
    ...common, type: 'failure_locations', log_bytes: Buffer.byteLength(text), ...counts,
    candidate_limit_reached: counts.candidate_count > counts.retained_candidate_count,
    last_observed_check_stage: expected.step === 8 ? lastStage : 'not_applicable',
    failure_boundary_observed: failureBoundaryObserved,
    stage_source: 'scripts/check.mjs', stage_is_root_cause: false, root_cause: 'UNKNOWN',
    locations, omitted_location_records: 0,
  };
  while (Buffer.byteLength(JSON.stringify(summary)) > maxAnnotationBytes && summary.locations.length) {
    summary.locations.pop(); summary.omitted_location_records += 1;
  }
  const diagnostics = { ...common, type: 'failure_messages', root_cause: 'UNKNOWN', messages, omitted_message_records: 0 };
  while (Buffer.byteLength(JSON.stringify(diagnostics)) > maxAnnotationBytes && diagnostics.messages.length) {
    diagnostics.messages.pop(); diagnostics.omitted_message_records += 1;
  }
  requireEvidence(Buffer.byteLength(JSON.stringify(summary)) <= maxAnnotationBytes
    && Buffer.byteLength(JSON.stringify(diagnostics)) <= maxAnnotationBytes, 'PROJECTION_TOO_LARGE');
  return [summary, diagnostics];
}

function logText(response) {
  requireEvidence(response.status === 200, 'LOG_RESPONSE_STATUS');
  const data = response.data;
  requireEvidence(typeof data === 'string' || data instanceof ArrayBuffer || ArrayBuffer.isView(data), 'LOG_RESPONSE_TYPE');
  const bytes = typeof data === 'string' ? Buffer.from(data) : Buffer.from(data instanceof ArrayBuffer ? data : data.buffer, data.byteOffset ?? 0, data.byteLength);
  requireEvidence(bytes.length <= maxLogBytes, 'LOG_TOO_LARGE');
  const text = new TextDecoder('utf-8', { fatal: true }).decode(bytes);
  requireEvidence(/^\d{4}-\d{2}-\d{2}T[\d:.]+Z /.test(text) && !text.includes('\0'), 'LOG_BODY_INVALID');
  return text;
}

export default async function report({ github, context, core }) {
  let phase = 'context';
  let jobId = null;
  const readerSource = /^[0-9a-f]{40}$/.test(context.sha ?? '') ? context.sha : 'UNKNOWN';
  const base = { original_run: runId, original_source: source, reader_source: readerSource, read_only: true, original_conclusion: 'failure' };
  const request = route => github.request(route, { owner, repo, request: { signal: AbortSignal.timeout(30000) } });
  try {
    requireEvidence(context.repo.owner === owner && context.repo.repo === repo && readerSource !== 'UNKNOWN', 'READER_CONTEXT_MISMATCH');
    phase = 'run-metadata';
    validateRunMetadata((await request(`GET /repos/{owner}/{repo}/actions/runs/${runId}`)).data);
    phase = 'job-metadata';
    for (const expected of jobs) {
      jobId = expected.id;
      validateJobMetadata((await request(`GET /repos/{owner}/{repo}/actions/jobs/${expected.id}`)).data);
    }
    await mkdir(outputDirectory, { recursive: true, mode: 0o700 });
    for (const expected of jobs) {
      jobId = expected.id;
      phase = 'job-logs';
      const text = logText(await request(`GET /repos/{owner}/{repo}/actions/jobs/${expected.id}/logs`));
      phase = 'projection';
      const records = projectLog(text, expected, { ...base, original_job: expected.id });
      phase = 'evidence-save';
      await writeFile(`${outputDirectory}/${expected.id}.json`, `${JSON.stringify(records, null, 2)}\n`, { flag: 'wx', mode: 0o600 });
      for (const record of records) core.warning(JSON.stringify(record), { title: 'Fixed main CI failure evidence' });
    }
  } catch (error) {
    // 不输出异常对象；它可能携带认证请求、签名 URL 或原日志正文。
    const failure = { ...base, type: 'read_failure', phase, original_job: jobId,
      code: error instanceof EvidenceError ? error.code : 'READ_FAILED',
      http_status: Number.isInteger(error?.status) && error.status >= 100 && error.status <= 599 ? error.status : null,
      root_cause: 'UNKNOWN' };
    core.error(JSON.stringify(failure), { title: 'Fixed main CI evidence read stopped' });
    try {
      await mkdir(outputDirectory, { recursive: true, mode: 0o700 });
      await writeFile(`${outputDirectory}/read-failure.json`, `${JSON.stringify(failure, null, 2)}\n`, { flag: 'wx', mode: 0o600 });
    } catch {
      core.error('Fixed main CI safe failure metadata could not be saved');
    }
    core.setFailed('Fixed main CI evidence read stopped; no retry or original task execution was requested');
  }
}
