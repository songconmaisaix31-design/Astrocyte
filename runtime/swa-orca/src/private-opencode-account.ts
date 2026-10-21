import { lstat, mkdir, readFile, readdir, realpath } from 'node:fs/promises'
import { basename, dirname, isAbsolute, join, relative, resolve, toNamespacedPath } from 'node:path'
import { isDeepStrictEqual } from 'node:util'
import { createFileWithDacl, currentUserSid, getDaclAces, setDaclSddl } from '@orca/windows-security'
import { z } from 'zod'
import { assertNativeOpenCodeCurrent, openCodeBuiltinPlugins, type NativeOpenCodeSnapshot } from './native-opencode-policy'
import { nativeConfigurationSource } from './native-pi-policy'

const systemSid = 'S-1-5-18'
const samePath = (a: string, b: string) => a.toLowerCase() === b.toLowerCase()
const deny = { action: '*', resource: '*', effect: 'deny' }
const privateDirectories = ['home', 'data', 'data/opencode', 'cache', 'config', 'config/opencode', 'state', 'programdata', 'appdata', 'localappdata', 'temp']

async function actualDirectory(path: string) {
  if (!isAbsolute(path) || !samePath(await realpath(path), resolve(path))) throw new Error('private_opencode_directory_redirected')
  for (let ancestor = resolve(path); ; ancestor = dirname(ancestor)) {
    const info = await lstat(ancestor)
    if (!info.isDirectory() || info.isSymbolicLink()) throw new Error('private_opencode_directory_redirected')
    if (dirname(ancestor) === ancestor) break
  }
}

function security(path: string, directory: boolean) {
  const user = currentUserSid(), acl = getDaclAces(toNamespacedPath(path)), trustees = new Set([user, systemSid])
  if (!acl?.protected || acl.ownerSid !== user || acl.aces.length !== trustees.size) throw new Error('private_opencode_security_changed')
  for (const ace of acl.aces) {
    if (ace.sid === null || !trustees.delete(ace.sid) || ace.type !== 0 || ace.flags !== (directory ? 3 : 0) || ace.mask !== 0x1f01ff) throw new Error('private_opencode_security_changed')
  }
  if (trustees.size) throw new Error('private_opencode_security_changed')
}

async function directory(path: string) {
  let created = false
  try { await mkdir(path); created = true }
  catch (error) { if (!(error instanceof Error) || !('code' in error) || error.code !== 'EEXIST') throw error }
  await actualDirectory(path)
  if (created) {
    if ((await readdir(path)).length) throw new Error('private_opencode_directory_not_empty')
    setDaclSddl(toNamespacedPath(path), `D:P${[...new Set([currentUserSid(), systemSid])].map(sid => `(A;OICI;FA;;;${sid})`).join('')}`)
  }
  security(path, true)
}

function writePrivateJSON(file: string, value: unknown) {
  const user = currentUserSid()
  createFileWithDacl(toNamespacedPath(file), `O:${user}D:P${[...new Set([user, systemSid])].map(sid => `(A;;FA;;;${sid})`).join('')}`, Buffer.from(JSON.stringify(value) + '\n'))
  security(file, false)
}

export async function createPrivateOpenCodeRoot(projectId: string) {
  z.uuid().parse(projectId)
  const data = await realpath(z.string().min(1).parse(process.env.SWA_DATA_DIR))
  await actualDirectory(data)
  await directory(join(data, 'opencode'))
  const root = join(data, 'opencode', projectId)
  await directory(root)
  return root
}

export function privateOpenCodeConfig(root: string, plugin: string, correlation: string) {
  return {
    model: 'deepseek/deepseek-v4-pro', default_agent: 'build', update: 'disable', share: 'disabled',
    warming: false, compaction: { auto: false }, snapshots: false, permissions: [deny],
    agents: { build: { permissions: [deny] }, ...Object.fromEntries(['general', 'explore', 'title', 'summary', 'compaction'].map(id => [id, { disabled: true }])) },
    plugins: ['-*', ...openCodeBuiltinPlugins.filter(id => !['opencode.provider.opencode', 'opencode.config.policy'].includes(id)), { package: plugin, options: { root, correlation, source: join(plugin, 'server.js') } }]
  }
}

export async function materializePrivateOpenCodeAccount(snapshot: NativeOpenCodeSnapshot, sessionId: string, plugin: string, correlation: string) {
  if (!/^[a-zA-Z0-9_-]+$/.test(sessionId) || !/^[a-f0-9]{64}$/.test(correlation)) throw new Error('private_opencode_worker_identity_invalid')
  const data = await realpath(z.string().parse(process.env.SWA_DATA_DIR))
  if (!samePath(dirname(snapshot.policy.account.private_root), join(data, 'opencode'))) throw new Error('private_opencode_root_changed')
  const expected = await createPrivateOpenCodeRoot(basename(snapshot.policy.account.private_root))
  if (!samePath(expected, snapshot.policy.account.private_root)) throw new Error('private_opencode_root_changed')
  assertNativeOpenCodeCurrent(snapshot)
  const root = join(expected, sessionId)
  for (const protectedRoot of [snapshot.policy.account.source_home, process.env.SWA_INSTALL_ROOT]) {
    if (!protectedRoot) continue
    const part = relative(await realpath(protectedRoot), root)
    if (!part || !part.startsWith('..') && !isAbsolute(part)) throw new Error('private_opencode_root_overlap')
  }
  await directory(root)
  if ((await readdir(root)).length) throw new Error('private_opencode_existing_account_unreviewed')
  for (const name of privateDirectories) await directory(join(root, name))
  const actualPlugin = await realpath(plugin)
  await actualDirectory(actualPlugin)
  for (const name of ['package.json', 'server.js']) {
    const source = await nativeConfigurationSource(join(actualPlugin, name))
    if (!samePath(source.path, join(actualPlugin, name)) || (await lstat(source.path)).isSymbolicLink()) throw new Error('private_opencode_plugin_redirected')
    if (!snapshot.sources.some(item => isDeepStrictEqual(item, source))) snapshot.sources.push(source)
  }
  writePrivateJSON(join(root, 'data/opencode/auth.json'), { deepseek: { type: 'api', key: snapshot.key } })
  writePrivateJSON(join(root, 'config/opencode/opencode.json'), privateOpenCodeConfig(root, actualPlugin, correlation))
  await verifyPrivateOpenCodeAccount(snapshot, root, actualPlugin, correlation)
  return root
}

export async function verifyPrivateOpenCodeAccount(snapshot: NativeOpenCodeSnapshot, root: string, plugin: string, correlation: string) {
  const within = relative(snapshot.policy.account.private_root, root)
  if (!within || within.startsWith('..') || isAbsolute(within) || within.includes('/') || within.includes('\\')) throw new Error('private_opencode_worker_root_invalid')
  await actualDirectory(root)
  security(root, true)
  for (const name of privateDirectories) { await actualDirectory(join(root, name)); security(join(root, name), true) }
  for (const [name, expected] of [['data/opencode/auth.json', { deepseek: { type: 'api', key: snapshot.key } }], ['config/opencode/opencode.json', privateOpenCodeConfig(root, plugin, correlation)]] as const) {
    const file = join(root, name), info = await lstat(file)
    if (!info.isFile() || info.isSymbolicLink() || !samePath(await realpath(file), file)) throw new Error('private_opencode_file_redirected')
    security(file, false)
    if (!isDeepStrictEqual(JSON.parse(await readFile(file, 'utf8')), expected)) throw new Error('private_opencode_configuration_changed')
  }
  assertNativeOpenCodeCurrent(snapshot)
}

export function privateOpenCodeEnvironment(root: string, inherited: Readonly<Record<string, string>>) {
  const allowed = new Set(['SYSTEMROOT', 'WINDIR', 'SYSTEMDRIVE', 'COMSPEC', 'PATHEXT', 'PATH', 'NUMBER_OF_PROCESSORS', 'PROCESSOR_ARCHITECTURE', 'PROCESSOR_IDENTIFIER', 'OS', 'LANG', 'LC_ALL', 'LC_CTYPE', 'TERM', 'COLORTERM', 'HTTP_PROXY', 'HTTPS_PROXY', 'ALL_PROXY', 'NO_PROXY'])
  const env: Record<string, string> = {}
  for (const [key, value] of Object.entries(inherited)) {
    const name = key.toUpperCase()
    if (!allowed.has(name)) continue
    if (env[name] !== undefined && env[name] !== value) throw new Error('native_environment_case_collision')
    env[name] = value
  }
  return Object.assign(env, {
    HOME: join(root, 'home'), USERPROFILE: join(root, 'home'), XDG_DATA_HOME: join(root, 'data'), XDG_CACHE_HOME: join(root, 'cache'),
    XDG_CONFIG_HOME: join(root, 'config'), XDG_STATE_HOME: join(root, 'state'), OPENCODE_CONFIG_DIR: join(root, 'config/opencode'),
    OPENCODE_DB: join(root, 'data/opencode/opencode.db'), ProgramData: join(root, 'programdata'), APPDATA: join(root, 'appdata'),
    LOCALAPPDATA: join(root, 'localappdata'), TEMP: join(root, 'temp'), TMP: join(root, 'temp'),
    OPENCODE_CONFIG_PROJECT_DISABLE: '1', OPENCODE_DISABLE_CLAUDE_CODE: '1', OPENCODE_DISABLE_MODELS_FETCH: '1',
    OPENCODE_DISABLE_FILEWATCHER: '1', OPENCODE_DISABLE_FFF: '1', OPENCODE_PRINT_LOGS: '1'
  })
}
