import assert from 'node:assert/strict'
import { execFileSync, spawn } from 'node:child_process'
import { createRequire } from 'node:module'
import { createHash } from 'node:crypto'
import { lstat, mkdir, readFile, readdir, realpath, writeFile } from 'node:fs/promises'
import { dirname, extname, isAbsolute, join } from 'node:path'
import { pathToFileURL } from 'node:url'
import { parseArgs } from 'node:util'
import { setTimeout } from 'node:timers/promises'
import { collectOwnedIncarnations } from './process-incarnation-closure.mjs'

const { values } = parseArgs({ options: { ...Object.fromEntries(['candidate', 'evidence', 'data-parent', 'protected-roots', 'port', 'mode', 'source-revision', 'edition'].map(name => [name, { type: 'string' }])), 'controller-input-check': { type: 'boolean', default: false }, 'controller-open-only': { type: 'boolean', default: false } }, allowPositionals: false })
assert.ok(!(values['controller-input-check'] && values['controller-open-only']), 'Controller diagnostic modes are mutually exclusive')
const controllerCheck = values['controller-input-check'] || values['controller-open-only']
const edition = values.edition ?? 'cli'
assert.ok(['cli', 'combined'].includes(edition), 'Unknown candidate edition')
for (const key of ['candidate', 'evidence', 'data-parent']) assert.ok(values[key] && isAbsolute(values[key]))
assert.ok(values['protected-roots'] && /^\d+$/.test(values.port))
const candidate = await realpath(values.candidate), evidence = values.evidence, dataParent = values['data-parent']
assert.equal((await realpath(process.execPath)).toLowerCase(), join(candidate, 'node', 'node.exe').toLowerCase())
await mkdir(evidence)
const source = (await readFile(join(candidate, 'VERSION'), 'utf8')).trim()
assert.match(source, /^[0-9a-f]{40}$/)
if (edition === 'combined') assert.equal((await readFile(join(candidate, 'EDITION'), 'utf8')).trim(), edition)
if (values['source-revision'] !== undefined) {
  assert.match(values['source-revision'], /^[0-9a-f]{40}$/)
  assert.equal(source, values['source-revision'])
}
assert.equal(execFileSync(join(candidate, 'swa.exe'), ['--version'], { encoding: 'utf8', windowsHide: true }).trim(), source)
const requireRuntime = createRequire(join(candidate, 'runtime', 'package.json'))
const native = requireRuntime('@vscode/windows-process-tree')
const security = requireRuntime('@orca/windows-security')
const sourceRoot = await realpath(join(dirname(import.meta.filename), '..', '..'))
assert.ok(process.env.ProgramFiles && isAbsolute(process.env.ProgramFiles))
const powershell = await realpath(join(process.env.ProgramFiles, 'PowerShell', '7', 'pwsh.exe'))
const processes = () => new Promise((resolve, reject) => native.getAllProcesses(rows => rows ? resolve(rows) : reject(new Error('Actual process identities unavailable')), 4))
const records = []
const modes = ['immediate-eof', 'startup-eof', 'readiness-eof', 'startup-stdout-close']
assert.ok(values.mode === undefined || modes.includes(values.mode))
const selectedModes = values.mode === undefined ? modes : [values.mode]
if (controllerCheck) assert.equal(values.mode, 'readiness-eof', 'Controller checking requires the explicit readiness-eof mode')

async function observeControllerInput(endpoint, data, directory, deadline, observeIdentity) {
  const observation = { status: 'FAIL', ownership_incomplete: false, producer: null, transport: null, transport_process: null, transport_actual_spawned: false }
  let control, controlClosed = false, controlExit, failure
  const output = [], errors = []
  let outputBytes = 0, errorBytes = 0, outputEOF = false, errorEOF = false, streamFailure = false
  const outputParent = join(directory, 'controller-input')
  await mkdir(outputParent)
  try {
    const { runControllerTaskCreateCheck } = await import('./controller-taskcreate-check.mjs')
    const produce = () => runControllerTaskCreateCheck({ candidate, endpoint, dataRoot: data,
      publicParent: join(dataParent, 'controller-input-public'), outputDir: join(outputParent, 'producer'),
      sourceRevision: source, deadline, onProcessIdentity: observeIdentity, ...(values['controller-open-only'] ? { stopAfter: 'controller-open' } : {}) })
    let producer
    if (!values['controller-open-only']) { producer = await produce(); observation.producer = producer.result }
    const transportOutput = join(outputParent, 'transport')
    control = spawn(process.execPath, [join(sourceRoot, 'tests/swa-cli-headless/controller-input-check.mjs'), '--install-root', candidate,
      ...(values['controller-open-only'] ? ['--short-only'] : ['--invocation-file', producer.invocationFile]), '--output-dir', transportOutput], { cwd: candidate, env: { ...process.env }, windowsHide: true, stdio: ['ignore', 'pipe', 'pipe'] })
    observation.transport_actual_spawned = Number.isSafeInteger(control.pid) && control.pid > 0
    const closed = new Promise(resolve => {
      control.once('error', error => { controlClosed = true; controlExit = { exit: null, signal: null, spawn_error: typeof error.code === 'string' ? error.code : 'UNKNOWN' }; resolve(controlExit) })
      control.once('close', (exit, signal) => { controlClosed = true; controlExit = { exit, signal, spawn_error: null }; resolve(controlExit) })
    })
    control.stdout.on('data', chunk => { outputBytes += chunk.length; if (outputBytes <= 64 * 1024) output.push(chunk); else streamFailure = true })
    control.stderr.on('data', chunk => { errorBytes += chunk.length; if (errorBytes <= 256 * 1024) errors.push(chunk); else streamFailure = true })
    control.stdout.on('end', () => { outputEOF = true })
    control.stderr.on('end', () => { errorEOF = true })
    for (const stream of [control.stdout, control.stderr]) stream.on('error', () => { streamFailure = true })
    assert.ok(Number.isSafeInteger(control.pid) && control.pid > 0)
    const created = native.getProcessCreationTime(control.pid)
    assert.ok(Number.isSafeInteger(created) && created > 0, 'Actual transport process identity is unavailable')
    observation.transport_process = { pid: control.pid, creationTimeMs: created }
    observeIdentity(observation.transport_process)
    const transportDeadline = Math.min(deadline, Date.now() + 120_000)
    while (!controlClosed && Date.now() < transportDeadline) await setTimeout(25)
    assert.equal(controlClosed, true, 'Installed helper transport completion remains unknown')
    await closed
    observation.transport_exit = controlExit
    const reportPath = join(transportOutput, 'report.json')
    assert.ok((await lstat(reportPath)).isFile() && (await lstat(reportPath)).size <= 2 * 1024 * 1024)
    const transportReport = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(await readFile(reportPath)))
    const summary = transportReport.summary
    assert.equal(summary?.schema, 'swa.controller-input-check.v1')
    assert.ok(['PASS', 'FAIL'].includes(summary.status)); assert.equal(summary.platform, 'win32')
    const safeTransport = { schema: summary.schema, status: summary.status, platform: summary.platform }
    for (const kind of ['tty', 'pipe']) {
      const row = summary[kind]
      assert.ok(row && ['PASS', 'FAIL'].includes(row.status))
      assert.equal(typeof row.natural_eof, 'boolean'); assert.equal(typeof row.forced_cleanup, 'boolean')
      assert.ok(Number.isSafeInteger(row.reply_count) && row.reply_count >= 0 && row.reply_count <= 3)
      assert.ok(row.long_bytes === null || Number.isSafeInteger(row.long_bytes) && row.long_bytes > 0 && row.long_bytes <= 1024 * 1024)
      assert.ok(row.exit_code === null || Number.isInteger(row.exit_code))
      safeTransport[kind] = { status: row.status, natural_eof: row.natural_eof, forced_cleanup: row.forced_cleanup,
        reply_count: row.reply_count, long_bytes: row.long_bytes, exit_code: row.exit_code }
    }
    assert.equal(summary.rpc_mutations, 'NOT_RUN'); assert.equal(summary.model_execution, 'NOT_RUN'); assert.equal(summary.kernel_job_empty, 'NOT_CLAIMED')
    Object.assign(safeTransport, { rpc_mutations: 'NOT_RUN', model_execution: 'NOT_RUN', kernel_job_empty: 'NOT_CLAIMED' })
    if (values['controller-open-only']) {
      assert.equal(summary.coverage, 'explicit_short_identify_only'); assert.equal(summary.long_input, 'NOT_RUN')
      Object.assign(safeTransport, { coverage: summary.coverage, long_input: summary.long_input })
    }
    observation.transport = safeTransport
    // 这些身份由原 checker 在进程仍可读时取得，不在退出后用裸 PID 补取创建时间。
    for (const identity of [transportReport.tty?.root, transportReport.tty?.helper, transportReport.pipe?.helper]) {
      assert.ok(identity && Number.isSafeInteger(identity.pid) && identity.pid > 0 && Number.isSafeInteger(identity.creationTimeMs) && identity.creationTimeMs > 0)
      observeIdentity({ pid: identity.pid, creationTimeMs: identity.creationTimeMs })
    }
    assert.deepEqual(controlExit, { exit: 0, signal: null, spawn_error: null })
    assert.ok(outputEOF && errorEOF && !streamFailure)
    assert.equal(safeTransport.status, 'PASS')
    for (const kind of ['tty', 'pipe']) {
      const row = safeTransport[kind]
      assert.equal(row.status, 'PASS'); assert.equal(row.natural_eof, true); assert.equal(row.forced_cleanup, false)
      assert.equal(row.reply_count, 3)
      if (values['controller-open-only']) assert.equal(row.long_bytes, null)
      else assert.ok(row.long_bytes > 4095)
      assert.equal(row.exit_code, 0)
    }
    if (values['controller-open-only']) { producer = await produce(); observation.producer = producer.result }
    observation.status = 'PASS'
  } catch (error) {
    failure = error
    if (error.observation) observation.producer = error.observation
    observation.ownership_incomplete ||= observation.producer?.ownership_incomplete === true || observation.transport_actual_spawned && !observation.transport_process
    observation.error = { name: error instanceof Error ? error.name : 'UnknownError', code: typeof error.code === 'string' ? error.code : null }
  } finally {
    if (control && !controlClosed) {
      const identity = observation.transport_process
      let cleanupAttempted = false
      if (identity && native.getProcessCreationTime(identity.pid) === identity.creationTimeMs) cleanupAttempted = control.kill()
      const cleanupDeadline = Date.now() + 5000
      while (!controlClosed && Date.now() < cleanupDeadline) await setTimeout(25)
      observation.transport_cleanup = { attempted: cleanupAttempted, exit: controlClosed ? controlExit : 'UNKNOWN', normal_exit_proven: false }
      if (!controlClosed) { control.stdout.destroy(); control.stderr.destroy(); control.unref() }
    }
    observation.transport_exit ??= controlExit ?? 'NOT_STARTED'
    observation.transport_streams = { stdout_bytes: outputBytes, stderr_bytes: errorBytes, stdout_eof: outputEOF, stderr_eof: errorEOF, failure: streamFailure }
    await writeFile(join(outputParent, 'transport.stdout'), Buffer.concat(output), { flag: 'wx' })
    await writeFile(join(outputParent, 'transport.stderr'), Buffer.concat(errors), { flag: 'wx' })
  }
  if (failure) {
    const error = new Error('The actual controller input check failed; bounded observations retained')
    error.code = 'ERR_SWA_CONTROLLER_INPUT'
    error.observation = observation
    throw error
  }
  return observation
}
async function webAssetInventory(installation) {
  const root = await realpath(join(installation, 'web'))
  const manifestFile = join(root, '.vite', 'manifest.json')
  assert.ok((await lstat(manifestFile)).size <= 1024 * 1024, 'Web manifest exceeds its read bound')
  const { checkWebAssets } = await import(pathToFileURL(join(installation, 'scripts', 'swa-package-config.mjs')).href)
  const closure = await checkWebAssets(root)
  const manifest = JSON.parse(await readFile(manifestFile, 'utf8'))
  const references = new Set(['index.html'])
  for (const entry of Object.values(manifest)) {
    references.add(entry.file)
    for (const field of ['css', 'assets']) for (const path of entry[field] ?? []) references.add(path)
  }
  assert.ok(references.size <= 128, 'Web asset count exceeds its read bound')
  const mimeTypes = { '.html': ['text/html'], '.js': ['text/javascript', 'application/javascript'], '.mjs': ['text/javascript', 'application/javascript'], '.css': ['text/css'], '.svg': ['image/svg+xml'], '.json': ['application/json'], '.png': ['image/png'], '.jpg': ['image/jpeg'], '.jpeg': ['image/jpeg'], '.gif': ['image/gif'], '.webp': ['image/webp'], '.ico': ['image/x-icon', 'image/vnd.microsoft.icon'], '.woff': ['font/woff'], '.woff2': ['font/woff2'] }
  const files = []
  let total = 0
  for (const path of references) {
    assert.ok(path.split('/').every(part => part && part !== '.' && part !== '..'), 'Web HTTP reference has an ambiguous path segment')
    const info = await lstat(join(root, path))
    assert.ok(info.isFile() && info.size <= 16 * 1024 * 1024, 'Web asset exceeds its read bound')
    total += info.size
    assert.ok(total <= 64 * 1024 * 1024, 'Total Web assets exceed their read bound')
    const types = mimeTypes[extname(path).toLowerCase()]
    assert.ok(types, 'Web asset MIME type is not covered by this check')
    const bytes = await readFile(join(root, path))
    assert.equal(bytes.length, info.size)
    files.push({ path, bytes, types, sha256: createHash('sha256').update(bytes).digest('hex') })
  }
  return { root, closure, files }
}

async function observeWebAssets(endpoint, deadline, output) {
  const observation = { source_revision: source, edition, endpoint, status: 'NOT_RUN', requests: [] }
  async function request(path, maximum) {
    const remaining = deadline - Date.now()
    assert.ok(remaining > 0, 'Web HTTP check reached the original service deadline')
    const response = await fetch(new URL(path, endpoint + '/'), { redirect: 'manual', signal: AbortSignal.timeout(Math.min(5000, remaining)) })
    const chunks = [], reader = response.body?.getReader()
    let size = 0, complete = false
    try {
      if (reader) for (;;) {
        const part = await reader.read()
        if (part.done) break
        size += part.value.byteLength
        assert.ok(size <= maximum, 'Web HTTP response exceeds its actual installed file bound')
        chunks.push(Buffer.from(part.value))
      }
      complete = true
    } finally {
      if (reader) {
        if (!complete) await reader.cancel()
        reader.releaseLock()
      }
    }
    return { response, bytes: Buffer.concat(chunks) }
  }
  try {
    const inventory = await webAssetInventory(candidate)
    observation.manifest = inventory.closure
    const redirect = await request('/index.html', 64 * 1024)
    observation.requests.push({ path: '/index.html', status: redirect.response.status, location: redirect.response.headers.get('location') })
    assert.equal(redirect.response.status, 301)
    assert.equal(redirect.response.headers.get('location'), './')
    assert.equal(new URL(redirect.response.headers.get('location'), endpoint + '/index.html').href, endpoint + '/')
    for (const file of inventory.files) {
      const path = file.path === 'index.html' ? '/' : '/' + file.path.split('/').map(encodeURIComponent).join('/')
      const actual = await request(path, file.bytes.length)
      const contentType = actual.response.headers.get('content-type')?.split(';')[0].trim().toLowerCase() ?? null
      const sha256 = createHash('sha256').update(actual.bytes).digest('hex')
      observation.requests.push({ path, status: actual.response.status, content_type: contentType, byte_length: actual.bytes.length, sha256, installed_sha256: file.sha256 })
      assert.equal(actual.response.status, 200)
      assert.ok(file.types.includes(contentType), 'Web HTTP MIME differs from the actual asset type')
      assert.equal(sha256, file.sha256)
      assert.ok(actual.bytes.equals(file.bytes), 'Web HTTP content differs from the installed asset')
    }
    observation.status = 'PASS'
  } catch (error) {
    observation.status = 'FAIL'
    observation.error = { name: error instanceof Error ? error.name : 'UnknownError', code: typeof error?.code === 'string' ? error.code : null }
    throw error
  } finally {
    await writeFile(output, JSON.stringify(observation, null, 2) + '\n', { flag: 'wx' })
  }
}

async function observeReadyService(ready, child, data, created, deadline) {
  const actualData = await realpath(data), endpoint = `http://127.0.0.1:${Number(values.port)}`
  const samePath = (actual, expected) => {
    assert.equal(typeof actual, 'string')
    assert.equal(actual.toLowerCase(), expected.toLowerCase())
  }
  samePath(ready.install_root, candidate)
  samePath(ready.data_root, actualData)
  samePath(ready.readiness?.dataRoot, actualData)
  samePath(ready.readiness?.node, join(candidate, 'node', 'node.exe'))
  assert.match(ready.readiness.runtimeId, /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/)
  assert.equal(ready.http, endpoint)
  const remaining = deadline - Date.now()
  assert.ok(remaining > 0, 'Health check reached the original service deadline')
  const response = await fetch(endpoint + '/api/v1/health', { redirect: 'error', signal: AbortSignal.timeout(Math.min(5000, remaining)) })
  assert.equal(response.status, 200)
  const health = await response.json()
  assert.deepEqual(health, { schema_version: 1, status: 'ok', service: 'astrocyte' })
  const metadata = JSON.parse(await readFile(join(actualData, 'orca-runtime.json'), 'utf8'))
  assert.equal(metadata.runtimeId, ready.readiness.runtimeId)
  assert.equal(metadata.pid, ready.runtime_pid)
  const daemon = JSON.parse(await readFile(join(actualData, 'daemon', 'daemon-v49.pid'), 'utf8'))
  const job = daemon.swa_service_job
  assert.match(job.name, /^Local\\SWAService[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/)
  assert.equal(job.platform, undefined)
  assert.equal(job.source_revision, source)
  samePath(job.data_root, actualData)
  samePath(job.install_root, candidate)
  assert.equal(job.serve_pid, child.pid)
  assert.ok(Number.isSafeInteger(created) && created > 0)
  assert.equal(job.serve_created_at_ms, created)
  assert.equal(job.daemon_pid, daemon.pid)
  assert.ok(Number.isSafeInteger(job.daemon_created_at_ms) && job.daemon_created_at_ms > 0)
  assert.match(job.daemon_session, /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/)
  assert.equal(job.daemon_session, daemon.launchNonce)
  const expected = [
    { role: 'serve', pid: child.pid, created, name: 'swa.exe' },
    { role: 'runtime', pid: ready.runtime_pid, name: 'node.exe' },
    { role: 'server', pid: ready.server_pid, name: 'server.exe' },
    { role: 'daemon', pid: job.daemon_pid, created: job.daemon_created_at_ms, name: 'node.exe' }
  ]
  assert.equal(new Set(expected.map(member => member.pid)).size, expected.length)
  const rows = await processes(), user = security.currentUserSid(), members = []
  for (const member of expected) {
    assert.ok(Number.isSafeInteger(member.pid) && member.pid > 0)
    const actual = security.readServiceJobMember(job.name, member.pid)
    assert.equal(actual.name, job.name)
    assert.equal(actual.pid, member.pid)
    assert.ok(Number.isSafeInteger(actual.creationTimeMs) && actual.creationTimeMs > 0)
    if (member.created !== undefined) assert.equal(actual.creationTimeMs, member.created)
    assert.equal(actual.ownerSid, user)
    assert.equal(actual.protected, true)
    const trustees = new Set([user, 'S-1-5-18'])
    assert.equal(actual.aces.length, trustees.size)
    for (const ace of actual.aces) {
      assert.equal(ace.flags, 0)
      assert.equal(ace.mask, 0x1f003f)
      assert.equal(trustees.delete(ace.sid), true)
    }
    assert.equal(trustees.size, 0)
    const process = rows.find(row => row.pid === member.pid)
    assert.ok(process)
    assert.equal(process.creationTimeMs, actual.creationTimeMs)
    assert.equal(process.name.toLowerCase(), member.name)
    if (member.role === 'runtime' || member.role === 'server') assert.equal(process.ppid, child.pid)
    members.push({ role: member.role, pid: member.pid, creation_time_ms: actual.creationTimeMs, name: process.name, parent_pid: process.ppid, actual_named_job_membership: true })
  }
  // 只保存核验所需字段，daemon PID 原记录中的 nonce 不进入证据。
  return { source_revision: source, runtime_id: ready.readiness.runtimeId, data_root: actualData, health, service_job_name: job.name, members, job_owner_and_dacl_verified: true, observed_at: new Date().toISOString() }
}
for (const mode of selectedModes) {
  const directory = join(evidence, mode)
  await mkdir(directory)
  const child = spawn(join(candidate, 'swa.exe'), ['serve', '--data-dir', join(dataParent, mode), '--port', values.port, '--protected-roots', values['protected-roots'], '--json', ...(edition === 'combined' ? ['--web-dir', join(candidate, 'web')] : [])], { windowsHide: true, stdio: ['pipe', 'pipe', 'pipe'], env: { ...process.env } })
  const stdout = [], stderr = [], observed = new Map(), events = [], unknown = [], additionalRoots = new Map()
  const rootCreationTimeMs = native.getProcessCreationTime(child.pid)
  observed.set(child.pid + ':' + rootCreationTimeMs, { pid: child.pid, creationTimeMs: rootCreationTimeMs, launched_executable: join(candidate, 'swa.exe'), parent_pid: process.pid })
  let settled = null, spawnError = null, streamError = null, closed = false, acted = false, readiness = null, readinessFailure = null
  let readinessStarted = false, readinessComplete = false, readinessWork, controllerInput = null
  let closureIncomplete = !Number.isSafeInteger(rootCreationTimeMs) || rootCreationTimeMs <= 0
  const exited = new Promise(resolve => {
    child.once('error', error => { spawnError = error })
    child.once('close', (exit, signal) => { closed = true; settled = { exit, signal, observed_at_ms: Date.now() }; resolve(settled) })
  })
  child.stdout.on('data', chunk => { stdout.push(chunk); assert.ok(Buffer.concat(stdout).length < 16 * 1024 * 1024) })
  child.stdout.on('error', error => { streamError = error.message })
  child.stderr.on('data', chunk => { stderr.push(chunk); assert.ok(Buffer.concat(stderr).length < 16 * 1024 * 1024) })
  child.stdin.on('error', error => { events.push({ type: 'stdin-error', code: error.code, observed_at_ms: Date.now() }) })
  events.push({ type: 'actual-spawn', pid: child.pid, creation_time_ms: rootCreationTimeMs, observed_at_ms: Date.now() })
  if (mode === 'immediate-eof') {
    child.stdin.end()
    acted = true
    events.push({ type: 'actual-owned-stdin-closed', observed_at_ms: Date.now() })
  }
  await writeFile(join(directory, 'actual-spawn.json'), JSON.stringify({ candidate, pid: child.pid, creation_time_ms: rootCreationTimeMs, mode, events }, null, 2) + '\n', { flag: 'wx' })
  const deadline = Date.now() + (controllerCheck ? 600_000 : 150_000)
  while (!closed && Date.now() < deadline) {
    const rows = await processes()
    const { rootVerified, ownedRows, excludedEdges, unknownRows } = collectOwnedIncarnations(rows, { pid: child.pid, creationTimeMs: rootCreationTimeMs })
    for (const anchor of additionalRoots.values()) {
      const extra = collectOwnedIncarnations(rows, anchor)
      for (const row of extra.ownedRows) if (!ownedRows.some(value => value.pid === row.pid && value.creationTimeMs === row.creationTimeMs)) ownedRows.push(row)
      excludedEdges.push(...extra.excludedEdges)
      unknownRows.push(...extra.unknownRows)
    }
    // Once provenance is incomplete, disappearance of its first unknown PID
    // cannot prove that unobserved descendants have also exited.
    closureIncomplete ||= unknownRows.length > 0
    for (const row of ownedRows) observed.set(row.pid + ':' + row.creationTimeMs, row)
    if (unknownRows.length) {
      for (const entry of unknownRows) unknown.push({ observed_at_ms: Date.now(), ...entry })
      await writeFile(join(directory, 'unknown-process-observations.json'), JSON.stringify(unknown, null, 2) + '\n')
    }
    await writeFile(join(directory, 'latest-actual-observation.json'), JSON.stringify({ mode, pid: child.pid, creation_time_ms: rootCreationTimeMs, events, current_owned_processes: ownedRows, excluded_parent_edges: excludedEdges, ownership_closure_incomplete: closureIncomplete, observed_processes: [...observed.values()], unknown_process_observations: unknown, stdout: Buffer.concat(stdout).toString('utf8'), stderr: Buffer.concat(stderr).toString('utf8'), observed_at_ms: Date.now() }, null, 2) + '\n')
    const currentStdout = Buffer.concat(stdout).toString('utf8')
    if (!readiness && currentStdout.includes('\n')) {
      readiness = JSON.parse(currentStdout)
      events.push({ type: 'actual-stdout-envelope', ok: readiness.ok, observed_at_ms: Date.now() })
    }
    if (!readinessStarted && mode === 'readiness-eof' && readiness?.ok) {
      readinessStarted = true
      readinessWork = (async () => { try {
        const observation = await observeReadyService(readiness, child, join(dataParent, mode), rootCreationTimeMs, deadline)
        for (const member of observation.members) {
          observed.set(member.pid + ':' + member.creation_time_ms, { pid: member.pid, creationTimeMs: member.creation_time_ms,
            ppid: member.parent_pid, name: member.name, role: member.role, actual_named_job_membership: true })
        }
        await writeFile(join(directory, 'ready-service-observation.json'), JSON.stringify(observation, null, 2) + '\n', { flag: 'wx' })
        if (edition === 'combined') await observeWebAssets(`http://127.0.0.1:${Number(values.port)}`, deadline, join(directory, 'web-http-observation.json'))
        if (controllerCheck) controllerInput = await observeControllerInput(`http://127.0.0.1:${Number(values.port)}`, join(dataParent, mode), directory, deadline, identity => {
          assert.ok(Number.isSafeInteger(identity.pid) && identity.pid > 0 && Number.isSafeInteger(identity.creationTimeMs) && identity.creationTimeMs > 0)
          const key = identity.pid + ':' + identity.creationTimeMs
          observed.set(key, identity)
          additionalRoots.set(key, identity)
        })
      } catch (error) {
        readinessFailure = error
        if (controllerCheck && error.observation) controllerInput = error.observation
        if (controllerCheck && (controllerInput?.ownership_incomplete || controllerInput?.producer?.ownership_incomplete)) closureIncomplete = true
      } finally {
        readinessComplete = true
      } })()
      if (!controllerCheck) await readinessWork
    }
    if (!acted && mode === 'readiness-eof' && readinessComplete) {
      child.stdin.end(); acted = true
      events.push({ type: 'actual-owned-stdin-closed', observed_at_ms: Date.now() })
    }
    if (!acted && rootVerified && mode.startsWith('startup-') && ownedRows.some(row => row.ppid === child.pid && row.name.toLowerCase() === 'node.exe')) {
      const noPublishedReadiness = !currentStdout.length
      events.push({ type: 'actual-owned-node-observed', no_published_readiness: noPublishedReadiness, observed_at_ms: Date.now() })
      if (mode === 'startup-eof') child.stdin.end()
      else child.stdout.destroy()
      acted = true
      events.push({ type: mode === 'startup-eof' ? 'actual-owned-stdin-closed' : 'actual-owned-stdout-read-end-closed', observed_at_ms: Date.now() })
    }
    if (!closed) await setTimeout(20)
  }
  if (readinessWork) await readinessWork
  await writeFile(join(directory, 'stdout.json'), Buffer.concat(stdout), { flag: 'wx' })
  await writeFile(join(directory, 'stderr.log'), Buffer.concat(stderr), { flag: 'wx' })
  if (!closed) {
    await writeFile(join(directory, 'unsettled.json'), JSON.stringify({ pid: child.pid, observed: [...observed.values()], events }, null, 2) + '\n', { flag: 'wx' })
    throw new Error('Actual owned serve remains unsettled; preserve its recorded identity')
  }
  await exited
  const observations = []
  for (let check = 0; check < 2; check++) {
    const rows = await processes()
    const remaining = [...observed.values()].filter(old => rows.some(row => row.pid === old.pid && (!(row.creationTimeMs > 0) || row.creationTimeMs === old.creationTimeMs)))
    observations.push({ observed_at_ms: Date.now(), remaining })
    if (check === 0) await setTimeout(250)
  }
  const cimRequest = join(directory, 'cim-request.json'), cimOutput = join(directory, 'cim-exit-observations.json')
  await writeFile(cimRequest, JSON.stringify({ schema_version: 1, known: [...observed.values()], unknown: unknown.map(entry => entry.actual) }, null, 2) + '\n', { flag: 'wx' })
  const cim = spawn(powershell, ['-NoProfile', '-File', join(sourceRoot, 'tests/swa-cli-headless/observe-process-exits.ps1'), '-RequestFile', cimRequest, '-OutputFile', cimOutput], { windowsHide: true, stdio: ['ignore', 'pipe', 'pipe'] })
  const cimStdout = [], cimStderr = []
  cim.stdout.on('data', chunk => cimStdout.push(chunk))
  cim.stderr.on('data', chunk => cimStderr.push(chunk))
  let cimSettled = false
  const cimClosed = new Promise(resolve => {
    cim.once('error', error => { cimSettled = true; resolve({ pid: cim.pid, spawn_error: String(error), exit: null, signal: null }) })
    cim.once('close', (exit, signal) => { cimSettled = true; resolve({ pid: cim.pid, exit, signal }) })
  })
  const cimDeadline = Date.now() + 15_000
  while (!cimSettled && Date.now() < cimDeadline) await setTimeout(50)
  const cimTimedOut = !cimSettled
  const cimTermination = cimTimedOut ? { attempted_at: new Date().toISOString(), actual_created_pid: cim.pid, signal_sent: cim.kill() } : null
  const cimCleanupDeadline = Date.now() + 5_000
  while (!cimSettled && Date.now() < cimCleanupDeadline) await setTimeout(50)
  const cimExit = { ...(cimSettled ? await cimClosed : { pid: cim.pid, exit: null, signal: null, completion_unknown: true }), query_timed_out: cimTimedOut, query_deadline_ms: 15_000, actual_helper_termination: cimTermination }
  if (!cimSettled) { cim.stdout.destroy(); cim.stderr.destroy(); cim.unref() }
  await writeFile(join(directory, 'cim.stdout.log'), Buffer.concat(cimStdout), { flag: 'wx' })
  await writeFile(join(directory, 'cim.stderr.log'), Buffer.concat(cimStderr), { flag: 'wx' })
  await writeFile(join(directory, 'cim-process-exit.json'), JSON.stringify(cimExit, null, 2) + '\n', { flag: 'wx' })
  let cimReport = null, cimError = null
  if ((await readdir(directory)).includes('cim-exit-observations.json')) {
    try { cimReport = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(await readFile(cimOutput))) }
    catch (error) { cimError = String(error) }
  }
  const completeCIM = !cimTimedOut && cimExit.exit === 0 && cimExit.signal === null && cimError === null && cimReport?.schema_version === 1 && cimReport.query_failure === null && cimReport.all_exited === true && Array.isArray(cimReport.observations) && cimReport.observations.length === 2 && cimReport.observations.every(row => row.query_succeeded === true && row.complete_result_count > 0 && row.required_identity_fields_complete === true && row.known_identities_exited === true && row.unknown_pids_absent === true) && cimReport.observations[1].monotonic_ms - cimReport.observations[0].monotonic_ms >= 50
  const identitiesExited = !closureIncomplete && observations.every(row => row.remaining.length === 0) && completeCIM
  const actualErrorText = Buffer.concat(stderr).toString('utf8')
  const startupHit = !mode.startsWith('startup-') || events.some(event => event.type === 'actual-owned-node-observed' && event.no_published_readiness)
  const record = { mode, edition, pid: child.pid, ...settled, events, actual_spawn_error: spawnError?.message ?? null, actual_stream_error: streamError, readiness_check_error: readinessFailure ? { name: readinessFailure instanceof Error ? readinessFailure.name : 'UnknownError', code: typeof readinessFailure?.code === 'string' ? readinessFailure.code : null } : null, ...(controllerCheck ? { controller_input_check: controllerInput } : {}), ownership_closure_incomplete: closureIncomplete, observed_processes: [...observed.values()], unknown_process_observations: unknown, exit_observations: observations, actual_cim_process_exit: cimExit, actual_cim_report: cimReport, actual_cim_error: cimError, status: !identitiesExited ? 'UNKNOWN' : startupHit ? 'AWAITING_ASSERTIONS' : 'INCONCLUSIVE' }
  await writeFile(join(directory, 'actual-observation.json'), JSON.stringify(record, null, 2) + '\n', { flag: 'wx' })
  if (!identitiesExited) {
    records.push(record)
    await writeFile(join(directory, 'result.json'), JSON.stringify(record, null, 2) + '\n', { flag: 'wx' })
    process.exitCode = 6
    break
  }
  if (readinessFailure) {
    record.status = 'FAIL'
    records.push(record)
    await writeFile(join(directory, 'result.json'), JSON.stringify(record, null, 2) + '\n', { flag: 'wx' })
    console.error(readinessFailure)
    process.exitCode = 1
    break
  }
  if (spawnError) throw spawnError
  assert.equal(settled.signal, null)
  assert.equal(acted, true)
  const beforeReady = mode === 'immediate-eof' || mode === 'startup-eof'
  if (beforeReady && startupHit) {
    assert.ok([130, 6].includes(settled.exit))
    const reply = JSON.parse(Buffer.concat(stdout).toString('utf8'))
    assert.equal(reply.ok, false)
    assert.equal(reply.error.code, 'delivery_unknown')
  } else if (mode === 'readiness-eof') {
    assert.equal(settled.exit, 0)
    assert.equal(JSON.parse(Buffer.concat(stdout).toString('utf8')).ok, true)
  } else if (mode === 'startup-stdout-close' && startupHit) {
    assert.equal(settled.exit, 6)
    assert.match(actualErrorText, /readiness|stdout|输出|R=/)
    assert.equal(actualErrorText.match(/^已知关联：.*$/gm)?.length, 1)
    assert.match(actualErrorText, /^已知关联：runtime_id="[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}"\r?$/m)
  }
  if (mode !== 'immediate-eof') {
    assert.match(actualErrorText, /active=1 consecutive=1/)
    assert.match(actualErrorText, /active=1 consecutive=2/)
  }
  record.status = startupHit ? 'OBSERVED' : 'INCONCLUSIVE'
  records.push(record)
  await writeFile(join(directory, 'result.json'), JSON.stringify(records.at(-1), null, 2) + '\n', { flag: 'wx' })
}
await writeFile(join(evidence, 'result.json'), JSON.stringify({ candidate, source_revision: source, edition, selected_modes: selectedModes, unselected_modes: modes.filter(mode => !selectedModes.includes(mode)), records, model_calls: 0, human_session_created: records.some(row => row.controller_input_check?.producer?.human_session_created), project_status_check: 'NOT_RUN', controller_check: controllerCheck ? records.some(row => row.controller_input_check?.status === 'PASS') ? 'PASS' : 'FAIL' : 'NOT_RUN', remainingCases: selectedModes.filter(mode => !records.some(row => row.mode === mode && row.status === 'OBSERVED')) }, null, 2) + '\n', { flag: 'wx' })
if (records.some(row => row.status !== 'OBSERVED') && !process.exitCode) process.exitCode = 6
