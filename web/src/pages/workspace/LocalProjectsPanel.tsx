import { useState } from 'react';
import type { components } from '../../api/schema';
import type { ReadApiState } from '../../hooks/useReadApi';
import { CollectionOverview, FilterChips } from '../../components/CollectionOverview';
import { SectionCard } from '../../components/SectionCard';
import { EmptyState } from '../../components/EmptyState';
import { ErrorState } from '../../components/ErrorState';
import { bindingStatusLabel, formatDateTime } from '../../utils/format';
import { localProjectViews, filterLocalProjects } from './localProjectView';
import { matchesQuery } from '../../utils/search';
import styles from './LocalProjectsPanel.module.css';

type Project = components['schemas']['ProjectV1'];
type Session = components['schemas']['SessionV1'];
export function LocalProjectsPanel({ projects, sessions, projectState, sessionState, fixture, onSelect, query }: {
  projects: Project[]; sessions: Session[]; fixture: boolean; onSelect: (project: Project) => void;
  projectState: ReadApiState<components['schemas']['ProjectListV1']>;
  sessionState: ReadApiState<components['schemas']['SessionListV1']>;
  query: string;
}) {
  const [platform, setPlatform] = useState('all');
  const [status, setStatus] = useState('all');
  const [group, setGroup] = useState('all');
  const [sort, setSort] = useState('source');
  const items = localProjectViews(projects, sessions);
  const platforms = [...new Set(items.flatMap(item => item.platforms))];
  const visible = filterLocalProjects(items, platform, status, group, sort).filter(item => matchesQuery(query, [item.project.name, item.project.root_path, item.project.environment_id]));
  const projectReady = fixture || !!projectState.data;
  const sessionReady = fixture || !!sessionState.data;
  const hasActivity = items.some(item => item.lastActivity !== null);
  return <SectionCard title="本地 Agent 与项目" tabs={['overview']} padded>
    <CollectionOverview title="项目总览" description={fixture ? '示例项目与会话，仅用于检查布局。' : '查看当前服务返回的项目与已记录会话。安装、配置和原生能力以逐项探测结果为准。'} metrics={[
      { label: '已加载项目', value: projectReady ? projects.length : projectState.loading ? '加载中' : '未知' },
      { label: '关联平台', value: projectReady && sessionReady ? platforms.length : '未知' },
      { label: '已加载会话', value: sessionReady ? sessions.length : sessionState.loading ? '加载中' : '未知' },
      { label: '最近活动', value: '未提供' },
    ]}>
      <p className={styles.note}>计数基于已加载记录，不代表全盘项目、已安装客户端或正在运行的 Agent。{!fixture && (projectState.data?.next_cursor || sessionState.data?.next_cursor) ? '服务还有未加载记录。' : ''}</p>
      <FilterChips label="平台" value={platform} onChange={setPlatform} disabled={!sessionReady} options={[{ value: 'all', label: '全部' }, ...platforms.map(value => ({ value, label: value })), { value: 'unknown', label: '未关联平台' }]} />
      <FilterChips label="绑定状态" value={status} onChange={setStatus} disabled={!sessionReady} options={[{ value: 'all', label: '全部' }, ...(['observed', 'bound', 'unavailable'] as const).map(value => ({ value, label: bindingStatusLabel(value) })), { value: 'unknown', label: '未记录' }]} />
      <div className={styles.controls}><label>分组<select aria-label="项目分组" value={group} onChange={event => setGroup(event.target.value)}><option value="all">全部分组</option><option value="unknown">未提供分组</option></select></label>
        <label>排序<select aria-label="项目排序" value={sort} onChange={event => setSort(event.target.value)}><option value="source">服务返回顺序</option><option value="name">项目名称</option><option value="activity" disabled={!hasActivity}>最近活动{!hasActivity ? ' · 未提供' : ''}</option></select></label></div>
    </CollectionOverview>
    {!fixture && sessionState.error && <p role="alert" className={styles.note}>会话关联读取失败：{sessionState.error}。平台与绑定状态{sessionState.data ? '显示上次返回记录' : '未知'}。<button type="button" className="ac-button secondary compact" onClick={sessionState.retry}>重试会话</button></p>}
    {!fixture && projectState.loading && !projectState.data ? <p role="status">加载项目中…</p> : !fixture && projectState.error && !projectState.data ? <ErrorState message={projectState.error} onRetry={projectState.retry} /> : !visible.length ? <EmptyState title={items.length ? '没有匹配的本地项目' : '暂无本地项目'} description={items.length ? '调整平台、分组或绑定状态筛选。' : '项目目录接入方式待确定；当前没有注册项目。'} /> : <>
      <p className={styles.note}>共 {visible.length} 条匹配项目</p><ul className={styles.cards}>{visible.map(item => <li key={item.project.id} className={styles.card}>
        <button className={styles.open} type="button" onClick={() => onSelect(item.project)}><h3 className="line-clamp-2">{item.project.name}</h3><span>查看项目详情 →</span></button>
        <div className={styles.tags}><span>项目</span>{item.platforms.length ? item.platforms.map(value => <span key={value}>{value}</span>) : <span>平台{sessionReady ? '未关联' : '未知'}</span>}</div>
        <p>来源目录</p><code className={styles.path}>{item.project.root_path}</code>
        <p>最近活动 · {item.lastActivity ? formatDateTime(item.lastActivity) : '未提供'}</p><p>文件数量 · {item.fileCount ?? '未提供'}</p>
        <footer><span>交接手册 · {item.handoff ?? '未提供'}</span><span>{item.project.worktree_id ? '已记录工作树' : '工作树未提供'}</span></footer>
      </li>)}</ul>
    </>}
  </SectionCard>;
}
