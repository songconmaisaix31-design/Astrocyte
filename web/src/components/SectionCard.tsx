import type { ReactNode } from 'react';
import styles from './SectionCard.module.css';
import { useActiveTab } from './activeTab';

interface SectionCardProps {
  title: string;
  count?: number;
  actions?: ReactNode;
  children: ReactNode;
  padded?: boolean;
  tabs?: string[];
}

export function SectionCard({ title, count, actions, children, padded, tabs }: SectionCardProps) {
  const active = useActiveTab();
  if (tabs && !tabs.includes(active)) return null;
  return (
    <section className={styles.card}>
      {title && <div className={styles.header}>
        <h2 className={styles.headerTitle}>
          {title}
          {count != null && <span className={styles.headerCount}>({count})</span>}
        </h2>
        {actions && <div>{actions}</div>}
      </div>}
      <div className={padded ? styles.bodyPadded : styles.body}>
        {children}
      </div>
    </section>
  );
}
