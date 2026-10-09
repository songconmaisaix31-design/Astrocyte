import { useState, useCallback } from 'react';
import { useProjects, useProposals, useSessions } from '../../hooks/useReadApi';
import { fixtureProjects, fixtureProposals, fixtureSessions } from '../../fixtures';
import { SectionCard } from '../../components/SectionCard';
import { StatusBadge } from '../../components/StatusBadge';
import { EmptyState } from '../../components/EmptyState';
import { ErrorState } from '../../components/ErrorState';
import { DetailPanel, Field, FieldRow, MutedValue } from '../../components/DetailPanel';
import { proposalStatusLabel, bindingStatusLabel, contextStateLabel, formatBudget, capabilityLabel } from '../../utils/format';
import type { components } from '../../api/schema';
import { PageFrame, IntroCard, RailSummary } from '../../components/PageFrame';
import { ResearchRoutes, DesignTimeline } from '../../components/DesignExamples';
import { matchesQuery } from '../../utils/search';
import { Icon, NetworkArt } from '../../components/DesignIcons';
import styles from './WorkspacePage.module.css';

type Project = components['schemas']['ProjectV1'];
type Proposal = components['schemas']['ProposalV1'];
type Session = components['schemas']['SessionV1'];

interface Props {
  fixture: boolean;
  query: string;
}

type SelectedItem =
  | { kind: 'project'; data: Project }
  | { kind: 'proposal'; data: Proposal }
  | { kind: 'session'; data: Session }
  | null;

export function WorkspacePage({ fixture, query }: Props) {
  const proj = useProjects();
  const prop = useProposals();
  const sess = useSessions();
  const [selected, setSelected] = useState<SelectedItem>(null);

  const projects = (fixture ? fixtureProjects : (proj.data?.items ?? [])).filter(p => matchesQuery(query, [p.name, p.root_path, p.environment_id]));
  const proposals = (fixture ? fixtureProposals : (prop.data?.items ?? [])).filter(p => matchesQuery(query, [p.title, p.id, p.goal]));
  const sessions = (fixture ? fixtureSessions : (sess.data?.items ?? [])).filter(p => matchesQuery(query, [p.adapter, p.id, p.project_id]));

  const handleClose = useCallback(() => setSelected(null), []);

  return (
    <PageFrame section="workspace" title="共同工作区" subtitle="围绕目标与项目，让想法有一个生长的地方。" fixture={fixture} onRefresh={() => { proj.retry(); prop.retry(); sess.retry(); }} rail={<><RailSummary title="工作区概览" rows={[{ label: '提案', value: fixture ? proposals.length : prop.loading ? '加载中…' : prop.error && !prop.data ? '无法获取' : proposals.length }, { label: 'Agent 会话', value: fixture ? sessions.length : sess.loading ? '加载中…' : sess.error && !sess.data ? '无法获取' : sessions.length }]} note="批准、原生接续与交接尚未启用，能力未知时不会推断可用。" /><RailSummary title="Agent 上下文" rows={sessions.map(s => ({ id: s.id, label: s.adapter, value: contextStateLabel(s.context_state), onClick: () => setSelected({ kind: 'session', data: s }) }))} note={sessions.length ? '客户端名称不代表底层模型；点击检查来源、绑定与能力。' : '暂无已连接会话；不展示模拟活跃 Agent。'} /></>}>
      {(!fixture && (proj.stale || prop.stale || sess.stale)) && (
        <div role="alert" style={{
          padding: 'var(--space-3) var(--space-5)', background: 'var(--color-warning-subtle)',
          borderBottom: '1px solid var(--color-warning)', fontSize: 'var(--text-sm)',
          color: '#92610a', display: 'flex', justifyContent: 'space-between', alignItems: 'center',
        }}>
          <span>⚠ 数据可能已过期（刷新失败），显示的是上次成功加载的数据</span>
          <button type="button" onClick={() => { proj.retry(); prop.retry(); sess.retry(); }}
            style={{ padding: 'var(--space-1) var(--space-3)', fontSize: 'var(--text-xs)',
              border: '1px solid var(--color-warning)', borderRadius: 'var(--radius-sm)',
              background: 'transparent', cursor: 'pointer' }}>重试</button>
        </div>
      )}

      <div className={styles.grid}>
        {!projects.length && <SectionCard title="" tabs={['overview']}><IntroCard section="workspace" /></SectionCard>}
        {/* ── Projects ── */}
        <SectionCard tabs={['overview']} title="项目" count={projects.length}>
          {fixture ? (
            <ProjectList items={projects} onSelect={(p) => setSelected({ kind: 'project', data: p })} />
          ) : proj.loading && !proj.data ? (
            <div className={styles.loadingRow}>加载中…</div>
          ) : proj.error && !proj.data ? (
            <ErrorState message={proj.error} onRetry={proj.retry} />
          ) : projects.length === 0 ? (
            <EmptyState icon="📁" title="暂无项目" description="创建项目后将在此显示" />
          ) : (
            <ProjectList items={projects} onSelect={(p) => setSelected({ kind: 'project', data: p })} />
          )}
        </SectionCard>

        {/* ── Proposals ── */}
        <SectionCard tabs={['proposals']} title="提案" count={proposals.length}>
          {fixture ? (
            <ProposalList items={proposals} onSelect={(p) => setSelected({ kind: 'proposal', data: p })} />
          ) : prop.loading && !prop.data ? (
            <div className={styles.loadingRow}>加载中…</div>
          ) : prop.error && !prop.data ? (
            <ErrorState message={prop.error} onRetry={prop.retry} />
          ) : proposals.length === 0 ? (
            <EmptyState icon="📋" title="暂无提案" description="系统生成提案后将在此显示" />
          ) : (
            <ProposalList items={proposals} onSelect={(p) => setSelected({ kind: 'proposal', data: p })} />
          )}
        </SectionCard>

        {/* ── Sessions (full width) ── */}
        <div className={styles.fullWidth}>
          <SectionCard tabs={['overview', 'sessions']} title="会话" count={sessions.length}>
            {fixture ? (
              <SessionList items={sessions} onSelect={(s) => setSelected({ kind: 'session', data: s })} />
            ) : sess.loading && !sess.data ? (
              <div className={styles.loadingRow}>加载中…</div>
            ) : sess.error && !sess.data ? (
              <ErrorState message={sess.error} onRetry={sess.retry} />
            ) : sessions.length === 0 ? (
              <EmptyState icon="💻" title="暂无会话" description="Agent 会话绑定后将在此显示" />
            ) : (
              <SessionList items={sessions} onSelect={(s) => setSelected({ kind: 'session', data: s })} />
            )}
          </SectionCard>
        </div>
        <SectionCard title="研究路线" tabs={['overview']}><ResearchRoutes fixture={fixture} /></SectionCard>
        <SectionCard title="研究动态" tabs={['overview']}><DesignTimeline fixture={fixture} /></SectionCard>
      </div>

      {/* ── Detail Panels ── */}
      {selected && (
        <DetailPanel
          title={selected.kind === 'project' ? '项目详情' : selected.kind === 'proposal' ? '提案详情' : '会话详情'}
          onClose={handleClose}
          disabledActions={selected.kind === 'proposal' ? ['批准', '拒绝'] : selected.kind === 'session' ? ['原生接续', '显式交接'] : ['编辑项目']}
          showDisabledNotice
          disabledNoticeText={
            selected.kind === 'proposal'
              ? '提案审批操作尚未启用'
              : selected.kind === 'session'
                ? '会话恢复与交接操作尚未启用'
                : '项目编辑操作尚未启用'
          }
        >
          {selected.kind === 'project' && <ProjectDetail item={selected.data} isFixture={fixture} />}
          {selected.kind === 'proposal' && <ProposalDetail item={selected.data} isFixture={fixture} />}
          {selected.kind === 'session' && <SessionDetail item={selected.data} isFixture={fixture} />}
        </DetailPanel>
      )}
    </PageFrame>
  );
}

// ── Lists ──

function ProjectList({ items, onSelect }: { items: Project[]; onSelect: (p: Project) => void }) {
  if (!items.length) return <EmptyState title="没有匹配的项目" description="尝试其他搜索词" />;
  return (
    <ul className={`${styles.list} ac-project-list`} role="list">
      {items.map((p, index) => (
        <li key={p.id} className={`${styles.listItem} ${index === 0 ? 'ac-project-hero' : 'ac-project-preview-card'}`} tabIndex={0} role="button"
          onClick={() => onSelect(p)}
          onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(p); }}}>
          {index === 0 && <span className="ac-overline">YOUR RESEARCH SPACE</span>}
          <div className={index === 0 ? 'ac-project-title' : styles.itemHeader}>
            <span className="ac-project-symbol">{p.name.slice(0, 1)}</span>
            <span className={`${styles.itemTitle} ${index === 0 ? 'ac-project-name' : ''} line-clamp-2`}>{p.name}</span>
          </div>
          <span className="ac-text-button">查看项目与工作树<Icon name="arrow" size={14} /></span>
          {index === 0 && <><NetworkArt /><span className="ac-hero-note">RESEARCH CONNECTIONS · 品牌插画</span></>}
          <div className={styles.itemMeta}>
            <span>{p.environment_id}</span>
            <span className={styles.itemDesc}>{p.root_path}</span>
            {p.worktree_id && <span>wt: {p.worktree_id}</span>}
          </div>
        </li>
      ))}
    </ul>
  );
}

function ProposalList({ items, onSelect }: { items: Proposal[]; onSelect: (p: Proposal) => void }) {
  if (!items.length) return <EmptyState title="没有匹配的提案" description="尝试其他搜索词" />;
  return (
    <ul className={styles.list} role="list">
      {items.map((p) => (
        <li key={p.id} className={`${styles.listItem} ac-proposal-card`} tabIndex={0} role="button"
          onClick={() => onSelect(p)}
          onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(p); }}}>
          <div className={styles.itemHeader}>
            <span className={`${styles.itemTitle} line-clamp-2`}>{p.title}</span>
            <StatusBadge value={p.status} label={proposalStatusLabel(p.status)} />
          </div>
          <div className={`${styles.itemDesc} line-clamp-2`}>{p.goal}</div>
          <div className={styles.itemMeta}>
            <span>r{p.revision}</span>
            <span>{p.deliverables.length} 项交付物</span>
            <span>{p.scope.project_ids.length} 个项目</span>
          </div>
          <span className="ac-text-button">查看范围与停止条件<Icon name="arrow" size={14} /></span>
        </li>
      ))}
    </ul>
  );
}

function SessionList({ items, onSelect }: { items: Session[]; onSelect: (s: Session) => void }) {
  if (!items.length) return <EmptyState title="没有匹配的会话" description="尝试其他搜索词" />;
  return (
    <ul className={`${styles.list} ac-session-grid`} role="list">
      {items.map((s) => (
        <li key={s.id} className={`${styles.listItem} ac-session-card`} tabIndex={0} role="button"
          onClick={() => onSelect(s)}
          onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onSelect(s); }}}>
          <span className="ac-avatar green">{s.adapter.slice(0, 1).toUpperCase()}</span><div className={styles.itemHeader}>
            <span className={styles.itemTitle}>{s.adapter} · {s.id}</span>
            <div style={{ display: 'flex', gap: 'var(--space-2)', flexShrink: 0 }}>
              <StatusBadge value={s.binding_status} label={bindingStatusLabel(s.binding_status)} />
              <StatusBadge value={s.context_state} label={contextStateLabel(s.context_state)} />
            </div>
          </div>
          <span className="ac-session-context">原生接续 · {capabilityLabel(s.capabilities.native_resume)}</span>
          <span className="ac-text-button">查看会话与能力<Icon name="arrow" size={14} /></span>
          <div className={styles.itemMeta}>
            <span>项目: {s.project_id}</span>
            {s.native_session_id && <span>Native: {s.native_session_id}</span>}
            {s.context_packet_id && <span>Context: {s.context_packet_id}</span>}
          </div>
        </li>
      ))}
    </ul>
  );
}

// ── Details ──

function ProjectDetail({ item, isFixture }: { item: Project; isFixture: boolean }) {
  return (
    <>
      <Field label="名称">{item.name}</Field>
      {isFixture && <FixtureTag />}
      <Field label="环境 ID"><code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)' }}>{item.environment_id}</code></Field>
      <Field label="根路径"><code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)' }}>{item.root_path}</code></Field>
      <Field label="仓库 ID">{item.repo_id ?? <MutedValue>无</MutedValue>}</Field>
      <Field label="工作树 ID">{item.worktree_id ?? <MutedValue>无</MutedValue>}</Field>
      <Field label="ID"><code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xs)' }}>{item.id}</code></Field>
    </>
  );
}

function ProposalDetail({ item, isFixture }: { item: Proposal; isFixture: boolean }) {
  return (
    <>
      <Field label="标题">{item.title}</Field>
      {isFixture && <FixtureTag />}
      <Field label="状态"><StatusBadge value={item.status} label={proposalStatusLabel(item.status)} /></Field>
      <Field label="目标">{item.goal}</Field>
      <Field label="交付物">
        {item.deliverables.length > 0
          ? <ul style={{ paddingLeft: 'var(--space-5)', margin: 0 }}>{item.deliverables.map((d, i) => <li key={i}>{d}</li>)}</ul>
          : <MutedValue>无</MutedValue>}
      </Field>
      <Field label="预算">{formatBudget(item.budget) === '—' ? <MutedValue>未知（未提供预算）</MutedValue> : formatBudget(item.budget)}</Field>
      <FieldRow>
        <Field label="调用上限">{item.budget?.calls ?? <MutedValue>未知</MutedValue>}</Field>
        <Field label="费用上限">{item.budget?.money != null ? `${item.budget.money} ${item.budget.currency ?? '币种未知'}` : <MutedValue>未知</MutedValue>}</Field>
      </FieldRow>
      <Field label="停止条件">
        {item.stop_conditions.length > 0
          ? item.stop_conditions.join('; ')
          : <MutedValue>无</MutedValue>}
      </Field>
      <Field label="范围">
        <div style={{ fontSize: 'var(--text-sm)' }}>
          <div>项目: {item.scope.project_ids.length > 0 ? item.scope.project_ids.join(', ') : '无'}</div>
          <div>允许操作: {item.scope.allowed_actions.length > 0 ? item.scope.allowed_actions.join(', ') : '无'}</div>
        </div>
      </Field>
      <FieldRow>
        <Field label="版本">r{item.revision}</Field>
        <Field label="ID"><code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xs)' }}>{item.id}</code></Field>
      </FieldRow>
    </>
  );
}

function SessionDetail({ item, isFixture }: { item: Session; isFixture: boolean }) {
  return (
    <>
      {isFixture && <FixtureTag />}
      <Field label="适配器">{item.adapter}</Field>
      <FieldRow>
        <Field label="绑定状态"><StatusBadge value={item.binding_status} label={bindingStatusLabel(item.binding_status)} /></Field>
        <Field label="上下文状态"><StatusBadge value={item.context_state} label={contextStateLabel(item.context_state)} /></Field>
      </FieldRow>
      <Field label="项目 ID"><code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-sm)' }}>{item.project_id}</code></Field>
      <Field label="原生会话 ID">{item.native_session_id ?? <MutedValue>无</MutedValue>}</Field>
      <Field label="工作树 ID">{item.worktree_id ?? <MutedValue>无</MutedValue>}</Field>
      <Field label="上下文包 ID">{item.context_packet_id ?? <MutedValue>无</MutedValue>}</Field>
      <Field label="能力">
        <div style={{ fontSize: 'var(--text-sm)' }}>
          <div>原生恢复: {capabilityLabel(item.capabilities.native_resume)}</div>
          <div>交接: {capabilityLabel(item.capabilities.handoff)}</div>
        </div>
      </Field>
      <Field label="ID"><code style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-xs)' }}>{item.id}</code></Field>
    </>
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
