import { useEffect, useState } from 'react';
import { navigate, useRoute } from './router';
import { Nav } from './components/Nav';
import { FixtureBanner } from './components/FixtureBanner';
import { Icon, Mark } from './components/DesignIcons';
import { AttentionPage } from './pages/attention/AttentionPage';
import { WorkspacePage } from './pages/workspace/WorkspacePage';
import { SwarmPage } from './pages/swarm/SwarmPage';
import { isFixtureMode } from './fixtures';
import './styles/preview.css';

export function App() {
  const route = useRoute();
  const fixture = isFixtureMode(route.query);
  const [query, setQuery] = useState('');
  const [menuOpen, setMenuOpen] = useState(false);
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      const target = event.target;
      if (document.querySelector('[role="dialog"]') || event.ctrlKey || event.metaKey || event.altKey || (target instanceof HTMLElement && (target.matches('input,textarea,select') || target.isContentEditable))) return;
      if (event.key === '/') { event.preventDefault(); document.getElementById('ac-global-search')?.focus(); }
      if (['1', '2', '3'].includes(event.key)) { const params = new URLSearchParams(window.location.search); params.delete('tab'); navigate(`/${['attention', 'workspace', 'swarm'][Number(event.key) - 1]}?${params}`); }
      if (event.key === 'Escape') setMenuOpen(false);
    };
    document.addEventListener('keydown', onKey);
    return () => document.removeEventListener('keydown', onKey);
  }, []);
  const props = { fixture, query };
  const modeKey = fixture ? 'fixture' : 'api';
  const page = route.path.startsWith('/workspace') ? <WorkspacePage key={modeKey} {...props} /> : route.path.startsWith('/swarm') ? <SwarmPage key={modeKey} {...props} /> : <AttentionPage key={modeKey} {...props} />;
  return <div className="ac-workbench">
    <a href="#ac-main" className="ac-skip">跳转到内容</a>
    <header className="ac-topbar"><div className="ac-topbar-inner">
      <button type="button" className="ac-icon-button ac-mobile-toggle" aria-label="展开导航" aria-expanded={menuOpen} onClick={() => setMenuOpen(!menuOpen)}><Icon name="menu" /></button>
      <button type="button" className="ac-brand" aria-label="Astrocyte 首页" onClick={() => navigate(`/attention${fixture ? '?fixture=1' : ''}`)}><Mark /><span>Astrocyte<small>个人研究与创造工作台</small></span></button>
      <div className="ac-global-search"><Icon name="search" size={18} /><input id="ac-global-search" aria-label="搜索当前页资料、任务或 Agent" placeholder="搜索资料、任务，或一个值得继续的问题…" value={query} onChange={event => setQuery(event.target.value)} onKeyDown={event => { if (event.key === 'Escape') { setQuery(''); event.currentTarget.blur(); } }} />{query ? <button type="button" className="ac-icon-button" aria-label="清空搜索" onClick={() => setQuery('')}><Icon name="close" size={15} /></button> : <kbd>/</kbd>}</div>
      <div className="ac-header-right"><span className="ac-preview-pill"><i />{fixture ? '示例数据' : '本地工作台'}</span><span className="ac-user-avatar" aria-label="本地用户">我</span></div>
    </div></header>
    <div className="ac-layout"><Nav currentPath={route.path} fixture={fixture} open={menuOpen} onNavigate={() => setMenuOpen(false)} /><main id="ac-main" className="ac-main" tabIndex={-1}>{fixture && <FixtureBanner />}{query && <p className="ac-search-note" role="status">当前页筛选：{query}（仅搜索已加载内容）</p>}{page}</main></div>
  </div>;
}
