import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { createHash, randomBytes } from 'node:crypto'
import { createReadStream } from 'node:fs'
import { copyFile, cp, lstat, mkdir, readFile, readdir, realpath, writeFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { dirname, isAbsolute, join, relative, resolve } from 'node:path'
import { createInterface } from 'node:readline'
import { fileURLToPath, pathToFileURL } from 'node:url'
import { parseArgs } from 'node:util'

async function main() {
// 只组合固定原包中的 Node/native 与原失败源码；不产生正式候选。
const sourceRevision = '7ff3a7d7d7e029c2fc9e2ae1fb6b81da8e0d4f8f'
const priorRevision = '1d3006fe0df0cb04b7f0de1e4563264bd0bfd9da'
const { values } = parseArgs({ options: { ...Object.fromEntries(['artifact-root', 'work-root', 'evidence-root', 'pnpm-cli'].map(name => [name, { type: 'string' }])),
  'native-compile-only': { type: 'boolean', default: false }, 'native-abi-only': { type: 'boolean', default: false }, 'service-observation-only': { type: 'boolean', default: false } }, strict: true, allowPositionals: false })
const nativeCompileOnly = values['native-compile-only'], nativeAbiOnly = values['native-abi-only'], nativeOnly = nativeCompileOnly || nativeAbiOnly
const serviceObservationOnly = values['service-observation-only']
assert.ok(!(nativeCompileOnly && nativeAbiOnly), 'Native compile-only and ABI-only modes are mutually exclusive')
assert.ok(!(serviceObservationOnly && nativeOnly), 'Service observation and native-only modes are mutually exclusive')
assert.equal(process.platform, 'win32'); assert.equal(process.arch, 'x64')
for (const name of ['artifact-root', 'work-root', 'evidence-root', 'pnpm-cli']) assert.ok(values[name] && isAbsolute(values[name]))
const checkout = await realpath(fileURLToPath(new URL('../', import.meta.url)))
const artifact = await realpath(values['artifact-root']), pnpm = await realpath(values['pnpm-cli'])
const work = resolve(values['work-root']), evidence = resolve(values['evidence-root'])
const contains = (parent, path) => { const rel = relative(parent, path); return !rel || rel !== '..' && !rel.startsWith('..\\') && !rel.startsWith('../') && !isAbsolute(rel) }
assert.ok(!contains(checkout, work) && !contains(artifact, work), 'development output must be outside source and retained artifact')
assert.equal(evidence.toLowerCase(), join(checkout, '.verification/swa-ci/windows-controller-open').toLowerCase())
assert.ok(!contains(work, evidence) && !contains(evidence, work))
assert.equal((await realpath(dirname(work))).toLowerCase(), (await realpath(process.env.RUNNER_TEMP)).toLowerCase())
await mkdir(work)
for (const path of [join(checkout, '.verification'), join(checkout, '.verification/swa-ci')]) {
  try { await mkdir(path) } catch (error) { if (error.code !== 'EEXIST') throw error }
  assert.ok((await lstat(path)).isDirectory() && !(await lstat(path)).isSymbolicLink())
  assert.equal((await realpath(path)).toLowerCase(), path.toLowerCase())
}
await mkdir(evidence)
const privateOutput = join(work, 'private')
await mkdir(privateOutput)
const save = (name, value) => writeFile(join(evidence, name), JSON.stringify(value, null, 2) + '\n', { flag: 'wx' })
const hash = bytes => createHash('sha256').update(bytes).digest('hex')
const digest = async path => ({ path: await realpath(path), sha256: hash(await readFile(path)) })
const result = { schema: 'swa.windows-controller-open-development.v1', status: 'FAIL', phase: 'preflight',
  source_revision: sourceRevision, input_revision: priorRevision, observation_revision: process.env.GITHUB_SHA ?? 'UNKNOWN',
  runtime_source_revision: nativeOnly ? priorRevision : sourceRevision, native_input_revision: priorRevision, go_source_revision: nativeOnly ? 'NOT_RUN' : 'UNKNOWN',
  scope: nativeCompileOnly ? 'development_native_compile_only_not_candidate' : nativeAbiOnly ? 'development_native_abi_only_not_candidate' : serviceObservationOnly ? 'development_service_observation_not_candidate' : 'development_composition_not_candidate',
  ...(nativeOnly ? { native_compile_only: nativeCompileOnly, native_abi_only: nativeAbiOnly, application_entries: 'COPIED_NOT_EXECUTED', go_build: 'NOT_RUN', windows_sdk: 'NOT_RUN',
    native_self: 'NOT_RUN', root_child_control: 'NOT_RUN', controller: 'NOT_RUN', service_execution: 'NOT_RUN' } : {}),
  model_calls: 0, worker_calls: 0, task_create: 'NOT_RUN', commands: [], private_raw_uploaded: false }
const env = {}
for (const key of Object.keys(process.env)) if (/^(?:systemroot|windir|comspec|pathext|programfiles|programfiles\(x86\)|programw6432|systemdrive|number_of_processors|processor_architecture|os)$/i.test(key)) env[key] = process.env[key]
const pathKeys = Object.keys(process.env).filter(key => key.toLowerCase() === 'path')
assert.equal(pathKeys.length, 1); assert.equal(typeof process.env[pathKeys[0]], 'string')
env.PATH = process.env[pathKeys[0]]
for (const name of ['home', 'temp', 'cache']) await mkdir(join(work, name))
Object.assign(env, { HOME: join(work, 'home'), USERPROFILE: join(work, 'home'), APPDATA: join(work, 'home'), LOCALAPPDATA: join(work, 'home'),
  TMP: join(work, 'temp'), TEMP: join(work, 'temp'), TMPDIR: join(work, 'temp'), GIT_CONFIG_NOSYSTEM: '1', GIT_TERMINAL_PROMPT: '0',
  GOCACHE: join(work, 'cache', 'go-build'), GOMODCACHE: join(work, 'cache', 'go-mod') })
let native, sequence = 0, stopBuild = () => {}, failure
const originalNative = []
function publicText(value) {
  if (typeof value !== 'string') return 'UNKNOWN'
  let text = value
  for (const [path, role] of [[checkout, 'observation-source'], [work, 'development-work'], [artifact, 'retained-package'], [evidence, 'evidence']]) text = text.split(path).join(`[${role}]`)
  if (/(?:authorization|\bbearer\b|api[-_]?key|client[-_]?secret|password|cookie|credential|nonce|launchTokenHash|S-1-\d|https?:\/\/)/i.test(text)) return '[sensitive error text omitted]'
  return text.replace(/[\x00-\x1f\x7f]/g, ' ').slice(0, 500)
}
const kind = error => typeof error?.code === 'string' && /^[A-Za-z][A-Za-z0-9_.-]{0,63}$/.test(error.code) ? error.code : error instanceof Error ? error.name : 'UNKNOWN'
const diagnosticErrorNames = ['Error', 'AssertionError', 'TypeError', 'RangeError', 'ReferenceError', 'SyntaxError', 'SystemError', 'UNKNOWN']
const diagnosticErrorCodes = ['ERR_ASSERTION', 'ERR_INVALID_ARG_TYPE', 'ERR_INVALID_ARG_VALUE', 'ERR_SWA_CONTROLLER_INPUT', 'ENOENT', 'EACCES', 'EPERM', 'EEXIST', 'ENOSPC', 'EMFILE', 'EIO', 'ETIMEDOUT', 'EPIPE', 'ECONNRESET', 'ECONNREFUSED', 'ABORT_ERR', 'UNKNOWN']
const diagnosticSignals = ['SIGTERM', 'SIGKILL', 'SIGINT', 'SIGABRT', 'SIGHUP', 'SIGPIPE', 'UNKNOWN']
const diagnosticCommandPhases = ['preflight', 'original-package-extraction', 'frozen-source-tools', 'native-runtime-copy', 'partial-application-build', 'windows-private-data-sdk-control', 'partial-go-build', 'diagnostic-process-tree-component', 'native-process-snapshot-self', 'native-process-snapshot-control', 'short-identify-then-controller-open', 'complete']
const diagnosticFailurePhases = {
  observer: ['spawn-record', 'query-freeze', 'native-query', 'collect-primary', 'collect-additional', 'record-query', 'write-unknown', 'write-latest', 'readiness-envelope', 'start-readiness-work', 'stdin-eof', 'loop-wait', 'await-readiness-work', 'save-service-streams', 'require-service-settled', 'native-exit-query', 'cim-request', 'cim-spawn', 'cim-wait', 'cim-read', 'exit-verdict', 'save-final-record', 'service-assertions'],
  readiness: ['ready-service', 'job-members', 'ready-record', 'controller-input'],
  'controller-input': ['prepare-output', 'import-producer', 'transport-spawn', 'transport-identity', 'transport-wait', 'transport-report', 'transport-validate', 'transport-receipts', 'producer-call'],
  'producer-command': ['identity', 'wait-close'], producer: ['public-clone', 'human-project-permission', 'controller-open', 'task-create', 'controller-observe-close', 'complete']
}
const diagnosticError = error => ({ name: typeof error?.name === 'string' && diagnosticErrorNames.includes(error.name) ? error.name : 'UNKNOWN',
  code: error?.code === undefined || error?.code === null ? null : typeof error.code === 'string' && diagnosticErrorCodes.includes(error.code) ? error.code : 'UNKNOWN' })
function serviceFirstFailureProjection() {
  const compact = value => {
    if (!value || value === 'NOT_CAPTURED') return 'NOT_CAPTURED'
    const { frames, status } = value.locations
    const facts = value.component === 'producer-command' ? {
      command_name: value.facts.command_name, command_sequence: value.facts.command_sequence, identity: value.facts.identity,
      exit_known: value.facts.exit !== null, exit_code: value.facts.exit?.exit ?? null, timed_out: value.facts.timed_out,
      stdout_eof: value.facts.stdout_eof, stderr_eof: value.facts.stderr_eof, stream_failure: value.facts.stream_failure, cleanup_attempted: value.facts.cleanup_attempted
    } : value.facts
    const projected = { component: value.component, phase: value.phase, at: value.at, error: value.error, facts,
      locations: { status, count: ['NOT_RECORDED', 'BOUND_EXCEEDED'].includes(status) ? null : frames.length, shown_count: frames.length ? 1 : 0,
        truncated: frames.length > 1, first: frames[0] ?? null } }
    assert.ok(Buffer.byteLength(JSON.stringify(projected)) <= 1024, 'Safe failure projection exceeds its bound; full files retained')
    return projected
  }
  const nonzero = result.first_nonzero_command ?? 'NOT_CAPTURED', outer = result.safe_outer_error ?? 'NOT_CAPTURED'
  assert.ok(Buffer.byteLength(JSON.stringify(nonzero)) <= 512 && Buffer.byteLength(JSON.stringify(outer)) <= 128)
  const envelope = { observation_revision: result.observation_revision, status: result.status, failure_records_match: result.service_evidence?.failure_records_match ?? 'NOT_CAPTURED', first_failure: null, observer_abort: null, first_nonzero_command: null, outer_error: null }
  assert.ok(Buffer.byteLength(JSON.stringify(envelope)) <= 512)
  const projection = { ...envelope, first_failure: compact(result.service_evidence?.first_failure), observer_abort: compact(result.service_evidence?.observer_abort), first_nonzero_command: nonzero, outer_error: outer }
  const text = JSON.stringify(projection)
  assert.ok(Buffer.byteLength(text) <= 3500, 'First failure annotation exceeds its bound; full files retained')
  return text
}
function utf8Fragment(text, limit, fromEnd = false) {
  let result = '', bytes = 0
  for (const character of fromEnd ? [...text].reverse() : text) {
    const size = Buffer.byteLength(character)
    if (bytes + size > limit) break
    result = fromEnd ? character + result : result + character
    bytes += size
  }
  return result
}
async function command(executable, args, cwd, cap = 120_000, extra = {}, allowNonzero = false) {
  const name = 'command-' + String(++sequence).padStart(2, '0'), out = [], err = []
  const diagnosticPhaseAtStart = serviceObservationOnly ? diagnosticCommandPhases.includes(result.phase) ? result.phase : 'UNKNOWN' : null
  let outBytes = 0, errBytes = 0, outEOF = false, errEOF = false, streamFailure = false, identity = null, exit, timeout = false
  const child = spawn(executable, args, { cwd, env: { ...env, ...extra }, windowsHide: true, stdio: ['ignore', 'pipe', 'pipe'] })
  const closed = new Promise(done => {
    child.once('error', error => done({ exit: null, signal: null, spawn_error: kind(error) }))
    child.once('close', (code, signal) => done({ exit: code, signal, spawn_error: null }))
  })
  child.stdout.on('data', bytes => { outBytes += bytes.length; if (outBytes <= 16 * 1024 * 1024) out.push(bytes); else streamFailure = true })
  child.stderr.on('data', bytes => { errBytes += bytes.length; if (errBytes <= 1024 * 1024) err.push(bytes); else streamFailure = true })
  child.stdout.on('end', () => { outEOF = true }); child.stderr.on('end', () => { errEOF = true })
  for (const stream of [child.stdout, child.stderr]) stream.on('error', () => { streamFailure = true })
  if (native && child.pid) {
    const creationTimeMs = native.getProcessCreationTime(child.pid)
    if (Number.isSafeInteger(creationTimeMs) && creationTimeMs > 0) identity = { pid: child.pid, creationTimeMs }
  }
  let timer
  try {
    exit = await Promise.race([closed, new Promise(done => { timer = setTimeout(() => { timeout = true; done(null) }, cap) })])
  } finally { clearTimeout(timer) }
  // 超时只尝试终止本次仍为同一 incarnation 的进程；结果继续失败。
  let cleanupAttempted = false
  if (!exit && identity && native.getProcessCreationTime(identity.pid) === identity.creationTimeMs) {
    cleanupAttempted = child.kill()
    let cleanupTimer
    try { exit = await Promise.race([closed, new Promise(done => { cleanupTimer = setTimeout(() => done(null), 5000) })]) }
    finally { clearTimeout(cleanupTimer) }
  }
  if (!exit) { child.stdout.destroy(); child.stderr.destroy(); child.unref() }
  const row = { name, exit: exit ?? 'UNKNOWN', timed_out: timeout, stdout_bytes: outBytes, stderr_bytes: errBytes, stdout_eof: outEOF, stderr_eof: errEOF,
    stream_failure: streamFailure, identity_captured: !!identity, cleanup_attempted: cleanupAttempted, kernel_tree_exit: 'NOT_CLAIMED' }
  result.commands.push(row)
  if (serviceObservationOnly && (!exit || timeout || exit.exit !== 0)) result.first_nonzero_command ??= {
    name, phase: diagnosticPhaseAtStart, exit: exit ? { exit: Number.isSafeInteger(exit.exit) ? exit.exit : null,
      signal: exit.signal === null ? null : diagnosticSignals.includes(exit.signal) ? exit.signal : 'UNKNOWN',
      spawn_error: exit.spawn_error === null ? null : diagnosticErrorCodes.includes(exit.spawn_error) ? exit.spawn_error : 'UNKNOWN' } : 'UNKNOWN',
    timed_out: timeout, stdout_bytes: outBytes, stderr_bytes: errBytes, stdout_eof: outEOF, stderr_eof: errEOF, stream_failure: streamFailure, cleanup_attempted: cleanupAttempted }
  await writeFile(join(privateOutput, name + '.stdout'), Buffer.concat(out), { flag: 'wx' })
  await writeFile(join(privateOutput, name + '.stderr'), Buffer.concat(err), { flag: 'wx' })
  if (!exit || timeout || exit.exit !== 0 && !allowNonzero) {
    const lines = Buffer.concat([...out, ...err]).toString('utf8').split(/\r?\n/).filter(line => /(?:\berror\b|\bfatal\b|ERR_[A-Z_]+|Cannot find|No projects matched|not recognized|SyntaxError|ReferenceError|TypeError)/i.test(line) || /^native-diagnostic phase=(?:validate|headers-download|headers-extract|node-library-download|official-msvc-toolchain|single-diagnostic-process-tree-build); kind=[A-Za-z][A-Za-z0-9]{0,63}$/.test(line))
    result.command_failure = { name, exit: exit ?? 'UNKNOWN', timed_out: timeout, error_line_total: lines.length,
      error_lines_truncated: lines.length > 20, errors: lines.slice(-20).map(publicText) }
  }
  assert.ok(!timeout && exit && outEOF && errEOF && !streamFailure, 'Actual preparation command did not settle completely')
  assert.equal(exit.signal, null); assert.equal(exit.spawn_error, null)
  if (!allowNonzero) assert.equal(exit.exit, 0, 'Actual preparation command exited nonzero')
  return { bytes: Buffer.concat(out), exit: exit.exit }
}
async function observeComponentFailure(nativeWork) {
  const observed = { status: 'UNKNOWN', actual_bytes: 'UNKNOWN', error_line_total: 'UNKNOWN', errors: [], errors_truncated: false }
  result.native_component_log = observed
  try {
    const command = result.commands.at(-1)
    if (command?.timed_out !== false || !command.stdout_eof || !command.stderr_eof || command.stream_failure ||
        !Number.isInteger(command.exit?.exit) || command.exit?.spawn_error !== null) {
      observed.status = 'NOT_CAPTURED'; observed.reason = 'command_not_settled'
      if (result.command_failure) result.command_failure.component_log = { status: observed.status, reason: observed.reason }
      return
    }
    const path = join(nativeWork, 'single-diagnostic-process-tree-build.raw.log'), info = await lstat(path)
    observed.actual_bytes = info.size
    assert.ok(info.isFile() && !info.isSymbolicLink() && info.nlink === 1 && info.size <= 16 * 1024 * 1024)
    assert.equal((await realpath(path)).toLowerCase(), resolve(path).toLowerCase())
    const input = createReadStream(path), reader = createInterface({ input, crlfDelay: Infinity })
    const first = [], last = [], codes = new Set()
    observed.error_line_total = 0
    try {
      for await (const line of reader) {
        const compiler = /\b(?:fatal )?error ((?:C|LNK|MSB)\d{4})\b/i.exec(line)
        const gyp = /^gyp ERR! (?:configure error\b|build error\b|stack Error: (?!Command failed\b))/i.test(line)
        const nodeError = /^(Error|TypeError|RangeError|ReferenceError|SyntaxError|SystemError|AssertionError)(?: \[(?:ERR_[A-Z0-9_]{1,96}|E[A-Z0-9_]{1,31})\])?: (?!Command failed\b)[^\r\n]*$/u.exec(line)
        if (!compiler && !gyp && !nodeError) continue
        observed.error_line_total++
        if (compiler) codes.add(compiler[1].toUpperCase())
        const text = /\b(?:command|argv)\b/i.test(line) ? (compiler ? compiler[1].toUpperCase() : nodeError ? nodeError[1] : 'gyp error') + ': [command text omitted]' : publicText(line)
        if (first.length < 10) first.push(text)
        else { last.push(text); if (last.length > 10) last.shift() }
      }
    } finally { reader.close(); input.destroy() }
    observed.status = 'OBSERVED'
    observed.errors = [...new Set(first.concat(last))]
    observed.errors_truncated = observed.error_line_total > observed.errors.length
    observed.error_code_total = codes.size; observed.error_codes = [...codes].slice(0, 8); observed.error_codes_truncated = codes.size > 8
    if (result.command_failure) {
      const original = result.command_failure
      original.component_log = { status: observed.status, actual_bytes: observed.actual_bytes, error_line_total: observed.error_line_total,
        error_codes: observed.error_codes, error_code_total: observed.error_code_total, error_codes_truncated: observed.error_codes_truncated }
      original.error_line_total += observed.error_line_total
      original.errors = observed.errors.concat(original.errors).slice(0, 20)
      original.error_lines_truncated ||= observed.errors_truncated || original.error_line_total > original.errors.length
    }
  } catch (error) {
    observed.status = error?.code === 'ENOENT' ? 'NOT_CAPTURED' : 'UNKNOWN'
    observed.error_line_total = 'UNKNOWN'
    observed.kind = kind(error)
    if (result.command_failure) result.command_failure.component_log = { status: observed.status, actual_bytes: observed.actual_bytes, kind: observed.kind }
  }
}
const git = async args => (await command('git', ['-c', 'credential.helper=', ...args], checkout)).bytes
async function fixed(path, revision = sourceRevision) { return git(['show', `${revision}:${path}`]) }
async function compareFixed(path) {
  const before = await fixed(path, priorRevision), after = await fixed(path)
  assert.ok(before.equals(after), `Retained native input differs: ${path}`)
  return { path, sha256: hash(after) }
}
async function ordinaryTree(root, base = root) {
  assert.ok(contains(base, await realpath(root)), 'Retained input escapes its ordinary package closure')
  const entry = await lstat(root)
  assert.ok(entry.isFile() || entry.isDirectory(), 'Retained package contains a nonordinary input')
  if (entry.isDirectory()) for (const name of await readdir(root)) await ordinaryTree(join(root, name), base)
  else if (/\.(?:node|dll)$/i.test(root)) originalNative.push({ path: relative(base, root).replaceAll('\\', '/'), sha256: (await digest(root)).sha256,
    dev: entry.dev, ino: entry.ino, nlink: entry.nlink })
}
async function retainServiceEvidence(directory) {
  const keys = (value, names) => { assert.ok(value && typeof value === 'object' && !Array.isArray(value)); assert.deepEqual(Object.keys(value).sort(), [...names].sort()) }
  const uint = value => Number.isSafeInteger(value) && value >= 0 && value <= 0xffffffff
  const birth = value => Number.isSafeInteger(value) && value > 0
  const finite = value => typeof value === 'number' && Number.isFinite(value) && value >= 0
  const list = (value, maximum, check) => { assert.ok(Array.isArray(value) && value.length <= maximum); for (const row of value) check(row) }
  const bool = value => assert.equal(typeof value, 'boolean')
  const atom = value => assert.ok(value === null || typeof value === 'string' && /^[A-Za-z][A-Za-z0-9_.-]{0,63}$/.test(value))
  const identity = value => { keys(value, ['pid', 'creationTimeMs']); assert.ok(uint(value.pid) && value.pid > 0 && birth(value.creationTimeMs)) }
  const instant = value => { keys(value, ['observed_at_ms', 'monotonic_ms']); assert.ok(birth(value.observed_at_ms) && finite(value.monotonic_ms)) }
  const sourceRoles = { serve_native_identity: ['serve'], callback_owned: ['callback'], named_job_member: ['serve', 'runtime', 'server', 'daemon'],
    transport_native_identity: ['transport'], transport_receipt_identity: ['tty_root', 'tty_helper', 'pipe_helper'],
    producer_command_native_identity: ['producer_command'], controller_receipt_native_identity: ['controller_process', 'controller_helper'] }
  const sourceInfo = value => {
    keys(value, ['kind', 'role', 'command_sequence', 'command_name', 'query_sequence', 'scope_index', 'scope_anchor'])
    assert.ok(typeof value.kind === 'string' && Object.hasOwn(sourceRoles, value.kind) && sourceRoles[value.kind].includes(value.role))
    if (value.kind === 'producer_command_native_identity') {
      assert.ok(birth(value.command_sequence)); assert.ok(typeof value.command_name === 'string' && /^(?:git|cli)-[0-9]{2,8}$/.test(value.command_name))
      assert.equal(Number(value.command_name.split('-')[1]), value.command_sequence)
    } else { assert.equal(value.command_sequence, null); assert.equal(value.command_name, null) }
    if (value.kind === 'callback_owned') { assert.ok(birth(value.query_sequence) && uint(value.scope_index) && value.scope_index <= 2048); identity(value.scope_anchor) }
    else { assert.equal(value.query_sequence, null); assert.equal(value.scope_index, null); assert.equal(value.scope_anchor, null) }
  }
  const origin = value => { keys(value, ['identity', 'confirmed_at', 'event_sequence', 'source']); identity(value.identity); instant(value.confirmed_at); assert.ok(birth(value.event_sequence)); sourceInfo(value.source) }
  const nativeRow = value => {
    keys(value, ['pid', 'ppid', 'name', 'creationTimeDiagnosticVersion', 'creationTimeProvider', 'creationTimeNtStatus', 'creationTimeSnapshotComplete', ...(Object.hasOwn(value, 'creationTimeMs') ? ['creationTimeMs'] : [])])
    assert.ok(uint(value.pid) && value.pid > 0 && uint(value.ppid)); assert.ok(typeof value.name === 'string' && value.name.length <= 260 && /^[\p{L}\p{N}_. ()-]*$/u.test(value.name))
    if (Object.hasOwn(value, 'creationTimeMs')) assert.ok(birth(value.creationTimeMs))
    assert.equal(value.creationTimeDiagnosticVersion, 2); assert.equal(value.creationTimeProvider, 'NtQuerySystemInformation'); assert.ok(uint(value.creationTimeNtStatus)); bool(value.creationTimeSnapshotComplete)
  }
  const processEvent = value => {
    if (value === null) return
    keys(value, ['exit', 'signal', 'observed_at_ms', 'monotonic_ms']); assert.ok(value.exit === null || Number.isSafeInteger(value.exit)); atom(value.signal)
    assert.ok(birth(value.observed_at_ms) && finite(value.monotonic_ms))
  }
  const queryInstant = value => { keys(value, ['observed_at_ms', 'monotonic_ms', 'root_exit_event', 'root_close_event']); assert.ok(birth(value.observed_at_ms) && finite(value.monotonic_ms)); processEvent(value.root_exit_event); processEvent(value.root_close_event) }
  const unknownReasons = ['invalid_native_process_row', 'current_root_identity_unverified', 'root_in_descendant_branch', 'parent_incarnation_unverified', 'process_identity_not_unique', 'creation_time_unavailable', 'parent_incarnation_changed']
  const closure = value => {
    keys(value, ['rootVerified', 'ownedRows', 'excludedEdges', 'unknownRows']); bool(value.rootVerified); list(value.ownedRows, 16384, nativeRow)
    list(value.excludedEdges, 32768, row => { keys(row, ['reason', 'actual', ...(Object.hasOwn(row, 'anchor') ? ['anchor'] : ['parent'])]); assert.ok(['different-root-incarnation', 'predates-root', 'predates-parent'].includes(row.reason)); nativeRow(row.actual); identity(row.anchor ?? row.parent) })
    list(value.unknownRows, 32768, row => {
      keys(row, ['reason', 'actual', ...(Object.hasOwn(row, 'source_anchor') ? ['source_anchor'] : [])]); assert.ok(unknownReasons.includes(row.reason)); nativeRow(row.actual)
      if (row.source_anchor) { keys(row.source_anchor, ['kind', 'pid', 'creation_time_ms']); assert.ok(['serve', 'additional'].includes(row.source_anchor.kind)); assert.ok(uint(row.source_anchor.pid) && row.source_anchor.pid > 0 && birth(row.source_anchor.creation_time_ms)) }
    })
  }
  const sample = value => {
    if (value === null) return
    keys(value, ['sequence', 'query_window', 'root', 'priors', 'before', 'rows', 'scopes', 'ownership_closure_incomplete']); assert.ok(birth(value.sequence)); identity(value.root)
    keys(value.query_window, ['start', 'end']); queryInstant(value.query_window.start); queryInstant(value.query_window.end)
    assert.ok(value.query_window.start.monotonic_ms <= value.query_window.end.monotonic_ms)
    list(value.priors, 2048, identity); keys(value.before, ['prior_sources', 'additional_roots']); list(value.before.prior_sources, 2048, origin); list(value.before.additional_roots, 2048, identity)
    assert.deepEqual(value.before.prior_sources.map(row => row.identity), value.priors)
    for (const row of value.before.prior_sources) assert.ok(row.confirmed_at.monotonic_ms <= value.query_window.start.monotonic_ms)
    list(value.rows, 16384, nativeRow); bool(value.ownership_closure_incomplete)
    list(value.scopes, 2049, row => {
      keys(row, ['anchor', 'source', 'existed_before_query', 'confirmed_during_query', 'result']); identity(row.anchor); origin(row.source); assert.deepEqual(row.anchor, row.source.identity)
      bool(row.existed_before_query); bool(row.confirmed_during_query); closure(row.result)
      assert.equal(row.confirmed_during_query, row.source.confirmed_at.monotonic_ms >= value.query_window.start.monotonic_ms && row.source.confirmed_at.monotonic_ms <= value.query_window.end.monotonic_ms)
    })
    assert.ok(value.scopes.length > 0); assert.deepEqual(value.scopes[0].anchor, value.root)
    value.scopes.forEach((row, index) => assert.equal(row.existed_before_query, (index === 0 ? value.priors : value.before.additional_roots).some(prior => prior.pid === row.anchor.pid && prior.creationTimeMs === row.anchor.creationTimeMs)))
    if (value.scopes.some(row => row.result.unknownRows.length > 0)) assert.equal(value.ownership_closure_incomplete, true)
  }
  const sourceHashes = { observer: result.observer_sources.find(row => row.name === 'serve-lifecycle.mjs').observation_sha256,
    producer: result.observer_sources.find(row => row.name === 'controller-taskcreate-check.mjs').observation_sha256,
    closure: result.observer_sources.find(row => row.name === 'process-incarnation-closure.mjs').observation_sha256, native: result.diagnostic_process_tree.binary_sha256 }
  const locationHashes = { observer: sourceHashes.observer, producer: sourceHashes.producer, closure: sourceHashes.closure,
    transport: result.observer_sources.find(row => row.name === 'controller-input-check.mjs').observation_sha256 }
  const failureRecord = value => {
    if (value === null) return
    keys(value, ['component', 'phase', 'at', 'error', 'facts', 'locations']); assert.ok(Buffer.byteLength(JSON.stringify(value)) <= 4096)
    assert.ok(typeof value.component === 'string' && Object.hasOwn(diagnosticFailurePhases, value.component) && diagnosticFailurePhases[value.component].includes(value.phase)); instant(value.at)
    keys(value.error, ['name', 'code']); assert.ok(diagnosticErrorNames.includes(value.error.name)); assert.ok(value.error.code === null || diagnosticErrorCodes.includes(value.error.code))
    const facts = value.facts, nullableIdentity = actual => { if (actual !== null) identity(actual) }
    const index = actual => assert.ok(actual === null || uint(actual))
    if (value.component === 'observer') {
      keys(facts, ['query_sequence', 'root', 'scope_index', 'check_index', 'closed', 'acted', 'readiness_started', 'readiness_complete', 'closure_incomplete', 'root_exit_seen', 'root_close_seen'])
      assert.ok(uint(facts.query_sequence)); nullableIdentity(facts.root); index(facts.scope_index); index(facts.check_index)
      for (const key of ['closed', 'acted', 'readiness_started', 'readiness_complete', 'closure_incomplete', 'root_exit_seen', 'root_close_seen']) bool(facts[key])
    } else if (value.component === 'readiness') {
      keys(facts, ['readiness_started', 'readiness_complete', 'controller_input_recorded']); for (const actual of Object.values(facts)) bool(actual)
    } else if (value.component === 'controller-input') {
      keys(facts, ['transport_identity', 'transport_spawned', 'transport_closed', 'producer_recorded']); nullableIdentity(facts.transport_identity)
      for (const key of ['transport_spawned', 'transport_closed', 'producer_recorded']) bool(facts[key])
    } else if (value.component === 'producer-command') {
      keys(facts, ['command_name', 'command_sequence', 'identity', 'exit', 'timed_out', 'stdout_eof', 'stderr_eof', 'stream_failure', 'cleanup_attempted'])
      assert.ok(typeof facts.command_name === 'string' && /^(?:git|cli)-[0-9]{2,8}$/.test(facts.command_name)); assert.ok(birth(facts.command_sequence)); assert.equal(Number(facts.command_name.split('-')[1]), facts.command_sequence)
      nullableIdentity(facts.identity)
      if (facts.exit !== null) {
        keys(facts.exit, ['exit', 'signal', 'spawn_error']); assert.ok(facts.exit.exit === null || Number.isSafeInteger(facts.exit.exit))
        assert.ok(facts.exit.signal === null || diagnosticSignals.includes(facts.exit.signal)); assert.ok(facts.exit.spawn_error === null || diagnosticErrorCodes.includes(facts.exit.spawn_error))
      }
      for (const key of ['timed_out', 'stdout_eof', 'stderr_eof', 'stream_failure', 'cleanup_attempted']) bool(facts[key])
    } else {
      keys(facts, ['operation', 'request_id_assigned', 'grant_created', 'controller_recorded']); assert.ok(['main', 'grant-revoke', 'dispatcher-close'].includes(facts.operation))
      for (const key of ['request_id_assigned', 'grant_created', 'controller_recorded']) bool(facts[key])
    }
    keys(value.locations, ['status', 'frames']); assert.ok(['CAPTURED', 'NOT_RECORDED', 'BOUND_EXCEEDED', 'NO_ALLOWED_FRAME'].includes(value.locations.status))
    list(value.locations.frames, 8, frame => {
      keys(frame, ['file', 'source_sha256', 'line', 'column']); assert.ok(typeof frame.file === 'string' && Object.hasOwn(locationHashes, frame.file)); assert.equal(frame.source_sha256, locationHashes[frame.file])
      assert.ok([frame.line, frame.column].every(actual => birth(actual) && actual <= 1000000))
    })
    assert.equal(value.locations.status === 'CAPTURED', value.locations.frames.length > 0)
  }
  const documents = {}, files = []
  for (const [name, maximum] of [['service-query-evidence.json', 8 * 1024 * 1024], ['service-process-lifecycle.json', 2 * 1024 * 1024]]) {
    const path = join(directory, name)
    try {
      let before
      try { before = await lstat(path) } catch (error) { if (error.code !== 'ENOENT') throw error; files.push({ file_name: name, status: 'NOT_CAPTURED' }); continue }
      assert.ok(before.isFile() && !before.isSymbolicLink() && before.nlink === 1 && before.size <= maximum)
      assert.equal((await realpath(path)).toLowerCase(), resolve(path).toLowerCase())
      const bytes = await readFile(path), after = await lstat(path)
      for (const key of ['dev', 'ino', 'size', 'mtimeMs', 'ctimeMs', 'nlink']) assert.equal(after[key], before[key])
      assert.equal(bytes.length, before.size)
      const value = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(bytes)), query = name === 'service-query-evidence.json'
      keys(value, ['schema_version', 'source_revision', 'sources', 'status', 'evidence_incomplete', 'error_kind', 'first_failure', 'observer_abort', ...(query ? ['query_count', 'ownership_closure_incomplete', 'first_sample', 'first_unknown_sample', 'last_sample'] : ['events'])])
      assert.equal(value.schema_version, 2); assert.equal(value.source_revision, result.observation_revision); assert.deepEqual(value.sources, sourceHashes)
      assert.ok(['OBSERVING', 'COMPLETE', 'INCOMPLETE'].includes(value.status)); bool(value.evidence_incomplete)
      assert.ok(value.error_kind === null || ['EVIDENCE_WRITE_OR_LIMIT', 'LIFECYCLE_INCOMPLETE', 'SOURCE_INCOMPLETE', 'PRIOR_SOURCE_INCOMPLETE', 'SCOPE_SOURCE_INCOMPLETE', 'NATIVE_BATCH_INCOMPLETE', 'QUERY_EVIDENCE_INCOMPLETE', 'OWNED_SCOPE_MISSING', 'QUERY_FAILED', 'OBSERVER_ABORTED', 'FAILURE_RECORD_INCOMPLETE'].includes(value.error_kind))
      failureRecord(value.first_failure); failureRecord(value.observer_abort)
      if (value.observer_abort) { assert.equal(value.observer_abort.component, 'observer'); assert.ok(value.first_failure && value.first_failure.at.monotonic_ms <= value.observer_abort.at.monotonic_ms) }
      if (query) {
        assert.ok(uint(value.query_count)); bool(value.ownership_closure_incomplete)
        for (const item of [value.first_sample, value.first_unknown_sample, value.last_sample]) { sample(item); if (item) assert.ok(item.sequence <= value.query_count) }
        if (value.first_unknown_sample) assert.ok(value.first_unknown_sample.scopes.some(row => row.result.unknownRows.length > 0))
      } else list(value.events, 4096, row => {
        keys(row, ['sequence', 'at', 'kind', 'source', 'pid', 'identity', 'exit', 'signal', 'spawn_error', 'stdio_policy']); assert.ok(birth(row.sequence)); instant(row.at); sourceInfo(row.source)
        assert.ok(['spawn', 'spawn-error', 'identity-confirm', 'receipt-observed', 'exit', 'close', 'stdout-end', 'stderr-end', 'stdin-end-request', 'stdin-finish', 'cleanup-request'].includes(row.kind))
        assert.ok(row.pid === null || uint(row.pid) && row.pid > 0); if (row.identity !== null) { identity(row.identity); assert.equal(row.pid, row.identity.pid) }
        assert.ok(row.exit === null || Number.isSafeInteger(row.exit)); atom(row.signal); atom(row.spawn_error)
        assert.equal(row.stdio_policy, row.kind === 'spawn' ? row.source.kind === 'serve_native_identity' ? 'pipe_pipe_pipe' : 'ignore_pipe_pipe' : null)
      })
      const sha256 = hash(bytes)
      await writeFile(join(evidence, name), bytes, { flag: 'wx' }); assert.equal((await digest(join(evidence, name))).sha256, sha256)
      documents[name] = value; files.push({ file_name: name, status: 'SAVED', sha256, byte_count: bytes.length })
    } catch (error) { files.push({ file_name: name, status: 'FAILED', error_kind: kind(error) }) }
  }
  const query = documents['service-query-evidence.json'], lifecycle = documents['service-process-lifecycle.json']
  let complete = !!query && !!lifecycle && [query, lifecycle].every(value => value.status === 'COMPLETE' && !value.evidence_incomplete)
  let binding = 'NOT_CHECKED'
  let failuresMatch = false
  if (query && lifecycle) {
    try {
      assert.deepEqual(query.first_failure, lifecycle.first_failure); assert.deepEqual(query.observer_abort, lifecycle.observer_abort)
      failuresMatch = true
      lifecycle.events.forEach((row, index) => { assert.equal(row.sequence, index + 1); if (index) assert.ok(row.at.monotonic_ms >= lifecycle.events[index - 1].at.monotonic_ms) })
      for (const sample of [query.first_sample, query.first_unknown_sample, query.last_sample].filter(Boolean)) {
        for (const origin of [...sample.before.prior_sources, ...sample.scopes.map(row => row.source)]) {
          const event = lifecycle.events[origin.event_sequence - 1]
          assert.ok(event && ['identity-confirm', 'receipt-observed'].includes(event.kind)); assert.deepEqual(event.identity, origin.identity); assert.deepEqual(event.at, origin.confirmed_at); assert.deepEqual(event.source, origin.source)
        }
      }
      assert.ok(query.first_sample && query.last_sample && query.query_count > 0)
      assert.equal(query.last_sample.sequence, query.query_count)
      binding = 'VERIFIED'
    } catch { binding = 'FAILED'; complete = false }
  }
  const describe = sample => sample === undefined ? 'NOT_CAPTURED' : sample === null ? null : { sequence: sample.sequence, query_window: sample.query_window,
    actual_row_count: sample.rows.length, prior_count: sample.priors.length, additional_before_count: sample.before.additional_roots.length,
    used_scope_count: sample.scopes.length, added_during_query_count: sample.scopes.slice(1).filter(row => !row.existed_before_query).length,
    root_verified: sample.scopes[0].result.rootVerified, unknown_count: sample.scopes.reduce((count, row) => count + row.result.unknownRows.length, 0) }
  const first = query?.first_unknown_sample, unknown = first?.scopes.flatMap((scope, index) => scope.result.unknownRows.map(row => ({ ...row, source_anchor: row.source_anchor ?? { kind: index === 0 ? 'serve' : 'additional', pid: scope.anchor.pid, creation_time_ms: scope.anchor.creationTimeMs } })))[0]
  result.service_evidence = { complete, binding, sources: sourceHashes, files, query_count: query?.query_count ?? 'NOT_RECORDED',
    failure_records_match: failuresMatch, first_failure: failuresMatch ? query.first_failure ?? 'NOT_CAPTURED' : 'NOT_CAPTURED', observer_abort: failuresMatch ? query.observer_abort ?? 'NOT_CAPTURED' : 'NOT_CAPTURED',
    ownership_closure_incomplete: query?.ownership_closure_incomplete ?? 'NOT_RECORDED', evidence_incomplete: query?.evidence_incomplete ?? 'NOT_RECORDED', error_kind: query ? query.error_kind : 'NOT_CAPTURED',
    first_sample: describe(query?.first_sample), first_unknown_sample: describe(first), last_sample: describe(query?.last_sample),
    lifecycle_event_count: lifecycle?.events.length ?? 'NOT_RECORDED', lifecycle_incomplete: lifecycle?.evidence_incomplete ?? 'NOT_RECORDED',
    source_kind_counts: lifecycle ? Object.keys(sourceRoles).map(kind => ({ kind, count: lifecycle.events.filter(row => row.source.kind === kind).length })) : 'NOT_RECORDED',
    first_unknown: !query ? 'NOT_CAPTURED' : unknown ? { reason: unknown.reason, actual: { pid: unknown.actual.pid, ppid: unknown.actual.ppid, creation_time_ms: unknown.actual.creationTimeMs ?? 'NOT_RECORDED' }, source_anchor: unknown.source_anchor,
      prior_exact_match: first.priors.some(row => row.pid === unknown.actual.pid && row.creationTimeMs === unknown.actual.creationTimeMs),
      pid_match_count: first.rows.filter(row => row.pid === unknown.actual.pid).length, parent_match_count: first.rows.filter(row => row.pid === unknown.actual.ppid).length,
      anchor_match_count: first.rows.filter(row => row.pid === unknown.source_anchor.pid).length } : null }
  if (result.service_evidence.first_failure !== 'NOT_CAPTURED') result.first_failure ??= { basis: 'CAPTURED_EXCEPTION', record: result.service_evidence.first_failure }
}
try {
  const goRevision = (await git(['rev-parse', 'HEAD'])).toString().trim()
  assert.match(goRevision, /^[0-9a-f]{40}$/)
  assert.equal(goRevision, process.env.GITHUB_SHA)
  const goScope = ['cmd', 'internal', 'contracts', 'go.mod', 'go.sum']
  const changedGo = (await git(['diff', '--name-only', sourceRevision, goRevision, '--', ...goScope])).toString().trim().split(/\r?\n/).filter(Boolean).sort()
  const approvedGo = ['cmd/swa/serve.go', 'cmd/swa/process_linux.go', 'cmd/swa/process_other.go', 'cmd/swa/process_windows.go', 'cmd/swa/data_directory_windows_test.go'].sort()
  assert.deepEqual(changedGo, approvedGo, 'Development Go source differs beyond the accepted private-data repair')
  assert.equal((await git(['status', '--porcelain', '--untracked-files=no', '--', ...goScope])).toString().trim(), '')
  if (!nativeOnly) result.go_source_revision = goRevision
  result.go_source_inputs = await Promise.all(approvedGo.map(async path => {
    const bytes = await readFile(join(checkout, path))
    assert.ok(bytes.equals(await fixed(path, goRevision)), 'Development Go input differs from its observation commit')
    return { path, sha256: hash(bytes) }
  }))
  const metadata = JSON.parse(await readFile(join(artifact, 'artifact-metadata.json'), 'utf8'))
  assert.equal(metadata.id, 12461580039); assert.equal(metadata.name, 'swa-windows-cli-' + priorRevision)
  assert.equal(metadata.workflow_run.id, 39858402174); assert.equal(metadata.expired, false)
  assert.equal(metadata.digest, 'sha256:2781fd68ec1c936c420572704a3528107f8a236bc421639307d0fe99066c9b8d')
  const files = await readdir(artifact), zipNames = files.filter(name => name.endsWith('.zip'))
  assert.equal(zipNames.length, 1)
  const archive = join(artifact, zipNames[0]), checksum = (await readFile(archive + '.sha256', 'utf8')).trim()
  const actualArchiveHash = hash(await readFile(archive))
  assert.equal(checksum, actualArchiveHash + '  ' + zipNames[0])
  result.archive_sha256 = actualArchiveHash
  const extracted = join(work, 'original-package')
  await mkdir(extracted)
  result.phase = 'original-package-extraction'
  const powershell = join(env.ProgramFiles ?? env.PROGRAMFILES, 'PowerShell', '7', 'pwsh.exe')
  await command(powershell, ['-NoProfile', '-NonInteractive', '-Command', '[IO.Compression.ZipFile]::ExtractToDirectory($env:SWA_PROBE_ZIP,$env:SWA_PROBE_EXTRACT,$false)'], work, 120_000, { SWA_PROBE_ZIP: archive, SWA_PROBE_EXTRACT: extracted })
  const roots = await readdir(extracted)
  assert.equal(roots.length, 1)
  const original = join(extracted, roots[0])
  assert.equal((await readFile(join(original, 'VERSION'), 'utf8')).trim(), priorRevision)
  await ordinaryTree(original)
  result.native_inputs_before = originalNative
  const originalNode = join(original, 'node', 'node.exe')
  const source = join(work, 'source-7ff3a7d7')
  await mkdir(source)
  const sourceArchive = join(work, 'source-7ff3a7d7.tar')
  await git(['archive', '--format=tar', '--output', sourceArchive, sourceRevision])
  await command('tar', ['-xf', sourceArchive, '-C', source], work)
  const pin = await import(pathToFileURL(join(source, 'third_party/orca-v1.4.230/src/shared/node-runtime-pin.ts')).href)
  assert.equal(pin.NODE_RUNTIME_PIN.version, '24.21.0')
  assert.equal((await digest(originalNode)).sha256, pin.NODE_RUNTIME_ASSETS['win32-x64'].executableSha256)
  assert.equal((await command(originalNode, ['--version'], work)).bytes.toString().trim(), 'v24.21.0')
  const inputPaths = ['runtime/swa-orca/package.json', 'third_party/orca-v1.4.230/src/shared/node-runtime-pin.ts', 'scripts/download-swa-node.ps1',
    'third_party/orca-v1.4.230/config/patches/node-pty@1.1.0.patch', 'third_party/orca-v1.4.230/config/patches/@vscode__windows-process-tree@0.8.0.patch']
  const changedNative = (await git(['diff', '--name-only', priorRevision, sourceRevision, '--', 'third_party/orca-v1.4.230/native', 'third_party/orca-v1.4.230/config/patches', 'third_party/orca-v1.4.230/src/packages/process-host'])).toString().trim()
  assert.equal(changedNative, '')
  result.inputs = await Promise.all(inputPaths.map(path => compareFixed(path)))
  const requireOriginal = createRequire(join(original, 'runtime/package.json'))
  native = requireOriginal('@vscode/windows-process-tree')
  result.phase = 'frozen-source-tools'
  assert.equal((await command(originalNode, [pnpm, '--version'], source)).bytes.toString().trim(), '11.27.0')
  await command(originalNode, [pnpm, '--filter', '{.}', '--filter', '{./web}', '--filter', '@swa/orca-runtime', '--filter', '@orca/process-host', 'install', '--frozen-lockfile', '--ignore-scripts'], source, 480_000)
  const requireRuntime = createRequire(join(source, 'runtime/swa-orca/package.json'))
  const ts = requireRuntime('typescript'), { build, stop } = requireRuntime('esbuild')
  stopBuild = stop
  assert.equal(ts.version, '6.0.3')
  const yaml = requireRuntime('yaml'), priorLock = yaml.parse((await fixed('pnpm-lock.yaml', priorRevision)).toString()), targetLock = yaml.parse(await readFile(join(source, 'pnpm-lock.yaml'), 'utf8'))
  for (const [name, pkg] of Object.entries(targetLock.packages)) assert.deepEqual(priorLock.packages[name], pkg, `Retained dependency resolution changed: ${name}`)
  for (const [name, snapshot] of Object.entries(targetLock.snapshots)) {
    const before = structuredClone(priorLock.snapshots[name]), after = structuredClone(snapshot)
    if (name === 'undici@8.11.2') { delete before.optional; delete after.optional }
    assert.deepEqual(before, after, `Retained dependency snapshot changed: ${name}`)
  }
  for (const name of ['runtime/swa-orca', 'third_party/orca-v1.4.230/native/windows-security', 'third_party/orca-v1.4.230/native/windows-registry', 'third_party/orca-v1.4.230/src/packages/process-host']) assert.deepEqual(priorLock.importers[name], targetLock.importers[name])
  assert.deepEqual(priorLock.patchedDependencies, targetLock.patchedDependencies)
  result.phase = nativeOnly ? 'native-runtime-copy' : 'partial-application-build'
  const development = join(work, 'development-7ff3a7d7')
  await mkdir(development)
  const entries = ['host.cjs', 'client.cjs', 'controller-helper.cjs', 'daemon-entry.js']
  for (const name of ['node', 'runtime', 'third-party-licenses']) await cp(join(original, name), join(development, name), { recursive: true, force: false, errorOnExist: true,
    filter: path => nativeOnly || !entries.some(entry => path === join(original, 'runtime', entry) || path === join(original, 'runtime', entry + '.metafile.json')) })
  const actualNode = join(development, 'node/node.exe')
  result.source_inputs = []
  if (!nativeOnly) {
  await writeFile(join(development, 'VERSION'), goRevision + '\n', { flag: 'wx' })
  for (const [name, text] of [['PLATFORM', 'win32-x64'], ['EDITION', 'cli']]) await writeFile(join(development, name), text + '\n', { flag: 'wx' })
  const builderPath = join(source, 'scripts/build-swa-runtime.mjs'), builderText = await readFile(builderPath, 'utf8')
  const tree = ts.createSourceFile(builderPath, builderText, ts.ScriptTarget.Latest, true, ts.ScriptKind.JS)
  assert.equal(tree.parseDiagnostics.length, 0)
  const functions = tree.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === 'buildSWARuntime')
  assert.equal(functions.length, 1)
  const declarations = functions[0].body.statements.filter(ts.isVariableStatement).flatMap(node => [...node.declarationList.declarations])
  const common = declarations.filter(node => node.name.getText(tree) === 'common'), mappings = declarations.filter(node => node.name.getText(tree) === 'entries')
  assert.equal(common.length, 1); assert.equal(mappings.length, 1)
  assert.ok(ts.isObjectLiteralExpression(common[0].initializer) && ts.isObjectLiteralExpression(mappings[0].initializer))
  const entryMap = new Map(mappings[0].initializer.properties.map(node => { assert.ok(ts.isPropertyAssignment(node) && ts.isStringLiteral(node.name)); return [node.name.text, node.initializer.getText(tree)] }))
  for (const name of entries) assert.ok(entryMap.has(name))
  const adapter = join(privateOutput, 'original-runtime-options.mjs')
  await writeFile(adapter, 'import { createRequire } from "node:module"; import { readFile } from "node:fs/promises"; import { dirname,join,resolve } from "node:path";\n' +
    `const root=${JSON.stringify(source)},runtime=join(root,"runtime/swa-orca"),upstream=join(root,"third_party/orca-v1.4.230"),requireRuntime=createRequire(join(runtime,"package.json"));\n` +
    'export const common=' + common[0].initializer.getText(tree) + ';\nexport const entries={' + entries.map(name => JSON.stringify(name) + ':' + entryMap.get(name)).join(',') + '};\n', { flag: 'wx' })
  const options = await import(pathToFileURL(adapter).href)
  const sourceInputs = new Set([builderPath])
  for (const [name, entry] of Object.entries(options.entries)) {
    const built = await build({ ...options.common, entryPoints: [entry], outfile: join(development, 'runtime', name) })
    await writeFile(join(development, 'runtime', name + '.metafile.json'), JSON.stringify(built.metafile), { flag: 'wx' })
    for (const name of Object.keys(built.metafile.inputs)) sourceInputs.add(resolve(source, name))
  }
  result.reused_child_entries = []
  for (const name of ['parcel-watcher-process-entry.js', 'profile-state-writer-worker-entry.js', 'profile-state-backup-worker-entry.js', 'foreign-sqlite-reader-entry.js', 'port-scan-command-worker-entry.js', 'session-scanner-service-entry.js', 'worker-thread-child-process-entry.js']) {
    const meta = JSON.parse(await readFile(join(original, 'runtime', name + '.metafile.json'), 'utf8'))
    let count = 0
    for (const path of Object.keys(meta.inputs)) {
      if (path.split(/[\\/]/).includes('node_modules') || path.startsWith('<')) continue
      assert.ok(!isAbsolute(path) && !path.split(/[\\/]/).includes('..'), 'Reused child input is not a source-relative Git path')
      await compareFixed(path); count++
    }
    assert.ok(count > 0)
    result.reused_child_entries.push({ name, git_input_count: count, sha256: (await digest(join(original, 'runtime', name))).sha256 })
  }
  result.source_inputs = await Promise.all([...sourceInputs].sort().map(async path => ({ path: relative(source, path).replaceAll('\\', '/'), sha256: (await digest(path)).sha256 })))
  assert.equal((await command('go', ['version'], checkout)).bytes.toString().trim().split(' ')[2], 'go1.27.2')
  await command('go', ['mod', 'download'], checkout, 240_000)
  await command('go', ['mod', 'verify'], checkout, 120_000)
  await command(originalNode, [join(source, 'scripts/check-dependencies.mjs')], source, 120_000)
  result.phase = 'windows-private-data-sdk-control'
  const sdk = await command('go', ['test', '-json', '-mod=readonly', '-count=1', '-timeout=90s', '-run', '^TestWindowsPrivateDataDirectory$', './cmd/swa'], checkout, 180_000, {}, true)
  const sdkCases = sdk.bytes.toString('utf8').split(/\r?\n/).filter(Boolean).map(line => JSON.parse(line))
    .filter(row => typeof row.Test === 'string' && ['pass', 'skip', 'fail'].includes(row.Action))
    .map(row => { assert.match(row.Test, /^TestWindowsPrivateDataDirectory(?:\/[a-z_]+)?$/); return { test: row.Test, action: row.Action } })
  const sdkNames = ['fresh_nested_and_existing_private', 'existing_extra_administrators_ace_unchanged', 'existing_wrong_owner_unchanged', 'ordinary_file_unchanged', 'reparse_target_and_ancestor', 'protected_roots_before_creation']
  result.windows_private_data_sdk = { exit: sdk.exit, cases: sdkCases }
  assert.equal(sdk.exit, 0)
  assert.deepEqual(sdkCases.filter(row => row.test.includes('/')).map(row => row.test.split('/')[1]).sort(), [...sdkNames].sort())
  for (const row of sdkCases) assert.ok(row.action === 'pass' || row.action === 'skip' && ['existing_wrong_owner_unchanged', 'reparse_target_and_ancestor'].includes(row.test.split('/')[1]))
  assert.ok(sdkCases.some(row => row.test === 'TestWindowsPrivateDataDirectory' && row.action === 'pass'))
  result.phase = 'partial-go-build'
  for (const [name, entry] of [['swa.exe', './cmd/swa'], ['server.exe', './cmd/swa-server']]) {
    await command('go', ['build', '-mod=readonly', '-p=4', '-trimpath', ...(name === 'swa.exe' ? ['-ldflags=-X main.version=' + goRevision] : []), '-o', join(development, name), entry], checkout, 480_000)
  }
  assert.equal((await command(join(development, 'swa.exe'), ['--version'], development)).bytes.toString().trim(), goRevision)
  }
  result.phase = 'diagnostic-process-tree-component'
  const developmentRuntime = join(development, 'runtime'), requireDevelopment = createRequire(join(developmentRuntime, 'package.json'))
  const originalPackage = await realpath(dirname(requireOriginal.resolve('@vscode/windows-process-tree/package.json')))
  const sourcePackage = await realpath(dirname(requireRuntime.resolve('@vscode/windows-process-tree/package.json')))
  const diagnosticPackage = await realpath(dirname(requireDevelopment.resolve('@vscode/windows-process-tree/package.json')))
  assert.ok(contains(developmentRuntime, diagnosticPackage) && !contains(original, diagnosticPackage) && !contains(source, diagnosticPackage))
  const diagnosticBinary = join(diagnosticPackage, 'build/Release/windows_process_tree.node')
  const retainedBinary = join(originalPackage, 'build/Release/windows_process_tree.node')
  const [diagnosticBinaryInfo, retainedBinaryInfo] = await Promise.all([lstat(diagnosticBinary), lstat(retainedBinary)])
  assert.ok(diagnosticBinaryInfo.isFile() && !diagnosticBinaryInfo.isSymbolicLink() && diagnosticBinaryInfo.nlink === 1)
  assert.ok(diagnosticBinaryInfo.dev !== retainedBinaryInfo.dev || diagnosticBinaryInfo.ino !== retainedBinaryInfo.ino, 'Diagnostic DLL aliases the retained native input')
  const diagnosticSources = ['src/process.h', 'src/process.cc', 'src/process_worker.cc']
  const unchangedSources = []
  result.diagnostic_process_tree = { scope: 'diagnostic_development_copy_only', observation_revision: goRevision, baseline_revision: priorRevision, source_inputs: [] }
  for (const name of diagnosticSources) {
    const originalPath = join(originalPackage, name), sourcePath = join(sourcePackage, name), target = join(diagnosticPackage, name)
    const [originalInfo, sourceInfo, targetInfo] = await Promise.all([lstat(originalPath), lstat(sourcePath), lstat(target)])
    assert.ok([originalInfo, sourceInfo, targetInfo].every(info => info.isFile() && !info.isSymbolicLink()))
    assert.equal(targetInfo.nlink, 1, 'Diagnostic source must not share a hardlink')
    for (const [path, info] of [[originalPath, originalInfo], [sourcePath, sourceInfo]]) {
      assert.ok(targetInfo.dev !== info.dev || targetInfo.ino !== info.ino, 'Diagnostic source aliases a retained or shared source')
      const bytes = await readFile(path)
      assert.ok(bytes.equals(await readFile(target)), 'Diagnostic source differs from its fixed retained input')
      unchangedSources.push({ path, sha256: hash(bytes), dev: info.dev, ino: info.ino, nlink: info.nlink })
    }
    assert.equal((await realpath(target)).toLowerCase(), resolve(target).toLowerCase())
    result.diagnostic_process_tree.source_inputs.push({ name, original_sha256: (await digest(target)).sha256 })
  }
  result.diagnostic_process_tree.vendor_sources = []
  for (const [name, targetName, sha256] of [
    ['ntextapi.h', 'src/psutil-ntextapi.h', 'f706626568c9a152e98a4631182745ec47ecf4de6261f332703ce7d56fb5968e'],
    ['LICENSE', 'psutil-LICENSE', 'b89c063b3786e28e0c0a38f1931db61fed35e69dd2a2966fbecffee0f46c8d10']
  ]) {
    const path = join(checkout, 'tests/swa-cli-headless/vendor/psutil-7.2.2', name), info = await lstat(path)
    assert.ok(info.isFile() && !info.isSymbolicLink())
    assert.equal((await realpath(path)).toLowerCase(), resolve(path).toLowerCase())
    const bytes = await readFile(path)
    assert.equal(hash(bytes), sha256)
    assert.ok(bytes.equals(await fixed('tests/swa-cli-headless/vendor/psutil-7.2.2/' + name, goRevision)))
    const target = join(diagnosticPackage, targetName)
    await writeFile(target, bytes, { flag: 'wx' })
    const targetInfo = await lstat(target)
    assert.ok(targetInfo.isFile() && !targetInfo.isSymbolicLink() && targetInfo.nlink === 1)
    assert.equal((await realpath(target)).toLowerCase(), resolve(target).toLowerCase())
    result.diagnostic_process_tree.vendor_sources.push({ name: targetName, sha256,
      upstream_revision: '9eea97dd6f1d16ea33f5144c8925f1ce7a0688e1', license: 'BSD-3-Clause' })
  }
  const diagnosticPatch = join(checkout, 'tests/swa-cli-headless/windows-process-query-diagnostic.patch')
  result.diagnostic_process_tree.patch_sha256 = (await digest(diagnosticPatch)).sha256
  await command('git', ['-c', 'core.autocrlf=false', 'apply', '--check', '--unidiff-zero', diagnosticPatch], diagnosticPackage)
  await command('git', ['-c', 'core.autocrlf=false', 'apply', '--unidiff-zero', diagnosticPatch], diagnosticPackage)
  for (const item of result.diagnostic_process_tree.source_inputs) item.diagnostic_sha256 = (await digest(join(diagnosticPackage, item.name))).sha256
  const buildRoot = join(checkout, '.verification/swa-cli-headless-20261020/build')
  await mkdir(buildRoot, { recursive: true })
  assert.equal((await realpath(buildRoot)).toLowerCase(), resolve(buildRoot).toLowerCase())
  const nativeWork = join(buildRoot, 'candidate-' + randomBytes(3).toString('hex'))
  await mkdir(nativeWork)
  const componentResult = join(privateOutput, 'diagnostic-native-preparation.json')
  try {
    await command(powershell, ['-NoProfile', '-NonInteractive', '-File', join(checkout, 'scripts/build-swa-windows-process-diagnostic.ps1'),
      '-SourceRoot', source, '-Runtime', developmentRuntime, '-Node', actualNode, '-Work', nativeWork, '-OutputFile', componentResult], checkout, 1_300_000, { RUNNER_TEMP: dirname(work) })
  } catch (error) {
    await observeComponentFailure(nativeWork)
    await sourceInputGuards()
    await originalInputGuards()
    throw error
  }
  const nativePreparation = JSON.parse(await readFile(componentResult, 'utf8'))
  assert.equal(nativePreparation.status, 'PASS'); assert.equal(nativePreparation.native_component, '@vscode/windows-process-tree')
  assert.equal(nativePreparation.native_component_index, 1); assert.equal(nativePreparation.other_native_components_built, false)
  result.diagnostic_process_tree.preparation = nativePreparation
  result.diagnostic_process_tree.binary_path = relative(development, diagnosticBinary).replaceAll('\\', '/')
  result.diagnostic_process_tree.binary_sha256 = (await digest(diagnosticBinary)).sha256
  const retainedProcessTree = originalNative.filter(item => item.path === result.diagnostic_process_tree.binary_path)
  assert.equal(retainedProcessTree.length, 1)
  assert.notEqual(result.diagnostic_process_tree.binary_sha256, retainedProcessTree[0].sha256)
  result.diagnostic_process_tree.baseline_binary_sha256 = retainedProcessTree[0].sha256
  async function sourceInputGuards() {
    for (const item of unchangedSources) {
      const info = await lstat(item.path)
      assert.equal(info.dev, item.dev); assert.equal(info.ino, item.ino); assert.equal(info.nlink, item.nlink); assert.equal((await digest(item.path)).sha256, item.sha256)
    }
    for (const item of result.diagnostic_process_tree.source_inputs) {
      const info = await lstat(join(diagnosticPackage, item.name))
      assert.equal(info.nlink, 1); assert.equal((await digest(join(diagnosticPackage, item.name))).sha256, item.diagnostic_sha256)
    }
    for (const item of result.diagnostic_process_tree.vendor_sources) {
      const path = join(diagnosticPackage, item.name), info = await lstat(path)
      assert.ok(info.isFile() && !info.isSymbolicLink() && info.nlink === 1)
      assert.equal((await realpath(path)).toLowerCase(), resolve(path).toLowerCase())
      assert.equal((await digest(path)).sha256, item.sha256)
    }
  }
  async function originalInputGuards() {
    result.native_inputs_after = await Promise.all(originalNative.map(async item => {
      const path = join(original, item.path), info = await lstat(path)
      return { path: item.path, sha256: (await digest(path)).sha256, dev: info.dev, ino: info.ino, nlink: info.nlink }
    }))
    assert.deepEqual(result.native_inputs_after, originalNative)
    const diagnosticRelative = relative(development, diagnosticBinary).replaceAll('\\', '/')
    for (const item of originalNative) if (item.path !== diagnosticRelative) assert.equal((await digest(join(development, item.path))).sha256, item.sha256)
    for (const item of result.go_source_inputs) assert.equal((await digest(join(checkout, item.path))).sha256, item.sha256)
    assert.equal((await git(['status', '--porcelain', '--untracked-files=no', '--', ...goScope])).toString().trim(), '')
    assert.equal(hash(await readFile(archive)), actualArchiveHash)
    result.native_input_protection = 'PASS'
  }
  await sourceInputGuards()
  if (nativeCompileOnly) {
    await originalInputGuards()
    for (const item of originalNative) assert.equal((await digest(join(development, item.path))).sha256,
      item.path === result.diagnostic_process_tree.binary_path ? result.diagnostic_process_tree.binary_sha256 : item.sha256)
    result.phase = 'native-process-tree-compile-complete'; result.status = 'PASS'
    return
  }
  result.phase = 'native-process-snapshot-self'
  if (nativeAbiOnly) result.native_self = 'UNKNOWN'
  native = requireDevelopment('@vscode/windows-process-tree')
  const diagnosticRows = await new Promise((resolveRows, reject) => {
    const timer = setTimeout(() => reject(new Error('Diagnostic native callback did not settle')), 15_000)
    native.getAllProcesses(rows => { clearTimeout(timer); Array.isArray(rows) ? resolveRows(rows) : reject(rows instanceof Error ? rows : new Error('Diagnostic native callback unavailable')) }, 4)
  })
  const diagnosticSelf = diagnosticRows.filter(row => row.pid === process.pid)
  assert.equal(diagnosticSelf.length, 1); assert.equal(diagnosticSelf[0].creationTimeDiagnosticVersion, 2)
  assert.equal(diagnosticSelf[0].creationTimeProvider, 'NtQuerySystemInformation')
  assert.equal(diagnosticSelf[0].creationTimeSnapshotComplete, true)
  assert.ok(Number.isInteger(diagnosticSelf[0].creationTimeNtStatus) && diagnosticSelf[0].creationTimeNtStatus >= 0 && diagnosticSelf[0].creationTimeNtStatus <= 0x7fffffff)
  assert.equal(diagnosticSelf[0].ppid, process.ppid)
  assert.ok(Number.isSafeInteger(diagnosticSelf[0].creationTimeMs) && diagnosticSelf[0].creationTimeMs > 0)
  assert.equal(diagnosticSelf[0].creationTimeMs, native.getProcessCreationTime(process.pid))
  for (const row of diagnosticRows) {
    assert.equal(row.creationTimeDiagnosticVersion, 2); assert.equal(row.creationTimeProvider, 'NtQuerySystemInformation')
    assert.equal(row.creationTimeSnapshotComplete, true); assert.equal(row.creationTimeNtStatus, diagnosticSelf[0].creationTimeNtStatus)
    assert.equal(typeof row.name, 'string')
    for (const key of ['creationTimeQueryStage', 'creationTimeWin32Error', 'creationTimeCloseError']) assert.equal(Object.hasOwn(row, key), false)
  }
  result.diagnostic_process_tree.actual_callback_version = 2
  result.native_snapshot = { provider: diagnosticSelf[0].creationTimeProvider, version: 2, ntstatus: diagnosticSelf[0].creationTimeNtStatus,
    self_identity_matches: true, root_child_control: 'NOT_RUN' }
  if (nativeAbiOnly) result.native_self = 'PASS'
  if (!nativeAbiOnly) result.development_entries = await Promise.all(['swa.exe', 'server.exe', 'VERSION', 'node/node.exe', ...entries.map(name => 'runtime/' + name)].map(async name => ({ name, sha256: (await digest(join(development, name))).sha256 })))
  // 观察脚本明确来自 observation revision，工具与所有产品输入仍解析自固定7ff目录。
  result.observer_sources = []
  for (const name of nativeAbiOnly ? ['process-incarnation-closure.mjs', 'windows-process-incarnation-control.mjs'] : ['controller-input-check.mjs', 'controller-taskcreate-check.mjs', 'serve-lifecycle.mjs', 'process-incarnation-closure.mjs', 'windows-process-incarnation-control.mjs']) {
    const destination = join(source, 'tests/swa-cli-headless', name), observation = join(checkout, 'tests/swa-cli-headless', name)
    assert.ok((await readFile(observation)).equals(await fixed('tests/swa-cli-headless/' + name, goRevision)))
    const originalSha = (await digest(destination)).sha256, actualSha = (await digest(observation)).sha256
    await copyFile(observation, destination)
    assert.equal((await digest(destination)).sha256, actualSha)
    result.observer_sources.push({ name, original_7ff_sha256: originalSha, observation_sha256: actualSha })
  }
  const requireObserver = createRequire(join(source, 'runtime/swa-orca/package.json'))
  if (!nativeAbiOnly) result.observer_tool_resolution = await Promise.all(['esbuild', 'typescript'].map(async name => {
    const path = await realpath(requireObserver.resolve(name))
    assert.ok(contains(source, path), 'Observer build dependency escaped the frozen source installation')
    return { name, path: relative(source, path).replaceAll('\\', '/') }
  }))
  let snapshotReport
  if (!serviceObservationOnly) {
  result.phase = 'native-process-snapshot-control'
  const snapshotData = join(work, 'native-process-control-data'), snapshotOutput = join(privateOutput, 'native-process-control')
  await mkdir(snapshotData)
  const controlScript = join(source, 'tests/swa-cli-headless/windows-process-incarnation-control.mjs')
  assert.ok((await readFile(controlScript)).equals(await fixed('tests/swa-cli-headless/windows-process-incarnation-control.mjs', goRevision)))
  if (nativeAbiOnly) result.root_child_control = 'UNKNOWN'
  result.native_snapshot.root_child_control = 'UNKNOWN'
  let snapshotCommand
  try {
    snapshotCommand = await command(actualNode, [controlScript, '--source-root', source, '--runtime-root', developmentRuntime,
      '--data-root', snapshotData, '--output-dir', snapshotOutput, '--source-revision', goRevision], source, 90_000, {}, true)
  } finally {
    // 完整实际帧先保全；fixture失败或未写完结果时也不丢掉已经取得的帧。
    result.continuity_frames = []
    let retentionFailure
    for (const label of ['a', 'b', 'c']) {
      const fileName = 'continuity-snapshot-' + label + '.json', path = join(snapshotOutput, fileName)
      try {
        let before
        try { before = await lstat(path) } catch (error) {
          if (error.code !== 'ENOENT') throw error
          result.continuity_frames.push({ label, file_name: fileName, status: 'NOT_CAPTURED' })
          continue
        }
        assert.ok(before.isFile() && !before.isSymbolicLink() && before.nlink === 1 && before.size <= 8 * 1024 * 1024)
        assert.equal((await realpath(path)).toLowerCase(), resolve(path).toLowerCase())
        const bytes = await readFile(path), after = await lstat(path)
        for (const key of ['dev', 'ino', 'size', 'mtimeMs', 'ctimeMs', 'nlink']) assert.equal(after[key], before[key])
        assert.equal(bytes.length, before.size)
        const frame = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(bytes))
        assert.equal(frame.schema_version, 1); assert.equal(frame.label, label); assert.ok(Array.isArray(frame.rows))
        for (const row of frame.rows) if (row && typeof row === 'object') assert.equal(Object.hasOwn(row, 'commandLine'), false)
        const sha256 = hash(bytes), destination = join(evidence, fileName)
        await writeFile(destination, bytes, { flag: 'wx' })
        assert.equal((await digest(destination)).sha256, sha256)
        result.continuity_frames.push({ label, file_name: fileName, status: 'SAVED', sha256, byte_count: bytes.length, actual_row_count: frame.rows.length })
      } catch (error) {
        result.continuity_frames.push({ label, file_name: fileName, status: 'FAILED', error_kind: kind(error) })
        retentionFailure ??= error
      }
    }
    const waitFileName = 'continuity-exit-wait.json', waitPath = join(snapshotOutput, waitFileName)
    result.continuity_exit_wait = { file_name: waitFileName, status: 'NOT_CAPTURED' }
    try {
      let before
      try { before = await lstat(waitPath) } catch (error) { if (error.code !== 'ENOENT') throw error }
      if (before) {
        assert.ok(before.isFile() && !before.isSymbolicLink() && before.nlink === 1 && before.size <= 2 * 1024 * 1024)
        assert.equal((await realpath(waitPath)).toLowerCase(), resolve(waitPath).toLowerCase())
        const bytes = await readFile(waitPath), after = await lstat(waitPath)
        for (const key of ['dev', 'ino', 'size', 'mtimeMs', 'ctimeMs', 'nlink']) assert.equal(after[key], before[key])
        assert.equal(bytes.length, before.size)
        const wait = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(bytes))
        const keys = (value, names) => {
          assert.ok(value && typeof value === 'object' && !Array.isArray(value))
          assert.deepEqual(Object.keys(value).sort(), [...names].sort())
        }
        const uint = value => Number.isSafeInteger(value) && value >= 0 && value <= 0xffffffff
        const birth = value => Number.isSafeInteger(value) && value > 0
        const finite = value => typeof value === 'number' && Number.isFinite(value) && value >= 0
        const identity = (value, withParent = false) => {
          keys(value, ['pid', 'creationTimeMs', ...(withParent ? ['ppid'] : [])])
          assert.ok(uint(value.pid) && value.pid > 0 && birth(value.creationTimeMs))
          if (withParent) assert.ok(uint(value.ppid))
        }
        const instant = value => { keys(value, ['observed_at_ms', 'monotonic_ms']); assert.ok(birth(value.observed_at_ms) && finite(value.monotonic_ms)) }
        const nativeRow = value => {
          keys(value, ['pid', 'ppid', 'name', 'creationTimeDiagnosticVersion', 'creationTimeProvider', 'creationTimeNtStatus', 'creationTimeSnapshotComplete', ...(Object.hasOwn(value, 'creationTimeMs') ? ['creationTimeMs'] : [])])
          assert.ok(uint(value.pid) && value.pid > 0 && uint(value.ppid))
          assert.ok(typeof value.name === 'string' && value.name.length <= 260 && /^[\p{L}\p{N}_. ()-]*$/u.test(value.name))
          if (Object.hasOwn(value, 'creationTimeMs')) assert.ok(birth(value.creationTimeMs))
          assert.equal(value.creationTimeDiagnosticVersion, 2); assert.equal(value.creationTimeProvider, 'NtQuerySystemInformation')
          assert.ok(uint(value.creationTimeNtStatus)); assert.equal(typeof value.creationTimeSnapshotComplete, 'boolean')
        }
        keys(wait, ['schema_version', 'source_revision', 'helper_sha256', 'closure_sha256', 'controller', 'parent', 'child', 'prior_b',
          'child_ready_observed', 'child_deadline_ms', 'wait_deadline_ms', 'budget_ms', 'interval_ms', 'maximum_samples',
          'status', 'stop_reason', 'retention_incomplete', 'unretained_sample_count', 'first_zero_index', 'samples'])
        assert.equal(wait.schema_version, 1); assert.equal(wait.source_revision, goRevision)
        assert.equal(wait.helper_sha256, (await digest(controlScript)).sha256)
        assert.equal(wait.closure_sha256, (await digest(join(source, 'tests/swa-cli-headless/process-incarnation-closure.mjs'))).sha256)
        for (const value of [wait.controller, wait.parent, wait.child]) identity(value, true)
        assert.ok(Array.isArray(wait.prior_b)); for (const value of wait.prior_b) identity(value)
        instant(wait.child_ready_observed)
        assert.equal(wait.child_deadline_ms, wait.child_ready_observed.monotonic_ms + 30_000)
        assert.equal(wait.wait_deadline_ms, wait.child_deadline_ms + 5000)
        assert.equal(wait.budget_ms, 5000); assert.equal(wait.interval_ms, 250); assert.equal(wait.maximum_samples, 21)
        assert.ok(['WAITING', 'ABSENCE_OBSERVED', 'FAIL'].includes(wait.status))
        assert.ok(wait.stop_reason === null || ['FIRST_ZERO', 'DEADLINE', 'QUERY_TIMEOUT', 'QUERY_FAILED', 'SAMPLE_LIMIT', 'EVIDENCE_LIMIT', 'VALIDATION_FAILED'].includes(wait.stop_reason))
        assert.equal(typeof wait.retention_incomplete, 'boolean'); assert.ok(uint(wait.unretained_sample_count))
        assert.ok(Array.isArray(wait.samples) && wait.samples.length <= 21)
        const unknownReasons = ['invalid_native_process_row', 'current_root_identity_unverified', 'root_in_descendant_branch',
          'parent_incarnation_unverified', 'process_identity_not_unique', 'creation_time_unavailable', 'parent_incarnation_changed']
        for (const [index, sample] of wait.samples.entries()) {
          keys(sample, ['index', 'query_window', 'rows', 'remaining', 'closure']); assert.equal(sample.index, index)
          keys(sample.query_window, ['start', 'end']); instant(sample.query_window.start); instant(sample.query_window.end)
          assert.ok(sample.query_window.start.monotonic_ms <= sample.query_window.end.monotonic_ms)
          assert.ok(Array.isArray(sample.rows)); for (const row of sample.rows) nativeRow(row)
          assert.ok(Array.isArray(sample.remaining) && sample.remaining.length <= 2)
          for (const row of sample.remaining) { identity(row, true); assert.ok([wait.parent, wait.child].some(old => JSON.stringify(old) === JSON.stringify(row))) }
          const closure = sample.closure
          keys(closure, ['rootVerified', 'ownedRows', 'excludedEdges', 'unknownRows'])
          assert.equal(typeof closure.rootVerified, 'boolean')
          for (const name of ['ownedRows', 'excludedEdges', 'unknownRows']) assert.ok(Array.isArray(closure[name]))
          for (const row of closure.ownedRows) nativeRow(row)
          for (const edge of closure.excludedEdges) {
            const anchorKey = edge.reason === 'different-root-incarnation' ? 'anchor' : 'parent'
            keys(edge, ['reason', 'actual', anchorKey]); assert.ok(['different-root-incarnation', 'predates-root', 'predates-parent'].includes(edge.reason))
            nativeRow(edge.actual); identity(edge[anchorKey])
          }
          for (const entry of closure.unknownRows) {
            keys(entry, ['reason', 'actual', ...(Object.hasOwn(entry, 'source_anchor') ? ['source_anchor'] : [])])
            assert.ok(unknownReasons.includes(entry.reason)); nativeRow(entry.actual)
            if (Object.hasOwn(entry, 'source_anchor')) {
              keys(entry.source_anchor, ['kind', 'pid', 'creation_time_ms']); assert.equal(entry.source_anchor.kind, 'additional')
              assert.ok(uint(entry.source_anchor.pid) && entry.source_anchor.pid > 0 && birth(entry.source_anchor.creation_time_ms))
            }
          }
        }
        assert.ok(wait.first_zero_index === null || Number.isSafeInteger(wait.first_zero_index) && wait.first_zero_index >= 0 && wait.first_zero_index < wait.samples.length)
        const firstZero = wait.first_zero_index === null ? null : wait.samples[wait.first_zero_index]
        if (firstZero) assert.equal(firstZero.remaining.length, 0)
        const lastNonzero = wait.samples.findLast(sample => sample.remaining.length > 0)
        const projectSample = sample => sample ? { index: sample.index, query_window: sample.query_window, remaining_count: sample.remaining.length,
          remaining: sample.remaining, root_verified: sample.closure.rootVerified, unknown_count: sample.closure.unknownRows.length } : null
        const sha256 = hash(bytes), destination = join(evidence, waitFileName)
        await writeFile(destination, bytes, { flag: 'wx' }); assert.equal((await digest(destination)).sha256, sha256)
        result.continuity_exit_wait = { file_name: waitFileName, status: 'SAVED', sha256, byte_count: bytes.length,
          source_revision: wait.source_revision, helper_sha256: wait.helper_sha256, closure_sha256: wait.closure_sha256,
          wait_status: wait.status, stop_reason: wait.stop_reason, retention_incomplete: wait.retention_incomplete,
          unretained_sample_count: wait.unretained_sample_count, sample_count: wait.samples.length, budget_ms: wait.budget_ms,
          child_ready_observed: wait.child_ready_observed, child_deadline_ms: wait.child_deadline_ms, wait_deadline_ms: wait.wait_deadline_ms,
          first_query: wait.samples[0]?.query_window ?? null, last_query: wait.samples.at(-1)?.query_window ?? null,
          first_zero: projectSample(firstZero), last_nonzero: projectSample(lastNonzero) }
        if (wait.status === 'ABSENCE_OBSERVED') {
          assert.equal(wait.retention_incomplete, false); assert.equal(wait.unretained_sample_count, 0)
          assert.equal(wait.stop_reason, 'FIRST_ZERO'); assert.ok(firstZero)
          assert.equal(firstZero.index, wait.samples.length - 1); assert.equal(firstZero.closure.rootVerified, true)
          assert.equal(firstZero.closure.unknownRows.length, 0); assert.ok(firstZero.query_window.end.monotonic_ms <= wait.wait_deadline_ms)
        }
      }
    } catch (error) {
      if (result.continuity_exit_wait.status !== 'SAVED') result.continuity_exit_wait = { file_name: waitFileName, status: 'FAILED', error_kind: kind(error) }
      retentionFailure ??= error
    }
    if (retentionFailure) throw retentionFailure
  }
  const snapshotReportPath = join(snapshotOutput, 'result.json'), snapshotReportInfo = await lstat(snapshotReportPath)
  assert.ok(snapshotReportInfo.isFile() && !snapshotReportInfo.isSymbolicLink() && snapshotReportInfo.nlink === 1 && snapshotReportInfo.size <= 1024 * 1024)
  assert.equal((await realpath(snapshotReportPath)).toLowerCase(), resolve(snapshotReportPath).toLowerCase())
  const snapshotReportBytes = await readFile(snapshotReportPath)
  snapshotReport = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(snapshotReportBytes))
  assert.equal(snapshotReport.schema_version, 1); assert.equal(snapshotReport.source_revision, goRevision)
  const continuityReport = snapshotReport.cases?.find(row => row.name === 'prior-owned-continuity')
  for (const frame of result.continuity_frames.filter(row => row.status === 'SAVED')) {
    const declaration = continuityReport?.snapshots?.find(row => row.label === frame.label)
    assert.ok(declaration)
    for (const key of ['file_name', 'sha256', 'byte_count', 'actual_row_count']) assert.equal(frame[key], declaration[key])
  }
  {
  const number = value => Number.isSafeInteger(value) && value >= 0 ? value : 'UNKNOWN'
  const finite = value => typeof value === 'number' && Number.isFinite(value) && value >= 0 ? value : 'UNKNOWN'
  const boolean = value => typeof value === 'boolean' ? value : 'UNKNOWN'
  const sha256 = value => typeof value === 'string' && /^[0-9a-f]{64}$/.test(value) ? value : 'UNKNOWN'
  const identity = row => ({ pid: number(row?.pid), ppid: number(row?.ppid), creationTimeMs: number(row?.creationTimeMs) })
  const instant = row => ({ observed_at_ms: number(row?.observed_at_ms), monotonic_ms: finite(row?.monotonic_ms) })
  const queryWindow = row => ({ start: instant(row?.start), end: instant(row?.end) })
  const actualExit = row => row === undefined ? 'UNKNOWN' : { exit: row?.exit === null ? null : number(row?.exit),
    signal: row?.signal === null ? null : /^SIG[A-Z0-9]{1,16}$/.test(row?.signal ?? '') ? row.signal : 'UNKNOWN', spawn_failed: boolean(row?.spawn_failed) }
  const observations = (rows, convert) => Array.isArray(rows) ? { count: rows.length, truncated: rows.length > 2, rows: rows.slice(0, 2).map(convert) } : 'UNKNOWN'
  const identities = rows => Array.isArray(rows) ? { count: rows.length, truncated: rows.length > 8, rows: rows.slice(0, 8).map(identity) } : 'UNKNOWN'
  result.root_child_fixture = { exit: snapshotCommand.exit,
    status: ['CONTROL_COMPLETE', 'FAIL', 'NOT_RUN'].includes(snapshotReport.status) ? snapshotReport.status : 'UNKNOWN',
    failed_phase: snapshotReport.failed_phase === undefined ? null : ['native-preflight', 'live-root-child', 'normal-eof', 'prior-owned-continuity'].includes(snapshotReport.failed_phase) ? snapshotReport.failed_phase : 'UNKNOWN',
    helper_sha256: sha256(snapshotReport.helper_sha256),
    probe_exit: number(snapshotReport.probe_exit), control_complete: boolean(snapshotReport.control_complete),
    cleanup_attempted: Object.hasOwn(snapshotReport, 'cleanup_attempted') ? boolean(snapshotReport.cleanup_attempted) : 'NOT_RECORDED',
    cleanup_unknown: Object.hasOwn(snapshotReport, 'cleanup_unknown') ? boolean(snapshotReport.cleanup_unknown) : 'NOT_RECORDED',
    kernel_tree_exit_proven: boolean(snapshotReport.kernel_tree_exit_proven), case_count: Array.isArray(snapshotReport.cases) ? snapshotReport.cases.length : 'UNKNOWN',
    cases_truncated: Array.isArray(snapshotReport.cases) ? snapshotReport.cases.length > 3 : 'UNKNOWN',
    cases: Array.isArray(snapshotReport.cases) ? snapshotReport.cases.slice(0, 3).map(row => ({
      name: ['live-root-child', 'normal-eof', 'prior-owned-continuity'].includes(row?.name) ? row.name : 'UNKNOWN',
      ...(row?.name === 'live-root-child' ? { observations: observations(row.observations, sample => ({ observed_at_ms: number(sample?.observed_at_ms), actual_row_count: number(sample?.actual_row_count),
        root_verified: boolean(sample?.root_verified), owned_count: number(sample?.owned_count), owned: identities(sample?.owned),
        excluded_edge_count: number(sample?.excluded_edge_count), unknown_count: number(sample?.unknown_count) })) } : {}),
      ...(row?.name === 'normal-eof' ? { root_exit: actualExit(row.root_exit), child_wait: actualExit(row.child_wait?.actual_exit),
        root_stdout_eof: boolean(row.root_stdout_eof), root_stderr_eof: boolean(row.root_stderr_eof), root_stdin_failed: boolean(row.root_stdin_failed),
        native_exit_observations: observations(row.native_exit_observations, sample => ({ observed_at_ms: number(sample?.observed_at_ms), actual_row_count: number(sample?.actual_row_count),
          remaining_count: number(sample?.remaining_count), remaining: identities(sample?.remaining), root_rows_count: number(sample?.root_rows_count),
          root_verified: boolean(sample?.root_verified), owned_count: number(sample?.owned_count), unknown_count: number(sample?.unknown_count) })),
        cim_process_exit: actualExit(row.cim_process_exit), cim_all_exited: boolean(row.cim_report?.all_exited),
        cim_observations: observations(row.cim_report?.observations, sample => ({ query_succeeded: boolean(sample?.query_succeeded),
          complete_result_count: number(sample?.complete_result_count), invalid_pid_record_count: number(sample?.invalid_pid_record_count),
          required_identity_fields_complete: boolean(sample?.required_identity_fields_complete), known_identities_exited: boolean(sample?.known_identities_exited),
          unknown_pids_absent: boolean(sample?.unknown_pids_absent), monotonic_ms: typeof sample?.monotonic_ms === 'number' && Number.isFinite(sample.monotonic_ms) && sample.monotonic_ms >= 0 ? sample.monotonic_ms : 'UNKNOWN',
          actual_match_count: number(sample?.actual_match_count) })) } : {}),
      ...(row?.name === 'prior-owned-continuity' ? {
        status: ['RUNNING', 'PASS', 'FAIL'].includes(row.status) ? row.status : 'UNKNOWN',
        stage: ['parent-ready', 'snapshot-a', 'snapshot-b', 'parent-exit', 'snapshot-c', 'prior-comparison', 'child-timer-exit', 'complete'].includes(row.stage) ? row.stage : 'UNKNOWN',
        error_kind: row.error_kind === undefined ? null : typeof row.error_kind === 'string' && /^[A-Za-z][A-Za-z0-9_.]{0,63}$/.test(row.error_kind) ? row.error_kind : 'UNKNOWN',
        controller: identity(row.controller), parent: identity(row.parent), child: identity(row.child),
        prior_a: identities(row.prior_a), prior_b: identities(row.prior_b),
        parent_exit: actualExit(row.parent_exit), parent_close_observed: instant(row.parent_close_observed),
        parent_stdout_eof: boolean(row.parent_stdout_eof), parent_stderr_eof: boolean(row.parent_stderr_eof),
        child_exit_code: row.child_exit_code === 'NOT_OBSERVABLE_FROM_NONPARENT' ? row.child_exit_code : 'UNKNOWN',
        forced_cleanup: boolean(row.forced_cleanup), cleanup_unknown: row.cleanup_unknown === undefined ? 'NOT_RECORDED' : boolean(row.cleanup_unknown),
        snapshots: Array.isArray(row.snapshots) ? { count: row.snapshots.length, truncated: row.snapshots.length > 3,
          rows: row.snapshots.slice(0, 3).map(sample => ({ label: ['a', 'b', 'c'].includes(sample?.label) ? sample.label : 'UNKNOWN',
            sha256: sha256(sample?.sha256), byte_count: number(sample?.byte_count), actual_row_count: number(sample?.actual_row_count), query_window: queryWindow(sample?.query_window),
            native_query: { provider: sample?.native_query?.provider === 'NtQuerySystemInformation' ? sample.native_query.provider : 'UNKNOWN',
              version: number(sample?.native_query?.version), ntstatus: number(sample?.native_query?.ntstatus), snapshot_complete: boolean(sample?.native_query?.snapshot_complete) } })) } : 'UNKNOWN',
        comparison: { original_owned_child: boolean(row.comparison?.original_owned_child), prior_a_owned_child: boolean(row.comparison?.prior_a_owned_child),
          prior_b_owned_child: boolean(row.comparison?.prior_b_owned_child), original_root_verified: boolean(row.comparison?.original_root_verified),
          prior_a_root_verified: boolean(row.comparison?.prior_a_root_verified), prior_b_root_verified: boolean(row.comparison?.prior_b_root_verified),
          original_unknown_count: number(row.comparison?.original_unknown_count), prior_a_unknown_count: number(row.comparison?.prior_a_unknown_count), prior_b_unknown_count: number(row.comparison?.prior_b_unknown_count),
          expected_unknown: Array.isArray(row.comparison?.expected_unknown) ? { count: row.comparison.expected_unknown.length, truncated: row.comparison.expected_unknown.length > 8,
            rows: row.comparison.expected_unknown.slice(0, 8).map(entry => ({ reason: entry?.reason === 'parent_incarnation_unverified' ? entry.reason : 'UNKNOWN',
              actual: identity(entry?.actual), source_anchor: { kind: entry?.source_anchor?.kind === 'additional' ? 'additional' : 'UNKNOWN',
                pid: number(entry?.source_anchor?.pid), creation_time_ms: number(entry?.source_anchor?.creation_time_ms) } })) } : 'UNKNOWN' },
        native_exit_observations: observations(row.native_exit_observations, sample => ({ query_window: queryWindow(sample?.query_window), actual_row_count: number(sample?.actual_row_count),
          remaining_count: number(sample?.remaining_count), remaining: identities(sample?.remaining), root_verified: boolean(sample?.root_verified), unknown_count: number(sample?.unknown_count) }))
      } : {}) })) : 'UNKNOWN' }
  if (nativeAbiOnly) result.root_child_control = result.root_child_fixture.status
  result.native_snapshot.root_child_control = result.root_child_fixture.status
  }
  assert.equal(snapshotCommand.exit, 0)
  assert.equal(snapshotReport.status, 'CONTROL_COMPLETE'); assert.equal(snapshotReport.control_complete, true); assert.equal(snapshotReport.probe_exit, 0)
  assert.equal(snapshotReport.native_component.addon_sha256, result.diagnostic_process_tree.binary_sha256)
  assert.equal(snapshotReport.helper_sha256, (await digest(controlScript)).sha256)
  assert.equal(snapshotReport.closure_sha256, (await digest(join(source, 'tests/swa-cli-headless/process-incarnation-closure.mjs'))).sha256)
  assert.equal(snapshotReport.model_calls, 0); assert.equal(snapshotReport.service_started, false); assert.equal(snapshotReport.account_session_started, false)
  assert.deepEqual(snapshotReport.cases.map(row => row.name), ['live-root-child', 'normal-eof', 'prior-owned-continuity'])
  assert.equal(continuityReport.status, 'PASS'); assert.equal(continuityReport.forced_cleanup, false)
  assert.equal(result.continuity_exit_wait.status, 'SAVED'); assert.equal(result.continuity_exit_wait.wait_status, 'ABSENCE_OBSERVED')
  assert.equal(result.continuity_exit_wait.retention_incomplete, false); assert.equal(result.continuity_exit_wait.unretained_sample_count, 0)
  assert.equal(continuityReport.exit_wait_file.sha256, result.continuity_exit_wait.sha256)
  assert.equal(continuityReport.exit_wait_file.sample_count, result.continuity_exit_wait.sample_count)
  assert.deepEqual(result.continuity_frames.map(row => row.status), ['SAVED', 'SAVED', 'SAVED'])
  result.diagnostic_process_tree.root_child_control = { status: snapshotReport.status, report_sha256: hash(snapshotReportBytes),
    live_observation_count: snapshotReport.cases[0].observations.length,
    native_exit_remaining: snapshotReport.cases[1].native_exit_observations.map(row => row.remaining_count),
    cim_exit: snapshotReport.cases[1].cim_process_exit.exit, cim_all_exited: snapshotReport.cases[1].cim_report.all_exited }
  result.native_snapshot.root_child_control = snapshotReport.status
  } else {
    assert.equal((await digest(join(source, 'tests/swa-cli-headless/windows-process-incarnation-control.mjs'))).sha256, '842060a11640491694dbb663057b71dd94bc58e24905acb006fdc61319a74ac4')
    assert.equal((await digest(join(source, 'tests/swa-cli-headless/process-incarnation-closure.mjs'))).sha256, '144d16fe707b432e9cfc05aa800d0719430ea112fc23c8706a46f9664436475f')
    result.native_snapshot.root_child_control = 'PRIOR_PASS_NOT_RERUN'
    result.prior_ordinary_control = { status: 'PRIOR_PASS_NOT_RERUN', head: 'ff33f6b9bcdd4b37eb330b61db49bd14d5476fb9', run: 40021139196, job: 120645694135,
      fixture_sha256: '842060a11640491694dbb663057b71dd94bc58e24905acb006fdc61319a74ac4', closure_sha256: '144d16fe707b432e9cfc05aa800d0719430ea112fc23c8706a46f9664436475f' }
  }
  for (const observer of result.observer_sources) {
    assert.equal((await digest(join(source, 'tests/swa-cli-headless', observer.name))).sha256, observer.observation_sha256)
    assert.equal((await digest(join(checkout, 'tests/swa-cli-headless', observer.name))).sha256, observer.observation_sha256)
  }
  await sourceInputGuards()
  if (nativeAbiOnly) {
    result.root_child_control = snapshotReport.status
    await originalInputGuards()
    for (const item of originalNative) assert.equal((await digest(join(development, item.path))).sha256,
      item.path === result.diagnostic_process_tree.binary_path ? result.diagnostic_process_tree.binary_sha256 : item.sha256)
    result.phase = 'native-process-tree-abi-control-complete'; result.status = 'PASS'
    return
  }
  const dataParent = join(work, 'data')
  await mkdir(dataParent)
  result.phase = 'short-identify-then-controller-open'
  const smoke = join(privateOutput, 'smoke')
  let observed
  try {
    observed = await command(actualNode, [join(source, 'tests/swa-cli-headless/serve-lifecycle.mjs'), '--candidate', development, '--evidence', smoke,
      '--data-parent', dataParent, '--protected-roots', env.ProgramFiles ?? env.PROGRAMFILES, '--port', '18793', '--mode', 'readiness-eof', '--source-revision', goRevision, '--controller-open-only', ...(serviceObservationOnly ? ['--retain-query-evidence'] : [])], checkout, 660_000, {}, true)
  } finally {
    if (serviceObservationOnly) await retainServiceEvidence(join(smoke, 'readiness-eof'))
  }
  result.fixture_exit = observed.exit
  result.native_inputs_after = await Promise.all(originalNative.map(async item => {
    const path = join(original, item.path), info = await lstat(path)
    return { path: item.path, sha256: (await digest(path)).sha256, dev: info.dev, ino: info.ino, nlink: info.nlink }
  }))
  assert.deepEqual(result.native_inputs_after, originalNative)
  for (const item of originalNative) assert.equal((await digest(join(development, item.path))).sha256,
    item.path === result.diagnostic_process_tree.binary_path ? result.diagnostic_process_tree.binary_sha256 : item.sha256)
  await sourceInputGuards()
  for (const item of result.source_inputs) assert.equal((await digest(join(source, item.path))).sha256, item.sha256)
  for (const item of result.go_source_inputs) assert.equal((await digest(join(checkout, item.path))).sha256, item.sha256)
  assert.equal((await git(['status', '--porcelain', '--untracked-files=no', '--', ...goScope])).toString().trim(), '')
  assert.equal(hash(await readFile(archive)), actualArchiveHash)
  const actual = JSON.parse(await readFile(join(smoke, 'result.json'), 'utf8'))
  assert.equal(actual.source_revision, goRevision); assert.equal(actual.records.length, 1)
  const record = actual.records[0], control = record.controller_input_check, producer = control?.producer
  result.control = { status: control?.status ?? 'UNKNOWN', transport: control?.transport ?? 'NOT_OBSERVED',
    producer: producer ? { status: producer.status, phase: producer.phase, error: producer.error ?? null, native_error: producer.native_error ?? null,
      open_request_id: producer.open_request_id ?? null, original_open_observation: producer.original_open_observation ?? null,
      open_receipt: producer.open_receipt ?? null, close_receipt: producer.close_receipt ?? null, cleanup: producer.cleanup,
      directory_observations: producer.directory_observations ?? [], ownership_incomplete: producer.ownership_incomplete } : 'NOT_STARTED' }
  try {
    const transportPath = join(smoke, 'readiness-eof/controller-input/transport/report.json')
    assert.ok((await lstat(transportPath)).size <= 2 * 1024 * 1024)
    const transport = JSON.parse(await readFile(transportPath, 'utf8'))
    assert.equal(transport.schema, 'swa.controller-input-check.v1')
    result.control.short_error = transport.error ? { kind: kind(transport.error), message: publicText(typeof transport.error.message === 'string' ? transport.error.message.split(/[\r\n]/u, 1)[0] : undefined) } : null
    const integer = value => Number.isSafeInteger(value) ? value : 'UNKNOWN'
    const boolean = value => typeof value === 'boolean' ? value : 'UNKNOWN'
    const tag = value => value === null ? null : typeof value === 'string' && /^[A-Za-z][A-Za-z0-9_.-]{0,63}$/.test(value) ? value : 'UNKNOWN'
    const traceEvent = event => !event || typeof event !== 'object' ? 'UNKNOWN' : {
      lane: ['tty', 'pipe'].includes(event.lane) ? event.lane : 'UNKNOWN',
      stage: ['create', 'spawn_identity', 'native_inputs', 'write', 'reply', 'identity', 'eof', 'absence', 'cleanup', 'exit'].includes(event.stage) ? event.stage : 'UNKNOWN',
      boundary: ['start', 'complete', 'observed'].includes(event.boundary) ? event.boundary : 'UNKNOWN',
      packet_index: event.packet_index === null ? null : integer(event.packet_index),
      packet_bytes: event.packet_bytes === null ? null : integer(event.packet_bytes),
      observed_at_ms: Number.isSafeInteger(event.observed_at_ms) ? event.observed_at_ms : 'UNKNOWN',
      monotonic_ms: Number.isFinite(event.monotonic_ms) ? event.monotonic_ms : 'UNKNOWN',
      elapsed_ms: Number.isFinite(event.elapsed_ms) ? event.elapsed_ms : 'UNKNOWN',
      deadline_remaining_ms: Number.isFinite(event.deadline_remaining_ms) ? event.deadline_remaining_ms : 'UNKNOWN' }
    const identity = value => value && Number.isSafeInteger(value.pid) && value.pid > 0 && Number.isSafeInteger(value.creationTimeMs) && value.creationTimeMs > 0
      ? { pid: value.pid, creation_time_ms: value.creationTimeMs } : null
    result.short_transport_diagnostic = { status: ['PASS', 'FAIL'].includes(transport.status) ? transport.status : 'UNKNOWN',
      error: transport.error ? { name: tag(transport.error.name), code: tag(transport.error.code) } : null,
      cleanup_unknown: boolean(transport.cleanup_unknown), trace_total: integer(transport.trace_total),
      trace: Array.isArray(transport.trace) ? transport.trace.slice(0, 64).map(traceEvent) : [],
      trace_truncated: transport.trace_truncated === true || transport.trace?.length > 64,
      last_event: traceEvent(transport.last_event), lanes: {}, identities_history: [] }
    for (const lane of ['tty', 'pipe']) {
      const row = transport[lane]
      result.short_transport_diagnostic.lanes[lane] = row === null ? null : row && typeof row === 'object' ? {
        status: ['PASS', 'FAIL'].includes(row.status) ? row.status : 'UNKNOWN', natural_eof: boolean(row.natural_eof), forced_cleanup: boolean(row.forced_cleanup),
        reply_count: Array.isArray(row.observations) ? row.observations.length : 'UNKNOWN',
        exit: row.exit === null ? null : row.exit ? { code: integer(row.exit.code), signal: row.exit.signal === null ? null : tag(row.exit.signal),
          cause: row.exit.cause ? { kind: ['operator_close', 'signaled', 'exited', 'unknown'].includes(row.exit.cause.kind) ? row.exit.cause.kind : 'UNKNOWN',
            ...(row.exit.cause.kind === 'exited' ? { exit_code: integer(row.exit.cause.exitCode) } : {}),
            ...(row.exit.cause.kind === 'signaled' ? { signal: integer(row.exit.cause.signal) } : {}),
            ...(row.exit.cause.kind === 'unknown' ? { reason: ['stop_unverified', 'host_status_unavailable', 'cause_unreported'].includes(row.exit.cause.reason) ? row.exit.cause.reason : 'UNKNOWN' } : {}) } : row.exit.cause === null ? null : 'NOT_RECORDED' } : 'UNKNOWN',
        failure_phase: row.failure_phase ? traceEvent(row.failure_phase) : null,
        sessions_after_dispose_count: Array.isArray(row.sessions_after_dispose) ? row.sessions_after_dispose.length : 'NOT_APPLICABLE',
        raw_output_bytes: row.raw_output_bytes === undefined ? 'NOT_RECORDED' : integer(row.raw_output_bytes),
        framing: row.framing && typeof row.framing === 'object' ? { valid_reply_records: integer(row.framing.valid_reply_records), schema_rejected_records: integer(row.framing.schema_rejected_records) } : 'NOT_RECORDED',
        framer_rejections: Array.isArray(row.rejected_output) ? ['invalid-json', 'line-too-long', 'UNKNOWN'].map(kind => ({ kind,
          count: row.rejected_output.filter(actual => kind === 'UNKNOWN' ? !['invalid-json', 'line-too-long'].includes(actual) : actual === kind).length })) : 'NOT_RECORDED' } : 'UNKNOWN'
    }
    const readyKinds = new Set(['startup-command-delivery', 'startup-command-stage-failed', 'shell-ready-timeout', 'shell-ready-wrapper-replaced'])
    const readyRows = Array.isArray(transport.tty?.readiness) ? transport.tty.readiness.map(row => ({
      event: readyKinds.has(row?.event) ? row.event : 'UNKNOWN', observed_at_ms: integer(row?.observed_at_ms),
      monotonic_ms: Number.isFinite(row?.monotonic_ms) ? row.monotonic_ms : 'UNKNOWN',
      ...(row?.event === 'startup-command-delivery' ? { written: boolean(row.details?.written), has_command: boolean(row.details?.hasCommand),
        command_length: integer(row.details?.commandLength), via_shell_args: boolean(row.details?.viaShellArgs), queued_by_shell_ready_barrier: boolean(row.details?.queuedByShellReadyBarrier) } : {}),
      ...(row?.event === 'shell-ready-timeout' ? { timeout_ms: integer(row.details?.timeoutMs) } : {}),
      ...(row?.event === 'startup-command-stage-failed' ? { reason: tag(row.details?.reason) } : {}) })) : null
    result.short_transport_diagnostic.ready_events = readyRows ? (readyRows.length > 32 ? readyRows.slice(0, 16).concat(readyRows.slice(-16)) : readyRows) : 'NOT_RECORDED'
    result.short_transport_diagnostic.ready_event_total = readyRows ? readyRows.length : 'NOT_RECORDED'
    result.short_transport_diagnostic.ready_events_truncated = !!readyRows && readyRows.length > 32
    result.short_transport_diagnostic.raw = { status: 'NOT_CAPTURED' }
    if (transport.coverage === 'explicit_short_identify_only' && transport.long_input === 'NOT_RUN' && transport.rpc_mutations === 'NOT_RUN' && transport.model_execution === 'NOT_RUN') {
      try {
        const rawPath = join(dirname(transportPath), 'tty-raw-transcript.txt'), rawInfo = await lstat(rawPath), actualPath = await realpath(rawPath)
        assert.ok(rawInfo.isFile() && !rawInfo.isSymbolicLink() && rawInfo.nlink === 1 && rawInfo.size <= 1024 * 1024)
        assert.equal(actualPath.toLowerCase(), resolve(rawPath).toLowerCase())
        const rawBytes = await readFile(rawPath)
        assert.equal(rawBytes.length, transport.tty.raw_output_bytes)
        let text = rawBytes.toString('utf8')
        for (const [path, role] of [[dirname(transportPath), 'short-control'], [checkout, 'observation-source'], [work, 'development-work'], [artifact, 'retained-package'], [evidence, 'evidence']]) text = text.split(path).join(`[${role}]`)
        const total = Buffer.byteLength(text), head = utf8Fragment(text, total <= 512 ? 512 : 256), tail = total <= 512 ? '' : utf8Fragment(text, 256, true)
        result.short_transport_diagnostic.raw = { status: 'OBSERVED', source: 'explicit_short_identify_only', original_bytes: rawBytes.length,
          projected_bytes: total, fragment_limit_bytes: 512, head, tail, fragment_bytes: Buffer.byteLength(head) + Buffer.byteLength(tail), truncated: total > 512 }
      } catch (error) { result.short_transport_diagnostic.raw = { status: error?.code === 'ENOENT' ? 'NOT_CAPTURED' : 'UNKNOWN', kind: kind(error) } }
    }
    for (const [role, value] of [['tty_root', transport.tty?.root ?? transport.tty?.spawn_incarnation], ['tty_helper', transport.tty?.helper],
      ['pipe_helper', transport.pipe?.helper], ['short_checker', control?.transport_process]]) {
      const actualIdentity = identity(value)
      if (actualIdentity) result.short_transport_diagnostic.identities_history.push({ role, ...actualIdentity })
    }
  } catch (error) { result.control.short_error = { status: 'UNKNOWN', kind: kind(error) } }
  result.service = { exit: record.exit, signal: record.signal, ownership_closure_incomplete: record.ownership_closure_incomplete,
    unknown_process_observation_count: record.unknown_process_observations.length,
    native_remaining: record.exit_observations.map(row => row.remaining.length), cim_exit: record.actual_cim_process_exit.exit,
    cim_all_exited: record.actual_cim_report?.all_exited ?? 'UNKNOWN' }
  const reasons = new Set(['invalid_native_process_row', 'current_root_identity_unverified', 'root_in_descendant_branch',
    'parent_incarnation_unverified', 'process_identity_not_unique', 'creation_time_unavailable', 'parent_incarnation_changed'])
  const number = value => Number.isSafeInteger(value) && value >= 0 ? value : 'UNKNOWN'
  const identifier = value => value === null ? null : typeof value === 'string' && /^[A-Za-z][A-Za-z0-9_.-]{0,63}$/.test(value) ? value : 'UNKNOWN'
  const boolean = value => typeof value === 'boolean' ? value : 'UNKNOWN'
  const processEvent = event => event === null ? null : event && typeof event === 'object' ? { exit: event.exit === null ? null : Number.isSafeInteger(event.exit) ? event.exit : 'UNKNOWN',
    signal: identifier(event.signal), observed_at_ms: number(event.observed_at_ms), monotonic_ms: Number.isFinite(event.monotonic_ms) ? event.monotonic_ms : 'UNKNOWN' } : 'UNKNOWN'
  const eventObserved = event => event === null ? false : event && typeof event === 'object' && Number.isSafeInteger(event.observed_at_ms) && Number.isFinite(event.monotonic_ms) ? true : 'UNKNOWN'
  const queryWindow = window => window && typeof window === 'object' ? {
    start_at_ms: number(window.start?.observed_at_ms), end_at_ms: number(window.end?.observed_at_ms),
    start_monotonic_ms: Number.isFinite(window.start?.monotonic_ms) ? window.start.monotonic_ms : 'UNKNOWN',
    end_monotonic_ms: Number.isFinite(window.end?.monotonic_ms) ? window.end.monotonic_ms : 'UNKNOWN',
    root_exit_observed_before_query: eventObserved(window.start?.root_exit_event), root_exit_observed_after_query: eventObserved(window.end?.root_exit_event),
    root_close_observed_before_query: eventObserved(window.start?.root_close_event), root_close_observed_after_query: eventObserved(window.end?.root_close_event) } : 'UNKNOWN'
  const counts = new Map(), groups = new Map(), queryCounts = new Map()
  assert.ok(Array.isArray(record.unknown_process_observations) && Array.isArray(record.observed_processes))
  for (const entry of record.unknown_process_observations) {
    const reason = reasons.has(entry?.reason) ? entry.reason : 'UNKNOWN'
    counts.set(reason, (counts.get(reason) ?? 0) + 1)
    const identity = { pid: number(entry?.actual?.pid), ppid: number(entry?.actual?.ppid), creation_time_ms: number(entry?.actual?.creationTimeMs) }
    const anchor = entry?.source_anchor && ['serve', 'additional'].includes(entry.source_anchor.kind)
      ? { kind: entry.source_anchor.kind, pid: number(entry.source_anchor.pid), creation_time_ms: number(entry.source_anchor.creation_time_ms) } : 'UNKNOWN'
    const diagnostic = entry?.actual?.creationTimeDiagnosticVersion === 2 ? {
      version: 2, api_stage: entry.actual.creationTimeProvider === 'NtQuerySystemInformation' ? entry.actual.creationTimeProvider : 'UNKNOWN',
      ntstatus: number(entry.actual.creationTimeNtStatus), snapshot_complete: boolean(entry.actual.creationTimeSnapshotComplete) } : entry?.actual?.creationTimeDiagnosticVersion === 1 ? {
      version: 1, api_stage: Number.isInteger(entry.actual.creationTimeQueryStage) && entry.actual.creationTimeQueryStage >= 0 && entry.actual.creationTimeQueryStage <= 4
        ? ['not_requested', 'OpenProcess', 'GetProcessTimes', 'invalid_filetime', 'complete'][entry.actual.creationTimeQueryStage] : 'UNKNOWN',
      win32_error: number(entry.actual.creationTimeWin32Error), close_error: number(entry.actual.creationTimeCloseError) } : 'UNKNOWN'
    const queryKey = JSON.stringify(diagnostic)
    const queryCount = queryCounts.get(queryKey)
    queryCounts.set(queryKey, { query: diagnostic, count: (queryCount?.count ?? 0) + 1 })
    const key = JSON.stringify([reason, identity.pid, identity.ppid, identity.creation_time_ms, anchor, diagnostic])
    const observedAt = number(entry?.observed_at_ms)
    if (groups.has(key)) {
      const group = groups.get(key)
      group.count++
      group.last_observed_at_ms = observedAt
      group.last_sample_sequence = number(entry?.sample_sequence)
      group.last_query_window = queryWindow(entry?.query_window)
      continue
    }
    const parents = new Map()
    if (typeof identity.ppid === 'number') for (const row of record.observed_processes) {
      if (row?.pid === identity.ppid && number(row.pid) !== 'UNKNOWN' && Number.isSafeInteger(row.creationTimeMs) && row.creationTimeMs > 0) {
        parents.set(row.pid + ':' + row.creationTimeMs, { pid: row.pid, creation_time_ms: row.creationTimeMs })
      }
    }
    const roleHistory = result.short_transport_diagnostic?.identities_history ?? []
    const roles = value => {
      const matches = value && typeof value.pid === 'number' && typeof value.creation_time_ms === 'number'
        ? [...new Set(roleHistory.filter(row => row.pid === value.pid && row.creation_time_ms === value.creation_time_ms).map(row => row.role))] : []
      return matches.length === 1 ? matches[0] : 'UNKNOWN'
    }
    groups.set(key, { reason, ...identity, source_anchor: anchor, diagnostic_query: diagnostic,
      actual_role_history: roles(identity), source_anchor_role_history: roles(anchor),
      count: 1, first_observed_at_ms: observedAt, last_observed_at_ms: observedAt,
      first_sample_sequence: number(entry?.sample_sequence), last_sample_sequence: number(entry?.sample_sequence),
      first_query_window: queryWindow(entry?.query_window), last_query_window: queryWindow(entry?.query_window),
      matches_recorded_identity_history: typeof identity.pid === 'number' && identity.pid > 0 && typeof identity.creation_time_ms === 'number' && identity.creation_time_ms > 0
        ? record.observed_processes.some(row => row?.pid === identity.pid && row.creationTimeMs === identity.creation_time_ms) : 'UNKNOWN',
      recorded_parent_identities_history: [...parents.values()].slice(0, 2), recorded_parent_identity_count: typeof identity.ppid === 'number' ? parents.size : 'UNKNOWN',
      recorded_parent_identities_truncated: parents.size > 2 })
  }
  const spawnEvents = record.events?.filter(event => event?.type === 'actual-spawn' && event.pid === record.pid) ?? []
  const rootCreation = spawnEvents.length === 1 ? number(spawnEvents[0].creation_time_ms) : 'UNKNOWN'
  const readinessError = record.readiness_check_error
  const unknownProjection = { go_source_revision: result.go_source_revision, runtime_source_revision: sourceRevision, native_input_revision: priorRevision,
    ownership_closure_incomplete: boolean(record.ownership_closure_incomplete),
    serve_root: { pid: number(record.pid), creation_time_ms: rootCreation },
    root_process_events: record.service_process_observations?.pid === record.pid && record.service_process_observations?.creation_time_ms === rootCreation
      ? { exit: processEvent(record.service_process_observations.exit_event), close: processEvent(record.service_process_observations.close_event) } : 'UNKNOWN',
    latch_inputs: { root_creation_time_readable: typeof rootCreation === 'number' && rootCreation > 0,
      controller_input_ownership_incomplete: boolean(control?.ownership_incomplete), producer_ownership_incomplete: boolean(producer?.ownership_incomplete),
      readiness_check_error: readinessError === null ? null : readinessError && typeof readinessError === 'object'
        ? { name: identifier(readinessError.name), code: identifier(readinessError.code) } : 'UNKNOWN' },
    unknown_item_count: record.unknown_process_observations.length,
    reason_counts: [...counts].map(([reason, count]) => ({ reason, count })),
    native_query_type_count: queryCounts.size, native_query_counts: [...queryCounts.values()].slice(0, 8), native_query_counts_truncated: queryCounts.size > 8,
    identity_group_count: groups.size, samples: groups.size > 8 ? [...groups.values()].slice(0, 4).concat([...groups.values()].slice(-4)) : [...groups.values()], samples_truncated: groups.size > 8,
    short_transport: result.short_transport_diagnostic ?? 'NOT_OBSERVED',
    identity_history_is_not_a_single_snapshot_chain: true }
  result.unknown_process_observation_projection = unknownProjection
  const logs = []
  let errorLineCount = 0
  for (const [role, path] of [['service', join(smoke, 'readiness-eof/stderr.log')], ['daemon', join(dataParent, 'readiness-eof/logs/daemon.log')], ['short-helper', join(smoke, 'readiness-eof/controller-input/transport.stderr')]]) {
    try {
      const info = await lstat(path)
      assert.ok(info.isFile() && info.size <= 16 * 1024 * 1024)
      const reader = createInterface({ input: createReadStream(path), crlfDelay: Infinity })
      for await (const line of reader) if (/(?:error|failed|unknown|exception|native_controller|native_codex|private_)/i.test(line)) {
        errorLineCount++
        logs.push({ role, text: publicText(line) })
        if (logs.length > 20) logs.shift()
      }
    } catch (error) { logs.push({ role, status: 'UNKNOWN', kind: kind(error) }) }
  }
  result.errors = logs
  result.error_line_total = errorLineCount
  result.error_lines_truncated = errorLineCount > 20
  if (serviceObservationOnly) {
    assert.equal(record.query_evidence?.evidence_incomplete, false)
    assert.equal(result.service_evidence?.complete, true)
    assert.equal(result.service_evidence.query_count, record.query_evidence.query_count)
    assert.equal(result.service_evidence.lifecycle_event_count, record.query_evidence.event_count)
    assert.equal(result.service_evidence.ownership_closure_incomplete, record.ownership_closure_incomplete)
  }
  assert.equal(observed.exit, 0); assert.equal(control?.status, 'PASS')
  result.phase = 'complete'; result.status = 'PASS'
} catch (error) {
  failure = error
  if (serviceObservationOnly) result.safe_outer_error = diagnosticError(error)
  result.error = { kind: kind(error), message: publicText(error?.message) }
  await writeFile(join(privateOutput, 'failure.txt'), String(error?.stack ?? error), { flag: 'wx' })
} finally {
  await stopBuild()
  if (serviceObservationOnly) result.first_failure ??= result.first_nonzero_command ? { basis: 'NONZERO_COMMAND', command: result.first_nonzero_command } : 'NOT_CAPTURED'
  await save('result.json', result)
  const firstFailureText = serviceObservationOnly ? serviceFirstFailureProjection() : null
  if (serviceObservationOnly && result.service_evidence) {
    const retained = result.service_evidence
    const annotations = [
      ['Windows service query evidence', { observation_revision: result.observation_revision, scope: result.scope, status: result.status,
        complete: retained.complete, binding: retained.binding, sources: retained.sources, files: retained.files,
        ordinary_control: result.prior_ordinary_control, evidence_incomplete: retained.evidence_incomplete, error_kind: retained.error_kind,
        ownership_closure_incomplete: retained.ownership_closure_incomplete, first_unknown: retained.first_unknown, kernel_tree_exit_proven: false }],
      ['Windows service query and lifecycle samples', { observation_revision: result.observation_revision, query_count: retained.query_count,
        first_sample: retained.first_sample, first_unknown_sample: retained.first_unknown_sample, last_sample: retained.last_sample,
        lifecycle_event_count: retained.lifecycle_event_count, lifecycle_incomplete: retained.lifecycle_incomplete, source_kind_counts: retained.source_kind_counts,
        task_create: 'NOT_RUN', model_calls: 0, worker_calls: 0 }]
    ].map(([title, value]) => ({ title, text: JSON.stringify(value) }))
    annotations.push({ title: 'Windows service first failure', text: firstFailureText })
    for (const row of annotations) assert.ok(Buffer.byteLength(row.text) <= 3500, 'Service evidence annotation exceeds its bound; full files retained')
    for (const row of annotations) process.stdout.write('::warning title=' + row.title + '::' + row.text.replaceAll('%', '%25').replaceAll('\r', '%0D').replaceAll('\n', '%0A') + '\n')
  } else if (serviceObservationOnly) {
    process.stdout.write('::warning title=Windows service first failure::' + firstFailureText.replaceAll('%', '%25').replaceAll('\r', '%0D').replaceAll('\n', '%0A') + '\n')
  }
  const rootChild = result.root_child_fixture
  const rootChildCases = Array.isArray(rootChild?.cases) ? rootChild.cases : []
  const liveRootChild = rootChildCases.find(row => row.name === 'live-root-child'), exitedRootChild = rootChildCases.find(row => row.name === 'normal-eof')
  const caseValues = (row, field, key) => Array.isArray(row?.[field]?.rows) ? row[field].rows.map(sample => sample[key]) : 'UNKNOWN'
  const projection = { status: result.status, phase: result.phase, scope: result.scope, source_revision: sourceRevision, input_revision: priorRevision,
    observation_revision: result.observation_revision, go_source_revision: result.go_source_revision, runtime_source_revision: result.runtime_source_revision, native_input_revision: priorRevision,
    ...(nativeOnly ? { native_compile_only: nativeCompileOnly, native_abi_only: nativeAbiOnly, application_entries: result.application_entries, go_build: 'NOT_RUN', windows_sdk: 'NOT_RUN',
      native_self: result.native_self, root_child_control: result.root_child_control, controller: 'NOT_RUN', service_execution: 'NOT_RUN' } : {}),
    native_input_scope: 'baseline_except_diagnostic_process_tree', diagnostic_process_tree_sha256: result.diagnostic_process_tree?.binary_sha256 ?? 'NOT_BUILT',
    native_snapshot: result.native_snapshot ?? 'NOT_RUN',
    ...(nativeAbiOnly && rootChild ? { root_child_fixture: { exit: rootChild.exit, status: rootChild.status,
      failed_phase: rootChild.failed_phase, probe_exit: rootChild.probe_exit,
      cleanup_attempted: rootChild.cleanup_attempted, cleanup_unknown: rootChild.cleanup_unknown,
      kernel_tree_exit_proven: rootChild.kernel_tree_exit_proven,
      live_root_verified: caseValues(liveRootChild, 'observations', 'root_verified'), live_unknown_counts: caseValues(liveRootChild, 'observations', 'unknown_count'),
      root_exit: exitedRootChild?.root_exit ?? 'UNKNOWN', child_wait: exitedRootChild?.child_wait ?? 'UNKNOWN',
      native_exit_remaining: caseValues(exitedRootChild, 'native_exit_observations', 'remaining_count'),
      native_exit_unknown_counts: caseValues(exitedRootChild, 'native_exit_observations', 'unknown_count'),
      native_exit_observed_at_ms: caseValues(exitedRootChild, 'native_exit_observations', 'observed_at_ms'),
      cim_exit: exitedRootChild?.cim_process_exit ?? 'UNKNOWN', cim_all_exited: exitedRootChild?.cim_all_exited ?? 'UNKNOWN',
      cim_monotonic_ms: caseValues(exitedRootChild, 'cim_observations', 'monotonic_ms') } } : {}),
    windows_private_data_sdk: result.windows_private_data_sdk ? { exit: result.windows_private_data_sdk.exit,
      pass_count: result.windows_private_data_sdk.cases.filter(row => row.action === 'pass').length,
      not_run: result.windows_private_data_sdk.cases.filter(row => row.action === 'skip').map(row => row.test.split('/')[1]),
      failures: result.windows_private_data_sdk.cases.filter(row => row.action === 'fail').map(row => row.test) } : 'NOT_RUN',
    fixture_exit: result.fixture_exit ?? 'NOT_RUN', error: result.error ?? null, control: structuredClone(result.control ?? null), service: result.service ?? null,
    command_failure: structuredClone(result.command_failure ?? null), errors: [...result.errors ?? []], error_line_total: result.error_line_total ?? 0, error_lines_truncated: result.error_lines_truncated ?? false, model_calls: 0, worker_calls: 0 }
  if (projection.control?.producer && typeof projection.control.producer === 'object') {
    const rows = projection.control.producer.directory_observations
    projection.control.producer.directory_observations = rows.filter((row, index) => row.role === 'data' || rows.findLastIndex(other => other.role === row.role) === index)
      .map(({ phase, role, status, owner_matches_current_user, protected: protectedDacl, ace_count, error_kind }) => ({ phase, role, status,
        ...(status === 'OBSERVED' ? { owner_matches_current_user, protected: protectedDacl, ace_count } : { error_kind: error_kind ?? null }) }))
  }
  while (Buffer.byteLength(JSON.stringify(projection)) > 3500 && projection.errors.length) { projection.errors.pop(); projection.error_lines_truncated = true }
  while (Buffer.byteLength(JSON.stringify(projection)) > 3500 && projection.command_failure?.errors.length > 1) { projection.command_failure.errors.pop(); projection.command_failure.error_lines_truncated = true }
  if (Buffer.byteLength(JSON.stringify(projection)) > 3500 && projection.command_failure?.errors.length === 1) {
    const errors = projection.command_failure.errors, original = errors[0]
    projection.command_failure.first_error_text_truncated = true
    errors[0] = ''
    assert.ok(Buffer.byteLength(JSON.stringify(projection)) <= 3500, 'Public diagnostic metadata leaves no first-error text budget')
    let low = 0, high = Buffer.byteLength(original)
    while (low < high) {
      const limit = Math.ceil((low + high) / 2)
      errors[0] = utf8Fragment(original, limit)
      if (Buffer.byteLength(JSON.stringify(projection)) <= 3500) low = limit
      else high = limit - 1
    }
    errors[0] = utf8Fragment(original, low)
    assert.ok(errors[0].length > 0 || original.length === 0, 'Public diagnostic metadata leaves no first-error text budget')
  }
  const text = JSON.stringify(projection)
  assert.ok(Buffer.byteLength(text) <= 3500, 'Public diagnostic metadata exceeds its total bound')
  const escaped = text.replaceAll('%', '%25').replaceAll('\r', '%0D').replaceAll('\n', '%0A')
  process.stdout.write('::warning title=Windows Controller open development control::' + escaped + '\n')
  const continuity = rootChildCases.find(row => row.name === 'prior-owned-continuity')
  if (continuity || result.continuity_frames) {
    const continuityAnnotations = [
      ['Windows process continuity', { observation_revision: result.observation_revision, fixture_status: rootChild?.status ?? 'UNKNOWN',
        status: continuity?.status ?? 'NOT_RUN', stage: continuity?.stage ?? 'NOT_RUN', error_kind: continuity?.error_kind ?? null,
        controller: continuity?.controller ?? 'UNKNOWN', parent: continuity?.parent ?? 'UNKNOWN', child: continuity?.child ?? 'UNKNOWN',
        parent_exit: continuity?.parent_exit ?? 'UNKNOWN', parent_close_observed: continuity?.parent_close_observed ?? 'UNKNOWN',
        parent_stdout_eof: continuity?.parent_stdout_eof ?? 'UNKNOWN', parent_stderr_eof: continuity?.parent_stderr_eof ?? 'UNKNOWN',
        child_exit_code: continuity?.child_exit_code ?? 'UNKNOWN', forced_cleanup: continuity?.forced_cleanup ?? 'UNKNOWN',
        cleanup_unknown: continuity?.cleanup_unknown ?? 'UNKNOWN', native_exit_observations: continuity?.native_exit_observations ?? 'UNKNOWN',
        kernel_tree_exit_proven: rootChild?.kernel_tree_exit_proven ?? 'UNKNOWN' }],
      ['Windows continuity prior identities', { prior_a: continuity?.prior_a ?? 'NOT_CAPTURED', prior_b: continuity?.prior_b ?? 'NOT_CAPTURED',
        comparison: continuity?.comparison ?? 'NOT_RUN', history_does_not_prove_d17_prior_ownership: true }],
      ['Windows continuity callback frames', { observation_revision: result.observation_revision,
        helper_sha256: rootChild?.helper_sha256 ?? 'UNKNOWN', closure_sha256: result.observer_sources?.find(row => row.name === 'process-incarnation-closure.mjs')?.observation_sha256 ?? 'UNKNOWN',
        native_sha256: result.diagnostic_process_tree?.binary_sha256 ?? 'NOT_BUILT', native_snapshot: result.native_snapshot ?? 'NOT_RUN',
        files: result.continuity_frames ?? 'NOT_CAPTURED', snapshots: continuity?.snapshots ?? 'NOT_CAPTURED' }]
    ]
    for (const [title, value] of continuityAnnotations) {
      const actual = JSON.stringify(value)
      assert.ok(Buffer.byteLength(actual) <= 3500, 'Continuity annotation exceeds its bound; full result retained')
      process.stdout.write('::warning title=' + title + '::' + actual.replaceAll('%', '%25').replaceAll('\r', '%0D').replaceAll('\n', '%0A') + '\n')
    }
  }
  if (result.continuity_exit_wait) {
    const text = JSON.stringify({ observation_revision: result.observation_revision, ...result.continuity_exit_wait,
      final_native_confirmations: continuity?.native_exit_observations ?? 'NOT_RUN', cim: rootChildCases.find(row => row.name === 'normal-eof')?.cim_process_exit ?? 'NOT_RUN' })
    assert.ok(Buffer.byteLength(text) <= 3500, 'Exit wait annotation exceeds its bound; complete file retained')
    process.stdout.write('::warning title=Windows continuity exit wait::' + text.replaceAll('%', '%25').replaceAll('\r', '%0D').replaceAll('\n', '%0A') + '\n')
  }
  if (result.unknown_process_observation_projection) {
    const diagnosticProjection = structuredClone(result.unknown_process_observation_projection)
    while (Buffer.byteLength(JSON.stringify(diagnosticProjection)) > 3500 && diagnosticProjection.short_transport?.trace?.length > 2) {
      diagnosticProjection.short_transport.trace.splice(1, 1); diagnosticProjection.short_transport.trace_truncated = true
    }
    while (Buffer.byteLength(JSON.stringify(diagnosticProjection)) > 3500 && diagnosticProjection.samples.length > 2) {
      diagnosticProjection.samples.splice(1, 1); diagnosticProjection.samples_truncated = true
    }
    while (Buffer.byteLength(JSON.stringify(diagnosticProjection)) > 3500 && Array.isArray(diagnosticProjection.short_transport?.ready_events) && diagnosticProjection.short_transport.ready_events.length > 1) {
      const rows = diagnosticProjection.short_transport.ready_events
      const index = rows.findIndex(row => row.event !== 'startup-command-delivery')
      rows.splice(index < 0 ? rows.length - 1 : index, 1)
      diagnosticProjection.short_transport.ready_events_truncated = true
    }
    while (Buffer.byteLength(JSON.stringify(diagnosticProjection)) > 3500 && diagnosticProjection.short_transport?.trace?.length) {
      diagnosticProjection.short_transport.trace.shift(); diagnosticProjection.short_transport.trace_truncated = true
    }
    for (const sample of diagnosticProjection.samples) if (JSON.stringify(sample.first_query_window) === JSON.stringify(sample.last_query_window) && sample.first_query_window !== 'UNKNOWN') sample.last_query_window = 'SAME_AS_FIRST'
    const annotationRaw = diagnosticProjection.short_transport?.raw
    for (const limit of [256, 128, 64]) {
      if (Buffer.byteLength(JSON.stringify(diagnosticProjection)) <= 3500 || annotationRaw?.status !== 'OBSERVED') break
      if (annotationRaw.fragment_bytes <= limit) continue
      const text = annotationRaw.head + annotationRaw.tail
      annotationRaw.head = utf8Fragment(text, limit / 2)
      annotationRaw.tail = utf8Fragment(text, limit / 2, true)
      annotationRaw.fragment_bytes = Buffer.byteLength(annotationRaw.head) + Buffer.byteLength(annotationRaw.tail)
      annotationRaw.truncated = true
    }
    if (Buffer.byteLength(JSON.stringify(diagnosticProjection)) > 3500 && diagnosticProjection.samples.length === 2) {
      diagnosticProjection.samples.shift(); diagnosticProjection.samples_truncated = true
    }
    while (Buffer.byteLength(JSON.stringify(diagnosticProjection)) > 3500 && diagnosticProjection.native_query_counts.length > 1) {
      diagnosticProjection.native_query_counts.splice(1, 1); diagnosticProjection.native_query_counts_truncated = true
    }
    for (const sample of diagnosticProjection.samples) while (Buffer.byteLength(JSON.stringify(diagnosticProjection)) > 3500 && sample.recorded_parent_identities_history.length) {
      sample.recorded_parent_identities_history.pop(); sample.recorded_parent_identities_truncated = true
    }
    const unknownText = JSON.stringify(diagnosticProjection)
    assert.ok(Buffer.byteLength(unknownText) <= 3500)
    process.stdout.write('::warning title=Windows closure unknown observations::' + unknownText.replaceAll('%', '%25').replaceAll('\r', '%0D').replaceAll('\n', '%0A') + '\n')
  }
}
if (failure) process.exitCode = 1
}
main().catch(error => {
  const code = typeof error?.code === 'string' && /^[A-Za-z][A-Za-z0-9_.-]{0,63}$/.test(error.code) ? error.code : 'UNEXPECTED_PREFLIGHT_ERROR'
  process.stderr.write('::error title=Windows Controller open preflight::phase=preflight; kind=' + code + '\n')
  process.exitCode = 1
})
