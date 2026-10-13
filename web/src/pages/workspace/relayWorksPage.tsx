import { ActiveTab } from '../../components/activeTab';
import { WorksAlbumPanel } from './WorksAlbumPanel';
import { useState } from 'react';

export function RelayWorksPage({ query, onClearQuery }: { query: string; onClearQuery: () => void }) {
  const [revision, setRevision] = useState(0);
  return <ActiveTab value="overview"><div className="relay-page">
    <div className="ac-breadcrumb"><span>我的工作台</span><span>作品寻回册</span><span className="ac-mode-label">本机数据</span></div>
    <div className="ac-page-header"><div><h1>作品寻回册</h1><p>找回作品，阅读交接手册，再由你决定下一步。</p></div><button type="button" className="ac-button secondary compact" onClick={() => setRevision(value => value + 1)}>↻ 刷新</button></div>
    <WorksAlbumPanel key={revision} fixture={false} query={query} onClearQuery={onClearQuery} />
  </div></ActiveTab>;
}
