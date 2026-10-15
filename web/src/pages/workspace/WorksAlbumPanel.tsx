import { useEffect, useRef, useState } from 'react';
import { BrandIcon } from '../../components/BrandIcon';
import { FilterChips } from '../../components/CollectionOverview';
import { SectionCard } from '../../components/SectionCard';
import { DetailPanel, Field } from '../../components/DetailPanel';
import { formatDateTime } from '../../utils/format';
import {
  attitudeLabel, attitudeOptions, fileVerifiedLabel, filterWorks,
  handoffSourceLabel, initialWorksFilters, platformLabel, worksSections,
  type WorksFilters, type WorksSnapshot, type WorksWork,
} from './worksAlbumPresentation';
import styles from './WorksAlbum.module.css';
import { ApiError } from '../../api/client';
import { localProjectsApi } from '../../api/s1';
import { worksAlbumApi } from './worksAlbumApi';
import { WorksMetadataForm, WorksDeliveryForm, WorksScanForm, type WorksMetadataEdit } from './WorksCommandForms';
import { relayAlbumMetrics } from './relayAlbumMetrics';
import { relayApi } from '../../api/relay';
import { useReadApi } from '../../hooks/useReadApi';
import { RelayIntentPanel } from './relayIntentPanel';
import { RelayOverview } from './relayOverview';

export function WorksAlbumPanel({ query, onClearQuery }: { query: string; onClearQuery: () => void }) {
  const [cached, setCached] = useState<WorksSnapshot>({ works: [], platforms: [], status: 'unknown', observed_at: null, failures: [] });
  const snapshot = cached;
  const relayTasks = useReadApi(signal => relayApi.tasks(signal), { refreshMs: 3000 });
  const [relayWorkID, setRelayWorkID] = useState('');
  const relayWork = snapshot.works.find(work => work.id === relayWorkID);
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [uncertain, setUncertain] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [modal, setModal] = useState<'' | 'scan' | 'delivery'>('');
  const [roots, setRoots] = useState<string[]>([]);
  const [rootsLoading, setRootsLoading] = useState(false);
  const inFlight = useRef(false);
  useEffect(() => {
    const controller = new AbortController();
    void worksAlbumApi.list(controller.signal).then(result => setCached(result.snapshot)).catch(cause => {
      if (!controller.signal.aborted) setError(`作品清单读取失败：${cause instanceof Error ? cause.message : '未知错误'}`);
    }).finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, []);
  const readCache = async () => {
    setLoading(true);
    try { setCached((await worksAlbumApi.list()).snapshot); setError(''); setNotice('已重新读取保存结果'); setUncertain(false); }
    catch (cause) { setError(`作品清单读取失败：${cause instanceof Error ? cause.message : '未知错误'}`); }
    finally { setLoading(false); }
  };
  const perform = async (action: (confirmed: () => void) => Promise<WorksSnapshot>, message: string) => {
    if (inFlight.current || uncertain) return;
    inFlight.current = true;
    setBusy(true); setError(''); setNotice('');
    let confirmed = false;
    try { setCached(await action(() => { confirmed = true; })); setNotice(message); }
    catch (cause) {
      const unknown = !confirmed && (!(cause instanceof ApiError) || ['delivery_unknown', 'internal_error'].includes(cause.code));
      if (unknown) setUncertain(true);
      setError(`${confirmed ? '命令已保存，列表读取失败' : unknown ? '请求结果未知，先重新读取保存结果；不会自动重试' : '请求未完成'}：${cause instanceof Error ? cause.message : '未知错误'}`);
    } finally { inFlight.current = false; setBusy(false); }
  };
  const loadRoots = async () => {
    const result = await localProjectsApi.listRegisteredProjects();
    if (!result.snapshot.observed_at || result.snapshot.status === 'unknown') throw new Error('登记项目观察尚未知，请等待登记项目发现完成后重新读取');
    return [...new Set(result.snapshot.projects.map(project => project.root))];
  };
  const openScan = async () => {
    setModal('scan'); setSelectedID(''); setRootsLoading(true);
    try { setRoots(await loadRoots()); }
    catch (cause) { setError(cause instanceof Error ? cause.message : '登记项目读取失败'); }
    finally { setRootsLoading(false); }
  };
  const scan = (selectedRoots: string[]) => perform(async () => (await worksAlbumApi.refresh(selectedRoots, crypto.randomUUID())).snapshot, '已保存文件观察');
  const sync = async () => {
    try {
      const approved = await loadRoots();
      const rootKey = (root: string) => root[1] === ':' ? root.split(String.fromCharCode(92)).join('/').toLowerCase() : root;
      const albumRoots = new Set(snapshot.works.map(work => rootKey(work.root)));
      const selectedRoots = approved.filter(root => albumRoots.has(rootKey(root)));
      if (!selectedRoots.length) { await openScan(); return; }
      await scan(selectedRoots);
    } catch (cause) { setError(cause instanceof Error ? cause.message : '登记项目读取失败'); }
  };
  const save = (item: WorksWork, edit: WorksMetadataEdit) => perform(async confirmed => {
    const result = await worksAlbumApi.edit(item, edit, crypto.randomUUID());
    confirmed();
    setCached(previous => ({ ...previous, works: previous.works.map(work => work.id === result.work.id ? result.work : work) }));
    return (await worksAlbumApi.list()).snapshot;
  }, '人工记录已保存');
  const disabled = busy || loading || uncertain;
  const [filters, setFilters] = useState<WorksFilters>(initialWorksFilters);
  const [selectedID, setSelectedID] = useState('');
  const countsKnown = snapshot.observed_at !== null && snapshot.status !== 'unknown';
  const filter = (key: keyof WorksFilters, value: string | boolean) => setFilters(previous => ({ ...previous, [key]: value }));
  const unique = (values: string[]) => [...new Set(values.filter(Boolean))].sort();
  const visible = filterWorks(snapshot.works, { ...filters, search: filters.search || query });
  const selected = snapshot.works.find(work => work.id === selectedID);
  const platforms = unique(snapshot.works.map(work => work.platform));
  const groups = unique(snapshot.works.map(work => work.group));

  return <SectionCard title="" tabs={['overview']} padded>
    <div className={styles.board} role="region" aria-label="作品寻回册">
      <div className={styles.heading}><h2 className="sr-only">作品寻回册</h2><div className={styles.actions}>
        <button type="button" className="ac-button compact" disabled={disabled} onClick={() => { void sync(); }}>同步最近改动</button>
        <button type="button" className="ac-button secondary compact" disabled={disabled} onClick={() => { void openScan(); }}>扫描导入项目</button>
        <button type="button" className="ac-button secondary compact" disabled={disabled} onClick={() => { setSelectedID(''); setModal('delivery'); }}>AI 投递 JSON</button>
      </div></div>
      <div className={styles.actions}><button type="button" className="ac-text-button" disabled={busy || loading} onClick={() => { void readCache(); }}>重新读取保存结果</button>{loading && <span role="status">正在读取已保存作品…</span>}</div>
      {error && <p role="alert" className={styles.empty}>{error}</p>}
      {notice && <p role="status" className={styles.note}>{notice}</p>}
      {snapshot.failures.length > 0 && <ul aria-label="未完整观察目录" className={styles.note}>{snapshot.failures.map((failure, index) => <li key={`${failure.root}-${index}`}>{failure.root || '扫描来源'} · {failure.reason}</li>)}</ul>}
      <dl className={`${styles.metrics} relay-metrics`} aria-label="作品统计">
        {relayAlbumMetrics(snapshot, !relayTasks.stale ? relayTasks.data?.dispatched_work_ids : undefined).map(metric => <div key={metric.label}><dd>{metric.value ?? '未知'}</dd><dt>{metric.label}</dt></div>)}
      </dl>
      <p className={styles.note}>接力过仅统计接入层确认的真实派发；交接手册不代表已派发。</p>
      {snapshot.platforms.length > 0 && <ul className={styles.platforms} aria-label="平台统计">{snapshot.platforms.map(stat => <li key={stat.platform}><b><BrandIcon name={stat.platform} size={18} />{platformLabel(stat.platform)}</b><span>{stat.works} 作品 · {stat.folders} 文件夹</span><span>{stat.handed_over} 有手册 · {snapshot.works.filter(work => work.platform === stat.platform).every(work => work.files.verified && work.files.mtime !== null) ? stat.recent : '未知'} 最近在动</span></li>)}</ul>}
      <div className={styles.viewbar}><FilterChips label="看的方式" value={filters.view} onChange={value => filter('view', value)} options={[{ value: 'flat', label: '平铺' }, { value: 'platform', label: '按平台' }, { value: 'group', label: '按分组' }, { value: 'timeline', label: '时间线' }]} /><p className={styles.note}>{snapshot.status === 'complete' ? '当前已保存清单' : snapshot.status === 'partial' ? '部分目录暂不可读 · 仅统计已返回作品' : snapshot.status === 'stale' ? '上次已保存观察 · 当前刷新未完成' : '尚无已保存的作品观察 · 数量未知'}{query ? ` · 页面搜索：${query}` : ''}</p></div>
      <div className={styles.filters}>
        <div className={styles.controls}>
          <label>搜索作品<input aria-label="搜索作品" type="search" value={filters.search || query} placeholder="名称、平台、目录或交接手册" onChange={event => filter('search', event.target.value)} /></label>
          <label>平台<select aria-label="平台" value={filters.platform} onChange={event => filter('platform', event.target.value)}><option value="all">全部平台</option>{platforms.map(value => <option key={value} value={value}>{platformLabel(value)}</option>)}</select></label>
          <label>分组<select aria-label="分组" value={filters.group} onChange={event => filter('group', event.target.value)}><option value="all">全部分组</option>{groups.map(value => <option key={value} value={value}>{value}</option>)}<option value="">未分组</option></select></label>
          <label>态度<select aria-label="态度" value={filters.attitude} onChange={event => filter('attitude', event.target.value)}><option value="all">全部态度</option>{attitudeOptions.map(option => <option key={option.value} value={option.value}>{option.label}</option>)}<option value="">尚未标记</option></select></label>
          <label>排序<select aria-label="排序" value={filters.sort} onChange={event => filter('sort', event.target.value)}><option value="recent">最近改动</option><option value="updated">最近更新</option><option value="name">作品名称</option></select></label>
        </div>
        <div className={styles.result}><span>{countsKnown ? `共 ${visible.length} 个已保存作品` : '作品数量未知'}</span><label style={{ display: 'flex', alignItems: 'center', gap: 6 }}><input type="checkbox" checked={filters.includeHidden} onChange={event => filter('includeHidden', event.target.checked)} />显示已隐藏</label><button type="button" className="ac-text-button" onClick={() => { setFilters(initialWorksFilters); onClearQuery(); }}>清除筛选</button></div>
      </div>
      {!snapshot.works.length && <p role="status" className={styles.empty}>尚无已保存的作品观察。可扫描已批准的登记项目，或保存 JSON 投递线索。</p>}
      {worksSections(visible, filters.view).map(section => <section key={section.label} aria-label={section.label}>
        {filters.view !== 'flat' && <h3 className={styles.groupTitle}>{section.label} <span>{section.works.length}</span></h3>}
        <ul className={`${styles.cards} ${filters.view === 'timeline' ? styles.timeline : ''}`}>{section.works.map(work => <li key={work.id} className="relay-work-card">
          <button type="button" data-work-id={work.id} className={styles.card} onClick={() => { setModal(''); setSelectedID(work.id); }} aria-label={`查看作品 ${work.name}`}>
            <div className={styles.cardHeader}><h3>{work.name}</h3>{work.hidden && <span className={styles.tag}>已隐藏</span>}</div>
            <div><span className={`${styles.badge} ${work.files.verified ? '' : styles.unverified}`}>{fileVerifiedLabel(work.files)}</span></div>
            <div className={styles.clients}><span><BrandIcon name={work.platform} size={18} />{platformLabel(work.platform)}</span></div>
            <p className={styles.handoff}>{work.handoff.content || '尚无交接手册'}</p>
            <code className={styles.path}>{work.files.source_path || '路径未提供'}{work.files.verified ? ` · ${work.files.file_count} 个文件` : ' · 文件数量未知'}</code>
            <small>最近改动 · {work.files.verified && work.files.mtime ? formatDateTime(work.files.mtime) : '未知'}</small>
            <div className={styles.cardFooter}><span>{attitudeLabel(work.attitude)} · {work.group || '未分组'}</span><span>查看详情 →</span></div>
          </button>
          <div className="relay-card-actions" aria-label={`${work.name} 的操作`}>
            <button type="button" className="ac-button secondary compact" disabled={disabled} onClick={() => { void perform(async () => {
              const approved = await loadRoots();
              const normalized = (root: string) => {
                const value = root.replaceAll('\\', '/').replace(/\/$/, '');
                return /^[a-z]:/i.test(value) ? value.toLowerCase() : value;
              };
              const root = approved.find(root => normalized(root) === normalized(work.root));
              if (!root) throw new Error('作品根目录尚未批准，请先在项目接入中登记并确认权限');
              return (await worksAlbumApi.refresh([root], crypto.randomUUID())).snapshot;
            }, '已保存作品文件观察'); }}>从文件核实</button>
            <button type="button" className="ac-button secondary compact" onClick={() => { setModal(''); setSelectedID(work.id); }}>交接手册</button>
            <button type="button" className="ac-button compact" disabled={disabled} onClick={() => { setModal(''); setSelectedID(''); setRelayWorkID(work.id); }}>接力</button>
          </div>
        </li>)}</ul>
      </section>)}
    </div>
    <RelayOverview workNames={Object.fromEntries(snapshot.works.map(work => [work.id, work.name]))} />
    {relayWork && <DetailPanel title={`接力 · ${relayWork.name}`} onClose={() => setRelayWorkID('')}><RelayIntentPanel key={relayWork.id} work={relayWork} onChanged={relayTasks.retry} /></DetailPanel>}
    {modal && <DetailPanel title={modal === 'scan' ? '扫描导入项目' : 'AI 投递 JSON'} onClose={() => setModal('')}>
      {error && <><p role="alert" className={styles.empty}>{error}</p><button type="button" className="ac-text-button" disabled={busy || loading} onClick={() => { void readCache(); }}>重新读取保存结果</button></>}{notice && <p role="status">{notice}</p>}
      {modal === 'scan' ? <>{rootsLoading && <p role="status">正在读取登记项目…</p>}<WorksScanForm roots={roots} busy={disabled || rootsLoading} submit={scan} /><button type="button" className="ac-text-button" disabled={busy || rootsLoading} onClick={() => { void openScan(); }}>重新读取登记项目</button></> : <WorksDeliveryForm busy={disabled} submit={content => perform(async () => (await worksAlbumApi.delivery(content, crypto.randomUUID())).snapshot, 'JSON 投递已保存，文件仍待核实')} />}
    </DetailPanel>}
    {selected && <DetailPanel title={`作品 · ${selected.name}`} onClose={() => setSelectedID('')}><WorksDetail item={selected} /><WorksMetadataForm key={selected.id} item={selected} busy={disabled} save={edit => save(selected, edit)} />{error && <><p role="alert" className={styles.empty}>{error}</p><button type="button" className="ac-text-button" disabled={busy || loading} onClick={() => { void readCache(); }}>重新读取保存结果</button></>}{notice && <p role="status">{notice}</p>}</DetailPanel>}
  </SectionCard>;
}

function WorksDetail({ item }: { item: WorksWork }) {
  return <>
    <Field label="平台"><BrandIcon name={item.platform} size={18} />{platformLabel(item.platform)}</Field>
    <Field label="来源路径"><code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)', overflowWrap: 'anywhere' }}>{item.files.source_path || '未提供'}</code></Field>
    <Field label="文件核实"><span className={`${styles.badge} ${item.files.verified ? '' : styles.unverified}`}>{fileVerifiedLabel(item.files)}</span>{item.files.verified ? ` · 最近改动 ${item.files.mtime ? formatDateTime(item.files.mtime) : '未知'} · ${item.files.file_count} 个文件` : ' · 文件数量未知'}</Field>
    <Field label="态度">{attitudeLabel(item.attitude)}</Field>
    <Field label="分组">{item.group || <span style={{ color: 'var(--color-muted)' }}>未分组</span>}</Field>
    <Field label="交接手册">{item.handoff.content || '尚无交接手册'}</Field>
    <Field label="手册来源">{item.handoff.content ? handoffSourceLabel(item.handoff.source) + (item.handoff.updated_at ? ` · 更新时间 ${formatDateTime(item.handoff.updated_at)}` : '') : '—'}</Field>
  </>;
}
