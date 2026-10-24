import assert from 'node:assert/strict'
import { createHash } from 'node:crypto'
import { lstat, mkdir, open, readdir, realpath, writeFile } from 'node:fs/promises'
import { dirname, isAbsolute, join, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { parseArgs } from 'node:util'

const original = { run: 40077869987, job: 120825053338, attempt: 1,
  head: '12c25800513a25e3de86455f99d7595383243f09', artifact: 12554530773, artifact_bytes: 5345,
  artifact_digest: 'sha256:dbda819cf786338f5bbfc86f1cd5153192898298aad032decb72da4d381b4e67' }
const source = '5c25b1ce6cf044eb2bcea0c9bf7142a8f1e84850'
const product = 'b464e70ba95967eb1ecc216f5f3d25407f1ecc66'
const resultHash = '90823848cedb2bda89370aab0049f42a22379f00b278c3c78eb13482c377f330'
const nativeHash = 'b8fcfad529c4cfad749ecccea3bcc5bf5afdc6159a4f66a7e45559261e2ac64d'
const patchHash = '658074a08f16d983f2aac160384d0159b8738d7000670e58eae0b1ef6a8db504'
const archiveHash = '42909a812d9c9c08b363d803e8de0ef1db85d6a1c71a5b83e0730a4da4e51f46'
const consumers = {
  'serve-lifecycle.mjs': 'bd45e0cff35fe9e5e6415587b38849ad2507b629218cb6f395e476cfe5e79671',
  'windows-service-evidence-summary.mjs': '52c45f5e4f282b43f27bde34b179ba0b9e41d6a38d46d15a8a15690e57e6a9c5',
  'controller-taskcreate-check.mjs': 'e1cba5dfdee30fd91232d38481475fef4b606a56166905d47ee20bc0b676ff34',
  'controller-input-check.mjs': '1a3d46e31c4be89aa334177339220e9791d36d24fa7d5d79cfeaf90fe8d25008',
  'process-incarnation-closure.mjs': '144d16fe707b432e9cfc05aa800d0719430ea112fc23c8706a46f9664436475f',
  'observe-process-exits.ps1': '1e91327b9723bd0a3f3a6b2dbba802afdc012034d9c0ae4ef9ec9301cc9682b7'
}
const sha = bytes => createHash('sha256').update(bytes).digest('hex')
const object = value => assert.ok(value && typeof value === 'object' && !Array.isArray(value))
const keys = (value, names) => { object(value); assert.deepEqual(Object.keys(value).sort(), [...names].sort()) }
const bool = value => assert.equal(typeof value, 'boolean')
const count = value => assert.ok(Number.isSafeInteger(value) && value >= 0)
const finite = value => assert.ok(typeof value === 'number' && Number.isFinite(value) && value >= 0)
const digest = value => assert.ok(typeof value === 'string' && /^[0-9a-f]{64}$/.test(value))
const choice = (value, choices) => assert.ok(choices.includes(value))
const fields = ['dev', 'ino', 'mode', 'nlink', 'uid', 'gid', 'size', 'mtimeMs', 'ctimeMs', 'birthtimeMs']
const inputs = ['SOURCE', 'preparation.json', 'result.json']
const state = { schema_version: 1, status: 'READ_FAILED', original, files: [], original_diagnostic_status: 'FAIL',
  exact_preparation_substep: 'UNKNOWN', product_execution: 'NOT_RUN', native_build: 'NOT_RUN', model_calls: 0 }
let output, attempts = 0

async function ordinary(path, directory) {
  path = resolve(path)
  for (let cursor = path; ; cursor = dirname(cursor)) {
    const stat = await lstat(cursor)
    assert.ok(!stat.isSymbolicLink()); assert.equal(await realpath(cursor), cursor)
    if (cursor === path) { assert.ok(directory ? stat.isDirectory() : stat.isFile()); if (!directory) assert.equal(stat.nlink, 1) }
    else assert.ok(stat.isDirectory())
    if (dirname(cursor) === cursor) break
  }
}

async function stable(path, cap) {
  await ordinary(path, false)
  const before = await lstat(path)
  count(before.size); assert.ok(before.size <= cap)
  const same = stat => {
    assert.ok(stat.isFile()); assert.equal(stat.nlink, 1)
    for (const key of fields) assert.equal(stat[key], before[key])
  }
  let handle, bytes, failure
  try {
    handle = await open(path, 'r'); same(await handle.stat())
    const limit = Math.min(before.size + 1, cap + 1), parts = []
    let position = 0
    while (position < limit) {
      const buffer = Buffer.allocUnsafe(Math.min(65536, limit - position))
      const read = await handle.read(buffer, 0, buffer.length, position)
      if (!read.bytesRead) break
      position += read.bytesRead; assert.ok(position <= before.size && position <= cap)
      parts.push(buffer.subarray(0, read.bytesRead))
    }
    bytes = Buffer.concat(parts); assert.equal(bytes.length, before.size)
    same(await handle.stat()); await ordinary(path, false); same(await lstat(path))
  } catch (error) { failure = error }
  try { if (handle) await handle.close() } catch {
    state.secondary_error_kind = 'FILE_CLOSE_FAILED'
    failure ??= new Error('Fixed input close failed')
  }
  if (failure) throw failure
  return bytes
}

function preparation(value) {
  keys(value, ['status', 'native_component', 'native_component_index', 'diagnostic_only', 'other_native_components_built',
    'original_inputs_modified', 'node_version', 'node_sha256', 'headers_sha256', 'import_library_sha256',
    'python_version', 'compiler_year', 'actual_vs_version', 'actual_cl_version'])
  assert.equal(value.status, 'PASS'); assert.equal(value.native_component, '@vscode/windows-process-tree')
  assert.equal(value.native_component_index, 1); assert.equal(value.diagnostic_only, true)
  assert.equal(value.other_native_components_built, false); assert.equal(value.original_inputs_modified, false)
  assert.equal(value.node_version, '24.21.0'); assert.equal(value.compiler_year, '2026')
  assert.equal(value.node_sha256, 'ba4e6d110e8c1592a1ecd390f6b05f3da124b13871a5be62b341a07a853c6c32')
  assert.equal(value.headers_sha256, '57c6bee2e30bbbee5bd51d6cc343eb992e174b56a2a1d0eab7a7510771c20ea2')
  assert.equal(value.import_library_sha256, 'a0a84aa03917b578d286010b7521837e9ff1136ffb2e4a406c14316c33fd06f7')
  for (const name of ['python_version', 'actual_vs_version', 'actual_cl_version'])
    assert.ok(typeof value[name] === 'string' && /^\d+(?:\.\d+){1,4}$/.test(value[name]))
  return value
}

function clock(value) {
  keys(value, ['observed_at_ms', 'monotonic_ms']); count(value.observed_at_ms); finite(value.monotonic_ms)
}

function commands(rows) {
  assert.ok(Array.isArray(rows) && rows.length <= 32)
  return rows.map((value, index) => {
    keys(value, ['stage', 'exit', 'signal', 'timed_out', 'close_observed', 'stdout_eof', 'stderr_eof', 'stdout_bytes', 'stderr_bytes',
      'stream_failure', 'spawn_error', 'output_complete', 'start', 'end'])
    choice(value.stage, ['preflight', 'source-archive', 'frozen-source-tools', 'diagnostic-copy', 'single-native-build', 'retained-product-extraction'])
    assert.ok(value.exit === null || Number.isSafeInteger(value.exit))
    choice(value.signal, [null, 'SIGTERM', 'SIGKILL', 'SIGINT', 'SIGABRT', 'SIGHUP', 'SIGPIPE', 'UNKNOWN'])
    for (const name of ['timed_out', 'close_observed', 'stdout_eof', 'stderr_eof', 'stream_failure', 'spawn_error', 'output_complete']) bool(value[name])
    count(value.stdout_bytes); count(value.stderr_bytes); clock(value.start); clock(value.end)
    const elapsed = value.end.monotonic_ms - value.start.monotonic_ms; finite(elapsed)
    // 位置只表示已保存数组中的次序，不补未返回命令的身份或结果。
    return { array_index: index, ...value, elapsed_ms: elapsed }
  })
}

function retained(value) {
  object(value)
  const expected = ['metadata', 'archive_sha256', 'version', 'installation']
  keys(value, Object.hasOwn(value, 'extraction') ? [...expected, 'extraction'] : expected)
  keys(value.metadata, ['id', 'name', 'size_in_bytes', 'digest', 'expired', 'workflow_run'])
  keys(value.metadata.workflow_run, ['id', 'head_sha'])
  assert.equal(value.metadata.id, 12542236014); assert.equal(value.metadata.name, 'swa-windows-combined-' + product)
  assert.equal(value.metadata.size_in_bytes, 310804282)
  assert.equal(value.metadata.digest, 'sha256:3579aec40c304d64255a9469eb6c69cb72112f8a5a6f43c30f9636d3e7b1c4d0')
  assert.equal(value.metadata.expired, false); assert.equal(value.metadata.workflow_run.id, 40047585581)
  assert.equal(value.metadata.workflow_run.head_sha, product); assert.equal(value.archive_sha256, archiveHash)
  assert.equal(value.version, product); assert.equal(value.installation, 'NOT_RUN')
  const result = { ...value, extraction: 'NOT_CAPTURED' }
  if (Object.hasOwn(value, 'extraction')) {
    const item = value.extraction
    keys(item, ['entry_count', 'file_count', 'expanded_bytes', 'maximum_file_bytes', 'free_before', 'free_before_copy', 'files_sha256', 'product_inputs_unchanged'])
    for (const name of ['entry_count', 'file_count', 'expanded_bytes', 'maximum_file_bytes', 'free_before', 'free_before_copy']) count(item[name])
    digest(item.files_sha256); bool(item.product_inputs_unchanged); result.extraction = item
  }
  return result
}

function progress(value) {
  object(value)
  assert.equal(value.schema_version, 1); assert.equal(value.source_revision, source)
  assert.equal(value.product_revision, product); assert.equal(value.observation_revision, original.head)
  assert.equal(value.scope, 'retained_b464_development_service_not_candidate')
  assert.equal(value.status, 'FAIL'); assert.equal(value.phase, 'retained-product-extraction'); assert.equal(value.error_kind, 'DEADLINE')
  assert.ok(Array.isArray(value.secondary_error_kinds) && value.secondary_error_kinds.length <= 1)
  for (const kind of value.secondary_error_kinds) assert.equal(kind, 'FILE_CLOSE_FAILED')
  assert.equal(value.native_sha256, nativeHash); assert.equal(value.diagnostic_patch_sha256, patchHash)
  assert.equal(value.model_calls, 0); assert.equal(value.service_started, false)
  assert.equal(value.other_native_components_built, false); assert.equal(value.kernel_tree_exit_proven, false)
  assert.equal(value.query, 'NOT_CAPTURED'); assert.equal(value.query_process, 'NOT_CAPTURED')
  for (const name of ['service_process', 'summary_process']) if (Object.hasOwn(value, name)) assert.equal(value[name], 'NOT_CAPTURED')
  let consumerSources = 'NOT_CAPTURED'
  if (Object.hasOwn(value, 'service_consumer_sources')) {
    assert.ok(Array.isArray(value.service_consumer_sources))
    assert.deepEqual(value.service_consumer_sources.map(row => row.name).sort(), Object.keys(consumers).sort())
    for (const row of value.service_consumer_sources) { keys(row, ['name', 'sha256']); assert.equal(row.sha256, consumers[row.name]) }
    consumerSources = value.service_consumer_sources
  }
  return { source_revision: source, product_revision: product, observation_revision: original.head,
    status: value.status, phase: value.phase, error_kind: value.error_kind, secondary_error_kinds: value.secondary_error_kinds,
    native_sha256: nativeHash, service_started: false, service_process: 'NOT_CAPTURED', summary_process: 'NOT_CAPTURED',
    query: 'NOT_CAPTURED', model_calls: 0, kernel_tree_exit_proven: false,
    retained_product: retained(value.retained_product), service_consumer_sources: consumerSources }
}

async function save(name, value) {
  const bytes = Buffer.from(JSON.stringify(value, null, 2) + '\n'); assert.ok(bytes.length <= 65536)
  const path = join(output, name)
  await writeFile(path, bytes, { flag: 'wx', mode: 0o600 }); assert.ok(bytes.equals(await stable(path, 65536)))
}

async function warning(message) {
  attempts++
  await new Promise((done, reject) => process.stdout.write('::warning::' + message.replaceAll('%', '%25').replaceAll('\r', '%0D').replaceAll('\n', '%0A') + '\n', error => error ? reject(error) : done()))
}

async function main() {
  const { values } = parseArgs({ options: { 'artifact-root': { type: 'string' } }, strict: true, allowPositionals: false })
  assert.equal(process.platform, 'linux'); assert.ok(isAbsolute(process.env.RUNNER_TEMP ?? ''))
  assert.match(process.env.GITHUB_SHA ?? '', /^[0-9a-f]{40}$/)
  const checkout = dirname(dirname(fileURLToPath(import.meta.url)))
  assert.equal(resolve(process.env.GITHUB_WORKSPACE ?? ''), checkout)
  const root = resolve(values['artifact-root'] ?? '')
  assert.equal(root, join(resolve(process.env.RUNNER_TEMP), 'swa-retained-preparation-12c25800'))
  for (const path of [join(checkout, '.verification'), join(checkout, '.verification/swa-ci')]) {
    try { await mkdir(path, { mode: 0o700 }) } catch (error) { if (error.code !== 'EEXIST') throw error }
    await ordinary(path, true)
  }
  const destination = join(checkout, '.verification/swa-ci/windows-preparation-retained')
  await mkdir(destination, { mode: 0o700 }); output = destination; await ordinary(output, true)
  await ordinary(root, true); assert.deepEqual((await readdir(root)).sort(), [...inputs].sort())
  const bytes = {}
  for (const name of inputs) {
    bytes[name] = await stable(join(root, name), name === 'SOURCE' ? 64 : 65536)
    state.files.push({ file_name: name, sha256: sha(bytes[name]), byte_count: bytes[name].length })
  }
  assert.equal(bytes.SOURCE.toString('utf8'), source + '\n'); assert.equal(sha(bytes['result.json']), resultHash)
  const json = name => JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(bytes[name]))
  state.preparation = preparation(json('preparation.json'))
  const result = json('result.json')
  state.progress = progress(result); state.commands = commands(result.commands)
  state.reader_revision = process.env.GITHUB_SHA
  state.reader_sha256 = sha(await stable(fileURLToPath(import.meta.url), 65536))
  for (const item of state.files) assert.equal(sha(await stable(join(root, item.file_name), item.file_name === 'SOURCE' ? 64 : 65536)), item.sha256)
  assert.deepEqual((await readdir(root)).sort(), [...inputs].sort())
  state.status = 'READ_COMPLETE'
  const { commands: rows, progress: savedProgress, ...overview } = state
  const messages = [JSON.stringify(overview), JSON.stringify({ schema_version: 1, kind: 'saved_preparation_progress',
    original_run: original.run, result_sha256: resultHash, progress: savedProgress })]
  // 只在完整对象之间分页；先核全部预算，再输出任何一条。
  let group = []
  const pack = items => JSON.stringify({ schema_version: 1, kind: 'saved_command_receipts', original_run: original.run, result_sha256: resultHash, commands: items })
  for (const row of rows) {
    if (Buffer.byteLength(pack([...group, row])) > 3500) { assert.ok(group.length); messages.push(pack(group)); group = [] }
    group.push(row)
  }
  if (group.length) messages.push(pack(group))
  assert.ok(messages.length <= 10)
  assert.ok(messages.every(message => Buffer.byteLength(message) <= 3500))
  assert.ok(messages.reduce((total, message) => total + Buffer.byteLength(message), 0) <= 32768)
  state.annotation_count = messages.length
  await save('summary.json', state)
  for (const message of messages) await warning(message)
}

try { await main() } catch {
  state.status = 'READ_FAILED'; state.error_kind = 'FIXED_EVIDENCE_READ_FAILED'
  if (output) { try { await save('failure.json', state) } catch { state.failure_file = 'NOT_SAVED' } }
  if (attempts === 0) {
    try { await warning(JSON.stringify({ schema_version: 1, status: 'READ_FAILED', original_run: original.run,
      error_kind: state.error_kind, original_diagnostic_status: 'FAIL', product_execution: 'NOT_RUN' })) } catch {}
  }
  process.exitCode = 1
}
