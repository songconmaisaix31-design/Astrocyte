import assert from 'node:assert/strict'
import { execFileSync, spawn } from 'node:child_process'
import { createHash, randomUUID } from 'node:crypto'
import { constants, closeSync, fsyncSync, lstatSync, openSync, realpathSync, statSync, writeFileSync } from 'node:fs'
import { lstat, mkdir, open, readFile, realpath, stat, writeFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { dirname, isAbsolute, join, relative, resolve } from 'node:path'
import { createInterface } from 'node:readline'
import { DatabaseSync } from 'node:sqlite'
import { Readable } from 'node:stream'
import { setTimeout as delay } from 'node:timers/promises'
import { fileURLToPath, pathToFileURL } from 'node:url'
import { parseArgs } from 'node:util'
import { publicCase, publicFiles, publicRevision, publicSource } from './public-case-definition.mjs'
import { codexEnds, codexStarts, isCodexDelta, originalLinuxTreeProof, validateCodexEventView, validateLinuxFinalCustody } from './codex-event-view.mjs'
import { isLinuxPiTarget, isPiDelta, originalPiTreeProof, piEnds, piStarts, readPiPrivateJson, validatePiEventView, validatePiLaunchRecord, validatePiSelection } from './pi-event-view.mjs'

const help = `真实安装物多Worker验收；不会签发Grant、启动服务或直接调用原生Agent。
使用同安装Node运行本文件，显式传入：
--candidate ABS --data-root ABS --project ABS --project-id ID --endpoint LOOPBACK
--agent ID --model ID --output ABS_NEW_DIRECTORY [--workers 3]
Windows Pi/OpenCode另需--service-log ABS；Linux Codex/Pi另需--source-repository ABS。
Linux StepFun Pi仅workers=3，另需--pi-launch-record ABS --pi-launch-sha256 SHA256，取自本次serve owner原件。
SWA_PROJECT_TOKEN仅通过受保护环境提供。来源、ToolsOff、当前Grant或完整资源证据缺失时不发送worker.start。
支持Windows Pi/OpenCode、Linux Codex0.163.0与封闭StepFun Pi1.0.1来源；缺项目化静态配置仍在启动前NOT_RUN。
业务成果必须由真实Reviewer/Agent控制端审阅；驱动不会自动生成通过记录。
退出：0=已执行断言与成果交接完成，1=失败，2=参数无效，3=未运行/前置不足；最终接受另由Reviewer决定。
`
if (process.argv.length === 3 && process.argv[2] === '--help') {
  process.stdout.write(help)
  process.exit(0)
}

class NotRun extends Error {}
const object = value => {
  assert.ok(value && typeof value === 'object' && !Array.isArray(value), 'product_object_missing')
  return value
}
const samePath = (left, right) => process.platform === 'win32' ? left.toLowerCase() === right.toLowerCase() : left === right
const within = (parent, child) => {
  const path = relative(parent, child)
  return !path || path !== '..' && !path.startsWith('../') && !path.startsWith('..\\') && !isAbsolute(path)
}
let values, workers
try {
  values = parseArgs({ options: Object.fromEntries(['candidate', 'data-root', 'project', 'project-id', 'endpoint', 'agent', 'model', 'output', 'service-log', 'source-repository', 'workers', 'pi-launch-record', 'pi-launch-sha256'].map(key => [key, { type: 'string' }])), allowPositionals: false }).values
  values.workers ??= '3'
  if (!/^[1-9]\d*$/.test(values.workers) || !Number.isSafeInteger(Number(values.workers)) || Number(values.workers) < 3 || Number(values.workers) > 100) throw new Error()
  workers = Number(values.workers)
  for (const key of ['candidate', 'data-root', 'project', 'output']) if (!values[key] || !isAbsolute(values[key])) throw new Error()
  if (process.platform === 'win32' && (!values['service-log'] || !isAbsolute(values['service-log']))) throw new Error()
  if (process.platform === 'linux' && ['codex', 'pi'].includes(values.agent) && (!values['source-repository'] || !isAbsolute(values['source-repository']))) throw new Error()
  if (process.platform === 'linux' && values.agent === 'pi') {
    if (!isLinuxPiTarget(process.platform, values.agent, values.model, workers) || !isAbsolute(values['pi-launch-record'] ?? '') || !/^[a-f0-9]{64}$/.test(values['pi-launch-sha256'] ?? '')) throw new Error()
  } else if (values['pi-launch-record'] !== undefined || values['pi-launch-sha256'] !== undefined) throw new Error()
  for (const key of ['project-id', 'endpoint', 'agent', 'model']) if (!values[key]?.trim()) throw new Error()
  const endpoint = new URL(values.endpoint)
  if (endpoint.protocol !== 'http:' || !['127.0.0.1', '[::1]'].includes(endpoint.hostname) || !endpoint.port || endpoint.username || endpoint.password || endpoint.pathname !== '/' || endpoint.search || endpoint.hash) throw new Error()
  values.endpoint = endpoint.origin
} catch {
  process.stderr.write('INVALID_ARGUMENTS: explicit current installation, isolated paths, loopback endpoint and workers=3..100 are required; use --help\n')
  process.exit(2)
}

process.umask(0o077)
const piLinux = isLinuxPiTarget(process.platform, values.agent, values.model, workers)
const linuxStructured = process.platform === 'linux' && (values.agent === 'codex' || piLinux)
const token = process.env.SWA_PROJECT_TOKEN
const report = { status: 'NOT_RUN', workers, agent: values.agent, model: values.model, model_execution: 'NOT_RUN', acceptance: 'REQUIRES_INDEPENDENT_REVIEW', checks: [], commands: [], unknown_requests: [], provider_events: [], started_workers: [] }
let output, candidate, dataRoot, project, inputDirectory, swa, runtimeId, controller, rootTask, runId, database, requireRuntime, processAPI, daemonOwner
let source, inputs, spec, jsonlParser, foldJournal, observer, capabilityRevision, completed = false
const sources = new Map(), witnesses = new Map(), stopped = new Set()
const sessionOwners = new Map()
const stopRequests = new Map(), releaseRequests = new Map()
const approvedFiles = new Map()
const codexViews = new Map()
const piViews = new Map()
let piLaunch, piSourceHome, piSourceRoot, piExecutable
let piProtocol
const piMetadata = new Map()
const piRecords = new Map(), piNativeFiles = new Map()
const piDriverSources = new Map()
const piPrivateSources = new Map()
const check = (name, detail) => report.checks.push({ name, status: 'PASS', detail })
const native = reply => object(object(reply.result).native)
async function ordinary(path, directory) {
  const canonical = await realpath(path)
  assert.ok(samePath(canonical, resolve(path)), 'redirected_path')
  for (let current = canonical; ; current = dirname(current)) {
    const entry = await lstat(current)
    assert.ok(!entry.isSymbolicLink(), 'symbolic_link_path')
    if (current === canonical) assert.ok(directory ? entry.isDirectory() : entry.isFile(), 'ordinary_input_required')
    else assert.ok(entry.isDirectory(), 'ordinary_ancestor_required')
    if (dirname(current) === current) return canonical
  }
}
async function piSourceMetadata(path, directory) {
  assert.equal(await ordinary(path, directory), path, 'pi_source_path_must_be_canonical')
  const info = await lstat(path, { bigint: true })
  assert.equal(info.uid, BigInt(process.getuid()))
  assert.equal(info.mode & 0o7777n, directory ? 0o700n : 0o600n)
  if (!directory) assert.equal(info.nlink, 1n)
  return { path, version: `${info.dev}:${info.ino}:${info.birthtimeNs}:${info.mtimeNs}:${info.size}`, identity: `${info.dev}:${info.ino}:${info.birthtimeNs}:${info.uid}:${info.mode}:${info.nlink}` }
}
async function readPiOwnedLaunch(revision, daemon) {
  const path = await ordinary(values['pi-launch-record'], false)
  const { value, bytes, source: before } = readPiPrivateJson(path, 16 * 1024, piProtocol.parseStrictPiJson)
  assert.equal(createHash('sha256').update(bytes).digest('hex'), values['pi-launch-sha256'], 'pi_owned_launch_hash_changed')
  const launch = validatePiLaunchRecord(value, { installation: candidate, source_revision: revision, data_root: dataRoot, endpoint: values.endpoint })
  assert.equal(launch.serve_pid, daemon.swa_service_job.serve_pid)
  const serve = await observeSource('incarnation', launch.serve_pid)
  assert.equal(serve.absent, false)
  assert.equal(serve.incarnation.boot_id, daemon.swa_service_job.boot_id)
  assert.equal(serve.incarnation.start_ticks, daemon.swa_service_job.serve_start_ticks)
  assert.equal(serve.incarnation.owner_uid, daemon.swa_service_job.owner_uid)
  assert.equal(await realpath(`/proc/${launch.serve_pid}/exe`), join(candidate, 'swa'))
  piSourceHome = launch.pi_source_home
  piSourceRoot = await piSourceMetadata(piSourceHome, true)
  const roots = [candidate, dataRoot, project, await ordinary(values['source-repository'], true), process.env.CODEX_HOME, process.env.CLAUDE_CONFIG_DIR, process.env.SWA_AGENT_PROFILE, ...(process.env.SWA_PROTECTED_ROOTS?.split(':') ?? [])].filter(Boolean)
  for (const root of roots) {
    const canonical = await realpath(root)
    assert.ok(!within(canonical, piSourceHome) && !within(piSourceHome, canonical), 'pi_source_overlaps_protected_root')
  }
  piMetadata.set(path, before)
  for (const name of ['auth.json', 'models.json', 'settings.json']) {
    const file = join(piSourceHome, name)
    piMetadata.set(file, await piSourceMetadata(file, false))
  }
  return { value: launch, serve: serve.incarnation, source: before }
}
async function verifyPiMetadata() {
  if (!piLinux) return
  assert.ok(piLaunch, 'pi_owned_launch_not_observed')
  assert.deepEqual(await piSourceMetadata(piSourceHome, true), piSourceRoot, 'pi_source_root_changed')
  for (const [path, expected] of piMetadata) assert.deepEqual(await piSourceMetadata(path, false), expected, 'pi_source_or_launch_metadata_changed')
  assert.deepEqual(await observeSource('incarnation', piLaunch.value.serve_pid), { absent: false, incarnation: piLaunch.serve }, 'pi_owned_serve_changed')
  assert.equal(await realpath(`/proc/${piLaunch.value.serve_pid}/exe`), join(candidate, 'swa'))
}
async function verifyPiDriverSources(revision) {
  const repository = await ordinary(values['source-repository'], true)
  for (const name of ['run-installed.mjs', 'public-case-definition.mjs', 'codex-event-view.mjs', 'prepare-source-observer.mjs', 'source-observer-entry.ts', 'extract-observer-source.py', 'pi-event-view.mjs', 'pi-source-observer-entry.ts']) {
    const relativePath = 'tests/swa-product-acceptance/' + name
    const path = await ordinary(fileURLToPath(new URL('./' + name, import.meta.url)), false)
    assert.equal(path, join(repository, relativePath), 'pi_driver_not_from_explicit_source_repository')
    const before = await stat(path, { bigint: true })
    assert.ok(before.size > 0n && before.size <= 1024n * 1024n)
    const actual = await readFile(path), original = execFileSync('git', ['--no-pager', 'show', `${revision}:${relativePath}`], { cwd: repository, timeout: 30_000, maxBuffer: 1024 * 1024, env: { ...process.env, GIT_OPTIONAL_LOCKS: '0', GIT_PAGER: 'cat' } })
    assert.ok(actual.equals(original), 'pi_driver_source_differs_from_installed_revision')
    const info = await stat(path, { bigint: true }), version = value => `${value.dev}:${value.ino}:${value.birthtimeNs}:${value.mtimeNs}:${value.size}`
    assert.equal(version(info), version(before))
    piDriverSources.set(path, { version: version(info), sha256: createHash('sha256').update(actual).digest('hex') })
  }
}
const safeText = text => {
  assert.ok(!token || !text.includes(token), 'credential_in_output_refused')
  return text
}
const save = async (name, value) => {
  const text = typeof value === 'string' ? value : JSON.stringify(value, null, 2) + '\n'
  await writeFile(join(output, name), safeText(text), { flag: 'wx', mode: 0o600 })
}
async function verifyObserverSources() {
  assert.ok(observer, 'same_revision_observer_not_prepared')
  for (const expected of [...observer.manifest.inputs, observer.manifest.journal_module, observer.manifest.observer, ...(piLinux ? [observer.manifest.protocol_module, observer.manifest.protocol_metafile] : [])]) {
    const path = await ordinary(expected.path, false)
    assert.ok(within(observer.root, path) || within(join(candidate, 'runtime'), path), 'observer_input_outside_approved_sources')
    const before = await stat(path, { bigint: true })
    const version = value => `${value.dev}:${value.ino}:${value.birthtimeNs}:${value.mtimeNs}:${value.size}`
    assert.equal(version(before), expected.version, 'observer_source_version_changed')
    assert.ok(before.size <= 8n * 1024n * 1024n)
    assert.equal(createHash('sha256').update(await readFile(path)).digest('hex'), expected.sha256, 'observer_source_bytes_changed')
    assert.equal(version(await stat(path, { bigint: true })), expected.version, 'observer_source_changed_during_read')
  }
}
async function observeSource(mode, input) {
  await verifyObserverSources()
  const node = join(candidate, 'node/node')
  const args = mode === 'incarnation' ? [observer.manifest.observer.path, mode, String(input)] : [observer.manifest.observer.path, mode]
  const text = execFileSync(node, args, { input: mode !== 'incarnation' ? input : undefined, encoding: 'utf8', timeout: 10_000, maxBuffer: 1024 * 1024, cwd: observer.home, env: { PATH: '/usr/bin:/bin', LANG: 'C.UTF-8', HOME: observer.home, TMPDIR: observer.home }, stdio: ['pipe', 'pipe', 'pipe'] })
  await verifyObserverSources()
  return object(JSON.parse(text))
}
function requestId(args) {
  const index = args.indexOf('--request-id')
  return index < 0 ? null : args[index + 1]
}
async function cli(args, expected = 0) {
  const ordinal = report.commands.length + 1
  const entry = { ordinal, args, request_id: requestId(args), mutation: !!requestId(args) && args[0] !== 'request', prepared_at: new Date().toISOString(), state: 'not-submitted' }
  report.commands.push(entry)
  // 原请求先独占落盘并同步，再允许修改到达产品。
  const commandFile = await open(join(output, `cli-${ordinal}.command.json`), 'wx', 0o600)
  try { await commandFile.writeFile(safeText(JSON.stringify({ ...entry, state: 'command-intent-only', effect: 'UNKNOWN_WITHOUT_ACTUAL_RECEIPT' }, null, 2) + '\n')); await commandFile.sync() }
  finally { await commandFile.close() }
  for (const flag of ['--input-file', '--spec-file', '--result-file']) {
    const index = args.indexOf(flag)
    if (index >= 0) {
      const path = args[index + 1]
      assert.ok(approvedFiles.has(path), 'public_input_was_not_verified')
      assert.equal(await readFile(await ordinary(path, false), 'utf8'), approvedFiles.get(path), 'public_input_changed_before_command')
    }
  }
  entry.state = 'pending'
  entry.started_at = new Date().toISOString()
  const child = spawn(swa, [...args, '--json'], { cwd: project, env: { ...process.env, SWA_ENDPOINT: values.endpoint, SWA_PROJECT_ID: values['project-id'] }, stdio: ['ignore', 'pipe', 'pipe'], windowsHide: true })
  const stdout = [], stderr = []
  let stdoutBytes = 0, stderrBytes = 0, transportUnknown = false
  const timer = setTimeout(() => { transportUnknown = true; child.kill('SIGTERM') }, 150_000)
  child.stdout.on('data', bytes => {
    stdoutBytes += bytes.length
    if (stdoutBytes > 16 * 1024 * 1024) { transportUnknown = true; child.kill('SIGTERM') }
    else stdout.push(bytes)
  })
  child.stderr.on('data', bytes => {
    stderrBytes += bytes.length
    if (stderrBytes <= 256 * 1024) stderr.push(bytes)
  })
  for (const stream of [child.stdout, child.stderr]) stream.on('error', () => { transportUnknown = true; child.kill('SIGTERM') })
  let result
  try {
    result = await new Promise((resolveExit, reject) => {
      child.once('error', reject)
      child.once('close', (code, signal) => resolveExit({ code, signal }))
    })
  } finally { clearTimeout(timer) }
  Object.assign(entry, result, { ended_at: new Date().toISOString(), pid: child.pid })
  const text = safeText(new TextDecoder('utf-8', { fatal: true }).decode(Buffer.concat(stdout)))
  await save(`cli-${ordinal}.stdout.json`, text)
  await save(`cli-${ordinal}.stderr.log`, safeText(Buffer.concat(stderr).toString('utf8')))
  const reply = text.trim() ? object(JSON.parse(text)) : null
  if (transportUnknown || result.signal || result.code !== expected || !reply || expected === 0 && reply.ok !== true) {
    entry.state = 'unknown-or-refused'
    if (entry.mutation) report.unknown_requests.push({ request_id: entry.request_id, args, original_command: ordinal })
    throw new Error('actual_cli_command_failed_preserve_original_request')
  }
  entry.state = expected === 0 ? 'completed-response-observed' : 'refused-response-observed'
  if (expected === 0 && runtimeId) assert.equal(reply.runtime_id, runtimeId, 'runtime_changed')
  return reply
}
const identity = value => ({ session: value.native_session_id, fence: value.fence, process: value.process, provider_link: value.provider_link })
async function queryWorker(dispatch) {
  const observed = native(await cli(['worker', 'show', '--controller', controller, '--dispatch', dispatch]))
  assert.equal(observed.worker.dispatchId, dispatch, 'worker_dispatch_changed')
  assert.equal(observed.worker.runtimeEpoch, runtimeId, 'worker_service_epoch_changed')
  return observed
}
async function originalStartBinding(worker, workerSource) {
  let request = object((await cli(['request', 'show', '--controller', controller, '--request-id', worker.request_id])).result)
  if (piLinux) for (const deadline = Date.now() + 30_000; request.current_input?.state !== 'accepted' && Date.now() < deadline;) {
    assert.ok(!['unknown', 'failed'].includes(request.current_input?.state), 'pi_original_input_unresolved')
    await delay(100)
    request = object((await cli(['request', 'show', '--controller', controller, '--request-id', worker.request_id])).result)
  }
  assert.equal(request.state, 'completed', 'original_start_not_completed')
  const binding = object(object(request.receipt).swa_input_binding)
  const input = object(request.current_input)
  assert.equal(input.state, 'accepted', 'original_start_input_not_accepted')
  assert.deepEqual(input.binding, binding, 'original_start_input_binding_changed')
  assert.equal(input.request_id, worker.request_id)
  assert.equal(input.dispatch_id, worker.dispatch_id)
  assert.equal(input.session_id, workerSource.native_session_id)
  assert.equal(binding.parent_request, worker.request_id)
  assert.equal(binding.dispatch_id, worker.dispatch_id)
  assert.equal(binding.session_id, workerSource.native_session_id)
  assert.equal(binding.agent, values.agent)
  assert.equal(binding.execution_mode, 'structured')
  assert.equal(binding.runtime_fence, workerSource.fence)
  assert.equal(input.operation_id, binding.operation_id)
  assert.equal(object(object(request.receipt).invocation).caller.settings_revision, capabilityRevision, 'project_settings_changed_after_static_discovery')
  if (values.agent === 'codex') {
    const selection = object(object(request.receipt).selection)
    assert.deepEqual(selection, workerSource.selection, 'original_native_selection_changed')
    assert.equal(selection.agent_id, 'codex')
    assert.equal(selection.version, '0.163.0')
    assert.equal(selection.model, values.model)
    assert.equal(selection.provider, 'openai')
    assert.equal(selection.native_launch.account.kind, 'codex-native-api-key')
    assert.equal(selection.native_launch.tools.mode, 'disabled')
    assert.equal(selection.native_launch.tools.native_contract, 'codex-cli-0.163.0')
  }
  if (piLinux) {
    await verifyPiMetadata()
    const selection = object(object(request.receipt).selection)
    assert.deepEqual(selection, workerSource.selection, 'original_pi_selection_changed')
    validatePiSelection(selection, { executable: piExecutable, sourceHome: piSourceHome, privateRoot: join(dataRoot, 'pi-workers', values['project-id']) })
    for (const source of [selection.native_launch.account.auth_source, selection.native_launch.account.models_source]) assert.equal((await piSourceMetadata(source.path, false)).version, source.version, 'pi_original_source_version_changed')
    assert.ok(selection.configuration_version.includes(`;${piMetadata.get(join(piSourceHome, 'settings.json')).version};${piMetadata.get(join(piSourceHome, 'auth.json')).version};`), 'pi_actual_configuration_version_changed')
  }
  return binding
}
async function sessionRecord(workerSource) {
  const path = await ordinary(workerSource.journal_database, false)
  assert.ok(within(dataRoot, path), 'session_journal_outside_private_root')
  const before = await stat(path, { bigint: true })
  const journalSource = { path, identity: `${before.dev}:${before.ino}:${before.birthtimeNs}` }
  const journal = new DatabaseSync(path, { readOnly: true })
  try {
    if (values.agent === 'codex' || piLinux) {
      const actual = journal.prepare(`SELECT session_id,
        CASE WHEN length(CAST(record_json AS BLOB)) BETWEEN 1 AND 1048576 THEN record_json ELSE NULL END AS record_json
        FROM agent_session_records WHERE session_id=?`).get(workerSource.native_session_id)
      assert.ok(actual && typeof actual.record_json === 'string', 'actual_session_record_missing_or_exceeds_bound')
      assert.equal(actual.session_id, workerSource.native_session_id)
      const decoded = await observeSource(piLinux ? 'decode-pi-record' : 'decode-record', actual.record_json)
      assert.equal(decoded.schema_version, 2)
      assert.equal(decoded.session_id, actual.session_id)
      assert.equal(decoded.provider, values.agent)
      assert.equal(decoded.model, values.model)
      assert.equal(decoded.head.handle.agent, values.agent)
      assert.equal(decoded.head.handle.transport, piLinux ? 'jsonl-rpc' : 'codex-app-server')
      assert.equal(decoded.lease.runtimeKind, 'native')
      assert.equal(decoded.lease.sessionId, actual.session_id)
      assert.equal(decoded.lease.handoffStage, null)
      assert.equal(decoded.lease.unreconciled, false)
      const after = await stat(path, { bigint: true })
      assert.equal(`${after.dev}:${after.ino}:${after.birthtimeNs}`, journalSource.identity, 'session_journal_replaced')
      if (piLinux) assert.deepEqual(decoded.account_home, { variable: 'PI_CODING_AGENT_DIR', path: join(dataRoot, 'pi-workers', values['project-id'], decoded.session_id) })
      if (piLinux) piRecords.set(decoded.session_id, { path, projection: decoded })
      return { journal_source: journalSource, session_id: decoded.session_id, provider: decoded.provider, model: decoded.model, head: decoded.head, lease: decoded.lease, ...(piLinux ? { account_home: decoded.account_home, journal_epoch: journal.prepare('SELECT epoch FROM journal_sessions WHERE session_id=?').get(decoded.session_id)?.epoch } : {}) }
    }
    const row = journal.prepare(`SELECT session_id,
      json_extract(record_json, '$.schemaVersion') AS schema_version,
      json_extract(record_json, '$.sessionId') AS record_session_id,
      json_extract(record_json, '$.provider') AS provider,
      json_extract(record_json, '$.options.model') AS model,
      json_extract(record_json, '$.providerHandleChain') AS handles,
      json_extract(record_json, '$.lease') AS lease
      FROM agent_session_records WHERE session_id=? AND json_valid(record_json)=1`).get(workerSource.native_session_id)
    assert.ok(row, 'actual_session_record_missing')
    assert.equal(row.schema_version, 2, 'session_record_schema_unreviewed')
    assert.equal(row.session_id, workerSource.native_session_id)
    assert.equal(row.record_session_id, row.session_id)
    assert.equal(row.provider, values.agent)
    assert.equal(row.model, values.model)
    const handles = JSON.parse(row.handles), lease = object(JSON.parse(row.lease))
    assert.ok(Array.isArray(handles) && handles.length > 0)
    // 此入口只支持现有 Pi/OpenCode neutral handles，typed handles 另行审查。
    const head = object(handles.at(-1)), handle = object(head.handle)
    assert.equal(handle.agent, values.agent)
    assert.equal(handle.transport, values.agent === 'pi' ? 'jsonl-rpc' : 'acp')
    assert.equal(lease.sessionId, row.session_id)
    assert.equal(lease.runtimeKind, 'native')
    assert.equal(lease.unreconciled, false, 'session_record_unreconciled')
    assert.equal(lease.handoffStage, null, 'session_record_handoff_in_progress')
    const after = await stat(path, { bigint: true })
    assert.equal(`${after.dev}:${after.ino}:${after.birthtimeNs}`, journalSource.identity, 'session_journal_replaced')
    return { journal_source: journalSource, session_id: row.session_id, provider: row.provider, model: row.model, head, lease: { claimStatus: lease.claimStatus, runtimeFence: lease.runtimeFence, provenHandleLinkId: lease.provenHandleLinkId, ownerProcess: lease.ownerProcess, deathEvidence: lease.deathEvidence } }
  } finally { journal.close() }
}
async function verifySessionOwner(record, workerSource, owner, worker) {
  assert.deepEqual(record.journal_source, owner.journal_source, 'original_session_journal_changed')
  assert.deepEqual(record.head, owner.head, 'original_provider_head_changed')
  assert.equal(record.head.mintedAtFence, owner.binding.runtime_fence)
  if (record.lease.claimStatus === 'live') {
    assert.equal(record.lease.runtimeFence, owner.binding.runtime_fence)
    assert.equal(record.lease.provenHandleLinkId, record.head.linkId)
    assert.deepEqual(record.lease.ownerProcess, owner.process, 'original_session_owner_changed')
  } else {
    assert.equal(record.lease.claimStatus, 'released', 'session_owner_unproven')
    assert.equal(record.lease.ownerProcess, null)
    assert.ok(record.lease.runtimeFence > owner.binding.runtime_fence)
    assert.equal(record.lease.deathEvidence?.kind, 'exit-observed')
    assert.equal(record.lease.deathEvidence?.ownerFence, owner.binding.runtime_fence)
    if (linuxStructured) {
      assert.ok(worker.stop?.reply && stopRequests.get(worker.dispatch_id)?.request_id === worker.stop.request_id, 'released_owner_has_no_original_stop_response')
      assert.equal(native(worker.stop.reply).state, 'stopped')
      const final = validateLinuxFinalCustody(workerSource.custody, (piLinux ? piViews : codexViews).get(workerSource.native_session_id))
      assert.deepEqual(final.observation.root, owner.incarnation)
    } else {
      assert.ok(stopped.has(worker.dispatch_id) && worker.stop?.proof, 'released_owner_has_no_original_complete_job_proof')
      assert.equal(worker.stop.proof.root.pid, owner.process.pid)
      assert.equal(worker.stop.proof.root.creationTimeMs, owner.process.processStartTimeMs)
    }
  }
  assert.equal(workerSource.process.runtime_id, owner.process.runtime)
  assert.equal(workerSource.process.pid, owner.process.pid)
  if (linuxStructured) {
    const { runtime_id, ...expected } = workerSource.process
    assert.deepEqual(expected, owner.incarnation)
    const observed = await observeSource('incarnation', owner.process.pid)
    if (record.lease.claimStatus === 'live') assert.deepEqual(observed, { absent: false, incarnation: expected })
    else assert.ok(observed.absent || JSON.stringify(observed.incarnation) !== JSON.stringify(expected), 'original_bootstrap_identity_still_alive')
  } else assert.equal(workerSource.process.creation_time_ms, owner.process.processStartTimeMs)
}
async function events(workerSource) {
  assert.equal(workerSource.availability, 'observed')
  assert.equal(workerSource.provider, values.agent)
  const worker = report.started_workers.find(item => item.source?.native_session_id === workerSource.native_session_id)
  assert.ok(worker, 'source_has_no_original_dispatch')
  const latest = object((await queryWorker(worker.dispatch_id)).source)
  assert.deepEqual(identity(latest), identity(workerSource), 'source_changed_its_original_binding')
  workerSource = latest
  let owner = sessionOwners.get(workerSource.native_session_id)
  const binding = owner?.binding ?? await originalStartBinding(worker, workerSource)
  const beforeRecord = await sessionRecord(workerSource)
  if (!owner) {
    assert.equal(beforeRecord.lease.claimStatus, 'live', 'first_owner_observation_must_be_live')
    const process = object(beforeRecord.lease.ownerProcess)
    assert.match(process.runtime, /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i, 'actual_process_runtime_incarnation_missing')
    assert.deepEqual(beforeRecord.head, workerSource.provider_link, 'provider_source_head_differs_from_actual_record')
    owner = { binding, process, head: beforeRecord.head, journal_source: beforeRecord.journal_source }
    if (linuxStructured) {
      const observed = await observeSource('incarnation', process.pid)
      assert.equal(observed.absent, false)
      owner.incarnation = observed.incarnation
    }
  }
  await verifySessionOwner(beforeRecord, workerSource, owner, worker)
  assert.ok(Number.isSafeInteger(workerSource.process.pid) && (linuxStructured || workerSource.process.creation_time_ms > 0))
  const path = await ordinary(workerSource.provider_events, false)
  assert.ok(within(dataRoot, path), 'provider_source_outside_private_root')
  const handle = await open(path, piLinux ? constants.O_RDONLY | constants.O_NOFOLLOW : 'r')
  let rows
  try {
    const before = await handle.stat({ bigint: true })
    assert.ok(before.size > 0n && before.size <= BigInt(linuxStructured ? 12 : 64) * 1024n * 1024n, 'provider_source_bound')
    if (linuxStructured) {
      let aggregate = before.size
      for (const [otherPath, value] of witnesses) if (otherPath !== path) aggregate += BigInt(value.bytes.length)
      assert.ok(aggregate <= 12n * 1024n * 1024n, 'aggregate_provider_bytes_exceed_bound')
    }
    const bytes = Buffer.concat(await Array.fromAsync(handle.createReadStream({ start: 0, end: Number(before.size) - 1, autoClose: false })))
    const signature = `${before.dev}:${before.ino}:${before.birthtimeNs}`
    const old = witnesses.get(path)
    if (old) assert.ok(old.signature === signature && bytes.length >= old.bytes.length && bytes.subarray(0, old.bytes.length).equals(old.bytes), 'provider_prefix_changed')
    assert.equal(bytes.length, Number(before.size))
    assert.equal(bytes.at(-1), 10, 'provider_source_incomplete_record')
    new TextDecoder('utf-8', { fatal: true }).decode(bytes)
    if (piLinux) await observeSource('validate-pi-trace', bytes)
    rows = []
    for await (const item of Readable.from([bytes]).pipe(jsonlParser.asStream())) rows.push(item.value)
    const after = await stat(path, { bigint: true })
    assert.equal(`${after.dev}:${after.ino}:${after.birthtimeNs}`, signature)
    assert.ok(after.size >= before.size)
    const afterRecord = await sessionRecord(workerSource)
    await verifySessionOwner(afterRecord, workerSource, owner, worker)
    if (beforeRecord.lease.claimStatus === afterRecord.lease.claimStatus) assert.deepEqual(afterRecord, beforeRecord, 'session_owner_changed_during_event_read')
    else assert.ok(beforeRecord.lease.claimStatus === 'live' && afterRecord.lease.claimStatus === 'released', 'unreviewed_session_owner_transition')
    witnesses.set(path, { signature, bytes, ...(piLinux ? { source_version: `${before.dev}:${before.ino}:${before.birthtimeNs}:${before.mtimeNs}:${before.size}` } : {}) })
  } finally { await handle.close() }
  if (values.agent === 'opencode') rows = rows.filter(row => row.session_id === workerSource.native_session_id && row.runtime_fence === workerSource.fence)
  if (values.agent === 'codex') {
    assert.equal(workerSource.runtime_epoch, runtimeId)
    const view = validateCodexEventView(rows, { session: workerSource.native_session_id, fence: owner.binding.runtime_fence, runtimeEpoch: runtimeId, process: workerSource.process, executable: workerSource.selection.executable, nodeExecutable: join(candidate, 'node/node'), nativeId: owner.head.handle.nativeId, model: values.model })
    const previous = codexViews.get(workerSource.native_session_id)
    if (previous) {
      assert.deepEqual(view.custody, previous.custody)
      assert.deepEqual(view.nativeCLI, previous.nativeCLI)
    } else {
      assert.deepEqual((await observeSource('incarnation', view.nativeCLI.pid)), { absent: false, incarnation: { platform: 'linux', pid: view.nativeCLI.pid, boot_id: view.nativeCLI.boot_id, start_ticks: view.nativeCLI.start_ticks, owner_uid: view.nativeCLI.owner_uid } })
      assert.deepEqual(view.custody.owner, daemonOwner)
    }
    if (rows.at(-1).event === 'custody_closed') assert.deepEqual(workerSource.custody, rows.at(-1).detail)
    codexViews.set(workerSource.native_session_id, view)
    let aggregate = Buffer.byteLength(JSON.stringify(rows))
    for (const [session, value] of sources) if (session !== workerSource.native_session_id) aggregate += Buffer.byteLength(JSON.stringify(value.rows))
    assert.ok(aggregate <= 12 * 1024 * 1024, 'aggregate_actual_provider_sources_exceed_bound')
  } else if (piLinux) {
    assert.equal(workerSource.runtime_epoch, runtimeId)
    const view = validatePiEventView(rows, { session: workerSource.native_session_id, fence: owner.binding.runtime_fence, runtimeEpoch: runtimeId, process: workerSource.process, executable: piExecutable, nodeExecutable: join(candidate, 'node/node'), sessionFile: owner.head.handle.nativeId, home: beforeRecord.account_home.path, journalEpoch: beforeRecord.journal_epoch })
    await piSourceMetadata(beforeRecord.account_home.path, true)
    await piSourceMetadata(join(beforeRecord.account_home.path, 'settings.json'), false)
    for (const name of ['auth.json', 'models.json']) {
      const path = join(beforeRecord.account_home.path, name), source = view.ready.sources.find(value => value.path === path)
      assert.equal((await piSourceMetadata(path, false)).version, source.version, 'pi_private_projection_source_changed')
      piPrivateSources.set(path, source)
    }
    const previous = piViews.get(workerSource.native_session_id)
    if (previous) {
      assert.deepEqual(view.custody, previous.custody)
      assert.deepEqual(view.nativeCLI, previous.nativeCLI)
      assert.deepEqual(view.ready, previous.ready)
    } else {
      assert.deepEqual(await observeSource('incarnation', view.nativeCLI.pid), { absent: false, incarnation: { platform: 'linux', pid: view.nativeCLI.pid, boot_id: view.nativeCLI.boot_id, start_ticks: view.nativeCLI.start_ticks, owner_uid: view.nativeCLI.owner_uid } })
      assert.deepEqual(view.custody.owner, daemonOwner)
    }
    if (view.finalRow) assert.deepEqual(workerSource.custody, view.finalRow.detail)
    piViews.set(workerSource.native_session_id, view)
  } else for (const row of rows) {
    assert.equal(row.process.pid, workerSource.process.pid)
    assert.equal(row.process.creationTimeMs, workerSource.process.creation_time_ms)
    assert.ok(Number.isFinite(row.monotonic_ms) && typeof row.observed_at === 'string')
    assert.ok(row.native_cli?.pid > 0 && row.native_cli.creationTimeMs > 0 && row.native_cli.pid !== row.process.pid)
    assert.equal(row.projection, values.agent === 'pi' ? 'actual_native_cli_lifecycle_and_delta_fields' : 'actual_native_acp_lifecycle_and_delta_fields')
  }
  if (!sessionOwners.has(workerSource.native_session_id)) check('worker.runtime_identity', { dispatch_id: worker.dispatch_id, session_id: workerSource.native_session_id, service_runtime_epoch: runtimeId, process_runtime_incarnation: owner.process.runtime, original_fence: owner.binding.runtime_fence })
  sessionOwners.set(workerSource.native_session_id, owner)
  sources.set(workerSource.native_session_id, { source: workerSource, rows })
  return rows
}
const isDelta = row => row.protocol === 'pi-jsonl-rpc' ? isPiDelta(row) : row.protocol === 'codex-app-server' ? isCodexDelta(row) : row.protocol === 'acp'
  ? row.event === 'session/update' && row.detail?.kind === 'known' && ['agent_message_chunk', 'agent_thought_chunk'].includes(row.detail.notification?.update?.sessionUpdate) && row.detail.notification.update.content?.type === 'text' && !!row.detail.notification.update.content.text
  : row.record?.type === 'message_update' && ['text_delta', 'thinking_delta'].includes(row.record.assistantMessageEvent?.type) && typeof row.record.assistantMessageEvent.delta === 'string' && !!row.record.assistantMessageEvent.delta
const starts = rows => piLinux ? piStarts(rows) : values.agent === 'codex' ? codexStarts(rows) : rows.filter(row => row.protocol === 'acp' ? row.event === 'session/prompt' : row.record?.type === 'agent_start')
const ends = rows => piLinux ? piEnds(rows) : values.agent === 'codex' ? codexEnds(rows) : rows.filter(row => row.protocol === 'acp' ? row.event === 'session/prompt/response' : row.record?.type === 'agent_end')
async function piReplayNativeInventory(view) {
  const path = view.ready.session_file
  const before = await piSourceMetadata(path, false)
  const handle = await open(path, constants.O_RDONLY | constants.O_NOFOLLOW)
  const version = info => `${info.dev}:${info.ino}:${info.birthtimeNs}:${info.mtimeNs}:${info.size}`
  let bytes
  try {
    const info = await handle.stat({ bigint: true })
    assert.equal(version(info), before.version)
    assert.ok(info.size > 0n && info.size <= 12n * 1024n * 1024n, 'pi_replay_native_bound')
    const buffer = Buffer.alloc(Number(info.size) + 1)
    let size = 0
    while (size < buffer.length) {
      const read = await handle.read(buffer, size, buffer.length - size, null)
      if (!read.bytesRead) break
      size += read.bytesRead
    }
    assert.equal(size, Number(info.size), 'pi_replay_native_changed_during_read')
    assert.equal(version(await handle.stat({ bigint: true })), before.version)
    assert.deepEqual(await piSourceMetadata(path, false), before)
    bytes = buffer.subarray(0, size)
  } finally { await handle.close() }
  const checked = await observeSource('validate-pi-native', bytes)
  assert.equal(checked.header.id, view.ready.session_id)
  const text = new TextDecoder('utf-8', { fatal: true }).decode(bytes)
  assert.equal(bytes.at(-1), 10)
  const entries = text.slice(0, -1).split('\n').map(line => piProtocol.parseStrictPiJson(line))
  assert.equal(entries.length, checked.record_count)
  return { path, source_version: before.version, sha256: createHash('sha256').update(bytes).digest('hex'), entry_ids: entries.slice(1).map(entry => entry.id) }
}
async function settledReplay(args, original) {
  const status = await cli(['request', 'show', '--controller', controller, '--request-id', requestId(args)])
  assert.equal(status.result.state, 'completed', 'unknown_request_must_not_replay')
  const executionCounts = async () => Promise.all([...sources.values()].map(async ({ source: actual }) => {
    const rows = await events(actual)
    const counts = { session: actual.native_session_id, fence: actual.fence, starts: starts(rows), ends: ends(rows) }
    if (piLinux) {
      const view = piViews.get(actual.native_session_id)
      counts.prompt_writes = rows.filter(row => row.event === 'write_intent' && row.detail.command === 'prompt')
      counts.user_echoes = rows.filter(row => row.event === 'native_record' && row.detail.frame.type === 'message_end' && row.detail.frame.message.role === 'user')
      counts.native = await piReplayNativeInventory(view)
    }
    return counts
  }))
  const before = await executionCounts()
  const replay = await cli(args)
  for (const key of ['request_id', 'controller', 'run_id', 'task_id', 'dispatch_id']) if (original[key] !== undefined) assert.equal(replay[key], original[key])
  const afterRequest = await cli(['request', 'show', '--controller', controller, '--request-id', requestId(args)])
  assert.deepEqual(afterRequest.result.receipt, status.result.receipt, 'replay_changed_original_receipt')
  assert.deepEqual(afterRequest.result.current_input, status.result.current_input, 'replay_changed_input_operation')
  assert.deepEqual(await executionCounts(), before, 'replay_started_another_native_operation')
  await delay(300)
  assert.deepEqual(await executionCounts(), before, 'replay_native_execution_count_changed_after_response')
  check('completed_request_replay.native_execution_counts', { request_id: requestId(args), sessions: before.map(row => ({ session: row.session, fence: row.fence, starts: row.starts.length, ends: row.ends.length, ...(piLinux ? { prompt_writes: row.prompt_writes.length, user_echoes: row.user_echoes.length, native_entry_ids: row.native.entry_ids, native_sha256: row.native.sha256 } : {}) })), observation_ms: 300 })
  return replay
}
async function proveWindowsTree(workerSource, checkpoint) {
  const proofs = []
  const log = await ordinary(values['service-log'], false)
  const offset = BigInt(checkpoint.size)
  for (const deadline = Date.now() + 2000; Date.now() < deadline && (await stat(log, { bigint: true })).size <= offset;) await delay(50)
  const handle = await open(log, 'r')
  const before = await handle.stat({ bigint: true })
  assert.equal(`${before.dev}:${before.ino}:${before.birthtimeNs}`, checkpoint.identity, 'service_log_replaced')
  assert.ok(before.size > offset, 'service_log_truncated_or_tree_proof_not_observed')
  assert.ok(before.size - offset <= 64n * 1024n * 1024n, 'service_log_observation_bound')
  const lines = createInterface({ input: handle.createReadStream({ start: Number(offset), end: Number(before.size) - 1, autoClose: false }), crlfDelay: Infinity })
  let lineNumber = 0
  try {
  for await (const line of lines) {
    lineNumber++
    if (!line.startsWith('{')) continue
    let row
    try { row = JSON.parse(line) } catch { continue }
    if (row.type !== 'swa_provider_job_tree_exited' || row.root?.pid !== workerSource.process.pid || row.root.creationTimeMs !== workerSource.process.creation_time_ms) continue
    assert.equal(row.observations.length, 2)
    assert.equal(row.independentIdentityObservations.length, 2)
    for (const observation of row.observations) {
      assert.equal(observation.observation.ownerPid, daemonOwner.pid)
      assert.equal(observation.observation.rootPid, workerSource.process.pid)
      assert.equal(observation.observation.rootCreationTimeMs, workerSource.process.creation_time_ms)
      assert.equal(observation.observation.assigned, true)
      assert.equal(observation.observation.rootExited, true)
      assert.deepEqual(observation.observation.members, [])
    }
    assert.ok(row.observations[1].elapsedMs > row.observations[0].elapsedMs)
    for (const observation of row.independentIdentityObservations) for (const item of observation.identities) assert.ok(item.observed == null || item.observed.creationTimeMs !== item.expected.creationTimeMs)
    proofs.push({ ...row, actual_source: { path: log, source_identity: checkpoint.identity, observed_after_byte: checkpoint.size.toString(), relative_line_number: lineNumber } })
  }
    const after = await stat(log, { bigint: true })
    assert.equal(`${after.dev}:${after.ino}:${after.birthtimeNs}`, checkpoint.identity)
    assert.ok(after.size >= before.size)
  } finally { lines.close(); await handle.close() }
  assert.equal(proofs.length, 1, 'complete_original_job_tree_proof_missing_or_ambiguous')
  for (let index = 0; index < 2; index++) {
    if (index) await delay(75)
    const rows = await new Promise((resolveRows, reject) => processAPI.getAllProcesses(rows => rows ? resolveRows(rows) : reject(new Error('process_inventory_unknown')), 4))
    const root = rows.find(row => row.pid === workerSource.process.pid)
    assert.ok(!root || root.creationTimeMs !== undefined && root.creationTimeMs !== workerSource.process.creation_time_ms)
  }
  return proofs[0]
}
async function stopWorker(worker, active = false) {
  assert.ok(!stopRequests.has(worker.dispatch_id) && !releaseRequests.has(worker.dispatch_id), 'preserve_original_stop_or_release_request')
  const args = ['worker', 'stop', '--controller', controller, '--dispatch', worker.dispatch_id, '--request-id', randomUUID()]
  const priorEvents = linuxStructured ? await events(worker.source) : null
  if (!piLinux) stopRequests.set(worker.dispatch_id, { args, request_id: requestId(args) })
  let checkpoint
  if (linuxStructured) {
    checkpoint = { sequence: priorEvents.at(-1).sequence, request_id: requestId(args) }
    if (piLinux && active) {
      const cycle = piViews.get(worker.source.native_session_id).active
      assert.ok(cycle && !cycle.end && !cycle.settled, 'pi_active_stop_checkpoint_is_not_running')
      Object.assign(checkpoint, { cid: cycle.cid, running: true })
    }
  }
  else {
    const log = await stat(await ordinary(values['service-log'], false), { bigint: true })
    checkpoint = active || !worker.source_checkpoint ? { identity: `${log.dev}:${log.ino}:${log.birthtimeNs}`, size: log.size.toString() } : worker.source_checkpoint
  }
  if (piLinux) stopRequests.set(worker.dispatch_id, { args, request_id: requestId(args) })
  const reply = await cli(args)
  assert.equal(native(reply).state, 'stopped')
  worker.stop = { request_id: requestId(args), args, reply, proof: null, active_stop_requested: active }
  let proof
  if (linuxStructured) {
    const rows = await events(worker.source), view = (piLinux ? piViews : codexViews).get(worker.source.native_session_id)
    proof = piLinux ? originalPiTreeProof(rows, view, checkpoint, active) : originalLinuxTreeProof(rows, view, checkpoint)
    for (const expected of [view.custody.root, view.custody.nativeState.identity]) {
      const observed = await observeSource('incarnation', expected.pid)
      assert.ok(observed.absent || JSON.stringify(observed.incarnation) !== JSON.stringify(expected), 'original_stopped_process_identity_still_alive')
    }
  } else proof = await proveWindowsTree(worker.source, checkpoint)
  stopped.add(worker.dispatch_id)
  worker.stop.proof = proof
  return worker.stop
}
async function readTranscript(dispatch) {
  const pages = [], seen = new Set(), messages = new Map()
  let cursor
  for (;;) {
    const reply = await cli(['worker', 'read', '--controller', controller, '--dispatch', dispatch, '--source', 'transcript', '--limit', '100', ...(cursor ? ['--cursor', cursor] : [])])
    const page = native(reply)
    assert.equal(page.source, 'transcript')
    assert.equal(page.sourceExact, true)
    assert.ok(Array.isArray(page.transcript.messages) && typeof page.transcript.limited === 'boolean')
    if (pages.length) assert.equal(page.sourceIdentity, pages[0].sourceIdentity)
    for (const message of page.transcript.messages) {
      if (messages.has(message.id)) assert.deepEqual(messages.get(message.id), message)
      messages.set(message.id, message)
    }
    pages.push(page)
    if (!page.transcript.limited) { assert.equal(page.contentComplete, true); break }
    cursor = page.transcript.nextCursor
    assert.ok(typeof cursor === 'string' && !seen.has(cursor), 'incomplete_or_cyclic_transcript_page')
    seen.add(cursor)
    assert.ok(pages.length < 1000, 'transcript_page_bound')
  }
  return { pages, messages: [...messages.values()] }
}

async function readCompleteNativeFile(value) {
  assert.ok(within(dataRoot, await ordinary(value.path, false)), 'native_rollout_outside_private_root')
  if (piLinux) await piSourceMetadata(value.path, false)
  const handle = await open(value.path, piLinux ? constants.O_RDONLY | constants.O_NOFOLLOW : 'r')
  let bytes
  try {
    const before = await handle.stat({ bigint: true })
    const version = info => `${info.dev}:${info.ino}:${info.birthtimeNs}:${info.mtimeNs}:${info.size}`
    assert.equal(version(before), value.source_version)
    assert.ok(before.size > 0n && before.size <= 12n * 1024n * 1024n)
    const buffer = Buffer.alloc(Number(before.size) + 1)
    let size = 0
    while (size < buffer.length) {
      const read = await handle.read(buffer, size, buffer.length - size, size)
      if (!read.bytesRead) break
      size += read.bytesRead
    }
    assert.equal(size, Number(before.size))
    assert.equal(version(await handle.stat({ bigint: true })), value.source_version)
    assert.equal(version(await stat(value.path, { bigint: true })), value.source_version)
    bytes = buffer.subarray(0, size)
  } finally { await handle.close() }
  assert.equal(bytes.at(-1), 10)
  assert.equal(new TextDecoder('utf-8', { fatal: true }).decode(bytes), value.content, 'complete_native_export_differs_from_actual_file')
  if (piLinux) piNativeFiles.set(value.path, { source_version: value.source_version, bytes })
  return bytes
}

async function piNativeRows(value) {
  assert.equal(value.format, 'pi-session-jsonl')
  const bytes = await readCompleteNativeFile(value)
  const checked = await observeSource('validate-pi-native', bytes)
  const rows = []
  for await (const item of Readable.from([bytes]).pipe(jsonlParser.asStream())) rows.push(item.value)
  assert.equal(rows.length, checked.record_count)
  assert.equal(rows.length, value.record_count)
  assert.deepEqual(rows[0], checked.header)
  assert.equal(checked.header.id, value.native_id)
  assert.equal(checked.header.cwd, project)
  assert.ok(value.path.endsWith(`_${value.native_id}.jsonl`))
  let parent = null
  const ids = new Set()
  for (const entry of rows.slice(1)) {
    assert.ok(!ids.has(entry.id) && entry.parentId === parent, 'pi_native_parent_chain_changed')
    ids.add(entry.id); parent = entry.id
    if (entry.type === 'model_change') { assert.equal(entry.provider, 'stepfun'); assert.equal(entry.modelId, 'step-3.5-flash') }
  }
  return { bytes, header: checked.header, entries: rows.slice(1) }
}

async function verifyPiJournalInputs(worker, original, bindings, actualNative) {
  const view = piViews.get(original.source.native_session_id)
  assert.ok(view && view.ready.session_id === worker.native_session.native_id && view.ready.session_file === worker.native_session.path)
  const folded = foldJournal({ sessionId: worker.journal.session_id, epoch: worker.journal.epoch })
  for (const row of worker.journal.rows) assert.equal(folded.add({ seq: row.seq, ts: row.ts, rowJson: JSON.stringify(row.row) }), true)
  const actual = folded.finish()
  assert.equal(actual.damage, null); assert.equal(actual.newer, null)
  for (const submission of actual.state.submissions.values()) assert.ok(!['pending', 'unknown'].includes(submission.dispatchState), 'pi_unresolved_journal_input')
  const messages = actualNative.entries.filter(entry => entry.type === 'message')
  const raw = sources.get(original.source.native_session_id).rows.filter(row => row.event === 'native_record' && row.detail.frame.type === 'message_end')
  assert.deepEqual(messages.map(entry => entry.message), raw.map(row => row.detail.frame.message), 'pi_raw_and_native_full_messages_differ')
  const submitted = []
  for (const binding of bindings) {
    const status = object((await cli(['request', 'show', '--controller', controller, '--request-id', binding.parent_request])).result.current_input)
    assert.equal(status.state, 'accepted'); assert.deepEqual(status.binding, binding)
    const submission = actual.state.submissions.get(binding.operation_id), receipt = actual.state.receipts.get(binding.operation_id)
    assert.deepEqual(submission, status.submission)
    assert.equal(submission?.dispatchState, 'accepted')
    assert.equal(submission?.fence, original.source.fence)
    assert.equal(submission?.payloadFingerprint, observer.module.fingerprintPiBody(binding.session_id, binding.body))
    assert.equal(submission?.clientMessageId, binding.operation_id)
    assert.ok(Number.isSafeInteger(submission?.submittedSequence) && submission.submittedSequence > 0)
    assert.equal(receipt?.providerItemId, submission.providerItemId)
    assert.equal(receipt?.clientMessageId, binding.operation_id)
    assert.equal(receipt?.cursor.epoch, worker.journal.epoch)
    assert.ok(Number.isSafeInteger(receipt?.cursor.sequence) && receipt.cursor.sequence > 0 && receipt.cursor.sequence <= actual.state.lastSequence)
    const item = actual.state.items.get(actual.state.aliases.get(submission.providerItemId) ?? submission.providerItemId)
    assert.equal(item?.body.kind, 'message'); assert.equal(item.body.role, 'user')
    assert.equal(item.agentId, undefined)
    assert.deepEqual(item.body.blocks, binding.body.blocks)
    assert.equal(item.turnScope?.kind, 'turn')
    assert.equal(actual.state.itemFences.get(item.itemId), binding.runtime_fence)
    const prompt = view.prompts.filter(value => value.cid === binding.operation_id)
    assert.equal(prompt.length, 1)
    assert.deepEqual(prompt[0].write.detail.binding, binding)
    assert.ok(prompt[0].echo && prompt[0].write.settled && prompt[0].write.reply)
    const rawIndex = raw.indexOf(prompt[0].echo)
    assert.ok(rawIndex >= 0 && messages[rawIndex].message.role === 'user')
    const baseline = prompt[0].write.detail.native_baseline
    if (baseline.source) {
      assert.equal(baseline.source.path, worker.native_session.path)
      assert.equal(baseline.source.version.split(':').slice(0, 3).join(':'), worker.native_session.source_version.split(':').slice(0, 3).join(':'))
      assert.ok(baseline.bytes >= 0 && baseline.bytes < actualNative.bytes.length)
      const prefix = actualNative.bytes.subarray(0, baseline.bytes)
      assert.equal(createHash('sha256').update(prefix).digest('hex'), baseline.sha256)
      if (prefix.length) {
        await observeSource('validate-pi-native', prefix)
        for await (const value of Readable.from([prefix]).pipe(jsonlParser.asStream())) assert.notEqual(value.value.id, messages[rawIndex].id, 'pi_user_entry_predates_original_submit')
      }
    } else assert.equal(baseline.absent, true)
    const turn = actual.state.items.get(item.turnScope.turnItemId)
    assert.equal(turn?.body.kind, 'turn'); assert.equal(turn.agentId, undefined)
    assert.equal(actual.state.itemFences.get(turn.itemId), binding.runtime_fence)
    submitted.push({ sequence: submission.submittedSequence, cid: binding.operation_id, nativeEntry: messages[rawIndex], baseline, turn })
  }
  submitted.sort((a, b) => a.sequence - b.sequence)
  assert.equal(submitted.length, view.prompts.length)
  for (const [index, row] of submitted.entries()) {
    if (index) {
      assert.ok(row.sequence > submitted[index - 1].sequence)
      assert.ok(!row.baseline.absent, 'pi_later_input_has_no_original_prefix')
    }
    assert.equal(row.cid, view.prompts[index].cid, 'pi_original_submission_order_changed')
  }
  for (const submission of actual.state.submissions.values()) if (submission.dispatchState === 'accepted') assert.ok(bindings.some(binding => binding.operation_id === submission.clientMessageId))
  const latest = submitted.at(-1)
  assert.ok(latest)
  for (const projection of view.projections) {
    const report = projection.detail.report
    assert.equal(report.journalEpoch, worker.journal.epoch)
    for (const step of report.steps.filter(value => value.written)) {
      const rows = worker.journal.rows.filter(value => value.epoch === report.journalEpoch && value.row.fence === original.source.fence)
      if (step.kind === 'settlement') assert.ok(rows.some(value => value.row.kind === 'lifecycle-batch' && value.row.settlementId === step.settlementId), 'pi_written_settlement_missing_from_actual_journal')
      else {
        const sameItem = itemId => itemId === step.itemId || (actual.state.aliases.get(itemId) ?? itemId) === (actual.state.aliases.get(step.itemId) ?? step.itemId)
        assert.ok(rows.some(value => value.row.kind === 'item' && sameItem(value.row.itemId) || value.row.kind === 'lifecycle-batch' && value.row.mutations.some(mutation => mutation.kind === 'item' && sameItem(mutation.itemId))), 'pi_written_item_missing_from_actual_journal')
      }
    }
  }
  check('pi.actual_chunks_and_full_native_messages', { session_id: original.source.native_session_id, native_id: view.ready.session_id, native_message_count: messages.length,
    cycles: view.cycles.map(cycle => ({ cid: cycle.cid, delta_count: cycle.deltas.length, assistant_messages: cycle.completedMessages.map(message => ({ timestamp: message.final.detail.frame.message.timestamp, content_indices: [...message.blocks.keys()], block_ends_observed: [...message.blocks.values()].every(block => !!block.ended), stop_reason: message.final.detail.frame.message.stopReason })) })) })
  if (worker.native_session.availability === 'complete') {
    const final = messages.at(-1), finalRaw = raw.at(-1)
    assert.equal(final.message.role, 'assistant'); assert.equal(final.message.stopReason, 'stop')
    assert.equal(final.id, worker.native_session.final.entry_id)
    assert.equal(worker.native_session.final.text, final.message.content.filter(block => block.type === 'text').map(block => block.text).join(''))
    assert.equal(latest.turn?.body.kind, 'turn'); assert.equal(latest.turn.body.state, 'completed'); assert.equal(latest.turn.body.outcome, 'success')
    const cycle = view.cycles.at(-1)
    assert.equal(cycle.cid, latest.cid)
    assert.ok(cycle.end && cycle.settled && cycle.idle && cycle.finalMessage === finalRaw)
    const closes = view.projections.filter(row => row.detail.eventType === 'text.close' && row.detail.sourceSequence === finalRaw.sequence)
    assert.ok(closes.length)
    for (const close of closes) {
      const candidates = view.projections.filter(row => row.sequence <= close.sequence && row.detail.providerItem === close.detail.providerItem)
      const itemIds = new Set(candidates.flatMap(row => row.detail.report.steps.filter(step => step.kind === 'item' && step.written).map(step => step.itemId)).filter(itemId => observer.module.matchesPiJournalItem(itemId, original.source.native_session_id, close.detail.providerItem)))
      assert.ok(itemIds.size)
      // 原transition可同时写barrier；只把真实本turn的assistant item作为当前文本证据。
      const assistantItems = [...itemIds].map(itemId => actual.state.items.get(actual.state.aliases.get(itemId) ?? itemId)).filter(item => item?.body.kind === 'message' && item.body.role === 'assistant' && item.turnScope?.turnItemId === latest.turn.itemId)
      assert.ok(assistantItems.length)
      for (const item of assistantItems) {
        assert.equal(item.agentId, undefined)
        assert.equal(item.turnScope?.turnItemId, latest.turn.itemId)
        assert.equal(actual.state.itemFences.get(item.itemId), original.source.fence)
        const index = final.message.content.findIndex((block, part) => close.detail.providerItem.endsWith(`:${part}:${block.type === 'text' ? 'assistant' : 'reasoning'}`))
        assert.ok(index >= 0)
        const block = final.message.content[index]
        assert.equal(item.body.blocks.flatMap(value => value.type === 'text' ? [value.text] : []).join(''), block.type === 'text' ? block.text : block.thinking)
      }
    }
  } else assert.equal(original.stop?.active_stop_requested, true, 'pi_only_real_stopped_worker_may_be_partial')
}

function piRootInventory() {
  const db = new DatabaseSync(database, { readOnly: true })
  try {
    db.exec('BEGIN')
    const receiptsWhere = `FROM mutation_receipts WHERE json_valid(receipt)=1 AND method='swa.controllerDeliver'
      AND json_extract(receipt,'$.project_id')=? AND json_extract(receipt,'$.run_id')=? AND json_extract(receipt,'$.root_task_id')=?`
    const params = [values['project-id'], runId, rootTask]
    const bound = db.prepare(`SELECT COUNT(*) AS count,SUM(LENGTH(CAST(receipt AS BLOB))) AS bytes FROM mutation_receipts WHERE request_id IN
      (SELECT request_id ${receiptsWhere} UNION SELECT request_id||'/native' ${receiptsWhere})`).get(...params, ...params)
    assert.ok(bound.count <= 100_000 && (bound.bytes ?? 0) <= 12 * 1024 * 1024)
    const receipts = db.prepare(`SELECT caller_fingerprint,request_id,method,state,payload_hash,receipt FROM mutation_receipts WHERE request_id IN
      (SELECT request_id ${receiptsWhere} UNION SELECT request_id||'/native' ${receiptsWhere}) ORDER BY caller_fingerprint,request_id`).all(...params, ...params)
    for (const receipt of receipts) assert.ok(!['pending', 'unknown'].includes(receipt.state), 'pi_root_has_unresolved_original_request')
    const taskBound = db.prepare('SELECT COUNT(*) AS count,SUM(LENGTH(CAST(spec AS BLOB))) AS bytes FROM tasks WHERE run_id=? AND (id=? OR parent_id=?)').get(runId, rootTask, rootTask)
    assert.ok(taskBound.count <= 101 && (taskBound.bytes ?? 0) <= 12 * 1024 * 1024)
    const tasks = db.prepare('SELECT id,run_id,parent_id,status,spec FROM tasks WHERE run_id=? AND (id=? OR parent_id=?) ORDER BY id').all(runId, rootTask, rootTask)
    assert.equal(tasks.filter(row => row.parent_id === rootTask).length, workers)
    const run = db.prepare('SELECT id,coordinator_handle,consumer_generation FROM runs WHERE id=?').get(runId)
    assert.equal(run.coordinator_handle, controller)
    const dispatches = db.prepare(`SELECT d.id,d.run_id,d.task_id,d.status,d.assignee_handle,d.assignee_orca_session_id,d.process_incarnation,d.contract_version,d.capability_revoked_at,
      w.runtime_epoch,w.state,w.stage,w.agent_terminal_handle FROM dispatch_contexts d JOIN tasks t ON t.id=d.task_id
      LEFT JOIN worker_dispatches w ON w.dispatch_id=d.id WHERE d.run_id=? AND t.parent_id=? ORDER BY d.id LIMIT 101`).all(runId, rootTask)
    assert.equal(dispatches.length, workers)
    assert.deepEqual(dispatches.map(row => row.id).sort(), report.started_workers.map(row => row.dispatch_id).sort())
    const resources = db.prepare(`SELECT id,origin_dispatch_id,owner_dispatch_id,terminal_handle,process_incarnation,ownership_state,release_state
      FROM worker_terminal_resources WHERE origin_dispatch_id IN (SELECT d.id FROM dispatch_contexts d JOIN tasks t ON t.id=d.task_id WHERE d.run_id=? AND t.parent_id=?)
      OR owner_dispatch_id IN (SELECT d.id FROM dispatch_contexts d JOIN tasks t ON t.id=d.task_id WHERE d.run_id=? AND t.parent_id=?) ORDER BY id LIMIT 101`).all(runId, rootTask, runId, rootTask)
    assert.equal(resources.length, workers)
    const value = JSON.stringify({ run, tasks, receipts, dispatches, resources })
    assert.ok(Buffer.byteLength(value) <= 12 * 1024 * 1024)
    return value
  } finally { db.close() }
}

function piFinalSourceCheck(inventory, result) {
  assert.equal(piRootInventory(), inventory, 'pi_complete_root_or_original_request_inventory_changed')
  const version = info => `${info.dev}:${info.ino}:${info.birthtimeNs}:${info.mtimeNs}:${info.size}`
  const privateFile = path => {
    assert.equal(realpathSync(path), path)
    const info = lstatSync(path, { bigint: true })
    assert.ok(info.isFile() && !info.isSymbolicLink() && info.uid === BigInt(process.getuid()) && (info.mode & 0o7777n) === 0o600n && info.nlink === 1n)
    return info
  }
  assert.equal(realpathSync(piSourceHome), piSourceHome)
  const rootInfo = lstatSync(piSourceHome, { bigint: true })
  assert.ok(rootInfo.isDirectory() && !rootInfo.isSymbolicLink() && rootInfo.uid === BigInt(process.getuid()) && (rootInfo.mode & 0o7777n) === 0o700n)
  assert.equal(version(rootInfo), piSourceRoot.version, 'pi_source_root_changed_before_export')
  for (const [path, expected] of piMetadata) assert.equal(version(privateFile(path)), expected.version, 'pi_source_metadata_changed_before_export')
  for (const [path, expected] of piNativeFiles) assert.equal(version(privateFile(path)), expected.source_version, 'pi_native_source_changed_before_export')
  for (const [path, expected] of piPrivateSources) assert.equal(version(privateFile(path)), expected.version, 'pi_private_projection_changed_before_export')
  for (const snapshot of piRecords.values()) {
    const home = snapshot.projection.account_home.path, info = lstatSync(home, { bigint: true })
    assert.equal(realpathSync(home), home)
    assert.ok(info.isDirectory() && !info.isSymbolicLink() && info.uid === BigInt(process.getuid()) && (info.mode & 0o7777n) === 0o700n)
    privateFile(join(home, 'settings.json'))
  }
  for (const expected of [...observer.manifest.inputs, observer.manifest.journal_module, observer.manifest.observer, observer.manifest.protocol_module, observer.manifest.protocol_metafile]) {
    assert.equal(realpathSync(expected.path), expected.path)
    assert.equal(version(statSync(expected.path, { bigint: true })), expected.version, 'pi_observer_source_changed_before_export')
  }
  for (const [path, expected] of piDriverSources) assert.equal(version(statSync(path, { bigint: true })), expected.version, 'pi_driver_changed_before_export')
  assert.deepEqual(observer.module.observeIncarnation(piLaunch.value.serve_pid), { absent: false, incarnation: piLaunch.serve })
  assert.deepEqual(observer.module.observeIncarnation(daemonOwner.pid), { absent: false, incarnation: daemonOwner })
  // 产品完整导出已含原 trace/native content，不能重复把同一原件另计一次。
  assert.ok(Buffer.byteLength(JSON.stringify(result)) <= 12 * 1024 * 1024, 'pi_total_export_observation_exceeds_bound')
  for (const worker of result.workers) {
    const session = worker.source.native_session_id, snapshot = piRecords.get(session)
    assert.ok(snapshot)
    const db = new DatabaseSync(snapshot.path, { readOnly: true })
    try {
      const original = db.prepare(`SELECT CASE WHEN length(CAST(record_json AS BLOB)) BETWEEN 1 AND 1048576 THEN record_json ELSE NULL END AS value
        FROM agent_session_records WHERE session_id=?`).get(session)
      assert.ok(typeof original?.value === 'string')
      assert.deepEqual(observer.module.decodePiRecord(piProtocol.parseStrictPiJson(original.value)), snapshot.projection, 'pi_record_or_original_owner_changed_before_export')
      assert.equal(db.prepare('SELECT epoch FROM journal_sessions WHERE session_id=?').get(session)?.epoch, worker.journal.epoch)
      assert.equal(db.prepare('SELECT MAX(seq) AS seq FROM journal_rows WHERE session_id=? AND epoch=?').get(session, worker.journal.epoch)?.seq, worker.journal.rows.at(-1)?.seq ?? 0)
    } finally { db.close() }
    const owner = sessionOwners.get(session), view = piViews.get(session)
    for (const incarnation of [owner.incarnation, view.custody.nativeState.identity]) {
      const observed = observer.module.observeIncarnation(incarnation.pid)
      if (snapshot.projection.lease.claimStatus === 'live') assert.deepEqual(observed, { absent: false, incarnation })
      else assert.ok(observed.absent || JSON.stringify(observed.incarnation) !== JSON.stringify(incarnation))
    }
  }
}

async function codexNativeUserTexts(value) {
  assert.equal(value.format, 'codex-rollout-jsonl')
  const bytes = await readCompleteNativeFile(value)
  const texts = []
  let count = 0
  for await (const item of Readable.from([bytes]).pipe(jsonlParser.asStream())) {
    assert.ok(++count <= 100_000)
    const row = object(item.value), payload = object(row.payload)
    if (row.type === 'response_item' && payload.type === 'message' && payload.role === 'user') texts.push(payload.content.map(part => part.text).join(''))
    if (row.type === 'event_msg' && payload.type === 'user_message') texts.push(payload.message)
    if (row.type === 'event_msg' && payload.type === 'item_completed' && payload.item?.type === 'UserMessage') texts.push(payload.item.content.map(part => part.text).join(''))
  }
  assert.equal(count, value.record_count)
  return texts
}

async function verifyCodexJournalInputs(worker, original, bindings) {
  const folded = foldJournal({ sessionId: worker.journal.session_id, epoch: worker.journal.epoch })
  for (const row of worker.journal.rows) assert.equal(folded.add({ seq: row.seq, ts: row.ts, rowJson: JSON.stringify(row.row) }), true)
  const actual = folded.finish()
  assert.equal(actual.damage, null)
  assert.equal(actual.newer, null)
  for (const submission of actual.state.submissions.values()) assert.ok(!['pending', 'unknown'].includes(submission.dispatchState), 'accepted_input_remains_unsettled')
  const submitted = []
  for (const binding of bindings) {
    const status = object((await cli(['request', 'show', '--controller', controller, '--request-id', binding.parent_request])).result.current_input)
    assert.equal(status.state, 'accepted')
    assert.deepEqual(status.binding, binding)
    const submission = actual.state.submissions.get(binding.operation_id), receipt = actual.state.receipts.get(binding.operation_id)
    assert.deepEqual(submission, status.submission, 'actual_original_submission_changed')
    assert.equal(submission?.fence, original.source.fence)
    assert.equal(receipt?.providerItemId, submission.providerItemId)
    const item = actual.state.items.get(actual.state.aliases.get(submission.providerItemId) ?? submission.providerItemId)
    assert.equal(item?.body.kind, 'message')
    assert.equal(item.body.role, 'user')
    assert.equal(item.turnScope?.kind, 'turn')
    const turn = actual.state.items.get(item.turnScope.turnItemId)
    assert.equal(turn?.body.kind, 'turn')
    assert.equal(actual.state.itemFences.get(turn.itemId), binding.runtime_fence)
    submitted.push({ sequence: submission.submittedSequence, turn: turn.body })
  }
  submitted.sort((left, right) => left.sequence - right.sequence)
  if (worker.native_session.availability === 'complete') {
    const latest = submitted.at(-1).turn
    assert.equal(latest.turnId, worker.native_session.final.turn_id)
    assert.equal(latest.state, 'completed')
    assert.equal(latest.outcome, 'success')
    assert.ok(ends(await events(original.source)).some(row => row.detail.frame.params.turn.id === latest.turnId && row.detail.frame.params.turn.status === 'completed'))
  }
}

async function run() {
  candidate = await ordinary(values.candidate, true)
  dataRoot = await ordinary(values['data-root'], true)
  project = await ordinary(values.project, true)
  await ordinary(dirname(values.output), true)
  const requestedOutput = resolve(values.output)
  for (const [left, right] of [[candidate, dataRoot], [candidate, project], [dataRoot, project], [requestedOutput, candidate], [requestedOutput, dataRoot], [requestedOutput, project]]) assert.ok(!within(left, right) && !within(right, left), 'input_roots_overlap')
  await mkdir(requestedOutput, { mode: 0o700 })
  output = requestedOutput
  const platform = (await readFile(join(candidate, 'PLATFORM'), 'utf8')).trim()
  assert.equal(platform, process.platform === 'win32' ? 'win32-x64' : 'linux-x64-glibc')
  assert.equal(process.arch, 'x64')
  const node = join(candidate, 'node', process.platform === 'win32' ? 'node.exe' : 'node')
  swa = join(candidate, process.platform === 'win32' ? 'swa.exe' : 'swa')
  assert.ok(samePath(await realpath(process.execPath), node), 'run_with_bundled_node')
  const revision = (await readFile(join(candidate, 'VERSION'), 'utf8')).trim()
  const edition = (await readFile(join(candidate, 'EDITION'), 'utf8')).trim()
  assert.ok(['cli', 'combined'].includes(edition), 'unknown_candidate_edition')
  assert.match(revision, /^[0-9a-f]{40}$/)
  assert.equal(execFileSync(swa, ['--version'], { encoding: 'utf8', timeout: 30_000, windowsHide: true }).trim(), revision)
  if (piLinux) await verifyPiDriverSources(revision)
  Object.assign(report, { source_revision: revision, candidate, data_root: dataRoot, project, endpoint: values.endpoint, platform, edition })
  check('candidate.version', revision)
  inputDirectory = join(project, '.verification', 'swa-product-input')
  source = JSON.parse(await readFile(await ordinary(join(inputDirectory, 'source.json'), false), 'utf8'))
  spec = JSON.parse(await readFile(await ordinary(join(inputDirectory, 'task.json'), false), 'utf8'))
  inputs = JSON.parse(await readFile(await ordinary(join(inputDirectory, 'inputs.json'), false), 'utf8'))
  assert.equal(source.source, publicSource)
  assert.equal(source.revision, publicRevision)
  assert.ok(samePath(source.project_root, project))
  assert.equal(spec.source_repo, source.source)
  assert.equal(spec.source_revision, source.revision)
  assert.equal(spec.max_workers, workers)
  assert.equal(inputs.length, workers)
  assert.equal(new Set(inputs.map(input => input.name)).size, workers)
  const git = args => execFileSync('git', ['-c', 'core.fsmonitor=false', '-c', 'core.hooksPath=' + join(dirname(project), 'empty-hooks'), '-C', project, ...args], { encoding: 'utf8', timeout: 30_000 })
  assert.equal(git(['rev-parse', 'HEAD']).trim(), source.revision)
  for (const original of source.originals) {
    assert.ok(publicFiles.includes(original.path))
    assert.deepEqual(Object.keys(original).sort(), ['path', 'source_revision', 'text'])
    assert.equal(original.source_revision, publicRevision)
    assert.equal(await readFile(await ordinary(join(project, original.path), false), 'utf8'), original.text)
    assert.equal(git(['show', source.revision + ':' + original.path]), original.text)
  }
  assert.equal(source.originals.length, 4)
  assert.deepEqual(source.originals.map(original => original.path), publicFiles)
  const prepared = publicCase(workers, source.originals)
  assert.deepEqual(spec, prepared.spec, 'public_task_spec_changed')
  assert.deepEqual(inputs, prepared.inputs.map(({ name, file }) => ({ name, input_file: join(inputDirectory, file) })), 'public_input_manifest_changed')
  for (const input of prepared.inputs) {
    const path = await ordinary(join(inputDirectory, input.file), false)
    assert.equal(await readFile(path, 'utf8'), input.body, 'public_worker_body_changed')
    approvedFiles.set(path, input.body)
  }
  assert.equal(await readFile(await ordinary(join(inputDirectory, 'followup.txt'), false), 'utf8'), prepared.followup, 'public_followup_changed')
  approvedFiles.set(join(inputDirectory, 'followup.txt'), prepared.followup)
  const specFile = join(inputDirectory, 'task.json'), specText = await readFile(await ordinary(specFile, false), 'utf8')
  assert.deepEqual(JSON.parse(specText), prepared.spec, 'public_task_spec_changed_before_snapshot')
  approvedFiles.set(specFile, specText)
  check('public.originals', { revision: source.revision, max_workers: workers })
  if (!token) throw new NotRun('current_project_grant_token_not_supplied_no_human_bootstrap')
  if (!(process.platform === 'win32' && ['pi', 'opencode'].includes(values.agent) || process.platform === 'linux' && values.agent === 'codex' && values.model === 'gpt-5.5' || piLinux)) throw new NotRun('selected_platform_and_agent_source_contract_not_implemented')
  requireRuntime = createRequire(join(candidate, 'runtime', 'package.json'))
  if (linuxStructured) {
    const processNode = await stat('/proc/self/exe', { bigint: true }), installedNode = await stat(node, { bigint: true })
    assert.equal(`${processNode.dev}:${processNode.ino}`, `${installedNode.dev}:${installedNode.ino}`, 'actual_running_node_inode_differs_from_installation')
    const toolEnvironment = { PATH: process.env.PATH ?? '/usr/bin:/bin', LANG: 'C.UTF-8', HOME: output, TMPDIR: output }
    const prepared = execFileSync(node, [fileURLToPath(new URL('./prepare-source-observer.mjs', import.meta.url)), '--candidate', candidate, '--source-repository', await ordinary(values['source-repository'], true), '--output', join(output, 'source-observer'), ...(piLinux ? ['--profile', 'stepfun-pi-n3'] : [])], { timeout: 90_000, maxBuffer: 4096, encoding: 'utf8', stdio: 'pipe', env: toolEnvironment })
    const root = join(output, 'source-observer'), manifestFile = await ordinary(join(root, 'manifest.json'), false)
    assert.ok((await stat(manifestFile)).size <= 4 * 1024 * 1024)
    observer = { root, home: join(root, 'observer-home'), manifest: JSON.parse(await readFile(manifestFile, 'utf8')) }
    assert.equal(observer.manifest.installed_revision, revision)
    if (piLinux) assert.equal(observer.manifest.profile, 'stepfun-pi-n3')
    await mkdir(observer.home, { mode: 0o700 })
    await verifyObserverSources()
    if (piLinux) {
      piProtocol = requireRuntime('./pi-protocol.cjs')
      assert.equal(typeof piProtocol.parseStrictPiJson, 'function')
      observer.module = createRequire(observer.manifest.observer.path)(observer.manifest.observer.path)
      assert.equal(typeof observer.module.decodePiRecord, 'function')
      assert.equal(typeof observer.module.observeIncarnation, 'function')
      assert.equal(typeof observer.module.matchesPiJournalItem, 'function')
      assert.equal(typeof observer.module.fingerprintPiBody, 'function')
      await verifyObserverSources()
    }
    check('same_revision_read_only_observer', JSON.parse(prepared))
  } else {
    assert.ok(within(dataRoot, await ordinary(values['service-log'], false)), 'select_only_this_services_private_log')
    processAPI = requireRuntime('@vscode/windows-process-tree')
  }
  const daemon = JSON.parse(await readFile(await ordinary(join(dataRoot, 'daemon', 'daemon-v49.pid'), false), 'utf8'))
  assert.equal(daemon.swa_service_job.source_revision, revision)
  assert.ok(samePath(daemon.swa_service_job.data_root, dataRoot) && samePath(daemon.swa_service_job.install_root, candidate))
  if (linuxStructured) {
    assert.equal(daemon.swa_service_job.platform, 'linux')
    const current = await observeSource('incarnation', daemon.pid)
    assert.equal(current.absent, false)
    daemonOwner = current.incarnation
    assert.equal(daemonOwner.boot_id, daemon.swa_service_job.boot_id)
    assert.equal(daemonOwner.start_ticks, daemon.swa_service_job.daemon_start_ticks)
    assert.equal(daemonOwner.owner_uid, daemon.swa_service_job.owner_uid)
    if (piLinux) piLaunch = await readPiOwnedLaunch(revision, daemon)
  } else {
    const security = requireRuntime('@orca/windows-security')
    daemonOwner = security.readServiceJobMember(daemon.swa_service_job.name, daemon.pid)
    assert.equal(daemonOwner.creationTimeMs, daemon.swa_service_job.daemon_created_at_ms)
    assert.equal(daemonOwner.ownerSid, security.currentUserSid())
  }
  jsonlParser = (await import(pathToFileURL(requireRuntime.resolve('stream-json/jsonl/parser.js')).href)).default
  foldJournal = requireRuntime('./journal-open.cjs').startJournalRowFold
  const checker = join(candidate, 'scripts', 'check-swa-runtime.mjs')
  try { await ordinary(checker, false) } catch { throw new NotRun('installed_production_checker_not_packaged_no_source_substitution') }
  execFileSync(node, [checker, '--candidate', candidate], { stdio: 'pipe', timeout: 60_000, windowsHide: true })
  check('candidate.complete_native_closure', 'existing_production_checker_completed')
  const status = await cli(['status'])
  runtimeId = status.runtime_id
  assert.ok(runtimeId)
  database = status.result.databases.find(row => row.name === 'main')?.file
  assert.ok(database && within(dataRoot, await ordinary(database, false)))
  const capabilities = (await cli(['capabilities'])).result
  assert.equal(capabilities.source_revision, revision)
  const agent = capabilities.agents.find(row => row.id === values.agent)
  if (!agent || agent.configured.model !== values.model) throw new NotRun('selected_project_agent_or_model_not_observed')
  for (const observation of [agent.registered, agent.enabled, agent.installed, agent.configured, ...['start', 'send', 'stop'].map(key => agent.control[key])]) {
    assert.ok(['available', 'unknown', 'unavailable'].includes(observation.status), 'invalid_observation_status')
    if (observation.status !== 'available') throw new NotRun('static_source_or_declared_control_not_observed_available')
  }
  for (const observation of [agent.launchable, ...['read', 'tools'].map(key => agent.control[key])]) {
    assert.ok(['available', 'unknown', 'unavailable'].includes(observation.status), 'invalid_observation_status')
    if (observation.status === 'unavailable') throw new NotRun('live_capability_explicitly_unavailable')
  }
  capabilityRevision = agent.requirements.settings_revision
  assert.ok(Number.isSafeInteger(capabilityRevision) && capabilityRevision > 0, 'actual_project_settings_revision_missing')
  if (values.agent === 'codex') {
    assert.equal(agent.installed.version, '0.163.0')
    assert.equal(agent.configured.provider, 'openai')
    assert.equal(agent.configured.account_source, 'codex-native-api-key')
    await ordinary(agent.installed.executable, false)
  }
  if (piLinux) {
    assert.equal(agent.installed.version, '1.0.1')
    assert.equal(agent.configured.provider, 'stepfun')
    assert.equal(agent.configured.account_source, 'pi-api-key-file')
    piExecutable = await ordinary(agent.installed.executable, false)
    assert.equal(piExecutable, agent.installed.executable)
    await verifyPiMetadata()
    check('pi.independent_owned_source', { launch: piLaunch.source, source_home: piSourceHome, provider: 'stepfun', model: values.model, auth: 'NOT_READ_BY_DRIVER' })
  }
  Object.assign(report, { runtime_id: runtimeId, capabilities })
  const opened = await cli(['controller', 'open', '--project', values['project-id'], '--workdir', project, '--request-id', randomUUID()])
  controller = opened.controller
  assert.ok(controller)
  const createArgs = ['task', 'create', '--controller', controller, '--spec-file', join(inputDirectory, 'task.json'), '--request-id', randomUUID()]
  const created = await cli(createArgs)
  rootTask = created.task_id; runId = created.run_id
  assert.ok(rootTask && runId)
  Object.assign(report, { controller, run_id: runId, root_task_id: rootTask })
  await settledReplay(createArgs, created)
  const startArgs = []
  const beforeStarts = linuxStructured ? null : await stat(values['service-log'], { bigint: true })
  const sourceCheckpoint = beforeStarts ? { identity: `${beforeStarts.dev}:${beforeStarts.ino}:${beforeStarts.birthtimeNs}`, size: beforeStarts.size.toString() } : null
  for (const [index, input] of inputs.entries()) {
    assert.ok(within(inputDirectory, await ordinary(input.input_file, false)))
    let body = await readFile(input.input_file, 'utf8')
    assert.equal(body, prepared.inputs[index].body, 'public_worker_body_changed_before_start')
    const publicCopy = '\n同一固定公开原文的复核副本，不增加事实：\n' + JSON.stringify(source.originals) + '\n'
    while (Buffer.byteLength(body) <= 2 * 16 * 1024) body += publicCopy
    const file = join(inputDirectory, input.name + '-chunked-' + randomUUID() + '.txt')
    await writeFile(file, body, { flag: 'wx', mode: 0o600 })
    approvedFiles.set(file, body)
    startArgs.push(['worker', 'start', '--controller', controller, '--task', rootTask, '--name', input.name, '--agent', values.agent, '--model', values.model, '--input-file', file, '--request-id', randomUUID()])
  }
  report.model_execution = 'REQUESTED_THROUGH_PRODUCT'
  const startsSettled = await Promise.allSettled(startArgs.map(args => cli(args)))
  for (let index = 0; index < startsSettled.length; index++) {
    const item = startsSettled[index]
    if (item.status === 'fulfilled') {
      const reply = item.value
      assert.equal(reply.run_id, runId)
      assert.ok(reply.dispatch_id && reply.task_id && reply.task_id !== rootTask)
      report.started_workers.push({ dispatch_id: reply.dispatch_id, task_id: reply.task_id, request_id: requestId(startArgs[index]), name: inputs[index].name, start: reply, source_checkpoint: sourceCheckpoint, actual_input_file: startArgs[index][startArgs[index].indexOf('--input-file') + 1] })
    }
  }
  assert.ok(startsSettled.every(item => item.status === 'fulfilled'), 'some_worker_starts_refused_or_unknown_preserve_all_requests')
  assert.equal(new Set(report.started_workers.map(worker => worker.dispatch_id)).size, workers)
  assert.equal(new Set(report.started_workers.map(worker => worker.task_id)).size, workers)
  for (const worker of report.started_workers) worker.source = object((await queryWorker(worker.dispatch_id)).source)
  assert.equal(new Set(report.started_workers.map(worker => worker.source.native_session_id)).size, workers)
  assert.equal(new Set(report.started_workers.map(worker => worker.source.process.pid)).size, workers)
  if (piLinux) {
    await Promise.all(report.started_workers.map(worker => events(worker.source)))
    assert.equal(new Set([...sessionOwners.values()].map(owner => owner.binding.session_id)).size, workers)
    assert.equal(new Set([...piViews.values()].map(view => view.ready.session_id)).size, workers)
    assert.equal(new Set([...piViews.values()].map(view => view.ready.session_file)).size, workers)
    assert.equal(new Set([...piViews.values()].map(view => view.nativeCLI.pid)).size, workers)
    assert.equal(new Set([...piViews.values()].map(view => view.ready.acquisition_generation)).size, workers)
  }
  let overlap
  for (const deadline = Date.now() + 180_000; Date.now() < deadline;) {
    const lists = await Promise.all(report.started_workers.map(worker => events(worker.source)))
    assert.ok(lists.every(rows => starts(rows).length === 1 && ends(rows).length === 0), 'turn_ended_before_all_worker_overlap')
    if (piLinux) for (const worker of report.started_workers) {
      const cycle = piViews.get(worker.source.native_session_id).active
      assert.ok(cycle && !cycle.end && !cycle.settled, 'pi_run_ended_before_all_worker_overlap')
    }
    const deltas = lists.map(rows => rows.filter(isDelta))
    if (deltas.every(rows => rows.length >= 2)) {
      const from = Math.max(...deltas.map(rows => rows[0].monotonic_ms)), to = Math.min(...deltas.map(rows => rows.at(-1).monotonic_ms))
      if (from < to) { overlap = { from_monotonic_ms: from, to_monotonic_ms: to, evidence: deltas.map(rows => ({ first: rows[0], last: rows.at(-1) })) }; break }
    }
    await delay(250)
  }
  assert.ok(overlap, 'no_common_actual_native_delta_interval')
  report.overlap = overlap
  const target = report.started_workers[0], survivors = report.started_workers.slice(1)
  assert.equal(ends(await events(target.source)).length, 0)
  const stoppedAt = Date.now()
  await stopWorker(target, true)
  for (const worker of survivors) {
    assert.deepEqual(identity((await queryWorker(worker.dispatch_id)).source), identity(worker.source))
    const rows = await events(worker.source)
    if (!piLinux) assert.ok(rows.some(row => isDelta(row) && Date.parse(row.observed_at) > stoppedAt), 'survivor_has_no_actual_output_after_stop')
  }
  for (const deadline = Date.now() + 600_000; Date.now() < deadline;) {
    const lists = await Promise.all(survivors.map(worker => events(worker.source)))
    if (lists.every((rows, index) => ends(rows).length === 1 && (!piLinux || piViews.get(survivors[index].source.native_session_id).cycles[0]?.idle))) break
    await delay(500)
  }
  for (const worker of survivors) assert.equal(ends(await events(worker.source)).length, 1, 'survivor_first_turn_not_complete')
  if (piLinux) {
    const proof = target.stop.proof
    for (const worker of survivors) {
      const view = piViews.get(worker.source.native_session_id)
      assert.deepEqual(view.custody.owner, piViews.get(target.source.native_session_id).custody.owner)
      const rows = await events(worker.source)
      assert.ok(rows.some(row => isPiDelta(row) && row.monotonic_ms > proof.stop_intent.monotonic_ms && row.monotonic_ms > proof.row.monotonic_ms), 'pi_survivor_has_no_same_daemon_monotonic_delta_after_full_stop')
      assert.ok(view.cycles[0].idle, 'pi_survivor_not_idle_before_followup')
    }
  }
  const follow = survivors[0], followupFile = join(inputDirectory, 'followup.txt'), sendRequest = randomUUID()
  const followup = await readFile(await ordinary(followupFile, false), 'utf8')
  assert.equal(followup, prepared.followup, 'public_followup_changed_before_send')
  const sendArgs = ['worker', 'send', '--controller', controller, '--dispatch', follow.dispatch_id, '--input-file', followupFile, '--request-id', sendRequest]
  const sent = await cli(sendArgs), input = native(sent)
  assert.equal(input.session_id, follow.source.native_session_id)
  assert.equal(input.request_id, sendRequest)
  assert.equal(input.swa_input_binding.dispatch_id, follow.dispatch_id)
  assert.equal(input.swa_input_binding.body.kind, 'message')
  assert.deepEqual(input.swa_input_binding.body.blocks, [{ type: 'text', text: followup }])
  for (const deadline = Date.now() + 600_000; Date.now() < deadline;) {
    if (ends(await events(follow.source)).length >= 2) break
    await delay(500)
  }
  assert.equal(starts(await events(follow.source)).length, 2)
  assert.equal(ends(await events(follow.source)).length, 2)
  await settledReplay(sendArgs, sent)
  assert.deepEqual(identity((await queryWorker(follow.dispatch_id)).source), identity(follow.source))
  const sendStatus = (await cli(['request', 'show', '--controller', controller, '--request-id', sendRequest])).result.current_input
  assert.equal(sendStatus.state, 'accepted')
  assert.equal(sendStatus.session_id, follow.source.native_session_id)
  assert.equal(sendStatus.operation_id, input.operation_id)
  report.followup = { request_id: sendRequest, reply: sent, current_input: sendStatus }
  const extra = await cli(['worker', 'start', '--controller', controller, '--task', rootTask, '--name', 'extra-quota-check', '--agent', values.agent, '--model', values.model, '--input-file', inputs[0].input_file, '--request-id', randomUUID()], 5)
  assert.equal(extra.error.code, 'resource_limit')
  const fleet = native(await cli(['worker', 'list', '--controller', controller, '--task', rootTask]))
  assert.equal(fleet.page.hasMore, false)
  assert.equal(fleet.page.total, workers)
  assert.deepEqual(fleet.workers.map(worker => worker.dispatchId).sort(), report.started_workers.map(worker => worker.dispatch_id).sort())
  report.quota = { extra, fleet }
  const inventory = native(await cli(['task', 'show', '--controller', controller, '--task', rootTask])).tasks
  assert.equal(inventory.filter(task => task.parent_id === rootTask).length, workers, 'quota_refusal_created_extra_child_task')
  for (let index = 0; index < startArgs.length; index++) await settledReplay(startArgs[index], startsSettled[index].value)
  const changedInput = [...startArgs[0]]
  changedInput[changedInput.indexOf('--input-file') + 1] = inputs[1].input_file
  const mismatched = await cli(changedInput, 2)
  assert.equal(mismatched.error.code, 'request_mismatch')
  report.changed_original_request_refusal = mismatched
  await settledReplay(target.stop.args, target.stop.reply)
  for (const worker of report.started_workers) worker.transcript = await readTranscript(worker.dispatch_id)
  const piInventory = piLinux ? piRootInventory() : null
  const exported = await cli(['task', 'result', '--controller', controller, '--task', rootTask, '--output-dir', join(inputDirectory, 'result-' + randomUUID())])
  const result = native(exported)
  assert.equal(result.workers.length, workers)
  for (const worker of result.workers) {
    const original = report.started_workers.find(item => item.dispatch_id === worker.worker.dispatchId)
    assert.ok(original && worker.source.native_session_id === original.source.native_session_id)
    if (original !== target) assert.equal(worker.native_session.availability, 'complete')
    const originalRequest = (await cli(['request', 'show', '--controller', controller, '--request-id', original.request_id])).result
    assert.equal(originalRequest.state, 'completed')
    const binding = originalRequest.receipt.swa_input_binding
    assert.equal(binding.parent_request, original.request_id)
    assert.equal(binding.dispatch_id, original.dispatch_id)
    assert.equal(binding.session_id, original.source.native_session_id)
    const body = await readFile(original.actual_input_file, 'utf8')
    assert.equal(originalRequest.receipt.invocation.input.text, body)
    const textOf = message => typeof message.content === 'string' ? message.content : message.content.filter(part => part.type === 'text').map(part => part.text).join('\n')
    const piNative = piLinux ? await piNativeRows(worker.native_session) : undefined
    const userTexts = piNative ? piNative.entries.filter(entry => entry.type === 'message' && entry.message.role === 'user').map(entry => textOf(entry.message)) : worker.native_session.format === 'codex-rollout-jsonl' ? await codexNativeUserTexts(worker.native_session) : worker.native_session.format === 'opencode-sqlite-v2'
      ? worker.native_session.messages.filter(message => message.type === 'user').map(message => message.data.text)
      : worker.native_session.entries.filter(entry => entry.type === 'message' && entry.message.role === 'user').map(entry => textOf(entry.message))
    const originalText = binding.body.blocks.filter(block => block.type === 'text').map(block => block.text).join('\n')
    assert.ok(originalText.includes(body) && userTexts.includes(originalText), 'actual_original_input_or_native_source_differs')
    if (original.dispatch_id === follow.dispatch_id) {
      assert.ok(userTexts.includes(followup), 'actual_followup_user_body_missing_from_native_source')
      assert.notEqual(binding.operation_id, input.operation_id, 'followup_must_have_its_own_operation')
    }
    if (values.agent === 'codex') {
      assert.equal(worker.native_session.native_id, original.source.provider_link.handle.nativeId)
      const actualRows = await events(original.source)
      assert.equal(worker.provider_events.format, 'codex-provider-events-ndjson')
      await readCompleteNativeFile(worker.provider_events)
      assert.equal(worker.provider_events.record_count, actualRows.length)
      await verifyCodexJournalInputs(worker, original, original.dispatch_id === follow.dispatch_id ? [binding, input.swa_input_binding] : [binding])
    }
    if (piLinux) {
      const actualRows = await events(original.source)
      const view = piViews.get(original.source.native_session_id)
      assert.equal(worker.native_session.native_id, view.ready.session_id)
      assert.equal(worker.native_session.path, view.ready.session_file)
      assert.equal(worker.provider_events.format, 'pi-native-rpc-events-ndjson')
      assert.equal(worker.provider_events.path, original.source.provider_events, 'pi_export_trace_path_not_original_source')
      const rawBytes = await readCompleteNativeFile(worker.provider_events)
      const actualWitness = witnesses.get(original.source.provider_events)
      assert.equal(worker.provider_events.source_version, actualWitness.source_version, 'pi_export_trace_version_not_actual_snapshot')
      assert.ok(rawBytes.equals(actualWitness.bytes), 'pi_export_trace_differs_from_actual_original_source')
      const validated = await observeSource('validate-pi-trace', rawBytes)
      assert.equal(validated.record_count, actualRows.length)
      assert.equal(worker.provider_events.record_count, actualRows.length)
      await verifyPiJournalInputs(worker, original, original.dispatch_id === follow.dispatch_id ? [binding, input.swa_input_binding] : [binding], piNative)
    }
    const journal = new DatabaseSync(original.source.journal_database, { readOnly: true })
    try {
      for (const row of worker.journal.rows) {
        const actual = journal.prepare('SELECT ts,row_json FROM journal_rows WHERE session_id=? AND epoch=? AND seq=?').get(worker.journal.session_id, row.epoch, row.seq)
        assert.ok(actual && actual.ts === row.ts)
        assert.deepEqual(JSON.parse(actual.row_json), row.row)
      }
    } finally { journal.close() }
  }
  const stoppedOutput = result.workers.find(worker => worker.worker.dispatchId === target.dispatch_id)
  const mark = stoppedOutput.journal.rows.find(row => row.row.kind === 'tombstone' && row.row.stopEvent?.reason === 'evict' && row.row.stopEvent.at >= stoppedAt)
  assert.ok(mark, 'actual_active_stop_mark_missing')
  if (!piLinux) for (const worker of survivors) assert.ok((await events(worker.source)).some(row => isDelta(row) && Date.parse(row.observed_at) > mark.row.stopEvent.at), 'survivor_has_no_actual_delta_after_native_stop_mark')
  const fold = foldJournal({ sessionId: stoppedOutput.journal.session_id, epoch: stoppedOutput.journal.epoch })
  for (const row of stoppedOutput.journal.rows.filter(row => row.seq < mark.seq)) assert.equal(fold.add({ seq: row.seq, ts: row.ts, rowJson: JSON.stringify(row.row) }), true)
  const prior = fold.finish()
  assert.equal(prior.damage, null)
  assert.equal(prior.newer, null)
  assert.ok([...prior.state.items.values()].some(item => item.body.kind === 'turn' && item.body.turnId === mark.row.stopEvent.turnId && item.body.state === 'running'))
  const endedTurn = stoppedOutput.journal.snapshot.items.find(item => item.body.kind === 'turn' && item.body.turnId === mark.row.stopEvent.turnId)
  const stoppedEvents = await events(target.source)
  const actualAbort = stoppedEvents.find(row => row.protocol === 'pi-jsonl-rpc'
    ? row.event === 'native_record' && row.detail.frame.type === 'message_end' && row.detail.frame.message.role === 'assistant' && row.detail.frame.message.stopReason === 'aborted' && row.sequence > target.stop.proof.stop_intent.sequence
    : Date.parse(row.observed_at) >= mark.row.stopEvent.at && (row.protocol === 'codex-app-server' ? row.event === 'notification' && row.detail.method === 'turn/completed' && row.detail.frame.params.turn.status === 'interrupted' : row.protocol === 'acp' ? row.event === 'session/prompt/response' && row.detail.response.stopReason === 'cancelled' : row.record?.type === 'message_end' && row.record.message?.role === 'assistant' && row.record.message.stopReason === 'aborted'))
  if (piLinux) {
    assert.ok(actualAbort && endedTurn && (endedTurn.body.state === 'interrupted' || endedTurn.body.outcome === 'cancellation'), 'pi_actual_same_cycle_abort_missing')
    assert.equal(target.stop.proof.stop_intent.detail.turn_id, mark.row.stopEvent.turnId, 'pi_stop_turn_differs_from_actual_journal_mark')
    assert.equal(target.stop.proof.stop_intent.detail.cid, sessionOwners.get(target.source.native_session_id).binding.operation_id)
  } else assert.ok(endedTurn && (endedTurn.body.state === 'interrupted' || endedTurn.body.outcome === 'cancellation' && actualAbort), 'active_native_turn_cancellation_not_proven')
  if (piLinux) {
    await verifyPiMetadata(); await verifyObserverSources()
    piFinalSourceCheck(piInventory, result)
    const fd = openSync(join(output, 'actual-result.json'), constants.O_CREAT | constants.O_EXCL | constants.O_WRONLY | constants.O_NOFOLLOW, 0o600)
    try { writeFileSync(fd, safeText(JSON.stringify(exported, null, 2) + '\n')); fsyncSync(fd) } finally { closeSync(fd) }
  } else await save('actual-result.json', exported)
  for (const worker of survivors) await stopWorker(worker)
  for (const worker of report.started_workers) {
    const args = ['worker', 'release', '--controller', controller, '--dispatch', worker.dispatch_id, '--request-id', randomUUID()]
    releaseRequests.set(worker.dispatch_id, { args, request_id: requestId(args) })
    const released = await cli(args)
    assert.equal(native(released).state, 'released')
    await settledReplay(args, released)
  }
  report.provider_events = [...sources.values()]
  await save('before-business-review.json', report)
  process.stdout.write(JSON.stringify({ state: 'ACTUAL_BUSINESS_REVIEW_REQUIRED', result: join(output, 'actual-result.json'), evidence: join(output, 'before-business-review.json'), input: '当前真实Reviewer/Agent读取成果后提供一行JSON：reviewed_result绝对路径、checked_result实际审阅说明；此输入不取得human身份。' }) + '\n')
  const lines = createInterface({ input: process.stdin, crlfDelay: Infinity, terminal: false })
  let answer
  try {
    answer = await Promise.race([new Promise((resolveLine, reject) => { lines.once('line', resolveLine); lines.once('close', () => reject(new NotRun('actual_business_review_not_supplied'))) }), delay(600_000, undefined, { ref: false }).then(() => { throw new NotRun('actual_business_review_deadline') })])
  } finally { lines.close() }
  const review = object(JSON.parse(answer))
  assert.equal(review.reviewed_result, join(output, 'actual-result.json'))
  assert.ok(typeof review.checked_result === 'string' && review.checked_result.trim().length >= 40)
  const checkedFile = join(inputDirectory, 'checked-result-' + randomUUID() + '.json')
  const checkedBody = JSON.stringify({ reviewed_result: review.reviewed_result, checked_result: review.checked_result, reviewed_by: 'current_actual_agent_controller' }) + '\n'
  await writeFile(checkedFile, checkedBody, { flag: 'wx', mode: 0o600 })
  approvedFiles.set(checkedFile, checkedBody)
  await cli(['task', 'complete', '--controller', controller, '--task', rootTask, '--result-file', checkedFile, '--request-id', randomUUID()])
  const tasks = native(await cli(['task', 'show', '--controller', controller, '--task', rootTask]))
  assert.equal(tasks.tasks.find(task => task.id === rootTask)?.status, 'completed')
  await cli(['controller', 'close', '--controller', controller, '--request-id', randomUUID()])
  report.business_review = review
  report.status = 'EXECUTED_ASSERTIONS_COMPLETED'
  report.model_execution = 'ACTUAL_NATIVE_EVENTS_OBSERVED'
  report.scenarios = { cli_multiworker: 'REQUIRES_INDEPENDENT_REVIEW', input_chunk_observation: piLinux ? 'REQUIRES_INDEPENDENT_REVIEW_OF_ACTUAL_PI_CHUNKS' : 'NOT_RUN_WITHOUT_PER_CHUNK_NATIVE_OBSERVATION', lost_effect_fault: 'INCONCLUSIVE', expired_or_cross_project_permission: 'NOT_RUN', human_cross_creator: 'NOT_RUN', gui_consistency: 'NOT_RUN', service_shutdown: 'REQUIRES_ACTUAL_SERVICE_OWNER_EVIDENCE' }
  completed = true
}

try {
  await run()
} catch (error) {
  report.status = error instanceof NotRun ? 'NOT_RUN' : 'FAIL'
  report.reason = error instanceof NotRun ? error.message : 'actual_check_failed_preserve_original_evidence'
  process.exitCode = error instanceof NotRun ? 3 : 1
  for (const command of report.commands) if (command.mutation && !['not-submitted', 'completed-response-observed', 'refused-response-observed'].includes(command.state) && !report.unknown_requests.some(request => request.request_id === command.request_id)) report.unknown_requests.push({ request_id: command.request_id, args: command.args, original_command: command.ordinal })
  // 修改结果未知时只查询原请求，不复发原修改或替换Worker。
  if (swa && token && output) for (const request of [...report.unknown_requests]) {
    try { await cli(['request', 'show', ...(controller ? ['--controller', controller] : ['--project', values['project-id']]), '--request-id', request.request_id]) }
    catch { report.checks.push({ name: 'original_request_query', status: 'UNKNOWN', request_id: request.request_id }) }
  }
  for (const worker of report.started_workers) if (!stopped.has(worker.dispatch_id)) {
    const pending = releaseRequests.get(worker.dispatch_id) ?? stopRequests.get(worker.dispatch_id)
    if (pending) {
      try { await cli(['request', 'show', '--controller', controller, '--request-id', pending.request_id]) }
      catch { /* 原请求仍保留，不能用新键重新停止。 */ }
      report.checks.push({ name: 'owned_worker_cleanup_original_stop_or_release_only', status: 'UNKNOWN', dispatch_id: worker.dispatch_id, request_id: pending.request_id })
      continue
    }
    try { worker.source ??= object((await queryWorker(worker.dispatch_id)).source); await stopWorker(worker) }
    catch { report.checks.push({ name: 'owned_worker_cleanup', status: 'UNKNOWN', dispatch_id: worker.dispatch_id }) }
  }
} finally {
  report.completed_at = new Date().toISOString()
  if (output) await save(completed ? 'driver-completed.json' : 'driver-incomplete.json', report)
  process.stdout.write(JSON.stringify({ status: report.status, reason: report.reason, evidence: output ?? null, acceptance: report.acceptance }) + '\n')
}
