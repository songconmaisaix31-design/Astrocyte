import { useCallback, useState } from 'react';
import { DetailPanel, Field } from './DetailPanel';
import { Icon } from './DesignIcons';
import { EmptyState } from './EmptyState';

const provenance = '来自用户提供的 Astrocyte-preview.html 固定设计示例；不是 API 数据、真实会话或实验结论。';
const tasks = [
  { id: 'T-11', name: '文献与假设', actor: 'Claude Code', tone: 'green', x: 5, y: 29, body: '整理软边界与硬边界的假设、适用条件与对照依据。' },
  { id: 'T-12', name: '实验设计', actor: '待认领', tone: 'amber', x: 38, y: 9, body: '固定预算、采样方式和验证方法，形成最小可执行方案。' },
  { id: 'T-13', name: '候选实现', actor: 'Codex', tone: 'blue', x: 40, y: 61, body: '保留候选产出、输入资料版本与复现步骤；提交候选不等于已采用。' },
  { id: 'T-14', name: '独立复核', actor: 'OpenCode', tone: 'purple', x: 72, y: 31, body: '从可见上下文包接手候选，独立检查来源与可复现性。' },
];
const events = [
  { time: '14:32', actor: 'Codex', title: '提交了一份候选产出', detail: '软边界实验代码已进入待复核队列。', tone: 'blue', target: 'T-13' },
  { time: '14:18', actor: 'Claude Code', title: '补充了两条研究线索', detail: '软边界与硬边界需要在同一预算下比较。', tone: 'green', target: 'T-11' },
  { time: '13:56', actor: '上下文交接', title: '准备了独立复核的上下文包', detail: '目标、候选版本和最近决定随工作项一起交付。', tone: 'purple', target: 'T-14' },
];

export function ResearchRoutes({ fixture }: { fixture: boolean }) {
  const [selected, setSelected] = useState<string | null>(null);
  const close = useCallback(() => setSelected(null), []);
  if (!fixture) return <div className="ac-panel"><p className="ac-rail-note">研究路线尚未接入。当前接口没有分支与依赖关系，项目详情保留真实路径和版本。</p></div>;
  return <><div className="ac-branch-grid">{[
    { name: '软边界路线', label: '已有候选', tone: 'green', body: '文献调研 → 实验设计 → 候选实现', footer: '等待独立复核' },
    { name: '硬边界路线', label: '待补证据', tone: 'amber', body: '同预算对照尚未完成，保留待查问题。', footer: '形成下一步候选' },
  ].map(route => <button key={route.name} type="button" onClick={() => setSelected(route.name)}><div className="ac-line"><span className={`ac-branch-label ${route.tone}`}><Icon name="branch" size={17} />{route.name}</span><span className={`ac-badge ${route.tone}`}>{route.label} · 示例</span></div><p>{route.body}</p><div className="ac-branch-footer"><span>{route.footer}</span><Icon name="arrow" size={16} /></div></button>)}</div>{selected && <DetailPanel title={`${selected} · 设计示例`} onClose={close} showDisabledNotice disabledNoticeText="研究路线写入尚未启用"><Field label="来源边界">{provenance}</Field><Field label="最小下一步">绑定真实论文、环境与固定采样预算，再完成独立复核。</Field><Field label="缺失依据">具体论文版本、实验环境与固定样本。</Field></DetailPanel>}</>;
}

export function DesignTimeline({ fixture }: { fixture: boolean }) {
  const [selected, setSelected] = useState<string | null>(null);
  const close = useCallback(() => setSelected(null), []);
  if (!fixture) return <EmptyState title="事件时间线尚未接入" description="当前 API 只返回任务快照；没有事件记录时不生成时间或协作动态。" />;
  const task = tasks.find(item => item.id === selected);
  return <div className="ac-panel"><div className="ac-timeline">{events.map(event => <article key={event.time}><div className="ac-event-marker"><i className={event.tone} /><span /></div><div className="ac-event-content"><div className="ac-line"><span><strong>{event.actor}</strong><span className="ac-muted"> · {event.title}</span></span><time>{event.time}</time></div><p>{event.detail}</p><button type="button" className="ac-text-button" onClick={() => setSelected(event.target)}>查看工作记录 · 示例<Icon name="arrow" size={14} /></button></div></article>)}</div><p className="ac-panel-note">{provenance}</p>{task && <DetailPanel title={`${task.id} · 设计示例`} onClose={close} showDisabledNotice disabledNoticeText="示例不执行认领、交接或资源预留"><Field label="来源">{provenance}</Field><Field label="工作项">{task.name}</Field><Field label="说明">{task.body}</Field><Field label="客户端示例">{task.actor}（CLI 名称不代表底层模型或活跃进程）</Field></DetailPanel>}</div>;
}

export function DesignGraph({ fixture }: { fixture: boolean }) {
  const [graph, setGraph] = useState(false);
  const [selected, setSelected] = useState<string | null>(null);
  const close = useCallback(() => setSelected(null), []);
  const task = tasks.find(item => item.id === selected);
  return <><div className="ac-section-head"><div><h2>把协调交给环境</h2><p>{fixture ? '设计示例 · 观察工作项与接续关系' : '拓扑需有认领、交接或复核事件支撑'}</p></div><button type="button" className="ac-button compact secondary" disabled={!fixture} title={!fixture ? '当前 API 未提供协作事件，拓扑尚未接入' : undefined} onClick={() => setGraph(!graph)}><Icon name={graph ? 'layers' : 'swarm'} size={16} />{graph ? '列表视图' : '拓扑视图'}</button></div>{fixture ? graph ? <div className="ac-graph"><svg viewBox="0 0 750 300" preserveAspectRatio="none" aria-hidden="true"><path d="M130 140 C200 140 220 65 335 65 M130 140 C210 140 230 220 345 220 M445 220 C520 220 525 150 620 150" fill="none" stroke="#a8c7ad" strokeWidth="1.5" strokeDasharray="5 5" /></svg>{tasks.map(item => <button type="button" key={item.id} className={`ac-graph-node ${item.tone}`} style={{ left: `${item.x}%`, top: `${item.y}%` }} onClick={() => setSelected(item.id)}><small>{item.id} · 示例</small><b>{item.name}</b><span>{item.actor}</span></button>)}<span className="ac-graph-caption">设计示意 · 非实时拓扑</span></div> : <div className="ac-list-panel">{tasks.map(item => <button type="button" key={item.id} className="ac-task-row" onClick={() => setSelected(item.id)}><span className={`ac-task-icon ${item.tone}`}><Icon name="file" size={19} /></span><span className="ac-task-copy"><strong>{item.name}</strong><span>{item.id} · {item.actor} · 设计示例</span></span><Icon name="chevron" size={15} /></button>)}</div> : <p className="ac-panel-note">拓扑尚未接入；下面显示真实任务与工作项快照。</p>}{task && <DetailPanel title={`${task.id} · 设计示例`} onClose={close} showDisabledNotice disabledNoticeText="此示意不执行真实协作"><Field label="来源边界">{provenance}</Field><Field label="工作项">{task.name}</Field><Field label="说明">{task.body}</Field></DetailPanel>}</>;
}
