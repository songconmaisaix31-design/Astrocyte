import assert from 'node:assert/strict'
import { createHash } from 'node:crypto'
import { lstat, mkdir, readFile, realpath, writeFile } from 'node:fs/promises'
import { dirname, isAbsolute, join, relative, resolve, win32 } from 'node:path'
import { fileURLToPath } from 'node:url'
import { parseArgs } from 'node:util'

const original = { revision: 'b464e70ba95967eb1ecc216f5f3d25407f1ecc66', run: 40047585581, job: 120732737576,
  attempt: 1, artifact: 12541713122, artifact_bytes: 65470,
  artifact_digest: 'sha256:378761a5153e5f71bc3f0f3cf7db7da28babc4e95d7ff10d789983cf349cec17' }
const hashes = { fixture: '842060a11640491694dbb663057b71dd94bc58e24905acb006fdc61319a74ac4',
  closure: '144d16fe707b432e9cfc05aa800d0719430ea112fc23c8706a46f9664436475f',
  native: '8b2d203527225d12e67920484aca4a07be35c8675b477e6a6fe84f2b608b389f',
  checker: '8a60c68130f77e94bc29367f56f07ff2368e01aab1e916a3b00a91b7c58e3a6d' }
// 本轮已发布的完整安全投影固定在源码中，运行时不读本地 ignored 证据。
const expectedService = {
  "source_revision": "b464e70ba95967eb1ecc216f5f3d25407f1ecc66",
  "status": "PASS",
  "observer": {
    "file_name": "service-observer-exit.json",
    "status": "SAVED",
    "sha256": "838f0a596bf17fcac9db4236a4ef6b48bf80c591c74455a588fe23bebb8cda75",
    "byte_count": 298,
    "started": true,
    "exit": 0,
    "timed_out": false,
    "stdout_eof": true,
    "stderr_eof": true
  },
  "actual": {
    "file_name": "result.json",
    "status": "SAVED",
    "sha256": "4b3b7c3636019ab6210b21e4d3bd30ad7c4670071c07bb2f64a62a48578bc998",
    "byte_count": 42231,
    "edition": "combined",
    "service_status": "OBSERVED",
    "exit": 0,
    "signal": null,
    "ownership_closure_incomplete": false,
    "owned_count": 88,
    "unknown_count": 0,
    "native_exit": [
      {
        "observed_at_ms": 1792803874159,
        "remaining": 0
      },
      {
        "observed_at_ms": 1792803874413,
        "remaining": 0
      }
    ],
    "controller": {
      "status": "PASS",
      "ownership_incomplete": false,
      "producer": {
        "status": "PASS",
        "phase": "complete",
        "model_calls": 0,
        "worker_start_calls": 0,
        "ownership_incomplete": false,
        "task_create_completed": true,
        "sqlite_original_request_unchanged": true,
        "cleanup": {
          "controller_closed": true,
          "grant_revoked": true,
          "dispatcher_closed": true
        }
      },
      "transport": {
        "status": "PASS",
        "tty": {
          "status": "PASS",
          "natural_eof": true,
          "forced_cleanup": false,
          "reply_count": 3,
          "long_bytes": 6230,
          "exit_code": 0
        },
        "pipe": {
          "status": "PASS",
          "natural_eof": true,
          "forced_cleanup": false,
          "reply_count": 3,
          "long_bytes": 6230,
          "exit_code": 0
        }
      }
    },
    "cim": {
      "exit": 0,
      "signal": null,
      "timed_out": false,
      "query_failed": false,
      "all_exited": true,
      "observations": [
        {
          "monotonic_ms": 353.3871,
          "query_succeeded": true,
          "complete_result_count": 140,
          "required_identity_fields_complete": true,
          "known_identities_exited": true,
          "unknown_pids_absent": true
        },
        {
          "monotonic_ms": 936.9141,
          "query_succeeded": true,
          "complete_result_count": 140,
          "required_identity_fields_complete": true,
          "known_identities_exited": true,
          "unknown_pids_absent": true
        }
      ]
    },
    "kernel_tree_exit_proven": false
  },
  "complete": true,
  "binding": "VERIFIED",
  "sources": {
    "observer": "9863caae32d3015eb2773b7342a4be41d0b9858136f76153dd30913ff22ed323",
    "producer": "e1cba5dfdee30fd91232d38481475fef4b606a56166905d47ee20bc0b676ff34",
    "closure": "144d16fe707b432e9cfc05aa800d0719430ea112fc23c8706a46f9664436475f",
    "native": "8b2d203527225d12e67920484aca4a07be35c8675b477e6a6fe84f2b608b389f"
  },
  "files": [
    {
      "file_name": "service-query-evidence.json",
      "status": "SAVED",
      "sha256": "9cdf185e5d487a519ad3a30614d14ae01fa792a1ad5898ac96b2c15f115768fd",
      "byte_count": 199537
    },
    {
      "file_name": "service-process-lifecycle.json",
      "status": "SAVED",
      "sha256": "5e56ca2041aeea8a06ef15dfea7f94d2f957b5b6ef70a968f740df364f981ba7",
      "byte_count": 128799
    }
  ],
  "query_count": 710,
  "ownership_closure_incomplete": false,
  "evidence_incomplete": false,
  "error_kind": null,
  "first_sample": {
    "sequence": 1,
    "query_start_ms": 74.9606,
    "query_end_ms": 77.5964,
    "actual_row_count": 140,
    "prior_count": 1,
    "used_scope_count": 1,
    "added_during_query_count": 0,
    "root_verified": true,
    "unknown_count": 0
  },
  "first_unknown_sample": null,
  "last_sample": {
    "sequence": 710,
    "query_start_ms": 28046.2709,
    "query_end_ms": 28048.0246,
    "actual_row_count": 139,
    "prior_count": 88,
    "used_scope_count": 25,
    "added_during_query_count": 0,
    "root_verified": true,
    "unknown_count": 0
  },
  "lifecycle_event_count": 199,
  "lifecycle_incomplete": false
}
const expectedFailure = { observation_revision: original.revision, status: 'PASS', failure_records_match: true,
  first_failure: 'NOT_CAPTURED', observer_abort: 'NOT_CAPTURED', first_nonzero_command: 'NOT_CAPTURED', outer_error: 'NOT_CAPTURED' }
const stages = ['frozen-install', 'go-download', 'go-verify', 'build', 'extracted-check', 'install', 'installed-check',
  'process-incarnation', 'service-smoke', 'service-evidence-summary']
const limits = new Map([['SOURCE', 128], ['result.json', 65536], ['logs/installed-check.log', 262144],
  ...stages.map(name => ['logs/' + name + '.exit.txt', 64]),
  ...['result.json', 'continuity-snapshot-a.json', 'continuity-snapshot-b.json', 'continuity-snapshot-c.json',
    'continuity-exit-wait.json', 'cim-exit-observations.json'].map(name => ['process-incarnation/' + name, 2 * 1024 * 1024]),
  ['service-observer-exit.json', 4096], ['service-evidence-summary.json', 16384]])
const sha = bytes => createHash('sha256').update(bytes).digest('hex')
const object = value => value !== null && typeof value === 'object' && !Array.isArray(value)
const keys = (value, required, optional = []) => {
  assert.ok(object(value)); assert.ok(required.every(key => Object.hasOwn(value, key)))
  assert.ok(Object.keys(value).every(key => required.includes(key) || optional.includes(key)))
}
const uint = value => Number.isSafeInteger(value) && value >= 0 && value <= 0xffffffff
const positive = value => Number.isSafeInteger(value) && value > 0
const finite = value => typeof value === 'number' && Number.isFinite(value) && value >= 0
const bool = value => assert.equal(typeof value, 'boolean')
const digest = value => assert.ok(typeof value === 'string' && /^[0-9a-f]{64}$/.test(value))
const list = (value, max, check) => { assert.ok(Array.isArray(value) && value.length <= max); value.forEach(check) }
const identity = (value, parent = false) => {
  keys(value, ['pid', 'creationTimeMs', ...(parent ? ['ppid'] : [])])
  assert.ok(uint(value.pid) && value.pid > 0 && positive(value.creationTimeMs))
  if (parent) assert.ok(uint(value.ppid))
}
const instant = value => { keys(value, ['observed_at_ms', 'monotonic_ms']); assert.ok(positive(value.observed_at_ms) && finite(value.monotonic_ms)) }
const window = value => {
  keys(value, ['start', 'end']); instant(value.start); instant(value.end)
  assert.ok(value.start.monotonic_ms <= value.end.monotonic_ms)
}
const exit0 = value => assert.deepEqual(value, { exit: 0, signal: null, spawn_failed: false })
const nativeRow = value => {
  keys(value, ['pid', 'ppid', 'name', 'creationTimeDiagnosticVersion', 'creationTimeProvider', 'creationTimeNtStatus', 'creationTimeSnapshotComplete'], ['creationTimeMs'])
  assert.ok(uint(value.pid) && value.pid > 0 && uint(value.ppid))
  assert.ok(typeof value.name === 'string' && value.name.length <= 260 && /^[\p{L}\p{N}_. ()-]*$/u.test(value.name))
  if (Object.hasOwn(value, 'creationTimeMs')) assert.ok(positive(value.creationTimeMs))
  assert.equal(value.creationTimeDiagnosticVersion, 2); assert.equal(value.creationTimeProvider, 'NtQuerySystemInformation')
  assert.equal(value.creationTimeSnapshotComplete, true); assert.equal(value.creationTimeNtStatus, 0)
}
const closure = value => {
  keys(value, ['rootVerified', 'ownedRows', 'excludedEdges', 'unknownRows']); bool(value.rootVerified)
  list(value.ownedRows, 16384, nativeRow)
  list(value.excludedEdges, 32768, edge => {
    assert.ok(['different-root-incarnation', 'predates-root', 'predates-parent'].includes(edge.reason))
    const key = edge.reason === 'different-root-incarnation' ? 'anchor' : 'parent'
    keys(edge, ['reason', 'actual', key]); nativeRow(edge.actual); identity(edge[key])
  })
  list(value.unknownRows, 32768, entry => {
    keys(entry, ['reason', 'actual'], ['source_anchor']); nativeRow(entry.actual)
    assert.ok(['invalid_native_process_row', 'current_root_identity_unverified', 'root_in_descendant_branch', 'parent_incarnation_unverified', 'process_identity_not_unique', 'creation_time_unavailable', 'parent_incarnation_changed'].includes(entry.reason))
    if (Object.hasOwn(entry, 'source_anchor')) { keys(entry.source_anchor, ['kind', 'pid', 'creation_time_ms']); assert.equal(entry.source_anchor.kind, 'additional'); assert.ok(uint(entry.source_anchor.pid) && entry.source_anchor.pid > 0 && positive(entry.source_anchor.creation_time_ms)) }
  })
}
const sameIdentity = (a, b) => a.pid === b.pid && a.creationTimeMs === b.creationTimeMs
const unique = (rows, id) => { const found = rows.filter(row => row.pid === id.pid); assert.equal(found.length, 1); assert.equal(found[0].creationTimeMs, id.creationTimeMs); return found[0] }
const state = { schema_version: 1, original, original_status: 'SUCCESS', status: 'READ_FAILED', phase: 'arguments', read_inputs: [], files: [],
  product_execution: 'NOT_RUN', fixture_execution: 'NOT_RUN', native_build: 'NOT_RUN', model_calls: 0, kernel_tree_exit_proven: false }
let output
const stableFields = ['dev', 'ino', 'size', 'mtimeMs', 'ctimeMs', 'nlink']
async function plainDirectory(path) {
  const stat = await lstat(path)
  assert.ok(stat.isDirectory() && !stat.isSymbolicLink()); assert.equal((await realpath(path)).toLowerCase(), resolve(path).toLowerCase())
}
function installedCheck(text) {
  // 全文只接受原 checker 的 native 行及外层行，不从其它日志捞取字段。
  const lines = text.split(/\r?\n/)
  if (lines.at(-1) === '') lines.pop()
  assert.equal(lines.length, 2); const [native, candidate] = lines.map(line => JSON.parse(line))
  keys(native, ['native_node', 'native_pid', 'native_creation_time_ms', 'provider_job_api', 'actual_empty_job', 'job_descendant_test', 'identity_snapshot'])
  assert.equal(native.native_node, 'v24.21.0'); assert.ok(uint(native.native_pid) && native.native_pid > 0 && positive(native.native_creation_time_ms))
  assert.equal(native.provider_job_api, 1); assert.equal(native.job_descendant_test, 'NOT_RUN')
  const job = native.actual_empty_job
  keys(job, ['ownerPid', 'rootPid', 'rootCreationTimeMs', 'assigned', 'rootExited', 'limitFlags', 'members'])
  assert.equal(job.ownerPid, native.native_pid); assert.equal(job.rootPid, 0); assert.equal(job.rootCreationTimeMs, 0)
  assert.equal(job.assigned, false); assert.equal(job.rootExited, false); assert.ok(uint(job.limitFlags)); assert.deepEqual(job.members, [])
  const snapshot = native.identity_snapshot
  keys(snapshot, ['rows', 'provider', 'version', 'complete', 'nt_status', 'self_identity_matches', 'loaded_addon_sha256', 'psutil_license_verified'])
  assert.ok(positive(snapshot.rows)); assert.equal(snapshot.provider, 'NtQuerySystemInformation'); assert.equal(snapshot.version, 2)
  assert.equal(snapshot.complete, true); assert.ok(uint(snapshot.nt_status) && snapshot.nt_status <= 0x7fffffff)
  assert.equal(snapshot.self_identity_matches, true); assert.equal(snapshot.loaded_addon_sha256, hashes.native); assert.equal(snapshot.psutil_license_verified, true)
  keys(candidate, ['candidate', 'source_revision', 'platform', 'edition', 'legacy_edition', 'web_manifest', 'codex_catalog', 'closure_paths', 'native_authentication', 'functional_acceptance', 'gui_http_acceptance'])
  assert.ok(typeof candidate.candidate === 'string' && candidate.candidate.length <= 4096 && win32.isAbsolute(candidate.candidate))
  const name = 'swa-windows-combined-' + original.revision
  assert.equal(win32.basename(candidate.candidate), name); assert.equal(candidate.source_revision, original.revision)
  assert.equal(candidate.platform, 'win32-x64'); assert.equal(candidate.edition, 'combined'); assert.equal(candidate.legacy_edition, false)
  keys(candidate.web_manifest, ['entries', 'referenced_files']); assert.ok(positive(candidate.web_manifest.entries) && positive(candidate.web_manifest.referenced_files))
  keys(candidate.codex_catalog, ['version', 'sha256']); assert.equal(candidate.codex_catalog.version, '0.163.0'); digest(candidate.codex_catalog.sha256)
  assert.ok(positive(candidate.closure_paths)); for (const key of ['native_authentication', 'functional_acceptance', 'gui_http_acceptance']) assert.equal(candidate[key], 'NOT_RUN')
  return { source_revision: candidate.source_revision, candidate_basename: name, platform: candidate.platform, edition: candidate.edition,
    checker_source_sha256: hashes.checker, native_node: native.native_node, identity_snapshot: snapshot,
    psutil_license_bytes: 'LICENSE_BYTES_NOT_RETAINED', candidate_closure_paths: candidate.closure_paths }
}
function ordinaryResult(value) {
  keys(value, ['schema_version', 'diagnostic_only', 'source_revision', 'helper_sha256', 'closure_sha256', 'status', 'control_complete', 'probe_exit', 'cases',
    'candidate_built', 'install_started', 'service_started', 'account_session_started', 'model_calls', 'kernel_tree_exit_proven', 'native_component'], ['cleanup_attempted', 'cleanup_unknown'])
  assert.equal(value.schema_version, 1); assert.equal(value.diagnostic_only, true); assert.equal(value.source_revision, original.revision)
  assert.equal(value.helper_sha256, hashes.fixture); assert.equal(value.closure_sha256, hashes.closure)
  assert.equal(value.status, 'CONTROL_COMPLETE'); assert.equal(value.control_complete, true); assert.equal(value.probe_exit, 0)
  for (const key of ['candidate_built', 'install_started', 'service_started', 'account_session_started', 'kernel_tree_exit_proven']) assert.equal(value[key], false)
  assert.equal(value.model_calls, 0)
  for (const key of ['cleanup_attempted', 'cleanup_unknown']) if (Object.hasOwn(value, key)) assert.equal(value[key], false)
  const component = value.native_component
  keys(component, ['name', 'version', 'addon_sha256', 'supported_process_data_flags', 'loaded_binary_matches_compiled'])
  assert.equal(component.name, '@vscode/windows-process-tree'); assert.equal(component.version, '0.8.0'); assert.equal(component.addon_sha256, hashes.native)
  assert.ok(uint(component.supported_process_data_flags) && (component.supported_process_data_flags & 4) === 4); assert.equal(component.loaded_binary_matches_compiled, true)
  assert.deepEqual(value.cases.map(row => row.name), ['live-root-child', 'normal-eof', 'prior-owned-continuity'])
  const [live, normal, continuity] = value.cases
  keys(live, ['name', 'observations']); assert.equal(live.observations.length, 2)
  for (const row of live.observations) {
    keys(row, ['observed_at_ms', 'actual_row_count', 'root_verified', 'owned_count', 'owned', 'owned_truncated', 'excluded_edge_count', 'unknown_count'])
    assert.ok(positive(row.observed_at_ms) && positive(row.actual_row_count) && positive(row.owned_count) && uint(row.excluded_edge_count))
    list(row.owned, 8, value => identity(value, true)); bool(row.owned_truncated)
    assert.equal(row.owned_truncated, row.owned_count > 8); assert.equal(row.owned.length, Math.min(row.owned_count, 8))
    assert.equal(row.root_verified, true); assert.equal(row.unknown_count, 0)
  }
  keys(normal, ['name', 'native_exit_observations', 'root_exit', 'child_wait', 'root_stdout_eof', 'root_stderr_eof', 'root_stdin_failed', 'cim_process_exit', 'cim_stdout_eof', 'cim_stderr_eof', 'cim_report'])
  exit0(normal.root_exit); exit0(normal.cim_process_exit)
  for (const key of ['root_stdout_eof', 'root_stderr_eof', 'cim_stdout_eof', 'cim_stderr_eof']) assert.equal(normal[key], true)
  assert.equal(normal.root_stdin_failed, false)
  keys(normal.child_wait, ['type', 'pid', 'actual_exit', 'stdout_eof', 'stderr_eof'])
  assert.equal(normal.child_wait.type, 'fixture-child-closed'); assert.ok(uint(normal.child_wait.pid) && normal.child_wait.pid > 0)
  exit0(normal.child_wait.actual_exit); assert.equal(normal.child_wait.stdout_eof, true); assert.equal(normal.child_wait.stderr_eof, true)
  assert.equal(normal.native_exit_observations.length, 2)
  for (const row of normal.native_exit_observations) {
    keys(row, ['observed_at_ms', 'actual_row_count', 'remaining_count', 'remaining', 'remaining_truncated', 'root_rows_count', 'root_verified', 'owned_count', 'unknown_count'])
    assert.ok(positive(row.observed_at_ms) && positive(row.actual_row_count)); assert.deepEqual(row.remaining, [])
    for (const key of ['remaining_count', 'root_rows_count', 'owned_count', 'unknown_count']) assert.equal(row[key], 0)
    assert.equal(row.remaining_truncated, false); assert.equal(row.root_verified, false)
  }
  assert.ok(normal.native_exit_observations[1].observed_at_ms >= normal.native_exit_observations[0].observed_at_ms)
  keys(continuity, ['name', 'status', 'stage', 'snapshots', 'forced_cleanup', 'parent_close_observed', 'controller', 'parent', 'prior_a', 'child_requested',
    'child_ready_observed', 'child', 'prior_b', 'parent_exit', 'parent_stdout_eof', 'parent_stderr_eof', 'comparison', 'exit_wait_file', 'native_exit_observations', 'child_exit_code'], ['cleanup_attempted', 'cleanup_unknown'])
  assert.equal(continuity.status, 'PASS'); assert.equal(continuity.stage, 'complete'); assert.equal(continuity.forced_cleanup, false)
  for (const key of ['cleanup_attempted', 'cleanup_unknown']) if (Object.hasOwn(continuity, key)) assert.equal(continuity[key], false)
  for (const key of ['controller', 'parent', 'child']) identity(continuity[key], true)
  instant(continuity.parent_close_observed); instant(continuity.child_ready_observed); exit0(continuity.parent_exit)
  for (const key of ['child_requested', 'parent_stdout_eof', 'parent_stderr_eof']) assert.equal(continuity[key], true)
  for (const key of ['prior_a', 'prior_b']) list(continuity[key], 2048, row => identity(row))
  assert.equal(continuity.child_exit_code, 'NOT_OBSERVABLE_FROM_NONPARENT')
  const comparison = continuity.comparison
  keys(comparison, ['original_owned_child', 'prior_a_owned_child', 'prior_b_owned_child', 'original_root_verified', 'prior_a_root_verified', 'prior_b_root_verified', 'original_unknown_count', 'prior_a_unknown_count', 'prior_b_unknown_count', 'expected_unknown'])
  assert.equal(comparison.original_owned_child, false); assert.equal(comparison.prior_a_owned_child, false); assert.equal(comparison.prior_b_owned_child, true)
  for (const key of ['original_root_verified', 'prior_a_root_verified', 'prior_b_root_verified']) assert.equal(comparison[key], true)
  assert.equal(comparison.original_unknown_count, 0); assert.equal(comparison.prior_b_unknown_count, 0)
  assert.ok(positive(comparison.prior_a_unknown_count)); assert.equal(comparison.expected_unknown.length, comparison.prior_a_unknown_count)
  list(comparison.expected_unknown, 16384, entry => {
    keys(entry, ['reason', 'actual', 'source_anchor']); assert.equal(entry.reason, 'parent_incarnation_unverified')
    assert.deepEqual(entry.actual, continuity.child)
    assert.deepEqual(entry.source_anchor, { kind: 'additional', pid: continuity.parent.pid, creation_time_ms: continuity.parent.creationTimeMs })
  })
  assert.equal(continuity.native_exit_observations.length, 2)
  for (const row of continuity.native_exit_observations) {
    keys(row, ['query_window', 'actual_row_count', 'remaining_count', 'remaining', 'root_verified', 'unknown_count']); window(row.query_window)
    assert.ok(positive(row.actual_row_count)); assert.equal(row.remaining_count, 0); assert.deepEqual(row.remaining, [])
    assert.equal(row.root_verified, true); assert.equal(row.unknown_count, 0)
  }
  assert.ok(continuity.native_exit_observations[1].query_window.start.monotonic_ms >= continuity.native_exit_observations[0].query_window.end.monotonic_ms)
  return { live, normal, continuity }
}
function cimReport(value, normal) {
  keys(value, ['schema_version', 'observations', 'query_failure', 'all_exited']); assert.equal(value.schema_version, 1)
  assert.equal(value.query_failure, null); assert.equal(value.all_exited, true); assert.equal(value.observations.length, 2)
  for (const row of value.observations) {
    keys(row, ['query_succeeded', 'observed_at', 'monotonic_ms', 'complete_result_count', 'invalid_pid_record_count', 'actual_matches', 'required_identity_fields_complete', 'known_identities_exited', 'unknown_pids_absent'])
    for (const key of ['query_succeeded', 'required_identity_fields_complete', 'known_identities_exited', 'unknown_pids_absent']) assert.equal(row[key], true)
    assert.ok(typeof row.observed_at === 'string' && row.observed_at.length <= 64 && Number.isFinite(Date.parse(row.observed_at)))
    assert.ok(finite(row.monotonic_ms) && positive(row.complete_result_count) && uint(row.invalid_pid_record_count))
    list(row.actual_matches, 16384, match => {
      keys(match, ['pid', 'parent_pid', 'name', 'creation_time_ms', 'creation_utc'])
      assert.ok(uint(match.pid) && match.pid > 0 && uint(match.parent_pid) && positive(match.creation_time_ms))
      assert.ok(typeof match.name === 'string' && match.name.length <= 260 && /^[\p{L}\p{N}_. ()-]*$/u.test(match.name))
      assert.ok(typeof match.creation_utc === 'string' && match.creation_utc.length <= 64 && Number.isFinite(Date.parse(match.creation_utc)))
    })
  }
  assert.ok(value.observations[1].monotonic_ms - value.observations[0].monotonic_ms >= 50)
  const projected = { schema_version: 1, all_exited: true, query_failure: null, observation_count: 2, observations_truncated: false,
    observations: value.observations.map(row => ({ query_succeeded: row.query_succeeded, complete_result_count: row.complete_result_count,
      invalid_pid_record_count: row.invalid_pid_record_count, required_identity_fields_complete: row.required_identity_fields_complete,
      known_identities_exited: row.known_identities_exited, unknown_pids_absent: row.unknown_pids_absent, monotonic_ms: row.monotonic_ms, actual_match_count: row.actual_matches.length })) }
  assert.deepEqual(normal.cim_report, projected)
  return projected
}
async function main() {
  const { values } = parseArgs({ options: { 'artifact-root': { type: 'string' } }, strict: true, allowPositionals: false })
  assert.ok(values['artifact-root'] && isAbsolute(values['artifact-root'])); assert.ok(expectedService)
  const checkout = await realpath(fileURLToPath(new URL('../', import.meta.url)))
  const input = resolve(values['artifact-root']); await plainDirectory(input)
  assert.equal(input.toLowerCase(), join(await realpath(process.env.RUNNER_TEMP), 'swa-retained-formal-b464').toLowerCase())
  for (const path of [join(checkout, '.verification'), join(checkout, '.verification/swa-ci')]) {
    try { await mkdir(path) } catch (error) { if (error.code !== 'EEXIST') throw error }
    await plainDirectory(path)
  }
  const destination = join(checkout, '.verification/swa-ci/windows-formal-retained')
  await mkdir(destination); await plainDirectory(destination); output = destination
  const readerPath = fileURLToPath(import.meta.url), readerBefore = await lstat(readerPath), readerSHA = sha(await readFile(readerPath))
  assert.ok(readerBefore.isFile() && !readerBefore.isSymbolicLink() && readerBefore.nlink === 1)
  assert.equal((await realpath(readerPath)).toLowerCase(), resolve(readerPath).toLowerCase())
  const tracked = [], readNames = new Set(); let total = 0
  const read = async (name, json = true) => {
    assert.ok(limits.has(name) && !readNames.has(name)); readNames.add(name); assert.ok(readNames.size <= 24)
    const path = join(input, name); assert.ok(!relative(input, path).startsWith('..'))
    for (let parent = dirname(path); ; parent = dirname(parent)) { await plainDirectory(parent); if (parent.toLowerCase() === input.toLowerCase()) break }
    const before = await lstat(path)
    assert.ok(before.isFile() && !before.isSymbolicLink() && before.nlink === 1 && before.size <= limits.get(name))
    assert.equal((await realpath(path)).toLowerCase(), resolve(path).toLowerCase()); total += before.size; assert.ok(total <= 24 * 1024 * 1024)
    const bytes = await readFile(path), after = await lstat(path)
    stableFields.forEach(key => assert.equal(after[key], before[key])); assert.equal(bytes.length, before.size)
    const inputSHA = sha(bytes)
    state.read_inputs.push({ source_name: name, byte_count: bytes.length, sha256: inputSHA })
    const text = new TextDecoder('utf-8', { fatal: true }).decode(bytes)
    const actual = { name, path, before, bytes, sha256: inputSHA, value: json ? JSON.parse(text) : text }; tracked.push(actual); return actual
  }
  const copy = async actual => {
    const name = actual.name.replaceAll('/', '--'); await writeFile(join(output, name), actual.bytes, { flag: 'wx' })
    assert.equal(sha(await readFile(join(output, name))), actual.sha256)
    state.files.push({ source_name: actual.name, saved_name: name, byte_count: actual.bytes.length, sha256: actual.sha256 })
  }
  state.phase = 'source-and-result'
  const source = await read('SOURCE', false); assert.equal(source.value.trim(), original.revision); await copy(source)
  const top = await read('result.json'), candidate = top.value
  keys(candidate, ['source_revision', 'platform', 'edition', 'status', 'model_calls', 'agent_acceptance', 'archive_sha256', 'installed_version', 'service', 'whole_service_job_shutdown', 'service_evidence'])
  assert.equal(candidate.source_revision, original.revision); assert.equal(candidate.installed_version, original.revision)
  assert.equal(candidate.platform, 'win32-x64'); assert.equal(candidate.edition, 'combined'); assert.equal(candidate.status, 'PASS')
  assert.equal(candidate.model_calls, 0); assert.equal(candidate.agent_acceptance, 'NOT_RUN'); digest(candidate.archive_sha256)
  for (const key of ['service', 'whole_service_job_shutdown', 'service_evidence']) assert.equal(candidate[key], 'PASS')
  await copy(top)
  state.phase = 'stage-exits'
  const exits = {}
  for (const name of stages) { const actual = await read('logs/' + name + '.exit.txt', false); assert.ok(/^\s*-?\d+\s*$/.test(actual.value)); const code = Number(actual.value.trim()); assert.ok(Number.isSafeInteger(code)); exits[name] = code; await copy(actual); assert.equal(code, 0) }
  state.phase = 'installed-check-json'
  const checker = await read('logs/installed-check.log', false), checked = installedCheck(checker.value)
  checked.original_byte_count = checker.bytes.length; checked.original_sha256 = checker.sha256
  const checkedBytes = Buffer.from(JSON.stringify(checked, null, 2) + '\n')
  await writeFile(join(output, 'installed-check-safe.json'), checkedBytes, { flag: 'wx' })
  const savedCheck = await readFile(join(output, 'installed-check-safe.json'))
  assert.equal(savedCheck.length, checkedBytes.length); assert.equal(sha(savedCheck), sha(checkedBytes))
  state.files.push({ source_name: checker.name, original_retained: false, original_byte_count: checker.bytes.length, original_sha256: checker.sha256,
    saved_name: 'installed-check-safe.json', byte_count: savedCheck.length, sha256: sha(savedCheck) })
  state.phase = 'ordinary-result'
  const fixture = await read('process-incarnation/result.json'), { live, normal, continuity } = ordinaryResult(fixture.value)
  state.phase = 'ordinary-frames'
  assert.deepEqual(continuity.snapshots.map(row => row.label), ['a', 'b', 'c'])
  const frames = []
  for (const declaration of continuity.snapshots) {
    keys(declaration, ['label', 'file_name', 'sha256', 'byte_count', 'actual_row_count', 'query_window', 'native_query'])
    assert.equal(declaration.file_name, 'continuity-snapshot-' + declaration.label + '.json')
    const actual = await read('process-incarnation/' + declaration.file_name), frame = actual.value
    keys(frame, ['schema_version', 'query_window', 'rows', 'label']); assert.equal(frame.schema_version, 1); assert.equal(frame.label, declaration.label)
    window(frame.query_window); list(frame.rows, 16384, nativeRow); assert.ok(frame.rows.length > 0)
    assert.equal(actual.sha256, declaration.sha256); assert.equal(actual.bytes.length, declaration.byte_count)
    assert.equal(frame.rows.length, declaration.actual_row_count); assert.deepEqual(frame.query_window, declaration.query_window)
    assert.deepEqual(declaration.native_query, { provider: 'NtQuerySystemInformation', version: 2, ntstatus: 0, snapshot_complete: true })
    frames.push(frame); await copy(actual)
  }
  const [a, b, c] = frames
  for (const frame of frames) unique(frame.rows, continuity.controller)
  assert.equal(unique(a.rows, continuity.parent).ppid, continuity.controller.pid)
  assert.equal(unique(b.rows, continuity.parent).ppid, continuity.controller.pid)
  assert.equal(unique(b.rows, continuity.child).ppid, continuity.parent.pid); assert.equal(unique(c.rows, continuity.child).ppid, continuity.parent.pid)
  assert.equal(c.rows.filter(row => row.pid === continuity.parent.pid).length, 0)
  assert.ok(!continuity.prior_a.some(row => sameIdentity(row, continuity.child))); assert.ok(continuity.prior_b.some(row => sameIdentity(row, continuity.child)))
  for (const [rows, priors] of [[a.rows, continuity.prior_a], [b.rows, continuity.prior_b]]) for (const id of priors) unique(rows, id)
  state.phase = 'ordinary-exit-wait'
  const waitFile = await read('process-incarnation/continuity-exit-wait.json'), wait = waitFile.value
  keys(wait, ['schema_version', 'source_revision', 'helper_sha256', 'closure_sha256', 'controller', 'parent', 'child', 'prior_b', 'child_ready_observed', 'child_deadline_ms', 'wait_deadline_ms', 'budget_ms', 'interval_ms', 'maximum_samples', 'status', 'stop_reason', 'retention_incomplete', 'unretained_sample_count', 'first_zero_index', 'samples'])
  assert.equal(wait.schema_version, 1); assert.equal(wait.source_revision, original.revision); assert.equal(wait.helper_sha256, hashes.fixture); assert.equal(wait.closure_sha256, hashes.closure)
  for (const key of ['controller', 'parent', 'child', 'prior_b', 'child_ready_observed']) assert.deepEqual(wait[key], continuity[key])
  assert.equal(wait.child_deadline_ms, wait.child_ready_observed.monotonic_ms + 30000); assert.equal(wait.wait_deadline_ms, wait.child_deadline_ms + 5000)
  assert.equal(wait.budget_ms, 5000); assert.equal(wait.interval_ms, 250); assert.equal(wait.maximum_samples, 21)
  assert.equal(wait.status, 'ABSENCE_OBSERVED'); assert.equal(wait.stop_reason, 'FIRST_ZERO'); assert.equal(wait.retention_incomplete, false); assert.equal(wait.unretained_sample_count, 0)
  assert.ok(Array.isArray(wait.samples) && wait.samples.length > 0 && wait.samples.length <= 21)
  for (const [index, sample] of wait.samples.entries()) {
    keys(sample, ['index', 'query_window', 'rows', 'remaining', 'closure']); assert.equal(sample.index, index); window(sample.query_window)
    list(sample.rows, 16384, nativeRow); assert.ok(sample.rows.length > 0); unique(sample.rows, wait.controller)
    list(sample.remaining, 2, row => { identity(row, true); assert.ok([wait.parent, wait.child].some(id => sameIdentity(id, row))) })
    const expectedRemaining = [wait.parent, wait.child].filter(id => sample.rows.some(row => row.pid === id.pid && (!positive(row.creationTimeMs) || row.creationTimeMs === id.creationTimeMs)))
    assert.deepEqual(sample.remaining, expectedRemaining)
    closure(sample.closure); assert.equal(sample.closure.rootVerified, true); assert.equal(sample.closure.unknownRows.length, 0)
    assert.ok(sample.query_window.end.monotonic_ms <= wait.wait_deadline_ms)
    if (index) assert.ok(sample.query_window.start.monotonic_ms >= wait.samples[index - 1].query_window.end.monotonic_ms)
    for (const row of sample.closure.ownedRows) assert.ok(sample.rows.some(actual => JSON.stringify(actual) === JSON.stringify(row)))
  }
  assert.equal(wait.first_zero_index, wait.samples.length - 1); assert.equal(wait.samples.at(-1).remaining.length, 0)
  assert.ok(wait.samples.slice(0, -1).every(sample => sample.remaining.length > 0))
  assert.deepEqual(continuity.exit_wait_file, { file_name: 'continuity-exit-wait.json', sha256: waitFile.sha256, status: wait.status, sample_count: wait.samples.length, retention_incomplete: false })
  await copy(waitFile)
  state.phase = 'ordinary-cim'
  const cim = await read('process-incarnation/cim-exit-observations.json'), cimSummary = cimReport(cim.value, normal)
  await copy(cim); await copy(fixture)
  state.phase = 'service-safe-originals'
  const observer = await read('service-observer-exit.json')
  assert.equal(observer.bytes.length, 298); assert.equal(observer.sha256, '838f0a596bf17fcac9db4236a4ef6b48bf80c591c74455a588fe23bebb8cda75')
  keys(observer.value, ['schema_version', 'source_revision', 'phase', 'started', 'exit', 'timed_out', 'stdout_bytes', 'stderr_bytes', 'stdout_eof', 'stderr_eof', 'outer_error'])
  assert.equal(observer.value.schema_version, 1); assert.equal(observer.value.source_revision, original.revision); assert.equal(observer.value.phase, 'service-observer')
  for (const key of ['started', 'stdout_eof', 'stderr_eof']) assert.equal(observer.value[key], true)
  assert.equal(observer.value.exit, 0); assert.equal(observer.value.timed_out, false); assert.equal(observer.value.outer_error, null)
  assert.ok(uint(observer.value.stdout_bytes) && uint(observer.value.stderr_bytes)); await copy(observer)
  const summary = await read('service-evidence-summary.json')
  assert.deepEqual(summary.value, { schema_version: 1, source_revision: original.revision, status: 'PASS', first_failure: expectedFailure, service_evidence: expectedService })
  await copy(summary)
  state.phase = 'source-stability'
  for (const actual of tracked) { const after = await lstat(actual.path); stableFields.forEach(key => assert.equal(after[key], actual.before[key])); assert.equal(sha(await readFile(actual.path)), actual.sha256) }
  const readerAfter = await lstat(readerPath); stableFields.forEach(key => assert.equal(readerAfter[key], readerBefore[key])); assert.equal(sha(await readFile(readerPath)), readerSHA)
  state.phase = 'annotation-budget'
  const projectSample = sample => sample ? { index: sample.index, query_window: sample.query_window, remaining: sample.remaining, root_verified: sample.closure.rootVerified, unknown_count: sample.closure.unknownRows.length } : null
  const messages = [
    ['Formal source and packaging', { original, reader_sha256: readerSHA, status: 'READ_COMPLETE', candidate, stage_exits: exits, product_execution: 'NOT_RUN' }],
    ['Formal installed checker', checked],
    ['Formal ordinary live and EOF', { source_revision: original.revision, status: fixture.value.status, native_component: fixture.value.native_component, live, normal,
      cleanup_attempted: fixture.value.cleanup_attempted ?? 'NOT_RECORDED', cleanup_unknown: fixture.value.cleanup_unknown ?? 'NOT_RECORDED', kernel_tree_exit_proven: false }],
    ['Formal ordinary continuity', { source_revision: original.revision, status: continuity.status, stage: continuity.stage, controller: continuity.controller, parent: continuity.parent,
      child: continuity.child, prior_a: continuity.prior_a, prior_b: continuity.prior_b, comparison: continuity.comparison, parent_exit: continuity.parent_exit,
      parent_close_observed: continuity.parent_close_observed, parent_stdout_eof: continuity.parent_stdout_eof, parent_stderr_eof: continuity.parent_stderr_eof,
      child_exit_code: continuity.child_exit_code, forced_cleanup: continuity.forced_cleanup, cleanup_unknown: continuity.cleanup_unknown ?? 'NOT_RECORDED' }],
    ['Formal ordinary exit wait', { source_revision: original.revision, status: wait.status, sample_count: wait.samples.length, budget_ms: wait.budget_ms,
      child_deadline_ms: wait.child_deadline_ms, wait_deadline_ms: wait.wait_deadline_ms, retention_incomplete: wait.retention_incomplete,
      unretained_sample_count: wait.unretained_sample_count, first_zero: projectSample(wait.samples[wait.first_zero_index]),
      last_nonzero: projectSample(wait.samples.findLast(sample => sample.remaining.length > 0)), final_native: continuity.native_exit_observations }],
    ['Formal ordinary CIM', { source_revision: original.revision, ...cimSummary }],
    ['Formal original files 1', { original, files: state.files.slice(0, 11) }], ['Formal original files 2', { original, files: state.files.slice(11) }],
    ['Formal service first failure', expectedFailure], ['Formal service evidence', expectedService]
  ].map(([title, value]) => ({ title, text: JSON.stringify(value) }))
  assert.equal(readNames.size, limits.size); assert.equal(state.files.length, 21); assert.ok(messages.length <= 10)
  assert.ok(messages.every(row => Buffer.byteLength(row.text) <= 3500)); assert.ok(messages.reduce((n, row) => n + Buffer.byteLength(row.text), 0) <= 32768)
  state.status = 'READ_COMPLETE'; state.phase = 'complete'; state.reader_sha256 = readerSHA
  state.annotation_bytes = messages.map(row => Buffer.byteLength(row.text)); state.original_archive_sha256 = candidate.archive_sha256
  await writeFile(join(output, 'summary.json'), JSON.stringify(state, null, 2) + '\n', { flag: 'wx' })
  for (const row of messages) process.stdout.write('::warning title=' + row.title + '::' + row.text.replaceAll('%', '%25').replaceAll('\r', '%0D').replaceAll('\n', '%0A') + '\n')
}
try { await main() } catch (error) {
  state.status = 'READ_FAILED'
  state.error_kind = ['ERR_ASSERTION', 'ENOENT', 'EACCES', 'EPERM', 'EEXIST', 'ENOSPC', 'EIO'].includes(error?.code) ? error.code : ['SyntaxError', 'TypeError', 'RangeError'].includes(error?.name) ? error.name : 'UNKNOWN'
  if (output) try { await writeFile(join(output, 'failure.json'), JSON.stringify(state, null, 2) + '\n', { flag: 'wx' }) } catch { /* 保留先前有效安全原件。 */ }
  process.stdout.write('::error title=Fixed formal evidence read failed::' + JSON.stringify({ status: 'READ_FAILED', phase: state.phase, error_kind: state.error_kind, original, product_execution: 'NOT_RUN' }) + '\n')
  process.exitCode = 1
}
