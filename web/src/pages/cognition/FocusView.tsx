import { useState } from 'react';
import { cognitionApi } from '../../api/cognition';
import type { CognitionFocus } from '../../api/cognitionTypes';
import { useReadApi } from '../../hooks/useReadApi';
import { SectionCard } from '../../components/SectionCard';
import { QueryState } from '../../components/QueryState';
import { EmptyState } from '../../components/EmptyState';
import { Icon } from '../../components/DesignIcons';
import { formatDate, formatNumber } from './format';

const colors = ['#82cba0', '#e4ad73', '#a99be0', '#8ab3d6', '#bfd58c', '#d59b93'];
function Trajectory({ series }: { series: CognitionFocus['series'] }) {
  for (const line of series) {
    if (line.points.some(point => !Number.isFinite(Date.parse(point.ts)) || !Number.isFinite(point.attention))) throw new Error('注意力轨迹包含无效时间或数值');
  }
  const lines = series.map(line => ({ ...line, points: [...line.points].sort((a, b) => Date.parse(a.ts) - Date.parse(b.ts)) })).filter(line => line.points.length);
  const points = lines.flatMap(line => line.points);
  if (!points.length) return <EmptyState title="暂无可用的注意力轨迹" />;
  const first = Math.min(...points.map(point => Date.parse(point.ts))), last = Math.max(...points.map(point => Date.parse(point.ts)));
  const low = Math.min(0, ...points.map(point => point.attention)), high = Math.max(low + 1, ...points.map(point => point.attention));
  const x = (ts: string) => 67 + (Date.parse(ts) - first) / Math.max(1, last - first) * 805;
  const y = (attention: number) => 235 - (attention - low) / (high - low) * 205;
  return <><div className="cg-chart-scroll"><svg viewBox="0 0 930 287" role="img" aria-label="多主题注意力历史与实验轨迹"><text x="12" y="16" className="cg-chart-label">注意力 A</text>{[0, 1, 2, 3, 4].map(step => { const value = low + (high - low) * step / 4; return <g key={step}><line x1="67" x2="872" y1={y(value)} y2={y(value)} stroke="var(--ac-line)" /><text x="8" y={y(value) + 4} className="cg-chart-label">{formatNumber(value)}</text></g>; })}{lines.map((line, index) => <g key={line.topic}>{line.points.map((point, pointIndex) => <g key={`${point.ts}-${pointIndex}`}>
      {/* 只连接间隔不超过一天的确证点，保留缺失日期造成的间隔。 */}
      {pointIndex > 0 && Date.parse(point.ts) - Date.parse(line.points[pointIndex - 1].ts) <= 86400000 && <line x1={x(line.points[pointIndex - 1].ts)} y1={y(line.points[pointIndex - 1].attention)} x2={x(point.ts)} y2={y(point.attention)} stroke={colors[index % colors.length]} strokeWidth="1.8" />}
      <circle cx={x(point.ts)} cy={y(point.attention)} r="3.8" fill={colors[index % colors.length]}><title>{`${line.topic} · ${formatDate(point.ts)} · A ${formatNumber(point.attention)}`}</title></circle></g>)}</g>)}<text x="67" y="273" className="cg-chart-label">{formatDate(new Date(first).toISOString())}</text><text x="872" y="273" textAnchor="end" className="cg-chart-label">{formatDate(new Date(last).toISOString())}</text></svg></div><div className="cg-legend">{lines.map((line, index) => <span key={line.topic}><i style={{ background: colors[index % colors.length] }} />{line.topic}<small>{line.points.length} 个确证点</small></span>)}</div></>;
}
export function FocusView({ revision, query, openTopic }: { revision: number; query: string; openTopic: (topic: string) => void }) {
  const state = useReadApi(signal => cognitionApi.focus(signal), { key: String(revision) });
  return <QueryState state={state}>{data => <FocusWorkspace data={data} query={query} openTopic={openTopic} />}</QueryState>;
}

function FocusWorkspace({ data, query, openTopic }: { data: CognitionFocus; query: string; openTopic: (topic: string) => void }) {
  const [hidden, setHidden] = useState<string[]>([]);
  const [selected, setSelected] = useState<string | null>(null);
  const [cutoff, setCutoff] = useState<string | null>(null);
    const topics = data.topics.filter(topic => topic.topic.toLowerCase().includes(query.toLowerCase())).sort((a, b) => b.attention - a.attention);
    const lines = data.series.filter(line => line.topic.toLowerCase().includes(query.toLowerCase()));
    const timestamps = [...new Set(lines.flatMap(line => line.points.map(point => point.ts)))].sort((a, b) => Date.parse(a) - Date.parse(b));
    const pointCount = lines.reduce((count, line) => count + line.points.length, 0);
    const focused = topics.find(topic => topic.topic === selected) ?? topics[0];
    const last = timestamps.at(-1);
    const visible = lines.filter(line => !hidden.includes(line.topic)).map(line => ({ ...line, points: line.points.filter(point => cutoff === null || Date.parse(point.ts) <= Date.parse(cutoff)) }));
    const cutoffIndex = cutoff === null ? timestamps.length - 1 : timestamps.indexOf(cutoff);
    return <><div className="cg-section-heading"><div><h2>焦点轨迹</h2><p>历史与实验期确证的 A 值，沿时间共同呈现；缺失点保留为空。</p></div></div>
      <div className="ac-observe-metrics" aria-label="注意力观察概览">
        <div className="ac-observe-metric"><span>当前焦点</span><strong>{topics[0]?.topic ?? '暂无主题'}</strong><small>{topics[0] ? `当前快照 A ${formatNumber(topics[0].attention)}` : '等待真实快照'}</small></div>
        <div className="ac-observe-metric"><span>关注主题</span><strong>{topics.length}</strong><small>当前筛选范围内的主题</small></div>
        <div className="ac-observe-metric"><span>历史确证点</span><strong>{pointCount}</strong><small>来自已保存的历史与实验记录</small></div>
        <div className="ac-observe-metric"><span>最近轨迹日期</span><strong>{last ? new Date(last).toLocaleDateString('zh-CN', { timeZone: 'Asia/Shanghai', month: '2-digit', day: '2-digit' }) : '未提供'}</strong><small>原记录时间 · 上海时间</small></div>
      </div>
      <div className="ac-observe-grid"><section aria-label="注意力历史画布"><div className="ac-observe-field"><div className="ac-observe-toolbar"><span>ATTENTION FIELD<small>注意力 A 的变化</small></span><span>历史确证记录</span></div><div className="cg-topic-toggles">{lines.map(line => <button key={line.topic} type="button" className={`cg-chip ${hidden.includes(line.topic) ? '' : 'selected'}`} aria-pressed={!hidden.includes(line.topic)} onClick={() => { setSelected(line.topic); setHidden(previous => previous.includes(line.topic) ? previous.filter(topic => topic !== line.topic) : [...previous, line.topic]); }}>{line.topic}</button>)}</div><Trajectory series={visible} /></div>
        <div className="ac-observe-playback"><label htmlFor="ac-focus-time">轨迹时间轴</label><input id="ac-focus-time" type="range" min="0" max={Math.max(0, timestamps.length - 1)} value={Math.max(0, cutoffIndex)} disabled={!timestamps.length} onChange={event => setCutoff(timestamps[Number(event.target.value)])} /><span>{cutoff === null ? '全部记录' : formatDate(cutoff)}</span><button type="button" className="ac-text-button" disabled={cutoff === null} onClick={() => setCutoff(null)}>返回全部</button></div>
      </section><aside className="ac-observe-inspector" aria-label="焦点详情"><span>焦点详情 / TOPIC</span><h3>{focused?.topic ?? '暂无焦点'}</h3><p>当前快照记录的注意力、变化动量与置信度。历史时间轴仅筛选左侧轨迹。</p>{focused && <dl><div><dt>注意力 A</dt><dd>{formatNumber(focused.attention)}</dd></div><div><dt>变化动量 M</dt><dd>{formatNumber(focused.momentum)}</dd></div><div><dt>置信度 C</dt><dd>{formatNumber(focused.confidence)}</dd></div><div><dt>历史确证点</dt><dd>{lines.find(line => line.topic === focused.topic)?.points.length ?? 0}</dd></div></dl>}<button type="button" className="ac-button" disabled={!focused} onClick={() => focused && openTopic(focused.topic)}>追溯关联证据<Icon name="arrow" size={13} /></button></aside></div>
      <SectionCard title="主题焦点" padded><p className="cg-muted cg-caption">按注意力 A 从高到低 · 点击主题回看证据</p>{topics.length ? <div className="cg-topic-list">{topics.map((topic, index) => <button className="cg-topic-row" type="button" key={topic.topic} onClick={() => { setSelected(topic.topic); openTopic(topic.topic); }}><span className="cg-rank">{String(index + 1).padStart(2, '0')}</span><strong>{topic.topic}</strong><span>A <b data-testid={`focus-a-${topic.topic}`} data-value={topic.attention}>{formatNumber(topic.attention)}</b></span><span>M <b data-testid={`focus-m-${topic.topic}`} data-value={topic.momentum}>{formatNumber(topic.momentum)}</b></span><span>C <b data-testid={`focus-c-${topic.topic}`} data-value={topic.confidence}>{formatNumber(topic.confidence)}</b></span><Icon name="chevron" size={15} /></button>)}</div> : <EmptyState title="没有可显示的主题" />}</SectionCard></>;
}
