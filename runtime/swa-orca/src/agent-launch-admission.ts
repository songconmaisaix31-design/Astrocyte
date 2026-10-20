import { isDeepStrictEqual } from 'node:util'
import { realpath } from 'node:fs/promises'
import { z } from 'zod'
import { isAbsolute, join } from 'node:path'
import type { OrcaRuntimeService } from '../../../third_party/orca-v1.4.230/src/main/runtime/orca-runtime'
import { installStructuredAgentLaunchAdmission } from '../../../third_party/orca-v1.4.230/src/main/runtime/structured-agent-launch-composition'
import { structuredWorkerIdentities } from '../../../third_party/orca-v1.4.230/src/main/runtime/structured-worker-identity'
import { getPendingWorkerStartRecovery } from '../../../third_party/orca-v1.4.230/src/main/runtime/rpc/orchestration-mutation-receipt'
import { agentSelectionSchema, invocationSchema, nativeObject } from './contracts'
import { selectAgent } from './agent-discovery'
import { assertNativePiSnapshotCurrent, inspectNativePiPolicy, sameNativeLaunchPolicy } from './native-pi-policy'
import { createNativeWorkerDirectory, materializePrivatePiAccount, privatePiEnvironment, verifyPrivatePiAccount } from './private-pi-account'
import { createDeliveryAdmission } from './delivery-admission'
import type { StructuredAgentInputAdmission } from '../../../third_party/orca-v1.4.230/src/main/runtime/structured-agent-launch-composition'
import { installWorkerInputApproval } from './worker-input'

const mutationRowSchema = z.object({ caller_fingerprint: z.string(), request_id: z.string(), method: z.string(), payload_hash: z.string(), state: z.enum(['pending', 'completed']), receipt: z.string().nullable() })
const inputBindingSchema = z.strictObject({ parent_request: z.string().min(1), child_request: z.string().min(1), dispatch_id: z.string().min(1), session_id: z.string().min(1), operation_id: z.string().min(1), runtime_fence: z.number().int().nonnegative(), consumer_generation: z.number().int().nonnegative(), body: z.unknown() })

export function installAgentLaunchAdmission(runtime: OrcaRuntimeService, owner: { ownerCapability: Buffer; ownerAlive(): boolean }) {
  const admitCurrentApproval = createDeliveryAdmission(owner.ownerCapability, owner.ownerAlive)
  const releaseInputApproval = installWorkerInputApproval(admitCurrentApproval)
  async function originalSelection(sessionId: string, agent: string, requireRecord = true) {
    if (!owner.ownerAlive()) throw new Error('The owning product service is unavailable')
    const identity = structuredWorkerIdentities.getBySessionId(sessionId)
    if (!identity?.dispatchId || identity.agent !== agent) throw new Error('The native Worker launch has no exact Dispatch association')
    const db = runtime.getOrchestrationDb()
    const parents = z.array(mutationRowSchema).parse(db.db.prepare(`SELECT * FROM mutation_receipts
      WHERE caller_fingerprint=? AND method='swa.controllerDeliver' AND state IN ('pending','completed') AND json_valid(receipt)=1
      AND json_extract(receipt, '$.action')='worker.start'`).all(db.getOrCreateLocalMutationCallerFingerprint()))
    const matches = []
    for (const row of parents) {
      if (row.receipt === null) continue
      const parent = nativeObject(JSON.parse(row.receipt))
      const request = z.string().min(1).parse(parent.child_request)
      const children = z.array(mutationRowSchema).parse(db.db.prepare(`SELECT * FROM mutation_receipts
        WHERE method='orchestration.workerStart' AND request_id=?`).all(request))
      if (children.length !== 1 || children[0]!.receipt === null) continue
      const child = children[0]!
      const dispatchId = child.state === 'pending' ? getPendingWorkerStartRecovery('orchestration.workerStart', child.receipt)?.dispatchId : nativeObject(JSON.parse(child.receipt!)).dispatchId
      if (dispatchId === identity.dispatchId) {
        const invocation = invocationSchema.parse(parent.invocation)
        const selection = agentSelectionSchema.parse(parent.selection)
        if (!isDeepStrictEqual(invocation.selection, selection) || invocation.project.id !== parent.project_id || invocation.scope.run_id !== parent.run_id || invocation.scope.root_task_id !== parent.root_task_id || invocation.scope.controller !== parent.controller || invocation.command.request_id + '/native' !== request) throw new Error('The original native Worker invocation changed')
        const dispatch = db.getDispatchContextById(identity.dispatchId)
        const task = dispatch && db.getTask(dispatch.task_id)
        if (!dispatch || dispatch.run_id !== invocation.scope.run_id || !task || task.parent_id !== invocation.scope.root_task_id) throw new Error('The native Worker Dispatch left its original Run and root task')
        matches.push({ selection, invocation, parentRow: row, childRow: child })
      }
    }
    if (matches.length !== 1 || matches[0]!.selection.agent_id !== agent) throw new Error('The original native Worker selection is unavailable or ambiguous')
    const { selection: original, invocation, parentRow, childRow } = matches[0]!
    const record = runtime.openedAgentSessionRecordStore()?.getRecord(sessionId)
    if (requireRecord || record) {
      if (!record || record.provider !== agent || record.options?.model !== original.model || !('variable' in record.accountHome) || record.accountHome.variable !== 'PI_CODING_AGENT_DIR' || record.accountHome.path !== original.native_launch.account.private_home) throw new Error('The native session Agent, account or model changed before launch')
    }
    const current = await selectAgent(runtime, invocation.project.id, original.agent_id, original.model)
    if (!isDeepStrictEqual(original, current)) throw new Error('The selected native Agent configuration changed before launch')
    return { original, invocation, parentRow, childRow }
  }
  async function policySnapshot(sessionId: string, agent: string, requireRecord = true) {
    const selected = await originalSelection(sessionId, agent, requireRecord)
    const { original } = selected
    const account = original.native_launch.account
    if (agent !== 'pi' || original.version !== '1.0.1') throw new Error('The selected native Agent has no reviewed disabled-tool policy')
    const snapshot = await inspectNativePiPolicy({ executable: original.executable, version: original.version, sourceHome: account.source_home, projectId: selected.invocation.project.id, provider: account.provider, modelId: account.model_target.model_id, env: process.env })
    if (!sameNativeLaunchPolicy(original.native_launch, snapshot.policy)) throw new Error('The original native account policy changed')
    return { ...selected, snapshot }
  }
  function saveBinding(row: z.infer<typeof mutationRowSchema>, binding: z.infer<typeof inputBindingSchema>) {
    const db = runtime.getOrchestrationDb()
    const current = db.getMutationReceipt(row.caller_fingerprint, row.request_id)
    if (!current || current.state !== 'pending' || current.payload_hash !== row.payload_hash || current.method !== row.method || !current.receipt) throw new Error('The original native input reservation is unavailable')
    const receipt = nativeObject(JSON.parse(current.receipt))
    if (receipt.swa_input_binding !== undefined && !isDeepStrictEqual(receipt.swa_input_binding, binding)) throw new Error('The original native input operation cannot be replaced')
    db.checkpointPendingMutationReceipt({ callerFingerprint: row.caller_fingerprint, requestId: row.request_id, method: row.method, payloadHash: row.payload_hash, receipt: JSON.stringify({ ...receipt, swa_input_binding: binding }) })
  }
  async function beforeInput(input: StructuredAgentInputAdmission) {
    const { original } = await policySnapshot(input.sessionId, input.agent)
    const db = runtime.getOrchestrationDb()
    const rows = z.array(mutationRowSchema).parse(db.db.prepare(`SELECT * FROM mutation_receipts WHERE caller_fingerprint=? AND method='swa.controllerDeliver' AND json_valid(receipt)=1 AND json_extract(receipt, '$.swa_input_binding.operation_id')=?`).all(db.getOrCreateLocalMutationCallerFingerprint(), input.clientMessageId))
    if (rows.length !== 1 || !rows[0]!.receipt) throw new Error('The original native input has no unique parent request')
    const receipt = nativeObject(JSON.parse(rows[0]!.receipt))
    const invocation = invocationSchema.parse(receipt.invocation)
    const binding = inputBindingSchema.parse(receipt.swa_input_binding)
    const childRows = z.array(mutationRowSchema).parse(db.db.prepare('SELECT * FROM mutation_receipts WHERE request_id=?').all(binding.child_request))
    if (childRows.length !== 1 || !childRows[0]!.receipt || !isDeepStrictEqual(nativeObject(JSON.parse(childRows[0]!.receipt)).swa_input_binding, binding)) throw new Error('The original native child input checkpoint is unavailable')
    const identity = structuredWorkerIdentities.getBySessionId(input.sessionId)
    const record = runtime.openedAgentSessionRecordStore()?.getRecord(input.sessionId)
    const dispatch = db.getDispatchContextById(binding.dispatch_id)
    const worker = db.getWorkerDispatch(binding.dispatch_id)
    const run = db.getRun(invocation.scope.run_id)
    if (!identity || identity.dispatchId !== binding.dispatch_id || binding.session_id !== input.sessionId || binding.operation_id !== input.clientMessageId || binding.runtime_fence !== input.fence || !isDeepStrictEqual(binding.body, input.body) || !isDeepStrictEqual(invocation.selection, original) || !record || record.lease.runtimeFence !== input.fence || record.lease.unreconciled || !record.lease.ownerProcess || record.lease.claimStatus !== 'live' || !dispatch || dispatch.run_id !== invocation.scope.run_id || !worker || !['starting', 'ready'].includes(worker.state) || worker.runtime_epoch !== runtime.getRuntimeId() || !run || run.consumer_generation !== binding.consumer_generation || run.coordinator_handle !== invocation.scope.controller) throw new Error('The native input Worker, consumer or session identity changed')
    await admitCurrentApproval(invocation)
    const fresh = await policySnapshot(input.sessionId, input.agent)
    await verifyPrivatePiAccount(fresh.snapshot)
    assertNativePiSnapshotCurrent(fresh.snapshot)
    const current = runtime.openedAgentSessionRecordStore()?.getRecord(input.sessionId)
    if (!owner.ownerAlive() || !isDeepStrictEqual(structuredWorkerIdentities.getBySessionId(input.sessionId), identity) || current?.lease.runtimeFence !== input.fence || current.lease.claimStatus !== 'live' || !isDeepStrictEqual(current.lease.ownerProcess, record.lease.ownerProcess) || db.getRun(run.id)?.consumer_generation !== binding.consumer_generation || !['starting', 'ready'].includes(db.getWorkerDispatch(binding.dispatch_id)?.state ?? '')) throw new Error('The native input identity changed during current approval')
  }
  const releaseLaunchAdmission = installStructuredAgentLaunchAdmission({
    commandSettings: async (sessionId, agent) => {
      const { original } = await originalSelection(sessionId, agent)
      return { agentCmdOverrides: { [agent]: `"${original.executable}"` } }
    },
    validate: async (sessionId, agent, command) => {
      const { original } = await originalSelection(sessionId, agent)
      const actual = await realpath(command)
      const same = process.platform === 'win32' ? actual.toLowerCase() === original.executable.toLowerCase() : actual === original.executable
      if (!same) throw new Error('The native adapter resolved a different executable from the approved Agent selection')
    },
    accountHome: async (sessionId, agent) => {
      const { invocation } = await policySnapshot(sessionId, agent, false)
      await admitCurrentApproval(invocation)
      const { snapshot } = await policySnapshot(sessionId, agent, false)
      await materializePrivatePiAccount(snapshot)
      assertNativePiSnapshotCurrent(snapshot)
      return { variable: 'PI_CODING_AGENT_DIR', path: snapshot.policy.account.private_home }
    },
    environment: async (sessionId, agent, inherited) => {
      const { invocation } = await policySnapshot(sessionId, agent)
      await admitCurrentApproval(invocation)
      const { snapshot } = await policySnapshot(sessionId, agent)
      await verifyPrivatePiAccount(snapshot)
      assertNativePiSnapshotCurrent(snapshot)
      return privatePiEnvironment(snapshot.policy.account.private_home, inherited)
    },
    disabledPiTools: async (sessionId, agent) => {
      const { invocation } = await policySnapshot(sessionId, agent)
      await admitCurrentApproval(invocation)
      const { snapshot } = await policySnapshot(sessionId, agent)
      await verifyPrivatePiAccount(snapshot)
      assertNativePiSnapshotCurrent(snapshot)
      const account = snapshot.policy.account
      return { nativeContract: 'pi-cli-1.0.1', provider: account.provider, modelId: account.model_target.model_id, api: account.model_target.api, baseUrl: account.model_target.base_url }
    },
    providerJobLauncher: async (sessionId, agent) => {
      const { invocation } = await policySnapshot(sessionId, agent)
      await admitCurrentApproval(invocation)
      const { snapshot } = await policySnapshot(sessionId, agent)
      await verifyPrivatePiAccount(snapshot)
      const node = z.string().min(1).parse(process.env.SWA_NODE_EXECUTABLE)
      const install = z.string().min(1).parse(process.env.SWA_INSTALL_ROOT)
      if (!isAbsolute(node) || !isAbsolute(install) || await realpath(node) !== join(await realpath(install), 'node', 'node.exe')) throw new Error('The installed native Job Node source is unavailable')
      const entry = await realpath(join(install, 'runtime', 'native-cli-entry.cjs'))
      const eventDirectory = await createNativeWorkerDirectory(sessionId)
      assertNativePiSnapshotCurrent(snapshot)
      return { node, entry, eventDirectory }
    },
    reserveInput: async input => {
      const { invocation, parentRow, childRow } = await policySnapshot(input.sessionId, input.agent)
      await admitCurrentApproval(invocation)
      const { snapshot } = await policySnapshot(input.sessionId, input.agent)
      await verifyPrivatePiAccount(snapshot)
      assertNativePiSnapshotCurrent(snapshot)
      const identity = structuredWorkerIdentities.getBySessionId(input.sessionId)
      const db = runtime.getOrchestrationDb()
      const run = db.getRun(invocation.scope.run_id)
      if (!identity?.dispatchId || !run || run.coordinator_handle !== invocation.scope.controller) throw new Error('The original native input Run association is unavailable')
      const binding = inputBindingSchema.parse({ parent_request: parentRow.request_id, child_request: childRow.request_id, dispatch_id: identity.dispatchId, session_id: input.sessionId, operation_id: input.clientMessageId, runtime_fence: input.fence, consumer_generation: run.consumer_generation, body: input.body })
      db.db.exec('BEGIN IMMEDIATE')
      try { saveBinding(parentRow, binding); saveBinding(childRow, binding); db.db.exec('COMMIT') }
      catch (error) { db.db.exec('ROLLBACK'); throw error }
    },
    beforeInput
  })
  return () => { releaseLaunchAdmission(); releaseInputApproval() }
}
