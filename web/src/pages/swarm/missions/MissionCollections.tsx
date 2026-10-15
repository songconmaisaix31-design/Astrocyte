import { QueryState } from '../../../components/QueryState';
import { EmptyState } from '../../../components/EmptyState';
import type { ReadApiState } from '../../../hooks/useReadApi';
import type { SwarmMission, SwarmMissionDetail, SwarmWorkItem } from '../../../api/swarm';
import { matchesQuery } from '../../../utils/search';
import { MissionTimeline, WorkItemList } from './MissionRecords';
import styles from '../SwarmPage.module.css';

type Missions = ReadApiState<{ items: SwarmMission[] }>;
type Records = ReadApiState<SwarmMissionDetail[]>;

export function WorkItemCollection({ missions, records, query, onSelect }: { missions: Missions; records: Records; query: string; onSelect: (item: SwarmWorkItem, missionId: string) => void }) {
  return <QueryState state={missions} empty={data => !data.items.length} emptyTitle="暂无工作项">{() =>
    <QueryState state={records} empty={data => !data.some(detail => detail.work_items?.length)} emptyTitle="暂无工作项">{data => {
      const works = data.flatMap(detail => detail.work_items ?? []).filter(item => matchesQuery(query, [item.title, item.id, item.mission_id, item.reason]));
      return works.length ? <WorkItemList items={works} missions={data.map(detail => detail.mission)} leases={data.flatMap(detail => detail.leases ?? [])} onSelect={onSelect} /> : <EmptyState title="没有匹配的工作项" description="请调整搜索词" />;
    }}</QueryState>
  }</QueryState>;
}

export function ArtifactCollection({ missions, records, query, onSelect }: { missions: Missions; records: Records; query: string; onSelect: (id: string, missionId: string) => void }) {
  return <QueryState state={missions} empty={data => !data.items.length} emptyTitle="暂无产物">{() =>
    <QueryState state={records} empty={data => !data.some(detail => detail.artifacts?.length)} emptyTitle="暂无产物">{data => {
      const artifacts = data.flatMap(detail => (detail.artifacts ?? []).map(artifact => ({ artifact, missionId: detail.mission.id }))).filter(({ artifact }) => matchesQuery(query, [artifact.id, artifact.work_item_id, artifact.commit, artifact.experience_id]));
      return artifacts.length ? <ul className={`${styles.list} ac-artifact-list`}>{artifacts.map(({ artifact, missionId }) => <li key={`${missionId}:${artifact.id}`}>
        <button type="button" className={`${styles.listItem} ac-artifact-card`} aria-label={`查看产物 ${artifact.id}`} onClick={() => onSelect(artifact.id, missionId)}><strong>{artifact.id}</strong><span>v{artifact.version} · {artifact.state}</span><span>工作项 {artifact.work_item_id}</span><span>Commit {artifact.commit}</span>{artifact.experience_id && <span>经验引用 {artifact.experience_id}</span>}</button>
      </li>)}</ul> : <EmptyState title="没有匹配的产物" description="请调整搜索词" />;
    }}</QueryState>
  }</QueryState>;
}

export function MissionEventCollection({ missions, records }: { missions: Missions; records: Records }) {
  return <QueryState state={missions} empty={data => !data.items.length} emptyTitle="暂无任务事件">{() =>
    <QueryState state={records} empty={data => !data.some(detail => detail.timeline?.length)} emptyTitle="暂无任务事件">{data =>
      <MissionTimeline items={data.flatMap(detail => detail.timeline ?? []).sort((left, right) => left.occurred_at.localeCompare(right.occurred_at) || left.event_id.localeCompare(right.event_id))} />
    }</QueryState>
  }</QueryState>;
}
