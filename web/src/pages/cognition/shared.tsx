import type { CognitionEvent, CognitionEvents } from '../../api/cognitionTypes';
import { EmptyState } from '../../components/EmptyState';
import { Icon } from '../../components/DesignIcons';
import { formatDate, formatNumber } from './format';

export function EventTable({ events }: { events: CognitionEvent[] }) {
  if (!events.length) return <EmptyState title="当前条件下没有事件" />;
  return <div className="cg-table-scroll"><table className="cg-events"><thead><tr><th>事件 ID</th><th>类型</th><th>主题</th><th>Token 数</th><th>范围</th><th>时间（上海）</th></tr></thead><tbody>{events.map(event => <tr key={event.event_id}><td className="cg-event-id">{event.event_id}</td><td>{event.kind}</td><td>{event.topic || '未标注'}</td><td>{formatNumber(event.tokens)}</td><td>{event.scope || '未标注'}</td><td className="cg-nowrap">{formatDate(event.ts)}</td></tr>)}</tbody></table></div>;
}
export function Pagination({ data, page, onPage }: { data: CognitionEvents; page: number; onPage: (page: number) => void }) {
  const pages = Math.max(1, Math.ceil(data.total / data.page_size));
  return <div className="cg-pagination"><span>共 <b data-testid="events-total" data-value={data.total}>{formatNumber(data.total)}</b> 条 · 第 {data.page} / {pages} 页</span><div><button className="ac-button secondary compact" type="button" disabled={page <= 1} onClick={() => onPage(page - 1)}>上一页</button><button className="ac-button secondary compact" type="button" disabled={page >= pages} onClick={() => onPage(page + 1)}>下一页<Icon name="chevron" size={12} /></button></div></div>;
}
export function FilterSelect({ label, value, values, onChange }: { label: string; value: string; values: string[]; onChange: (value: string) => void }) {
  return <label className="cg-field"><span>{label}</span><select aria-label={label} value={value} onChange={event => onChange(event.target.value)}><option value="">全部{label}</option>{values.filter(Boolean).map(item => <option key={item} value={item}>{item}</option>)}</select></label>;
}
