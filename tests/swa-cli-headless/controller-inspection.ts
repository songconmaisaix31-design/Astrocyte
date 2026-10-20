import assert from 'node:assert/strict'
import { readFile, realpath, writeFile } from 'node:fs/promises'
import { isAbsolute, join, relative } from 'node:path'
import { parseArgs } from 'node:util'
import { z } from 'zod'
import Database from '../../third_party/orca-v1.4.230/src/main/sqlite/sync-database'
import { DaemonClient } from '../../third_party/orca-v1.4.230/src/main/daemon/client'
import { getDaemonPidPath, getDaemonSocketPath, getDaemonTokenPath } from '../../third_party/orca-v1.4.230/src/main/daemon/daemon-spawner'
import type { ListSessionsResult } from '../../third_party/orca-v1.4.230/src/main/daemon/types'
import { controllerOwnerSchema } from '../../runtime/swa-orca/src/controller-scope'
import { swaDaemonServiceJobSchema } from '../../third_party/orca-v1.4.230/src/main/daemon/swa-service-job'

async function main() {
  const { values } = parseArgs({ options: Object.fromEntries(['candidate', 'data-root', 'owner-file', 'database', 'output'].map(key => [key, { type: 'string' as const }])), allowPositionals: false })
  const required = (key: string) => { const value = z.string().refine(isAbsolute).parse(values[key]); return value }
  const candidate = await realpath(required('candidate')), data = await realpath(required('data-root'))
  assert.equal((await realpath(process.execPath)).toLowerCase(), join(candidate, 'node', 'node.exe').toLowerCase())
  const envelope = z.object({ ok: z.literal(true), request_id: z.uuid(), authority_domain: z.string().min(1), result: controllerOwnerSchema }).parse(JSON.parse(await readFile(required('owner-file'), 'utf8')))
  const databasePath = await realpath(required('database')), rel = relative(data, databasePath)
  assert.ok(rel && rel !== '..' && !rel.startsWith('..\\') && !isAbsolute(rel))
  const db = new Database(databasePath, { readonly: true, fileMustExist: true })
  try {
    const row = z.object({ method: z.enum(['swa.controllerOpen', 'swa.controllerRecover']), state: z.literal('completed'), receipt: z.string() }).parse(db.prepare('SELECT method, state, receipt FROM mutation_receipts WHERE caller_fingerprint=? AND request_id=?').get(envelope.authority_domain, envelope.request_id))
    assert.deepEqual(controllerOwnerSchema.parse(JSON.parse(row.receipt)), envelope.result)
  } finally { db.close() }
  const directory = join(data, 'daemon'), pidPath = getDaemonPidPath(directory)
  assert.equal((await realpath(pidPath)).toLowerCase(), pidPath.toLowerCase())
  const record = z.object({ pid: z.number().int().positive(), launchNonce: z.uuid(), swa_service_job: swaDaemonServiceJobSchema }).parse(JSON.parse(await readFile(pidPath, 'utf8')))
  const client = new DaemonClient({ socketPath: getDaemonSocketPath(directory), tokenPath: getDaemonTokenPath(directory) })
  try {
    await client.ensureConnectedWithin(5000)
    const daemon = client.getDaemonIdentity()
    assert.ok(daemon && daemon.pid === record.pid && daemon.launchNonce === record.launchNonce)
    assert.deepEqual(daemon.swa_service_job, record.swa_service_job)
    const ownership = record.swa_service_job
    assert.equal(ownership.install_root.toLowerCase(), candidate.toLowerCase())
    assert.equal(ownership.data_root.toLowerCase(), data.toLowerCase())
    assert.equal(ownership.source_revision, (await readFile(join(candidate, 'VERSION'), 'utf8')).trim())
    const inventory = await client.request<ListSessionsResult>('listSessions', undefined, 5000)
    const sessions = z.array(z.looseObject({ sessionId: z.string(), incarnationId: z.string().optional(), terminalHandle: z.string().optional(), pid: z.number().nullable(), isAlive: z.boolean(), wslDistro: z.string().nullable().optional() })).parse(inventory.sessions)
    const matches = sessions.filter(session => session.sessionId === envelope.result.binding.ptyId || session.terminalHandle === envelope.result.controller)
    assert.equal(matches.length, 1)
    const session = matches[0]!
    assert.equal(session.sessionId, envelope.result.binding.ptyId)
    assert.equal(session.incarnationId, envelope.result.binding.processIncarnation)
    assert.equal(session.terminalHandle, envelope.result.controller)
    assert.equal(session.pid, envelope.result.process.pid)
    assert.equal(session.isAlive, true)
    assert.ok(session.wslDistro == null)
    await writeFile(required('output'), JSON.stringify({ observed_at: new Date().toISOString(), source_revision: ownership.source_revision, actual_native_owner: envelope.result, daemon, service_job: ownership, session, database: databasePath }, null, 2) + '\n', { flag: 'wx' })
  } finally { client.disconnect() }
}
main().catch(error => { process.stderr.write(String(error) + '\n'); process.exitCode = 1 })
