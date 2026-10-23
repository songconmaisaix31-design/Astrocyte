import assert from 'node:assert/strict'
import { createHash } from 'node:crypto'
import { lstat, mkdir, readFile, realpath, writeFile } from 'node:fs/promises'
import { isAbsolute, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { parseArgs } from 'node:util'
import { gzipSync, gunzipSync } from 'node:zlib'

const revision = '101e46815ef86de3bd73a2a1de7448ce68b908a6'
const run = 40027206964, job = 120666286706, artifact = 12532883615
const sourceHashes = {
  observer: 'f9f9af16b2958e9bdc47239ee9e5b27eaeec33e28490710d49b923b645955e9a',
  producer: '7a7009267cb781b56bada56c908d0dc294f99eb384188cbde46a2e8fe6511f13',
  closure: '144d16fe707b432e9cfc05aa800d0719430ea112fc23c8706a46f9664436475f',
  native: '50891e0c2b415a46f5775d2c83b1db1090f45a474059f35e7c905873af771509'
}
const originals = [
  { label: 'query', file_name: 'service-query-evidence.json', byte_count: 204079, sha256: 'c5584661c77378ebe28d5c5a343f66e29917f3018194bba58b43cfd59da9b4bf' },
  { label: 'lifecycle', file_name: 'service-process-lifecycle.json', byte_count: 75487, sha256: 'bb2a7fd5cf253f46ac84c91ac07dc914602882fc067834db1dd59fd5f2852649' }
]
const sha = bytes => createHash('sha256').update(bytes).digest('hex')
const count = value => Number.isSafeInteger(value) && value >= 0
const keys = (value, expected) => {
  assert.ok(value && typeof value === 'object' && !Array.isArray(value))
  assert.deepEqual(Object.keys(value).sort(), [...expected].sort())
}
const escape = value => value.replaceAll('%', '%25').replaceAll('\r', '%0D').replaceAll('\n', '%0A')
const state = { schema_version: 1, original_run: run, original_job: job, original_artifact: artifact, original_revision: revision,
  original_status: 'FAIL', status: 'READ_FAILED', phase: 'arguments', copied_files: [], service_execution: 'NOT_RUN',
  sdk_execution: 'NOT_RUN', native_build: 'NOT_RUN', fixture_execution: 'NOT_RUN', model_calls: 0, worker_calls: 0 }
let output

// Mechanical copy of the reviewed probe's complete safe evidence validators.
function validateDocuments(queryDocument, lifecycle) {
  const keys = (value, names) => { assert.ok(value && typeof value === 'object' && !Array.isArray(value)); assert.deepEqual(Object.keys(value).sort(), [...names].sort()) }
  const uint = value => Number.isSafeInteger(value) && value >= 0 && value <= 0xffffffff
  const birth = value => Number.isSafeInteger(value) && value > 0
  const finite = value => typeof value === 'number' && Number.isFinite(value) && value >= 0
  const list = (value, maximum, check) => { assert.ok(Array.isArray(value) && value.length <= maximum); for (const row of value) check(row) }
  const bool = value => assert.equal(typeof value, 'boolean')
  const atom = value => assert.ok(value === null || typeof value === 'string' && /^[A-Za-z][A-Za-z0-9_.-]{0,63}$/.test(value))
  const identity = value => { keys(value, ['pid', 'creationTimeMs']); assert.ok(uint(value.pid) && value.pid > 0 && birth(value.creationTimeMs)) }
  const instant = value => { keys(value, ['observed_at_ms', 'monotonic_ms']); assert.ok(birth(value.observed_at_ms) && finite(value.monotonic_ms)) }
  const sourceRoles = { serve_native_identity: ['serve'], callback_owned: ['callback'], named_job_member: ['serve', 'runtime', 'server', 'daemon'],
    transport_native_identity: ['transport'], transport_receipt_identity: ['tty_root', 'tty_helper', 'pipe_helper'],
    producer_command_native_identity: ['producer_command'], controller_receipt_native_identity: ['controller_process', 'controller_helper'] }
  const sourceInfo = value => {
    keys(value, ['kind', 'role', 'command_sequence', 'command_name', 'query_sequence', 'scope_index', 'scope_anchor'])
    assert.ok(typeof value.kind === 'string' && Object.hasOwn(sourceRoles, value.kind) && sourceRoles[value.kind].includes(value.role))
    if (value.kind === 'producer_command_native_identity') {
      assert.ok(birth(value.command_sequence)); assert.ok(typeof value.command_name === 'string' && /^(?:git|cli)-[0-9]{2,8}$/.test(value.command_name))
      assert.equal(Number(value.command_name.split('-')[1]), value.command_sequence)
    } else { assert.equal(value.command_sequence, null); assert.equal(value.command_name, null) }
    if (value.kind === 'callback_owned') { assert.ok(birth(value.query_sequence) && uint(value.scope_index) && value.scope_index <= 2048); identity(value.scope_anchor) }
    else { assert.equal(value.query_sequence, null); assert.equal(value.scope_index, null); assert.equal(value.scope_anchor, null) }
  }
  const origin = value => { keys(value, ['identity', 'confirmed_at', 'event_sequence', 'source']); identity(value.identity); instant(value.confirmed_at); assert.ok(birth(value.event_sequence)); sourceInfo(value.source) }
  const nativeRow = value => {
    keys(value, ['pid', 'ppid', 'name', 'creationTimeDiagnosticVersion', 'creationTimeProvider', 'creationTimeNtStatus', 'creationTimeSnapshotComplete', ...(Object.hasOwn(value, 'creationTimeMs') ? ['creationTimeMs'] : [])])
    assert.ok(uint(value.pid) && value.pid > 0 && uint(value.ppid)); assert.ok(typeof value.name === 'string' && value.name.length <= 260 && /^[\p{L}\p{N}_. ()-]*$/u.test(value.name))
    if (Object.hasOwn(value, 'creationTimeMs')) assert.ok(birth(value.creationTimeMs))
    assert.equal(value.creationTimeDiagnosticVersion, 2); assert.equal(value.creationTimeProvider, 'NtQuerySystemInformation'); assert.ok(uint(value.creationTimeNtStatus)); bool(value.creationTimeSnapshotComplete)
  }
  const processEvent = value => {
    if (value === null) return
    keys(value, ['exit', 'signal', 'observed_at_ms', 'monotonic_ms']); assert.ok(value.exit === null || Number.isSafeInteger(value.exit)); atom(value.signal)
    assert.ok(birth(value.observed_at_ms) && finite(value.monotonic_ms))
  }
  const queryInstant = value => { keys(value, ['observed_at_ms', 'monotonic_ms', 'root_exit_event', 'root_close_event']); assert.ok(birth(value.observed_at_ms) && finite(value.monotonic_ms)); processEvent(value.root_exit_event); processEvent(value.root_close_event) }
  const unknownReasons = ['invalid_native_process_row', 'current_root_identity_unverified', 'root_in_descendant_branch', 'parent_incarnation_unverified', 'process_identity_not_unique', 'creation_time_unavailable', 'parent_incarnation_changed']
  const closure = value => {
    keys(value, ['rootVerified', 'ownedRows', 'excludedEdges', 'unknownRows']); bool(value.rootVerified); list(value.ownedRows, 16384, nativeRow)
    list(value.excludedEdges, 32768, row => { keys(row, ['reason', 'actual', ...(Object.hasOwn(row, 'anchor') ? ['anchor'] : ['parent'])]); assert.ok(['different-root-incarnation', 'predates-root', 'predates-parent'].includes(row.reason)); nativeRow(row.actual); identity(row.anchor ?? row.parent) })
    list(value.unknownRows, 32768, row => {
      keys(row, ['reason', 'actual', ...(Object.hasOwn(row, 'source_anchor') ? ['source_anchor'] : [])]); assert.ok(unknownReasons.includes(row.reason)); nativeRow(row.actual)
      if (row.source_anchor) { keys(row.source_anchor, ['kind', 'pid', 'creation_time_ms']); assert.ok(['serve', 'additional'].includes(row.source_anchor.kind)); assert.ok(uint(row.source_anchor.pid) && row.source_anchor.pid > 0 && birth(row.source_anchor.creation_time_ms)) }
    })
  }
  const sample = value => {
    if (value === null) return
    keys(value, ['sequence', 'query_window', 'root', 'priors', 'before', 'rows', 'scopes', 'ownership_closure_incomplete']); assert.ok(birth(value.sequence)); identity(value.root)
    keys(value.query_window, ['start', 'end']); queryInstant(value.query_window.start); queryInstant(value.query_window.end)
    assert.ok(value.query_window.start.monotonic_ms <= value.query_window.end.monotonic_ms)
    list(value.priors, 2048, identity); keys(value.before, ['prior_sources', 'additional_roots']); list(value.before.prior_sources, 2048, origin); list(value.before.additional_roots, 2048, identity)
    assert.deepEqual(value.before.prior_sources.map(row => row.identity), value.priors)
    for (const row of value.before.prior_sources) assert.ok(row.confirmed_at.monotonic_ms <= value.query_window.start.monotonic_ms)
    list(value.rows, 16384, nativeRow); bool(value.ownership_closure_incomplete)
    list(value.scopes, 2049, row => {
      keys(row, ['anchor', 'source', 'existed_before_query', 'confirmed_during_query', 'result']); identity(row.anchor); origin(row.source); assert.deepEqual(row.anchor, row.source.identity)
      bool(row.existed_before_query); bool(row.confirmed_during_query); closure(row.result)
      assert.equal(row.confirmed_during_query, row.source.confirmed_at.monotonic_ms >= value.query_window.start.monotonic_ms && row.source.confirmed_at.monotonic_ms <= value.query_window.end.monotonic_ms)
    })
    assert.ok(value.scopes.length > 0); assert.deepEqual(value.scopes[0].anchor, value.root)
    value.scopes.forEach((row, index) => assert.equal(row.existed_before_query, (index === 0 ? value.priors : value.before.additional_roots).some(prior => prior.pid === row.anchor.pid && prior.creationTimeMs === row.anchor.creationTimeMs)))
    if (value.scopes.some(row => row.result.unknownRows.length > 0)) assert.equal(value.ownership_closure_incomplete, true)
  }
  for (const [query, value] of [[true, queryDocument], [false, lifecycle]]) {
      keys(value, ['schema_version', 'source_revision', 'sources', 'status', 'evidence_incomplete', 'error_kind', ...(query ? ['query_count', 'ownership_closure_incomplete', 'first_sample', 'first_unknown_sample', 'last_sample'] : ['events'])])
      assert.equal(value.schema_version, 1); assert.equal(value.source_revision, revision); assert.deepEqual(value.sources, sourceHashes)
      assert.ok(['OBSERVING', 'COMPLETE', 'INCOMPLETE'].includes(value.status)); bool(value.evidence_incomplete)
      assert.ok(value.error_kind === null || ['EVIDENCE_WRITE_OR_LIMIT', 'LIFECYCLE_INCOMPLETE', 'SOURCE_INCOMPLETE', 'PRIOR_SOURCE_INCOMPLETE', 'SCOPE_SOURCE_INCOMPLETE', 'NATIVE_BATCH_INCOMPLETE', 'QUERY_EVIDENCE_INCOMPLETE', 'OWNED_SCOPE_MISSING', 'QUERY_FAILED', 'OBSERVER_ABORTED'].includes(value.error_kind))
      if (query) {
        assert.ok(uint(value.query_count)); bool(value.ownership_closure_incomplete)
        for (const item of [value.first_sample, value.first_unknown_sample, value.last_sample]) { sample(item); if (item) assert.ok(item.sequence <= value.query_count) }
        if (value.first_unknown_sample) assert.ok(value.first_unknown_sample.scopes.some(row => row.result.unknownRows.length > 0))
      } else list(value.events, 4096, row => {
        keys(row, ['sequence', 'at', 'kind', 'source', 'pid', 'identity', 'exit', 'signal', 'spawn_error', 'stdio_policy']); assert.ok(birth(row.sequence)); instant(row.at); sourceInfo(row.source)
        assert.ok(['spawn', 'spawn-error', 'identity-confirm', 'receipt-observed', 'exit', 'close', 'stdout-end', 'stderr-end', 'stdin-end-request', 'stdin-finish', 'cleanup-request'].includes(row.kind))
        assert.ok(row.pid === null || uint(row.pid) && row.pid > 0); if (row.identity !== null) { identity(row.identity); assert.equal(row.pid, row.identity.pid) }
        assert.ok(row.exit === null || Number.isSafeInteger(row.exit)); atom(row.signal); atom(row.spawn_error)
        assert.equal(row.stdio_policy, row.kind === 'spawn' ? row.source.kind === 'serve_native_identity' ? 'pipe_pipe_pipe' : 'ignore_pipe_pipe' : null)
      })
  }
  const query = queryDocument
      lifecycle.events.forEach((row, index) => { assert.equal(row.sequence, index + 1); if (index) assert.ok(row.at.monotonic_ms >= lifecycle.events[index - 1].at.monotonic_ms) })
      for (const sample of [query.first_sample, query.first_unknown_sample, query.last_sample].filter(Boolean)) {
        for (const origin of [...sample.before.prior_sources, ...sample.scopes.map(row => row.source)]) {
          const event = lifecycle.events[origin.event_sequence - 1]
          assert.ok(event && ['identity-confirm', 'receipt-observed'].includes(event.kind)); assert.deepEqual(event.identity, origin.identity); assert.deepEqual(event.at, origin.confirmed_at); assert.deepEqual(event.source, origin.source)
        }
      }
      assert.ok(query.first_sample && query.last_sample && query.query_count > 0)
      assert.equal(query.last_sample.sequence, query.query_count)
  assert.equal(query.status, 'INCOMPLETE'); assert.equal(lifecycle.status, 'INCOMPLETE')
  for (const value of [query, lifecycle]) { assert.equal(value.evidence_incomplete, true); assert.equal(value.error_kind, 'OBSERVER_ABORTED') }
  assert.equal(query.query_count, 649); assert.equal(query.ownership_closure_incomplete, false)
  assert.equal(query.first_unknown_sample, null); assert.equal(lifecycle.events.length, 116)
  for (const [sample, sequence, rows, priors, additional, scopes] of [
    [query.first_sample, 1, 137, 1, 0, 1], [query.last_sample, 649, 145, 55, 14, 15]
  ]) {
    assert.equal(sample.sequence, sequence); assert.equal(sample.rows.length, rows); assert.equal(sample.priors.length, priors)
    assert.equal(sample.before.additional_roots.length, additional); assert.equal(sample.scopes.length, scopes)
    assert.equal(sample.ownership_closure_incomplete, false)
    assert.equal(sample.scopes[0].result.rootVerified, true)
    assert.ok(sample.scopes.every(scope => scope.result.unknownRows.length === 0))
  }
}

function commandRows(rows) {
  assert.ok(Array.isArray(rows) && rows.length > 0 && rows.length <= 16384)
  const names = new Set()
  return rows.map(row => {
    keys(row, ['name', 'exit', 'timed_out', 'stdout_bytes', 'stderr_bytes', 'stdout_eof', 'stderr_eof',
      'stream_failure', 'identity_captured', 'cleanup_attempted', 'kernel_tree_exit'])
    assert.ok(typeof row.name === 'string' && /^command-[0-9]{2,8}$/.test(row.name))
    assert.ok(Number(row.name.slice(8)) > 0 && !names.has(row.name)); names.add(row.name)
    assert.equal(row.kernel_tree_exit, 'NOT_CLAIMED')
    for (const key of ['timed_out', 'stdout_eof', 'stderr_eof', 'stream_failure', 'identity_captured', 'cleanup_attempted']) assert.equal(typeof row[key], 'boolean')
    for (const key of ['stdout_bytes', 'stderr_bytes']) assert.ok(count(row[key]))
    if (row.exit !== 'UNKNOWN') {
      keys(row.exit, ['exit', 'signal', 'spawn_error'])
      assert.ok(row.exit.exit === null || Number.isSafeInteger(row.exit.exit))
      for (const key of ['signal', 'spawn_error']) assert.ok(row.exit[key] === null || typeof row.exit[key] === 'string' && /^[A-Za-z][A-Za-z0-9_.-]{0,63}$/.test(row.exit[key]))
    }
    return row
  })
}

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
  const destination = join(checkout, '.verification/swa-ci/windows-service-query-retained')
  await mkdir(destination)
  output = destination
  const tracked = [], stableFields = ['dev', 'ino', 'size', 'mtimeMs', 'ctimeMs', 'nlink']
  const readerPath = fileURLToPath(import.meta.url), readerBefore = await lstat(readerPath), readerSHA = sha(await readFile(readerPath))
  const read = async (name, maximum) => {
    const path = join(input, name), before = await lstat(path)
    assert.ok(before.isFile() && !before.isSymbolicLink() && before.nlink === 1 && before.size <= maximum)
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
    schema: 'swa.windows-controller-open-development.v1', status: 'FAIL', phase: 'short-identify-then-controller-open',
    scope: 'development_service_observation_not_candidate', observation_revision: revision, go_source_revision: revision,
    source_revision: '7ff3a7d7d7e029c2fc9e2ae1fb6b81da8e0d4f8f', runtime_source_revision: '7ff3a7d7d7e029c2fc9e2ae1fb6b81da8e0d4f8f',
    input_revision: '1d3006fe0df0cb04b7f0de1e4563264bd0bfd9da', native_input_revision: '1d3006fe0df0cb04b7f0de1e4563264bd0bfd9da',
    model_calls: 0, worker_calls: 0, task_create: 'NOT_RUN', fixture_exit: 1, private_raw_uploaded: false
  }
  for (const [key, value] of Object.entries(expected)) assert.equal(saved[key], value)
  assert.equal(saved.error?.kind, 'ENOENT')
  assert.deepEqual(saved.native_snapshot, { provider: 'NtQuerySystemInformation', version: 2, ntstatus: 0,
    self_identity_matches: true, root_child_control: 'PRIOR_PASS_NOT_RERUN' })
  assert.equal(saved.diagnostic_process_tree.binary_sha256, sourceHashes.native)
  for (const [role, name] of [['observer', 'serve-lifecycle.mjs'], ['producer', 'controller-taskcreate-check.mjs'], ['closure', 'process-incarnation-closure.mjs']]) {
    const matches = saved.observer_sources.filter(row => row.name === name)
    assert.equal(matches.length, 1); assert.equal(matches[0].observation_sha256, sourceHashes[role])
  }
  assert.equal(saved.windows_private_data_sdk.exit, 0)
  assert.equal(saved.windows_private_data_sdk.cases.filter(row => row.action === 'pass').length, 7)
  assert.equal(saved.windows_private_data_sdk.cases.filter(row => ['fail', 'skip'].includes(row.action)).length, 0)
  assert.deepEqual(saved.prior_ordinary_control, { status: 'PRIOR_PASS_NOT_RERUN', head: 'ff33f6b9bcdd4b37eb330b61db49bd14d5476fb9',
    run: 40021139196, job: 120645694135, fixture_sha256: '842060a11640491694dbb663057b71dd94bc58e24905acb006fdc61319a74ac4', closure_sha256: sourceHashes.closure })
  assert.equal(saved.service_evidence.complete, false); assert.equal(saved.service_evidence.binding, 'VERIFIED')
  assert.deepEqual(saved.service_evidence.sources, sourceHashes)
  assert.equal(saved.service_evidence.error_kind, 'OBSERVER_ABORTED'); assert.equal(saved.service_evidence.evidence_incomplete, true)
  assert.equal(saved.service_evidence.ownership_closure_incomplete, false)
  assert.equal(saved.service_evidence.query_count, 649); assert.equal(saved.service_evidence.first_unknown, null)
  assert.equal(saved.service_evidence.lifecycle_event_count, 116); assert.equal(saved.service_evidence.lifecycle_incomplete, true)
  const commands = commandRows(saved.commands)
  await copy('original-result.json', original)
  state.phase = 'complete-safe-originals'
  const actualFiles = []
  assert.ok(Array.isArray(saved.service_evidence.files)); assert.equal(saved.service_evidence.files.length, originals.length)
  for (const expectedFile of originals) {
    const actual = await read(expectedFile.file_name, expectedFile.byte_count)
    assert.equal(actual.bytes.length, expectedFile.byte_count); assert.equal(actual.sha256, expectedFile.sha256)
    const declarations = saved.service_evidence.files.filter(row => row.file_name === expectedFile.file_name)
    assert.equal(declarations.length, 1)
    assert.deepEqual(declarations[0], { file_name: expectedFile.file_name, status: 'SAVED', sha256: actual.sha256, byte_count: actual.bytes.length })
    actualFiles.push(actual)
  }
  validateDocuments(actualFiles[0].value, actualFiles[1].value)
  for (let index = 0; index < originals.length; index++) await copy(originals[index].file_name, actualFiles[index])
  state.phase = 'typed-original-result'
  const failedCommands = commands.filter(row => row.exit === 'UNKNOWN' || row.exit.exit !== 0 || row.timed_out)
  assert.ok(failedCommands.length <= 8)
  const report = { original_run: run, original_job: job, original_artifact: artifact, original_revision: revision,
    original_status: saved.status, phase: saved.phase, scope: saved.scope, sources: sourceHashes,
    original_result_byte_count: original.bytes.length, original_result_sha256: original.sha256, reader_sha256: readerSHA,
    command_count: commands.length, failed_command_count: failedCommands.length, failed_commands: failedCommands,
    windows_private_data_sdk: { exit: 0, pass_count: 7, skip_count: 0 }, native_self: saved.native_snapshot,
    ordinary_control: 'PRIOR_PASS_NOT_RERUN', service_acceptance: 'FAIL', native_absence: 'NOT_ESTABLISHED', cim: 'NOT_ESTABLISHED',
    outer_error_kind: 'ENOENT', missing_file: 'smoke/result.json', original_observer_exception: 'NOT_CAPTURED', producer_error_message: 'NOT_CAPTURED',
    original_query_count: 649, retained_callback_count: 2, lifecycle_event_count: 116,
    service_execution: 'NOT_RUN', model_calls: 0, worker_calls: 0, d17_prior_ownership: 'NOT_CAPTURED', kernel_tree_exit_proven: false }
  state.phase = 'lossless-budget'
  const combined = Buffer.concat(actualFiles.map(actual => actual.bytes))
  assert.equal(combined.length, 279566)
  const compressed = gzipSync(combined, { level: 9 }), compressedSHA = sha(compressed)
  assert.ok(compressed.length <= 14400)
  assert.ok(gunzipSync(compressed, { maxOutputLength: 279566 }).equals(combined))
  const encoded = compressed.toString('base64')
  assert.ok(encoded.length <= 19200 && /^[A-Za-z0-9+/]+={0,2}$/.test(encoded))
  assert.ok(Buffer.from(encoded, 'base64').equals(compressed))
  const chunks = []
  for (let offset = 0; offset < encoded.length; offset += 2400) chunks.push(encoded.slice(offset, offset + 2400))
  assert.ok(chunks.length > 0 && chunks.length <= 8)
  const manifest = { original_run: run, original_job: job, original_artifact: artifact, original_revision: revision, reader_sha256: readerSHA,
    codec: 'gzip+base64', order: ['query', 'lifecycle'], files: originals, original_byte_count: combined.length, original_sha256: sha(combined),
    gzip_byte_count: compressed.length, gzip_sha256: compressedSHA, base64_length: encoded.length, chunk_count: chunks.length }
  const annotations = [
    ['Retained service query original result', report], ['Retained service query manifest', manifest],
    ...chunks.map((data, index) => ['Retained service query chunk', { original_run: run, original_job: job, original_artifact: artifact,
      original_revision: revision, gzip_sha256: compressedSHA, index: index + 1, count: chunks.length, data }])
  ].map(([title, value]) => ({ title, text: JSON.stringify(value) }))
  assert.ok(annotations.length <= 10)
  assert.ok(annotations.every(row => Buffer.byteLength(row.text) <= 3500))
  assert.ok(annotations.reduce((total, row) => total + Buffer.byteLength(row.text), 0) <= 32768)
  state.phase = 'source-stability'
  for (const actual of tracked) {
    const after = await lstat(actual.path)
    for (const field of stableFields) assert.equal(after[field], actual.before[field])
    assert.equal(sha(await readFile(actual.path)), actual.sha256)
  }
  const readerAfter = await lstat(readerPath)
  for (const field of stableFields) assert.equal(readerAfter[field], readerBefore[field])
  assert.equal(sha(await readFile(readerPath)), readerSHA)
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
  const failure = { original_run: run, original_job: job, original_artifact: artifact, original_revision: revision, original_status: 'FAIL',
    status: 'READ_FAILED', phase: state.phase, error_kind: state.error_kind, copied_file_count: state.copied_files.length }
  process.stdout.write('::warning title=Retained service query reader failure::' + escape(JSON.stringify(failure)) + '\n')
  process.exitCode = 1
}

