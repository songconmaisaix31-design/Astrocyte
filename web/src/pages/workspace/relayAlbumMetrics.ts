import type { WorksSnapshot } from './worksAlbumPresentation';

/** Saved scanner facts and service-projected task evidence have separate meanings. */
export function relayAlbumMetrics(snapshot: WorksSnapshot, relayedWorkIDs?: readonly string[]) {
  const known = snapshot.observed_at !== null && snapshot.status !== 'unknown';
  const recentKnown = known && snapshot.works.every(work => work.files.verified && work.files.mtime !== null);
  const platformNames = new Set(snapshot.platforms.map(stat => stat.platform).filter(name => name && name !== 'unknown'));
  const workIDs = new Set(snapshot.works.map(work => work.id));
  return [
    { label: 'AI 平台', value: known ? platformNames.size : null },
    { label: '作品', value: known ? snapshot.works.length : null },
    { label: '文件夹', value: known ? snapshot.works.filter(work => work.kind === 'folder').length : null },
    { label: '接力过', value: known && relayedWorkIDs ? new Set(relayedWorkIDs.filter(id => workIDs.has(id))).size : null },
    { label: '最近在动', value: recentKnown ? snapshot.platforms.reduce((sum, stat) => sum + stat.recent, 0) : null },
  ];
}
