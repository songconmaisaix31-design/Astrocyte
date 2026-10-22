import { isPersistedAgentSessionRecord } from './source/third_party/orca-v1.4.230/src/shared/agent-session-record'
import { decodePersistedAgentSessionRecord } from './source/third_party/orca-v1.4.230/src/shared/agent-session-record-stored-form'
import { agentSessionProviderHandleChainHead } from './source/third_party/orca-v1.4.230/src/shared/agent-session-provider-handle'
import { isSwaLinuxProcessAbsent, readSwaLinuxProcessIncarnation } from './source/third_party/orca-v1.4.230/src/main/daemon/swa-linux-process-identity'

// 观察入口只调用安装版本的成熟 decoder 和明确 PID 的只读内核观察。
async function main() {
  const [mode, pidText] = process.argv.slice(2)
  if (mode === 'incarnation') {
    if (!/^[1-9][0-9]*$/.test(pidText ?? '')) throw new Error('observer_invalid_pid')
    const pid = Number(pidText)
    if (!Number.isSafeInteger(pid)) throw new Error('observer_invalid_pid')
    const value = isSwaLinuxProcessAbsent(pid) ? { absent: true, pid } : { absent: false, incarnation: readSwaLinuxProcessIncarnation(pid) }
    process.stdout.write(JSON.stringify(value) + '\n')
    return
  }
  if (mode !== 'decode-record' || pidText !== undefined) throw new Error('observer_invalid_mode')
  let size = 0
  const chunks: Buffer[] = []
  for await (const chunk of process.stdin) {
    if (!(chunk instanceof Uint8Array)) throw new Error('observer_input_not_bytes')
    size += chunk.length
    if (size > 1024 * 1024) throw new Error('observer_input_exceeds_bound')
    chunks.push(Buffer.from(chunk))
  }
  const parsed: unknown = JSON.parse(new TextDecoder('utf-8', { fatal: true }).decode(Buffer.concat(chunks)))
  if (!isPersistedAgentSessionRecord(parsed)) throw new Error('observer_record_unreadable')
  const record = decodePersistedAgentSessionRecord(parsed).record
  const head = agentSessionProviderHandleChainHead(record.providerHandleChain)
  const owner = record.lease.ownerProcess
  const death = record.lease.deathEvidence
  const projection = {
    schema_version: record.schemaVersion,
    session_id: record.sessionId,
    provider: record.provider,
    model: record.options?.model ?? null,
    head,
    lease: {
      sessionId: record.lease.sessionId,
      runtimeKind: record.lease.runtimeKind,
      claimStatus: record.lease.claimStatus,
      runtimeFence: record.lease.runtimeFence,
      handoffStage: record.lease.handoffStage,
      provenHandleLinkId: record.lease.provenHandleLinkId,
      unreconciled: record.lease.unreconciled,
      ownerProcess: owner ? { pid: owner.pid, processStartTimeMs: owner.processStartTimeMs, runtime: owner.runtime ?? null } : null,
      deathEvidence: death ? { kind: death.kind, observedAt: death.observedAt, ownerFence: death.ownerFence ?? null } : null,
    },
  }
  process.stdout.write(JSON.stringify(projection) + '\n')
}

void main().catch(() => {
  // 原始 record、账户目录和原生错误不进入观察日志。
  process.stderr.write('source_observer_unavailable\n')
  process.exitCode = 1
})
