import { useId, useMemo, useState } from 'react';
import { ApiError } from '../../../api/client';
import { swarmApi, type SwarmMission, type SwarmMissionDAG } from '../../../api/swarm';
import { useReadApi, type ReadApiState } from '../../../hooks/useReadApi';
import { QueryState } from '../../../components/QueryState';
import { layoutDAG, validateDAG } from './layout';
import styles from './MissionDAGView.module.css';

const states: Record<string, string> = { open: '开放', leased: '已认领', running: '执行中', submitted: '待验收', accepted: '已验收', revision_needed: '待修改', blocked: '受阻', cancelled: '已取消' };

export function MissionDAGView({ missions, onOpenWorkItem }: { missions: ReadApiState<{ items: SwarmMission[] }>; onOpenWorkItem: (id: string, workItemID: string) => void }) {
  const [missionID, setMissionID] = useState('');
  const dag = useReadApi(signal => swarmApi.dag(missionID, signal).then(validateDAG), {
    key: missionID, enabled: !!missionID, refreshMs: 5000,
    retainDataOnError: error => !(error instanceof ApiError && [401, 403, 404].includes(error.status)),
  });
  return <section aria-label="动态任务依赖">
    <p>依赖来自实际工作项，每5秒重新读取。执行往返与 Worker 观察保留在独立记录中。</p>
    <div className={styles.toolbar}>
      <label>任务<select aria-label="DAG任务" value={missionID} onChange={event => setMissionID(event.target.value)}><option value="">请选择任务</option>{missions.data?.items.map(mission => <option key={mission.id} value={mission.id}>{mission.grant.goal} · {mission.id}</option>)}</select></label>
      <button className="ac-button secondary compact" type="button" onClick={() => { missions.retry(); if (missionID) dag.retry(); }}>刷新任务依赖</button>
    </div>
    {!missions.data?.items.length && !missions.loading && <p role="status">暂无可读取的任务依赖。请先从已有批准建立任务。</p>}
    {missionID && <QueryState state={dag}>{data => <DAGCanvas key={missionID} data={data} disabled={dag.stale} onOpenWorkItem={id => onOpenWorkItem(missionID, id)} />}</QueryState>}
  </section>;
}

function DAGCanvas({ data, disabled, onOpenWorkItem }: { data: SwarmMissionDAG; disabled: boolean; onOpenWorkItem: (id: string) => void }) {
  const [direction, setDirection] = useState<'LR' | 'TB'>('LR');
  const [zoom, setZoom] = useState(1);
  const [selectedID, setSelectedID] = useState('');
  const layout = useMemo(() => layoutDAG(data, direction), [data, direction]);
  const marker = useId().replaceAll(':', '');
  const selected = data.nodes.find(node => node.work_item.id === selectedID);
  const byID = new Map(data.nodes.map(node => [node.work_item.id, node]));
  return <>
    <p role="status">工作项 {data.nodes.length} · 依赖 {data.edges.length} · Worker执行记录 {data.workers.length} · 任务版本 {data.mission_version} · {data.mission_state}</p>
    <p>服务观察时间 <time dateTime={data.observed_at}>{data.observed_at}</time>{data.reason && <> · 原因：{data.reason}</>}</p>
    <div className={styles.toolbar}>
      <label>布局方向<select aria-label="依赖布局方向" value={direction} onChange={event => setDirection(event.target.value as 'LR' | 'TB')}><option value="LR">从左到右</option><option value="TB">从上到下</option></select></label>
      <label>缩放<select aria-label="依赖布局缩放" value={zoom} onChange={event => setZoom(Number(event.target.value))}><option value={0.6}>60%</option><option value={0.8}>80%</option><option value={1}>100%</option><option value={1.25}>125%</option></select></label>
    </div>
    {data.nodes.length ? <div className={styles.canvas} role="region" aria-label="工作项依赖画布" tabIndex={0}>
      <div style={{ width: layout.width * zoom, height: layout.height * zoom }}>
        <div className={styles.plane} style={{ width: layout.width, height: layout.height, transform: `scale(${zoom})` }}>
          <svg width={layout.width} height={layout.height} aria-hidden="true"><defs><marker id={marker} markerWidth="8" markerHeight="8" refX="7" refY="4" orient="auto"><path d="M0,0 L8,4 L0,8 Z" fill="currentColor" /></marker></defs>{layout.paths.map(edge => <path key={`${edge.from}:${edge.to}`} d={edge.path} fill="none" stroke="currentColor" strokeWidth="2" markerEnd={`url(#${marker})`} />)}</svg>
          {layout.nodes.map(node => <button type="button" className={styles.node} key={node.work_item.id} aria-pressed={selectedID === node.work_item.id} onClick={() => setSelectedID(node.work_item.id)} style={{ left: node.x, top: node.y }}><strong>{node.work_item.title}</strong><span>{states[node.work_item.state] ?? node.work_item.state}</span></button>)}
        </div>
      </div>
    </div> : <p role="status">当前任务尚无工作项，依赖图为空。</p>}
    <div className={styles.details}>
      <section aria-label="依赖节点详情">
        <h3>节点详情</h3>
        {selected ? <>
          <h4>{selected.work_item.title}</h4><p><code>{selected.work_item.id}</code> · {states[selected.work_item.state] ?? selected.work_item.state} · v{selected.work_item.version}</p>
          <p>创建主体：{selected.work_item.creator_kind === 'human' ? '人工' : selected.work_item.creator_kind === 'agent' ? 'Agent主体（模型来源未核实）' : '来源未记录'} {selected.work_item.created_by}</p>
          <p>创建事件：{selected.creation_event ? <><code>{selected.creation_event.event_id}</code> · {selected.creation_event.occurred_at}</> : '来源未记录'}</p>
          <p>写入范围：{selected.work_item.write_paths?.join('、')}</p>
          <p>固定输入：{selected.work_item.inputs?.map(input => `${input.id}@${input.revision}`).join('、')}</p>
          <button className="ac-button secondary compact" type="button" disabled={disabled} onClick={() => onOpenWorkItem(selected.work_item.id)}>打开工作项详情</button>
        </> : <p>{selectedID ? '所选节点已不在当前快照中。' : '使用Tab选择节点，按Enter查看来源与工作项。'}</p>}
      </section>
      <section aria-label="任务依赖事实"><h3>依赖事实</h3>{data.edges.length ? <ul>{data.edges.map(edge => <li key={`${edge.from}:${edge.to}`}><button type="button" className="ac-link-button" onClick={() => setSelectedID(edge.from)}>{byID.get(edge.from)?.work_item.title}</button> → <button type="button" className="ac-link-button" onClick={() => setSelectedID(edge.to)}>{byID.get(edge.to)?.work_item.title}</button><p>required · 创建事件 <code>{edge.event_id || '来源未记录'}</code></p></li>)}</ul> : <p>暂无任务依赖。</p>}</section>
    </div>
    <section aria-label="Worker参与与退出历史"><h3>Worker参与与退出历史</h3>{data.workers.length ? <ul className={styles.workers}>{data.workers.map(worker => <li key={worker.execution.operation_id}>
      <p>Worker <code>{worker.execution.holder_id || worker.execution.session_id}</code> · operation <code>{worker.execution.operation_id}</code></p>
      <p>所属工作项 <code>{worker.execution.work_item_id}</code> · binding <code>{worker.execution.binding_id}</code> · fence {worker.execution.fence} <button type="button" className="ac-link-button" disabled={disabled} onClick={() => onOpenWorkItem(worker.execution.work_item_id)}>查看所属工作项</button></p>
      <p>原生观察：{worker.observation.state === 'exited' ? '确认退出' : worker.observation.state === 'observed' ? '原控制器观察可用' : '无法核实'} · 租约：{worker.lease_state === 'current' ? '有效' : worker.lease_state === 'expired' ? '已过期，退出状态独立核查' : worker.lease_state === 'retained' ? '预留保留，写权限无效' : '当前不存在'} · 资源释放：{worker.execution.release_confirmed ? '已有确认记录' : '未确认'}</p>
      <p>观察来源：{worker.observation.source} · {worker.observation.reason} · {worker.observation.observed_at}</p>
      <p>本轮输入：{worker.observation.current_input === null ? '当前身份未核实' : worker.observation.current_input ? '当前原生输入' : '历史原生输入'}</p>
      <p>实际退出观察时间：{worker.observation.exit_observed_at ?? '未知'}{worker.execution.exit_event && <> · 退出确认事件 <code>{worker.execution.exit_event.event_id}</code> · 确认记录时间 {worker.execution.exit_event.occurred_at}</>}</p>
      <p>停止原因：{worker.execution.stop_reason || '未知（缺少关联的停止请求）'}{worker.execution.stop_event && <> · 来源事件 <code>{worker.execution.stop_event.event_id}</code> · {worker.execution.stop_event.occurred_at}</>}</p>
      <p>历史对账说明：{worker.execution.reconciliation_reason || '暂无独立对账说明'}</p>
      <ul>{worker.execution.events?.map(event => <li key={event.event_id}>{event.type} · <code>{event.event_id}</code> · {event.occurred_at}</li>)}</ul>
    </li>)}</ul> : <p>暂无实际 Worker 执行记录；工作项创建不会产生执行或退出事实。</p>}</section>
    <details><summary>真实任务与协作事件 · {data.events.length}</summary><ul>{data.events.map(event => <li key={event.event_id}>{event.type} · <code>{event.event_id}</code> · 对象 <code>{event.aggregate_id}</code> · {event.occurred_at}</li>)}</ul></details>
  </>;
}
