import { mkdir, realpath, lstat, readFile } from 'node:fs/promises'
import { isAbsolute, join, relative, sep } from 'node:path'
import { z } from 'zod'
import type { StructuredAgentLaunchPartResolver } from '../runtime/structured-agent-launch-composition'
import type { PiRpcResolvedLaunch } from './rpc-launch-resolution'
import { agentSessionProviderHandleChainHead } from '../../shared/agent-session-provider-handle'
import { PI_RPC_SESSION_ENV_TO_DELETE } from './rpc-launch'
import { isWindowsProcessStartTimeAvailable } from '../windows/windows-process-table'
import { structuredWorkerIdentities } from '../runtime/structured-worker-identity'
import { AGENT_SESSION_FOUNDING_FENCE } from '../runtime/agent-session-record-founding'

const requiredPath = async (key: string): Promise<string> => {
  const value = process.env[key]
  if (!value || !isAbsolute(value)) throw new Error(`SWA requires an absolute ${key}`)
  return realpath(value)
}

async function directory(root: string, name: string): Promise<string> {
  const target = join(root, name)
  try {
    if ((await lstat(target)).isSymbolicLink()) throw new Error('SWA private directory is redirected')
  } catch (error) {
    if (!(error instanceof Error) || !('code' in error) || error.code !== 'ENOENT') throw error
  }
  await mkdir(target, { recursive: true })
  const path = await realpath(target)
  const rel = relative(root, path)
  if (!rel || rel === '..' || rel.startsWith(`..${sep}`) || isAbsolute(rel)) throw new Error('SWA private path escaped')
  return path
}

export const swaPiLaunchPart: StructuredAgentLaunchPartResolver<PiRpcResolvedLaunch> = async (basis, identity) => {
  if (process.platform === 'win32' && !isWindowsProcessStartTimeAvailable()) throw new Error('SWA requires Windows process creation-time proof')
  const worker = structuredWorkerIdentities.getBySessionId(identity.sessionId)
  if (worker?.agent !== 'pi' || basis.record.lease.runtimeFence !== AGENT_SESSION_FOUNDING_FENCE + 1 || basis.record.lease.minimumNextFence !== undefined || basis.record.providerHandleChain.length) throw new Error('SWA refuses to restart or replace a prior Worker process; preserve its original Dispatch and request')
  if (!/^[a-zA-Z0-9_-]+$/.test(identity.sessionId)) throw new Error('Invalid native session ID')
  const data = await requiredPath('SWA_DATA_DIR')
  const workers = await directory(data, 'workers')
  const root = await directory(workers, identity.sessionId)
  const home = await directory(root, 'home')
  const agentDir = await directory(root, 'pi')
  const sessions = await directory(root, 'sessions')
  const temporary = await directory(root, 'temp')
  const node = await requiredPath('SWA_NODE_EXECUTABLE')
  const sdkEntry = await requiredPath('SWA_PI_ENTRY')
  const packageDir = await requiredPath('SWA_PI_PACKAGE_DIR')
  z.object({ name: z.literal('@earendil-works/pi-coding-agent'), version: z.literal('1.0.1') }).parse(JSON.parse(await readFile(join(packageDir, 'package.json'), 'utf8')))
  const authFile = await requiredPath('SWA_CODEX_AUTH_FILE')
  const cwd = await realpath(await basis.launchDirectory())
  const previous = agentSessionProviderHandleChainHead(basis.record.providerHandleChain)
  const env: Record<string, string> = {
    HOME: home, USERPROFILE: home, APPDATA: home, LOCALAPPDATA: home,
    TEMP: temporary, TMP: temporary,
    SWA_WORKER_ROOT: root, SWA_PROJECT_ROOT: cwd, SWA_CODEX_AUTH_FILE: authFile,
    PI_CODING_AGENT_DIR: agentDir, PI_CODING_AGENT_SESSION_DIR: sessions,
    PI_PACKAGE_DIR: packageDir, NODE_OPTIONS: '--max-old-space-size=1024'
  }
  const privateKeys = new Set(Object.keys(env).map(key => key.toUpperCase()))
  const remove = Object.keys(process.env).filter(key =>
    /^(ORCA_|CODEX_|CLAUDE_|ANTHROPIC_|PI_|SWA_|NODE_|ASTROCYTE_)/i.test(key) || key.toUpperCase() === 'OPENAI_API_KEY')
    .filter(key => !privateKeys.has(key.toUpperCase()))
  return {
    launch: { command: node, sdkEntry, fullAccess: false, previous },
    env, inheritedEnvToDelete: [...PI_RPC_SESSION_ENV_TO_DELETE.filter(key => !privateKeys.has(key.toUpperCase())), ...remove]
  }
}
