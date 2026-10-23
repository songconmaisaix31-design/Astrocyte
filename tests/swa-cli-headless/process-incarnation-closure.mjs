const unsignedPID = value => Number.isSafeInteger(value) && value >= 0 && value <= 0xffffffff
const positivePID = value => unsignedPID(value) && value > 0
const creation = value => Number.isSafeInteger(value) && value > 0

export function incarnationOrder(parentCreationTimeMs, childCreationTimeMs) {
  if (!creation(parentCreationTimeMs) || !creation(childCreationTimeMs)) return 'unknown'
  return childCreationTimeMs < parentCreationTimeMs ? 'predates-parent' : 'ordered'
}

// Each call consumes one actual native callback batch. Historical rows must
// never fill a missing parent in this batch or become a different incarnation.
export function collectOwnedIncarnations(actualRows, root) {
  if (!Array.isArray(actualRows)) throw new TypeError('Actual process snapshot unavailable')
  const byPID = new Map(), byParent = new Map()
  const ownedRows = [], excludedEdges = [], unknownRows = []
  const unknownSeen = new Set(), visitedRows = new Set(), queuedPIDs = new Set()
  const unknown = (reason, actual) => {
    if (!unknownSeen.has(actual)) {
      unknownSeen.add(actual)
      unknownRows.push({ reason, actual })
    }
  }
  for (const row of actualRows) {
    if (!row || typeof row !== 'object' || !unsignedPID(row.pid) || !unsignedPID(row.ppid)) {
      unknown('invalid_native_process_row', row)
      continue
    }
    const identities = byPID.get(row.pid) ?? []
    identities.push(row)
    byPID.set(row.pid, identities)
    const children = byParent.get(row.ppid) ?? []
    children.push(row)
    byParent.set(row.ppid, children)
  }
  const roots = byPID.get(root.pid) ?? []
  const rootVerified = positivePID(root.pid) && creation(root.creationTimeMs) && roots.length === 1 && roots[0].creationTimeMs === root.creationTimeMs
  if (rootVerified) {
    ownedRows.push(roots[0])
    visitedRows.add(roots[0])
  } else {
    for (const row of roots) {
      unknown('current_root_identity_unverified', row)
      visitedRows.add(row)
    }
  }
  // A completely absent root and no raw candidates is an ordinary EOF race.
  // It adds no invented unknown identity; the caller retains its known root.
  const queue = [{ pid: root.pid, creationTimeMs: root.creationTimeMs, verified: rootVerified }]
  queuedPIDs.add(root.pid)
  const enqueue = (row, verified) => {
    if (positivePID(row.pid) && !queuedPIDs.has(row.pid)) {
      queuedPIDs.add(row.pid)
      queue.push({ pid: row.pid, creationTimeMs: row.creationTimeMs, verified })
    }
  }
  for (let index = 0; index < queue.length; index++) {
    const parent = queue[index]
    for (const row of byParent.get(parent.pid) ?? []) {
      if (row.pid === root.pid) unknown('root_in_descendant_branch', row)
      if (visitedRows.has(row)) continue
      visitedRows.add(row)
      // 已知原根缺席时，先排除比该根更早的唯一直接子行；不补根或推断退出。
      if (roots.length === 0 && parent.pid === root.pid && parent.creationTimeMs === root.creationTimeMs &&
          positivePID(root.pid) && positivePID(row.pid) && byPID.get(row.pid)?.length === 1 &&
          incarnationOrder(root.creationTimeMs, row.creationTimeMs) === 'predates-parent') {
        excludedEdges.push({ reason: 'predates-root', actual: row, parent: { pid: parent.pid, creationTimeMs: parent.creationTimeMs } })
        continue
      }
      // Unknown parentage propagates through every raw-reachable child.
      // A readable child timestamp cannot repair that missing provenance.
      if (!parent.verified) {
        unknown('parent_incarnation_unverified', row)
        enqueue(row, false)
        continue
      }
      if (!positivePID(row.pid) || byPID.get(row.pid)?.length !== 1) {
        unknown('process_identity_not_unique', row)
        enqueue(row, false)
        continue
      }
      const rootOrder = incarnationOrder(root.creationTimeMs, row.creationTimeMs)
      const parentOrder = incarnationOrder(parent.creationTimeMs, row.creationTimeMs)
      if (rootOrder === 'unknown' || parentOrder === 'unknown') {
        unknown('creation_time_unavailable', row)
        enqueue(row, false)
        continue
      }
      if (rootOrder === 'predates-parent' || parentOrder === 'predates-parent') {
        excludedEdges.push({ reason: rootOrder === 'predates-parent' ? 'predates-root' : 'predates-parent', actual: row,
          parent: { pid: parent.pid, creationTimeMs: parent.creationTimeMs } })
        // This excludes only an impossible PPID edge, never a positive Job
        // membership. Do not expand this branch or erase any known identity.
        continue
      }
      const currentParent = byPID.get(parent.pid)
      if (currentParent?.length !== 1 || currentParent[0].creationTimeMs !== parent.creationTimeMs) {
        unknown('parent_incarnation_changed', row)
        enqueue(row, false)
        continue
      }
      ownedRows.push(row)
      enqueue(row, true)
    }
  }
  return { rootVerified, ownedRows, excludedEdges, unknownRows }
}
