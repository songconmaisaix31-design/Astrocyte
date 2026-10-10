import { navigate } from '../router';
import { useFoundation, useProjects } from '../hooks/useReadApi';
import { fixtureProjects } from '../fixtures';
import { Icon } from './DesignIcons';

const items = [
  { path: '/attention', label: '资料沉淀', icon: 'library' },
  { path: '/workspace', label: '共同工作区', icon: 'grid' },
  { path: '/swarm', label: '蜂群空间', icon: 'swarm' },
] as const;

export function Nav({ currentPath, fixture, open, onNavigate }: { currentPath: string; fixture: boolean; open: boolean; onNavigate: () => void }) {
  const foundation = useFoundation();
  const projects = useProjects();
  const caps = foundation.data?.capabilities;
  const attentionPage = currentPath.startsWith('/attention') || currentPath === '/';
  const canImport = attentionPage && !fixture && !!caps?.imports && !foundation.stale;
  const projectItems = fixture ? fixtureProjects : projects.data?.items ?? [];
  const go = (path: string) => { const query = new URLSearchParams(window.location.search); query.delete('tab'); navigate(`${path}${query.size ? `?${query}` : ''}`); onNavigate(); };
  return <aside className={`ac-sidebar ${open ? 'open' : ''}`}>
    <div className="ac-nav-label">我的工作台</div>
    <nav aria-label="主导航">{items.map(item => <button key={item.path} type="button" className={currentPath.startsWith(item.path) || (currentPath === '/' && item.path === '/attention') ? 'active' : ''} aria-current={currentPath.startsWith(item.path) || (currentPath === '/' && item.path === '/attention') ? 'page' : undefined} onClick={() => go(item.path)}><Icon name={item.icon} /><span>{item.label}</span></button>)}</nav>
    <button type="button" className="ac-new-button" disabled={!canImport} onClick={() => { const query = new URLSearchParams(window.location.search); query.delete('tab'); query.set('import', '1'); navigate(`/attention?${query}`); onNavigate(); }} title={canImport ? '导入 arXiv 或 summarize 既有导出' : '示例模式或服务未提供此写入能力'}><Icon name="plus" size={18} />{canImport ? '添加资料' : '添加资料 · 未启用'}</button>
    <div className="ac-disabled-caption">{canImport ? '保存资料不会启动任务' : '写入操作尚未启用'}</div>
    <div className="ac-sidebar-divider" /><div className="ac-nav-label">我的项目{fixture && <span>示例</span>}</div>
    {projectItems.map(project => <button type="button" className="ac-project-nav" key={project.id} onClick={() => go('/workspace')}><span className="ac-project-symbol">{project.name.slice(0, 1)}</span><span className="text-truncate">{project.name}</span></button>)}
    {!projectItems.length && <p className="ac-sidebar-empty">{projects.loading ? '加载中…' : projects.error ? '项目列表无法获取' : '暂无已注册项目'}</p>}
    {!fixture && projects.error && <button type="button" className="ac-text-button" onClick={projects.retry}>重试项目列表</button>}
    {!fixture && projects.stale && <p className="ac-sidebar-empty" role="status">项目列表已过期</p>}
    <button type="button" className="ac-register-button" onClick={() => go('/workspace')}><Icon name="plus" size={15} />管理本地项目</button>
    <div className="ac-sidebar-bottom">{fixture && <div className="ac-small-quote"><Icon name="branch" size={22} /><p>把一次次探索，<br />连接成可以继承的经验。</p><button type="button" onClick={() => go('/swarm')}>进入开发空间<Icon name="arrow" size={14} /></button></div>}
      <details className="ac-capabilities" aria-label="系统能力"><summary>系统能力</summary>{caps ? <div>{Object.entries({ 导入: caps.imports, 审核: caps.approvals, 执行: caps.execution, 恢复: caps.native_resume, 交接: caps.handoff }).map(([label, value]) => <span key={label}>{label} · {value ? '可用' : '未启用'}</span>)}</div> : <p>{foundation.loading ? '加载中…' : '无法获取'}</p>}{foundation.stale && <p>能力信息已过期</p>}{foundation.error && <button type="button" className="ac-text-button" onClick={foundation.retry}>重试能力</button>}</details>
    </div>
  </aside>;
}
