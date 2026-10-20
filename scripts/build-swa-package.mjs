import { execFileSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import { access, copyFile, cp, lstat, mkdir, mkdtemp, readFile, realpath, writeFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { dirname, isAbsolute, join, relative, resolve } from 'node:path'
import { parseArgs } from 'node:util'
import { root, findGo } from './process.mjs'
import { NODE_RUNTIME_PIN, NODE_RUNTIME_ASSETS, nodeRuntimeReleaseUrl } from '../third_party/orca-v1.4.230/src/shared/node-runtime-pin.ts'
import { buildSWARuntime } from './build-swa-runtime.mjs'
import { slotSourceFiles } from '../third_party/orca-v1.4.230/config/scripts/orcad-prebuild-slot-contents.mjs'

const { values } = parseArgs({ options: { output: { type: 'string' } }, allowPositionals: false })
if (process.platform !== 'win32' || process.arch !== 'x64') throw new Error('This candidate targets Windows x64')
if (!values.output || !isAbsolute(values.output)) throw new Error('An explicit absolute --output inside dist is required')
const output = resolve(values.output), beneath = relative(join(root, 'dist'), output)
if (!beneath || beneath === '..' || beneath.startsWith('..\\') || isAbsolute(beneath)) throw new Error('Choose a new candidate directory inside dist')
const pnpm = process.env.SWA_PNPM_CLI
if (!pnpm || !isAbsolute(pnpm)) throw new Error('SWA_PNPM_CLI must identify the installed pinned pnpm JavaScript entry')
await access(pnpm)
const runtimeSource = join(root, 'runtime', 'swa-orca')
const sourceRequire = createRequire(join(runtimeSource, 'package.json'))
const { unzipSync } = sourceRequire('fflate')
const gyp = sourceRequire.resolve('node-gyp/bin/node-gyp.js')
const python = await realpath('C:/Python313/python.exe')
const taskWork = join(root, '.verification', 'swa-cli-headless-20261020')
const store = join(taskWork, 'pnpm-store')
await mkdir(store, { recursive: true })
if ((await lstat(store)).isSymbolicLink() || (await realpath(store)).toLowerCase() !== join(await realpath(root), '.verification', 'swa-cli-headless-20261020', 'pnpm-store').toLowerCase()) throw new Error('The private pnpm store is redirected')
const temporaryRoot = join(taskWork, 'build')
await mkdir(temporaryRoot, { recursive: true })
const work = await mkdtemp(join(temporaryRoot, 'candidate-'))
const temp = join(work, 'temp')
await mkdir(temp)
const environment = { ...process.env, TEMP: temp, TMP: temp, TMPDIR: temp, NODE_OPTIONS: '--max-old-space-size=1024', GOMAXPROCS: '4', GOMEMLIMIT: '512MiB', NODE_GYP_FORCE_PYTHON: python }
const run = (program, args, cwd = root) => execFileSync(program, args, { cwd, env: environment, stdio: 'inherit', windowsHide: true, timeout: 20 * 60_000 })
const capture = (program, args) => execFileSync(program, args, { cwd: root, env: environment, encoding: 'utf8', windowsHide: true, timeout: 60_000 }).trim()
if (capture(process.execPath, [pnpm, '--version']) !== '11.27.0') throw new Error('The build requires pnpm 11.27.0')
const pythonInfo = JSON.parse(capture(python, ['-c', 'import json,sys; print(json.dumps({"executable":sys.executable,"version":list(sys.version_info[:3])}))']))
if (!pythonInfo || typeof pythonInfo.executable !== 'string' || !isAbsolute(pythonInfo.executable) || (await realpath(pythonInfo.executable)).toLowerCase() !== python.toLowerCase()) throw new Error('The actual Python executable differs from the selected native build input')
const pythonVersion = pythonInfo.version
if (!Array.isArray(pythonVersion) || pythonVersion.length !== 3 || !pythonVersion.every(Number.isSafeInteger) || pythonVersion.some(value => value < 0) || pythonVersion[0] < 3 || pythonVersion[0] === 3 && pythonVersion[1] < 6) throw new Error('The actual selected Python version is unsupported by node-gyp')
console.log(JSON.stringify({ stage: 'native-python-preflight', executable: python, version: pythonVersion.join('.') }))
const distribution = join(root, 'dist')
await mkdir(distribution, { recursive: true })
const actualDistribution = await realpath(distribution)
if ((await lstat(distribution)).isSymbolicLink() || actualDistribution.toLowerCase() !== join(await realpath(root), 'dist').toLowerCase()) throw new Error('The distribution root must be the actual repository directory')
const parent = await realpath(dirname(output)), parentRelative = relative(actualDistribution, parent)
if (parentRelative === '..' || parentRelative.startsWith('..\\') || isAbsolute(parentRelative)) throw new Error('The candidate parent escaped the actual distribution root')
// 只创建受控父目录；候选目录仍独占创建，已有路径立即失败。
await mkdir(output)
const runtime = join(output, 'runtime')
run(process.execPath, [pnpm, '--filter', '@swa/orca-runtime', '--prod', 'deploy', '--ignore-scripts', '--verify-store-integrity', runtime, '--store-dir', store])
const revision = capture('git', ['rev-parse', 'HEAD'])
if (!/^[0-9a-f]{40}$/.test(revision)) throw new Error('Exact source revision is unavailable')
if (capture('git', ['status', '--porcelain', '--untracked-files=normal'])) throw new Error('Commit the candidate source before building the installable package')
await buildSWARuntime(runtime)
const nodeDirectory = join(output, 'node')
await mkdir(nodeDirectory)
const nodeAsset = NODE_RUNTIME_ASSETS['win32-x64']
if (!process.env.ProgramFiles || !isAbsolute(process.env.ProgramFiles)) throw new Error('The installed PowerShell location is unavailable')
const powershell = await realpath(join(process.env.ProgramFiles, 'PowerShell', '7', 'pwsh.exe'))
const downloader = join(root, 'scripts', 'download-swa-node.ps1')
await access(downloader)

async function fetchPinned(file, expected, target) {
  execFileSync(powershell, ['-NoProfile', '-NonInteractive', '-File', downloader, '-Uri', nodeRuntimeReleaseUrl('official', file), '-WorkDirectory', work, '-OutputFile', target], { cwd: root, env: environment, stdio: 'inherit', windowsHide: true, timeout: 10 * 60_000 })
  const body = await readFile(target)
  if (createHash('sha256').update(body).digest('hex') !== expected) throw new Error(`Pinned Node input differs: ${file}`)
  return body
}

const archive = await fetchPinned(nodeAsset.archive, nodeAsset.archiveSha256, join(work, nodeAsset.archive))
const top = `node-v${NODE_RUNTIME_PIN.version}-win-x64`
const selected = unzipSync(archive, { filter: file => file.name === `${top}/node.exe` || file.name === `${top}/LICENSE` })
const executable = selected[`${top}/node.exe`]
if (!executable || executable.length !== nodeAsset.executableSize || createHash('sha256').update(executable).digest('hex') !== nodeAsset.executableSha256) throw new Error('Pinned Node executable differs')
if (!selected[`${top}/LICENSE`]) throw new Error('Pinned Node license is missing')
const node = join(nodeDirectory, 'node.exe')
await writeFile(node, executable, { flag: 'wx' })
await writeFile(join(nodeDirectory, 'LICENSE'), selected[`${top}/LICENSE`], { flag: 'wx' })
if (capture(node, ['--version']) !== 'v' + NODE_RUNTIME_PIN.version) throw new Error('Installed Node version differs from its pin')
const headers = join(work, NODE_RUNTIME_PIN.headers.file)
await fetchPinned(NODE_RUNTIME_PIN.headers.file, NODE_RUNTIME_PIN.headers.sha256, headers)
const headerRoot = join(work, 'headers')
await mkdir(headerRoot)
run(join(process.env.SystemRoot, 'System32', 'tar.exe'), ['-xzf', headers, '-C', headerRoot])
const nodeDir = join(headerRoot, `node-v${NODE_RUNTIME_PIN.version}`)
await mkdir(join(nodeDir, 'Release'))
const library = NODE_RUNTIME_PIN.windowsImportLibs['win32-x64']
await fetchPinned(library.file, library.sha256, join(nodeDir, 'Release', 'node.lib'))
environment.PATH = nodeDirectory + ';' + environment.PATH
environment.npm_config_msvs_version = '2022'
environment.npm_config_devdir = join(work, 'gyp-cache')
// 原生源码使用UTF-8；编译参数只作用于本候选的子进程。
for (const key of Object.keys(environment)) if (/^(?:CL|_CL_)$/i.test(key)) delete environment[key]
environment.CL = '/utf-8'
console.log(JSON.stringify({ stage: 'native-msvc-options', CL: environment.CL }))
const installed = createRequire(join(runtime, 'package.json'))
for (const name of ['node-pty', '@vscode/windows-process-tree', '@orca/windows-security', '@orca/windows-registry']) {
  const packageDir = await realpath(dirname(installed.resolve(name + '/package.json')))
  const rel = relative(output, packageDir)
  if (!rel || rel === '..' || rel.startsWith('..\\') || isAbsolute(rel)) throw new Error('Native build target escaped the private candidate')
  if ((await lstat(packageDir)).isSymbolicLink()) throw new Error('Native build target is redirected')
  if (name === 'node-pty') await copyFile(join(root, 'third_party', 'orca-v1.4.230', 'native', 'swa-provider-job.h'), join(packageDir, 'src', 'win', 'swa_provider_job.h'))
  if (name === '@vscode/windows-process-tree') {
    const headersDir = dirname(createRequire(join(packageDir, 'package.json')).resolve('node-addon-api/package.json'))
    await cp(headersDir, join(packageDir, 'deps', 'node-addon-api'), { recursive: true, dereference: true, errorOnExist: true })
  }
  run(node, [gyp, 'rebuild', '--python=' + python, '--arch=x64', '--jobs=4', '--nodedir=' + nodeDir], packageDir)
  if (name === 'node-pty') {
    const buildDir = join(packageDir, 'build', 'Release')
    for (const [file, source] of slotSourceFiles({ platform: 'win32', arch: 'x64', buildDir, nodePtyDir: packageDir })) {
      const destination = join(buildDir, file)
      if (resolve(source) === resolve(destination)) continue
      await mkdir(dirname(destination), { recursive: true })
      await copyFile(source, destination)
    }
  }
}
const go = await findGo()
for (const [name, entry] of [['swa.exe', './cmd/swa'], ['server.exe', './cmd/server']]) {
  const args = ['build', '-p=4', '-trimpath']
  if (name === 'swa.exe') args.push('-ldflags=-X main.version=' + revision)
  run(go, [...args, '-o', join(output, name), entry])
}
run(go, ['build', '-p=4', '-trimpath', '-o', join(runtime, 'controller-fault.exe'), './tests/swa-cli-headless/controller-fault'])
for (const [source, target] of [['LICENSE', 'LICENSE'], ['THIRD_PARTY_SWA.md', 'THIRD_PARTY_SWA.md'], ['third_party/orca-v1.4.230/LICENSE', 'ORCA_LICENSE'], ['scripts/install-swa.ps1', 'install-swa.ps1'], ['runtime/swa-orca/USAGE.md', 'USAGE.md']]) await copyFile(join(root, source), join(output, target))
await writeFile(join(output, 'VERSION'), revision + '\n', { flag: 'wx' })
console.log(JSON.stringify({ source_revision: revision, candidate: output, node: NODE_RUNTIME_PIN.version, native_modules_built: ['node-pty', '@vscode/windows-process-tree', '@orca/windows-security', '@orca/windows-registry'] }))
