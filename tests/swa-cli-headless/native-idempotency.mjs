import assert from 'node:assert/strict'
import { spawn, execFileSync } from 'node:child_process'
import { randomUUID } from 'node:crypto'
import { DatabaseSync } from 'node:sqlite'
import { readFile, realpath, writeFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { dirname, isAbsolute, join, relative } from 'node:path'
import { setTimeout } from 'node:timers/promises'
import { parseArgs } from 'node:util'

const { values } = parseArgs({ options: Object.fromEntries(['candidate', 'data-root', 'project', 'project-id', 'endpoint', 'output'].map(key => [key, { type: 'string' }])), allowPositionals: false })
for (const key of ['candidate', 'data-root', 'project', 'output']) assert.ok(values[key] && isAbsolute(values[key]))
assert.ok(values['project-id'] && values.endpoint && process.env.SWA_PROJECT_TOKEN)
const candidate = await realpath(values.candidate), dataRoot = await realpath(values['data-root']), project = await realpath(values.project)
const node = join(candidate, 'node', 'node.exe'), swa = join(candidate, 'swa.exe'), evidence = dirname(values.output)
assert.equal((await realpath(process.execPath)).toLowerCase(), node.toLowerCase())
const revision = (await readFile(join(candidate, 'VERSION'), 'utf8')).trim()
assert.equal(execFileSync(swa, ['--version'], { encoding: 'utf8', windowsHide: true }).trim(), revision)
const input = join(project, '.verification', 'swa-cli-input'), source = JSON.parse(await readFile(join(input, 'source.json'), 'utf8'))
assert.equal(source.revision, 'da3bdddb085808863090a05f60cc1c64415039cf')
assert.equal((await realpath(source.project_root)).toLowerCase(), project.toLowerCase())
const environment = { ...process.env, SWA_ENDPOINT: values.endpoint, SWA_PROJECT_ID: values['project-id'], NODE_OPTIONS: '--max-old-space-size=1024' }
const requireRuntime = createRequire(join(candidate, 'runtime', 'package.json'))
const processAPI = requireRuntime('@vscode/windows-process-tree')
const commands = [], checks = []
let ordinal = 0, runtimeId, database

async function cli(args, allowed = [0]) {
  const index = ++ordinal, started = new Date().toISOString()
  const child = spawn(swa, [...args, '--json'], { cwd: project, env: environment, stdio: ['ignore', 'pipe', 'pipe'], windowsHide: true })
  const stdout = [], stderr = []
  let length = 0
  child.stdout.on('data', bytes => { length += bytes.length; if (length > 16 * 1024 * 1024) child.stdout.destroy(new Error('Actual CLI output exceeds its declared bound')); else stdout.push(bytes) })
  child.stderr.on('data', bytes => { if (Buffer.concat(stderr).length < 256 * 1024) stderr.push(bytes) })
  const outcome = await new Promise((resolve, reject) => { child.once('error', reject); child.stdout.once('error', reject); child.once('close', (exit, signal) => resolve({ exit, signal })) })
  const text = Buffer.concat(stdout).toString('utf8')
  await writeFile(join(evidence, `idempotency-${index}.stdout.json`), text, { flag: 'wx' })
  await writeFile(join(evidence, `idempotency-${index}.stderr.log`), Buffer.concat(stderr), { flag: 'wx' })
  const envelope = text.trim() ? JSON.parse(text) : null
  commands.push({ ordinal: index, args, started, ended: new Date().toISOString(), pid: child.pid, ...outcome, request_id: envelope?.request_id ?? null })
  assert.ok(envelope && allowed.includes(outcome.exit) && !outcome.signal, `Actual CLI failed; preserve idempotency-${index} and the original request`)
  if (outcome.exit === 0) { assert.equal(envelope.ok, true); if (runtimeId) assert.equal(envelope.runtime_id, runtimeId) }
  return { envelope, ...outcome }
}

function receipt(request, method) {
  const rows = database.prepare('SELECT request_id, method, state, receipt FROM mutation_receipts WHERE caller_fingerprint=? AND request_id=?').all(authority, request)
  assert.equal(rows.length, 1)
  assert.equal(rows[0].method, method)
  assert.equal(rows[0].state, 'completed')
  return { ...rows[0], receipt: JSON.parse(rows[0].receipt) }
}
async function inspect(request) {
  const observed = (await cli(['request', 'show', '--project', values['project-id'], '--request-id', request])).envelope.result
  assert.equal(observed.state, 'completed')
  return observed
}
async function mismatch(args) {
  const result = await cli(args, [2])
  assert.equal(result.envelope.ok, false)
  assert.equal(result.envelope.error.code, 'request_mismatch')
}
async function gone(owner) {
  const identities = [owner.process, owner.helper]
  assert.ok(identities.length === 2 && identities.every(row => row.pid > 0 && row.creationTimeMs > 0))
  const samples = []
  for (let index = 0; index < 2; index++) {
    if (index) await setTimeout(75)
    const rows = await new Promise((resolve, reject) => processAPI.getAllProcesses(value => value ? resolve(value) : reject(new Error('Actual native process inventory is unavailable')), 4))
    for (const expected of identities) {
      const row = rows.find(item => item.pid === expected.pid)
      assert.ok(!row || row.creationTimeMs !== undefined && row.creationTimeMs !== expected.creationTimeMs)
    }
    samples.push({ utc: new Date().toISOString(), expected: identities, observed: rows.filter(row => identities.some(old => old.pid === row.pid)) })
  }
  return samples
}

const status = (await cli(['status'])).envelope
runtimeId = status.runtime_id
const authority = status.authority_domain, dbPath = status.result.databases.find(row => row.name === 'main')?.file
assert.ok(authority && dbPath)
const rel = relative(dataRoot, await realpath(dbPath))
assert.ok(rel && !rel.startsWith('..') && !isAbsolute(rel))
database = new DatabaseSync(dbPath, { readOnly: true })
const remainingCases = ['actual_response_loss_after_each_effect_boundary', 'runtime_restart_pending_queries', 'worker_start_send_concurrent_and_disconnected_replays', 'different_request_close_create_start_interleavings']
try {
  const openR = randomUUID(), openArgs = ['controller', 'open', '--project', values['project-id'], '--workdir', project, '--request-id', openR]
  const opened = await Promise.all([cli(openArgs), cli(openArgs)])
  const controller = opened[0].envelope.controller
  assert.ok(controller)
  assert.equal(opened[1].envelope.controller, controller)
  assert.equal((await cli(openArgs)).envelope.controller, controller)
  const openReceipt = receipt(openR, 'swa.controllerOpen')
  assert.equal(openReceipt.receipt.controller, controller)
  assert.equal(openReceipt.receipt.client_mutation_id, `${openR}/controller-pty`)
  await inspect(openR)
  await mismatch(['controller', 'open', '--project', values['project-id'], '--workdir', input, '--request-id', openR])
  checks.push({ case: 'controller_open_same_request_concurrent_completed_and_changed_workdir', request: openR, receipt: openReceipt })

  const taskR = randomUUID(), specPath = join(input, `idempotency-${taskR}.json`), contextPath = join(input, `idempotency-${taskR}.txt`)
  const originalContext = await readFile(join(project, 'README.md'))
  await writeFile(contextPath, originalContext, { flag: 'wx' })
  const spec = JSON.parse(await readFile(join(input, 'task.json'), 'utf8'))
  spec.context.files.push(relative(project, contextPath).replaceAll('\\', '/'))
  await writeFile(specPath, JSON.stringify(spec) + '\n', { flag: 'wx' })
  const taskArgs = ['task', 'create', '--controller', controller, '--spec-file', specPath, '--request-id', taskR]
  const created = await Promise.all([cli(taskArgs, [0, 6]), cli(taskArgs, [0, 6])])
  assert.ok(created.some(row => row.exit === 0))
  for (const row of created.filter(row => row.exit === 6)) assert.equal(row.envelope.error.code, 'operation_unknown')
  const replay = (await cli(taskArgs)).envelope
  for (const row of created.filter(row => row.exit === 0)) { assert.equal(row.envelope.run_id, replay.run_id); assert.equal(row.envelope.task_id, replay.task_id) }
  const nativeTask = database.prepare('SELECT id, run_id, spec FROM tasks WHERE id=?').get(replay.task_id)
  assert.ok(nativeTask && nativeTask.run_id === replay.run_id)
  const fixedContext = JSON.parse(nativeTask.spec).context
  assert.ok(fixedContext.id && fixedContext.created_at)
  const taskReceipt = receipt(taskR, 'swa.taskCreate'), observedTask = await inspect(taskR)
  assert.deepEqual(taskReceipt.receipt.children, [`${taskR}/run-create`, `${taskR}/root-task`])
  assert.deepEqual(observedTask.children.map(row => row.requestId), taskReceipt.receipt.children)
  assert.ok(observedTask.children.every(row => row.state === 'completed'))
  for (const child of taskReceipt.receipt.children) receipt(child, child.endsWith('/run-create') ? 'orchestration.runCreate' : 'orchestration.taskCreate')
  await writeFile(specPath, JSON.stringify({ ...spec, goal: spec.goal + ' 请同时报告准确公开来源。' }) + '\n')
  try { await mismatch(taskArgs) } finally { await writeFile(specPath, JSON.stringify(spec) + '\n') }
  await writeFile(contextPath, Buffer.concat([originalContext, Buffer.from('\n来源：' + source.source + '\n')]))
  try { await mismatch(taskArgs) } finally { await writeFile(contextPath, originalContext) }
  assert.equal((await cli(taskArgs)).envelope.task_id, replay.task_id)
  assert.deepEqual(JSON.parse(database.prepare('SELECT spec FROM tasks WHERE id=?').get(replay.task_id).spec).context, fixedContext)
  assert.equal(database.prepare('SELECT COUNT(*) AS total FROM coordinator_runs WHERE coordinator_handle=?').get(controller).total, 1)
  assert.equal(database.prepare('SELECT COUNT(*) AS total FROM tasks WHERE run_id=? AND parent_id IS NULL').get(replay.run_id).total, 1)
  checks.push({ case: 'task_create_same_request_concurrent_completed_and_changed_spec_or_real_context', request: taskR, receipt: taskReceipt, fixedContext, concurrentUnknownResponses: created.filter(row => row.exit === 6).length })

  const closeR = randomUUID(), closeArgs = ['controller', 'close', '--controller', controller, '--request-id', closeR]
  await Promise.all([cli(closeArgs), cli(closeArgs)])
  await cli(closeArgs)
  const closed = receipt(closeR, 'swa.controllerClose'), exited = await gone(openReceipt.receipt)
  await inspect(closeR)
  const another = (await cli(['controller', 'open', '--project', values['project-id'], '--workdir', project, '--request-id', randomUUID()])).envelope.controller
  assert.ok(another && another !== controller)
  await mismatch(['controller', 'close', '--controller', another, '--request-id', closeR])
  const anotherOwner = (await cli(['controller', 'show', '--controller', another])).envelope.result.owner
  await cli(['controller', 'close', '--controller', another, '--request-id', randomUUID()])
  const anotherExited = await gone(anotherOwner)
  checks.push({ case: 'controller_close_same_request_concurrent_completed_and_other_controller', request: closeR, receipt: closed, exited, anotherExited })
  await writeFile(values.output, JSON.stringify({ candidate, source_revision: revision, runtime_id: runtimeId, checks, commands, model_calls: 0, remainingCases, scope: 'actual_native_non_model_idempotency_checks_only' }, null, 2) + '\n', { flag: 'wx' })
} finally { database.close() }
