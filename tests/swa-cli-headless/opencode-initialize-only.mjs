// 同 S 受控初始化客户端；只准备真实账户与 ACP 会话，不发送模型输入。
import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import { createHash, randomBytes, randomUUID } from 'node:crypto'
import { constants } from 'node:fs'
import { lstat, mkdir, open, readFile, realpath } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { dirname, isAbsolute, join, relative, resolve } from 'node:path'
import { DatabaseSync } from 'node:sqlite'
import { fileURLToPath } from 'node:url'
import { parseArgs } from 'node:util'
import { publicFiles, publicRevision, publicSource } from '../swa-product-acceptance/public-case-definition.mjs'

process.umask(0o077)
assert.equal(process.umask(), 0o077)
const hash = bytes => createHash('sha256').update(bytes).digest('hex')
const inside = (root, path) => { const part = relative(root, path); return !part || !isAbsolute(part) && part !== '..' && !part.startsWith('../') }
const stamp = info => `${info.dev}:${info.ino}:${info.birthtimeNs}:${info.mtimeNs}:${info.size}`
const safeCode = error => error instanceof Error && /^(?:native_opencode|private_opencode|opencode_initialization)_[a-z0-9_]+$/.test(error.message) ? error.message : 'opencode_initialization_check_failed'
async function ordinary(path, directory = false, privatePath = false) {
  assert.ok(isAbsolute(path) && resolve(path) === path)
  assert.equal(await realpath(path), path)
  for (let current = path; ; current = dirname(current)) {
    const info = await lstat(current)
    assert.ok(!info.isSymbolicLink() && (current === path && !directory ? info.isFile() : info.isDirectory()))
    if (dirname(current) === current) break
  }
  const info = await lstat(path, { bigint: true })
  if (privatePath) {
    assert.equal(info.uid, BigInt(process.getuid())); assert.equal(info.mode & 0o7777n, directory ? 0o700n : 0o600n)
    if (!directory) assert.equal(info.nlink, 1n)
  }
  return info
}
async function publicBytes(path, maximum = 8 * 1024 * 1024) {
  const before = await ordinary(path)
  assert.ok(before.size >= 0n && before.size <= BigInt(maximum))
  const file = await open(path, constants.O_RDONLY | constants.O_NOFOLLOW)
  try {
    assert.equal(stamp(await file.stat({ bigint: true })), stamp(before))
    const buffer = Buffer.alloc(Number(before.size) + 1)
    let count = 0
    while (count < buffer.length) { const { bytesRead } = await file.read(buffer, count, buffer.length - count, null); if (!bytesRead) break; count += bytesRead }
    assert.equal(count, Number(before.size)); assert.equal(stamp(await file.stat({ bigint: true })), stamp(before))
    assert.equal(stamp(await ordinary(path)), stamp(before))
    return buffer.subarray(0, count)
  } finally { await file.close() }
}
async function persist(path, value) {
  const file = await open(path, constants.O_WRONLY | constants.O_CREAT | constants.O_EXCL | constants.O_NOFOLLOW, 0o600)
  try { await file.writeFile(JSON.stringify(value, null, 2) + '\n'); await file.sync() } finally { await file.close() }
}
const timeout = (promise, milliseconds) => {
  let timer
  return Promise.race([promise, new Promise((_, reject) => { timer = setTimeout(() => reject(new Error('opencode_initialization_timeout')), milliseconds) })]).finally(() => clearTimeout(timer))
}
async function main() {
  const values = parseArgs({ options: Object.fromEntries(['candidate', 'source-repository', 'source-home', 'executable', 'public-project', 'project-id', 'output', 'data-root'].map(key => [key, { type: 'string' }])), allowPositionals: false }).values
  for (const key of ['candidate', 'source-repository', 'source-home', 'executable', 'public-project', 'output', 'data-root']) assert.ok(isAbsolute(values[key] ?? ''))
  assert.equal(process.platform, 'linux'); assert.equal(process.arch, 'x64')
  assert.match(values['project-id'], /^[a-f0-9]{8}(?:-[a-f0-9]{4}){3}-[a-f0-9]{12}$/i)
  const candidate = await realpath(values.candidate), repository = await realpath(values['source-repository']), project = await realpath(values['public-project']), sourceHome = await realpath(values['source-home'])
  const output = resolve(values.output), dataRoot = resolve(values['data-root']), executable = await realpath(values.executable)
  for (const root of [candidate, repository, project]) { await ordinary(root, true); for (const privateRoot of [sourceHome, dataRoot, output]) assert.ok(!inside(root, privateRoot) && !inside(privateRoot, root)) }
  for (const [left, right] of [[sourceHome, dataRoot], [sourceHome, output], [dataRoot, output]]) assert.ok(!inside(left, right) && !inside(right, left))
  await ordinary(sourceHome, true, true); await ordinary(executable)
  for (const path of [output, dataRoot]) await ordinary(dirname(path), true)
  assert.equal(await realpath(process.execPath), join(candidate, 'node/node'))
  assert.equal(process.env.NODE_PATH, join(candidate, 'runtime/node_modules')); assert.equal(process.env.NODE_OPTIONS, undefined)
  const revision = (await publicBytes(join(candidate, 'VERSION'), 4096)).toString('ascii').trim()
  assert.match(revision, /^[a-f0-9]{40}$/)
  const gitOptions = { cwd: repository, env: { PATH: process.env.PATH, LANG: 'C.UTF-8', GIT_OPTIONAL_LOCKS: '0', GIT_PAGER: 'cat' }, timeout: 30_000, maxBuffer: 8 * 1024 * 1024 }
  assert.equal(execFileSync('git', ['rev-parse', '--verify', `${revision}^{commit}`], gitOptions).toString('ascii').trim(), revision)
  for (const path of ['tests/swa-cli-headless/opencode-initialize-only.mjs', 'tests/swa-product-acceptance/public-case-definition.mjs', 'tests/swa-product-acceptance/prepare-source-observer.mjs', 'tests/swa-product-acceptance/extract-observer-source.py', 'tests/swa-product-acceptance/opencode-initialization-entry.ts']) assert.ok((await publicBytes(join(repository, path))).equals(execFileSync('git', ['show', `${revision}:${path}`], gitOptions)))
  assert.equal(fileURLToPath(import.meta.url), join(repository, 'tests/swa-cli-headless/opencode-initialize-only.mjs'))
  const publicGit = args => execFileSync('git', ['-c', 'core.fsmonitor=false', '-c', 'core.hooksPath=' + join(dirname(project), 'empty-hooks'), '-C', project, ...args], { ...gitOptions, cwd: project })
  assert.equal(publicGit(['rev-parse', 'HEAD']).toString('ascii').trim(), publicRevision)
  assert.deepEqual(publicGit(['ls-tree', '--name-only', publicRevision]).toString('utf8').trim().split('\n').sort(), [...publicFiles].sort())
  for (const name of publicFiles) assert.ok((await publicBytes(join(project, name), 1024 * 1024)).equals(publicGit(['show', `${publicRevision}:${name}`])))
  await mkdir(output, { mode: 0o700 }); await mkdir(dataRoot, { mode: 0o700 })
  const report = { status: 'NOT_RUN', source_revision: revision, candidate, public_source: publicSource, public_revision: publicRevision, project, data_root: dataRoot, source_home: sourceHome, orca_dispatch_created: false, prompt_calls: 0, observed_model_requests: 0, source_account_inference: 'NOT_RUN', phase: 'prepare-client', error: null, close_error: null, custody_error: null, database_close_error: null, report_error: null }
  let snapshot, resource, tree, managed, client, primary, streamsClosed, closing = false
  let stdoutEnded = false, stderrEnded = false, protocolComplete = false, otherStderrLines = 0, sequence = 0, policyActive = false, server
  let rejectStreams
  const streamErrors = new Promise((_, reject) => { rejectStreams = reject }); void streamErrors.catch(() => undefined)
  const fail = code => { primary ??= new Error(code); rejectStreams(primary) }
  const frames = [], statuses = []; let frameBytes = 0
  try {
    const preparedRoot = join(output, 'initialization-client')
    execFileSync(process.execPath, [join(repository, 'tests/swa-product-acceptance/prepare-source-observer.mjs'), '--candidate', candidate, '--source-repository', repository, '--output', preparedRoot, '--profile', 'opencode-initialize-only'], { timeout: 120_000, maxBuffer: 8192, env: { PATH: process.env.PATH, LANG: 'C.UTF-8', HOME: output, TMPDIR: output, NODE_PATH: process.env.NODE_PATH } })
    const manifest = JSON.parse((await publicBytes(join(preparedRoot, 'manifest.json'), 8 * 1024 * 1024)).toString('utf8'))
    assert.equal(manifest.profile, 'opencode-initialize-only'); assert.equal(manifest.installed_revision, revision)
    const verifyInputs = async () => {
      for (const expected of [...manifest.inputs, ...manifest.dynamic_dependencies, ...manifest.installed_external_entries, manifest.process_host_package, manifest.observer]) {
        assert.ok(inside(preparedRoot, expected.path) || inside(join(candidate, 'runtime'), expected.path))
        const info = await ordinary(expected.path), bytes = await publicBytes(expected.path)
        assert.equal(stamp(info), expected.version); assert.equal(hash(bytes), expected.sha256)
      }
    }
    await verifyInputs()
    const load = createRequire(manifest.observer.path)
    const pty = await realpath(load.resolve('node-pty/lib/utils')), security = await realpath(load.resolve('@orca/windows-security'))
    assert.ok(inside(await realpath(join(candidate, 'runtime/node_modules/node-pty')), pty))
    assert.ok(inside(await realpath(join(candidate, 'runtime/node_modules/@orca/windows-security')), security))
    const api = load(manifest.observer.path), { z } = createRequire(join(candidate, 'runtime/package.json'))('zod')
    const common = { type: z.literal('swa_opencode_policy'), correlation: z.string(), pid: z.number().int().positive(), root: z.string(), sequence: z.number().int().positive() }
    const statusSchema = z.discriminatedUnion('state', [z.strictObject({ ...common, state: z.literal('active'), tools_count: z.literal(0), directory: z.string(), roots: z.record(z.string(), z.string()), plugins: z.array(z.strictObject({ id: z.string(), state: z.literal('active') })) }), z.strictObject({ ...common, state: z.literal('inactive') })])
    const plugin = await realpath(join(candidate, 'runtime/opencode-policy')), correlation = randomBytes(32).toString('hex')
    const inherited = { PATH: process.env.PATH ?? '/usr/bin:/bin', LANG: 'C.UTF-8', NODE_PATH: process.env.NODE_PATH, SWA_INSTALL_ROOT: candidate, SWA_DATA_DIR: dataRoot, SWA_OPENCODE_API_KEY_HOME: sourceHome }
    process.env.SWA_INSTALL_ROOT = candidate; process.env.SWA_DATA_DIR = dataRoot; process.env.SWA_OPENCODE_API_KEY_HOME = sourceHome
    report.phase = 'inspect-source'
    snapshot = await api.inspectNativeOpenCodePolicy({ executable, version: '2.0.31', sourceHome, projectId: values['project-id'], privateRoot: join(dataRoot, 'opencode', values['project-id']), env: inherited, protectedRoots: [candidate, repository, project, dataRoot] })
    assert.equal(snapshot.policy.account.kind, 'opencode-native-stepfun-api-key')
    report.phase = 'materialize-private-account'
    const root = await api.materializePrivateOpenCodeAccount(snapshot, randomUUID(), plugin, correlation)
    report.private_root = root
    const env = api.privateOpenCodeEnvironment(root, inherited)
    report.phase = 'import-original-credential'
    report.import_result = await api.importPrivateOpenCodeCredential({ snapshot, root, plugin, correlation, node: process.execPath, entry: join(candidate, 'runtime/native-cli-entry.cjs'), executable, cwd: project, inherited, observe: observed => { resource = observed } })
    assert.ok(resource); report.import_custody = resource.tree.readFinalObservation()
    assert.equal(report.import_custody.acknowledgement_delivered, true); assert.equal(report.import_custody.observation.fullTreeProof.forced, false)
    await api.verifyPrivateOpenCodeAccount(snapshot, root, plugin, correlation)
    report.phase = 'spawn-acp'
    tree = new api.SwaProviderProcessTree()
    managed = api.spawnManagedProviderProcess({ command: process.execPath, args: [join(candidate, 'runtime/native-cli-entry.cjs'), JSON.stringify({ command: executable, args: ['acp'] })], cwd: project, env, envToDelete: Object.keys(process.env).filter(name => !(name in env)), swaJobIPC: true }, {
      site: 'opencode-initialize-only', acceptClose: result => result.root === 'exited' && (result.tree === 'exited' || managed.processless), stderrLineMaxChars: 32768,
      onStderrLine(line, cut) {
        if (!line.startsWith('SWA_OPENCODE_POLICY ')) { otherStderrLines++; return true }
        try {
          assert.equal(cut, false)
          const parsed = JSON.parse(line.slice('SWA_OPENCODE_POLICY '.length))
          if (parsed?.state === 'http_request') report.observed_model_requests++
          const row = statusSchema.parse(parsed)
          assert.equal(row.correlation, correlation); assert.equal(row.root, root); assert.equal(row.sequence, ++sequence)
          const observed = row.state === 'inactive' && server ? server : tree.requireOwnedExecutable(row.pid, executable)
          if (server) assert.deepEqual(observed, server); else server = observed
          assert.equal(row.pid, server.pid)
          if (row.state === 'active') {
            assert.equal(policyActive, false); assert.equal(row.directory, project)
            assert.deepEqual(row.plugins.map(plugin => plugin.id).sort(), [...api.openCodePlugins('stepfun'), api.openCodePolicyPluginId].sort())
            assert.deepEqual(row.roots, { home: env.HOME, data: env.XDG_DATA_HOME, cache: env.XDG_CACHE_HOME, config: env.XDG_CONFIG_HOME, state: env.XDG_STATE_HOME, programdata: env.ProgramData, temp: env.TEMP, database: env.OPENCODE_DB })
            policyActive = true
          } else policyActive = false
          const { correlation: _secret, ...publicRow } = row
          assert.ok(statuses.length < 16); statuses.push({ row: publicRow, server })
        } catch { fail('opencode_initialization_policy_failed') }
        return true
      }
    })
    const streamError = () => fail('opencode_initialization_stream_failed')
    managed.child.stdin.on('error', streamError); managed.child.stdout.on('error', streamError); managed.child.stderr.on('error', streamError); managed.child.once('error', streamError)
    managed.child.stdout.once('end', () => { stdoutEnded = true }); managed.child.stderr.once('end', () => { stderrEnded = true })
    streamsClosed = new Promise(resolveClosed => managed.child.once('close', resolveClosed))
    tree.bind(managed.child, () => ({ observed: managed.rootExitObserved, processless: managed.processless }), true)
    client = api.createOpenCodeInitializationClient(managed.child.stdout, managed.child.stdin, {
      onRecord(frame, raw_line) {
        try {
          const serialized = JSON.stringify({ frame, raw_line }); frameBytes += Buffer.byteLength(serialized)
          assert.ok(frameBytes <= 12 * 1024 * 1024 && frames.length < 100_000 && !serialized.includes(snapshot.key))
          frames.push({ frame, raw_line })
        } catch { fail('opencode_initialization_frame_bound_or_secret') }
      },
      onProtocolFailure() { fail('opencode_initialization_protocol_failed') }, onInputEnd() { protocolComplete = true },
      onClose() { if (!closing) fail('opencode_initialization_peer_closed_early') }, onDiagnostic() { fail('opencode_initialization_protocol_diagnostic') }
    })
    await timeout(Promise.race([tree.ready(), streamErrors]), 30_000)
    report.phase = 'initialize-session'
    report.acp = await timeout(Promise.race([client.initializeAndStart(project), streamErrors]), 60_000)
    assert.ok(policyActive && server)
    assert.deepEqual(tree.requireOwnedExecutable(server.pid, executable), server)
    report.phase = 'normal-close'
    closing = true; client.beginClose()
    let endError
    try { managed.child.stdin.end() } catch { endError = new Error('opencode_initialization_stdin_end_failed') }
    report.closed = await managed.close(tree)
    if (endError) throw endError
    await timeout(Promise.race([streamsClosed, streamErrors]), 10_000)
    assert.equal(report.closed.root, 'exited'); assert.equal(report.closed.tree, 'exited')
    assert.ok(stdoutEnded && stderrEnded && protocolComplete); assert.equal(managed.child.exitCode, 0); assert.equal(managed.child.signalCode, null)
    report.custody = tree.readFinalObservation()
    assert.ok(report.custody?.acknowledgement_delivered && report.custody.root_absent && report.custody.root_exit_observed)
    assert.equal(report.custody.observation.fullTreeProof.forced, false)
    if (primary) throw primary
    report.phase = 'inspect-closed-session'
    await api.verifyPrivateOpenCodeAccount(snapshot, root, plugin, correlation)
    const credential = api.readOpenCodeCredential(env.OPENCODE_DB, 'stepfun')
    assert.equal(credential.count, 1); assert.equal(credential.id, snapshot.policy.account.credential_id); assert.equal(credential.key, snapshot.key)
    const db = new DatabaseSync(env.OPENCODE_DB, { readOnly: true, allowExtension: false })
    let databaseError
    try {
      db.exec('PRAGMA query_only=ON; BEGIN')
      assert.equal(db.prepare('SELECT count(*) AS count FROM credential').get().count, 1)
      assert.equal(db.prepare('SELECT active FROM credential WHERE id=?').get(credential.id).active, 1)
      const session = db.prepare('SELECT id,agent,model,directory FROM session_v2 WHERE id=?').get(report.acp.started.sessionId)
      assert.equal(session.agent, 'build'); assert.deepEqual(JSON.parse(session.model), { id: 'step-3.5-flash', providerID: 'stepfun' }); assert.equal(session.directory, project)
      const counts = db.prepare('SELECT type,count(*) AS count FROM session_message WHERE session_id=? GROUP BY type').all(session.id)
      assert.ok(counts.every(row => ['agent-switched', 'model-switched'].includes(row.type)))
      report.native_session = session; report.native_message_counts = counts
      report.credential = { id: credential.id, unique: true, active: true, matches_source: true }
      db.exec('COMMIT')
    } catch (error) { databaseError = error }
    finally { try { db.close() } catch { report.database_close_error = 'opencode_initialization_database_close_failed' } }
    if (databaseError) throw databaseError
    if (report.database_close_error) throw new Error(report.database_close_error)
    api.assertNativeOpenCodeCurrent(snapshot)
    await verifyInputs()
    report.phase = 'complete'; report.status = 'INITIALIZED_WITHOUT_PROMPT'
  } catch (error) {
    primary ??= error; report.error = { phase: report.phase, code: safeCode(primary) }; report.status = 'FAIL_PRESERVE_ORIGINAL_ROOT'
  } finally {
    if (tree && !managed) {
      try { tree.closeBeforeSpawnFailure() } catch { report.close_error ??= 'opencode_initialization_spawn_cleanup_unknown' }
    }
    if (resource && !report.import_custody) {
      try { report.import_custody = resource.tree.readFinalObservation() } catch { report.close_error ??= 'opencode_initialization_import_custody_unavailable' }
    }
    if (managed) {
      closing = true
      try { client?.beginClose() } catch { report.close_error ??= 'opencode_initialization_drain_failed' }
      try { managed.child.stdin.end() } catch { report.close_error ??= 'opencode_initialization_stdin_end_failed' }
      try {
        const closed = await managed.close(tree)
        report.closed = closed
        if (closed.root !== 'exited' || closed.tree !== 'exited') report.close_error ??= 'opencode_initialization_cleanup_unknown'
      } catch { report.close_error ??= 'opencode_initialization_cleanup_unknown' }
      if (streamsClosed) {
        try { await timeout(streamsClosed, 10_000) } catch { report.close_error ??= 'opencode_initialization_stream_close_unknown' }
      }
      try {
        const custody = tree.readFinalObservation() ?? null
        if (report.custody) assert.deepEqual(custody, report.custody)
        report.custody = custody
      } catch { report.custody_error = 'opencode_initialization_final_custody_unavailable' }
      try { client?.close() } catch { report.close_error ??= 'opencode_initialization_peer_close_failed' }
    }
    report.stdout_eof = stdoutEnded; report.stderr_eof = stderrEnded; report.strict_stdout_complete = protocolComplete
    report.other_stderr_line_count = otherStderrLines; report.policy = statuses; report.frames = frames
    if (report.close_error || report.custody_error || primary) report.status = 'FAIL_PRESERVE_ORIGINAL_ROOT'
    try {
      const serialized = JSON.stringify(report)
      assert.ok(!snapshot || !serialized.includes(snapshot.key))
      await persist(join(output, 'result.json'), report)
    } catch { report.report_error = 'opencode_initialization_report_unpublished'; report.status = 'FAIL_PRESERVE_ORIGINAL_ROOT' }
    snapshot = undefined
    console.log(JSON.stringify({ status: report.status, phase: report.phase, error: report.error, close_error: report.close_error, custody_error: report.custody_error, database_close_error: report.database_close_error, report_error: report.report_error, report: join(output, 'result.json'), prompt_calls: 0, observed_model_requests: report.observed_model_requests }))
    if (report.status !== 'INITIALIZED_WITHOUT_PROMPT') process.exitCode = 1
  }
}
void main().catch(error => { console.log(JSON.stringify({ status: 'REFUSED_BEFORE_INITIALIZATION', code: safeCode(error), prompt_calls: 0 })); process.exitCode = 2 })
