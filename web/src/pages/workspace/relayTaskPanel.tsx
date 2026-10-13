import { useState } from 'react';
import { relayApi, type RelayReceipt, type RelayTaskView } from '../../api/relay';
import { useReadApi } from '../../hooks/useReadApi';
import { QueryState } from '../../components/QueryState';
import { Field } from '../../components/DetailPanel';
import { EventTable } from '../cognition/shared';
import { navigate } from '../../router';
import { useRelayCommand } from './relayCommand';
import { RelayIntentPanel } from './relayIntentPanel';
import { relayActionLabel, relayDisplayState, relayElapsed, relayLivenessLabel, relayReceiptLabel, relayUnresolved } from './relayPresentation';
import { RelayReceipts } from './relayReceipts';

export function RelayTaskSummary({ view, workName }: { view: RelayTaskView; workName?: string }) {
  return <div className="relay-task-summary">
    <Field label="任务 ID"><code>{view.task.task_id}</code></Field>
    <Field label="Agent">{view.task.agent}</Field><Field label="作品">{workName || view.task.work_id}</Field>
    <Field label="分支">{view.worktree?.branch || '工作区尚未准备'}</Field>
    <Field label="当前状态">{relayDisplayState(view)}</Field>
    <Field label="已持续时间">{relayElapsed(view)}</Field>
    <Field label="终端活性"><span className={`relay-liveness relay-${view.terminal?.liveness || 'unverifiable'}`}>{view.terminal ? relayLivenessLabel[view.terminal.liveness] : '终端尚未创建 · 活性未知'}</span></Field>
    {view.terminal && <Field label="终端 ID">{view.terminal.terminal_id}</Field>}
    {view.task.degraded_launch && <p role="status">degraded_launch · 服务端已安全降级为终端路径</p>}
  </div>;
}

export function RelayTaskPanel({ taskID, onChanged }: { taskID: string; onChanged: () => void }) {
  const task = useReadApi(signal => relayApi.task(taskID, signal), { key: taskID, refreshMs: 3000 });
  const events = useReadApi(signal => relayApi.events(taskID, signal), { key: taskID, refreshMs: 3000 });
  const receipts = useReadApi(signal => relayApi.receipts(taskID, signal), { key: taskID, refreshMs: 3000 });
  const projects = useReadApi(signal => relayApi.projects(signal), { refreshMs: 5000 });
  const [input, setInput] = useState('');
  const send = useRelayCommand(), read = useRelayCommand(), stop = useRelayCommand();
  const changed = () => { task.retry(); events.retry(); receipts.retry(); onChanged(); };
  const view = task.data?.view;
  const permission = projects.data?.projects.find(project => project.project_id === view?.task.project_id);
  const active = view && ['accepted', 'working'].includes(view.task.state) && view.task.condition !== 'terminal_exited' && view.terminal?.liveness !== 'exited';
  const unavailable = !active || task.stale || receipts.stale || projects.stale || !receipts.data || !permission;
  const pending = send.pending || read.pending || stop.pending;
  const disabled = (action: RelayReceipt['action']) => unavailable || pending || !permission?.allowed_actions.includes(action === 'read' ? 'observe' : action) || relayUnresolved(receipts.data?.receipts || [], action);
  return <div className="relay-flow">
    <QueryState state={task}>{data => <><RelayTaskSummary view={data.view} />
      {data.view.task.state === 'ready' && <RelayIntentPanel existing={data.view} onChanged={changed} />}
      {data.view.task.condition === 'terminal_exited' && <p role="alert">终端退出，任务未完成；不会自动重派。</p>}
      {data.view.worktree && <Field label="中性工作区"><code>{data.view.worktree.worktree_id} · {data.view.worktree.root}</code></Field>}
      <details><summary>服务端持久状态迁移</summary><ol>{data.view.transitions.map((transition, index) => <li key={`${transition.ts}-${index}`}>{transition.from} → {transition.to} · {transition.ts} · {transition.operator}</li>)}</ol>{!data.view.transitions.length && <p>尚无状态迁移记录。</p>}</details>
    </>}</QueryState>
    <QueryState state={projects}>{() => null}</QueryState>
    <form onSubmit={event => { event.preventDefault(); if (!disabled('send') && !send.uncertain && input.trim()) void send.run(key => relayApi.send(taskID, input, key), result => { if (result.receipt.status === 'accepted') setInput(''); changed(); }); }}>
      <label>发送输入<textarea aria-label="发送输入" value={input} onChange={event => setInput(event.target.value)} rows={4} disabled={pending} /></label>
      <div className="relay-card-actions">
        <button type="submit" className="ac-button compact" disabled={!!disabled('send') || send.uncertain || !input.trim()}>发送输入</button>
        <button type="button" className="ac-button secondary compact" disabled={!!disabled('read') || read.uncertain} onClick={() => { void read.run(key => relayApi.read(taskID, key), changed); }}>读取会话并留收据</button>
        <button type="button" className="ac-button secondary compact" disabled={!!disabled('stop') || stop.uncertain} onClick={() => { void stop.run(key => relayApi.stop(taskID, key), changed); }}>停止 Agent</button>
      </div>
    </form>
    {[send, read, stop].map((command, index) => <div key={index}>{command.error && <p role="alert">{command.error} · 操作 {command.operationID}</p>}{command.uncertain && <p role="alert">动作结果未知，保留输入并禁止重复该动作；读取收据核对。</p>}{command.result && <p role="status">{relayActionLabel[command.result.receipt.action]}收据 {command.result.receipt.receipt_id} · {relayReceiptLabel[command.result.receipt.status]}</p>}</div>)}
    <button type="button" className="ac-button secondary compact" onClick={changed}>只读取任务、事件和收据</button>
    <section aria-label="任务收据"><h3>动作收据</h3><QueryState state={receipts}>{data => <RelayReceipts receipts={data.receipts} />}</QueryState></section>
    <section aria-label="任务事件流"><h3>任务事件</h3><p className="relay-note">按时间倒序 · 仅六个派生字段</p><QueryState state={events}>{data => <EventTable events={[...data.events].sort((a, b) => Date.parse(b.ts) - Date.parse(a.ts) || a.event_id.localeCompare(b.event_id))} />}</QueryState></section>
    <button type="button" className="ac-button secondary compact" onClick={() => navigate('/cognition/today')}>在 Today 查看认知活动</button>
  </div>;
}
