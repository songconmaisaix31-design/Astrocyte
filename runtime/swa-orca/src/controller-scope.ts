import { z } from 'zod'
import type { OrcaRuntimeService } from '../../../third_party/orca-v1.4.230/src/main/runtime/orca-runtime'
import { OrchestrationError } from '../../../third_party/orca-v1.4.230/src/main/runtime/orchestration/orchestration-error'
import { getLocalPtyProvider } from '../../../third_party/orca-v1.4.230/src/main/ipc/pty/provider/registry'
import { readWindowsProcessTableFresh } from '../../../third_party/orca-v1.4.230/src/main/windows/windows-process-table'
import { contextSchema, nativeObject, type Scope } from './contracts'

const openReceipt = z.object({
  project_id: z.string(), controller: z.string(), worktree_id: z.string(),
  binding: z.object({ ptyId: z.string(), processIncarnation: z.string(), generation: z.number() }),
  process: z.object({ pid: z.number().int().positive(), creationTimeMs: z.number().positive(), name: z.string() }),
  helper: z.object({ pid: z.number().int().positive(), creationTimeMs: z.number().positive(), name: z.string() }),
  caller_authority: z.object({ terminalHandle: z.string().min(1), paneKey: z.string().min(1), processIncarnation: z.string().min(1), launchTokenHash: z.string().min(1), hostScope: z.looseObject({ kind: z.literal('local') }) })
})
export type Controller = z.infer<typeof openReceipt>
const fail = (code: string, message: string): never => { throw new OrchestrationError(code, message) }
export const authority = (runtime: OrcaRuntimeService) => runtime.getOrchestrationDb().getOrCreateLocalMutationCallerFingerprint()

export function recorded(runtime: OrcaRuntimeService, method: string, controller: string) {
  const db = runtime.getOrchestrationDb()
  const row = db.db.prepare(`SELECT receipt, state FROM mutation_receipts
    WHERE caller_fingerprint=? AND method=? AND json_valid(receipt)=1 AND json_extract(receipt, '$.controller')=?
    ORDER BY created_at DESC LIMIT 1`).get(authority(runtime), method, controller)
  const parsed = z.object({ receipt: z.string(), state: z.string() }).optional().parse(row)
  return parsed ? { state: parsed.state, receipt: nativeObject(JSON.parse(parsed.receipt)) } : undefined
}

export function controllerScope(runtime: OrcaRuntimeService, project: string, controller: string): { owner: Controller; scope: Scope } {
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

export async function exactController(runtime: OrcaRuntimeService, owner: Controller) {
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
