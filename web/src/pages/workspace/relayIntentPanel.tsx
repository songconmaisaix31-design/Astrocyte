import { useState } from 'react';
import { relayApi, type RelayTaskView, type RelayTaskResult } from '../../api/relay';
import { useReadApi } from '../../hooks/useReadApi';
import { QueryState } from '../../components/QueryState';
import { Field } from '../../components/DetailPanel';
import type { WorksWork } from './worksAlbumPresentation';
import { useRelayCommand } from './relayCommand';
import { relayAgentAvailability, relayProjectMatches, relayReceiptLabel, relayUnresolved } from './relayPresentation';
import { RelayReceipts } from './relayReceipts';
import { RelayExistingTask } from './relayExistingTask';

export function RelayIntentPanel({ work, existing, onChanged }: { work?: WorksWork; existing?: RelayTaskView; onChanged: () => void }) {
  const projects = useReadApi(signal => relayApi.projects(signal));
  const agents = useReadApi(signal => relayApi.agents(signal));
  const [projectID, setProjectID] = useState('');
  const [agent, setAgent] = useState('');
  const [description, setDescription] = useState('');
  const [view, setView] = useState<RelayTaskView | null>(existing ?? null);
  const command = useRelayCommand<RelayTaskResult>();
  const compatible = projects.data?.projects.filter(project => !work || relayProjectMatches(work.root, project)) ?? [];
  const project = projects.data?.projects.find(project => project.project_id === (view?.task.project_id || projectID));
  const unavailable = projects.loading || projects.stale || !projects.data || agents.loading || agents.stale || !agents.data;
  return <div className="relay-flow">
    {view ? <RelayConfirmation key={view.task.task_id} view={view} onChanged={onChanged} /> : <>
      <p>声明意图只保存任务，下一页确认权限后才会启动 Agent。</p>
      {work && <><Field label="作品">{work.name}</Field><Field label="交接手册">{work.handoff.content || '尚无交接手册，请先补充当前状态和下一步。'}</Field></>}
      <QueryState state={projects}>{() => !compatible.length ? <p role="status">作品尚无根目录完全匹配的获准项目。请在共同工作区登记并确认项目许可；当前服务不接受嵌套根目录代替作品根。</p> : null}</QueryState>
      <QueryState state={agents}>{() => null}</QueryState>
      <form onSubmit={event => {
        event.preventDefault();
        if (unavailable || !project || !work || !agent.trim() || !description.trim()) return;
        void command.run(key => relayApi.intent({ work_id: work.id, project_id: project!.project_id, agent, task: description.trim() }, key), result => {
          setView(result.view);
          onChanged();
        });
      }}>
        <fieldset disabled={unavailable || command.pending || command.uncertain}>
          <legend>你希望如何接力</legend>
          <label>获准项目<select aria-label="获准项目" value={projectID} onChange={event => setProjectID(event.target.value)} required><option value="">选择项目…</option>{compatible.map(project => <option key={project.project_id} value={project.project_id}>{project.name} · {project.root}</option>)}</select></label>
          <label>偏好 Agent<select aria-label="偏好 Agent" value={agent} onChange={event => setAgent(event.target.value)} required><option value="">选择注册 Agent…</option>{agents.data?.agents.map(spec => <option key={spec.name} value={spec.name}>{spec.name} · {spec.drive_mode}</option>)}</select></label>
          {agent && agents.data?.agents.filter(spec => spec.name === agent).map(spec => <p className="relay-note" key={spec.name}>{relayAgentAvailability(spec, agents.data!.inventory)}</p>)}
          <label>任务描述<textarea aria-label="任务描述" value={description} onChange={event => setDescription(event.target.value)} rows={5} required /></label>
          <p className="relay-note">首条输入由服务端合并交接手册与任务描述。注册、安装和原生能力分别核实。</p>
          <button type="submit" className="ac-button" disabled={!projectID || !agent || !description.trim()}>保存意图并查看权限</button>
        </fieldset>
      </form>
    </>}
    {command.error && <p role="alert">{command.error} · 操作 {command.operationID}</p>}
    {command.existingTaskID && <RelayExistingTask key={command.existingTaskID} taskID={command.existingTaskID} />}
    {command.uncertain && <p role="alert">意图结果未知，保留草稿；请读取任务与收据核对，不会自动重试。</p>}
    {command.result && <p role="status">意图已保存：{command.result.view.task.task_id} · 启动前仍须确认权限。</p>}
  </div>;
}

function RelayConfirmation({ view, onChanged }: { view: RelayTaskView; onChanged: () => void }) {
  const projects = useReadApi(signal => relayApi.projects(signal), { refreshMs: 5000 });
  const task = useReadApi(signal => relayApi.task(view.task.task_id, signal), { refreshMs: 3000 });
  const receipts = useReadApi(signal => relayApi.receipts(view.task.task_id, signal), { refreshMs: 3000 });
  const [confirmedRevision, setConfirmedRevision] = useState<number | null>(null);
  const command = useRelayCommand();
  const current = task.data?.view;
  const project = projects.data?.projects.find(project => project.project_id === view.task.project_id);
  const denied = !project || !current || current.task.state !== 'ready' || projects.stale || task.stale || receipts.stale || !receipts.data || relayUnresolved(receipts.data.receipts, 'dispatch');
  const changed = project && project.project_settings_revision !== current?.task.project_settings_revision;
  const checked = !!project && confirmedRevision === project.project_settings_revision;
  const locked = denied || changed || command.pending || command.uncertain;
  return <section className="relay-flow" aria-label="权限确认页">
    <h3>启动前确认权限</h3>
    <Field label="任务 ID">{view.task.task_id}</Field><Field label="Agent">{view.task.agent}</Field><Field label="任务描述">{view.task.description}</Field>
    <QueryState state={projects}>{() => project ? <>
      <Field label="项目根目录"><code>{project.root}</code></Field>
      <Field label="现有项目权限">{project.permission_level} · 版本 {project.project_settings_revision}</Field>
      <Field label="B · 明确目录许可">{project.allow_directory ? '已获准' : '未获准'}</Field><Field label="C · 引用展开许可">{project.expand_references ? '已获准' : '未获准'}</Field>
      <Field label="允许目录 / 文件">{project.allowed_subdirs.length ? project.allowed_subdirs.join('、') : '未许可额外目录 / 文件'}</Field>
      <Field label="允许动作">{project.allowed_actions.length ? project.allowed_actions.join('、') : '尚无获准动作'}</Field>
      <p className="relay-note">A：主动纳入的固定版本资料；B：明确获准目录 / 文件；C：展开获准资料引用。B / C 独立，此页沿用项目现有许可。资料许可不授权 CLI 自由工具；中性工作区执行权限尚待用户裁决，由服务端拒绝未获准请求。</p>
    </> : <p role="alert">项目许可无法读取，不能确认启动。</p>}</QueryState>
    <QueryState state={task}>{data => <p>服务端当前状态：{data.view.task.state} · {data.view.task.condition || '无附加状态'}</p>}</QueryState>
    <section aria-label="派发与意图收据"><h3>派发与意图收据</h3><QueryState state={receipts}>{data => <RelayReceipts receipts={data.receipts.filter(receipt => ['intent', 'dispatch'].includes(receipt.action))} />}</QueryState></section>
    {changed && <p role="alert">项目权限已变化，任务的权限版本已过期。请核对项目许可和保存的意图，当前不能启动。</p>}
    <form onSubmit={event => {
      event.preventDefault();
      if (locked || !checked || !project) return;
      void command.run(key => relayApi.confirm(view.task.task_id, { permission_confirmed: true, permission_level: project.permission_level, project_settings_revision: project.project_settings_revision }, key), () => { task.retry(); receipts.retry(); onChanged(); });
    }}><fieldset disabled={!!locked}><legend>此次接力的权限范围</legend>
      <label>权限级别<input aria-label="权限级别" readOnly value={project?.permission_level || '未获取'} /></label>
      <p className="relay-note">新项目默认为 A。此处确认现有项目许可摘要；修改 B / C 范围需回到项目设置，由你另行保存。</p>
      <label className="relay-check"><input type="checkbox" checked={checked} onChange={event => setConfirmedRevision(event.target.checked && project ? project.project_settings_revision : null)} />我已核对根目录、资料权限和允许动作，确认请求此 Agent 接力</label>
      <button type="submit" className="ac-button" disabled={!checked}>确认资料权限并请求接力</button>
    </fieldset></form>
    {command.error && <p role="alert">{command.error} · 操作 {command.operationID}</p>}
    {command.existingTaskID && <RelayExistingTask key={command.existingTaskID} taskID={command.existingTaskID} />}
    {command.uncertain && <p role="alert">派发结果未知，禁止重派。只读取任务和收据核对结果。</p>}
    {command.result && <p role="status">派发收据 {command.result.receipt.receipt_id} · {relayReceiptLabel[command.result.receipt.status]}</p>}
    <button type="button" className="ac-button secondary compact" onClick={() => { projects.retry(); task.retry(); receipts.retry(); }}>读取权限、任务和收据</button>
  </section>;
}
