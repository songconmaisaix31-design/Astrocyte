import { useState } from 'react';
import { useReadApi } from '../../hooks/useReadApi';
import { cognitionApi } from '../../api/cognition';
import { QueryState } from '../../components/QueryState';
import { SectionCard } from '../../components/SectionCard';
import { EmptyState } from '../../components/EmptyState';
import { Icon } from '../../components/DesignIcons';
import { EventTable, FilterSelect } from './shared';
import { formatDate, formatNumber } from './format';
import { RelaySessionEvents } from './relaySessionEvents';

const sleepLabels: Record<string, string> = { completed: '已完成', running: '运行中', not_produced: '尚未产出' };
export function TodayView({ revision, query, openTopic }: { revision: number; query: string; openTopic: (topic: string) => void }) {
  const [source, setSource] = useState('');
  const state = useReadApi(signal => cognitionApi.today(source, signal), { key: `${revision}:${source}` });
  const filters = useReadApi(signal => cognitionApi.filters(signal), { key: String(revision) });
  return <><QueryState state={state}>{data => {
    const focus = [...data.focus].filter(topic => topic.topic.toLowerCase().includes(query.toLowerCase())).sort((a, b) => b.attention - a.attention);
    const days = [...data.daily_counts].sort((a, b) => a.day.localeCompare(b.day));
    const max = Math.max(1, ...days.map(day => day.count));
    return <><div className="cg-section-heading"><div><h2>今日概览</h2><p>从焦点、活动与沉淀，理解今天的认知节奏。</p></div><span className="cg-day">{data.day}</span></div><div className="cg-section-heading"><h3>当前焦点</h3><span>A 注意力 · M 动量 · C 置信度</span></div>{focus.length ? <div className="cg-focus-grid">{focus.map(topic => <button className="cg-focus-card" type="button" key={topic.topic} onClick={() => openTopic(topic.topic)}><div className="cg-card-heading"><Icon name="spark" size={17} /><strong>{topic.topic}</strong><Icon name="arrow" size={16} /></div><div className="cg-attention"><b data-testid={`today-a-${topic.topic}`} data-value={topic.attention}>{formatNumber(topic.attention)}</b><small>A · 注意力</small></div><p className="cg-delta" data-testid={`today-delta-${topic.topic}`}>{topic.attention_delta == null ? '日环比暂无 · 缺少昨日确证值' : `较昨日 ${topic.attention_delta > 0 ? '+' : ''}${formatNumber(topic.attention_delta)}`}</p><div className="cg-card-metrics"><span>M 动量 <b data-testid={`today-m-${topic.topic}`} data-value={topic.momentum}>{formatNumber(topic.momentum)}</b></span><span>C 置信度 <b data-testid={`today-c-${topic.topic}`} data-value={topic.confidence}>{formatNumber(topic.confidence)}</b></span></div></button>)}</div> : <EmptyState title={query ? '没有匹配的焦点' : '今日尚无焦点记录'} />}
      <div className="cg-overview-grid"><SectionCard title="昨夜沉淀" padded><span className="cg-day">{data.sleep.day}</span><p className="cg-sleep" data-testid="sleep-status">{sleepLabels[data.sleep.status] || `状态：${data.sleep.status}`}</p><p className="cg-muted">{data.sleep.completed_at ? `完成于 ${formatDate(data.sleep.completed_at)}` : '快照中尚无完成时间'}</p></SectionCard><SectionCard title="近七日活动" padded>{days.length ? <div className="cg-mini-chart" aria-label="近七日事件计数">{days.map(day => <div className="cg-mini-column" key={day.day}><b data-testid={`daily-${day.day}`} data-value={day.count}>{formatNumber(day.count)}</b><div className="cg-bar-track"><div className="cg-bar" style={{ height: `${day.count / max * 100}%` }} /></div><span>{day.day.slice(5)}</span></div>)}</div> : <EmptyState title="暂无七日计数" />}</SectionCard></div>
      <SectionCard title="今日活动" padded actions={<FilterSelect label="来源" value={source} values={filters.data?.sources || data.sources.map(item => item.source)} onChange={setSource} />}><p className="cg-muted cg-caption">按时间倒序 · 按来源回看当天记录</p>{filters.error && <QueryState state={filters}>{() => null}</QueryState>}<EventTable events={[...data.activity].filter(event => [event.topic, event.kind, event.event_id].some(value => value.toLowerCase().includes(query.toLowerCase()))).sort((a, b) => Date.parse(b.ts) - Date.parse(a.ts) || a.event_id.localeCompare(b.event_id))} /></SectionCard>
    </>;
  }}</QueryState><RelaySessionEvents revision={revision} /></>;
}
