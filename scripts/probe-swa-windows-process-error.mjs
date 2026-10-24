import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { createHash, randomBytes } from 'node:crypto'
import { createWriteStream } from 'node:fs'
import { copyFile, cp, lstat, mkdir, open, readdir, realpath, writeFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { dirname, isAbsolute, join, relative, resolve } from 'node:path'
import { performance } from 'node:perf_hooks'
import { fileURLToPath, pathToFileURL } from 'node:url'
import { parseArgs } from 'node:util'

const sourceRevision = '5c25b1ce6cf044eb2bcea0c9bf7142a8f1e84850'
const productRevision = 'b464e70ba95967eb1ecc216f5f3d25407f1ecc66'
const retainedArchiveHash = '42909a812d9c9c08b363d803e8de0ef1db85d6a1c71a5b83e0730a4da4e51f46'
const approvedChanges = {
  'src/process_worker.cc': ['b418168745b8950bc5841cbabebdc314e7d3e27c1dad79ecfe051b9aa33c617f', '02adc3e67300a23e0f87ab63d55c9680e44bf86e3b846427e2475e9be6fcf9cc'],
  'lib/index.js': ['b615a3788e37bd6484cf66b0a231d8083c09b3c95fb2f5fc22efa2e3de3b61dd', 'a601d37ded7090f1314ff30e5841cc2cf2a3e6104cffb6d30ca0c8f940c5bc4c'],
  'lib/index.ts': ['589e64c4d518acbe38f13401423a3dac4c77b49773adb5886a30f7e35bb17b53', '9383d5d83afeccc86145aae71302a2fd1d6dcce3505900b32d4c8ab4ff51e5bd'],
  'typings/windows-process-tree.d.ts': ['a653368535baa4b66571dd8e5f945e3e3ebf2f275858f314055c90afb9f5da9a', 'cf68f5fc3a2218c5a66adc6f04cd2ae9b9d5eb2d30d6f0a5d89bf046ef0f82aa']
}
const checkout = resolve(fileURLToPath(new URL('../', import.meta.url)))
const { values } = parseArgs({ options: { 'pnpm-cli': { type: 'string' }, 'query-only': { type: 'boolean', default: false },
  'parent-budget-ms': { type: 'string' }, 'service-only': { type: 'boolean', default: false }, 'artifact-root': { type: 'string' } }, strict: true, allowPositionals: false })
const serviceOnly = values['service-only']
assert.ok(!(serviceOnly && values['query-only']))
if (!serviceOnly) assert.equal(values['artifact-root'], undefined)
const sha = bytes => createHash('sha256').update(bytes).digest('hex')
const pairs = { Win32: ['load-system-module', 'resolve-api', 'image-conversion'],
  NTSTATUS: ['buffer-limit', 'query', 'buffer-attempts'], InvalidBuffer: ['returned-length', 'buffer-address', 'record-header',
    'process-id', 'image-length', 'image-address', 'image-limit', 'record-offset', 'empty-snapshot'] }
const positive = value => Number.isSafeInteger(value) && value > 0
const uint = value => Number.isInteger(value) && value >= 0 && value <= 0xffffffff
const contains = (parent, path) => { const part = relative(parent, path); return !part || !isAbsolute(part) && part !== '..' && !part.startsWith('../') && !part.startsWith('..\\') }
const keys = (value, names) => { assert.ok(value && typeof value === 'object' && !Array.isArray(value)); assert.deepEqual(Object.keys(value).sort(), [...names].sort()) }
const instant = () => ({ observed_at_ms: Date.now(), monotonic_ms: performance.now() })
const fail = code => Object.assign(new Error('The bounded native diagnostic did not complete'), { diagnosticCode: code })
const safeCode = error => ['DEADLINE', 'COMMAND_FAILED', 'OUTPUT_BOUND', 'CALLBACK_TIMEOUT', 'BINDING_FAILED', 'SAFE_RECORD_FAILED', 'FILE_CLOSE_FAILED'].includes(error?.diagnosticCode) ? error.diagnosticCode : 'UNKNOWN'
function remaining(deadline) { const value = deadline - performance.now(); if (!(value > 0)) throw fail('DEADLINE'); return value }
async function within(deadline, action) {
  let timer
  try {
    const result = await Promise.race([Promise.resolve().then(() => { remaining(deadline); return action() }),
      new Promise((_, reject) => { timer = setTimeout(() => reject(fail('DEADLINE')), remaining(deadline)) })])
    remaining(deadline)
    return result
  } finally { clearTimeout(timer) }
}
async function ordinary(path, directory, copied = false) {
  path = resolve(path)
  let cursor = path
  while (cursor) {
    const item = await lstat(cursor)
    assert.ok(!item.isSymbolicLink())
    assert.equal((await realpath(cursor)).toLowerCase(), cursor.toLowerCase())
    if (cursor === path) { assert.ok(directory ? item.isDirectory() : item.isFile()); if (copied && !directory) assert.equal(item.nlink, 1) }
    const parent = dirname(cursor)
    if (parent === cursor) break
    cursor = parent
  }
  return path
}
async function stable(path, maximum = 1024 * 1024, copied = false) {
  await ordinary(path, false, copied)
  const before = await lstat(path)
  assert.ok(before.isFile() && Number.isSafeInteger(before.size) && before.size >= 0 && before.size <= maximum)
  const same = value => {
    assert.ok(value.isFile()); if (copied) assert.equal(value.nlink, 1)
    for (const key of ['dev', 'ino', 'mode', 'nlink', 'uid', 'gid', 'size', 'mtimeMs', 'ctimeMs', 'birthtimeMs']) assert.equal(value[key], before[key])
  }
  let handle, bytes, failure
  try {
    handle = await open(path, 'r')
    same(await handle.stat())
    const pieces = [], limit = Math.min(before.size + 1, maximum + 1)
    let position = 0
    while (position < limit) {
      const buffer = Buffer.allocUnsafe(Math.min(64 * 1024, limit - position))
      const actual = await handle.read(buffer, 0, buffer.length, position)
      if (!actual.bytesRead) break
      position += actual.bytesRead
      assert.ok(position <= before.size && position <= maximum)
      pieces.push(buffer.subarray(0, actual.bytesRead))
    }
    bytes = Buffer.concat(pieces)
    assert.equal(bytes.length, before.size)
    same(await handle.stat())
    await ordinary(path, false, copied)
    same(await lstat(path))
  } catch (error) { failure = error }
  try { if (handle) await handle.close() } catch (error) {
    if (failure) failure = Object.assign(new AggregateError([failure, error], 'The bounded file read and close did not complete', { cause: failure }),
      { diagnosticCode: failure.diagnosticCode, diagnosticSecondaryCodes: ['FILE_CLOSE_FAILED'] })
    else failure = Object.assign(error, { diagnosticCode: 'FILE_CLOSE_FAILED' })
  }
  if (failure) throw failure
  return bytes
}
function nativeDiagnostic(value) {
  if (value === undefined) return 'NOT_CAPTURED'
  keys(value, ['phase', 'kind', 'code'])
  assert.equal(typeof value.kind, 'string'); assert.equal(typeof value.phase, 'string')
  assert.ok(Object.hasOwn(pairs, value.kind) && pairs[value.kind].includes(value.phase) && uint(value.code))
  return { phase: value.phase, kind: value.kind, code: value.code }
}
function sample(value) {
  if (value === null) return
  keys(value, ['sequence', 'start', 'end', 'row_count', 'self', 'native'])
  assert.ok(positive(value.sequence) && positive(value.row_count))
  for (const at of [value.start, value.end]) { keys(at, ['observed_at_ms', 'monotonic_ms']); assert.ok(positive(at.observed_at_ms) && Number.isFinite(at.monotonic_ms) && at.monotonic_ms >= 0) }
  assert.ok(value.end.monotonic_ms >= value.start.monotonic_ms)
  keys(value.self, ['pid', 'creationTimeMs']); assert.ok(positive(value.self.pid) && positive(value.self.creationTimeMs))
  keys(value.native, ['provider', 'version', 'ntstatus', 'snapshot_complete'])
  assert.equal(value.native.provider, 'NtQuerySystemInformation'); assert.equal(value.native.version, 2)
  assert.ok(uint(value.native.ntstatus) && value.native.ntstatus <= 0x7fffffff); assert.equal(value.native.snapshot_complete, true)
}
function queryRecord(value) {
  keys(value, ['schema_version', 'status', 'completed_count', 'first', 'last', 'failure', 'model_calls', 'service_started'])
  assert.equal(value.schema_version, 1); assert.ok(['NOT_REPRODUCED', 'FAIL'].includes(value.status))
  assert.ok(Number.isSafeInteger(value.completed_count) && value.completed_count >= 0 && value.completed_count <= 512)
  assert.equal(value.model_calls, 0); assert.equal(value.service_started, false)
  sample(value.first); sample(value.last)
  assert.equal(value.completed_count === 0, value.first === null); assert.equal(value.completed_count === 0, value.last === null)
  if (value.first) { assert.equal(value.first.sequence, 1); assert.equal(value.last.sequence, value.completed_count) }
  if (value.failure !== null) {
    keys(value.failure, ['sequence', 'start', 'end', 'callback_kind', 'native_snapshot', 'error_kind'])
    assert.equal(value.failure.sequence, value.completed_count + 1)
    for (const at of [value.failure.start, value.failure.end]) { keys(at, ['observed_at_ms', 'monotonic_ms']); assert.ok(positive(at.observed_at_ms) && Number.isFinite(at.monotonic_ms)) }
    assert.ok(['NOT_CAPTURED', 'undefined', 'null', 'array', 'object', 'boolean', 'number', 'string', 'function', 'symbol', 'bigint'].includes(value.failure.callback_kind))
    if (value.failure.native_snapshot !== 'NOT_CAPTURED') nativeDiagnostic(value.failure.native_snapshot)
    assert.ok(['CALLBACK_TIMEOUT', 'CALLBACK_UNAVAILABLE', 'BATCH_INVALID', 'DIAGNOSTIC_INVALID', 'UNKNOWN'].includes(value.failure.error_kind))
    assert.equal(value.status, 'FAIL')
  } else assert.equal(value.status, 'NOT_REPRODUCED')
}
async function queryOnly(runtime) {
  const budget = Number(values['parent-budget-ms'])
  assert.ok(Number.isFinite(budget) && budget > 0 && budget <= 30_000)
  const overall = performance.now() + budget, queryDeadline = Math.min(overall, performance.now() + 20_000)
  const result = { schema_version: 1, status: 'NOT_REPRODUCED', completed_count: 0, first: null, last: null,
    failure: null, model_calls: 0, service_started: false }
  const requireRuntime = createRequire(join(runtime, 'package.json'))
  const native = requireRuntime('@vscode/windows-process-tree')
  const identity = { pid: process.pid, creationTimeMs: native.getProcessCreationTime(process.pid) }
  assert.ok(positive(identity.creationTimeMs)); remaining(queryDeadline)
  while (result.completed_count < 512 && performance.now() < queryDeadline) {
    const start = instant(), sequence = result.completed_count + 1
    let captured, ended, failureKind = 'UNKNOWN'
    try {
      const callbackDeadline = Math.min(queryDeadline, performance.now() + 5000)
      try {
        captured = await within(callbackDeadline, () => new Promise(resolveCallback => {
          native.getAllProcesses((rows, diagnostic) => { resolveCallback({ rows, diagnostic, at: instant() }) }, 4)
        }))
      } catch (error) { failureKind = error?.diagnosticCode === 'DEADLINE' ? 'CALLBACK_TIMEOUT' : 'UNKNOWN'; throw error }
      ended = captured.at
      failureKind = 'DIAGNOSTIC_INVALID'
      const diagnostic = nativeDiagnostic(captured.diagnostic)
      failureKind = 'CALLBACK_UNAVAILABLE'
      assert.ok(Array.isArray(captured.rows))
      failureKind = 'BATCH_INVALID'
      assert.equal(diagnostic, 'NOT_CAPTURED')
      const self = captured.rows.filter(row => row?.pid === process.pid)
      assert.equal(self.length, 1); assert.equal(self[0].creationTimeMs, identity.creationTimeMs)
      assert.equal(native.getProcessCreationTime(process.pid), identity.creationTimeMs)
      assert.equal(self[0].ppid, process.ppid)
      for (const row of captured.rows) {
        assert.ok(row && typeof row === 'object' && uint(row.pid) && positive(row.pid) && uint(row.ppid))
        assert.equal(row.creationTimeProvider, 'NtQuerySystemInformation'); assert.equal(row.creationTimeDiagnosticVersion, 2)
        assert.equal(row.creationTimeSnapshotComplete, true)
        assert.ok(uint(row.creationTimeNtStatus) && row.creationTimeNtStatus <= 0x7fffffff)
        assert.equal(row.creationTimeNtStatus, self[0].creationTimeNtStatus)
      }
      remaining(queryDeadline)
      const actual = { sequence, start, end: ended, row_count: captured.rows.length, self: identity,
        native: { provider: self[0].creationTimeProvider, version: self[0].creationTimeDiagnosticVersion,
          ntstatus: self[0].creationTimeNtStatus, snapshot_complete: self[0].creationTimeSnapshotComplete } }
      sample(actual); result.completed_count++; result.first ??= actual; result.last = actual
    } catch {
      let diagnostic = 'NOT_CAPTURED'
      try { if (captured) diagnostic = nativeDiagnostic(captured.diagnostic) } catch { failureKind = 'DIAGNOSTIC_INVALID' }
      result.status = 'FAIL'
      result.failure = { sequence, start, end: ended ?? instant(),
        callback_kind: !captured ? 'NOT_CAPTURED' : captured.rows === null ? 'null' : Array.isArray(captured.rows) ? 'array' : typeof captured.rows,
        native_snapshot: diagnostic, error_kind: failureKind }
      break
    }
    const pause = Math.min(25, Math.max(0, queryDeadline - performance.now()))
    if (pause) await within(overall, () => new Promise(resolvePause => setTimeout(resolvePause, pause)))
  }
  queryRecord(result)
  const bytes = Buffer.from(JSON.stringify(result))
  assert.ok(bytes.length <= 64 * 1024)
  await within(overall, () => new Promise((resolveWrite, reject) => process.stdout.write(bytes, error => error ? reject(error) : resolveWrite())))
  process.exitCode = result.status === 'NOT_REPRODUCED' ? 0 : 1
}

async function main() {
  assert.equal(process.platform, 'win32'); assert.equal(process.arch, 'x64')
  assert.ok(isAbsolute(process.env.RUNNER_TEMP ?? ''))
  const temporary = await ordinary(process.env.RUNNER_TEMP, true)
  const work = join(temporary, 'swa-native-error-5c25'), source = join(work, 'source-5c25')
  const runtime = join(work, 'development/runtime'), privateRoot = join(work, 'private')
  if (values['query-only']) { await ordinary(runtime, true); return queryOnly(runtime) }
  assert.equal(values['parent-budget-ms'], undefined)
  assert.ok(isAbsolute(values['pnpm-cli'] ?? '')); const pnpm = await ordinary(values['pnpm-cli'], false)
  const evidence = join(checkout, '.verification/swa-ci/windows-native-error')
  assert.ok(!contains(checkout, work) && !contains(work, checkout))
  for (const path of [join(checkout, '.verification'), join(checkout, '.verification/swa-ci')]) {
    try { await mkdir(path) } catch (error) { if (error.code !== 'EEXIST') throw error }
    await ordinary(path, true)
  }
  await mkdir(evidence); await mkdir(work)
  for (const name of ['source-5c25', 'private', 'development', 'home', 'temp', 'cache']) await mkdir(join(work, name))
  const env = {}
  for (const key of Object.keys(process.env)) if (/^(?:systemroot|windir|comspec|pathext|programfiles|programfiles\(x86\)|programw6432|systemdrive|number_of_processors|processor_architecture|os)$/i.test(key)) env[key] = process.env[key]
  const paths = Object.keys(process.env).filter(key => key.toLowerCase() === 'path'); assert.equal(paths.length, 1)
  Object.assign(env, { PATH: process.env[paths[0]], RUNNER_TEMP: temporary, HOME: join(work, 'home'), USERPROFILE: join(work, 'home'),
    APPDATA: join(work, 'home'), LOCALAPPDATA: join(work, 'home'), TEMP: join(work, 'temp'), TMP: join(work, 'temp'),
    TMPDIR: join(work, 'temp'), GIT_CONFIG_NOSYSTEM: '1', GIT_TERMINAL_PROMPT: '0', GIT_CEILING_DIRECTORIES: work })
  const result = { schema_version: 1, status: 'FAIL', phase: 'preflight', source_revision: sourceRevision,
    observation_revision: process.env.GITHUB_SHA, model_calls: 0, service_started: false, other_native_components_built: false,
    kernel_tree_exit_proven: false, commands: [], query: 'NOT_CAPTURED', query_process: 'NOT_CAPTURED', error_kind: null, secondary_error_kinds: [] }
  let counter = 0, finalDeadline, primary, warningAttempts = 0
  const save = async (name, value, deadline) => {
    const bytes = Buffer.from(typeof value === 'string' ? value : JSON.stringify(value, null, 2) + '\n')
    assert.ok(bytes.length <= 64 * 1024)
    if (deadline) remaining(deadline)
    await writeFile(join(evidence, name), bytes, { flag: 'wx' })
    if (deadline) remaining(deadline)
    assert.ok(bytes.equals(await stable(join(evidence, name), 64 * 1024, true)))
    if (deadline) remaining(deadline)
  }
  async function command(executable, args, cwd, milliseconds, options = {}) {
    const id = ++counter, stage = options.stage ?? result.phase, deadline = options.deadline ?? performance.now() + milliseconds
    const output = { stage, exit: null, signal: null, timed_out: false, close_observed: false, stdout_eof: false, stderr_eof: false,
      stdout_bytes: 0, stderr_bytes: 0, stream_failure: false, spawn_error: false, output_complete: true }
    const streams = [createWriteStream(join(privateRoot, `${id}.stdout.raw`), { flags: 'wx' }), createWriteStream(join(privateRoot, `${id}.stderr.raw`), { flags: 'wx' })]
    for (const stream of streams) stream.on('error', () => { output.stream_failure = true })
    const before = instant(); let child, timer, forced, captured = []
    let settled
    const closed = new Promise(resolveClose => { settled = resolveClose })
    try {
      remaining(deadline)
      child = spawn(executable, args, { cwd, env, stdio: ['ignore', 'pipe', 'pipe'], windowsHide: true })
      child.once('error', () => { output.spawn_error = true })
      child.once('close', (code, signal) => { output.exit = Number.isInteger(code) ? code : null
        output.signal = signal === null ? null : ['SIGTERM', 'SIGKILL', 'SIGINT', 'SIGABRT', 'SIGHUP', 'SIGPIPE'].includes(signal) ? signal : 'UNKNOWN'
        output.close_observed = true; settled() })
      for (const [index, stream] of [child.stdout, child.stderr].entries()) {
        const key = index === 0 ? 'stdout' : 'stderr'
        stream.on('data', bytes => {
          output[key + '_bytes'] += bytes.length
          if (output[key + '_bytes'] <= 4 * 1024 * 1024) streams[index].write(bytes)
          else { output.output_complete = false; if (!forced) { forced = true; child.kill() } }
          if (index === 0 && options.capture && output.stdout_bytes <= 64 * 1024) captured.push(bytes)
        })
        stream.once('end', () => { output[key + '_eof'] = true })
        stream.once('error', () => { output.stream_failure = true })
      }
      // 关闭请求预留在原绝对界内；未见 close 时仍报告 UNKNOWN。
      timer = setTimeout(() => { output.timed_out = true; forced = true; child.kill() }, Math.max(1, remaining(deadline) - 1000))
      await within(deadline, () => closed)
    } catch (error) {
      output.timed_out ||= error?.diagnosticCode === 'DEADLINE'
      if (child && !output.close_observed) { if (!forced) child.kill(); child.stdout.destroy(); child.stderr.destroy(); child.unref() }
    } finally {
      clearTimeout(timer)
      for (const stream of streams) stream.end()
      try { await within(deadline, () => Promise.all(streams.map(stream => stream.closed ? Promise.resolve() : new Promise(resolveClose => stream.once('close', resolveClose))))) }
      catch { output.stream_failure = true }
    }
    const receipt = { ...output, start: before, end: instant() }
    result.commands.push(receipt)
    const bytes = options.capture && output.stdout_bytes <= 64 * 1024 ? Buffer.concat(captured) : null
    if (!options.allowFailure && (output.exit !== 0 || !output.close_observed || output.timed_out || !output.stdout_eof || !output.stderr_eof || output.stream_failure || output.spawn_error || !output.output_complete)) throw fail('COMMAND_FAILED')
    return { receipt, bytes, started: !!child && positive(child.pid) }
  }
  const git = async args => {
    const value = await command('git', ['-c', 'core.autocrlf=false', ...args], checkout, 30_000, { capture: true })
    assert.ok(value.bytes); return value.bytes
  }
  async function tree(directory, copied = false, prefix = '') {
    const files = []
    for (const entry of await readdir(directory, { withFileTypes: true })) {
      if (!prefix && ['build', 'node_modules', 'deps'].includes(entry.name)) continue
      const path = join(directory, entry.name), name = prefix + entry.name
      assert.ok(!entry.isSymbolicLink())
      if (entry.isDirectory()) files.push(...await tree(path, copied, name + '/'))
      else { const bytes = await stable(path, 2 * 1024 * 1024, copied); files.push({ name, sha256: sha(bytes), bytes: bytes.length }) }
    }
    return files.sort((a, b) => a.name.localeCompare(b.name))
  }
  let serviceMessages = []
  async function serviceWindow({ node, powershell, binary, diagnosticPackage, sourceGuard }) {
    const compositionDeadline = performance.now() + 240_000
    const runService = await within(compositionDeadline, async () => {
    result.product_revision = productRevision
    result.scope = 'retained_b464_development_service_not_candidate'
    const artifact = resolve(values['artifact-root'] ?? '')
    assert.equal(artifact.toLowerCase(), join(temporary, 'swa-b464-retained').toLowerCase())
    await ordinary(artifact, true)
    const metadata = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(await stable(join(artifact, 'private/artifact-metadata.json'), 4096, true)))
    keys(metadata, ['id', 'name', 'size_in_bytes', 'digest', 'expired', 'workflow_run'])
    keys(metadata.workflow_run, ['id', 'head_sha'])
    assert.equal(metadata.id, 12542236014)
    assert.equal(metadata.name, 'swa-windows-combined-' + productRevision)
    assert.equal(metadata.size_in_bytes, 310804282)
    assert.equal(metadata.digest, 'sha256:3579aec40c304d64255a9469eb6c69cb72112f8a5a6f43c30f9636d3e7b1c4d0')
    assert.equal(metadata.expired, false); assert.equal(metadata.workflow_run.id, 40047585581)
    assert.equal(metadata.workflow_run.head_sha, productRevision)
    const base = 'swa-windows-combined-' + productRevision, archive = join(artifact, base + '.zip')
    assert.deepEqual((await readdir(artifact)).sort(), ['private', base + '.zip', base + '.zip.sha256'].sort())
    assert.equal(sha(await stable(archive, 512 * 1024 * 1024, true)), retainedArchiveHash)
    const originalSidecar = await stable(archive + '.sha256', 1024, true)
    assert.equal(new TextDecoder('utf-8', { fatal: true }).decode(originalSidecar).trim(), retainedArchiveHash + '  ' + base + '.zip')
    result.retained_product = { metadata, archive_sha256: retainedArchiveHash, version: productRevision, installation: 'NOT_RUN' }
    result.phase = 'retained-product-extraction'
    const extracted = join(work, 'retained-product'), manifestPath = join(privateRoot, 'retained-entries.json')
    const extractScript = join(privateRoot, 'extract-retained.ps1')
    // 复用正式包的 .NET 提取器，先完整核实际条目和空间，再允许一次独占写入。
    const extractText = [
      'param([string]$Archive,[string]$Target,[string]$Manifest,[string]$RootName)',
      "$ErrorActionPreference='Stop'",
      'try {',
      'if (Test-Path -LiteralPath $Target) { throw "Target exists" }',
      '$zip=[IO.Compression.ZipFile]::OpenRead($Archive)',
      'try {',
      'if ($zip.Entries.Count -gt 65536 -or $zip.Entries.Count -lt 1) { throw "Entry bound" }',
      '$names=[Collections.Generic.HashSet[string]]::new([StringComparer]::OrdinalIgnoreCase)',
      '$files=[Collections.Generic.List[object]]::new(); [long]$total=0; [long]$maximum=0',
      'foreach($entry in $zip.Entries) {',
      '$name=$entry.FullName; $parts=$name.Split("/")',
      'if (-not $name.StartsWith($RootName+"/",[StringComparison]::Ordinal) -or $name.Contains("\\") -or $name.Contains(":") -or [IO.Path]::IsPathRooted($name) -or $parts -contains ".." -or $parts -contains "." -or -not $names.Add($name.TrimEnd("/"))) { throw "Entry path" }',
      '$isDir=$name.EndsWith("/"); $relative=$name.Substring($RootName.Length+1)',
      'if ($name.Contains("//") -or (-not $relative -and -not $isDir)) { throw "Empty entry" }',
      '[long]$attributes=$entry.ExternalAttributes; $unix=($attributes -shr 16) -band 61440',
      'if (($attributes -band 1024) -ne 0 -or ($unix -notin @(0,32768,16384)) -or ($isDir -and $unix -eq 32768) -or (-not $isDir -and $unix -eq 16384)) { throw "Entry type" }',
      'if ($entry.Length -lt 0 -or $entry.CompressedLength -lt 0 -or $entry.Length -gt 268435456 -or ($isDir -and $entry.Length -ne 0)) { throw "Entry size" }',
      '$targetPath=[IO.Path]::GetFullPath([IO.Path]::Combine($Target,$name))',
      '$prefix=[IO.Path]::GetFullPath($Target).TrimEnd("\\")+"\\"',
      'if (-not $targetPath.StartsWith($prefix,[StringComparison]::OrdinalIgnoreCase)) { throw "Entry escape" }',
      '$total += $entry.Length; if ($total -gt 2147483648) { throw "Expanded bound" }; $maximum=[Math]::Max($maximum,$entry.Length)',
      'if (-not $isDir) { $files.Add([ordered]@{name=$relative;bytes=[long]$entry.Length}) }',
      '}',
      '$entryCount=$zip.Entries.Count',
      '} finally { $zip.Dispose() }',
      '$free=[IO.DriveInfo]::new([IO.Path]::GetPathRoot($Target)).AvailableFreeSpace',
      '$needed=2*$total+(Get-Item -LiteralPath $Archive).Length+1073741824',
      'if ($free -lt $needed) { throw "Space bound" }',
      '[IO.Compression.ZipFile]::ExtractToDirectory($Archive,$Target,$false)',
      '$report=[ordered]@{schema_version=1;entry_count=$entryCount;expanded_bytes=$total;maximum_file_bytes=$maximum;free_before=$free;needed_before=$needed;files=@($files)}',
      '$bytes=[Text.UTF8Encoding]::new($false).GetBytes(($report|ConvertTo-Json -Depth 5 -Compress))',
      'if ($bytes.Length -gt 33554432) { throw "Manifest bound" }',
      '$stream=[IO.File]::Open($Manifest,[IO.FileMode]::CreateNew,[IO.FileAccess]::Write,[IO.FileShare]::None)',
      'try {$stream.Write($bytes,0,$bytes.Length);$stream.Flush($true)} finally {$stream.Dispose()}',
      '} catch {[Console]::Error.WriteLine("retained-product-extraction-failed");exit 1}'
    ].join('\n')
    remaining(compositionDeadline)
    await writeFile(extractScript, extractText, { flag: 'wx' })
    await command(powershell, ['-NoProfile', '-NonInteractive', '-File', extractScript, '-Archive', archive, '-Target', extracted, '-Manifest', manifestPath, '-RootName', base], work, 120_000,
      { deadline: Math.min(compositionDeadline, performance.now() + 120_000) })
    remaining(compositionDeadline)
    const manifest = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(await stable(manifestPath, 32 * 1024 * 1024, true)))
    keys(manifest, ['schema_version', 'entry_count', 'expanded_bytes', 'maximum_file_bytes', 'free_before', 'needed_before', 'files'])
    assert.equal(manifest.schema_version, 1)
    for (const key of ['entry_count', 'expanded_bytes', 'maximum_file_bytes', 'free_before', 'needed_before']) assert.ok(positive(manifest[key]))
    assert.ok(manifest.entry_count <= 65536 && manifest.expanded_bytes <= 2 * 1024 ** 3 && manifest.maximum_file_bytes <= 256 * 1024 ** 2)
    assert.ok(manifest.free_before >= manifest.needed_before && Array.isArray(manifest.files) && manifest.files.length <= manifest.entry_count)
    const expected = new Map(), folded = new Set()
    for (const entry of manifest.files) {
      keys(entry, ['name', 'bytes'])
      assert.ok(typeof entry.name === 'string' && entry.name && !entry.name.includes('\\') && !entry.name.includes(':'))
      assert.ok(!isAbsolute(entry.name) && entry.name.split('/').every(part => part && part !== '.' && part !== '..'))
      assert.ok(Number.isSafeInteger(entry.bytes) && entry.bytes >= 0 && entry.bytes <= 256 * 1024 ** 2)
      assert.ok(!folded.has(entry.name.toLowerCase())); folded.add(entry.name.toLowerCase()); expected.set(entry.name, entry.bytes)
    }
    assert.equal([...expected.values()].reduce((sum, value) => sum + value, 0), manifest.expanded_bytes)
    const development = join(extracted, base), replacementSizes = new Map()
    assert.deepEqual(await readdir(extracted), [base])
    remaining(compositionDeadline)
    async function productTree(root, deadline, prefix = '', changed = false) {
      const files = []
      remaining(deadline)
      await ordinary(root, true)
      remaining(deadline)
      const entries = await readdir(root, { withFileTypes: true })
      remaining(deadline)
      for (const entry of entries) {
        remaining(deadline)
        const name = prefix + entry.name, path = join(root, entry.name)
        assert.ok(!entry.isSymbolicLink())
        if (entry.isDirectory()) {
          const nested = await productTree(path, deadline, name + '/', changed)
          remaining(deadline)
          files.push(...nested)
        }
        else {
          assert.ok(expected.has(name)); const bytes = await stable(path, 256 * 1024 ** 2, true)
          remaining(deadline)
          assert.equal(bytes.length, changed && replacementSizes.has(name) ? replacementSizes.get(name) : expected.get(name))
          const item = { name, bytes: bytes.length, sha256: sha(bytes) }
          remaining(deadline)
          files.push(item)
        }
      }
      files.sort((a, b) => a.name.localeCompare(b.name))
      remaining(deadline)
      return files
    }
    const originalFiles = await productTree(development, compositionDeadline)
    assert.equal(originalFiles.length, expected.size)
    const baselineHash = sha(Buffer.from(JSON.stringify(originalFiles)))
    remaining(compositionDeadline)
    result.retained_product.extraction = { entry_count: manifest.entry_count, file_count: originalFiles.length, expanded_bytes: manifest.expanded_bytes,
      maximum_file_bytes: manifest.maximum_file_bytes, free_before: manifest.free_before,
      files_sha256: baselineHash, product_inputs_unchanged: false }
    for (const [name, value] of [['VERSION', productRevision], ['PLATFORM', 'win32-x64'], ['EDITION', 'combined']]) {
      remaining(compositionDeadline)
      assert.equal(new TextDecoder('utf-8', { fatal: true }).decode(await stable(join(development, name), 1024, true)).trim(), value)
      remaining(compositionDeadline)
    }
    const serviceNode = join(development, 'node/node.exe')
    const serviceNodeBytes = await stable(serviceNode, 128 * 1024 ** 2, true)
    remaining(compositionDeadline)
    const pinnedNodeBytes = await stable(node, 128 * 1024 ** 2)
    remaining(compositionDeadline)
    assert.equal(sha(serviceNodeBytes), sha(pinnedNodeBytes))
    const requireDevelopment = createRequire(join(development, 'runtime/package.json'))
    const servicePackage = await realpath(dirname(requireDevelopment.resolve('@vscode/windows-process-tree/package.json')))
    remaining(compositionDeadline)
    assert.ok(contains(join(development, 'runtime'), servicePackage))
    const packageJSON = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(await stable(join(servicePackage, 'package.json'), 65536, true)))
    remaining(compositionDeadline)
    assert.equal(packageJSON.version, '0.8.0'); assert.equal(packageJSON.main, 'lib/index.js')
    const serviceJS = join(servicePackage, 'lib/index.js'), serviceBinary = join(servicePackage, 'build/Release/windows_process_tree.node')
    assert.equal(sha(await stable(serviceJS, 65536, true)), approvedChanges['lib/index.js'][0])
    remaining(compositionDeadline)
    assert.equal(sha(await stable(serviceBinary, 16 * 1024 ** 2, true)), '8b2d203527225d12e67920484aca4a07be35c8675b477e6a6fe84f2b608b389f')
    remaining(compositionDeadline)
    await copyFile(join(diagnosticPackage, 'lib/index.js'), serviceJS)
    remaining(compositionDeadline)
    await copyFile(binary, serviceBinary)
    remaining(compositionDeadline)
    const changedFiles = new Map([
      [relative(development, serviceJS).replaceAll('\\', '/'), approvedChanges['lib/index.js'][1]],
      [relative(development, serviceBinary).replaceAll('\\', '/'), result.native_sha256]
    ])
    replacementSizes.set(relative(development, serviceJS).replaceAll('\\', '/'), (await lstat(serviceJS)).size)
    remaining(compositionDeadline)
    replacementSizes.set(relative(development, serviceBinary).replaceAll('\\', '/'), (await lstat(serviceBinary)).size)
    remaining(compositionDeadline)
    const guardArchive = async deadline => {
      remaining(deadline)
      const archiveBytes = await stable(archive, 512 * 1024 ** 2, true)
      remaining(deadline)
      assert.equal(sha(archiveBytes), retainedArchiveHash)
      remaining(deadline)
      const sidecarBytes = await stable(archive + '.sha256', 1024, true)
      remaining(deadline)
      assert.ok(sidecarBytes.equals(originalSidecar))
      assert.equal(new TextDecoder('utf-8', { fatal: true }).decode(sidecarBytes).trim(), retainedArchiveHash + '  ' + base + '.zip')
      remaining(deadline)
    }
    const guardReplacements = async deadline => {
      for (const [name, hash] of changedFiles) {
        remaining(deadline)
        const bytes = await stable(join(development, name), 16 * 1024 ** 2, true)
        remaining(deadline)
        assert.equal(bytes.length, replacementSizes.get(name)); assert.equal(sha(bytes), hash)
        remaining(deadline)
      }
    }
    const guardProduct = async deadline => {
      await guardArchive(deadline)
      const copied = await productTree(development, deadline, '', true)
      remaining(deadline)
      assert.equal(copied.length, originalFiles.length)
      for (let index = 0; index < copied.length; index++) {
        remaining(deadline)
        assert.equal(copied[index].name, originalFiles[index].name)
        assert.equal(copied[index].sha256, changedFiles.get(copied[index].name) ?? originalFiles[index].sha256)
      }
    }
    // 只替换两个观察消费者；实际产品 VERSION 仍为 B464。
    const observerNames = ['serve-lifecycle.mjs', 'windows-service-evidence-summary.mjs']
    const consumerNames = [...observerNames, 'controller-taskcreate-check.mjs', 'controller-input-check.mjs', 'process-incarnation-closure.mjs', 'observe-process-exits.ps1']
    for (const name of consumerNames) {
      remaining(compositionDeadline)
      assert.equal((await git(['diff', '--name-only', productRevision, sourceRevision, '--', 'tests/swa-cli-headless/' + name])).toString().trim(), '')
      remaining(compositionDeadline)
    }
    for (const name of observerNames) {
      remaining(compositionDeadline)
      await copyFile(join(checkout, 'tests/swa-cli-headless', name), join(source, 'tests/swa-cli-headless', name))
      remaining(compositionDeadline)
    }
    const consumerPins = await Promise.all(consumerNames.map(async name => {
      remaining(compositionDeadline)
      const bytes = await stable(join(source, 'tests/swa-cli-headless', name), 256 * 1024, true)
      remaining(compositionDeadline)
      return { name, sha256: sha(bytes) }
    }))
    remaining(compositionDeadline)
    result.service_consumer_sources = consumerPins
    const guardConsumers = async deadline => {
      for (const pin of consumerPins) {
        remaining(deadline)
        assert.equal(sha(await stable(join(source, 'tests/swa-cli-headless', pin.name), 256 * 1024, true)), pin.sha256)
        remaining(deadline)
      }
      for (const name of observerNames) {
        remaining(deadline)
        assert.equal(sha(await stable(join(checkout, 'tests/swa-cli-headless', name), 256 * 1024)), consumerPins.find(pin => pin.name === name).sha256)
        remaining(deadline)
      }
    }
    // 基线已完整核验；本轮只改两个叶文件，服务结束后再作完整比较。
    await guardReplacements(compositionDeadline); await guardArchive(compositionDeadline)
    await guardConsumers(compositionDeadline); remaining(compositionDeadline)
    await sourceGuard(); remaining(compositionDeadline)
    const serviceEvidence = join(source, '.verification/swa-ci/windows'), smoke = join(serviceEvidence, 'smoke')
    remaining(compositionDeadline)
    await mkdir(serviceEvidence, { recursive: true }); await ordinary(serviceEvidence, true)
    remaining(compositionDeadline)
    const dataParent = join(work, 'service-data'); await mkdir(dataParent)
    const exitPath = join(serviceEvidence, 'service-observer-exit.json'), summaryPath = join(serviceEvidence, 'service-evidence-summary.json')
    const serviceCase = join(smoke, 'readiness-eof')
    remaining(compositionDeadline)
    return async () => {
    result.phase = 'retained-service-observation'
    result.service_started = 'NOT_CAPTURED'
    // 这一个父截止同时覆盖服务、摘要和最终安全原件，不把前置准备计为服务耗时。
    finalDeadline = performance.now() + 660_000
    const wallDeadline = String(Date.now() + Math.floor(remaining(finalDeadline)))
    let serviceError
    try {
      const observed = await command(serviceNode, ['--max-old-space-size=1024', join(source, 'tests/swa-cli-headless/serve-lifecycle.mjs'),
        '--candidate', development, '--source-revision', productRevision, '--edition', 'combined', '--evidence', smoke,
        '--data-parent', dataParent, '--protected-roots', checkout + ';' + (env.ProgramFiles ?? env.PROGRAMFILES), '--port', '18793',
        '--mode', 'readiness-eof', '--controller-input-check', '--retain-query-evidence', '--parent-deadline-ms', wallDeadline], source, 660_000,
        { deadline: finalDeadline, allowFailure: true })
      result.service_process = observed.receipt
      const row = observed.receipt
      const exit = { schema_version: 1, source_revision: productRevision, phase: 'service-observer', started: observed.started,
        exit: row.exit, timed_out: row.timed_out, stdout_bytes: row.stdout_bytes, stderr_bytes: row.stderr_bytes,
        stdout_eof: row.stdout_eof, stderr_eof: row.stderr_eof, outer_error: row.stream_failure || row.spawn_error || !row.close_observed ? 'UNKNOWN' : null }
      await within(finalDeadline, () => writeFile(exitPath, JSON.stringify(exit) + '\n', { flag: 'wx' }))
    } catch (error) { serviceError = error }
    try {
      const summarized = await command(serviceNode, [join(source, 'tests/swa-cli-headless/windows-service-evidence-summary.mjs'),
        '--source-root', source, '--install-root', development, '--evidence', serviceCase, '--source-revision', productRevision,
        '--observer-exit-file', exitPath, '--output', summaryPath, '--parent-deadline-ms', wallDeadline], source, 30_000,
        { deadline: Math.min(finalDeadline, performance.now() + 30_000), capture: true, allowFailure: true })
      result.summary_process = summarized.receipt
      const bytes = await within(finalDeadline, () => stable(summaryPath, 65536, true))
      const summary = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(bytes))
      keys(summary, ['schema_version', 'source_revision', 'status', 'first_failure', 'service_evidence', 'native_failure'])
      assert.equal(summary.schema_version, 1); assert.equal(summary.source_revision, productRevision)
      assert.ok(['PASS', 'FAIL'].includes(summary.status))
      assert.ok(summarized.bytes)
      assert.deepEqual(JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(summarized.bytes)), summary)
      for (const value of [summary.first_failure, summary.service_evidence, summary.native_failure]) assert.ok(value && typeof value === 'object' && !Array.isArray(value))
      assert.ok(Buffer.byteLength(JSON.stringify(summary.native_failure)) <= 1800)
      serviceMessages = [summary.service_evidence, summary.first_failure, summary.native_failure].map(value => JSON.stringify(value))
      assert.ok(serviceMessages.every(value => Buffer.byteLength(value) <= 3500))
      result.service_summary = { status: summary.status, sha256: sha(bytes), native_failure: summary.native_failure }
      // 摘要已完整验证原安全文件；只复制固定两原件，原始日志始终不公开。
      for (const [name, maximum] of [['service-query-evidence.json', 8 * 1024 ** 2], ['service-process-lifecycle.json', 2 * 1024 ** 2]]) {
        const originalBytes = await within(finalDeadline, () => stable(join(serviceCase, name), maximum, true))
        assert.ok(Array.isArray(summary.service_evidence.files))
        const matching = summary.service_evidence.files.filter(file => file.file_name === name)
        assert.equal(matching.length, 1)
        keys(matching[0], ['file_name', 'status', 'sha256', 'byte_count'])
        assert.equal(matching[0].status, 'SAVED')
        assert.equal(matching[0].sha256, sha(originalBytes))
        assert.equal(matching[0].byte_count, originalBytes.length)
        const document = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(originalBytes))
        assert.equal(document.source_revision, productRevision)
        assert.equal(document.sources.observer, consumerPins.find(pin => pin.name === 'serve-lifecycle.mjs').sha256)
        assert.equal(document.sources.native, result.native_sha256)
        if (name === 'service-process-lifecycle.json' && document.events.some(event =>
          event.kind === 'spawn' && event.source.kind === 'serve_native_identity' && event.source.role === 'serve' && positive(event.pid))) {
          result.service_started = true
        }
        await within(finalDeadline, () => writeFile(join(evidence, name), originalBytes, { flag: 'wx' }))
        assert.ok(originalBytes.equals(await within(finalDeadline, () => stable(join(evidence, name), maximum, true))))
      }
      await within(finalDeadline, () => writeFile(join(evidence, 'service-summary.json'), bytes, { flag: 'wx' }))
      assert.ok(bytes.equals(await within(finalDeadline, () => stable(join(evidence, 'service-summary.json'), 65536, true))))
      const row = result.service_process, summaryRow = summarized.receipt
      if (summary.status !== 'PASS' || result.service_started !== true || !row || row.exit !== 0 || row.timed_out || !row.close_observed ||
          !row.stdout_eof || !row.stderr_eof || row.stream_failure || row.spawn_error || !row.output_complete ||
          summaryRow.exit !== 0 || summaryRow.timed_out || !summaryRow.close_observed || !summaryRow.stdout_eof || !summaryRow.stderr_eof ||
          summaryRow.stream_failure || summaryRow.spawn_error || !summaryRow.output_complete) serviceError ??= fail('COMMAND_FAILED')
    } catch (error) { serviceError ??= error }
    try {
      await within(finalDeadline, async () => {
        await guardProduct(finalDeadline); await guardConsumers(finalDeadline); remaining(finalDeadline)
        await sourceGuard(); remaining(finalDeadline)
      })
      result.retained_product.extraction.product_inputs_unchanged = true
    } catch (error) { serviceError ??= error }
    if (serviceError) throw serviceError
    result.status = 'NOT_REPRODUCED'; result.phase = 'complete'
    }
    })
    await runService()
  }
  try {
    assert.match(result.observation_revision, /^[0-9a-f]{40}$/)
    assert.equal((await git(['rev-parse', 'HEAD'])).toString().trim(), result.observation_revision)
    assert.equal((await git(['status', '--porcelain', '--untracked-files=no'])).toString().trim(), '')
    result.phase = 'source-archive'
    const archive = join(work, 'source.tar')
    await git(['archive', '--format=tar', '--output', archive, sourceRevision])
    await command('tar', ['-xf', archive, '-C', source], work, 120_000)
    const node = await ordinary(process.execPath, false)
    const pin = await import(pathToFileURL(join(source, 'third_party/orca-v1.4.230/src/shared/node-runtime-pin.ts')).href)
    assert.equal(pin.NODE_RUNTIME_PIN.version, '24.21.0'); assert.equal(process.version, 'v24.21.0')
    assert.equal(sha(await stable(node, 128 * 1024 * 1024)), pin.NODE_RUNTIME_ASSETS['win32-x64'].executableSha256)
    const sourceNames = ['pnpm-lock.yaml', 'pnpm-workspace.yaml', 'package.json', 'runtime/swa-orca/package.json', 'scripts/build-swa-native.mjs',
      'scripts/download-swa-node.ps1', 'scripts/swa-package-config.mjs', 'third_party/orca-v1.4.230/src/shared/node-runtime-pin.ts',
      'third_party/orca-v1.4.230/config/patches/@vscode__windows-process-tree@0.8.0.patch', 'third_party/orca-v1.4.230/pnpm-lock.yaml']
    const sourcePins = await Promise.all(sourceNames.map(async name => ({ name, sha256: sha(await stable(join(source, name), 4 * 1024 * 1024)) })))
    result.phase = 'frozen-source-tools'
    assert.equal((await command(node, [pnpm, '--version'], source, 30_000, { capture: true })).bytes.toString().trim(), '11.27.0')
    await command(node, [pnpm, '--filter', '{.}', '--filter', '@swa/orca-runtime', '--filter', '@orca/process-host', 'install', '--frozen-lockfile', '--ignore-scripts'], source, 480_000)
    const requireSource = createRequire(join(source, 'runtime/swa-orca/package.json'))
    const originalPackage = await realpath(dirname(requireSource.resolve('@vscode/windows-process-tree/package.json')))
    const requirePackage = createRequire(join(originalPackage, 'package.json'))
    const addon = await realpath(dirname(requirePackage.resolve('node-addon-api/package.json')))
    assert.equal(requirePackage('./package.json').version, '0.8.0'); assert.equal(requirePackage('node-addon-api/package.json').version, '7.1.0')
    assert.equal(requireSource('node-gyp/package.json').version, '12.3.0')
    assert.ok(contains(source, originalPackage) && contains(source, addon))
    const originalFiles = await tree(originalPackage), addonFiles = await tree(addon)
    for (const [name, hashes] of Object.entries(approvedChanges)) assert.equal(originalFiles.find(row => row.name === name)?.sha256, hashes[0])
    result.phase = 'diagnostic-copy'
    await mkdir(runtime); await mkdir(join(runtime, 'node_modules')); await mkdir(join(runtime, 'node_modules/@vscode'))
    await copyFile(join(source, 'runtime/swa-orca/package.json'), join(runtime, 'package.json'))
    const diagnosticPackage = join(runtime, 'node_modules/@vscode/windows-process-tree')
    await cp(originalPackage, diagnosticPackage, { recursive: true, force: false, errorOnExist: true,
      filter: path => !['build', 'node_modules', 'deps'].includes(relative(originalPackage, path).split(/[\\/]/)[0]) })
    const copiedAddon = join(runtime, 'node_modules/node-addon-api')
    await cp(addon, copiedAddon, { recursive: true, force: false, errorOnExist: true })
    assert.deepEqual(await tree(diagnosticPackage, true), originalFiles)
    assert.deepEqual(await tree(copiedAddon, true), addonFiles)
    const patch = join(checkout, 'tests/swa-cli-headless/windows-process-error-diagnostic.patch')
    result.diagnostic_patch_sha256 = sha(await stable(patch))
    assert.equal(result.diagnostic_patch_sha256, '658074a08f16d983f2aac160384d0159b8738d7000670e58eae0b1ef6a8db504')
    await command('git', ['-c', 'core.autocrlf=false', 'apply', '--ignore-space-change', '--check', patch], diagnosticPackage, 30_000)
    await command('git', ['-c', 'core.autocrlf=false', 'apply', '--ignore-space-change', patch], diagnosticPackage, 30_000)
    const diagnosticFiles = await tree(diagnosticPackage, true)
    assert.deepEqual(diagnosticFiles.map(row => row.name), originalFiles.map(row => row.name))
    const changed = diagnosticFiles.filter((row, index) => row.sha256 !== originalFiles[index].sha256).map(row => row.name).sort()
    assert.deepEqual(changed, ['lib/index.js', 'lib/index.ts', 'src/process_worker.cc', 'typings/windows-process-tree.d.ts'].sort())
    for (const [name, hashes] of Object.entries(approvedChanges)) assert.equal(diagnosticFiles.find(row => row.name === name)?.sha256, hashes[1])
    result.inputs = { source: sourcePins, original_package: originalFiles, diagnostic_package: diagnosticFiles, node_addon_api: addonFiles }
    const sourceGuard = async () => {
      assert.deepEqual(await tree(originalPackage), originalFiles)
      assert.deepEqual(await tree(diagnosticPackage, true), diagnosticFiles)
      assert.deepEqual(await tree(addon), addonFiles)
      assert.deepEqual(await tree(copiedAddon, true), addonFiles)
      for (const item of sourcePins) assert.equal(sha(await stable(join(source, item.name), 4 * 1024 * 1024)), item.sha256)
    }
    await sourceGuard()
    result.phase = 'single-native-build'
    const buildRoot = join(checkout, '.verification/swa-cli-headless-20261020/build')
    await mkdir(buildRoot, { recursive: true }); await ordinary(buildRoot, true)
    const nativeWork = join(buildRoot, 'candidate-' + randomBytes(3).toString('hex')); await mkdir(nativeWork)
    const powershell = join(env.ProgramFiles ?? env.PROGRAMFILES, 'PowerShell/7/pwsh.exe')
    const preparationPath = join(privateRoot, 'diagnostic-native-preparation.json')
    await command(powershell, ['-NoProfile', '-NonInteractive', '-File', join(checkout, 'scripts/build-swa-windows-process-diagnostic.ps1'),
      '-NativeErrorOnly', '-SourceRoot', source, '-Runtime', runtime, '-Node', node, '-Work', nativeWork, '-OutputFile', preparationPath], checkout, 1_230_000)
    const preparation = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(await stable(preparationPath, 64 * 1024, true)))
    keys(preparation, ['status', 'native_component', 'native_component_index', 'diagnostic_only', 'other_native_components_built', 'original_inputs_modified',
      'node_version', 'node_sha256', 'headers_sha256', 'import_library_sha256', 'python_version', 'compiler_year', 'actual_vs_version', 'actual_cl_version'])
    assert.equal(preparation.status, 'PASS'); assert.equal(preparation.native_component, '@vscode/windows-process-tree'); assert.equal(preparation.native_component_index, 1)
    assert.equal(preparation.diagnostic_only, true); assert.equal(preparation.other_native_components_built, false); assert.equal(preparation.original_inputs_modified, false)
    assert.equal(preparation.node_version, '24.21.0'); assert.equal(preparation.node_sha256, pin.NODE_RUNTIME_ASSETS['win32-x64'].executableSha256)
    assert.equal(preparation.headers_sha256, pin.NODE_RUNTIME_PIN.headers.sha256); assert.equal(preparation.import_library_sha256, pin.NODE_RUNTIME_PIN.windowsImportLibs['win32-x64'].sha256)
    for (const key of ['python_version', 'actual_vs_version', 'actual_cl_version']) assert.ok(typeof preparation[key] === 'string' && /^\d+(?:\.\d+){1,4}$/.test(preparation[key]))
    assert.equal(preparation.compiler_year, '2026')
    const binary = join(diagnosticPackage, 'build/Release/windows_process_tree.node')
    result.native_sha256 = sha(await stable(binary, 16 * 1024 * 1024, true))
    await sourceGuard(); await save('SOURCE', sourceRevision + '\n'); await save('preparation.json', preparation)
    if (serviceOnly) {
      await serviceWindow({ node, powershell, binary, diagnosticPackage, sourceGuard })
    } else {
    result.phase = 'bounded-native-query'
    // 同一个截止涵盖真实子进程、双流、后核与最终证据；准备阶段有独立前置 timeout。
    finalDeadline = performance.now() + 30_000
    const queried = await command(node, [fileURLToPath(import.meta.url), '--query-only', '--parent-budget-ms', String(remaining(finalDeadline))], checkout, 30_000,
      { capture: true, allowFailure: true, deadline: finalDeadline })
    result.query_process = queried.receipt
    await within(finalDeadline, async () => {
      if (queried.bytes) {
        const actual = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(queried.bytes))
        queryRecord(actual); remaining(finalDeadline); result.query = actual
      }
      await sourceGuard()
      remaining(finalDeadline)
      assert.equal(sha(await stable(binary, 16 * 1024 * 1024, true)), result.native_sha256)
      remaining(finalDeadline)
      const actual = queried.receipt
      if (actual.exit !== 0 || !actual.close_observed || actual.timed_out || !actual.stdout_eof || !actual.stderr_eof || actual.stream_failure || actual.spawn_error || !actual.output_complete || result.query === 'NOT_CAPTURED' || result.query.status !== 'NOT_REPRODUCED') throw fail('COMMAND_FAILED')
      result.status = 'NOT_REPRODUCED'; result.phase = 'complete'
    })
    }
  } catch (error) { primary = error; result.status = 'FAIL'; result.error_kind = safeCode(error)
    result.secondary_error_kinds = Array.isArray(error?.diagnosticSecondaryCodes) && error.diagnosticSecondaryCodes.includes('FILE_CLOSE_FAILED') ? ['FILE_CLOSE_FAILED'] : [] }
  const emit = async deadline => {
    const projected = { schema_version: 1, source_revision: sourceRevision, observation_revision: result.observation_revision,
      status: result.status, phase: result.phase, error_kind: result.error_kind, secondary_error_kinds: result.secondary_error_kinds, native_sha256: result.native_sha256 ?? 'NOT_CAPTURED',
      query: result.query, query_process: result.query_process, kernel_tree_exit_proven: false, model_calls: 0, service_started: false }
    const component = { schema_version: 1, source_revision: sourceRevision,
      diagnostic_patch_sha256: result.diagnostic_patch_sha256 ?? 'NOT_CAPTURED', native_components_built: result.native_sha256 ? 1 : 'UNKNOWN',
      other_native_components_built: false, result_sha256: sha(Buffer.from(JSON.stringify(result, null, 2) + '\n')) }
    const messages = serviceOnly ? [...serviceMessages, JSON.stringify({ ...component, product_revision: productRevision,
      observation_revision: result.observation_revision, scope: 'retained_b464_development_service_not_candidate',
      status: result.status, phase: result.phase, error_kind: result.error_kind, secondary_error_kinds: result.secondary_error_kinds,
      native_sha256: result.native_sha256 ?? 'NOT_CAPTURED', service_process: result.service_process ?? 'NOT_CAPTURED',
      summary_process: result.summary_process ?? 'NOT_CAPTURED', model_calls: 0, worker_calls: 0, service_started: result.service_started,
      kernel_tree_exit_proven: false })] : [JSON.stringify(projected), JSON.stringify(component)]
    assert.ok(messages.length <= (serviceOnly ? 4 : 2))
    assert.ok(messages.every(value => Buffer.byteLength(value) <= 3500))
    await save('result.json', result, deadline)
    for (const message of messages) {
      remaining(deadline)
      const encoded = message.replaceAll('%', '%25').replaceAll('\r', '%0D').replaceAll('\n', '%0A')
      warningAttempts++
      await new Promise((resolveWrite, reject) => process.stdout.write('::warning file=.github/workflows/swa-windows-native.yml,line=1::' + encoded + '\n', error => error ? reject(error) : resolveWrite()))
      remaining(deadline)
    }
  }
  const outputDeadline = finalDeadline ?? performance.now() + 5000
  try { await within(outputDeadline, () => emit(outputDeadline)) }
  catch (error) {
    primary ??= error; result.status = 'FAIL'
    if (warningAttempts === 0) {
      warningAttempts++
      process.stdout.write('::warning file=.github/workflows/swa-windows-native.yml,line=1::{"schema_version":1,"status":"FAIL","error_kind":"SAFE_RECORD_FAILED"}\n')
    }
    // 到界后不等待未完成的文件 I/O；这条失败摘要也不冒成功落盘或已取消写入。
    if (finalDeadline && performance.now() >= finalDeadline) process.exit(1)
  }
  process.exitCode = primary || result.status !== 'NOT_REPRODUCED' ? 1 : 0
}
main().catch(() => { process.stdout.write('::warning file=.github/workflows/swa-windows-native.yml,line=1::{"schema_version":1,"status":"FAIL","error_kind":"PREFLIGHT_FAILED"}\n'); process.exitCode = 1 })
