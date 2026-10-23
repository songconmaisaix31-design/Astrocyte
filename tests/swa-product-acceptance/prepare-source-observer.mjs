// 验收观察入口准备；只核公开源，不修改产品或读取 Session。
import assert from 'node:assert/strict'
import { execFileSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { constants } from 'node:fs'
import { lstat, mkdir, open, readFile, realpath, stat, writeFile } from 'node:fs/promises'
import { builtinModules, createRequire } from 'node:module'
import { dirname, isAbsolute, join, relative, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'
import { parseArgs } from 'node:util'

const values = parseArgs({ options: Object.fromEntries(['candidate', 'source-repository', 'output', 'profile'].map(key => [key, { type: 'string' }])), allowPositionals: false }).values
for (const key of ['candidate', 'source-repository', 'output']) assert.ok(values[key] && isAbsolute(values[key]), 'observer_absolute_paths_required')
assert.ok(values.profile === undefined || ['stepfun-pi-n3', 'stepfun-pi-questions'].includes(values.profile), 'observer_profile_unreviewed')
const pi = ['stepfun-pi-n3', 'stepfun-pi-questions'].includes(values.profile)
const questions = values.profile === 'stepfun-pi-questions'
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
const questionContractSources = questions ? ['runtime/swa-orca/src/contracts.ts', 'runtime/swa-orca/src/task-budget.ts',
  'third_party/orca-v1.4.230/src/main/runtime/orchestration/orchestration-error.ts'] : []
snapshotPaths.push(...questionContractSources)
const questionExports = questions ? "\nexport { questionRequestID, questionWaitID, readQuestionMutationResult, questionResumeBody } from './source/runtime/swa-orca/src/contracts'\n" : ''
const sha = bytes => createHash('sha256').update(bytes).digest('hex')
const inside = (root, path) => {
  const suffix = relative(root, path)
  return !suffix || suffix !== '..' && !suffix.startsWith('../') && !isAbsolute(suffix)
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
assert.equal(await realpath(process.execPath), await ordinary(join(candidate, 'node/node')), 'observer_requires_actual_installed_node')
const installedRevision = (await readFile(await ordinary(join(candidate, 'VERSION')), 'utf8')).trim()
assert.match(installedRevision, /^[0-9a-f]{40}$/)
const commandOptions = { cwd: repository, env: { ...process.env, GIT_OPTIONAL_LOCKS: '0', GIT_PAGER: 'cat' }, timeout: 30_000, maxBuffer: 64 * 1024 * 1024 }
const sourceRevision = execFileSync('git', ['rev-parse', '--verify', `${installedRevision}^{commit}`], commandOptions).toString('ascii').trim()
assert.equal(sourceRevision, installedRevision, 'observer_installed_commit_missing')
if (!pi) assert.equal(execFileSync(join(candidate, 'swa'), ['--version'], { timeout: 5000, maxBuffer: 4096, encoding: 'utf8' }).trim(), installedRevision, 'observer_actual_cli_version_mismatch')
const requireInstalled = createRequire(join(candidate, 'runtime/package.json'))
const journalModule = await version(await ordinary(requireInstalled.resolve('./journal-open.cjs')))
const protocolModule = pi ? await version(await ordinary(join(candidate, 'runtime/pi-protocol.cjs')), true) : undefined
const protocolMetafile = pi ? await version(await ordinary(join(candidate, 'runtime/pi-protocol.cjs.metafile.json')), true) : undefined
if (pi) assert.equal(execFileSync(join(candidate, 'swa'), ['--version'], { timeout: 5000, maxBuffer: 4096, encoding: 'utf8' }).trim(), installedRevision, 'observer_actual_cli_version_mismatch')
await mkdir(output, { mode: 0o700 })
const archive = execFileSync('git', ['archive', '--format=zip', installedRevision, '--', ...snapshotPaths], commandOptions)
const archivePath = join(output, 'source.zip')
await writeFile(archivePath, archive, { flag: 'wx', mode: 0o600 })
const extraction = execFileSync('python3', ['-I', join(prototype, 'extract-observer-source.py'), archivePath, repository, installedRevision, join(output, 'source'), ...(pi ? ['--profile', values.profile] : [])], { timeout: 60_000, maxBuffer: 4 * 1024 * 1024, encoding: 'utf8' })
const archiveProof = JSON.parse(extraction)
assert.equal(archiveProof.revision, installedRevision)
const entryPath = join(output, 'entry.ts')
const prototypeEntry = await version(join(prototype, pi ? 'pi-source-observer-entry.ts' : 'source-observer-entry.ts'))
const entryBytes = await readFile(prototypeEntry.path)
assert.equal(sha(entryBytes), prototypeEntry.sha256, 'observer_entry_source_changed')
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
const bundlePath = join(output, 'observer.cjs')
let build
try { build = await esbuild.build({
  absWorkingDir: output, entryPoints: [entryPath], outfile: bundlePath, platform: 'node',
  target: 'node24', format: 'cjs', bundle: true, metafile: true, write: false,
  tsconfigRaw: { compilerOptions: { target: 'ES2022', module: 'CommonJS', esModuleInterop: true } },
  logLevel: 'silent',
  plugins: [{ name: 'installed-source-only', setup(plugin) {
    plugin.onResolve({ filter: /^[^./]/ }, async args => {
      if (builtins.has(args.path)) return { path: args.path, external: true }
      const resolved = await ordinary(requireInstalled.resolve(args.path))
      assert.ok(inside(join(candidate, 'runtime/node_modules'), resolved), 'observer_dependency_outside_installation')
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
assert.deepEqual(await version(journalModule.path), journalModule, 'observer_journal_module_changed')
if (pi) for (const source of [protocolModule, protocolMetafile]) assert.deepEqual(await version(source.path, true), source, 'observer_protocol_artifact_changed')
await writeFile(bundlePath, build.outputFiles[0].contents, { flag: 'wx', mode: 0o600 })
const manifest = {
  scope: 'read-only-source-observer-preparation', product_or_model_execution: 'NOT_RUN',
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
}
await writeFile(join(output, 'manifest.json'), JSON.stringify(manifest, null, 2) + '\n', { flag: 'wx', mode: 0o600 })
process.stdout.write(JSON.stringify({ status: 'PREPARED_ONLY', installed_revision: installedRevision, source_entries: archiveProof.files, bundled_inputs: seenInputs.size, observer_sha256: manifest.observer.sha256, model_execution: 'NOT_RUN' }) + '\n')
