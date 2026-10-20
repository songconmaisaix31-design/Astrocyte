import assert from 'node:assert/strict'
import { spawn, execFileSync } from 'node:child_process'
import { randomUUID } from 'node:crypto'
import { readFile, realpath, writeFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { DatabaseSync } from 'node:sqlite'
import { dirname, isAbsolute, join, relative } from 'node:path'
import { parseArgs } from 'node:util'
import { setTimeout } from 'node:timers/promises'

const { values } = parseArgs({ options: Object.fromEntries(['candidate', 'data-root', 'project', 'project-id', 'endpoint', 'protected-pids', 'output'].map(key => [key, { type: 'string' }])), allowPositionals: false })
for (const key of ['candidate', 'data-root', 'project', 'output']) assert.ok(values[key] && isAbsolute(values[key]), `Actual absolute ${key} is required`)
assert.ok(values['project-id'] && process.env.SWA_PROJECT_TOKEN && values['protected-pids'])
const candidate = await realpath(values.candidate), data = await realpath(values['data-root']), project = await realpath(values.project), evidence = dirname(values.output)
assert.equal((await realpath(process.execPath)).toLowerCase(), join(candidate, 'node', 'node.exe').toLowerCase())
const sourceRevision = (await readFile(join(candidate, 'VERSION'), 'utf8')).trim()
const codeRoot = await realpath(join(dirname(import.meta.filename), '..', '..'))
assert.equal(execFileSync('git', ['-C', codeRoot, 'rev-parse', 'HEAD'], { encoding: 'utf8', windowsHide: true }).trim(), sourceRevision)
assert.equal(execFileSync('git', ['-C', codeRoot, 'status', '--porcelain'], { encoding: 'utf8', windowsHide: true }).trim(), '')
const inputs = join(project, '.verification', 'swa-cli-input')
const source = JSON.parse(await readFile(join(inputs, 'source.json'), 'utf8'))
assert.equal(source.revision, 'da3bdddb085808863090a05f60cc1c64415039cf')
assert.equal((await realpath(source.project_root)).toLowerCase(), project.toLowerCase())
const spec = JSON.parse(await readFile(join(inputs, 'task.json'), 'utf8'))
const environment = { ...process.env, SWA_ENDPOINT: values.endpoint, SWA_PROJECT_ID: values['project-id'], NODE_OPTIONS: '--max-old-space-size=1024' }
const swa = join(candidate, 'swa.exe'), requireRuntime = createRequire(join(candidate, 'runtime', 'package.json'))
const processAPI = requireRuntime('@vscode/windows-process-tree'), { readServiceJobMember } = requireRuntime('@orca/windows-security')
const commands = [], checks = []
let ordinal = 0, runtimeId
async function cli(args, expected = 0) {
  const index = ++ordinal, child = spawn(swa, [...args, '--json'], { env: environment, cwd: project, stdio: ['ignore', 'pipe', 'pipe'], windowsHide: true })
  const stdout = [], stderr = []
  let bytes = 0
  child.stdout.on('data', chunk => { bytes += chunk.length; if (bytes > 16 * 1024 * 1024) child.stdout.destroy(new Error('Actual CLI output exceeds its bound')); else stdout.push(chunk) })
  child.stderr.on('data', chunk => { if (Buffer.concat(stderr).length < 256 * 1024) stderr.push(chunk) })
  const actual = await new Promise((resolve, reject) => { child.once('error', reject); child.stdout.once('error', reject); child.once('close', (exit, signal) => resolve({ exit, signal })) })
  const text = Buffer.concat(stdout).toString('utf8')
  await writeFile(join(evidence, `recovery-${index}.stdout.json`), text, { flag: 'wx' })
  await writeFile(join(evidence, `recovery-${index}.stderr.log`), Buffer.concat(stderr), { flag: 'wx' })
  const reply = JSON.parse(text)
  commands.push({ args, pid: child.pid, ...actual, request_id: reply.request_id })
  assert.equal(actual.exit, expected, `Actual CLI failed; preserve recovery-${index}`)
  assert.equal(actual.signal, null)
  assert.equal(reply.ok, expected === 0)
  if (runtimeId) assert.equal(reply.runtime_id, runtimeId)
  return reply
}
const status = await cli(['status'])
runtimeId = status.runtime_id
const dbPath = status.result.databases.find(row => row.name === 'main')?.file
assert.ok(dbPath && relative(data, await realpath(dbPath)) && !relative(data, await realpath(dbPath)).startsWith('..'))
const database = new DatabaseSync(dbPath, { readOnly: true }), fingerprint = status.authority_domain
const row = request => database.prepare('SELECT request_id, method, state, payload_hash, receipt FROM mutation_receipts WHERE caller_fingerprint=? AND request_id=?').get(fingerprint, request)
const counts = () => ({ runs: database.prepare('SELECT COUNT(*) AS count FROM coordinator_runs').get().count, tasks: database.prepare('SELECT COUNT(*) AS count FROM tasks').get().count })
const openRequest = randomUUID(), taskRequest = randomUUID()
const opened = await cli(['controller', 'open', '--project', values['project-id'], '--workdir', project, '--request-id', openRequest])
const controller = opened.controller
assert.ok(controller && opened.result.process && opened.result.helper)
const ownerFile = join(evidence, 'actual-controller-owner.json')
await writeFile(ownerFile, JSON.stringify(opened, null, 2) + '\n', { flag: 'wx' })
const created = await cli(['task', 'create', '--controller', controller, '--spec-file', join(inputs, 'task.json'), '--request-id', taskRequest])
const fixedRun = created.run_id, root = created.task_id, initialGeneration = created.result.native.run.consumer_generation
assert.ok(fixedRun && root && Number.isInteger(initialGeneration))
const before = counts(), originalChild = row(`${taskRequest}/run-create`)
assert.equal(originalChild.state, 'completed')
const probeState = join(evidence, 'strict-probe-source.json'), probeOutput = join(evidence, 'strict-probe-result.json')
await writeFile(probeState, JSON.stringify({ candidate, data_root: data, source_revision: sourceRevision, runtime_id: runtimeId, controller, task_request: taskRequest, objective: spec.goal }, null, 2) + '\n', { flag: 'wx' })
const probeEnv = Object.fromEntries(Object.entries(process.env).filter(([key]) => /^(SYSTEMROOT|WINDIR|COMSPEC|PATH|PATHEXT|TEMP|TMP)$/i.test(key)))
execFileSync(process.execPath, ['--max-old-space-size=1024', join(candidate, 'runtime', 'strict-caller-probe.cjs'), '--state', probeState, '--output', probeOutput], { env: probeEnv, windowsHide: true, timeout: 45_000 })
assert.deepEqual(row(`${taskRequest}/run-create`), originalChild)
assert.deepEqual(counts(), before)
checks.push({ case: 'missing_real_helper_evidence_before_completed_native_replay', source: probeOutput })
const liveRequest = randomUUID()
const live = await cli(['controller', 'recover', '--controller', controller, '--request-id', liveRequest], 6)
assert.equal(live.error.code, 'controller_unverifiable')
assert.equal(row(liveRequest), undefined)
assert.deepEqual(counts(), before)
checks.push({ case: 'live_prior_helper_effect_free_rejection', request_id: liveRequest, actual_discard: true })
const daemonDirectory = join(data, 'daemon')
const { readdir } = await import('node:fs/promises')
const pidFiles = (await readdir(daemonDirectory)).filter(file => /^daemon-v.*\.pid$/.test(file))
assert.equal(pidFiles.length, 1)
assert.equal(pidFiles[0], 'daemon-v49.pid')
const daemon = JSON.parse(await readFile(join(daemonDirectory, pidFiles[0]), 'utf8'))
assert.equal(daemon.swa_service_job.data_root.toLowerCase(), data.toLowerCase())
assert.equal(daemon.swa_service_job.install_root.toLowerCase(), candidate.toLowerCase())
for (const expected of [opened.result.process, opened.result.helper]) assert.equal(readServiceJobMember(daemon.swa_service_job.name, expected.pid).creationTimeMs, expected.creationTimeMs)
execFileSync('pwsh', ['-NoProfile', '-File', join(dirname(import.meta.filename), 'stop-owned-controller.ps1'), '-OwnerFile', ownerFile, '-ProtectedProcessIds', values['protected-pids'], '-OutputFile', join(evidence, 'actual-controller-fault.json')], { windowsHide: true, timeout: 30_000 })
for (let attempt = 0; ; attempt++) {
  const rows = await new Promise((resolve, reject) => processAPI.getAllProcesses(result => result ? resolve(result) : reject(new Error('Actual process inventory unavailable')), 4))
  if (![opened.result.process, opened.result.helper].some(expected => rows.some(actual => actual.pid === expected.pid))) break
  if (attempt >= 40) throw new Error('The actual prior helper exit remains unknown')
  await setTimeout(100)
}
const unavailable = await cli(['controller', 'show', '--controller', controller], 6)
assert.equal(unavailable.error.code, 'controller_unverifiable')
const historical = await cli(['task', 'show', '--controller', controller, '--task', root])
assert.ok(historical.result.native.tasks.some(task => task.id === root))
const request = randomUUID(), args = ['controller', 'recover', '--controller', controller, '--request-id', request]
const recovered = await cli(args)
assert.ok(recovered.controller && recovered.controller !== controller)
assert.equal(recovered.run_id, fixedRun)
assert.equal(recovered.task_id, root)
assert.ok(recovered.result.consumer_generation > initialGeneration)
assert.equal(recovered.result.native.run.id, fixedRun)
assert.equal(recovered.result.native.run.coordinator_handle, recovered.controller)
assert.equal(recovered.result.prior_exit_observations.length, 2)
const completedRecovery = row(request), replay = await cli(args)
assert.equal(replay.controller, recovered.controller)
assert.equal(replay.result.consumer_generation, recovered.result.consumer_generation)
assert.deepEqual(row(request), completedRecovery)
assert.deepEqual(counts(), before)
const stale = await cli(['controller', 'recover', '--controller', controller, '--request-id', randomUUID()], 3)
assert.equal(stale.error.code, 'scope_denied')
const duplicateTask = await cli(['task', 'create', '--controller', recovered.controller, '--spec-file', join(inputs, 'task.json'), '--request-id', randomUUID()], 2)
assert.equal(duplicateTask.error.code, 'request_mismatch')
await cli(['controller', 'close', '--controller', recovered.controller, '--request-id', randomUUID()])
checks.push({ case: 'same_run_root_context_new_helper_and_completed_replay', request_id: request, prior_controller: controller, controller: recovered.controller, run_id: fixedRun, root_task_id: root, consumer_generation: recovered.result.consumer_generation })
await writeFile(values.output, JSON.stringify({ candidate, source_revision: sourceRevision, runtime_id: runtimeId, model_calls: 0, commands, checks, same_native_counts: before, remainingCases: ['runtime_and_serve_restart_same_named_job_and_authenticated_daemon', 'real_response_loss_each_recovery_effect_boundary', 'pending_recovery_observation_and_competing_request_barrier', 'real_restored_helper_attestation_and_old_generation', 'other_real_project_and_run_scope', 'same_live_worker_session_and_unknown_resource_preservation'] }, null, 2) + '\n', { flag: 'wx' })
database.close()
