import { useEffect, useState } from 'react';
import type { components } from '../../api/schema';
import { attentionApi } from '../../api/client';
import { useMaterialDetail, useMaterialContent, useDomains } from '../../hooks/useAttention';
import { useCommand } from '../../hooks/useCommand';
import { Field, FieldRow, MutedValue } from '../../components/DetailPanel';
import { StatusBadge } from '../../components/StatusBadge';
import { QueryState } from '../../components/QueryState';
import { CommandState } from '../../components/CommandState';
import { EmptyState } from '../../components/EmptyState';
import { formatDateTime, lifecycleLabel, importStatusLabel, materialKindLabel } from '../../utils/format';
import { ProvenanceFields, SourceRefs } from './SourceFields';
import { DistillationForm } from './DistillationForm';
import { AutomaticDistillationForm } from './AutomaticDistillationForm';
import { OpportunityForm } from './OpportunityForm';
import { TextField, SelectField } from './FormControls';
import { MaterialDomainForm } from './DomainsPanel';
import { stages, type Material, type Opportunity, type Distillation } from './model';
import styles from './AttentionPage.module.css';

type Detail = components['schemas']['MaterialDetailV1'];
export function MaterialWorkspace({ item, id, fixture, refreshToken, materials, domains, onChanged, onQueued, onOpportunity }: { item?: Material; id: string; fixture: boolean; refreshToken: number; materials: Material[]; domains: ReturnType<typeof useDomains>; onChanged: () => void; onQueued: () => void; onOpportunity: (opportunity: Opportunity) => void }) {
  const state = useMaterialDetail(id, !fixture);
  const { retry } = state;
  useEffect(() => { if (refreshToken) retry(); }, [refreshToken, retry]);
  if (fixture && item) return <><MaterialSummary item={item} /><p role="note">示例数据 · 不提供真实原文、来源版本或沉淀历史；所有写操作尚未启用。</p></>;
  return <><button type="button" className="ac-button secondary compact" onClick={state.retry} disabled={state.loading}>刷新资料详情</button><QueryState state={state}>{data => <MaterialBody detail={data} materials={materials} domains={domains} disabled={state.loading || state.stale} onChanged={onChanged} onQueued={onQueued} onOpportunity={onOpportunity} />}</QueryState></>;
}

function MaterialBody({ detail, materials, domains, disabled, onChanged, onQueued, onOpportunity }: { detail: Detail; materials: Material[]; domains: ReturnType<typeof useDomains>; disabled: boolean; onChanged: () => void; onQueued: () => void; onOpportunity: (opportunity: Opportunity) => void }) {
  const item = detail.material;
  const [revision, setRevision] = useState(item.current_revision);
  const [editing, setEditing] = useState(false);
  const [distilling, setDistilling] = useState(false);
  const [creating, setCreating] = useState(false);
  const [suggestion, setSuggestion] = useState<Distillation | null>(null);
  const source = detail.revisions.find(entry => entry.revision === revision);
  const modelContentRecords = detail.distillations.filter(record => record.stage === 'content' && record.status === 'succeeded' && (record.provenance.mode === 'selected_project_fixed_text' || record.provenance.processor === 'codex-cli') && record.input_refs.some(ref => ref.material_id === item.id && ref.revision === revision));
  const content = useMaterialContent(item.id, revision, item.lifecycle !== 'withdrawn');
  const command = useCommand(disabled || !item.version);
  const unavailable = disabled || !item.version || item.lifecycle === 'withdrawn';
  return <>
    <MaterialSummary item={item} />
    {!item.version && <p role="alert">服务未提供元数据版本，写操作禁用，请升级或重新连接服务。</p>}
    <div className={styles.actions}>
      <button type="button" className="ac-button secondary compact" disabled={disabled || !item.version} onClick={() => setEditing(!editing)}>整理收藏与状态</button>
      <button type="button" className="ac-button secondary compact" disabled={unavailable || command.pending} onClick={() => { const request = command.prepare({ expected_version: item.version!, action: 'reread' as const }); void command.run(() => attentionApi.recordMaterialUse(item.id, request.body, request.key), () => { content.retry(); onChanged(); }, '已记录主动重读'); }}>主动重读</button>
      <button type="button" className="ac-button secondary compact" disabled={unavailable || command.pending} onClick={() => { const request = command.prepare({ expected_version: item.version!, action: 'annotate' as const }); void command.run(() => attentionApi.recordMaterialUse(item.id, request.body, request.key), onChanged, '已记录人工标记'); }}>标记已关注</button>
      <button type="button" className="ac-button secondary compact" disabled={unavailable || command.pending} onClick={() => { const request = command.prepare({ expected_version: item.version!, action: 'project_reuse' as const }); void command.run(() => attentionApi.recordMaterialUse(item.id, request.body, request.key), onChanged, '已记录这次明确的项目复用'); }}>记录本次项目复用</button>
    </div>
    <p className={styles.note}>普通查看、GET 和自动刷新不增加人工关注；主动重读、标记、固定收藏、分类、项目复用和 @ 提及可记录人工行为。</p>
    <CommandState {...command} />
    {editing && <MetadataForm item={item} disabled={disabled || !item.version} onSaved={onChanged} />}
    <MaterialDomainForm material={item} state={domains} disabled={unavailable} onSaved={onChanged} />
    <section className={styles.record} aria-label="来源与原文版本">
      <h4>来源与原文版本</h4>
      <div className={styles.form}><SelectField label="原文版本" value={String(revision)} onChange={value => setRevision(Number(value))} options={detail.revisions.map(entry => ({ value: String(entry.revision), label: `v${entry.revision} · ${formatDateTime(entry.created_at)}` }))} /></div>
      {source ? <>
        <Field label="固定来源">{source.source_locator}</Field>
        <Field label="真实原文位置">{source.source_spans.length ? source.source_spans.join('；') : <MutedValue>未提供真实位置</MutedValue>}</Field>
        <ProvenanceFields value={source.provenance} />
        {source.summary && <Field label="导入摘要（独立于原文）">{source.summary}</Field>}
        <Field label="本版本模型内容整理">{modelContentRecords.length ? `已保存 ${modelContentRecords.length} 轮模型整理结果，见各轮沉淀的固定输入与实际输出。` : '尚无此版本的模型内容整理结果。导入正文或既有摘要不会自动完成模型整理。'}</Field>
        <Field label="原始附件">{source.attachments?.length ? source.attachments.map(attachment => <p key={attachment.name}><a href={`/api/v1/materials/${encodeURIComponent(item.id)}/revisions/${revision}/attachments/${encodeURIComponent(attachment.name)}`} target="_blank" rel="noreferrer">{attachment.name} · {attachment.media_type}</a><br />{attachment.source_locator}</p>) : <MutedValue>未提供附件</MutedValue>}</Field>
      </> : <p role="alert">此版本的来源记录缺失，请刷新资料。</p>}
      {item.lifecycle === 'withdrawn' ? <p>资料已撤回，原文分发与新沉淀已停止。</p> : <QueryState state={content}>{data => <><Field label="导入正文状态">{data.text ? '已保存可读文本；内容获取与模型整理分别记录。' : '未提供可读正文，请检查原始附件与导入队列。'}</Field><ProvenanceFields value={data.provenance} /><Field label="保存的原文 / 提取文本"><pre>{data.text || '此版本未提供可读原文；请检查原始附件。'}</pre></Field></>}</QueryState>}
    </section>
    <section aria-label="各轮沉淀">
      <h4>各轮沉淀</h4>
      {!detail.distillations.length && <EmptyState title="暂无沉淀记录" description="展开继续沉淀，选择内容整理与本轮问题，再按实际处理器能力明确提交；也可保存人工记录。" />}
      {detail.distillations.map(record => <article key={record.id} className={styles.record}>
        <h4>{stages.find(stage => stage.value === record.stage)?.label ?? record.stage} · {formatDateTime(record.created_at)}</h4>
        <ProvenanceFields value={record.provenance} />
        <Field label="本轮问题">{record.question || '未提供'}</Field>
        <Field label="输入版本"><SourceRefs refs={record.input_refs} /></Field>
        <Field label="明确选入的前轮记录">{record.prior_distillation_ids?.join('；') || '本轮未选入前轮记录'}</Field>
        <Field label="整理结果"><pre>{record.output_text}</pre></Field>
        <Field label="关联资料"><SourceRefs refs={record.related_refs} /></Field>
        <Field label="关联观点 / 冲突">{[...record.related_ideas, ...record.conflicts].join('；') || '未提供'}</Field>
        <Field label="待查问题">{record.pending_questions.join('；') || '未提供'}</Field>
        <Field label="下一轮问题">{record.next_question || '未提供'}</Field>
        {record.candidate_suggestion && <section className={styles.record}><h4>此轮处理器候选建议</h4><p>{record.candidate_suggestion.title}</p><Field label="建议用途">{record.candidate_suggestion.purpose}</Field><SourceRefs refs={record.candidate_suggestion.evidence_refs} /><Field label="建议下一步">{record.candidate_suggestion.next_step}</Field><Field label="建议缺失依据">{record.candidate_suggestion.missing_evidence.join('；') || '未记录'}</Field><button type="button" className="ac-button secondary compact" disabled={unavailable} onClick={() => { setSuggestion(record); setCreating(true); }}>按此建议编辑候选</button></section>}
        {record.stage === 'project' && <><Field label="关联目标">{record.goal_refs.join('；') || '未提供'}</Field><Field label="现有资产">{record.existing_assets.join('；') || '未提供'}</Field><Field label="预期改进">{record.expected_improvement || '未提供'}</Field><Field label="最小成果">{record.minimum_artifact || '未提供'}</Field><Field label="缺失依据">{record.missing_evidence.join('；') || '未提供'}</Field></>}
      </article>)}
      <div className={styles.actions}>
        <button className="ac-button" type="button" disabled={unavailable} onClick={() => setDistilling(!distilling)}>继续沉淀</button>
        <button className="ac-button secondary" type="button" disabled={unavailable} onClick={() => { setSuggestion(null); setCreating(!creating); }}>形成候选</button>
      </div>
      {distilling && source && <><DistillationForm key={`manual:${item.id}@${revision}`} input={{ material_id: item.id, revision, locator: source.source_locator, span: null }} materials={materials} disabled={unavailable} onSaved={onChanged} /><AutomaticDistillationForm key={`automatic:${item.id}@${revision}`} input={{ material_id: item.id, revision, locator: source.source_locator, span: null }} sourceKey={source.source_key} records={detail.distillations} disabled={unavailable} onQueued={onQueued} /></>}
      {creating && <OpportunityForm key={suggestion?.id ?? 'manual'} suggestion={suggestion?.candidate_suggestion ?? undefined} suggestionRecordID={suggestion?.id} materials={materials} distillations={detail.distillations} disabled={unavailable} onSaved={result => { onChanged(); onOpportunity(result); }} />}
    </section>
    <Field label="关注行为记录">{detail.uses.length ? <ul className={styles.timeline}>{detail.uses.map(use => <li key={use.id}>{use.actor_kind === 'human' ? '人类关注' : '机器使用'} · {use.action} · {formatDateTime(use.occurred_at)}</li>)}</ul> : <MutedValue>暂无行为记录</MutedValue>}</Field>
  </>;
}

function MetadataForm({ item, disabled, onSaved }: { item: Material; disabled: boolean; onSaved: () => void }) {
  const [expectedVersion, setExpectedVersion] = useState(item.version);
  const [reason, setReason] = useState(item.collection_reason ?? '');
  const [lifecycle, setLifecycle] = useState(item.lifecycle);
  const [pinned, setPinned] = useState(item.pinned ?? false);
  const command = useCommand(disabled);
  return <form className={styles.form} onSubmit={event => { event.preventDefault(); const request = command.prepare({ expected_version: expectedVersion!, lifecycle, pinned, collection_reason: reason.trim() || null }); void command.run(() => attentionApi.updateMaterial(item.id, request.body, request.key), result => { setExpectedVersion(result.material.version); onSaved(); }, '已保存收藏与资料状态'); }}>
    {item.version !== expectedVersion && <p role="note">资料已更新，当前输入保留。请关闭并重新展开此表单载入最新状态，避免覆盖其他修改。</p>}
    <fieldset disabled={disabled || command.pending}><legend>整理收藏与状态</legend>
      <TextField label="收藏理由" value={reason} onChange={setReason} multiline hint="留空明确为未提供；收藏不会创建任务。" />
      <label><span><input type="checkbox" checked={pinned} onChange={event => setPinned(event.target.checked)} style={{ width: 'auto' }} /> 固定到我的收藏</span></label>
      <SelectField label="资料生命周期" value={lifecycle} onChange={value => setLifecycle(value as typeof lifecycle)} options={[{ value: 'active', label: '活跃' }, { value: 'archived', label: '归档' }, { value: 'withdrawn', label: '撤回' }]} />
      {lifecycle === 'withdrawn' && <p role="note">保存撤回后，停止新的原文分发与沉淀；历史记录保留。</p>}
    </fieldset>
    <CommandState {...command} />
    <button className="ac-button" type="submit" disabled={disabled || command.pending}>保存收藏与状态</button>
  </form>;
}

function MaterialSummary({ item }: { item: Material }) {
  return <>
    <Field label="标题">{item.title || <MutedValue>无标题</MutedValue>}</Field>
    <FieldRow><Field label="生命周期"><StatusBadge value={item.lifecycle} label={lifecycleLabel(item.lifecycle)} /></Field><Field label="类型">{materialKindLabel(item.kind)}</Field></FieldRow>
    {item.import_status && <Field label="导入状态"><StatusBadge value={item.import_status} label={importStatusLabel(item.import_status)} /></Field>}
    <Field label="来源定位">{item.source_locator}</Field>
    <Field label="收藏理由">{item.collection_reason?.trim() || <MutedValue>未提供</MutedValue>}</Field>
    <Field label="来源片段">{item.source_spans?.length ? item.source_spans.join('；') : <MutedValue>未提供真实位置</MutedValue>}</Field>
    <FieldRow><Field label="当前内容版本">v{item.current_revision}</Field><Field label="元数据版本">{item.version == null ? '未知' : `v${item.version}`}</Field></FieldRow>
    <FieldRow><Field label="人工使用 / 人类关注">{item.human_usage_count == null ? '未知' : `${item.human_usage_count} 次`}</Field><Field label="Agent 使用 / 机器使用">{item.agent_usage_count == null ? '未知' : `${item.agent_usage_count} 次`}</Field></FieldRow>
    <FieldRow><Field label="人类关注活跃度">{item.attention_score ?? '未知'}</Field><Field label="资料长期价值">{item.long_term_value ?? '未知'}</Field></FieldRow>
    <Field label="算法排序策略">{item.ranking_strategy || '未提供'}</Field>
    <Field label="排序原因">{item.ranking_reason || '未提供'}</Field>
    <Field label="关注半衰期">{item.attention_half_life_seconds == null ? '未知' : `${item.attention_half_life_seconds} 秒`}</Field>
    <Field label="关注事件权重">{item.attention_weights ? Object.entries(item.attention_weights).map(([action, weight]) => `${action}: ${weight}`).join('；') : '未知'}</Field>
    <Field label="固定收藏">{item.pinned == null ? '未知' : item.pinned ? '已固定' : '未固定'}</Field>
  </>;
}
