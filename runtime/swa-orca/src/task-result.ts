import { lstat, open, realpath, stat } from 'node:fs/promises'
import { basename, dirname, isAbsolute, join, relative } from 'node:path'
import { Readable } from 'node:stream'
import parser from 'stream-json/jsonl/parser.js'
import { z } from 'zod'
import type { OrcaRuntimeService } from '../../../third_party/orca-v1.4.230/src/main/runtime/orca-runtime'
import { OrchestrationError } from '../../../third_party/orca-v1.4.230/src/main/runtime/orchestration/orchestration-error'
import { sessionIdFromStructuredWorkerIncarnation } from '../../../third_party/orca-v1.4.230/src/main/runtime/structured-worker-identity'
import { getStructuredAgentSessionHost } from '../../../third_party/orca-v1.4.230/src/main/native-chat/agent-session-wire/structured-agent-session-registry'
import { agentSessionProviderHandleChainHead } from '../../../third_party/orca-v1.4.230/src/shared/agent-session-provider-handle'
import type { AgentSessionRecord } from '../../../third_party/orca-v1.4.230/src/shared/agent-session-record'
import { replayJournal } from '../../../third_party/orca-v1.4.230/src/main/native-chat/agent-session-journal/journal-open'
import { renderJournalState } from '../../../third_party/orca-v1.4.230/src/main/native-chat/agent-session-journal/journal-reducer'
import { iterateJournalEpochRows, readJournalSessionEpoch } from '../../../third_party/orca-v1.4.230/src/main/native-chat/agent-session-journal/journal-row-table'
import { nativeObject, type Scope } from './contracts'

const maximum = 12 * 1024 * 1024
const sourceWitnesses = new Map<string, { path: string; identity: string; sdkSession: string; bytes: Buffer }>()
const headerSchema = z.looseObject({ type: z.literal('session'), version: z.literal(3), id: z.uuid(), timestamp: z.iso.datetime(), cwd: z.string().min(1), parentSession: z.never().optional() })
const textContent = z.looseObject({ type: z.literal('text'), text: z.string(), textSignature: z.string().optional() })
const thinkingContent = z.looseObject({ type: z.literal('thinking'), thinking: z.string(), thinkingSignature: z.string().optional(), redacted: z.boolean().optional() })
const toolCallContent = z.looseObject({ type: z.literal('toolCall'), id: z.string().min(1), name: z.string().min(1), arguments: z.record(z.string(), z.json()), thoughtSignature: z.string().optional(), namespace: z.string().optional() })
const sdkUsage = z.looseObject({
  input: z.number().nonnegative(), output: z.number().nonnegative(), cacheRead: z.number().nonnegative(), cacheWrite: z.number().nonnegative(),
  cacheWrite1h: z.number().nonnegative().optional(), reasoning: z.number().nonnegative().optional(), totalTokens: z.number().nonnegative(),
  cost: z.looseObject({ input: z.number().nonnegative(), output: z.number().nonnegative(), cacheRead: z.number().nonnegative(), cacheWrite: z.number().nonnegative(), total: z.number().nonnegative() })
})
const timestamp = z.number().nonnegative()
const diagnostic = z.looseObject({ type: z.string(), timestamp, error: z.looseObject({ name: z.string().optional(), message: z.string(), stack: z.string().optional(), code: z.union([z.string(), z.number()]).optional() }).optional(), details: z.record(z.string(), z.json()).optional() })
const deferred = z.looseObject({ provider: z.string(), modelId: z.string(), api: z.string(), id: z.string(), expiresAt: z.number().optional(), pollAfterMs: z.number().optional(), data: z.json().optional() })
const messageSchema = z.discriminatedUnion('role', [
  z.looseObject({ role: z.literal('user'), content: z.union([z.string(), z.array(textContent)]), timestamp }),
  z.looseObject({ role: z.literal('system'), content: z.union([z.string(), z.array(textContent)]), sections: z.record(z.string(), z.string().nullable()).optional(), toolsAdded: z.array(z.never()).optional(), toolsRemoved: z.array(z.never()).optional(), timestamp }),
  z.looseObject({ role: z.literal('assistant'), content: z.array(z.discriminatedUnion('type', [textContent, thinkingContent, toolCallContent])), api: z.string().min(1),
    provider: z.literal('openai-codex'), model: z.literal('gpt-6.1-sol'), responseModel: z.string().optional(), responseId: z.string().optional(),
    providerThinkingLevel: z.string().optional(), thinkingLevel: z.string().optional(), usage: sdkUsage,
    stopReason: z.enum(['pending', 'stop', 'length', 'toolUse', 'error', 'aborted', 'deferred']), deferred: deferred.optional(), diagnostics: z.array(diagnostic).optional(), errorMessage: z.string().optional(), rawStopReason: z.string().optional(), endTurn: z.boolean().optional(), timestamp })
])
const entryBase = {
  id: z.string().min(1), parentId: z.string().min(1).nullable(), timestamp: z.iso.datetime()
}
const entrySchema = z.discriminatedUnion('type', [
  z.looseObject({ ...entryBase, type: z.literal('message'), message: messageSchema }),
  z.looseObject({ ...entryBase, type: z.literal('thinking_level_change'), thinkingLevel: z.literal('xhigh') }),
  z.looseObject({ ...entryBase, type: z.literal('model_change'), provider: z.literal('openai-codex'), modelId: z.literal('gpt-6.1-sol') }),
  z.looseObject({ ...entryBase, type: z.literal('usage'), kind: z.string().min(1), provider: z.literal('openai-codex'), model: z.literal('gpt-6.1-sol'), usage: sdkUsage, note: z.string().optional() }),
  z.looseObject({ ...entryBase, type: z.literal('session_info'), name: z.string().optional() }),
  z.looseObject({ ...entryBase, type: z.literal('label'), targetId: z.string().min(1), label: z.string().optional() })
])
const fail = (message: string): never => { throw new OrchestrationError('evidence_missing', message) }
const equalPath = (left: string, right: string) => left.toLowerCase() === right.toLowerCase()
const sourceSignature = (record: AgentSessionRecord) =>
  JSON.stringify({ session: record.sessionId, location: record.location, fence: record.lease.runtimeFence, head: agentSessionProviderHandleChainHead(record.providerHandleChain) })

async function exactPrivatePath(root: string, path: string, missingLeaf = false) {
  if (!isAbsolute(path)) return fail('The native source path is not absolute')
  const rel = relative(root, path)
  if (!rel || rel.startsWith('..') || isAbsolute(rel)) return fail('The native source escaped its exact private session directory')
  for (let item = path; !equalPath(item, root); item = dirname(item)) {
    if (dirname(item) === item) return fail('The private source path has no owned ancestor')
    try { if ((await lstat(item)).isSymbolicLink()) return fail('The private source path contains a reparse point') }
    catch (error) {
      if (!missingLeaf || !equalPath(item, path) || !(error instanceof Error) || !('code' in error) || error.code !== 'ENOENT') throw error
    }
  }
  if (!equalPath(await realpath(dirname(path)), dirname(path))) return fail('The native source parent changed its real path')
}

export async function readSDKSessionSource(path: string, sessionRoot: string, cwd: string, sessionId: string, stopped: boolean) {
  const witnessed = sourceWitnesses.get(sessionId)
  if (witnessed && !equalPath(witnessed.path, path)) return fail('A previously observed SDK source changed its path')
  await exactPrivatePath(sessionRoot, path, true)
  let file
  try { file = await open(path, 'r') } catch (error) {
    if (!(error instanceof Error) || !('code' in error) || error.code !== 'ENOENT') throw error
    if (witnessed || !stopped) return fail('The required SDK session source is missing')
    return { availability: 'unavailable', reason: 'ENOENT', persistence: 'unknown', path, entries: null, usage: { known: false, value: null, reason: 'sdk_source_unavailable' } }
  }
  try {
    const before = await file.stat({ bigint: true })
    if (!before.isFile() || before.size <= 0n || before.size > BigInt(maximum)) return fail('SDK session file is empty, unsupported or exceeds the declared result bound')
    const identity = `${before.dev}:${before.ino}:${before.birthtimeNs}`
    if (witnessed && witnessed.identity !== identity) return fail('A previously observed SDK source was replaced')
    const bytes = await file.readFile()
    if (witnessed && (bytes.length < witnessed.bytes.length || !bytes.subarray(0, witnessed.bytes.length).equals(witnessed.bytes))) return fail('A previously read SDK source prefix changed')
    new TextDecoder('utf-8', { fatal: true }).decode(bytes)
    if (bytes.length !== Number(before.size) || bytes.at(-1) !== 10) return fail('SDK JSONL source is truncated or changed during its read')
    const stream = Readable.from([bytes]).pipe(parser.asStream())
    const rows: unknown[] = []
    for await (const item of stream) rows.push(item.value)
    const header = headerSchema.parse(rows[0])
    if (!basename(path).endsWith(`_${header.id}.jsonl`) || witnessed && witnessed.sdkSession !== header.id || !equalPath(await realpath(header.cwd), cwd)) return fail('SDK header identity or cwd does not match the verified native source')
    const seen = new Set<string>()
    const entries = rows.slice(1).map(value => {
      const entry = entrySchema.parse(value)
      if (seen.has(entry.id) || entry.parentId !== null && !seen.has(entry.parentId)) return fail('SDK entry identity is duplicated or its real parent is missing')
      seen.add(entry.id)
      if (entry.type === 'message') {
        const message = messageSchema.parse(entry.message)
        if (message.role === 'assistant') {
          if (message.content.some(item => item.type === 'toolCall')) return fail('A restricted SDK source unexpectedly contains tool calls')
        }
      }
      return entry
    })
    const after = await file.stat({ bigint: true })
    const actualPath = await realpath(path)
    if (!equalPath(actualPath, path) || `${after.dev}:${after.ino}:${after.birthtimeNs}` !== identity || before.size !== after.size || before.mtimeNs !== after.mtimeNs) return fail('SDK source changed during its consistent read')
    const messages = entries.filter(entry => entry.type === 'message').map(entry => ({ entry_id: entry.id, message: messageSchema.parse(entry.message) }))
    const last = messages.at(-1)?.message
    const complete = last?.role === 'assistant' && last.stopReason === 'stop' && !last.errorMessage && last.content.some(item => item.type === 'text' && item.text.trim().length > 0)
    if (!complete && !stopped) return fail('The required final Worker SDK output is not complete')
    sourceWitnesses.set(sessionId, { path, identity, sdkSession: header.id, bytes })
    const usage = messages.flatMap(row => row.message.role === 'assistant' ? [{ entry_id: row.entry_id, value: row.message.usage }] : [])
    return { availability: complete ? 'complete' : 'incomplete', reason: complete ? null : 'no_complete_final_assistant_entry', persistence: 'observed', path, file_identity: identity, size: Number(after.size), mtime_ns: after.mtimeNs.toString(), header, entries, usage: { known: usage.length > 0, value: usage.length ? usage : null, reason: usage.length ? null : 'usage_not_present_in_actual_sdk_entries', source: 'actual_sdk_message_usage', billed_cost: { known: false, value: null, reason: 'sdk_cost_fields_do_not_prove_billed_amount' } } }
  } finally { await file.close() }
}

export async function observeWorkerSource(runtime: OrcaRuntimeService, scope: Scope, workspaceId: string, dispatchId: string) {
  const db = runtime.getOrchestrationDb(), dispatch = db.getDispatchContextById(dispatchId)
  const task = dispatch && db.getTask(dispatch.task_id), resource = db.getWorkerTerminalResourceByOwner(dispatchId)
  const nativeWorker = db.getWorkerDispatch(dispatchId)
  if (!dispatch || dispatch.run_id !== scope.run_id || task?.parent_id !== scope.root_task_id) return fail('The observed source does not belong to the exact native root task')
  if (!resource || !nativeWorker) return { availability: 'unavailable', reason: 'native_resource_not_observed', worker_state: nativeWorker?.state ?? null }
  const sessionId = sessionIdFromStructuredWorkerIncarnation(resource.process_incarnation)
  const host = getStructuredAgentSessionHost(), record = sessionId && host?.deps.store.getRecord(sessionId)
  if (!sessionId || !host || !record) return { availability: 'unavailable', reason: 'native_session_record_not_observed', worker_state: nativeWorker.state }
  if (!/^[a-zA-Z0-9_-]+$/.test(sessionId) || dispatch.process_incarnation !== resource.process_incarnation || record.sessionId !== sessionId || record.provider !== 'pi' || record.location.workspaceId !== workspaceId || record.location.wslDistro !== null) return fail('The actual native Worker source identity is unavailable')
  const head = agentSessionProviderHandleChainHead(record.providerHandleChain)
  if (!head) return { availability: 'unavailable', reason: 'native_pi_link_not_observed', worker_state: nativeWorker.state }
  if (head.handle.transport !== 'jsonl-rpc' || head.handle.agent !== 'pi') return fail('The actual native Pi source link has a different provider')
  const data = process.env.SWA_DATA_DIR
  if (!data || !isAbsolute(data) || !host) return fail('The actual private source root is unavailable')
  const root = await realpath(data), workerRoot = join(root, 'workers', sessionId)
  const journal = await realpath(join(host.deps.journalDatabase.stateDirectory, 'agent-session-journal.db'))
  await exactPrivatePath(root, journal)
  await exactPrivatePath(root, workerRoot)
  const ownerProcess = record.lease.ownerProcess
  return { availability: 'observed', native_session_id: sessionId, provider: record.provider, fence: record.lease.runtimeFence, claim_status: record.lease.claimStatus,
    process: ownerProcess ? { pid: ownerProcess.pid, creation_time_ms: ownerProcess.processStartTimeMs, runtime_id: ownerProcess.runtime ?? null } : null,
    provider_link: head, journal_database: journal, provider_events: join(workerRoot, 'provider-events.ndjson'), worker_state: nativeWorker.state, release_state: resource.release_state }
}

export async function completeTaskResult(runtime: OrcaRuntimeService, scope: Scope, workspaceId: string, metadata: unknown) {
  const summary = nativeObject(metadata)
  const workers = z.array(z.looseObject({ worker: z.looseObject({ dispatchId: z.string(), taskId: z.string(), runId: z.string() }), observation: z.unknown() })).parse(summary.workers)
  if (workers.length > 2) return fail('The exact fixed task Worker inventory exceeds its approved bound')
  const db = runtime.getOrchestrationDb()
  const children = db.listTasksWithDispatch({ runId: scope.run_id }).filter(task => task.parent_id === scope.root_task_id)
  if (children.length !== workers.length || new Set(workers.map(row => row.worker.dispatchId)).size !== workers.length) return fail('Native Worker output inventory is incomplete or duplicated')
  const host = getStructuredAgentSessionHost()
  if (!host) return fail('The actual structured session host is unavailable')
  const privateData = process.env.SWA_DATA_DIR
  if (!privateData || !isAbsolute(privateData)) return fail('The explicit private source root is missing')
  const data = await realpath(privateData)
  const outputs = []
  const checks: { sessionId: string; signature: string; epoch: string; sequence: number; path: string; identity: string | null; size: number | null; mtime: string | null }[] = []
  let used = Buffer.byteLength(JSON.stringify(summary))
  for (const row of workers) {
    const worker = row.worker, dispatch = db.getDispatchContextById(worker.dispatchId)
    const nativeWorker = db.getWorkerDispatch(worker.dispatchId), resource = db.getWorkerTerminalResourceByOwner(worker.dispatchId)
    if (worker.runId !== scope.run_id || !dispatch || dispatch.run_id !== scope.run_id || dispatch.task_id !== worker.taskId || !children.some(task => task.id === worker.taskId) || !nativeWorker || !resource) return fail('The exact Worker output does not belong to this native Run/root Task')
    const sessionId = sessionIdFromStructuredWorkerIncarnation(resource.process_incarnation)
    if (!sessionId || !/^[a-zA-Z0-9_-]+$/.test(sessionId) || dispatch.process_incarnation !== resource.process_incarnation) return fail('The native Worker session incarnation is missing or changed')
    const record = host.deps.store.getRecord(sessionId)
    if (!record || record.sessionId !== sessionId || record.provider !== 'pi' || record.location.workspaceId !== workspaceId || record.location.wslDistro !== null) return fail('The provider record has a different native identity or workspace')
    const head = agentSessionProviderHandleChainHead(record.providerHandleChain)
    if (!head || head.handle.transport !== 'jsonl-rpc' || head.handle.agent !== 'pi' || !isAbsolute(head.handle.nativeId)) return fail('The verified native Pi provider link is unavailable')
    const signature = sourceSignature(record)
    if (!host.deps.resolveWorkspacePath) return fail('The actual native workspace resolution port is unavailable')
    const cwd = await realpath(await host.deps.resolveWorkspacePath(workspaceId))
    const sessionRoot = join(data, 'workers', sessionId, 'sessions')
    await exactPrivatePath(data, sessionRoot)
    if (!equalPath(await realpath(sessionRoot), sessionRoot)) return fail('The private SDK session directory changed its real identity')
    const sdk = await readSDKSessionSource(head.handle.nativeId, sessionRoot, cwd, sessionId, nativeWorker.state === 'stopped')
    const journalDB = host.deps.journalDatabase.db
    const bound = z.object({ count: z.number(), bytes: z.number().nullable(), max_seq: z.number().nullable() }).parse(journalDB.prepare('SELECT COUNT(*) AS count, SUM(LENGTH(row_json)) AS bytes, MAX(seq) AS max_seq FROM journal_rows WHERE session_id=? AND epoch=?').get(sessionId, readJournalSessionEpoch(journalDB, sessionId)))
    used += Buffer.byteLength(JSON.stringify(sdk)) + (bound.bytes ?? 0)
    if (used > maximum || bound.count <= 0 || bound.count > 100_000) return fail('The required journal is missing or exceeds the declared result bound')
    const journal = replayJournal(journalDB, sessionId)
    if (!journal || journal.damage || journal.newer || journal.state.lastSequence !== bound.max_seq) return fail('The actual native journal is damaged, incomplete or has a newer schema')
    const rows = [...iterateJournalEpochRows(journalDB, sessionId, journal.state.epoch)]
    if (rows.length !== bound.count) return fail('Native journal rows changed during their consistent read')
    const snapshot = renderJournalState(journal.state)
    const current = host.deps.store.getRecord(sessionId)
    if (!current || sourceSignature(current) !== signature || readJournalSessionEpoch(journalDB, sessionId) !== journal.state.epoch) return fail('The native provider identity or journal epoch changed during result collection')
    checks.push({ sessionId, signature, epoch: journal.state.epoch, sequence: journal.state.lastSequence, path: sdk.path, identity: 'file_identity' in sdk ? sdk.file_identity ?? null : null, size: 'size' in sdk ? sdk.size ?? null : null, mtime: 'mtime_ns' in sdk ? sdk.mtime_ns ?? null : null })
    outputs.push({ ...row, source: { native_session_id: sessionId, location: record.location, fence: record.lease.runtimeFence, provider_link: head, owner_process: record.lease.ownerProcess ? { pid: record.lease.ownerProcess.pid, creation_time_ms: record.lease.ownerProcess.processStartTimeMs, runtime_id: record.lease.ownerProcess.runtime ?? null } : null }, sdk, journal: { session_id: sessionId, epoch: journal.state.epoch, rows: rows.map(item => ({ epoch: item.epoch, seq: item.seq, ts: item.ts, row: JSON.parse(item.rowJson) })), snapshot }, usage: sdk.usage })
  }
  for (const check of checks) {
    try {
      const latest = await stat(check.path, { bigint: true })
      if (check.identity === null || !equalPath(await realpath(check.path), check.path) || `${latest.dev}:${latest.ino}:${latest.birthtimeNs}` !== check.identity || latest.size !== BigInt(check.size!) || latest.mtimeNs.toString() !== check.mtime) return fail('SDK source changed before the complete task result was returned')
    } catch (error) {
      if (check.identity !== null || !(error instanceof Error) || !('code' in error) || error.code !== 'ENOENT') throw error
    }
  }
  for (const check of checks) {
    const current = host.deps.store.getRecord(check.sessionId), journalDB = host.deps.journalDatabase.db
    const sequence = z.object({ seq: z.number().nullable() }).parse(journalDB.prepare('SELECT MAX(seq) AS seq FROM journal_rows WHERE session_id=? AND epoch=?').get(check.sessionId, check.epoch)).seq
    if (!current || sourceSignature(current) !== check.signature || readJournalSessionEpoch(journalDB, check.sessionId) !== check.epoch || sequence !== check.sequence) return fail('Native result sources changed before the complete result was returned')
  }
  const result = { ...summary, workers: outputs, source: 'native_journal_and_sdk_session', collected_at: new Date().toISOString() }
  if (Buffer.byteLength(JSON.stringify(result)) > maximum) return fail('The full result exceeds the declared 12 MiB export bound; no partial export is returned')
  return result
}
