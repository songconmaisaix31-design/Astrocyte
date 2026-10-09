/**
 * Shared formatting utilities with pure functions (testable).
 * Updated for V1 schema types.
 */

/** Format ISO timestamp to locale display. Returns '—' for null/invalid. */
export function formatDate(iso: string | null | undefined): string {
  if (!iso) return '—';
  const d = new Date(iso);
  if (isNaN(d.getTime())) return '—';
  return d.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  });
}

/** Format ISO timestamp with time. Returns '—' for null/invalid. */
export function formatDateTime(iso: string | null | undefined): string {
  if (!iso) return '—';
  const d = new Date(iso);
  if (isNaN(d.getTime())) return '—';
  return d.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
}

/** Format byte count to human-readable size. Returns '—' for null. */
export function formatBytes(bytes: number | null | undefined): string {
  if (bytes == null) return '—';
  if (bytes === 0) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB'];
  const i = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  const val = bytes / Math.pow(1024, i);
  return `${val.toFixed(i === 0 ? 0 : 1)} ${units[i]}`;
}

/** Format a dimension score value (0-1 or null) to percentage. Returns '—' for null. */
export function formatDimScore(value: number | null | undefined): string {
  if (value == null) return '—';
  return `${Math.round(value * 100)}%`;
}

/** Format token usage with comma separators. Returns '—' for null. */
export function formatTokens(tokens: number | null | undefined): string {
  if (tokens == null) return '—';
  return tokens.toLocaleString('zh-CN');
}

/** Format budget display. */
export function formatBudget(budget: { calls: number | null; tokens: number | null; money: number | null; currency: string | null } | null | undefined): string {
  if (!budget) return '—';
  const parts: string[] = [];
  if (budget.calls != null) parts.push(`${budget.calls} 次调用`);
  if (budget.tokens != null) parts.push(`${formatTokens(budget.tokens)} tokens`);
  if (budget.money != null) parts.push(`${budget.money} ${budget.currency ?? ''}`.trim());
  return parts.length > 0 ? parts.join(' / ') : '—';
}

// ── Status labels for V1 enums ──

const LIFECYCLE_LABELS: Record<string, string> = {
  active:    '活跃',
  archived:  '已归档',
  withdrawn: '已撤回',
};

const IMPORT_STATUS_LABELS: Record<string, string> = {
  queued:    '排队中',
  running:   '导入中',
  succeeded: '已完成',
  failed:    '失败',
  cancelled: '已取消',
};

const OPPORTUNITY_STATE_LABELS: Record<string, string> = {
  incubating:       '孵化中',
  ready_for_review: '待审核',
  admitted:         '已采纳',
  deferred:         '已延期',
  rejected:         '已拒绝',
  withdrawn:        '已撤回',
};

const PROPOSAL_STATUS_LABELS: Record<string, string> = {
  draft:       '草稿',
  in_review:   '审核中',
  approved:    '已通过',
  declined:    '已拒绝',
  superseded:  '已替代',
};

const MISSION_STATUS_LABELS: Record<string, string> = {
  pending:   '等待中',
  running:   '执行中',
  paused:    '已暂停',
  blocked:   '已阻塞',
  completed: '已完成',
  cancelled: '已取消',
  failed:    '失败',
};

const WORKITEM_STATUS_LABELS: Record<string, string> = {
  open:            '开放',
  claimed:         '已认领',
  in_progress:     '进行中',
  submitted:       '已提交',
  revision_needed: '需修改',
  blocked:         '已阻塞',
  accepted:        '已通过',
  cancelled:       '已取消',
};

const BINDING_STATUS_LABELS: Record<string, string> = {
  observed:     '已观察',
  bound:        '已绑定',
  unavailable:  '不可用',
};

const CONTEXT_STATE_LABELS: Record<string, string> = {
  unknown: '未知',
  current: '当前',
  stale:   '过期',
};

const MATERIAL_KIND_LABELS: Record<string, string> = {
  paper: '论文',
  video: '视频',
  text:  '文本',
  file:  '文件',
};

/** Get lifecycle label. */
export function lifecycleLabel(lifecycle: string): string {
  return LIFECYCLE_LABELS[lifecycle] ?? lifecycle;
}

/** Get import status label. */
export function importStatusLabel(status: string): string {
  return IMPORT_STATUS_LABELS[status] ?? status;
}

/** Get opportunity state label. */
export function opportunityStateLabel(state: string): string {
  return OPPORTUNITY_STATE_LABELS[state] ?? state;
}

/** Get proposal status label. */
export function proposalStatusLabel(status: string): string {
  return PROPOSAL_STATUS_LABELS[status] ?? status;
}

/** Get mission status label. */
export function missionStatusLabel(status: string): string {
  return MISSION_STATUS_LABELS[status] ?? status;
}

/** Get work item status label. */
export function workItemStatusLabel(status: string): string {
  return WORKITEM_STATUS_LABELS[status] ?? status;
}

/** Get binding status label. */
export function bindingStatusLabel(status: string): string {
  return BINDING_STATUS_LABELS[status] ?? status;
}

/** Get context state label. */
export function contextStateLabel(state: string): string {
  return CONTEXT_STATE_LABELS[state] ?? state;
}

/** Get material kind label. */
export function materialKindLabel(kind: string): string {
  return MATERIAL_KIND_LABELS[kind] ?? kind;
}

/**
 * Three-state capability label.
 * null = unprobed / unknown; false = verified unsupported; true = available.
 */
export function capabilityLabel(value: boolean | null): string {
  if (value === true) return '● 可用';
  if (value === false) return '○ 不可用';
  return '◇ 未知';
}
