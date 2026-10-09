import { useState } from 'react';
import { attentionApi } from '../../api/client';
import { useCommand } from '../../hooks/useCommand';
import { CommandState } from '../../components/CommandState';
import { TextField, SelectField } from './FormControls';
import { ReferencePicker } from './SourceFields';
import { lines, sourceRef, stages, type Material, type SourceRef } from './model';
import type { components } from '../../api/schema';
import styles from './AttentionPage.module.css';

type Stage = components['schemas']['RecordDistillationRequestV1']['stage'];
export function DistillationForm({ material, revision, materials, disabled, onSaved }: { material: Material; revision: number; materials: Material[]; disabled: boolean; onSaved: () => void }) {
  const [stage, setStage] = useState<Stage>('content');
  const [output, setOutput] = useState('');
  const [question, setQuestion] = useState('');
  const [nextQuestion, setNextQuestion] = useState('');
  const [relatedRefs, setRelatedRefs] = useState<SourceRef[]>([]);
  const [ideas, setIdeas] = useState('');
  const [conflicts, setConflicts] = useState('');
  const [pending, setPending] = useState('');
  const [goals, setGoals] = useState('');
  const [assets, setAssets] = useState('');
  const [improvement, setImprovement] = useState('');
  const [artifact, setArtifact] = useState('');
  const [missing, setMissing] = useState('');
  const command = useCommand(disabled);
  return <form className={styles.form} onSubmit={event => {
    event.preventDefault();
    const request = command.prepare({ expected_version: 1, input_refs: [sourceRef(material, revision)], stage, output_text: output.trim(), question: question.trim(), processing_config: 'manual:v1', next_question: nextQuestion.trim() || null, related_refs: relatedRefs, related_ideas: lines(ideas), conflicts: lines(conflicts), pending_questions: lines(pending), goal_refs: lines(goals), existing_assets: lines(assets), expected_improvement: improvement.trim(), minimum_artifact: artifact.trim(), missing_evidence: lines(missing) });
    void command.run(() => attentionApi.recordDistillation(request.body, request.key), result => {
      onSaved();
      // Keep the completed record and inputs visible for reuse or the next question.
      if (result.reused) return;
    }, '已保存人工沉淀；相同输入、配置和问题由服务复用结果');
  }}>
    <h4>继续沉淀 · 人工记录</h4>
    <p className={styles.note}>输入固定到资料 v{revision}。此表单保存你已整理的结果，来源明确为人工；导入摘要不自动等于三层沉淀完成。</p>
    <button className="ac-button secondary compact" type="button" disabled title="当前自动整理服务尚未配置">自动总结 · 未配置</button>
    <fieldset disabled={disabled || command.pending}><legend>本轮记录</legend>
      <SelectField label="沉淀层次" value={stage} onChange={value => setStage(value as Stage)} options={stages} />
      <TextField label="本轮问题" value={question} onChange={setQuestion} required />
      <TextField label="人工整理结果" value={output} onChange={setOutput} multiline required />
      <TextField label="下一轮问题（可选）" value={nextQuestion} onChange={setNextQuestion} multiline />
      {stage !== 'content' && <>
        <ReferencePicker label="关联资料" materials={materials} value={relatedRefs} onChange={setRelatedRefs} />
        <TextField label="关联观点（每行一项）" value={ideas} onChange={setIdeas} multiline />
        <TextField label="冲突（每行一项）" value={conflicts} onChange={setConflicts} multiline />
        <TextField label="待查问题（每行一项）" value={pending} onChange={setPending} multiline hint="资料不足时保留待查问题；不要补造关联依据。" />
      </>}
      {stage === 'project' && <>
        <TextField label="关联目标（每行一项）" value={goals} onChange={setGoals} multiline required />
        <TextField label="现有资产（每行一项）" value={assets} onChange={setAssets} multiline required />
        <TextField label="预期改进" value={improvement} onChange={setImprovement} multiline required />
        <TextField label="最小成果" value={artifact} onChange={setArtifact} required />
        <TextField label="缺失依据（每行一项）" value={missing} onChange={setMissing} multiline />
      </>}
    </fieldset>
    <CommandState {...command} />
    <button className="ac-button" type="submit" disabled={disabled || command.pending}>保存人工沉淀</button>
  </form>;
}
