import { type ReadApiState, useReadApi } from '../../../hooks/useReadApi';
import { swarmApi, type SwarmMission } from '../../../api/swarm';

export function useMissionCollections(missions: ReadApiState<{ items: SwarmMission[] }>) {
  const items = missions.data?.items;
  const key = items?.map(item => `${item.id}@${item.version}`).join('|') ?? '';
  return useReadApi(signal => Promise.all((items ?? []).map(item => swarmApi.mission(item.id, signal))), { key, enabled: !!items?.length });
}
