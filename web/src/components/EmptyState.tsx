interface EmptyStateProps {
  icon?: string;
  title: string;
  description?: string;
}

/** 数据为空时显示当前状态与明确说明。 */
export function EmptyState({ icon = '◇', title, description }: EmptyStateProps) {
  return (
    <div
      role="status"
      aria-label={title}
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 'var(--space-12) var(--space-6)',
        textAlign: 'center',
        color: 'var(--color-text-muted)',
      }}
    >
      <span style={{ fontSize: '2rem', marginBottom: 'var(--space-3)', opacity: 0.5 }} aria-hidden="true">
        {icon}
      </span>
      <div style={{ fontSize: 'var(--text-md)', fontWeight: 'var(--weight-medium)', color: 'var(--color-text-secondary)' }}>
        {title}
      </div>
      {description && (
        <div style={{ fontSize: 'var(--text-sm)', marginTop: 'var(--space-2)', maxWidth: '320px' }}>
          {description}
        </div>
      )}
    </div>
  );
}
