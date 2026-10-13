import React, { useEffect, useRef, useState } from 'react';
import { createRoot } from 'react-dom/client';
import { useApi, query, type Health, type Today, type Focus, type Agents, type Events, type Filters, type TopicEvidence, type Event } from './api';
import './style.css';

const nav = [{ path: '/', name: '今日', icon: '◉' }, { path: '/focus', name: '焦点', icon: '◎' }, { path: '/agents', name: '来源', icon: '◈' }, { path: '/evidence', name: '证据', icon: '▤' }];
const number = (value: number | null | undefined) => value == null ? '暂无' : new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 4 }).format(value);
const date = (value: string | null | undefined) => !value ? '暂无' : new Date(value).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai', hour12: false });
const colors = ['#387765', '#ad7934', '#627fb1', '#a16683', '#7b8442', '#9b604a'];

function State({ loading, error }: { loading: boolean; error?: string }) {
  if (loading) return <div className="state" role="status">正在读取快照…</div>;
  if (error) return <div className="state error" role="alert"><strong>暂时无法读取数据</strong><p>{error}。请确认本机服务可用，再重新读取。</p></div>;
  return null;
}
function Empty({ children = '当前条件下没有记录' }: { children?: React.ReactNode }) { return <div className="empty">{children}</div>; }
function EventTable({ events }: { events: Event[] }) {
  if (!events.length) return <Empty />;
  return <div className="table-scroll"><table className="events"><thead><tr><th>事件 ID</th><th>类型</th><th>主题</th><th>Token 数</th><th>范围</th><th>时间（上海）</th></tr></thead><tbody>{events.map(event => <tr key={event.event_id}><td className="event-id">{event.event_id}</td><td>{event.kind}</td><td>{event.topic || '未标注'}</td><td>{number(event.tokens)}</td><td>{event.scope || '未标注'}</td><td className="nowrap">{date(event.ts)}</td></tr>)}</tbody></table></div>;
}
function Pager({ data, page, setPage }: { data: Events; page: number; setPage: (page: number) => void }) {
  const pages = Math.max(1, Math.ceil(data.total / data.page_size));
  return <div className="pager"><span>共 <b data-testid="events-total">{number(data.total)}</b> 条 · 第 {data.page} / {pages} 页</span><div><button disabled={page <= 1} onClick={() => setPage(page - 1)}>上一页</button><button disabled={page >= pages} onClick={() => setPage(page + 1)}>下一页</button></div></div>;
}
function Select({ label, value, values, onChange }: { label: string; value: string; values: string[]; onChange: (value: string) => void }) {
  return <label className="field"><span>{label}</span><select aria-label={label} value={value} onChange={event => onChange(event.target.value)}><option value="">全部{label}</option>{values.map(item => <option key={item} value={item}>{item || '未标注'}</option>)}</select></label>;
}
function MiniChart({ counts }: { counts: Today['daily_counts'] }) {
  const sorted = [...counts].sort((a, b) => a.day.localeCompare(b.day));
  const max = Math.max(1, ...sorted.map(point => point.count));
  return <div className="mini-chart" aria-label="近七日事件计数">{sorted.map(point => <div className="mini-column" key={point.day}><span className="mini-value" data-testid={`daily-${point.day}`}>{number(point.count)}</span><div className="bar-track"><div className="bar" style={{ height: `${point.count / max * 100}%` }} /></div><span>{point.day.slice(5)}</span></div>)}</div>;
}
function TodayView({ revision, openTopic }: { revision: number; openTopic: (topic: string) => void }) {
  const [source, setSource] = useState('');
  const filters = useApi<Filters>('/api/filters', revision);
  const state = useApi<Today>(query('/api/today', { source }), revision);
  const data = state.data;
  return <><div className="page-heading"><div><p className="eyebrow">把注意力留给重要的事</p><h1>今日概览</h1><p>从焦点、活动与沉淀，理解今天的认知节奏。</p></div><span className="day-tag">{data?.day || '上海自然日'}</span></div><State {...state} />{data && <>
    <div className="section-title"><h2>当前焦点</h2><span>A 注意力 · M 动量 · C 置信度</span></div>
    {data.focus.length ? <div className="focus-grid">{[...data.focus].sort((a, b) => b.attention - a.attention).map(topic => <button className="focus-card" key={topic.topic} onClick={() => openTopic(topic.topic)}><div className="card-top"><span className="dot" /><strong>{topic.topic}</strong><span>↗</span></div><div className="attention"><span data-testid={`attention-${topic.topic}`}>{number(topic.attention)}</span><small>A · 注意力</small></div><div className="delta">{topic.attention_delta === null ? '日环比暂无' : `较昨日 ${topic.attention_delta > 0 ? '+' : ''}${number(topic.attention_delta)}`}</div><div className="metrics"><span>M 动量 <b data-testid={`momentum-${topic.topic}`}>{number(topic.momentum)}</b></span><span>C 置信度 <b data-testid={`confidence-${topic.topic}`}>{number(topic.confidence)}</b></span></div></button>)}</div> : <Empty>今日尚无焦点记录</Empty>}
    <div className="overview-grid"><section className="panel"><div className="section-title"><h2>昨夜沉淀</h2><span>{data.sleep.day}</span></div><span className="sleep-badge" data-testid="sleep-status">{({ completed: '已完成', pending: '待完成', not_run: '尚未运行', missing: '暂无记录', not_available: '暂无记录', no_data: '暂无记录', not_produced: '尚未产出', failed: '未完成', unknown: '状态未知' } as Record<string, string>)[data.sleep.status] || `状态：${data.sleep.status}`}</span><p className="muted">{data.sleep.completed_at ? `完成于 ${date(data.sleep.completed_at)}` : '快照中尚无完成时间'}</p><p className="small muted">沉淀状态来自实际记录。</p></section><section className="panel"><div className="section-title"><h2>近七日活动</h2><span>事件数</span></div>{data.daily_counts.length ? <MiniChart counts={data.daily_counts} /> : <Empty />}</section></div>
    <section className="panel"><div className="section-title"><div><h2>今日活动</h2><p className="muted small">按时间倒序 · 六字段证据</p></div><Select label="来源" value={source} values={filters.data?.sources || data.sources.map(item => item.source)} onChange={setSource} /></div><State {...filters} /><EventTable events={[...data.activity].sort((a, b) => b.ts.localeCompare(a.ts) || a.event_id.localeCompare(b.event_id))} /></section>
  </>}</>;
}
function Trajectory({ series }: { series: Focus['series'] }) {
  const actual = series.filter(line => line.points.length > 0);
  if (!actual.length) return <Empty>暂无可用的注意力轨迹</Empty>;
  const all = actual.flatMap(line => line.points).filter(point => Number.isFinite(Date.parse(point.ts)) && Number.isFinite(point.attention));
  if (!all.length) return <Empty>暂无有效轨迹点</Empty>;
  const first = Math.min(...all.map(point => Date.parse(point.ts))), last = Math.max(...all.map(point => Date.parse(point.ts)));
  const low = Math.min(0, ...all.map(point => point.attention)), high = Math.max(low + 1, ...all.map(point => point.attention));
  const x = (ts: string) => 64 + (Date.parse(ts) - first) / Math.max(1, last - first) * 860;
  const y = (value: number) => 244 - (value - low) / (high - low) * 216;
  return <><div className="chart-scroll"><svg viewBox="0 0 980 295" role="img" aria-label="多主题注意力历史与实验轨迹"><text x="12" y="17" className="chart-text">注意力 A</text>{[0, 1, 2, 3, 4].map(step => { const value = low + (high - low) * step / 4; return <g key={step}><line x1="64" x2="924" y1={y(value)} y2={y(value)} stroke="#e7ebe6" /><text x="8" y={y(value) + 4} className="chart-text">{number(value)}</text></g>; })}{actual.map((line, index) => { const points = [...line.points].filter(point => Number.isFinite(Date.parse(point.ts)) && Number.isFinite(point.attention)).sort((a, b) => a.ts.localeCompare(b.ts)); return <g key={line.topic}>{/* Points remain discrete: missing historical days are never interpolated. */}{points.map((point, pointIndex) => <circle key={`${point.ts}-${pointIndex}`} cx={x(point.ts)} cy={y(point.attention)} r="4" fill={colors[index % colors.length]}><title>{`${line.topic} · ${date(point.ts)} · A ${number(point.attention)}`}</title></circle>)}</g>; })}<text x="64" y="281" className="chart-text">{date(new Date(first).toISOString())}</text><text x="924" y="281" textAnchor="end" className="chart-text">{date(new Date(last).toISOString())}</text></svg></div><div className="legend">{actual.map((line, index) => <span key={line.topic}><i style={{ background: colors[index % colors.length] }} />{line.topic}<small>{line.points.length} 个实测点</small></span>)}</div></>;
}
function FocusView({ revision, openTopic }: { revision: number; openTopic: (topic: string) => void }) {
  const state = useApi<Focus>('/api/focus', revision);
  const [selected, setSelected] = useState<string[]>([]);
  const series = state.data?.series || [];
  return <><div className="page-heading"><div><p className="eyebrow">注意力如何生长</p><h1>焦点轨迹</h1><p>历史与实验期的真实观测点，沿时间共同呈现。</p></div></div><State {...state} />{state.data && <><section className="panel"><div className="section-title"><h2>注意力 A 的变化</h2><span>缺失点保留为空</span></div><div className="topic-toggles">{series.map(line => <button className={selected.length === 0 || selected.includes(line.topic) ? 'chip selected' : 'chip'} aria-pressed={selected.length === 0 || selected.includes(line.topic)} key={line.topic} onClick={() => setSelected(previous => { const current = previous.length === 0 ? series.map(item => item.topic) : previous; const next = current.includes(line.topic) ? current.filter(topic => topic !== line.topic) : [...current, line.topic]; return next.length ? next : []; })}>{line.topic}</button>)}</div><Trajectory series={selected.length ? series.filter(line => selected.includes(line.topic)) : series} /></section><section className="panel"><div className="section-title"><h2>主题焦点</h2><span>按注意力 A 从高到低</span></div>{state.data.topics.length ? <div className="topic-list">{[...state.data.topics].sort((a, b) => b.attention - a.attention).map((topic, index) => <button className="topic-row" key={topic.topic} onClick={() => openTopic(topic.topic)}><span className="rank">{String(index + 1).padStart(2, '0')}</span><strong>{topic.topic}</strong><span>A <b data-testid={`focus-attention-${topic.topic}`}>{number(topic.attention)}</b></span><span>M <b>{number(topic.momentum)}</b></span><span>C <b>{number(topic.confidence)}</b></span><span className="evidence-link">查看证据 →</span></button>)}</div> : <Empty />}</section></>}</>;
}
function AgentsView({ revision }: { revision: number }) {
  const state = useApi<Agents>('/api/agents', revision);
  return <><div className="page-heading"><div><p className="eyebrow">不同来源，共同构成工作</p><h1>来源概览</h1><p>了解各来源的活动频率，以及最近留下的证据。</p></div></div><State {...state} />{state.data && (state.data.groups.length ? <div className="agent-list">{state.data.groups.map(group => <section className="panel" key={group.source}><div className="agent-heading"><div className="source-title"><span className="source-icon">◈</span><div><h2>{group.source || '未标注来源'}</h2><p className="muted small">来源分组</p></div></div><div className="source-metrics">{(['today', 'days7', 'days30'] as const).map((key, index) => <div key={key}><span>{['今日', '近 7 日', '近 30 日'][index]}</span><strong data-testid={`agent-${group.source}-${key}`}>{number(group[key])}</strong></div>)}</div></div><h3>最近事件</h3><EventTable events={group.recent} /></section>)}</div> : <Empty>尚无来源活动记录</Empty>)}</>;
}
function EvidenceView({ revision }: { revision: number }) {
  const filters = useApi<Filters>('/api/filters', revision);
  const [values, setValues] = useState({ topic: '', kind: '', source: '', start: '', end: '' });
  const [page, setPage] = useState(1), [size, setSize] = useState(20);
  const invalid = Boolean(values.start && values.end && values.start > values.end);
  const state = useApi<Events>(query('/api/events', { ...values, page, page_size: size }), revision);
  const change = (key: keyof typeof values, value: string) => { setPage(1); setValues(previous => ({ ...previous, [key]: value })); };
  return <><div className="page-heading"><div><p className="eyebrow">每个判断，都有迹可循</p><h1>证据浏览</h1><p>按主题、类型、日期与来源，回看完整的事件记录。</p></div></div><section className="panel"><div className="filters"><Select label="主题" value={values.topic} values={filters.data?.topics || []} onChange={value => change('topic', value)} /><Select label="类型" value={values.kind} values={filters.data?.kinds || []} onChange={value => change('kind', value)} /><Select label="来源" value={values.source} values={filters.data?.sources || []} onChange={value => change('source', value)} /><label className="field"><span>开始日期</span><input type="date" aria-label="开始日期" value={values.start} max={values.end || undefined} onChange={event => change('start', event.target.value)} /></label><label className="field"><span>结束日期</span><input type="date" aria-label="结束日期" value={values.end} min={values.start || undefined} onChange={event => change('end', event.target.value)} /></label><label className="field"><span>每页条数</span><select aria-label="每页条数" value={size} onChange={event => { setSize(Number(event.target.value)); setPage(1); }}>{[20, 50, 100].map(value => <option key={value} value={value}>{value} 条</option>)}</select></label></div><div className="filter-footer"><span className="muted small">日期范围包含首尾两日，按上海自然日计算。</span><button onClick={() => { setValues({ topic: '', kind: '', source: '', start: '', end: '' }); setPage(1); }}>清除筛选</button></div><State {...filters} />{invalid ? <div className="state error" role="alert">开始日期不能晚于结束日期</div> : <><State {...state} />{state.data && <><EventTable events={state.data.events} /><Pager data={state.data} page={page} setPage={setPage} /></>}</>}</section></>;
}
function EvidenceDrawer({ topic, revision, close }: { topic: string; revision: number; close: () => void }) {
  const [page, setPage] = useState(1);
  const state = useApi<TopicEvidence>(query('/api/topics/evidence', { topic, page, page_size: 20 }), revision);
  const ref = useRef<HTMLDivElement>(null);
  useEffect(() => {
    const previous = document.activeElement as HTMLElement | null;
    const overflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    ref.current?.focus();
    function keyboard(event: KeyboardEvent) {
      if (event.key === 'Escape') close();
      if (event.key === 'Tab') {
        const elements = [...(ref.current?.querySelectorAll<HTMLElement>('button:not([disabled]), [tabindex="0"]') || [])];
        const first = elements[0], last = elements[elements.length - 1];
        if (event.shiftKey && (document.activeElement === first || document.activeElement === ref.current)) { event.preventDefault(); last?.focus(); }
        else if (!event.shiftKey && document.activeElement === last) { event.preventDefault(); first?.focus(); }
      }
    }
    document.addEventListener('keydown', keyboard);
    return () => { document.body.style.overflow = overflow; document.removeEventListener('keydown', keyboard); previous?.focus(); };
  }, [close]);
  return <div className="drawer-overlay" onClick={event => { if (event.target === event.currentTarget) close(); }}><div className="drawer" role="dialog" aria-modal="true" aria-labelledby="drawer-title" ref={ref} tabIndex={-1}><div className="drawer-header"><div><p className="eyebrow">主题证据</p><h2 id="drawer-title">{topic}</h2></div><button className="close" aria-label="关闭证据" onClick={close}>×</button></div><State {...state} />{state.data && <><h3>按日与类型汇总</h3>{state.data.aggregates.length ? <div className="table-scroll aggregates"><table><thead><tr><th>日期</th><th>类型</th><th>事件数</th></tr></thead><tbody>{state.data.aggregates.map(item => <tr key={`${item.day}-${item.kind}`}><td>{item.day}</td><td>{item.kind}</td><td data-testid={`aggregate-${item.day}-${item.kind}`}>{number(item.count)}</td></tr>)}</tbody></table></div> : <Empty />}<h3>完整事件证据</h3><EventTable events={state.data.events} /><Pager data={state.data} page={page} setPage={setPage} /></>}</div></div>;
}
function App() {
  const [path, setPath] = useState(window.location.pathname);
  const [revision, setRevision] = useState(0), [topic, setTopic] = useState<string | null>(null);
  const health = useApi<Health>('/api/health', revision);
  useEffect(() => { const update = () => { setPath(window.location.pathname); setTopic(null); }; window.addEventListener('popstate', update); return () => window.removeEventListener('popstate', update); }, []);
  useEffect(() => { document.title = `${nav.find(item => item.path === path)?.name || '页面未找到'} · Astrocyte 认知工作台`; }, [path]);
  const close = React.useCallback(() => setTopic(null), []);
  return <div className="app-shell"><aside className="sidebar"><a className="brand" href="/" onClick={event => { event.preventDefault(); history.pushState({}, '', '/'); setPath('/'); setTopic(null); }}><span className="brand-mark">✳</span><span>Astrocyte<small>认知工作台</small></span></a><p className="nav-label">我的工作空间</p><nav aria-label="主导航">{nav.map(item => <a href={item.path} key={item.path} aria-current={path === item.path ? 'page' : undefined} className={path === item.path ? 'nav-item active' : 'nav-item'} onClick={event => { if (event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return; event.preventDefault(); history.pushState({}, '', item.path); setPath(item.path); setTopic(null); window.scrollTo(0, 0); }}><span>{item.icon}</span>{item.name}</a>)}</nav><div className="sidebar-note"><span className="dot" />只读工作台<p>观察焦点，回看证据。<br />数据来自本机快照。</p><small>版本 0.1</small></div></aside><div className="main-shell"><header className="topbar"><div className="breadcrumb">我的工作空间 <span>/</span> {nav.find(item => item.path === path)?.name || '页面未找到'}</div><div className="snapshot"><span className={`status-dot ${health.error ? 'offline' : ''}`} /><span>{health.loading ? '读取快照时间…' : health.error ? '本机服务未连接' : `快照时间：${date(health.data?.snapshot_at)}`}<small>上海时间 · 本机快照</small></span><button aria-label="重新读取快照" onClick={() => setRevision(value => value + 1)}>↻ <span>重新读取</span></button></div></header><main>{path === '/' ? <TodayView revision={revision} openTopic={setTopic} /> : path === '/focus' ? <FocusView revision={revision} openTopic={setTopic} /> : path === '/agents' ? <AgentsView revision={revision} /> : path === '/evidence' ? <EvidenceView revision={revision} /> : <Empty>页面未找到，请通过左侧导航返回工作台。</Empty>}</main><footer>Astrocyte · 本机只读认知工作台 <span>快照刷新由本机备份程序完成</span></footer></div>{topic !== null && <EvidenceDrawer key={topic} topic={topic} revision={revision} close={close} />}</div>;
}

createRoot(document.getElementById('root')!).render(<App />);
