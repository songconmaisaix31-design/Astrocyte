import { randomUUID } from 'node:crypto'
import { realpath } from 'node:fs/promises'
import { join } from 'node:path'
import { isDeepStrictEqual, stripVTControlCharacters } from 'node:util'
import { z } from 'zod'
import { defineMethod, type RpcContext } from '../../../third_party/orca-v1.4.230/src/main/runtime/rpc/core'
import type { OrcaRuntimeService } from '../../../third_party/orca-v1.4.230/src/main/runtime/orca-runtime'
import { OrchestrationError } from '../../../third_party/orca-v1.4.230/src/main/runtime/orchestration/orchestration-error'
import { deriveRemoteRuntimeTerminalCreateHandle } from '../../../third_party/orca-v1.4.230/src/main/runtime/remote-runtime-terminal-create-identity'
import { getRepoMainWorktreeId } from '../../../third_party/orca-v1.4.230/src/shared/worktree/id'
import { getLocalPtyProvider } from '../../../third_party/orca-v1.4.230/src/main/ipc/pty/provider/registry'
import { readWindowsProcessTableFresh } from '../../../third_party/orca-v1.4.230/src/main/windows/windows-process-table'
import { createIncrementalNdjsonFramer } from '../../../third_party/orca-v1.4.230/src/shared/main-process-ndjson-framer'
import { getPendingWorkerStartRecovery } from '../../../third_party/orca-v1.4.230/src/main/runtime/rpc/orchestration-mutation-receipt'
import { invocationSchema, helperReply, nativeObject, contextSchema, commandSchema, capabilityRequirementsSchema, agentSelectionSchema, type AgentSelection, type Invocation, type Scope } from './contracts'
import { discoverAgents, selectAgent } from './agent-discovery'
import { admitWorkerResources } from './resource-admission'
import { completeTaskResult, observeWorkerSource } from './task-result'

const openReceipt = z.object({
  project_id: z.string(), controller: z.string(), worktree_id: z.string(),
  binding: z.object({ ptyId: z.string(), processIncarnation: z.string(), generation: z.number() }),
  process: z.object({ pid: z.number().int().positive(), creationTimeMs: z.number().positive(), name: z.string() }),
  helper: z.object({ pid: z.number().int().positive(), creationTimeMs: z.number().positive(), name: z.string() })
})
type Controller = z.infer<typeof openReceipt>
const wrapper = z.object({ invocation: invocationSchema })
const lookup = z.object({ project_id: z.string().min(1), controller: z.string().min(1) })
const fail = (code: string, message: string): never => { throw new OrchestrationError(code, message) }
const authority = (runtime: OrcaRuntimeService) => runtime.getOrchestrationDb().getOrCreateLocalMutationCallerFingerprint()

function recorded(runtime: OrcaRuntimeService, method: string, controller: string) {
  const db = runtime.getOrchestrationDb()
  const row = db.db.prepare(`SELECT receipt, state FROM mutation_receipts
    WHERE caller_fingerprint=? AND method=? AND json_valid(receipt)=1 AND json_extract(receipt, '$.controller')=?
    ORDER BY created_at DESC LIMIT 1`).get(authority(runtime), method, controller)
  const parsed = z.object({ receipt: z.string(), state: z.string() }).optional().parse(row)
  return parsed ? { state: parsed.state, receipt: nativeObject(JSON.parse(parsed.receipt)) } : undefined
}

function controllerScope(runtime: OrcaRuntimeService, project: string, controller: string): { owner: Controller; scope: Scope } {
  const open = recorded(runtime, 'swa.controllerOpen', controller)
  if (!open || open.state !== 'completed') return fail('operation_unknown', 'The exact controller open receipt is unavailable')
  const owner = openReceipt.parse(open.receipt)
  if (owner.project_id !== project) return fail('scope_denied', 'The native controller belongs to another project')
  const created = recorded(runtime, 'swa.taskCreate', controller)
  const scope: Scope = { project_id: project, controller, run_id: '', root_task_id: '', authority_domain: authority(runtime), context: {} }
  if (created) {
    if (created.state !== 'completed') return fail('operation_unknown', 'The original task creation remains pending')
    const ids = z.object({ run_id: z.string(), root_task_id: z.string() }).parse(created.receipt)
    const task = runtime.getOrchestrationDb().getTask(ids.root_task_id)
    const run = runtime.getOrchestrationDb().getRun(ids.run_id)
    if (!task || !run || task.run_id !== run.id || run.coordinator_handle !== controller) return fail('scope_denied', 'The native Run/Task/controller association changed')
    const fixed = z.object({ project_id: z.string(), context: contextSchema }).parse(JSON.parse(task.spec))
    if (fixed.project_id !== project) return fail('scope_denied', 'The fixed task belongs to another project')
    Object.assign(scope, ids, { context: fixed.context })
  }
  return { owner, scope }
}

async function exactController(runtime: OrcaRuntimeService, owner: Controller) {
  const inventory = await getLocalPtyProvider().listProcesses()
  const matches = inventory.filter(row => row.id === owner.binding.ptyId && row.terminalHandle === owner.controller && row.incarnationId === owner.binding.processIncarnation && row.rootProcessId === owner.process.pid)
  if (matches.length !== 1) return fail('worker_identity_changed', 'The exact controller PTY incarnation is not present')
  const binding = runtime.getTerminalPromptRequestBinding(owner.controller)
  if (binding.ptyId !== owner.binding.ptyId || binding.processIncarnation !== owner.binding.processIncarnation) return fail('worker_identity_changed', 'The controller binding changed')
  const rows = await readWindowsProcessTableFresh()
  for (const identity of [owner.process, owner.helper]) {
    const row = rows.find(item => item.pid === identity.pid)
    if (!row || row.creationTimeMs !== identity.creationTimeMs || row.name !== identity.name) return fail('worker_identity_changed', 'The controller process identity changed')
  }
}

async function helperCall(runtime: OrcaRuntimeService, controller: string, requestId: string, action: string, invocation?: Invocation) {
  const ptyId = runtime.getTerminalPtyIdForHandle(controller)
  if (!ptyId) return fail('operation_unknown', 'The native helper is unavailable')
  let stop = () => {}
  let timer: ReturnType<typeof setTimeout> | undefined
  let rejectReply: (error: Error) => void = () => {}
  const response = new Promise<z.infer<typeof helperReply>>((resolve, reject) => {
    rejectReply = reject
    const framer = createIncrementalNdjsonFramer(value => {
      const parsed = helperReply.safeParse(value)
      if (!parsed.success || parsed.data.request_id !== requestId || parsed.data.controller !== controller) return
      resolve(parsed.data)
    }, rejected => {
      if (rejected.kind === 'line-too-long') reject(new OrchestrationError('operation_unknown', 'Controller response exceeded the declared bound'))
    }, { maxLineBytes: 16 * 1024 * 1024 })
    stop = runtime.subscribeToTerminalData(ptyId, data => framer.feed(stripVTControlCharacters(data)))
    timer = setTimeout(() => reject(new OrchestrationError('operation_unknown', 'The controller action may have reached its exact PTY')), 120_000)
  })
  try {
    const written = await runtime.sendTerminal(controller, { text: JSON.stringify({ type: 'swa_controller_action', request_id: requestId, action, ...(invocation ? { invocation } : {}) }), enter: true }, { inputKind: 'driving', requireWriteSettlement: true })
    if (!written.accepted) rejectReply(new OrchestrationError('operation_unknown', 'The controller input settlement is unknown'))
    const reply = await response
    if (reply.error) throw new OrchestrationError(reply.error.code, reply.error.message)
    return reply
  } finally {
    clearTimeout(timer)
    stop()
  }
}

function checkpoint(context: RpcContext, receipt: unknown) {
  const mutation = context.orchestrationMutation
  if (!mutation) return fail('request_mismatch', 'A durable native mutation identity is required')
  context.runtime.getOrchestrationDb().checkpointPendingMutationReceipt({ ...mutation, receipt: JSON.stringify(receipt) })
  context.markMutationEffectPossible?.()
  return mutation.requestId
}

async function effect<T>(run: () => Promise<T>): Promise<T> {
  try { return await run() } catch (error) {
    throw new OrchestrationError('operation_unknown', error instanceof Error ? error.message : 'The native effect is unknown')
  }
}

export const SWA_METHODS = [
  defineMethod({ name: 'swa.status', permission: 'workspace', params: null, handler: (_params, { runtime }) => ({
    native: runtime.getStatus(), authority_domain: authority(runtime),
    databases: runtime.getOrchestrationDb().db.pragma('database_list'),
    agent_discovery: 'swa.capabilities', execution_validation: 'requires_actual_selected_native_agent_session'
  }) }),
  defineMethod({ name: 'swa.capabilities', permission: 'workspace', params: z.object({ project_id: z.string().min(1), requirements: capabilityRequirementsSchema }), handler: ({ requirements }, { runtime }) => discoverAgents(runtime, requirements) }),
  defineMethod({ name: 'swa.resolveSelection', permission: 'workspace', params: z.object({ project_id: z.string().min(1), command: commandSchema }), handler: ({ project_id, command }, { runtime }) => resolveSelection(runtime, project_id, command) }),
  defineMethod({ name: 'swa.scope', permission: 'workspace', params: lookup, handler: (params, { runtime }) => controllerScope(runtime, params.project_id, params.controller).scope }),
  defineMethod({ name: 'swa.controllerShow', permission: 'workspace', params: lookup, handler: async (params, { runtime }) => {
    const owned = controllerScope(runtime, params.project_id, params.controller)
    await exactController(runtime, owned.owner)
    return { ...owned, observation: await helperCall(runtime, params.controller, randomUUID(), 'observe') }
  } }),
  defineMethod({ name: 'swa.controllerOpen', permission: 'workspace', params: wrapper, handler: async ({ invocation }, context) => {
    const { runtime } = context
    if (invocation.command.action !== 'controller.open' || !invocation.command.workdir || invocation.project.id !== invocation.caller.project_id) return fail('scope_denied', 'The project open scope is invalid')
    const workdir = await realpath(invocation.command.workdir)
    checkpoint(context, { project_id: invocation.project.id, workdir, stage: 'workspace-registration' })
    return effect(async () => {
      const repo = await runtime.addRepo(workdir, 'folder')
      const workspace = getRepoMainWorktreeId(repo)
      const requestId = context.orchestrationMutation!.requestId
      const childId = `${requestId}/controller-pty`
      const controller = deriveRemoteRuntimeTerminalCreateHandle('swa-controller-local', workspace, childId)
      checkpoint(context, { project_id: invocation.project.id, controller, worktree_id: workspace, client_identity: 'swa-controller-local', client_mutation_id: childId, stage: 'pty-spawn' })
      const node = process.env.SWA_NODE_EXECUTABLE
      const entry = join(process.env.SWA_INSTALL_ROOT!, 'runtime', 'controller-helper.cjs')
      if (!node) throw new Error('installed_node_missing')
      const terminal = await runtime.dedupeTerminalCreate('swa-controller-local', `id:${workspace}`, childId, false,
        (canonical, preAllocatedHandle) => runtime.createTerminal(canonical, {
          launchConfig: { agentCommand: `"${node}"`, agentArgs: `--max-old-space-size=1024 "${entry}"`, agentEnv: {} },
          title: 'SWA controller', presentation: 'background', focus: false, activate: false, preAllocatedHandle
        }))
      if (terminal.handle !== controller) throw new Error('native_controller_handle_mismatch')
      const observed = await helperCall(runtime, controller, requestId, 'observe')
      const binding = runtime.getTerminalPromptRequestBinding(controller)
      const listed = await getLocalPtyProvider().listProcesses()
      const pty = listed.find(row => row.id === binding.ptyId && row.terminalHandle === controller && row.incarnationId === binding.processIncarnation)
      if (!pty?.rootProcessId) throw new Error('native_controller_process_missing')
      const rows = await readWindowsProcessTableFresh()
      const root = rows.find(row => row.pid === pty.rootProcessId)
      const helper = rows.find(row => row.pid === observed.pid)
      if (!root?.creationTimeMs || !helper?.creationTimeMs) throw new Error('native_controller_creation_time_missing')
      const ancestors = new Set<number>()
      for (let item = helper; item && !ancestors.has(item.pid); item = rows.find(row => row.pid === item!.ppid)) { ancestors.add(item.pid); if (item.pid === root.pid) break }
      if (!ancestors.has(root.pid)) throw new Error('helper_not_owned_by_controller_pty')
      return { project_id: invocation.project.id, controller, worktree_id: workspace, client_identity: 'swa-controller-local', client_mutation_id: childId, binding,
        process: { pid: root.pid, creationTimeMs: root.creationTimeMs, name: root.name }, helper: { pid: helper.pid, creationTimeMs: helper.creationTimeMs, name: helper.name }, observation: observed.result }
    })
  } }),
  defineMethod({ name: 'swa.taskCreate', permission: 'workspace', params: wrapper, handler: async ({ invocation }, context) => {
    const c = invocation.command
    if (!c.controller || c.action !== 'task.create' || !invocation.context || !invocation.spec) return fail('invalid_argument', 'A fixed task and context are required')
    const owned = controllerScope(context.runtime, invocation.project.id, c.controller)
    await exactController(context.runtime, owned.owner)
    const requestId = reserve(context, invocation, { project_id: invocation.project.id, controller: c.controller, children: [`${c.request_id}/run-create`, `${c.request_id}/root-task`], stage: 'helper-input' })
    return effect(async () => {
      const reply = nativeObject((await helperCall(context.runtime, c.controller!, requestId, c.action, invocation)).result)
      const run = nativeObject(reply.run), task = nativeObject(reply.task)
      return { project_id: invocation.project.id, controller: c.controller, run_id: z.string().parse(run.id), root_task_id: z.string().parse(task.id), children: [`${requestId}/run-create`, `${requestId}/root-task`], native: reply }
    })
  } }),
  defineMethod({ name: 'swa.controllerDeliver', permission: 'workspace', params: wrapper, handler: async ({ invocation }, context) => {
    const c = invocation.command
    if (!c.controller || !['worker.start', 'worker.send', 'worker.stop', 'worker.release', 'task.complete'].includes(c.action)) return fail('invalid_argument', 'Unsupported controller mutation')
    const { scope, owner } = controllerScope(context.runtime, invocation.project.id, c.controller)
    await exactController(context.runtime, owner)
    validateTaskAction(context.runtime, invocation, scope)
    if (c.action === 'worker.start' || c.action === 'worker.send') {
      const selected = await resolveSelection(context.runtime, invocation.project.id, c)
      if (!invocation.selection || !isDeepStrictEqual(selected, invocation.selection)) return fail('request_mismatch', 'The exact selected Agent configuration changed before native execution')
    }
    if (c.action === 'worker.start') await admitWorkerResources()
    const requestId = reserve(context, invocation, { project_id: invocation.project.id, controller: c.controller, run_id: scope.run_id, root_task_id: scope.root_task_id, action: c.action, selection: invocation.selection, child_request: `${c.request_id}/native`, stage: 'helper-input' })
    return effect(async () => ({ project_id: invocation.project.id, controller: c.controller, run_id: scope.run_id, root_task_id: scope.root_task_id, action: c.action, selection: invocation.selection, child_request: `${requestId}/native`, native: (await helperCall(context.runtime, c.controller!, requestId, c.action, { ...invocation, scope })).result }))
  } }),
  defineMethod({ name: 'swa.observe', permission: 'workspace', params: wrapper, handler: async ({ invocation }, { runtime }) => {
    const c = invocation.command
    if (!c.controller || !['task.show', 'task.list', 'task.result', 'worker.show', 'worker.list', 'worker.read'].includes(c.action)) return fail('invalid_argument', 'Unsupported observation')
    const { scope, owner } = controllerScope(runtime, invocation.project.id, c.controller)
    await exactController(runtime, owner)
    validateTaskAction(runtime, invocation, scope)
    const native = (await helperCall(runtime, c.controller, randomUUID(), c.action, { ...invocation, scope })).result
    const result = c.action === 'task.result' ? await completeTaskResult(runtime, scope, owner.worktree_id, native)
      : c.action === 'worker.show' ? { ...nativeObject(native), source: await observeWorkerSource(runtime, scope, owner.worktree_id, c.dispatch_id!) } : native
    return { scope, ...(c.action === 'task.result' ? { output_dir: c.output_dir } : {}), native: result }
  } }),
  defineMethod({ name: 'swa.controllerClose', permission: 'workspace', params: wrapper, handler: async ({ invocation }, context) => {
    const c = invocation.command
    if (!c.controller || c.action !== 'controller.close') return fail('invalid_argument', 'Exact controller close required')
    const { owner, scope } = controllerScope(context.runtime, invocation.project.id, c.controller)
    await exactController(context.runtime, owner)
    reserve(context, invocation, { project_id: invocation.project.id, controller: c.controller, action: c.action, binding: owner.binding, process: owner.process, helper: owner.helper, stage: 'controller-close' })
    return effect(async () => {
      await exactController(context.runtime, owner)
      const close = await context.runtime.closeTerminal(c.controller!)
      if (!close.ptyKilled) throw new Error('controller_exit_unproven')
      const remaining = await getLocalPtyProvider().listProcesses()
      if (remaining.some(row => row.id === owner.binding.ptyId)) throw new Error('controller_pty_still_present')
      const rows = await readWindowsProcessTableFresh()
      if (rows.some(row => [owner.process, owner.helper].some(old => old.pid === row.pid && old.creationTimeMs === row.creationTimeMs))) throw new Error('controller_process_exit_unproven')
      return { project_id: invocation.project.id, controller: c.controller, binding: owner.binding, process: owner.process, helper: owner.helper, close }
    })
  } })
]

async function resolveSelection(runtime: OrcaRuntimeService, project: string, command: Invocation['command']): Promise<AgentSelection> {
  if (command.action === 'worker.start') {
    if (!command.agent || !command.model) return fail('invalid_argument', 'An explicit native Agent and model are required')
    return selectAgent(runtime, command.agent, command.model)
  }
  if (command.action !== 'worker.send' || !command.controller || !command.dispatch_id) return fail('invalid_argument', 'Only a start or an exact existing Worker input resolves a model selection')
  const { scope } = controllerScope(runtime, project, command.controller)
  validateTaskAction(runtime, { command }, scope)
  const db = runtime.getOrchestrationDb()
  const rows = z.array(z.object({ receipt: z.string() })).parse(db.db.prepare(`SELECT receipt FROM mutation_receipts
    WHERE caller_fingerprint=? AND method='swa.controllerDeliver' AND state='completed'
    AND json_valid(receipt)=1 AND json_extract(receipt, '$.project_id')=?
    AND json_extract(receipt, '$.controller')=? AND json_extract(receipt, '$.action')='worker.start'
    AND json_extract(receipt, '$.native.dispatchId')=?`).all(authority(runtime), project, command.controller, command.dispatch_id))
  if (rows.length !== 1) return fail('operation_unknown', 'The exact original Worker start selection is unavailable')
  const original = agentSelectionSchema.parse(nativeObject(JSON.parse(rows[0]!.receipt)).selection)
  const dispatch = db.getDispatchContextById(command.dispatch_id)
  const record = dispatch?.assignee_orca_session_id && runtime.openedAgentSessionRecordStore()?.getRecord(dispatch.assignee_orca_session_id)
  if (!record || record.provider !== original.agent_id || record.options?.model !== original.model) return fail('worker_identity_changed', 'The actual Worker session does not match its original Agent selection')
  const current = await selectAgent(runtime, original.agent_id, original.model)
  if (!isDeepStrictEqual(original, current)) return fail('request_mismatch', 'The original Worker Agent source changed; preserve the original selection')
  return original
}

function requireReleased(runtime: OrcaRuntimeService, scope: Scope) {
  if (!scope.run_id) return
  const resources = runtime.getOrchestrationDb().listWorkerTerminalResources({ runId: scope.run_id, limit: 100 })
  if (resources.some(row => !row.resource || row.resource.release_state !== 'released')) fail('worker_release_unconfirmed', 'All exact Worker resources must be archived and released')
}

// 准入与 checkpoint 在同一个原生事务中完成，未知请求继续占用原生 receipt。
function reserve(context: RpcContext, invocation: Invocation, receipt: Record<string, unknown>) {
  const runtime = context.runtime, db = runtime.getOrchestrationDb(), c = invocation.command
  if (!context.orchestrationMutation || context.orchestrationMutation.requestId !== c.request_id || !c.controller) fail('request_mismatch', 'The public and native request IDs must match')
  db.db.exec('BEGIN IMMEDIATE')
  try {
    const fresh = controllerScope(runtime, invocation.project.id, c.controller)
    if (recorded(runtime, 'swa.controllerClose', c.controller)) fail('controller_closing', 'The exact controller already reserved its close; inspect the original request')
    if (c.action === 'task.create') {
      if (fresh.scope.run_id || recorded(runtime, 'swa.taskCreate', c.controller)) fail('request_mismatch', 'This controller already reserved its one user task')
    } else if (c.action === 'controller.close') {
      const pending = z.object({ count: z.number().int() }).parse(db.db.prepare(`SELECT COUNT(*) AS count FROM mutation_receipts
        WHERE caller_fingerprint=? AND state='pending' AND json_valid(receipt)=1
        AND json_extract(receipt, '$.controller')=?
        AND (method='swa.taskCreate' OR (method='swa.controllerDeliver' AND json_extract(receipt, '$.action')='worker.start'))`).get(authority(runtime), c.controller))
      if (pending.count > 0) fail('operation_in_progress', 'A native task or Worker creation still owns its reservation')
      requireReleased(runtime, fresh.scope)
    } else {
      validateTaskAction(runtime, invocation, fresh.scope)
      if (c.action === 'worker.start') {
        const children = db.listTasksWithDispatch({ runId: fresh.scope.run_id }).filter(task => task.parent_id === fresh.scope.root_task_id)
        const taskIds = new Set(children.map(task => task.id))
        const pending = z.array(z.object({ receipt: z.string() })).parse(db.db.prepare(`SELECT receipt FROM mutation_receipts
          WHERE caller_fingerprint=? AND method='swa.controllerDeliver' AND state='pending'
          AND json_valid(receipt)=1 AND json_extract(receipt, '$.controller')=?
          AND json_extract(receipt, '$.action')='worker.start'`).all(authority(runtime), c.controller))
        let uncreated = 0
        for (const item of pending) {
          const held = z.object({ child_request: z.string() }).parse(JSON.parse(item.receipt))
          const child = db.getMutationReceipt(authority(runtime), held.child_request)
          const accepted = child && getPendingWorkerStartRecovery('orchestration.workerStart', child.receipt)
          const complete = child?.receipt ? z.object({ dispatchId: z.string().optional() }).parse(JSON.parse(child.receipt)).dispatchId : undefined
          const dispatchId = accepted?.dispatchId ?? complete
          const dispatch = dispatchId ? db.getDispatchContextById(dispatchId) : undefined
          if (!dispatch || !taskIds.has(dispatch.task_id)) uncreated++
        }
        if (children.length + uncreated >= 2) fail('resource_limit', 'The fixed task has reserved both Worker slots')
      }
    }
    const requestId = checkpoint(context, receipt)
    db.db.exec('COMMIT')
    return requestId
  } catch (error) {
    db.db.exec('ROLLBACK')
    throw error
  }
}

function validateTaskAction(runtime: OrcaRuntimeService, invocation: Pick<Invocation, 'command' | 'input'>, scope: Scope) {
  const c = invocation.command, db = runtime.getOrchestrationDb()
  if (!scope.run_id || c.task_id && c.task_id !== scope.root_task_id) fail('scope_denied', 'The command must name its exact root task')
  if (c.dispatch_id) {
    const dispatch = db.getDispatchContextById(c.dispatch_id)
    const task = dispatch && db.getTask(dispatch.task_id)
    if (!dispatch || dispatch.run_id !== scope.run_id || !task || task.parent_id !== scope.root_task_id) fail('scope_denied', 'The exact Worker does not belong to this root task')
    if (['worker.send', 'worker.stop', 'worker.release'].includes(c.action)) {
      const worker = db.getWorkerDispatch(c.dispatch_id)
      if (!worker || worker.runtime_epoch !== runtime.getRuntimeId()) fail('operation_unknown', 'The original Worker Job belongs to a different runtime generation; preserve its resources')
      if (c.action === 'worker.send' && worker.state !== 'ready') fail('operation_unknown', 'The exact Worker is not ready for a new approved input')
    }
  }
  const tasks = db.listTasksWithDispatch({ runId: scope.run_id })
  if (c.action === 'worker.start' && tasks.filter(task => task.parent_id === scope.root_task_id).length >= 2) fail('resource_limit', 'The fixed task permits exactly two Workers')
  if (c.action === 'task.complete') {
    requireReleased(runtime, scope)
    if (tasks.filter(task => task.parent_id === scope.root_task_id).length !== 2 || !invocation.input) fail('evidence_missing', 'The two Worker tasks and explicit result are required')
  }
}
