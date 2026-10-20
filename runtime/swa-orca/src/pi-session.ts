import { realpathSync, mkdirSync, lstatSync, readFileSync } from 'node:fs'
import { isAbsolute, relative, resolve, dirname, join } from 'node:path'
import { z } from 'zod'
import { readonlyCodexCredentials, MODEL, PROVIDER, REQUEST_TIMEOUT_MS } from './readonly-codex-credentials'

const required = (name: string): string => {
  const value = process.env[name]
  if (!value || !isAbsolute(value)) throw new Error(`missing_absolute_${name}`)
  return value
}

function privateDirectory(root: string, input: string): string {
  const path = resolve(input)
  const rel = relative(root, path)
  if (!rel || rel === '..' || rel.startsWith(`..${process.platform === 'win32' ? '\\' : '/'}`) || isAbsolute(rel)) {
    throw new Error('pi_directory_outside_worker_root')
  }
  for (let item = path; relative(root, item) !== ''; item = dirname(item)) {
    if (dirname(item) === item) throw new Error('pi_directory_outside_worker_root')
    try {
      if (lstatSync(item).isSymbolicLink()) throw new Error('pi_directory_redirected')
    } catch (error) {
      if (!(error instanceof Error) || !('code' in error) || error.code !== 'ENOENT') throw error
    }
  }
  mkdirSync(path, { recursive: true })
  const actual = realpathSync.native(path)
  const actualRelative = relative(root, actual)
  if (!actualRelative || actualRelative === '..' || actualRelative.startsWith('..') || isAbsolute(actualRelative)) {
    throw new Error('pi_directory_redirected')
  }
  return actual
}

export async function runRestrictedSession() {
  const root = realpathSync.native(required('SWA_WORKER_ROOT'))
  const privateHome = privateDirectory(root, required('HOME'))
  if (realpathSync.native(required('USERPROFILE')) !== privateHome) throw new Error('pi_home_mismatch')
  const agentDir = privateDirectory(root, required('PI_CODING_AGENT_DIR'))
  const sessionDir = privateDirectory(root, required('PI_CODING_AGENT_SESSION_DIR'))
  const packageDir = realpathSync.native(required('PI_PACKAGE_DIR'))
  z.object({ name: z.literal('@earendil-works/pi-coding-agent'), version: z.literal('1.0.1') }).parse(
    JSON.parse(readFileSync(join(packageDir, 'package.json'), 'utf8'))
  )
  const cwd = realpathSync.native(required('SWA_PROJECT_ROOT'))
  const auth = readonlyCodexCredentials(required('SWA_CODEX_AUTH_FILE'))
  const sdk = await import('@earendil-works/pi-coding-agent')
  const settings = sdk.SettingsManager.inMemory({
    defaultProvider: PROVIDER,
    defaultModel: MODEL,
    defaultThinkingLevel: 'xhigh',
    defaultProjectTrust: 'never',
    defaultTools: [],
    packages: [], extensions: [], skills: [], prompts: [], themes: [],
    enableSkillCommands: false,
    enableInstallTelemetry: false,
    enableAnalytics: false,
    images: { blockImages: true },
    cacheWarming: 'off',
    compaction: { enabled: false },
    retry: { enabled: false, maxRetries: 0, provider: { maxRetries: 0, timeoutMs: REQUEST_TIMEOUT_MS } },
    transport: 'sse'
  }, { projectTrusted: false })
  const runtime = await sdk.ModelRuntime.create({
    credentials: auth.store,
    modelsPath: null,
    allowModelNetwork: false,
    refreshOnCreate: false
  })
  const loader = new sdk.DefaultResourceLoader({
    cwd, agentDir, settingsManager: settings,
    noExtensions: true, noSkills: true, noPromptTemplates: true, noContextFiles: true, noThemes: true,
    additionalExtensionPaths: [], additionalSkillPaths: [], additionalPromptTemplatePaths: [], additionalThemePaths: [],
    extensionFactories: [],
    systemPrompt: 'Work only on the fixed public task and context supplied in the prompt. All native tools are disabled. Return your analysis and task result as text; the host reads the result and controls completion.',
    appendSystemPrompt: []
  })
  await loader.reload()
  const extensions = loader.getExtensions()
  if (extensions.errors.length || extensions.extensions.length || extensions.runtime.pendingProviderRegistrations.length || extensions.runtime.pendingNativeProviderRegistrations.length || extensions.runtime.pendingVirtualModelRegistrations.length) {
    throw new Error('pi_unapproved_extensions_or_providers')
  }
  const refreshed = await runtime.refresh({ providers: [PROVIDER], allowNetwork: false })
  if (refreshed.aborted || refreshed.errors.size || runtime.getError()) throw new Error('pi_selected_provider_refresh_failed')
  const model = runtime.getModel(PROVIDER, MODEL)
  if (!model || !runtime.isUsingOAuth(PROVIDER) || !runtime.hasConfiguredAuth(PROVIDER)) throw new Error('pi_selected_model_or_oauth_unavailable')
  const resolvedAuth = await runtime.getAuth(model)
  if (!resolvedAuth || resolvedAuth.source !== 'OAuth') throw new Error('pi_selected_oauth_source_mismatch')
  const services = { cwd, agentDir, modelRuntime: runtime, settingsManager: settings, resourceLoader: loader, diagnostics: [] }
  const assertRestrictions = (session: Awaited<ReturnType<typeof sdk.createAgentSession>>['session']) => {
    if (session.model?.provider !== PROVIDER || session.model.id !== MODEL || session.getActiveToolNames().length || session.getAllTools().length) {
      throw new Error('pi_model_or_tools_changed')
    }
    if (loader.getSkills().skills.length || loader.getPrompts().prompts.length || loader.getThemes().themes.length || loader.getAgentsFiles().agentsFiles.length || loader.getAppendSystemPrompt().length) {
      throw new Error('pi_unapproved_context_resources')
    }
    if (settings.getCacheWarmingMode() !== 'off' || settings.getCompactionEnabled() || settings.getRetryEnabled() || settings.getProviderRetrySettings().maxRetries !== 0 || settings.getTransport() !== 'sse') {
      throw new Error('pi_implicit_replay_enabled')
    }
    if (auth.observations.some(item => item.provider !== PROVIDER || item.operation !== 'read')) throw new Error('pi_unapproved_credential_operation')
  }
  const sessionManager = sdk.SessionManager.create(cwd, sessionDir)
  const host = await sdk.createAgentSessionRuntime(async options => {
    if (options.cwd !== cwd || options.agentDir !== agentDir) throw new Error('pi_session_scope_changed')
    const result = await sdk.createAgentSession({
      ...services, model, thinkingLevel: 'xhigh',
      sessionManager: options.sessionManager,
      sessionStartEvent: options.sessionStartEvent,
      tools: [], noTools: 'all', customTools: []
    })
    if (result.modelFallbackMessage) throw new Error('pi_model_fallback_refused')
    assertRestrictions(result.session)
    result.session.subscribe(event => {
      if (event.type === 'agent_start') assertRestrictions(result.session)
    })
    return { ...result, services, diagnostics: [] }
  }, { cwd, agentDir, sessionManager })
  assertRestrictions(host.session)
  if (process.argv.includes('--prepare')) {
    process.stdout.write(JSON.stringify({
      provider: PROVIDER, model: MODEL, authentication: 'oauth',
      home: privateHome, agentDir, sessionDir, packageDir,
      activeTools: host.session.getActiveToolNames(), registeredTools: host.session.getAllTools(),
      credentialOperations: auth.observations,
      settings: { cacheWarming: settings.getCacheWarmingMode(), retry: settings.getRetryEnabled(), compaction: settings.getCompactionEnabled(), transport: settings.getTransport() },
      cacheWarmingStatus: host.session.cacheWarmingStatus
    }) + '\n')
    await host.dispose()
    return
  }
  if (process.argv.length !== 2) throw new Error('pi_sdk_entry_accepts_no_cli_resource_options')
  await sdk.runRpcMode(host)
}
