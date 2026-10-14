import { useState } from 'react';
import { relayApi } from '../../api/relay';
import { useReadApi } from '../../hooks/useReadApi';
import { QueryState } from '../../components/QueryState';
import { SectionCard } from '../../components/SectionCard';
import { DetailPanel } from '../../components/DetailPanel';
import { relayAgentAvailability } from './relayPresentation';
import { RelayTaskPanel, RelayTaskSummary } from './relayTaskPanel';

export function RelayOverview({ workNames }: { workNames?: Readonly<Record<string, string>> }) {
  const registry = useReadApi(signal => relayApi.agents(signal));
  const tasks = useReadApi(signal => relayApi.tasks(signal), { refreshMs: 3000 });
  const [taskID, setTaskID] = useState('');
  return <div className="relay-overview">
    <SectionCard title="注册 Agent" padded actions={<button className="ac-text-button" type="button" onClick={registry.retry}>读取注册表</button>}>
      <QueryState state={registry}>{data => data.agents.length ? <ul className="relay-agents">{data.agents.map(agent => <li key={agent.name}>
        <b>{agent.name}</b><span>驱动路径：{agent.drive_mode}</span><p>{relayAgentAvailability(agent, data.inventory)}</p>
        <dl>{Object.entries(agent.capabilities).map(([name, capability]) => <div key={name}><dt>{name}</dt><dd>{capability.status} · {capability.reason}{capability.checked_at ? ` · ${capability.checked_at}` : ''}</dd></div>)}</dl>
      </li>)}</ul> : <p>尚无数据注册的 Agent。</p>}</QueryState>
    </SectionCard>
    <SectionCard title="接力任务" padded actions={<button className="ac-text-button" type="button" onClick={tasks.retry}>只读取任务</button>}>
      <p className="relay-note">终端、任务和工作区状态均由接入层投影。刷新页面不派发任务。</p>
      <QueryState state={tasks}>{data => data.tasks.length ? <ul className="relay-tasks">{[...data.tasks].sort((a, b) => Date.parse(b.task.created_at) - Date.parse(a.task.created_at)).map(view => <li key={view.task.task_id}>
        <RelayTaskSummary view={view} workName={workNames?.[view.task.work_id]} /><button type="button" className="ac-button secondary compact" onClick={() => setTaskID(view.task.task_id)}>{view.task.state === 'ready' ? '查看意图与权限' : '查看任务、事件和收据'}</button>
      </li>)}</ul> : <p>尚无人工接力意图。</p>}</QueryState>
    </SectionCard>
    {taskID && <DetailPanel title={`接力任务 · ${taskID}`} onClose={() => setTaskID('')}><RelayTaskPanel key={taskID} taskID={taskID} onChanged={tasks.retry} /></DetailPanel>}
  </div>;
}
