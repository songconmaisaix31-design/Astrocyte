import assert from 'node:assert/strict'
// 普通进程 fixture 使用安装候选的真实 Job 和关闭适配，不调用模型。
import { readFile, realpath, writeFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { dirname, isAbsolute, join } from 'node:path'
import { createInterface } from 'node:readline'
import { parseArgs } from 'node:util'

const { values } = parseArgs({ options: { candidate: { type: 'string' }, output: { type: 'string' } }, allowPositionals: false })
if (!values.candidate || !values.output || !isAbsolute(values.candidate) || !isAbsolute(values.output)) throw new Error('Explicit installed candidate and new evidence file are required')
const candidate = await realpath(values.candidate)
const node = join(candidate, 'node', 'node.exe')
if ((await realpath(process.execPath)).toLowerCase() !== node.toLowerCase()) throw new Error('This fixture must run with the installed candidate Node')
const requireRuntime = createRequire(join(candidate, 'runtime', 'package.json'))
const { SwaProviderProcessTree } = requireRuntime('./swa-provider-process-tree.cjs')
const { spawnManagedProviderProcess } = requireRuntime('./managed-provider-process.cjs')
const fixture = join(dirname(import.meta.filename), 'job-fixture.mjs')
const records = []
const unrelatedTree = new SwaProviderProcessTree()
const unrelated = spawnManagedProviderProcess({ command: node, args: [fixture, 'root', 'forced', join(candidate, 'runtime', 'provider-job-bootstrap.mjs')], swaJobIPC: true }, { site: 'swa-unrelated-owned-process-fixture', acceptClose: value => value.root === 'exited' && value.tree === 'exited' })
unrelated.child.stdout.resume()
unrelatedTree.bind(unrelated.child, () => ({ observed: unrelated.rootExitObserved, processless: unrelated.processless }))
await unrelatedTree.ready()
const unrelatedBefore = unrelatedTree.readObservation()
const native = requireRuntime('node-pty/lib/utils').loadNativeModule('conpty').module
const negativeJob = native.createProviderJob()
assert.throws(() => native.assignProviderJob(negativeJob, unrelatedBefore.rootPid, unrelatedBefore.rootCreationTimeMs + 1), /creation identity changed/)
assert.throws(() => native.assignProviderJob(negativeJob, unrelatedBefore.rootPid, unrelatedBefore.rootCreationTimeMs), /another owned Provider Job/)
assert.equal(native.queryProviderJob(negativeJob).assigned, false)
assert.deepEqual(native.queryProviderJob(negativeJob).members, [])
assert.equal(native.closeProviderJob(negativeJob), true)
assert.throws(() => native.queryProviderJob(negativeJob), /ownership unavailable/)
for (const mode of ['normal', 'forced', 'crash', 'late-child']) {
  const tree = new SwaProviderProcessTree()
  const managed = spawnManagedProviderProcess({ command: node, args: [fixture, 'root', mode, join(candidate, 'runtime', 'provider-job-bootstrap.mjs')], swaJobIPC: true }, { site: 'swa-owned-process-fixture', acceptClose: value => value.root === 'exited' && value.tree === 'exited' })
  tree.bind(managed.child, () => ({ observed: managed.rootExitObserved, processless: managed.processless }))
  const seen = []
  const ready = new Promise((resolve, reject) => {
    const timer = setTimeout(() => reject(new Error('Actual fixture descendants did not become ready')), 10_000)
    const lines = createInterface({ input: managed.child.stdout, crlfDelay: Infinity })
    lines.on('line', line => {
      try {
        const row = JSON.parse(line)
        if (row.type !== 'owned_process_fixture_ready') throw new Error('Unexpected process fixture output')
        seen.push(row)
        if (seen.length === 3) { clearTimeout(timer); resolve() }
      } catch (error) { clearTimeout(timer); reject(error) }
    })
    managed.onExit(() => { if (seen.length < 3) { clearTimeout(timer); reject(new Error('Actual fixture root exited before descendants were observed')) } })
  })
  await tree.ready()
  await ready
  const before = tree.readObservation()
  assert.equal(before.members.length, 3)
  for (const row of seen) assert.ok(before.members.some(member => member.pid === row.pid && member.creationTimeMs > 0))
  const closed = await managed.close(tree)
  assert.deepEqual(closed, { root: 'exited', tree: 'exited' })
  assert.equal(managed.rootExitObserved, true)
  assert.equal(tree.treeVerdict, 'exited')
  const unrelatedAfter = unrelatedTree.readObservation()
  assert.equal(unrelated.rootExitObserved, false)
  assert.equal(unrelatedAfter.rootPid, unrelatedBefore.rootPid)
  assert.equal(unrelatedAfter.rootCreationTimeMs, unrelatedBefore.rootCreationTimeMs)
  assert.ok(unrelatedAfter.members.some(member => member.pid === unrelatedBefore.rootPid && member.creationTimeMs === unrelatedBefore.rootCreationTimeMs))
  records.push({ mode, actualProcesses: seen, admittedJob: before, close: closed, unrelatedAfter })
}
assert.deepEqual(await unrelated.close(unrelatedTree), { root: 'exited', tree: 'exited' })
await writeFile(values.output, JSON.stringify({ candidate, source_revision: (await readFile(join(candidate, 'VERSION'), 'utf8')).trim(), cases: records, model_calls: 0, scope: 'ordinary_owned_process_fixture', unrelatedBefore, negativeCases: ['actual_closed_job_rejected', 'wrong_creation_identity_rejected', 'different_owned_job_reassignment_rejected'], remainingCases: ['parent_loss', 'actual_sdk_abort_then_close', 'missing_native_pre_input'] }, null, 2) + '\n', { flag: 'wx' })
console.log(JSON.stringify({ process_fixture_cases: records.length, output: values.output, real_agent_acceptance: 'NOT_RUN' }))
