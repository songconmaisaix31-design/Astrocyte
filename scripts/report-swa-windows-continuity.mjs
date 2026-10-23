import assert from 'node:assert/strict'
import { createHash } from 'node:crypto'
import { lstat, mkdir, readFile, realpath, writeFile } from 'node:fs/promises'
import { dirname, isAbsolute, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { parseArgs } from 'node:util'
import { gzipSync, gunzipSync } from 'node:zlib'

const revision = 'b234cf81d07bc1ce5af9e470db3cf34515b6855a'
const run = 40012110754, artifact = 12527445614
const fixtureSHA = '5ac3aec5ae32682577566a00ddff1776994881b97f19c5257fd9b46d095d975a'
const closureSHA = '144d16fe707b432e9cfc05aa800d0719430ea112fc23c8706a46f9664436475f'
const nativeSHA = '9bb7d35e47f06e06fc93482520a33d512312e4890c08e8a0de90e8f6fdaaa338'
const frames = [
  { label: 'a', file_name: 'continuity-snapshot-a.json', byte_count: 41436, row_count: 139, sha256: '65deef3fe51b72cb8f06e580e835af6fe056067afb74549d924b384858e75203' },
  { label: 'b', file_name: 'continuity-snapshot-b.json', byte_count: 42025, row_count: 141, sha256: 'b49f411a35fac122365f4d83bdfe232381a878104a3d5fead3cfd9b2508bc084' },
  { label: 'c', file_name: 'continuity-snapshot-c.json', byte_count: 40848, row_count: 137, sha256: '781b20c2d61691cd8dac5e310fd8d3d1fbcef9648cbd69b6ba4dfdc3f1fee07f' }
]
const sha = bytes => createHash('sha256').update(bytes).digest('hex')
const object = value => value !== null && typeof value === 'object' && !Array.isArray(value) && Object.getPrototypeOf(value) === Object.prototype
const keys = (value, expected) => { assert.ok(object(value)); assert.deepEqual(Object.keys(value).sort(), [...expected].sort()) }
const count = value => Number.isSafeInteger(value) && value >= 0
const pid = value => count(value) && value > 0 && value <= 0xffffffff
const birth = value => count(value) && value > 0
const finite = value => typeof value === 'number' && Number.isFinite(value) && value >= 0
const identity = value => {
  assert.ok(object(value) && pid(value.pid) && birth(value.creationTimeMs))
  assert.ok(value.ppid === 'UNKNOWN' || count(value.ppid) && value.ppid <= 0xffffffff)
  return { pid: value.pid, ppid: value.ppid, creationTimeMs: value.creationTimeMs }
}
const instant = value => {
  keys(value, ['observed_at_ms', 'monotonic_ms'])
  assert.ok(birth(value.observed_at_ms) && finite(value.monotonic_ms))
  return { observed_at_ms: value.observed_at_ms, monotonic_ms: value.monotonic_ms }
}
const window = value => {
  keys(value, ['start', 'end'])
  const result = { start: instant(value.start), end: instant(value.end) }
  assert.ok(result.start.observed_at_ms <= result.end.observed_at_ms && result.start.monotonic_ms <= result.end.monotonic_ms)
  return result
}
const state = { schema_version: 1, original_run: run, original_artifact: artifact, original_revision: revision,
  original_status: 'FAIL', status: 'READ_FAILED', phase: 'arguments', copied_files: [], fixture_execution: 'NOT_RUN', native_build: 'NOT_RUN', model_calls: 0 }
let output
const escape = value => value.replaceAll('%', '%25').replaceAll('\r', '%0D').replaceAll('\n', '%0A')
async function main() {
  const { values } = parseArgs({ options: { 'artifact-root': { type: 'string' } }, strict: true, allowPositionals: false })
  assert.ok(values['artifact-root'] && isAbsolute(values['artifact-root']))
  const checkout = await realpath(fileURLToPath(new URL('../', import.meta.url)))
  const input = await realpath(values['artifact-root'])
  assert.equal(input.toLowerCase(), resolve(values['artifact-root']).toLowerCase())
  assert.ok((await lstat(input)).isDirectory() && !(await lstat(input)).isSymbolicLink())
  for (const path of [join(checkout, '.verification'), join(checkout, '.verification/swa-ci')]) {
    try { await mkdir(path) } catch (error) { if (error.code !== 'EEXIST') throw error }
    assert.ok((await lstat(path)).isDirectory() && !(await lstat(path)).isSymbolicLink())
    assert.equal((await realpath(path)).toLowerCase(), resolve(path).toLowerCase())
  }
  const destination = join(checkout, '.verification/swa-ci/windows-continuity-retained')
  await mkdir(destination)
  output = destination
  const tracked = []
  const stableFields = ['dev', 'ino', 'size', 'mtimeMs', 'ctimeMs', 'nlink']
  const read = async (name, limit) => {
    const path = join(input, name), before = await lstat(path)
    assert.ok(before.isFile() && !before.isSymbolicLink() && before.nlink === 1 && before.size <= limit)
    assert.equal((await realpath(path)).toLowerCase(), resolve(path).toLowerCase())
    const bytes = await readFile(path), after = await lstat(path)
    for (const field of stableFields) assert.equal(after[field], before[field])
    assert.equal(bytes.length, before.size)
    const value = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(bytes))
    const actual = { path, before, bytes, value, sha256: sha(bytes) }
    tracked.push(actual)
    return actual
  }
  const copy = async (name, actual) => {
    await writeFile(join(output, name), actual.bytes, { flag: 'wx' })
    assert.equal(sha(await readFile(join(output, name))), actual.sha256)
    state.copied_files.push({ file_name: name, byte_count: actual.bytes.length, sha256: actual.sha256 })
  }
  state.phase = 'result-source'
  const original = await read('result.json', 2 * 1024 * 1024), saved = original.value
  const expected = {
    schema: 'swa.windows-controller-open-development.v1', status: 'FAIL', phase: 'native-process-snapshot-control',
    scope: 'development_native_abi_only_not_candidate', observation_revision: revision,
    source_revision: '7ff3a7d7d7e029c2fc9e2ae1fb6b81da8e0d4f8f', runtime_source_revision: '1d3006fe0df0cb04b7f0de1e4563264bd0bfd9da',
    input_revision: '1d3006fe0df0cb04b7f0de1e4563264bd0bfd9da', native_input_revision: '1d3006fe0df0cb04b7f0de1e4563264bd0bfd9da',
    go_source_revision: 'NOT_RUN', go_build: 'NOT_RUN', windows_sdk: 'NOT_RUN', controller: 'NOT_RUN', service_execution: 'NOT_RUN',
    task_create: 'NOT_RUN', model_calls: 0, worker_calls: 0, native_compile_only: false, native_abi_only: true
  }
  for (const [name, value] of Object.entries(expected)) assert.equal(saved[name], value)
  const fixture = saved.root_child_fixture
  assert.equal(fixture.helper_sha256, fixtureSHA)
  assert.equal(saved.diagnostic_process_tree.binary_sha256, nativeSHA)
  const closures = saved.observer_sources.filter(row => row.name === 'process-incarnation-closure.mjs')
  assert.equal(closures.length, 1); assert.equal(closures[0].observation_sha256, closureSHA)
  assert.equal(fixture.status, 'FAIL'); assert.equal(fixture.exit, 6); assert.equal(fixture.failed_phase, 'prior-owned-continuity')
  assert.deepEqual(fixture.cases.map(row => row.name), ['live-root-child', 'normal-eof', 'prior-owned-continuity'])
  const continuity = fixture.cases[2]
  assert.equal(continuity.status, 'FAIL'); assert.equal(continuity.stage, 'snapshot-c'); assert.equal(continuity.error_kind, 'ERR_ASSERTION')
  assert.equal(continuity.snapshots.count, 3); assert.equal(continuity.snapshots.truncated, false)
  assert.equal(continuity.snapshots.rows.length, 3); assert.equal(saved.continuity_frames.length, 3)
  await copy('original-result.json', original)
  state.phase = 'original-frames'
  const actualFrames = [], queries = []
  for (const expectedFrame of frames) {
    const actual = await read(expectedFrame.file_name, 8 * 1024 * 1024)
    assert.equal(actual.bytes.length, expectedFrame.byte_count); assert.equal(actual.sha256, expectedFrame.sha256)
    const frame = actual.value
    keys(frame, ['schema_version', 'query_window', 'rows', 'label'])
    assert.equal(frame.schema_version, 1); assert.equal(frame.label, expectedFrame.label)
    assert.ok(Array.isArray(frame.rows)); assert.equal(frame.rows.length, expectedFrame.row_count)
    const query = window(frame.query_window)
    if (queries.length) assert.ok(queries.at(-1).end.monotonic_ms <= query.start.monotonic_ms)
    queries.push(query)
    for (const row of frame.rows) {
      keys(row, ['pid', 'ppid', 'name', 'creationTimeDiagnosticVersion', 'creationTimeProvider', 'creationTimeNtStatus', 'creationTimeSnapshotComplete',
        ...(Object.hasOwn(row, 'creationTimeMs') ? ['creationTimeMs'] : [])])
      assert.ok(pid(row.pid) && count(row.ppid) && row.ppid <= 0xffffffff)
      if (Object.hasOwn(row, 'creationTimeMs')) assert.ok(birth(row.creationTimeMs))
      assert.ok(typeof row.name === 'string' && row.name.length <= 260 && /^[\p{L}\p{N}_. ()-]*$/u.test(row.name))
      assert.equal(row.creationTimeDiagnosticVersion, 2); assert.equal(row.creationTimeProvider, 'NtQuerySystemInformation')
      assert.equal(row.creationTimeNtStatus, 0); assert.equal(row.creationTimeSnapshotComplete, true)
    }
    const declarations = saved.continuity_frames.filter(row => row.label === expectedFrame.label)
    const observations = continuity.snapshots.rows.filter(row => row.label === expectedFrame.label)
    assert.equal(declarations.length, 1); assert.equal(observations.length, 1)
    const declaration = declarations[0], observation = observations[0]
    assert.equal(declaration.status, 'SAVED'); assert.equal(declaration.file_name, expectedFrame.file_name)
    for (const row of [declaration, observation]) {
      assert.equal(row.sha256, actual.sha256); assert.equal(row.byte_count, actual.bytes.length); assert.equal(row.actual_row_count, frame.rows.length)
    }
    assert.deepEqual(observation.query_window, query)
    assert.deepEqual(observation.native_query, { provider: 'NtQuerySystemInformation', version: 2, ntstatus: 0, snapshot_complete: true })
    await copy(expectedFrame.file_name, actual)
    actualFrames.push(actual)
  }
  state.phase = 'typed-original-result'
  const prior = (value, expectedCount) => {
    assert.equal(value.count, expectedCount); assert.equal(value.truncated, false)
    assert.ok(Array.isArray(value.rows)); assert.equal(value.rows.length, expectedCount)
    return { count: value.count, truncated: false, rows: value.rows.map(identity) }
  }
  assert.deepEqual(continuity.parent_exit, { exit: 0, signal: null, spawn_failed: false })
  assert.equal(continuity.parent_stdout_eof, true); assert.equal(continuity.parent_stderr_eof, true)
  assert.equal(continuity.forced_cleanup, false); assert.equal(continuity.cleanup_unknown, 'NOT_RECORDED')
  assert.equal(continuity.child_exit_code, 'UNKNOWN'); assert.equal(continuity.native_exit_observations, 'UNKNOWN')
  assert.equal(fixture.cases[1].cim_process_exit, 'UNKNOWN')
  const report = { original_run: run, original_artifact: artifact, original_revision: revision, original_status: saved.status,
    original_result_sha256: original.sha256, scope: saved.scope, stage: continuity.stage, error_kind: continuity.error_kind,
    original_assertion_stack: 'NOT_CAPTURED', helper_sha256: fixtureSHA, closure_sha256: closureSHA, native_sha256: nativeSHA,
    controller: identity(continuity.controller), parent: identity(continuity.parent), child_identity_from_b: identity(continuity.child),
    prior_a: prior(continuity.prior_a, 4), prior_b: prior(continuity.prior_b, 6), parent_exit: continuity.parent_exit,
    parent_close_observed: instant(continuity.parent_close_observed), parent_stdout_eof: true, parent_stderr_eof: true,
    forced_cleanup: false, cleanup_unknown: 'NOT_RECORDED', child_exit_code: 'UNKNOWN', cim: 'NOT_REACHED',
    fixture_execution: 'NOT_RUN', model_calls: 0, d17_prior_ownership: 'NOT_CAPTURED' }
  state.phase = 'lossless-budget'
  const combined = Buffer.concat(actualFrames.map(frame => frame.bytes))
  assert.equal(combined.length, 124309)
  const compressed = gzipSync(combined, { level: 9 }), compressedSHA = sha(compressed)
  assert.ok(compressed.length <= 12288)
  assert.ok(gunzipSync(compressed, { maxOutputLength: 124309 }).equals(combined))
  const encoded = compressed.toString('base64')
  assert.ok(encoded.length <= 16384 && /^[A-Za-z0-9+/]+={0,2}$/.test(encoded))
  assert.ok(Buffer.from(encoded, 'base64').equals(compressed))
  const chunks = []
  for (let offset = 0; offset < encoded.length; offset += 2400) chunks.push(encoded.slice(offset, offset + 2400))
  assert.ok(chunks.length > 0 && chunks.length <= 7)
  const manifest = { original_run: run, original_artifact: artifact, original_revision: revision,
    codec: 'gzip+base64', order: ['a', 'b', 'c'], frames, original_byte_count: combined.length, original_sha256: sha(combined),
    gzip_byte_count: compressed.length, gzip_sha256: compressedSHA, base64_length: encoded.length, chunk_count: chunks.length }
  const annotations = [
    ['Retained continuity original result', report], ['Retained continuity frame manifest', manifest],
    ...chunks.map((data, index) => ['Retained continuity frame chunk', { original_run: run, original_artifact: artifact,
      original_revision: revision, gzip_sha256: compressedSHA, index: index + 1, count: chunks.length, data }])
  ].map(([title, value]) => ({ title, text: JSON.stringify(value) }))
  assert.ok(annotations.length <= 9)
  assert.ok(annotations.every(row => Buffer.byteLength(row.text) <= 3500))
  assert.ok(annotations.reduce((total, row) => total + Buffer.byteLength(row.text), 0) <= 32768)
  state.phase = 'source-stability'
  for (const actual of tracked) {
    const after = await lstat(actual.path)
    for (const field of stableFields) assert.equal(after[field], actual.before[field])
    assert.equal(sha(await readFile(actual.path)), actual.sha256)
  }
  state.status = 'LOSSLESS_TRANSPORT_COMPLETE'; state.phase = 'complete'
  state.manifest = manifest; state.original_result = report
  state.annotation_count = annotations.length; state.annotation_utf8_bytes = annotations.map(row => Buffer.byteLength(row.text))
  await writeFile(join(output, 'summary.json'), JSON.stringify(state, null, 2) + '\n', { flag: 'wx' })
  for (const annotation of annotations) process.stdout.write('::warning title=' + annotation.title + '::' + escape(annotation.text) + '\n')
}
try { await main() } catch (error) {
  state.status = 'READ_FAILED'
  state.error_kind = typeof error?.code === 'string' && /^[A-Z][A-Z0-9_]{0,63}$/.test(error.code) ? error.code : 'UNKNOWN'
  if (output) try { await writeFile(join(output, 'failure.json'), JSON.stringify(state, null, 2) + '\n', { flag: 'wx' }) } catch {}
  const failure = { original_run: run, original_artifact: artifact, original_revision: revision, original_status: 'FAIL',
    status: 'READ_FAILED', phase: state.phase, error_kind: state.error_kind, copied_file_count: state.copied_files.length }
  process.stdout.write('::warning title=Retained continuity reader failure::' + escape(JSON.stringify(failure)) + '\n')
  process.exitCode = 1
}
