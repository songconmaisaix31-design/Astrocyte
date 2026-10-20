import { readFile, realpath, stat } from 'node:fs/promises'
import { createHash } from 'node:crypto'
import { delimiter, isAbsolute, join } from 'node:path'
import { parse as parseTOML } from 'smol-toml'
import { z } from 'zod'
import type { OrcaRuntimeService } from '../../../third_party/orca-v1.4.230/src/main/runtime/orca-runtime'
import { getStructuredAgentSessionHost } from '../../../third_party/orca-v1.4.230/src/main/native-chat/agent-session-wire/structured-agent-session-registry'
import { resolveStructuredAgentCommand } from '../../../third_party/orca-v1.4.230/src/main/native-chat/structured-agent-command-resolution'
import { readAgentCliVersion } from '../../../third_party/orca-v1.4.230/src/main/agent-cli-version-probe'
import { ACP_LAUNCH_SPECS } from '../../../third_party/orca-v1.4.230/src/main/acp/acp-launch-specs'
import { isTuiAgent } from '../../../third_party/orca-v1.4.230/src/shared/tui-agent-config'
import { normalizeCodexRuntimeSelection } from '../../../third_party/orca-v1.4.230/src/main/codex-accounts/runtime-selection'
import { normalizeClaudeRuntimeSelection } from '../../../third_party/orca-v1.4.230/src/main/claude-accounts/runtime-selection'
import { OrchestrationError } from '../../../third_party/orca-v1.4.230/src/main/runtime/orchestration/orchestration-error'
import { readAgentSettings, type AgentSettingsSnapshot } from './agent-settings'
import { agentCapabilitiesSchema, agentSelectionSchema, type AgentSelection, type CapabilityRequirements } from './contracts'

type Observation = { status: 'available' | 'unavailable' | 'unknown'; reason: string; source: string; checked_at?: string }
const observation = (status: Observation['status'], reason: string, source: string): Observation => ({ status, reason, source })
const unknown = (reason: string, source: string) => observation('unknown', reason, source)
const available = (reason: string, source: string) => observation('available', reason, source)
const object = z.record(z.string(), z.unknown())
const text = (value: unknown): string | undefined => typeof value === 'string' && value.trim() ? value : undefined

async function optionalConfiguration(file: string, format: 'json' | 'toml' = 'json') {
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
  return { value: object.parse(format === 'toml' ? parseTOML(contents) : JSON.parse(contents)), version: `${info.dev}:${info.ino}:${info.mtimeNs}:${info.size}` }
}

// 认证内容只用于判断来源类型；返回值只包含非秘密配置事实。
async function configuration(agent: string, snapshot: AgentSettingsSnapshot) {
  const home = snapshot.home
  if (agent === 'codex') {
    const accountID = normalizeCodexRuntimeSelection(snapshot.settings).host
    const account = accountID ? snapshot.settings.codexManagedAccounts?.find(item => item.id === accountID) : undefined
    if (accountID && (!account || account.managedHomeRuntime === 'wsl')) throw new Error('The explicit native Codex account selection has no local read-only source')
    const accountHome = account?.managedHomePath ?? snapshot.settings.agentDefaultEnv?.codex?.CODEX_HOME ?? (process.env.SWA_DISCOVERY_CODEX_HOME || join(home, '.codex'))
    if (!isAbsolute(accountHome)) throw new Error('The selected Codex account home must be absolute')
    const config = await optionalConfiguration(join(accountHome, 'config.toml'), 'toml')
    const auth = await optionalConfiguration(join(accountHome, 'auth.json'))
    const tokens = auth?.value.tokens
    const oauth = typeof tokens === 'object' && tokens !== null && text(object.parse(tokens).access_token)
    const key = text(auth?.value.OPENAI_API_KEY)
    return { model: text(config?.value.model), provider: text(config?.value.model_provider), account: oauth ? 'codex-oauth-file' : key ? 'codex-api-key-file' : undefined, version: `${config?.version ?? 'absent'};${auth?.version ?? 'absent'}` }
  }
  if (agent === 'claude') {
    const accountID = normalizeClaudeRuntimeSelection(snapshot.settings).host
    const account = accountID ? snapshot.settings.claudeManagedAccounts?.find(item => item.id === accountID) : undefined
    if (accountID && (!account || account.managedAuthRuntime === 'wsl')) throw new Error('The explicit native Claude account selection has no local read-only source')
    const configHome = snapshot.settings.agentDefaultEnv?.claude?.CLAUDE_CONFIG_DIR ?? (process.env.SWA_DISCOVERY_CLAUDE_HOME || join(home, '.claude'))
    if (!isAbsolute(configHome)) throw new Error('The selected Claude configuration home must be absolute')
    const config = await optionalConfiguration(join(configHome, 'settings.json'))
    const credentials = await optionalConfiguration(account?.managedAuthPath ?? join(configHome, '.credentials.json'))
    const env = config?.value.env ? object.parse(config.value.env) : {}
    const oauth = credentials?.value.claudeAiOauth
    const token = typeof oauth === 'object' && oauth !== null && text(object.parse(oauth).accessToken)
    const key = text(env.ANTHROPIC_AUTH_TOKEN) ?? text(env.ANTHROPIC_API_KEY)
    return { model: text(config?.value.model) ?? text(env.ANTHROPIC_MODEL), provider: token ? 'anthropic' : undefined, account: token ? 'claude-oauth-file' : key ? 'claude-settings-key' : undefined, version: `${config?.version ?? 'absent'};${credentials?.version ?? 'absent'}` }
  }
  if (agent === 'pi') {
    const accountHome = snapshot.settings.agentDefaultEnv?.pi?.PI_CODING_AGENT_DIR ?? (process.env.SWA_DISCOVERY_PI_HOME || join(home, '.pi', 'agent'))
    if (!isAbsolute(accountHome)) throw new Error('The selected Pi account home must be absolute')
    const config = await optionalConfiguration(join(accountHome, 'settings.json'))
    const auth = await optionalConfiguration(join(accountHome, 'auth.json'))
    const provider = text(config?.value.defaultProvider)
    const entry = provider && auth?.value[provider] ? object.parse(auth.value[provider]) : undefined
    const account = entry?.type === 'oauth' && text(entry.access) ? 'pi-oauth-file' : entry?.type === 'api_key' && text(entry.key) ? 'pi-api-key-file' : undefined
    return { model: text(config?.value.defaultModel), provider, account, version: `${config?.version ?? 'absent'};${auth?.version ?? 'absent'}` }
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

async function inspectAgents(runtime: OrcaRuntimeService, requirements?: CapabilityRequirements, onlyAgent?: string) {
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
        const result = await readAgentCliVersion({ program: executable, cwd: process.env.SWA_DATA_DIR!, env })
        version = result.version ?? undefined
        installed = version ? { ...available('The selected native executable returned its real CLI version', 'native-version-probe'), executable, version } : unknown('The selected native executable did not prove its CLI version', 'native-version-probe')
      } else installed = observation('unavailable', 'Native command resolution found no installed executable', 'native-command-resolution')
    }
    const config = await configuration(id, snapshot)
    const proven = !!(config.model && config.provider && config.account)
    const configured = proven ? { ...available('Native configuration supplies model, provider and an account source', 'native-configuration'), model: config.model, provider: config.provider, account_source: config.account } : unknown('Native configuration has not proved model, provider and an account source', 'native-configuration')
    const declared = available('The real adapter implements this structured operation; no live handshake is claimed', `adapter-declaration:${adapterName}`)
    const control = { start: declared, send: declared, read: unknown('No live session or complete transcript was observed', 'native-session'), stop: declared, resume: unknown('No native recovery handshake was observed', 'native-session'), tools: unknown('The selected Agent tool policy requires its own verified native contract', 'native-policy') }
    agents.push({ id, adapter: adapterName, execution_mode: 'structured' as const, registered, enabled, installed, configured, launchable: unknown('No actual Agent handshake was performed', 'native-session'), control, ...(requirements ? { requirements } : {}) })
    const selection = enabled.status === 'available' && executable && version && proven ? agentSelectionSchema.parse({ agent_id: id, adapter: adapterName, execution_mode: 'structured', executable, version, model: config.model, provider: config.provider, external_target: id, settings_source: snapshot.source, settings_revision: snapshot.revision, configuration_version: `${snapshot.version};${config.version}` }) : undefined
    details.set(id, { selection, configuration: config })
  }
  return { agents, snapshot, details }
}

export async function discoverAgents(runtime: OrcaRuntimeService, requirements: CapabilityRequirements) {
  const { agents, snapshot } = await inspectAgents(runtime, requirements)
  const sourceRevision = (await readFile(join(process.env.SWA_INSTALL_ROOT!, 'VERSION'), 'utf8')).trim()
  if (!/^[a-f0-9]{40}$/.test(sourceRevision)) throw new Error('The actual installed product source revision is invalid')
  const artifactSHA = createHash('sha256').update(await readFile(join(process.env.SWA_INSTALL_ROOT!, 'runtime', 'host.cjs'))).digest('hex')
  const result = agentCapabilitiesSchema.parse({ schema: 'swa.capabilities.v1', source_revision: sourceRevision, registry: { orca_version: '1.4.230', artifact_sha256: artifactSHA }, discovery: { host: 'local', platform: process.platform, observed_at: new Date().toISOString(), settings_source: snapshot.source }, agents })
  return result
}

export async function selectAgent(runtime: OrcaRuntimeService, agent: string, model: string) {
  const { details } = await inspectAgents(runtime, undefined, agent)
  const item = details.get(agent)
  if (!item?.selection) throw new OrchestrationError('agent_unconfigured', 'The selected native Agent lacks proven installation or configuration')
  return agentSelectionSchema.parse({ ...item.selection, model })
}
