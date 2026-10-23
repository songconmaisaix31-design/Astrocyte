import assert from 'node:assert/strict'
import { spawn } from 'node:child_process'
import { createHash, randomBytes } from 'node:crypto'
import { createReadStream } from 'node:fs'
import { copyFile, cp, lstat, mkdir, readFile, readdir, realpath, writeFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { dirname, isAbsolute, join, relative, resolve } from 'node:path'
import { createInterface } from 'node:readline'
import { fileURLToPath, pathToFileURL } from 'node:url'
import { parseArgs } from 'node:util'

async function main() {
// 只组合固定原包中的 Node/native 与原失败源码；不产生正式候选。
const sourceRevision = '7ff3a7d7d7e029c2fc9e2ae1fb6b81da8e0d4f8f'
const priorRevision = '1d3006fe0df0cb04b7f0de1e4563264bd0bfd9da'
const { values } = parseArgs({ options: Object.fromEntries(['artifact-root', 'work-root', 'evidence-root', 'pnpm-cli'].map(name => [name, { type: 'string' }])), strict: true, allowPositionals: false })
assert.equal(process.platform, 'win32'); assert.equal(process.arch, 'x64')
for (const name of ['artifact-root', 'work-root', 'evidence-root', 'pnpm-cli']) assert.ok(values[name] && isAbsolute(values[name]))
const checkout = await realpath(fileURLToPath(new URL('../', import.meta.url)))
const artifact = await realpath(values['artifact-root']), pnpm = await realpath(values['pnpm-cli'])
const work = resolve(values['work-root']), evidence = resolve(values['evidence-root'])
const contains = (parent, path) => { const rel = relative(parent, path); return !rel || rel !== '..' && !rel.startsWith('..\\') && !rel.startsWith('../') && !isAbsolute(rel) }
assert.ok(!contains(checkout, work) && !contains(artifact, work), 'development output must be outside source and retained artifact')
assert.equal(evidence.toLowerCase(), join(checkout, '.verification/swa-ci/windows-controller-open').toLowerCase())
assert.ok(!contains(work, evidence) && !contains(evidence, work))
assert.equal((await realpath(dirname(work))).toLowerCase(), (await realpath(process.env.RUNNER_TEMP)).toLowerCase())
await mkdir(work)
for (const path of [join(checkout, '.verification'), join(checkout, '.verification/swa-ci')]) {
  try { await mkdir(path) } catch (error) { if (error.code !== 'EEXIST') throw error }
  assert.ok((await lstat(path)).isDirectory() && !(await lstat(path)).isSymbolicLink())
  assert.equal((await realpath(path)).toLowerCase(), path.toLowerCase())
}
await mkdir(evidence)
const privateOutput = join(work, 'private')
await mkdir(privateOutput)
const save = (name, value) => writeFile(join(evidence, name), JSON.stringify(value, null, 2) + '\n', { flag: 'wx' })
const hash = bytes => createHash('sha256').update(bytes).digest('hex')
const digest = async path => ({ path: await realpath(path), sha256: hash(await readFile(path)) })
const result = { schema: 'swa.windows-controller-open-development.v1', status: 'FAIL', phase: 'preflight',
  source_revision: sourceRevision, input_revision: priorRevision, observation_revision: process.env.GITHUB_SHA ?? 'UNKNOWN',
  runtime_source_revision: sourceRevision, native_input_revision: priorRevision, go_source_revision: 'UNKNOWN',
  scope: 'development_composition_not_candidate', model_calls: 0, worker_calls: 0, task_create: 'NOT_RUN', commands: [], private_raw_uploaded: false }
const env = {}
for (const key of Object.keys(process.env)) if (/^(?:systemroot|windir|comspec|pathext|programfiles|programfiles\(x86\)|programw6432|systemdrive|number_of_processors|processor_architecture|os)$/i.test(key)) env[key] = process.env[key]
const pathKeys = Object.keys(process.env).filter(key => key.toLowerCase() === 'path')
assert.equal(pathKeys.length, 1); assert.equal(typeof process.env[pathKeys[0]], 'string')
env.PATH = process.env[pathKeys[0]]
for (const name of ['home', 'temp', 'cache']) await mkdir(join(work, name))
Object.assign(env, { HOME: join(work, 'home'), USERPROFILE: join(work, 'home'), APPDATA: join(work, 'home'), LOCALAPPDATA: join(work, 'home'),
  TMP: join(work, 'temp'), TEMP: join(work, 'temp'), TMPDIR: join(work, 'temp'), GIT_CONFIG_NOSYSTEM: '1', GIT_TERMINAL_PROMPT: '0',
  GOCACHE: join(work, 'cache', 'go-build'), GOMODCACHE: join(work, 'cache', 'go-mod') })
let native, sequence = 0, stopBuild = () => {}, failure
const originalNative = []
function publicText(value) {
  if (typeof value !== 'string') return 'UNKNOWN'
  let text = value
  for (const [path, role] of [[checkout, 'observation-source'], [work, 'development-work'], [artifact, 'retained-package'], [evidence, 'evidence']]) text = text.split(path).join(`[${role}]`)
  if (/(?:authorization|\bbearer\b|api[-_]?key|client[-_]?secret|password|cookie|credential|nonce|launchTokenHash|S-1-\d|https?:\/\/)/i.test(text)) return '[sensitive error text omitted]'
  return text.replace(/[\x00-\x1f\x7f]/g, ' ').slice(0, 500)
}
const kind = error => typeof error?.code === 'string' && /^[A-Za-z][A-Za-z0-9_.-]{0,63}$/.test(error.code) ? error.code : error instanceof Error ? error.name : 'UNKNOWN'
function utf8Fragment(text, limit, fromEnd = false) {
  let result = '', bytes = 0
  for (const character of fromEnd ? [...text].reverse() : text) {
    const size = Buffer.byteLength(character)
    if (bytes + size > limit) break
    result = fromEnd ? character + result : result + character
    bytes += size
  }
  return result
}
async function command(executable, args, cwd, cap = 120_000, extra = {}, allowNonzero = false) {
  const name = 'command-' + String(++sequence).padStart(2, '0'), out = [], err = []
  let outBytes = 0, errBytes = 0, outEOF = false, errEOF = false, streamFailure = false, identity = null, exit, timeout = false
  const child = spawn(executable, args, { cwd, env: { ...env, ...extra }, windowsHide: true, stdio: ['ignore', 'pipe', 'pipe'] })
  const closed = new Promise(done => {
    child.once('error', error => done({ exit: null, signal: null, spawn_error: kind(error) }))
    child.once('close', (code, signal) => done({ exit: code, signal, spawn_error: null }))
  })
  child.stdout.on('data', bytes => { outBytes += bytes.length; if (outBytes <= 16 * 1024 * 1024) out.push(bytes); else streamFailure = true })
  child.stderr.on('data', bytes => { errBytes += bytes.length; if (errBytes <= 1024 * 1024) err.push(bytes); else streamFailure = true })
  child.stdout.on('end', () => { outEOF = true }); child.stderr.on('end', () => { errEOF = true })
  for (const stream of [child.stdout, child.stderr]) stream.on('error', () => { streamFailure = true })
  if (native && child.pid) {
    const creationTimeMs = native.getProcessCreationTime(child.pid)
    if (Number.isSafeInteger(creationTimeMs) && creationTimeMs > 0) identity = { pid: child.pid, creationTimeMs }
  }
  let timer
  try {
    exit = await Promise.race([closed, new Promise(done => { timer = setTimeout(() => { timeout = true; done(null) }, cap) })])
  } finally { clearTimeout(timer) }
  // 超时只尝试终止本次仍为同一 incarnation 的进程；结果继续失败。
  let cleanupAttempted = false
  if (!exit && identity && native.getProcessCreationTime(identity.pid) === identity.creationTimeMs) {
    cleanupAttempted = child.kill()
    let cleanupTimer
    try { exit = await Promise.race([closed, new Promise(done => { cleanupTimer = setTimeout(() => done(null), 5000) })]) }
    finally { clearTimeout(cleanupTimer) }
  }
  if (!exit) { child.stdout.destroy(); child.stderr.destroy(); child.unref() }
  const row = { name, exit: exit ?? 'UNKNOWN', timed_out: timeout, stdout_bytes: outBytes, stderr_bytes: errBytes, stdout_eof: outEOF, stderr_eof: errEOF,
    stream_failure: streamFailure, identity_captured: !!identity, cleanup_attempted: cleanupAttempted, kernel_tree_exit: 'NOT_CLAIMED' }
  result.commands.push(row)
  await writeFile(join(privateOutput, name + '.stdout'), Buffer.concat(out), { flag: 'wx' })
  await writeFile(join(privateOutput, name + '.stderr'), Buffer.concat(err), { flag: 'wx' })
  if (!exit || timeout || exit.exit !== 0 && !allowNonzero) {
    const lines = Buffer.concat([...out, ...err]).toString('utf8').split(/\r?\n/).filter(line => /(?:\berror\b|\bfatal\b|ERR_[A-Z_]+|Cannot find|No projects matched|not recognized|SyntaxError|ReferenceError|TypeError)/i.test(line) || /^native-diagnostic phase=(?:validate|headers-download|headers-extract|node-library-download|official-msvc-toolchain|single-diagnostic-process-tree-build); kind=[A-Za-z][A-Za-z0-9]{0,63}$/.test(line))
    result.command_failure = { name, exit: exit ?? 'UNKNOWN', timed_out: timeout, error_line_total: lines.length,
      error_lines_truncated: lines.length > 20, errors: lines.slice(-20).map(publicText) }
  }
  assert.ok(!timeout && exit && outEOF && errEOF && !streamFailure, 'Actual preparation command did not settle completely')
  assert.equal(exit.signal, null); assert.equal(exit.spawn_error, null)
  if (!allowNonzero) assert.equal(exit.exit, 0, 'Actual preparation command exited nonzero')
  return { bytes: Buffer.concat(out), exit: exit.exit }
}
const git = async args => (await command('git', ['-c', 'credential.helper=', ...args], checkout)).bytes
async function fixed(path, revision = sourceRevision) { return git(['show', `${revision}:${path}`]) }
async function compareFixed(path) {
  const before = await fixed(path, priorRevision), after = await fixed(path)
  assert.ok(before.equals(after), `Retained native input differs: ${path}`)
  return { path, sha256: hash(after) }
}
async function ordinaryTree(root, base = root) {
  assert.ok(contains(base, await realpath(root)), 'Retained input escapes its ordinary package closure')
  const entry = await lstat(root)
  assert.ok(entry.isFile() || entry.isDirectory(), 'Retained package contains a nonordinary input')
  if (entry.isDirectory()) for (const name of await readdir(root)) await ordinaryTree(join(root, name), base)
  else if (/\.(?:node|dll)$/i.test(root)) originalNative.push({ path: relative(base, root).replaceAll('\\', '/'), sha256: (await digest(root)).sha256,
    dev: entry.dev, ino: entry.ino, nlink: entry.nlink })
}
try {
  const goRevision = (await git(['rev-parse', 'HEAD'])).toString().trim()
  assert.match(goRevision, /^[0-9a-f]{40}$/)
  assert.equal(goRevision, process.env.GITHUB_SHA)
  const goScope = ['cmd', 'internal', 'contracts', 'go.mod', 'go.sum']
  const changedGo = (await git(['diff', '--name-only', sourceRevision, goRevision, '--', ...goScope])).toString().trim().split(/\r?\n/).filter(Boolean).sort()
  const approvedGo = ['cmd/swa/serve.go', 'cmd/swa/process_linux.go', 'cmd/swa/process_other.go', 'cmd/swa/process_windows.go', 'cmd/swa/data_directory_windows_test.go'].sort()
  assert.deepEqual(changedGo, approvedGo, 'Development Go source differs beyond the accepted private-data repair')
  assert.equal((await git(['status', '--porcelain', '--untracked-files=no', '--', ...goScope])).toString().trim(), '')
  result.go_source_revision = goRevision
  result.go_source_inputs = await Promise.all(approvedGo.map(async path => {
    const bytes = await readFile(join(checkout, path))
    assert.ok(bytes.equals(await fixed(path, goRevision)), 'Development Go input differs from its observation commit')
    return { path, sha256: hash(bytes) }
  }))
  const metadata = JSON.parse(await readFile(join(artifact, 'artifact-metadata.json'), 'utf8'))
  assert.equal(metadata.id, 12461580039); assert.equal(metadata.name, 'swa-windows-cli-' + priorRevision)
  assert.equal(metadata.workflow_run.id, 39858402174); assert.equal(metadata.expired, false)
  assert.equal(metadata.digest, 'sha256:2781fd68ec1c936c420572704a3528107f8a236bc421639307d0fe99066c9b8d')
  const files = await readdir(artifact), zipNames = files.filter(name => name.endsWith('.zip'))
  assert.equal(zipNames.length, 1)
  const archive = join(artifact, zipNames[0]), checksum = (await readFile(archive + '.sha256', 'utf8')).trim()
  const actualArchiveHash = hash(await readFile(archive))
  assert.equal(checksum, actualArchiveHash + '  ' + zipNames[0])
  result.archive_sha256 = actualArchiveHash
  const extracted = join(work, 'original-package')
  await mkdir(extracted)
  result.phase = 'original-package-extraction'
  const powershell = join(env.ProgramFiles ?? env.PROGRAMFILES, 'PowerShell', '7', 'pwsh.exe')
  await command(powershell, ['-NoProfile', '-NonInteractive', '-Command', '[IO.Compression.ZipFile]::ExtractToDirectory($env:SWA_PROBE_ZIP,$env:SWA_PROBE_EXTRACT,$false)'], work, 120_000, { SWA_PROBE_ZIP: archive, SWA_PROBE_EXTRACT: extracted })
  const roots = await readdir(extracted)
  assert.equal(roots.length, 1)
  const original = join(extracted, roots[0])
  assert.equal((await readFile(join(original, 'VERSION'), 'utf8')).trim(), priorRevision)
  await ordinaryTree(original)
  result.native_inputs_before = originalNative
  const originalNode = join(original, 'node', 'node.exe')
  const source = join(work, 'source-7ff3a7d7')
  await mkdir(source)
  const sourceArchive = join(work, 'source-7ff3a7d7.tar')
  await git(['archive', '--format=tar', '--output', sourceArchive, sourceRevision])
  await command('tar', ['-xf', sourceArchive, '-C', source], work)
  const pin = await import(pathToFileURL(join(source, 'third_party/orca-v1.4.230/src/shared/node-runtime-pin.ts')).href)
  assert.equal(pin.NODE_RUNTIME_PIN.version, '24.21.0')
  assert.equal((await digest(originalNode)).sha256, pin.NODE_RUNTIME_ASSETS['win32-x64'].executableSha256)
  assert.equal((await command(originalNode, ['--version'], work)).bytes.toString().trim(), 'v24.21.0')
  const inputPaths = ['runtime/swa-orca/package.json', 'third_party/orca-v1.4.230/src/shared/node-runtime-pin.ts', 'scripts/download-swa-node.ps1',
    'third_party/orca-v1.4.230/config/patches/node-pty@1.1.0.patch', 'third_party/orca-v1.4.230/config/patches/@vscode__windows-process-tree@0.8.0.patch']
  const changedNative = (await git(['diff', '--name-only', priorRevision, sourceRevision, '--', 'third_party/orca-v1.4.230/native', 'third_party/orca-v1.4.230/config/patches', 'third_party/orca-v1.4.230/src/packages/process-host'])).toString().trim()
  assert.equal(changedNative, '')
  result.inputs = await Promise.all(inputPaths.map(path => compareFixed(path)))
  const requireOriginal = createRequire(join(original, 'runtime/package.json'))
  native = requireOriginal('@vscode/windows-process-tree')
  result.phase = 'frozen-source-tools'
  assert.equal((await command(originalNode, [pnpm, '--version'], source)).bytes.toString().trim(), '11.27.0')
  await command(originalNode, [pnpm, '--filter', '{.}', '--filter', '{./web}', '--filter', '@swa/orca-runtime', '--filter', '@orca/process-host', 'install', '--frozen-lockfile', '--ignore-scripts'], source, 480_000)
  const requireRuntime = createRequire(join(source, 'runtime/swa-orca/package.json'))
  const ts = requireRuntime('typescript'), { build, stop } = requireRuntime('esbuild')
  stopBuild = stop
  assert.equal(ts.version, '6.0.3')
  const yaml = requireRuntime('yaml'), priorLock = yaml.parse((await fixed('pnpm-lock.yaml', priorRevision)).toString()), targetLock = yaml.parse(await readFile(join(source, 'pnpm-lock.yaml'), 'utf8'))
  for (const [name, pkg] of Object.entries(targetLock.packages)) assert.deepEqual(priorLock.packages[name], pkg, `Retained dependency resolution changed: ${name}`)
  for (const [name, snapshot] of Object.entries(targetLock.snapshots)) {
    const before = structuredClone(priorLock.snapshots[name]), after = structuredClone(snapshot)
    if (name === 'undici@8.11.2') { delete before.optional; delete after.optional }
    assert.deepEqual(before, after, `Retained dependency snapshot changed: ${name}`)
  }
  for (const name of ['runtime/swa-orca', 'third_party/orca-v1.4.230/native/windows-security', 'third_party/orca-v1.4.230/native/windows-registry', 'third_party/orca-v1.4.230/src/packages/process-host']) assert.deepEqual(priorLock.importers[name], targetLock.importers[name])
  assert.deepEqual(priorLock.patchedDependencies, targetLock.patchedDependencies)
  result.phase = 'partial-application-build'
  const development = join(work, 'development-7ff3a7d7')
  await mkdir(development)
  const entries = ['host.cjs', 'client.cjs', 'controller-helper.cjs', 'daemon-entry.js']
  for (const name of ['node', 'runtime', 'third-party-licenses']) await cp(join(original, name), join(development, name), { recursive: true, force: false, errorOnExist: true,
    filter: path => !entries.some(entry => path === join(original, 'runtime', entry) || path === join(original, 'runtime', entry + '.metafile.json')) })
  await writeFile(join(development, 'VERSION'), goRevision + '\n', { flag: 'wx' })
  for (const [name, text] of [['PLATFORM', 'win32-x64'], ['EDITION', 'cli']]) await writeFile(join(development, name), text + '\n', { flag: 'wx' })
  const builderPath = join(source, 'scripts/build-swa-runtime.mjs'), builderText = await readFile(builderPath, 'utf8')
  const tree = ts.createSourceFile(builderPath, builderText, ts.ScriptTarget.Latest, true, ts.ScriptKind.JS)
  assert.equal(tree.parseDiagnostics.length, 0)
  const functions = tree.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === 'buildSWARuntime')
  assert.equal(functions.length, 1)
  const declarations = functions[0].body.statements.filter(ts.isVariableStatement).flatMap(node => [...node.declarationList.declarations])
  const common = declarations.filter(node => node.name.getText(tree) === 'common'), mappings = declarations.filter(node => node.name.getText(tree) === 'entries')
  assert.equal(common.length, 1); assert.equal(mappings.length, 1)
  assert.ok(ts.isObjectLiteralExpression(common[0].initializer) && ts.isObjectLiteralExpression(mappings[0].initializer))
  const entryMap = new Map(mappings[0].initializer.properties.map(node => { assert.ok(ts.isPropertyAssignment(node) && ts.isStringLiteral(node.name)); return [node.name.text, node.initializer.getText(tree)] }))
  for (const name of entries) assert.ok(entryMap.has(name))
  const adapter = join(privateOutput, 'original-runtime-options.mjs')
  await writeFile(adapter, 'import { createRequire } from "node:module"; import { readFile } from "node:fs/promises"; import { dirname,join,resolve } from "node:path";\n' +
    `const root=${JSON.stringify(source)},runtime=join(root,"runtime/swa-orca"),upstream=join(root,"third_party/orca-v1.4.230"),requireRuntime=createRequire(join(runtime,"package.json"));\n` +
    'export const common=' + common[0].initializer.getText(tree) + ';\nexport const entries={' + entries.map(name => JSON.stringify(name) + ':' + entryMap.get(name)).join(',') + '};\n', { flag: 'wx' })
  const options = await import(pathToFileURL(adapter).href)
  const sourceInputs = new Set([builderPath])
  for (const [name, entry] of Object.entries(options.entries)) {
    const built = await build({ ...options.common, entryPoints: [entry], outfile: join(development, 'runtime', name) })
    await writeFile(join(development, 'runtime', name + '.metafile.json'), JSON.stringify(built.metafile), { flag: 'wx' })
    for (const name of Object.keys(built.metafile.inputs)) sourceInputs.add(resolve(source, name))
  }
  result.reused_child_entries = []
  for (const name of ['parcel-watcher-process-entry.js', 'profile-state-writer-worker-entry.js', 'profile-state-backup-worker-entry.js', 'foreign-sqlite-reader-entry.js', 'port-scan-command-worker-entry.js', 'session-scanner-service-entry.js', 'worker-thread-child-process-entry.js']) {
    const meta = JSON.parse(await readFile(join(original, 'runtime', name + '.metafile.json'), 'utf8'))
    let count = 0
    for (const path of Object.keys(meta.inputs)) {
      if (path.split(/[\\/]/).includes('node_modules') || path.startsWith('<')) continue
      assert.ok(!isAbsolute(path) && !path.split(/[\\/]/).includes('..'), 'Reused child input is not a source-relative Git path')
      await compareFixed(path); count++
    }
    assert.ok(count > 0)
    result.reused_child_entries.push({ name, git_input_count: count, sha256: (await digest(join(original, 'runtime', name))).sha256 })
  }
  result.source_inputs = await Promise.all([...sourceInputs].sort().map(async path => ({ path: relative(source, path).replaceAll('\\', '/'), sha256: (await digest(path)).sha256 })))
  assert.equal((await command('go', ['version'], checkout)).bytes.toString().trim().split(' ')[2], 'go1.27.2')
  await command('go', ['mod', 'download'], checkout, 240_000)
  await command('go', ['mod', 'verify'], checkout, 120_000)
  await command(originalNode, [join(source, 'scripts/check-dependencies.mjs')], source, 120_000)
  result.phase = 'windows-private-data-sdk-control'
  const sdk = await command('go', ['test', '-json', '-mod=readonly', '-count=1', '-timeout=90s', '-run', '^TestWindowsPrivateDataDirectory$', './cmd/swa'], checkout, 180_000, {}, true)
  const sdkCases = sdk.bytes.toString('utf8').split(/\r?\n/).filter(Boolean).map(line => JSON.parse(line))
    .filter(row => typeof row.Test === 'string' && ['pass', 'skip', 'fail'].includes(row.Action))
    .map(row => { assert.match(row.Test, /^TestWindowsPrivateDataDirectory(?:\/[a-z_]+)?$/); return { test: row.Test, action: row.Action } })
  const sdkNames = ['fresh_nested_and_existing_private', 'existing_extra_administrators_ace_unchanged', 'existing_wrong_owner_unchanged', 'ordinary_file_unchanged', 'reparse_target_and_ancestor', 'protected_roots_before_creation']
  result.windows_private_data_sdk = { exit: sdk.exit, cases: sdkCases }
  assert.equal(sdk.exit, 0)
  assert.deepEqual(sdkCases.filter(row => row.test.includes('/')).map(row => row.test.split('/')[1]).sort(), [...sdkNames].sort())
  for (const row of sdkCases) assert.ok(row.action === 'pass' || row.action === 'skip' && ['existing_wrong_owner_unchanged', 'reparse_target_and_ancestor'].includes(row.test.split('/')[1]))
  assert.ok(sdkCases.some(row => row.test === 'TestWindowsPrivateDataDirectory' && row.action === 'pass'))
  result.phase = 'partial-go-build'
  for (const [name, entry] of [['swa.exe', './cmd/swa'], ['server.exe', './cmd/swa-server']]) {
    await command('go', ['build', '-mod=readonly', '-p=4', '-trimpath', ...(name === 'swa.exe' ? ['-ldflags=-X main.version=' + goRevision] : []), '-o', join(development, name), entry], checkout, 480_000)
  }
  const actualNode = join(development, 'node/node.exe')
  assert.equal((await command(join(development, 'swa.exe'), ['--version'], development)).bytes.toString().trim(), goRevision)
  result.phase = 'diagnostic-process-tree-component'
  const developmentRuntime = join(development, 'runtime'), requireDevelopment = createRequire(join(developmentRuntime, 'package.json'))
  const originalPackage = await realpath(dirname(requireOriginal.resolve('@vscode/windows-process-tree/package.json')))
  const sourcePackage = await realpath(dirname(requireRuntime.resolve('@vscode/windows-process-tree/package.json')))
  const diagnosticPackage = await realpath(dirname(requireDevelopment.resolve('@vscode/windows-process-tree/package.json')))
  assert.ok(contains(developmentRuntime, diagnosticPackage) && !contains(original, diagnosticPackage) && !contains(source, diagnosticPackage))
  const diagnosticBinary = join(diagnosticPackage, 'build/Release/windows_process_tree.node')
  const retainedBinary = join(originalPackage, 'build/Release/windows_process_tree.node')
  const [diagnosticBinaryInfo, retainedBinaryInfo] = await Promise.all([lstat(diagnosticBinary), lstat(retainedBinary)])
  assert.ok(diagnosticBinaryInfo.isFile() && !diagnosticBinaryInfo.isSymbolicLink() && diagnosticBinaryInfo.nlink === 1)
  assert.ok(diagnosticBinaryInfo.dev !== retainedBinaryInfo.dev || diagnosticBinaryInfo.ino !== retainedBinaryInfo.ino, 'Diagnostic DLL aliases the retained native input')
  const diagnosticSources = ['src/process.h', 'src/process.cc', 'src/process_worker.cc']
  const unchangedSources = []
  result.diagnostic_process_tree = { scope: 'diagnostic_development_copy_only', observation_revision: goRevision, baseline_revision: priorRevision, source_inputs: [] }
  for (const name of diagnosticSources) {
    const originalPath = join(originalPackage, name), sourcePath = join(sourcePackage, name), target = join(diagnosticPackage, name)
    const [originalInfo, sourceInfo, targetInfo] = await Promise.all([lstat(originalPath), lstat(sourcePath), lstat(target)])
    assert.ok([originalInfo, sourceInfo, targetInfo].every(info => info.isFile() && !info.isSymbolicLink()))
    assert.equal(targetInfo.nlink, 1, 'Diagnostic source must not share a hardlink')
    for (const [path, info] of [[originalPath, originalInfo], [sourcePath, sourceInfo]]) {
      assert.ok(targetInfo.dev !== info.dev || targetInfo.ino !== info.ino, 'Diagnostic source aliases a retained or shared source')
      const bytes = await readFile(path)
      assert.ok(bytes.equals(await readFile(target)), 'Diagnostic source differs from its fixed retained input')
      unchangedSources.push({ path, sha256: hash(bytes), dev: info.dev, ino: info.ino, nlink: info.nlink })
    }
    assert.equal((await realpath(target)).toLowerCase(), resolve(target).toLowerCase())
    result.diagnostic_process_tree.source_inputs.push({ name, original_sha256: (await digest(target)).sha256 })
  }
  const diagnosticPatch = join(checkout, 'tests/swa-cli-headless/windows-process-query-diagnostic.patch')
  result.diagnostic_process_tree.patch_sha256 = (await digest(diagnosticPatch)).sha256
  await command('git', ['-c', 'core.autocrlf=false', 'apply', '--check', '--unidiff-zero', diagnosticPatch], diagnosticPackage)
  await command('git', ['-c', 'core.autocrlf=false', 'apply', '--unidiff-zero', diagnosticPatch], diagnosticPackage)
  for (const item of result.diagnostic_process_tree.source_inputs) item.diagnostic_sha256 = (await digest(join(diagnosticPackage, item.name))).sha256
  const buildRoot = join(checkout, '.verification/swa-cli-headless-20261020/build')
  await mkdir(buildRoot, { recursive: true })
  assert.equal((await realpath(buildRoot)).toLowerCase(), resolve(buildRoot).toLowerCase())
  const nativeWork = join(buildRoot, 'candidate-' + randomBytes(3).toString('hex'))
  await mkdir(nativeWork)
  const componentResult = join(privateOutput, 'diagnostic-native-preparation.json')
  await command(powershell, ['-NoProfile', '-NonInteractive', '-File', join(checkout, 'scripts/build-swa-windows-process-diagnostic.ps1'),
    '-SourceRoot', source, '-Runtime', developmentRuntime, '-Node', actualNode, '-Work', nativeWork, '-OutputFile', componentResult], checkout, 1_300_000, { RUNNER_TEMP: dirname(work) })
  const nativePreparation = JSON.parse(await readFile(componentResult, 'utf8'))
  assert.equal(nativePreparation.status, 'PASS'); assert.equal(nativePreparation.native_component, '@vscode/windows-process-tree')
  assert.equal(nativePreparation.native_component_index, 1); assert.equal(nativePreparation.other_native_components_built, false)
  result.diagnostic_process_tree.preparation = nativePreparation
  result.diagnostic_process_tree.binary_path = relative(development, diagnosticBinary).replaceAll('\\', '/')
  result.diagnostic_process_tree.binary_sha256 = (await digest(diagnosticBinary)).sha256
  const retainedProcessTree = originalNative.filter(item => item.path === result.diagnostic_process_tree.binary_path)
  assert.equal(retainedProcessTree.length, 1)
  assert.notEqual(result.diagnostic_process_tree.binary_sha256, retainedProcessTree[0].sha256)
  result.diagnostic_process_tree.baseline_binary_sha256 = retainedProcessTree[0].sha256
  const sourceInputGuards = async () => {
    for (const item of unchangedSources) {
      const info = await lstat(item.path)
      assert.equal(info.dev, item.dev); assert.equal(info.ino, item.ino); assert.equal(info.nlink, item.nlink); assert.equal((await digest(item.path)).sha256, item.sha256)
    }
    for (const item of result.diagnostic_process_tree.source_inputs) {
      const info = await lstat(join(diagnosticPackage, item.name))
      assert.equal(info.nlink, 1); assert.equal((await digest(join(diagnosticPackage, item.name))).sha256, item.diagnostic_sha256)
    }
  }
  await sourceInputGuards()
  native = requireDevelopment('@vscode/windows-process-tree')
  const diagnosticRows = await new Promise((resolveRows, reject) => {
    const timer = setTimeout(() => reject(new Error('Diagnostic native callback did not settle')), 15_000)
    native.getAllProcesses(rows => { clearTimeout(timer); rows ? resolveRows(rows) : reject(new Error('Diagnostic native callback unavailable')) }, 4)
  })
  const diagnosticSelf = diagnosticRows.filter(row => row.pid === process.pid)
  assert.equal(diagnosticSelf.length, 1); assert.equal(diagnosticSelf[0].creationTimeDiagnosticVersion, 1)
  assert.equal(diagnosticSelf[0].creationTimeQueryStage, 4); assert.equal(diagnosticSelf[0].creationTimeWin32Error, 0)
  assert.equal(diagnosticSelf[0].creationTimeCloseError, 0)
  result.diagnostic_process_tree.actual_callback_version = 1
  result.development_entries = await Promise.all(['swa.exe', 'server.exe', 'VERSION', 'node/node.exe', ...entries.map(name => 'runtime/' + name)].map(async name => ({ name, sha256: (await digest(join(development, name))).sha256 })))
  // 观察脚本明确来自 observation revision，工具与所有产品输入仍解析自固定7ff目录。
  result.observer_sources = []
  for (const name of ['controller-input-check.mjs', 'controller-taskcreate-check.mjs', 'serve-lifecycle.mjs']) {
    const destination = join(source, 'tests/swa-cli-headless', name), observation = join(checkout, 'tests/swa-cli-headless', name)
    const originalSha = (await digest(destination)).sha256, actualSha = (await digest(observation)).sha256
    await copyFile(observation, destination)
    assert.equal((await digest(destination)).sha256, actualSha)
    result.observer_sources.push({ name, original_7ff_sha256: originalSha, observation_sha256: actualSha })
  }
  const requireObserver = createRequire(join(source, 'runtime/swa-orca/package.json'))
  result.observer_tool_resolution = await Promise.all(['esbuild', 'typescript'].map(async name => {
    const path = await realpath(requireObserver.resolve(name))
    assert.ok(contains(source, path), 'Observer build dependency escaped the frozen source installation')
    return { name, path: relative(source, path).replaceAll('\\', '/') }
  }))
  const dataParent = join(work, 'data')
  await mkdir(dataParent)
  result.phase = 'short-identify-then-controller-open'
  const smoke = join(privateOutput, 'smoke')
  const observed = await command(actualNode, [join(source, 'tests/swa-cli-headless/serve-lifecycle.mjs'), '--candidate', development, '--evidence', smoke,
    '--data-parent', dataParent, '--protected-roots', env.ProgramFiles ?? env.PROGRAMFILES, '--port', '18793', '--mode', 'readiness-eof', '--source-revision', goRevision, '--controller-open-only'], checkout, 660_000, {}, true)
  result.fixture_exit = observed.exit
  result.native_inputs_after = await Promise.all(originalNative.map(async item => {
    const path = join(original, item.path), info = await lstat(path)
    return { path: item.path, sha256: (await digest(path)).sha256, dev: info.dev, ino: info.ino, nlink: info.nlink }
  }))
  assert.deepEqual(result.native_inputs_after, originalNative)
  for (const item of originalNative) assert.equal((await digest(join(development, item.path))).sha256,
    item.path === result.diagnostic_process_tree.binary_path ? result.diagnostic_process_tree.binary_sha256 : item.sha256)
  await sourceInputGuards()
  for (const item of result.source_inputs) assert.equal((await digest(join(source, item.path))).sha256, item.sha256)
  for (const item of result.go_source_inputs) assert.equal((await digest(join(checkout, item.path))).sha256, item.sha256)
  assert.equal((await git(['status', '--porcelain', '--untracked-files=no', '--', ...goScope])).toString().trim(), '')
  assert.equal(hash(await readFile(archive)), actualArchiveHash)
  const actual = JSON.parse(await readFile(join(smoke, 'result.json'), 'utf8'))
  assert.equal(actual.source_revision, goRevision); assert.equal(actual.records.length, 1)
  const record = actual.records[0], control = record.controller_input_check, producer = control?.producer
  result.control = { status: control?.status ?? 'UNKNOWN', transport: control?.transport ?? 'NOT_OBSERVED',
    producer: producer ? { status: producer.status, phase: producer.phase, error: producer.error ?? null, native_error: producer.native_error ?? null,
      open_request_id: producer.open_request_id ?? null, original_open_observation: producer.original_open_observation ?? null,
      open_receipt: producer.open_receipt ?? null, close_receipt: producer.close_receipt ?? null, cleanup: producer.cleanup,
      directory_observations: producer.directory_observations ?? [], ownership_incomplete: producer.ownership_incomplete } : 'NOT_STARTED' }
  try {
    const transportPath = join(smoke, 'readiness-eof/controller-input/transport/report.json')
    assert.ok((await lstat(transportPath)).size <= 2 * 1024 * 1024)
    const transport = JSON.parse(await readFile(transportPath, 'utf8'))
    assert.equal(transport.schema, 'swa.controller-input-check.v1')
    result.control.short_error = transport.error ? { kind: kind(transport.error), message: publicText(typeof transport.error.message === 'string' ? transport.error.message.split(/[\r\n]/u, 1)[0] : undefined) } : null
    const integer = value => Number.isSafeInteger(value) ? value : 'UNKNOWN'
    const boolean = value => typeof value === 'boolean' ? value : 'UNKNOWN'
    const tag = value => value === null ? null : typeof value === 'string' && /^[A-Za-z][A-Za-z0-9_.-]{0,63}$/.test(value) ? value : 'UNKNOWN'
    const traceEvent = event => !event || typeof event !== 'object' ? 'UNKNOWN' : {
      lane: ['tty', 'pipe'].includes(event.lane) ? event.lane : 'UNKNOWN',
      stage: ['create', 'spawn_identity', 'native_inputs', 'write', 'reply', 'identity', 'eof', 'absence', 'cleanup', 'exit'].includes(event.stage) ? event.stage : 'UNKNOWN',
      boundary: ['start', 'complete', 'observed'].includes(event.boundary) ? event.boundary : 'UNKNOWN',
      packet_index: event.packet_index === null ? null : integer(event.packet_index),
      packet_bytes: event.packet_bytes === null ? null : integer(event.packet_bytes),
      observed_at_ms: Number.isSafeInteger(event.observed_at_ms) ? event.observed_at_ms : 'UNKNOWN',
      monotonic_ms: Number.isFinite(event.monotonic_ms) ? event.monotonic_ms : 'UNKNOWN',
      elapsed_ms: Number.isFinite(event.elapsed_ms) ? event.elapsed_ms : 'UNKNOWN',
      deadline_remaining_ms: Number.isFinite(event.deadline_remaining_ms) ? event.deadline_remaining_ms : 'UNKNOWN' }
    const identity = value => value && Number.isSafeInteger(value.pid) && value.pid > 0 && Number.isSafeInteger(value.creationTimeMs) && value.creationTimeMs > 0
      ? { pid: value.pid, creation_time_ms: value.creationTimeMs } : null
    result.short_transport_diagnostic = { status: ['PASS', 'FAIL'].includes(transport.status) ? transport.status : 'UNKNOWN',
      error: transport.error ? { name: tag(transport.error.name), code: tag(transport.error.code) } : null,
      cleanup_unknown: boolean(transport.cleanup_unknown), trace_total: integer(transport.trace_total),
      trace: Array.isArray(transport.trace) ? transport.trace.slice(0, 64).map(traceEvent) : [],
      trace_truncated: transport.trace_truncated === true || transport.trace?.length > 64,
      last_event: traceEvent(transport.last_event), lanes: {}, identities_history: [] }
    for (const lane of ['tty', 'pipe']) {
      const row = transport[lane]
      result.short_transport_diagnostic.lanes[lane] = row === null ? null : row && typeof row === 'object' ? {
        status: ['PASS', 'FAIL'].includes(row.status) ? row.status : 'UNKNOWN', natural_eof: boolean(row.natural_eof), forced_cleanup: boolean(row.forced_cleanup),
        reply_count: Array.isArray(row.observations) ? row.observations.length : 'UNKNOWN',
        exit: row.exit === null ? null : row.exit ? { code: integer(row.exit.code), signal: row.exit.signal === null ? null : tag(row.exit.signal),
          cause: row.exit.cause ? { kind: ['operator_close', 'signaled', 'exited', 'unknown'].includes(row.exit.cause.kind) ? row.exit.cause.kind : 'UNKNOWN',
            ...(row.exit.cause.kind === 'exited' ? { exit_code: integer(row.exit.cause.exitCode) } : {}),
            ...(row.exit.cause.kind === 'signaled' ? { signal: integer(row.exit.cause.signal) } : {}),
            ...(row.exit.cause.kind === 'unknown' ? { reason: ['stop_unverified', 'host_status_unavailable', 'cause_unreported'].includes(row.exit.cause.reason) ? row.exit.cause.reason : 'UNKNOWN' } : {}) } : row.exit.cause === null ? null : 'NOT_RECORDED' } : 'UNKNOWN',
        failure_phase: row.failure_phase ? traceEvent(row.failure_phase) : null,
        sessions_after_dispose_count: Array.isArray(row.sessions_after_dispose) ? row.sessions_after_dispose.length : 'NOT_APPLICABLE',
        raw_output_bytes: row.raw_output_bytes === undefined ? 'NOT_RECORDED' : integer(row.raw_output_bytes),
        framing: row.framing && typeof row.framing === 'object' ? { valid_reply_records: integer(row.framing.valid_reply_records), schema_rejected_records: integer(row.framing.schema_rejected_records) } : 'NOT_RECORDED',
        framer_rejections: Array.isArray(row.rejected_output) ? ['invalid-json', 'line-too-long', 'UNKNOWN'].map(kind => ({ kind,
          count: row.rejected_output.filter(actual => kind === 'UNKNOWN' ? !['invalid-json', 'line-too-long'].includes(actual) : actual === kind).length })) : 'NOT_RECORDED' } : 'UNKNOWN'
    }
    const readyKinds = new Set(['startup-command-delivery', 'startup-command-stage-failed', 'shell-ready-timeout', 'shell-ready-wrapper-replaced'])
    const readyRows = Array.isArray(transport.tty?.readiness) ? transport.tty.readiness.map(row => ({
      event: readyKinds.has(row?.event) ? row.event : 'UNKNOWN', observed_at_ms: integer(row?.observed_at_ms),
      monotonic_ms: Number.isFinite(row?.monotonic_ms) ? row.monotonic_ms : 'UNKNOWN',
      ...(row?.event === 'startup-command-delivery' ? { written: boolean(row.details?.written), has_command: boolean(row.details?.hasCommand),
        command_length: integer(row.details?.commandLength), via_shell_args: boolean(row.details?.viaShellArgs), queued_by_shell_ready_barrier: boolean(row.details?.queuedByShellReadyBarrier) } : {}),
      ...(row?.event === 'shell-ready-timeout' ? { timeout_ms: integer(row.details?.timeoutMs) } : {}),
      ...(row?.event === 'startup-command-stage-failed' ? { reason: tag(row.details?.reason) } : {}) })) : null
    result.short_transport_diagnostic.ready_events = readyRows ? (readyRows.length > 32 ? readyRows.slice(0, 16).concat(readyRows.slice(-16)) : readyRows) : 'NOT_RECORDED'
    result.short_transport_diagnostic.ready_event_total = readyRows ? readyRows.length : 'NOT_RECORDED'
    result.short_transport_diagnostic.ready_events_truncated = !!readyRows && readyRows.length > 32
    result.short_transport_diagnostic.raw = { status: 'NOT_CAPTURED' }
    if (transport.coverage === 'explicit_short_identify_only' && transport.long_input === 'NOT_RUN' && transport.rpc_mutations === 'NOT_RUN' && transport.model_execution === 'NOT_RUN') {
      try {
        const rawPath = join(dirname(transportPath), 'tty-raw-transcript.txt'), rawInfo = await lstat(rawPath), actualPath = await realpath(rawPath)
        assert.ok(rawInfo.isFile() && !rawInfo.isSymbolicLink() && rawInfo.nlink === 1 && rawInfo.size <= 1024 * 1024)
        assert.equal(actualPath.toLowerCase(), resolve(rawPath).toLowerCase())
        const rawBytes = await readFile(rawPath)
        assert.equal(rawBytes.length, transport.tty.raw_output_bytes)
        let text = rawBytes.toString('utf8')
        for (const [path, role] of [[dirname(transportPath), 'short-control'], [checkout, 'observation-source'], [work, 'development-work'], [artifact, 'retained-package'], [evidence, 'evidence']]) text = text.split(path).join(`[${role}]`)
        const total = Buffer.byteLength(text), head = utf8Fragment(text, total <= 512 ? 512 : 256), tail = total <= 512 ? '' : utf8Fragment(text, 256, true)
        result.short_transport_diagnostic.raw = { status: 'OBSERVED', source: 'explicit_short_identify_only', original_bytes: rawBytes.length,
          projected_bytes: total, fragment_limit_bytes: 512, head, tail, fragment_bytes: Buffer.byteLength(head) + Buffer.byteLength(tail), truncated: total > 512 }
      } catch (error) { result.short_transport_diagnostic.raw = { status: error?.code === 'ENOENT' ? 'NOT_CAPTURED' : 'UNKNOWN', kind: kind(error) } }
    }
    for (const [role, value] of [['tty_root', transport.tty?.root ?? transport.tty?.spawn_incarnation], ['tty_helper', transport.tty?.helper],
      ['pipe_helper', transport.pipe?.helper], ['short_checker', control?.transport_process]]) {
      const actualIdentity = identity(value)
      if (actualIdentity) result.short_transport_diagnostic.identities_history.push({ role, ...actualIdentity })
    }
  } catch (error) { result.control.short_error = { status: 'UNKNOWN', kind: kind(error) } }
  result.service = { exit: record.exit, signal: record.signal, ownership_closure_incomplete: record.ownership_closure_incomplete,
    unknown_process_observation_count: record.unknown_process_observations.length,
    native_remaining: record.exit_observations.map(row => row.remaining.length), cim_exit: record.actual_cim_process_exit.exit,
    cim_all_exited: record.actual_cim_report?.all_exited ?? 'UNKNOWN' }
  const reasons = new Set(['invalid_native_process_row', 'current_root_identity_unverified', 'root_in_descendant_branch',
    'parent_incarnation_unverified', 'process_identity_not_unique', 'creation_time_unavailable', 'parent_incarnation_changed'])
  const number = value => Number.isSafeInteger(value) && value >= 0 ? value : 'UNKNOWN'
  const identifier = value => value === null ? null : typeof value === 'string' && /^[A-Za-z][A-Za-z0-9_.-]{0,63}$/.test(value) ? value : 'UNKNOWN'
  const boolean = value => typeof value === 'boolean' ? value : 'UNKNOWN'
  const processEvent = event => event === null ? null : event && typeof event === 'object' ? { exit: event.exit === null ? null : Number.isSafeInteger(event.exit) ? event.exit : 'UNKNOWN',
    signal: identifier(event.signal), observed_at_ms: number(event.observed_at_ms), monotonic_ms: Number.isFinite(event.monotonic_ms) ? event.monotonic_ms : 'UNKNOWN' } : 'UNKNOWN'
  const eventObserved = event => event === null ? false : event && typeof event === 'object' && Number.isSafeInteger(event.observed_at_ms) && Number.isFinite(event.monotonic_ms) ? true : 'UNKNOWN'
  const queryWindow = window => window && typeof window === 'object' ? {
    start_at_ms: number(window.start?.observed_at_ms), end_at_ms: number(window.end?.observed_at_ms),
    start_monotonic_ms: Number.isFinite(window.start?.monotonic_ms) ? window.start.monotonic_ms : 'UNKNOWN',
    end_monotonic_ms: Number.isFinite(window.end?.monotonic_ms) ? window.end.monotonic_ms : 'UNKNOWN',
    root_exit_observed_before_query: eventObserved(window.start?.root_exit_event), root_exit_observed_after_query: eventObserved(window.end?.root_exit_event),
    root_close_observed_before_query: eventObserved(window.start?.root_close_event), root_close_observed_after_query: eventObserved(window.end?.root_close_event) } : 'UNKNOWN'
  const counts = new Map(), groups = new Map(), queryCounts = new Map()
  assert.ok(Array.isArray(record.unknown_process_observations) && Array.isArray(record.observed_processes))
  for (const entry of record.unknown_process_observations) {
    const reason = reasons.has(entry?.reason) ? entry.reason : 'UNKNOWN'
    counts.set(reason, (counts.get(reason) ?? 0) + 1)
    const identity = { pid: number(entry?.actual?.pid), ppid: number(entry?.actual?.ppid), creation_time_ms: number(entry?.actual?.creationTimeMs) }
    const anchor = entry?.source_anchor && ['serve', 'additional'].includes(entry.source_anchor.kind)
      ? { kind: entry.source_anchor.kind, pid: number(entry.source_anchor.pid), creation_time_ms: number(entry.source_anchor.creation_time_ms) } : 'UNKNOWN'
    const diagnostic = entry?.actual?.creationTimeDiagnosticVersion === 1 ? {
      version: 1, api_stage: Number.isInteger(entry.actual.creationTimeQueryStage) && entry.actual.creationTimeQueryStage >= 0 && entry.actual.creationTimeQueryStage <= 4
        ? ['not_requested', 'OpenProcess', 'GetProcessTimes', 'invalid_filetime', 'complete'][entry.actual.creationTimeQueryStage] : 'UNKNOWN',
      win32_error: number(entry.actual.creationTimeWin32Error), close_error: number(entry.actual.creationTimeCloseError) } : 'UNKNOWN'
    const queryKey = JSON.stringify(diagnostic)
    const queryCount = queryCounts.get(queryKey)
    queryCounts.set(queryKey, { query: diagnostic, count: (queryCount?.count ?? 0) + 1 })
    const key = JSON.stringify([reason, identity.pid, identity.ppid, identity.creation_time_ms, anchor, diagnostic])
    const observedAt = number(entry?.observed_at_ms)
    if (groups.has(key)) {
      const group = groups.get(key)
      group.count++
      group.last_observed_at_ms = observedAt
      group.last_sample_sequence = number(entry?.sample_sequence)
      group.last_query_window = queryWindow(entry?.query_window)
      continue
    }
    const parents = new Map()
    if (typeof identity.ppid === 'number') for (const row of record.observed_processes) {
      if (row?.pid === identity.ppid && number(row.pid) !== 'UNKNOWN' && Number.isSafeInteger(row.creationTimeMs) && row.creationTimeMs > 0) {
        parents.set(row.pid + ':' + row.creationTimeMs, { pid: row.pid, creation_time_ms: row.creationTimeMs })
      }
    }
    const roleHistory = result.short_transport_diagnostic?.identities_history ?? []
    const roles = value => {
      const matches = value && typeof value.pid === 'number' && typeof value.creation_time_ms === 'number'
        ? [...new Set(roleHistory.filter(row => row.pid === value.pid && row.creation_time_ms === value.creation_time_ms).map(row => row.role))] : []
      return matches.length === 1 ? matches[0] : 'UNKNOWN'
    }
    groups.set(key, { reason, ...identity, source_anchor: anchor, diagnostic_query: diagnostic,
      actual_role_history: roles(identity), source_anchor_role_history: roles(anchor),
      count: 1, first_observed_at_ms: observedAt, last_observed_at_ms: observedAt,
      first_sample_sequence: number(entry?.sample_sequence), last_sample_sequence: number(entry?.sample_sequence),
      first_query_window: queryWindow(entry?.query_window), last_query_window: queryWindow(entry?.query_window),
      matches_recorded_identity_history: typeof identity.pid === 'number' && identity.pid > 0 && typeof identity.creation_time_ms === 'number' && identity.creation_time_ms > 0
        ? record.observed_processes.some(row => row?.pid === identity.pid && row.creationTimeMs === identity.creation_time_ms) : 'UNKNOWN',
      recorded_parent_identities_history: [...parents.values()].slice(0, 2), recorded_parent_identity_count: typeof identity.ppid === 'number' ? parents.size : 'UNKNOWN',
      recorded_parent_identities_truncated: parents.size > 2 })
  }
  const spawnEvents = record.events?.filter(event => event?.type === 'actual-spawn' && event.pid === record.pid) ?? []
  const rootCreation = spawnEvents.length === 1 ? number(spawnEvents[0].creation_time_ms) : 'UNKNOWN'
  const readinessError = record.readiness_check_error
  const unknownProjection = { go_source_revision: result.go_source_revision, runtime_source_revision: sourceRevision, native_input_revision: priorRevision,
    ownership_closure_incomplete: boolean(record.ownership_closure_incomplete),
    serve_root: { pid: number(record.pid), creation_time_ms: rootCreation },
    root_process_events: record.service_process_observations?.pid === record.pid && record.service_process_observations?.creation_time_ms === rootCreation
      ? { exit: processEvent(record.service_process_observations.exit_event), close: processEvent(record.service_process_observations.close_event) } : 'UNKNOWN',
    latch_inputs: { root_creation_time_readable: typeof rootCreation === 'number' && rootCreation > 0,
      controller_input_ownership_incomplete: boolean(control?.ownership_incomplete), producer_ownership_incomplete: boolean(producer?.ownership_incomplete),
      readiness_check_error: readinessError === null ? null : readinessError && typeof readinessError === 'object'
        ? { name: identifier(readinessError.name), code: identifier(readinessError.code) } : 'UNKNOWN' },
    unknown_item_count: record.unknown_process_observations.length,
    reason_counts: [...counts].map(([reason, count]) => ({ reason, count })),
    native_query_type_count: queryCounts.size, native_query_counts: [...queryCounts.values()].slice(0, 8), native_query_counts_truncated: queryCounts.size > 8,
    identity_group_count: groups.size, samples: groups.size > 8 ? [...groups.values()].slice(0, 4).concat([...groups.values()].slice(-4)) : [...groups.values()], samples_truncated: groups.size > 8,
    short_transport: result.short_transport_diagnostic ?? 'NOT_OBSERVED',
    identity_history_is_not_a_single_snapshot_chain: true }
  result.unknown_process_observation_projection = unknownProjection
  const logs = []
  let errorLineCount = 0
  for (const [role, path] of [['service', join(smoke, 'readiness-eof/stderr.log')], ['daemon', join(dataParent, 'readiness-eof/logs/daemon.log')], ['short-helper', join(smoke, 'readiness-eof/controller-input/transport.stderr')]]) {
    try {
      const info = await lstat(path)
      assert.ok(info.isFile() && info.size <= 16 * 1024 * 1024)
      const reader = createInterface({ input: createReadStream(path), crlfDelay: Infinity })
      for await (const line of reader) if (/(?:error|failed|unknown|exception|native_controller|native_codex|private_)/i.test(line)) {
        errorLineCount++
        logs.push({ role, text: publicText(line) })
        if (logs.length > 20) logs.shift()
      }
    } catch (error) { logs.push({ role, status: 'UNKNOWN', kind: kind(error) }) }
  }
  result.errors = logs
  result.error_line_total = errorLineCount
  result.error_lines_truncated = errorLineCount > 20
  assert.equal(observed.exit, 0); assert.equal(control?.status, 'PASS')
  result.phase = 'complete'; result.status = 'PASS'
} catch (error) {
  failure = error
  result.error = { kind: kind(error), message: publicText(error?.message) }
  await writeFile(join(privateOutput, 'failure.txt'), String(error?.stack ?? error), { flag: 'wx' })
} finally {
  await stopBuild()
  await save('result.json', result)
  const projection = { status: result.status, phase: result.phase, scope: result.scope, source_revision: sourceRevision, input_revision: priorRevision,
    go_source_revision: result.go_source_revision, runtime_source_revision: sourceRevision, native_input_revision: priorRevision,
    native_input_scope: 'baseline_except_diagnostic_process_tree', diagnostic_process_tree_sha256: result.diagnostic_process_tree?.binary_sha256 ?? 'NOT_BUILT',
    windows_private_data_sdk: result.windows_private_data_sdk ? { exit: result.windows_private_data_sdk.exit,
      pass_count: result.windows_private_data_sdk.cases.filter(row => row.action === 'pass').length,
      not_run: result.windows_private_data_sdk.cases.filter(row => row.action === 'skip').map(row => row.test.split('/')[1]),
      failures: result.windows_private_data_sdk.cases.filter(row => row.action === 'fail').map(row => row.test) } : 'NOT_RUN',
    fixture_exit: result.fixture_exit ?? 'NOT_RUN', error: result.error ?? null, control: structuredClone(result.control ?? null), service: result.service ?? null,
    command_failure: result.command_failure ?? null, errors: [...result.errors ?? []], error_line_total: result.error_line_total ?? 0, error_lines_truncated: result.error_lines_truncated ?? false, model_calls: 0, worker_calls: 0 }
  if (projection.control?.producer && typeof projection.control.producer === 'object') {
    const rows = projection.control.producer.directory_observations
    projection.control.producer.directory_observations = rows.filter((row, index) => row.role === 'data' || rows.findLastIndex(other => other.role === row.role) === index)
      .map(({ phase, role, status, owner_matches_current_user, protected: protectedDacl, ace_count, error_kind }) => ({ phase, role, status,
        ...(status === 'OBSERVED' ? { owner_matches_current_user, protected: protectedDacl, ace_count } : { error_kind: error_kind ?? null }) }))
  }
  while (Buffer.byteLength(JSON.stringify(projection)) > 3500 && projection.errors.length) { projection.errors.pop(); projection.error_lines_truncated = true }
  while (Buffer.byteLength(JSON.stringify(projection)) > 3500 && projection.command_failure?.errors.length) { projection.command_failure.errors.pop(); projection.command_failure.error_lines_truncated = true }
  const text = JSON.stringify(projection)
  assert.ok(Buffer.byteLength(text) <= 3500, 'Public diagnostic metadata exceeds its total bound')
  const escaped = text.replaceAll('%', '%25').replaceAll('\r', '%0D').replaceAll('\n', '%0A')
  process.stdout.write('::warning title=Windows Controller open development control::' + escaped + '\n')
  if (result.unknown_process_observation_projection) {
    const diagnosticProjection = structuredClone(result.unknown_process_observation_projection)
    while (Buffer.byteLength(JSON.stringify(diagnosticProjection)) > 3500 && diagnosticProjection.short_transport?.trace?.length > 2) {
      diagnosticProjection.short_transport.trace.splice(1, 1); diagnosticProjection.short_transport.trace_truncated = true
    }
    while (Buffer.byteLength(JSON.stringify(diagnosticProjection)) > 3500 && diagnosticProjection.samples.length > 2) {
      diagnosticProjection.samples.splice(1, 1); diagnosticProjection.samples_truncated = true
    }
    while (Buffer.byteLength(JSON.stringify(diagnosticProjection)) > 3500 && Array.isArray(diagnosticProjection.short_transport?.ready_events) && diagnosticProjection.short_transport.ready_events.length > 1) {
      const rows = diagnosticProjection.short_transport.ready_events
      const index = rows.findIndex(row => row.event !== 'startup-command-delivery')
      rows.splice(index < 0 ? rows.length - 1 : index, 1)
      diagnosticProjection.short_transport.ready_events_truncated = true
    }
    while (Buffer.byteLength(JSON.stringify(diagnosticProjection)) > 3500 && diagnosticProjection.short_transport?.trace?.length) {
      diagnosticProjection.short_transport.trace.shift(); diagnosticProjection.short_transport.trace_truncated = true
    }
    for (const sample of diagnosticProjection.samples) if (JSON.stringify(sample.first_query_window) === JSON.stringify(sample.last_query_window) && sample.first_query_window !== 'UNKNOWN') sample.last_query_window = 'SAME_AS_FIRST'
    const annotationRaw = diagnosticProjection.short_transport?.raw
    for (const limit of [256, 128, 64]) {
      if (Buffer.byteLength(JSON.stringify(diagnosticProjection)) <= 3500 || annotationRaw?.status !== 'OBSERVED') break
      if (annotationRaw.fragment_bytes <= limit) continue
      const text = annotationRaw.head + annotationRaw.tail
      annotationRaw.head = utf8Fragment(text, limit / 2)
      annotationRaw.tail = utf8Fragment(text, limit / 2, true)
      annotationRaw.fragment_bytes = Buffer.byteLength(annotationRaw.head) + Buffer.byteLength(annotationRaw.tail)
      annotationRaw.truncated = true
    }
    if (Buffer.byteLength(JSON.stringify(diagnosticProjection)) > 3500 && diagnosticProjection.samples.length === 2) {
      diagnosticProjection.samples.shift(); diagnosticProjection.samples_truncated = true
    }
    while (Buffer.byteLength(JSON.stringify(diagnosticProjection)) > 3500 && diagnosticProjection.native_query_counts.length > 1) {
      diagnosticProjection.native_query_counts.splice(1, 1); diagnosticProjection.native_query_counts_truncated = true
    }
    for (const sample of diagnosticProjection.samples) while (Buffer.byteLength(JSON.stringify(diagnosticProjection)) > 3500 && sample.recorded_parent_identities_history.length) {
      sample.recorded_parent_identities_history.pop(); sample.recorded_parent_identities_truncated = true
    }
    const unknownText = JSON.stringify(diagnosticProjection)
    assert.ok(Buffer.byteLength(unknownText) <= 3500)
    process.stdout.write('::warning title=Windows closure unknown observations::' + unknownText.replaceAll('%', '%25').replaceAll('\r', '%0D').replaceAll('\n', '%0A') + '\n')
  }
}
if (failure) process.exitCode = 1
}
main().catch(error => {
  const code = typeof error?.code === 'string' && /^[A-Za-z][A-Za-z0-9_.-]{0,63}$/.test(error.code) ? error.code : 'UNEXPECTED_PREFLIGHT_ERROR'
  process.stderr.write('::error title=Windows Controller open preflight::phase=preflight; kind=' + code + '\n')
  process.exitCode = 1
})
