import { useState } from 'react';
import { useReadApi } from '../../../hooks/useReadApi';
import { usePersistentCommand } from '../../../hooks/usePersistentCommand';
import { QueryState } from '../../../components/QueryState';
import { CommandState } from '../../../components/CommandState';
import { Field } from '../../../components/DetailPanel';
import { JudgmentResultDetails } from './JudgmentResultDetails';
import { JudgmentFeedbackFields, type JudgmentFeedbackDraft } from './JudgmentFeedbackFields';
import { judgmentApi, type JudgmentResult } from './judgmentApi';
import styles from './DelegationWorkspace.module.css';

const feedbackLabels = { adopt: '采用建议', later: '以后再看', reject: '拒绝建议', revise: '需要修改', already_solved: '已经解决' };

export function JudgmentResultWorkspace({ projectID, decisionID, disabled }: { projectID: string; decisionID: string; disabled: boolean }) {
  const result = useReadApi(signal => judgmentApi.result(projectID, decisionID, signal));
  const feedback = useReadApi(signal => judgmentApi.feedback(projectID, decisionID, signal));
  const blocked = disabled || result.loading || result.stale || !result.data;
  return <section className={styles.result}>
    <button className="ac-button secondary compact" type="button" onClick={() => { result.retry(); feedback.retry(); }}>只读刷新判断与反馈</button>
    <QueryState state={result}>{({ judgment }) => <><JudgmentResultDetails result={judgment} /><JudgmentResultActions projectID={projectID} result={judgment} disabled={blocked || feedback.loading || feedback.stale || !feedback.data} onFeedback={feedback.retry} /></>}</QueryState>
    <section aria-label="固定判断的人工反馈"><h3>人工反馈记录</h3><QueryState state={feedback} empty={data => !data.items.length} emptyTitle="暂无人工反馈">{data => <ul className={styles.list}>{data.items.map(item => <li key={item.id}><Field label="反馈类型">{feedbackLabels[item.kind as keyof typeof feedbackLabels] ?? item.kind}</Field><Field label="反馈理由">{item.reason}</Field><Field label="目标结果版本">r{item.result_version}</Field><Field label="反馈人"><code>{item.actor_id}</code></Field><Field label="时间">{item.created_at}</Field></li>)}</ul>}</QueryState></section>
  </section>;
}

function JudgmentResultActions({ projectID, result, disabled, onFeedback }: { projectID: string; result: JudgmentResult; disabled: boolean; onFeedback: () => void }) {
  const [draft, setDraft] = useState<JudgmentFeedbackDraft>({ kind: '', reason: '' });
  const [expectedVersion, setExpectedVersion] = useState(result.result_version);
  const [confirmed, setConfirmed] = useState(false);
  const feedback = usePersistentCommand<Awaited<ReturnType<typeof judgmentApi.recordFeedback>>>(`judgment:feedback:${projectID}:${result.decision_id}`, () => ({ status: 'succeeded', message: '人工反馈已保存' }));
  const admission = usePersistentCommand<Awaited<ReturnType<typeof judgmentApi.admit>>>(`judgment:admission:${projectID}:${result.decision_id}`, () => ({ status: 'succeeded', message: '资料已准入并建立草稿提案' }));
  const busy = disabled || feedback.pending || feedback.uncertain || admission.pending || admission.uncertain;
  const context = result.context;
  return <>
    <form className={styles.result} onSubmit={event => {
      event.preventDefault();
      if (busy || !draft.kind || !draft.reason.trim() || expectedVersion !== result.result_version) return;
      void feedback.run(() => judgmentApi.recordFeedback(projectID, result.decision_id, { result_version: expectedVersion, kind: draft.kind as Exclude<JudgmentFeedbackDraft['kind'], ''>, reason: draft.reason }), (_, outcome) => { if (outcome.status === 'succeeded') onFeedback(); });
    }}>
      <JudgmentFeedbackFields draft={draft} disabled={busy} onChange={setDraft} />
      <Field label="本次反馈固定版本">r{expectedVersion}</Field>
      {expectedVersion !== result.result_version && <div role="alert"><p>结果版本已变化，原反馈输入保留。请核查当前结果后明确选择。</p><button type="button" className="ac-button secondary compact" disabled={busy} onClick={() => setExpectedVersion(result.result_version)}>使用已核对的结果版本 r{result.result_version}</button></div>}
      <CommandState {...feedback} />{feedback.uncertain && <p role="alert">人工反馈请求 {feedback.operationID} 的结果未知，关闭或刷新后继续禁止重发。请只读核查反馈记录。</p>}
      <button className="ac-button" type="submit" disabled={busy || !draft.kind || !draft.reason.trim() || expectedVersion !== result.result_version}>保存该固定版本的人工反馈</button>
    </form>
    {context && <section className={styles.fields} aria-label="判断后的资料准入"><h3>固定候选的资料准入</h3><p>准入保存资料关系并建立草稿提案。任务批准与建立任务分别确认。</p><Field label="候选固定版本">{context.opportunity.title} · r{context.opportunity.revision} · v{context.opportunity.version}</Field>
      <label className={styles.confirmation}><input type="checkbox" checked={confirmed} disabled={busy} onChange={event => setConfirmed(event.target.checked)} /><span>已核查当前固定判断、人工反馈、候选资料及准入范围</span></label>
      <button className="ac-button" type="button" disabled={busy || !!admission.result || !confirmed || result.decision !== 'recommend' || result.truncated} onClick={() => {
        if (busy || !confirmed) return;
        void admission.run(key => judgmentApi.admit(projectID, result.decision_id, { schema_version: 1, request_id: key, expected_opportunity_version: context.opportunity.version }, key));
      }}>准入该固定候选并创建草稿提案</button>
      <CommandState {...admission} />{admission.uncertain && <p role="alert">准入请求 {admission.operationID} 的结果未知，原幂等身份继续保留，关闭或刷新后禁止重发。</p>}
      {admission.result && <><Field label="准入记录"><code>{admission.result.admission.admission_id}</code> · v{admission.result.admission.admission_version}</Field><Field label="草稿提案"><code>{admission.result.admission.proposal_id}</code> · v{admission.result.admission.proposal_version}</Field><a href="/workspace?tab=proposals">查看工作区提案与批准</a></>}
    </section>}
  </>;
}
