import { useState } from 'react';
import type { components } from '../../api/schema';
import { attentionApi } from '../../api/client';
import { useCommand } from '../../hooks/useCommand';
import { CommandState } from '../../components/CommandState';
import { TextField } from './FormControls';
import { ReferencePicker } from './SourceFields';
import { dimensionLabels, lines, sourceIdentity, unknownDimensions, type Dimensions, type Distillation, type Material, type Opportunity, type SourceRef } from './model';
import styles from './AttentionPage.module.css';

export function OpportunityForm({ initial, suggestion, suggestionRecordID, materials, distillations, disabled, onSaved }: { initial?: Opportunity; suggestion?: components['schemas']['CandidateSuggestionV1']; suggestionRecordID?: string; materials: Material[]; distillations: Distillation[]; disabled: boolean; onSaved: (opportunity: Opportunity) => void }) {
  const [expectedVersion, setExpectedVersion] = useState(initial?.version ?? 1);
  const [title, setTitle] = useState(initial?.title ?? suggestion?.title ?? '');
  const [purpose, setPurpose] = useState(initial?.purpose ?? suggestion?.purpose ?? '');
  const [nextStep, setNextStep] = useState(initial?.next_step ?? suggestion?.next_step ?? '');
  const [missing, setMissing] = useState((initial?.missing_evidence ?? suggestion?.missing_evidence)?.join('\n') ?? '');
  const [goals, setGoals] = useState((initial?.goal_refs ?? suggestion?.goal_refs)?.join('\n') ?? '');
  const [refs, setRefs] = useState<SourceRef[]>(initial?.evidence_refs ?? suggestion?.evidence_refs ?? []);
  const [ids, setIds] = useState<string[]>(initial?.distillation_ids ?? (suggestionRecordID ? [suggestionRecordID] : []));
  const [dimensions, setDimensions] = useState<Dimensions>(initial?.dimensions ?? suggestion?.dimensions ?? unknownDimensions());
  const command = useCommand(disabled || (!!initial && !initial.version));
  return <form className={styles.form} onSubmit={event => {
    event.preventDefault();
    const request = command.prepare({ expected_version: expectedVersion, title: title.trim(), purpose: purpose.trim(), next_step: nextStep.trim(), evidence_refs: refs, distillation_ids: ids, dimensions, goal_refs: lines(goals), missing_evidence: lines(missing) });
    void command.run(() => initial ? attentionApi.reviseOpportunity(initial.id, request.body, request.key) : attentionApi.createOpportunity(request.body, request.key), result => { setExpectedVersion(result.opportunity.version ?? expectedVersion); onSaved(result.opportunity); }, initial ? '已保存候选新版本；历史版本保留' : '已保存候选；尚未准入或启动任务');
  }}>
    <h4>{initial ? '编辑候选版本' : '形成候选'}</h4>
    {suggestion && <p className={styles.note}>从实际处理器记录 {suggestionRecordID} 载入建议；你可修订后明确保存，未自动生成任务。</p>}
    {initial && initial.version !== expectedVersion && <p role="note">候选已更新，当前输入保留。请关闭并重新展开编辑载入最新版本，避免覆盖其他修改。</p>}
    <p className={styles.note}>候选保留依据、用途与下一步，并引用固定版本的主题关联或明确待查问题，才可进入待审阅。条件不足仍可保存为酝酿中；服务核对记录与来源版本。保存和采用反馈均不会批准任务。</p>
    <fieldset disabled={disabled || command.pending || (!!initial && !initial.version)}><legend>候选内容</legend>
      <TextField label="候选标题" value={title} onChange={setTitle} required />
      <TextField label="候选用途 / 为什么值得做" value={purpose} onChange={setPurpose} multiline required />
      <TextField label="最小下一步" value={nextStep} onChange={setNextStep} multiline required />
      <TextField label="缺失依据（每行一项）" value={missing} onChange={setMissing} multiline hint="未知或不足请明确填写，不把没有记录显示为已齐备。" />
      <TextField label="关联目标（每行一项）" value={goals} onChange={setGoals} multiline />
      <ReferencePicker label="候选依据（固定内容版本）" materials={materials} value={refs} onChange={setRefs} />
      <fieldset><legend>引用沉淀记录</legend>
        {!distillations.length && <p>暂无沉淀记录；可先回资料详情继续沉淀。</p>}
        {distillations.map(record => <label key={record.id}><span><input type="checkbox" checked={ids.includes(record.id)} style={{ width: 'auto' }} onChange={event => {
          setIds(event.target.checked ? [...ids, record.id] : ids.filter(id => id !== record.id));
          if (event.target.checked) {
            const existing = new Set(refs.map(sourceIdentity));
            const additions = [...record.input_refs, ...record.related_refs].filter(ref => { const key = sourceIdentity(ref); if (existing.has(key)) return false; existing.add(key); return true; });
            setRefs([...refs, ...additions]);
          }
        }} /> {record.stage} · {record.question} · {record.id}</span></label>)}
        {ids.filter(id => !distillations.some(record => record.id === id)).map(id => <p key={id}>已引用记录 {id}（未在当前查询返回，保留原引用）</p>)}
      </fieldset>
      <DimensionEditor value={dimensions} onChange={setDimensions} />
    </fieldset>
    <CommandState {...command} />
    <button type="submit" className="ac-button" disabled={disabled || command.pending || (!!initial && !initial.version)}>{initial ? '保存候选新版本' : '保存候选'}</button>
  </form>;
}

export function DimensionEditor({ value, onChange }: { value: Dimensions; onChange: (value: Dimensions) => void }) {
  return <fieldset><legend>四维评估 · 未知留空</legend>
    <p className={styles.note}>0 表示明确的零分；留空表示未知。综合排序由已保存的用户配置决定，未知维度不填零。</p>
    {(Object.keys(dimensionLabels) as (keyof Dimensions)[]).map(key => <div key={key} className={styles.record}>
      <TextField label={`${dimensionLabels[key]}（0–1，未知留空）`} type="number" value={value[key].value == null ? '' : String(value[key].value)} onChange={score => onChange({ ...value, [key]: { ...value[key], value: score === '' ? null : Number(score) } })} />
      <TextField label={`${dimensionLabels[key]}依据`} value={value[key].reason} onChange={reason => onChange({ ...value, [key]: { ...value[key], reason } })} required />
    </div>)}
  </fieldset>;
}
