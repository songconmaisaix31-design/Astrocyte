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
import styles from './AttentionPage.module.css';

type Material = components['schemas']['MaterialV1'];
type Opportunity = components['schemas']['OpportunityV1'];

interface Props {
  fixture: boolean;
}

export function AttentionPage({ fixture }: Props) {
  const mat = useMaterials();
  const opp = useOpportunities();
  const [selectedMat, setSelectedMat] = useState<Material | null>(null);
  const [selectedOpp, setSelectedOpp] = useState<Opportunity | null>(null);

  const materials = fixture ? fixtureMaterials : (mat.data?.items ?? []);
  const opportunities = fixture ? fixtureOpportunities : (opp.data?.items ?? []);

  const handleClose = useCallback(() => {
    setSelectedMat(null);
    setSelectedOpp(null);
  }, []);

  return (
    <div>
      <div className={styles.pageHeader}>
        <h1 className={styles.pageTitle}>注意力</h1>
        <p className={styles.pageSub}>追踪文献素材与研究机会</p>
      </div>

      <div className={styles.grid}>
        {/* ── Materials Section ── */}
        <SectionCard title="素材" count={materials.length}>
          {fixture ? (
            <MaterialList items={materials} onSelect={setSelectedMat} />
          ) : mat.loading ? (
            <div className={styles.loadingRow}>加载中…</div>
          ) : mat.error ? (
            <ErrorState message={mat.error} onRetry={mat.retry} />
          ) : materials.length === 0 ? (
            <EmptyState icon="📄" title="暂无素材" description="导入文献后，素材将在此显示" />
          ) : (
            <MaterialList items={materials} onSelect={setSelectedMat} />
          )}
        </SectionCard>

        {/* ── Opportunities Section ── */}
        <SectionCard title="机会" count={opportunities.length}>
          {fixture ? (
            <OpportunityList items={opportunities} onSelect={setSelectedOpp} />
          ) : opp.loading ? (
            <div className={styles.loadingRow}>加载中…</div>
          ) : opp.error ? (
            <ErrorState message={opp.error} onRetry={opp.retry} />
          ) : opportunities.length === 0 ? (
            <EmptyState icon="💡" title="暂无机会" description="系统识别研究机会后将在此显示" />
          ) : (
            <OpportunityList items={opportunities} onSelect={setSelectedOpp} />
          )}
        </SectionCard>
      </div>

      {/* ── Detail Panels ── */}
      {selectedMat && (
        <DetailPanel
          title="素材详情"
          onClose={handleClose}
          showDisabledNotice
          disabledNoticeText="导入与审核操作将在后续切片中启用"
        >
          <MaterialDetail item={selectedMat} isFixture={fixture} />
        </DetailPanel>
      )}
      {selectedOpp && (
        <DetailPanel
          title="机会详情"
          onClose={handleClose}
          showDisabledNotice
          disabledNoticeText="审核与批准操作将在后续切片中启用"
        >
          <OpportunityDetail item={selectedOpp} isFixture={fixture} />
        </DetailPanel>
      )}
    </div>
  );
}

function MaterialList({ items, onSelect }: { items: Material[]; onSelect: (m: Material) => void }) {
  return (
    <ul className={styles.list} role="list">
      {items.map((m) => (
        <li
          key={m.id}
          className={styles.listItem}
          tabIndex={0}
          role="button"
          onClick={() => onSelect(m)}
          onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(m); }}}
        >
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
        </li>
      ))}
    </ul>
  );
}

function OpportunityList({ items, onSelect }: { items: Opportunity[]; onSelect: (o: Opportunity) => void }) {
  return (
    <ul className={styles.list} role="list">
      {items.map((o) => (
        <li
          key={o.id}
          className={styles.listItem}
          tabIndex={0}
          role="button"
          onClick={() => onSelect(o)}
          onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(o); }}}
        >
          <div className={styles.itemHeader}>
            <span className={`${styles.itemTitle} line-clamp-2`}>{o.title ?? o.id}</span>
            <StatusBadge value={o.state} label={opportunityStateLabel(o.state)} />
          </div>
          <div className={styles.itemMeta}>
            <span>{o.evidence_refs.length} 条证据</span>
            <span>r{o.revision}</span>
            <span className={styles.itemDesc} style={{ flex: 1 }}>{o.next_step}</span>
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
        <Field label="人工使用">{item.human_usage_count ?? 0} 次</Field>
        <Field label="Agent 使用">{item.agent_usage_count ?? 0} 次</Field>
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
