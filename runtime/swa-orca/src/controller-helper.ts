import { createInterface } from 'node:readline'
import { z } from 'zod'
import { RuntimeClient } from '../../../third_party/orca-v1.4.230/src/cli/runtime/client'
import { helperInput, nativeObject, nativeError, type Invocation } from './contracts'

const controller = process.env.ORCA_TERMINAL_HANDLE
const data = process.env.ORCA_USER_DATA
const executable = process.env.SWA_EXECUTABLE
if (!controller || !data || !executable) throw new Error('native_controller_environment_missing')
const client = new RuntimeClient(data, 120_000, null, null, executable)
const call = async (method: string, params?: unknown, requestId?: string) =>
  (await client.call(method, params, requestId ? { orchestrationRequestId: requestId, timeoutMs: 120_000 } : undefined)).result

async function execute(action: string, requestId: string, invocation?: Invocation): Promise<unknown> {
  if (action === 'observe') return { current: await call('orchestration.runCurrent', { from: controller }) }
  if (!invocation || invocation.command.controller !== controller) throw new Error('controller_scope_mismatch')
  const c = invocation.command
  const run = invocation.scope.run_id
  if (action === 'task.create') {
    if (!invocation.spec || !invocation.context) throw new Error('fixed_task_context_missing')
    const created = nativeObject(await call('orchestration.runCreate', { from: controller, objective: invocation.spec.goal }, `${requestId}/run-create`))
    const nativeRun = nativeObject(created.run)
    const root = nativeObject(await call('orchestration.taskCreate', {
      spec: JSON.stringify({ spec: invocation.spec, context: invocation.context, project_id: invocation.project.id, settings_revision: invocation.caller.settings_revision }),
      taskTitle: invocation.spec.goal, callerTerminalHandle: controller, run: nativeRun.id
    }, `${requestId}/root-task`))
    return { ...created, ...root }
  }
  if (!run) throw new Error('native_run_missing')
  switch (action) {
    case 'task.list':
    case 'task.show': return call('orchestration.taskList', { run, callerTerminalHandle: controller })
    case 'task.result': return collectTaskResult(invocation)
    case 'task.complete': return call('orchestration.taskUpdate', { id: c.task_id, status: 'completed', result: invocation.input?.text, run, callerTerminalHandle: controller }, `${requestId}/native`)
    case 'worker.start': {
      if (!invocation.input || c.agent !== 'pi' || c.model !== 'openai-codex/gpt-6.1-sol') throw new Error('fixed_worker_input_required')
      return call('orchestration.workerStart', {
        from: controller, run, parent: invocation.scope.root_task_id, worktree: 'current', name: c.name,
        taskTitle: c.name, agent: 'pi', model: c.model, effort: 'xhigh', timeoutMs: 90_000,
        spec: `${invocation.input.text}\n\n固定公开上下文：\n${JSON.stringify(invocation.scope.context)}`
      }, `${requestId}/native`)
    }
    case 'worker.show': return call('orchestration.workerShow', { dispatch: c.dispatch_id })
    case 'worker.list': return call('orchestration.workerList', { run, paginate: true, limit: 100 })
    case 'worker.read': return call('orchestration.workerRead', { dispatch: c.dispatch_id, source: 'transcript', limit: c.limit, ...(c.cursor ? { cursor: c.cursor } : {}) })
    case 'worker.stop': return call('orchestration.workerStop', { dispatch: c.dispatch_id }, `${requestId}/native`)
    case 'worker.release': return call('orchestration.workerRelease', { dispatch: c.dispatch_id }, `${requestId}/native`)
    case 'worker.send': {
      const observed = nativeObject(await call('orchestration.workerShow', { dispatch: c.dispatch_id }))
      const dispatch = nativeObject(observed.dispatch)
      if (!invocation.input || typeof dispatch.assigneeHandle !== 'string') throw new Error('exact_worker_recipient_missing')
      return call('orchestration.send', { from: controller, to: dispatch.assigneeHandle, subject: 'SWA approved follow-up', body: invocation.input.text, type: 'message', run }, `${requestId}/native`)
    }
    default: throw new Error('unsupported_controller_action')
  }
}

async function collectTaskResult(invocation: Invocation) {
  const run = invocation.scope.run_id, root = invocation.scope.root_task_id
  const tasks = nativeObject(await call('orchestration.taskList', { run, callerTerminalHandle: controller }))
  const rows = z.array(z.looseObject({ id: z.string(), parent_id: z.string().nullable() })).parse(tasks.tasks)
  if (!rows.some(task => task.id === root)) throw new Error('The exact native root task is unavailable')
  const children = new Set(rows.filter(task => task.parent_id === root).map(task => task.id))
  const fleet = nativeObject(await call('orchestration.workerList', { run, paginate: true, limit: 100 }))
  const page = z.object({ hasMore: z.boolean() }).parse(fleet.page)
  if (page.hasMore) throw new Error('The fixed task worker inventory is unexpectedly incomplete')
  const workers = z.array(z.looseObject({ dispatchId: z.string(), taskId: z.string(), runId: z.string() })).parse(fleet.workers)
  const outputs = []
  for (const worker of workers) {
    if (worker.runId !== run || !children.has(worker.taskId)) throw new Error('Native output belongs to another task')
    const observation = await call('orchestration.workerShow', { dispatch: worker.dispatchId })
    outputs.push({ worker, observation })
  }
  return { run: await call('orchestration.runShow', { run }), tasks, workers: outputs }
}

async function main() {
  const lines = createInterface({ input: process.stdin, terminal: false, crlfDelay: Infinity })
  for await (const line of lines) {
    const command = helperInput.parse(JSON.parse(line))
    try {
      const result = await execute(command.action, command.request_id, command.invocation)
      process.stdout.write(JSON.stringify({ type: 'swa_controller_reply', request_id: command.request_id, controller, pid: process.pid, result }) + '\n')
    } catch (error) {
      process.stdout.write(JSON.stringify({ type: 'swa_controller_reply', request_id: command.request_id, controller, pid: process.pid, error: nativeError(error) }) + '\n')
    }
  }
}
main().catch(error => { process.stderr.write(JSON.stringify(nativeError(error)) + '\n'); process.exitCode = 1 })
