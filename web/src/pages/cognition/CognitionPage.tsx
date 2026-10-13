import { useCallback, useState } from 'react';
import { navigate } from '../../router';
import { useReadApi } from '../../hooks/useReadApi';
import { QueryState } from '../../components/QueryState';
import { Icon } from '../../components/DesignIcons';
import { EmptyState } from '../../components/EmptyState';
import { cognitionApi } from '../../api/cognition';
import { TodayView } from './TodayView';
import { FocusView } from './FocusView';
import { AgentsView } from './AgentsView';
import { EvidenceView, EvidenceDrawer } from './EvidenceView';
import { formatDate } from './format';
import '../../styles/cognition.css';

const tabs = [{ id: 'today', name: '今日' }, { id: 'focus', name: '焦点' }, { id: 'agents', name: '来源' }, { id: 'evidence', name: '证据' }];
export function CognitionPage({ path, query }: { path: string; query: string }) {
  const active = path === '/cognition' || path === '/cognition/' ? 'today' : path.split('/')[2];
  const [revision, setRevision] = useState(0), [topic, setTopic] = useState<string | null>(null);
  const health = useReadApi(signal => cognitionApi.health(signal), { key: String(revision) });
  const close = useCallback(() => setTopic(null), []);
  const select = (id: string) => { setTopic(null); navigate(`/cognition/${id}`); };
  return <div className="cg-page"><div className="ac-breadcrumb"><span>我的工作台</span><Icon name="chevron" size={12} /><span>认知</span><span className="ac-mode-label">本机只读</span></div><div className="ac-page-header"><div><h1>认知工作台</h1><p>观察注意力的变化，回看每一步留下的证据。</p></div><button type="button" className="ac-button secondary compact" onClick={() => setRevision(value => value + 1)}>↻ 重新读取</button></div><div className="cg-snapshot"><Icon name="lock" size={14} /><QueryState state={health}>{data => <span data-testid="snapshot-time">快照时间：{formatDate(data.snapshot_at)} · 上海时间</span>}</QueryState></div><div className="ac-tabs" role="tablist" aria-label="认知子页面">{tabs.map((tab, index) => <button key={tab.id} type="button" id={`cg-${tab.id}-tab`} role="tab" aria-selected={active === tab.id} aria-controls="cg-panel" tabIndex={active === tab.id ? 0 : -1} className={active === tab.id ? 'active' : ''} onClick={() => select(tab.id)} onKeyDown={event => { if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return; event.preventDefault(); const next = event.key === 'Home' ? 0 : event.key === 'End' ? tabs.length - 1 : (index + (event.key === 'ArrowRight' ? 1 : -1) + tabs.length) % tabs.length; select(tabs[next].id); document.getElementById(`cg-${tabs[next].id}-tab`)?.focus(); }}>{tab.name}</button>)}</div><div id="cg-panel" role="tabpanel" aria-labelledby={`cg-${active}-tab`} tabIndex={0}>{active === 'today' ? <TodayView revision={revision} query={query} openTopic={setTopic} /> : active === 'focus' ? <FocusView revision={revision} query={query} openTopic={setTopic} /> : active === 'agents' ? <AgentsView revision={revision} query={query} /> : active === 'evidence' ? <EvidenceView revision={revision} /> : <EmptyState title="未找到认知子页面" />}</div><footer className="ac-main-footer"><span>仅展示快照记录 · 快照更新由本机备份程序完成</span><span>ASTROCYTE · 认知</span></footer>{topic !== null && <EvidenceDrawer key={topic} topic={topic} revision={revision} close={close} />}</div>;
}
