import { useState } from 'react';
import type { components } from '../../api/schema';
import type { RelayActionResult, RelayTaskView } from '../../api/relay';
import { useRelayCommand } from './relayCommand';
import { relayReceiptLabel } from './relayPresentation';

export function RelayTransitionControls({ view, disabled, transition, onChanged }: {
  view: RelayTaskView;
  disabled: boolean;
  transition: (body: components['schemas']['RelayTransitionRequestV1'], key: string) => Promise<RelayActionResult>;
  onChanged: () => void;
}) {
  const [confirmedVersion, setConfirmedVersion] = useState('');
  const command = useRelayCommand();
  const version = `${view.task.state}:${view.task.updated_at}`;
  const checked = confirmedVersion === version;
  const canComplete = view.task.state === 'working';
  const exitConfirmed = view.terminal?.liveness === 'exited' && !!view.terminal.exited_at;
  const canRelease = view.task.state === 'done' && exitConfirmed;
  const locked = disabled || command.pending || command.uncertain || (!canComplete && !canRelease);
  if (!['working', 'done'].includes(view.task.state)) return null;
  return <section className="relay-flow" aria-label="人工完成与交接">
    <h3>{canComplete ? '由你确认任务完成' : '释放工作区，准备下一次接力'}</h3>
    <p className="relay-note">任务完成需要你核对目标并明确确认。Git 改动、事件和终端退出不会自动完成任务。已完成任务须取得终端退出证据后，才能释放给下一位 Agent。</p>
    {view.task.state === 'done' && !exitConfirmed && <p role="status">尚无终端退出确证，不能释放工作区。可先停止 Agent，再读取退出证据。</p>}
    <fieldset disabled={locked}><legend>{canComplete ? '核对任务目标' : '核对退出与工作区'}</legend>
      <label className="relay-check"><input type="checkbox" checked={checked} onChange={event => setConfirmedVersion(event.target.checked ? version : '')} />{canComplete ? '我已核对任务目标，明确确认此任务完成' : '我已核对终端退出证据，确认释放此中性工作区'}</label>
      <div className="relay-card-actions"><button type="button" className="ac-button compact" disabled={!checked} onClick={() => {
        if (locked || !checked) return;
        void command.run(key => transition({ from: canComplete ? 'working' : 'done', to: canComplete ? 'done' : 'released', completion_confirmed: canComplete }, key), () => { setConfirmedVersion(''); onChanged(); });
      }}>{canComplete ? '确认任务完成' : '确认释放工作区'}</button><button type="button" className="ac-button secondary compact" onClick={() => setConfirmedVersion('')}>取消确认</button></div>
    </fieldset>
    {command.error && <p role="alert">{command.error} · 操作 {command.operationID}</p>}
    {command.uncertain && <p role="alert">迁移结果未知，不重复提交；请只读取服务端状态与收据。</p>}
    {command.result && <p role="status">状态迁移收据 {command.result.receipt.receipt_id} · {relayReceiptLabel[command.result.receipt.status]}</p>}
  </section>;
}
