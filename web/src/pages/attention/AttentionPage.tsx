import { useState, useCallback } from 'react';
import { useMaterials, useOpportunities } from '../../hooks/useReadApi';
import { fixtureMaterials, fixtureOpportunities } from '../../fixtures';
import { SectionCard } from '../../components/SectionCard';
import { StatusBadge } from '../../components/StatusBadge';
import { EmptyState } from '../../components/EmptyState';
import { ErrorState } from '../../components/ErrorState';
import { DetailPanel, Field, FieldRow, MutedValue } from '../../components/DetailPanel';
import { formatDimScore, lifecycleLabel, importStatusLabel, opportunityStateLabel, materialKindLabel } from '../../utils/format';
import type { components } from '../../api/schema';
import { PageFrame, IntroCard, RailSummary } from '../../components/PageFrame';
import { matchesQuery } from '../../utils/search';
import { MaterialCover } from '../../components/MaterialCover';
import { Icon } from '../../components/DesignIcons';
import styles from './AttentionPage.module.css';

type Material = components['schemas']['MaterialV1'];
type Opportunity = components['schemas']['OpportunityV1'];

interface Props {
  fixture: boolean;
  query: string;
}

export function AttentionPage({ fixture, query }: Props) {
  const mat = useMaterials();
  const opp = useOpportunities();
  const [selectedMat, setSelectedMat] = useState<Material | null>(null);
  const [selectedOpp, setSelectedOpp] = useState<Opportunity | null>(null);
  const [kind, setKind] = useState<Material['kind'] | 'all'>('all');

  const materials = (fixture ? fixtureMaterials : (mat.data?.items ?? [])).filter(m => (kind === 'all' || kind === m.kind) && matchesQuery(query, [m.title, m.source_locator, m.kind]));
  const opportunities = (fixture ? fixtureOpportunities : (opp.data?.items ?? [])).filter(o => matchesQuery(query, [o.title, o.id, o.next_step]));

  const handleClose = useCallback(() => {
    setSelectedMat(null);
    setSelectedOpp(null);
  }, []);

  return (
    <PageFrame section="attention" title="资料沉淀" subtitle="每一份关注，都可以长出新的可能。" fixture={fixture} onRefresh={() => { mat.retry(); opp.retry(); }} rail={<><RailSummary title="资料概览" rows={[{ label: '素材', value: fixture ? materials.length : mat.loading ? '加载中…' : mat.error && !mat.data ? '无法获取' : materials.length }, { label: '候选机会', value: fixture ? opportunities.length : opp.loading ? '加载中…' : opp.error && !opp.data ? '无法获取' : opportunities.length }]} note="保存资料不会自动启动任务；候选仍需人工准入。" /><RailSummary title="值得继续的问题" rows={opportunities.slice(0, 2).map(o => ({ id: o.id, label: o.title ?? o.id, value: opportunityStateLabel(o.state), onClick: () => setSelectedOpp(o) }))} note={opportunities.length ? '点击查看四维评分、最小下一步与缺失依据。' : '暂无候选机会。'} /></>}>
      {(!fixture && (mat.stale || opp.stale)) && (
        <div role="alert" style={{
          padding: 'var(--space-3) var(--space-5)', background: 'var(--color-warning-subtle)',
          borderBottom: '1px solid var(--color-warning)', fontSize: 'var(--text-sm)',
          color: '#92610a', display: 'flex', justifyContent: 'space-between', alignItems: 'center',
        }}>
          <span>⚠ 数据可能已过期（刷新失败），显示的是上次成功加载的数据</span>
          <button type="button" onClick={() => { mat.retry(); opp.retry(); }}
            style={{ padding: 'var(--space-1) var(--space-3)', fontSize: 'var(--text-xs)',
              border: '1px solid var(--color-warning)', borderRadius: 'var(--radius-sm)',
              background: 'transparent', cursor: 'pointer' }}>重试</button>
        </div>
      )}

      <div className={styles.grid}>
        <SectionCard title="" tabs={['overview']}><IntroCard section="attention" /></SectionCard>
        {/* ── Materials Section ── */}
        <SectionCard tabs={['overview']} title="素材" count={materials.length}>
          <div className="ac-filter-row" role="group" aria-label="素材类型">{(['all', 'paper', 'video', 'text', 'file'] as const).map(value => <button key={value} type="button" className={kind === value ? 'active' : ''} aria-pressed={kind === value} onClick={() => setKind(value)}>{value === 'all' ? '全部资料' : materialKindLabel(value)}</button>)}</div>
          {fixture ? (
            <MaterialList items={materials} onSelect={setSelectedMat} />
          ) : mat.loading && !mat.data ? (
            <div className={styles.loadingRow}>加载中…</div>
          ) : mat.error && !mat.data ? (
            <ErrorState message={mat.error} onRetry={mat.retry} />
          ) : materials.length === 0 ? (
            <EmptyState icon="📄" title="暂无素材" description="导入文献后，素材将在此显示" />
          ) : (
            <MaterialList items={materials} onSelect={setSelectedMat} />
          )}
        </SectionCard>

        {/* ── Opportunities Section ── */}
        <SectionCard tabs={['overview', 'opportunities']} title="机会" count={opportunities.length}>
          {fixture ? (
            <OpportunityList items={opportunities} onSelect={setSelectedOpp} />
          ) : opp.loading && !opp.data ? (
            <div className={styles.loadingRow}>加载中…</div>
          ) : opp.error && !opp.data ? (
            <ErrorState message={opp.error} onRetry={opp.retry} />
          ) : opportunities.length === 0 ? (
            <EmptyState icon="💡" title="暂无机会" description="系统识别研究机会后将在此显示" />
          ) : (
            <OpportunityList items={opportunities} onSelect={setSelectedOpp} />
          )}
        </SectionCard>
      </div>

      {/* ── Detail Panels ── */}
      <SectionCard title="我的收藏" tabs={['saved']} padded><EmptyState icon="☆" title="收藏尚未接入" description="当前接口不提供收藏状态，不能从采集或使用次数推断收藏。" /><button className="ac-button disabled" type="button" disabled title="收藏接口尚未实现">收藏操作尚未启用</button></SectionCard>

      {selectedMat && (
        <DetailPanel
          title="素材详情"
          disabledActions={['继续沉淀', '以后再看']}
          onClose={handleClose}
          showDisabledNotice
          disabledNoticeText="导入与审核操作尚未启用"
        >
          <MaterialDetail item={selectedMat} isFixture={fixture} />
        </DetailPanel>
      )}
      {selectedOpp && (
        <DetailPanel
          title="机会详情"
          disabledActions={['拒绝', '准入']}
          onClose={handleClose}
          showDisabledNotice
          disabledNoticeText="审核与批准操作尚未启用"
        >
          <OpportunityDetail item={selectedOpp} isFixture={fixture} />
        </DetailPanel>
      )}
    </PageFrame>
  );
}

function MaterialList({ items, onSelect }: { items: Material[]; onSelect: (m: Material) => void }) {
  if (!items.length) return <EmptyState title="没有匹配的素材" description="尝试其他搜索词或资料类型" />;
  return (
    <ul className={`${styles.list} ac-material-grid`} role="list">
      {items.map((m) => (
        <li
          key={m.id}
          className={`${styles.listItem} ac-material-card`}
          tabIndex={0}
          role="button"
          onClick={() => onSelect(m)}
          onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(m); }}}
        >
          <MaterialCover kind={m.kind} />
          <div className="ac-material-body">
          <div className={styles.itemHeader}>
            <span className={`${styles.itemTitle} line-clamp-2`}>{m.title ?? m.source_locator}</span>
            <div style={{ display: 'flex', gap: 'var(--space-2)', flexShrink: 0 }}>
              <StatusBadge value={m.lifecycle} label={lifecycleLabel(m.lifecycle)} />
              {m.import_status && <StatusBadge value={m.import_status} label={importStatusLabel(m.import_status)} />}
            </div>
          </div>
          <div className={styles.itemMeta}>
            <span>{materialKindLabel(m.kind)}</span>
            <span>{m.source_locator}</span>
            <span>v{m.current_revision}</span>
          </div>
          <p>{m.collection_reason ?? '采集原因未记录'}</p>
          </div>
        </li>
      ))}
    </ul>
  );
}

function OpportunityList({ items, onSelect }: { items: Opportunity[]; onSelect: (o: Opportunity) => void }) {
  if (!items.length) return <EmptyState title="没有匹配的机会" description="尝试其他搜索词" />;
  return (
    <ul className={styles.list} role="list">
      {items.map((o) => (
        <li
          key={o.id}
          className={`${styles.listItem} ac-opportunity`}
          tabIndex={0}
          role="button"
          onClick={() => onSelect(o)}
          onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(o); }}}
        >
          <span className="ac-op-icon"><Icon name="spark" size={19} /></span><div className="ac-op-body">
          <div className={styles.itemHeader}>
            <span className={`${styles.itemTitle} line-clamp-2`}>{o.title ?? o.id}</span>
            <StatusBadge value={o.state} label={opportunityStateLabel(o.state)} />
          </div>
          <div className={styles.itemMeta}>
            <span>{o.evidence_refs.length} 条证据</span>
            <span>r{o.revision}</span>
          </div>
          <p>{o.missing_evidence?.length ? `缺失依据：${o.missing_evidence!.join('；')}` : '缺失依据未记录'}</p>
          <div className="ac-op-footer"><span>最小下一步 · {o.next_step}</span><span className="ac-text-button">查看依据<Icon name="arrow" size={14} /></span></div>
          </div>
        </li>
      ))}
    </ul>
  );
}

function MaterialDetail({ item, isFixture }: { item: Material; isFixture: boolean }) {
  return (
    <>
      <Field label="标题">{item.title ?? <MutedValue>无标题</MutedValue>}</Field>
      {isFixture && <FixtureTag />}
      <FieldRow>
        <Field label="生命周期"><StatusBadge value={item.lifecycle} label={lifecycleLabel(item.lifecycle)} /></Field>
        <Field label="类型">{materialKindLabel(item.kind)}</Field>
      </FieldRow>
      {item.import_status && (
        <Field label="导入状态"><StatusBadge value={item.import_status} label={importStatusLabel(item.import_status)} /></Field>
      )}
      <Field label="来源定位">
        <code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)' }}>{item.source_locator}</code>
      </Field>
      <Field label="采集原因">{item.collection_reason ?? <MutedValue>无</MutedValue>}</Field>
      <Field label="来源片段">
        {item.source_spans?.length
          ? item.source_spans.join(', ')
          : <MutedValue>无</MutedValue>}
      </Field>
      <FieldRow>
        <Field label="当前版本">v{item.current_revision}</Field>
        <Field label="ID"><code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xs)' }}>{item.id}</code></Field>
      </FieldRow>
      <FieldRow>
        <Field label="人工使用">{item.human_usage_count != null ? `${item.human_usage_count} 次` : <MutedValue>未知</MutedValue>}</Field>
        <Field label="Agent 使用">{item.agent_usage_count != null ? `${item.agent_usage_count} 次` : <MutedValue>未知</MutedValue>}</Field>
      </FieldRow>
    </>
  );
}

function OpportunityDetail({ item, isFixture }: { item: Opportunity; isFixture: boolean }) {
  const dims = item.dimensions;
  return (
    <>
      <Field label="标题">{item.title ?? <MutedValue>无标题</MutedValue>}</Field>
      {isFixture && <FixtureTag />}
      <Field label="状态"><StatusBadge value={item.state} label={opportunityStateLabel(item.state)} /></Field>
      <Field label="下一步">{item.next_step}</Field>

      <Field label="评估维度">
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 'var(--space-2)', fontSize: 'var(--text-sm)' }}>
          <DimRow label="目标进展" value={dims.goal_progress} />
          <DimRow label="当前兴趣" value={dims.current_interest} />
          <DimRow label="项目改善" value={dims.project_improvement} />
          <DimRow label="创新性" value={dims.originality} />
        </div>
      </Field>

      <Field label="证据引用">
        {item.evidence_refs.length > 0
          ? item.evidence_refs.map(r => (
              <div key={r.material_id} style={{ fontSize: 'var(--text-sm)', fontFamily: 'var(--font-mono)' }}>
                {r.locator}{r.span ? ` (${r.span})` : ''}
              </div>
            ))
          : <MutedValue>无证据</MutedValue>}
      </Field>

      <Field label="缺失证据">
        {item.missing_evidence?.length
          ? item.missing_evidence.join(', ')
          : <MutedValue>无</MutedValue>}
      </Field>

      <FieldRow>
        <Field label="版本">r{item.revision}</Field>
        <Field label="ID"><code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xs)' }}>{item.id}</code></Field>
      </FieldRow>
    </>
  );
}

function DimRow({ label, value }: { label: string; value: { value: number | null; reason: string } }) {
  return (
    <div>
      <span style={{ color: 'var(--color-text-muted)', fontSize: 'var(--text-xs)' }}>{label}</span>
      <div>{formatDimScore(value.value)} <span style={{ color: 'var(--color-text-muted)', fontSize: 'var(--text-xs)' }}>({value.reason})</span></div>
    </div>
  );
}

function FixtureTag() {
  return (
    <div style={{
      display: 'inline-block', padding: '2px 8px', fontSize: 'var(--text-xs)',
      fontWeight: 'var(--weight-medium)', background: 'var(--color-warning-subtle)',
      color: '#e67700', borderRadius: 'var(--radius-sm)', border: '1px solid var(--color-warning)',
      marginBottom: 'var(--space-3)',
    }}>⚑ 示例数据</div>
  );
}
