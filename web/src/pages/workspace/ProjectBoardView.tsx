import { useState } from 'react';
import { BrandIcon } from '../../components/BrandIcon';
import { FilterChips } from '../../components/CollectionOverview';
import { formatDateTime } from '../../utils/format';
import { boardSections, filterBoardProjects, initialBoardFilters, humanIntentLabel, projectSourceLabel, type BoardFilters, type BoardProject } from './projectBoardPresentation';
import styles from './ProjectBoard.module.css';

export function ProjectBoardView({ items, query, complete, onSelect, onClearQuery }: { items: BoardProject[]; query: string; complete: boolean; onSelect: (item: BoardProject) => void; onClearQuery: () => void }) {
  const [filters, setFilters] = useState(initialBoardFilters);
  const [view, setView] = useState('flat');
  function filter(key: keyof BoardFilters, value: string) { setFilters(previous => ({ ...previous, [key]: value })); }
  const unique = (values: string[]) => [...new Set(values)].sort();
  const visible = filterBoardProjects(filterBoardProjects(items, { ...initialBoardFilters, archive: 'all', search: query, sort: 'source' }), filters);
  const folders = unique(items.flatMap(item => item.folders));
  const clients = unique(items.flatMap(item => item.clients));
  return <div className={styles.board}>
    <dl className={styles.metrics} aria-label="项目统计">
      {[{ label: '项目', value: items.length }, { label: '来源文件夹', value: folders.length }, { label: '已记录客户端', value: clients.length }, { label: '有活动时间', value: items.filter(item => item.activityAt).length }, { label: '已归档', value: items.filter(item => item.archived).length }].map(metric => <div key={metric.label}><dd>{metric.value}</dd><dt>{metric.label}</dt></div>)}
    </dl>
    <p className={styles.note}>{complete ? '当前已保存清单' : '部分来源暂不可读 · 仅统计已返回项目'}{query ? ` · 页面搜索：${query}` : ''}</p>
    <FilterChips label="看的方式" value={view} onChange={setView} options={[{ value: 'flat', label: '平铺' }, { value: 'platform', label: '按平台' }, { value: 'group', label: '按分组' }, { value: 'timeline', label: '时间线' }]} />
    <div className={styles.filters}>
      <div className={styles.controls}>
        <label>搜索项目<input aria-label="搜索项目" type="search" value={filters.search} placeholder="名称、目录或笔记" onChange={event => filter('search', event.target.value)} /></label>
        <label>继续意愿<select aria-label="继续意愿" value={filters.intent} onChange={event => filter('intent', event.target.value)}><option value="all">全部意愿</option>{unique(items.map(item => item.intent).filter(Boolean)).map(value => <option key={value}>{value}</option>)}<option value="__unset">尚未标记</option></select></label>
        <label>排序<select aria-label="排序" value={filters.sort} onChange={event => filter('sort', event.target.value)}><option value="activity">最近记录活动</option><option value="name">项目名称</option><option value="source">来源顺序</option></select></label>
      </div>
      <details className={styles.advanced}><summary>更多筛选（客户端、分组、来源、活动与归档）</summary><div className={styles.controls}>
        <label>客户端<select aria-label="客户端" value={filters.client} onChange={event => filter('client', event.target.value)}><option value="all">全部客户端</option>{clients.map(value => <option key={value}>{value}</option>)}<option value="unknown">客户端未知</option></select></label>
        <label>分组<select aria-label="分组" value={filters.group} onChange={event => filter('group', event.target.value)}><option value="all">全部分组</option>{unique(items.map(item => item.group).filter(Boolean)).map(value => <option key={value}>{value}</option>)}<option value="unknown">未分组</option></select></label>
        <label>来源<select aria-label="来源" value={filters.source} onChange={event => filter('source', event.target.value)}><option value="all">全部来源</option>{unique(items.flatMap(item => item.sources)).map(value => <option key={value} value={value}>{projectSourceLabel(value)}</option>)}</select></label>
        <label>活动状态<select aria-label="活动状态" value={filters.status} onChange={event => filter('status', event.target.value)}><option value="all">全部状态</option>{unique(items.map(item => item.activityStatus)).map(value => <option key={value}>{value}</option>)}</select></label>
        <label>归档<select aria-label="归档" value={filters.archive} onChange={event => filter('archive', event.target.value)}><option value="active">未归档</option><option value="archived">只看归档</option><option value="all">全部看</option></select></label>
      </div></details>
      <div className={styles.result}><span>共 {visible.length} 个项目</span><button type="button" className="ac-text-button" onClick={() => { setFilters(initialBoardFilters); onClearQuery(); }}>清除筛选</button></div>
    </div>
    {!visible.length && <p role="status" className={styles.empty}>{items.length ? '没有匹配的项目，请调整筛选。' : '已保存清单中没有项目。可从登记目录发现项目，或手动登记允许接入的目录。'}</p>}
    {boardSections(visible, view).map(section => <section key={section.label} aria-label={section.label}>
      {view !== 'flat' && <h3 className={styles.groupTitle}>{section.label} <span>{section.projects.length}</span></h3>}
      <ul className={`${styles.cards} ${view === 'timeline' ? styles.timeline : ''}`}>{section.projects.map(item => <li key={item.id}>
        <button type="button" data-project-id={item.id} className={styles.card} onClick={() => onSelect(item)} aria-label={`查看项目 ${item.name}`}>
          <div className={styles.cardHeader}><h3>{item.name}</h3>{item.archived && <span className={styles.tag}>已归档</span>}</div>
          <div className={styles.clients}>{item.clients.length ? item.clients.map(client => <span key={client}><BrandIcon name={client} size={18} />{client}</span>) : <span>客户端未知</span>}</div>
          <p className={styles.notes}>{item.notes || '尚无备注'}</p>
          <code className={styles.path}>{item.folders[0] ?? '目录未提供'}{item.folders.length > 1 ? ` · 另 ${item.folders.length - 1} 个目录` : ''}</code>
          <p>最近记录活动 · {item.activityAt ? formatDateTime(item.activityAt) : '未知'}</p>
          <small>{item.activitySource || '活动来源未提供'}</small>
          <div className={styles.cardFooter}><span>{humanIntentLabel(item.intent)} · {item.group || '未分组'}</span><span>打开项目 →</span></div>
        </button>
      </li>)}</ul>
    </section>)}
  </div>;
}
