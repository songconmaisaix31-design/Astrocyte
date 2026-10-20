import assert from 'node:assert/strict'
import { access, mkdir, readFile, realpath, writeFile } from 'node:fs/promises'
import { createRequire } from 'node:module'
import { dirname, isAbsolute, join, relative } from 'node:path'
import { pathToFileURL } from 'node:url'
import { parseArgs } from 'node:util'
import { createInterface } from 'node:readline'

const { values } = parseArgs({ options: { candidate: { type: 'string' }, output: { type: 'string' }, project: { type: 'string' }, 'auth-file': { type: 'string' } }, allowPositionals: false })
for (const key of ['candidate', 'output', 'project', 'auth-file']) if (!values[key] || !isAbsolute(values[key])) throw new Error(`Explicit absolute ${key} is required`)
const candidate = await realpath(values.candidate)
const node = join(candidate, 'node', 'node.exe')
assert.equal((await realpath(process.execPath)).toLowerCase(), node.toLowerCase())
const project = await realpath(values.project)
const authFile = await realpath(values['auth-file'])
const sourceBytes = await readFile(authFile)
const requireRuntime = createRequire(join(candidate, 'runtime', 'package.json'))
const { JsonlRpcAgentConnection } = requireRuntime('./agent-connection.cjs')
const { SwaProviderProcessTree } = requireRuntime('./swa-provider-process-tree.cjs')
const { spawnManagedProviderProcess } = requireRuntime('./managed-provider-process.cjs')
const processAPI = requireRuntime('@vscode/windows-process-tree')
const { readonlyCodexCredentials, PROVIDER } = await import(pathToFileURL(join(candidate, 'runtime', 'readonly-codex-credentials.mjs')).href)
const root = join(dirname(values.output), 'actual-sdk-sessions')
await mkdir(root)
const marker = join(root, 'forbidden-refresh-callback.txt')
const credentials = readonlyCodexCredentials(authFile)
const realCredential = await credentials.store.read(PROVIDER)
assert.ok(realCredential.expires - Date.now() > 900_000)
await assert.rejects(credentials.store.modify(PROVIDER, async () => {
  await writeFile(marker, 'Actual callback was incorrectly reached', { flag: 'wx' })
  return realCredential
}), /selected_codex_oauth_read_only/)
await assert.rejects(credentials.store.delete(PROVIDER), /selected_codex_oauth_read_only/)
await assert.rejects(credentials.store.read('unapproved-provider'), /unapproved_credential_provider/)
await assert.rejects(access(marker), error => error.code === 'ENOENT')

async function launch(name, args = []) {
  const worker = join(root, name)
  for (const directory of ['home', 'pi', 'sessions', 'temp']) await mkdir(join(worker, directory), { recursive: true })
  const home = join(worker, 'home')
  const env = {
    HOME: home, USERPROFILE: home, APPDATA: home, LOCALAPPDATA: home,
    TEMP: join(worker, 'temp'), TMP: join(worker, 'temp'),
    SWA_WORKER_ROOT: worker, SWA_PROJECT_ROOT: project, SWA_CODEX_AUTH_FILE: authFile,
    PI_CODING_AGENT_DIR: join(worker, 'pi'), PI_CODING_AGENT_SESSION_DIR: join(worker, 'sessions'),
    PI_PACKAGE_DIR: await realpath(join(candidate, 'runtime', 'node_modules', '@earendil-works', 'pi-coding-agent')),
    NODE_OPTIONS: '--max-old-space-size=1024'
  }
  const retained = new Set(Object.keys(env).map(key => key.toUpperCase()))
  const envToDelete = Object.keys(process.env).filter(key => /^(ORCA_|PI_|CODEX_|CLAUDE_|ANTHROPIC_|OPENAI_|SWA_|NODE_|ASTROCYTE_)/i.test(key) && !retained.has(key.toUpperCase()))
  return { command: node, args: [join(candidate, 'runtime', 'pi-entry.mjs'), ...args], cwd: project, env, envToDelete, swaJobIPC: true }
}

const prepareTree = new SwaProviderProcessTree()
const preparation = spawnManagedProviderProcess(await launch('prepare', ['--prepare']), { site: 'actual-restricted-sdk-preparation', acceptClose: value => value.root === 'exited' && value.tree === 'exited' })
preparation.child.stderr.pipe(process.stderr, { end: false })
const prepareRecords = []
const prepareRead = (async () => {
  for await (const line of createInterface({ input: preparation.child.stdout, crlfDelay: Infinity })) prepareRecords.push(JSON.parse(line))
})()
prepareTree.bind(preparation.child, () => ({ observed: preparation.rootExitObserved, processless: preparation.processless }))
await prepareTree.ready()
const prepareJob = prepareTree.readObservation()
await Promise.race([preparation.exitPromise, new Promise((_, reject) => setTimeout(() => reject(new Error('Actual SDK preparation did not exit')), 90_000).unref())])
await prepareRead
assert.equal(preparation.child.exitCode, 0)
assert.equal(preparation.child.signalCode, null)
assert.equal(prepareRecords.length, 1)
const restrictions = prepareRecords[0]
assert.equal(restrictions.provider, 'openai-codex')
assert.equal(restrictions.model, 'gpt-6.1-sol')
assert.equal(restrictions.authentication, 'oauth')
assert.deepEqual(restrictions.activeTools, [])
assert.deepEqual(restrictions.registeredTools, [])
assert.equal(restrictions.settings.cacheWarming, 'off')
assert.equal(restrictions.settings.retry, false)
assert.equal(restrictions.settings.compaction, false)
assert.equal(restrictions.settings.transport, 'sse')
assert.ok(restrictions.credentialOperations.length > 0)
assert.ok(restrictions.credentialOperations.every(row => row.provider === PROVIDER && row.operation === 'read'))
assert.deepEqual(await preparation.close(prepareTree), { root: 'exited', tree: 'exited' })

const rpcRecords = []
const connection = new JsonlRpcAgentConnection(await launch('abort'), {
  allowedCommands: new Set(['get_state', 'abort']),
  onRecord: record => rpcRecords.push(record),
  onDiagnostic: message => process.stderr.write(message + '\n')
})
const state = await connection.request('get_state', {}, { timeoutMs: 90_000 })
assert.equal(state.model.provider, PROVIDER)
assert.equal(state.model.id, 'gpt-6.1-sol')
assert.equal(state.thinkingLevel, 'xhigh')
assert.equal(state.isStreaming, false)
assert.equal(state.autoCompactionEnabled, false)
assert.equal(state.messageCount, 0)
const expectedSessions = join(root, 'abort', 'sessions')
const rel = relative(expectedSessions, state.sessionFile)
assert.ok(rel && !rel.startsWith('..') && !isAbsolute(rel))
const actualRoot = { pid: connection.pid, creationTimeMs: processAPI.getProcessCreationTime(connection.pid) }
assert.ok(actualRoot.pid > 0 && actualRoot.creationTimeMs > 0)
await assert.rejects(connection.request('prompt', { message: 'This forbidden request must never reach the SDK' }), /SWA refuses Pi RPC command/)
await connection.request('abort', {}, { timeoutMs: 30_000 })
assert.deepEqual(await connection.close(), { root: 'exited', tree: 'exited' })
assert.equal(connection.rootVerdict, 'exited')
assert.ok(rpcRecords.every(row => row.type !== 'agent_start' && row.type !== 'message_start' && row.type !== 'message_update'))
assert.deepEqual(await readFile(authFile), sourceBytes)
await assert.rejects(access(marker), error => error.code === 'ENOENT')
await writeFile(values.output, JSON.stringify({
  candidate, source_revision: (await readFile(join(candidate, 'VERSION'), 'utf8')).trim(),
  real_sdk: '1.0.1', model_calls: 0, preparation: { restrictions, admittedJob: prepareJob, close: preparation.lastCloseResult },
  actualRPC: { state, root: actualRoot, abortAcknowledged: true, close: connection.lastCloseResult, records: rpcRecords },
  credentialOperations: credentials.observations, refreshCallbackReached: false, originalAuthBytesUnchanged: true,
  remainingCases: []
}, null, 2) + '\n', { flag: 'wx' })
console.log(JSON.stringify({ output: values.output, actual_sdk_abort_then_close: 'passed', model_calls: 0, multi_worker_acceptance: 'NOT_RUN' }))
