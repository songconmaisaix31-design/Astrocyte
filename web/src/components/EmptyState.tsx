interface EmptyStateProps {
  icon?: string;
  title: string;
  description?: string;
}

/** Honest no-data empty state. No fake science, no misleading content. */
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
