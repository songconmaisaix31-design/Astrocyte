import assert from 'node:assert/strict'
import { mkdir, mkdtemp, writeFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { test } from 'node:test'

const root = fileURLToPath(new URL('../../', import.meta.url))
const { buildSync } = createRequire(join(root, 'runtime/swa-orca/package.json'))('esbuild')
await mkdir(join(root, '.verification/run-view'), { recursive: true })
const directory = await mkdtemp(join(root, '.verification/run-view/case-'))
const bundle = join(directory, 'native-run-view.cjs')
const built = buildSync({ stdin: { contents: "export { OrchestrationDb } from './third_party/orca-v1.4.230/src/main/runtime/orchestration/db'; export { readRunView } from './runtime/swa-orca/src/run-view'", resolveDir: root }, platform: 'node', format: 'cjs', bundle: true, write: false, nodePaths: [join(root, 'runtime/swa-orca/node_modules')] })
await writeFile(bundle, built.outputFiles[0].contents)
const { OrchestrationDb, readRunView } = createRequire(bundle)(bundle)

function createCase(name, maximum = 120) {
  const path = join(directory, `${name}.sqlite`)
  const db = new OrchestrationDb(path)
  const run = db.createRun({ objective: 'Observe native task records', coordinatorHandle: null, coordinatorPaneKey: null })
  const task = db.createTask({ runId: run.id, spec: JSON.stringify({ project_id: name, spec: { max_workers: maximum } }) })
  return { db, path, root: task, scope: { project_id: name, controller: 'database-observation', run_id: run.id, root_task_id: task.id } }
}
function child(value, parent = value.root.id, deps = []) {
  return value.db.createTask({ runId: value.scope.run_id, parentId: parent, deps, spec: 'Read the public task record', taskTitle: 'Database observation' })
}

test('one native transaction reads every descendant and preserves strict failures', () => {
  const value = createCase('recursive')
  try {
    const lead = child(value), grandchild = child(value, lead.id)
    value.db.updateTaskStatus(lead.id, 'completed', 'Lead native task result')
    value.db.updateTaskStatus(value.root.id, 'completed', 'Root native task result')
    let view = readRunView(value.db, value.scope)
    assert.equal(view.complete, false)
    assert.equal(view.tasks.length, 3)
    assert.equal(view.tasks.find(row => row.id === grandchild.id).parent_id, lead.id)
    value.db.updateTaskStatus(grandchild.id, 'failed', 'Nested task failed')
    view = readRunView(value.db, value.scope)
    assert.equal(view.complete, true)
    assert.equal(view.status, 'completed')
    assert.deepEqual(view.failures, [{ task_id: grandchild.id, reason: 'native_task_failed' }])
    assert.equal(view.deliverables.length, 3)
    assert.equal(value.db.db.isTransaction, false)
  } finally { value.db.close() }
})

test('native pending questions remain visible without consuming their mailbox', () => {
  const value = createCase('questions')
  try {
    const task = child(value)
    const dispatch = value.db.createDispatchContext({ taskId: task.id, assigneeHandle: 'database-consult', creator: { kind: 'system' }, maxDepth: 4 })
    const recorded = value.db.createQuestion({ runId: value.scope.run_id, dispatchId: dispatch.id, askerHandle: 'database-consult', question: 'Which public requirement applies?' })
    const before = value.db.getMessageById(recorded.message.id)
    const view = readRunView(value.db, value.scope)
    assert.equal(view.status, 'waiting_input')
    assert.equal(view.complete, false)
    assert.equal(view.questions[0].id, recorded.message.id)
    assert.equal(view.questions[0].task_id, task.id)
    assert.deepEqual(value.db.getMessageById(recorded.message.id), before)
  } finally { value.db.close() }
})

test('unknown worker and missing resource custody never become completed', () => {
  const value = createCase('unknown')
  try {
    const task = child(value)
    const created = value.db.createStartingWorkerDispatch({ taskId: task.id, startOptions: {}, creator: { kind: 'system' }, maxDepth: 4 })
    value.db.markWorkerStartUnknown(created.dispatch.id, 'database_observation', 'A process was never observed by this database test')
    const view = readRunView(value.db, value.scope)
    assert.equal(view.status, 'unknown')
    assert.equal(view.complete, false)
    assert.ok(view.blockers.some(row => row.reason === 'start_unknown'))
    assert.ok(view.blockers.some(row => row.reason === 'resource_inventory_unavailable'))
  } finally { value.db.close() }
})

test('a reopened native database reads more than one hundred task rows completely', () => {
  const value = createCase('inventory', 110)
  const ids = []
  try {
    for (let index = 0; index < 105; index++) ids.push(child(value).id)
    const first = readRunView(value.db, value.scope)
    assert.equal(first.tasks.length, 106)
    assert.equal(first.complete, false)
  } finally { value.db.close() }
  const reopened = new OrchestrationDb(value.path)
  try {
    const view = readRunView(reopened, value.scope)
    assert.equal(view.tasks.length, 106)
    for (const id of ids) assert.ok(view.tasks.some(task => task.id === id))
    assert.equal(view.complete, false)
  } finally { reopened.close() }
})

test('stored parent or dependency corruption is refused and the read transaction closes', () => {
  for (const variant of ['root', 'orphan', 'parent-cycle', 'dependency-cycle']) {
    const value = createCase(variant)
    try {
      const a = child(value), b = child(value, a.id)
      if (variant === 'root') value.db.createTask({ runId: value.scope.run_id, spec: 'Second root' })
      if (variant === 'orphan') value.db.db.prepare('UPDATE tasks SET parent_id=? WHERE id=?').run('missing-native-task', a.id)
      if (variant === 'parent-cycle') value.db.db.prepare('UPDATE tasks SET parent_id=? WHERE id=?').run(b.id, a.id)
      if (variant === 'dependency-cycle') {
        value.db.db.prepare('UPDATE tasks SET deps=? WHERE id=?').run(JSON.stringify([b.id]), a.id)
        value.db.db.prepare('UPDATE tasks SET deps=? WHERE id=?').run(JSON.stringify([a.id]), b.id)
      }
      assert.throws(() => readRunView(value.db, value.scope), { code: 'evidence_missing' })
      assert.equal(value.db.db.isTransaction, false)
    } finally { value.db.close() }
  }
})

test('project identity and approved capacity remain mandatory for native Run reads', () => {
  const value = createCase('bounds', 1)
  try {
    assert.throws(() => readRunView(value.db, { ...value.scope, project_id: 'another-project' }), { code: 'scope_denied' })
    assert.equal(readRunView(value.db, value.scope).complete, false)
    child(value)
    child(value)
    assert.throws(() => readRunView(value.db, value.scope), { code: 'evidence_missing' })
  } finally { value.db.close() }
})

test('the approved initial Worker count prevents premature Run completion', () => {
  const value = createCase('initial-workers', 6)
  try {
    value.db.db.prepare('UPDATE tasks SET spec=? WHERE id=?').run(JSON.stringify({ project_id: value.scope.project_id, spec: { max_workers: 6, initial_workers: 4 } }), value.root.id)
    const first = child(value)
    value.db.updateTaskStatus(first.id, 'completed', 'Stored task result')
    value.db.updateTaskStatus(value.root.id, 'completed', 'Stored root result')
    const view = readRunView(value.db, value.scope)
    assert.equal(view.complete, false)
    assert.equal(view.acceptance, 'unreviewed')
    assert.ok(view.blockers.some(row => row.reason === 'initial_workers_missing:3'))
  } finally { value.db.close() }
})

test('native handoff observations preserve original mail without acknowledging it', () => {
  const value = createCase('handoff-observation')
  try {
    const task = child(value)
    const message = value.db.insertMessage({ runId: value.scope.run_id, from: 'database-author', to: 'database-recipient', type: 'handoff', deliveryContract: 'audit_only', subject: 'Domain mailbox observation', body: 'A stored notification', payload: JSON.stringify({ schema: 'swa.handoff.v1', mode: 'notify', parent_task_id: task.id }) })
    const before = value.db.getMessageById(message.id)
    const view = readRunView(value.db, value.scope)
    assert.deepEqual(view.handoffs, [{ id: message.id, from: before.from_handle, to: before.to_handle, text: before.body, read: false, record: JSON.parse(before.payload) }])
    assert.deepEqual(value.db.getMessageById(message.id), before)
    assert.equal(before.delivery_contract, 'audit_only')
    assert.deepEqual(value.db.getUnreadMessages(before.to_handle), [])
    assert.equal(view.tasks.length, 2)
    assert.equal(view.dispatches.length, 0)
  } finally { value.db.close() }
})

test('audit-only lifecycle observations expose an unresolved blocker', () => {
  const value = createCase('lifecycle-observation')
  try {
    child(value)
    const message = value.db.insertMessage({ runId: value.scope.run_id, from: 'database-observer', to: value.scope.controller, type: 'escalation', deliveryContract: 'audit_only', subject: 'Read-only domain observation', body: 'The original operation remains unconfirmed', payload: JSON.stringify({ schema: 'swa.run-lifecycle-observation.v1' }) })
    const view = readRunView(value.db, value.scope)
    assert.equal(view.status, 'unknown')
    assert.equal(view.complete, false)
    assert.ok(view.blockers.some(row => row.kind === 'lifecycle' && row.reason === value.db.getMessageById(message.id).body))
    assert.equal(value.db.getMessageById(message.id).read, 0)
  } finally { value.db.close() }
})

test('the original atomic settlement preserves succeeded instead of abandoned bookkeeping', () => {
  const value = createCase('atomic-worker-settlement')
  try {
    // 仅验证原数据库状态机；没有 Session、进程、资源或模型执行证明。
    const abandonedTask = child(value), succeededTask = child(value)
    const first = value.db.createStartingWorkerDispatch({ taskId: abandonedTask.id, startOptions: {}, creator: { kind: 'system' }, maxDepth: 4 })
    value.db.markWorkerDispatchReady(first.dispatch.id)
    value.db.completeDispatch(first.dispatch.id)
    assert.equal(value.db.getWorkerDispatch(first.dispatch.id).state, 'abandoned')
    const second = value.db.createStartingWorkerDispatch({ taskId: succeededTask.id, startOptions: {}, creator: { kind: 'system' }, maxDepth: 4 })
    value.db.markWorkerDispatchReady(second.dispatch.id)
    const settle = () => value.db.settleWorkerReportInTransaction({ taskId: succeededTask.id, dispatchId: second.dispatch.id, outcome: 'succeeded', result: 'Storage transaction value' })
    value.db.db.exec('BEGIN IMMEDIATE')
    assert.equal(settle().action, 'settled')
    value.db.db.exec('ROLLBACK')
    assert.equal(value.db.getWorkerDispatch(second.dispatch.id).state, 'ready')
    assert.equal(value.db.getDispatchContextById(second.dispatch.id).status, 'dispatched')
    assert.equal(value.db.getTask(succeededTask.id).status, 'dispatched')
    value.db.db.exec('BEGIN IMMEDIATE')
    assert.equal(settle().action, 'settled')
    value.db.db.exec('COMMIT')
    assert.equal(value.db.getWorkerDispatch(second.dispatch.id).state, 'succeeded')
    assert.equal(value.db.getDispatchContextById(second.dispatch.id).status, 'completed')
    assert.equal(value.db.getTask(succeededTask.id).status, 'completed')
    assert.equal(value.db.getTask(succeededTask.id).result, 'Storage transaction value')
    assert.equal(value.db.getWorkerTerminalResourceByOwner(second.dispatch.id), undefined)
  } finally { value.db.close() }
})
