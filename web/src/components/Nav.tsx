import { navigate } from '../router';
import { useFoundation } from '../hooks/useReadApi';
import styles from './Nav.module.css';

interface NavProps {
  currentPath: string;
}

const NAV_ITEMS = [
  { path: '/attention', label: '注意力', icon: '◎', shortcut: '1' },
  { path: '/workspace', label: '工作台', icon: '▦', shortcut: '2' },
  { path: '/swarm',     label: '集群',   icon: '◈', shortcut: '3' },
];

export function Nav({ currentPath }: NavProps) {
  const foundation = useFoundation();
  const caps = foundation.data?.capabilities;

  return (
    <nav className={styles.sidebar} role="navigation" aria-label="主导航">
      <div className={styles.brand}>
        <div className={styles.brandTitle}>Astrocyte</div>
        <div className={styles.brandSub}>
          S0 · 研究工作台
          {foundation.data && <span style={{ marginLeft: 6, opacity: 0.6 }}>({foundation.data.stage})</span>}
        </div>
      </div>

      <ul className={styles.navList} role="list">
        {NAV_ITEMS.map((item) => {
          const active = currentPath.startsWith(item.path);
          return (
            <li key={item.path} className={styles.navItem}>
              <button
                className={`${styles.navLink} ${active ? styles.navLinkActive : ''}`}
                onClick={() => navigate(item.path)}
                aria-current={active ? 'page' : undefined}
                type="button"
              >
                <span className={styles.navIcon} aria-hidden="true">{item.icon}</span>
                <span className={styles.navLabel}>{item.label}</span>
                <kbd className={styles.navShortcut}>{item.shortcut}</kbd>
              </button>
            </li>
          );
        })}
      </ul>

      <div className={styles.footer}>
        <div style={{ marginBottom: 'var(--space-2)', fontWeight: 'var(--weight-medium)' }}>
          系统能力
        </div>
        {caps ? (
          <>
            <CapRow label="导入" value={caps.imports} />
            <CapRow label="审核" value={caps.approvals} />
            <CapRow label="执行" value={caps.execution} />
            <CapRow label="恢复" value={caps.native_resume} />
            <CapRow label="交接" value={caps.handoff} />
          </>
        ) : (
          <div style={{ color: 'var(--color-text-disabled)' }}>
            {foundation.loading ? '加载中…' : foundation.error ? '无法获取' : '—'}
          </div>
        )}
      </div>
    </nav>
  );
}

function CapRow({ label, value }: { label: string; value: boolean }) {
  return (
    <div className={styles.capRow}>
      <span>{label}</span>
      <span className={value ? styles.capTrue : styles.capFalse}>
        {value ? '●' : '○'}
      </span>
    </div>
  );
}
