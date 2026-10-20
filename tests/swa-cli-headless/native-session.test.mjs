import assert from 'node:assert/strict'
import { readFile, realpath, writeFile } from 'node:fs/promises'
import { isAbsolute, join } from 'node:path'
import { parseArgs } from 'node:util'
import { createRequire } from 'node:module'

const { values } = parseArgs({ options: { candidate: { type: 'string' }, 'flow-evidence': { type: 'string' }, output: { type: 'string' } }, allowPositionals: false })
for (const key of ['candidate', 'flow-evidence', 'output']) if (!values[key] || !isAbsolute(values[key])) throw new Error(`Actual absolute ${key} is required`)
const candidate = await realpath(values.candidate)
assert.equal((await realpath(process.execPath)).toLowerCase(), join(candidate, 'node', 'node.exe').toLowerCase())
const flow = JSON.parse(await readFile(values['flow-evidence'], 'utf8'))
const revision = (await readFile(join(candidate, 'VERSION'), 'utf8')).trim()
assert.equal((await realpath(flow.candidate)).toLowerCase(), candidate.toLowerCase())
assert.equal(flow.source_revision, revision)
assert.ok(flow.actualOverlap && flow.dispatchA && flow.dispatchB && flow.dispatchA !== flow.dispatchB)
assert.ok(Array.isArray(flow.goneA) && flow.goneA.length === 2 && Array.isArray(flow.goneB) && flow.goneB.length === 2)
assert.ok(flow.activeStop.beforeTurn.body.state === 'running' && flow.activeStop.endedTurn.body.state === 'interrupted' || flow.activeStop.actualNativeAbort)
const exported = JSON.parse(await readFile(flow.resultFile, 'utf8'))
assert.ok(exported.ok && exported.runtime_id === flow.runtime_id && exported.run_id === flow.run_id)
const workers = exported.result.native.workers
assert.equal(workers.length, 2)
const requireRuntime = createRequire(join(candidate, 'runtime', 'package.json'))
const { readNativePiSessionSource } = requireRuntime('./task-result.cjs')
const observations = []
for (const worker of workers) {
  const selection = worker.source.selection
  assert.equal(selection.agent_id, 'pi')
  assert.equal(selection.native_launch.tools.mode, 'disabled')
  assert.equal(selection.native_launch.tools.native_contract, 'pi-cli-1.0.1')
  assert.equal(selection.native_launch.account.kind, 'pi-native-api-key')
  const actual = worker.native_session
  const source = await readNativePiSessionSource(actual.path, join(selection.native_launch.account.private_home, 'sessions'), await realpath(actual.header.cwd), worker.source.native_session_id, worker.worker.dispatchId === flow.dispatchA, selection)
  assert.deepEqual(source.entries, actual.entries)
  const events = flow.provider_events.find(item => item.source.native_session_id === worker.source.native_session_id)
  assert.ok(events?.rows.length > 0)
  for (const event of events.rows) {
    assert.equal(event.projection, 'actual_native_cli_lifecycle_and_delta_fields')
    assert.ok(event.native_cli.pid > 0 && event.native_cli.creationTimeMs > 0 && event.native_cli.pid !== event.process.pid)
  }
  observations.push({ dispatch_id: worker.worker.dispatchId, native_session_id: worker.source.native_session_id, source_file: actual.path, availability: source.availability, native_cli: events.rows[0].native_cli })
}
assert.ok(observations.some(row => row.dispatch_id === flow.dispatchB && row.availability === 'complete'))
await writeFile(values.output, JSON.stringify({ candidate, source_revision: revision, flow_evidence: values['flow-evidence'], runtime_id: flow.runtime_id, actual_workers: observations, actual_stop: flow.activeStop, additional_model_calls: 0, scope: 'actual_installed_native_cli_flow_and_sources', remainingCases: [] }, null, 2) + '\n', { flag: 'wx' })
console.log(JSON.stringify({ output: values.output, actual_native_worker_sources: observations.length, additional_model_calls: 0 }))
