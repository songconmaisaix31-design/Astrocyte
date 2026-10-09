import { useState } from 'react';
import { attentionApi } from '../../api/client';
import { useDistillerStatus } from '../../hooks/useAttention';
import { useCommand } from '../../hooks/useCommand';
import { QueryState } from '../../components/QueryState';
import { CommandState } from '../../components/CommandState';
import { SelectField, TextField } from './FormControls';
import { SourceRefs } from './SourceFields';
import { sourceIdentity, stages, type Distillation, type SourceRef } from './model';
import styles from './AttentionPage.module.css';

export function AutomaticDistillationForm({ input, sourceKey, records, disabled, onQueued }: { input: SourceRef; sourceKey: string; records: Distillation[]; disabled: boolean; onQueued: () => void }) {
  const state = useDistillerStatus();
  const [stage, setStage] = useState<Distillation['stage']>('content');
  const [question, setQuestion] = useState('');
  const [priorIDs, setPriorIDs] = useState<string[]>([]);
  const selected = records.filter(record => priorIDs.includes(record.id));
  const inputRefs = [...new Map([input, ...selected.flatMap(record => [...record.input_refs, ...record.related_refs])].map(ref => [sourceIdentity(ref), ref])).values()];
  const blocked = disabled || state.loading || state.stale || !state.data?.available || !state.data.configuration_id || !state.data.allowed_source_keys.includes(sourceKey);
  const command = useCommand(blocked);
  return <section className={styles.record} aria-label="自动沉淀">
    <h4>自动沉淀 · 独立作业</h4>
    <button className="ac-button secondary compact" type="button" disabled={state.loading} onClick={state.retry}>检查自动处理服务</button>
    <QueryState state={state}>{processor => <>
      <p>处理器 · {processor.processor || '未提供'} · 模型 {processor.model || '未知'} · 配置 {processor.configuration_id || '未提供'}</p>
      {!processor.available && <p role="note">自动处理未启用 · {processor.reason || '服务未提供可用处理器'} · {processor.required_action}</p>}
      {processor.available && !processor.allowed_source_keys.includes(sourceKey) && <p role="note">此来源未纳入已批准的自动处理范围，提交禁用。</p>}
      <form className={styles.form} onSubmit={event => { event.preventDefault(); const request = command.prepare({ expected_version: 1, input_refs: inputRefs, stage, question: question.trim(), processing_config: processor.configuration_id!, prior_distillation_ids: priorIDs }); void command.run(() => attentionApi.requestDistillation(request.body, request.key), onQueued, '已提交自动沉淀作业；以持久化队列和实际输出为准'); }}>
        <p className={styles.note}>仅提交你明确选择的版本和本轮问题。选入历史记录会同时列出其来源；服务逐项核对获准范围。人工记录不改标为模型生成，GET 不会触发自动处理。</p>
        <fieldset disabled={blocked || command.pending}><legend>本轮自动请求</legend>
          <SelectField label="自动沉淀层次" value={stage} onChange={value => setStage(value as Distillation['stage'])} options={stages} />
          <TextField label="自动沉淀本轮问题" value={question} onChange={setQuestion} multiline required />
          <fieldset><legend>明确选入前轮记录</legend>{records.map(record => <label key={record.id}><span><input style={{ width: 'auto' }} type="checkbox" checked={priorIDs.includes(record.id)} onChange={event => setPriorIDs(event.target.checked ? [...priorIDs, record.id] : priorIDs.filter(id => id !== record.id))} /> {stages.find(entry => entry.value === record.stage)?.label} · {record.question} · {record.id}</span></label>)}{!records.length && <p>暂无前轮记录；本次只处理当前来源版本。</p>}</fieldset>
        </fieldset>
        <FieldSources refs={inputRefs} />
        <CommandState {...command} /><button className="ac-button" type="submit" disabled={blocked || command.pending}>提交自动沉淀</button>
      </form>
    </>}</QueryState>
  </section>;
}
function FieldSources({ refs }: { refs: SourceRef[] }) {
  return <div><p>本次明确选择的输入版本</p><SourceRefs refs={refs} /></div>;
}
