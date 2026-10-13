import { useState } from 'react';
import { relayApi } from '../../api/relay';
import { useReadApi } from '../../hooks/useReadApi';
import { QueryState } from '../../components/QueryState';
import { Field } from '../../components/DetailPanel';
import { relayDisplayState, relayLivenessLabel } from './relayPresentation';

/** Opening a conflict target performs only GET; it never changes the saved intent. */
export function RelayExistingTask({ taskID }: { taskID: string }) {
  const [opened, setOpened] = useState(false);
  const task = useReadApi(signal => relayApi.task(taskID, signal), { enabled: opened, key: taskID });
  return <section aria-label={`已有任务 ${taskID}`} className="relay-flow">
    <p>同一工作区已有任务：<code>{taskID}</code>。此次请求被拒绝，不会自动重派。</p>
    <button type="button" className="ac-button secondary compact" onClick={() => { if (opened) task.retry(); else setOpened(true); }}>只读查看已有任务</button>
    {opened && <QueryState state={task}>{data => <>
      <Field label="已有任务 ID"><code>{data.view.task.task_id}</code></Field>
      <Field label="已有任务 Agent">{data.view.task.agent}</Field>
      <Field label="已有任务作品">{data.view.task.work_id}</Field>
      <Field label="已有任务分支">{data.view.worktree?.branch || '工作区尚未准备'}</Field>
      <Field label="已有任务状态">{relayDisplayState(data.view)}</Field>
      <Field label="已有任务终端活性">{data.view.terminal ? relayLivenessLabel[data.view.terminal.liveness] : '终端尚未创建 · 活性未知'}</Field>
    </>}</QueryState>}
  </section>;
}
