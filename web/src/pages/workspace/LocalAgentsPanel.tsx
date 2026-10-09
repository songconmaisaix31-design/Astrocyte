import type { components } from '../../api/schema';
import type { ReadApiState } from '../../hooks/useReadApi';
import { QueryState } from '../../components/QueryState';
import { SectionCard } from '../../components/SectionCard';
import { CollectionOverview } from '../../components/CollectionOverview';
import { formatDateTime } from '../../utils/format';
import styles from './LocalAgentsPanel.module.css';

type Agent = components['schemas']['LocalAgentV1'];
type Observation = components['schemas']['AgentObservationV1'] | components['schemas']['NativeCapabilityObservationV1'];
const capabilityNames: Record<keyof Agent['capabilities'], string> = {
  discover: '发现会话', read_context: '读取上下文', start: '启动', resume: '原生接续',
  send: '发送输入', stop: '停止', observe: '观察状态', reconcile: '核对结果',
};
const statusNames: Record<Observation['status'], string> = { available: '可用', unavailable: '不可用', unknown: '未知', unsupported: '不支持', supported: '已支持' };
const reasonNames: Record<string, string> = {
  native_runtime_not_tested: '尚未验证原生能力',
  configuration_not_inspected: '未检查登录与配置',
  cli_not_probed: '尚未检查客户端',
  cli_not_found: '未找到 CLI 入口',
  executable_not_found: '未找到 CLI 入口',
};

export function LocalAgentsPanel({ state, fixture }: { state: ReadApiState<components['schemas']['LocalAgentListV1']>; fixture: boolean }) {
  return <SectionCard title="本机 Agent 清单" tabs={['overview', 'sessions']} padded actions={!fixture && <button type="button" className="ac-button secondary compact" disabled={state.loading} onClick={state.retry}>刷新清单</button>}>
    <p className={styles.note}>安装、配置、可启动和原生能力分别核实。清单与版本不代表已连接项目或运行中的会话。</p>
    {fixture ? <p className={styles.note}>示例模式不读取本机 Agent 清单。</p> : <QueryState state={state} empty={data => !data.items.length} emptyTitle="暂无本机 Agent 记录">{data => <>
      <CollectionOverview title="客户端概览" description="显示服务最近保存的探测结果。刷新清单不会启动 Agent。" metrics={[
        { label: '已加载客户端', value: data.items.length },
        { label: '已核实安装', value: data.items.filter(agent => agent.installed.status === 'available').length },
        { label: '已核实配置', value: data.items.filter(agent => agent.configured.status === 'available').length },
        { label: '已核实可启动', value: data.items.filter(agent => agent.startable.status === 'available').length },
      ]} />
      {data.next_cursor && <p className={styles.note}>还有未加载客户端，统计仅覆盖本页。</p>}
      <ul className={styles.agents}>{data.items.map(agent => <li key={agent.id}>
        <div className={styles.heading}><h3>{agent.display_name}</h3><span>版本 · {agent.version ?? '未提供'}</span></div>
        <dl className={styles.readiness}>{([['installed', '安装'], ['configured', '配置'], ['startable', '可启动']] as const).map(([key, label]) => <div key={key}><dt>{label}</dt><dd><ObservationValue observation={agent[key]} /></dd></div>)}</dl>
        <details><summary>原生能力 · {Object.values(agent.capabilities).filter(observation => observation.status === 'supported').length}/8 已支持</summary>
          <dl className={styles.capabilities}>{(Object.keys(capabilityNames) as (keyof Agent['capabilities'])[]).map(key => <div key={key}><dt>{capabilityNames[key]}</dt><dd><ObservationValue observation={agent.capabilities[key]} /></dd></div>)}</dl>
        </details>
      </li>)}</ul>
    </>}</QueryState>}
  </SectionCard>;
}

function ObservationValue({ observation }: { observation: Observation }) {
  const reason = reasonNames[observation.reason];
  return <><strong>{statusNames[observation.status]}</strong><small title={observation.reason}>{reason ?? (observation.reason ? '服务返回的原因见诊断详情' : '未提供原因')}</small>
    {!reason && observation.reason && <details><summary>诊断详情</summary><code>{observation.reason}</code></details>}
    <small>核实时间 · {observation.checked_at ? formatDateTime(observation.checked_at) : '未核实'}</small></>;
}
