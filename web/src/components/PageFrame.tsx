import { type ReactNode } from 'react';
import { navigate, useRoute } from '../router';
import { Icon, NetworkArt } from './DesignIcons';
import { ActiveTab } from './activeTab';

export type Section = 'attention' | 'workspace' | 'swarm';
const tabs = {
  attention: [{ id: 'overview', label: '资料与线索' }, { id: 'opportunities', label: '研究机会' }, { id: 'saved', label: '我的收藏' }],
  workspace: [{ id: 'overview', label: '研究空间' }, { id: 'sessions', label: 'Agent 会话' }, { id: 'proposals', label: '提案与批准' }],
  swarm: [{ id: 'overview', label: '协作现场' }, { id: 'timeline', label: '任务时间线' }, { id: 'artifacts', label: '成果与继承' }],
};

export function PageFrame({ section, title, subtitle, fixture, onRefresh, rail, children }: {
  section: Section; title: string; subtitle: string; fixture: boolean;
  onRefresh: () => void; rail: ReactNode; children: ReactNode;
}) {
  const route = useRoute();
  const options = tabs[section];
  const requested = route.query.get('tab');
  const active = options.some(t => t.id === requested) ? requested! : 'overview';
  function selectTab(id: string) {
    const query = new URLSearchParams(route.query);
    query.set('tab', id);
    navigate(`/${section}?${query}`);
  }
  return <ActiveTab value={active}>
    <div className="ac-breadcrumb"><span>我的工作台</span><Icon name="chevron" size={12} /><span>{title}</span><span className="ac-mode-label">{fixture ? '示例数据' : '真实 API'}</span></div>
    <div className="ac-page-header"><div><h1>{title}</h1><p>{subtitle}</p></div>{!fixture && <button className="ac-button secondary compact" type="button" onClick={onRefresh}>↻ 刷新</button>}</div>
    <div className="ac-content-grid"><div className="ac-center">
      <div className="ac-tabs" role="tablist" aria-label={`${title}子页面`}>
        {options.map((tab, index) => <button key={tab.id} id={`${section}-${tab.id}-tab`} type="button" role="tab" aria-selected={active === tab.id} aria-controls={`${section}-panel`} tabIndex={active === tab.id ? 0 : -1} className={active === tab.id ? 'active' : ''} onClick={() => selectTab(tab.id)} onKeyDown={event => {
          if (!['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return;
          event.preventDefault();
          const next = event.key === 'Home' ? 0 : event.key === 'End' ? options.length - 1 : (index + (event.key === 'ArrowRight' ? 1 : -1) + options.length) % options.length;
          selectTab(options[next].id);
          document.getElementById(`${section}-${options[next].id}-tab`)?.focus();
        }}>{tab.label}</button>)}
      </div>
      <div id={`${section}-panel`} role="tabpanel" aria-labelledby={`${section}-${active}-tab`} tabIndex={0}>{children}</div>
      <footer className="ac-main-footer"><span>让线索连接，让经验生长。</span><span>ASTROCYTE · {fixture ? '示例数据预览' : '本地工作台'}</span></footer>
    </div><aside className="ac-right-rail" aria-label="研究提示">{rail}<section className="ac-rail-card ac-knowledge-card"><div className="ac-knowledge-top"><Icon name="layers" size={26} /><span>KNOWLEDGE THAT LASTS</span></div><h3>别让经验，<br />停在一个会话里。</h3><p>带着来源与适用条件，<br />把这次的发现交给下一次研究。</p><button className="ac-rail-more" type="button" onClick={() => { const query = new URLSearchParams(route.query); query.set('tab', 'artifacts'); navigate(`/swarm?${query}`); }}>查看成果与继承<Icon name="arrow" size={14} /></button></section><p className="ac-rail-foot">本地优先 · 上下文可见 · 由你批准<br /><span>ASTROCYTE RESEARCH WORKBENCH</span></p></aside></div>
  </ActiveTab>;
}

export function RailSummary({ title, rows, note }: { title: string; rows: { id?: string; label: string; value: string | number; onClick?: () => void }[]; note: string }) {
  return <section className="ac-rail-card"><h3>{title}</h3>{rows.map(row => <button type="button" key={row.id ?? row.label} className="ac-attention-item" onClick={row.onClick} disabled={!row.onClick}><span className="ac-mini-icon green"><Icon name="layers" size={16} /></span><span><b>{row.label}</b><small>{row.value}</small></span>{row.onClick && <Icon name="chevron" size={14} />}</button>)}<p className="ac-rail-note">{note}</p></section>;
}

export function IntroCard({ section, onAdd }: { section: Section; onAdd?: () => void }) {
  if (section === 'attention') return <section className="ac-attention-banner"><div><span className="ac-overline">COLLECT → CONNECT → CREATE</span><h2>先让线索沉淀，<br />再让想法生长。</h2><p>收藏不是终点。找到资料之间、资料与项目之间的联系。</p><button type="button" className={`ac-button${onAdd ? '' : ' disabled'}`} disabled={!onAdd} onClick={onAdd} title={onAdd ? '导入 arXiv 或 summarize 既有导出' : '示例模式写操作尚未启用'}><Icon name="plus" size={16} />{onAdd ? '添加资料' : '添加资料 · 尚未启用'}</button></div><div className="ac-banner-paper" aria-hidden="true"><span>IDEA / 01</span><b>把好问题<br />留在这里。</b><Icon name="spark" size={40} /><i>论文 · 视频 · 笔记</i></div></section>;
  if (section === 'workspace') return <section className="ac-project-hero"><div className="ac-project-copy"><span className="ac-overline">YOUR RESEARCH SPACE</span><div className="ac-project-title"><h2>让想法有一个生长的地方。</h2></div><p>围绕目标与项目，连接资料、提案和 Agent 会话。</p><div className="ac-project-tags"><span><Icon name="branch" size={13} />来源与版本可回查</span><span><Icon name="lock" size={13} />由你决定批准边界</span></div><button type="button" className="ac-button disabled" disabled title="提案写入尚未启用"><Icon name="plus" size={16} />记录提案 · 尚未启用</button></div><NetworkArt /><span className="ac-hero-note">RESEARCH CONNECTIONS · 品牌插画</span></section>;
  return <section className="ac-inherit-banner"><Icon name="swarm" size={45} /><div><span className="ac-overline">LOCAL WORK · SHARED CONTEXT</span><h2>让每一步协作，都有迹可循。</h2><p>查看当前任务、有效持有者、阻塞与产物。</p></div></section>;
}
