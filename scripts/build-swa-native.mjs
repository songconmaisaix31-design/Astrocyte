import { execFileSync } from 'node:child_process'
import { copyFile, cp, lstat, mkdir, readdir, realpath } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { dirname, isAbsolute, join, relative, resolve } from 'node:path'
import { root } from './process.mjs'
import { slotSourceFiles } from '../third_party/orca-v1.4.230/config/scripts/orcad-prebuild-slot-contents.mjs'

export const nativePackages = ['node-pty', '@vscode/windows-process-tree', '@orca/windows-security', '@orca/windows-registry']

export async function buildNativePackage({ name, index, runtime, work, node, nodeDir, python, gyp, environment }) {
  if (nativePackages[index] !== name) throw new Error('Unknown native package build')
  const installed = createRequire(join(runtime, 'package.json'))
  const packageDir = await realpath(dirname(installed.resolve(name + '/package.json')))
  const packageRelative = relative(await realpath(runtime), packageDir)
  if (!packageRelative || packageRelative.startsWith('..') || isAbsolute(packageRelative)) throw new Error('Native source escaped the controlled runtime')
  const source = join(work, 'n' + index), analysisDirectory = join(work, 'a' + index)
  await mkdir(source)
  await mkdir(analysisDirectory)
  for (const directory of [source, analysisDirectory]) {
    if ((await lstat(directory)).isSymbolicLink() || (await realpath(directory)).toLowerCase() !== resolve(directory).toLowerCase()) throw new Error('The native build directory is redirected')
  }
  // 原生编译完整保留默认跟踪与分析功能；源码、依赖和中间文件使用本次独占的短目录。
  await cp(packageDir, source, { recursive: true, dereference: true, filter: path => !['build', 'node_modules'].includes(relative(packageDir, path).split(/[\\/]/)[0]) })
  const addon = await realpath(dirname(createRequire(join(packageDir, 'package.json')).resolve('node-addon-api/package.json')))
  const addonDestination = name === '@vscode/windows-process-tree' ? join(source, 'deps', 'node-addon-api') : join(source, 'node_modules', 'node-addon-api')
  await mkdir(dirname(addonDestination), { recursive: true })
  await cp(addon, addonDestination, { recursive: true, dereference: true, errorOnExist: true, force: false })
  if (name === 'node-pty') await copyFile(join(root, 'third_party', 'orca-v1.4.230', 'native', 'swa-provider-job.h'), join(source, 'src', 'win', 'swa_provider_job.h'))
  const childEnvironment = { ...environment, CL: '/utf-8', _CL_: '/analyze:log "' + analysisDirectory + '\\\\"' }
  if (childEnvironment._CL_.length > 1024) throw new Error('The native compiler environment exceeds the MSVC limit')
  console.log(JSON.stringify({ stage: 'native-module', name, source, analysisDirectory, CL: childEnvironment.CL, _CL_: childEnvironment._CL_ }))
  execFileSync(node, [gyp, 'rebuild', '--python=' + python, '--arch=x64', '--jobs=4', '--nodedir=' + nodeDir], { cwd: source, env: childEnvironment, stdio: 'inherit', windowsHide: true, timeout: 20 * 60_000 })
  const buildDir = join(source, 'build', 'Release')
  const binary = ['conpty.node', 'windows_process_tree.node', 'orca_windows_security.node', 'orca_windows_registry.node'][index]
  if (!(await lstat(join(buildDir, binary))).isFile()) throw new Error('The actual native binary is missing')
  if (name === 'node-pty' && !(await lstat(join(analysisDirectory, 'conpty.nativecodeanalysis.xml'))).size) throw new Error('The actual ConPTY analysis log is missing')
  const pending = [join(source, 'build')]
  let maximumPathLength = 0, trackedFiles = 0
  while (pending.length) {
    for (const entry of await readdir(pending.pop(), { withFileTypes: true })) {
      const path = join(entry.parentPath, entry.name)
      maximumPathLength = Math.max(maximumPathLength, path.length)
      if (path.length >= 260) throw new Error('A native intermediate path exceeds MAX_PATH: ' + path)
      if (entry.isDirectory()) pending.push(path)
      else if (entry.name.endsWith('.tlog')) trackedFiles++
    }
  }
  if (!trackedFiles) throw new Error('The actual native tracking logs are missing')
  const destination = join(packageDir, 'build', 'Release')
  await mkdir(destination, { recursive: true })
  await copyFile(join(buildDir, binary), join(destination, binary))
  if (name === 'node-pty') {
    for (const [file, fileSource] of slotSourceFiles({ platform: 'win32', arch: 'x64', buildDir: destination, nodePtyDir: packageDir })) {
      const target = join(destination, file)
      if (resolve(fileSource) === resolve(target)) continue
      await mkdir(dirname(target), { recursive: true })
      await copyFile(fileSource, target)
    }
  }
  console.log(JSON.stringify({ stage: 'native-module-built', name, binary: join(destination, binary), maximumPathLength, trackedFiles }))
}
