// 验收观察入口准备；只核公开源，不修改产品或读取 Session。
import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { constants } from 'node:fs'
import { lstat, mkdir, open, readFile, realpath, stat, writeFile } from 'node:fs/promises'
import { builtinModules, createRequire, isBuiltin } from 'node:module'
import { dirname, isAbsolute, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { parseArgs } from 'node:util'

const values = parseArgs({ options: Object.fromEntries(['candidate', 'source-repository', 'output', 'profile'].map(key => [key, { type: 'string' }])), allowPositionals: false }).values
for (const key of ['candidate', 'source-repository', 'output']) assert.ok(values[key] && isAbsolute(values[key]), 'observer_absolute_paths_required')
assert.ok(values.profile === undefined || ['stepfun-pi-n3', 'stepfun-pi-questions', 'stepfun-opencode', 'opencode-initialize-only'].includes(values.profile), 'observer_profile_unreviewed')
const pi = ['stepfun-pi-n3', 'stepfun-pi-questions'].includes(values.profile)
const questions = values.profile === 'stepfun-pi-questions'
const openCode = values.profile === 'stepfun-opencode'
const initializeOpenCode = values.profile === 'opencode-initialize-only'
if (initializeOpenCode) assert.ok(['linux', 'win32'].includes(process.platform) && process.arch === 'x64', 'initializer_platform_unreviewed')
process.umask(0o077)
const prototype = dirname(fileURLToPath(import.meta.url))
const snapshotPaths = [
  'third_party/orca-v1.4.230/src/shared',
  'third_party/orca-v1.4.230/src/main/daemon/swa-linux-process-identity.ts',
  'third_party/orca-v1.4.230/src/main/agent-hooks/managed-hook-owner-identity.ts',
  'third_party/orca-v1.4.230/src/main/pty-process-table-parser.ts',
  'third_party/orca-v1.4.230/package.json',
  'third_party/orca-v1.4.230/tsconfig.json',
  'third_party/orca-v1.4.230/config/tsconfig.node.json',
]
if (pi) snapshotPaths.push('third_party/orca-v1.4.230/src/main/pi/rpc-protocol.ts', 'scripts/build-swa-runtime.mjs',
  'third_party/orca-v1.4.230/src/main/runtime/rpc/orchestration-mutation-receipt.ts',
  'third_party/orca-v1.4.230/src/main/native-chat/agent-session-timeline/provider-timeline-identity.ts',
  'third_party/orca-v1.4.230/src/main/native-chat/agent-session-journal/journal-payload-bounds.ts')
if (openCode) snapshotPaths.push(
  'third_party/orca-v1.4.230/src/main/runtime/rpc/orchestration-mutation-receipt.ts',
  'third_party/orca-v1.4.230/src/main/native-chat/agent-session-timeline/provider-timeline-identity.ts',
  'third_party/orca-v1.4.230/src/main/native-chat/agent-session-journal/journal-payload-bounds.ts',
  'third_party/orca-v1.4.230/src/main/acp/acp-rpc-envelope.ts',
  'third_party/orca-v1.4.230/src/main/acp/acp-errors.ts',
  'third_party/orca-v1.4.230/src/main/acp/generated/acp-protocol.generated.ts')
if (initializeOpenCode) snapshotPaths.push(
  ...['native-opencode-policy.ts', 'private-opencode-account.ts', 'native-opencode-import.ts', 'native-pi-policy.ts',
    'native-codex-policy.ts', 'private-posix-security.ts', 'contracts.ts', 'task-budget.ts', 'pi-official-sources.json']
    .map(name => `runtime/swa-orca/src/${name}`),
  'third_party/orca-v1.4.230/src/packages/process-host/package.json',
  'third_party/orca-v1.4.230/src/packages/process-host/src',
  ...[
    'acp/acp-session-runtime.ts', 'acp/acp-errors.ts', 'acp/acp-response-settlement.ts', 'acp/acp-json-rpc-peer.ts',
    'acp/acp-permission-requests.ts', 'acp/acp-session-events.ts', 'acp/acp-session-setup.ts', 'acp/acp-incoming-requests.ts',
    'acp/acp-peer-limits.ts', 'acp/acp-oversized-lines.ts', 'acp/acp-write-queue.ts', 'acp/acp-stdio-error-boundary.ts',
    'acp/acp-rpc-envelope.ts', 'acp/generated/acp-protocol.generated.ts', 'zod-schema-compilation.ts',
    'provider-process/managed-provider-process.ts', 'provider-process/swa-provider-process-tree.ts',
    'provider-process/swa-linux-provider-process-tree.ts', 'provider-process/provider-process-close.ts',
    'provider-process/provider-process-supervisor.ts', 'provider-process/provider-process-teardown.ts',
    'provider-process/provider-executable-missing.ts', 'provider-process/provider-spawn-failure-report.ts',
    'provider-process/provider-process-exit-deadline.ts', 'provider-process/provider-process-launch.ts',
    'provider-process/provider-stdio-write-queue.ts', 'provider-process/provider-stdio-error-boundary.ts',
    'windows/windows-process-table.ts', 'windows/windows-process-table-timeout-error.ts',
    'windows/windows-process-table-cim-scan.ts', 'windows/windows-command-line-recovery-health.ts',
    'windows-process-tree-kill.ts', 'windows-pty-root-identity.ts', 'pty-descendant-exit-verification.ts',
    'pty-descendant-termination.ts', 'own-chromium-tree-kill-guard.ts', 'orca-chromium-process-pids.ts',
    'providers/windows-foreground-process-rows.ts', 'crash-reporting/self-initiated-tree-kill-log.ts',
    'crash-reporting/durable-crash-breadcrumb.ts', 'crash-reporting/main-process-lifecycle-identity.ts',
    'crash-reporting/crash-breadcrumb-store.ts', 'observability/tracer.ts', 'observability/redactor.ts',
    'agent-hooks/windows-powershell-hook-launcher.ts', 'pi/rpc-protocol.ts', 'runtime/orchestration/orchestration-error.ts',
    'opencode/opencode-structured-account-home.ts', 'managed-data-accounts/service.ts',
    'managed-data-accounts/credential-capture.ts', 'managed-data-accounts/profile-removal.ts',
    'host-tree-removal.ts', 'asar-transparent-fs.ts', 'opencode-usage/schema-helpers.ts',
    'sqlite/sync-database.ts', 'sqlite/node-sqlite-reader-api.ts', 'sqlite/node-sqlite-statement.ts', 'sqlite/sqlite-integer-reader.ts'
  ].map(name => `third_party/orca-v1.4.230/src/main/${name}`))
const questionContractSources = questions ? ['runtime/swa-orca/src/contracts.ts', 'runtime/swa-orca/src/task-budget.ts',
  'third_party/orca-v1.4.230/src/main/runtime/orchestration/orchestration-error.ts'] : []
snapshotPaths.push(...questionContractSources)
const questionExports = questions ? "\nexport { questionRequestID, questionWaitID, readQuestionMutationResult, questionResumeBody } from './source/runtime/swa-orca/src/contracts'\n" : ''
const sha = bytes => createHash('sha256').update(bytes).digest('hex')
const inside = (root, path) => {
  const suffix = relative(root, path)
  return !suffix || suffix !== '..' && !suffix.startsWith('../') && !suffix.startsWith('..\\') && !isAbsolute(suffix)
}
async function ordinary(path, directory = false) {
  const canonical = await realpath(path)
  assert.equal(canonical, resolve(path), 'observer_redirected_path')
  for (let current = canonical; ; current = dirname(current)) {
    const value = await lstat(current)
    assert.ok(!value.isSymbolicLink(), 'observer_link_path')
    assert.ok(current === canonical ? directory ? value.isDirectory() : value.isFile() : value.isDirectory(), 'observer_ordinary_path_required')
    if (dirname(current) === current) return canonical
  }
}
const serialize = value => `${value.dev}:${value.ino}:${value.birthtimeNs}:${value.mtimeNs}:${value.size}`
async function boundedArtifact(path) {
  await ordinary(path)
  const before = await stat(path, { bigint: true })
  assert.ok(before.size > 0n && before.size <= 4n * 1024n * 1024n, 'observer_protocol_artifact_bound')
  const handle = await open(path, constants.O_RDONLY | constants.O_NOFOLLOW)
  try {
    assert.equal(serialize(await handle.stat({ bigint: true })), serialize(before), 'observer_protocol_fd_changed')
    const buffer = Buffer.alloc(Number(before.size) + 1)
    let count = 0
    while (count < buffer.length) {
      const value = await handle.read(buffer, count, buffer.length - count, null)
      if (!value.bytesRead) break
      count += value.bytesRead
    }
    assert.equal(count, Number(before.size), 'observer_protocol_changed_during_read')
    assert.equal(serialize(await handle.stat({ bigint: true })), serialize(before), 'observer_protocol_fd_changed')
    assert.equal(serialize(await stat(path, { bigint: true })), serialize(before), 'observer_protocol_path_changed')
    return { bytes: buffer.subarray(0, count), info: before }
  } finally { await handle.close() }
}
async function version(path, protocolArtifact = false) {
  if (protocolArtifact) {
    const { bytes, info } = await boundedArtifact(path)
    return { path, version: serialize(info), sha256: sha(bytes) }
  }
  await ordinary(path)
  const before = await stat(path, { bigint: true })
  const bytes = await readFile(path)
  const after = await stat(path, { bigint: true })
  assert.equal(serialize(after), serialize(before), 'observer_source_changed_during_read')
  return { path, version: serialize(before), sha256: sha(bytes) }
}
const candidate = await ordinary(values.candidate, true)
const repository = await ordinary(values['source-repository'], true)
await ordinary(dirname(values.output), true)
const output = resolve(values.output)
assert.ok(!inside(candidate, output) && !inside(output, candidate), 'observer_output_overlaps_installation')
const installedNode = join(candidate, initializeOpenCode && process.platform === 'win32' ? 'node/node.exe' : 'node/node')
const installedCli = join(candidate, initializeOpenCode && process.platform === 'win32' ? 'swa.exe' : 'swa')
assert.equal(await realpath(process.execPath), await ordinary(installedNode), 'observer_requires_actual_installed_node')
const installedRevision = (await readFile(await ordinary(join(candidate, 'VERSION')), 'utf8')).trim()
assert.match(installedRevision, /^[0-9a-f]{40}$/)
const commandOptions = { cwd: repository, env: { ...process.env, GIT_OPTIONAL_LOCKS: '0', GIT_PAGER: 'cat' }, timeout: 30_000, maxBuffer: 64 * 1024 * 1024 }
const sourceRevision = execFileSync('git', ['rev-parse', '--verify', `${installedRevision}^{commit}`], commandOptions).toString('ascii').trim()
assert.equal(sourceRevision, installedRevision, 'observer_installed_commit_missing')
if (!pi) assert.equal(execFileSync(installedCli, ['--version'], { timeout: 5000, maxBuffer: 4096, encoding: 'utf8' }).trim(), installedRevision, 'observer_actual_cli_version_mismatch')
const requireInstalled = createRequire(join(candidate, 'runtime/package.json'))
const bundlePath = join(output, 'observer.cjs')
const installedExternals = new Map()
const dynamicDependencies = []
let initializationEnvironment
if (initializeOpenCode) {
  const nodePath = await ordinary(join(candidate, 'runtime/node_modules'), true)
  assert.equal(process.env.NODE_PATH, nodePath, 'initializer_node_path_must_match_installation')
  assert.ok(!Object.keys(process.env).some(key => key.toUpperCase() === 'NODE_OPTIONS'), 'initializer_node_options_forbidden')
  const requireBundle = createRequire(bundlePath)
  for (const name of ['node-pty/lib/utils', '@orca/windows-security', ...(process.platform === 'win32' ? ['@vscode/windows-process-tree'] : [])]) {
    const path = await ordinary(requireBundle.resolve(name))
    const installed = await ordinary(requireInstalled.resolve(name))
    const packageName = name === 'node-pty/lib/utils' ? 'node-pty' : name
    assert.equal(path, installed, 'initializer_runtime_dependency_differs_from_installation')
    assert.ok(inside(join(nodePath, packageName), path), 'initializer_runtime_dependency_outside_package')
    dynamicDependencies.push({ name, ...await version(path) })
  }
  initializationEnvironment = { platform: process.platform, architecture: process.arch, node_path: nodePath, node_options_absent: true, bundle_path: bundlePath }
}
const journalModule = await version(await ordinary(requireInstalled.resolve('./journal-open.cjs')))
const protocolModule = pi ? await version(await ordinary(join(candidate, 'runtime/pi-protocol.cjs')), true) : undefined
const protocolMetafile = pi ? await version(await ordinary(join(candidate, 'runtime/pi-protocol.cjs.metafile.json')), true) : undefined
if (pi) assert.equal(execFileSync(join(candidate, 'swa'), ['--version'], { timeout: 5000, maxBuffer: 4096, encoding: 'utf8' }).trim(), installedRevision, 'observer_actual_cli_version_mismatch')
await mkdir(output, { mode: 0o700 })
const archive = execFileSync('git', ['archive', '--format=zip', installedRevision, '--', ...snapshotPaths], commandOptions)
const archivePath = join(output, 'source.zip')
await writeFile(archivePath, archive, { flag: 'wx', mode: 0o600 })
const extraction = execFileSync('python3', ['-I', join(prototype, 'extract-observer-source.py'), archivePath, repository, installedRevision, join(output, 'source'), ...(pi || openCode || initializeOpenCode ? ['--profile', values.profile] : [])], { timeout: 60_000, maxBuffer: 4 * 1024 * 1024, encoding: 'utf8' })
const archiveProof = JSON.parse(extraction)
assert.equal(archiveProof.revision, installedRevision)
const entryPath = join(output, 'entry.ts')
const entryName = initializeOpenCode ? 'opencode-initialization-entry.ts' : openCode ? 'opencode-source-observer-entry.ts' : pi ? 'pi-source-observer-entry.ts' : 'source-observer-entry.ts'
const prototypeEntry = await version(join(prototype, entryName))
const entryBytes = await readFile(prototypeEntry.path)
assert.equal(sha(entryBytes), prototypeEntry.sha256, 'observer_entry_source_changed')
if (openCode || initializeOpenCode) assert.ok(execFileSync('git', ['show', `${installedRevision}:tests/swa-product-acceptance/${entryName}`], commandOptions).equals(entryBytes), 'observer_entry_differs_from_installed_revision')
await writeFile(entryPath, Buffer.concat([entryBytes, Buffer.from(questionExports, 'utf8')]), { flag: 'wx', mode: 0o600 })
// 编译器单独记录为既有工具；产品源码与依赖只取本安装版本快照和安装原件。
const requireCompiler = createRequire(new URL('../../runtime/swa-orca/package.json', import.meta.url))
const esbuildPath = await ordinary(requireCompiler.resolve('esbuild'))
const esbuild = requireCompiler('esbuild')
assert.equal(esbuild.version, '0.28.2', 'observer_compiler_version_unreviewed')
let protocolMapping, typescriptModule
if (pi) {
  const ts = requireCompiler('typescript')
  typescriptModule = await version(await ordinary(requireCompiler.resolve('typescript')))
  const builderPath = 'scripts/build-swa-runtime.mjs'
  const builder = ts.createSourceFile(builderPath, await readFile(join(output, 'source', builderPath), 'utf8'), ts.ScriptTarget.Latest, true, ts.ScriptKind.JS)
  const mappings = []
  const visit = node => {
    if (ts.isVariableDeclaration(node) && ts.isIdentifier(node.name) && node.name.text === 'entries') {
      assert.ok(node.initializer && ts.isObjectLiteralExpression(node.initializer), 'observer_builder_entries_not_literal')
      for (const property of node.initializer.properties) {
        if (!ts.isPropertyAssignment(property) || !ts.isStringLiteral(property.name) || property.name.text !== 'pi-protocol.cjs') continue
        const call = property.initializer
        assert.ok(ts.isCallExpression(call) && ts.isIdentifier(call.expression) && call.expression.text === 'join' && call.arguments.length === 2 && ts.isIdentifier(call.arguments[0]) && call.arguments[0].text === 'upstream' && ts.isStringLiteral(call.arguments[1]), 'observer_protocol_mapping_unreviewed')
        mappings.push(call.arguments[1].text)
      }
    }
    ts.forEachChild(node, visit)
  }
  visit(builder)
  assert.deepEqual(mappings, ['src/main/pi/rpc-protocol.ts'], 'observer_protocol_mapping_changed')
  const protocolSource = 'third_party/orca-v1.4.230/' + mappings[0]
  protocolMapping = { builder: archiveProof.entries.find(row => row.path === builderPath), source: archiveProof.entries.find(row => row.path === protocolSource) }
  assert.ok(protocolMapping.builder && protocolMapping.source, 'observer_protocol_git_source_missing')
  const metaBytes = await boundedArtifact(protocolMetafile.path)
  assert.equal(sha(metaBytes.bytes), protocolMetafile.sha256, 'observer_protocol_metafile_changed')
  const meta = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(metaBytes.bytes))
  const outputs = Object.values(meta.outputs)
  assert.equal(outputs.length, 1, 'observer_protocol_output_inventory')
  assert.equal(outputs[0].entryPoint.replaceAll('\\', '/').replace(/^.*\/third_party\//, 'third_party/'), protocolSource, 'observer_protocol_actual_entrypoint_changed')
  assert.deepEqual(outputs[0].imports, [], 'observer_protocol_external_dependency_unreviewed')
  const inputs = Object.keys(meta.inputs).map(path => path.replaceAll('\\', '/'))
  assert.ok(inputs.some(path => path.endsWith(protocolSource)) && inputs.some(path => path.includes('/jsonc-parser/') && path.endsWith('/lib/esm/main.js')) && inputs.some(path => path.includes('/zod/')), 'observer_protocol_actual_inputs_missing')
  assert.ok(inputs.every(path => !/claude-agent-sdk|pi-coding-agent|pi-ai|native-pi-policy|pi-session-policy|process-host/.test(path)), 'observer_protocol_effectful_input_unreviewed')
  if (questions) {
    const protocol = requireInstalled(protocolModule.path)
    for (const name of ['strictPiQuestionRawSchema', 'strictPiQuestionNativeEntrySchema', 'strictPiQuestionResultSchema']) assert.equal(typeof protocol[name]?.safeParse, 'function', 'observer_installed_question_schema_missing')
  }
}
const builtins = new Set(builtinModules.flatMap(name => [name, `node:${name}`]))
const seenInputs = new Map()
const processHostRoot = join(output, 'source/third_party/orca-v1.4.230/src/packages/process-host')
const processHostPackage = initializeOpenCode ? await version(join(processHostRoot, 'package.json')) : undefined
const processHostExports = initializeOpenCode ? JSON.parse(await readFile(processHostPackage.path, 'utf8')).exports : undefined
let build
try { build = await esbuild.build({
  absWorkingDir: output, entryPoints: [entryPath], outfile: bundlePath, platform: 'node',
  target: 'node24', format: 'cjs', bundle: true, metafile: true, write: false,
  tsconfigRaw: { compilerOptions: { target: 'ES2022', module: 'CommonJS', esModuleInterop: true } },
  logLevel: 'silent',
  plugins: [{ name: 'installed-source-only', setup(plugin) {
    plugin.onResolve({ filter: /^[^./]/ }, async args => {
      if (builtins.has(args.path) || initializeOpenCode && isBuiltin(args.path)) return { path: args.path, external: true }
      if (initializeOpenCode && isAbsolute(args.path)) return undefined
      if (initializeOpenCode && (args.path === '@orca/process-host' || args.path.startsWith('@orca/process-host/'))) {
        const subpath = args.path === '@orca/process-host' ? '.' : '.' + args.path.slice('@orca/process-host'.length)
        const entry = processHostExports[subpath]?.['orca-source']
        assert.ok(typeof entry === 'string' && entry.startsWith('./src/'), 'initializer_process_host_export_unreviewed')
        const path = await ordinary(resolve(processHostRoot, entry))
        assert.ok(inside(join(processHostRoot, 'src'), path), 'initializer_process_host_export_outside_source')
        return { path }
      }
      const resolved = await ordinary(requireInstalled.resolve(args.path))
      assert.ok(inside(join(candidate, 'runtime/node_modules'), resolved), 'observer_dependency_outside_installation')
      if (initializeOpenCode && args.path === '@orca/windows-security') {
        const runtime = dynamicDependencies.find(entry => entry.name === args.path)
        assert.equal(resolved, runtime.path, 'initializer_external_runtime_mismatch')
        installedExternals.set(resolved, await version(resolved))
        return { path: resolved, external: true }
      }
      return { path: resolved }
    })
    plugin.onLoad({ filter: /.*/ }, async args => {
      const path = await ordinary(args.path)
      assert.ok(path === entryPath || inside(join(output, 'source'), path) || inside(join(candidate, 'runtime/node_modules'), path), 'observer_compiler_input_outside_snapshot')
      seenInputs.set(path, await version(path))
      return undefined
    })
  } }],
}) } finally { await esbuild.stop() }
assert.equal(build.outputFiles.length, 1, 'observer_unexpected_compiler_outputs')
for (const path of Object.keys(build.metafile.inputs)) assert.ok(seenInputs.has(resolve(output, path)), 'observer_untracked_metafile_input')
for (const [path, expected] of seenInputs) assert.deepEqual(await version(path), expected, 'observer_source_changed_during_build')
if (initializeOpenCode) {
  assert.deepEqual(await version(processHostPackage.path), processHostPackage, 'initializer_process_host_package_changed')
  for (const expected of installedExternals.values()) assert.deepEqual(await version(expected.path), expected, 'initializer_external_changed')
  for (const { name: _name, ...expected } of dynamicDependencies) assert.deepEqual(await version(expected.path), expected, 'initializer_dynamic_dependency_changed')
}
assert.deepEqual(await version(journalModule.path), journalModule, 'observer_journal_module_changed')
if (pi) for (const source of [protocolModule, protocolMetafile]) assert.deepEqual(await version(source.path, true), source, 'observer_protocol_artifact_changed')
await writeFile(bundlePath, build.outputFiles[0].contents, { flag: 'wx', mode: 0o600 })
const manifest = {
  scope: initializeOpenCode ? 'controlled-opencode-initialization-client-preparation' : 'read-only-source-observer-preparation', product_or_model_execution: 'NOT_RUN',
  installed_revision: installedRevision, archive_sha256: sha(archive), archive_proof: archiveProof,
  compiler: { version: esbuild.version, module: await version(esbuildPath) },
  journal_module: journalModule, inputs: [...seenInputs.values()], metafile: build.metafile,
  observer: await version(bundlePath),
  ...(questions ? { entry_source: prototypeEntry, entry_exports: questionExports, question_contract_sources: questionContractSources.map(path => {
    const source = archiveProof.entries.find(entry => entry.path === path)
    assert.ok(source, 'observer_question_contract_git_source_missing')
    return source
  }) } : {}),
  ...(pi ? { profile: values.profile, protocol_module: protocolModule, protocol_metafile: protocolMetafile, protocol_mapping: protocolMapping, typescript_module: typescriptModule } : {}),
  ...(openCode ? { profile: values.profile, entry_source: prototypeEntry } : {}),
  ...(initializeOpenCode ? { profile: values.profile, entry_source: prototypeEntry, process_host_package: processHostPackage,
    runtime_environment: initializationEnvironment, installed_external_entries: [...installedExternals.values()], dynamic_dependencies: dynamicDependencies } : {}),
}
await writeFile(join(output, 'manifest.json'), JSON.stringify(manifest, null, 2) + '\n', { flag: 'wx', mode: 0o600 })
process.stdout.write(JSON.stringify({ status: 'PREPARED_ONLY', installed_revision: installedRevision, source_entries: archiveProof.files, bundled_inputs: seenInputs.size, observer_sha256: manifest.observer.sha256, model_execution: 'NOT_RUN' }) + '\n')
