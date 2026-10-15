import { useState } from 'react';
import { proposalsApi, type Goal } from '../../../api/proposals';
import { useReadApi } from '../../../hooks/useReadApi';
import { QueryState } from '../../../components/QueryState';
import { GoalEditor } from './GoalEditor';
import { GoalRevision, GoalRevisionDiff } from './GoalRevision';
import styles from './ProposalWorkspace.module.css';

export function GoalDetailWorkspace({ id, onChanged }: { id: string; onChanged: () => void }) {
  const current = useReadApi(signal => proposalsApi.goal(id, undefined, signal), { key: id });
  const [revision, setRevision] = useState(0);
  const [editing, setEditing] = useState<Goal | null>(null);
  const historical = useReadApi(signal => proposalsApi.goal(id, revision, signal), { enabled: revision > 0, key: `${id}@${revision}` });
  const state = revision ? historical : current;
  const previousRevision = (state.data?.goal.revision ?? 1) - 1;
  const previous = useReadApi(signal => proposalsApi.goal(id, previousRevision, signal), { enabled: previousRevision > 0, key: `${id}@previous:${previousRevision}` });
  const unavailable = state.loading || state.stale || !state.data || current.loading || current.stale || !current.data;
  return <>
    <div className={styles.actions}><button className="ac-button secondary compact" type="button" disabled={current.loading} onClick={() => { current.retry(); if (revision) historical.retry(); }}>刷新当前目标</button>
      {current.data && <label>查看目标版本<select aria-label="查看目标版本" value={revision} disabled={!!editing} onChange={event => setRevision(Number(event.target.value))}><option value={0}>当前版本 r{current.data.goal.revision}</option>{Array.from({ length: current.data.goal.revision }, (_, index) => index + 1).map(value => <option key={value} value={value}>固定版本 r{value}</option>)}</select></label>}
    </div>
    {!!revision && current.error && <p role="alert">当前目标读取未完成，请使用“刷新当前目标”核查最新记录。固定历史版本继续独立显示。</p>}
    <QueryState state={state}>{({ goal }) => <>
      {!!revision && <p>正在查看固定历史目标版本。</p>}<GoalRevision goal={goal} />
      {previousRevision > 0 && <QueryState state={previous}>{({ goal: before }) => <GoalRevisionDiff before={before} after={goal} />}</QueryState>}
      <button className="ac-button secondary" type="button" disabled={unavailable || !!revision || !!editing} onClick={() => setEditing(goal)}>编辑目标新版本</button>
    </>}</QueryState>
    {editing && <section aria-label="编辑目标新版本"><h4>编辑目标新版本</h4><GoalEditor key={editing.id} initial={editing} latest={current.data?.goal} disabled={current.loading || current.stale || !current.data} onSaved={() => { setEditing(null); setRevision(0); current.retry(); onChanged(); }} /><button className="ac-button secondary compact" type="button" onClick={() => setEditing(null)}>关闭目标编辑</button></section>}
  </>;
}
