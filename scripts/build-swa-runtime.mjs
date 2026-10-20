import { createRequire } from 'node:module'
import { mkdir, readFile, writeFile } from 'node:fs/promises'
import { dirname, isAbsolute, join, relative, resolve } from 'node:path'
import { root } from './process.mjs'

const runtime = join(root, 'runtime', 'swa-orca')
const requireRuntime = createRequire(join(runtime, 'package.json'))
const { build } = requireRuntime('esbuild')
const upstream = join(root, 'third_party', 'orca-v1.4.230')

export async function buildSWARuntime(output) {
  if (!isAbsolute(output)) throw new Error('SWA build output must be absolute')
  const target = relative(join(root, 'dist'), output)
  if (target === '..' || target.startsWith('../') || target.startsWith('..\\') || isAbsolute(target)) throw new Error('SWA build output must stay inside dist')
  await mkdir(output, { recursive: true })
  const common = {
    absWorkingDir: root, bundle: true, platform: 'node', target: 'node24', format: 'cjs',
    conditions: ['orca-source'], nodePaths: [join(runtime, 'node_modules'), join(root, 'node_modules')],
    external: ['electron', 'node-pty', '@parcel/watcher', 'fsevents', '@orca/windows-security', '@orca/windows-registry', '@vscode/windows-process-tree'],
    banner: { js: 'var __swaModuleUrl = require("node:url").pathToFileURL(__filename).href;' },
    define: { 'import.meta.url': '__swaModuleUrl', 'process.env.NODE_ENV': '"production"' },
    logOverride: { 'empty-import-meta': 'error' }, metafile: true, sourcemap: false, minify: false,
    plugins: [{ name: 'swa-process-host-source', setup(builder) { builder.onResolve({ filter: /^@orca\/process-host(?:\/|$)/ }, async args => {
      const base = join(upstream, 'src', 'packages', 'process-host')
      const pkg = JSON.parse(await readFile(join(base, 'package.json'), 'utf8'))
      const key = args.path === '@orca/process-host' ? '.' : '.' + args.path.slice('@orca/process-host'.length)
      const entry = pkg.exports[key]?.['orca-source']
      if (typeof entry !== 'string') throw new Error(`Missing exact process-host source export ${key}`)
      return { path: resolve(base, entry) }
    }) } }, { name: 'swa-native-addons', setup(builder) { builder.onResolve({ filter: /\.node$/ }, args => ({ path: args.path, external: true })) } },
      { name: 'swa-jsonc-esm', setup(builder) { builder.onResolve({ filter: /^jsonc-parser$/ }, () => ({ path: join(dirname(requireRuntime.resolve('jsonc-parser/package.json')), 'lib', 'esm', 'main.js') })) } }]
  }
  const entries = {
    'host.cjs': join(runtime, 'src', 'host-entry.ts'),
    'client.cjs': join(runtime, 'src', 'client.ts'),
    'controller-helper.cjs': join(runtime, 'src', 'controller-helper.ts'),
    'native-cli-entry.cjs': join(runtime, 'src', 'native-cli-entry.ts'),
    'strict-caller-probe.cjs': join(root, 'tests', 'swa-cli-headless', 'strict-caller-probe.mjs'),
    'swa-provider-process-tree.cjs': join(upstream, 'src/main/provider-process/swa-provider-process-tree.ts'),
    'managed-provider-process.cjs': join(upstream, 'src/main/provider-process/managed-provider-process.ts'),
    'agent-connection.cjs': join(upstream, 'src/main/jsonl-rpc/agent-connection.ts'),
    'task-result.cjs': join(runtime, 'src', 'task-result.ts'),
    'journal-open.cjs': join(upstream, 'src/main/native-chat/agent-session-journal/journal-open.ts'),
    'parcel-watcher-process-entry.js': join(upstream, 'src/main/ipc/parcel-watcher-process-entry.ts'),
    'daemon-entry.js': join(upstream, 'src/main/daemon/daemon-entry.ts'),
    'profile-state-writer-worker-entry.js': join(upstream, 'src/main/persistence/profile-state/profile-state-writer-worker-entry.ts'),
    'profile-state-backup-worker-entry.js': join(upstream, 'src/main/persistence/profile-state/profile-state-backup-worker-entry.ts'),
    'foreign-sqlite-reader-entry.js': join(upstream, 'src/main/foreign-sqlite-readers/foreign-sqlite-reader-entry.ts'),
    'port-scan-command-worker-entry.js': join(upstream, 'src/main/ports/port-scan-command-worker-entry.ts'),
    'session-scanner-service-entry.js': join(upstream, 'src/main/ai-vault/session-scanner-service-entry.ts'),
    'worker-thread-child-process-entry.js': join(upstream, 'src/main/worker-thread-child-process/worker-thread-child-process-entry.ts'),
    'claude-profile-setup-worker-entry.js': join(upstream, 'src/main/claude-accounts/claude-profile-setup-worker-entry.ts')
  }
  for (const [filename, entry] of Object.entries(entries)) {
    const result = await build({ ...common, entryPoints: [entry], outfile: join(output, filename) })
    await writeFile(join(output, filename + '.metafile.json'), JSON.stringify(result.metafile, null, 2) + '\n', { flag: 'wx' })
  }
  for (const name of ['provider-job-bootstrap']) {
    const result = await build({ ...common, entryPoints: [join(runtime, 'src', name + '.ts')], outfile: join(output, name + '.mjs'), format: 'esm', banner: {}, define: { 'process.env.NODE_ENV': '"production"' }, packages: 'external', external: common.external })
    await writeFile(join(output, name + '.mjs.metafile.json'), JSON.stringify(result.metafile, null, 2) + '\n', { flag: 'wx' })
  }
  const pkg = JSON.parse(await readFile(join(output, 'package.json'), 'utf8'))
  if (pkg.type !== 'commonjs') throw new Error('The deployed runtime must preserve CommonJS child entry semantics')
}

if (process.argv[1] && resolve(process.argv[1]) === resolve(import.meta.filename)) {
  if (process.argv.length !== 3) throw new Error('Usage: build-swa-runtime.mjs ABSOLUTE_DIST_RUNTIME')
  await buildSWARuntime(resolve(process.argv[2]))
}
