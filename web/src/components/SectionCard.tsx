import type { ReactNode } from 'react';
import styles from './SectionCard.module.css';

interface SectionCardProps {
  title: string;
  count?: number;
  actions?: ReactNode;
  children: ReactNode;
  padded?: boolean;
}

export function SectionCard({ title, count, actions, children, padded }: SectionCardProps) {
  return (
    <section className={styles.card}>
      <div className={styles.header}>
        <h2 className={styles.headerTitle}>
          {title}
          {count != null && <span className={styles.headerCount}>({count})</span>}
        </h2>
        {actions && <div>{actions}</div>}
      </div>
      <div className={padded ? styles.bodyPadded : styles.body}>
        {children}
      </div>
    </section>
  );
}
