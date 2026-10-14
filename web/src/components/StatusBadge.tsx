/** V1 状态同时使用图标、文字与颜色显示，便于辨认。 */

interface StatusBadgeProps {
  /** 状态或生命周期值。 */
  value: string;
  /** 已翻译的显示标签。 */
  label: string;
}

const STATUS_ICONS: Record<string, string> = {
  // 资料生命周期。
  active:      '●',
  archived:    '▪',
  withdrawn:   '⊘',
  // 资料导入状态。
  queued:      '◷',
  running:     '▶',
  succeeded:   '✓',
  failed:      '✗',
  cancelled:   '⊘',
  // 机会状态。
  incubating:       '◇',
  ready_for_review: '◷',
  admitted:         '✓',
  deferred:         '⟳',
  rejected:         '✗',
  // 提案状态。
  draft:       '○',
  in_review:   '◷',
  approved:    '✓',
  declined:    '✗',
  superseded:  '⟳',
  // 任务状态。
  pending:     '◷',
  paused:      '‖',
  blocked:     '⊘',
  completed:   '✓',
  // 工作项状态。
  open:            '○',
  claimed:         '◈',
  in_progress:     '▶',
  submitted:       '↑',
  revision_needed: '⟳',
  accepted:        '✓',
  // 会话绑定状态。
  observed:    '👁',
  bound:       '🔗',
  unavailable: '—',
  // 上下文状态。
  unknown: '?',
  current: '●',
  stale:   '◌',
};

type ColorSet = { bg: string; text: string; border: string };

const GREEN: ColorSet   = { bg: 'var(--color-success-subtle)', text: '#2b8a3e', border: '#2b8a3e' };
const BLUE: ColorSet    = { bg: 'var(--color-info-subtle)', text: 'var(--color-info)', border: 'var(--color-info)' };
const YELLOW: ColorSet  = { bg: 'var(--color-warning-subtle)', text: '#e67700', border: '#e67700' };
const RED: ColorSet     = { bg: 'var(--color-error-subtle)', text: 'var(--color-error)', border: 'var(--color-error)' };
const GRAY: ColorSet    = { bg: 'var(--color-bg-muted)', text: 'var(--color-text-muted)', border: 'var(--color-border)' };
const NEUTRAL: ColorSet = { bg: 'var(--color-bg-muted)', text: 'var(--color-text-secondary)', border: 'var(--color-border)' };

const STATUS_COLORS: Record<string, ColorSet> = {
  active: BLUE, succeeded: GREEN, admitted: GREEN, approved: GREEN, accepted: GREEN,
  completed: GREEN, bound: GREEN, current: GREEN,
  running: BLUE, in_progress: BLUE, claimed: BLUE, observed: NEUTRAL,
  incubating: NEUTRAL, draft: NEUTRAL, open: NEUTRAL,
  queued: YELLOW, pending: YELLOW, ready_for_review: YELLOW, in_review: YELLOW,
  deferred: YELLOW, stale: YELLOW, submitted: YELLOW, paused: YELLOW,
  blocked: RED, failed: RED, rejected: RED, declined: RED,
  cancelled: GRAY, archived: GRAY, withdrawn: GRAY,
  superseded: GRAY, revision_needed: YELLOW,
  unavailable: GRAY, unknown: GRAY, ended: GRAY,
};

export function StatusBadge({ value, label }: StatusBadgeProps) {
  const icon = STATUS_ICONS[value] ?? '○';
  const colors = STATUS_COLORS[value] ?? NEUTRAL;

  return (
    <span
      role="status"
      aria-label={`状态: ${label}`}
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: 'var(--space-1)',
        padding: '2px 8px',
        fontSize: 'var(--text-sm)',
        fontWeight: 'var(--weight-medium)',
        borderRadius: 'var(--radius-full)',
        background: colors.bg,
        color: colors.text,
        border: `1px solid ${colors.border}`,
        whiteSpace: 'nowrap',
      }}
    >
      <span aria-hidden="true">{icon}</span>
      {label}
    </span>
  );
}
