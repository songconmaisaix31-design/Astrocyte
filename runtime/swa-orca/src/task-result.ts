import { constants, lstatSync } from 'node:fs'
import { lstat, open, realpath, stat } from 'node:fs/promises'
import { basename, dirname, isAbsolute, join, relative } from 'node:path'
import { Readable } from 'node:stream'
import { isDeepStrictEqual } from 'node:util'
import parser from 'stream-json/jsonl/parser.js'
import { z } from 'zod'
import type { OrcaRuntimeService } from '../../../third_party/orca-v1.4.230/src/main/runtime/orca-runtime'
import { OrchestrationError } from '../../../third_party/orca-v1.4.230/src/main/runtime/orchestration/orchestration-error'
import { approvedWorkerLimit } from './task-budget'
import { sessionIdFromStructuredWorkerIncarnation } from '../../../third_party/orca-v1.4.230/src/main/runtime/structured-worker-identity'
import { getStructuredAgentSessionHost } from '../../../third_party/orca-v1.4.230/src/main/native-chat/agent-session-wire/structured-agent-session-registry'
import { agentSessionProviderHandleChainHead, agentSessionProviderHandleKey } from '../../../third_party/orca-v1.4.230/src/shared/agent-session-provider-handle'
import type { AgentSessionRecord } from '../../../third_party/orca-v1.4.230/src/shared/agent-session-record'
import Database from '../../../third_party/orca-v1.4.230/src/main/sqlite/sync-database'
import { replayJournal } from '../../../third_party/orca-v1.4.230/src/main/native-chat/agent-session-journal/journal-open'
import { renderJournalState, type JournalReducerState } from '../../../third_party/orca-v1.4.230/src/main/native-chat/agent-session-journal/journal-reducer'
import { parseAgentJournalItemKey } from '../../../third_party/orca-v1.4.230/src/shared/agent-session-journal-item-key'
import { agentSessionSendBodyFingerprint } from '../../../third_party/orca-v1.4.230/src/shared/structured-agent-session-send-mutation'
import { SendBodyStructure } from '../../../third_party/orca-v1.4.230/src/shared/rpc-contract/structured-agent-session-message-params'
import type { OrchestrationDb } from '../../../third_party/orca-v1.4.230/src/main/runtime/orchestration/db/orchestration-db'
import { iterateJournalEpochRows, readJournalSessionEpoch } from '../../../third_party/orca-v1.4.230/src/main/native-chat/agent-session-journal/journal-row-table'
import { agentSelectionSchema, nativeObject, structuredInputBindingSchema, type AgentSelection, type Scope } from './contracts'
import { readNativeOpenCodeSessionSource } from './opencode-session-result'
import { readNativeCodexSessionSource } from './codex-session-result'
import { assertNativeCodexCurrent, inspectNativeCodexPolicy, requireNativeCodexPrivatePath } from './native-codex-policy'
import { verifyPrivateCodexAccount } from './private-codex-account'
import { validateCodexExport } from './codex-session-policy'
import { nativeConfigurationSource } from './native-pi-policy'
import { codexProtectedRoots } from './agent-discovery'
import { readAgentSettings } from './agent-settings'
import { swaLinuxProcessIncarnationSchema } from '../../../third_party/orca-v1.4.230/src/main/daemon/swa-linux-process-identity'
import { swaLinuxChildMessageSchema } from '../../../third_party/orca-v1.4.230/src/main/provider-process/swa-linux-provider-process-tree'

const maximum = 12 * 1024 * 1024
const sourceWitnesses = new Map<string, { path: string; identity: string; nativeSession: string; bytes: Buffer }>()
const headerSchema = z.looseObject({ type: z.literal('session'), version: z.literal(3), id: z.uuid(), timestamp: z.iso.datetime(), cwd: z.string().min(1), parentSession: z.string().optional() })
const textContent = z.looseObject({ type: z.literal('text'), text: z.string(), textSignature: z.string().optional() })
const thinkingContent = z.looseObject({ type: z.literal('thinking'), thinking: z.string(), thinkingSignature: z.string().optional(), redacted: z.boolean().optional() })
const toolCallContent = z.looseObject({ type: z.literal('toolCall'), id: z.string().min(1), name: z.string().min(1), arguments: z.record(z.string(), z.json()), thoughtSignature: z.string().optional(), namespace: z.string().optional() })
const nativeUsage = z.looseObject({
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
    provider: z.string().min(1), model: z.string().min(1), responseModel: z.string().optional(), responseId: z.string().optional(),
    providerThinkingLevel: z.string().optional(), thinkingLevel: z.string().optional(), usage: nativeUsage,
    stopReason: z.enum(['pending', 'stop', 'length', 'toolUse', 'error', 'aborted', 'deferred']), deferred: deferred.optional(), diagnostics: z.array(diagnostic).optional(), errorMessage: z.string().optional(), rawStopReason: z.string().optional(), endTurn: z.boolean().optional(), timestamp })
])
const entryBase = {
  id: z.string().min(1), parentId: z.string().min(1).nullable(), timestamp: z.iso.datetime()
}
const entrySchema = z.discriminatedUnion('type', [
  z.looseObject({ ...entryBase, type: z.literal('message'), message: messageSchema }),
  z.looseObject({ ...entryBase, type: z.literal('thinking_level_change'), thinkingLevel: z.enum(['off', 'minimal', 'low', 'medium', 'high', 'xhigh', 'max']) }),
  z.looseObject({ ...entryBase, type: z.literal('model_change'), provider: z.string().min(1), modelId: z.string().min(1) }),
  z.looseObject({ ...entryBase, type: z.literal('usage'), kind: z.string().min(1), provider: z.string().min(1), model: z.string().min(1), usage: nativeUsage, note: z.string().optional() }),
  z.looseObject({ ...entryBase, type: z.literal('session_info'), name: z.string().optional() }),
  z.looseObject({ ...entryBase, type: z.literal('label'), targetId: z.string().min(1), label: z.string().optional() })
])
const fail = (message: string): never => { throw new OrchestrationError('evidence_missing', message) }
const equalPath = (left: string, right: string) => process.platform === 'win32' ? left.toLowerCase() === right.toLowerCase() : left === right
const sourceSignature = (record: AgentSessionRecord) =>
  JSON.stringify({ session: record.sessionId, location: record.location, account: record.accountHome, options: record.options, fence: record.lease.runtimeFence, claim: record.lease.claimStatus, owner: record.lease.ownerProcess, death: record.lease.deathEvidence, unreconciled: record.lease.unreconciled, head: agentSessionProviderHandleChainHead(record.providerHandleChain) })

function originalWorkerSelection(runtime: OrcaRuntimeService, scope: Scope, dispatchId: string, record: AgentSessionRecord) {
  const db = runtime.getOrchestrationDb()
  const rows = z.array(z.object({ receipt: z.string() })).parse(db.db.prepare(`SELECT receipt FROM mutation_receipts
    WHERE caller_fingerprint=? AND method='swa.controllerDeliver' AND state IN ('pending','completed') AND json_valid(receipt)=1
    AND json_extract(receipt, '$.action')='worker.start' AND json_extract(receipt, '$.project_id')=?
    AND json_extract(receipt, '$.run_id')=? AND json_extract(receipt, '$.root_task_id')=?
    AND (json_extract(receipt, '$.native.dispatchId')=? OR json_extract(receipt, '$.swa_input_binding.dispatch_id')=?)`).all(db.getOrCreateLocalMutationCallerFingerprint(), scope.project_id, scope.run_id, scope.root_task_id, dispatchId, dispatchId))
  if (rows.length !== 1) return fail('The original native Worker selection is unavailable or ambiguous')
  const receipt = nativeObject(JSON.parse(rows[0]!.receipt)), selection = agentSelectionSchema.parse(receipt.selection)
  if (record.provider !== selection.agent_id || record.options?.model !== selection.model) return fail('The native output Agent or model differs from its original selection')
  const account = selection.native_launch.account
  if (account.kind === 'pi-native-api-key') {
    if (!('variable' in record.accountHome) || record.accountHome.variable !== 'PI_CODING_AGENT_DIR' || record.accountHome.path !== account.private_home) return fail('The native output Pi account changed')
  } else if (account.kind === 'codex-native-api-key') {
    if (!('variable' in record.accountHome) || record.accountHome.variable !== 'CODEX_HOME' || record.accountHome.path !== join(account.private_root, record.sessionId)) return fail('The native output Codex account changed')
  } else if (!('kind' in record.accountHome) || record.accountHome.kind !== 'opencode' || record.accountHome.locator.kind !== 'private' || record.accountHome.locator.root !== join(account.private_root, record.sessionId)) return fail('The native output OpenCode account changed')
  return { selection, binding: receipt.swa_input_binding === undefined ? null : structuredInputBindingSchema.parse(receipt.swa_input_binding) }
}

const processSchema = z.union([
  swaLinuxProcessIncarnationSchema.extend({ runtime_id: z.string().min(1) }),
  z.strictObject({ pid: z.number().int().positive(), creation_time_ms: z.number().positive(), runtime_id: z.string().min(1) })
])
const boundInputBody = SendBodyStructure.extend({ from: z.unknown().optional() })
const traceRowSchema = z.strictObject({ protocol: z.literal('codex-app-server'), projection: z.literal('actual_native_app_server_frames'), session_id: z.string().min(1), runtime_fence: z.number().int().nonnegative(), runtime_epoch: z.string().min(1), sequence: z.number().int().positive(), observed_at: z.iso.datetime(), monotonic_ms: z.number().nonnegative(), event: z.enum(['custody', 'policy_ready', 'thread_opened', 'notification', 'custody_closed']), process: processSchema, native_cli: z.record(z.string(), z.unknown()), detail: z.unknown() })

async function readCodexTrace(input: { path: string; record: AgentSessionRecord; binding: z.infer<typeof structuredInputBindingSchema> | null; nativeId: string; runtimeId: string; maximum: number; requireClosed: boolean }) {
  const { record, binding } = input
  const head = agentSessionProviderHandleChainHead(record.providerHandleChain)
  if (!binding || binding.agent !== 'codex' || binding.session_id !== record.sessionId || binding.dispatch_id === '' || !head || head.mintedAtFence !== binding.runtime_fence || head.handle.nativeId !== input.nativeId || head.origin !== 'created' || head.forkedFromKey) return fail('The original Codex owner fence or provider handle is unavailable')
  requireNativeCodexPrivatePath(input.path, false)
  const source = await nativeConfigurationSource(input.path)
  const handle = await open(input.path, constants.O_RDONLY | (constants.O_NOFOLLOW ?? 0))
  let content: string
  try {
    const before = await handle.stat({ bigint: true })
    if (before.size < 1n || before.size > BigInt(Math.max(0, input.maximum))) return fail('The full provider event source exceeds the remaining result bound')
    const buffer = Buffer.alloc(Number(before.size) + 1)
    let count = 0
    while (count < buffer.length) { const value = await handle.read(buffer, count, buffer.length - count, count); if (!value.bytesRead) break; count += value.bytesRead }
    const after = await handle.stat({ bigint: true })
    const version = (info: typeof before) => `${info.dev}:${info.ino}:${info.birthtimeNs}:${info.mtimeNs}:${info.size}`
    if (count !== Number(before.size) || version(before) !== source.version || version(after) !== source.version || buffer[count - 1] !== 10) return fail('The complete provider event source changed or has an incomplete frame')
    content = new TextDecoder('utf-8', { fatal: true }).decode(buffer.subarray(0, count))
  } finally { await handle.close() }
  const rows: z.infer<typeof traceRowSchema>[] = []
  for await (const item of Readable.from([Buffer.from(content)]).pipe(parser.asStream())) { if (rows.length >= 100_000) return fail('The complete provider event inventory exceeds its bound'); rows.push(traceRowSchema.parse(item.value)) }
  const first = rows[0]
  if (!first || first.event !== 'custody') return fail('The original Codex custody observation is missing')
  for (const [index, row] of rows.entries()) {
    if (row.sequence !== index + 1 || row.session_id !== record.sessionId || row.runtime_fence !== binding.runtime_fence || row.runtime_epoch !== input.runtimeId || !isDeepStrictEqual(row.process, first.process) || !isDeepStrictEqual(row.native_cli, first.native_cli) || index && row.monotonic_ms < rows[index - 1]!.monotonic_ms) return fail('The original provider event source changed its session, owner or sequence')
    if ((row.event === 'notification' || row.event === 'thread_opened')) {
      const detail = nativeObject(row.detail)
      if (typeof detail.raw_line !== 'string' || !isDeepStrictEqual(JSON.parse(detail.raw_line), detail.frame)) return fail('The raw native provider frame differs from its projection')
    }
    if (row.event === 'custody_closed' && index !== rows.length - 1) return fail('The native provider source contains input after its complete closure')
  }
  if (rows.filter(row => row.event === 'policy_ready').length !== 1) return fail('The actual native Codex policy observation is missing or ambiguous')
  const opened = rows.filter(row => row.event === 'thread_opened')
  if (opened.length !== 1) return fail('The actual native Codex thread reply is missing or ambiguous')
  const reply = nativeObject(nativeObject(opened[0]!.detail).frame), actualThread = nativeObject(nativeObject(reply.result).thread)
  if (actualThread.id !== input.nativeId || typeof actualThread.path !== 'string' || !isAbsolute(actualThread.path)) return fail('The actual native Codex thread or rollout changed')
  const closed = rows.at(-1)?.event === 'custody_closed' ? nativeObject(rows.at(-1)!.detail) : null
  if ('platform' in first.process) {
    const observed = nativeObject(first.detail)
    if (!isDeepStrictEqual(observed.root, Object.fromEntries(Object.entries(first.process).filter(([key]) => key !== 'runtime_id'))) || observed.custodyAdmitted !== true || !isDeepStrictEqual(observed.nativeCLI, first.native_cli)) return fail('The Linux native source differs from its actual custody')
    if (closed) {
      const final = z.object({ platform: z.literal('linux'), observation: z.looseObject({ custody_id: z.string(), owner: swaLinuxProcessIncarnationSchema, root: swaLinuxProcessIncarnationSchema, nativeCLI: z.record(z.string(), z.unknown()), fullTreeProof: swaLinuxChildMessageSchema.options[3] }), acknowledgement_delivered: z.literal(true), root_exit_observed: z.literal(true), root_absent: z.literal(true) }).parse(closed)
      if (!isDeepStrictEqual(final.observation.root, observed.root) || !isDeepStrictEqual(final.observation.owner, observed.owner) || final.observation.custody_id !== observed.custody_id || !isDeepStrictEqual(final.observation.nativeCLI, first.native_cli) || !isDeepStrictEqual(final.observation.fullTreeProof.root, observed.root) || !isDeepStrictEqual(final.observation.fullTreeProof.owner, observed.owner) || !isDeepStrictEqual(final.observation.fullTreeProof.native_cli, observed.nativeState) || final.observation.fullTreeProof.custody_id !== observed.custody_id || final.observation.fullTreeProof.observations[1].monotonic_ms <= final.observation.fullTreeProof.observations[0].monotonic_ms) return fail('The Linux final source is not the original complete custody')
    }
  } else {
    const windowsProcess = first.process
    const observed = nativeObject(first.detail)
    if (observed.assigned !== true || observed.rootPid !== windowsProcess.pid || observed.rootCreationTimeMs !== windowsProcess.creation_time_ms || !isDeepStrictEqual(observed.nativeCLI, first.native_cli)) return fail('The Windows native source differs from its actual Job')
    if (closed) {
      const final = z.object({ platform: z.literal('win32'), root: z.object({ pid: z.literal(windowsProcess.pid), creationTimeMs: z.literal(windowsProcess.creation_time_ms) }), nativeCLI: z.record(z.string(), z.unknown()), observations: z.array(z.looseObject({ elapsedMs: z.number(), observation: z.looseObject({ assigned: z.literal(true), rootExited: z.literal(true), members: z.tuple([]) }) })).length(2), independentIdentityObservations: z.array(z.looseObject({ elapsedMs: z.number(), identities: z.array(z.object({ expected: z.object({ pid: z.number(), creationTimeMs: z.number() }), observed: z.unknown() })) })).length(2) }).parse(closed)
      if (!isDeepStrictEqual(final.nativeCLI, first.native_cli) || final.observations[1]!.elapsedMs <= final.observations[0]!.elapsedMs || final.independentIdentityObservations[1]!.elapsedMs <= final.independentIdentityObservations[0]!.elapsedMs) return fail('The Windows final source has incomplete independent observations')
      for (const observation of final.independentIdentityObservations) {
        if (!observation.identities.some(item => item.expected.pid === windowsProcess.pid && item.expected.creationTimeMs === windowsProcess.creation_time_ms)) return fail('The original Windows root is missing from the independent inventory')
        for (const item of observation.identities) if (item.observed !== null) {
          const actual = z.object({ pid: z.literal(item.expected.pid), creationTimeMs: z.number().positive() }).parse(item.observed)
          if (actual.creationTimeMs === item.expected.creationTimeMs) return fail('A known original Windows process remains in the complete source')
        }
      }
    }
  }
  const owner = record.lease.ownerProcess
  if (record.lease.claimStatus === 'released') {
    if (!closed || owner !== null || record.lease.runtimeFence <= binding.runtime_fence || record.lease.deathEvidence?.kind !== 'exit-observed' || record.lease.deathEvidence.ownerFence !== binding.runtime_fence) return fail('The released native source has no exact complete owner death evidence')
  } else if (record.lease.claimStatus !== 'live' || !owner || owner.pid !== first.process.pid || owner.runtime !== first.process.runtime_id || record.lease.runtimeFence !== binding.runtime_fence || 'creation_time_ms' in first.process && owner.processStartTimeMs !== first.process.creation_time_ms) return fail('The live native source differs from its current exact owner')
  if (input.requireClosed && !closed) return fail('The complete native result requires its actual full-tree closure')
  const verify = async () => { requireNativeCodexPrivatePath(input.path, false); if (!isDeepStrictEqual(source, await nativeConfigurationSource(input.path))) return fail('The provider event source changed before export') }
  await verify()
  return { process: first.process, fence: binding.runtime_fence, custody: closed ?? first.detail, thread: actualThread, source, rows, value: { format: 'codex-provider-events-ndjson', path: input.path, source_version: source.version, record_count: rows.length, content }, verify }
}

const inputReceiptRow = z.object({ request_id: z.string(), state: z.enum(['pending', 'completed']), payload_hash: z.string(), receipt: z.string(), child_caller: z.string().nullable(), child_method: z.string().nullable(), child_state: z.string().nullable(), child_payload_hash: z.string().nullable(), child_receipt: z.string().nullable() })
const inputReceiptQuery = `FROM mutation_receipts parent LEFT JOIN mutation_receipts child
  ON child.request_id=json_extract(parent.receipt,'$.child_request')
  WHERE parent.caller_fingerprint=? AND parent.method='swa.controllerDeliver' AND json_valid(parent.receipt)=1
  AND json_extract(parent.receipt,'$.project_id')=? AND json_extract(parent.receipt,'$.run_id')=?
  AND json_extract(parent.receipt,'$.root_task_id')=? AND json_extract(parent.receipt,'$.action') IN ('worker.start','worker.send')
  AND (json_extract(parent.receipt,'$.swa_input_binding.dispatch_id')=?
    OR json_extract(parent.receipt,'$.native.dispatchId')=? OR json_extract(parent.receipt,'$.invocation.command.dispatch_id')=?)`

export function readCodexInputInventory(db: OrchestrationDb, scope: Scope, dispatchId: string) {
  const parameters = [db.getOrCreateLocalMutationCallerFingerprint(), scope.project_id, scope.run_id, scope.root_task_id, dispatchId, dispatchId, dispatchId]
  const bound = z.object({ count: z.number().int(), bytes: z.number().nullable() }).parse(db.db.prepare(`SELECT COUNT(*) AS count,
    SUM(LENGTH(CAST(parent.receipt AS BLOB))+COALESCE(LENGTH(CAST(child.receipt AS BLOB)),0)) AS bytes ${inputReceiptQuery}`).get(...parameters))
  if (bound.count < 1 || bound.count > 100_000 || (bound.bytes ?? 0) > maximum) return fail('The complete original Codex input request inventory is missing or exceeds its bound')
  const rows = z.array(inputReceiptRow).parse(db.db.prepare(`SELECT parent.request_id,parent.state,parent.payload_hash,parent.receipt,
    child.caller_fingerprint AS child_caller,child.method AS child_method,child.state AS child_state,
    child.payload_hash AS child_payload_hash,child.receipt AS child_receipt ${inputReceiptQuery}
    ORDER BY parent.request_id,child.caller_fingerprint`).all(...parameters))
  if (rows.length !== bound.count || new Set(rows.map(row => row.request_id)).size !== rows.length) return fail('The original Codex input request or native child is ambiguous')
  if (rows.some(row => row.state !== 'completed')) return fail('An original Codex input request is still pending or unknown')
  return { rows, signature: JSON.stringify(rows) }
}

function codexInputBindings(inventory: ReturnType<typeof readCodexInputInventory>, input: { scope: Scope; dispatchId: string; sessionId: string; fence: number; selection: AgentSelection }) {
  const bindings: z.infer<typeof structuredInputBindingSchema>[] = []
  for (const row of inventory.rows) {
    const receipt = nativeObject(JSON.parse(row.receipt))
    if (row.state !== 'completed') return fail('An original Codex input request is still pending or unknown')
    const child = row.child_receipt === null ? null : nativeObject(JSON.parse(row.child_receipt))
    const native = receipt.native === undefined ? null : nativeObject(receipt.native)
    if (receipt.action === 'worker.send' && receipt.child_request === `${row.request_id}/native` && receipt.swa_input_binding === undefined && row.child_method === 'swa.workerSend' && row.child_state === 'completed' && native?.state === 'rejected' && native.effects_applied === false && native.native_request_id === receipt.child_request && native.request_id === row.request_id && nativeObject(native.source).kind === 'native_pre_input_admission' && nativeObject(native.source).child_request === receipt.child_request && child !== null && child.swa_input_binding === undefined && child.state === 'rejected' && child.effects_applied === false && child.native_request_id === receipt.child_request && child.request_id === row.request_id && nativeObject(child.source).kind === 'native_pre_input_admission' && nativeObject(child.source).child_request === receipt.child_request) continue
    if (!child || row.child_state !== 'completed' || row.child_method !== (receipt.action === 'worker.start' ? 'orchestration.workerStart' : 'swa.workerSend')) return fail('The original Codex input native settlement is unavailable')
    const binding = structuredInputBindingSchema.parse(receipt.swa_input_binding)
    const invocation = nativeObject(receipt.invocation), command = nativeObject(invocation.command), originalScope = nativeObject(invocation.scope)
    if (binding.parent_request !== row.request_id || binding.child_request !== `${row.request_id}/native` || binding.child_request !== receipt.child_request || binding.agent !== 'codex' || binding.dispatch_id !== input.dispatchId || binding.session_id !== input.sessionId || binding.runtime_fence !== input.fence || !isDeepStrictEqual(binding, child.swa_input_binding) || !isDeepStrictEqual(receipt.selection, input.selection) || !isDeepStrictEqual(invocation.selection, input.selection) || command.request_id !== row.request_id || command.action !== receipt.action || nativeObject(invocation.project).id !== input.scope.project_id || originalScope.run_id !== input.scope.run_id || originalScope.root_task_id !== input.scope.root_task_id) return fail('The original Codex input changed its request, operation, Dispatch or session binding')
    bindings.push(binding)
  }
  if (!bindings.length || new Set(bindings.map(binding => binding.operation_id)).size !== bindings.length) return fail('The original Codex input operations are missing or ambiguous')
  return bindings
}

function assertCodexInputCompletion(bindings: ReturnType<typeof codexInputBindings>, state: JournalReducerState, nativeId: string, native: { availability: string; final: { turn_id: string; text: string } | null }, events: Awaited<ReturnType<typeof readCodexTrace>>['rows']) {
  if ([...state.submissions.values()].some(submission => submission.dispatchState === 'pending' || submission.dispatchState === 'unknown')) return fail('The native Codex journal still has queued, pending or unknown input')
  const providerItems = new Set<string>(), submissions = []
  for (const binding of bindings) {
    const { from: _sender, ...body } = boundInputBody.parse(binding.body)
    const submission = state.submissions.get(binding.operation_id), receipt = state.receipts.get(binding.operation_id)
    if (!submission || submission.clientMessageId !== binding.operation_id || submission.fence !== binding.runtime_fence || submission.dispatchState !== 'accepted' || !submission.providerItemId || !receipt || receipt.providerItemId !== submission.providerItemId || receipt.clientMessageId !== binding.operation_id || receipt.cursor.epoch !== state.epoch || submission.payloadFingerprint !== agentSessionSendBodyFingerprint(binding.session_id, body) || !Number.isSafeInteger(submission.submittedSequence) || submission.submittedSequence! <= 0 || providerItems.has(submission.providerItemId)) return fail('The original Codex input has no exact accepted native journal receipt')
    providerItems.add(submission.providerItemId)
    const identity = parseAgentJournalItemKey(submission.providerItemId)
    const item = state.items.get(state.aliases.get(submission.providerItemId) ?? submission.providerItemId)
    if (!identity || identity.provider !== 'codex' || identity.threadId !== nativeId || !item || item.body.kind !== 'message' || item.body.role !== 'user' || item.turnScope?.kind !== 'turn' || state.itemFences.get(item.itemId) !== binding.runtime_fence) return fail('The accepted Codex input has no original native user item or turn')
    const turn = state.items.get(item.turnScope.turnItemId)
    if (!turn || turn.agentId !== undefined || turn.body.kind !== 'turn' || turn.body.turnId !== identity.turnId || state.itemFences.get(turn.itemId) !== binding.runtime_fence || turn.body.state === 'running') return fail('The accepted Codex input turn is still running or has changed identity')
    submissions.push({ sequence: submission.submittedSequence!, turn: turn.body })
  }
  submissions.sort((left, right) => left.sequence - right.sequence)
  const latest = submissions.at(-1)!
  if (submissions.length > 1 && latest.sequence === submissions.at(-2)!.sequence) return fail('The latest accepted Codex input order is ambiguous')
  if (native.availability === 'complete') {
    if (!native.final || latest.turn.turnId !== native.final.turn_id || latest.turn.state !== 'completed' || latest.turn.outcome !== 'success') return fail('The native final output does not complete the latest original input')
    const completed = events.filter(row => row.event === 'notification' && nativeObject(row.detail).method === 'turn/completed').map(row => nativeObject(nativeObject(nativeObject(row.detail).frame).params))
    if (!completed.some(params => params.threadId === nativeId && nativeObject(params.turn).id === native.final!.turn_id && nativeObject(params.turn).status === 'completed' && nativeObject(params.turn).error == null)) return fail('The final Codex turn has no exact actual provider completion notification')
  }
}

export function readRootResultInventory(db: OrchestrationDb, scope: Scope, workers: { dispatchId: string; taskId: string; runId: string }[]) {
  const startWhere = `FROM mutation_receipts WHERE caller_fingerprint=? AND method='swa.controllerDeliver'
    AND json_valid(receipt)=1 AND json_extract(receipt,'$.project_id')=? AND json_extract(receipt,'$.run_id')=?
    AND json_extract(receipt,'$.root_task_id')=? AND json_extract(receipt,'$.action')='worker.start'`
  const parameters = [db.getOrCreateLocalMutationCallerFingerprint(), scope.project_id, scope.run_id, scope.root_task_id]
  const startsBound = z.object({ count: z.number().int(), bytes: z.number().nullable(), pending: z.number().nullable() }).parse(db.db.prepare(`SELECT COUNT(*) AS count,SUM(LENGTH(CAST(receipt AS BLOB))) AS bytes,SUM(CASE WHEN state='pending' THEN 1 ELSE 0 END) AS pending ${startWhere}`).get(...parameters))
  if (startsBound.count > 100_000 || (startsBound.bytes ?? 0) > maximum) return fail('The complete root Worker start reservation inventory exceeds its bound')
  if (startsBound.pending) return fail('An accepted root Worker start still has a pending reservation')
  const starts = db.db.prepare(`SELECT request_id,state,payload_hash,receipt ${startWhere} ORDER BY request_id`).all(...parameters)
  const childBound = z.object({ count: z.number().int(), bytes: z.number().nullable() }).parse(db.db.prepare('SELECT COUNT(*) AS count,SUM(LENGTH(CAST(spec AS BLOB))) AS bytes FROM tasks WHERE run_id=? AND parent_id=?').get(scope.run_id, scope.root_task_id))
  const rootBound = z.object({ bytes: z.number() }).parse(db.db.prepare('SELECT LENGTH(CAST(spec AS BLOB)) AS bytes FROM tasks WHERE id=? AND run_id=?').get(scope.root_task_id, scope.run_id))
  if (childBound.count > 100 || (childBound.bytes ?? 0) + rootBound.bytes > maximum) return fail('The complete root task inventory exceeds the current result observation bound')
  const root = db.getTask(scope.root_task_id), limit = approvedWorkerLimit(root, scope), run = db.getRun(scope.run_id)
  if (!root || !run || root.parent_id !== null || childBound.count > limit) return fail('The complete result root or approved Worker budget changed')
  const children = z.array(z.object({ id: z.string(), run_id: z.string(), parent_id: z.string(), status: z.string(), spec: z.string() })).parse(db.db.prepare('SELECT id,run_id,parent_id,status,spec FROM tasks WHERE run_id=? AND parent_id=? ORDER BY id').all(scope.run_id, scope.root_task_id))
  if (children.length !== childBound.count || children.length !== workers.length || new Set(workers.map(worker => worker.dispatchId)).size !== workers.length || new Set(workers.map(worker => worker.taskId)).size !== workers.length || workers.some(worker => worker.runId !== scope.run_id || !children.some(child => child.id === worker.taskId))) return fail('The complete root Worker inventory differs from the result')
  const listed = db.listWorkerTerminalResources({ runId: scope.run_id, limit: 101 })
  if (listed.length > 100) return fail('The complete Worker resource inventory exceeds the current observation bound')
  const resources = listed.filter(resource => resource.parentTaskId === scope.root_task_id).sort((left, right) => left.dispatchId.localeCompare(right.dispatchId))
  if (resources.length !== workers.length || resources.some(resource => !workers.some(worker => worker.dispatchId === resource.dispatchId && worker.taskId === resource.taskId) || !resource.resource)) return fail('The exact root Worker resource inventory is incomplete')
  const identities = [...workers].sort((left, right) => left.dispatchId.localeCompare(right.dispatchId)).map(worker => {
    const actual = db.getWorkerDispatch(worker.dispatchId), dispatch = db.getDispatchContextById(worker.dispatchId)
    if (!actual || !dispatch || dispatch.run_id !== scope.run_id || dispatch.task_id !== worker.taskId) return fail('The exact root Worker Dispatch identity is unavailable')
    return { worker: actual, dispatch: { id: dispatch.id, run_id: dispatch.run_id, task_id: dispatch.task_id, status: dispatch.status, assignee_handle: dispatch.assignee_handle, assignee_orca_session_id: dispatch.assignee_orca_session_id, process_incarnation: dispatch.process_incarnation, contract_version: dispatch.contract_version } }
  })
  const signature = JSON.stringify({ root: { id: root.id, run_id: root.run_id, parent_id: root.parent_id, spec: root.spec, status: root.status }, run: { id: run.id, coordinator_handle: run.coordinator_handle, consumer_generation: run.consumer_generation }, children, starts, resources, identities })
  if (Buffer.byteLength(signature) > maximum) return fail('The complete root inventory exceeds its bounded snapshot')
  return { children, signature }
}

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

async function readPrivateJournal<T>(data: string, directory: string, read: (database: Database.Database) => T): Promise<T> {
  const path = await realpath(join(directory, 'agent-session-journal.db'))
  await exactPrivatePath(data, path)
  const database = new Database(path, { readonly: true, fileMustExist: true })
  try { return read(database) } finally { database.close() }
}

export async function readNativeInputStatus(runtime: OrcaRuntimeService, projectId: string, requestId: string) {
  const db = runtime.getOrchestrationDb()
  const row = db.getMutationReceipt(db.getOrCreateLocalMutationCallerFingerprint(), requestId)
  if (!row?.receipt || row.method !== 'swa.controllerDeliver') return fail('The original native input receipt is unavailable')
  const receipt = nativeObject(JSON.parse(row.receipt))
  if (receipt.project_id !== projectId) throw new OrchestrationError('scope_denied', 'The original input belongs to another project')
  const binding = structuredInputBindingSchema.parse(receipt.swa_input_binding)
  const children = z.array(z.object({ receipt: z.string().nullable() })).parse(db.db.prepare('SELECT receipt FROM mutation_receipts WHERE request_id=?').all(binding.child_request))
  if (children.length !== 1 || !children[0]!.receipt || !isDeepStrictEqual(nativeObject(JSON.parse(children[0]!.receipt)).swa_input_binding, receipt.swa_input_binding)) return fail('The original native child input binding is unavailable')
  const host = getStructuredAgentSessionHost()
  const data = process.env.SWA_DATA_DIR
  if (!host || !data || !isAbsolute(data)) return fail('The native input journal source is unavailable')
  return readPrivateJournal(await realpath(data), host.deps.journalDatabase.stateDirectory, journalDB => {
    const replay = replayJournal(journalDB, binding.session_id)
    if (!replay || replay.damage || replay.newer) return fail('The original input journal is missing or damaged')
    const submission = replay.state.submissions.get(binding.operation_id)
    if (submission && (submission.fence !== binding.runtime_fence || submission.clientMessageId !== binding.operation_id)) return fail('The original native input journal operation changed its identity')
    const state = !submission ? 'unknown' : submission.dispatchState === 'pending' ? submission.handoverRecorded && submission.handedOverAt === undefined ? 'queued' : 'journaled' : submission.dispatchState
    return { request_id: binding.parent_request, native_request_id: binding.child_request, dispatch_id: binding.dispatch_id, agent: binding.agent, execution_mode: binding.execution_mode, session_id: binding.session_id, operation_id: binding.operation_id, client_message_id: submission?.clientMessageId, binding: receipt.swa_input_binding, state, reason: submission ? null : 'original_operation_not_observed_in_journal', submission: submission ?? null, source: { session_id: binding.session_id, epoch: replay.state.epoch, sequence: replay.state.lastSequence } }
  })
}

export async function readNativePiSessionSource(path: string, sessionRoot: string, cwd: string, sessionId: string, stopped: boolean, selected: AgentSelection) {
  const selection = agentSelectionSchema.parse(selected)
  const target = selection.native_launch.account.model_target
  if (selection.agent_id !== 'pi' || selection.native_launch.account.provider !== selection.provider || selection.model !== `${selection.provider}/${target.model_id}`) return fail('The original native Pi selection is unavailable')
  const witnessed = sourceWitnesses.get(sessionId)
  if (witnessed && !equalPath(witnessed.path, path)) return fail('A previously observed native source changed its path')
  await exactPrivatePath(sessionRoot, path, true)
  let file
  try { file = await open(path, 'r') } catch (error) {
    if (!(error instanceof Error) || !('code' in error) || error.code !== 'ENOENT') throw error
    if (witnessed || !stopped) return fail('The required native session source is missing')
    return { availability: 'unavailable', reason: 'ENOENT', persistence: 'unknown', path, entries: null, usage: { known: false, value: null, reason: 'native_source_unavailable' } }
  }
  try {
    const before = await file.stat({ bigint: true })
    if (!before.isFile() || before.size <= 0n || before.size > BigInt(maximum)) return fail('Native session file is empty, unsupported or exceeds the declared result bound')
    const identity = `${before.dev}:${before.ino}:${before.birthtimeNs}`
    if (witnessed && witnessed.identity !== identity) return fail('A previously observed native source was replaced')
    const bytes = await file.readFile()
    if (witnessed && (bytes.length < witnessed.bytes.length || !bytes.subarray(0, witnessed.bytes.length).equals(witnessed.bytes))) return fail('A previously read native source prefix changed')
    new TextDecoder('utf-8', { fatal: true }).decode(bytes)
    if (bytes.length !== Number(before.size) || bytes.at(-1) !== 10) return fail('Native JSONL source is truncated or changed during its read')
    const stream = Readable.from([bytes]).pipe(parser.asStream())
    const rows: unknown[] = []
    for await (const item of stream) rows.push(item.value)
    const header = headerSchema.parse(rows[0])
    if (!basename(path).endsWith(`_${header.id}.jsonl`) || witnessed && witnessed.nativeSession !== header.id || !equalPath(await realpath(header.cwd), cwd)) return fail('Native header identity or cwd does not match the verified native source')
    const seen = new Set<string>()
    const entries = rows.slice(1).map(value => {
      const entry = entrySchema.parse(value)
      if (seen.has(entry.id) || entry.parentId !== null && !seen.has(entry.parentId)) return fail('Native entry identity is duplicated or its real parent is missing')
      seen.add(entry.id)
      if (entry.type === 'message') {
        const message = messageSchema.parse(entry.message)
        if (message.role === 'assistant') {
          if (message.provider !== selection.provider || message.model !== target.model_id || message.api !== target.api || message.content.some(item => item.type === 'toolCall')) return fail('The actual native response violates its approved model or disabled-tool policy')
        }
      }
      if (entry.type === 'model_change' && (entry.provider !== selection.provider || entry.modelId !== target.model_id) || entry.type === 'usage' && (entry.provider !== selection.provider || entry.model !== target.model_id)) return fail('The actual native session changed its approved provider or model')
      return entry
    })
    const after = await file.stat({ bigint: true })
    const actualPath = await realpath(path)
    const currentPath = await stat(path, { bigint: true })
    if (!equalPath(actualPath, path) || `${after.dev}:${after.ino}:${after.birthtimeNs}` !== identity || `${currentPath.dev}:${currentPath.ino}:${currentPath.birthtimeNs}` !== identity || before.size !== after.size || before.mtimeNs !== after.mtimeNs) return fail('Native source changed during its consistent read')
    const messages = entries.filter(entry => entry.type === 'message').map(entry => ({ entry_id: entry.id, message: messageSchema.parse(entry.message) }))
    const last = messages.at(-1)?.message
    const complete = last?.role === 'assistant' && last.stopReason === 'stop' && !last.errorMessage && last.content.some(item => item.type === 'text' && item.text.trim().length > 0)
    if (!complete && !stopped) return fail('The required final Worker native output is not complete')
    sourceWitnesses.set(sessionId, { path, identity, nativeSession: header.id, bytes })
    const usage = messages.flatMap(row => row.message.role === 'assistant' ? [{ entry_id: row.entry_id, value: row.message.usage }] : [])
    return { availability: complete ? 'complete' : 'incomplete', reason: complete ? null : 'no_complete_final_assistant_entry', persistence: 'observed', path, file_identity: identity, size: Number(after.size), mtime_ns: after.mtimeNs.toString(), header, entries, usage: { known: usage.length > 0, value: usage.length ? usage : null, reason: usage.length ? null : 'usage_not_present_in_actual_native_entries', source: 'actual_native_message_usage', billed_cost: { known: false, value: null, reason: 'native_cost_fields_do_not_prove_billed_amount' } } }
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
  if (!/^[a-zA-Z0-9_-]+$/.test(sessionId) || dispatch.process_incarnation !== resource.process_incarnation || record.sessionId !== sessionId || !['pi', 'opencode', 'codex'].includes(record.provider) || record.location.workspaceId !== workspaceId || record.location.wslDistro !== null) return fail('The actual native Worker source identity is unavailable')
  const head = agentSessionProviderHandleChainHead(record.providerHandleChain)
  if (!head) return { availability: 'unavailable', reason: 'native_provider_link_not_observed', worker_state: nativeWorker.state }
  if (head.handle.agent !== record.provider || head.handle.transport !== (record.provider === 'pi' ? 'jsonl-rpc' : record.provider === 'opencode' ? 'acp' : 'codex-app-server')) return fail('The actual native source link has a different provider')
  const data = process.env.SWA_DATA_DIR
  if (!data || !isAbsolute(data) || !host) return fail('The actual private source root is unavailable')
  const root = await realpath(data), workerRoot = join(root, 'workers', sessionId)
  const journal = await realpath(join(host.deps.journalDatabase.stateDirectory, 'agent-session-journal.db'))
  await exactPrivatePath(root, journal)
  await exactPrivatePath(root, workerRoot)
  const ownerProcess = record.lease.ownerProcess
  const original = originalWorkerSelection(runtime, scope, dispatchId, record), selection = original.selection
  if (record.provider === 'codex') {
    if (original.binding?.dispatch_id !== dispatchId) return fail('The native Codex source belongs to another Dispatch')
    const trace = await readCodexTrace({ path: join(workerRoot, 'provider-events.ndjson'), record, binding: original.binding, nativeId: head.handle.nativeId, runtimeId: z.string().min(1).parse(nativeWorker.runtime_epoch), maximum, requireClosed: resource.release_state === 'released' })
    return { availability: 'observed', native_session_id: sessionId, provider: record.provider, fence: trace.fence, runtime_epoch: nativeWorker.runtime_epoch, claim_status: record.lease.claimStatus, selection, process: trace.process, custody: trace.custody, provider_link: head, journal_database: journal, provider_events: trace.source.path, provider_events_source: trace.source, worker_state: nativeWorker.state, release_state: resource.release_state }
  }
  return { availability: 'observed', native_session_id: sessionId, provider: record.provider, fence: record.lease.runtimeFence, claim_status: record.lease.claimStatus,
    selection,
    process: ownerProcess ? { pid: ownerProcess.pid, creation_time_ms: ownerProcess.processStartTimeMs, runtime_id: ownerProcess.runtime ?? null } : null,
    provider_link: head, journal_database: journal, provider_events: join(workerRoot, 'provider-events.ndjson'), worker_state: nativeWorker.state, release_state: resource.release_state }
}

export async function completeTaskResult(runtime: OrcaRuntimeService, scope: Scope, workspaceId: string, metadata: unknown) {
  const summary = nativeObject(metadata)
  const workers = z.array(z.looseObject({ worker: z.looseObject({ dispatchId: z.string(), taskId: z.string(), runId: z.string() }), observation: z.unknown() })).parse(summary.workers)
  const db = runtime.getOrchestrationDb()
  const rootInventory = readRootResultInventory(db, scope, workers.map(row => row.worker))
  const children = rootInventory.children
  const host = getStructuredAgentSessionHost()
  if (!host) return fail('The actual structured session host is unavailable')
  const privateData = process.env.SWA_DATA_DIR
  if (!privateData || !isAbsolute(privateData)) return fail('The explicit private source root is missing')
  const data = await realpath(privateData)
  const outputs = []
  const checks: { sessionId: string; signature: string; epoch: string; sequence: number; path: string; identity: string | null; size: number | null; mtime: string | null; nativeVerified: boolean }[] = []
  const nativeChecks: (() => Promise<void>)[] = []
  const finalNativeChecks: (() => void)[] = []
  const exportChecks: ((text: string) => void)[] = []
  const inputInventories: { dispatchId: string; signature: string }[] = []
  let used = Buffer.byteLength(JSON.stringify(summary))
  for (const row of workers) {
    const worker = row.worker, dispatch = db.getDispatchContextById(worker.dispatchId)
    const nativeWorker = db.getWorkerDispatch(worker.dispatchId), resource = db.getWorkerTerminalResourceByOwner(worker.dispatchId)
    if (worker.runId !== scope.run_id || !dispatch || dispatch.run_id !== scope.run_id || dispatch.task_id !== worker.taskId || !children.some(task => task.id === worker.taskId) || !nativeWorker || !resource) return fail('The exact Worker output does not belong to this native Run/root Task')
    const sessionId = sessionIdFromStructuredWorkerIncarnation(resource.process_incarnation)
    if (!sessionId || !/^[a-zA-Z0-9_-]+$/.test(sessionId) || dispatch.process_incarnation !== resource.process_incarnation) return fail('The native Worker session incarnation is missing or changed')
    const record = host.deps.store.getRecord(sessionId)
    if (!record || record.sessionId !== sessionId || !['pi', 'opencode', 'codex'].includes(record.provider) || record.location.workspaceId !== workspaceId || record.location.wslDistro !== null) return fail('The provider record has a different native identity or workspace')
    const head = agentSessionProviderHandleChainHead(record.providerHandleChain)
    if (!head || head.handle.agent !== record.provider || head.handle.transport !== (record.provider === 'pi' ? 'jsonl-rpc' : record.provider === 'opencode' ? 'acp' : 'codex-app-server') || record.provider === 'pi' && !isAbsolute(head.handle.nativeId)) return fail('The verified native provider link is unavailable')
    const signature = sourceSignature(record)
    const original = originalWorkerSelection(runtime, scope, worker.dispatchId, record), selection = original.selection
    if (!host.deps.resolveWorkspacePath) return fail('The actual native workspace resolution port is unavailable')
    const cwd = await realpath(await host.deps.resolveWorkspacePath(workspaceId))
    const account = selection.native_launch.account
    let nativeSession
    let providerEvents: Awaited<ReturnType<typeof readCodexTrace>> | undefined
    let codexInputs: ReturnType<typeof codexInputBindings> | undefined
    let nativeVerified = false
    if (account.kind === 'opencode-native-api-key') {
      const native = await readNativeOpenCodeSessionSource({ selection, sessionId, nativeId: head.handle.nativeId, cwd, stopped: nativeWorker.state === 'stopped', projectId: scope.project_id })
      nativeSession = native.value
      nativeChecks.push(native.verify)
      exportChecks.push(native.validateExport)
      nativeVerified = true
    } else if (account.kind === 'codex-native-api-key') {
      if (original.binding?.dispatch_id !== worker.dispatchId) return fail('The native Codex output belongs to another Dispatch')
      const inventory = readCodexInputInventory(db, scope, worker.dispatchId)
      codexInputs = codexInputBindings(inventory, { scope, dispatchId: worker.dispatchId, sessionId, fence: original.binding.runtime_fence, selection })
      inputInventories.push({ dispatchId: worker.dispatchId, signature: inventory.signature })
      const snapshot = await inspectNativeCodexPolicy({ executable: selection.executable, version: selection.version, sourceHome: account.source_home, projectId: scope.project_id, privateRoot: account.private_root, catalogPath: account.catalog_source.path, env: process.env, protectedRoots: codexProtectedRoots(await readAgentSettings(), cwd) })
      if (!isDeepStrictEqual(snapshot.policy, selection.native_launch)) return fail('The original Codex account source changed before export')
      const privateSource = await verifyPrivateCodexAccount(snapshot, sessionId)
      providerEvents = await readCodexTrace({ path: join(data, 'workers', sessionId, 'provider-events.ndjson'), record, binding: original.binding, nativeId: head.handle.nativeId, runtimeId: z.string().min(1).parse(nativeWorker.runtime_epoch), maximum: maximum - used, requireClosed: resource.release_state === 'released' })
      used += Buffer.byteLength(JSON.stringify(providerEvents.value))
      const path = z.string().min(1).parse(providerEvents.thread.path)
      const validateExport = (text: string) => validateCodexExport(snapshot, text)
      const native = await readNativeCodexSessionSource({ path, privateRoot: join(account.private_root, sessionId), nativeId: head.handle.nativeId, cwd, model: selection.model, provider: selection.provider, stopped: nativeWorker.state === 'stopped', nativeVersion: '0.163.0', source: await nativeConfigurationSource(path), validateExport })
      nativeSession = native.value
      nativeChecks.push(native.verify, providerEvents.verify, async () => { assertNativeCodexCurrent(snapshot); if (!isDeepStrictEqual(privateSource, await verifyPrivateCodexAccount(snapshot, sessionId))) return fail('The private native Codex account changed before export') })
      const finalSources = [providerEvents.source, { path: native.value.path, version: native.value.source_version }, privateSource]
      finalNativeChecks.push(() => {
        assertNativeCodexCurrent(snapshot)
        for (const source of finalSources) {
          requireNativeCodexPrivatePath(source.path, false)
          const info = lstatSync(source.path, { bigint: true })
          if (`${info.dev}:${info.ino}:${info.birthtimeNs}:${info.mtimeNs}:${info.size}` !== source.version) return fail('The native Codex source changed before the final synchronous export')
        }
      })
      exportChecks.push(native.validateExport)
      nativeVerified = true
    } else {
      const sessionRoot = join(account.private_home, 'sessions')
      await exactPrivatePath(data, sessionRoot)
      if (!equalPath(await realpath(sessionRoot), sessionRoot)) return fail('The private native session directory changed its real identity')
      nativeSession = await readNativePiSessionSource(head.handle.nativeId, sessionRoot, cwd, sessionId, nativeWorker.state === 'stopped', selection)
    }
    if ('header' in nativeSession && nativeSession.header) {
      const parent = head.forkedFromKey ? record.providerHandleChain.find(link => agentSessionProviderHandleKey(link.handle) === head.forkedFromKey) : undefined
      if (nativeSession.header.parentSession !== undefined && (!parent || parent.handle.nativeId !== nativeSession.header.parentSession) || head.origin === 'forked' && nativeSession.header.parentSession === undefined) return fail('The actual native session fork source differs from its provider handle chain')
    }
    const { journal, rows, snapshot } = await readPrivateJournal(data, host.deps.journalDatabase.stateDirectory, journalDB => {
      const bound = z.object({ count: z.number(), bytes: z.number().nullable(), max_seq: z.number().nullable() }).parse(journalDB.prepare('SELECT COUNT(*) AS count, SUM(LENGTH(row_json)) AS bytes, MAX(seq) AS max_seq FROM journal_rows WHERE session_id=? AND epoch=?').get(sessionId, readJournalSessionEpoch(journalDB, sessionId)))
      used += Buffer.byteLength(JSON.stringify(nativeSession)) + (bound.bytes ?? 0)
      if (used > maximum || bound.count <= 0 || bound.count > 100_000) return fail('The required journal is missing or exceeds the declared result bound')
      const journal = replayJournal(journalDB, sessionId)
      if (!journal || journal.damage || journal.newer || journal.state.lastSequence !== bound.max_seq) return fail('The actual native journal is damaged, incomplete or has a newer schema')
      const rows = [...iterateJournalEpochRows(journalDB, sessionId, journal.state.epoch)]
      if (rows.length !== bound.count || readJournalSessionEpoch(journalDB, sessionId) !== journal.state.epoch) return fail('Native journal rows changed during their consistent read')
      return { journal, rows, snapshot: renderJournalState(journal.state) }
    })
    if (codexInputs && providerEvents && 'final' in nativeSession) assertCodexInputCompletion(codexInputs, journal.state, head.handle.nativeId, nativeSession, providerEvents.rows)
    const current = host.deps.store.getRecord(sessionId)
    if (!current || sourceSignature(current) !== signature) return fail('The native provider identity changed during result collection')
    checks.push({ sessionId, signature, epoch: journal.state.epoch, sequence: journal.state.lastSequence, path: nativeSession.path, identity: 'file_identity' in nativeSession ? nativeSession.file_identity ?? null : null, size: 'size' in nativeSession ? nativeSession.size ?? null : null, mtime: 'mtime_ns' in nativeSession ? nativeSession.mtime_ns ?? null : null, nativeVerified })
    outputs.push({ ...row, source: { native_session_id: sessionId, selection, location: record.location, fence: providerEvents?.fence ?? record.lease.runtimeFence, provider_link: head, owner_process: providerEvents?.process ?? (record.lease.ownerProcess ? { pid: record.lease.ownerProcess.pid, creation_time_ms: record.lease.ownerProcess.processStartTimeMs, runtime_id: record.lease.ownerProcess.runtime ?? null } : null), ...(providerEvents ? { runtime_epoch: nativeWorker.runtime_epoch, custody: providerEvents.custody } : {}) }, native_session: nativeSession, ...(providerEvents ? { provider_events: providerEvents.value } : {}), journal: { session_id: sessionId, epoch: journal.state.epoch, rows: rows.map(item => ({ epoch: item.epoch, seq: item.seq, ts: item.ts, row: JSON.parse(item.rowJson) })), snapshot }, usage: nativeSession.usage })
  }
  for (const check of checks) {
    if (check.nativeVerified) continue
    try {
      const latest = await stat(check.path, { bigint: true })
      if (check.identity === null || !equalPath(await realpath(check.path), check.path) || `${latest.dev}:${latest.ino}:${latest.birthtimeNs}` !== check.identity || latest.size !== BigInt(check.size!) || latest.mtimeNs.toString() !== check.mtime) return fail('Native source changed before the complete task result was returned')
    } catch (error) {
      if (check.identity !== null || !(error instanceof Error) || !('code' in error) || error.code !== 'ENOENT') throw error
    }
  }
  for (const verify of nativeChecks) await verify()
  for (const verify of finalNativeChecks) verify()
  const journalDB = host.deps.journalDatabase.db
  for (const check of checks) {
    const sequence = z.object({ seq: z.number().nullable() }).parse(journalDB.prepare('SELECT MAX(seq) AS seq FROM journal_rows WHERE session_id=? AND epoch=?').get(check.sessionId, check.epoch)).seq
    const unchanged = readJournalSessionEpoch(journalDB, check.sessionId) === check.epoch && sequence === check.sequence
    const current = host.deps.store.getRecord(check.sessionId)
    if (!current || sourceSignature(current) !== check.signature || !unchanged) return fail('Native result sources changed before the complete result was returned')
  }
  for (const inventory of inputInventories) if (readCodexInputInventory(db, scope, inventory.dispatchId).signature !== inventory.signature) return fail('An original Codex input request was accepted or changed before the complete result was returned')
  if (readRootResultInventory(db, scope, workers.map(row => row.worker)).signature !== rootInventory.signature) return fail('The complete root Worker inventory changed before the result was returned')
  const result = { ...summary, workers: outputs, source: 'native_journal_and_native_cli_session', collected_at: new Date().toISOString() }
  const serialized = JSON.stringify(result)
  if (Buffer.byteLength(serialized) > maximum) return fail('The full result exceeds the declared 12 MiB export bound; no partial export is returned')
  for (const validate of exportChecks) validate(serialized)
  return result
}
