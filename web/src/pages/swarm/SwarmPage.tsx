import { useState, useCallback, useMemo } from 'react';
import { useMissions } from '../../hooks/useReadApi';
import { fixtureMissions, flattenWorkItems, flattenArtifactIds } from '../../fixtures';
import { SectionCard } from '../../components/SectionCard';
import { StatusBadge } from '../../components/StatusBadge';
import { EmptyState } from '../../components/EmptyState';
import { ErrorState } from '../../components/ErrorState';
import { DetailPanel, Field, FieldRow, MutedValue } from '../../components/DetailPanel';
import { missionStatusLabel, workItemStatusLabel } from '../../utils/format';
import type { components } from '../../api/schema';
import { PageFrame, IntroCard, RailSummary } from '../../components/PageFrame';
import { DesignGraph, DesignTimeline } from '../../components/DesignExamples';
import { Icon } from '../../components/DesignIcons';
import { matchesQuery } from '../../utils/search';
import styles from './SwarmPage.module.css';

type Mission = components['schemas']['MissionV1'];
type WorkItem = components['schemas']['WorkItemV1'];

/** Stable empty array to avoid changing useMemo deps on every render. */
const EMPTY_MISSIONS: Mission[] = [];

interface Props {
  fixture: boolean;
  query: string;
}

type SelectedItem =
  | { kind: 'mission'; data: Mission }
  | { kind: 'workitem'; data: WorkItem; missionId: string }
  | null;

export function SwarmPage({ fixture, query }: Props) {
  const miss = useMissions();
  const [selected, setSelected] = useState<SelectedItem>(null);

  const missions = useMemo(
    () => (fixture ? fixtureMissions : (miss.data?.items ?? EMPTY_MISSIONS)).filter(m => matchesQuery(query, [m.goal, m.id, ...(m.work_items ?? []).map(w => w.id), ...(m.artifact_ids ?? [])])),
    [fixture, miss.data, query],
  );
  const workItems = useMemo(() => flattenWorkItems(missions), [missions]);
  const artifactIds = useMemo(() => flattenArtifactIds(missions), [missions]);

  const handleClose = useCallback(() => setSelected(null), []);

  return (
    <PageFrame section="swarm" title="蜂群空间" subtitle="先把项目放进开发空间，再由你决定下一步。" fixture={fixture} onRefresh={() => { miss.retry(); }} rail={<RailSummary title="执行概览" rows={[{ label: '任务', value: fixture ? missions.length : miss.loading ? '加载中…' : miss.error && !miss.data ? '无法获取' : missions.length }, { label: '工作项', value: fixture ? workItems.length : miss.loading ? '加载中…' : miss.error && !miss.data ? '无法获取' : workItems.length }, { label: '产物引用', value: fixture ? artifactIds.length : miss.loading ? '加载中…' : miss.error && !miss.data ? '无法获取' : artifactIds.length }]} note="产物引用不等于已采用；暂停、取消与采用尚未启用。" />}>
      {(!fixture && miss.stale) && (
        <div role="alert" style={{
          padding: 'var(--space-3) var(--space-5)', background: 'var(--color-warning-subtle)',
          borderBottom: '1px solid var(--color-warning)', fontSize: 'var(--text-sm)',
          color: '#92610a', display: 'flex', justifyContent: 'space-between', alignItems: 'center',
        }}>
          <span>⚠ 数据可能已过期（刷新失败），显示的是上次成功加载的数据</span>
          <button type="button" onClick={() => miss.retry()}
            style={{ padding: 'var(--space-1) var(--space-3)', fontSize: 'var(--text-xs)',
              border: '1px solid var(--color-warning)', borderRadius: 'var(--radius-sm)',
              background: 'transparent', cursor: 'pointer' }}>重试</button>
        </div>
      )}

      <div className={styles.grid}>
        <SectionCard title="" tabs={['overview']}><IntroCard section="swarm" /></SectionCard>
        <SectionCard title="协作视图" tabs={['overview']}><DesignGraph fixture={fixture} /></SectionCard>
        <SectionCard title="任务时间线" tabs={['timeline']}><DesignTimeline fixture={fixture} /></SectionCard>
        {/* ── Missions ── */}
        <SectionCard tabs={['overview', 'timeline']} title="任务" count={missions.length}>
          {fixture ? (
            <MissionList items={missions} onSelect={(m) => setSelected({ kind: 'mission', data: m })} />
          ) : miss.loading && !miss.data ? (
            <div className={styles.loadingRow}>加载中…</div>
          ) : miss.error && !miss.data ? (
            <ErrorState message={miss.error} onRetry={miss.retry} />
          ) : missions.length === 0 ? (
            <EmptyState icon="🎯" title="暂无任务" description="暂无任务" />
          ) : (
            <MissionList items={missions} onSelect={(m) => setSelected({ kind: 'mission', data: m })} />
          )}
        </SectionCard>

        {/* ── Work Items (flattened from missions — propagate mission state) ── */}
        <SectionCard tabs={['overview', 'timeline']} title="工作项" count={workItems.length}>
          {fixture ? (
            workItems.length === 0 ? (
              <EmptyState icon="⚙" title="暂无工作项" description="示例任务无工作项" />
            ) : (
              <WorkItemList
                items={workItems}
                missions={missions}
                onSelect={(w, mid) => setSelected({ kind: 'workitem', data: w, missionId: mid })}
              />
            )
          ) : miss.loading && !miss.data ? (
            <div className={styles.loadingRow}>加载中…</div>
          ) : miss.error && !miss.data ? (
            <ErrorState message={miss.error} onRetry={miss.retry} />
          ) : workItems.length === 0 ? (
            <EmptyState icon="⚙" title="暂无工作项" description="工作项由任务分解产生" />
          ) : (
            <WorkItemList
              items={workItems}
              missions={missions}
              onSelect={(w, mid) => setSelected({ kind: 'workitem', data: w, missionId: mid })}
            />
          )}
        </SectionCard>

        {/* ── Artifacts (IDs from missions — propagate mission state) ── */}
        <SectionCard tabs={['overview', 'artifacts']} title="产物引用" count={artifactIds.length}>
          {fixture ? (
            artifactIds.length === 0 ? (
              <EmptyState icon="📦" title="暂无产物引用" description="示例任务无产物" />
            ) : (
              <ul className={`${styles.list} ac-artifact-list`} role="list">
                {artifactIds.map((id) => (
                  <li key={id} className={`${styles.listItem} ac-artifact-card`} style={{ cursor: 'default' }}><span className="ac-artifact-icon blue"><Icon name="layers" size={24} /></span>
                    <div className={styles.itemHeader}>
                      <span className={styles.itemTitle}>
                        <code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)' }}>{id}</code>
                      </span>
                    </div>
                  </li>
                ))}
              </ul>
            )
          ) : miss.loading && !miss.data ? (
            <div className={styles.loadingRow}>加载中…</div>
          ) : miss.error && !miss.data ? (
            <ErrorState message={miss.error} onRetry={miss.retry} />
          ) : artifactIds.length === 0 ? (
            <EmptyState icon="📦" title="暂无产物引用" description="工作项完成后产物引用将在此显示" />
          ) : (
            <ul className={`${styles.list} ac-artifact-list`} role="list">
              {artifactIds.map((id) => (
                <li key={id} className={`${styles.listItem} ac-artifact-card`} style={{ cursor: 'default' }}><span className="ac-artifact-icon blue"><Icon name="layers" size={24} /></span>
                  <div className={styles.itemHeader}>
                    <span className={styles.itemTitle}>
                      <code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)' }}>{id}</code>
                    </span>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </SectionCard>
      </div>

      {/* ── Detail Panels ── */}
      {selected && (
        <DetailPanel
          title={selected.kind === 'mission' ? '任务详情' : '工作项详情'}
          onClose={handleClose}
          disabledActions={selected.kind === 'mission' ? ['暂停', '取消'] : ['查看 diff', '采用', '退回']}
          showDisabledNotice
          disabledNoticeText="任务创建与执行操作尚未启用"
        >
          {selected.kind === 'mission' && <MissionDetail item={selected.data} isFixture={fixture} />}
          {selected.kind === 'workitem' && <WorkItemDetail item={selected.data} missionId={selected.missionId} isFixture={fixture} />}
        </DetailPanel>
      )}
    </PageFrame>
  );
}

// ── Lists ──

function MissionList({ items, onSelect }: { items: Mission[]; onSelect: (m: Mission) => void }) {
  return (
    <ul className={styles.list} role="list">
      {items.map((m) => {
        const wiCount = m.work_items?.length ?? 0;
        const artCount = m.artifact_ids?.length ?? 0;
        return (
          <li key={m.id} className={`${styles.listItem} ac-mission-card`} tabIndex={0} role="button"
            onClick={() => onSelect(m)}
            onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(m); }}}>
            <div className={styles.itemHeader}>
              <span className={`${styles.itemTitle} line-clamp-2`}>{m.goal}</span>
              <StatusBadge value={m.status} label={missionStatusLabel(m.status)} />
            </div>
            <div className={styles.itemMeta}>
              <span>v{m.version}</span>
              <span>Grant: {m.grant_id}</span>
              <span>{wiCount} 工作项</span>
              <span>{artCount} 产物</span>
            </div>
            {m.blockers && m.blockers.length > 0 && (
              <div className={styles.blockerRow}>
                ⚠ {m.blockers.join('; ')}
              </div>
            )}
          </li>
        );
      })}
    </ul>
  );
}

function WorkItemList({
  items,
  missions,
  onSelect,
}: {
  items: WorkItem[];
  missions: Mission[];
  onSelect: (w: WorkItem, missionId: string) => void;
}) {
  // Build a map of work_item_id → mission_id
  const wiToMission = useMemo(() => {
    const map = new Map<string, string>();
    for (const m of missions) {
      for (const wi of m.work_items ?? []) {
        map.set(wi.id, m.id);
      }
    }
    return map;
  }, [missions]);

  return (
    <ul className={styles.list} role="list">
      {items.map((w) => (
        <li key={w.id} className={styles.listItem} tabIndex={0} role="button"
          onClick={() => onSelect(w, wiToMission.get(w.id) ?? '')}
          onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(w, wiToMission.get(w.id) ?? ''); }}}>
          <div className={styles.itemHeader}>
            <span className={styles.itemTitle}>
              <code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)' }}>{w.id}</code>
            </span>
            <StatusBadge value={w.status} label={workItemStatusLabel(w.status)} />
          </div>
          <div className={styles.itemMeta}>
            <span>v{w.version}</span>
            <span>持有者: {w.holder_id ?? '未认领'}</span>
            <span>{w.artifact_ids.length} 产物</span>
          </div>
        </li>
      ))}
    </ul>
  );
}

// ── Details ──

function MissionDetail({ item, isFixture }: { item: Mission; isFixture: boolean }) {
  return (
    <>
      <Field label="目标">{item.goal}</Field>
      {isFixture && <FixtureTag />}
      <FieldRow>
        <Field label="状态"><StatusBadge value={item.status} label={missionStatusLabel(item.status)} /></Field>
        <Field label="版本">v{item.version}</Field>
      </FieldRow>
      <Field label="Grant ID">
        <code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)' }}>{item.grant_id}</code>
      </Field>
      <Field label="工作项">
        {(item.work_items?.length ?? 0) > 0 ? (
          <div style={{ fontSize: 'var(--text-sm)' }}>
            {item.work_items!.map((wi) => (
              <div key={wi.id} style={{ display: 'flex', justifyContent: 'space-between', padding: 'var(--space-1) 0', borderBottom: '1px solid var(--color-border-light)' }}>
                <code style={{ fontFamily: 'var(--font-mono)' }}>{wi.id}</code>
                <StatusBadge value={wi.status} label={workItemStatusLabel(wi.status)} />
              </div>
            ))}
          </div>
        ) : <MutedValue>无工作项</MutedValue>}
      </Field>
      <Field label="产物引用">
        {(item.artifact_ids?.length ?? 0) > 0
          ? item.artifact_ids!.join(', ')
          : <MutedValue>无</MutedValue>}
      </Field>
      <Field label="阻塞原因">
        {item.blockers && item.blockers.length > 0
          ? item.blockers.join('; ')
          : <MutedValue>无</MutedValue>}
      </Field>
      <Field label="ID"><code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xs)' }}>{item.id}</code></Field>
    </>
  );
}

function WorkItemDetail({ item, missionId, isFixture }: { item: WorkItem; missionId: string; isFixture: boolean }) {
  return (
    <>
      {isFixture && <FixtureTag />}
      <FieldRow>
        <Field label="状态"><StatusBadge value={item.status} label={workItemStatusLabel(item.status)} /></Field>
        <Field label="版本">v{item.version}</Field>
      </FieldRow>
      <Field label="所属任务">
        <code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)' }}>{missionId}</code>
      </Field>
      <Field label="持有者">{item.holder_id ?? <MutedValue>未认领</MutedValue>}</Field>
      <Field label="上下文包 ID">{item.context_packet_id ?? <MutedValue>无</MutedValue>}</Field>
      <Field label="产物引用">
        {item.artifact_ids.length > 0
          ? item.artifact_ids.join(', ')
          : <MutedValue>无</MutedValue>}
      </Field>
      <Field label="ID"><code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xs)' }}>{item.id}</code></Field>
    </>
  );
}

// ── Helpers ──

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
