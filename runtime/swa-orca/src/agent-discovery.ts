import { readFile, realpath, stat } from 'node:fs/promises'
import { realpathSync, statSync } from 'node:fs'
import { createHash } from 'node:crypto'
import { delimiter, dirname, extname, isAbsolute, join, normalize, resolve } from 'node:path'
import { isDeepStrictEqual } from 'node:util'
import { parse as parseTOML } from 'smol-toml'
import { z } from 'zod'
import type { OrcaRuntimeService } from '../../../third_party/orca-v1.4.230/src/main/runtime/orca-runtime'
import { getStructuredAgentSessionHost } from '../../../third_party/orca-v1.4.230/src/main/native-chat/agent-session-wire/structured-agent-session-registry'
import { resolveStructuredAgentCommand } from '../../../third_party/orca-v1.4.230/src/main/native-chat/structured-agent-command-resolution'
import { readAgentCliVersion } from '../../../third_party/orca-v1.4.230/src/main/agent-cli-version-probe'
import { ACP_LAUNCH_SPECS } from '../../../third_party/orca-v1.4.230/src/main/acp/acp-launch-specs'
import { isTuiAgent } from '../../../third_party/orca-v1.4.230/src/shared/tui-agent-config'
import { OrchestrationError } from '../../../third_party/orca-v1.4.230/src/main/runtime/orchestration/orchestration-error'
import { readAgentSettings, type AgentSettingsSnapshot } from './agent-settings'
import { parsePiModelOptionId } from '../../../third_party/orca-v1.4.230/src/main/pi/rpc-options'
import { agentCapabilitiesSchema, agentSelectionSchema, registeredAgentSchema, type AgentSelection, type CapabilityRequirements } from './contracts'
import { inspectNativePiPolicy } from './native-pi-policy'
import { inspectNativeOpenCodePolicy, readOpenCodeCredential } from './native-opencode-policy'
import { parse as parseJSONC, type ParseError } from 'jsonc-parser'
import { inspectNativeCodexAccountSource, inspectNativeCodexPolicy } from './native-codex-policy'
import { inspectNativeClaudeAccountSource, inspectNativeClaudePolicy } from './native-claude-policy'

type Observation = { status: 'available' | 'unavailable' | 'unknown'; reason: string; source: string; checked_at?: string }
const observation = (status: Observation['status'], reason: string, source: string): Observation => ({ status, reason, source })
const unknown = (reason: string, source: string) => observation('unknown', reason, source)
const available = (reason: string, source: string) => observation('available', reason, source)
const object = z.record(z.string(), z.unknown())
const text = (value: unknown): string | undefined => typeof value === 'string' && value.trim() ? value : undefined

async function optionalConfiguration(file: string, format: 'json' | 'toml' | 'jsonc' = 'json') {
  let contents: string
  let sourceVersion: string
  try {
    const before = await stat(file, { bigint: true })
    sourceVersion = `${before.dev}:${before.ino}:${before.mtimeNs}:${before.size}`
    contents = await readFile(file, 'utf8')
  } catch (error) {
    if (error instanceof Error && 'code' in error && error.code === 'ENOENT') return undefined
    throw error
  }
  const info = await stat(file, { bigint: true })
  if (sourceVersion !== `${info.dev}:${info.ino}:${info.mtimeNs}:${info.size}`) throw new Error('The selected native Agent configuration changed during reading')
  const errors: ParseError[] = []
  const value: unknown = format === 'toml' ? parseTOML(contents) : format === 'jsonc' ? parseJSONC(contents, errors, { allowTrailingComma: true }) : JSON.parse(contents)
  if (errors.length) throw new Error('The selected native configuration is not valid JSONC')
  return { value: object.parse(value), version: `${info.dev}:${info.ino}:${info.mtimeNs}:${info.size}` }
}

// 认证内容只用于判断来源类型；返回值只包含非秘密配置事实。
export function codexProtectedRoots(snapshot: AgentSettingsSnapshot, projectRoot: string) {
  return [projectRoot, process.env.SWA_AGENT_PROFILE, snapshot.settings.agentDefaultEnv?.codex?.CODEX_HOME, snapshot.settings.agentDefaultEnv?.claude?.CLAUDE_CONFIG_DIR,
    ...snapshot.settings.codexManagedAccounts?.map(account => account.managedHomePath) ?? [],
    ...snapshot.settings.claudeManagedAccounts?.map(account => dirname(account.managedAuthPath)) ?? []].filter((path): path is string => !!path && isAbsolute(path))
}

async function configuration(agent: string, snapshot: AgentSettingsSnapshot, requestedModel?: string, projectRoot?: string) {
  const home = snapshot.home
  if (agent === 'codex') {
    const sourceHome = process.env.SWA_CODEX_API_KEY_HOME
    if (requestedModel !== undefined && requestedModel !== 'gpt-5.5') throw new Error('native_codex_model_unreviewed')
    if (!sourceHome || !projectRoot) return { model: undefined, provider: undefined, account: undefined, sourceHome, version: 'explicit-api-key-source-unverified' }
    const account = await inspectNativeCodexAccountSource({ sourceHome, protectedRoots: codexProtectedRoots(snapshot, projectRoot), env: process.env })
    return { model: 'gpt-5.5', provider: 'openai', account: 'codex-native-api-key', sourceHome: account.sourceHome, version: account.authSource.version }
  }
  if (agent === 'claude') {
    const sourceHome = process.env.SWA_CLAUDE_API_KEY_HOME
    if (requestedModel !== undefined && requestedModel !== 'claude-opus-5') throw new Error('native_claude_model_unreviewed')
    if (!sourceHome || !projectRoot) return { model: undefined, provider: undefined, account: undefined, sourceHome, version: 'explicit-api-key-source-unverified' }
    const account = await inspectNativeClaudeAccountSource({ sourceHome, protectedRoots: codexProtectedRoots(snapshot, projectRoot), env: process.env })
    return { model: 'claude-opus-5', provider: 'anthropic', account: 'claude-native-api-key', sourceHome: account.sourceHome, version: account.authSource.version }
  }
  if (agent === 'pi') {
    const accountHome = snapshot.settings.agentDefaultEnv?.pi?.PI_CODING_AGENT_DIR ?? (process.env.SWA_DISCOVERY_PI_HOME || join(home, '.pi', 'agent'))
    if (!isAbsolute(accountHome)) throw new Error('The selected Pi account home must be absolute')
    const config = await optionalConfiguration(join(accountHome, 'settings.json'))
    const auth = await optionalConfiguration(join(accountHome, 'auth.json'))
    const explicit = requestedModel === undefined ? undefined : parsePiModelOptionId(requestedModel)
    const provider = explicit?.provider ?? text(config?.value.defaultProvider)
    const entry = provider && auth?.value[provider] ? object.parse(auth.value[provider]) : undefined
    const oauth = entry?.type === 'oauth' && text(entry.access) && text(entry.refresh) && typeof entry.expires === 'number' && Number.isFinite(entry.expires)
    const account = oauth ? 'pi-oauth-file' : entry?.type === 'api_key' && text(entry.key) ? 'pi-api-key-file' : undefined
    const defaultModel = text(config?.value.defaultModel)
    return { model: requestedModel ?? (provider && defaultModel ? `${provider}/${defaultModel}` : undefined), provider, account, sourceHome: accountHome, version: `${config?.version ?? 'absent'};${auth?.version ?? 'absent'}` }
  }
  if (agent === 'opencode') {
    const sourceHome = process.env.SWA_DISCOVERY_OPENCODE_HOME || home
    if (!isAbsolute(sourceHome)) throw new Error('The selected OpenCode source home must be absolute')
    const config = await optionalConfiguration(join(sourceHome, '.config/opencode/opencode.jsonc'), 'jsonc')
    const auth = await optionalConfiguration(join(sourceHome, '.local/share/opencode/auth.json'))
    const entry = auth?.value.deepseek === undefined ? undefined : object.parse(auth.value.deepseek)
    const model = requestedModel ?? text(config?.value.model)
    const file = join(sourceHome, '.local/share/opencode/opencode.db')
    let databaseVersion = 'absent', databaseKeyMatches = false
    try {
      const info = await stat(file, { bigint: true })
      databaseVersion = `${info.dev}:${info.ino}:${info.mtimeNs}:${info.size}`
      const credential = readOpenCodeCredential(file)
      databaseKeyMatches = credential.key === entry?.key
      databaseVersion += `;${credential.id};${credential.updatedAt}`
    } catch (error) { if (!(error instanceof Error) || !('code' in error) || error.code !== 'ENOENT') throw error }
    return { model, provider: model === 'deepseek/deepseek-v4-pro' ? 'deepseek' : undefined, account: entry?.type === 'api' && text(entry.key) && databaseKeyMatches ? 'opencode-native-key-database' : undefined, sourceHome, version: `${config?.version ?? 'absent'};${auth?.version ?? 'absent'};${databaseVersion}` }
  }
  return { model: undefined, provider: undefined, account: undefined, version: 'native-configuration-unobserved' }
}

function discoveryEnvironment(): Record<string, string> {
  const env: Record<string, string> = {}
  for (const [key, value] of Object.entries(process.env)) {
    if (value !== undefined && /^(?:PATH|PATHEXT|COMSPEC|SYSTEMROOT|WINDIR|TEMP|TMP|HOME|USERPROFILE|APPDATA|LOCALAPPDATA|LANG|LC_ALL|LC_CTYPE)$/i.test(key)) env[key] = value
  }
  // 原 home 只参与命令解析；version 子进程继续使用私有 HOME。
  return env
}

async function inspectAgents(runtime: OrcaRuntimeService, requirements?: CapabilityRequirements, onlyAgent?: string, requestedModel?: string, projectId?: string, projectRoot?: string) {
  const snapshot = await readAgentSettings()
  await runtime.ensureStructuredAgentSessionHost({ prewarmModelCatalog: false })
  const host = getStructuredAgentSessionHost()
  if (!host) throw new Error('The actual product structured Agent registry is unavailable')
  const disabled = new Set<string>(snapshot.launch.disabledTuiAgents)
  const agents = []
  const details = new Map<string, { selection: AgentSelection | undefined; configuration: Awaited<ReturnType<typeof configuration>> }>()
  for (const { definition, adapter } of host.deps.agents.registrations()) {
    const id = definition.agent
    if (onlyAgent !== undefined && id !== onlyAgent) continue
    const adapterName = adapter.constructor.name
    const registered = available('The real native adapter passed StructuredAgentRegistry validation', `StructuredAgentRegistry:${adapterName}`)
    const enabled = observation(disabled.has(id) ? 'unavailable' : 'available', disabled.has(id) ? 'Disabled in the selected native settings' : 'Enabled in the selected native settings', 'native-settings')
    let installed: Observation & { executable?: string; version?: string } = unknown('No absolute native command has been resolved', 'native-command-resolution')
    let executable: string | undefined, version: string | undefined
    if (isTuiAgent(id)) {
      const spec = ACP_LAUNCH_SPECS.find(item => item.agent === id)
      const env = discoveryEnvironment()
      const sourceEnv: Record<string, string> = { ...env, HOME: snapshot.home, USERPROFILE: snapshot.home }
      if (process.env.SWA_DISCOVERY_APPDATA) sourceEnv.APPDATA = process.env.SWA_DISCOVERY_APPDATA
      if (process.env.SWA_DISCOVERY_LOCALAPPDATA) sourceEnv.LOCALAPPDATA = process.env.SWA_DISCOVERY_LOCALAPPDATA
      const overlay = snapshot.settings.agentDefaultEnv?.[id]
      const pathEnv = [overlay?.PATH ?? overlay?.Path ?? env.PATH ?? env.Path, ...(spec?.installDirectories({ env: sourceEnv, homePath: snapshot.home }) ?? [])].filter(Boolean).join(delimiter)
      const command = resolveStructuredAgentCommand(id, spec?.ignoreAgentCommandSetting ? {} : { agentCmdOverrides: snapshot.launch.agentCmdOverrides }, { pathEnv, homePath: snapshot.home }, { command: spec?.command ?? id })
      if (isAbsolute(command)) {
        executable = await realpath(command)
        if (id === 'opencode' && extname(executable).toLowerCase() === '.cmd') executable = await realpath(join(dirname(executable), 'node_modules/@opencode/cli/bin/opencode.exe'))
        const result = await readAgentCliVersion({ program: executable, cwd: process.env.SWA_DATA_DIR!, env })
        version = result.version ?? undefined
        installed = version ? { ...available('The selected native executable returned its real CLI version', 'native-version-probe'), executable, version } : unknown('The selected native executable did not prove its CLI version', 'native-version-probe')
      } else installed = observation('unavailable', 'Native command resolution found no installed executable', 'native-command-resolution')
    }
    const config = await configuration(id, snapshot, requestedModel, projectRoot)
    const proven = !!(config.model && config.provider && config.account)
    const configured = { ...(proven ? available('Native configuration and the explicit selection supply model, provider and an account source', 'native-configuration') : unknown('The observed native defaults do not supply every model, provider and account field', 'native-configuration')), model: config.model, provider: config.provider, account_source: config.account }
    const declared = available('The real adapter implements this structured operation; no live handshake is claimed', `adapter-declaration:${adapterName}`)
    const control = { start: declared, send: declared, read: unknown('No live session or complete transcript was observed', 'native-session'), stop: declared, resume: unknown('No native recovery handshake was observed', 'native-session'), tools: unknown('The selected Agent tool policy requires its own verified native contract', 'native-policy') }
    agents.push({ id, adapter: adapterName, execution_mode: 'structured' as const, registered, enabled, installed, configured, launchable: unknown('No actual Agent handshake was performed', 'native-session'), control, ...(requirements ? { requirements } : {}) })
    let selection: AgentSelection | undefined
    if (projectId && enabled.status === 'available' && executable && version && proven) {
      if (id !== 'pi' && id !== 'opencode' && id !== 'codex' && id !== 'claude' || !('sourceHome' in config) || !config.sourceHome || !config.model) throw new OrchestrationError('unsupported_capability', 'The selected native Agent account and disabled-tool policy have no reviewed contract')
      if (snapshot.settings.agentDefaultArgs?.[id]?.trim()) throw new OrchestrationError('unsupported_capability', 'The selected native Agent extra arguments have not been approved')
      const native = id === 'pi'
        ? await inspectNativePiPolicy({ executable, version, sourceHome: config.sourceHome, projectId, ...parsePiModelOptionId(config.model), env: discoveryEnvironment() })
        : id === 'opencode' ? await inspectNativeOpenCodePolicy({ executable, version, sourceHome: config.sourceHome, projectId, privateRoot: join(await realpath(process.env.SWA_DATA_DIR!), 'opencode', projectId) })
          : id === 'codex' ? await inspectNativeCodexPolicy({ executable, version, sourceHome: config.sourceHome, projectId, privateRoot: join(await realpath(process.env.SWA_DATA_DIR!), 'codex', projectId), catalogPath: join(await realpath(process.env.SWA_INSTALL_ROOT!), 'runtime/codex-policy/models.json'), env: process.env, protectedRoots: codexProtectedRoots(snapshot, projectRoot!) })
            : await inspectNativeClaudePolicy({ executable, version, sourceHome: config.sourceHome, projectId, privateRoot: join(await realpath(process.env.SWA_DATA_DIR!), 'claude', projectId), env: process.env, protectedRoots: codexProtectedRoots(snapshot, projectRoot!) })
      selection = agentSelectionSchema.parse({ agent_id: id, adapter: adapterName, execution_mode: 'structured', executable, version, model: config.model, provider: native.policy.account.provider, external_target: id, settings_source: snapshot.source, settings_revision: snapshot.revision, configuration_version: `${snapshot.version};${config.version};${JSON.stringify(native.sources)}`, native_launch: native.policy })
    }
    details.set(id, { selection, configuration: config })
  }
  return { agents, snapshot, details }
}

async function installedRegistrySource() {
  const sourceRevision = (await readFile(join(process.env.SWA_INSTALL_ROOT!, 'VERSION'), 'utf8')).trim()
  if (!/^[a-f0-9]{40}$/.test(sourceRevision)) throw new Error('The actual installed product source revision is invalid')
  const artifactSHA = createHash('sha256').update(await readFile(join(process.env.SWA_INSTALL_ROOT!, 'runtime', 'host.cjs'))).digest('hex')
  return { source_revision: sourceRevision, registry: { orca_version: '1.4.230' as const, artifact_sha256: artifactSHA } }
}

export async function registeredAgent(runtime: OrcaRuntimeService, agent: string) {
  const host = getStructuredAgentSessionHost()
  const registration = host?.deps.agents.registration(agent)
  if (!registration) throw new OrchestrationError('unsupported_capability', 'The requested Agent has no actual structured registration or bound terminal launcher')
  return registeredAgentSchema.parse({
    schema: 'swa.registered-agent.v1', id: registration.definition.agent, adapter: registration.adapter.constructor.name,
    execution_mode: 'structured', registered: true, source: { kind: 'structured-registry' },
    ...await installedRegistrySource(), runtime: {
      runtime_id: runtime.getRuntimeId(), authority_domain: runtime.getOrchestrationDb().getOrCreateLocalMutationCallerFingerprint(),
      data_root: await realpath(process.env.SWA_DATA_DIR!), install_root: await realpath(process.env.SWA_INSTALL_ROOT!)
    }
  })
}

export function readDiscoveryProjectRoot(root: string) {
  try {
    if (!isAbsolute(root)) throw new Error()
    const path = realpathSync.native(root)
    const spelling = (value: string) => process.platform === 'win32' ? normalize(value).toLowerCase() : normalize(value)
    if (spelling(path) !== spelling(resolve(root))) throw new Error()
    const info = statSync(path, { bigint: true })
    if (!info.isDirectory()) throw new Error()
    return { path, identity: `${info.dev}:${info.ino}:${info.birthtimeNs}` }
  } catch { throw new OrchestrationError('context_stale', 'The authorized project directory is unavailable or redirected') }
}

export function assertDiscoverySnapshot(root: ReturnType<typeof readDiscoveryProjectRoot>, before: AgentSettingsSnapshot, after: AgentSettingsSnapshot) {
  if (!isDeepStrictEqual(root, readDiscoveryProjectRoot(root.path)) || !isDeepStrictEqual(before, after)) throw new OrchestrationError('context_stale', 'The project directory or native Agent settings changed during discovery')
}

export async function discoverAgents(runtime: OrcaRuntimeService, requirements: CapabilityRequirements, projectRoot: string) {
  const root = readDiscoveryProjectRoot(projectRoot)
  // 能力查询只检查静态配置；不传 projectId，避免为全部 Agent 建立启动选择。
  const { agents, snapshot } = await inspectAgents(runtime, requirements, undefined, undefined, undefined, root.path)
  const source = await installedRegistrySource()
  const current = await readAgentSettings()
  assertDiscoverySnapshot(root, snapshot, current)
  const result = agentCapabilitiesSchema.parse({ schema: 'swa.capabilities.v1', ...source, discovery: { host: 'local', platform: process.platform, observed_at: new Date().toISOString(), settings_source: snapshot.source }, agents })
  return result
}

export async function selectAgent(runtime: OrcaRuntimeService, projectId: string, agent: string, model: string, projectRoot?: string) {
  const { details } = await inspectAgents(runtime, undefined, agent, model, projectId, projectRoot)
  const item = details.get(agent)
  if (!item?.selection) throw new OrchestrationError('agent_unconfigured', 'The selected native Agent lacks proven installation or configuration')
  return agentSelectionSchema.parse({ ...item.selection, model })
}
