import { Field, FieldRow, MutedValue } from '../../../components/DetailPanel';
import { StatusBadge } from '../../../components/StatusBadge';
import type { components } from '../../../api/schema';
import { missionStatusLabel, workItemStatusLabel } from '../../../utils/format';
import styles from '../SwarmPage.module.css';
import detailStyles from './MissionWorkspace.module.css';

type S = components['schemas'];
type Mission = S['SwarmMissionV1'];
type WorkItem = S['SwarmWorkItemV1'];
type Artifact = S['SwarmArtifactV1'];
type Lease = S['SwarmLeaseV1'];
type Execution = S['SwarmExecutionV1'];

export function MissionList({ items, onSelect }: { items: Mission[]; onSelect: (item: Mission) => void }) {
  return <ul className={styles.list}>{items.map(item => <li key={item.id}>
    <button type="button" className={`${styles.listItem} ac-mission-card`} aria-label={`查看任务 ${item.grant.goal}`} onClick={() => onSelect(item)}>
      <span className={styles.itemHeader}><strong className={`${styles.itemTitle} line-clamp-2`}>{item.grant.goal}</strong><StatusBadge value={item.state} label={item.state === 'ready' ? '待执行' : missionStatusLabel(item.state)} /></span>
      <span className={styles.itemMeta}><span>v{item.version}</span><span>Grant: {item.grant.id}</span><span>{item.work_item_count} 工作项</span><span>项目 {item.grant.project_id}</span></span>
      {item.reason && <span className={styles.blockerRow}>{item.reason}</span>}
    </button>
  </li>)}</ul>;
}

export function MissionDetail({ item, workItems, artifacts }: { item: Mission; workItems?: WorkItem[]; artifacts?: Artifact[] }) {
  const grant = item.grant;
  return <section aria-label="任务账本与批准快照">
    <Field label="目标">{grant.goal}</Field>
    <FieldRow><Field label="状态"><StatusBadge value={item.state} label={item.state === 'ready' ? '待执行' : missionStatusLabel(item.state)} /></Field><Field label="版本">v{item.version}</Field></FieldRow>
    <Field label="任务 ID"><code>{item.id}</code></Field>
    <Field label="Grant ID"><code>{grant.id}</code> · Epoch {grant.epoch}</Field>
    <Field label="批准的提案版本"><code>{grant.proposal_id}</code> · r{grant.proposal_revision}</Field>
    <Field label="项目"><code>{grant.project_id}</code></Field>
    <Field label="允许根目录"><code>{grant.root}</code></Field>
    <Field label="允许修改路径">{grant.write_paths?.join('；') || <MutedValue>无</MutedValue>}</Field>
    <Field label="允许动作">{grant.actions?.join('；') || <MutedValue>无</MutedValue>}</Field>
    <Field label="固定输入版本">{grant.inputs?.length ? <ul>{grant.inputs.map(input => <li key={`${input.id}@${input.revision}`}><code>{input.id}</code> · {input.revision}</li>)}</ul> : <MutedValue>无</MutedValue>}</Field>
    <Field label="批准有效期">{grant.expires_at}</Field>
    <ResourceRecord label="资源上限" resources={grant.budget} />
    <ResourceRecord label="当前预留" resources={item.reserved} />
    <ResourceRecord label="实际用量" resources={item.used} />
    <Field label="停止条件"><ul><li>截止时间：{grant.limits.deadline}</li><li>最大工作项数：{grant.limits.max_work_items}</li><li>每项最大尝试次数：{grant.limits.max_attempts}</li><li>最大并发数：{grant.limits.max_active}</li><li>连续无进展轮数上限：{grant.limits.max_no_progress}</li><li>无进展时限（秒）：{grant.limits.no_progress_seconds}</li></ul></Field>
    <Field label="工作项">{workItems ? workItems.length ? <ul>{workItems.map(work => <li key={work.id}><code>{work.id}</code> · {work.title} · {workItemStatusLabel(work.state)}</li>)}</ul> : <MutedValue>无工作项</MutedValue> : <MutedValue>共 {item.work_item_count} 项；详情尚未读取</MutedValue>}</Field>
    <Field label="产物引用">{artifacts ? artifacts.length ? <ul>{artifacts.map(artifact => <li key={artifact.id}><code>{artifact.id}</code> · v{artifact.version} · {workItemStatusLabel(artifact.state)}</li>)}</ul> : <MutedValue>无</MutedValue> : <MutedValue>详情尚未读取</MutedValue>}</Field>
    <Field label="阻塞或停止原因">{item.reason || <MutedValue>无</MutedValue>}</Field>
    <Field label="连续无进展轮数">{item.no_progress}</Field>
    <Field label="最近有效进展">{item.last_progress_at}</Field>
    <Field label="已核查操作数">{item.settled_operations}</Field>
    <Field label="创建时间">{item.created_at}</Field>
  </section>;
}

export function ResourceRecord({ label, resources }: { label: string; resources: S['SwarmBudgetV1'] | S['SwarmResourcesV1'] | S['SwarmUsageV1'] }) {
  return <Field label={label}><dl className={detailStyles.resources}>
    <div><dt>调用量</dt><dd>{resources.calls === null ? '未知' : resources.calls}</dd></div>
    <div><dt>Token</dt><dd>{resources.tokens === null ? '未知' : resources.tokens}</dd></div>
    <div><dt>费用</dt><dd>{resources.money === null ? '未知' : `${resources.money} ${resources.currency}（最小货币单位）`}</dd></div>
  </dl></Field>;
}

export function WorkItemList({ items, missions, leases, onSelect }: { items: WorkItem[]; missions?: Mission[]; leases?: Lease[]; onSelect: (item: WorkItem, missionId: string) => void }) {
  return <ul className={styles.list}>{items.map(item => {
    const mission = missions?.find(value => value.id === item.mission_id);
    const lease = leases?.find(value => value.work_item_id === item.id && value.fence === item.fence);
    return <li key={item.id}><button type="button" className={styles.listItem} aria-label={`查看工作项 ${item.title}`} onClick={() => onSelect(item, item.mission_id)}>
      <span className={styles.itemHeader}><strong className={styles.itemTitle}>{item.title}</strong><StatusBadge value={item.state} label={workItemStatusLabel(item.state)} /></span>
      <code>{item.id}</code><span className={styles.itemMeta}><span>v{item.version}</span><span>尝试 {item.attempts}</span><span>{lease ? `持有者：${lease.holder_id}` : item.state === 'open' ? '未认领' : '查看当前租约'}</span><span>{mission ? mission.grant.goal : item.mission_id}</span></span>
      {item.reason && <span className={styles.blockerRow}>{item.reason}</span>}
    </button></li>;
  })}</ul>;
}

export function WorkItemDetail({ item, missionId, lease, executions }: { item: WorkItem; missionId: string; lease?: Lease; executions?: Execution[] }) {
  if (item.mission_id !== missionId) throw new Error('工作项所属任务与当前任务不一致');
  if (lease && (lease.work_item_id !== item.id || lease.mission_id !== missionId || lease.fence !== item.fence)) throw new Error('当前租约与工作项身份或 fencing token 不一致');
  return <section aria-label="工作项详情">
    <Field label="工作项标题">{item.title}</Field>
    <FieldRow><Field label="状态"><StatusBadge value={item.state} label={workItemStatusLabel(item.state)} /></Field><Field label="版本">v{item.version}</Field></FieldRow>
    <Field label="工作项 ID"><code>{item.id}</code></Field><Field label="所属任务"><code>{missionId}</code></Field>
    <Field label="允许修改路径">{item.write_paths?.join('；') || <MutedValue>无</MutedValue>}</Field>
    <Field label="固定输入版本">{item.inputs?.length ? <ul>{item.inputs.map(input => <li key={`${input.id}@${input.revision}`}><code>{input.id}</code> · {input.revision}</li>)}</ul> : <MutedValue>无</MutedValue>}</Field>
    <Field label="依赖工作项">{item.dependencies?.join('；') || <MutedValue>无</MutedValue>}</Field>
    <Field label="所需能力">{item.required_capabilities?.join('；') || <MutedValue>无</MutedValue>}</Field>
    <Field label="采用的固定产物">{item.artifact_inputs?.length ? <ul>{item.artifact_inputs.map(input => <li key={`${input.id}@${input.version}`}><code>{input.id}</code> · v{input.version} · Commit {input.commit} · {input.object_ref}</li>)}</ul> : <MutedValue>无</MutedValue>}</Field>
    <Field label="尝试次数">{item.attempts}</Field><Field label="当前 fencing token">{item.fence}</Field>
    <Field label="当前持有者">{lease ? lease.holder_id : <MutedValue>{item.state === 'open' ? '未认领' : '当前无有效租约'}</MutedValue>}</Field>
    <Field label="上下文包 ID">{lease?.context_packet_id || <MutedValue>当前无有效租约上下文；历史引用见执行记录</MutedValue>}</Field>
    {lease && <><Field label="绑定及原生会话"><code>{lease.binding_id}</code> · <code>{lease.session_id}</code> · {lease.member_version}</Field><Field label="原操作 ID"><code>{lease.operation_id}</code></Field><Field label="执行工作树"><code>{lease.worktree}</code></Field><Field label="基准 Commit"><code>{lease.base_commit}</code></Field><Field label="租约有效期">{lease.expires_at}</Field><Field label="策略版本">{lease.policy_version}</Field><ResourceRecord label="租约资源预留" resources={lease.reserved} /></>}
    <Field label="产物引用">{item.artifact_id || <MutedValue>无</MutedValue>}</Field><Field label="阻塞或退回原因">{item.reason || <MutedValue>无</MutedValue>}</Field>
    <ExecutionRecords items={executions} />
  </section>;
}

export function ExecutionRecords({ items }: { items?: Execution[] }) {
  return <section aria-label="实际执行记录"><h4>实际执行记录</h4>{items ? items.length ? <ul className={detailStyles.list}>{items.map(item => <li className={detailStyles.record} key={item.operation_id}>
    <Field label="原操作 ID"><code>{item.operation_id}</code></Field><Field label="原生会话 ID"><code>{item.session_id}</code></Field><Field label="绑定 ID"><code>{item.binding_id}</code></Field><Field label="上下文包 ID">{item.context_packet_id || <MutedValue>未提供</MutedValue>}</Field><Field label="fencing token">{item.fence}</Field><Field label="策略版本">{item.policy_version}</Field>
    <MissionTimeline items={item.events ?? []} />
    <Field label="工具事件读取状态">{item.tool_events_status}</Field>
    {item.tool_events?.length ? <ul>{item.tool_events.map(event => <li key={event.event_id}><time>{event.occurred_at}</time> · {event.type} · {event.tool} · <code>{event.event_id}</code></li>)}</ul> : <p>当前没有可显示的工具事件引用。</p>}
  </li>)}</ul> : <p>暂无实际执行记录。</p> : <p>执行记录尚未读取。</p>}</section>;
}

export function MissionTimeline({ items }: { items: S['SwarmTimelineEventV1'][] }) {
  return <section aria-label="真实事件时间线"><h4>任务时间线</h4>{items.length ? <ol className={detailStyles.timeline}>{items.map(event => <li key={event.event_id}><time>{event.occurred_at}</time><p>{event.type} · v{event.aggregate_version}</p><p>对象 <code>{event.aggregate_id}</code></p><p>事件 <code>{event.event_id}</code></p></li>)}</ol> : <p>暂无已保存的任务事件。</p>}</section>;
}
