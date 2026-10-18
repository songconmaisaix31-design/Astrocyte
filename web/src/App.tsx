import { useEffect, useState } from 'react';
import { navigate, useRoute } from './router';
import { Nav } from './components/Nav';
import { Icon } from './components/DesignIcons';
import { AttentionPage } from './pages/attention/AttentionPage';
import { WorkspacePage } from './pages/workspace/WorkspacePage';
import { SwarmPage } from './pages/swarm/SwarmPage';
import { CognitionPage } from './pages/cognition/CognitionPage';
import { RelayWorksPage } from './pages/workspace/relayWorksPage';
import './styles/preview.css';
import './styles/workbench.css';
import './styles/relay.css';
import './styles/observatory.css';

export function App() {
  const route = useRoute();
  const cognition = route.path.startsWith('/cognition');
  const relay = route.path.startsWith('/workspace/works');
  const [query, setQuery] = useState('');
  const [menuOpen, setMenuOpen] = useState(false);
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      const target = event.target;
      if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k' && !document.querySelector('[role="dialog"]')) { event.preventDefault(); document.getElementById('ac-global-search')?.focus(); return; }
      if (document.querySelector('[role="dialog"]') || event.ctrlKey || event.metaKey || event.altKey || (target instanceof HTMLElement && (target.matches('input,textarea,select') || target.isContentEditable))) return;
      if (event.key === '/') { event.preventDefault(); document.getElementById('ac-global-search')?.focus(); }
      if (['1', '2', '3'].includes(event.key)) { const params = new URLSearchParams(window.location.search); params.delete('tab'); navigate(`/${['attention', 'workspace', 'swarm'][Number(event.key) - 1]}?${params}`); }
      if (event.key === 'Escape') setMenuOpen(false);
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, []);
  const props = { query };
  const workspace = route.path.startsWith('/workspace');
  const workspaceTab = route.query.get('tab') ?? 'overview';
  const searchName = workspace ? workspaceTab === 'sessions' ? '搜索已加载 Agent 会话' : workspaceTab === 'proposals' ? '搜索已加载提案' : '搜索已加载项目、目录与备注' : '搜索当前页资料、任务或 Agent';
  const searchPlaceholder = workspace ? workspaceTab === 'sessions' ? '搜索客户端、会话或项目 ID…' : workspaceTab === 'proposals' ? '搜索提案标题、目标或 ID…' : '搜索项目、目录或备注…' : '搜索资料、任务，或一个值得继续的问题…';
  const pageName = route.path === '/cognition/focus' ? '注意力流' : route.path === '/cognition/evidence' ? '事件与证据' : cognition ? '认知工作台' : relay ? '作品寻回册' : workspace ? '共同工作区' : route.path.startsWith('/swarm') ? '协作网络' : '资料来源';
  const page = cognition ? <CognitionPage path={route.path} query={query} /> : relay ? <RelayWorksPage query={query} onClearQuery={() => setQuery('')} /> : workspace ? <WorkspacePage {...props} onClearQuery={() => setQuery('')} /> : route.path.startsWith('/swarm') ? <SwarmPage {...props} /> : <AttentionPage {...props} />;
  return <div className="ac-workbench ac-observatory">
    <a href="#ac-main" className="ac-skip">跳转到内容</a>
    <Nav currentPath={route.path} open={menuOpen} onNavigate={() => setMenuOpen(false)} />
    <div className="ac-app-body">
    <header className="ac-topbar"><div className="ac-topbar-inner">
      <button type="button" className="ac-icon-button ac-mobile-toggle" aria-label="展开导航" aria-expanded={menuOpen} onClick={() => setMenuOpen(!menuOpen)}><Icon name="menu" /></button>
      <div className="ac-shell-breadcrumb"><span>我的工作空间</span><Icon name="chevron" size={12} /><strong>{pageName}</strong></div>
      <div className="ac-global-search"><Icon name="search" size={18} /><input id="ac-global-search" aria-label={searchName} placeholder={searchPlaceholder} value={query} onChange={event => setQuery(event.target.value)} onKeyDown={event => { if (event.key === 'Escape') { setQuery(''); event.currentTarget.blur(); } }} />{query ? <button type="button" className="ac-icon-button" aria-label="清空搜索" onClick={() => setQuery('')}><Icon name="close" size={15} /></button> : <kbd>/</kbd>}</div>
      <div className="ac-header-right"><span className="ac-preview-pill"><i />本地工作台</span><span className="ac-user-avatar" aria-label="本地用户">我</span></div>
    </div></header>
    <main id="ac-main" className="ac-main" tabIndex={-1}>{query && <p className="ac-search-note" role="status">当前页筛选：{query}（仅搜索已加载内容）</p>}{page}</main>
    </div>
  </div>;
}
