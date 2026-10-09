import type { ReactNode } from 'react';
import styles from './CollectionOverview.module.css';

export function CollectionOverview({ title, description, metrics, children }: {
  title: string; description: string;
  metrics: { label: string; value: ReactNode }[];
  children?: ReactNode;
}) {
  return <div className={styles.overview}>
    <div className={styles.heading}><h3>{title}</h3><p>{description}</p></div>
    <dl className={styles.metrics}>{metrics.map(metric => <div key={metric.label}><dd>{metric.value}</dd><dt>{metric.label}</dt></div>)}</dl>
    {children}
  </div>;
}

export function FilterChips({ label, value, options, onChange, disabled = false }: {
  label: string; value: string; options: { value: string; label: string }[];
  onChange: (value: string) => void; disabled?: boolean;
}) {
  return <div className={styles.filter} role="group" aria-label={label}><span>{label}</span><div>{options.map(option => <button
    key={option.value} type="button" disabled={disabled} aria-pressed={value === option.value}
    className={value === option.value ? styles.active : undefined} onClick={() => onChange(option.value)}
  >{option.label}</button>)}</div></div>;
}
