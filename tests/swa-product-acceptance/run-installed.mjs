import assert from 'node:assert/strict'
import { execFileSync, spawn } from 'node:child_process'
import { randomUUID } from 'node:crypto'
import { lstat, mkdir, open, readFile, realpath, stat, writeFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { dirname, isAbsolute, join, relative, resolve } from 'node:path'
import { createInterface } from 'node:readline'
import { DatabaseSync } from 'node:sqlite'
import { Readable } from 'node:stream'
import { setTimeout as delay } from 'node:timers/promises'
import { pathToFileURL } from 'node:url'
import { parseArgs } from 'node:util'
import { publicCase, publicFiles, publicRevision, publicSource } from './public-case-definition.mjs'

const help = `真实安装物多Worker验收；不会签发Grant、启动服务或直接调用原生Agent。
使用同安装Node运行本文件，显式传入：
--candidate ABS --data-root ABS --project ABS --project-id ID --endpoint LOOPBACK
--agent ID --model ID --output ABS_NEW_DIRECTORY --service-log ABS [--workers 3]
SWA_PROJECT_TOKEN仅通过受保护环境提供。来源、ToolsOff、当前Grant或完整资源证据缺失时不发送worker.start。
当前产品只有Windows Pi/OpenCode的原生事件/成果消费者；Linux及Codex来源接入完成前为NOT_RUN。
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
  values = parseArgs({ options: Object.fromEntries(['candidate', 'data-root', 'project', 'project-id', 'endpoint', 'agent', 'model', 'output', 'service-log', 'workers'].map(key => [key, { type: 'string' }])), allowPositionals: false }).values
  values.workers ??= '3'
  if (!/^[1-9]\d*$/.test(values.workers) || !Number.isSafeInteger(Number(values.workers)) || Number(values.workers) < 3 || Number(values.workers) > 100) throw new Error()
  workers = Number(values.workers)
  for (const key of ['candidate', 'data-root', 'project', 'output', 'service-log']) if (!values[key] || !isAbsolute(values[key])) throw new Error()
  for (const key of ['project-id', 'endpoint', 'agent', 'model']) if (!values[key]?.trim()) throw new Error()
  const endpoint = new URL(values.endpoint)
  if (endpoint.protocol !== 'http:' || !['127.0.0.1', '[::1]'].includes(endpoint.hostname) || !endpoint.port || endpoint.username || endpoint.password || endpoint.pathname !== '/' || endpoint.search || endpoint.hash) throw new Error()
  values.endpoint = endpoint.origin
} catch {
  process.stderr.write('INVALID_ARGUMENTS: explicit current installation, isolated paths, loopback endpoint and workers=3..100 are required; use --help\n')
  process.exit(2)
}

process.umask(0o077)
const token = process.env.SWA_PROJECT_TOKEN
const report = { status: 'NOT_RUN', workers, agent: values.agent, model: values.model, model_execution: 'NOT_RUN', acceptance: 'REQUIRES_INDEPENDENT_REVIEW', checks: [], commands: [], unknown_requests: [], provider_events: [], started_workers: [] }
let output, candidate, dataRoot, project, inputDirectory, swa, runtimeId, controller, rootTask, runId, database, requireRuntime, processAPI, daemonOwner
let source, inputs, spec, jsonlParser, foldJournal, completed = false
const sources = new Map(), witnesses = new Map(), stopped = new Set()
const sessionOwners = new Map()
const stopRequests = new Map(), releaseRequests = new Map()
const approvedFiles = new Map()
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
const safeText = text => {
  assert.ok(!token || !text.includes(token), 'credential_in_output_refused')
  return text
}
const save = async (name, value) => {
  const text = typeof value === 'string' ? value : JSON.stringify(value, null, 2) + '\n'
  await writeFile(join(output, name), safeText(text), { flag: 'wx', mode: 0o600 })
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
  const request = object((await cli(['request', 'show', '--controller', controller, '--request-id', worker.request_id])).result)
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
  return binding
}
async function sessionRecord(workerSource) {
  const path = await ordinary(workerSource.journal_database, false)
  assert.ok(within(dataRoot, path), 'session_journal_outside_private_root')
  const before = await stat(path, { bigint: true })
  const journalSource = { path, identity: `${before.dev}:${before.ino}:${before.birthtimeNs}` }
  const journal = new DatabaseSync(path, { readOnly: true })
  try {
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
function verifySessionOwner(record, workerSource, owner, worker) {
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
    assert.ok(stopped.has(worker.dispatch_id) && worker.stop?.proof, 'released_owner_has_no_original_complete_job_proof')
    assert.equal(worker.stop.proof.root.pid, owner.process.pid)
    assert.equal(worker.stop.proof.root.creationTimeMs, owner.process.processStartTimeMs)
  }
  assert.equal(workerSource.process.runtime_id, owner.process.runtime)
  assert.equal(workerSource.process.pid, owner.process.pid)
  assert.equal(workerSource.process.creation_time_ms, owner.process.processStartTimeMs)
}
async function events(workerSource) {
  assert.equal(workerSource.availability, 'observed')
  assert.equal(workerSource.provider, values.agent)
  const worker = report.started_workers.find(item => item.source?.native_session_id === workerSource.native_session_id)
  assert.ok(worker, 'source_has_no_original_dispatch')
  await queryWorker(worker.dispatch_id)
  let owner = sessionOwners.get(workerSource.native_session_id)
  const binding = owner?.binding ?? await originalStartBinding(worker, workerSource)
  const beforeRecord = await sessionRecord(workerSource)
  if (!owner) {
    assert.equal(beforeRecord.lease.claimStatus, 'live', 'first_owner_observation_must_be_live')
    const process = object(beforeRecord.lease.ownerProcess)
    assert.match(process.runtime, /^[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}$/i, 'actual_process_runtime_incarnation_missing')
    assert.deepEqual(beforeRecord.head, workerSource.provider_link, 'provider_source_head_differs_from_actual_record')
    owner = { binding, process, head: beforeRecord.head, journal_source: beforeRecord.journal_source }
  }
  verifySessionOwner(beforeRecord, workerSource, owner, worker)
  assert.ok(Number.isSafeInteger(workerSource.process.pid) && workerSource.process.creation_time_ms > 0)
  const path = await ordinary(workerSource.provider_events, false)
  assert.ok(within(dataRoot, path), 'provider_source_outside_private_root')
  const handle = await open(path, 'r')
  let rows
  try {
    const before = await handle.stat({ bigint: true })
    assert.ok(before.size > 0n && before.size <= 64n * 1024n * 1024n, 'provider_source_bound')
    const bytes = Buffer.concat(await Array.fromAsync(handle.createReadStream({ start: 0, end: Number(before.size) - 1, autoClose: false })))
    const signature = `${before.dev}:${before.ino}:${before.birthtimeNs}`
    const old = witnesses.get(path)
    if (old) assert.ok(old.signature === signature && bytes.length >= old.bytes.length && bytes.subarray(0, old.bytes.length).equals(old.bytes), 'provider_prefix_changed')
    assert.equal(bytes.length, Number(before.size))
    assert.equal(bytes.at(-1), 10, 'provider_source_incomplete_record')
    new TextDecoder('utf-8', { fatal: true }).decode(bytes)
    rows = []
    for await (const item of Readable.from([bytes]).pipe(jsonlParser.asStream())) rows.push(item.value)
    const after = await stat(path, { bigint: true })
    assert.equal(`${after.dev}:${after.ino}:${after.birthtimeNs}`, signature)
    assert.ok(after.size >= before.size)
    const afterRecord = await sessionRecord(workerSource)
    verifySessionOwner(afterRecord, workerSource, owner, worker)
    if (beforeRecord.lease.claimStatus === afterRecord.lease.claimStatus) assert.deepEqual(afterRecord, beforeRecord, 'session_owner_changed_during_event_read')
    else assert.ok(beforeRecord.lease.claimStatus === 'live' && afterRecord.lease.claimStatus === 'released', 'unreviewed_session_owner_transition')
    witnesses.set(path, { signature, bytes })
  } finally { await handle.close() }
  if (values.agent === 'opencode') rows = rows.filter(row => row.session_id === workerSource.native_session_id && row.runtime_fence === workerSource.fence)
  for (const row of rows) {
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
const isDelta = row => row.protocol === 'acp'
  ? row.event === 'session/update' && row.detail?.kind === 'known' && ['agent_message_chunk', 'agent_thought_chunk'].includes(row.detail.notification?.update?.sessionUpdate) && row.detail.notification.update.content?.type === 'text' && !!row.detail.notification.update.content.text
  : row.record?.type === 'message_update' && ['text_delta', 'thinking_delta'].includes(row.record.assistantMessageEvent?.type) && typeof row.record.assistantMessageEvent.delta === 'string' && !!row.record.assistantMessageEvent.delta
const starts = rows => rows.filter(row => row.protocol === 'acp' ? row.event === 'session/prompt' : row.record?.type === 'agent_start')
const ends = rows => rows.filter(row => row.protocol === 'acp' ? row.event === 'session/prompt/response' : row.record?.type === 'agent_end')
async function settledReplay(args, original) {
  const status = await cli(['request', 'show', '--controller', controller, '--request-id', requestId(args)])
  assert.equal(status.result.state, 'completed', 'unknown_request_must_not_replay')
  const executionCounts = async () => Promise.all([...sources.values()].map(async ({ source: actual }) => {
    const rows = await events(actual)
    return { session: actual.native_session_id, fence: actual.fence, starts: starts(rows), ends: ends(rows) }
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
  check('completed_request_replay.native_execution_counts', { request_id: requestId(args), sessions: before.map(row => ({ session: row.session, fence: row.fence, starts: row.starts.length, ends: row.ends.length })), observation_ms: 300 })
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
  stopRequests.set(worker.dispatch_id, { args, request_id: requestId(args) })
  const log = await stat(await ordinary(values['service-log'], false), { bigint: true })
  const checkpoint = active || !worker.source_checkpoint ? { identity: `${log.dev}:${log.ino}:${log.birthtimeNs}`, size: log.size.toString() } : worker.source_checkpoint
  const reply = await cli(args)
  assert.equal(native(reply).state, 'stopped')
  const proof = await proveWindowsTree(worker.source, checkpoint)
  stopped.add(worker.dispatch_id)
  worker.stop = { request_id: requestId(args), args, reply, proof, active_stop_requested: active }
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
  if (!['pi', 'opencode'].includes(values.agent)) throw new NotRun('selected_agent_native_delta_and_complete_result_reader_not_implemented')
  if (process.platform !== 'win32') throw new NotRun('linux_native_result_and_process_source_contract_awaits_runtime_implementation')
  assert.ok(within(dataRoot, await ordinary(values['service-log'], false)), 'select_only_this_services_private_log')
  requireRuntime = createRequire(join(candidate, 'runtime', 'package.json'))
  processAPI = requireRuntime('@vscode/windows-process-tree')
  const daemon = JSON.parse(await readFile(await ordinary(join(dataRoot, 'daemon', 'daemon-v49.pid'), false), 'utf8'))
  assert.equal(daemon.swa_service_job.source_revision, revision)
  assert.ok(samePath(daemon.swa_service_job.data_root, dataRoot) && samePath(daemon.swa_service_job.install_root, candidate))
  const security = requireRuntime('@orca/windows-security')
  daemonOwner = security.readServiceJobMember(daemon.swa_service_job.name, daemon.pid)
  assert.equal(daemonOwner.creationTimeMs, daemon.swa_service_job.daemon_created_at_ms)
  assert.equal(daemonOwner.ownerSid, security.currentUserSid())
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
  assert.ok(agent && agent.configured.model === values.model)
  for (const observation of [agent.registered, agent.enabled, agent.installed, agent.configured, agent.launchable, ...['start', 'send', 'read', 'stop', 'tools'].map(key => agent.control[key])]) assert.equal(observation.status, 'available', 'native_launch_tools_off_or_control_unavailable')
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
  const beforeStarts = await stat(values['service-log'], { bigint: true })
  const sourceCheckpoint = { identity: `${beforeStarts.dev}:${beforeStarts.ino}:${beforeStarts.birthtimeNs}`, size: beforeStarts.size.toString() }
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
  let overlap
  for (const deadline = Date.now() + 180_000; Date.now() < deadline;) {
    const lists = await Promise.all(report.started_workers.map(worker => events(worker.source)))
    assert.ok(lists.every(rows => starts(rows).length === 1 && ends(rows).length === 0), 'turn_ended_before_all_worker_overlap')
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
    assert.ok(rows.some(row => isDelta(row) && Date.parse(row.observed_at) > stoppedAt), 'survivor_has_no_actual_output_after_stop')
  }
  for (const deadline = Date.now() + 600_000; Date.now() < deadline;) {
    if ((await Promise.all(survivors.map(worker => events(worker.source)))).every(rows => ends(rows).length === 1)) break
    await delay(500)
  }
  for (const worker of survivors) assert.equal(ends(await events(worker.source)).length, 1, 'survivor_first_turn_not_complete')
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
    const userTexts = worker.native_session.format === 'opencode-sqlite-v2'
      ? worker.native_session.messages.filter(message => message.type === 'user').map(message => message.data.text)
      : worker.native_session.entries.filter(entry => entry.type === 'message' && entry.message.role === 'user').map(entry => textOf(entry.message))
    const originalText = binding.body.blocks.filter(block => block.type === 'text').map(block => block.text).join('\n')
    assert.ok(originalText.includes(body) && userTexts.includes(originalText), 'actual_original_input_or_native_source_differs')
    if (original.dispatch_id === follow.dispatch_id) {
      assert.ok(userTexts.includes(followup), 'actual_followup_user_body_missing_from_native_source')
      assert.notEqual(binding.operation_id, input.operation_id, 'followup_must_have_its_own_operation')
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
  for (const worker of survivors) assert.ok((await events(worker.source)).some(row => isDelta(row) && Date.parse(row.observed_at) > mark.row.stopEvent.at), 'survivor_has_no_actual_delta_after_native_stop_mark')
  const fold = foldJournal({ sessionId: stoppedOutput.journal.session_id, epoch: stoppedOutput.journal.epoch })
  for (const row of stoppedOutput.journal.rows.filter(row => row.seq < mark.seq)) assert.equal(fold.add({ seq: row.seq, ts: row.ts, rowJson: JSON.stringify(row.row) }), true)
  const prior = fold.finish()
  assert.equal(prior.damage, null)
  assert.equal(prior.newer, null)
  assert.ok([...prior.state.items.values()].some(item => item.body.kind === 'turn' && item.body.turnId === mark.row.stopEvent.turnId && item.body.state === 'running'))
  const endedTurn = stoppedOutput.journal.snapshot.items.find(item => item.body.kind === 'turn' && item.body.turnId === mark.row.stopEvent.turnId)
  const stoppedEvents = await events(target.source)
  const actualAbort = stoppedEvents.find(row => Date.parse(row.observed_at) >= mark.row.stopEvent.at && (row.protocol === 'acp' ? row.event === 'session/prompt/response' && row.detail.response.stopReason === 'cancelled' : row.record?.type === 'message_end' && row.record.message?.role === 'assistant' && row.record.message.stopReason === 'aborted'))
  assert.ok(endedTurn && (endedTurn.body.state === 'interrupted' || endedTurn.body.outcome === 'cancellation' && actualAbort), 'active_native_turn_cancellation_not_proven')
  await save('actual-result.json', exported)
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
  report.scenarios = { cli_multiworker: 'REQUIRES_INDEPENDENT_REVIEW', input_chunk_observation: 'NOT_RUN_WITHOUT_PER_CHUNK_NATIVE_OBSERVATION', lost_effect_fault: 'INCONCLUSIVE', expired_or_cross_project_permission: 'NOT_RUN', human_cross_creator: 'NOT_RUN', gui_consistency: 'NOT_RUN', service_shutdown: 'REQUIRES_ACTUAL_SERVICE_OWNER_EVIDENCE' }
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
