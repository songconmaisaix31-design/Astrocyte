import type { components } from '../../api/schema';

export type WorksAttitude = 'continue' | 'park' | 'done';
export type WorksHandoff = components['schemas']['WorkHandoffV1'];
export type WorksFileFact = components['schemas']['WorkFileFactV1'];
export type WorksWork = components['schemas']['WorkV1'];
export type WorksPlatformStat = components['schemas']['WorkPlatformStatV1'];
export type WorksSnapshot = components['schemas']['WorksSnapshotV1'];

export const attitudeLabels: Record<WorksAttitude, string> = {
  continue: '想继续做',
  park: '暂时搁置',
  done: '已经收尾',
};

export const attitudeOptions: { value: WorksAttitude; label: string }[] = [
  { value: 'continue', label: attitudeLabels.continue },
  { value: 'park', label: attitudeLabels.park },
  { value: 'done', label: attitudeLabels.done },
];

export const attitudeLabel = (attitude: string) => attitudeLabels[attitude as WorksAttitude] ?? '尚未标记';

export const platformLabel = (platform: string) => platform && platform !== 'unknown' ? platform : '平台未知';

export const handoffSourceLabel = (source: string) => (source === 'generated' ? 'AI 生成' : source === 'human' ? '人工编写' : '来源未知');

export interface WorksFilters {
  view: 'flat' | 'platform' | 'group' | 'timeline';
  platform: string;
  group: string;
  attitude: string;
  includeHidden: boolean;
  sort: 'name' | 'recent' | 'updated';
  search: string;
}

export const initialWorksFilters: WorksFilters = {
  view: 'flat', platform: 'all', group: 'all', attitude: 'all', includeHidden: false, sort: 'recent', search: '',
};

/** 默认过滤已隐藏作品；用户明确选择后才重新显示。 */
export function filterWorks(items: WorksWork[], filters: WorksFilters) {
  const query = filters.search.trim().toLocaleLowerCase();
  return items
    .filter(item =>
      (filters.includeHidden || !item.hidden) &&
      (filters.platform === 'all' || item.platform === filters.platform) &&
      (filters.group === 'all' || item.group === filters.group) &&
      (filters.attitude === 'all' || item.attitude === filters.attitude) &&
      (!query || [item.name, item.platform, item.group, item.root, item.handoff.content].some(value => (value ?? '').toLocaleLowerCase().includes(query))))
    .sort((left, right) => {
      if (filters.sort === 'name') return left.name.localeCompare(right.name, 'zh-CN');
      if (filters.sort === 'updated') return (Date.parse(right.updated_at ?? '') || 0) - (Date.parse(left.updated_at ?? '') || 0);
      return (Date.parse(right.files.mtime ?? '') || 0) - (Date.parse(left.files.mtime ?? '') || 0) || left.name.localeCompare(right.name, 'zh-CN');
    });
}

/** 根据所选视图分组；平铺时保留一个分组。 */
export function worksSections(items: WorksWork[], view: WorksFilters['view']) {
  const groups = new Map<string, WorksWork[]>();
  for (const item of items) {
    const label = view === 'platform' ? platformLabel(item.platform)
      : view === 'group' ? item.group || '未分组'
      : view === 'timeline' ? item.files.mtime?.slice(0, 10) || '活动时间未知'
      : '全部作品';
    groups.set(label, [...(groups.get(label) ?? []), item]);
  }
  return [...groups].map(([label, works]) => ({ label, works }));
}

/** 文件核实标记取自服务明确提供的观察值。 */
export const fileVerifiedLabel = (fact: WorksFileFact) => fact.verified ? '从文件核实' : fact.reason || '待核实';
