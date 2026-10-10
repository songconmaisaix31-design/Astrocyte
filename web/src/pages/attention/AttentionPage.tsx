import { useState, useCallback, useRef } from 'react';
import { navigate, useRoute } from '../../router';
import { useMaterials, useOpportunities } from '../../hooks/useReadApi';
import { useDomains } from '../../hooks/useAttention';
import { fixtureMaterials, fixtureOpportunities } from '../../fixtures';
import { SectionCard } from '../../components/SectionCard';
import { StatusBadge } from '../../components/StatusBadge';
import { EmptyState } from '../../components/EmptyState';
import { ErrorState } from '../../components/ErrorState';
import { DetailPanel } from '../../components/DetailPanel';
import { lifecycleLabel, importStatusLabel, opportunityStateLabel, materialKindLabel } from '../../utils/format';
import { PageFrame, IntroCard, RailSummary } from '../../components/PageFrame';
import { matchesQuery } from '../../utils/search';
import { MaterialCover } from '../../components/MaterialCover';
import { BrandIcon } from '../../components/BrandIcon';
import { Icon } from '../../components/DesignIcons';
import { ImportForm } from './ImportForm';
import { useImportDraft } from './useImportDraft';
import { JobsPanel } from './JobsPanel';
import { AccountsPanel } from './AccountsPanel';
import { PaperSearchPanel } from './PaperSearchPanel';
import { PluginSnapshotReview } from './PluginSnapshotReview';
import { DomainsPanel } from './DomainsPanel';
import { SpacesPanel } from './SpacesPanel';
import { RankingPanel } from './RankingPanel';
import { SelectField } from './FormControls';
import { MaterialWorkspace } from './MaterialWorkspace';
import { OpportunityWorkspace, DimensionScores } from './OpportunityWorkspace';
import type { Material, Opportunity } from './model';
import styles from './AttentionPage.module.css';

export function AttentionPage({ fixture, query }: { fixture: boolean; query: string }) {
  const route = useRoute();
  const mat = useMaterials(!fixture);
  const opp = useOpportunities(!fixture);
  const domains = useDomains(!fixture);
  const [selectedMat, setSelectedMat] = useState<Material | null>(null);
  const [selectedMatId, setSelectedMatId] = useState<string | null>(null);
  const [selectedOpp, setSelectedOpp] = useState<Opportunity | null>(null);
  const [importing, setImporting] = useState(false);
  const sourcesPanel = useRef<HTMLDetailsElement>(null);
  const jobsPanel = useRef<HTMLDetailsElement>(null);
  const settingsPanel = useRef<HTMLDetailsElement>(null);
  const paperPanel = useRef<HTMLDetailsElement>(null);
  const pluginPanel = useRef<HTMLDetailsElement>(null);
  const openFlow = (panel: HTMLDetailsElement | null) => { if (panel) { panel.open = true; panel.scrollIntoView({ behavior: 'smooth', block: 'start' }); panel.querySelector<HTMLElement>('summary')?.focus(); } };
  const importDraft = useImportDraft(fixture);
  const [queueToken, setQueueToken] = useState(0);
  const [detailToken, setDetailToken] = useState(0);
  const [kind, setKind] = useState<Material['kind'] | 'all'>('all');
  const [domain, setDomain] = useState('all');
  const allMaterials = fixture ? fixtureMaterials : mat.data?.items ?? [];
  const materials = allMaterials.filter(m => (kind === 'all' || kind === m.kind) && (domain === 'all' || (domain === 'unclassified' ? m.domain_ids?.length === 0 : m.domain_ids?.includes(domain))) && matchesQuery(query, [m.title, m.source_locator, m.collection_reason, m.kind]));
  const opportunities = (fixture ? fixtureOpportunities : opp.data?.items ?? []).filter(o => matchesQuery(query, [o.title, o.id, o.next_step, o.purpose]));
  const saved = materials.filter(m => m.pinned === true);
  const retryMaterials = mat.retry;
  const retryOpportunities = opp.retry;
  const refreshCollections = useCallback(() => { retryMaterials(); retryOpportunities(); setDetailToken(value => value + 1); }, [retryMaterials, retryOpportunities]);
  const handleClose = useCallback(() => { setSelectedMat(null); setSelectedMatId(null); setSelectedOpp(null); setImporting(false); const params = new URLSearchParams(window.location.search); if (params.has('import')) { params.delete('import'); navigate(`/attention${params.size ? `?${params}` : ''}`, true); } }, []);
  const selectOpportunity = (item: Opportunity) => { setSelectedMat(null); setSelectedMatId(null); setSelectedOpp(item); };
  return <PageFrame section="attention" title="资料沉淀" subtitle="每一份关注，都可以长出新的可能。" fixture={fixture} onRefresh={() => { refreshCollections(); domains.retry(); setQueueToken(value => value + 1); }} rail={<>
    <RailSummary title="资料概览" rows={[{ label: '素材', value: mat.loading ? '加载中…' : mat.error && !mat.data ? '无法获取' : materials.length }, { label: '候选机会', value: opp.loading ? '加载中…' : opp.error && !opp.data ? '无法获取' : opportunities.length }]} note="保存资料与候选不会自动启动任务；准入和执行尚未实现。" />
    <RailSummary title="值得继续的问题" rows={opportunities.slice(0, 2).map(o => ({ id: o.id, label: o.title ?? o.id, value: opportunityStateLabel(o.state), onClick: () => selectOpportunity(o) }))} note={opportunities.length ? '查看依据、用途、四维评估、最小下一步与缺失信息。' : '暂无候选机会。'} />
  </>}>
    {!fixture && (mat.stale || opp.stale) && <div className={styles.stale} role="alert">数据可能已过期（刷新失败），显示的是上次成功加载的数据。<button className="ac-button secondary compact" type="button" onClick={refreshCollections}>重试</button></div>}
    <div className={styles.grid}>
      <SectionCard title="" tabs={['overview']}>
        {fixture ? <IntroCard section="attention" /> : <div className="ac-library-actions"><div><h2>把值得继续的线索留在这里。</h2><p>导入一份资料，或从来源更新里挑选。</p></div><button type="button" className="ac-button" onClick={() => setImporting(true)}><Icon name="plus" size={17} />添加资料</button></div>}
        <div className="ac-inline-flows"><button type="button" className="ac-button secondary" onClick={() => openFlow(sourcesPanel.current)}><BrandIcon name="bilibili" size={22} /><BrandIcon name="douyin" size={22} />查看来源更新</button><button type="button" className="ac-button secondary" onClick={() => openFlow(paperPanel.current)}><Icon name="search" size={19} />检索论文</button><button type="button" className="ac-button secondary" onClick={() => openFlow(jobsPanel.current)}><Icon name="layers" size={19} />查看处理队列</button><span>搜索论文，勾选后导入。只处理你选中的资料，正文获取进度在处理队列查看。</span></div>
      </SectionCard>
      <SectionCard tabs={['overview']} title="素材" count={materials.length}>
        <div className="ac-filter-row" role="group" aria-label="素材类型">{(['all', 'paper', 'video', 'text', 'file'] as const).map(value => <button key={value} type="button" className={kind === value ? 'active' : ''} aria-pressed={kind === value} onClick={() => setKind(value)}>{value === 'all' ? '全部资料' : materialKindLabel(value)}</button>)}</div>
        {!fixture && <div className={styles.form}><SelectField label="资料域筛选" value={domain} onChange={setDomain} disabled={domains.loading || domains.stale || !!domains.error} options={[{ value: 'all', label: '所有域' }, { value: 'unclassified', label: '未分类' }, ...(domains.data?.items ?? []).map(item => ({ value: item.id, label: item.title }))]} /></div>}
        {!fixture && mat.loading && !mat.data ? <div className={styles.loadingRow} role="status">加载中…</div> : !fixture && mat.error && !mat.data ? <ErrorState message={mat.error} onRetry={mat.retry} /> : !materials.length && !query && kind === 'all' && domain === 'all' ? <EmptyState icon="📄" title="暂无素材" description="导入文献后，素材将在此显示" /> : <MaterialList items={materials} onSelect={setSelectedMat} />}
      </SectionCard>
      <SectionCard tabs={opportunities.length || opp.error ? ['overview', 'opportunities'] : ['opportunities']} title="机会" count={opportunities.length}>
        {!fixture && opp.loading && !opp.data ? <div className={styles.loadingRow} role="status">加载中…</div> : !fixture && opp.error && !opp.data ? <ErrorState message={opp.error} onRetry={opp.retry} /> : !opportunities.length && !query ? <EmptyState icon="💡" title="暂无机会" description="从资料详情形成候选；自动候选生成尚未配置" /> : <OpportunityList items={opportunities} onSelect={selectOpportunity} />}
      </SectionCard>
      <SectionCard title="我的收藏" tabs={['saved']} padded>
        {fixture ? <><EmptyState title="收藏尚未接入" description="固定样本没有真实收藏状态；不能从采集或使用次数推断收藏。" /><button className="ac-button disabled" type="button" disabled>收藏操作尚未启用</button></> : mat.loading && !mat.data ? <p role="status">加载中…</p> : mat.error && !mat.data ? <ErrorState message={mat.error} onRetry={mat.retry} /> : saved.length ? <MaterialList items={saved} onSelect={setSelectedMat} /> : <EmptyState title="暂无收藏" description="在资料详情中固定到收藏。仅显示服务明确返回的固定项。" />}
      </SectionCard>
      <SectionCard title="来源与整理" tabs={['overview']}><div className="ac-progressive-flows">
        <details ref={sourcesPanel}><summary>查看来源更新</summary><AccountsPanel fixture={fixture} onChanged={() => { refreshCollections(); setQueueToken(value => value + 1); }} onMaterial={id => { setSelectedMat(null); setSelectedMatId(id); }} /></details>
        <details ref={paperPanel}><summary>论文检索</summary><PaperSearchPanel fixture={fixture} onImported={() => { refreshCollections(); setQueueToken(value => value + 1); }} /></details>
        <details ref={pluginPanel}><summary>插件快照复核导入</summary><PluginSnapshotReview fixture={fixture} onImported={() => { refreshCollections(); setQueueToken(value => value + 1); }} /></details>
        <details ref={jobsPanel}><summary>查看处理队列</summary><JobsPanel fixture={fixture} refreshToken={queueToken} onChanged={refreshCollections} onMaterial={id => { setSelectedMat(null); setSelectedMatId(id); }} /></details>
        <details ref={settingsPanel}><summary>管理资料域、排序与空间引用</summary><div className={styles.grid}>
          <DomainsPanel state={domains} fixture={fixture} />
          <RankingPanel fixture={fixture} refreshToken={queueToken} onChanged={refreshCollections} />
          <SpacesPanel fixture={fixture} materials={allMaterials} refreshToken={queueToken} onChanged={refreshCollections} />
        </div></details>
      </div></SectionCard>
    </div>
    {(importing || route.query.get('import') === '1') && <DetailPanel title="添加资料" onClose={handleClose}><ImportForm fixture={fixture} draft={importDraft} onImported={() => { refreshCollections(); setQueueToken(value => value + 1); }} /></DetailPanel>}
    {(selectedMat || selectedMatId) && <DetailPanel title="素材详情" onClose={handleClose} disabledActions={fixture ? ['继续沉淀', '以后再看'] : []} showDisabledNotice disabledNoticeText={fixture ? '示例模式：所有写操作尚未启用，请退出示例模式使用真实 API' : '资料准入和任务执行尚未实现，保存不会创建 Mission'}><MaterialWorkspace key={selectedMat?.id ?? selectedMatId!} id={selectedMat?.id ?? selectedMatId!} item={selectedMat ?? undefined} fixture={fixture} refreshToken={detailToken} materials={allMaterials} domains={domains} onChanged={refreshCollections} onQueued={() => { refreshCollections(); setQueueToken(value => value + 1); }} onOpportunity={selectOpportunity} /></DetailPanel>}
    {selectedOpp && <DetailPanel title="机会详情" onClose={handleClose} disabledActions={fixture ? ['拒绝', '准入'] : ['准入', '任务执行']} showDisabledNotice disabledNoticeText={fixture ? '示例模式：所有写操作尚未启用，请退出示例模式使用真实 API' : '准入和任务执行尚未实现；人工反馈不会批准或启动任务'}><OpportunityWorkspace key={selectedOpp.id} item={selectedOpp} fixture={fixture} materials={allMaterials} onChanged={refreshCollections} /></DetailPanel>}
  </PageFrame>;
}

function MaterialList({ items, onSelect }: { items: Material[]; onSelect: (m: Material) => void }) {
  if (!items.length) return <EmptyState title="没有匹配的素材" description="尝试其他搜索词或资料类型" />;
  return <ul className={`${styles.list} ac-material-grid`} role="list">{items.map(m => <li key={m.id} className={`${styles.listItem} ac-material-card`} tabIndex={0} role="button" onClick={() => onSelect(m)} onKeyDown={event => { if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); onSelect(m); } }}>
    <MaterialCover kind={m.kind} /><div className="ac-material-body"><div className={styles.itemHeader}><span className={`${styles.itemTitle} line-clamp-2`}>{m.title || m.source_locator}</span><div className={styles.actions}><StatusBadge value={m.lifecycle} label={lifecycleLabel(m.lifecycle)} />{m.import_status && <StatusBadge value={m.import_status} label={importStatusLabel(m.import_status)} />}</div></div>
    <div className={styles.itemMeta}><span>{materialKindLabel(m.kind)}</span><span>{m.source_locator}</span><span>v{m.current_revision}</span></div><p>收藏理由 · {m.collection_reason?.trim() || '未提供'}</p><p className={styles.note}>人类关注 {m.human_usage_count ?? '未知'} · 机器使用 {m.agent_usage_count ?? '未知'}</p>{m.ranking_reason && <p className={styles.note}>排序依据 · {m.ranking_reason}</p>}</div>
  </li>)}</ul>;
}
function OpportunityList({ items, onSelect }: { items: Opportunity[]; onSelect: (o: Opportunity) => void }) {
  if (!items.length) return <EmptyState title="没有匹配的机会" description="尝试其他搜索词" />;
  return <ul className={styles.list} role="list">{items.map(o => <li key={o.id} className={`${styles.listItem} ac-opportunity`} tabIndex={0} role="button" onClick={() => onSelect(o)} onKeyDown={event => { if (event.key === 'Enter' || event.key === ' ') { event.preventDefault(); onSelect(o); } }}>
    <span className="ac-op-icon"><Icon name="spark" size={19} /></span><div className="ac-op-body"><div className={styles.itemHeader}><span className={`${styles.itemTitle} line-clamp-2`}>{o.title || o.id}</span><StatusBadge value={o.state} label={opportunityStateLabel(o.state)} /></div>
    <div className={styles.itemMeta}><span>{o.evidence_refs.length} 条证据</span><span>r{o.revision}</span></div><p>用途 · {o.purpose || '未提供'}</p><p>{o.missing_evidence?.length ? `缺失依据：${o.missing_evidence.join('；')}` : '缺失依据未记录'}</p><DimensionScores value={o.dimensions} />{o.ranking_reason && <p className={styles.note}>排序依据 · {o.ranking_reason} · 综合分 {o.composite_score ?? '未知'} · 配置版本 {o.ranking_profile_version ?? '未提供'}</p>}<div className="ac-op-footer"><span>最小下一步 · {o.next_step || '未提供'}</span><span className="ac-text-button">查看依据<Icon name="arrow" size={14} /></span></div></div>
  </li>)}</ul>;
}
