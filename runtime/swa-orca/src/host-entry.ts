import { mkdirSync, realpathSync, lstatSync } from 'node:fs'
import { dirname, isAbsolute, join, relative } from 'node:path'
import { setAppEnvironment, type AppPathName } from '../../../third_party/orca-v1.4.230/src/shared/app-environment'
import { setSecretStore } from '../../../third_party/orca-v1.4.230/src/shared/secret-store'

const absolute = (key: string) => {
  const input = process.env[key]
  if (!input || !isAbsolute(input)) throw new Error(`SWA requires an absolute ${key}`)
  return realpathSync.native(input)
}
const data = absolute('SWA_DATA_DIR')
const install = absolute('SWA_INSTALL_ROOT')
if (process.platform !== 'win32') throw new Error('This SWA package requires Windows')
if (absolute('ORCA_USER_DATA') !== data || absolute('ORCA_USER_DATA_PATH') !== data) throw new Error('SWA native profile root mismatch')
const privatePath = (name: string) => {
  const target = join(data, name)
  try { if (lstatSync(target).isSymbolicLink()) throw new Error('SWA private path redirected') }
  catch (error) { if (!(error instanceof Error) || !('code' in error) || error.code !== 'ENOENT') throw error }
  mkdirSync(target, { recursive: true })
  const actual = realpathSync.native(target)
  const rel = relative(data, actual)
  if (!rel || rel.startsWith('..') || isAbsolute(rel)) throw new Error('SWA private path escaped')
  return actual
}
const home = privatePath('home'), temp = privatePath('temp'), logs = privatePath('logs'), downloads = privatePath('downloads')
for (const key of ['HOME', 'USERPROFILE', 'APPDATA', 'LOCALAPPDATA']) if (absolute(key) !== home) throw new Error(`SWA private ${key} mismatch`)
for (const key of ['TEMP', 'TMP']) if (absolute(key) !== temp) throw new Error(`SWA private ${key} mismatch`)
const callbacks: (() => void)[] = []
const paths: Record<AppPathName, string> = { userData: data, home, appData: home, temp, downloads, logs, exe: absolute('SWA_NODE_EXECUTABLE') }
setAppEnvironment({
  getPath: name => paths[name], getAppPath: () => join(install, 'runtime'), getVersion: () => '1.4.230-swa',
  isPackaged: () => true, getCliLauncherPath: () => absolute('SWA_EXECUTABLE'),
  onWillQuit: callback => callbacks.push(callback), exit: code => process.exit(code), getAppMetrics: () => []
})

async function main() {
  const { createOrcadSecretStore } = await import('../../../third_party/orca-v1.4.230/src/main/orcad/orcad-secret-store')
  const secret = createOrcadSecretStore()
  setSecretStore(secret)
  if (!(await secret.prepare({ keyDir: privatePath('keys'), userDataPath: data }))) throw new Error('SWA private secret protection unavailable')
  const { startOrcadWithHost, installOrcadShutdownSignals, flushOrcadProfileStoreForShutdown } = await import('../../../third_party/orca-v1.4.230/src/main/orcad/orcad-lifecycle')
  const running = await startOrcadWithHost(data, async registerCleanup => {
    const { createOrcadProfileStateStartup } = await import('../../../third_party/orca-v1.4.230/src/main/orcad/orcad-profile-state-startup')
    const { store, authority } = await createOrcadProfileStateStartup(data)
    registerCleanup(() => flushOrcadProfileStoreForShutdown(store))
    privatePath('pi')
    store.updateSettings({ agentStatusHooksEnabled: false, workspaceDir: privatePath('workspaces'), browserClientHostedRemoteEnabled: false, browserSshWorkspaceRoutingEnabled: false })
    const { initDaemonPtyProvider } = await import('../../../third_party/orca-v1.4.230/src/main/daemon/daemon-provider-init')
    const { daemonOwnsFreshPersistentPtys, disconnectDaemon, readDaemonPidRecord } = await import('../../../third_party/orca-v1.4.230/src/main/daemon/daemon-provider-state')
    await initDaemonPtyProvider(undefined, { macosLoginSessionWatch: false, startupTimeoutMs: 30_000 })
    if (!daemonOwnsFreshPersistentPtys()) throw new Error('SWA requires its actual persistent daemon')
    registerCleanup(() => disconnectDaemon())
    const { getLocalPtyProvider } = await import('../../../third_party/orca-v1.4.230/src/main/ipc/pty/provider/registry')
    const { agentHookServer } = await import('../../../third_party/orca-v1.4.230/src/main/agent-hooks/server')
    await agentHookServer.start({ env: 'production', userDataPath: data, statusHooksEnabled: false })
    registerCleanup(() => agentHookServer.stop())
    const { agentHookRuntimeDeps } = await import('../../../third_party/orca-v1.4.230/src/main/agent-hooks/agent-hook-runtime-deps')
    const { AgentStatusObservedPaneIdentities, AgentStatusObservedPaneIdentityCapture } = await import('../../../third_party/orca-v1.4.230/src/main/runtime/agent-status-observed-pane-identity')
    const observed = new AgentStatusObservedPaneIdentities()
    const capture = new AgentStatusObservedPaneIdentityCapture(observed)
    registerCleanup(agentHookServer.subscribeEnrichedStatus(status => capture.observe(status)))
    const { OrcaRuntimeService } = await import('../../../third_party/orca-v1.4.230/src/main/runtime/orca-runtime')
    const { loadAgentSessionClaimSigner } = await import('../../../third_party/orca-v1.4.230/src/main/runtime/agent-session-claim-identity')
    const runtime = new OrcaRuntimeService(store, undefined, {
      agentSessionClaimSigner: loadAgentSessionClaimSigner(data, data),
      ...agentHookRuntimeDeps(agentHookServer, observed, () => store.getSettings()),
      getLocalProvider: () => getLocalPtyProvider(), canRecoverPersistentLocalPtys: () => daemonOwnsFreshPersistentPtys(),
      getDesktopWindowStatus: () => 'blocked'
    })
    const { registerHeadlessPtyRuntime } = await import('../../../third_party/orca-v1.4.230/src/main/ipc/pty/register-headless-runtime')
    await registerHeadlessPtyRuntime(runtime, undefined, () => store.getSettings(), undefined, store)
    await runtime.refreshRestoredOrchestrationAuthority()
    await runtime.reconcileLegacyWorkerTerminals()
    capture.attach(runtime)
    const { publishHeadlessRuntimeGraph } = await import('../../../third_party/orca-v1.4.230/src/main/runtime/headless-runtime-graph')
    publishHeadlessRuntimeGraph(runtime)
    registerCleanup(() => runtime.stopLegacyWorkerTerminalRecovery())
    registerCleanup(() => runtime.closeNotificationDismissals())
    const { disposeWatcherProcessAndWait } = await import('../../../third_party/orca-v1.4.230/src/main/ipc/parcel-watcher-process')
    registerCleanup(() => disposeWatcherProcessAndWait())
    const { OrcaRuntimeRpcServer } = await import('../../../third_party/orca-v1.4.230/src/main/runtime/runtime-rpc')
    const { ALL_RPC_METHODS } = await import('../../../third_party/orca-v1.4.230/src/main/runtime/rpc/methods')
    const { SWA_METHODS } = await import('./native-methods')
    const allowed = new Set(['status.get', 'orchestration.callerShow', 'orchestration.runCurrent', 'orchestration.runCreate', 'orchestration.runShow', 'orchestration.taskCreate', 'orchestration.taskList', 'orchestration.taskUpdate', 'orchestration.workerStart', 'orchestration.workerShow', 'orchestration.workerList', 'orchestration.workerRead', 'orchestration.workerStop', 'orchestration.workerRelease', 'orchestration.send', 'orchestration.requestShow'])
    const rpc = new OrcaRuntimeRpcServer({ runtime, userDataPath: data, enableWebSocket: false, methods: [...ALL_RPC_METHODS.filter(method => allowed.has(method.name)), ...SWA_METHODS] })
    registerCleanup(() => rpc.stop())
    await rpc.start()
    return { runtimeId: runtime.getRuntimeId(), daemon: readDaemonPidRecord(), authority, dataRoot: data, node: process.execPath }
  }, () => { for (const callback of callbacks.splice(0)) callback() }, { browserProvider: false })
  installOrcadShutdownSignals(() => running.stop())
  process.stdout.write(JSON.stringify({ ok: true, readiness: running }) + '\n')
}
main().catch(error => { process.stderr.write(JSON.stringify({ code: 'swa_runtime_start_failed', message: error instanceof Error ? error.message : 'SWA startup failed' }) + '\n'); process.exitCode = 1 })
