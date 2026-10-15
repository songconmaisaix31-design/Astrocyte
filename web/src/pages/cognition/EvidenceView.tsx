import { useState } from 'react';
import { cognitionApi } from '../../api/cognition';
import { useReadApi } from '../../hooks/useReadApi';
import { QueryState } from '../../components/QueryState';
import { SectionCard } from '../../components/SectionCard';
import { DetailPanel } from '../../components/DetailPanel';
import { EmptyState } from '../../components/EmptyState';
import { EventTable, FilterSelect, Pagination } from './shared';
import { formatNumber } from './format';

export function EvidenceView({ revision }: { revision: number }) {
  const filters = useReadApi(signal => cognitionApi.filters(signal), { key: String(revision) });
  const [values, setValues] = useState({ topic: '', kind: '', source: '', start: '', end: '' });
  const [page, setPage] = useState(1), [size, setSize] = useState(20);
  const invalid = Boolean(values.start && values.end && values.start > values.end);
  const state = useReadApi(signal => cognitionApi.events({ ...Object.fromEntries(Object.entries(values).filter(([, value]) => value)), page, page_size: size }, signal), { enabled: !invalid, key: `${revision}:${JSON.stringify(values)}:${page}:${size}` });
  const change = (key: keyof typeof values, value: string) => { setPage(1); setValues(previous => ({ ...previous, [key]: value })); };
  return <><div className="cg-section-heading"><div><h2>证据浏览</h2><p>按主题、类型、日期与来源，回看完整的事件记录。</p></div></div><SectionCard title="事件证据" padded><div className="cg-filters"><FilterSelect label="主题" value={values.topic} values={filters.data?.topics || []} onChange={value => change('topic', value)} /><FilterSelect label="类型" value={values.kind} values={filters.data?.kinds || []} onChange={value => change('kind', value)} /><FilterSelect label="来源" value={values.source} values={filters.data?.sources || []} onChange={value => change('source', value)} /><label className="cg-field"><span>开始日期</span><input type="date" aria-label="开始日期" value={values.start} max={values.end || undefined} onChange={event => change('start', event.target.value)} /></label><label className="cg-field"><span>结束日期</span><input type="date" aria-label="结束日期" value={values.end} min={values.start || undefined} onChange={event => change('end', event.target.value)} /></label><label className="cg-field"><span>每页条数</span><select aria-label="每页条数" value={size} onChange={event => { setPage(1); setSize(Number(event.target.value)); }}>{[20, 50, 100].map(count => <option value={count} key={count}>{count} 条</option>)}</select></label></div><div className="cg-filter-footer"><span className="cg-muted">日期包含首尾两日，按上海自然日计算。</span><button className="ac-button secondary compact" type="button" onClick={() => { setValues({ topic: '', kind: '', source: '', start: '', end: '' }); setPage(1); }}>清除筛选</button></div>{filters.error && <QueryState state={filters}>{() => null}</QueryState>}{invalid ? <p role="alert">开始日期不能晚于结束日期</p> : <QueryState state={state}>{data => <><EventTable events={data.events} /><Pagination data={data} page={page} onPage={setPage} /></>}</QueryState>}</SectionCard></>;
}
export function EvidenceDrawer({ topic, revision, close }: { topic: string; revision: number; close: () => void }) {
  const [page, setPage] = useState(1);
  const state = useReadApi(signal => cognitionApi.evidence(topic, page, 20, signal), { key: `${revision}:${topic}:${page}` });
  return <DetailPanel title={`${topic} · 主题证据`} onClose={close}><div className="cg-drawer"><QueryState state={state}>{data => <><h3 className="cg-subheading">按日与类型汇总</h3>{data.aggregates.length ? <div className="cg-table-scroll cg-aggregates"><table><thead><tr><th>日期</th><th>类型</th><th>事件数</th></tr></thead><tbody>{data.aggregates.map(item => <tr key={`${item.day}-${item.kind}`}><td>{item.day}</td><td>{item.kind}</td><td data-testid={`aggregate-${item.day}-${item.kind}`} data-value={item.count}>{formatNumber(item.count)}</td></tr>)}</tbody></table></div> : <EmptyState title="暂无主题统计" />}<h3 className="cg-subheading">完整事件证据</h3><EventTable events={data.events} /><Pagination data={data} page={page} onPage={setPage} /></>}</QueryState></div></DetailPanel>;
}
