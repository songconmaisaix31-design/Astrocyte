import { isDeepStrictEqual } from 'node:util'
import { realpath } from 'node:fs/promises'
import { z } from 'zod'
import type { OrcaRuntimeService } from '../../../third_party/orca-v1.4.230/src/main/runtime/orca-runtime'
import { installStructuredAgentLaunchAdmission } from '../../../third_party/orca-v1.4.230/src/main/runtime/structured-agent-launch-composition'
import { structuredWorkerIdentities } from '../../../third_party/orca-v1.4.230/src/main/runtime/structured-worker-identity'
import { getPendingWorkerStartRecovery } from '../../../third_party/orca-v1.4.230/src/main/runtime/rpc/orchestration-mutation-receipt'
import { agentSelectionSchema, nativeObject } from './contracts'
import { selectAgent } from './agent-discovery'

export function installAgentLaunchAdmission(runtime: OrcaRuntimeService) {
  async function originalSelection(sessionId: string, agent: string) {
    const identity = structuredWorkerIdentities.getBySessionId(sessionId)
    if (!identity?.dispatchId || identity.agent !== agent) throw new Error('The native Worker launch has no exact Dispatch association')
    const db = runtime.getOrchestrationDb()
    const parents = z.array(z.object({ receipt: z.string() })).parse(db.db.prepare(`SELECT receipt FROM mutation_receipts
      WHERE method='swa.controllerDeliver' AND state IN ('pending','completed') AND json_valid(receipt)=1
      AND json_extract(receipt, '$.action')='worker.start'`).all())
    const matches = []
    for (const row of parents) {
      const parent = nativeObject(JSON.parse(row.receipt))
      const request = z.string().min(1).parse(parent.child_request)
      const children = z.array(z.object({ state: z.string(), receipt: z.string().nullable() })).parse(db.db.prepare(`SELECT state, receipt FROM mutation_receipts
        WHERE method='orchestration.workerStart' AND request_id=?`).all(request))
      if (children.length !== 1 || children[0]!.receipt === null) continue
      const child = children[0]!
      const dispatchId = child.state === 'pending' ? getPendingWorkerStartRecovery('orchestration.workerStart', child.receipt)?.dispatchId : nativeObject(JSON.parse(child.receipt!)).dispatchId
      if (dispatchId === identity.dispatchId) matches.push(agentSelectionSchema.parse(parent.selection))
    }
    if (matches.length !== 1 || matches[0]!.agent_id !== agent) throw new Error('The original native Worker selection is unavailable or ambiguous')
    const original = matches[0]!
    const record = runtime.openedAgentSessionRecordStore()?.getRecord(sessionId)
    if (!record || record.provider !== agent || record.options?.model !== original.model) throw new Error('The native session Agent and model changed before launch')
    const current = await selectAgent(runtime, original.agent_id, original.model)
    if (!isDeepStrictEqual(original, current)) throw new Error('The selected native Agent configuration changed before launch')
    return original
  }
  return installStructuredAgentLaunchAdmission({
    commandSettings: async (sessionId, agent) => {
      const original = await originalSelection(sessionId, agent)
      return { agentCmdOverrides: { [agent]: `"${original.executable}"` } }
    },
    validate: async (sessionId, agent, command) => {
      const original = await originalSelection(sessionId, agent)
      const actual = await realpath(command)
      const same = process.platform === 'win32' ? actual.toLowerCase() === original.executable.toLowerCase() : actual === original.executable
      if (!same) throw new Error('The native adapter resolved a different executable from the approved Agent selection')
    }
  })
}
